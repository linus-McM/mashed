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
)

// BmadPhase groups processes into lifecycle stages.
type BmadPhase string

const (
	PhaseAnalysis       BmadPhase = "analysis"
	PhasePlanning       BmadPhase = "planning"
	PhaseSolutioning    BmadPhase = "solutioning"
	PhaseImplementation BmadPhase = "implementation"
	PhaseSupport        BmadPhase = "support"
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
}

// Position is a 2D coordinate for canvas placement.
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// WorkflowEdge connects two nodes.
type WorkflowEdge struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Target string `json:"target"`
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
