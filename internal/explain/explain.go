package explain

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os/exec"
	"strings"
	"sync"
)

// Explainer calls the Claude CLI to explain diff hunks, caching results
// by a SHA-256 key of (repoPath + filePath + hunkText).
type Explainer struct {
	cache map[string]string
	mu    sync.RWMutex
}

// New creates an Explainer with an initialized cache.
func New() *Explainer {
	return &Explainer{
		cache: make(map[string]string),
	}
}

// Explain returns a 1-2 sentence explanation of why a diff hunk was changed.
// Results are cached by a SHA-256 hash of repoPath+filePath+hunkText.
func (e *Explainer) Explain(ctx context.Context, repoPath, filePath, hunkText string) (string, error) {
	key := cacheKey(repoPath, filePath, hunkText)

	// Check cache (read lock)
	e.mu.RLock()
	if cached, ok := e.cache[key]; ok {
		e.mu.RUnlock()
		return cached, nil
	}
	e.mu.RUnlock()

	// Truncate oversized hunks to first 100 lines
	hunk := truncateHunk(hunkText, 100)

	prompt := fmt.Sprintf(
		"You are a code reviewer. Given this diff hunk from file %q, explain WHY this change was made in 1-2 sentences. Be concise and focus on intent, not mechanics.\n\n```diff\n%s\n```",
		filePath, hunk,
	)

	cmd := exec.CommandContext(ctx, "claude", "-p", prompt)
	cmd.Dir = repoPath
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("claude CLI failed: %w", err)
	}

	explanation := strings.TrimSpace(string(out))
	if explanation == "" {
		return "", fmt.Errorf("claude returned empty response")
	}

	// Cache result (write lock)
	e.mu.Lock()
	e.cache[key] = explanation
	e.mu.Unlock()

	return explanation, nil
}

// cacheKey returns the hex-encoded SHA-256 of repoPath+filePath+hunkText.
func cacheKey(repoPath, filePath, hunkText string) string {
	h := sha256.Sum256([]byte(repoPath + filePath + hunkText))
	return fmt.Sprintf("%x", h)
}

// truncateHunk returns the first n lines of text if it exceeds n lines.
func truncateHunk(text string, n int) string {
	lines := strings.SplitN(text, "\n", n+1)
	if len(lines) <= n {
		return text
	}
	return strings.Join(lines[:n], "\n")
}
