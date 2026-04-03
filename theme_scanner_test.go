package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// --- Helper: create a mock VSCodium extension directory ---

// mockExtension creates a single extension directory with a package.json
// containing the given theme contributions.
type mockTheme struct {
	Label   string
	Path    string
	UITheme string
}

func createMockExtension(t *testing.T, extDir, extName string, themes []mockTheme) {
	t.Helper()
	dir := filepath.Join(extDir, extName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}

	// Build contributes.themes array
	type themeEntry struct {
		Label   string `json:"label"`
		UITheme string `json:"uiTheme"`
		Path    string `json:"path"`
	}
	var entries []themeEntry
	for _, th := range themes {
		entries = append(entries, themeEntry{
			Label:   th.Label,
			UITheme: th.UITheme,
			Path:    th.Path,
		})
	}

	pkg := map[string]interface{}{
		"name": extName,
		"contributes": map[string]interface{}{
			"themes": entries,
		},
	}

	data, err := json.MarshalIndent(pkg, "", "  ")
	if err != nil {
		t.Fatalf("marshal package.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "package.json"), data, 0644); err != nil {
		t.Fatalf("write package.json: %v", err)
	}

	// Create the theme JSON files
	for _, th := range themes {
		thPath := filepath.Join(dir, th.Path)
		if err := os.MkdirAll(filepath.Dir(thPath), 0755); err != nil {
			t.Fatalf("mkdir theme dir: %v", err)
		}
		themeContent := fmt.Sprintf(`{"name": %q, "type": "dark", "colors": {}}`, th.Label)
		if err := os.WriteFile(thPath, []byte(themeContent), 0644); err != nil {
			t.Fatalf("write theme file: %v", err)
		}
	}
}

// --- Helper: create a mock .vsix (zip archive) ---

// createMockVSIX creates a .vsix zip file in dir with the given name.
// files maps zip-internal paths (e.g. "extension/package.json") to their content.
func createMockVSIX(t *testing.T, dir, name string, files map[string]string) {
	t.Helper()
	vsixPath := filepath.Join(dir, name)
	f, err := os.Create(vsixPath)
	if err != nil {
		t.Fatalf("create vsix %s: %v", vsixPath, err)
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	for path, content := range files {
		w, err := zw.Create(path)
		if err != nil {
			t.Fatalf("create zip entry %s: %v", path, err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			t.Fatalf("write zip entry %s: %v", path, err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip writer: %v", err)
	}
}

// --- Test: stripJSONC ---

func TestStripJSONC(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "line comment",
			input: "{\n  // comment\n  \"name\": \"test\"\n}",
			want:  "{\n  \n  \"name\": \"test\"\n}",
		},
		{
			name:  "block comment",
			input: "{\n  /* block\n  comment */\n  \"name\": \"test\"\n}",
			want:  "{\n  \n  \"name\": \"test\"\n}",
		},
		{
			name:  "inline block comment",
			input: `{"name": "test" /* comment */, "type": "dark"}`,
			want:  `{"name": "test" , "type": "dark"}`,
		},
		{
			name:  "comment-like inside string preserved",
			input: `{"url": "https://example.com/path"}`,
			want:  `{"url": "https://example.com/path"}`,
		},
		{
			name:  "double-slash in string preserved",
			input: `{"desc": "use // for comments"}`,
			want:  `{"desc": "use // for comments"}`,
		},
		{
			name:  "block comment markers in string preserved",
			input: `{"desc": "use /* and */ in text"}`,
			want:  `{"desc": "use /* and */ in text"}`,
		},
		{
			name:  "trailing comma before closing brace",
			input: `{"a": 1, "b": 2,}`,
			want:  `{"a": 1, "b": 2}`,
		},
		{
			name:  "trailing comma before closing bracket",
			input: `["a", "b",]`,
			want:  `["a", "b"]`,
		},
		{
			name: "trailing comma with whitespace",
			input: `{
  "a": 1,
  "b": 2  ,
}`,
			want: `{
  "a": 1,
  "b": 2
}`,
		},
		{
			name:  "escaped quote in string",
			input: `{"val": "he said \"hello\" // not a comment"}`,
			want:  `{"val": "he said \"hello\" // not a comment"}`,
		},
		{
			name:  "no comments or trailing commas",
			input: `{"name": "clean"}`,
			want:  `{"name": "clean"}`,
		},
		{
			name:  "multiple line comments",
			input: "// first\n// second\n{\"a\": 1}",
			want:  "\n\n{\"a\": 1}",
		},
		{
			name:  "comment after value on same line",
			input: `{"a": 1 // inline}`,
			want:  `{"a": 1 `,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(stripJSONC([]byte(tt.input)))
			if got != tt.want {
				t.Errorf("stripJSONC(%q)\n  got:  %q\n  want: %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestStripJSONC_ResultIsValidJSON(t *testing.T) {
	input := `{
  // This is a comment
  "name": "Test Theme",
  "colors": {
    "editor.background": "#282a36" /* inline comment */
  },
  "tokenColors": [
    {
      "scope": "comment",
      "settings": {
        "foreground": "#6272a4",
      },
    },
  ],
}`
	got := stripJSONC([]byte(input))
	var parsed map[string]interface{}
	if err := json.Unmarshal(got, &parsed); err != nil {
		t.Errorf("stripJSONC output is not valid JSON: %v\nOutput: %s", err, string(got))
	}
	if parsed["name"] != "Test Theme" {
		t.Errorf("name = %v, want %q", parsed["name"], "Test Theme")
	}
}

// --- Test: ListVSCodiumThemes ---

func TestListVSCodiumThemes_HappyPath(t *testing.T) {
	extDir := t.TempDir()

	createMockVSIX(t, extDir, "dracula-theme.theme-dracula-2.24.3.vsix", map[string]string{
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
	createMockVSIX(t, extDir, "github.github-vscode-theme-6.3.5.vsix", map[string]string{
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

	app := &App{}
	cfg := loadConfig()
	cfg.VSCodiumExtPath = extDir
	if err := saveConfig(cfg); err != nil {
		t.Fatalf("save config: %v", err)
	}
	defer func() {
		cfg := loadConfig()
		cfg.VSCodiumExtPath = ""
		_ = saveConfig(cfg)
	}()

	themes, err := app.ListVSCodiumThemes()
	if err != nil {
		t.Fatalf("ListVSCodiumThemes error: %v", err)
	}

	if len(themes) != 3 {
		t.Fatalf("got %d themes, want 3", len(themes))
	}

	// Verify sorted alphabetically by label
	wantLabels := []string{"Dracula", "GitHub Dark", "GitHub Light"}
	for i, want := range wantLabels {
		if themes[i].Label != want {
			t.Errorf("themes[%d].Label = %q, want %q", i, themes[i].Label, want)
		}
	}

	// Verify ThemePath contains ::vsix:: separator
	for _, th := range themes {
		if !strings.Contains(th.ThemePath, vsixSeparator) {
			t.Errorf("ThemePath %q does not contain %q", th.ThemePath, vsixSeparator)
		}
	}

	// Verify ExtensionID is filename minus .vsix extension
	if themes[0].ExtensionID != "dracula-theme.theme-dracula-2.24.3" {
		t.Errorf("ExtensionID = %q, want %q", themes[0].ExtensionID, "dracula-theme.theme-dracula-2.24.3")
	}
}

func TestListVSCodiumThemes_FiltersTmTheme(t *testing.T) {
	extDir := t.TempDir()

	// Single vsix with both a .tmTheme path and a .json path
	createMockVSIX(t, extDir, "mixed-ext.vsix", map[string]string{
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

	cfg := loadConfig()
	cfg.VSCodiumExtPath = extDir
	if err := saveConfig(cfg); err != nil {
		t.Fatalf("save config: %v", err)
	}
	defer func() {
		cfg := loadConfig()
		cfg.VSCodiumExtPath = ""
		_ = saveConfig(cfg)
	}()

	app := &App{}
	themes, err := app.ListVSCodiumThemes()
	if err != nil {
		t.Fatalf("ListVSCodiumThemes error: %v", err)
	}

	for _, th := range themes {
		if strings.Contains(strings.ToLower(th.Label), "tm theme") {
			t.Errorf("tmTheme entry should have been filtered: %+v", th)
		}
	}
	if len(themes) != 1 {
		t.Errorf("got %d themes, want 1 (only JSON)", len(themes))
	}
}

func TestListVSCodiumThemes_NotConfigured(t *testing.T) {
	cfg := loadConfig()
	origPath := cfg.VSCodiumExtPath
	cfg.VSCodiumExtPath = ""
	if err := saveConfig(cfg); err != nil {
		t.Fatalf("save config: %v", err)
	}
	defer func() {
		cfg := loadConfig()
		cfg.VSCodiumExtPath = origPath
		_ = saveConfig(cfg)
	}()

	app := &App{}
	_, err := app.ListVSCodiumThemes()
	if err == nil {
		t.Fatal("expected error for unconfigured path")
	}
	if !strings.Contains(err.Error(), "not configured") {
		t.Errorf("error = %q, want to contain 'not configured'", err.Error())
	}
}

func TestListVSCodiumThemes_CorruptPackageJSON(t *testing.T) {
	extDir := t.TempDir()

	// Create a valid vsix
	createMockVSIX(t, extDir, "good-ext.vsix", map[string]string{
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

	// Create a corrupt vsix (valid zip, but package.json is invalid JSON)
	createMockVSIX(t, extDir, "bad-ext.vsix", map[string]string{
		"extension/package.json": `not json{{{`,
	})

	cfg := loadConfig()
	cfg.VSCodiumExtPath = extDir
	saveConfig(cfg)
	defer func() {
		cfg := loadConfig()
		cfg.VSCodiumExtPath = ""
		saveConfig(cfg)
	}()

	app := &App{}
	themes, err := app.ListVSCodiumThemes()
	if err != nil {
		t.Fatalf("should not error on corrupt extension: %v", err)
	}
	// Should still return the good extension's theme
	if len(themes) != 1 {
		t.Errorf("got %d themes, want 1 (corrupt ext skipped)", len(themes))
	}
}

// --- Test: ReadThemeFile ---

func TestReadThemeFile_HappyPath(t *testing.T) {
	extDir := t.TempDir()

	createMockExtension(t, extDir, "test-ext", []mockTheme{
		{Label: "Test", Path: "theme.json", UITheme: "vs-dark"},
	})

	cfg := loadConfig()
	cfg.VSCodiumExtPath = extDir
	saveConfig(cfg)
	defer func() {
		cfg := loadConfig()
		cfg.VSCodiumExtPath = ""
		saveConfig(cfg)
	}()

	themeFile := filepath.Join(extDir, "test-ext", "theme.json")
	app := &App{}
	result, err := app.ReadThemeFile(themeFile)
	if err != nil {
		t.Fatalf("ReadThemeFile error: %v", err)
	}

	// Verify it's valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("result is not valid JSON: %v\n%s", err, result)
	}
	if parsed["name"] != "Test" {
		t.Errorf("name = %v, want %q", parsed["name"], "Test")
	}
}

func TestReadThemeFile_JSONCStripped(t *testing.T) {
	extDir := t.TempDir()
	themeDir := filepath.Join(extDir, "test-ext")
	os.MkdirAll(themeDir, 0755)

	jsonc := `{
  // This is a comment
  "name": "Test Theme",
  "colors": {
    "editor.background": "#282a36" /* inline comment */
  }
}`
	themeFile := filepath.Join(themeDir, "theme.json")
	os.WriteFile(themeFile, []byte(jsonc), 0644)

	cfg := loadConfig()
	cfg.VSCodiumExtPath = extDir
	saveConfig(cfg)
	defer func() {
		cfg := loadConfig()
		cfg.VSCodiumExtPath = ""
		saveConfig(cfg)
	}()

	app := &App{}
	result, err := app.ReadThemeFile(themeFile)
	if err != nil {
		t.Fatalf("ReadThemeFile error: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("result should be valid JSON after JSONC stripping: %v\n%s", err, result)
	}
	if parsed["name"] != "Test Theme" {
		t.Errorf("name = %v, want %q", parsed["name"], "Test Theme")
	}
}

func TestReadThemeFile_PathTraversal(t *testing.T) {
	// Create a parent dir that holds both the ext dir and the "outside" file
	parentDir := t.TempDir()
	extDir := filepath.Join(parentDir, "extensions")
	os.MkdirAll(extDir, 0755)

	// Create a file outside the ext dir but in a resolvable location
	outsideFile := filepath.Join(parentDir, "secret.json")
	os.WriteFile(outsideFile, []byte(`{"secret": true}`), 0644)

	// Also create a valid ext dir structure
	themeDir := filepath.Join(extDir, "test-ext")
	os.MkdirAll(themeDir, 0755)
	os.WriteFile(filepath.Join(themeDir, "theme.json"), []byte(`{"name":"ok"}`), 0644)

	cfg := loadConfig()
	cfg.VSCodiumExtPath = extDir
	saveConfig(cfg)
	defer func() {
		cfg := loadConfig()
		cfg.VSCodiumExtPath = ""
		saveConfig(cfg)
	}()

	app := &App{}

	tests := []struct {
		name string
		path string
	}{
		{
			name: "relative traversal",
			path: filepath.Join(extDir, "test-ext", "..", "..", "secret.json"),
		},
		{
			name: "absolute outside path",
			path: outsideFile,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := app.ReadThemeFile(tt.path)
			if err == nil {
				t.Fatal("expected error for path traversal")
			}
			if !strings.Contains(err.Error(), "outside extensions directory") {
				t.Errorf("error = %q, want to contain 'outside extensions directory'", err.Error())
			}
		})
	}
}

func TestReadThemeFile_SymlinkOutside(t *testing.T) {
	extDir := t.TempDir()
	themeDir := filepath.Join(extDir, "evil-ext")
	os.MkdirAll(themeDir, 0755)

	// Create a file outside
	outsideDir := t.TempDir()
	outsideFile := filepath.Join(outsideDir, "passwd.json")
	os.WriteFile(outsideFile, []byte(`{"leaked": true}`), 0644)

	// Create symlink inside ext dir pointing outside
	symlinkPath := filepath.Join(themeDir, "theme.json")
	if err := os.Symlink(outsideFile, symlinkPath); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}

	cfg := loadConfig()
	cfg.VSCodiumExtPath = extDir
	saveConfig(cfg)
	defer func() {
		cfg := loadConfig()
		cfg.VSCodiumExtPath = ""
		saveConfig(cfg)
	}()

	app := &App{}
	_, err := app.ReadThemeFile(symlinkPath)
	if err == nil {
		t.Fatal("expected error for symlink outside extensions directory")
	}
	if !strings.Contains(err.Error(), "outside extensions directory") {
		t.Errorf("error = %q, want to contain 'outside extensions directory'", err.Error())
	}
}

func TestReadThemeFile_PrefixConfusion(t *testing.T) {
	// Ensure /ext-other/theme.json doesn't match /ext/ prefix
	parentDir := t.TempDir()
	extDir := filepath.Join(parentDir, "ext")
	extOtherDir := filepath.Join(parentDir, "ext-other")
	os.MkdirAll(extDir, 0755)
	os.MkdirAll(extOtherDir, 0755)

	// Put a theme file in ext-other
	os.WriteFile(filepath.Join(extOtherDir, "theme.json"), []byte(`{"name":"evil"}`), 0644)

	cfg := loadConfig()
	cfg.VSCodiumExtPath = extDir
	saveConfig(cfg)
	defer func() {
		cfg := loadConfig()
		cfg.VSCodiumExtPath = ""
		saveConfig(cfg)
	}()

	app := &App{}
	_, err := app.ReadThemeFile(filepath.Join(extOtherDir, "theme.json"))
	if err == nil {
		t.Fatal("expected error for prefix confusion path")
	}
	if !strings.Contains(err.Error(), "outside extensions directory") {
		t.Errorf("error = %q, want to contain 'outside extensions directory'", err.Error())
	}
}

func TestReadThemeFile_SizeLimit(t *testing.T) {
	extDir := t.TempDir()
	themeDir := filepath.Join(extDir, "big-ext")
	os.MkdirAll(themeDir, 0755)

	// Create a file larger than 512KB
	bigData := make([]byte, 600*1024)
	for i := range bigData {
		bigData[i] = ' '
	}
	bigFile := filepath.Join(themeDir, "big.json")
	os.WriteFile(bigFile, bigData, 0644)

	cfg := loadConfig()
	cfg.VSCodiumExtPath = extDir
	saveConfig(cfg)
	defer func() {
		cfg := loadConfig()
		cfg.VSCodiumExtPath = ""
		saveConfig(cfg)
	}()

	app := &App{}
	_, err := app.ReadThemeFile(bigFile)
	if err == nil {
		t.Fatal("expected error for oversized file")
	}
	if !strings.Contains(err.Error(), "too large") {
		t.Errorf("error = %q, want to contain 'too large'", err.Error())
	}
}

// --- Test: Include resolution ---

func TestReadThemeFile_IncludeResolution(t *testing.T) {
	extDir := t.TempDir()
	themeDir := filepath.Join(extDir, "include-ext")
	os.MkdirAll(themeDir, 0755)

	// Base theme
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
	os.WriteFile(filepath.Join(themeDir, "base.json"), []byte(baseTheme), 0644)

	// Child theme includes base, overrides one color
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
	os.WriteFile(filepath.Join(themeDir, "child.json"), []byte(childTheme), 0644)

	cfg := loadConfig()
	cfg.VSCodiumExtPath = extDir
	saveConfig(cfg)
	defer func() {
		cfg := loadConfig()
		cfg.VSCodiumExtPath = ""
		saveConfig(cfg)
	}()

	app := &App{}
	result, err := app.ReadThemeFile(filepath.Join(themeDir, "child.json"))
	if err != nil {
		t.Fatalf("ReadThemeFile error: %v", err)
	}

	var parsed struct {
		Name        string                 `json:"name"`
		Colors      map[string]string      `json:"colors"`
		TokenColors []map[string]interface{} `json:"tokenColors"`
	}
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("result not valid JSON: %v\n%s", err, result)
	}

	// Child overrides editor.background
	if parsed.Colors["editor.background"] != "#111" {
		t.Errorf("editor.background = %q, want #111", parsed.Colors["editor.background"])
	}
	// Inherits editor.foreground from base
	if parsed.Colors["editor.foreground"] != "#fff" {
		t.Errorf("editor.foreground = %q, want #fff", parsed.Colors["editor.foreground"])
	}
	// Name should be child's name
	if parsed.Name != "Child" {
		t.Errorf("name = %q, want Child", parsed.Name)
	}
	// TokenColors: base prepended before child
	if len(parsed.TokenColors) != 2 {
		t.Fatalf("tokenColors length = %d, want 2", len(parsed.TokenColors))
	}
}

func TestReadThemeFile_IncludeMultiLevel(t *testing.T) {
	extDir := t.TempDir()
	themeDir := filepath.Join(extDir, "multi-ext")
	os.MkdirAll(themeDir, 0755)

	// grandparent -> parent -> child
	grandparent := `{"name": "GP", "colors": {"a": "1", "b": "2", "c": "3"}}`
	parent := `{"name": "P", "include": "./grandparent.json", "colors": {"b": "22"}}`
	child := `{"name": "C", "include": "./parent.json", "colors": {"c": "333"}}`

	os.WriteFile(filepath.Join(themeDir, "grandparent.json"), []byte(grandparent), 0644)
	os.WriteFile(filepath.Join(themeDir, "parent.json"), []byte(parent), 0644)
	os.WriteFile(filepath.Join(themeDir, "child.json"), []byte(child), 0644)

	cfg := loadConfig()
	cfg.VSCodiumExtPath = extDir
	saveConfig(cfg)
	defer func() {
		cfg := loadConfig()
		cfg.VSCodiumExtPath = ""
		saveConfig(cfg)
	}()

	app := &App{}
	result, err := app.ReadThemeFile(filepath.Join(themeDir, "child.json"))
	if err != nil {
		t.Fatalf("ReadThemeFile error: %v", err)
	}

	var parsed struct {
		Colors map[string]string `json:"colors"`
	}
	json.Unmarshal([]byte(result), &parsed)

	if parsed.Colors["a"] != "1" {
		t.Errorf("a = %q, want 1 (from grandparent)", parsed.Colors["a"])
	}
	if parsed.Colors["b"] != "22" {
		t.Errorf("b = %q, want 22 (from parent override)", parsed.Colors["b"])
	}
	if parsed.Colors["c"] != "333" {
		t.Errorf("c = %q, want 333 (from child override)", parsed.Colors["c"])
	}
}

func TestReadThemeFile_IncludeDepthLimit(t *testing.T) {
	extDir := t.TempDir()
	themeDir := filepath.Join(extDir, "deep-ext")
	os.MkdirAll(themeDir, 0755)

	// Create a chain of 7 includes (exceeds depth limit of 5)
	for i := 0; i < 7; i++ {
		name := fmt.Sprintf("level%d.json", i)
		var content string
		if i < 6 {
			next := fmt.Sprintf("level%d.json", i+1)
			content = fmt.Sprintf(`{"name": "level%d", "include": "./%s", "colors": {"l%d": "v%d"}}`, i, next, i, i)
		} else {
			content = fmt.Sprintf(`{"name": "level%d", "colors": {"l%d": "v%d"}}`, i, i, i)
		}
		os.WriteFile(filepath.Join(themeDir, name), []byte(content), 0644)
	}

	cfg := loadConfig()
	cfg.VSCodiumExtPath = extDir
	saveConfig(cfg)
	defer func() {
		cfg := loadConfig()
		cfg.VSCodiumExtPath = ""
		saveConfig(cfg)
	}()

	app := &App{}
	// Reading level0 should work but stop resolving at depth 5
	result, err := app.ReadThemeFile(filepath.Join(themeDir, "level0.json"))
	if err != nil {
		t.Fatalf("ReadThemeFile error: %v", err)
	}

	var parsed struct {
		Colors map[string]string `json:"colors"`
	}
	json.Unmarshal([]byte(result), &parsed)

	// level0 color should be present (it's the top level)
	if parsed.Colors["l0"] != "v0" {
		t.Errorf("l0 = %q, want v0", parsed.Colors["l0"])
	}
	// Levels beyond the depth cap should NOT be resolved
	// With depth limit 5, starting from level0 (depth 0), we can resolve up to level4 (depth 4)
	// level5 can be read (depth 5), but level5's include of level6 would be depth 6, which is skipped
	if parsed.Colors["l5"] != "v5" {
		t.Errorf("l5 should be present (depth 5 is the last resolved)")
	}
}

// --- Test: ReadThemeFile from VSIX ---

func TestReadThemeFile_VSIX_HappyPath(t *testing.T) {
	extDir := t.TempDir()

	createMockVSIX(t, extDir, "test-ext.vsix", map[string]string{
		"extension/package.json":      `{"name":"test-ext","contributes":{"themes":[{"label":"Test","uiTheme":"vs-dark","path":"./themes/test.json"}]}}`,
		"extension/themes/test.json": `{"name":"Test Theme","type":"dark","colors":{"editor.background":"#000"}}`,
	})

	cfg := loadConfig()
	cfg.VSCodiumExtPath = extDir
	saveConfig(cfg)
	defer func() {
		cfg := loadConfig()
		cfg.VSCodiumExtPath = ""
		saveConfig(cfg)
	}()

	app := &App{}
	themePath := makeVSIXThemePath(filepath.Join(extDir, "test-ext.vsix"), "extension/themes/test.json")
	result, err := app.ReadThemeFile(themePath)
	if err != nil {
		t.Fatalf("ReadThemeFile error: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("result is not valid JSON: %v\n%s", err, result)
	}
	if parsed["name"] != "Test Theme" {
		t.Errorf("name = %v, want %q", parsed["name"], "Test Theme")
	}
}

func TestReadThemeFile_VSIX_JSONCStripped(t *testing.T) {
	extDir := t.TempDir()

	jsoncContent := `{
  // This is a comment
  "name": "Commented Theme",
  "colors": {
    "editor.background": "#282a36" /* inline comment */
  }
}`
	createMockVSIX(t, extDir, "jsonc-ext.vsix", map[string]string{
		"extension/package.json":       `{"name":"jsonc-ext"}`,
		"extension/themes/theme.json": jsoncContent,
	})

	cfg := loadConfig()
	cfg.VSCodiumExtPath = extDir
	saveConfig(cfg)
	defer func() {
		cfg := loadConfig()
		cfg.VSCodiumExtPath = ""
		saveConfig(cfg)
	}()

	app := &App{}
	themePath := makeVSIXThemePath(filepath.Join(extDir, "jsonc-ext.vsix"), "extension/themes/theme.json")
	result, err := app.ReadThemeFile(themePath)
	if err != nil {
		t.Fatalf("ReadThemeFile error: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("result should be valid JSON after JSONC stripping: %v\n%s", err, result)
	}
	if parsed["name"] != "Commented Theme" {
		t.Errorf("name = %v, want %q", parsed["name"], "Commented Theme")
	}
}

func TestReadThemeFile_VSIX_IncludeResolution(t *testing.T) {
	extDir := t.TempDir()

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
	createMockVSIX(t, extDir, "include-ext.vsix", map[string]string{
		"extension/package.json":     `{"name":"include-ext"}`,
		"extension/themes/base.json":  baseTheme,
		"extension/themes/child.json": childTheme,
	})

	cfg := loadConfig()
	cfg.VSCodiumExtPath = extDir
	saveConfig(cfg)
	defer func() {
		cfg := loadConfig()
		cfg.VSCodiumExtPath = ""
		saveConfig(cfg)
	}()

	app := &App{}
	themePath := makeVSIXThemePath(filepath.Join(extDir, "include-ext.vsix"), "extension/themes/child.json")
	result, err := app.ReadThemeFile(themePath)
	if err != nil {
		t.Fatalf("ReadThemeFile error: %v", err)
	}

	var parsed struct {
		Name        string                   `json:"name"`
		Colors      map[string]string        `json:"colors"`
		TokenColors []map[string]interface{} `json:"tokenColors"`
	}
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("result not valid JSON: %v\n%s", err, result)
	}

	// Child overrides editor.background
	if parsed.Colors["editor.background"] != "#111" {
		t.Errorf("editor.background = %q, want #111", parsed.Colors["editor.background"])
	}
	// Inherits editor.foreground from base
	if parsed.Colors["editor.foreground"] != "#fff" {
		t.Errorf("editor.foreground = %q, want #fff", parsed.Colors["editor.foreground"])
	}
	// Name should be child's name
	if parsed.Name != "Child" {
		t.Errorf("name = %q, want Child", parsed.Name)
	}
	// TokenColors: base prepended before child
	if len(parsed.TokenColors) != 2 {
		t.Fatalf("tokenColors length = %d, want 2", len(parsed.TokenColors))
	}
}

func TestReadThemeFile_VSIX_IncludeMultiLevel(t *testing.T) {
	extDir := t.TempDir()

	grandparent := `{"name": "GP", "colors": {"a": "1", "b": "2", "c": "3"}}`
	parent := `{"name": "P", "include": "./grandparent.json", "colors": {"b": "22"}}`
	child := `{"name": "C", "include": "./parent.json", "colors": {"c": "333"}}`

	createMockVSIX(t, extDir, "multi-ext.vsix", map[string]string{
		"extension/package.json":            `{"name":"multi-ext"}`,
		"extension/themes/grandparent.json": grandparent,
		"extension/themes/parent.json":      parent,
		"extension/themes/child.json":       child,
	})

	cfg := loadConfig()
	cfg.VSCodiumExtPath = extDir
	saveConfig(cfg)
	defer func() {
		cfg := loadConfig()
		cfg.VSCodiumExtPath = ""
		saveConfig(cfg)
	}()

	app := &App{}
	themePath := makeVSIXThemePath(filepath.Join(extDir, "multi-ext.vsix"), "extension/themes/child.json")
	result, err := app.ReadThemeFile(themePath)
	if err != nil {
		t.Fatalf("ReadThemeFile error: %v", err)
	}

	var parsed struct {
		Colors map[string]string `json:"colors"`
	}
	json.Unmarshal([]byte(result), &parsed)

	if parsed.Colors["a"] != "1" {
		t.Errorf("a = %q, want 1 (from grandparent)", parsed.Colors["a"])
	}
	if parsed.Colors["b"] != "22" {
		t.Errorf("b = %q, want 22 (from parent override)", parsed.Colors["b"])
	}
	if parsed.Colors["c"] != "333" {
		t.Errorf("c = %q, want 333 (from child override)", parsed.Colors["c"])
	}
}

func TestReadThemeFile_VSIX_IncludeDepthLimit(t *testing.T) {
	extDir := t.TempDir()

	// Create a chain of 7 includes inside a single vsix (exceeds depth limit of 5)
	files := map[string]string{
		"extension/package.json": `{"name":"deep-ext"}`,
	}
	for i := 0; i < 7; i++ {
		name := fmt.Sprintf("extension/themes/level%d.json", i)
		var content string
		if i < 6 {
			next := fmt.Sprintf("./level%d.json", i+1)
			content = fmt.Sprintf(`{"name": "level%d", "include": %q, "colors": {"l%d": "v%d"}}`, i, next, i, i)
		} else {
			content = fmt.Sprintf(`{"name": "level%d", "colors": {"l%d": "v%d"}}`, i, i, i)
		}
		files[name] = content
	}
	createMockVSIX(t, extDir, "deep-ext.vsix", files)

	cfg := loadConfig()
	cfg.VSCodiumExtPath = extDir
	saveConfig(cfg)
	defer func() {
		cfg := loadConfig()
		cfg.VSCodiumExtPath = ""
		saveConfig(cfg)
	}()

	app := &App{}
	themePath := makeVSIXThemePath(filepath.Join(extDir, "deep-ext.vsix"), "extension/themes/level0.json")
	// Reading level0 should work but stop resolving at depth 5
	result, err := app.ReadThemeFile(themePath)
	if err != nil {
		t.Fatalf("ReadThemeFile error: %v", err)
	}

	var parsed struct {
		Colors map[string]string `json:"colors"`
	}
	json.Unmarshal([]byte(result), &parsed)

	// level0 color should be present (it's the top level)
	if parsed.Colors["l0"] != "v0" {
		t.Errorf("l0 = %q, want v0", parsed.Colors["l0"])
	}
	// level5 should be present (depth 5 is the last resolved)
	if parsed.Colors["l5"] != "v5" {
		t.Errorf("l5 should be present (depth 5 is the last resolved)")
	}
}

func TestReadThemeFile_VSIX_PathTraversal(t *testing.T) {
	// Create a vsix file outside the extensions directory
	outsideDir := t.TempDir()
	extDir := t.TempDir()

	createMockVSIX(t, outsideDir, "evil.vsix", map[string]string{
		"extension/package.json":     `{"name":"evil"}`,
		"extension/themes/evil.json": `{"name":"Evil","colors":{}}`,
	})

	cfg := loadConfig()
	cfg.VSCodiumExtPath = extDir
	saveConfig(cfg)
	defer func() {
		cfg := loadConfig()
		cfg.VSCodiumExtPath = ""
		saveConfig(cfg)
	}()

	app := &App{}
	themePath := makeVSIXThemePath(filepath.Join(outsideDir, "evil.vsix"), "extension/themes/evil.json")
	_, err := app.ReadThemeFile(themePath)
	if err == nil {
		t.Fatal("expected error for vsix path outside extensions directory")
	}
	if !strings.Contains(err.Error(), "outside extensions directory") {
		t.Errorf("error = %q, want to contain 'outside extensions directory'", err.Error())
	}
}

func TestReadThemeFile_VSIX_SizeLimit(t *testing.T) {
	extDir := t.TempDir()

	// Create a theme file inside vsix that's > 512KB
	bigContent := strings.Repeat(" ", 600*1024)
	createMockVSIX(t, extDir, "big-ext.vsix", map[string]string{
		"extension/package.json":      `{"name":"big-ext"}`,
		"extension/themes/big.json": bigContent,
	})

	cfg := loadConfig()
	cfg.VSCodiumExtPath = extDir
	saveConfig(cfg)
	defer func() {
		cfg := loadConfig()
		cfg.VSCodiumExtPath = ""
		saveConfig(cfg)
	}()

	app := &App{}
	themePath := makeVSIXThemePath(filepath.Join(extDir, "big-ext.vsix"), "extension/themes/big.json")
	_, err := app.ReadThemeFile(themePath)
	if err == nil {
		t.Fatal("expected error for oversized file inside vsix")
	}
	if !strings.Contains(err.Error(), "too large") {
		t.Errorf("error = %q, want to contain 'too large'", err.Error())
	}
}

func TestReadThemeFile_VSIX_FileNotFound(t *testing.T) {
	extDir := t.TempDir()

	createMockVSIX(t, extDir, "sparse-ext.vsix", map[string]string{
		"extension/package.json": `{"name":"sparse-ext"}`,
	})

	cfg := loadConfig()
	cfg.VSCodiumExtPath = extDir
	saveConfig(cfg)
	defer func() {
		cfg := loadConfig()
		cfg.VSCodiumExtPath = ""
		saveConfig(cfg)
	}()

	app := &App{}
	themePath := makeVSIXThemePath(filepath.Join(extDir, "sparse-ext.vsix"), "extension/themes/missing.json")
	_, err := app.ReadThemeFile(themePath)
	if err == nil {
		t.Fatal("expected error for missing file inside vsix")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %q, want to contain 'not found'", err.Error())
	}
}

func TestReadThemeFile_VSIX_BackwardCompat(t *testing.T) {
	// Verify that a regular filesystem path (no ::vsix::) still works
	// via the existing readThemeFileWithDepth code path
	extDir := t.TempDir()
	themeDir := filepath.Join(extDir, "compat-ext")
	os.MkdirAll(themeDir, 0755)

	themeContent := `{"name":"Compat Theme","type":"dark","colors":{"editor.background":"#000"}}`
	themeFile := filepath.Join(themeDir, "theme.json")
	os.WriteFile(themeFile, []byte(themeContent), 0644)

	cfg := loadConfig()
	cfg.VSCodiumExtPath = extDir
	saveConfig(cfg)
	defer func() {
		cfg := loadConfig()
		cfg.VSCodiumExtPath = ""
		saveConfig(cfg)
	}()

	app := &App{}
	result, err := app.ReadThemeFile(themeFile)
	if err != nil {
		t.Fatalf("ReadThemeFile error: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("result is not valid JSON: %v\n%s", err, result)
	}
	if parsed["name"] != "Compat Theme" {
		t.Errorf("name = %v, want %q", parsed["name"], "Compat Theme")
	}
}

// --- Test: Tilde expansion ---

func TestTildeExpansion(t *testing.T) {
	// We can't test full ListVSCodiumThemes with ~ since we can't mock os.UserHomeDir,
	// but we verify the expansion logic works by testing with a path that exists

	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot get home dir")
	}

	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name:   "tilde slash prefix",
			input:  "~/.vscode-oss/extensions",
			expect: filepath.Join(home, ".vscode-oss/extensions"),
		},
		{
			name:   "tilde alone",
			input:  "~",
			expect: home,
		},
		{
			name:   "tilde other user NOT expanded",
			input:  "~otheruser/path",
			expect: "~otheruser/path",
		},
		{
			name:   "no tilde",
			input:  "/absolute/path",
			expect: "/absolute/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandTilde(tt.input)
			if got != tt.expect {
				t.Errorf("expandTilde(%q) = %q, want %q", tt.input, got, tt.expect)
			}
		})
	}
}

// --- Test: SetImportedTheme ---

func TestSetImportedTheme(t *testing.T) {
	app := &App{}

	err := app.SetImportedTheme("/path/to/theme.json")
	if err != nil {
		t.Fatalf("SetImportedTheme error: %v", err)
	}

	cfg := loadConfig()
	if cfg.ImportedTheme != "/path/to/theme.json" {
		t.Errorf("ImportedTheme = %q, want %q", cfg.ImportedTheme, "/path/to/theme.json")
	}

	// Clean up
	cfg.ImportedTheme = ""
	saveConfig(cfg)
}

// --- Test: Mutex protection (concurrent writes) ---

func TestConcurrentConfigWrites(t *testing.T) {
	app := &App{}

	// Save a clean config first
	cfg := loadConfig()
	origTheme := cfg.Theme
	origImported := cfg.ImportedTheme
	defer func() {
		cfg := loadConfig()
		cfg.Theme = origTheme
		cfg.ImportedTheme = origImported
		saveConfig(cfg)
	}()

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		for i := 0; i < 20; i++ {
			app.SetTheme(fmt.Sprintf("theme-%d", i))
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 20; i++ {
			app.SetImportedTheme(fmt.Sprintf("/path/%d.json", i))
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 20; i++ {
			app.SetVSCodiumExtPath(fmt.Sprintf("/ext/%d", i))
		}
	}()

	wg.Wait()

	// After all goroutines complete, verify config isn't corrupted
	final := loadConfig()
	if final.Theme == "" {
		t.Error("Theme should not be empty after concurrent writes")
	}
	if final.ImportedTheme == "" {
		t.Error("ImportedTheme should not be empty after concurrent writes")
	}
	if final.VSCodiumExtPath == "" {
		t.Error("VSCodiumExtPath should not be empty after concurrent writes")
	}

	// Verify the config is valid JSON by re-loading
	data, err := os.ReadFile(configPath())
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	var check conductorConfig
	if err := json.Unmarshal(data, &check); err != nil {
		t.Fatalf("config is corrupt JSON after concurrent writes: %v", err)
	}
}

// --- Test: ReadThemeFile with ext path not configured ---

func TestReadThemeFile_ExtPathNotConfigured(t *testing.T) {
	cfg := loadConfig()
	origPath := cfg.VSCodiumExtPath
	cfg.VSCodiumExtPath = ""
	saveConfig(cfg)
	defer func() {
		cfg := loadConfig()
		cfg.VSCodiumExtPath = origPath
		saveConfig(cfg)
	}()

	app := &App{}
	_, err := app.ReadThemeFile("/some/path/theme.json")
	if err == nil {
		t.Fatal("expected error when ext path not configured")
	}
	if !strings.Contains(err.Error(), "not configured") {
		t.Errorf("error = %q, want to contain 'not configured'", err.Error())
	}
}

// --- Test: ListVSCodiumThemes with inaccessible directory ---

func TestListVSCodiumThemes_InaccessibleDirectory(t *testing.T) {
	cfg := loadConfig()
	cfg.VSCodiumExtPath = "/nonexistent/path/that/does/not/exist"
	saveConfig(cfg)
	defer func() {
		cfg := loadConfig()
		cfg.VSCodiumExtPath = ""
		saveConfig(cfg)
	}()

	app := &App{}
	_, err := app.ListVSCodiumThemes()
	if err == nil {
		t.Fatal("expected error for inaccessible directory")
	}
}

// --- Test: Extension ID from directory name ---

func TestListVSCodiumThemes_ExtensionID(t *testing.T) {
	extDir := t.TempDir()

	createMockVSIX(t, extDir, "publisher.extension-name-1.2.3.vsix", map[string]string{
		"extension/package.json": `{
			"name": "extension-name",
			"contributes": {
				"themes": [
					{"label": "My Theme", "uiTheme": "vs-dark", "path": "./theme.json"}
				]
			}
		}`,
		"extension/theme.json": `{"name":"My Theme","type":"dark","colors":{}}`,
	})

	cfg := loadConfig()
	cfg.VSCodiumExtPath = extDir
	saveConfig(cfg)
	defer func() {
		cfg := loadConfig()
		cfg.VSCodiumExtPath = ""
		saveConfig(cfg)
	}()

	app := &App{}
	themes, err := app.ListVSCodiumThemes()
	if err != nil {
		t.Fatalf("ListVSCodiumThemes error: %v", err)
	}
	if len(themes) != 1 {
		t.Fatalf("got %d themes, want 1", len(themes))
	}
	// ExtensionID should be the vsix filename minus the .vsix extension
	if themes[0].ExtensionID != "publisher.extension-name-1.2.3" {
		t.Errorf("ExtensionID = %q, want %q", themes[0].ExtensionID, "publisher.extension-name-1.2.3")
	}
}
