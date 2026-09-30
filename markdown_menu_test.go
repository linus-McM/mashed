package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// expectedMarkdownMenuDefaults returns the authoritative default struct.
// Any drift here is a spec violation — update the plan before changing.
func expectedMarkdownMenuDefaults() MarkdownMenuSettings {
	return MarkdownMenuSettings{
		Bold:          true,
		Italic:        true,
		Strikethrough: true,
		Code:          true,
		Link:          true,
		Latex:         false,
	}
}

// =============================================================================
// AC-1: Defaults are exactly the values in the table
// =============================================================================

func TestDefaultMarkdownMenuSettings(t *testing.T) {
	app := &App{}
	got := app.DefaultMarkdownMenuSettings()

	assert.True(t, got.Bold, "Bold default must be true")
	assert.True(t, got.Italic, "Italic default must be true")
	assert.True(t, got.Strikethrough, "Strikethrough default must be true")
	assert.True(t, got.Code, "Code default must be true")
	assert.True(t, got.Link, "Link default must be true")
	assert.False(t, got.Latex, "Latex default must be false")

	assert.Equal(t, expectedMarkdownMenuDefaults(), got,
		"DefaultMarkdownMenuSettings must match the authoritative defaults table")
}

func TestDefaultMarkdownMenuSettings_Consistent(t *testing.T) {
	app := &App{}
	assert.Equal(t, app.DefaultMarkdownMenuSettings(), app.DefaultMarkdownMenuSettings(),
		"consecutive calls must return identical defaults")
}

// =============================================================================
// AC-2: GetMarkdownMenuSettings returns defaults when unset (no disk write)
// =============================================================================

func TestGetMarkdownMenuSettings_DefaultsWhenNil(t *testing.T) {
	app := setupTestConfig(t, `{"devDir":"/tmp"}`)

	// Snapshot config bytes before read.
	before, err := os.ReadFile(configPath())
	require.NoError(t, err)

	got := app.GetMarkdownMenuSettings()
	assert.Equal(t, expectedMarkdownMenuDefaults(), got,
		"Get must return defaults when markdownMenu key is absent")

	// File must be byte-identical: Get is pure read.
	after, err := os.ReadFile(configPath())
	require.NoError(t, err)
	assert.Equal(t, before, after,
		"GetMarkdownMenuSettings must not mutate the on-disk config")
}

func TestGetMarkdownMenuSettings_NoConfigFile(t *testing.T) {
	app := setupTestConfig(t, "")

	got := app.GetMarkdownMenuSettings()
	assert.Equal(t, expectedMarkdownMenuDefaults(), got,
		"Get must return defaults when config file is missing")

	// No file must have been created by a read.
	_, err := os.Stat(configPath())
	assert.True(t, os.IsNotExist(err),
		"GetMarkdownMenuSettings must not create a config file")
}

// =============================================================================
// AC-3: Set/Get round-trip persists and reloads identically
// =============================================================================

func TestSetGetMarkdownMenuSettings_RoundTrip(t *testing.T) {
	app := setupTestConfig(t, `{"devDir":"/tmp"}`)

	input := MarkdownMenuSettings{
		Bold:          false,
		Italic:        true,
		Strikethrough: false,
		Code:          true,
		Link:          false,
		Latex:         true,
	}
	require.NoError(t, app.SetMarkdownMenuSettings(input))

	// A *new* App instance must see the persisted value — this is the
	// end-to-end reload semantics the frontend depends on.
	fresh := &App{}
	got := fresh.GetMarkdownMenuSettings()
	assert.Equal(t, input, got, "round-trip must preserve all six fields exactly")

	// The raw JSON file must actually contain the markdownMenu key.
	raw := readRawConfig(t)
	_, ok := raw["markdownMenu"]
	assert.True(t, ok, "config file must contain markdownMenu JSON key after Set")
}

// =============================================================================
// AC-4: omitempty — nil MarkdownMenu serialises without the key
// =============================================================================

func TestConfigMarkdownMenuOmitempty(t *testing.T) {
	cfg := mashedConfig{
		DevDir:       "/tmp",
		MarkdownMenu: nil,
	}
	data, err := json.Marshal(cfg)
	require.NoError(t, err)

	assert.False(t, bytes.Contains(data, []byte("markdownMenu")),
		"nil MarkdownMenu must be omitted from JSON entirely, not emitted as null")

	var decoded mashedConfig
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Nil(t, decoded.MarkdownMenu,
		"round-tripped absent key must unmarshal as nil pointer")
}

func TestConfigMarkdownMenu_PresentWhenSet(t *testing.T) {
	s := MarkdownMenuSettings{Bold: true, Latex: true}
	cfg := mashedConfig{
		DevDir:       "/tmp",
		MarkdownMenu: &s,
	}
	data, err := json.Marshal(cfg)
	require.NoError(t, err)

	assert.True(t, bytes.Contains(data, []byte(`"markdownMenu"`)),
		"non-nil MarkdownMenu must appear in JSON")

	var decoded mashedConfig
	require.NoError(t, json.Unmarshal(data, &decoded))
	require.NotNil(t, decoded.MarkdownMenu)
	assert.Equal(t, s, *decoded.MarkdownMenu)
}

// =============================================================================
// AC-5: Set preserves other config fields
// =============================================================================

func TestSetMarkdownMenuSettings_PreservesOtherConfig(t *testing.T) {
	app := setupTestConfig(t, `{
		"devDir": "/work",
		"theme": "mashed-dark",
		"monoFont": "JetBrains Mono"
	}`)

	input := expectedMarkdownMenuDefaults()
	input.Latex = true
	require.NoError(t, app.SetMarkdownMenuSettings(input))

	raw := readRawConfig(t)

	var devDir string
	require.NoError(t, json.Unmarshal(raw["devDir"], &devDir))
	assert.Equal(t, "/work", devDir, "devDir must be preserved across Set")

	var theme string
	require.NoError(t, json.Unmarshal(raw["theme"], &theme))
	assert.Equal(t, "mashed-dark", theme, "theme must be preserved across Set")

	var monoFont string
	require.NoError(t, json.Unmarshal(raw["monoFont"], &monoFont))
	assert.Equal(t, "JetBrains Mono", monoFont, "monoFont must be preserved across Set")

	// And the value we actually set is reflected.
	var menu MarkdownMenuSettings
	require.NoError(t, json.Unmarshal(raw["markdownMenu"], &menu))
	assert.Equal(t, input, menu)
}

// =============================================================================
// AC-6 supporting: backward-compat — legacy configs leave MarkdownMenu nil
// =============================================================================

func TestBackwardCompat_LegacyConfigNoMarkdownMenu(t *testing.T) {
	setupTestConfig(t, `{"devDir":"/tmp","theme":"mashed-dark"}`)
	cfg := loadConfig()
	assert.Nil(t, cfg.MarkdownMenu,
		"MarkdownMenu must be nil when the key is absent from config")
}

// =============================================================================
// Concurrency: 10 goroutines writing distinct values must be race-clean
// =============================================================================

func TestSetMarkdownMenuSettings_ConcurrentRace(t *testing.T) {
	app := setupTestConfig(t, minimalConfigJSON)

	// Build 10 distinct settings by toggling bits of `i`.
	const goroutines = 10
	values := make([]MarkdownMenuSettings, goroutines)
	for i := 0; i < goroutines; i++ {
		values[i] = MarkdownMenuSettings{
			Bold:          i&1 != 0,
			Italic:        i&2 != 0,
			Strikethrough: i&4 != 0,
			Code:          i&8 != 0,
			Link:          i&16 != 0,
			Latex:         i%2 == 0,
		}
	}

	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(v MarkdownMenuSettings) {
			defer wg.Done()
			_ = app.SetMarkdownMenuSettings(v)
		}(values[i])
	}
	wg.Wait()

	// Final persisted value must equal one of the 10 candidates
	// (last-writer-wins is acceptable per the story).
	got := app.GetMarkdownMenuSettings()
	matched := false
	for _, v := range values {
		if v == got {
			matched = true
			break
		}
	}
	assert.True(t, matched,
		"persisted value must equal one of the values written by the 10 goroutines; got %+v", got)
}

// =============================================================================
// AC-6 supporting: struct JSON tag shape (guards against rename drift)
// =============================================================================

func TestMarkdownMenuSettings_JSONTags(t *testing.T) {
	s := MarkdownMenuSettings{
		Bold:          true,
		Italic:        true,
		Strikethrough: true,
		Code:          true,
		Link:          true,
		Latex:         true,
	}
	data, err := json.Marshal(s)
	require.NoError(t, err)

	var raw map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &raw))

	expected := []string{"bold", "italic", "strikethrough", "code", "link", "latex"}
	for _, key := range expected {
		_, ok := raw[key]
		assert.True(t, ok, "JSON must contain key %q", key)
	}
	assert.Len(t, raw, 6, "MarkdownMenuSettings must serialise to exactly 6 keys")
}

func TestMarkdownMenuSettings_JSONRoundTrip(t *testing.T) {
	original := MarkdownMenuSettings{
		Bold:          true,
		Italic:        false,
		Strikethrough: true,
		Code:          false,
		Link:          true,
		Latex:         true,
	}
	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded MarkdownMenuSettings
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, original, decoded, "JSON round-trip must preserve all fields")
}

// Sanity: the config file written by setupTestConfig lives where we think it does.
func TestMarkdownMenu_ConfigPathLocation(t *testing.T) {
	setupTestConfig(t, minimalConfigJSON)
	home, err := os.UserHomeDir()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(home, ".mashed", "config.json"), configPath())
}
