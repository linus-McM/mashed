package uiadapter

import (
	"context"
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

// op-attr constants for contextguard.* records — single source of truth so
// reviewers can grep by operation.
const (
	contextguardOllamaOp = "contextguard.ollama"
	contextguardClaudeOp = "contextguard.claude"
)

// approxBytesPerToken is the coarse byte→token ratio (4 bytes ≈ 1 token).
// Mirrored across both Apply* helpers so the truncation telemetry stays
// consistent with the byte-budget arithmetic.
const approxBytesPerToken = 4

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
// occurred. Story 5: emits contextguard.ollama.start + contextguard.ollama.truncated;
// neither record carries any substring of raw.
func (g *ContextGuard) ApplyOllama(raw string) (string, bool) {
	ctx := context.Background()
	if g.logger.Enabled(ctx, slog.LevelDebug) {
		g.logger.LogAttrs(ctx, slog.LevelDebug, "contextguard.ollama.start",
			slog.String("op", contextguardOllamaOp),
			slog.Int("bytes_in", len(raw)),
		)
	}
	numCtx := g.cfg.NumCtx
	if numCtx <= 0 {
		numCtx = 8192
	}
	const reserve = 1024
	budgetTokens := numCtx - reserve
	if budgetTokens < 128 {
		budgetTokens = 128
	}
	budgetBytes := budgetTokens * approxBytesPerToken
	originalTokens := len(raw) / approxBytesPerToken

	emitTruncated := func(out string, truncated bool) {
		if !g.logger.Enabled(ctx, slog.LevelDebug) {
			return
		}
		fittedTokens := originalTokens
		if truncated {
			fittedTokens = len(out) / approxBytesPerToken
		}
		g.logger.LogAttrs(ctx, slog.LevelDebug, "contextguard.ollama.truncated",
			slog.String("op", contextguardOllamaOp),
			slog.Bool("truncated", truncated),
			slog.Int("original_tokens", originalTokens),
			slog.Int("fitted_tokens", fittedTokens),
			slog.Int("reserve_tokens", reserve),
		)
	}

	if len(raw) <= budgetBytes {
		emitTruncated(raw, false)
		return raw, false
	}
	out := truncateToBudget(raw, budgetBytes)
	emitTruncated(out, true)
	return out, true
}

// truncateToBudget keeps the head and tail of raw with a visible sentinel
// replacing the dropped middle. Prefers a line-aware variant when raw has
// at least four lines; falls back to byte-aware truncation otherwise or when
// the line-aware output overshoots the budget.
func truncateToBudget(raw string, budgetBytes int) string {
	half := budgetBytes / 2
	byteVariant := func() string {
		return raw[:half] + "\n[... content elided ...]\n" + raw[len(raw)-half:]
	}
	lines := strings.Split(raw, "\n")
	if len(lines) < 4 {
		return byteVariant()
	}
	headLines := lines[:len(lines)/4]
	tailLines := lines[3*len(lines)/4:]
	dropped := len(lines) - len(headLines) - len(tailLines)
	sentinel := fmt.Sprintf("[... %d lines elided ...]", dropped)
	out := strings.Join(headLines, "\n") + "\n" + sentinel + "\n" + strings.Join(tailLines, "\n")
	if len(out) > budgetBytes {
		return byteVariant()
	}
	return out
}

// ApplyClaude enforces the hard-limit contract. Claude returns HTTP 400 if
// input+max_tokens exceeds the context window; we refuse early so the
// fallback chain (Story v3-11) can escalate to a secondary backend.
// Returns the raw verbatim on the happy path; ErrContextOverflow otherwise.
//
// maxModelContext should be the backend's Capabilities().MaxContextTokens
// (Story C). Claude Haiku / Sonnet both quote 200_000 tokens.
//
// Story 5: emits contextguard.claude.start + contextguard.claude.truncated;
// neither record carries any substring of raw.
func (g *ContextGuard) ApplyClaude(raw string, maxModelContext int) (string, error) {
	ctx := context.Background()
	if g.logger.Enabled(ctx, slog.LevelDebug) {
		g.logger.LogAttrs(ctx, slog.LevelDebug, "contextguard.claude.start",
			slog.String("op", contextguardClaudeOp),
			slog.Int("bytes_in", len(raw)),
			slog.Int("max_model_context", maxModelContext),
		)
	}
	approxTokens := len(raw) / approxBytesPerToken
	headroom := maxModelContext - g.cfg.ClaudeMaxTokens
	if headroom <= 0 {
		headroom = maxModelContext
	}
	overflow := approxTokens > headroom
	if g.logger.Enabled(ctx, slog.LevelDebug) {
		fittedTokens := approxTokens
		if overflow {
			fittedTokens = headroom
		}
		g.logger.LogAttrs(ctx, slog.LevelDebug, "contextguard.claude.truncated",
			slog.String("op", contextguardClaudeOp),
			slog.Bool("truncated", overflow),
			slog.Int("original_tokens", approxTokens),
			slog.Int("fitted_tokens", fittedTokens),
		)
	}
	if overflow {
		return "", fmt.Errorf("%w: %d tokens exceeds Claude headroom %d",
			ErrContextOverflow, approxTokens, headroom)
	}
	return raw, nil
}

// OllamaOptions emits the Ollama options block for a chat request. Sets
// `num_ctx` + `keep_alive` from the Config so the model loader sizes the
// KV cache correctly and the model stays resident (Plan §3 Story 3 +
// Story 7).
//
// Story 5: emits a single contextguard.ollama.options Debug record per call
// for accessor visibility (called per-request via Translate).
func (g *ContextGuard) OllamaOptions() map[string]any {
	opts := map[string]any{}
	if g.cfg.NumCtx > 0 {
		opts["num_ctx"] = g.cfg.NumCtx
	}
	if g.cfg.KeepAlive != "" {
		opts["keep_alive"] = g.cfg.KeepAlive
	}
	ctx := context.Background()
	if g.logger.Enabled(ctx, slog.LevelDebug) {
		g.logger.LogAttrs(ctx, slog.LevelDebug, "contextguard.ollama.options",
			slog.String("op", contextguardOllamaOp),
		)
	}
	return opts
}
