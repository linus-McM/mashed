package bmad

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ConditionType is the type of comparison a condition performs.
type ConditionType string

const (
	CondContains    ConditionType = "contains"
	CondNotContains ConditionType = "notContains"
	CondRegex       ConditionType = "regex"
	CondExitCode    ConditionType = "exitCode"
	CondFileExists  ConditionType = "fileExists"
	CondAlways      ConditionType = "always"
)

// validConditionTypes is the set of recognised condition types.
var validConditionTypes = map[ConditionType]struct{}{
	CondContains:    {},
	CondNotContains: {},
	CondRegex:       {},
	CondExitCode:    {},
	CondFileExists:  {},
	CondAlways:      {},
}

// Condition defines a test to run against node output.
type Condition struct {
	Type       ConditionType `json:"type"`
	Pattern    string        `json:"pattern"`
	SourceNode string        `json:"sourceNode,omitempty"`
}

// Evaluate tests the condition against the nodeOutputs map.
// It looks up nodeOutputs[c.SourceNode] and evaluates based on Type.
// Returns false for unknown types or missing source nodes (except notContains and always).
func (c *Condition) Evaluate(nodeOutputs map[string]string, repoPath string) bool {
	// fileExists does not depend on node output.
	if c.Type == CondFileExists {
		_, err := os.Stat(filepath.Join(repoPath, c.Pattern))
		return err == nil
	}

	if c.Type == CondAlways {
		return true
	}

	output, found := nodeOutputs[c.SourceNode]
	if !found {
		switch c.Type {
		case CondNotContains:
			return true
		default:
			return false
		}
	}

	switch c.Type {
	case CondContains:
		return strings.Contains(output, c.Pattern)
	case CondNotContains:
		return !strings.Contains(output, c.Pattern)
	case CondRegex:
		matched, err := regexp.MatchString(c.Pattern, output)
		if err != nil {
			return false
		}
		return matched
	case CondExitCode:
		// Match patterns like "exit code: N", "exited with code N", "exit status: N".
		// The pattern value is the expected exit code (e.g. "0", "1", "127").
		re, err := regexp.Compile(`(?i)(?:exit code|exited with code|exit status)[:\s]+` + regexp.QuoteMeta(c.Pattern) + `\b`)
		if err != nil {
			return false
		}
		return re.MatchString(output)
	default:
		return false
	}
}

// ParseCondition deserializes a Condition from JSON.
// Returns ErrInvalidCondition for invalid JSON, unknown types, or path traversal in fileExists.
func ParseCondition(configJSON string) (*Condition, error) {
	var cond Condition
	if err := json.Unmarshal([]byte(configJSON), &cond); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidCondition, err.Error())
	}

	if _, ok := validConditionTypes[cond.Type]; !ok {
		return nil, fmt.Errorf("%w: unknown type %q", ErrInvalidCondition, cond.Type)
	}

	if cond.Type == CondFileExists && strings.Contains(cond.Pattern, "..") {
		return nil, fmt.Errorf("%w: path traversal not allowed in fileExists pattern", ErrInvalidCondition)
	}

	return &cond, nil
}
