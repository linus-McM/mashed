package uiadapter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"regexp"
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

// Config is the adapter-wide construction knob. Plan v3 §3 Story B
// collapses every per-backend knob the v3 pipeline introduces into one
// dependency-injected surface — no package-level constants (§6.5 "DI via
// Config"). Zero-valued fields are backfilled by DefaultConfig / the
// mergeWithDefaults helper inside NewDefault.
//
// Legacy fields (Enabled, MaxInflight, Deterministic) are retained for
// backward compatibility with callers in app.go and pre-v3 tests; new code
// should prefer the v3 knobs (Backend, RouterPolicy, CacheCapacity, etc.).
type Config struct {
	// Legacy gates (Story ui-ast-U2). Enabled collapses
	// UIAdapterEnabled && OllamaEnabled at the call-site. Deterministic
	// pins options.temperature=0 on every Ollama request — set by the
	// offline eval harness only.
	Enabled       bool
	MaxInflight   int
	Deterministic bool

	// Backend selection (plan §3 Story B).
	Backend       string   // "ollama" | "claude-api" | "claude-cli" | "router"
	RouterPolicy  string   // "claude-first" | "ollama-first" | "claude-for-hard" | "local-only" | "claude-only" | "cost-aware" | "privacy-strict"
	FallbackOrder []string // cross-backend fallback chain (Story v3-11 / v3-16)

	// Ollama.
	OllamaEndpoint      string // default localhost:11434 (see DefaultConfig)
	Model               string // default "gemma3:4b"
	AllowUnvettedModels bool   // Story v3-12 override
	NumCtx              int    // default 8192 (Story v3-03)
	KeepAlive           string // default "30m" (Story v3-07)
	LooseFormat         bool   // rollback to format:"json" string (Story v3-06)

	// Claude API.
	AnthropicAPIKeyEnv string // default "ANTHROPIC_API_KEY"
	ClaudeModelPrimary string // default "claude-haiku-4-5"
	ClaudeModelHard    string // default "claude-sonnet-4-6"
	ClaudeMaxTokens    int    // default 2048
	AnthropicVersion   string // header pin; default "2023-06-01"
	PromptCacheTTL     string // "5m" | "1h" | "off"; default "5m"

	// Claude CLI.
	ClaudeCLIBinary     string
	ClaudeCLIExtraFlags []string

	// Cost / rate accountant (Story v3-11b).
	UsdBudgetPerSession float64
	RPMSoftLimit        int
	TPMSoftLimit        int

	// Sampling (Story v3-09).
	Temperature float32
	Seed        int64

	// Timeouts. TimeoutMs default 5000 (plan §3 Story B, was 2900 in legacy).
	TimeoutMs       int
	WarmUpTimeoutMs int

	// Runtime behaviour.
	CacheCapacity        int     // default 1024; 0 disables (Story v3-04)
	EnableSemanticCache  bool    // default false
	EnableFastPath       bool    // default true (Story v3-02)
	EnableSpotlighting   bool    // default true (Story v3-08)
	RepairMaxRetries     int     // default 1 (Ollama/Haiku), 0 (Sonnet) — Story v3-10
	BreakerFailThreshold int     // default 3 (Story v3-11)
	BreakerResetMs       int     // default 30000
	ShadowSampleRate     float64 // default 0.05 (Story v3-13)

	// Privacy / policy.
	PrivacyPatterns []*regexp.Regexp

	// Lifecycle (Story v3-17).
	DisableHealthTicker bool

	// Streaming (Story v3-14, default off in v3.0).
	Streaming bool
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
