package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestU1_AC4_MashedConfig_UIAdapterFields_RoundTrip(t *testing.T) {
	setupTestConfig(t, "")

	cfg := mashedConfig{
		DevDir:             "/some/dev",
		OllamaEnabled:      true,
		OllamaModel:        "gemma3:4b",
		UIAdapterEnabled:   true,
		UIAdapterTimeoutMs: 3000,
	}
	require.NoError(t, saveConfig(cfg))

	loaded := loadConfig()
	assert.True(t, loaded.OllamaEnabled)
	assert.Equal(t, "gemma3:4b", loaded.OllamaModel)
	assert.True(t, loaded.UIAdapterEnabled)
	assert.Equal(t, 3000, loaded.UIAdapterTimeoutMs)

	raw, err := os.ReadFile(configPath())
	require.NoError(t, err)
	var m map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(raw, &m))
	for _, k := range []string{"ollamaEnabled", "ollamaModel", "uiAdapterEnabled", "uiAdapterTimeoutMs"} {
		assert.Contains(t, m, k, "round-tripped config must retain %q key on disk", k)
	}
}

// Pre-U1 configs have none of the four new keys. Per the 2026-04-21 user
// decision (overriding §4.4's default-false), the enabled flags default TRUE
// on load — a missing key must be treated as opt-in, not opt-out.
func TestU1_AC4_MashedConfig_LegacyLoad_DefaultsApplied(t *testing.T) {
	setupTestConfig(t, `{"devDir":"/Users/x/dev","theme":"mashed-dark"}`)

	cfg := loadConfig()
	assert.True(t, cfg.UIAdapterEnabled, "missing uiAdapterEnabled must default TRUE (2026-04-21 user decision)")
	assert.True(t, cfg.OllamaEnabled, "missing ollamaEnabled must default TRUE (2026-04-21 user decision)")
	assert.Equal(t, "gemma3:4b", cfg.OllamaModel, "missing ollamaModel must default to gemma3:4b (§4.5)")
	assert.Equal(t, 30000, cfg.UIAdapterTimeoutMs, "missing uiAdapterTimeoutMs must default to 30000ms (raised from 3000ms to cover Ollama gemma3:4b cold-load + generation)")
}

// Explicit opt-out must survive the default-fill pass — otherwise a user who
// disabled the adapter would get silently re-enabled on every restart.
func TestU1_AC4_MashedConfig_ExplicitOptOut_Honored(t *testing.T) {
	setupTestConfig(t, `{"devDir":"/x","uiAdapterEnabled":false}`)

	cfg := loadConfig()
	assert.False(t, cfg.UIAdapterEnabled, "explicit uiAdapterEnabled:false must round-trip as false")
}

// A corrupt config.json must not crash the load path. The legacy contract
// (pre-U1) returned a zero-value struct; the new contract returns
// defaultConfig() so the UI AST layer stays enabled-by-default.
func TestU1_AC4_MashedConfig_MalformedJSON_ReturnsDefaults(t *testing.T) {
	setupTestConfig(t, `{this is not json`)

	cfg := loadConfig()
	assert.True(t, cfg.UIAdapterEnabled, "malformed config must fall back to enabled-by-default")
	assert.True(t, cfg.OllamaEnabled)
	assert.Equal(t, "gemma3:4b", cfg.OllamaModel)
	assert.Equal(t, 30000, cfg.UIAdapterTimeoutMs)
}

// Missing config file (fresh install) must return defaults, not zero values.
func TestU1_AC4_MashedConfig_MissingFile_ReturnsDefaults(t *testing.T) {
	setupTestConfig(t, "")

	cfg := loadConfig()
	assert.True(t, cfg.UIAdapterEnabled)
	assert.True(t, cfg.OllamaEnabled)
	assert.Equal(t, "gemma3:4b", cfg.OllamaModel)
	assert.Equal(t, 30000, cfg.UIAdapterTimeoutMs)
}

// M1 (security): a hand-edited config.json must not bypass the binding-layer
// OllamaModel regex — loadConfig revalidates and falls back to the default.
func TestU5_LoadConfig_InvalidModelFallsBackToDefault(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "path traversal", input: `{"devDir":"/x","ollamaModel":"../../etc/passwd"}`},
		{name: "forward slash", input: `{"devDir":"/x","ollamaModel":"foo/bar"}`},
		{name: "whitespace", input: `{"devDir":"/x","ollamaModel":"bad name"}`},
		{name: "shell metachar", input: `{"devDir":"/x","ollamaModel":"foo;rm -rf"}`},
		{name: "over 64 chars", input: `{"devDir":"/x","ollamaModel":"a23456789012345678901234567890123456789012345678901234567890123456"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTestConfig(t, tt.input)
			cfg := loadConfig()
			assert.Equal(t, "gemma3:4b", cfg.OllamaModel,
				"invalid on-disk OllamaModel must fall back to default, not surface to bindings")
		})
	}
}

// R13: config and theme files are owner-only (0600) in an owner-only (0700)
// directory, including a pre-existing 0755 directory.
func TestConfig_FilesAre0600In0700Dir(t *testing.T) {
	app := setupTestConfig(t, `{}`) // creates ~/.mashed as 0755
	require.NoError(t, app.SetTheme("dark"))
	require.NoError(t, app.SaveTheme("t1", `{"label":"x"}`))

	home, _ := os.UserHomeDir()
	for _, p := range []string{configPath(), themesPath()} {
		fi, err := os.Stat(p)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o600), fi.Mode().Perm(), p)
	}
	fi, err := os.Stat(filepath.Join(home, ".mashed"))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o700), fi.Mode().Perm(), "config dir")
}

// R14 regression guard: 50 concurrent SaveTheme calls with distinct ids all
// survive (setters serialise on a.mu; this must stay true).
func TestConfig_ConcurrentSetters_NoLostUpdate(t *testing.T) {
	app := setupTestConfig(t, "")
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			assert.NoError(t, app.SaveTheme(fmt.Sprintf("theme-%d", i), `{}`))
		}(i)
	}
	wg.Wait()
	var all map[string]json.RawMessage
	require.NoError(t, json.Unmarshal([]byte(app.GetSavedThemes()), &all))
	assert.Len(t, all, 50)
}

// R13: readers never observe a half-written config.json while setters write.
func TestConfig_ReadsNeverSeeTornWrite(t *testing.T) {
	app := setupTestConfig(t, `{"theme":"theme-A","sidebarWidth":300}`)
	stop := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; ; i++ {
			select {
			case <-stop:
				return
			default:
			}
			_ = app.SetSidebarWidth(300 + i%7)
		}
	}()
	for i := 0; i < 2000; i++ {
		if got := app.GetConfig().Theme; got != "theme-A" {
			close(stop)
			wg.Wait()
			t.Fatalf("read %d saw Theme=%q (torn write)", i, got)
		}
	}
	close(stop)
	wg.Wait()
}
