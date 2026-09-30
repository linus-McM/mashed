package uiadapter

import (
	"context"
	"log/slog"
)

// Plan §3 Story 9 — deterministic sampling. Per-backend because Claude has
// no seed; Ollama has a flaky one. Story 13's eval harness applies the
// per-backend tolerance.

// op-attr constants for sampling.* records.
const (
	samplingOllamaOp = "sampling.ollama"
	samplingClaudeOp = "sampling.claude"
)

// OllamaSamplingOptions returns the Ollama-specific options for a
// deterministic call (Config.Temperature, Seed, top_k, top_p). Every knob
// is sourced from Config per §6.5 "DI via Config" — no package constants.
// logger may be nil; nilSafeLogger normalises it so any future story can
// emit telemetry without an inline guard.
//
// Story 5: emits sampling.ollama with op, seed, temperature.
func OllamaSamplingOptions(cfg Config, logger *slog.Logger) map[string]any {
	lg := nilSafeLogger(logger)
	if lg.Enabled(context.Background(), slog.LevelDebug) {
		lg.LogAttrs(context.Background(), slog.LevelDebug, "sampling.ollama",
			slog.String("op", samplingOllamaOp),
			slog.Int64("seed", cfg.Seed),
			slog.Float64("temperature", float64(cfg.Temperature)),
		)
	}
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
// Story v3-16 router) by swapping this helper. logger may be nil;
// nilSafeLogger normalises it so any future story can emit telemetry
// without an inline guard.
//
// Story 5: emits sampling.claude with op + temperature only — Claude has
// no seed knob so the attribute is deliberately absent.
func ClaudeSamplingOptions(cfg Config, logger *slog.Logger) map[string]any {
	lg := nilSafeLogger(logger)
	if lg.Enabled(context.Background(), slog.LevelDebug) {
		lg.LogAttrs(context.Background(), slog.LevelDebug, "sampling.claude",
			slog.String("op", samplingClaudeOp),
			slog.Float64("temperature", float64(cfg.Temperature)),
		)
	}
	return map[string]any{
		"temperature": cfg.Temperature,
	}
}
