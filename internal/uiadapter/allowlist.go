package uiadapter

import (
	"context"
	"log/slog"
	"slices"
)

// Plan §3 Story 12 — per-backend model allowlist. Config-layer decisions
// warn, they don't panic (§6.5 "Never panic in library code").

// allowlistOp is the canonical op-attr value for allowlist.* records.
const allowlistOp = "allowlist.check"

// OllamaAllowlist — models with golden fixtures behind them.
var OllamaAllowlist = []string{
	"gemma3:1b",
	"gemma3:4b",
	"gemma3:4b-it-qat",
	"gemma3:12b",
}

// ClaudeAllowlist — Anthropic models this v3.0 release has been tested
// against.
var ClaudeAllowlist = []string{
	"claude-haiku-4-5",
	"claude-sonnet-4-6",
	"claude-opus-4-6",
}

// CheckModelAllowlist emits a Warn-level slog line if the configured
// Ollama or Claude model is outside its allowlist and AllowUnvettedModels
// is false. Returns true if a warning fired (callers may use the return
// for test assertions).
//
// Story 5: nil-logger fallback now routes through nilSafeLogger; the
// process-default fallback was removed so all package logging stays scoped
// to the caller-provided logger. On the vetted-OK path emits a single
// allowlist.ok Debug record carrying op + model.
func CheckModelAllowlist(cfg Config, logger *slog.Logger) (warned bool) {
	logger = nilSafeLogger(logger)
	if cfg.AllowUnvettedModels {
		return false
	}
	if cfg.Model != "" && !slices.Contains(OllamaAllowlist, cfg.Model) {
		logger.Warn("uiadapter.model.unvetted",
			slog.String("provider", "ollama"),
			slog.String("model", cfg.Model),
			slog.String("allowlist", joinAllowlist(OllamaAllowlist)),
			slog.String("override_flag", "Config.AllowUnvettedModels"),
		)
		warned = true
	}
	claudeModels := []string{cfg.ClaudeModelPrimary, cfg.ClaudeModelHard}
	for _, m := range claudeModels {
		if m == "" {
			continue
		}
		if !slices.Contains(ClaudeAllowlist, m) {
			logger.Warn("uiadapter.model.unvetted",
				slog.String("provider", "claude"),
				slog.String("model", m),
				slog.String("allowlist", joinAllowlist(ClaudeAllowlist)),
				slog.String("override_flag", "Config.AllowUnvettedModels"),
			)
			warned = true
		}
	}
	if !warned && logger.Enabled(context.Background(), slog.LevelDebug) {
		logger.LogAttrs(context.Background(), slog.LevelDebug, "allowlist.ok",
			slog.String("op", allowlistOp),
			slog.String("model", cfg.Model),
		)
	}
	return warned
}

func joinAllowlist(list []string) string {
	out := ""
	for i, s := range list {
		if i > 0 {
			out += ","
		}
		out += s
	}
	return out
}
