package main

import (
	"encoding/json"
	"os"
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
	assert.Equal(t, 3000, cfg.UIAdapterTimeoutMs, "missing uiAdapterTimeoutMs must default to 3000 (§4.4)")
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
	assert.Equal(t, 3000, cfg.UIAdapterTimeoutMs)
}

// Missing config file (fresh install) must return defaults, not zero values.
func TestU1_AC4_MashedConfig_MissingFile_ReturnsDefaults(t *testing.T) {
	setupTestConfig(t, "")

	cfg := loadConfig()
	assert.True(t, cfg.UIAdapterEnabled)
	assert.True(t, cfg.OllamaEnabled)
	assert.Equal(t, "gemma3:4b", cfg.OllamaModel)
	assert.Equal(t, 3000, cfg.UIAdapterTimeoutMs)
}
