package uiadapter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestOllamaSamplingOptions_DeterministicDefaults — AC-9.1 Ollama branch.
// Story 9 pins temperature=0, seed=42, top_k=1, top_p=1.0 by default.
func TestOllamaSamplingOptions_DeterministicDefaults(t *testing.T) {
	t.Parallel()
	opts := OllamaSamplingOptions(DefaultConfig(), nil)
	assert.EqualValues(t, 0, opts["temperature"])
	assert.EqualValues(t, 42, opts["seed"])
	assert.Equal(t, 1, opts["top_k"])
	assert.EqualValues(t, 1.0, opts["top_p"])
}

// TestClaudeSamplingOptions_NoSeed — AC-9.1 Claude branch. Claude has no
// seed exposed — temperature is the only lever.
func TestClaudeSamplingOptions_NoSeed(t *testing.T) {
	t.Parallel()
	opts := ClaudeSamplingOptions(DefaultConfig(), nil)
	assert.Contains(t, opts, "temperature")
	assert.NotContains(t, opts, "seed", "Anthropic Messages API has no seed param")
}

// TestSamplingOptions_ConfigurableThroughConfig — AC-9.2 — the eval
// harness can vary every knob through Config.
func TestSamplingOptions_ConfigurableThroughConfig(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cfg.Temperature = 0.5
	cfg.Seed = 1234
	ollamaOpts := OllamaSamplingOptions(cfg, nil)
	assert.EqualValues(t, 0.5, ollamaOpts["temperature"])
	assert.EqualValues(t, 1234, ollamaOpts["seed"])
}
