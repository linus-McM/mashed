package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Test helpers: git repo setup
// ---------------------------------------------------------------------------

// initTestGitRepo creates a temporary git repo with an initial commit.
func initTestGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	gitRun(t, dir, "init")
	gitRun(t, dir, "config", "user.email", "test@test.com")
	gitRun(t, dir, "config", "user.name", "Test")
	readme := filepath.Join(dir, "README.md")
	require.NoError(t, os.WriteFile(readme, []byte("# Test\n"), 0o644))
	gitRun(t, dir, "add", ".")
	gitRun(t, dir, "commit", "-m", "initial")
	return dir
}

// gitRun runs a git command in the given directory.
func gitRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %v failed: %s", args, string(out))
}

// commitFile creates/overwrites a file, stages, and commits it.
func commitFile(t *testing.T, repoPath, name, content string) {
	t.Helper()
	p := filepath.Join(repoPath, name)
	require.NoError(t, os.WriteFile(p, []byte(content), 0o644))
	gitRun(t, repoPath, "add", name)
	gitRun(t, repoPath, "commit", "-m", "add "+name)
}

// modifyTrackedFile writes new content to a committed file without staging.
func modifyTrackedFile(t *testing.T, repoPath, name, newContent string) {
	t.Helper()
	p := filepath.Join(repoPath, name)
	require.NoError(t, os.WriteFile(p, []byte(newContent), 0o644))
}

// createUntrackedFile writes a new file that is not tracked by git.
func createUntrackedFile(t *testing.T, repoPath, name, content string) {
	t.Helper()
	p := filepath.Join(repoPath, name)
	require.NoError(t, os.WriteFile(p, []byte(content), 0o644))
}

// ---------------------------------------------------------------------------
// buildScopedDiff — AC-1: Scoped diff generation
// ---------------------------------------------------------------------------

func TestBuildScopedDiff_AC1_OnlySelectedFiles(t *testing.T) {
	t.Parallel()
	// BDD Scenario 1: Only selected files appear in diff
	// Given a git repo with tracked changes in "main.go", "util.go", and "config.go"
	// When buildScopedDiff is called with filePaths ["main.go", "util.go"]
	// Then diff contains content from "main.go" and "util.go"
	// And diff does NOT contain content from "config.go"

	repoPath := initTestGitRepo(t)

	commitFile(t, repoPath, "main.go", "package main\n")
	commitFile(t, repoPath, "util.go", "package main\n")
	commitFile(t, repoPath, "config.go", "package main\n")

	modifyTrackedFile(t, repoPath, "main.go", "package main\n// main changed\n")
	modifyTrackedFile(t, repoPath, "util.go", "package main\n// util changed\n")
	modifyTrackedFile(t, repoPath, "config.go", "package main\n// config changed\n")

	diff, err := buildScopedDiff(context.Background(), repoPath, []string{"main.go", "util.go"})
	require.NoError(t, err)

	assert.Contains(t, diff, "main.go", "selected file main.go must be in diff")
	assert.Contains(t, diff, "util.go", "selected file util.go must be in diff")
	assert.NotContains(t, diff, "config.go", "unselected file config.go must NOT be in diff")
}

// ---------------------------------------------------------------------------
// buildScopedDiff — AC-2: Untracked file fallback
// ---------------------------------------------------------------------------

func TestBuildScopedDiff_AC2_UntrackedFallback(t *testing.T) {
	t.Parallel()
	// BDD Scenario 2: Untracked file uses no-index fallback
	// Given a git repo with an untracked file "new_feature.go"
	// When buildScopedDiff is called with filePaths ["new_feature.go"]
	// Then the method falls back to --no-index and the diff is non-empty

	repoPath := initTestGitRepo(t)
	createUntrackedFile(t, repoPath, "new_feature.go", "package main\n\nfunc NewFeature() {}\n")

	diff, err := buildScopedDiff(context.Background(), repoPath, []string{"new_feature.go"})
	require.NoError(t, err)

	assert.NotEmpty(t, diff, "untracked file should produce non-empty diff via --no-index fallback")
	assert.Contains(t, diff, "new_feature.go")
	assert.Contains(t, diff, "NewFeature")
}

// ---------------------------------------------------------------------------
// buildScopedDiff — Mixed tracked and untracked
// ---------------------------------------------------------------------------

func TestBuildScopedDiff_MixedTrackedUntracked(t *testing.T) {
	t.Parallel()
	// BDD Scenario 3: Mixed tracked and untracked files
	// Given a git repo with tracked change "main.go" and untracked file "new.go"
	// When buildScopedDiff is called with filePaths ["main.go", "new.go"]
	// Then both files appear in the concatenated diff output

	repoPath := initTestGitRepo(t)

	commitFile(t, repoPath, "main.go", "package main\n")
	modifyTrackedFile(t, repoPath, "main.go", "package main\n// tracked change\n")
	createUntrackedFile(t, repoPath, "new.go", "package main\n\nfunc New() {}\n")

	diff, err := buildScopedDiff(context.Background(), repoPath, []string{"main.go", "new.go"})
	require.NoError(t, err)

	assert.Contains(t, diff, "main.go", "tracked file must appear in diff")
	assert.Contains(t, diff, "new.go", "untracked file must appear in diff")
	assert.Contains(t, diff, "tracked change", "tracked file modification content must be present")
	assert.Contains(t, diff, "New", "untracked file content must be present")
}

// ---------------------------------------------------------------------------
// buildScopedDiff — AC-4: Empty file list validation
// ---------------------------------------------------------------------------

func TestBuildScopedDiff_AC4_EmptyInput(t *testing.T) {
	t.Parallel()
	// BDD Scenario: Empty filePaths emits error
	// Given filePaths is empty or nil
	// When buildScopedDiff is called
	// Then it returns an error containing "no files selected"

	repoPath := initTestGitRepo(t)

	tests := []struct {
		name      string
		filePaths []string
	}{
		{name: "empty slice", filePaths: []string{}},
		{name: "nil slice", filePaths: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := buildScopedDiff(context.Background(), repoPath, tt.filePaths)
			require.Error(t, err, "%s must return an error", tt.name)
			assert.Contains(t, err.Error(), "no files selected")
		})
	}
}

// ---------------------------------------------------------------------------
// buildScopedDiff — All files unchanged (empty diff guard)
// ---------------------------------------------------------------------------

func TestBuildScopedDiff_AllFilesUnchanged(t *testing.T) {
	t.Parallel()
	// BDD Scenario: All selected files have no changes
	// Given filePaths contains "unchanged.go" which has no diff
	// When buildScopedDiff assembles the diff
	// Then the returned diff is empty

	repoPath := initTestGitRepo(t)
	commitFile(t, repoPath, "unchanged.go", "package main\n")

	diff, err := buildScopedDiff(context.Background(), repoPath, []string{"unchanged.go"})
	require.NoError(t, err)
	assert.Empty(t, diff, "unchanged tracked file must produce empty diff")
}

// ---------------------------------------------------------------------------
// buildScopedDiff — Path traversal defence
// ---------------------------------------------------------------------------

func TestBuildScopedDiff_RejectsPathTraversal(t *testing.T) {
	// Security: a filePath like "../../etc/passwd" must not escape the repo
	// and leak external file content through the --no-index fallback.
	// Traversal paths are silently skipped.

	repoPath := initTestGitRepo(t)

	// Plant a sentinel file *outside* the repo whose content must NOT leak.
	outsideDir := t.TempDir()
	secretName := "leaked_secret.txt"
	secretContent := "SENTINEL_SECRET_MUST_NOT_LEAK_123"
	secretPath := filepath.Join(outsideDir, secretName)
	require.NoError(t, os.WriteFile(secretPath, []byte(secretContent), 0o600))

	// Build a relative traversal from repoPath up to outsideDir/secretName.
	// e.g. "../<basename>/leaked_secret.txt" (or deeper) depending on TempDir layout.
	rel, err := filepath.Rel(repoPath, secretPath)
	require.NoError(t, err)
	require.Contains(t, rel, "..", "traversal relative path must contain ..")

	// Also include a classic absolute-escape pattern.
	traversalPaths := []string{
		rel,
		"../../etc/passwd",
	}

	diff, err := buildScopedDiff(context.Background(), repoPath, traversalPaths)
	require.NoError(t, err, "traversal paths are silently skipped, not errored")

	assert.NotContains(t, diff, secretContent,
		"external file content must never appear in diff output")
	assert.NotContains(t, diff, "root:", "passwd-shaped content must not leak")
}

// ---------------------------------------------------------------------------
// buildScopedDiff — Separator between files
// ---------------------------------------------------------------------------

func TestBuildScopedDiff_SeparatorBetweenFiles(t *testing.T) {
	t.Parallel()
	repoPath := initTestGitRepo(t)

	commitFile(t, repoPath, "a.go", "package main\n")
	commitFile(t, repoPath, "b.go", "package main\n")
	modifyTrackedFile(t, repoPath, "a.go", "package main\n// a\n")
	modifyTrackedFile(t, repoPath, "b.go", "package main\n// b\n")

	diff, err := buildScopedDiff(context.Background(), repoPath, []string{"a.go", "b.go"})
	require.NoError(t, err)
	assert.Contains(t, diff, "\n---\n", "multiple file diffs must be separated by ---")
}

// ---------------------------------------------------------------------------
// assembleScopedPayload — AC-3: Additional context prepend
// ---------------------------------------------------------------------------

func TestAssembleScopedPayload_AC3_AdditionalContextPrepended(t *testing.T) {
	t.Parallel()
	// BDD Scenario: Prior advice is prepended to diff
	// Given additionalContext is non-empty
	// When the stdin payload is assembled
	// Then it starts with "## Prior Context", contains the context and "## Code Changes"

	tests := []struct {
		name              string
		diff              string
		additionalContext string
		wantContains      []string
	}{
		{
			name:              "single line context",
			diff:              "diff --git a/main.go b/main.go\n+// changed",
			additionalContext: "Previous advice: refactor error handling in main.go",
			wantContains: []string{
				"Previous advice: refactor error handling in main.go",
				"## Code Changes",
				"diff --git a/main.go b/main.go",
			},
		},
		{
			name:              "multiline context",
			diff:              "diff --git a/util.go b/util.go\n+func Helper() {}",
			additionalContext: "Line 1 of context\nLine 2 of context",
			wantContains: []string{
				"Line 1 of context",
				"Line 2 of context",
				"## Code Changes",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			payload := assembleScopedPayload(tt.diff, tt.additionalContext)

			assert.True(t, strings.HasPrefix(payload, "## Prior Context"),
				"payload must start with '## Prior Context', got prefix: %q",
				payload[:min(len(payload), 50)])

			for _, want := range tt.wantContains {
				assert.Contains(t, payload, want)
			}

			// Ordering: Prior Context must appear before Code Changes
			priorIdx := strings.Index(payload, "## Prior Context")
			changesIdx := strings.Index(payload, "## Code Changes")
			assert.Greater(t, changesIdx, priorIdx,
				"## Prior Context must appear before ## Code Changes")
		})
	}
}

func TestAssembleScopedPayload_EmptyContext(t *testing.T) {
	t.Parallel()
	// BDD Scenario: Empty additional context is omitted
	// Given additionalContext is ""
	// When the stdin payload is assembled
	// Then it does not contain "## Prior Context" and returns raw diff

	diff := "diff --git a/main.go b/main.go\n+// changed"

	payload := assembleScopedPayload(diff, "")

	assert.NotContains(t, payload, "## Prior Context",
		"empty context must not include Prior Context header")
	assert.Equal(t, diff, payload,
		"empty context must return raw diff unchanged")
}

// ---------------------------------------------------------------------------
// scopedAdviceEvent — AC-5: Event compatibility
// ---------------------------------------------------------------------------

func TestScopedAdviceEvent_AC5_MatchesStreamAdviceShape(t *testing.T) {
	t.Parallel()
	// AC-5: Events have the same shape as StreamAdvice events
	// Required keys: repoPath, text, done, error

	tests := []struct {
		name     string
		repoPath string
		text     string
		done     bool
		errMsg   string
	}{
		{
			name:     "progress event",
			repoPath: "/tmp/repo",
			text:     "analyzing code...",
			done:     false,
			errMsg:   "",
		},
		{
			name:     "done event",
			repoPath: "/tmp/repo",
			text:     "",
			done:     true,
			errMsg:   "",
		},
		{
			name:     "error event",
			repoPath: "/tmp/repo",
			text:     "",
			done:     true,
			errMsg:   "Claude CLI not found",
		},
	}

	requiredKeys := []string{"repoPath", "text", "done", "error"}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			event := scopedAdviceEvent(tt.repoPath, tt.text, tt.done, tt.errMsg)
			require.NotNil(t, event, "event map must not be nil")

			for _, key := range requiredKeys {
				_, exists := event[key]
				assert.True(t, exists,
					"event must contain key %q for StreamAdvice compatibility", key)
			}

			assert.Equal(t, tt.repoPath, event["repoPath"])
			assert.Equal(t, tt.text, event["text"])
			assert.Equal(t, tt.done, event["done"])
			assert.Equal(t, tt.errMsg, event["error"])
		})
	}
}
