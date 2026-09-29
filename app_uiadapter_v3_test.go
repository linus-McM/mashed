package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTempConfig redirects HOME to a temp dir so each test works on its
// own config.json without polluting the user's ~/.mashed.
func setupTempConfig(t *testing.T) {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	_ = os.MkdirAll(filepath.Join(tmp, ".mashed"), 0o755)
}

// TestApp_SetBackend_Validates — AC-18.8. Unknown backend wraps
// ErrInvalidBackend.
func TestApp_SetBackend_Validates(t *testing.T) {
	setupTempConfig(t)
	a := &App{}
	require.NoError(t, a.SetBackend("ollama"))
	require.NoError(t, a.SetBackend("claude-api"))
	require.NoError(t, a.SetBackend("claude-cli"))

	err := a.SetBackend("injected-via-devtools")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidBackend), "must wrap sentinel")
}

// TestApp_SetClaudeModel_Validates — AC-18.8. Only allowlisted Claude
// models accepted.
func TestApp_SetClaudeModel_Validates(t *testing.T) {
	setupTempConfig(t)
	a := &App{}
	require.NoError(t, a.SetClaudeModel("claude-haiku-4-5"))
	require.NoError(t, a.SetClaudeModel("claude-sonnet-4-6"))

	err := a.SetClaudeModel("claude-foo-9")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidClaudeModel))
}

// TestApp_SetRouterPolicy_Validates — AC-18.8. All 6 plan policies +
// rejection.
func TestApp_SetRouterPolicy_Validates(t *testing.T) {
	setupTempConfig(t)
	a := &App{}
	for _, p := range []string{"local-only", "claude-only", "claude-first", "ollama-first", "cost-aware", "privacy-strict"} {
		require.NoError(t, a.SetRouterPolicy(p), "policy %q must be accepted", p)
	}
	err := a.SetRouterPolicy("policy-that-does-not-exist")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidRouterPolicy))
}

// TestApp_ListBackendsAvailable_StableOrder — the three enums return
// deterministic order so the frontend render is stable.
func TestApp_ListBackendsAvailable_StableOrder(t *testing.T) {
	t.Parallel()
	a := &App{}
	assert.Equal(t, []string{"ollama", "claude-api", "claude-cli"}, a.ListBackendsAvailable())
	assert.Equal(t, []string{"claude-haiku-4-5", "claude-sonnet-4-6", "claude-opus-4-6"}, a.ListClaudeModels())
	assert.Equal(t, []string{"local-only", "claude-only", "claude-first", "ollama-first", "cost-aware", "privacy-strict"}, a.ListRouterPolicies())
}

// TestApp_Setters_RoundtripToConfig — values persist through saveConfig +
// loadConfig.
func TestApp_Setters_RoundtripToConfig(t *testing.T) {
	setupTempConfig(t)
	a := &App{}
	require.NoError(t, a.SetBackend("claude-api"))
	require.NoError(t, a.SetClaudeModel("claude-sonnet-4-6"))
	require.NoError(t, a.SetCLIModel("claude-opus-4-6"))
	require.NoError(t, a.SetRouterPolicy("cost-aware"))

	cfg := loadConfig()
	assert.Equal(t, "claude-api", cfg.Backend)
	assert.Equal(t, "claude-sonnet-4-6", cfg.ClaudeModel)
	assert.Equal(t, "claude-opus-4-6", cfg.CLIModel)
	assert.Equal(t, "cost-aware", cfg.RouterPolicy)
}
