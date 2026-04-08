package main

import (
	"os/exec"
)

// ExplainDiffHunk returns an AI-generated explanation of why a diff hunk was changed.
func (a *App) ExplainDiffHunk(repoPath, filePath, hunkText string) (string, error) {
	return a.explainer.Explain(a.ctx, repoPath, filePath, hunkText)
}

// IsExplainAvailable returns true if the claude CLI is on PATH.
func (a *App) IsExplainAvailable() bool {
	_, err := exec.LookPath("claude")
	return err == nil
}
