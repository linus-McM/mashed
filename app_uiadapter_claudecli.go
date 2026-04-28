package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"mashed/internal/uiadapter"
	"mashed/internal/uiadapter/backend/claudecli"
)

// claudeCLIAdapter satisfies uiadapter.Adapter by delegating to the
// claudecli backend client. Used when cfg.Backend == "claude-cli" so the
// modal's structured-UI translation runs against Claude (Haiku by default,
// for cost) instead of a local Ollama model — Ollama's small models (gemma3,
// gemma4) frequently mis-format the JSON envelope, blow the 30 s budget on
// cold loads, or emit empty AST nodes, leaving users in a raw JSON textarea.
type claudeCLIAdapter struct {
	client  *claudecli.Client
	logger  *slog.Logger
	timeout time.Duration
}

// newClaudeCLIAdapter builds the adapter. cfg.ClaudeCLIExtraFlags is
// expected to carry --model so every spawned `claude -p` call lands on the
// chosen Haiku/Sonnet variant; without it the user's current shell-default
// model wins (typically Opus, which is overkill for this translation).
func newClaudeCLIAdapter(cfg uiadapter.Config, logger *slog.Logger) uiadapter.Adapter {
	return &claudeCLIAdapter{
		client:  claudecli.NewClient(cfg),
		logger:  logger,
		timeout: time.Duration(cfg.TimeoutMs) * time.Millisecond,
	}
}

func (a *claudeCLIAdapter) Translate(ctx context.Context, raw, procID string) *uiadapter.UIAST {
	start := time.Now()
	sanitized, _ := uiadapter.SanitizeCapture(raw, a.logger)
	if a.timeout <= 0 {
		a.timeout = 30 * time.Second
	}
	ctxT, cancel := context.WithTimeout(ctx, a.timeout)
	defer cancel()

	// Use the full embedded prompt (covers free/choice/multi/approval) so
	// menu-shaped turns surface as decision_group radio buttons. The
	// `GenerateSingleShot` shortcut routes to the text-only stage prompt
	// which always degrades to a `free` widget regardless of source.
	ast, rawResp, err := a.client.TranslateWithFullPrompt(ctxT, sanitized)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		reason := "claude-cli:" + truncReason(err.Error())
		if a.logger != nil {
			// Persist the raw assistant text on failure so we can see
			// WHY parsing degraded — was the model emitting prose, an
			// unfenced JSON, an apology message, etc. Truncated head /
			// tail keep the log size bounded; full bodies go to a
			// per-failure debug file when MASHED_UIADAPTER_DUMP_RAW=1.
			head, tail := splitHeadTail(rawResp, 800, 800)
			a.logger.Warn("claude-cli.translate.error",
				"op", "claude-cli.translate",
				"proc_id", procID,
				"latency_ms", latency,
				"input_bytes", len(sanitized),
				"raw_bytes", len(rawResp),
				"raw_head", head,
				"raw_tail", tail,
				"reason", reason,
				"err", err.Error())
			dumpRawIfRequested(rawResp, procID, a.logger)
		}
		return uiadapter.FallbackAST(sanitized, reason, a.logger)
	}
	if ast == nil {
		if a.logger != nil {
			a.logger.Warn("claude-cli.translate.nil",
				"op", "claude-cli.translate",
				"proc_id", procID,
				"latency_ms", latency,
				"input_bytes", len(sanitized))
		}
		return uiadapter.FallbackAST(sanitized, "claude-cli:nil", a.logger)
	}
	if ast.GeneratedBy == "" {
		ast.GeneratedBy = fmt.Sprintf("claude-cli:%s", capModel(a.client))
	}
	if a.logger != nil {
		a.logger.Info("claude-cli.translate.ok",
			"op", "claude-cli.translate",
			"proc_id", procID,
			"latency_ms", latency,
			"input_bytes", len(sanitized),
			"generated_by", ast.GeneratedBy,
			"node_count", len(ast.Nodes),
			"fallback_answer_shape", ast.FallbackAnswerShape)
	}
	return ast
}

// truncReason caps an error string so a 4 KiB Anthropic body does not bloat
// the GeneratedBy field on the persisted PendingPrompt.
func truncReason(s string) string {
	const max = 60
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= max {
		return s
	}
	return s[:max]
}

// splitHeadTail returns (head, tail) of s capped at headN / tailN bytes.
// When s fits inside head + tail no overlap occurs — the head is returned
// verbatim and the tail is empty. Used to keep failure logs bounded while
// still showing both the model's leading prose AND any trailing JSON.
func splitHeadTail(s string, headN, tailN int) (string, string) {
	if len(s) <= headN+tailN {
		return s, ""
	}
	return s[:headN], s[len(s)-tailN:]
}

// dumpRawIfRequested writes the full assistant response to /tmp when
// MASHED_UIADAPTER_DUMP_RAW=1. Useful when the head/tail snippets in the
// log aren't enough — e.g. when the JSON envelope spans the middle of a
// long response. Filename includes timestamp + procID for easy correlation.
func dumpRawIfRequested(raw, procID string, logger *slog.Logger) {
	if os.Getenv("MASHED_UIADAPTER_DUMP_RAW") != "1" {
		return
	}
	stamp := time.Now().UTC().Format("20060102T150405Z")
	safeProcID := strings.NewReplacer("/", "-", " ", "-").Replace(procID)
	path := fmt.Sprintf("/tmp/mashed-uiadapter-raw-%s-%s.txt", stamp, safeProcID)
	if err := os.WriteFile(path, []byte(raw), 0o644); err != nil {
		if logger != nil {
			logger.Warn("claude-cli.translate.dump_failed", "op", "claude-cli.translate", "path", path, "err", err.Error())
		}
		return
	}
	if logger != nil {
		logger.Info("claude-cli.translate.dump", "op", "claude-cli.translate", "path", path, "bytes", len(raw))
	}
}

// capModel surfaces the configured model in GeneratedBy for telemetry. The
// claudecli.Client doesn't expose its config; capabilities is the closest
// public surface.
func capModel(c *claudecli.Client) string {
	caps := c.Capabilities()
	if caps.Model == "" {
		return "default"
	}
	return caps.Model
}
