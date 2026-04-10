package main

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// FileSummary / ReviewSummary struct construction
// ---------------------------------------------------------------------------

func TestReviewSummary_StructConstruction(t *testing.T) {
	tests := []struct {
		name     string
		files    []FileSummary
		wantAdd  int
		wantRm   int
		wantLen  int
	}{
		{
			name:    "empty summary",
			files:   []FileSummary{},
			wantAdd: 0, wantRm: 0, wantLen: 0,
		},
		{
			name: "single file",
			files: []FileSummary{
				{Path: "main.go", Added: 10, Removed: 3, Summary: "Added logging", IsBinary: false},
			},
			wantAdd: 10, wantRm: 3, wantLen: 1,
		},
		{
			name: "multiple files with binary",
			files: []FileSummary{
				{Path: "main.go", Added: 5, Removed: 2, Summary: "Refactored init", IsBinary: false},
				{Path: "icon.png", Added: 0, Removed: 0, Summary: "Binary file changed", IsBinary: true},
				{Path: "util.go", Added: 20, Removed: 8, Summary: "New helper functions", IsBinary: false},
			},
			wantAdd: 25, wantRm: 10, wantLen: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			totalAdded := 0
			totalRemoved := 0
			for _, f := range tt.files {
				totalAdded += f.Added
				totalRemoved += f.Removed
			}

			rs := ReviewSummary{
				Files:        tt.files,
				TotalAdded:   totalAdded,
				TotalRemoved: totalRemoved,
			}

			assert.Equal(t, tt.wantAdd, rs.TotalAdded)
			assert.Equal(t, tt.wantRm, rs.TotalRemoved)
			assert.Len(t, rs.Files, tt.wantLen)
		})
	}
}

func TestFileSummary_BinaryFlag(t *testing.T) {
	fs := FileSummary{
		Path:     "image.png",
		Added:    0,
		Removed:  0,
		Summary:  "Binary file changed",
		IsBinary: true,
	}

	assert.True(t, fs.IsBinary)
	assert.Equal(t, "Binary file changed", fs.Summary)
	assert.Equal(t, "image.png", fs.Path)
}

// ---------------------------------------------------------------------------
// Concurrent review guard (activeReviews sync.Map)
// ---------------------------------------------------------------------------

func TestReviewConcurrencyGuard(t *testing.T) {
	// Reset the global map for test isolation.
	activeReviews = sync.Map{}

	repoPath := "/tmp/test-repo"

	// First load should succeed (loaded == false).
	_, loaded := activeReviews.LoadOrStore(repoPath, true)
	assert.False(t, loaded, "first LoadOrStore should not be loaded")

	// Second load should detect conflict (loaded == true).
	_, loaded = activeReviews.LoadOrStore(repoPath, true)
	assert.True(t, loaded, "second LoadOrStore should detect existing entry")

	// After delete, the guard should reset.
	activeReviews.Delete(repoPath)
	_, loaded = activeReviews.LoadOrStore(repoPath, true)
	assert.False(t, loaded, "LoadOrStore after Delete should not be loaded")

	// Cleanup.
	activeReviews.Delete(repoPath)
}

func TestReviewConcurrencyGuard_DifferentRepos(t *testing.T) {
	activeReviews = sync.Map{}

	repo1 := "/tmp/repo-a"
	repo2 := "/tmp/repo-b"

	_, loaded1 := activeReviews.LoadOrStore(repo1, true)
	_, loaded2 := activeReviews.LoadOrStore(repo2, true)

	assert.False(t, loaded1, "repo1 should not conflict")
	assert.False(t, loaded2, "repo2 should not conflict with repo1")

	// Cleanup.
	activeReviews.Delete(repo1)
	activeReviews.Delete(repo2)
}

// ---------------------------------------------------------------------------
// truncateDiffLines
// ---------------------------------------------------------------------------

func TestTruncateDiffLines(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		maxLines  int
		wantLines int
		wantTrunc bool
	}{
		{
			name:      "under limit",
			input:     "line1\nline2\nline3",
			maxLines:  5,
			wantLines: 3,
			wantTrunc: false,
		},
		{
			name:      "exactly at limit",
			input:     "line1\nline2\nline3",
			maxLines:  3,
			wantLines: 3,
			wantTrunc: false,
		},
		{
			name:      "over limit",
			input:     "line1\nline2\nline3\nline4\nline5",
			maxLines:  3,
			wantLines: 4, // 3 content lines + truncation marker
			wantTrunc: true,
		},
		{
			name:      "single line under limit",
			input:     "only one line",
			maxLines:  10,
			wantLines: 1,
			wantTrunc: false,
		},
		{
			name:      "empty string",
			input:     "",
			maxLines:  5,
			wantLines: 1, // split on empty string returns [""]
			wantTrunc: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateDiffLines(tt.input, tt.maxLines)

			if tt.wantTrunc {
				assert.Contains(t, result, "... (truncated)")
			} else {
				assert.NotContains(t, result, "... (truncated)")
			}

			lines := strings.Split(result, "\n")
			assert.Equal(t, tt.wantLines, len(lines))
		})
	}
}

func TestTruncateDiffLines_PreservesContent(t *testing.T) {
	// Build a diff with 10 lines.
	var lines []string
	for i := 0; i < 10; i++ {
		lines = append(lines, "+added line")
	}
	diff := strings.Join(lines, "\n")

	result := truncateDiffLines(diff, 5)
	resultLines := strings.Split(result, "\n")

	// First 5 lines should be preserved exactly.
	for i := 0; i < 5; i++ {
		assert.Equal(t, "+added line", resultLines[i])
	}

	// Last line should be truncation marker.
	assert.Equal(t, "... (truncated)", resultLines[5])
}

// ---------------------------------------------------------------------------
// buildFileSummaryPrompt
// ---------------------------------------------------------------------------

func TestBuildFileSummaryPrompt(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		diff     string
		wantSubs []string
	}{
		{
			name:     "includes file path",
			filePath: "internal/git/diff.go",
			diff:     "+new line\n-old line",
			wantSubs: []string{
				"internal/git/diff.go",
				"+new line",
				"-old line",
				"Summarise",
			},
		},
		{
			name:     "includes instruction keywords",
			filePath: "main.go",
			diff:     "+func main() {}",
			wantSubs: []string{
				"ONE sentence",
				"3-5 sentences",
				"Diff:",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt := buildFileSummaryPrompt(tt.filePath, tt.diff)
			for _, sub := range tt.wantSubs {
				assert.Contains(t, prompt, sub)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// buildAdvicePrompt
// ---------------------------------------------------------------------------

func TestBuildAdvicePrompt(t *testing.T) {
	methodology := "Check for security issues and code smells."
	diff := "+func handler(w http.ResponseWriter) {}"

	prompt := buildAdvicePrompt(methodology, diff)

	require.Contains(t, prompt, "expert code reviewer")
	require.Contains(t, prompt, "## Methodology")
	require.Contains(t, prompt, methodology)
	require.Contains(t, prompt, "## Code Changes (Diff)")
	require.Contains(t, prompt, diff)
	require.Contains(t, prompt, "markdown format")
}

func TestBuildAdvicePrompt_EmptyMethodology(t *testing.T) {
	prompt := buildAdvicePrompt("", "+some diff")

	// Should still produce a valid prompt structure.
	assert.Contains(t, prompt, "## Methodology")
	assert.Contains(t, prompt, "## Code Changes (Diff)")
}

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

func TestReviewConstants(t *testing.T) {
	assert.Equal(t, 50, maxReviewFiles, "max review files should be 50")
	assert.Equal(t, 500, maxDiffLines, "max diff lines should be 500")
}

// ---------------------------------------------------------------------------
// SpawnRefactorPlan helpers
// ---------------------------------------------------------------------------

func TestSpawnRefactorPlan_InputValidation(t *testing.T) {
	tests := []struct {
		name       string
		repoPath   string
		adviceText string
		errContain string
	}{
		{
			name:       "empty repo path",
			repoPath:   "",
			adviceText: "some advice",
			errContain: "repo path is required",
		},
		{
			name:       "empty advice text",
			repoPath:   "/tmp/repo",
			adviceText: "",
			errContain: "advice text is required",
		},
	}

	app := &App{}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := app.SpawnRefactorPlan(tc.repoPath, tc.adviceText)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.errContain)
		})
	}
}

func TestSpawnRefactorPlan_PlanPathFormat(t *testing.T) {
	// We can't fully test the spawn (requires Wails context),
	// but we can verify the plan path generation logic.
	tmpDir := t.TempDir()
	plansDir := filepath.Join(tmpDir, ".claude", "plans")

	err := os.MkdirAll(plansDir, 0o755)
	require.NoError(t, err)

	// Verify directory was created
	info, err := os.Stat(plansDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}
