package uiadapter

// Plan §3 Story 9 — deterministic sampling. Per-backend because Claude has
// no seed; Ollama has a flaky one. Story 13's eval harness applies the
// per-backend tolerance.

// OllamaSamplingOptions returns the Ollama-specific options for a
// deterministic call (Config.Temperature, Seed, top_k, top_p). Every knob
// is sourced from Config per §6.5 "DI via Config" — no package constants.
func OllamaSamplingOptions(cfg Config) map[string]any {
	return map[string]any{
		"temperature": cfg.Temperature,
		"seed":        cfg.Seed,
		"top_k":       1,
		"top_p":       1.0,
	}
}

// ClaudeSamplingOptions returns the Anthropic-compatible sampling block.
// No seed available — temperature=0 is the only lever. Included here so
// a single helper keeps the v3 plan's "one place per concern" invariant
// and later Stories can inject variant policies (Story v3-10 repair,
// Story v3-16 router) by swapping this helper.
func ClaudeSamplingOptions(cfg Config) map[string]any {
	return map[string]any{
		"temperature": cfg.Temperature,
	}
}
