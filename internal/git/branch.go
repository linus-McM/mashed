package git

import (
	"context"
	"os/exec"
	"strings"
)

// command builds `git -C repoPath <args...>` bound to ctx. It is shared by
// every ctx-taking helper in branch.go, commit.go, remote.go and files.go.
func command(ctx context.Context, repoPath string, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, "git", append([]string{"-C", repoPath}, args...)...)
}

// output runs git and returns its stdout. The error is the raw exec error so
// callers keep their own message wording.
func output(ctx context.Context, repoPath string, args ...string) (string, error) {
	out, err := command(ctx, repoPath, args...).Output()
	return string(out), err
}

// combined runs git and returns stdout+stderr. Callers inspect this text
// (e.g. "nothing to commit", "non-fast-forward"), so it is returned even
// when err is non-nil.
func combined(ctx context.Context, repoPath string, args ...string) (string, error) {
	out, err := command(ctx, repoPath, args...).CombinedOutput()
	return string(out), err
}

// Branch is a local branch as listed by ListBranches.
type Branch struct {
	Name    string
	Current bool
}

// ListBranches returns the local branches of repoPath, marking the checked-out one.
func ListBranches(ctx context.Context, repoPath string) ([]Branch, error) {
	out, err := output(ctx, repoPath, "branch", "--format=%(refname:short)\t%(HEAD)")
	if err != nil {
		return nil, err
	}
	var branches []Branch
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		branches = append(branches, Branch{
			Name:    parts[0],
			Current: len(parts) > 1 && strings.TrimSpace(parts[1]) == "*",
		})
	}
	return branches, nil
}

// Checkout runs `git checkout <ref> --`; the trailing "--" keeps ref from
// being read as a pathspec. It returns the combined output. Callers must
// validate ref (ValidateBranchName) first.
func Checkout(ctx context.Context, repoPath, ref string) (string, error) {
	return combined(ctx, repoPath, "checkout", ref, "--")
}

// CheckoutBranch runs `git checkout <branch>` with no separator. It is used
// only to return to a branch name git itself reported (merge-failure restore).
func CheckoutBranch(ctx context.Context, repoPath, branch string) error {
	return command(ctx, repoPath, "checkout", branch).Run()
}

// CreateBranch runs `git checkout -b <name>` and returns the combined output.
func CreateBranch(ctx context.Context, repoPath, name string) (string, error) {
	return combined(ctx, repoPath, "checkout", "-b", name)
}

// CurrentBranch returns the abbreviated name of HEAD ("HEAD" when detached).
func CurrentBranch(ctx context.Context, repoPath string) (string, error) {
	out, err := output(ctx, repoPath, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// VerifyRef returns nil when ref resolves via `git rev-parse --verify`.
func VerifyRef(ctx context.Context, repoPath, ref string) error {
	return command(ctx, repoPath, "rev-parse", "--verify", ref).Run()
}

// Merge runs `git merge <branch>` and returns the combined output.
func Merge(ctx context.Context, repoPath, branch string) (string, error) {
	return combined(ctx, repoPath, "merge", branch)
}

// MergeAbort runs `git merge --abort`.
func MergeAbort(ctx context.Context, repoPath string) error {
	return command(ctx, repoPath, "merge", "--abort").Run()
}
