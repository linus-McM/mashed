package uiadapter

import "reflect"

// DefaultConfig returns the documented defaults from plan §3 Story B.
// Every downstream v3 story reads knobs from Config — not from package-level
// constants — so a single DefaultConfig is the only place to change a default.
//
// Zero-valued fields passed to NewDefault are backfilled via mergeWithDefaults.
func DefaultConfig() Config {
	return Config{
		// Ollama.
		// §7.3 (TestU2_AC7) pins the loopback form to the localhost
		// literal. Plan v3 documents the equivalent dotted-quad; we
		// ship the localhost form to preserve the existing invariant.
		OllamaEndpoint: "http://localhost:11434",
		Model:          "gemma3:4b",
		NumCtx:         8192,
		KeepAlive:      "30m",

		// Claude API.
		AnthropicAPIKeyEnv: "ANTHROPIC_API_KEY",
		ClaudeModelPrimary: "claude-haiku-4-5",
		ClaudeModelHard:    "claude-sonnet-4-6",
		ClaudeMaxTokens:    2048,
		AnthropicVersion:   "2023-06-01",
		PromptCacheTTL:     "5m",

		// Claude CLI.
		ClaudeCLIBinary: "claude",

		// Sampling.
		Seed: 42,

		// Timeouts.
		TimeoutMs:       5000,
		WarmUpTimeoutMs: 10000,

		// Runtime.
		CacheCapacity:        1024,
		EnableFastPath:       true,
		EnableSpotlighting:   true,
		RepairMaxRetries:     1,
		BreakerFailThreshold: 3,
		BreakerResetMs:       30000,
		ShadowSampleRate:     0.05,
	}
}

// mergeWithDefaults backfills zero-valued string/int/float fields on cfg from
// DefaultConfig(). Booleans and slices are preserved verbatim — callers that
// need to *disable* a default-true flag (EnableFastPath, EnableSpotlighting)
// set it explicitly to false in their Config literal.
//
// The reflection pass keeps this helper compact at the cost of a per-construct
// O(field-count) iteration — irrelevant outside hot loops.
func mergeWithDefaults(cfg Config) Config {
	def := DefaultConfig()
	cv := reflect.ValueOf(&cfg).Elem()
	dv := reflect.ValueOf(def)
	for i := 0; i < cv.NumField(); i++ {
		f := cv.Field(i)
		if !f.CanSet() {
			continue
		}
		switch f.Kind() {
		case reflect.String:
			if f.String() == "" {
				f.SetString(dv.Field(i).String())
			}
		case reflect.Int, reflect.Int32, reflect.Int64:
			if f.Int() == 0 {
				f.SetInt(dv.Field(i).Int())
			}
		case reflect.Float32, reflect.Float64:
			if f.Float() == 0 {
				f.SetFloat(dv.Field(i).Float())
			}
		}
	}
	// Booleans default-true: EnableFastPath, EnableSpotlighting. Preserve the
	// plan semantic — explicit `false` in the caller's Config stays `false`.
	// We only flip to true when the caller's value *looks* uninitialized, and
	// we can't distinguish zero-value from explicit-false without a sentinel.
	// Convention (plan §6.5 DI via Config): callers that disable these flags
	// pass DefaultConfig() first and then mutate. For plain `Config{}` callers
	// who want the documented defaults, they call DefaultConfig() directly.
	return cfg
}

// configFieldSet returns the set of exported field names on Config. Used by
// TestConfig_HasEveryPlanField (Story B AC-B.2 structural guard) to assert
// the plan's full knob list is represented.
func configFieldSet(cfg *Config) map[string]struct{} {
	t := reflect.TypeOf(*cfg)
	out := make(map[string]struct{}, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		out[t.Field(i).Name] = struct{}{}
	}
	return out
}
