package uiadapter

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync/atomic"
	"time"
)

// repairOp / repairPromptOp are the canonical `op` values for every repair
// log emission (Story 4 §14 sanitize discipline — closed enum).
const (
	repairOp       = "repair.run"
	repairPromptOp = "repair.prompt"
)

// firstReason returns the first validator reason or "" when the slice is
// empty. Validator reasons are short closed-enum tokens (Story 4 §14) so
// they may be logged verbatim.
func firstReason(errs []string) string {
	if len(errs) == 0 {
		return ""
	}
	return errs[0]
}

// Plan §3 Story 10 — bounded self-repair loop. Schema-constrained decoding
// prevents structural failure; semantic failures (wrong enum, slug drift)
// are retried once with validator-error feedback. Per-backend budget.

// ErrRepairExhausted is surfaced when the repair budget is spent.
var ErrRepairExhausted = errors.New("uiadapter: repair exhausted")

// RepairAttempt captures one round of invalid-output → repair-prompt
// resubmit. GenerateFn executes the retry against the same backend.
type RepairAttempt struct {
	Kind           StageKind
	StaticPrefix   string
	SanitizedRaw   string
	PreviousOutput string
	Errors         []string
}

// BuildRepairPrompt assembles the repair user message.
//
//	<static_prefix>
//	RAW CAPTURE: <raw>
//	KIND: <kind>
//	PREVIOUS ATTEMPT (invalid):
//	<previous bad output>
//	VALIDATION ERRORS:
//	- <err>
//	...
//	Emit a corrected UIAST that addresses each error.
//
// logger may be nil; nilSafeLogger normalises it so the Story 4 telemetry
// emission below stays safe at every call site.
func BuildRepairPrompt(a RepairAttempt, logger *slog.Logger) string {
	log := nilSafeLogger(logger)
	prompt := buildRepairPromptString(a)
	ctx := context.Background()
	if log.Enabled(ctx, slog.LevelDebug) {
		log.LogAttrs(ctx, slog.LevelDebug, "repair.prompt.build",
			slog.String("op", repairPromptOp),
			slog.Int("bytes_out", len(prompt)))
	}
	return prompt
}

// buildRepairPromptString assembles the prompt body without any logging
// side-effect. Used by Repairer.Run so each retry does not double-emit a
// `repair.prompt.build` record alongside the loop's `repair.attempt` /
// `repair.failure` script (Story 4 AC-4.2 ordering invariant).
func buildRepairPromptString(a RepairAttempt) string {
	var b strings.Builder
	b.WriteString(a.StaticPrefix)
	b.WriteString("\n\nRAW CAPTURE: ")
	b.WriteString(a.SanitizedRaw)
	b.WriteString("\nKIND: ")
	b.WriteString(string(a.Kind))
	b.WriteString("\nPREVIOUS ATTEMPT (invalid):\n")
	b.WriteString(a.PreviousOutput)
	b.WriteString("\nVALIDATION ERRORS:\n")
	for _, e := range a.Errors {
		b.WriteString("- ")
		b.WriteString(e)
		b.WriteString("\n")
	}
	b.WriteString("\nEmit a corrected UIAST that addresses each error. Do not repeat the previous errors.\n")
	return b.String()
}

// Repairer tracks repair-attempt counters across Translate calls so
// Story 13's scorecard can compute recovery rate. Safe for concurrent
// use.
type Repairer struct {
	maxRetries int
	attempts   atomic.Int64
	succeeded  atomic.Int64
	logger     *slog.Logger
}

// NewRepairer constructs a Repairer from the adapter Config. maxRetries
// ≤ 0 disables repair (Sonnet path per plan §3 Story 10). logger may be
// nil; nilSafeLogger normalises it so the field is always usable.
func NewRepairer(cfg Config, logger *slog.Logger) *Repairer {
	return &Repairer{
		maxRetries: cfg.RepairMaxRetries,
		logger:     nilSafeLogger(logger),
	}
}

// MaxRetries returns the configured budget. Used by the router to short-
// circuit repair on Sonnet.
func (r *Repairer) MaxRetries() int { return r.maxRetries }

// Run applies up to maxRetries repair attempts. validate returns the
// validation errors (nil/empty on success). generate runs one more
// downstream call; if validate succeeds the result is returned. Returns
// ErrRepairExhausted wrapped on budget exhaustion.
func (r *Repairer) Run(
	ctx context.Context,
	initial *UIAST,
	initialBody string,
	attempt RepairAttempt,
	validate func(*UIAST) []string,
	generate func(ctx context.Context, repairPrompt string) (*UIAST, string, error),
) (*UIAST, error) {
	debug := r.logger.Enabled(ctx, slog.LevelDebug)
	if debug {
		r.logger.LogAttrs(ctx, slog.LevelDebug, "repair.start",
			slog.String("op", repairOp),
			slog.Int("max_retries", r.maxRetries),
			slog.Int("bytes_in", len(initialBody)))
	}
	ast := initial
	body := initialBody
	errorsList := validate(ast)
	if len(errorsList) == 0 {
		// Fast path — no repair needed.
		return ast, nil
	}
	for i := 0; i < r.maxRetries; i++ {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		attemptNum := i + 1
		if debug {
			r.logger.LogAttrs(ctx, slog.LevelDebug, "repair.attempt",
				slog.String("op", repairOp),
				slog.Int("attempt", attemptNum),
				slog.String("reason", firstReason(errorsList)))
		}
		r.attempts.Add(1)
		attempt.PreviousOutput = body
		attempt.Errors = errorsList
		repairPrompt := buildRepairPromptString(attempt)
		var attemptStart time.Time
		if debug {
			attemptStart = time.Now()
		}
		next, nextBody, err := generate(ctx, repairPrompt)
		if err != nil {
			return nil, fmt.Errorf("uiadapter: repair attempt %d: %w", attemptNum, err)
		}
		ast = next
		body = nextBody
		errorsList = validate(ast)
		if len(errorsList) == 0 {
			r.succeeded.Add(1)
			if debug {
				r.logger.LogAttrs(ctx, slog.LevelDebug, "repair.success",
					slog.String("op", repairOp),
					slog.Int("attempt", attemptNum),
					slog.Int64("latency_ms", time.Since(attemptStart).Milliseconds()))
			}
			return ast, nil
		}
		if debug {
			r.logger.LogAttrs(ctx, slog.LevelDebug, "repair.failure",
				slog.String("op", repairOp),
				slog.Int("attempt", attemptNum),
				slog.String("reason", firstReason(errorsList)))
		}
	}
	if debug {
		r.logger.LogAttrs(ctx, slog.LevelDebug, "repair.exhausted",
			slog.String("op", repairOp),
			slog.Int("attempts_total", r.maxRetries),
			slog.String("final_reason", firstReason(errorsList)))
	}
	return nil, fmt.Errorf("%w: %d attempts exhausted", ErrRepairExhausted, r.maxRetries)
}

// Metrics returns (attempts, succeeded). attempts is the total repair
// invocations across all Translate calls; succeeded is the subset where
// the repaired AST passed validation. Feeds Story v3-13 recovery-rate
// metric.
func (r *Repairer) Metrics() (int64, int64) {
	return r.attempts.Load(), r.succeeded.Load()
}
