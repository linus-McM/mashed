package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"mashed/internal/domain"
	"mashed/internal/fsutil"
	"mashed/internal/git"
	"mashed/internal/pathguard"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// BranchInfo represents a single local branch returned by GitListBranches.
type BranchInfo struct {
	Name    string `json:"name"`
	Current bool   `json:"current"`
}

// RepoStatusInfo represents the git status of a repository.
type RepoStatusInfo struct {
	Dirty     bool `json:"dirty"`
	OpenPRs   int  `json:"openPRs"`
	Ahead     int  `json:"ahead"`
	Behind    int  `json:"behind"`
	Protected bool `json:"protected"`
}

// RepoChoice represents a repository available for selection.
type RepoChoice struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	Branch string `json:"branch"`
}

// RepoMtimes returns the modification times of .git/index and the repo root directory.
// The frontend polls this cheaply to detect when a full refresh is needed.
func (a *App) RepoMtimes(repoPath string) (map[string]int64, error) {
	if repoPath == "" {
		return nil, fmt.Errorf("empty repo path")
	}
	result := map[string]int64{"index": 0, "root": 0}

	indexPath := filepath.Join(repoPath, ".git", "index")
	if fi, err := os.Stat(indexPath); err == nil {
		result["index"] = fi.ModTime().UnixMilli()
	}

	if fi, err := os.Stat(repoPath); err == nil {
		result["root"] = fi.ModTime().UnixMilli()
	}

	return result, nil
}

// ListRepoChoices returns the repos available for spawning agents.
func (a *App) ListRepoChoices() []RepoChoice {
	st := a.scanSnapshot()
	if st.repoScanner == nil {
		return nil
	}
	repos, err := st.repoScanner.ScanRepos(nil)
	if err != nil {
		return nil
	}
	choices := make([]RepoChoice, 0, len(repos))
	for _, r := range repos {
		choices = append(choices, RepoChoice{
			Name:   r.Name,
			Path:   r.Path,
			Branch: r.Branch,
		})
	}
	return choices
}

// CreateRepo creates a new repository in the dev directory with optional GitHub
// remote and BMAD method installation. Emits repo:create:progress events.
func (a *App) CreateRepo(name string, isPublic bool, installBmad bool) {
	emit := func(step, message, errMsg string, done bool) {
		runtime.EventsEmit(a.ctx, "repo:create:progress", map[string]interface{}{
			"step":    step,
			"message": message,
			"error":   errMsg,
			"done":    done,
		})
	}

	st := a.scanSnapshot() // one snapshot for the whole operation (R17)
	go func() {
		// Validate name
		if name == "" || strings.ContainsAny(name, "/\\. ") {
			emit("validate", "", "Invalid repo name: must be non-empty with no spaces, dots, or slashes", true)
			return
		}
		if st.devDir == "" {
			emit("validate", "", "No development directory configured — set it in Settings first", true)
			return
		}
		targetDir := filepath.Join(st.devDir, name)
		if _, err := os.Stat(targetDir); err == nil {
			emit("validate", "", fmt.Sprintf("Directory %q already exists", name), true)
			return
		}

		// Check gh auth
		emit("gh-auth", "Checking GitHub CLI...", "", false)
		if out, err := exec.CommandContext(a.ctx, "gh", "auth", "status").CombinedOutput(); err != nil {
			emit("gh-auth", "", fmt.Sprintf("GitHub CLI not authenticated: %s", strings.TrimSpace(string(out))), true)
			return
		}

		// Create repo via gh
		visibility := "--private"
		if isPublic {
			visibility = "--public"
		}
		emit("gh-create", fmt.Sprintf("Creating %s repo %q...", visibility[2:], name), "", false)
		ghCmd := exec.CommandContext(a.ctx, "gh", "repo", "create", name, visibility, "--clone")
		ghCmd.Dir = st.devDir
		if out, err := ghCmd.CombinedOutput(); err != nil {
			emit("gh-create", "", fmt.Sprintf("gh repo create failed: %s", strings.TrimSpace(string(out))), true)
			return
		}

		// Install BMAD (best-effort)
		if installBmad {
			emit("bmad-install", "Installing BMAD method...", "", false)
			npxPath, err := exec.LookPath("npx")
			if err != nil {
				emit("bmad-install", "npx not found — skipping BMAD install", "", false)
			} else {
				bmadCmd := exec.CommandContext(a.ctx, npxPath, "bmad-method", "install", "--tools", "claude-code", "--directory", ".", "-y")
				bmadCmd.Dir = targetDir
				if out, err := bmadCmd.CombinedOutput(); err != nil {
					emit("bmad-install", fmt.Sprintf("BMAD install warning: %s", strings.TrimSpace(string(out))), "", false)
				} else {
					emit("bmad-install", "BMAD method installed", "", false)
				}
			}
		}

		// Git add + commit (best-effort)
		emit("git-commit", "Committing initial scaffold...", "", false)
		addCmd := exec.CommandContext(a.ctx, "git", "-C", targetDir, "add", "-A")
		if out, err := addCmd.CombinedOutput(); err != nil {
			emit("git-commit", fmt.Sprintf("git add warning: %s", strings.TrimSpace(string(out))), "", false)
		} else {
			commitCmd := exec.CommandContext(a.ctx, "git", "-C", targetDir, "commit", "-m", "feat: initial BMAD method scaffold")
			if out, err := commitCmd.CombinedOutput(); err != nil {
				emit("git-commit", fmt.Sprintf("git commit warning: %s", strings.TrimSpace(string(out))), "", false)
			} else {
				emit("git-commit", "Initial commit created", "", false)
			}
		}

		// Refresh repo list — rescan and emit the same "repos" event the feed listens for
		if st.repoScanner != nil {
			st.repoScanner.InvalidateCache(targetDir)
			if repos, err := st.repoScanner.ScanRepos(nil); err == nil {
				runtime.EventsEmit(a.ctx, "repos", repos)
			}
		}

		emit("done", fmt.Sprintf("Repository %q created successfully", name), "", true)
	}()
}

// GitListBranches returns all local branches for a repo, with the current branch marked.
func (a *App) GitListBranches(repoPath string) ([]BranchInfo, error) {
	if _, err := a.repoDir(repoPath); err != nil {
		return nil, err
	}
	if repoPath == "" {
		return nil, fmt.Errorf("repo path is required")
	}
	cmd := exec.CommandContext(a.ctx, "git", "-C", repoPath, "branch", "--format=%(refname:short)\t%(HEAD)")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git branch: %w", err)
	}

	var branches []BranchInfo
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		name := parts[0]
		current := len(parts) > 1 && strings.TrimSpace(parts[1]) == "*"
		branches = append(branches, BranchInfo{
			Name:    name,
			Current: current,
		})
	}
	return branches, nil
}

// GitSwitchBranch switches to an existing branch with optional auto-commit.
func (a *App) GitSwitchBranch(repoPath, branch string, autoCommit bool) error {
	if _, err := a.repoDir(repoPath); err != nil {
		return err
	}
	if repoPath == "" || branch == "" {
		return fmt.Errorf("repo path and branch name are required")
	}
	// R10: validate before any auto-commit or git process runs.
	if err := git.ValidateBranchName(branch); err != nil {
		return err
	}

	if autoCommit {
		if _, err := a.GitCommit(repoPath); err != nil {
			if !strings.Contains(err.Error(), "nothing to commit") {
				return fmt.Errorf("auto-commit failed: %w", err)
			}
		}
	}

	cmd := exec.CommandContext(a.ctx, "git", "-C", repoPath, "checkout", branch, "--")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git checkout: %w (%s)", err, string(out))
	}
	return nil
}

// GitCreateBranch creates a new branch with optional auto-commit of current changes.
// prefix is e.g. "feature", "hotfix"; name is the branch slug.
func (a *App) GitCreateBranch(repoPath, prefix, name string, autoCommit bool) error {
	if _, err := a.repoDir(repoPath); err != nil {
		return err
	}
	if repoPath == "" || name == "" {
		return fmt.Errorf("repo path and branch name are required")
	}

	branchName := name
	if prefix != "" {
		branchName = prefix + "/" + name
	}
	// R10: validate each part and the composed name before auto-commit.
	for _, part := range []string{prefix, name, branchName} {
		if part == "" {
			continue
		}
		if err := git.ValidateBranchName(part); err != nil {
			return err
		}
	}

	// Auto-commit current changes if requested
	if autoCommit {
		if _, err := a.GitCommit(repoPath); err != nil {
			// Ignore "nothing to commit" — that's fine
			if !strings.Contains(err.Error(), "nothing to commit") {
				return fmt.Errorf("auto-commit failed: %w", err)
			}
		}
	}

	// Create and checkout the new branch
	cmd := exec.CommandContext(a.ctx, "git", "-C", repoPath, "checkout", "-b", branchName)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git checkout -b: %w (%s)", err, string(out))
	}

	return nil
}

// RepoStatus returns git dirty state, open PR count, and ahead/behind counts for a repo.
func (a *App) RepoStatus(repoPath string) RepoStatusInfo {
	if _, err := a.repoDir(repoPath); err != nil {
		return RepoStatusInfo{}
	}
	result := RepoStatusInfo{}
	if repoPath == "" {
		return result
	}

	// Check dirty (uncommitted changes including untracked files)
	statusCmd := exec.CommandContext(a.ctx, "git", "-C", repoPath, "status", "--porcelain")
	if out, err := statusCmd.Output(); err == nil && len(out) > 0 {
		result.Dirty = true
	}

	// Check open PRs for the current branch
	branchCmd := exec.CommandContext(a.ctx, "git", "-C", repoPath, "rev-parse", "--abbrev-ref", "HEAD")
	branchOut, err := branchCmd.Output()
	if err == nil {
		branch := strings.TrimSpace(string(branchOut))
		ghCmd := exec.CommandContext(a.ctx, "gh", "pr", "list",
			"--state", "open",
			"--head", branch,
			"--json", "number",
			"--jq", "length",
		)
		ghCmd.Dir = repoPath
		if prOut, err := ghCmd.Output(); err == nil {
			count := strings.TrimSpace(string(prOut))
			if count != "" && count != "0" {
				n := 0
				fmt.Sscanf(count, "%d", &n)
				result.OpenPRs = n
			}
		}

		// Check ahead/behind remote tracking branch
		revCmd := exec.CommandContext(a.ctx, "git", "-C", repoPath,
			"rev-list", "--left-right", "--count", "HEAD...@{upstream}")
		if revOut, err := revCmd.Output(); err == nil {
			parts := strings.Fields(strings.TrimSpace(string(revOut)))
			if len(parts) == 2 {
				var ahead, behind int
				fmt.Sscanf(parts[0], "%d", &ahead)
				fmt.Sscanf(parts[1], "%d", &behind)
				result.Ahead = ahead
				result.Behind = behind
			}
		}

		// Check branch protection rules via gh API
		ghProtCmd := exec.CommandContext(a.ctx, "gh", "api",
			fmt.Sprintf("repos/{owner}/{repo}/branches/%s/protection", branch),
			"--jq", ".required_status_checks // empty",
		)
		ghProtCmd.Dir = repoPath
		if protOut, err := ghProtCmd.Output(); err == nil && len(strings.TrimSpace(string(protOut))) > 0 {
			result.Protected = true
		}
	}

	return result
}

// gitCommitCore stages all changes, generates an AI commit message, and commits.
// It returns the commit message. The onProgress callback, if non-nil, is called
// at each step so callers can stream status to the frontend.
//
// On commit failure (e.g. pre-commit hooks, lint errors), it spawns a Claude
// session to auto-fix the issues and retries. The stat-based fallback commit
// message is only used as an absolute last resort.
func (a *App) gitCommitCore(repoPath string, onProgress func(step, detail string)) (string, error) {
	progress := func(step, detail string) {
		if onProgress != nil {
			onProgress(step, detail)
		}
	}

	// Stage all changes
	progress("Staging changes...", "")
	addCmd := exec.CommandContext(a.ctx, "git", "-C", repoPath, "add", "-A")
	if out, err := addCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("git add: %w (%s)", err, string(out))
	}
	progress("Staged all changes", "")

	// Check there's something to commit
	statusCmd := exec.CommandContext(a.ctx, "git", "-C", repoPath, "diff", "--cached", "--stat")
	statusOut, err := statusCmd.Output()
	if err != nil || len(statusOut) == 0 {
		return "", fmt.Errorf("nothing to commit")
	}
	progress("Changes found", strings.TrimSpace(string(statusOut)))

	// Get the diff for the AI to summarize
	progress("Generating commit message...", "")
	diffCmd := exec.CommandContext(a.ctx, "git", "-C", repoPath, "diff", "--cached")
	diffOut, _ := diffCmd.Output()
	diffText := string(diffOut)
	if len(diffText) > 8000 {
		diffText = diffText[:8000] + "\n... (truncated)"
	}

	// Generate commit message using Claude CLI — retry with simpler prompt before falling back
	commitMsg := a.generateCommitMessage(repoPath, diffText, strings.TrimSpace(string(statusOut)), progress)

	// Attempt commit — on failure, auto-fix with Claude and retry
	const maxFixAttempts = 2
	for attempt := 0; attempt <= maxFixAttempts; attempt++ {
		progress("Committing...", "")
		commitCmd := exec.CommandContext(a.ctx, "git", "-C", repoPath, "commit", "-m", commitMsg)
		commitOut, commitErr := commitCmd.CombinedOutput()
		if commitErr == nil {
			return commitMsg, nil
		}

		errText := strings.TrimSpace(string(commitOut))

		// Don't retry "nothing to commit"
		if strings.Contains(errText, "nothing to commit") {
			return "", fmt.Errorf("nothing to commit")
		}

		if attempt >= maxFixAttempts {
			return "", fmt.Errorf("git commit: %w (%s)", commitErr, errText)
		}

		// Auto-fix: ask Claude to diagnose and fix the issues
		progress(fmt.Sprintf("Commit failed (attempt %d/%d), auto-fixing...", attempt+1, maxFixAttempts+1), errText)

		fixPrompt := fmt.Sprintf(
			"A git commit in this repository failed with this error:\n\n%s\n\n"+
				"Diagnose and fix the issue. Common causes: pre-commit hook failures, "+
				"lint errors, formatting issues, type errors. Fix the source files directly. "+
				"Do NOT run git commit — just fix the code so the next commit will succeed.",
			errText,
		)
		fixCmd := claudeCommand(a.ctx, "--dangerously-skip-permissions", "-p", fixPrompt)
		fixCmd.Dir = repoPath
		fixOut, fixErr := fixCmd.Output()
		fixSummary := strings.TrimSpace(string(fixOut))

		if fixErr != nil {
			progress("Auto-fix failed, retrying commit as-is...", "")
		} else {
			// Truncate for display
			if len(fixSummary) > 500 {
				fixSummary = fixSummary[:500] + "..."
			}
			progress("Auto-fix applied", fixSummary)
		}

		// Re-stage everything (including Claude's fixes)
		reAddCmd := exec.CommandContext(a.ctx, "git", "-C", repoPath, "add", "-A")
		if out, err := reAddCmd.CombinedOutput(); err != nil {
			progress("Re-staging failed", strings.TrimSpace(string(out)))
		}

		// Regenerate commit message to cover the fixes
		progress("Regenerating commit message...", "")
		reDiffCmd := exec.CommandContext(a.ctx, "git", "-C", repoPath, "diff", "--cached")
		reDiffOut, _ := reDiffCmd.Output()
		reDiffText := string(reDiffOut)
		if len(reDiffText) > 8000 {
			reDiffText = reDiffText[:8000] + "\n... (truncated)"
		}
		reStatCmd := exec.CommandContext(a.ctx, "git", "-C", repoPath, "diff", "--cached", "--stat")
		reStatOut, _ := reStatCmd.Output()

		commitMsg = a.generateCommitMessage(repoPath, reDiffText, strings.TrimSpace(string(reStatOut)), progress)
	}

	return commitMsg, nil
}

// generateCommitMessage tries Claude CLI to produce a commit message with retries.
// Falls back to a stat-based message only as an absolute last resort.
func (a *App) generateCommitMessage(repoPath, diffText, statSummary string, progress func(string, string)) string {
	// Primary attempt: full diff context
	prompt := fmt.Sprintf(
		"Write a concise git commit message (1-2 lines max, no quotes, no markdown) for this diff:\n\n%s",
		diffText,
	)
	claudeCmd := claudeCommand(a.ctx, "-p", prompt)
	claudeCmd.Dir = repoPath
	if msgOut, err := claudeCmd.Output(); err == nil {
		msg := strings.TrimSpace(string(msgOut))
		if msg != "" {
			progress("Commit message ready", msg)
			return msg
		}
	}

	// Retry: simpler prompt with just the stat summary
	progress("Retrying commit message generation...", "")
	retryPrompt := fmt.Sprintf(
		"Write a one-line git commit message (no quotes, no markdown) summarising these changes:\n\n%s",
		statSummary,
	)
	retryCmd := claudeCommand(a.ctx, "-p", retryPrompt)
	retryCmd.Dir = repoPath
	if retryOut, err := retryCmd.Output(); err == nil {
		msg := strings.TrimSpace(string(retryOut))
		if msg != "" {
			progress("Commit message ready (retry)", msg)
			return msg
		}
	}

	// Absolute last resort: stat-based fallback
	fallback := "update: " + statSummary
	if idx := strings.IndexByte(fallback, '\n'); idx > 0 {
		fallback = fallback[:idx]
	}
	progress("Using fallback commit message", fallback)
	return fallback
}

// GitCommit stages all changes, generates an AI commit message, and commits.
// Returns the commit message used.
func (a *App) GitCommit(repoPath string) (string, error) {
	if _, err := a.repoDir(repoPath); err != nil {
		return "", err
	}
	return a.gitCommitCore(repoPath, nil)
}

// GitCommitStreaming stages, generates an AI commit message, and commits,
// emitting progress events to the frontend at each step. The core function
// auto-fixes errors via Claude and retries. If all attempts fail, Claude
// explains the remaining issue to the user.
func (a *App) GitCommitStreaming(repoPath string) {
	if _, err := a.repoDir(repoPath); err != nil {
		runtime.EventsEmit(a.ctx, "git:commit:progress", map[string]interface{}{
			"repoPath": repoPath, "step": "Error", "error": err.Error(), "done": true,
		})
		return
	}
	emit := func(step, output, errMsg, explanation string, done bool) {
		runtime.EventsEmit(a.ctx, "git:commit:progress", map[string]interface{}{
			"repoPath":    repoPath,
			"step":        step,
			"output":      output,
			"error":       errMsg,
			"explanation": explanation,
			"done":        done,
		})
	}

	go func() {
		msg, err := a.gitCommitCore(repoPath, func(step, detail string) {
			emit(step, detail, "", "", false)
		})
		if err != nil {
			errMsg := err.Error()

			// "nothing to commit" is not worth explaining via Claude
			if strings.Contains(errMsg, "nothing to commit") {
				emit("Nothing to commit", "", "Nothing to commit — working tree clean", "", true)
				return
			}

			// Ask Claude to explain the failure
			explanation := ""
			prompt := fmt.Sprintf(
				"A git commit operation failed. Explain this error concisely (2-3 sentences) and suggest a fix.\n\nError:\n%s",
				errMsg,
			)
			claudeCmd := claudeCommand(a.ctx, "-p", prompt)
			claudeCmd.Dir = repoPath
			if expOut, expErr := claudeCmd.Output(); expErr == nil {
				explanation = strings.TrimSpace(string(expOut))
			}

			emit("Error", "", errMsg, explanation, true)
			return
		}

		emit("Committed", msg, "", "", true)
	}()
}

// GitCommitAndPush commits (via GitCommit) then pushes to origin.
// Creates the remote branch if it doesn't exist.
func (a *App) GitCommitAndPush(repoPath string) (string, error) {
	if _, err := a.repoDir(repoPath); err != nil {
		return "", err
	}
	msg, err := a.GitCommit(repoPath)
	if err != nil {
		return "", err
	}

	// Push with -u to set upstream, creating branch if needed
	pushCmd := exec.CommandContext(a.ctx, "git", "-C", repoPath, "push", "-u", "origin", "HEAD")
	if out, err := pushCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("git push: %w (%s)", err, string(out))
	}

	return msg, nil
}

// GitPush pushes the current branch to origin without committing first.
// Returns a structured result: "ok" on success, or "conflict:<message>" when
// the push is rejected due to diverged history (non-fast-forward).
func (a *App) GitPush(repoPath string) (string, error) {
	if _, err := a.repoDir(repoPath); err != nil {
		return "", err
	}
	if repoPath == "" {
		return "", fmt.Errorf("repo path is required")
	}

	pushCmd := exec.CommandContext(a.ctx, "git", "-C", repoPath, "push", "-u", "origin", "HEAD")
	out, err := pushCmd.CombinedOutput()
	if err != nil {
		outStr := string(out)
		// Detect non-fast-forward (diverged history) vs other errors
		if strings.Contains(outStr, "non-fast-forward") ||
			strings.Contains(outStr, "rejected") ||
			strings.Contains(outStr, "fetch first") {
			return "conflict:" + strings.TrimSpace(outStr), nil
		}
		return "", fmt.Errorf("git push: %w (%s)", err, outStr)
	}
	return "ok", nil
}

// GitForcePush force-pushes the current branch to origin with --force-with-lease
// for safety (fails if someone else pushed since your last fetch).
func (a *App) GitForcePush(repoPath string) (string, error) {
	if _, err := a.repoDir(repoPath); err != nil {
		return "", err
	}
	if repoPath == "" {
		return "", fmt.Errorf("repo path is required")
	}

	pushCmd := exec.CommandContext(a.ctx, "git", "-C", repoPath, "push", "--force-with-lease", "-u", "origin", "HEAD")
	out, err := pushCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git force push: %w (%s)", err, string(out))
	}
	return strings.TrimSpace(string(out)), nil
}

// GitPull pulls remote changes into the current branch.
func (a *App) GitPull(repoPath string) (string, error) {
	if _, err := a.repoDir(repoPath); err != nil {
		return "", err
	}
	if repoPath == "" {
		return "", fmt.Errorf("repo path is required")
	}
	cmd := exec.CommandContext(a.ctx, "git", "-C", repoPath, "pull")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git pull: %w (%s)", err, string(out))
	}
	return strings.TrimSpace(string(out)), nil
}

// GitMergeInto merges the current branch into targetBranch.
// If autoCommit is true, commits current changes before merging.
// On merge failure, aborts the merge and checks out the original branch.
func (a *App) GitMergeInto(repoPath, targetBranch string, autoCommit bool) (string, error) {
	if _, err := a.repoDir(repoPath); err != nil {
		return "", err
	}
	if repoPath == "" || targetBranch == "" {
		return "", fmt.Errorf("repo path and target branch are required")
	}
	// R10: validate before any auto-commit or git process runs.
	if err := git.ValidateBranchName(targetBranch); err != nil {
		return "", err
	}

	// Get current branch name
	branchCmd := exec.CommandContext(a.ctx, "git", "-C", repoPath, "rev-parse", "--abbrev-ref", "HEAD")
	branchOut, err := branchCmd.Output()
	if err != nil {
		return "", fmt.Errorf("get current branch: %w", err)
	}
	sourceBranch := strings.TrimSpace(string(branchOut))

	if sourceBranch == targetBranch {
		return "", fmt.Errorf("already on %s — nothing to merge", targetBranch)
	}

	// Auto-commit current changes if requested
	if autoCommit {
		if _, err := a.GitCommit(repoPath); err != nil {
			if !strings.Contains(err.Error(), "nothing to commit") {
				return "", fmt.Errorf("auto-commit failed: %w", err)
			}
		}
	}

	// Switch to target branch
	checkoutCmd := exec.CommandContext(a.ctx, "git", "-C", repoPath, "checkout", targetBranch, "--")
	if out, err := checkoutCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("checkout %s: %w (%s)", targetBranch, err, string(out))
	}

	// Merge source into target
	mergeCmd := exec.CommandContext(a.ctx, "git", "-C", repoPath, "merge", sourceBranch)
	mergeOut, mergeErr := mergeCmd.CombinedOutput()
	if mergeErr != nil {
		// Abort the failed merge and return to the original branch
		_ = exec.CommandContext(a.ctx, "git", "-C", repoPath, "merge", "--abort").Run()
		_ = exec.CommandContext(a.ctx, "git", "-C", repoPath, "checkout", sourceBranch).Run()
		return "", fmt.Errorf("merge %s into %s failed: %w (%s)", sourceBranch, targetBranch, mergeErr, string(mergeOut))
	}

	return fmt.Sprintf("Merged %s into %s", sourceBranch, targetBranch), nil
}

// GitCommitPushAndPR commits, pushes, and creates a PR with an extensive description.
// Returns the PR URL.
func (a *App) GitCommitPushAndPR(repoPath string) (string, error) {
	if _, err := a.repoDir(repoPath); err != nil {
		return "", err
	}
	_, err := a.GitCommitAndPush(repoPath)
	if err != nil {
		return "", err
	}

	// Get the current branch
	branchCmd := exec.CommandContext(a.ctx, "git", "-C", repoPath, "rev-parse", "--abbrev-ref", "HEAD")
	branchOut, err := branchCmd.Output()
	if err != nil {
		return "", fmt.Errorf("get branch: %w", err)
	}
	branch := strings.TrimSpace(string(branchOut))

	// Get the full diff against main/master for the PR body
	var baseBranch string
	for _, candidate := range []string{"main", "master"} {
		checkCmd := exec.CommandContext(a.ctx, "git", "-C", repoPath, "rev-parse", "--verify", candidate)
		if checkCmd.Run() == nil {
			baseBranch = candidate
			break
		}
	}
	if baseBranch == "" {
		baseBranch = "main"
	}

	diffCmd := exec.CommandContext(a.ctx, "git", "-C", repoPath, "diff", baseBranch+"...HEAD")
	diffOut, _ := diffCmd.Output()
	diffText := string(diffOut)
	if len(diffText) > 12000 {
		diffText = diffText[:12000] + "\n... (truncated)"
	}

	// Generate extensive PR description using Claude CLI
	prompt := fmt.Sprintf(`Write an extensive pull request description for this diff from branch "%s". Include:
- A clear title (first line, under 70 chars)
- ## Summary section with bullet points
- ## Changes section describing each file changed
- ## Test Plan section
Keep it factual based on the diff.

%s`, branch, diffText)
	claudeCmd := claudeCommand(a.ctx, "-p", prompt)
	claudeCmd.Dir = repoPath
	prOut, err := claudeCmd.Output()
	prBody := strings.TrimSpace(string(prOut))
	if err != nil || prBody == "" {
		prBody = fmt.Sprintf("Changes from branch %s", branch)
	}

	// Extract title (first line) and body (rest)
	prTitle := branch
	if idx := strings.IndexByte(prBody, '\n'); idx > 0 {
		prTitle = strings.TrimSpace(prBody[:idx])
		prBody = strings.TrimSpace(prBody[idx+1:])
		// Strip markdown heading prefix from title
		prTitle = strings.TrimLeft(prTitle, "# ")
	}
	if len(prTitle) > 70 {
		prTitle = prTitle[:67] + "..."
	}

	// Create PR using gh CLI
	ghCmd := exec.CommandContext(a.ctx, "gh", "pr", "create",
		"--title", prTitle,
		"--body", prBody,
		"--base", baseBranch,
	)
	ghCmd.Dir = repoPath
	ghOut, err := ghCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("gh pr create: %w (%s)", err, string(ghOut))
	}

	return strings.TrimSpace(string(ghOut)), nil
}

// GetScopedDiff returns the changed files for a directory.
func (a *App) GetScopedDiff(dir string) (*domain.ScopedDiff, error) {
	if _, err := a.repoDir(dir); err != nil {
		return nil, err
	}
	if dir == "" {
		return nil, fmt.Errorf("directory path is required")
	}
	return git.ScopedDiff(dir)
}

// GetWorktrees returns worktrees for a repo.
func (a *App) GetWorktrees(repoPath string) ([]domain.WorktreeInfo, error) {
	if _, err := a.repoDir(repoPath); err != nil {
		return nil, err
	}
	if repoPath == "" {
		return nil, fmt.Errorf("repo path is required")
	}
	return git.DetectWorktrees(repoPath)
}

// ListRepoFiles returns all tracked (and untracked non-ignored) files in a repo.
func (a *App) ListRepoFiles(repoPath string) ([]string, error) {
	if _, err := a.repoDir(repoPath); err != nil {
		return nil, err
	}
	if repoPath == "" {
		return nil, fmt.Errorf("empty repo path")
	}
	// git ls-files returns tracked files; --others --exclude-standard adds untracked non-ignored
	cmd := exec.CommandContext(a.ctx, "git", "-C", repoPath, "ls-files", "--cached", "--others", "--exclude-standard")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git ls-files: %w", err)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var files []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			files = append(files, line)
		}
	}
	sort.Strings(files)
	return files, nil
}

// fileRoots returns the directories the file bindings may touch: $HOME and
// the configured DevDir (spec R7, C1/C2).
func (a *App) fileRoots() []string {
	return pathguard.AllowedRoots(a.GetDevDir())
}

// repoDir validates a UI-supplied repository path: it must be an existing
// directory inside $HOME or DevDir (spec R11). It returns the resolved path.
func (a *App) repoDir(repoPath string) (string, error) {
	if repoPath == "" {
		return "", fmt.Errorf("repo path is required")
	}
	resolved, err := pathguard.ResolveExisting(a.fileRoots(), repoPath)
	if err != nil {
		return "", fmt.Errorf("repo: %w", err)
	}
	if fi, err := os.Stat(resolved); err != nil || !fi.IsDir() {
		return "", fmt.Errorf("repo %q is not a directory", repoPath)
	}
	return resolved, nil
}

// WriteFile writes content to a file on disk. The path must resolve inside
// $HOME or DevDir and must not be a persistence-sensitive file. A symlinked
// file is written through to its target, so the link is preserved.
func (a *App) WriteFile(path, content string) error {
	if path == "" {
		return fmt.Errorf("empty file path")
	}
	resolved, err := pathguard.ResolveForWrite(a.fileRoots(), path)
	if err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	if home, err := pathguard.HomeRoot(); err == nil {
		if err := pathguard.CheckWriteDenylist(home, resolved); err != nil {
			return fmt.Errorf("write file: %w", err)
		}
	}
	return fsutil.WriteFileAtomic(resolved, []byte(content), 0644)
}

// ReadFile returns the contents of a file as a string. The path must
// resolve inside $HOME or DevDir.
func (a *App) ReadFile(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("file path is required")
	}
	resolved, err := pathguard.ResolveExisting(a.fileRoots(), path)
	if err != nil {
		return "", fmt.Errorf("read file: %w", err)
	}
	data, err := os.ReadFile(resolved)
	if err != nil {
		return "", fmt.Errorf("read file %s: %w", path, err)
	}
	// Cap at 1MB to avoid sending huge files to frontend
	if len(data) > 1024*1024 {
		return string(data[:1024*1024]) + "\n... (truncated at 1MB)", nil
	}
	return string(data), nil
}

const maxImageFileSize = 10 * 1024 * 1024 // 10 MB

// mimeForExt returns a MIME type for the given file extension.
func mimeForExt(ext string) string {
	switch strings.ToLower(ext) {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".webp":
		return "image/webp"
	case ".bmp":
		return "image/bmp"
	case ".ico":
		return "image/x-icon"
	default:
		return "application/octet-stream"
	}
}

// ReadFileBase64 reads a file and returns it as a base64-encoded data URI.
func (a *App) ReadFileBase64(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("empty file path")
	}
	resolved, err := pathguard.ResolveExisting(a.fileRoots(), path)
	if err != nil {
		return "", fmt.Errorf("ReadFileBase64: %w", err)
	}

	info, err := os.Stat(resolved)
	if err != nil {
		return "", fmt.Errorf("ReadFileBase64 %s: %w", path, err)
	}
	if info.Size() > maxImageFileSize {
		return "", fmt.Errorf("file too large: %s (%d bytes)", path, info.Size())
	}

	data, err := os.ReadFile(resolved)
	if err != nil {
		return "", fmt.Errorf("ReadFileBase64 %s: %w", path, err)
	}

	mime := mimeForExt(filepath.Ext(path))
	return fmt.Sprintf("data:%s;base64,%s", mime, base64.StdEncoding.EncodeToString(data)), nil
}

// repoRelPath checks lexically that filePath stays inside its repo: it must
// be relative, must not climb with "..", and must not look like an option.
// Lexical, so a tracked file deleted from the working tree still resolves.
func repoRelPath(filePath string) (string, error) {
	rel := filepath.Clean(filePath)
	if filePath == "" || filepath.IsAbs(rel) || rel == ".." ||
		strings.HasPrefix(rel, ".."+string(filepath.Separator)) || strings.HasPrefix(rel, "-") {
		return "", fmt.Errorf("%q: %w", filePath, pathguard.ErrOutsideRoot)
	}
	return rel, nil
}

// ReadFileDiff returns the git diff for a specific file.
func (a *App) ReadFileDiff(repoPath, filePath string) (string, error) {
	if _, err := a.repoDir(repoPath); err != nil {
		return "", err
	}
	if repoPath == "" {
		return "", fmt.Errorf("repo path is required")
	}
	rel, err := repoRelPath(filePath)
	if err != nil {
		return "", fmt.Errorf("git diff: %w", err)
	}
	cmd := exec.CommandContext(a.ctx, "git", "-C", repoPath, "diff", "HEAD", "--", rel)
	out, err := cmd.Output()
	if err != nil {
		// Untracked file: diff against /dev/null. Resolve through symlinks so
		// the fallback can never read outside the repo (R9).
		abs, rerr := pathguard.ResolveExisting([]string{repoPath}, filepath.Join(repoPath, rel))
		if rerr != nil {
			return "", fmt.Errorf("git diff: %w", rerr)
		}
		cmd2 := exec.CommandContext(a.ctx, "git", "-C", repoPath, "diff", "--no-index", "--", "/dev/null", abs)
		out2, _ := cmd2.Output()
		if len(out2) > 0 {
			return string(out2), nil
		}
		return "", fmt.Errorf("git diff %s: %w", filePath, err)
	}
	return string(out), nil
}

// ReadFileAtHead returns the content of a file at the HEAD commit.
func (a *App) ReadFileAtHead(repoPath, filePath string) (string, error) {
	if _, err := a.repoDir(repoPath); err != nil {
		return "", err
	}
	if repoPath == "" {
		return "", fmt.Errorf("repo path is required")
	}
	rel, err := repoRelPath(filePath)
	if err != nil {
		return "", fmt.Errorf("git show: %w", err)
	}
	cmd := exec.CommandContext(a.ctx, "git", "-C", repoPath, "show", "HEAD:"+filepath.ToSlash(rel))
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git show HEAD:%s: %w", filePath, err)
	}
	if len(out) > 1024*1024 {
		return string(out[:1024*1024]) + "\n... (truncated at 1MB)", nil
	}
	return string(out), nil
}

// MarkRead marks a notification as read.
func (a *App) MarkRead(agentID string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for i, n := range a.notifications {
		if n.AgentID == agentID {
			a.notifications[i].Read = true
			break
		}
	}
}

// SpawnPRReview spawns a Claude agent to do an adversarial review of the latest PR.
// Returns the tmux pane target.
func (a *App) SpawnPRReview(repoPath string) (string, error) {
	if _, err := a.repoDir(repoPath); err != nil {
		return "", err
	}
	findPR := a.prNumber
	if findPR == nil {
		findPR = latestOpenPR
	}
	prNumber, err := findPR(a.ctx, repoPath)
	if err != nil {
		return "", err
	}
	if !isPRNumber(prNumber) {
		return "", fmt.Errorf("invalid PR number %q", prNumber)
	}

	// The app fetches the diff itself so the agent needs no shell (below).
	fetchDiff := a.prDiff
	if fetchDiff == nil {
		fetchDiff = ghPRDiff
	}
	diff, err := fetchDiff(a.ctx, repoPath, prNumber)
	if err != nil {
		return "", err
	}
	diffFile, err := os.CreateTemp("", "mashed-pr-"+prNumber+"-*.diff") // 0600
	if err != nil {
		return "", fmt.Errorf("write PR diff: %w", err)
	}
	_, werr := diffFile.WriteString(diff)
	if cerr := diffFile.Close(); werr == nil {
		werr = cerr
	}
	if werr != nil {
		os.Remove(diffFile.Name())
		return "", fmt.Errorf("write PR diff: %w", werr)
	}

	prompt := fmt.Sprintf(`You are an adversarial code reviewer. Review PR #%s in this repo thoroughly.
Look for: bugs, security vulnerabilities, race conditions, edge cases, performance issues,
missing error handling, breaking changes, and any code that could fail in production.
Be specific — cite file names and line numbers. Don't be nice, be thorough.
The PR diff is in this file; start by reading it with the Read tool: %s`, prNumber, diffFile.Name())

	// R16 / C3: the prompt is one argv element and the agent is read-only.
	// No shell at all: Claude Code auto-approves read-only-looking commands
	// such as `git log`, which accept --output=<file> and would let a
	// prompt-injected PR write arbitrary files (PR 2 security review).
	defaultModel := domain.DefaultAlias(a.ListModels())
	argv := []string{"claude", "--model", defaultModel,
		"--allowedTools", "Read,Grep,Glob", "--disallowedTools", "Bash", "-p", prompt}
	return a.spawnSessionArgv("review", repoPath, argv, domain.SessionAgent, defaultModel)
}

// ghPRDiff returns `gh pr diff <n>` for the repo.
func ghPRDiff(ctx context.Context, repoPath, n string) (string, error) {
	cmd := exec.CommandContext(ctx, "gh", "pr", "diff", n)
	cmd.Dir = repoPath
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("gh pr diff %s: %w", n, err)
	}
	return string(out), nil
}

// latestOpenPR returns the number of the most recent open PR in repoPath.
func latestOpenPR(ctx context.Context, repoPath string) (string, error) {
	ghCmd := exec.CommandContext(ctx, "gh", "pr", "list", "--state", "open", "--limit", "1", "--json", "number", "--jq", ".[0].number")
	ghCmd.Dir = repoPath
	prOut, err := ghCmd.Output()
	if err != nil {
		return "", fmt.Errorf("no open PRs found: %w", err)
	}
	n := strings.TrimSpace(string(prOut))
	if n == "" {
		return "", fmt.Errorf("no open PRs found")
	}
	return n, nil
}

// isPRNumber reports whether s is a non-empty string of ASCII digits.
func isPRNumber(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
