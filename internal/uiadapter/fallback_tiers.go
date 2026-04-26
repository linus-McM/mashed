package uiadapter

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"
)

// fallbackTierOp is the canonical `op` attribute value for fallback-tier
// log emissions (Story 4 §14 sanitize discipline — closed enum).
const fallbackTierOp = "fallback.tier"

// tierFailureReason classifies a backend failure into the closed-enum
// `reason` token Story 4 §14 permits on `fallback.tier.failure` records.
// Unknown errors collapse to "transport" so err.Error() text never reaches
// the log, regardless of the upstream error wrapping.
func tierFailureReason(err error) string {
	if err == nil {
		return ""
	}
	switch {
	case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
		return "context_overflow"
	default:
		return "transport"
	}
}

// Plan §3 Story 11 — Tiered Fallback (richest → simplest).
//
//  1. Primary backend (validate + repair if enabled).
//  2. Secondary backend per Config.FallbackOrder.
//  3. Minimal-kind UIAST if classifier succeeded.
//  4. Plaintext widget (existing FallbackAST).
//
// TierResult tags the serve path with `escalated_from` on the slog line.

// FallbackTier enumerates the tiers a Translate result may have been
// sourced from. Stamped onto Diagnostics.FallbackReasons via the
// adapter's logging path.
type FallbackTier string

const (
	TierPrimary    FallbackTier = "primary"
	TierSecondary  FallbackTier = "secondary"
	TierMinimal    FallbackTier = "minimal-kind"
	TierPlaintext  FallbackTier = "plaintext"
)

// RunWithFallback dispatches fn against each backend in order. Errors
// from one backend (including breaker-open, context overflow, cost
// budget) advance to the next. Returns the served AST plus the tier
// that produced it. escalatedFrom is the name of the upstream backend
// whose failure triggered escalation (empty string on primary success).
func RunWithFallback(
	ctx context.Context,
	order []string,
	fn func(ctx context.Context, backendName string) (*UIAST, error),
	minimalKind func(ctx context.Context) (*UIAST, error),
	plaintext func() *UIAST,
	logger *slog.Logger,
) (*UIAST, FallbackTier, string, error) {
	log := nilSafeLogger(logger)
	debug := log.Enabled(ctx, slog.LevelDebug)
	if debug {
		log.LogAttrs(ctx, slog.LevelDebug, "fallback.tier.start",
			slog.String("op", fallbackTierOp),
			slog.Int("tiers_count", len(order)))
	}
	if len(order) == 0 {
		return plaintext(), TierPlaintext, "", nil
	}
	var escalatedFrom string
	var errs []error
	for i, name := range order {
		if ctx.Err() != nil {
			return nil, "", escalatedFrom, ctx.Err()
		}
		var attemptStart time.Time
		if debug {
			attemptStart = time.Now()
			log.LogAttrs(ctx, slog.LevelDebug, "fallback.tier.enter",
				slog.String("op", fallbackTierOp),
				slog.String("tier", name),
				slog.Int("attempt_index", i))
		}
		ast, err := fn(ctx, name)
		if err == nil && ast != nil {
			tier := TierPrimary
			if i > 0 {
				tier = TierSecondary
			}
			if debug {
				log.LogAttrs(ctx, slog.LevelDebug, "fallback.tier.success",
					slog.String("op", fallbackTierOp),
					slog.String("tier", name),
					slog.Int64("latency_ms", time.Since(attemptStart).Milliseconds()))
			}
			return ast, tier, escalatedFrom, nil
		}
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", name, err))
			if debug {
				log.LogAttrs(ctx, slog.LevelDebug, "fallback.tier.failure",
					slog.String("op", fallbackTierOp),
					slog.String("tier", name),
					slog.String("reason", tierFailureReason(err)))
			}
		}
		escalatedFrom = name
	}
	// All backends failed — try minimal-kind. Successful rescue returns
	// nil error because the caller served a valid AST; the failed upstream
	// errors are carried on the returned ast's Diagnostics (stamped by the
	// adapter at the call site).
	if minimalKind != nil {
		ast, err := minimalKind(ctx)
		if err == nil && ast != nil {
			return ast, TierMinimal, escalatedFrom, nil
		}
		errs = append(errs, fmt.Errorf("minimal-kind: %w", err))
	}
	if debug {
		log.LogAttrs(ctx, slog.LevelDebug, "fallback.tier.exhausted",
			slog.String("op", fallbackTierOp),
			slog.Int("tiers_tried", len(order)))
	}
	// Last resort — plaintext. errors.Join preserves all upstream
	// diagnoses for the §6.1 log line.
	return plaintext(), TierPlaintext, escalatedFrom, errors.Join(errs...)
}
