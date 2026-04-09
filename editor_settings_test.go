package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Minimal config JSON used by validation tests that only need a valid config file.
const minimalConfigJSON = `{"devDir":"/dev"}`

// setupTestConfig redirects HOME to a temp dir with optional seed config.
func setupTestConfig(t *testing.T, configJSON string) *App {
	t.Helper()
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	if configJSON != "" {
		cfgDir := filepath.Join(tmpDir, ".mashed")
		require.NoError(t, os.MkdirAll(cfgDir, 0755))
		require.NoError(t, os.WriteFile(
			filepath.Join(cfgDir, "config.json"),
			[]byte(configJSON),
			0644,
		))
	}
	return &App{}
}

func readRawConfig(t *testing.T) map[string]json.RawMessage {
	t.Helper()
	data, err := os.ReadFile(configPath())
	require.NoError(t, err)
	var raw map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(data, &raw))
	return raw
}

func expectedDefaults() EditorSettings {
	return EditorSettings{
		TabSize:                 2,
		InsertSpaces:            true,
		CursorStyle:             "line",
		CursorBlinking:          "blink",
		MinimapEnabled:          false,
		WordWrap:                "off",
		LineNumbers:             "on",
		RenderWhitespace:        "none",
		BracketPairColorization: true,
		RenderLineHighlight:     "line",
		FontLigatures:           false,
		ScrollBeyondLastLine:    false,
		SmoothScrolling:         false,
	}
}

// =============================================================================
// BDD Scenario 1: Default settings
// =============================================================================

// TestDefaultEditorSettings_AC1_AllFieldsPresent verifies all 13 defaults match spec.
func TestDefaultEditorSettings_AC1_AllFieldsPresent(t *testing.T) {
	app := &App{}
	assert.Equal(t, expectedDefaults(), app.DefaultEditorSettings())
}

func TestDefaultEditorSettings_ConsistentValues(t *testing.T) {
	app := &App{}
	first := app.DefaultEditorSettings()
	second := app.DefaultEditorSettings()
	assert.Equal(t, first, second, "consecutive calls must return identical defaults")
}

// =============================================================================
// BDD Scenario 1 + AC-2: GetEditorSettings returns defaults when nil
// =============================================================================

func TestGetEditorSettings_AC2_DefaultsWhenNil(t *testing.T) {
	app := setupTestConfig(t, `{"devDir":"/Users/x/dev","theme":"mashed-dark"}`)
	got := app.GetEditorSettings()
	expected := expectedDefaults()

	assert.Equal(t, expected, got, "should return defaults when editorSettings is absent")
}

func TestGetEditorSettings_AC2_NoConfigFile(t *testing.T) {
	app := setupTestConfig(t, "") // no config file created
	got := app.GetEditorSettings()
	expected := expectedDefaults()

	assert.Equal(t, expected, got, "should return defaults when config file is missing")
}

func TestGetEditorSettings_ExistingConfig(t *testing.T) {
	configJSON := `{
		"devDir": "/Users/x/dev",
		"editorSettings": {
			"tabSize": 4,
			"insertSpaces": false,
			"cursorStyle": "block",
			"cursorBlinking": "smooth",
			"minimapEnabled": true,
			"wordWrap": "on",
			"lineNumbers": "relative",
			"renderWhitespace": "all",
			"bracketPairColorization": false,
			"renderLineHighlight": "all",
			"fontLigatures": true,
			"scrollBeyondLastLine": true,
			"smoothScrolling": true
		}
	}`
	app := setupTestConfig(t, configJSON)
	got := app.GetEditorSettings()

	assert.Equal(t, 4, got.TabSize)
	assert.False(t, got.InsertSpaces)
	assert.Equal(t, "block", got.CursorStyle)
	assert.Equal(t, "smooth", got.CursorBlinking)
	assert.True(t, got.MinimapEnabled)
	assert.Equal(t, "on", got.WordWrap)
	assert.Equal(t, "relative", got.LineNumbers)
	assert.Equal(t, "all", got.RenderWhitespace)
	assert.False(t, got.BracketPairColorization)
	assert.Equal(t, "all", got.RenderLineHighlight)
	assert.True(t, got.FontLigatures)
	assert.True(t, got.ScrollBeyondLastLine)
	assert.True(t, got.SmoothScrolling)
}

// =============================================================================
// BDD Scenario 2: Persistence round-trip (AC-3)
// =============================================================================

func TestSetEditorSettings_AC3_ValidRoundTrip(t *testing.T) {
	app := setupTestConfig(t, `{"devDir":"/Users/x/dev"}`)

	input := expectedDefaults()
	input.TabSize = 4
	input.CursorStyle = "block"
	input.MinimapEnabled = true

	err := app.SetEditorSettings(input)
	require.NoError(t, err)

	got := app.GetEditorSettings()
	assert.Equal(t, 4, got.TabSize)
	assert.Equal(t, "block", got.CursorStyle)
	assert.True(t, got.MinimapEnabled)

	// Verify JSON file contains the key
	raw := readRawConfig(t)
	_, hasKey := raw["editorSettings"]
	assert.True(t, hasKey, "config file must contain editorSettings JSON key")
}

func TestSetEditorSettings_AC3_PreservesOtherConfig(t *testing.T) {
	app := setupTestConfig(t, `{"devDir":"/Users/x/dev","theme":"mashed-dark","monoFont":"JetBrains Mono"}`)

	input := expectedDefaults()
	input.TabSize = 4

	err := app.SetEditorSettings(input)
	require.NoError(t, err)

	raw := readRawConfig(t)
	// Check devDir preserved
	var devDir string
	require.NoError(t, json.Unmarshal(raw["devDir"], &devDir))
	assert.Equal(t, "/Users/x/dev", devDir)

	// Check theme preserved
	var theme string
	require.NoError(t, json.Unmarshal(raw["theme"], &theme))
	assert.Equal(t, "mashed-dark", theme)

	// Check monoFont preserved
	var monoFont string
	require.NoError(t, json.Unmarshal(raw["monoFont"], &monoFont))
	assert.Equal(t, "JetBrains Mono", monoFont)
}

// =============================================================================
// BDD Scenario 3: Validation rejects bad values (AC-4)
// =============================================================================

func TestSetEditorSettings_AC4_InvalidTabSize(t *testing.T) {
	tests := []struct {
		name    string
		tabSize int
	}{
		{"zero", 0},
		{"negative", -1},
		{"one_below_min", 1},
		{"above_max", 9},
		{"way_above_max", 99},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := setupTestConfig(t, minimalConfigJSON)
			settings := expectedDefaults()
			settings.TabSize = tt.tabSize

			err := app.SetEditorSettings(settings)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "tabSize",
				"error should mention the invalid field name")
		})
	}
}

func TestSetEditorSettings_AC4_ValidTabSizeBounds(t *testing.T) {
	tests := []struct {
		name    string
		tabSize int
	}{
		{"min_valid", 2},
		{"mid_valid", 4},
		{"max_valid", 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := setupTestConfig(t, minimalConfigJSON)
			settings := expectedDefaults()
			settings.TabSize = tt.tabSize

			err := app.SetEditorSettings(settings)
			assert.NoError(t, err)
		})
	}
}

func TestSetEditorSettings_AC4_InvalidCursorStyle(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{"blinky", "blinky"},
		{"empty", ""},
		{"uppercase", "LINE"},
		{"random", "cursor-fancy"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := setupTestConfig(t, minimalConfigJSON)
			settings := expectedDefaults()
			settings.CursorStyle = tt.value

			err := app.SetEditorSettings(settings)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "cursorStyle")
		})
	}
}

func TestSetEditorSettings_AC4_ValidCursorStyles(t *testing.T) {
	validStyles := []string{"line", "block", "underline", "line-thin", "block-outline", "underline-thin"}

	for _, style := range validStyles {
		t.Run(style, func(t *testing.T) {
			app := setupTestConfig(t, minimalConfigJSON)
			settings := expectedDefaults()
			settings.CursorStyle = style

			err := app.SetEditorSettings(settings)
			assert.NoError(t, err, "cursor style %q should be valid", style)
		})
	}
}

func TestSetEditorSettings_AC4_InvalidWordWrap(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{"maybe", "maybe"},
		{"empty", ""},
		{"true_string", "true"},
		{"wrap", "wrap"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := setupTestConfig(t, minimalConfigJSON)
			settings := expectedDefaults()
			settings.WordWrap = tt.value

			err := app.SetEditorSettings(settings)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "wordWrap")
		})
	}
}

func TestSetEditorSettings_AC4_ValidWordWrapValues(t *testing.T) {
	validValues := []string{"off", "on", "wordWrapColumn", "bounded"}

	for _, val := range validValues {
		t.Run(val, func(t *testing.T) {
			app := setupTestConfig(t, minimalConfigJSON)
			settings := expectedDefaults()
			settings.WordWrap = val

			err := app.SetEditorSettings(settings)
			assert.NoError(t, err, "wordWrap %q should be valid", val)
		})
	}
}

func TestSetEditorSettings_AC4_InvalidLineNumbers(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{"empty", ""},
		{"true_string", "true"},
		{"numbers", "numbers"},
		{"yes", "yes"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := setupTestConfig(t, minimalConfigJSON)
			settings := expectedDefaults()
			settings.LineNumbers = tt.value

			err := app.SetEditorSettings(settings)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "lineNumbers")
		})
	}
}

func TestSetEditorSettings_AC4_ValidLineNumbers(t *testing.T) {
	validValues := []string{"on", "off", "relative", "interval"}

	for _, val := range validValues {
		t.Run(val, func(t *testing.T) {
			app := setupTestConfig(t, minimalConfigJSON)
			settings := expectedDefaults()
			settings.LineNumbers = val

			err := app.SetEditorSettings(settings)
			assert.NoError(t, err, "lineNumbers %q should be valid", val)
		})
	}
}

func TestSetEditorSettings_AC4_InvalidRenderWhitespace(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{"empty", ""},
		{"visible", "visible"},
		{"true_string", "true"},
		{"spaces", "spaces"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := setupTestConfig(t, minimalConfigJSON)
			settings := expectedDefaults()
			settings.RenderWhitespace = tt.value

			err := app.SetEditorSettings(settings)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "renderWhitespace")
		})
	}
}

func TestSetEditorSettings_AC4_ValidRenderWhitespace(t *testing.T) {
	validValues := []string{"none", "boundary", "selection", "trailing", "all"}

	for _, val := range validValues {
		t.Run(val, func(t *testing.T) {
			app := setupTestConfig(t, minimalConfigJSON)
			settings := expectedDefaults()
			settings.RenderWhitespace = val

			err := app.SetEditorSettings(settings)
			assert.NoError(t, err, "renderWhitespace %q should be valid", val)
		})
	}
}

func TestSetEditorSettings_AC4_InvalidCursorBlinking(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{"empty", ""},
		{"fast", "fast"},
		{"on", "on"},
		{"true_string", "true"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := setupTestConfig(t, minimalConfigJSON)
			settings := expectedDefaults()
			settings.CursorBlinking = tt.value

			err := app.SetEditorSettings(settings)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "cursorBlinking")
		})
	}
}

func TestSetEditorSettings_AC4_ValidCursorBlinking(t *testing.T) {
	validValues := []string{"blink", "smooth", "phase", "expand", "solid"}

	for _, val := range validValues {
		t.Run(val, func(t *testing.T) {
			app := setupTestConfig(t, minimalConfigJSON)
			settings := expectedDefaults()
			settings.CursorBlinking = val

			err := app.SetEditorSettings(settings)
			assert.NoError(t, err, "cursorBlinking %q should be valid", val)
		})
	}
}

func TestSetEditorSettings_AC4_InvalidRenderLineHighlight(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{"empty", ""},
		{"highlight", "highlight"},
		{"true_string", "true"},
		{"yes", "yes"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := setupTestConfig(t, minimalConfigJSON)
			settings := expectedDefaults()
			settings.RenderLineHighlight = tt.value

			err := app.SetEditorSettings(settings)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "renderLineHighlight")
		})
	}
}

func TestSetEditorSettings_AC4_ValidRenderLineHighlight(t *testing.T) {
	validValues := []string{"none", "gutter", "line", "all"}

	for _, val := range validValues {
		t.Run(val, func(t *testing.T) {
			app := setupTestConfig(t, minimalConfigJSON)
			settings := expectedDefaults()
			settings.RenderLineHighlight = val

			err := app.SetEditorSettings(settings)
			assert.NoError(t, err, "renderLineHighlight %q should be valid", val)
		})
	}
}

// Verifies config file is not modified when validation fails.
func TestSetEditorSettings_AC4_ConfigUnchangedOnError(t *testing.T) {
	app := setupTestConfig(t, `{"devDir":"/Users/x/dev","theme":"mashed-dark"}`)

	// Read original config
	originalData, err := os.ReadFile(configPath())
	require.NoError(t, err)

	// Attempt invalid settings
	settings := expectedDefaults()
	settings.TabSize = 0 // invalid
	err = app.SetEditorSettings(settings)
	require.Error(t, err)

	// Config file should be unchanged
	afterData, err := os.ReadFile(configPath())
	require.NoError(t, err)
	assert.Equal(t, string(originalData), string(afterData),
		"config file must not be modified when validation fails")
}

// =============================================================================
// BDD Scenario 4: Concurrent access
// =============================================================================

// Run with: go test -race
func TestSetEditorSettings_ConcurrentAccess(t *testing.T) {
	app := setupTestConfig(t, minimalConfigJSON)

	var wg sync.WaitGroup
	const goroutines = 10

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			settings := expectedDefaults()
			settings.TabSize = 2 + (n % 7) // valid range 2-8
			_ = app.SetEditorSettings(settings)
		}(i)
	}
	wg.Wait()

	// After all goroutines, config should be valid JSON
	got := app.GetEditorSettings()
	assert.GreaterOrEqual(t, got.TabSize, 2)
	assert.LessOrEqual(t, got.TabSize, 8)
}

// =============================================================================
// AC-5: Backward compatibility with existing configs
// =============================================================================

func TestBackwardCompat_AC5_LegacyConfigNoEditorSettings(t *testing.T) {
	setupTestConfig(t, `{"devDir":"/Users/x/dev","theme":"mashed-dark"}`)

	cfg := loadConfig()
	assert.Nil(t, cfg.EditorSettings,
		"EditorSettings pointer must be nil when key is absent, not zero-value struct")
	// Other fields should be fine
	assert.Equal(t, "/Users/x/dev", cfg.DevDir)
	assert.Equal(t, "mashed-dark", cfg.Theme)
}

func TestBackwardCompat_AC5_ConfigWithEditorSettings(t *testing.T) {
	configJSON := `{
		"devDir": "/Users/x/dev",
		"editorSettings": {
			"tabSize": 4,
			"insertSpaces": true,
			"cursorStyle": "block",
			"cursorBlinking": "blink",
			"minimapEnabled": false,
			"wordWrap": "off",
			"lineNumbers": "on",
			"renderWhitespace": "none",
			"bracketPairColorization": true,
			"renderLineHighlight": "line",
			"fontLigatures": false,
			"scrollBeyondLastLine": false,
			"smoothScrolling": false
		}
	}`
	setupTestConfig(t, configJSON)

	cfg := loadConfig()
	require.NotNil(t, cfg.EditorSettings,
		"EditorSettings pointer must not be nil when key is present")
	assert.Equal(t, 4, cfg.EditorSettings.TabSize)
	assert.Equal(t, "block", cfg.EditorSettings.CursorStyle)
}

func TestBackwardCompat_AC5_EmptyConfigFile(t *testing.T) {
	setupTestConfig(t, `{}`)

	cfg := loadConfig()
	assert.Nil(t, cfg.EditorSettings,
		"EditorSettings should be nil for empty config object")
}

// =============================================================================
// AC-1: EditorSettings struct JSON tags
// =============================================================================

func TestEditorSettings_AC1_JSONTags(t *testing.T) {
	es := EditorSettings{
		TabSize:                 4,
		InsertSpaces:            true,
		CursorStyle:             "block",
		CursorBlinking:          "blink",
		MinimapEnabled:          true,
		WordWrap:                "on",
		LineNumbers:             "relative",
		RenderWhitespace:        "all",
		BracketPairColorization: true,
		RenderLineHighlight:     "all",
		FontLigatures:           true,
		ScrollBeyondLastLine:    true,
		SmoothScrolling:         true,
	}

	data, err := json.Marshal(es)
	require.NoError(t, err)

	var raw map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &raw))

	expectedKeys := []string{
		"tabSize", "insertSpaces", "cursorStyle", "cursorBlinking",
		"minimapEnabled", "wordWrap", "lineNumbers", "renderWhitespace",
		"bracketPairColorization", "renderLineHighlight", "fontLigatures",
		"scrollBeyondLastLine", "smoothScrolling",
	}

	for _, key := range expectedKeys {
		_, exists := raw[key]
		assert.True(t, exists, "JSON should contain key %q", key)
	}
	assert.Len(t, raw, 13, "EditorSettings should serialize to exactly 13 JSON keys")
}

func TestEditorSettings_AC1_JSONRoundTrip(t *testing.T) {
	original := EditorSettings{
		TabSize:                 8,
		InsertSpaces:            false,
		CursorStyle:             "underline",
		CursorBlinking:          "phase",
		MinimapEnabled:          true,
		WordWrap:                "bounded",
		LineNumbers:             "interval",
		RenderWhitespace:        "trailing",
		BracketPairColorization: false,
		RenderLineHighlight:     "gutter",
		FontLigatures:           true,
		ScrollBeyondLastLine:    true,
		SmoothScrolling:         true,
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded EditorSettings
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, original, decoded, "JSON round-trip must preserve all fields")
}
