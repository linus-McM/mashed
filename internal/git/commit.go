package git

import "context"

// StageAll runs `git add -A` and returns the combined output.
func StageAll(ctx context.Context, repoPath string) (string, error) {
	return combined(ctx, repoPath, "add", "-A")
}

// StagedStat returns `git diff --cached --stat` (empty when nothing is staged).
func StagedStat(ctx context.Context, repoPath string) (string, error) {
	return output(ctx, repoPath, "diff", "--cached", "--stat")
}

// StagedDiff returns the full staged diff (`git diff --cached`).
func StagedDiff(ctx context.Context, repoPath string) (string, error) {
	return output(ctx, repoPath, "diff", "--cached")
}

// Commit runs `git commit -m <msg>` and returns the combined output, which
// callers inspect for "nothing to commit" and hook failures.
func Commit(ctx context.Context, repoPath, msg string) (string, error) {
	return combined(ctx, repoPath, "commit", "-m", msg)
}
