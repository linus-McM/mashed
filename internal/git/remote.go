package git

import (
	"context"
	"fmt"
	"strings"
)

// Push runs `git push -u origin HEAD` and returns the combined output, which
// callers inspect for "non-fast-forward" / "rejected" / "fetch first".
func Push(ctx context.Context, repoPath string) (string, error) {
	return combined(ctx, repoPath, "push", "-u", "origin", "HEAD")
}

// ForcePush runs `git push --force-with-lease -u origin HEAD`.
func ForcePush(ctx context.Context, repoPath string) (string, error) {
	return combined(ctx, repoPath, "push", "--force-with-lease", "-u", "origin", "HEAD")
}

// Pull runs `git pull` and returns the combined output.
func Pull(ctx context.Context, repoPath string) (string, error) {
	return combined(ctx, repoPath, "pull")
}

// AheadBehind returns how many commits HEAD is ahead of and behind its
// upstream. It errors when there is no upstream.
func AheadBehind(ctx context.Context, repoPath string) (ahead, behind int, err error) {
	out, err := output(ctx, repoPath, "rev-list", "--left-right", "--count", "HEAD...@{upstream}")
	if err != nil {
		return 0, 0, err
	}
	parts := strings.Fields(strings.TrimSpace(out))
	if len(parts) != 2 {
		return 0, 0, &GitError{Op: "rev-list", RepoPath: repoPath, Err: fmt.Errorf("unexpected output %q", out)}
	}
	fmt.Sscanf(parts[0], "%d", &ahead)
	fmt.Sscanf(parts[1], "%d", &behind)
	return ahead, behind, nil
}

// PorcelainStatus returns `git status --porcelain` (empty when clean).
func PorcelainStatus(ctx context.Context, repoPath string) (string, error) {
	return output(ctx, repoPath, "status", "--porcelain")
}
