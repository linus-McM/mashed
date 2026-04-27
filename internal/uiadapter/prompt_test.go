package uiadapter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	minPromptBytes = 1500
	// maxPromptBytes was raised from 7000 to 8500 in the v2 (classification-
	// first) rewrite — the new "Step 1 — Classify the turn" preamble and
	// matching examples added ~1.5 KiB of decision-rule prose so Haiku reads
	// the full turn and picks INFORMING / ASKING-* / AMBIGUOUS before
	// generating UI. Still well under the ≤ ~4 KiB static prefix +
	// per-kind prompts ceiling implied by plan §6.3.
	maxPromptBytes = 8500
)

// readFixture loads a fixture file from testdata/prompts/ and returns its
// string contents. Fails the test on any read error.
func readFixture(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "prompts", name))
	require.NoError(t, err, "read fixture %s", name)
	return string(data)
}

// newStubbedOllama starts an httptest.Server that replies to /api/chat with
// {"message":{"role":"assistant","content":<fixtureJSON>}} — the exact shape
// Client.Chat decodes. Callers must still call withOllamaHost(t, srv.URL) to
// wire the package's ollamaHost var at the stub.
func newStubbedOllama(t *testing.T, fixtureJSON string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(ollamaChatPayload(fixtureJSON)))
	}))
	t.Cleanup(srv.Close)
	return srv
}

// decisionGroups filters an AST's nodes to only decision_group entries.
// Helper for the golden-path tests to assert widget shape without looping.
func decisionGroups(ast *UIAST) []UINode {
	var out []UINode
	for _, n := range ast.Nodes {
		if n.Type == decisionGroupNodeType {
			out = append(out, n)
		}
	}
	return out
}

// goldenAdapter is the adapter construction used across every golden test:
// enabled, model gemma3:4b, 3s timeout, single in-flight. Kept in one place so
// per-fixture tests don't drift on Config values.
func goldenAdapter(t *testing.T) Adapter {
	t.Helper()
	return NewDefault(
		Config{Enabled: true, Model: "gemma3:4b", TimeoutMs: 3000, MaxInflight: 1},
		discardLogger,
	)
}

// runGolden wires up a stubbed Ollama against the named fixture pair and
// returns the raw capture alongside the AST Translate produced. procID is
// passed to Translate for telemetry symmetry with real callers.
func runGolden(t *testing.T, name, procID string) (string, *UIAST) {
	t.Helper()
	raw := readFixture(t, name+"-raw.txt")
	response := readFixture(t, name+"-response.json")
	srv := newStubbedOllama(t, response)
	withOllamaHost(t, srv.URL)
	return raw, goldenAdapter(t).Translate(context.Background(), raw, procID)
}

// TestSystemPrompt_ContainsAllSections covers AC-1: the embedded prompt must
// carry the five §4.6 section headings and stay within the byte-budget guard.
func TestSystemPrompt_ContainsAllSections(t *testing.T) {
	prompt := SystemPrompt()

	for _, section := range []string{"Role", "Schema", "Examples", "Output contract", "Safety"} {
		assert.Contains(t, prompt, section,
			"AC-1: SystemPrompt must contain the %q section heading", section)
	}

	require.GreaterOrEqual(t, len(prompt), minPromptBytes,
		"AC-1: prompt is too short (%d bytes); readability floor is %d", len(prompt), minPromptBytes)
	require.LessOrEqual(t, len(prompt), maxPromptBytes,
		"AC-1: prompt exceeds %d-byte budget (%d bytes)", maxPromptBytes, len(prompt))
}

// TestPromptVersion_IsV2 covers AC-2: any edit to prompt.md must be
// accompanied by a deliberate version bump — pin the constant at "v2" so the
// telemetry `prompt_version` field stays stable across non-behavioural edits.
// v2 introduced the classification-first preamble (INFORMING / ASKING-* /
// AMBIGUOUS) so Haiku decides shape before generating UI.
func TestPromptVersion_IsV2(t *testing.T) {
	assert.Equal(t, "v2", promptVersion,
		"AC-2: promptVersion must equal \"v2\"; bump only on behaviour-changing prompt edits")
}

// TestAdapter_SendsSystemPromptInRequest covers AC-8: the embedded system
// prompt must land in the chat request body's messages[0] slot. A recording
// server captures and decodes the request body so the assertion is on the
// actual wire payload, not on internal plumbing.
func TestAdapter_SendsSystemPromptInRequest(t *testing.T) {
	var captured chatRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewDecoder(r.Body).Decode(&captured),
			"AC-8: recording server must decode the adapter's chat request body")
		_, _ = w.Write([]byte(ollamaChatPayload(validASTJSON)))
	}))
	t.Cleanup(srv.Close)
	withOllamaHost(t, srv.URL)

	a := goldenAdapter(t)
	_ = a.Translate(context.Background(), "hello raw turn", "proc")

	require.GreaterOrEqual(t, len(captured.Messages), 1,
		"AC-8: chat request must carry at least one message")
	assert.Equal(t, "system", captured.Messages[0].Role,
		"AC-8: messages[0].role must be \"system\"")
	assert.Equal(t, SystemPrompt(), captured.Messages[0].Content,
		"AC-8: messages[0].content must equal SystemPrompt()")
}

// TestPrompt_Golden_Brainstorming covers AC-3: one decision_group carrying a
// choice widget with exactly five options, non-fallback envelope.
func TestPrompt_Golden_Brainstorming(t *testing.T) {
	_, ast := runGolden(t, "brainstorming", "bmad-brainstorming")

	require.NotNil(t, ast)
	require.Equal(t, "1", ast.Version)
	require.False(t, strings.HasPrefix(ast.GeneratedBy, "fallback:"),
		"AC-3: expected non-fallback envelope, got %q", ast.GeneratedBy)

	groups := decisionGroups(ast)
	require.Len(t, groups, 1, "AC-3: brainstorming fixture must produce exactly one decision_group")
	require.NotNil(t, groups[0].Widget, "AC-3: decision_group must carry a widget")
	assert.Equal(t, "choice", groups[0].Widget.Type, "AC-3: widget.type must be choice")
	require.Len(t, groups[0].Widget.Options, 5, "AC-3: choice widget must expose five options")
	assert.Equal(t, "brainwriting", groups[0].Widget.Options[0].Value,
		"AC-3: first option.value must be %q — catches option-value scrambling regressions", "brainwriting")
}

// TestPrompt_Golden_Elicitation covers AC-4: one decision_group with a multi
// widget whose options use the numeric values "1".."5" (mirrors the source
// "or r/a/x" shape).
func TestPrompt_Golden_Elicitation(t *testing.T) {
	_, ast := runGolden(t, "elicitation", "bmad-elicitation")

	require.NotNil(t, ast)
	require.False(t, strings.HasPrefix(ast.GeneratedBy, "fallback:"),
		"AC-4: expected non-fallback envelope, got %q", ast.GeneratedBy)

	groups := decisionGroups(ast)
	require.Len(t, groups, 1, "AC-4: elicitation fixture must produce exactly one decision_group")
	require.NotNil(t, groups[0].Widget)
	assert.Equal(t, "multi", groups[0].Widget.Type, "AC-4: widget.type must be multi")
	require.Len(t, groups[0].Widget.Options, 5, "AC-4: multi widget must expose five options")

	wantValues := []string{"1", "2", "3", "4", "5"}
	for i, opt := range groups[0].Widget.Options {
		assert.Equal(t, wantValues[i], opt.Value,
			"AC-4: elicitation option[%d].value must be numeric %q", i, wantValues[i])
	}
}

// TestPrompt_Golden_ProductBrief covers AC-5: three decision_group nodes, the
// widget mix includes both choice and free, and the validator leaves
// Diagnostics.Collapsed false (frontend owns the collapse rule).
func TestPrompt_Golden_ProductBrief(t *testing.T) {
	_, ast := runGolden(t, "product-brief", "bmad-product-brief")

	require.NotNil(t, ast)
	require.False(t, strings.HasPrefix(ast.GeneratedBy, "fallback:"),
		"AC-5: expected non-fallback envelope, got %q", ast.GeneratedBy)

	groups := decisionGroups(ast)
	require.Len(t, groups, 3, "AC-5: product-brief fixture must produce three decision_groups")

	widgetTypes := make(map[string]bool, len(groups))
	var choiceGroup *UINode
	for i, g := range groups {
		require.NotNil(t, g.Widget, "AC-5: decision_group[%d] must carry a widget", i)
		widgetTypes[g.Widget.Type] = true
		if g.Widget.Type == "choice" && choiceGroup == nil {
			choiceGroup = &groups[i]
		}
	}
	assert.True(t, widgetTypes["choice"], "AC-5: widget mix must include choice")
	assert.True(t, widgetTypes["free"], "AC-5: widget mix must include free")
	require.NotNil(t, choiceGroup, "AC-5: at least one decision_group must carry a choice widget")
	require.NotEmpty(t, choiceGroup.Widget.Options, "AC-5: choice widget must expose options")
	assert.Equal(t, "end-user", choiceGroup.Widget.Options[0].Value,
		"AC-5: first choice option.value must be %q — catches option-value scrambling regressions", "end-user")
	assert.False(t, ast.Diagnostics.Collapsed,
		"AC-5: validator must leave Diagnostics.Collapsed=false; collapse is a frontend concern")
}

// TestPrompt_Golden_PartyMode covers AC-6: one approval-style decision_group
// and a trusted envelope (the markdown code-block sub-test enforces the
// preservation invariant in isolation).
func TestPrompt_Golden_PartyMode(t *testing.T) {
	_, ast := runGolden(t, "party-mode", "bmad-party-mode")

	require.NotNil(t, ast)
	require.False(t, strings.HasPrefix(ast.GeneratedBy, "fallback:"),
		"AC-6: expected non-fallback envelope, got %q", ast.GeneratedBy)

	groups := decisionGroups(ast)
	require.Len(t, groups, 1, "AC-6: party-mode must produce one decision_group")
	require.NotNil(t, groups[0].Widget)
	assert.Equal(t, "approval", groups[0].Widget.Type, "AC-6: widget.type must be approval")
	assert.False(t, ast.Diagnostics.Untrusted,
		"AC-6: contentPreserved passes when the fenced code block is copied verbatim")
}

// TestPrompt_Golden_PartyMode_CodeBlockPreserved is AC-6's preservation
// sub-test: the party-mode markdown node must contain the raw fenced block
// byte-for-byte so contentPreserved never flags Untrusted.
func TestPrompt_Golden_PartyMode_CodeBlockPreserved(t *testing.T) {
	raw, ast := runGolden(t, "party-mode", "bmad-party-mode")

	require.NotNil(t, ast)

	fence := codeFencePattern.FindString(raw)
	require.NotEmpty(t, fence, "AC-6 setup: party-mode raw must contain a fenced code block")

	var preserved bool
	for _, n := range ast.Nodes {
		if n.Type == markdownNodeType && strings.Contains(n.Content, fence) {
			preserved = true
			break
		}
	}
	require.True(t, preserved,
		"AC-6: a markdown node must preserve the raw fenced code block verbatim")
	assert.False(t, ast.Diagnostics.Untrusted,
		"AC-6: Diagnostics.Untrusted must stay false when the code block is preserved")
}

// TestPrompt_Golden_Freeform covers AC-7: single open-ended question yields
// zero decision_groups and fallback_answer_shape "free".
func TestPrompt_Golden_Freeform(t *testing.T) {
	_, ast := runGolden(t, "freeform", "bmad-freeform")

	require.NotNil(t, ast)
	require.False(t, strings.HasPrefix(ast.GeneratedBy, "fallback:"),
		"AC-7: expected non-fallback envelope, got %q", ast.GeneratedBy)
	assert.Empty(t, decisionGroups(ast), "AC-7: freeform must produce zero decision_group nodes")
	assert.Equal(t, "free", ast.FallbackAnswerShape,
		"AC-7: FallbackAnswerShape must be \"free\" for a single open-ended question")
}
