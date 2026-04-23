package uiadapter

import (
	"context"
	"io"
	"log/slog"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDefaultConfig_Bootable_AllBackends — Story B AC-B.1. Zero-valued Config
// passed through DefaultConfig boots a working adapter. Every field has a
// sane default; NewDefault does not panic; Translate returns a valid UIAST
// even when the underlying backend is unreachable (falls through to the
// FallbackAST chain).
func TestDefaultConfig_Bootable_AllBackends(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	require.NotNil(t, cfg, "DefaultConfig must return a non-nil pointer receiver")

	// Plan §3 Story B defaults.
	assert.Equal(t, "http://localhost:11434", cfg.OllamaEndpoint,
		"OllamaEndpoint default — §7.3 pins localhost form over the 127.0.0.1 literal")
	assert.Equal(t, "gemma3:4b", cfg.Model, "Model default")
	assert.Equal(t, 8192, cfg.NumCtx, "NumCtx default")
	assert.Equal(t, "30m", cfg.KeepAlive, "KeepAlive default")
	assert.Equal(t, "ANTHROPIC_API_KEY", cfg.AnthropicAPIKeyEnv, "AnthropicAPIKeyEnv default")
	assert.Equal(t, "claude-haiku-4-5", cfg.ClaudeModelPrimary, "ClaudeModelPrimary default")
	assert.Equal(t, "claude-sonnet-4-6", cfg.ClaudeModelHard, "ClaudeModelHard default")
	assert.Equal(t, 2048, cfg.ClaudeMaxTokens, "ClaudeMaxTokens default")
	assert.Equal(t, "2023-06-01", cfg.AnthropicVersion, "AnthropicVersion pin")
	assert.Equal(t, "5m", cfg.PromptCacheTTL, "PromptCacheTTL default")
	assert.Equal(t, "claude", cfg.ClaudeCLIBinary, "ClaudeCLIBinary default")
	assert.InEpsilon(t, 0.05, cfg.ShadowSampleRate, 0.0001, "ShadowSampleRate default")
	assert.Equal(t, 1024, cfg.CacheCapacity, "CacheCapacity default")
	assert.True(t, cfg.EnableFastPath, "EnableFastPath default true")
	assert.True(t, cfg.EnableSpotlighting, "EnableSpotlighting default true")
	assert.Equal(t, 1, cfg.RepairMaxRetries, "RepairMaxRetries default")
	assert.Equal(t, 3, cfg.BreakerFailThreshold, "BreakerFailThreshold default")
	assert.Equal(t, 30000, cfg.BreakerResetMs, "BreakerResetMs default")
	assert.Equal(t, 5000, cfg.TimeoutMs, "TimeoutMs default")
	assert.Equal(t, 10000, cfg.WarmUpTimeoutMs, "WarmUpTimeoutMs default")
	assert.EqualValues(t, 42, cfg.Seed, "Seed default")
}

// TestDefaultConfig_ZeroValueBootsAdapter — Plan Story B AC-B.1 scenario:
// `Given the caller passes Config{} to DefaultConfig() And feeds the resulting
// struct to NewDefault(cfg) When the adapter is constructed Then no panic
// fires`.
func TestDefaultConfig_ZeroValueBootsAdapter(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cfg.Enabled = true
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	adapter := NewDefault(cfg, logger)
	require.NotNil(t, adapter, "NewDefault must not return nil")

	// Translate must not panic even when the backend is unreachable.
	ast := adapter.Translate(context.Background(), "trivial raw capture", "proc-test")
	require.NotNil(t, ast, "Translate must never return nil (§Adapter interface contract)")
	assert.Equal(t, "1", ast.Version, "UIAST envelope version must be '1'")
}

// TestDefaultConfig_MergeOntoZeroValued — AC-B.1 coverage for the merge
// semantics: an empty Config passed through a merge helper inherits every
// default without overriding explicitly set fields.
func TestDefaultConfig_MergeOntoZeroValued(t *testing.T) {
	t.Parallel()
	merged := mergeWithDefaults(Config{Model: "custom-model", TimeoutMs: 7000})
	assert.Equal(t, "custom-model", merged.Model, "explicit Model wins")
	assert.Equal(t, 7000, merged.TimeoutMs, "explicit TimeoutMs wins")
	assert.Equal(t, 1024, merged.CacheCapacity, "zero-value CacheCapacity backfilled to default")
	assert.Equal(t, "http://localhost:11434", merged.OllamaEndpoint, "zero-value OllamaEndpoint backfilled")
}

// TestConfig_HasEveryPlanField — AC-B.2 structural assertion. Every knob the
// plan enumerates (§3 Story B) must be a field on Config so downstream stories
// can dependency-inject it. Failure names the missing field.
func TestConfig_HasEveryPlanField(t *testing.T) {
	t.Parallel()
	planFields := []string{
		"Backend", "RouterPolicy", "FallbackOrder",
		"OllamaEndpoint", "Model", "AllowUnvettedModels", "NumCtx", "KeepAlive", "LooseFormat",
		"AnthropicAPIKeyEnv", "ClaudeModelPrimary", "ClaudeModelHard", "ClaudeMaxTokens",
		"AnthropicVersion", "PromptCacheTTL",
		"ClaudeCLIBinary", "ClaudeCLIExtraFlags",
		"UsdBudgetPerSession", "RPMSoftLimit", "TPMSoftLimit",
		"Temperature", "Seed",
		"TimeoutMs", "WarmUpTimeoutMs",
		"CacheCapacity", "EnableSemanticCache", "EnableFastPath", "EnableSpotlighting",
		"RepairMaxRetries", "BreakerFailThreshold", "BreakerResetMs", "ShadowSampleRate",
		"PrivacyPatterns", "DisableHealthTicker", "Streaming",
	}
	cfg := Config{}
	have := configFieldSet(&cfg)
	for _, f := range planFields {
		assert.Contains(t, have, f,
			"Plan §3 Story B mandates Config.%s; add the field to internal/uiadapter/adapter.go", f)
	}
}

// TestConfig_PrivacyPatternsUsable — AC-B.2: the PrivacyPatterns field must
// carry compiled *regexp.Regexp values so router Story 16 can branch on match
// without re-compiling per-request.
func TestConfig_PrivacyPatternsUsable(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cfg.PrivacyPatterns = []*regexp.Regexp{
		regexp.MustCompile(`AKIA[0-9A-Z]{16}`),
		regexp.MustCompile(`ghp_[A-Za-z0-9]{36,}`),
	}
	assert.Len(t, cfg.PrivacyPatterns, 2)
	assert.True(t, cfg.PrivacyPatterns[0].MatchString("AKIAIOSFODNN7EXAMPLE"))
	assert.True(t, cfg.PrivacyPatterns[1].MatchString("ghp_"+strings.Repeat("a", 36)))
}
