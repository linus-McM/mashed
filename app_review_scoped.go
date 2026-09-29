package main

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"mashed/internal/advice"
	"mashed/internal/pathguard"
)

// diffSeparator joins per-file diff outputs in the scoped diff payload.
const diffSeparator = "\n---\n"

// buildScopedDiff runs git diff for each specified file path, falling back to
// --no-index for untracked files. It concatenates non-empty per-file diffs
// separated by "\n---\n". If filePaths is empty, it returns an error.
// Paths that escape repoPath via ".." or absolute prefixes are silently
// skipped as a defence against traversal-based content leaks.
func buildScopedDiff(ctx context.Context, repoPath string, filePaths []string) (string, error) {
	if len(filePaths) == 0 {
		return "", fmt.Errorf("no files selected")
	}

	absRepo, err := filepath.Abs(repoPath)
	if err != nil {
		return "", fmt.Errorf("resolve repo path: %w", err)
	}

	tracked := listTrackedFiles(ctx, repoPath, filePaths)

	var parts []string
	for _, fp := range filePaths {
		absPath, ok := containedPath(absRepo, fp)
		if !ok {
			// Path escapes the repo — skip to prevent leaking external files.
			continue
		}

		if tracked[fp] {
			out, _ := exec.CommandContext(ctx, "git", "-C", repoPath, "diff", "HEAD", "--", fp).Output()
			if len(out) > 0 {
				parts = append(parts, string(out))
			}
			continue
		}

		// Untracked file: use --no-index fallback. git diff --no-index
		// exits with code 1 when differences exist, but cmd.Output() still
		// returns the captured stdout, so ignoring the error is safe.
		// Resolve symlinks too: a lexically contained path can still reach
		// outside the repo through a symlinked directory (R11).
		resolved, err := pathguard.ResolveExisting([]string{absRepo}, absPath)
		if err != nil {
			continue
		}
		noIdxOut, _ := exec.CommandContext(ctx, "git", "-C", repoPath, "diff", "--no-index", "--", "/dev/null", resolved).Output()
		if len(noIdxOut) > 0 {
			parts = append(parts, string(noIdxOut))
		}
	}

	return strings.Join(parts, diffSeparator), nil
}

// containedPath joins fp onto an already-absolute absRepo and returns the
// cleaned absolute path, provided the result stays inside absRepo. If fp
// escapes the repo (via ".." segments, an absolute sibling path, or a path
// that filepath.Rel cannot relate), it returns ("", false).
func containedPath(absRepo, fp string) (string, bool) {
	absFile := filepath.Clean(filepath.Join(absRepo, fp))
	rel, err := filepath.Rel(absRepo, absFile)
	if err != nil {
		return "", false
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", false
	}
	return absFile, true
}

// listTrackedFiles returns the set of filePaths that are tracked in the given
// repo, computed in a single git ls-files call.
func listTrackedFiles(ctx context.Context, repoPath string, filePaths []string) map[string]bool {
	args := append([]string{"-C", repoPath, "ls-files", "--"}, filePaths...)
	out, _ := exec.CommandContext(ctx, "git", args...).Output()

	tracked := make(map[string]bool, len(filePaths))
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line != "" {
			tracked[line] = true
		}
	}
	return tracked
}

// assembleScopedPayload combines additionalContext and diff into the stdin
// payload for Claude CLI. If additionalContext is non-empty, it is prepended
// under a "## Prior Context" header, with the diff under "## Code Changes".
// If additionalContext is empty, the raw diff is returned unchanged.
func assembleScopedPayload(diff, additionalContext string) string {
	if additionalContext == "" {
		return diff
	}
	return fmt.Sprintf("## Prior Context\n%s\n\n## Code Changes\n%s", additionalContext, diff)
}

// scopedAdviceEvent constructs the event payload map for StreamScopedAdvice,
// matching the same shape as StreamAdvice events (repoPath, text, done, error).
func scopedAdviceEvent(repoPath, text string, done bool, errMsg string) map[string]interface{} {
	return map[string]interface{}{
		"repoPath": repoPath,
		"text":     text,
		"done":     done,
		"error":    errMsg,
	}
}

// StreamScopedAdvice streams advice from Claude CLI restricted to a subset of
// changed files. It mirrors StreamAdvice but takes explicit filePaths and an
// optional additionalContext that is prepended to the stdin payload. Events
// are emitted on the "review:advice:progress" channel.
func (a *App) StreamScopedAdvice(repoPath, modeName, model string, filePaths []string, additionalContext string) {
	if _, err := a.repoDir(repoPath); err != nil {
		a.emitEvent("review:advice:progress", scopedAdviceEvent(repoPath, "", true, err.Error()))
		return
	}
	if model == "" {
		model = "sonnet"
	}

	go func() {
		emitProgress := func(text string, done bool, errMsg string) {
			runtime.EventsEmit(a.ctx, "review:advice:progress", scopedAdviceEvent(repoPath, text, done, errMsg))
		}

		if _, err := exec.LookPath("claude"); err != nil {
			emitProgress("", true, "Claude CLI not found")
			return
		}

		body, err := advice.LoadAdviceBody(repoPath, modeName)
		if err != nil {
			emitProgress("", true, fmt.Sprintf("failed to load advice mode %q: %s", modeName, err))
			return
		}

		diff, err := buildScopedDiff(a.ctx, repoPath, filePaths)
		if err != nil {
			emitProgress("", true, fmt.Sprintf("failed to build scoped diff: %s", err))
			return
		}

		if diff == "" {
			emitProgress("No changes to review.", true, "")
			return
		}

		payload := assembleScopedPayload(diff, additionalContext)

		systemPrompt := fmt.Sprintf(`You are an expert code reviewer. Apply the following methodology to review the code changes provided.

%s

Provide your analysis in markdown format.`, body)

		cmd := claudeCommand(a.ctx,
			"--print",
			"--model", model,
			"--system-prompt", systemPrompt,
			"--no-session-persistence",
		)
		cmd.Dir = repoPath
		cmd.Stdin = strings.NewReader(payload)

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			emitProgress("", true, fmt.Sprintf("failed to create pipe: %s", err))
			return
		}

		if err := cmd.Start(); err != nil {
			emitProgress("", true, fmt.Sprintf("failed to start Claude: %s", err))
			return
		}

		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			emitProgress(scanner.Text(), false, "")
		}

		if err := cmd.Wait(); err != nil {
			emitProgress("", true, fmt.Sprintf("Claude exited with error: %s", err))
			return
		}

		emitProgress("", true, "")
	}()
}
