package uiadapter

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidate_Rule1_UnknownTypeBecomesMarkdown — §3.3 rule (1): unknown node
// type rewritten IN PLACE to markdown; never dropped; content preserved.
func TestValidate_Rule1_UnknownTypeBecomesMarkdown(t *testing.T) {
	t.Parallel()
	ast := &UIAST{
		Version: "1",
		Nodes:   []UINode{{Type: "snarkfish", Content: "preserved content"}},
	}
	reasons := Validate(ast, "raw", nil)

	require.Len(t, ast.Nodes, 1, "rule 1 rewrites in place; never drops")
	assert.Equal(t, "markdown", ast.Nodes[0].Type)
	assert.Equal(t, "preserved content", ast.Nodes[0].Content,
		"original content must survive the rewrite")
	assert.Contains(t, reasons, "unknown_type")
}

// TestValidate_Rule2_EmptyOptionsDropsGroup — §3.3 rule (2): empty options on
// choice/multi discards the enclosing decision_group (required or not).
func TestValidate_Rule2_EmptyOptionsDropsGroup(t *testing.T) {
	t.Parallel()
	t.Run("choice", func(t *testing.T) {
		ast := &UIAST{
			Version: "1",
			Nodes: []UINode{
				{Type: "decision_group", ResponseKey: "a",
					Widget: &WidgetNode{Type: "choice", Options: nil}},
			},
		}
		reasons := Validate(ast, "raw", nil)
		assert.Len(t, ast.Nodes, 0, "decision_group with empty choice options must be dropped")
		assert.Contains(t, reasons, "empty_options")
	})
	t.Run("multi", func(t *testing.T) {
		ast := &UIAST{
			Version: "1",
			Nodes: []UINode{
				{Type: "decision_group", ResponseKey: "a",
					Widget: &WidgetNode{Type: "multi", Options: nil}},
			},
		}
		reasons := Validate(ast, "raw", nil)
		assert.Len(t, ast.Nodes, 0, "decision_group with empty multi options must be dropped")
		assert.Contains(t, reasons, "empty_options")
	})
}

// TestValidate_Rule3_NoWidgetDropsGroup — §3.3 rule (3): decision_group with
// no widget is discarded.
func TestValidate_Rule3_NoWidgetDropsGroup(t *testing.T) {
	t.Parallel()
	ast := &UIAST{
		Version: "1",
		Nodes:   []UINode{{Type: "decision_group", ResponseKey: "a", Widget: nil}},
	}
	reasons := Validate(ast, "raw", nil)
	assert.Len(t, ast.Nodes, 0)
	assert.Contains(t, reasons, "no_widget")
}

// TestValidate_Rule4_DuplicateKeySuffix — §3.3 rule (4): duplicates suffixed
// -2, -3; first occurrence retains the bare key.
func TestValidate_Rule4_DuplicateKeySuffix(t *testing.T) {
	t.Parallel()

	// a / a → a / a-2
	pair := &UIAST{
		Version: "1",
		Nodes: []UINode{
			{Type: "decision_group", ResponseKey: "a", Widget: &WidgetNode{Type: "free"}},
			{Type: "decision_group", ResponseKey: "a", Widget: &WidgetNode{Type: "free"}},
		},
	}
	reasons := Validate(pair, "raw", nil)
	require.Len(t, pair.Nodes, 2)
	assert.Equal(t, "a", pair.Nodes[0].ResponseKey, "first occurrence keeps bare key")
	assert.Equal(t, "a-2", pair.Nodes[1].ResponseKey, "second occurrence suffixed -2")
	assert.Contains(t, reasons, "dup_key")

	// a / a / a → a / a-2 / a-3
	triple := &UIAST{
		Version: "1",
		Nodes: []UINode{
			{Type: "decision_group", ResponseKey: "a", Widget: &WidgetNode{Type: "free"}},
			{Type: "decision_group", ResponseKey: "a", Widget: &WidgetNode{Type: "free"}},
			{Type: "decision_group", ResponseKey: "a", Widget: &WidgetNode{Type: "free"}},
		},
	}
	Validate(triple, "raw", nil)
	require.Len(t, triple.Nodes, 3)
	assert.Equal(t, "a", triple.Nodes[0].ResponseKey)
	assert.Equal(t, "a-2", triple.Nodes[1].ResponseKey)
	assert.Equal(t, "a-3", triple.Nodes[2].ResponseKey)
}

// TestValidate_Rule5_LongKeyTruncated — §3.3 rule (5): response_key > 64
// truncated to first 64 chars preserving prefix.
func TestValidate_Rule5_LongKeyTruncated(t *testing.T) {
	t.Parallel()
	long := strings.Repeat("x", 100)
	ast := &UIAST{
		Version: "1",
		Nodes: []UINode{
			{Type: "decision_group", ResponseKey: long, Widget: &WidgetNode{Type: "free"}},
		},
	}
	reasons := Validate(ast, "raw", nil)
	require.Len(t, ast.Nodes, 1)
	assert.Len(t, ast.Nodes[0].ResponseKey, 64, "response_key must truncate to 64 chars")
	assert.Equal(t, strings.Repeat("x", 64), ast.Nodes[0].ResponseKey,
		"truncation must preserve prefix (drop suffix, not middle)")
	assert.Contains(t, reasons, "key_truncated")
}

// TestValidate_Rule6_CountCapOptionalDrop — §3.3 rule (6): > 8 groups, none
// required → keep first 8, no terminal fallback reason.
func TestValidate_Rule6_CountCapOptionalDrop(t *testing.T) {
	t.Parallel()
	var nodes []UINode
	for i := 0; i < 9; i++ {
		nodes = append(nodes, UINode{
			Type:        "decision_group",
			ResponseKey: fmt.Sprintf("k%d", i),
			Required:    false,
			Widget:      &WidgetNode{Type: "free"},
		})
	}
	ast := &UIAST{Version: "1", Nodes: nodes}
	reasons := Validate(ast, "raw", nil)

	groups := 0
	for _, n := range ast.Nodes {
		if n.Type == "decision_group" {
			groups++
		}
	}
	assert.Equal(t, 8, groups,
		"count > 8 with no required drops keeps the first 8 groups in declaration order")
	assert.Equal(t, "k0", ast.Nodes[0].ResponseKey, "first group retained")
	assert.Equal(t, "k7", ast.Nodes[7].ResponseKey, "eighth group retained")
	assert.NotContains(t, reasons, "required_dropped",
		"optional drop must NOT emit the terminal required_dropped reason")
}

// TestValidate_Rule6_CountCapRequiredDropFallback — §3.3 rule (6): dropping a
// required group → terminal reason "required_dropped" forces full fallback.
func TestValidate_Rule6_CountCapRequiredDropFallback(t *testing.T) {
	t.Parallel()
	var nodes []UINode
	for i := 0; i < 9; i++ {
		nodes = append(nodes, UINode{
			Type:        "decision_group",
			ResponseKey: fmt.Sprintf("k%d", i),
			Required:    i == 8, // 9th is required and will be dropped by the cap
			Widget:      &WidgetNode{Type: "free"},
		})
	}
	ast := &UIAST{Version: "1", Nodes: nodes}
	reasons := Validate(ast, "raw", nil)
	assert.Contains(t, reasons, "required_dropped",
		"dropping a required decision_group past the cap must emit the terminal sentinel")
}

// TestValidate_Rule7_NodeCapRequiredDropFallback — §3.3 rule (7): > 32 total
// nodes with a required group beyond the cap → terminal "required_dropped".
func TestValidate_Rule7_NodeCapRequiredDropFallback(t *testing.T) {
	t.Parallel()
	nodes := make([]UINode, 33)
	for i := 0; i < 32; i++ {
		nodes[i] = UINode{Type: "markdown", Content: "n"}
	}
	nodes[32] = UINode{
		Type:        "decision_group",
		ResponseKey: "zz",
		Required:    true,
		Widget:      &WidgetNode{Type: "free"},
	}
	ast := &UIAST{Version: "1", Nodes: nodes}
	reasons := Validate(ast, "raw", nil)
	assert.Contains(t, reasons, "required_dropped",
		"> 32 nodes with a required group in the dropped remainder must trigger full fallback")
}

// TestValidate_Rule8_OversizeFullFallback — §3.3 rule (8): serialized size >
// 6 KiB after rules (1)–(7) → terminal "oversize".
func TestValidate_Rule8_OversizeFullFallback(t *testing.T) {
	t.Parallel()
	huge := strings.Repeat("A", 7*1024)
	ast := &UIAST{
		Version: "1",
		Nodes:   []UINode{{Type: "markdown", Content: huge}},
	}
	reasons := Validate(ast, "raw", nil)
	assert.Contains(t, reasons, "oversize",
		"AST serialising above 6 KiB must emit the terminal oversize sentinel")
}

// TestValidate_RuleOrdering_Deterministic — §3.3 declaration-order contract:
// input triggering rules 1 + 4 + 5 emits ["unknown_type","dup_key","key_truncated"]
// in that exact first-appearance order.
func TestValidate_RuleOrdering_Deterministic(t *testing.T) {
	t.Parallel()
	longKey := strings.Repeat("x", 100)
	ast := &UIAST{
		Version: "1",
		Nodes: []UINode{
			{Type: "snarkfish", Content: "x"}, // rule 1
			{Type: "decision_group", ResponseKey: "dup",
				Widget: &WidgetNode{Type: "free"}}, // baseline
			{Type: "decision_group", ResponseKey: "dup",
				Widget: &WidgetNode{Type: "free"}}, // rule 4 fires
			{Type: "decision_group", ResponseKey: longKey,
				Widget: &WidgetNode{Type: "free"}}, // rule 5 fires
		},
	}
	reasons := Validate(ast, "raw", nil)

	want := map[string]bool{"unknown_type": true, "dup_key": true, "key_truncated": true}
	seen := []string{}
	already := map[string]bool{}
	for _, r := range reasons {
		if want[r] && !already[r] {
			seen = append(seen, r)
			already[r] = true
		}
	}
	assert.Equal(t, []string{"unknown_type", "dup_key", "key_truncated"}, seen,
		"reasons must appear in rule-order 1 → 4 → 5 regardless of node placement")
}

// TestValidate_Security71_FileWidgetPreserved — §7.1: the validator must NOT
// strip repoRootRelative or rewrite accept[]; path-traversal rejection lives
// at submission time in internal/bmad/validate.go, not here.
func TestValidate_Security71_FileWidgetPreserved(t *testing.T) {
	t.Parallel()
	ast := &UIAST{
		Version: "1",
		Nodes: []UINode{
			{
				Type:        "decision_group",
				ResponseKey: "input-file",
				Widget: &WidgetNode{
					Type:        "file",
					RepoRootRel: true,
					Accept:      []string{".md", ".txt"},
				},
			},
		},
	}
	Validate(ast, "raw", nil)

	require.Len(t, ast.Nodes, 1)
	require.NotNil(t, ast.Nodes[0].Widget)
	assert.Equal(t, "file", ast.Nodes[0].Widget.Type,
		"validator must not rewrite file widget type")
	assert.True(t, ast.Nodes[0].Widget.RepoRootRel,
		"validator must not strip repoRootRelative")
	assert.Equal(t, []string{".md", ".txt"}, ast.Nodes[0].Widget.Accept,
		"validator must not rewrite the accept list")
}

// TestContentPreservation_URLDropped_SetsUntrusted — §7.2 / §4.7.4: a URL
// present in raw but absent from any node content must cause
// contentPreserved to return false (caller flips Diagnostics.Untrusted).
func TestContentPreservation_URLDropped_SetsUntrusted(t *testing.T) {
	t.Parallel()
	raw := "See https://example.com/foo for details."
	ast := &UIAST{
		Version: "1",
		Nodes:   []UINode{{Type: "markdown", Content: "See the link for details."}},
	}
	preserved := contentPreserved(raw, ast)
	assert.False(t, preserved,
		"dropping a URL from raw → contentPreserved must return false (caller sets Untrusted)")
}

// TestContentPreservation_CodeBlockDropped_SetsUntrusted — §7.2: a fenced
// code block present in raw but absent from any node content must cause
// contentPreserved to return false.
func TestContentPreservation_CodeBlockDropped_SetsUntrusted(t *testing.T) {
	t.Parallel()
	raw := "Run:\n```bash\necho hi\n```\nafter install."
	ast := &UIAST{
		Version: "1",
		Nodes:   []UINode{{Type: "markdown", Content: "Run the command after install."}},
	}
	preserved := contentPreserved(raw, ast)
	assert.False(t, preserved,
		"dropping a fenced code block → contentPreserved must return false")
}

// TestContentPreservation_NumberedListToChoice_NoUntrusted — §7.2 carve-out:
// converting "1. stdout / 2. file / 3. both" into a choice widget is the
// feature's main job; preservation must NOT flag it as untrusted.
func TestContentPreservation_NumberedListToChoice_NoUntrusted(t *testing.T) {
	t.Parallel()
	raw := "Choose one: 1. stdout 2. file 3. both"
	ast := &UIAST{
		Version: "1",
		Nodes: []UINode{
			{
				Type:        "decision_group",
				ResponseKey: "sink",
				Widget: &WidgetNode{
					Type: "choice",
					Options: []WidgetOption{
						{Value: "stdout"}, {Value: "file"}, {Value: "both"},
					},
				},
			},
		},
	}
	preserved := contentPreserved(raw, ast)
	assert.True(t, preserved,
		"numbered-list → choice conversion is the feature — preservation must remain true")
}

// TestValidate_RuleOrdering_PerNodeDeclaration — §3.3: rules are applied per
// node in declaration order. Earlier offenders must report before later ones;
// reversing the input must reverse the reason slice.
func TestValidate_RuleOrdering_PerNodeDeclaration(t *testing.T) {
	t.Parallel()

	empty := UINode{Type: "decision_group", ResponseKey: "a",
		Widget: &WidgetNode{Type: "choice", Options: nil}}
	nilw := UINode{Type: "decision_group", ResponseKey: "b", Widget: nil}

	t.Run("empty_then_nil", func(t *testing.T) {
		ast := &UIAST{Version: "1", Nodes: []UINode{empty, nilw}}
		reasons := Validate(ast, "raw", nil)
		assert.Equal(t, []string{"empty_options", "no_widget"}, reasons)
	})

	t.Run("nil_then_empty", func(t *testing.T) {
		ast := &UIAST{Version: "1", Nodes: []UINode{nilw, empty}}
		reasons := Validate(ast, "raw", nil)
		assert.Equal(t, []string{"no_widget", "empty_options"}, reasons)
	})
}

// TestValidate_WidgetUnknownField_Rejected — §4.7.1: widget-level decoding is
// strict. An unknown field on a widget is a validation error, unlike unknown
// node types which are rewritten to markdown.
func TestValidate_WidgetUnknownField_Rejected(t *testing.T) {
	t.Parallel()
	raw := []byte(`{"type":"choice","options":[],"sneaky":1}`)
	var w WidgetNode
	err := json.Unmarshal(raw, &w)
	require.Error(t, err, "widget with unknown field must fail to decode")
	assert.Contains(t, err.Error(), "sneaky",
		"error must identify the offending field")
}
