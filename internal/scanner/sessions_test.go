package scanner

import (
	"os"
	"path/filepath"
	"testing"

	"conductor/internal/domain"
)

// helper to write JSONL content to a temp file and return its path.
func writeTempJSONL(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test-session.jsonl")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writing temp JSONL: %v", err)
	}
	return path
}

// newTestProvider creates a minimal ClaudeCodeProvider for test use.
func newTestProvider(t *testing.T) *ClaudeCodeProvider {
	t.Helper()
	return &ClaudeCodeProvider{
		claudeDir:   t.TempDir(),
		pidDirCache: make(map[int]pidDirEntry),
		parsers:     make(map[string]*sessionParserState),
	}
}

func TestParseSession_AssistantTokenAccumulation(t *testing.T) {
	jsonl := `{"type":"assistant","message":{"usage":{"input_tokens":100,"output_tokens":50,"cache_read_input_tokens":10,"cache_creation_input_tokens":5},"content":[{"type":"text","text":"Hello, I will help you with that task."}]}}
{"type":"assistant","message":{"usage":{"input_tokens":200,"output_tokens":75,"cache_read_input_tokens":20,"cache_creation_input_tokens":10},"content":[{"type":"text","text":"Here is the updated implementation code."}]}}
`

	path := writeTempJSONL(t, jsonl)
	p := newTestProvider(t)

	data, err := p.ParseSession(path)
	if err != nil {
		t.Fatalf("ParseSession error: %v", err)
	}

	if data.InputTokens != 300 {
		t.Errorf("InputTokens = %d, want 300", data.InputTokens)
	}
	if data.OutputTokens != 125 {
		t.Errorf("OutputTokens = %d, want 125", data.OutputTokens)
	}
	if data.CacheReadTokens != 30 {
		t.Errorf("CacheReadTokens = %d, want 30", data.CacheReadTokens)
	}
	if data.CacheCreateTokens != 15 {
		t.Errorf("CacheCreateTokens = %d, want 15", data.CacheCreateTokens)
	}
	if data.TotalTokens != 300+125+30+15 {
		t.Errorf("TotalTokens = %d, want %d", data.TotalTokens, 300+125+30+15)
	}
}

func TestParseSession_ToolUseLogKinds(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		wantKind domain.LogKind
	}{
		{"Bash maps to info", "Bash", domain.LogInfo},
		{"Read maps to dim", "Read", domain.LogDim},
		{"Write maps to ok", "Write", domain.LogOK},
		{"Edit maps to ok", "Edit", domain.LogOK},
		{"Glob maps to dim", "Glob", domain.LogDim},
		{"Grep maps to dim", "Grep", domain.LogDim},
		{"Agent maps to system", "Agent", domain.LogSystem},
		{"Skill maps to system", "Skill", domain.LogSystem},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonl := `{"type":"assistant","message":{"usage":{"input_tokens":10,"output_tokens":5,"cache_read_input_tokens":0,"cache_creation_input_tokens":0},"content":[{"type":"tool_use","id":"tool_1","name":"` + tt.toolName + `","input":{"command":"echo hello"}}]}}
`
			path := writeTempJSONL(t, jsonl)
			p := newTestProvider(t)

			data, err := p.ParseSession(path)
			if err != nil {
				t.Fatalf("ParseSession error: %v", err)
			}

			if len(data.LogLines) == 0 {
				t.Fatal("expected at least one LogLine")
			}

			got := data.LogLines[0].Kind
			if got != tt.wantKind {
				t.Errorf("tool %q: LogLine.Kind = %q, want %q", tt.toolName, got, tt.wantKind)
			}
		})
	}
}

func TestParseSession_SubAgentDetection(t *testing.T) {
	jsonl := `{"type":"assistant","message":{"usage":{"input_tokens":10,"output_tokens":5,"cache_read_input_tokens":0,"cache_creation_input_tokens":0},"content":[{"type":"tool_use","id":"tool_agent_1","name":"Agent","input":{"name":"test-agent","description":"runs tests"}}]}}
{"type":"user","message":{"content":[{"type":"tool_result","tool_use_id":"tool_agent_1","content":"all tests passed"}]}}
`

	path := writeTempJSONL(t, jsonl)
	p := newTestProvider(t)

	data, err := p.ParseSession(path)
	if err != nil {
		t.Fatalf("ParseSession error: %v", err)
	}

	if len(data.SubAgents) != 1 {
		t.Fatalf("SubAgents count = %d, want 1", len(data.SubAgents))
	}

	sub := data.SubAgents[0]
	if sub.Name != "test-agent" {
		t.Errorf("SubAgent.Name = %q, want %q", sub.Name, "test-agent")
	}
	if sub.Description != "runs tests" {
		t.Errorf("SubAgent.Description = %q, want %q", sub.Description, "runs tests")
	}
	if sub.ToolUseID != "tool_agent_1" {
		t.Errorf("SubAgent.ToolUseID = %q, want %q", sub.ToolUseID, "tool_agent_1")
	}
	if sub.Status != "done" {
		t.Errorf("SubAgent.Status = %q, want %q", sub.Status, "done")
	}
	if sub.Result != "all tests passed" {
		t.Errorf("SubAgent.Result = %q, want %q", sub.Result, "all tests passed")
	}
}

func TestParseSession_MalformedJSONLSkipped(t *testing.T) {
	jsonl := `{"type":"assistant","message":{"usage":{"input_tokens":50,"output_tokens":25,"cache_read_input_tokens":0,"cache_creation_input_tokens":0},"content":[{"type":"text","text":"First valid message here."}]}}
this is not valid json at all{{{
{"type":"assistant","message":{"usage":{"input_tokens":100,"output_tokens":50,"cache_read_input_tokens":0,"cache_creation_input_tokens":0},"content":[{"type":"text","text":"Second valid message here."}]}}
`

	path := writeTempJSONL(t, jsonl)
	p := newTestProvider(t)

	data, err := p.ParseSession(path)
	if err != nil {
		t.Fatalf("ParseSession error: %v", err)
	}

	// Should have accumulated tokens from both valid lines, skipping the malformed one.
	if data.InputTokens != 150 {
		t.Errorf("InputTokens = %d, want 150 (malformed line should be skipped)", data.InputTokens)
	}
	if data.OutputTokens != 75 {
		t.Errorf("OutputTokens = %d, want 75", data.OutputTokens)
	}
}

func TestParseSession_EmptyFile(t *testing.T) {
	path := writeTempJSONL(t, "")
	p := newTestProvider(t)

	data, err := p.ParseSession(path)
	if err != nil {
		t.Fatalf("ParseSession error: %v", err)
	}

	if data.TotalTokens != 0 {
		t.Errorf("TotalTokens = %d, want 0 for empty file", data.TotalTokens)
	}
	if len(data.LogLines) != 0 {
		t.Errorf("LogLines count = %d, want 0 for empty file", len(data.LogLines))
	}
	if len(data.SubAgents) != 0 {
		t.Errorf("SubAgents count = %d, want 0 for empty file", len(data.SubAgents))
	}
}

func TestParseSession_UnknownToolName(t *testing.T) {
	jsonl := `{"type":"assistant","message":{"usage":{"input_tokens":10,"output_tokens":5,"cache_read_input_tokens":0,"cache_creation_input_tokens":0},"content":[{"type":"tool_use","id":"tool_unk","name":"UnknownTool","input":{"foo":"bar"}}]}}
`
	path := writeTempJSONL(t, jsonl)
	p := newTestProvider(t)

	data, err := p.ParseSession(path)
	if err != nil {
		t.Fatalf("ParseSession error: %v", err)
	}

	if len(data.LogLines) != 1 {
		t.Fatalf("LogLines count = %d, want 1", len(data.LogLines))
	}

	// Unknown tool should default to LogInfo.
	got := data.LogLines[0].Kind
	if got != domain.LogInfo {
		t.Errorf("unknown tool LogLine.Kind = %q, want %q (default)", got, domain.LogInfo)
	}
}
