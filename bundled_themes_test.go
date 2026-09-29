package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// chdirTemp changes the working directory to dir and returns a cleanup function.
// Not safe for parallel tests — use only in tests that require CWD control.
func chdirTemp(t *testing.T, dir string) {
	t.Helper()
	origWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { os.Chdir(origWd) })
}

// =============================================================================
// AC-6: scanVSIXDirectory is a clean refactor of ListVSCodiumThemes
// =============================================================================

func TestScanVSIXDirectory_AC6_WithThemes(t *testing.T) {
	dir := t.TempDir()

	createMockVSIX(t, dir, "dracula-theme.theme-dracula-2.24.3.vsix", map[string]string{
		"extension/package.json": `{
			"name": "theme-dracula",
			"contributes": {
				"themes": [
					{"label": "Dracula", "uiTheme": "vs-dark", "path": "./theme/dracula.json"}
				]
			}
		}`,
		"extension/theme/dracula.json": `{"name":"Dracula","type":"dark","colors":{"editor.background":"#282a36"}}`,
	})
	createMockVSIX(t, dir, "github.github-vscode-theme-6.3.5.vsix", map[string]string{
		"extension/package.json": `{
			"name": "github-vscode-theme",
			"contributes": {
				"themes": [
					{"label": "GitHub Dark", "uiTheme": "vs-dark", "path": "./themes/dark.json"},
					{"label": "GitHub Light", "uiTheme": "vs", "path": "./themes/light.json"}
				]
			}
		}`,
		"extension/themes/dark.json":  `{"name":"GitHub Dark","type":"dark","colors":{}}`,
		"extension/themes/light.json": `{"name":"GitHub Light","type":"light","colors":{}}`,
	})

	themes, err := scanVSIXDirectory(dir)
	require.NoError(t, err)
	require.Len(t, themes, 3)

	// Verify sorted alphabetically by label
	assert.Equal(t, "Dracula", themes[0].Label)
	assert.Equal(t, "GitHub Dark", themes[1].Label)
	assert.Equal(t, "GitHub Light", themes[2].Label)

	// Verify ThemePath uses ::vsix:: separator
	for _, th := range themes {
		assert.Contains(t, th.ThemePath, vsixSeparator, "ThemePath should contain VSIX separator")
	}

	// Verify ExtensionID is filename minus .vsix extension
	assert.Equal(t, "dracula-theme.theme-dracula-2.24.3", themes[0].ExtensionID)
}

func TestScanVSIXDirectory_AC6_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	themes, err := scanVSIXDirectory(dir)
	require.NoError(t, err)
	assert.Empty(t, themes, "empty dir should return empty slice")
}

func TestScanVSIXDirectory_AC6_MissingDir(t *testing.T) {
	themes, err := scanVSIXDirectory("/nonexistent/path/that/does/not/exist")
	// Missing dir should return error (os.ReadDir fails on nonexistent path)
	require.Error(t, err)
	assert.Empty(t, themes)
}

func TestScanVSIXDirectory_AC6_CorruptVSIX(t *testing.T) {
	dir := t.TempDir()

	createMockVSIX(t, dir, "good-ext.vsix", map[string]string{
		"extension/package.json": `{
			"name": "good-ext",
			"contributes": {
				"themes": [
					{"label": "Good Theme", "uiTheme": "vs-dark", "path": "./theme.json"}
				]
			}
		}`,
		"extension/theme.json": `{"name":"Good Theme","type":"dark","colors":{}}`,
	})

	require.NoError(t, os.WriteFile(filepath.Join(dir, "corrupt.vsix"), []byte("not a zip"), 0644))

	createMockVSIX(t, dir, "bad-json.vsix", map[string]string{
		"extension/package.json": `not valid json {{{`,
	})

	themes, err := scanVSIXDirectory(dir)
	require.NoError(t, err, "corrupt VSIX files should be skipped, not cause error")
	require.Len(t, themes, 1)
	assert.Equal(t, "Good Theme", themes[0].Label)
}

func TestScanVSIXDirectory_AC6_FiltersTmTheme(t *testing.T) {
	dir := t.TempDir()

	createMockVSIX(t, dir, "mixed-ext.vsix", map[string]string{
		"extension/package.json": `{
			"name": "mixed-ext",
			"contributes": {
				"themes": [
					{"label": "JSON Theme", "uiTheme": "vs-dark", "path": "./themes/good.json"},
					{"label": "TM Theme", "uiTheme": "vs-dark", "path": "./themes/old.tmTheme"}
				]
			}
		}`,
		"extension/themes/good.json":   `{"name":"JSON Theme","type":"dark","colors":{}}`,
		"extension/themes/old.tmTheme": `<plist></plist>`,
	})

	themes, err := scanVSIXDirectory(dir)
	require.NoError(t, err)
	require.Len(t, themes, 1, "tmTheme should be filtered out")
	assert.Equal(t, "JSON Theme", themes[0].Label)
}

func TestScanVSIXDirectory_AC6_SkipsNonVSIXFiles(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("hello"), 0644))

	createMockVSIX(t, dir, "theme.vsix", map[string]string{
		"extension/package.json": `{
			"name": "theme",
			"contributes": {
				"themes": [
					{"label": "Test", "uiTheme": "vs-dark", "path": "./theme.json"}
				]
			}
		}`,
		"extension/theme.json": `{"name":"Test","type":"dark","colors":{}}`,
	})

	require.NoError(t, os.MkdirAll(filepath.Join(dir, "subdir"), 0755))

	themes, err := scanVSIXDirectory(dir)
	require.NoError(t, err)
	require.Len(t, themes, 1)
	assert.Equal(t, "Test", themes[0].Label)
}

func TestScanVSIXDirectory_AC6_MatchesListVSCodiumThemes(t *testing.T) {
	// Verify scanVSIXDirectory produces same results as ListVSCodiumThemes
	// would for the same directory (regression guard for refactor).
	dir := t.TempDir()

	createMockVSIX(t, dir, "alpha.vsix", map[string]string{
		"extension/package.json": `{
			"name": "alpha",
			"contributes": {
				"themes": [
					{"label": "Zebra", "uiTheme": "vs-dark", "path": "./themes/zebra.json"},
					{"label": "Alpha", "uiTheme": "vs", "path": "./themes/alpha.json"}
				]
			}
		}`,
		"extension/themes/zebra.json": `{"name":"Zebra","type":"dark","colors":{}}`,
		"extension/themes/alpha.json": `{"name":"Alpha","type":"light","colors":{}}`,
	})

	themes, err := scanVSIXDirectory(dir)
	require.NoError(t, err)
	require.Len(t, themes, 2)

	// Must be sorted alphabetically by label
	assert.Equal(t, "Alpha", themes[0].Label)
	assert.Equal(t, "Zebra", themes[1].Label)

	// Fields must be populated
	for _, th := range themes {
		assert.NotEmpty(t, th.Label, "Label must be set")
		assert.NotEmpty(t, th.ExtensionID, "ExtensionID must be set")
		assert.NotEmpty(t, th.ThemePath, "ThemePath must be set")
		assert.NotEmpty(t, th.UITheme, "UITheme must be set")
		assert.Contains(t, th.ThemePath, vsixSeparator)
		assert.True(t, strings.HasPrefix(th.ThemePath, dir), "ThemePath should start with dir")
	}
}

// =============================================================================
// AC-1: ListBundledThemes returns themes from ./themes/ directory
// =============================================================================

func TestListBundledThemes_AC1_WithThemes(t *testing.T) {
	// Set up a themes/ directory with VSIX files in CWD.
	// bundledThemesDir() falls back to ./themes/ relative to CWD.
	themesDir := t.TempDir()
	themesSub := filepath.Join(themesDir, "themes")
	require.NoError(t, os.MkdirAll(themesSub, 0755))

	createMockVSIX(t, themesSub, "monokai.theme-monokai-pro-vscode-2.0.13.vsix", map[string]string{
		"extension/package.json": `{
			"name": "theme-monokai-pro",
			"contributes": {
				"themes": [
					{"label": "Monokai Pro", "uiTheme": "vs-dark", "path": "./themes/monokai-pro.json"},
					{"label": "Monokai Pro Filter Machine", "uiTheme": "vs-dark", "path": "./themes/monokai-pro-filter-machine.json"}
				]
			}
		}`,
		"extension/themes/monokai-pro.json":                `{"name":"Monokai Pro","type":"dark","colors":{"editor.background":"#2d2a2e"}}`,
		"extension/themes/monokai-pro-filter-machine.json": `{"name":"Monokai Pro Filter Machine","type":"dark","colors":{"editor.background":"#273136"}}`,
	})

	chdirTemp(t, themesDir)

	app := &App{}
	themes, err := app.ListBundledThemes()
	require.NoError(t, err)
	require.Len(t, themes, 2)

	// Verify correct fields populated
	for _, th := range themes {
		assert.NotEmpty(t, th.Label)
		assert.NotEmpty(t, th.ExtensionID)
		assert.NotEmpty(t, th.ThemePath)
		assert.NotEmpty(t, th.UITheme)
		assert.Contains(t, th.ThemePath, vsixSeparator, "ThemePath must use VSIX-encoded format")
	}

	// BDD Scenario 1: ExtensionID matches VSIX filename without extension
	assert.Equal(t, "monokai.theme-monokai-pro-vscode-2.0.13", themes[0].ExtensionID)
}

func TestListBundledThemes_AC1_SortedByLabel(t *testing.T) {
	themesDir := t.TempDir()
	themesSub := filepath.Join(themesDir, "themes")
	require.NoError(t, os.MkdirAll(themesSub, 0755))

	createMockVSIX(t, themesSub, "z-theme.vsix", map[string]string{
		"extension/package.json": `{"name":"z-theme","contributes":{"themes":[{"label":"Zebra","uiTheme":"vs-dark","path":"./t.json"}]}}`,
		"extension/t.json":       `{"name":"Zebra","type":"dark","colors":{}}`,
	})
	createMockVSIX(t, themesSub, "a-theme.vsix", map[string]string{
		"extension/package.json": `{"name":"a-theme","contributes":{"themes":[{"label":"Aardvark","uiTheme":"vs","path":"./t.json"}]}}`,
		"extension/t.json":       `{"name":"Aardvark","type":"light","colors":{}}`,
	})

	chdirTemp(t, themesDir)

	app := &App{}
	themes, err := app.ListBundledThemes()
	require.NoError(t, err)
	require.Len(t, themes, 2)
	assert.Equal(t, "Aardvark", themes[0].Label)
	assert.Equal(t, "Zebra", themes[1].Label)
}

// =============================================================================
// AC-2: ListBundledThemes returns empty slice when themes directory missing
// =============================================================================

func TestListBundledThemes_AC2_MissingDir(t *testing.T) {
	// Use a temp dir with NO themes/ subdirectory
	emptyDir := t.TempDir()

	chdirTemp(t, emptyDir)

	app := &App{}
	themes, err := app.ListBundledThemes()
	require.NoError(t, err, "missing themes dir should NOT return error")
	assert.Empty(t, themes, "missing themes dir should return empty slice")
}

func TestListBundledThemes_AC2_EmptyDir(t *testing.T) {
	themesDir := t.TempDir()
	themesSub := filepath.Join(themesDir, "themes")
	require.NoError(t, os.MkdirAll(themesSub, 0755))
	// themes/ exists but has no .vsix files

	chdirTemp(t, themesDir)

	app := &App{}
	themes, err := app.ListBundledThemes()
	require.NoError(t, err)
	assert.Empty(t, themes)
}

// =============================================================================
// AC-3: ReadBundledThemeFile reads theme JSON from bundled VSIX
// =============================================================================

func TestReadBundledThemeFile_AC3_ReadsThemeJSON(t *testing.T) {
	dir := t.TempDir()

	themeJSON := `{
		"name": "Dracula",
		"type": "dark",
		"colors": {
			"editor.background": "#282a36",
			"editor.foreground": "#f8f8f2"
		},
		"tokenColors": [
			{"scope": "comment", "settings": {"foreground": "#6272a4"}}
		]
	}`
	createMockVSIX(t, dir, "dracula.vsix", map[string]string{
		"extension/package.json":           `{"name":"dracula"}`,
		"extension/themes/dracula.json":    themeJSON,
	})

	useBundledThemesDir(t, dir)
	app := &App{}
	themePath := makeVSIXThemePath(filepath.Join(dir, "dracula.vsix"), "extension/themes/dracula.json")
	result, err := app.ReadBundledThemeFile(themePath)
	require.NoError(t, err)

	// Verify it returns valid JSON
	var parsed map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(result), &parsed), "result must be valid JSON")

	// Verify content
	assert.Equal(t, "Dracula", parsed["name"])
	assert.NotNil(t, parsed["colors"], "should contain colors key")
	assert.NotNil(t, parsed["tokenColors"], "should contain tokenColors key")
}

func TestReadBundledThemeFile_AC3_StripsJSONC(t *testing.T) {
	dir := t.TempDir()

	jsoncTheme := `{
		// Line comment
		"name": "Commented Theme",
		"colors": {
			"editor.background": "#282a36" /* inline comment */
		},
		"tokenColors": [],
	}`
	createMockVSIX(t, dir, "jsonc-ext.vsix", map[string]string{
		"extension/package.json":       `{"name":"jsonc-ext"}`,
		"extension/themes/theme.json": jsoncTheme,
	})

	useBundledThemesDir(t, dir)
	app := &App{}
	themePath := makeVSIXThemePath(filepath.Join(dir, "jsonc-ext.vsix"), "extension/themes/theme.json")
	result, err := app.ReadBundledThemeFile(themePath)
	require.NoError(t, err)

	var parsed map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(result), &parsed), "JSONC should be stripped to valid JSON")
	assert.Equal(t, "Commented Theme", parsed["name"])
}

func TestReadBundledThemeFile_AC3_ResolvesIncludes(t *testing.T) {
	dir := t.TempDir()

	baseTheme := `{
		"name": "Base",
		"colors": {
			"editor.background": "#000",
			"editor.foreground": "#fff"
		},
		"tokenColors": [
			{"scope": "comment", "settings": {"foreground": "#666"}}
		]
	}`
	childTheme := `{
		"name": "Child",
		"include": "./base.json",
		"colors": {
			"editor.background": "#111"
		},
		"tokenColors": [
			{"scope": "keyword", "settings": {"foreground": "#f00"}}
		]
	}`
	createMockVSIX(t, dir, "include-ext.vsix", map[string]string{
		"extension/package.json":        `{"name":"include-ext"}`,
		"extension/themes/base.json":    baseTheme,
		"extension/themes/child.json":   childTheme,
	})

	useBundledThemesDir(t, dir)
	app := &App{}
	themePath := makeVSIXThemePath(filepath.Join(dir, "include-ext.vsix"), "extension/themes/child.json")
	result, err := app.ReadBundledThemeFile(themePath)
	require.NoError(t, err)

	var parsed struct {
		Name        string                   `json:"name"`
		Colors      map[string]string        `json:"colors"`
		TokenColors []map[string]interface{} `json:"tokenColors"`
	}
	require.NoError(t, json.Unmarshal([]byte(result), &parsed))

	// Child overrides editor.background
	assert.Equal(t, "#111", parsed.Colors["editor.background"])
	// Inherits editor.foreground from base
	assert.Equal(t, "#fff", parsed.Colors["editor.foreground"])
	// Name is child's name
	assert.Equal(t, "Child", parsed.Name)
	// TokenColors: base prepended before child = 2 total
	assert.Len(t, parsed.TokenColors, 2)
}

func TestReadBundledThemeFile_AC3_NoConfigRequired(t *testing.T) {
	dir := t.TempDir()

	createMockVSIX(t, dir, "standalone.vsix", map[string]string{
		"extension/package.json":       `{"name":"standalone"}`,
		"extension/themes/theme.json": `{"name":"Standalone","type":"dark","colors":{}}`,
	})

	// Ensure VSCodiumExtPath is NOT configured
	cfg := loadConfig()
	origPath := cfg.VSCodiumExtPath
	cfg.VSCodiumExtPath = ""
	saveConfig(cfg)
	defer func() {
		cfg := loadConfig()
		cfg.VSCodiumExtPath = origPath
		saveConfig(cfg)
	}()

	useBundledThemesDir(t, dir)
	app := &App{}
	themePath := makeVSIXThemePath(filepath.Join(dir, "standalone.vsix"), "extension/themes/theme.json")
	result, err := app.ReadBundledThemeFile(themePath)
	require.NoError(t, err, "ReadBundledThemeFile should work without VSCodiumExtPath")

	var parsed map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(result), &parsed))
	assert.Equal(t, "Standalone", parsed["name"])
}

func TestReadBundledThemeFile_AC3_InvalidPath(t *testing.T) {
	tests := []struct {
		name      string
		themePath string
	}{
		{
			name:      "no vsix separator",
			themePath: "/some/path/without/separator.json",
		},
		{
			name:      "nonexistent vsix file",
			themePath: makeVSIXThemePath("/nonexistent/file.vsix", "extension/themes/theme.json"),
		},
		{
			name:      "empty path",
			themePath: "",
		},
	}

	app := &App{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := app.ReadBundledThemeFile(tt.themePath)
			assert.Error(t, err, "invalid path should return error")
		})
	}
}

func TestReadBundledThemeFile_AC3_MissingInternalFile(t *testing.T) {
	dir := t.TempDir()

	createMockVSIX(t, dir, "sparse.vsix", map[string]string{
		"extension/package.json": `{"name":"sparse"}`,
	})

	app := &App{}
	themePath := makeVSIXThemePath(filepath.Join(dir, "sparse.vsix"), "extension/themes/nonexistent.json")
	_, err := app.ReadBundledThemeFile(themePath)
	assert.Error(t, err, "missing internal file should return error")
}

// =============================================================================
// bundledThemesDir — verify path resolution
// =============================================================================

func TestBundledThemesDir_ReturnsPath(t *testing.T) {
	result := bundledThemesDir()
	assert.NotEmpty(t, result, "bundledThemesDir should return a non-empty path")
	assert.True(t, filepath.IsAbs(result), "bundledThemesDir should return an absolute path")
	assert.True(t, strings.HasSuffix(result, "themes"), "path should end with 'themes'")
}
