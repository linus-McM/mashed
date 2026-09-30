package uiadapter

import (
	"context"
	"log/slog"
	"regexp"
	"strings"
	"sync/atomic"
	"time"
)

// fastpathOp is the canonical `op` attribute value for every fastpath log
// emission (Story 4 §14 sanitize discipline — closed enum).
const fastpathOp = "fastpath.classify"

// FastPathClassifier is the pre-LLM short-circuit for highly-repetitive
// Claude-Code turns (Plan §3 Story 2). A first-match-wins rule table maps
// common shapes (y/n prompts, numbered menus, press-enter acknowledgements,
// file-confirm prompts, single-line free-text questions) straight to a
// UIAST. Every hit is one fewer LLM call — zero latency, zero failure
// surface, zero cost on paid backends.
//
// Gated by Config.EnableFastPath (Story B). When disabled, Classify returns
// hit=false unconditionally so callers unconditionally fall through to the
// LLM pipeline.
type FastPathClassifier struct {
	enabled     bool
	rules       []fastRule
	hitsPerRule []atomic.Int64
	totalCalls  atomic.Int64
	totalHits   atomic.Int64
	logger      *slog.Logger
}

type fastRule struct {
	Name  string
	Match *regexp.Regexp
	Build func(raw string, matches []string) *UIAST
}

// NewFastPathClassifier builds the classifier with the v3.0 rule set. The
// Config.EnableFastPath flag determines whether Classify short-circuits.
// logger may be nil; nilSafeLogger normalises it so the field is always
// usable.
func NewFastPathClassifier(enabled bool, logger *slog.Logger) *FastPathClassifier {
	rules := defaultFastRules()
	return &FastPathClassifier{
		enabled:     enabled,
		rules:       rules,
		hitsPerRule: make([]atomic.Int64, len(rules)),
		logger:      nilSafeLogger(logger),
	}
}

// Classify tries each rule in order. First match wins.
func (f *FastPathClassifier) Classify(raw string) (hit bool, rule string, ast *UIAST) {
	f.totalCalls.Add(1)
	ctx := context.Background()
	if !f.enabled {
		if f.logger.Enabled(ctx, slog.LevelDebug) {
			f.logger.LogAttrs(ctx, slog.LevelDebug, "fastpath.classify.disabled",
				slog.String("op", fastpathOp))
		}
		return false, "", nil
	}
	if raw == "" {
		return false, "", nil
	}
	debug := f.logger.Enabled(ctx, slog.LevelDebug)
	var start time.Time
	if debug {
		start = time.Now()
		f.logger.LogAttrs(ctx, slog.LevelDebug, "fastpath.classify.start",
			slog.String("op", fastpathOp),
			slog.Int("bytes_in", len(raw)),
			slog.Bool("enabled", f.enabled))
	}
	for i, r := range f.rules {
		if m := r.Match.FindStringSubmatch(raw); m != nil {
			built := r.Build(raw, m)
			if built == nil {
				// Some rules (numbered-menu) veto via nil Build result when
				// a post-match invariant fails — treat as a miss, try next.
				continue
			}
			f.hitsPerRule[i].Add(1)
			f.totalHits.Add(1)
			built.GeneratedBy = "fastpath:" + r.Name
			if debug {
				f.logger.LogAttrs(ctx, slog.LevelDebug, "fastpath.classify.hit",
					slog.String("op", fastpathOp),
					slog.String("rule", r.Name),
					slog.Int64("latency_ms", time.Since(start).Milliseconds()))
			}
			return true, r.Name, built
		}
	}
	if debug {
		f.logger.LogAttrs(ctx, slog.LevelDebug, "fastpath.classify.skip",
			slog.String("op", fastpathOp),
			slog.Int64("latency_ms", time.Since(start).Milliseconds()))
	}
	return false, "", nil
}

// HitRate is the running ratio of hits to calls. Feeds the slog
// `fast_path_hit_rate` attribute (AC-2.2).
func (f *FastPathClassifier) HitRate() float64 {
	calls := f.totalCalls.Load()
	if calls == 0 {
		return 0
	}
	return float64(f.totalHits.Load()) / float64(calls)
}

// HitsPerRule is a snapshot map ruleName→hits.
func (f *FastPathClassifier) HitsPerRule() map[string]int64 {
	out := make(map[string]int64, len(f.rules))
	for i, r := range f.rules {
		out[r.Name] = f.hitsPerRule[i].Load()
	}
	return out
}

// defaultFastRules returns the v3.0 minimum rule set. Order is material —
// first-match-wins; narrower rules precede broader ones.
func defaultFastRules() []fastRule {
	return []fastRule{
		{
			Name:  "yn-prompt",
			Match: regexp.MustCompile(`(?i)(\by/n\b|\[Y/n\]|\[y/N\]|\(yes/no\))`),
			Build: buildYN,
		},
		{
			Name:  "press-enter",
			Match: regexp.MustCompile(`(?i)press\s+(any\s+key|enter)\s+to\s+(continue|exit)`),
			Build: buildPressEnter,
		},
		{
			Name:  "file-confirm",
			Match: regexp.MustCompile(`(?im)^\s*(apply|write|save)\b.*\?$`),
			Build: buildFileConfirm,
		},
		{
			Name:  "numbered-menu",
			Match: regexp.MustCompile(`(?m)^\s*\d+[\.\)]\s+\S`),
			Build: buildNumberedMenu,
		},
		{
			Name:  "free-text-prompt",
			Match: regexp.MustCompile(`\A([^\n]*\?)\s*\z`),
			Build: buildFreeText,
		},
	}
}

func buildYN(raw string, _ []string) *UIAST {
	prompt := firstNonEmptyLine(raw)
	return &UIAST{
		Version:     "1",
		TurnSummary: truncate(prompt, 120),
		Nodes: []UINode{{
			Type:        "decision_group",
			Prompt:      prompt,
			ResponseKey: "answer",
			Widget:      &WidgetNode{Type: "approval"},
		}},
		FallbackAnswerShape: "approval",
	}
}

func buildPressEnter(raw string, _ []string) *UIAST {
	return &UIAST{
		Version:     "1",
		TurnSummary: "press enter to continue",
		Nodes: []UINode{{
			Type:        "decision_group",
			Prompt:      firstNonEmptyLine(raw),
			ResponseKey: "ack",
			Widget:      &WidgetNode{Type: "approval"},
		}},
		FallbackAnswerShape: "approval",
	}
}

func buildFileConfirm(raw string, _ []string) *UIAST {
	return &UIAST{
		Version:     "1",
		TurnSummary: "confirm file action",
		Nodes: []UINode{{
			Type:        "decision_group",
			Prompt:      firstNonEmptyLine(raw),
			ResponseKey: "confirm",
			Widget:      &WidgetNode{Type: "approval"},
		}},
		FallbackAnswerShape: "approval",
	}
}

var reMenuLine = regexp.MustCompile(`^\s*(\d+)[\.\)]\s+(.+)$`)

func buildNumberedMenu(raw string, _ []string) *UIAST {
	var options []WidgetOption
	var prompt string
	for _, line := range strings.Split(raw, "\n") {
		if m := reMenuLine.FindStringSubmatch(line); m != nil {
			options = append(options, WidgetOption{
				Value: m[1],
				Label: strings.TrimSpace(m[2]),
			})
			continue
		}
		if prompt == "" {
			if trim := strings.TrimSpace(line); trim != "" {
				prompt = trim
			}
		}
	}
	// Require ≥2 consecutive (or at least ≥2 total) numbered lines; the
	// single-item case is ambiguous and better handled by the free-text or
	// LLM pipeline.
	if len(options) < 2 {
		return nil
	}
	if prompt == "" {
		prompt = "select one"
	}
	return &UIAST{
		Version:     "1",
		TurnSummary: truncate(prompt, 120),
		Nodes: []UINode{{
			Type:        "decision_group",
			Prompt:      prompt,
			ResponseKey: "selection",
			Widget:      &WidgetNode{Type: "choice", Options: options},
		}},
		FallbackAnswerShape: "choice",
	}
}

func buildFreeText(raw string, _ []string) *UIAST {
	prompt := strings.TrimSpace(raw)
	return &UIAST{
		Version:     "1",
		TurnSummary: truncate(prompt, 120),
		Nodes: []UINode{{
			Type:        "decision_group",
			Prompt:      prompt,
			ResponseKey: "answer",
			Widget:      &WidgetNode{Type: "free"},
		}},
		FallbackAnswerShape: "free",
	}
}

func firstNonEmptyLine(raw string) string {
	for _, line := range strings.Split(raw, "\n") {
		trim := strings.TrimSpace(line)
		if trim != "" {
			return trim
		}
	}
	return ""
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
