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

// activeReviews maps repoPath → context.CancelFunc for in-progress reviews.
// A new review for the same repo cancels the previous one.
var activeReviews sync.Map

// maxReviewFiles caps the number of files processed in a single review.
const maxReviewFiles = 50

// maxDiffLines caps the number of diff lines sent to Claude per file.
const maxDiffLines = 500

// summaryTimeout is the per-file timeout for Claude summarisation calls.
const summaryTimeout = 60 * time.Second

// reviewSkipExts contains file extensions to exclude from code review summaries.
// These are non-code files (docs, configs, dotfiles) that add noise.
var reviewSkipExts = map[string]bool{
	".md":       true,
	".txt":      true,
	".json":     true,
	".yaml":     true,
	".yml":      true,
	".toml":     true,
	".xml":      true,
	".csv":      true,
	".lock":     true,
	".sum":      true,
	".mod":      true,
	".env":      true,
	".gitignore": true,
}

// reviewSkipNames contains exact filenames (case-insensitive) to exclude.
var reviewSkipNames = map[string]bool{
	"makefile":    true,
	"justfile":    true,
	"dockerfile":  true,
	"rakefile":    true,
	"gemfile":     true,
	"procfile":    true,
	"license":     true,
	"licence":     true,
	"changelog":   true,
}

// isReviewableFile returns true if the file should be included in code review.
// Excludes dotfiles (paths starting with '.'), markdown, config files,
// and build/task runner files (Makefile, justfile, Dockerfile).
func isReviewableFile(path string) bool {
	base := filepath.Base(path)
	// Skip dotfiles (e.g. .gitignore, .eslintrc, .prettierrc)
	if strings.HasPrefix(base, ".") {
		return false
	}
	// Skip files in dot-directories (e.g. .wolf/, .github/, .claude/)
	for _, part := range strings.Split(path, "/") {
		if strings.HasPrefix(part, ".") && part != "." && part != ".." {
			return false
		}
	}
	// Skip known non-code filenames (Makefile, justfile, etc.)
	if reviewSkipNames[strings.ToLower(base)] {
		return false
	}
	ext := strings.ToLower(filepath.Ext(path))
	return !reviewSkipExts[ext]
}

// ListAdviceModes returns the available advice/review methodology modes.
func (a *App) ListAdviceModes(repoPath string) ([]advice.AdviceMode, error) {
	return advice.LoadAdviceModes(repoPath)
}

// StreamCodeReviewSummary generates AI summaries for each changed file and
// streams progress events to the frontend. Runs asynchronously.
// The model parameter selects which Claude model to use (alias or full ID).
func (a *App) StreamCodeReviewSummary(repoPath, model string) {
	if _, err := a.repoDir(repoPath); err != nil {
		a.emitEvent("review:summary:done", map[string]interface{}{
			"repoPath": repoPath, "summary": nil, "error": err.Error(),
		})
		return
	}
	if model == "" {
		model = "sonnet" // fast model for per-file summaries
	}

	// Cancel any in-progress review for this repo.
	if prev, loaded := activeReviews.Load(repoPath); loaded {
		if cancel, ok := prev.(context.CancelFunc); ok {
			cancel()
		}
	}

	// Create a cancellable context for this review.
	reviewCtx, reviewCancel := context.WithCancel(a.ctx)
	activeReviews.Store(repoPath, reviewCancel)

	go func() {
		defer activeReviews.Delete(repoPath)
		defer reviewCancel()

		emitDone := func(summary interface{}, errMsg string) {
			runtime.EventsEmit(a.ctx, "review:summary:done", map[string]interface{}{
				"repoPath": repoPath,
				"summary":  summary,
				"error":    errMsg,
			})
		}

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

		// Filter to code-only files (skip markdown, dotfiles, configs).
		var files []domain.DiffFileStat
		for _, f := range scopedDiff.Files {
			if isReviewableFile(f.Path) {
				files = append(files, f)
			}
		}
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
			case <-reviewCtx.Done():
				return // silently exit — a new review replaced us
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
					summary, claudeErr := runClaudePrompt(reviewCtx, repoPath, model,
						fileSummarySystemPrompt, diffText, summaryTimeout)
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

			// Only emit if we haven't been cancelled.
			select {
			case <-reviewCtx.Done():
				return
			default:
			}

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
// The model parameter selects which Claude model to use (alias or full ID).
func (a *App) StreamAdvice(repoPath, modeName, model string) {
	if _, err := a.repoDir(repoPath); err != nil {
		a.emitEvent("review:advice:progress", map[string]interface{}{
			"repoPath": repoPath, "text": "", "done": true, "error": err.Error(),
		})
		return
	}
	if model == "" {
		model = "sonnet"
	}

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

		// Load advice body (used as system prompt).
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

		// Build system prompt from advice body.
		systemPrompt := fmt.Sprintf(`You are an expert code reviewer. Apply the following methodology to review the code changes provided.

%s

Provide your analysis in markdown format.`, body)

		// Spawn Claude: system prompt via --system-prompt, diff via stdin.
		// --no-session-persistence avoids writing throwaway sessions to disk.
		// Note: --bare is NOT used because it disables OAuth/keychain auth.
		cmd := claudeCommand(a.ctx,
			"--print",
			"--model", model,
			"--system-prompt", systemPrompt,
			"--no-session-persistence",
		)
		cmd.Dir = repoPath
		cmd.Stdin = strings.NewReader(string(diffOut))

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

// SpawnRefactorPlan spawns a Claude agent that produces a refactor plan
// based on the advice text. The optional filePaths are referenced in the
// generated plan filename so reviewers can tell at a glance which source
// files a plan covers. Returns the plan file path.
func (a *App) SpawnRefactorPlan(repoPath, adviceText string, filePaths []string) (string, error) {
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

	planPath := filepath.Join(plansDir, refactorPlanFilename(filePaths, time.Now().Unix()))

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

	defaultModel := domain.DefaultAlias(a.ListModels())
	cmd := fmt.Sprintf("claude --dangerously-skip-permissions --model %s -p %q", defaultModel, prompt)
	_, err := a.spawnSession("refactor", repoPath, cmd, domain.SessionAgent, defaultModel, 0, 0)
	if err != nil {
		return "", fmt.Errorf("spawn refactor plan agent: %w", err)
	}

	return planPath, nil
}

// refactorPlanFilename builds the plan filename. It references the source
// files the plan covers so reviewers can match plans to code at a glance.
// The '.' separating a file's extension is replaced with '_' so the filename
// has exactly one real extension (`.md`). Multiple files are joined with '+';
// beyond maxNamedFiles the extras collapse into a "+Nmore" tail to keep the
// filename bounded.
func refactorPlanFilename(filePaths []string, unixTime int64) string {
	const maxNamedFiles = 3

	if len(filePaths) == 0 {
		return fmt.Sprintf("refactor-%d.md", unixTime)
	}

	named := filePaths
	extra := 0
	if len(filePaths) > maxNamedFiles {
		named = filePaths[:maxNamedFiles]
		extra = len(filePaths) - maxNamedFiles
	}

	slugs := make([]string, 0, len(named))
	for _, p := range named {
		if slug := slugifyPlanPath(p); slug != "" {
			slugs = append(slugs, slug)
		}
	}

	if len(slugs) == 0 {
		return fmt.Sprintf("refactor-%d.md", unixTime)
	}

	joined := strings.Join(slugs, "+")
	if extra > 0 {
		joined = fmt.Sprintf("%s+%dmore", joined, extra)
	}
	return fmt.Sprintf("refactor-%s-%d.md", joined, unixTime)
}

// slugifyPlanPath reduces a repo-relative file path to a filename-safe slug.
// Directory separators become '-', the extension dot becomes '_', and any
// remaining characters outside [A-Za-z0-9_-] are replaced with '_'. Returns
// an empty string for inputs that would otherwise yield nothing usable.
func slugifyPlanPath(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return ""
	}

	base := filepath.Base(p)
	if base == "." || base == "/" {
		return ""
	}

	// Replace the final '.' (extension separator) with '_'. Only the last
	// dot is converted so multi-dotted names like "foo.bar.go" become
	// "foo.bar_go" rather than "foo_bar_go".
	if dot := strings.LastIndex(base, "."); dot > 0 && dot < len(base)-1 {
		base = base[:dot] + "_" + base[dot+1:]
	}

	var b strings.Builder
	b.Grow(len(base))
	for _, r := range base {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '_', r == '-', r == '.':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	return b.String()
}

// truncateDiffLines caps diff text at the given number of lines.
func truncateDiffLines(diff string, maxLines int) string {
	lines := strings.SplitN(diff, "\n", maxLines+1)
	if len(lines) <= maxLines {
		return diff
	}
	return strings.Join(lines[:maxLines], "\n") + "\n... (truncated)"
}

// fileSummarySystemPrompt is the system prompt for per-file diff summarisation.
const fileSummarySystemPrompt = `Summarise the code changes provided.
- If the change is simple (rename, one-liner, import change), respond with ONE sentence.
- If the change is complex (new function, refactor, logic change), respond with a short paragraph (3-5 sentences).
- Be specific about what changed and why it matters.`

// runClaudePrompt executes `claude --print` with a system prompt and user
// content piped via stdin. Uses --no-session-persistence to avoid disk clutter.
// Note: --bare is NOT used because it disables OAuth/keychain auth.
func runClaudePrompt(parentCtx context.Context, repoPath, model, systemPrompt, userContent string, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(parentCtx, timeout)
	defer cancel()

	cmd := claudeCommand(ctx,
		"--print",
		"--model", model,
		"--system-prompt", systemPrompt,
		"--no-session-persistence",
	)
	cmd.Dir = repoPath
	cmd.Stdin = strings.NewReader(userContent)

	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("claude --print: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}
