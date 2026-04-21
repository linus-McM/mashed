package uiadapter

import (
	"encoding/json"
	"fmt"
	"regexp"
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
// AST instead of emitting a partially-valid tree.
func Validate(ast *UIAST, raw string) []string {
	var reasons []string

	reasons = applyRule1(ast, reasons)
	reasons = applyRules2And3(ast, reasons)
	reasons = applyRule4(ast, reasons)
	reasons = applyRule5(ast, reasons)
	if r := applyRule6(ast, reasons); terminalReason(r) {
		return r
	} else {
		reasons = r
	}
	if r := applyRule7(ast, reasons); terminalReason(r) {
		return r
	} else {
		reasons = r
	}
	if r := applyRule8(ast, reasons); terminalReason(r) {
		return r
	} else {
		reasons = r
	}
	return reasons
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
func applyRule1(ast *UIAST, reasons []string) []string {
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
		return append(reasons, "unknown_type")
	}
	return reasons
}

// Rules 2 + 3: drop decision_groups that cannot render, emitting one reason
// per offending node in declaration order (spec §3.3: "Validator walks nodes
// in declaration order"). Rule 3 (nil Widget) is checked before rule 2 (empty
// Options) within a single node so a nil widget never tries to read Options.
func applyRules2And3(ast *UIAST, reasons []string) []string {
	kept := ast.Nodes[:0]
	for _, n := range ast.Nodes {
		if n.Type == decisionGroupNodeType {
			if n.Widget == nil {
				reasons = append(reasons, "no_widget")
				continue
			}
			if (n.Widget.Type == widgetTypeChoice || n.Widget.Type == widgetTypeMulti) && len(n.Widget.Options) == 0 {
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
func applyRule4(ast *UIAST, reasons []string) []string {
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
		return append(reasons, "dup_key")
	}
	return reasons
}

// Rule 5: response_key > 64 characters truncated to the first 64 (prefix
// preserved).
func applyRule5(ast *UIAST, reasons []string) []string {
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
		return append(reasons, "key_truncated")
	}
	return reasons
}

// Rule 6: at most 8 decision_groups. If any dropped group had Required=true,
// return the terminal sentinel so the caller substitutes a full fallback —
// users must never silently lose a required input.
func applyRule6(ast *UIAST, reasons []string) []string {
	groupCount := 0
	overflow := false
	kept := ast.Nodes[:0]
	for _, n := range ast.Nodes {
		if n.Type == decisionGroupNodeType {
			groupCount++
			if groupCount > maxDecisionGroupsCount {
				overflow = true
				if n.Required {
					return []string{"required_dropped"}
				}
				continue
			}
		}
		kept = append(kept, n)
	}
	ast.Nodes = kept
	if overflow {
		reasons = append(reasons, "count_capped")
	}
	return reasons
}

// Rule 7: at most 32 total nodes. Same required-drop guard as rule 6.
func applyRule7(ast *UIAST, reasons []string) []string {
	if len(ast.Nodes) <= maxTotalNodes {
		return reasons
	}
	for _, n := range ast.Nodes[maxTotalNodes:] {
		if n.Type == decisionGroupNodeType && n.Required {
			return []string{"required_dropped"}
		}
	}
	ast.Nodes = ast.Nodes[:maxTotalNodes]
	return append(reasons, "node_capped")
}

// Rule 8: serialized AST size > 6 KiB forces a full fallback. Must run after
// rules (1)–(7) so the size check reflects all prior mutations. A marshal
// error is genuinely abnormal (UIAST is well-defined); surface it with its
// own sentinel so operators can distinguish it from legitimate oversize.
func applyRule8(ast *UIAST, reasons []string) []string {
	blob, err := json.Marshal(ast)
	if err != nil {
		return []string{"marshal_error"}
	}
	if len(blob) > maxSerializedASTBytes {
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
