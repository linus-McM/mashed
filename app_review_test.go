package main

import (
	"context"
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

	// Store a cancel func (as production code does).
	ctx, cancel1 := context.WithCancel(context.Background())
	defer cancel1()
	activeReviews.Store(repoPath, cancel1)

	// Second store for same repo should find existing entry.
	prev, loaded := activeReviews.Load(repoPath)
	assert.True(t, loaded, "should find existing entry")
	assert.NotNil(t, prev, "stored value should be a cancel func")

	// Cancel the old and replace.
	if c, ok := prev.(context.CancelFunc); ok {
		c()
	}
	assert.Error(t, ctx.Err(), "old context should be cancelled")

	_, cancel2 := context.WithCancel(context.Background())
	defer cancel2()
	activeReviews.Store(repoPath, cancel2)

	// Delete clears it.
	activeReviews.Delete(repoPath)
	_, loaded = activeReviews.Load(repoPath)
	assert.False(t, loaded, "should be empty after Delete")
}

func TestReviewConcurrencyGuard_DifferentRepos(t *testing.T) {
	activeReviews = sync.Map{}

	repo1 := "/tmp/repo-a"
	repo2 := "/tmp/repo-b"

	_, cancel1 := context.WithCancel(context.Background())
	defer cancel1()
	_, cancel2 := context.WithCancel(context.Background())
	defer cancel2()

	activeReviews.Store(repo1, cancel1)
	activeReviews.Store(repo2, cancel2)

	_, loaded1 := activeReviews.Load(repo1)
	_, loaded2 := activeReviews.Load(repo2)

	assert.True(t, loaded1, "repo1 should be stored")
	assert.True(t, loaded2, "repo2 should be stored independently")

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

func TestModelRegistry_FallbackHasDefault(t *testing.T) {
	models := domain.FallbackModels()
	require.NotEmpty(t, models)

	defaultAlias := domain.DefaultAlias(models)
	assert.NotEmpty(t, defaultAlias)

	var found bool
	for _, m := range models {
		if m.Alias == defaultAlias {
			found = true
			assert.True(t, m.IsDefault)
		}
	}
	assert.True(t, found, "default model should exist in FallbackModels")
}

func TestModelRegistry_AliasLookup(t *testing.T) {
	models := domain.FallbackModels()
	tests := []struct {
		alias     string
		wantAlias string
	}{
		{"opus", "opus"},
		{"sonnet", "sonnet"},
		{"haiku", "haiku"},
		{"claude-opus-4-6", "opus"}, // full ID lookup works too
	}

	for _, tt := range tests {
		t.Run(tt.alias, func(t *testing.T) {
			m := domain.ModelByAlias(models, tt.alias)
			assert.Equal(t, tt.wantAlias, m.Alias)
		})
	}
}

func TestModelRegistry_UnknownAliasFallsBack(t *testing.T) {
	models := domain.FallbackModels()
	m := domain.ModelByAlias(models, "nonexistent-model")
	assert.Equal(t, domain.DefaultAlias(models), m.Alias, "unknown alias should fall back to default")
}

func TestParseModelResponse_ValidJSON(t *testing.T) {
	// Plain text JSON from --json-schema (no --output-format json)
	input := `{"models":[{"alias":"opus","id":"claude-opus-4-6","displayName":"Opus 4.6","contextWindow":1000000,"tier":"powerful","isDefault":true},{"alias":"sonnet","id":"claude-sonnet-4-6","displayName":"Sonnet 4.6","contextWindow":200000,"tier":"balanced","isDefault":false}]}`

	models, err := parseModelResponse([]byte(input))
	require.NoError(t, err)
	require.Len(t, models, 2)
	assert.Equal(t, "opus", models[0].Alias)
	assert.Equal(t, "sonnet", models[1].Alias)
	assert.True(t, models[0].IsDefault)
	assert.Equal(t, 1000000, models[0].ContextWindow)
}

func TestParseModelResponse_EmptyInput(t *testing.T) {
	_, err := parseModelResponse([]byte(""))
	assert.Error(t, err)
}

func TestParseModelResponse_InvalidJSON(t *testing.T) {
	_, err := parseModelResponse([]byte("not json at all"))
	assert.Error(t, err)
}

func TestParseModelResponse_JSONInCodeFence(t *testing.T) {
	// Claude sometimes wraps JSON in markdown code fences
	input := "```json\n" + `{"models":[{"alias":"opus","id":"claude-opus-4-6","displayName":"Opus 4.6","contextWindow":1000000,"tier":"powerful","isDefault":true}]}` + "\n```"

	models, err := parseModelResponse([]byte(input))
	require.NoError(t, err)
	require.Len(t, models, 1)
	assert.Equal(t, "opus", models[0].Alias)
}

func TestParseModelResponse_JSONWithSurroundingProse(t *testing.T) {
	input := "Here are the models:\n" + `{"models":[{"alias":"sonnet","id":"claude-sonnet-4-6","displayName":"Sonnet 4.6","contextWindow":200000,"tier":"balanced","isDefault":false}]}` + "\nHope that helps!"

	models, err := parseModelResponse([]byte(input))
	require.NoError(t, err)
	require.Len(t, models, 1)
	assert.Equal(t, "sonnet", models[0].Alias)
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
			_, err := app.SpawnRefactorPlan(tc.repoPath, tc.adviceText, nil)
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

// ---------------------------------------------------------------------------
// refactorPlanFilename — file-referencing plan filenames
// ---------------------------------------------------------------------------

func TestRefactorPlanFilename(t *testing.T) {
	t.Parallel()

	const ts int64 = 1712700000

	tests := []struct {
		name      string
		filePaths []string
		want      string
	}{
		{
			name:      "nil filePaths falls back to bare refactor name",
			filePaths: nil,
			want:      "refactor-1712700000.md",
		},
		{
			name:      "empty slice falls back to bare refactor name",
			filePaths: []string{},
			want:      "refactor-1712700000.md",
		},
		{
			name:      "single file replaces extension dot with underscore",
			filePaths: []string{"main.go"},
			want:      "refactor-main_go-1712700000.md",
		},
		{
			name:      "single file in subdirectory uses basename only",
			filePaths: []string{"internal/bmad/executor.go"},
			want:      "refactor-executor_go-1712700000.md",
		},
		{
			name:      "multi-dotted filename converts only the final dot",
			filePaths: []string{"foo.bar.go"},
			want:      "refactor-foo.bar_go-1712700000.md",
		},
		{
			name:      "two files join with plus",
			filePaths: []string{"main.go", "util.go"},
			want:      "refactor-main_go+util_go-1712700000.md",
		},
		{
			name:      "three files join with plus",
			filePaths: []string{"main.go", "util.go", "config.go"},
			want:      "refactor-main_go+util_go+config_go-1712700000.md",
		},
		{
			name:      "four files collapse the extras into +Nmore tail",
			filePaths: []string{"main.go", "util.go", "config.go", "server.go"},
			want:      "refactor-main_go+util_go+config_go+1more-1712700000.md",
		},
		{
			name:      "five files collapse the extras into +Nmore tail",
			filePaths: []string{"main.go", "util.go", "config.go", "server.go", "client.go"},
			want:      "refactor-main_go+util_go+config_go+2more-1712700000.md",
		},
		{
			name:      "typescript extension is slugified",
			filePaths: []string{"SummarisationModal.svelte"},
			want:      "refactor-SummarisationModal_svelte-1712700000.md",
		},
		{
			name:      "unsafe characters in filename become underscores",
			filePaths: []string{"weird name (v2).ts"},
			want:      "refactor-weird_name__v2__ts-1712700000.md",
		},
		{
			name:      "dotfile without extension stays intact",
			filePaths: []string{".gitignore"},
			want:      "refactor-.gitignore-1712700000.md",
		},
		{
			name:      "file without extension is kept verbatim",
			filePaths: []string{"Makefile"},
			want:      "refactor-Makefile-1712700000.md",
		},
		{
			name:      "whitespace-only path is skipped",
			filePaths: []string{"   "},
			want:      "refactor-1712700000.md",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := refactorPlanFilename(tc.filePaths, ts)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestSlugifyPlanPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{"plain go file", "main.go", "main_go"},
		{"nested path uses basename", "pkg/util/helper.go", "helper_go"},
		{"multi-dot name only converts last", "server.handlers.go", "server.handlers_go"},
		{"dotfile without extension", ".env", ".env"},
		{"no extension", "Makefile", "Makefile"},
		{"empty string", "", ""},
		{"only whitespace", "  \t", ""},
		{"punctuation becomes underscores", "a b+c.ts", "a_b_c_ts"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := slugifyPlanPath(tc.in)
			assert.Equal(t, tc.want, got)
		})
	}
}
