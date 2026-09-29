package git

import (
	"context"
	"path/filepath"
	"sort"
	"strings"
)

// ListFiles returns tracked plus untracked non-ignored files, sorted
// (`git ls-files --cached --others --exclude-standard`).
func ListFiles(ctx context.Context, repoPath string) ([]string, error) {
	out, err := output(ctx, repoPath, "ls-files", "--cached", "--others", "--exclude-standard")
	if err != nil {
		return nil, err
	}
	var files []string
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			files = append(files, line)
		}
	}
	sort.Strings(files)
	return files, nil
}

// FileDiff returns `git diff HEAD -- <rel>`. The caller must already have
// validated rel as repo-relative.
func FileDiff(ctx context.Context, repoPath, rel string) (string, error) {
	return output(ctx, repoPath, "diff", "HEAD", "--", rel)
}

// NoIndexDiff diffs /dev/null against abs (`git diff --no-index -- /dev/null
// <abs>`), showing an untracked file as all-added. git exits 1 when the files
// differ, so callers should use the output even when err != nil. The caller
// must already have resolved abs inside the repo.
func NoIndexDiff(ctx context.Context, repoPath, abs string) (string, error) {
	return output(ctx, repoPath, "diff", "--no-index", "--", "/dev/null", abs)
}

// ShowAtHead returns the content of rel at HEAD (`git show HEAD:<rel>`).
func ShowAtHead(ctx context.Context, repoPath, rel string) (string, error) {
	return output(ctx, repoPath, "show", "HEAD:"+filepath.ToSlash(rel))
}

// BranchDiff returns `git diff <base>...HEAD`.
func BranchDiff(ctx context.Context, repoPath, base string) (string, error) {
	return output(ctx, repoPath, "diff", base+"...HEAD")
}
