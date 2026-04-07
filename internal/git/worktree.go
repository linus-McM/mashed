package git

import (
	"mashed/internal/domain"
	"os/exec"
	"strings"
)

// DetectWorktrees returns all worktrees for a git repo at repoPath.
// It runs `git worktree list --porcelain` and parses the output into WorktreeInfo structs.
// If a worktree's branch no longer exists, IsOrphaned is set to true.
func DetectWorktrees(repoPath string) ([]domain.WorktreeInfo, error) {
	cmd := exec.Command("git", "-C", repoPath, "worktree", "list", "--porcelain")
	out, err := cmd.Output()
	if err != nil {
		return nil, &GitError{Op: "worktree-list", RepoPath: repoPath, Err: err}
	}

	raw := string(out)
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}

	// Split into blocks separated by blank lines.
	blocks := splitWorktreeBlocks(raw)

	var worktrees []domain.WorktreeInfo
	for _, block := range blocks {
		wt, ok := parseWorktreeBlock(block)
		if !ok {
			continue
		}

		// Check if the branch still exists (skip if detached HEAD / no branch).
		if wt.Branch != "" {
			orphaned, err := isBranchOrphaned(repoPath, wt.Branch)
			if err != nil {
				// If we can't verify, assume not orphaned.
				orphaned = false
			}
			wt.IsOrphaned = orphaned
		}

		worktrees = append(worktrees, wt)
	}

	return worktrees, nil
}

// splitWorktreeBlocks splits porcelain output into blocks separated by blank lines.
func splitWorktreeBlocks(raw string) [][]string {
	var blocks [][]string
	var current []string

	for _, line := range strings.Split(raw, "\n") {
		if line == "" {
			if len(current) > 0 {
				blocks = append(blocks, current)
				current = nil
			}
			continue
		}
		current = append(current, line)
	}
	if len(current) > 0 {
		blocks = append(blocks, current)
	}

	return blocks
}

// parseWorktreeBlock parses a single porcelain block into a WorktreeInfo.
// Each block contains lines like:
//
//	worktree /path/to/worktree
//	HEAD abc123def
//	branch refs/heads/branch-name
//
// If the "branch" line is absent, the worktree is in detached HEAD state.
func parseWorktreeBlock(lines []string) (domain.WorktreeInfo, bool) {
	var wt domain.WorktreeInfo
	found := false

	for _, line := range lines {
		if strings.HasPrefix(line, "worktree ") {
			wt.Path = strings.TrimPrefix(line, "worktree ")
			found = true
		} else if strings.HasPrefix(line, "branch ") {
			ref := strings.TrimPrefix(line, "branch ")
			// Strip refs/heads/ prefix to get the short branch name.
			wt.Branch = strings.TrimPrefix(ref, "refs/heads/")
		}
	}

	return wt, found
}

// isBranchOrphaned checks whether a branch still exists in the repository.
func isBranchOrphaned(repoPath, branch string) (bool, error) {
	cmd := exec.Command("git", "-C", repoPath, "rev-parse", "--verify", "refs/heads/"+branch)
	err := cmd.Run()
	if err != nil {
		// If rev-parse fails, the branch does not exist.
		return true, nil
	}
	return false, nil
}
