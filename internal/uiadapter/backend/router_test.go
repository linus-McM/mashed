package backend

import (
	"errors"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mashed/internal/uiadapter"
)

func newTestRouter(t *testing.T, policy RouterPolicy) *Router {
	t.Helper()
	cfg := uiadapter.DefaultConfig()
	cfg.RouterPolicy = string(policy)
	backends := map[string]LLMBackend{
		"ollama":     NewStub("ollama", Capabilities{Provider: "ollama"}),
		"claude-api": NewStub("claude-api", Capabilities{Provider: "anthropic-api"}),
	}
	return NewRouter(cfg, backends)
}

// TestRouter_CoversEveryPolicy — AC-16.1 matrix (subset). Every policy
// produces a Decision with a non-empty Primary.
func TestRouter_CoversEveryPolicy(t *testing.T) {
	t.Parallel()
	policies := []RouterPolicy{
		PolicyLocalOnly,
		PolicyClaudeOnly,
		PolicyClaudeFirst,
		PolicyOllamaFirst,
		PolicyCostAware,
		PolicyPrivacyStrict,
	}
	for _, p := range policies {
		t.Run(string(p), func(t *testing.T) {
			r := newTestRouter(t, p)
			d := r.Decide("neutral raw")
			assert.NotEmpty(t, d.Primary, "policy %s must name a primary", p)
			assert.NotEmpty(t, d.Reason)
		})
	}
}

// TestRouter_PrivacyStrictBlocksClaude — AC-16.2. AWS-key-shaped raw
// with privacy-strict policy picks ollama.
func TestRouter_PrivacyStrictBlocksClaude(t *testing.T) {
	t.Parallel()
	r := newTestRouter(t, PolicyClaudeFirst) // even under claude-first, privacy match overrides
	d := r.Decide("my key is AKIAIOSFODNN7EXAMPLE")
	assert.Equal(t, "ollama", d.Primary)
	assert.Contains(t, d.Reason, "privacy-strict override")
}

// TestRouter_DefaultPolicyFallsBackToLocalWhenClaudeAbsent — Config with
// no Claude backend registered + empty RouterPolicy defaults to local-only.
func TestRouter_DefaultPolicyFallsBackToLocalWhenClaudeAbsent(t *testing.T) {
	t.Parallel()
	cfg := uiadapter.DefaultConfig()
	cfg.RouterPolicy = "" // default
	backends := map[string]LLMBackend{"ollama": NewStub("ollama", Capabilities{})}
	r := NewRouter(cfg, backends)
	d := r.Decide("any")
	assert.Equal(t, "ollama", d.Primary)
}

// TestRouter_CostAwareChoosesOllamaWhenHealthy — cost-aware with healthy
// Ollama routes to Ollama.
func TestRouter_CostAwareChoosesOllamaWhenHealthy(t *testing.T) {
	t.Parallel()
	r := newTestRouter(t, PolicyCostAware)
	r.SetHealth("ollama", nil)
	d := r.Decide("anything")
	assert.Equal(t, "ollama", d.Primary)
	assert.Equal(t, []string{"claude-api"}, d.Fallbacks)
}

// TestRouter_CostAwareFallsToClaudeWhenOllamaDown — cost-aware with
// Ollama unhealthy routes to Claude.
func TestRouter_CostAwareFallsToClaudeWhenOllamaDown(t *testing.T) {
	t.Parallel()
	r := newTestRouter(t, PolicyCostAware)
	r.SetHealth("ollama", errors.New("connection refused"))
	d := r.Decide("anything")
	assert.Equal(t, "claude-api", d.Primary)
}

// TestRouter_ConfigurablePrivacyPatterns — callers can override the
// defaults via Config.PrivacyPatterns.
func TestRouter_ConfigurablePrivacyPatterns(t *testing.T) {
	t.Parallel()
	cfg := uiadapter.DefaultConfig()
	cfg.RouterPolicy = "claude-first"
	cfg.PrivacyPatterns = []*regexp.Regexp{regexp.MustCompile(`SECRET-\d+`)}
	backends := map[string]LLMBackend{
		"ollama":     NewStub("ollama", Capabilities{}),
		"claude-api": NewStub("claude-api", Capabilities{}),
	}
	r := NewRouter(cfg, backends)
	d := r.Decide("captured: SECRET-12345")
	assert.Equal(t, "ollama", d.Primary)

	d2 := r.Decide("no secrets here, just prose")
	assert.Equal(t, "claude-api", d2.Primary)
}

// TestRouter_ResolveUnknownBackend — missing backend surfaces errNoBackend.
func TestRouter_ResolveUnknownBackend(t *testing.T) {
	t.Parallel()
	r := newTestRouter(t, PolicyClaudeFirst)
	_, err := r.Resolve(Decision{Primary: "nonexistent"})
	require.Error(t, err)
	assert.ErrorIs(t, err, errNoBackend)
}
