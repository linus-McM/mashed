package uiadapter

import (
	"log/slog"
	"slices"
)

// Plan §3 Story 12 — per-backend model allowlist. Config-layer decisions
// warn, they don't panic (§6.5 "Never panic in library code").

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
func CheckModelAllowlist(cfg Config, logger *slog.Logger) (warned bool) {
	if logger == nil {
		logger = slog.Default()
	}
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
