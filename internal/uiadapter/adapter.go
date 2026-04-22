package uiadapter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"
)

// Adapter translates a raw Claude-turn capture into a Mashed UI AST tree.
// Translate never returns nil: every failure mode degrades to a
// fallback AST with GeneratedBy="fallback:<reason>". A mid-call context
// cancellation also populates the returned Diagnostics.CancelReason so the
// executor's round loop can distinguish user-initiated aborts.
type Adapter interface {
	Translate(ctx context.Context, raw, procID string) *UIAST
}

// Config is the adapter-wide construction knob. Enabled collapses
// UIAdapterEnabled && OllamaEnabled at the call-site — the adapter only
// sees a single boolean so callers cannot smuggle in a partial gate.
// Deterministic pins options.temperature=0 on every Ollama request — set by
// the offline eval harness only (Story U9 §4.5 / §Risks); production callers
// leave it false so users see the model's natural sampling.
type Config struct {
	Enabled       bool
	Model         string
	TimeoutMs     int
	MaxInflight   int
	Deterministic bool
}

type defaultAdapter struct {
	client        *Client
	model         string
	timeout       time.Duration
	sem           semaphore
	logger        *slog.Logger
	deterministic bool
}

// disabledAdapter short-circuits every Translate call to a fallback:disabled
// AST. Kept as its own type so the call-site branch happens at construction
// time — no nil checks, no per-call flag inspection.
type disabledAdapter struct{}

// NewDefault constructs the adapter. When Config.Enabled is false (the
// combined UIAdapterEnabled && OllamaEnabled gate), a disabledAdapter is
// returned that never touches the HTTP client. Otherwise the full default
// adapter with a bounded-concurrency semaphore is produced.
func NewDefault(cfg Config, logger *slog.Logger) Adapter {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	if !cfg.Enabled {
		return &disabledAdapter{}
	}
	if cfg.MaxInflight <= 0 {
		cfg.MaxInflight = 1
	}
	return &defaultAdapter{
		client:        NewClient(ClientConfig{TimeoutMs: cfg.TimeoutMs}),
		model:         cfg.Model,
		timeout:       time.Duration(cfg.TimeoutMs) * time.Millisecond,
		sem:           newSemaphore(cfg.MaxInflight),
		logger:        logger,
		deterministic: cfg.Deterministic,
	}
}

// Translate on disabledAdapter is a pure function — no HTTP client exists so
// AC-9's "no HTTP call on disabled" is a structural guarantee, not a runtime
// check.
func (*disabledAdapter) Translate(_ context.Context, raw, _ string) *UIAST {
	return FallbackAST(raw, "disabled")
}

// Translate executes the full §4.7 reliability pipeline. Order of operations:
// sem acquire (with ctx.Done fallback → saturated), HTTP call under the
// adapter's TimeoutMs, JSON decode, §3.3 validator, content-preservation
// check, telemetry. Every failure mode maps to a fallback AST per §4.8.
func (a *defaultAdapter) Translate(ctx context.Context, raw, procID string) *UIAST {
	start := time.Now()

	if !a.sem.acquire(ctx) {
		return a.emitFallback(raw, "saturated", ctx, start, 0)
	}
	defer a.sem.release()

	ctxT, cancel := context.WithTimeout(ctx, a.timeout)
	defer cancel()

	body, err := a.chat(ctxT, raw)
	if err != nil {
		return a.emitFallback(raw, classifyChatErr(err), ctx, start, 0)
	}

	ast := &UIAST{}
	if decErr := json.Unmarshal([]byte(body), ast); decErr != nil {
		return a.emitFallback(raw, "validation:malformed", ctx, start, len(body))
	}

	reasons := Validate(ast, raw)
	for _, r := range reasons {
		if r == "required_dropped" || r == "oversize" || r == "marshal_error" {
			return a.emitFallback(raw, "validation:"+r, ctx, start, len(body))
		}
	}

	if !contentPreserved(raw, ast) {
		ast.Diagnostics.Untrusted = true
	}
	ast.Diagnostics.FallbackReasons = append(ast.Diagnostics.FallbackReasons, reasons...)
	a.stampSuccessMetadata(ast, raw, body, start)

	a.logTelemetry(slog.LevelInfo, procID, len(raw), len(body), start, reasons, ast.Diagnostics.Untrusted)
	return ast
}

// chat dispatches to ChatDeterministic when the harness pinned temperature=0,
// otherwise to plain Chat so production keeps Ollama's natural sampling.
func (a *defaultAdapter) chat(ctx context.Context, raw string) (string, error) {
	if a.deterministic {
		return a.client.ChatDeterministic(ctx, a.model, SystemPrompt(), raw)
	}
	return a.client.Chat(ctx, a.model, SystemPrompt(), raw)
}

// stampSuccessMetadata fills in envelope + diagnostics fields the model is
// allowed to omit. Version defaults to "1", GeneratedBy to "ollama:<model>",
// GeneratedAt to the current Unix second.
func (a *defaultAdapter) stampSuccessMetadata(ast *UIAST, raw, body string, start time.Time) {
	if ast.Version == "" {
		ast.Version = "1"
	}
	if ast.GeneratedBy == "" {
		ast.GeneratedBy = "ollama:" + a.model
	}
	if ast.GeneratedAt == 0 {
		ast.GeneratedAt = time.Now().Unix()
	}
	ast.Diagnostics.InputBytes = len(raw)
	ast.Diagnostics.OutputBytes = len(body)
	ast.Diagnostics.LatencyMs = int(time.Since(start).Milliseconds())
}

// emitFallback builds the fallback AST for reason, tags Diagnostics with
// byte counts + latency, and writes the telemetry line at Warn level.
// Sets CancelReason only on "canceled" so the executor can differentiate
// user-initiated aborts from other failure paths.
func (a *defaultAdapter) emitFallback(raw, reason string, ctx context.Context, start time.Time, outBytes int) *UIAST {
	ast := FallbackAST(raw, reason)
	ast.Diagnostics.InputBytes = len(raw)
	ast.Diagnostics.OutputBytes = outBytes
	ast.Diagnostics.LatencyMs = int(time.Since(start).Milliseconds())
	if reason == "canceled" {
		if ctxErr := ctx.Err(); ctxErr != nil {
			ast.Diagnostics.CancelReason = ctxErr.Error()
		}
	}
	a.logTelemetry(slog.LevelWarn, reason, len(raw), outBytes, start, []string{reason}, false)
	return ast
}

// logTelemetry emits exactly one structured line per Translate call (§4.7.7).
// byte counts only — raw content never reaches the logger.
func (a *defaultAdapter) logTelemetry(level slog.Level, tag string, inBytes, outBytes int, start time.Time, reasons []string, untrusted bool) {
	a.logger.LogAttrs(context.Background(), level, "uiadapter.translate",
		slog.String("op", "translate"),
		slog.String("model", a.model),
		slog.String("prompt_version", promptVersion),
		slog.String("tag", tag),
		slog.Int("input_bytes", inBytes),
		slog.Int("output_bytes", outBytes),
		slog.Int("latency_ms", int(time.Since(start).Milliseconds())),
		slog.String("validation", strings.Join(reasons, ",")),
		slog.Bool("untrusted", untrusted),
	)
}

// classifyChatErr maps a Client.Chat error into the §4.8 reason string.
// Canceled beats DeadlineExceeded because an externally-canceled parent
// context propagates as context.Canceled even through a
// context.WithTimeout child.
func classifyChatErr(err error) string {
	if errors.Is(err, context.Canceled) {
		return "canceled"
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "timeout"
	}
	var httpErr *HTTPStatusError
	if errors.As(err, &httpErr) {
		return fmt.Sprintf("server:%d", httpErr.StatusCode)
	}
	return "unreachable"
}
