package uiadapter

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync/atomic"
)

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
// logger may be nil; nilSafeLogger normalises it so any future story can
// emit telemetry without an inline guard.
func BuildRepairPrompt(a RepairAttempt, logger *slog.Logger) string {
	_ = nilSafeLogger(logger)
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
		r.attempts.Add(1)
		attempt.PreviousOutput = body
		attempt.Errors = errorsList
		repairPrompt := BuildRepairPrompt(attempt, r.logger)
		next, nextBody, err := generate(ctx, repairPrompt)
		if err != nil {
			return nil, fmt.Errorf("uiadapter: repair attempt %d: %w", i+1, err)
		}
		ast = next
		body = nextBody
		errorsList = validate(ast)
		if len(errorsList) == 0 {
			r.succeeded.Add(1)
			return ast, nil
		}
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
