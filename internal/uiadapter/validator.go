package uiadapter

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"sort"
	"strings"
)

// Node type sentinels (spec §3.2 enumeration).
const (
	markdownNodeType       = "markdown"
	hintNodeType           = "hint"
	summaryNodeType        = "summary"
	codeNodeType           = "code"
	tableNodeType          = "table"
	decisionGroupNodeType  = "decision_group"
	widgetTypeChoice       = "choice"
	widgetTypeMulti        = "multi"
	maxResponseKeyLen      = 64
	maxDecisionGroupsCount = 8
	maxTotalNodes          = 32
	maxSerializedASTBytes  = 6 * 1024
)

// op-attr constants for validator.* records. The bare "validator" value is
// used by the start/rule.fail/done records (AC-5.4 contract); the dotted
// "validator.scan" form on the always-firing scan record satisfies AC-5.8's
// per-file emission coverage.
const (
	validatorOp     = "validator"
	validatorScanOp = "validator.scan"
)

var (
	// urlPattern is a lax RFC-3986-style http(s) URL match — enough for the
	// preservation check, deliberately not a full grammar.
	urlPattern = regexp.MustCompile(`https?://[^\s"'<>\])}]+`)

	// codeFencePattern captures fenced code blocks (```...```) across newlines.
	codeFencePattern = regexp.MustCompile("(?s)```[^`]*```")
)

// Validate applies spec §3.3 rules (1)–(8) in declaration order, mutating ast
// in place where possible. Returns a slice of reason strings in the order
// each rule first fires. Terminal reasons — "required_dropped" and "oversize"
// — short-circuit the pipeline so the caller can substitute a full fallback
// AST instead of emitting a partially-valid tree. logger may be nil;
// nilSafeLogger normalises it so any future story can emit telemetry without
// an inline guard.
//
// Story 5: emits validator.start at entry, validator.rule.fail per offending
// rule (rule_id + bounded reason enum), and validator.done at return with
// fail_count + sorted distinct reasons. Reasons are bounded enums — never
// user content — so logging them is safe under §14 sanitize discipline.
func Validate(ast *UIAST, raw string, logger *slog.Logger) []string {
	lg := nilSafeLogger(logger)
	ctx := context.Background()
	if lg.Enabled(ctx, slog.LevelDebug) {
		lg.LogAttrs(ctx, slog.LevelDebug, "validator.start",
			slog.String("op", validatorOp),
			slog.Int("node_count", len(ast.Nodes)),
			slog.Int("bytes_in", len(raw)),
		)
		// Per-call scan record carries a dotted op so AC-5.8's emission-
		// coverage scan reliably sees a `validator.*` op even when no rule
		// fires.
		lg.LogAttrs(ctx, slog.LevelDebug, "validator.scan",
			slog.String("op", validatorScanOp),
			slog.Int("node_count", len(ast.Nodes)),
		)
	}
	var reasons []string

	reasons = applyRule1(ast, reasons, lg)
	reasons = applyRules2And3(ast, reasons, lg)
	reasons = applyRule4(ast, reasons, lg)
	reasons = applyRule5(ast, reasons, lg)
	if r := applyRule6(ast, reasons, lg); terminalReason(r) {
		emitValidatorDone(lg, r)
		return r
	} else {
		reasons = r
	}
	if r := applyRule7(ast, reasons, lg); terminalReason(r) {
		emitValidatorDone(lg, r)
		return r
	} else {
		reasons = r
	}
	if r := applyRule8(ast, reasons, lg); terminalReason(r) {
		emitValidatorDone(lg, r)
		return r
	} else {
		reasons = r
	}
	emitValidatorDone(lg, reasons)
	return reasons
}

// emitValidatorDone writes the aggregate validator.done record with a sorted,
// distinct reasons slice. Centralised so every return path through Validate
// emits exactly one done record with consistent attributes.
func emitValidatorDone(logger *slog.Logger, reasons []string) {
	ctx := context.Background()
	if !logger.Enabled(ctx, slog.LevelDebug) {
		return
	}
	distinct := dedupeAndSort(reasons)
	logger.LogAttrs(ctx, slog.LevelDebug, "validator.done",
		slog.String("op", validatorOp),
		slog.Int("fail_count", len(reasons)),
		slog.Any("reasons", distinct),
	)
}

// emitRuleFail emits a validator.rule.fail record carrying the rule id and
// the bounded reason enum. Called from each applyRule* helper when a new
// reason is appended.
func emitRuleFail(logger *slog.Logger, ruleID int, reason string) {
	ctx := context.Background()
	if !logger.Enabled(ctx, slog.LevelDebug) {
		return
	}
	logger.LogAttrs(ctx, slog.LevelDebug, "validator.rule.fail",
		slog.String("op", validatorOp),
		slog.Int("rule_id", ruleID),
		slog.String("reason", reason),
	)
}

// dedupeAndSort returns a sorted slice of the distinct reason enum strings.
// Pre-allocates against the expected upper bound (the eight rule reasons)
// so the hot path stays cheap.
func dedupeAndSort(reasons []string) []string {
	if len(reasons) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(reasons))
	out := make([]string, 0, len(reasons))
	for _, r := range reasons {
		if _, ok := seen[r]; ok {
			continue
		}
		seen[r] = struct{}{}
		out = append(out, r)
	}
	sort.Strings(out)
	return out
}

// terminalReason reports whether the slice contains a short-circuit sentinel
// that forces the caller to substitute a full fallback AST (spec §3.3 rules
// 6/7/8).
func terminalReason(reasons []string) bool {
	for _, r := range reasons {
		if r == "required_dropped" || r == "oversize" || r == "marshal_error" {
			return true
		}
	}
	return false
}

// Rule 1: unknown node type rewritten IN PLACE to markdown. Content preserved
// verbatim when present; otherwise the original node is marshalled and the
// JSON becomes the markdown body.
func applyRule1(ast *UIAST, reasons []string, logger *slog.Logger) []string {
	fired := false
	for i := range ast.Nodes {
		n := &ast.Nodes[i]
		if knownNodeType(n.Type) {
			continue
		}
		content := n.Content
		if content == "" {
			if blob, err := json.Marshal(*n); err == nil {
				content = string(blob)
			}
		}
		*n = UINode{Type: markdownNodeType, Content: content}
		fired = true
	}
	if fired {
		emitRuleFail(logger, 1, "unknown_type")
		return append(reasons, "unknown_type")
	}
	return reasons
}

// Rules 2 + 3: drop decision_groups that cannot render, emitting one reason
// per offending node in declaration order (spec §3.3: "Validator walks nodes
// in declaration order"). Rule 3 (nil Widget) is checked before rule 2 (empty
// Options) within a single node so a nil widget never tries to read Options.
func applyRules2And3(ast *UIAST, reasons []string, logger *slog.Logger) []string {
	kept := ast.Nodes[:0]
	for _, n := range ast.Nodes {
		if n.Type == decisionGroupNodeType {
			if n.Widget == nil {
				emitRuleFail(logger, 3, "no_widget")
				reasons = append(reasons, "no_widget")
				continue
			}
			if (n.Widget.Type == widgetTypeChoice || n.Widget.Type == widgetTypeMulti) && len(n.Widget.Options) == 0 {
				emitRuleFail(logger, 2, "empty_options")
				reasons = append(reasons, "empty_options")
				continue
			}
		}
		kept = append(kept, n)
	}
	ast.Nodes = kept
	return reasons
}

// Rule 4: duplicate response_key suffixed -2, -3, … in first-appearance order.
// Only walks decision_group nodes — rule 1 rewrites strip response_key off
// anything else by reconstructing the node.
func applyRule4(ast *UIAST, reasons []string, logger *slog.Logger) []string {
	seen := map[string]int{}
	fired := false
	for i := range ast.Nodes {
		n := &ast.Nodes[i]
		if n.Type != decisionGroupNodeType || n.ResponseKey == "" {
			continue
		}
		seen[n.ResponseKey]++
		if seen[n.ResponseKey] > 1 {
			n.ResponseKey = fmt.Sprintf("%s-%d", n.ResponseKey, seen[n.ResponseKey])
			fired = true
		}
	}
	if fired {
		emitRuleFail(logger, 4, "dup_key")
		return append(reasons, "dup_key")
	}
	return reasons
}

// Rule 5: response_key > 64 characters truncated to the first 64 (prefix
// preserved).
func applyRule5(ast *UIAST, reasons []string, logger *slog.Logger) []string {
	fired := false
	for i := range ast.Nodes {
		n := &ast.Nodes[i]
		if n.Type != decisionGroupNodeType {
			continue
		}
		if len(n.ResponseKey) > maxResponseKeyLen {
			n.ResponseKey = n.ResponseKey[:maxResponseKeyLen]
			fired = true
		}
	}
	if fired {
		emitRuleFail(logger, 5, "key_truncated")
		return append(reasons, "key_truncated")
	}
	return reasons
}

// Rule 6: at most 8 decision_groups. If any dropped group had Required=true,
// return the terminal sentinel so the caller substitutes a full fallback —
// users must never silently lose a required input.
func applyRule6(ast *UIAST, reasons []string, logger *slog.Logger) []string {
	groupCount := 0
	overflow := false
	kept := ast.Nodes[:0]
	for _, n := range ast.Nodes {
		if n.Type == decisionGroupNodeType {
			groupCount++
			if groupCount > maxDecisionGroupsCount {
				overflow = true
				if n.Required {
					emitRuleFail(logger, 6, "required_dropped")
					return []string{"required_dropped"}
				}
				continue
			}
		}
		kept = append(kept, n)
	}
	ast.Nodes = kept
	if overflow {
		emitRuleFail(logger, 6, "count_capped")
		reasons = append(reasons, "count_capped")
	}
	return reasons
}

// Rule 7: at most 32 total nodes. Same required-drop guard as rule 6.
func applyRule7(ast *UIAST, reasons []string, logger *slog.Logger) []string {
	if len(ast.Nodes) <= maxTotalNodes {
		return reasons
	}
	for _, n := range ast.Nodes[maxTotalNodes:] {
		if n.Type == decisionGroupNodeType && n.Required {
			emitRuleFail(logger, 7, "required_dropped")
			return []string{"required_dropped"}
		}
	}
	ast.Nodes = ast.Nodes[:maxTotalNodes]
	emitRuleFail(logger, 7, "node_capped")
	return append(reasons, "node_capped")
}

// Rule 8: serialized AST size > 6 KiB forces a full fallback. Must run after
// rules (1)–(7) so the size check reflects all prior mutations. A marshal
// error is genuinely abnormal (UIAST is well-defined); surface it with its
// own sentinel so operators can distinguish it from legitimate oversize.
func applyRule8(ast *UIAST, reasons []string, logger *slog.Logger) []string {
	blob, err := json.Marshal(ast)
	if err != nil {
		emitRuleFail(logger, 8, "marshal_error")
		return []string{"marshal_error"}
	}
	if len(blob) > maxSerializedASTBytes {
		emitRuleFail(logger, 8, "oversize")
		return []string{"oversize"}
	}
	return reasons
}

// knownNodeType enumerates the six §3.2 node shapes the UI layer can render.
// Any other type is rewritten by rule 1.
func knownNodeType(t string) bool {
	switch t {
	case markdownNodeType, hintNodeType, summaryNodeType, codeNodeType, tableNodeType, decisionGroupNodeType:
		return true
	}
	return false
}

// contentPreserved implements §4.7.4 / §7.2: a raw URL or fenced code block
// present in the source capture must appear in at least one node's content.
// Numbered list entries are deliberately NOT checked — converting
// "1. stdout / 2. file / 3. both" into a choice widget is the feature's
// primary job and must never flag Untrusted.
func contentPreserved(raw string, ast *UIAST) bool {
	urls := urlPattern.FindAllString(raw, -1)
	codes := codeFencePattern.FindAllString(raw, -1)
	if len(urls) == 0 && len(codes) == 0 {
		return true
	}

	haystack := collectRenderedText(ast)
	for _, u := range urls {
		if !strings.Contains(haystack, u) {
			return false
		}
	}
	for _, c := range codes {
		if !strings.Contains(haystack, c) {
			return false
		}
	}
	return true
}

// collectRenderedText flattens every user-visible string in the AST so the
// preservation check can scan for URL / code-block presence without caring
// which node variant carries the content.
func collectRenderedText(ast *UIAST) string {
	var b strings.Builder
	for _, n := range ast.Nodes {
		b.WriteString(n.Content)
		b.WriteByte('\n')
		b.WriteString(n.Heading)
		b.WriteByte('\n')
		b.WriteString(n.Prompt)
		b.WriteByte('\n')
		b.WriteString(n.Help)
		b.WriteByte('\n')
		for _, bullet := range n.Bullets {
			b.WriteString(bullet)
			b.WriteByte('\n')
		}
		for _, col := range n.Columns {
			b.WriteString(col)
			b.WriteByte('\n')
		}
		for _, row := range n.Rows {
			for _, cell := range row {
				b.WriteString(cell)
				b.WriteByte('\n')
			}
		}
		if n.Widget != nil {
			for _, opt := range n.Widget.Options {
				b.WriteString(opt.Value)
				b.WriteByte('\n')
				b.WriteString(opt.Label)
				b.WriteByte('\n')
			}
			b.WriteString(n.Widget.Default)
			b.WriteByte('\n')
			b.WriteString(n.Widget.Placeholder)
			b.WriteByte('\n')
		}
	}
	return b.String()
}
