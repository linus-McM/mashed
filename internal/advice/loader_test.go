package advice

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeAdviceFile is a test helper that creates a .md advice file with frontmatter.
func writeAdviceFile(t *testing.T, dir, filename, content string) string {
	t.Helper()
	err := os.MkdirAll(dir, 0o755)
	require.NoError(t, err)
	path := filepath.Join(dir, filename)
	err = os.WriteFile(path, []byte(content), 0o644)
	require.NoError(t, err)
	return path
}

func TestParseAdviceFile(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		source    string
		wantMode  AdviceMode
		wantErr   bool
		errSubstr string
	}{
		{
			name: "valid file with all fields",
			content: `---
name: strategic
displayName: Strategic Advisor
icon: compass
order: 10
---
This is the body.

It has multiple lines.
`,
			source: "local",
			wantMode: AdviceMode{
				Name:        "strategic",
				DisplayName: "Strategic Advisor",
				Icon:        "compass",
				Order:       10,
				Body:        "This is the body.\n\nIt has multiple lines.\n",
				Source:      "local",
			},
		},
		{
			name: "missing displayName defaults to name",
			content: `---
name: tactical
order: 5
---
Body here.
`,
			source: "global",
			wantMode: AdviceMode{
				Name:        "tactical",
				DisplayName: "tactical",
				Icon:        "",
				Order:       5,
				Body:        "Body here.\n",
				Source:      "global",
			},
		},
		{
			name: "missing icon defaults to empty",
			content: `---
name: review
displayName: Code Reviewer
order: 1
---
Review body.
`,
			source: "bundled",
			wantMode: AdviceMode{
				Name:        "review",
				DisplayName: "Code Reviewer",
				Icon:        "",
				Order:       1,
				Body:        "Review body.\n",
				Source:      "bundled",
			},
		},
		{
			name: "missing order defaults to 0",
			content: `---
name: minimal
displayName: Minimal Mode
icon: dot
---
Minimal body.
`,
			source: "local",
			wantMode: AdviceMode{
				Name:        "minimal",
				DisplayName: "Minimal Mode",
				Icon:        "dot",
				Order:       0,
				Body:        "Minimal body.\n",
				Source:      "local",
			},
		},
		{
			name:      "missing name returns error",
			content:   "---\ndisplayName: No Name\norder: 1\n---\nBody.\n",
			source:    "local",
			wantErr:   true,
			errSubstr: "name",
		},
		{
			name:      "no frontmatter returns error",
			content:   "Just a plain markdown file.\nNo frontmatter here.\n",
			source:    "local",
			wantErr:   true,
			errSubstr: "frontmatter",
		},
		{
			name:      "only one delimiter returns error",
			content:   "---\nname: broken\nNo closing delimiter.\n",
			source:    "local",
			wantErr:   true,
			errSubstr: "frontmatter",
		},
		{
			name:      "malformed YAML returns error",
			content:   "---\nname: [invalid yaml\norder: not a number\n---\nBody.\n",
			source:    "local",
			wantErr:   true,
			errSubstr: "", // any error from YAML parsing
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := writeAdviceFile(t, dir, "test.md", tt.content)

			mode, err := parseAdviceFile(path, tt.source)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errSubstr != "" {
					assert.Contains(t, err.Error(), tt.errSubstr)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantMode.Name, mode.Name)
			assert.Equal(t, tt.wantMode.DisplayName, mode.DisplayName)
			assert.Equal(t, tt.wantMode.Icon, mode.Icon)
			assert.Equal(t, tt.wantMode.Order, mode.Order)
			assert.Equal(t, tt.wantMode.Body, mode.Body)
			assert.Equal(t, tt.wantMode.Source, mode.Source)
			assert.Equal(t, path, mode.FilePath)
		})
	}
}

func TestLoadAdviceModes_MergePriority(t *testing.T) {
	// Set up three sources: global and local directories.
	// Bundled defaults come from the embedded FS (empty in tests).
	// We test global + local merge priority.
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "repo")

	globalDir := filepath.Join(tmpDir, "global-advice")
	localDir := filepath.Join(repoPath, ".claude", "advice")

	// Global has "strategic" and "review"
	writeAdviceFile(t, globalDir, "strategic.md", `---
name: strategic
displayName: Global Strategic
icon: globe
order: 10
---
Global strategic body.
`)
	writeAdviceFile(t, globalDir, "review.md", `---
name: review
displayName: Code Review
icon: eye
order: 20
---
Global review body.
`)

	// Local has "strategic" (override) and "tactical" (new)
	writeAdviceFile(t, localDir, "strategic.md", `---
name: strategic
displayName: Local Strategic
icon: target
order: 5
---
Local strategic body.
`)
	writeAdviceFile(t, localDir, "tactical.md", `---
name: tactical
displayName: Tactical Advisor
icon: sword
order: 15
---
Local tactical body.
`)

	// Use custom loading to inject test paths
	modes, err := loadAdviceModesFromDirs(nil, globalDir, localDir)
	require.NoError(t, err)

	require.Len(t, modes, 3)

	// Sorted by Order: strategic(5), tactical(15), review(20)
	assert.Equal(t, "strategic", modes[0].Name)
	assert.Equal(t, "Local Strategic", modes[0].DisplayName) // local overrides global
	assert.Equal(t, "target", modes[0].Icon)
	assert.Equal(t, 5, modes[0].Order)
	assert.Equal(t, "local", modes[0].Source)

	assert.Equal(t, "tactical", modes[1].Name)
	assert.Equal(t, 15, modes[1].Order)

	assert.Equal(t, "review", modes[2].Name)
	assert.Equal(t, "Code Review", modes[2].DisplayName)
	assert.Equal(t, 20, modes[2].Order)
}

func TestLoadAdviceModes_SortOrderThenDisplayName(t *testing.T) {
	dir := t.TempDir()

	writeAdviceFile(t, dir, "bravo.md", `---
name: bravo
displayName: Bravo
order: 1
---
B body.
`)
	writeAdviceFile(t, dir, "alpha.md", `---
name: alpha
displayName: Alpha
order: 1
---
A body.
`)
	writeAdviceFile(t, dir, "charlie.md", `---
name: charlie
displayName: Charlie
order: 0
---
C body.
`)

	modes, err := loadAdviceModesFromDirs(nil, "", dir)
	require.NoError(t, err)
	require.Len(t, modes, 3)

	// Charlie (order 0) first, then Alpha (order 1, "Alpha" < "Bravo"), then Bravo (order 1)
	assert.Equal(t, "charlie", modes[0].Name)
	assert.Equal(t, "alpha", modes[1].Name)
	assert.Equal(t, "bravo", modes[2].Name)
}

func TestLoadAdviceModes_MissingDirectories(t *testing.T) {
	// All three directories don't exist — should return empty, no error
	modes, err := loadAdviceModesFromDirs(nil, "/nonexistent/global", "/nonexistent/local")
	require.NoError(t, err)
	assert.Empty(t, modes)
}

func TestLoadAdviceModes_NonMdFilesIgnored(t *testing.T) {
	dir := t.TempDir()

	writeAdviceFile(t, dir, "valid.md", `---
name: valid
displayName: Valid Mode
order: 1
---
Valid body.
`)

	// Write non-.md files that should be ignored
	writeAdviceFile(t, dir, "readme.txt", "not a mode")
	writeAdviceFile(t, dir, "config.yaml", "not: a mode")
	writeAdviceFile(t, dir, "notes", "plain file")

	modes, err := loadAdviceModesFromDirs(nil, "", dir)
	require.NoError(t, err)
	require.Len(t, modes, 1)
	assert.Equal(t, "valid", modes[0].Name)
}

func TestLoadAdviceModes_MalformedFilesSkipped(t *testing.T) {
	dir := t.TempDir()

	// Valid file
	writeAdviceFile(t, dir, "good.md", `---
name: good
displayName: Good Mode
order: 1
---
Good body.
`)

	// Malformed file (no name)
	writeAdviceFile(t, dir, "bad.md", `---
displayName: Bad No Name
order: 2
---
Bad body.
`)

	// Another malformed file (no frontmatter)
	writeAdviceFile(t, dir, "ugly.md", "No frontmatter at all.")

	modes, err := loadAdviceModesFromDirs(nil, "", dir)
	require.NoError(t, err)
	require.Len(t, modes, 1)
	assert.Equal(t, "good", modes[0].Name)
}

func TestLoadAdviceModes_EmptyRepoPathSkipsLocal(t *testing.T) {
	// With empty repoPath, LoadAdviceModes should not attempt local dir.
	// Returns bundled defaults (8) plus any global files.
	modes, err := LoadAdviceModes("")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(modes), 8, "should include at least the 8 bundled defaults")
}

func TestLoadAdviceBody(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "repo")
	localDir := filepath.Join(repoPath, ".claude", "advice")
	globalDir := filepath.Join(tmpDir, "global-advice")

	// Global has "strategic"
	writeAdviceFile(t, globalDir, "strategic.md", `---
name: strategic
displayName: Strategic
order: 1
---
Global strategic body content.
`)

	// Local has "strategic" (higher priority)
	writeAdviceFile(t, localDir, "strategic.md", `---
name: strategic
displayName: Strategic
order: 1
---
Local strategic body content.
`)

	// Local also has "tactical"
	writeAdviceFile(t, localDir, "tactical.md", `---
name: tactical
displayName: Tactical
order: 2
---
Tactical body content.
`)

	body, err := loadAdviceBodyFromDirs("strategic", nil, globalDir, localDir)
	require.NoError(t, err)
	assert.Equal(t, "Local strategic body content.\n", body) // local wins

	body, err = loadAdviceBodyFromDirs("tactical", nil, globalDir, localDir)
	require.NoError(t, err)
	assert.Equal(t, "Tactical body content.\n", body)

	_, err = loadAdviceBodyFromDirs("nonexistent", nil, globalDir, localDir)
	assert.Error(t, err)
}

func TestParseAdviceFileFromFS(t *testing.T) {
	// Test parsing a valid embedded file.
	mode, err := parseAdviceFileFromFS(defaultsFS, "defaults/strategic.md")
	require.NoError(t, err)
	assert.Equal(t, "strategic", mode.Name)
	assert.Equal(t, "Strategic Advisor", mode.DisplayName)
	assert.Equal(t, "compass", mode.Icon)
	assert.Equal(t, 10, mode.Order)
	assert.Equal(t, "bundled", mode.Source)
	assert.NotEmpty(t, mode.Body)

	// Test parsing a non-existent embedded file returns an error.
	_, err = parseAdviceFileFromFS(defaultsFS, "defaults/nonexistent.md")
	assert.Error(t, err)
}

func TestParseAdviceFile_NonexistentFile(t *testing.T) {
	_, err := parseAdviceFile("/nonexistent/path/file.md", "local")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reading advice file")
}

func TestLoadAdviceBody_GlobalOnly(t *testing.T) {
	tmpDir := t.TempDir()
	globalDir := filepath.Join(tmpDir, "global")

	writeAdviceFile(t, globalDir, "design.md", `---
name: design
displayName: Design Advisor
order: 1
---
Design body from global.
`)

	body, err := loadAdviceBodyFromDirs("design", nil, globalDir, "")
	require.NoError(t, err)
	assert.Equal(t, "Design body from global.\n", body)
}

func TestBundledDefaults_AllFilesValid(t *testing.T) {
	// AC-1: All 8 advice files exist and load.
	// AC-2: Each file has valid frontmatter with all 4 fields.
	// AC-3: Each file has substantive body (>100 words).
	// AC-4: Files load via go:embed and are sorted by order.

	modes, err := loadAdviceModesFromDirs(&defaultsFS, "", "")
	require.NoError(t, err)
	require.Len(t, modes, 8, "expected 8 bundled defaults")

	expectedModes := []struct {
		name        string
		displayName string
		icon        string
		order       int
	}{
		{"strategic", "Strategic Advisor", "compass", 10},
		{"extreme-programming", "Extreme Programming", "zap", 20},
		{"clean-code", "Clean Code", "sparkles", 30},
		{"solid", "SOLID Principles", "blocks", 40},
		{"domain-driven-design", "Domain-Driven Design", "layers", 50},
		{"security-first", "Security First", "shield", 60},
		{"performance", "Performance", "gauge", 70},
		{"pragmatic", "Pragmatic Developer", "wrench", 80},
	}

	for i, expected := range expectedModes {
		mode := modes[i]
		t.Run(expected.name, func(t *testing.T) {
			// AC-2: All frontmatter fields present and correct
			assert.Equal(t, expected.name, mode.Name, "name mismatch")
			assert.Equal(t, expected.displayName, mode.DisplayName, "displayName mismatch")
			assert.Equal(t, expected.icon, mode.Icon, "icon mismatch")
			assert.Equal(t, expected.order, mode.Order, "order mismatch")
			assert.Equal(t, "bundled", mode.Source, "source should be bundled")

			// AC-3: Body is non-empty and has at least 100 words
			assert.NotEmpty(t, mode.Body, "body should not be empty")
			wordCount := len(strings.Fields(mode.Body))
			assert.GreaterOrEqual(t, wordCount, 100,
				"body should have at least 100 words, got %d", wordCount)
		})
	}

	// AC-4: Verify order is strictly ascending
	for i := 1; i < len(modes); i++ {
		assert.Less(t, modes[i-1].Order, modes[i].Order,
			"modes should be sorted by order: %s(%d) should come before %s(%d)",
			modes[i-1].Name, modes[i-1].Order, modes[i].Name, modes[i].Order)
	}
}

func TestBundledDefaults_BodyLoadable(t *testing.T) {
	// Verify that LoadAdviceBody can retrieve each bundled mode's body.
	expectedNames := []string{
		"strategic", "extreme-programming", "clean-code", "solid",
		"domain-driven-design", "security-first", "performance", "pragmatic",
	}

	for _, name := range expectedNames {
		t.Run(name, func(t *testing.T) {
			body, err := loadAdviceBodyFromDirs(name, &defaultsFS, "", "")
			require.NoError(t, err)
			assert.NotEmpty(t, body)
			assert.Contains(t, body, "## Focus Areas")
			assert.Contains(t, body, "## Review Approach")
			assert.Contains(t, body, "## What to Look For")
		})
	}
}

func TestBundledDefaults_LocalOverridesBundled(t *testing.T) {
	// Verify that a local file with the same name overrides the bundled version.
	tmpDir := t.TempDir()
	localDir := filepath.Join(tmpDir, "local")

	writeAdviceFile(t, localDir, "strategic.md", `---
name: strategic
displayName: My Custom Strategic
icon: star
order: 1
---
Custom local strategic body.
`)

	modes, err := loadAdviceModesFromDirs(&defaultsFS, "", localDir)
	require.NoError(t, err)

	// Should have 8 total (7 bundled + 1 local override)
	require.Len(t, modes, 8)

	// Find strategic - it should be the local version
	var strategic AdviceMode
	for _, m := range modes {
		if m.Name == "strategic" {
			strategic = m
			break
		}
	}
	assert.Equal(t, "My Custom Strategic", strategic.DisplayName)
	assert.Equal(t, "star", strategic.Icon)
	assert.Equal(t, 1, strategic.Order)
	assert.Equal(t, "local", strategic.Source)
}

func TestLoadAdviceBody_BundledFallback(t *testing.T) {
	// With the real embedded FS (which is empty), this should return not found.
	_, err := loadAdviceBodyFromDirs("nonexistent", &defaultsFS, "", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestLoadAdviceModes_BundledWithEmbeddedFS(t *testing.T) {
	// Test that the real embedded FS loads all 8 bundled defaults.
	modes, err := loadAdviceModesFromDirs(&defaultsFS, "", "")
	require.NoError(t, err)
	require.Len(t, modes, 8)

	// Verify sorted by order 10..80
	expectedNames := []string{
		"strategic", "extreme-programming", "clean-code", "solid",
		"domain-driven-design", "security-first", "performance", "pragmatic",
	}
	for i, name := range expectedNames {
		assert.Equal(t, name, modes[i].Name, "mode at index %d", i)
		assert.Equal(t, "bundled", modes[i].Source)
		assert.Equal(t, (i+1)*10, modes[i].Order)
	}
}

func TestLoadAdviceModes_AllThreeSources(t *testing.T) {
	tmpDir := t.TempDir()
	globalDir := filepath.Join(tmpDir, "global")
	localDir := filepath.Join(tmpDir, "local")

	writeAdviceFile(t, globalDir, "global-only.md", `---
name: globalonly
displayName: Global Only
order: 95
---
Global only body.
`)

	writeAdviceFile(t, localDir, "local-only.md", `---
name: localonly
displayName: Local Only
order: 5
---
Local only body.
`)

	// Pass the real embedded FS (8 bundled defaults) + global + local
	modes, err := loadAdviceModesFromDirs(&defaultsFS, globalDir, localDir)
	require.NoError(t, err)
	require.Len(t, modes, 10) // 8 bundled + 1 global + 1 local

	// Sorted by order: localonly(5), strategic(10), ..., pragmatic(80), globalonly(95)
	assert.Equal(t, "localonly", modes[0].Name)
	assert.Equal(t, "local", modes[0].Source)
	assert.Equal(t, "globalonly", modes[9].Name)
	assert.Equal(t, "global", modes[9].Source)

	// Bundled modes are in the middle
	assert.Equal(t, "strategic", modes[1].Name)
	assert.Equal(t, "bundled", modes[1].Source)
	assert.Equal(t, "pragmatic", modes[8].Name)
	assert.Equal(t, "bundled", modes[8].Source)
}

func TestFindModeInDir_NotFound(t *testing.T) {
	dir := t.TempDir()
	writeAdviceFile(t, dir, "other.md", `---
name: other
displayName: Other
order: 1
---
Other body.
`)

	_, err := findModeInDir(dir, "nonexistent", "local")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestFindModeInDir_NonexistentDir(t *testing.T) {
	_, err := findModeInDir("/nonexistent/dir", "test", "local")
	require.Error(t, err)
}

func TestSplitFrontmatter_EmptyBody(t *testing.T) {
	content := "---\nname: test\n---\n"
	fm, body, err := splitFrontmatter(content)
	require.NoError(t, err)
	assert.Equal(t, "name: test", fm)
	assert.Equal(t, "", body)
}

func TestLoadModesFromDir_NonexistentDir(t *testing.T) {
	modes, err := loadModesFromDir("/nonexistent/dir", "local")
	require.NoError(t, err)
	assert.Nil(t, modes)
}

func TestLoadAdviceBody_PublicAPI(t *testing.T) {
	// Test the public LoadAdviceBody function with empty repoPath
	// (no local dir). It will check global dir (likely empty) and bundled (empty).
	_, err := LoadAdviceBody("", "nonexistent-mode")
	assert.Error(t, err)
}

func TestSplitFrontmatter_CRLFLineEndings(t *testing.T) {
	content := "---\r\nname: test\r\n---\r\nBody with CRLF.\r\n"
	fm, body, err := splitFrontmatter(content)
	require.NoError(t, err)
	assert.Contains(t, fm, "name: test")
	assert.Contains(t, body, "Body with CRLF.")
}

func TestSplitFrontmatter_OpeningCRLF(t *testing.T) {
	content := "---\r\nname: crlftest\n---\nBody text.\n"
	fm, body, err := splitFrontmatter(content)
	require.NoError(t, err)
	assert.Contains(t, fm, "name: crlftest")
	assert.Equal(t, "Body text.\n", body)
}

func TestGlobalAdviceDir(t *testing.T) {
	dir, err := globalAdviceDir()
	require.NoError(t, err)
	assert.Contains(t, dir, ".mashed")
	assert.Contains(t, dir, "advice")
}

func TestLocalAdviceDir(t *testing.T) {
	result := localAdviceDir("/some/repo")
	assert.Equal(t, "/some/repo/.claude/advice", result)
}
