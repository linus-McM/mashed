package bmad

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNodeTypeConstants(t *testing.T) {
	tests := []struct {
		constant NodeType
		expected string
	}{
		{NodeTypeProcess, "process"},
		{NodeTypeCondition, "condition"},
		{NodeTypeLoop, "loop"},
		{NodeTypeLoopUntil, "loopUntil"},
		{NodeTypeTransform, "transform"},
		{NodeTypeMerge, "merge"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.expected, string(tt.constant), "NodeType constant mismatch")
	}
}

func TestWorkflowNode_EffectiveType(t *testing.T) {
	tests := []struct {
		name     string
		nodeType NodeType
		want     NodeType
	}{
		{"empty string defaults to process", "", NodeTypeProcess},
		{"explicit process stays process", NodeTypeProcess, NodeTypeProcess},
		{"condition stays condition", NodeTypeCondition, NodeTypeCondition},
		{"loop stays loop", NodeTypeLoop, NodeTypeLoop},
		{"loopUntil stays loopUntil", NodeTypeLoopUntil, NodeTypeLoopUntil},
		{"transform stays transform", NodeTypeTransform, NodeTypeTransform},
		{"merge stays merge", NodeTypeMerge, NodeTypeMerge},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := WorkflowNode{NodeType: tt.nodeType}
			assert.Equal(t, tt.want, node.EffectiveType())
		})
	}
}

func TestWorkflowNode_NodeType_JSONRoundTrip(t *testing.T) {
	t.Run("node with nodeType serializes correctly", func(t *testing.T) {
		node := WorkflowNode{
			ID:        "n1",
			ProcessID: "p1",
			Label:     "Test",
			Position:  Position{X: 10, Y: 20},
			Status:    NodePending,
			Config:    map[string]string{},
			NodeType:  NodeTypeCondition,
		}
		data, err := json.Marshal(node)
		require.NoError(t, err)
		assert.Contains(t, string(data), `"nodeType":"condition"`)

		var restored WorkflowNode
		require.NoError(t, json.Unmarshal(data, &restored))
		assert.Equal(t, NodeTypeCondition, restored.NodeType)
		assert.Equal(t, NodeTypeCondition, restored.EffectiveType())
	})

	t.Run("node without nodeType omits field in JSON", func(t *testing.T) {
		node := WorkflowNode{
			ID:       "n1",
			Label:    "Test",
			Position: Position{X: 0, Y: 0},
			Status:   NodePending,
			Config:   map[string]string{},
		}
		data, err := json.Marshal(node)
		require.NoError(t, err)
		assert.NotContains(t, string(data), `"nodeType"`)
	})

	t.Run("legacy JSON without nodeType deserializes cleanly", func(t *testing.T) {
		legacy := `{"id":"n1","processId":"p1","label":"A","position":{"x":0,"y":0},"status":"pending","config":{},"tmuxTarget":""}`
		var node WorkflowNode
		require.NoError(t, json.Unmarshal([]byte(legacy), &node))
		assert.Equal(t, NodeType(""), node.NodeType)
		assert.Equal(t, NodeTypeProcess, node.EffectiveType())
	})
}

func TestWorkflowEdge_Handles_JSONRoundTrip(t *testing.T) {
	t.Run("edge with handles round-trips correctly", func(t *testing.T) {
		edge := WorkflowEdge{
			ID:           "e1",
			Source:       "n1",
			Target:       "n2",
			SourceHandle: "true",
			TargetHandle: "left",
		}
		data, err := json.Marshal(edge)
		require.NoError(t, err)
		assert.Contains(t, string(data), `"sourceHandle":"true"`)
		assert.Contains(t, string(data), `"targetHandle":"left"`)

		var restored WorkflowEdge
		require.NoError(t, json.Unmarshal(data, &restored))
		assert.Equal(t, "true", restored.SourceHandle)
		assert.Equal(t, "left", restored.TargetHandle)
	})

	t.Run("edge without handles omits fields in JSON", func(t *testing.T) {
		edge := WorkflowEdge{
			ID:     "e1",
			Source: "n1",
			Target: "n2",
		}
		data, err := json.Marshal(edge)
		require.NoError(t, err)
		assert.NotContains(t, string(data), `"sourceHandle"`)
		assert.NotContains(t, string(data), `"targetHandle"`)
	})
}

func TestWorkflowExecution_NodeOutputs_JSONRoundTrip(t *testing.T) {
	t.Run("execution with NodeOutputs round-trips correctly", func(t *testing.T) {
		exec := WorkflowExecution{
			ID:          "exec-1",
			WorkflowID:  "wf-1",
			Status:      ExecRunning,
			NodeOutputs: map[string]string{"node-1": "hello world"},
		}
		data, err := json.Marshal(exec)
		require.NoError(t, err)
		assert.Contains(t, string(data), `"nodeOutputs"`)

		var restored WorkflowExecution
		require.NoError(t, json.Unmarshal(data, &restored))
		assert.Equal(t, "hello world", restored.NodeOutputs["node-1"])
	})

	t.Run("execution without NodeOutputs omits field", func(t *testing.T) {
		exec := WorkflowExecution{
			ID:     "exec-1",
			Status: ExecIdle,
		}
		data, err := json.Marshal(exec)
		require.NoError(t, err)
		assert.NotContains(t, string(data), `"nodeOutputs"`)
	})
}
