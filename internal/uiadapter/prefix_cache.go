package uiadapter

import (
	"encoding/json"
	"log/slog"
)

// Plan §3 Story 7 — prompt prefix caching. Both Ollama (implicit via byte-
// stable KV prefix reuse) and Claude (explicit cache_control markers)
// benefit; this helper owns the Claude wire-format so §7.4 byte-
// stability + §7.6 cache-hit invariants can be tested in isolation.

// ClaudeSystemBlock builds the system block on an Anthropic Messages API
// request. Static prefix is wrapped in a cache_control:ephemeral marker
// so Claude prompt caching takes effect from the second call onward. TTL
// is 5m or 1h per Config.PromptCacheTTL (Story B). logger may be nil;
// nilSafeLogger normalises it so any future story can emit telemetry without
// an inline guard.
func ClaudeSystemBlock(staticPrefix string, cfg Config, logger *slog.Logger) []map[string]any {
	_ = nilSafeLogger(logger)
	if staticPrefix == "" {
		return nil
	}
	ttl := cfg.PromptCacheTTL
	if ttl == "" {
		ttl = "5m"
	}
	block := map[string]any{
		"type": "text",
		"text": staticPrefix,
	}
	if ttl != "off" {
		block["cache_control"] = map[string]any{
			"type": cacheControlType(ttl),
		}
	}
	return []map[string]any{block}
}

// cacheControlType maps the Config TTL string to the Anthropic wire value.
// Anthropic accepts "ephemeral" (5m) or — with a specific API version —
// extended 1h. We emit ephemeral unconditionally for v3.0; the TTL is
// communicated via the `ttl` header in Story v3-14.
func cacheControlType(ttl string) string {
	if ttl == "off" {
		return ""
	}
	return "ephemeral"
}

// OllamaKeepAliveEncoded marshals Config.KeepAlive for the Ollama options
// block. The option value is a duration string ("30m", "-1") and rides
// alongside num_ctx in ContextGuard.OllamaOptions (Story v3-03).
// Exposed separately so tests can assert the duration string shape.
func OllamaKeepAliveEncoded(cfg Config) string {
	if cfg.KeepAlive == "" {
		return "30m"
	}
	return cfg.KeepAlive
}

// ClaudeSystemBlockJSON returns the marshalled system block. Used by
// tests that compare the wire payload byte-for-byte across calls for
// AC-7.4 byte-stability assertion. logger may be nil; nilSafeLogger
// normalises it so any future story can emit telemetry without an inline
// guard.
func ClaudeSystemBlockJSON(staticPrefix string, cfg Config, logger *slog.Logger) ([]byte, error) {
	return json.Marshal(ClaudeSystemBlock(staticPrefix, cfg, logger))
}
