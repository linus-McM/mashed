// Package bmad defines domain types and a process registry for the BMAD method.
package bmad

import "errors"

// Sentinel errors for registry lookups.
var (
	ErrProcessNotFound  = errors.New("bmad: process not found")
	ErrWorkflowNotFound = errors.New("bmad: workflow not found")
	ErrModuleNotFound   = errors.New("bmad: module not found")
	ErrAgentNotFound   = errors.New("bmad: agent not found")
	ErrInvalidID       = errors.New("bmad: invalid ID")
	ErrExecNotFound    = errors.New("bmad: execution not found")
	ErrExecNotRunning  = errors.New("bmad: execution not running")
	ErrExecNotPaused   = errors.New("bmad: execution not paused")
	ErrCyclicWorkflow      = errors.New("bmad: workflow contains a cycle")
	ErrSprintFileNotFound  = errors.New("bmad: sprint status file not found")
	ErrSprintFileMalformed = errors.New("bmad: sprint status file is malformed")
	ErrStoryNotFound       = errors.New("bmad: story not found in sprint status")
	ErrInvalidCondition    = errors.New("bmad: invalid condition")
	ErrAnswerTooLong       = errors.New("bmad: answer exceeds size limit")
	// ErrIdleTimeoutNoStart is returned when pollForIdle exhausts its
	// timeout budget before claude produces any new output after command
	// injection — the "stuck at idle baseline" failure mode, distinct
	// from a completion timeout that fires during normal processing.
	ErrIdleTimeoutNoStart = errors.New("bmad: idle wait timed out before claude produced output")
)

// BmadPhase groups processes into lifecycle stages.
type BmadPhase string

const (
	PhaseAnalysis       BmadPhase = "analysis"
	PhasePlanning       BmadPhase = "planning"
	PhaseSolutioning    BmadPhase = "solutioning"
	PhaseImplementation BmadPhase = "implementation"
	PhaseSupport        BmadPhase = "support"
	PhaseUtilities      BmadPhase = "utilities"
)

// BmadAgentRole is the agent persona that executes a process.
type BmadAgentRole string

const (
	RoleAnalyst    BmadAgentRole = "analyst"
	RolePM         BmadAgentRole = "pm"
	RoleUXDesigner BmadAgentRole = "ux-designer"
	RoleArchitect  BmadAgentRole = "architect"
	RoleDeveloper  BmadAgentRole = "developer"
	RoleTechWriter BmadAgentRole = "tech-writer"
	RoleQA         BmadAgentRole = "qa"
)

// ProcessDef is a single BMAD process in the registry.
type ProcessDef struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Phase       BmadPhase     `json:"phase"`
	AgentRole   BmadAgentRole `json:"agentRole"`
	SkillName   string        `json:"skillName"`
	Description string        `json:"description"`
	Inputs      []string      `json:"inputs"`
	Outputs     []string      `json:"outputs"`
	ModuleID    string        `json:"moduleId"`
	Version     string        `json:"version"`
}

// WorkflowNodeStatus tracks execution state of a single node.
type WorkflowNodeStatus string

const (
	NodePending  WorkflowNodeStatus = "pending"
	NodeRunning  WorkflowNodeStatus = "running"
	NodeComplete WorkflowNodeStatus = "complete"
	NodeFailed   WorkflowNodeStatus = "failed"
	NodeSkipped  WorkflowNodeStatus = "skipped"
)

// NodeType discriminates between process nodes and control flow nodes.
type NodeType string

const (
	NodeTypeProcess   NodeType = "process"
	NodeTypeCondition NodeType = "condition"
	NodeTypeLoop      NodeType = "loop"
	NodeTypeLoopUntil NodeType = "loopUntil"
	NodeTypeTransform NodeType = "transform"
	NodeTypeMerge     NodeType = "merge"
	NodeTypeCommand   NodeType = "command"
)

// WorkflowNode is a process instance placed on the canvas.
type WorkflowNode struct {
	ID         string             `json:"id"`
	ProcessID  string             `json:"processId"`
	Label      string             `json:"label"`
	Position   Position           `json:"position"`
	Status     WorkflowNodeStatus `json:"status"`
	Config     map[string]string  `json:"config"`
	TmuxTarget string             `json:"tmuxTarget"`
	StoryID    string             `json:"storyId,omitempty"`
	NodeType   NodeType           `json:"nodeType,omitempty"`
}

// EffectiveType returns the node's type, defaulting to NodeTypeProcess for
// legacy nodes that have an empty NodeType field.
func (n WorkflowNode) EffectiveType() NodeType {
	if n.NodeType == "" {
		return NodeTypeProcess
	}
	return n.NodeType
}

// Position is a 2D coordinate for canvas placement.
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// WorkflowEdge connects two nodes.
type WorkflowEdge struct {
	ID           string `json:"id"`
	Source       string `json:"source"`
	Target       string `json:"target"`
	SourceHandle string `json:"sourceHandle,omitempty"`
	TargetHandle string `json:"targetHandle,omitempty"`
}

// WorkflowDef is a saveable/loadable workflow definition.
type WorkflowDef struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	RepoPath    string         `json:"repoPath,omitempty"`
	Nodes       []WorkflowNode `json:"nodes"`
	Edges       []WorkflowEdge `json:"edges"`
	IsTemplate  bool           `json:"isTemplate"`
	TemplateID  string         `json:"templateId,omitempty"`
	CreatedAt   string         `json:"createdAt"`
	UpdatedAt   string         `json:"updatedAt"`
}

// WorkflowExecStatus tracks overall execution state.
type WorkflowExecStatus string

const (
	ExecIdle     WorkflowExecStatus = "idle"
	ExecRunning  WorkflowExecStatus = "running"
	ExecPaused   WorkflowExecStatus = "paused"
	ExecComplete WorkflowExecStatus = "complete"
	ExecFailed   WorkflowExecStatus = "failed"
)

// WorkflowExecution is a running instance of a WorkflowDef.
type WorkflowExecution struct {
	ID          string             `json:"id"`
	WorkflowID  string            `json:"workflowId"`
	RepoPath    string            `json:"repoPath"`
	Status      WorkflowExecStatus `json:"status"`
	Nodes       []WorkflowNode    `json:"nodes"`
	StartedAt   string            `json:"startedAt"`
	CurrentNode string            `json:"currentNode"`
	NodeOutputs map[string]string  `json:"nodeOutputs,omitempty"`
}

// BmadAgentConfig defines a custom BMAD user agent.
type BmadAgentConfig struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	Role      BmadAgentRole `json:"role"`
	Persona   string        `json:"persona"`
	Skills    []string      `json:"skills"`
	Model     string        `json:"model"`
	CreatedAt string        `json:"createdAt"`
}

// AgentInfo is a lightweight agent reference for local/global Claude agents.
type AgentInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// GroupedAgents contains agents organized by source for the UI dropdown.
type GroupedAgents struct {
	BmadAgents   []BmadAgentConfig `json:"bmadAgents"`
	LocalAgents  []AgentInfo       `json:"localAgents"`
	GlobalAgents []AgentInfo       `json:"globalAgents"`
}

// ArtifactType classifies the format of a BMAD artifact.
type ArtifactType string

const (
	ArtifactMarkdown  ArtifactType = "markdown"
	ArtifactYAML      ArtifactType = "yaml"
	ArtifactDirectory ArtifactType = "directory"
	ArtifactCode      ArtifactType = "code"
)

// ArtifactSpec describes a single artifact produced or consumed by a BMAD process.
type ArtifactSpec struct {
	Name        string       `json:"name"`
	Type        ArtifactType `json:"type"`
	Path        string       `json:"path"`
	Description string       `json:"description"`
	Optional    bool         `json:"optional"`
}

// NodeArtifactEvent reports which expected output artifacts were found after
// a process node completes execution.
type NodeArtifactEvent struct {
	ExecID  string   `json:"execId"`
	NodeID  string   `json:"nodeId"`
	Found   []string `json:"found"`
	Missing []string `json:"missing"`
}

// ControlFlowNodeDef describes a control flow node type for the frontend sidebar.
type ControlFlowNodeDef struct {
	Type        NodeType `json:"type"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Icon        string   `json:"icon"`
}
