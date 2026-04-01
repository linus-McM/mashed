package git

import "fmt"

// GitError wraps errors from git operations with operation name and repo context.
type GitError struct {
	Op       string
	RepoPath string
	Err      error
}

func (e *GitError) Error() string {
	return fmt.Sprintf("git %s on %s: %v", e.Op, e.RepoPath, e.Err)
}

func (e *GitError) Unwrap() error { return e.Err }
