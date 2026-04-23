package uiadapter

import (
	"context"
	"errors"
	"fmt"
)

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
) (*UIAST, FallbackTier, string, error) {
	if len(order) == 0 {
		return plaintext(), TierPlaintext, "", nil
	}
	var escalatedFrom string
	var errs []error
	for i, name := range order {
		if ctx.Err() != nil {
			return nil, "", escalatedFrom, ctx.Err()
		}
		ast, err := fn(ctx, name)
		if err == nil && ast != nil {
			tier := TierPrimary
			if i > 0 {
				tier = TierSecondary
			}
			return ast, tier, escalatedFrom, nil
		}
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", name, err))
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
	// Last resort — plaintext. errors.Join preserves all upstream
	// diagnoses for the §6.1 log line.
	return plaintext(), TierPlaintext, escalatedFrom, errors.Join(errs...)
}
