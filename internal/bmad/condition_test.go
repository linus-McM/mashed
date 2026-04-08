package bmad

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCondition_Evaluate_Contains(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		output   string
		expected bool
	}{
		{"exact match", "hello", "hello", true},
		{"substring match", "world", "hello world!", true},
		{"no match", "missing", "hello world", false},
		{"empty pattern matches anything", "", "some output", true},
		{"empty output with empty pattern", "", "", true},
		{"empty output with non-empty pattern", "something", "", false},
		{"case sensitive", "Hello", "hello world", false},
		{"multiline output", "line2", "line1\nline2\nline3", true},
		{"special characters", "exit(0)", "called exit(0) successfully", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Condition{
				Type:       CondContains,
				Pattern:    tt.pattern,
				SourceNode: "node-1",
			}
			outputs := map[string]string{"node-1": tt.output}
			assert.Equal(t, tt.expected, c.Evaluate(outputs, ""))
		})
	}
}

func TestCondition_Evaluate_NotContains(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		output   string
		expected bool
	}{
		{"pattern absent", "missing", "hello world", true},
		{"pattern present", "hello", "hello world", false},
		{"empty pattern always matches so notContains is false", "", "some output", false},
		{"empty output with non-empty pattern", "something", "", true},
		{"case sensitive mismatch", "Hello", "hello world", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Condition{
				Type:       CondNotContains,
				Pattern:    tt.pattern,
				SourceNode: "node-1",
			}
			outputs := map[string]string{"node-1": tt.output}
			assert.Equal(t, tt.expected, c.Evaluate(outputs, ""))
		})
	}
}

func TestCondition_Evaluate_Regex(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		output   string
		expected bool
	}{
		{"simple match", `\d+`, "abc123def", true},
		{"no match", `^\d+$`, "abc123def", false},
		{"full line match", `^hello$`, "hello", true},
		{"multiline", `line\d`, "line1\nline2", true},
		{"invalid regex returns false", `[invalid`, "anything", false},
		{"empty pattern matches anything", ``, "hello", true},
		{"complex pattern", `^(ok|success|passed)$`, "passed", true},
		{"complex pattern no match", `^(ok|success|passed)$`, "failed", false},
		{"dot star", `err.*found`, "error was found here", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Condition{
				Type:       CondRegex,
				Pattern:    tt.pattern,
				SourceNode: "node-1",
			}
			outputs := map[string]string{"node-1": tt.output}
			assert.Equal(t, tt.expected, c.Evaluate(outputs, ""))
		})
	}
}

func TestCondition_Evaluate_ExitCode(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		output   string
		expected bool
	}{
		{"exit code colon format", "0", "process finished exit code: 0", true},
		{"exit code colon nonzero", "1", "exit code: 1", true},
		{"exited with code", "0", "program exited with code 0", true},
		{"exit status", "0", "exit status: 0", true},
		{"exit status nonzero", "127", "command exit status: 127", true},
		{"wrong code", "0", "exit code: 1", false},
		{"no exit pattern in output", "0", "everything is fine", false},
		{"case insensitive", "0", "EXIT CODE: 0", true},
		{"multidigit code", "42", "exit code: 42 encountered", true},
		{"partial code mismatch", "1", "exit code: 12", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Condition{
				Type:       CondExitCode,
				Pattern:    tt.pattern,
				SourceNode: "node-1",
			}
			outputs := map[string]string{"node-1": tt.output}
			assert.Equal(t, tt.expected, c.Evaluate(outputs, ""))
		})
	}
}

func TestCondition_Evaluate_FileExists(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test file.
	testFile := "somedir/testfile.txt"
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "somedir"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, testFile), []byte("data"), 0o644))

	tests := []struct {
		name     string
		pattern  string
		expected bool
	}{
		{"file exists", testFile, true},
		{"file does not exist", "nonexistent.txt", false},
		{"directory exists", "somedir", true},
		{"nested nonexistent", "somedir/missing.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Condition{
				Type:       CondFileExists,
				Pattern:    tt.pattern,
				SourceNode: "node-1",
			}
			// fileExists does not use output, but still receives the map.
			outputs := map[string]string{"node-1": "irrelevant"}
			assert.Equal(t, tt.expected, c.Evaluate(outputs, tmpDir))
		})
	}
}

func TestCondition_Evaluate_Always(t *testing.T) {
	tests := []struct {
		name   string
		output string
	}{
		{"with output", "some output"},
		{"empty output", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Condition{
				Type:       CondAlways,
				SourceNode: "node-1",
			}
			outputs := map[string]string{"node-1": tt.output}
			assert.True(t, c.Evaluate(outputs, ""))
		})
	}
}

func TestCondition_Evaluate_MissingSourceNode(t *testing.T) {
	outputs := map[string]string{"other-node": "data"}

	tests := []struct {
		name     string
		condType ConditionType
		expected bool
	}{
		{"contains returns false", CondContains, false},
		{"notContains returns true", CondNotContains, true},
		{"regex returns false", CondRegex, false},
		{"exitCode returns false", CondExitCode, false},
		{"always returns true", CondAlways, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Condition{
				Type:       tt.condType,
				Pattern:    "test",
				SourceNode: "missing-node",
			}
			assert.Equal(t, tt.expected, c.Evaluate(outputs, ""))
		})
	}

	// fileExists does not depend on source node output.
	t.Run("fileExists proceeds normally", func(t *testing.T) {
		c := &Condition{
			Type:       CondFileExists,
			Pattern:    "nonexistent-file.txt",
			SourceNode: "missing-node",
		}
		assert.False(t, c.Evaluate(outputs, t.TempDir()))
	})
}

func TestCondition_Evaluate_UnknownType(t *testing.T) {
	c := &Condition{
		Type:       "bogus",
		Pattern:    "test",
		SourceNode: "node-1",
	}
	outputs := map[string]string{"node-1": "test"}
	assert.False(t, c.Evaluate(outputs, ""))
}

func TestParseCondition_Valid(t *testing.T) {
	tests := []struct {
		name         string
		json         string
		expectedType ConditionType
		expectedPat  string
		expectedSrc  string
	}{
		{
			"contains",
			`{"type":"contains","pattern":"hello","sourceNode":"n1"}`,
			CondContains, "hello", "n1",
		},
		{
			"notContains",
			`{"type":"notContains","pattern":"error","sourceNode":"n2"}`,
			CondNotContains, "error", "n2",
		},
		{
			"regex",
			`{"type":"regex","pattern":"\\d+","sourceNode":"n1"}`,
			CondRegex, `\d+`, "n1",
		},
		{
			"exitCode",
			`{"type":"exitCode","pattern":"0","sourceNode":"n1"}`,
			CondExitCode, "0", "n1",
		},
		{
			"fileExists",
			`{"type":"fileExists","pattern":"go.mod"}`,
			CondFileExists, "go.mod", "",
		},
		{
			"always",
			`{"type":"always","pattern":""}`,
			CondAlways, "", "",
		},
		{
			"always without pattern",
			`{"type":"always"}`,
			CondAlways, "", "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cond, err := ParseCondition(tt.json)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedType, cond.Type)
			assert.Equal(t, tt.expectedPat, cond.Pattern)
			assert.Equal(t, tt.expectedSrc, cond.SourceNode)
		})
	}
}

func TestParseCondition_Invalid(t *testing.T) {
	tests := []struct {
		name string
		json string
	}{
		{"bad JSON", `{not json}`},
		{"unknown type", `{"type":"unknown","pattern":"x"}`},
		{"empty type", `{"type":"","pattern":"x"}`},
		{"path traversal in fileExists", `{"type":"fileExists","pattern":"../../../etc/passwd"}`},
		{"path traversal mid-path", `{"type":"fileExists","pattern":"some/../../../etc/passwd"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cond, err := ParseCondition(tt.json)
			assert.Nil(t, cond)
			assert.ErrorIs(t, err, ErrInvalidCondition)
		})
	}
}

func TestCondition_Evaluate_NilOutputs(t *testing.T) {
	c := &Condition{
		Type:       CondContains,
		Pattern:    "hello",
		SourceNode: "node-1",
	}
	// nil map should behave same as missing source node.
	assert.False(t, c.Evaluate(nil, ""))
}
