package uiadapter

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
)

// ErrContextOverflow is surfaced by ContextGuard when a hard-limit backend
// (Claude API/CLI) cannot accept a raw capture. The fallback chain
// (Story v3-11) translates this into a graceful degraded response.
var ErrContextOverflow = errors.New("uiadapter: context overflow")

// ContextGuard sizes a raw capture for a given backend before the LLM call.
// Ollama silently truncates when num_ctx is too small, so we pre-truncate
// with a visible sentinel. Claude returns HTTP 400 on overflow, so we
// refuse explicitly and let the fallback chain handle it.
type ContextGuard struct {
	cfg    Config
	logger *slog.Logger
}

// NewContextGuard returns a guard configured from cfg. Pass the same Config
// the adapter uses — NumCtx, ClaudeMaxTokens and Backend are all read from
// it (Plan §3 Story 3 + §6.5 DI via Config). logger may be nil;
// nilSafeLogger normalises it so the field is always usable.
func NewContextGuard(cfg Config, logger *slog.Logger) *ContextGuard {
	return &ContextGuard{cfg: cfg, logger: nilSafeLogger(logger)}
}

// ApplyOllama fits raw into the Ollama num_ctx budget. We approximate tokens
// as len(raw)/4 (Gemma has no public Go tokenizer; ±20% is sufficient for
// v3.0 — Plan §3 Story 3 Ollama branch). A reserve of 1024 tokens is held
// back for the system prompt + response. On overflow we keep the head and
// the tail, replacing the middle with a visible sentinel; the tail is
// always preserved because the active prompt lives there.
//
// Returns the fitted text and a boolean indicating whether truncation
// occurred. `sanitize_delta_bytes`-style telemetry lives in the adapter's
// log line; callers should surface the `truncated` flag through their own
// attribute.
func (g *ContextGuard) ApplyOllama(raw string) (string, bool) {
	numCtx := g.cfg.NumCtx
	if numCtx <= 0 {
		numCtx = 8192
	}
	const reserve = 1024
	budgetTokens := numCtx - reserve
	if budgetTokens < 128 {
		budgetTokens = 128
	}
	budgetBytes := budgetTokens * 4 // coarse token→byte estimate
	if len(raw) <= budgetBytes {
		return raw, false
	}
	// Keep head and tail; drop middle with sentinel.
	half := budgetBytes / 2
	lines := strings.Split(raw, "\n")
	if len(lines) < 4 {
		head := raw[:half]
		tail := raw[len(raw)-half:]
		return head + "\n[... content elided ...]\n" + tail, true
	}
	// Line-aware variant — preserve readable boundaries.
	headLines := lines[:len(lines)/4]
	tailLines := lines[3*len(lines)/4:]
	dropped := len(lines) - len(headLines) - len(tailLines)
	sentinel := fmt.Sprintf("[... %d lines elided ...]", dropped)
	out := strings.Join(headLines, "\n") + "\n" + sentinel + "\n" + strings.Join(tailLines, "\n")
	// If we over-shot the budget, fall back to byte-aware truncation.
	if len(out) > budgetBytes {
		head := raw[:half]
		tail := raw[len(raw)-half:]
		return head + "\n[... content elided ...]\n" + tail, true
	}
	return out, true
}

// ApplyClaude enforces the hard-limit contract. Claude returns HTTP 400 if
// input+max_tokens exceeds the context window; we refuse early so the
// fallback chain (Story v3-11) can escalate to a secondary backend.
// Returns the raw verbatim on the happy path; ErrContextOverflow otherwise.
//
// maxModelContext should be the backend's Capabilities().MaxContextTokens
// (Story C). Claude Haiku / Sonnet both quote 200_000 tokens.
func (g *ContextGuard) ApplyClaude(raw string, maxModelContext int) (string, error) {
	approxTokens := len(raw) / 4
	headroom := maxModelContext - g.cfg.ClaudeMaxTokens
	if headroom <= 0 {
		headroom = maxModelContext
	}
	if approxTokens > headroom {
		return "", fmt.Errorf("%w: %d tokens exceeds Claude headroom %d",
			ErrContextOverflow, approxTokens, headroom)
	}
	return raw, nil
}

// OllamaOptions emits the Ollama options block for a chat request. Sets
// `num_ctx` + `keep_alive` from the Config so the model loader sizes the
// KV cache correctly and the model stays resident (Plan §3 Story 3 +
// Story 7).
func (g *ContextGuard) OllamaOptions() map[string]any {
	opts := map[string]any{}
	if g.cfg.NumCtx > 0 {
		opts["num_ctx"] = g.cfg.NumCtx
	}
	if g.cfg.KeepAlive != "" {
		opts["keep_alive"] = g.cfg.KeepAlive
	}
	return opts
}
