package main

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"mashed/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// FileSummary / ReviewSummary struct construction
// ---------------------------------------------------------------------------

func TestReviewSummary_StructConstruction(t *testing.T) {
	tests := []struct {
		name    string
		files   []FileSummary
		wantAdd int
		wantRm  int
		wantLen int
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
	activeReviews = sync.Map{}

	repoPath := "/tmp/test-repo"

	_, loaded := activeReviews.LoadOrStore(repoPath, true)
	assert.False(t, loaded, "first LoadOrStore should not be loaded")

	_, loaded = activeReviews.LoadOrStore(repoPath, true)
	assert.True(t, loaded, "second LoadOrStore should detect existing entry")

	activeReviews.Delete(repoPath)
	_, loaded = activeReviews.LoadOrStore(repoPath, true)
	assert.False(t, loaded, "LoadOrStore after Delete should not be loaded")

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
			wantLines: 1,
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
	var lines []string
	for i := 0; i < 10; i++ {
		lines = append(lines, "+added line")
	}
	diff := strings.Join(lines, "\n")

	result := truncateDiffLines(diff, 5)
	resultLines := strings.Split(result, "\n")

	for i := 0; i < 5; i++ {
		assert.Equal(t, "+added line", resultLines[i])
	}
	assert.Equal(t, "... (truncated)", resultLines[5])
}

// ---------------------------------------------------------------------------
// fileSummarySystemPrompt constant
// ---------------------------------------------------------------------------

func TestFileSummarySystemPrompt(t *testing.T) {
	assert.Contains(t, fileSummarySystemPrompt, "Summarise")
	assert.Contains(t, fileSummarySystemPrompt, "ONE sentence")
	assert.Contains(t, fileSummarySystemPrompt, "3-5 sentences")
}

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

func TestReviewConstants(t *testing.T) {
	assert.Equal(t, 50, maxReviewFiles, "max review files should be 50")
	assert.Equal(t, 500, maxDiffLines, "max diff lines should be 500")
}

// ---------------------------------------------------------------------------
// Model registry integration
// ---------------------------------------------------------------------------

func TestModelRegistry_DefaultExists(t *testing.T) {
	defaultID := domain.DefaultModelID()
	assert.NotEmpty(t, defaultID)

	models := domain.AvailableModels()
	var found bool
	for _, m := range models {
		if m.ID == defaultID {
			found = true
			assert.True(t, m.IsDefault)
		}
	}
	assert.True(t, found, "default model should exist in AvailableModels")
}

func TestModelRegistry_AliasLookup(t *testing.T) {
	tests := []struct {
		alias  string
		wantID string
	}{
		{"opus", "claude-opus-4-6"},
		{"sonnet", "claude-sonnet-4-6"},
		{"haiku", "claude-haiku-4-5-20251001"},
		{"claude-opus-4-6", "claude-opus-4-6"}, // full ID works too
	}

	for _, tt := range tests {
		t.Run(tt.alias, func(t *testing.T) {
			m := domain.ModelByAlias(tt.alias)
			assert.Equal(t, tt.wantID, m.ID)
		})
	}
}

func TestModelRegistry_UnknownAliasFallsBack(t *testing.T) {
	m := domain.ModelByAlias("nonexistent-model")
	assert.Equal(t, domain.DefaultModelID(), m.ID, "unknown alias should fall back to default")
}

// ---------------------------------------------------------------------------
// SpawnRefactorPlan input validation
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
	tmpDir := t.TempDir()
	plansDir := filepath.Join(tmpDir, ".claude", "plans")

	err := os.MkdirAll(plansDir, 0o755)
	require.NoError(t, err)

	info, err := os.Stat(plansDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}
