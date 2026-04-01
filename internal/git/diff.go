package git

import (
	"conductor/internal/domain"
	"os/exec"
	"strconv"
	"strings"
)

// ScopedDiff returns the files changed in a working directory relative to HEAD.
// It combines `git diff --stat HEAD` (staged + unstaged changes) with
// `git ls-files --others --exclude-standard` (untracked files).
func ScopedDiff(dir string) (*domain.ScopedDiff, error) {
	var files []domain.DiffFileStat

	// 1. Get tracked file changes via git diff --stat HEAD.
	diffFiles, err := parseDiffStat(dir)
	if err != nil {
		return nil, err
	}
	files = append(files, diffFiles...)

	// 2. Get untracked files via git ls-files.
	untrackedFiles, err := parseUntrackedFiles(dir)
	if err != nil {
		return nil, err
	}
	files = append(files, untrackedFiles...)

	return &domain.ScopedDiff{Files: files}, nil
}

// parseDiffStat runs `git diff --stat HEAD` and parses each line.
// Lines look like:
//
//	path/to/file | 10 ++++------
//	path/to/bin  | Bin 0 -> 1234 bytes
//
// The last line is a summary like " 3 files changed, ..." which we skip.
func parseDiffStat(dir string) ([]domain.DiffFileStat, error) {
	cmd := exec.Command("git", "-C", dir, "diff", "--stat", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return nil, &GitError{Op: "diff-stat", RepoPath: dir, Err: err}
	}

	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return nil, nil
	}

	lines := strings.Split(raw, "\n")
	var files []domain.DiffFileStat

	for _, line := range lines {
		// Skip the summary line (e.g., " 3 files changed, 10 insertions(+), 5 deletions(-)")
		if strings.Contains(line, "files changed") || strings.Contains(line, "file changed") {
			continue
		}

		fs, ok := parseDiffStatLine(line)
		if ok {
			files = append(files, fs)
		}
	}

	return files, nil
}

// parseDiffStatLine parses a single --stat line into a DiffFileStat.
// Format: " path | N +++---" or " path | Bin X -> Y bytes"
func parseDiffStatLine(line string) (domain.DiffFileStat, bool) {
	// Split on " | " to separate path from stats.
	parts := strings.SplitN(line, " | ", 2)
	if len(parts) != 2 {
		return domain.DiffFileStat{}, false
	}

	path := strings.TrimSpace(parts[0])
	stat := strings.TrimSpace(parts[1])

	if path == "" {
		return domain.DiffFileStat{}, false
	}

	fs := domain.DiffFileStat{Path: path}

	// Check for binary files: "Bin X -> Y bytes"
	if strings.HasPrefix(stat, "Bin") {
		fs.IsBinary = true
		return fs, true
	}

	// Parse text stat: "N +++---" where N is the total changes count.
	// Count + and - characters in the visual bar.
	fs.Added = strings.Count(stat, "+")
	fs.Removed = strings.Count(stat, "-")

	// Also try to parse the numeric count for accuracy when the bar is truncated.
	// The stat starts with a number like "10 +++---".
	statParts := strings.Fields(stat)
	if len(statParts) >= 2 {
		if total, err := strconv.Atoi(statParts[0]); err == nil {
			// The bar may be truncated. Use the numeric total and bar ratio for accuracy.
			plusCount := strings.Count(stat, "+")
			minusCount := strings.Count(stat, "-")
			barTotal := plusCount + minusCount
			if barTotal > 0 {
				fs.Added = total * plusCount / barTotal
				fs.Removed = total * minusCount / barTotal
			}
		}
	}

	return fs, true
}

// parseUntrackedFiles runs `git ls-files --others --exclude-standard` and returns
// each untracked file as a DiffFileStat with IsNew=true.
func parseUntrackedFiles(dir string) ([]domain.DiffFileStat, error) {
	cmd := exec.Command("git", "-C", dir, "ls-files", "--others", "--exclude-standard")
	out, err := cmd.Output()
	if err != nil {
		return nil, &GitError{Op: "ls-files", RepoPath: dir, Err: err}
	}

	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return nil, nil
	}

	lines := strings.Split(raw, "\n")
	files := make([]domain.DiffFileStat, 0, len(lines))

	for _, line := range lines {
		path := strings.TrimSpace(line)
		if path == "" {
			continue
		}
		files = append(files, domain.DiffFileStat{
			Path:  path,
			IsNew: true,
		})
	}

	return files, nil
}
