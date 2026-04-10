package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"mashed/internal/advice"
	"mashed/internal/domain"
	"mashed/internal/git"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// FileSummary describes a single file's changes and an AI-generated summary.
type FileSummary struct {
	Path     string `json:"path"`
	Added    int    `json:"added"`
	Removed  int    `json:"removed"`
	Summary  string `json:"summary"`
	IsBinary bool   `json:"isBinary"`
}

// ReviewSummary aggregates all file summaries and total line counts.
type ReviewSummary struct {
	Files        []FileSummary `json:"files"`
	TotalAdded   int           `json:"totalAdded"`
	TotalRemoved int           `json:"totalRemoved"`
}

// activeReviews guards against concurrent reviews of the same repo.
var activeReviews sync.Map

// maxReviewFiles caps the number of files processed in a single review.
const maxReviewFiles = 50

// maxDiffLines caps the number of diff lines sent to Claude per file.
const maxDiffLines = 500

// summaryTimeout is the per-file timeout for Claude summarisation calls.
const summaryTimeout = 60 * time.Second

// ListAdviceModes returns the available advice/review methodology modes.
func (a *App) ListAdviceModes(repoPath string) ([]advice.AdviceMode, error) {
	return advice.LoadAdviceModes(repoPath)
}

// StreamCodeReviewSummary generates AI summaries for each changed file and
// streams progress events to the frontend. Runs asynchronously.
func (a *App) StreamCodeReviewSummary(repoPath string) {
	go func() {
		emitDone := func(summary interface{}, errMsg string) {
			runtime.EventsEmit(a.ctx, "review:summary:done", map[string]interface{}{
				"repoPath": repoPath,
				"summary":  summary,
				"error":    errMsg,
			})
		}

		// Concurrency guard: one review per repo at a time.
		if _, loaded := activeReviews.LoadOrStore(repoPath, true); loaded {
			emitDone(nil, "review already in progress")
			return
		}
		defer activeReviews.Delete(repoPath)

		// Verify Claude CLI is available.
		if _, err := exec.LookPath("claude"); err != nil {
			emitDone(nil, "Claude CLI not found")
			return
		}

		// Get scoped diff.
		scopedDiff, err := git.ScopedDiff(repoPath)
		if err != nil {
			emitDone(nil, fmt.Sprintf("failed to get diff: %s", err))
			return
		}

		files := scopedDiff.Files
		if len(files) == 0 {
			emitDone(ReviewSummary{Files: []FileSummary{}}, "")
			return
		}

		if len(files) > maxReviewFiles {
			files = files[:maxReviewFiles]
		}

		total := len(files)
		summaries := make([]FileSummary, 0, total)
		totalAdded := 0
		totalRemoved := 0

		for i, file := range files {
			select {
			case <-a.ctx.Done():
				emitDone(nil, "review cancelled")
				return
			default:
			}

			fs := FileSummary{
				Path:     file.Path,
				Added:    file.Added,
				Removed:  file.Removed,
				IsBinary: file.IsBinary,
			}

			if file.IsBinary {
				fs.Summary = "Binary file changed"
			} else {
				diffText, diffErr := a.ReadFileDiff(repoPath, file.Path)
				if diffErr != nil {
					fs.Summary = fmt.Sprintf("Could not read diff: %s", diffErr)
				} else {
					diffText = truncateDiffLines(diffText, maxDiffLines)
					prompt := buildFileSummaryPrompt(file.Path, diffText)
					summary, claudeErr := runClaudePrompt(a.ctx, repoPath, prompt, summaryTimeout)
					if claudeErr != nil {
						fs.Summary = fmt.Sprintf("Summary unavailable: %s", claudeErr)
					} else {
						fs.Summary = summary
					}
				}
			}

			totalAdded += fs.Added
			totalRemoved += fs.Removed
			summaries = append(summaries, fs)

			runtime.EventsEmit(a.ctx, "review:summary:progress", map[string]interface{}{
				"repoPath": repoPath,
				"file":     fs,
				"index":    i,
				"total":    total,
				"error":    "",
			})
		}

		emitDone(ReviewSummary{
			Files:        summaries,
			TotalAdded:   totalAdded,
			TotalRemoved: totalRemoved,
		}, "")
	}()
}

// StreamAdvice runs a methodology-based code review and streams the output
// line-by-line to the frontend.
func (a *App) StreamAdvice(repoPath, modeName string) {
	go func() {
		emitProgress := func(text string, done bool, errMsg string) {
			runtime.EventsEmit(a.ctx, "review:advice:progress", map[string]interface{}{
				"repoPath": repoPath,
				"text":     text,
				"done":     done,
				"error":    errMsg,
			})
		}

		// Verify Claude CLI is available.
		if _, err := exec.LookPath("claude"); err != nil {
			emitProgress("", true, "Claude CLI not found")
			return
		}

		// Load advice body.
		body, err := advice.LoadAdviceBody(repoPath, modeName)
		if err != nil {
			emitProgress("", true, fmt.Sprintf("failed to load advice mode %q: %s", modeName, err))
			return
		}

		// Get full diff.
		diffCmd := exec.CommandContext(a.ctx, "git", "-C", repoPath, "diff", "HEAD")
		diffOut, err := diffCmd.Output()
		if err != nil {
			emitProgress("", true, fmt.Sprintf("failed to get diff: %s", err))
			return
		}

		if len(diffOut) == 0 {
			emitProgress("No changes to review.", true, "")
			return
		}

		prompt := buildAdvicePrompt(body, string(diffOut))

		// Spawn Claude with streaming output.
		cmd := exec.CommandContext(a.ctx, "claude", "-p", prompt)
		cmd.Dir = repoPath

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

// truncateDiffLines caps diff text at the given number of lines.
func truncateDiffLines(diff string, maxLines int) string {
	lines := strings.SplitN(diff, "\n", maxLines+1)
	if len(lines) <= maxLines {
		return diff
	}
	return strings.Join(lines[:maxLines], "\n") + "\n... (truncated)"
}

// buildFileSummaryPrompt constructs the Claude prompt for summarising a single file diff.
func buildFileSummaryPrompt(filePath, diff string) string {
	return fmt.Sprintf(`Summarise the following code changes in %s.
- If the change is simple (rename, one-liner, import change), respond with ONE sentence.
- If the change is complex (new function, refactor, logic change), respond with a short paragraph (3-5 sentences).
- Be specific about what changed and why it matters.

Diff:
%s`, filePath, diff)
}

// buildAdvicePrompt constructs the Claude prompt for methodology-based review.
func buildAdvicePrompt(methodologyBody, diff string) string {
	return fmt.Sprintf(`You are an expert code reviewer. Apply the following methodology to review these code changes.

## Methodology
%s

## Code Changes (Diff)
%s

Provide your analysis in markdown format.`, methodologyBody, diff)
}

// SpawnRefactorPlan spawns a Claude agent that produces a refactor plan
// based on the advice text. Returns the plan file path.
func (a *App) SpawnRefactorPlan(repoPath, adviceText string) (string, error) {
	if repoPath == "" {
		return "", fmt.Errorf("repo path is required")
	}
	if adviceText == "" {
		return "", fmt.Errorf("advice text is required")
	}

	// Ensure plans directory exists.
	plansDir := filepath.Join(repoPath, ".claude", "plans")
	if err := os.MkdirAll(plansDir, 0o755); err != nil {
		return "", fmt.Errorf("create plans directory: %w", err)
	}

	planPath := filepath.Join(plansDir, fmt.Sprintf("refactor-%d.md", time.Now().Unix()))

	// Truncate advice to avoid CLI arg length limits.
	if len(adviceText) > 10000 {
		adviceText = adviceText[:10000] + "\n... (truncated)"
	}

	prompt := fmt.Sprintf(`You are a senior engineer creating a refactor plan. Based on the following code review advice, create a detailed refactor plan and write it to: %s

## Advice
%s

## Instructions
- Write a markdown plan to the file path above using Write tool
- Structure: Summary, Priority Items, File-by-File Changes, Testing Strategy
- Be specific: include file paths, function names, line references
- Order by priority (critical first, cosmetic last)`, planPath, adviceText)

	cmd := fmt.Sprintf("claude --dangerously-skip-permissions --model claude-opus-4-6 -p %q", prompt)
	_, err := a.spawnSession("refactor", repoPath, cmd, domain.SessionAgent, "claude-opus-4-6")
	if err != nil {
		return "", fmt.Errorf("spawn refactor plan agent: %w", err)
	}

	return planPath, nil
}

// runClaudePrompt executes `claude -p` with the given prompt and timeout.
func runClaudePrompt(parentCtx context.Context, repoPath, prompt string, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(parentCtx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "claude", "-p", prompt)
	cmd.Dir = repoPath

	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("claude -p: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}
