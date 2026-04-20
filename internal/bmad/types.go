// Package bmad defines domain types and a process registry for the BMAD method.
package bmad

import "errors"

// Sentinel errors for registry lookups.
var (
	ErrProcessNotFound     = errors.New("bmad: process not found")
	ErrWorkflowNotFound    = errors.New("bmad: workflow not found")
	ErrModuleNotFound      = errors.New("bmad: module not found")
	ErrAgentNotFound       = errors.New("bmad: agent not found")
	ErrInvalidID           = errors.New("bmad: invalid ID")
	ErrExecNotFound        = errors.New("bmad: execution not found")
	ErrExecNotRunning      = errors.New("bmad: execution not running")
	ErrExecNotPaused       = errors.New("bmad: execution not paused")
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
	// ErrDuplicateMultiFileLabel is returned when a MultiFileLoader node's
	// entries contain two labels with the same non-empty value. Positional
	// fallback labels (empty user-provided label) never collide because the
	// executor assigns unique file[N] suffixes by position.
	ErrDuplicateMultiFileLabel = errors.New("bmad: duplicate multiFileLoader label")
	// ErrMultiFileTooMany is returned when a MultiFileLoader config exceeds
	// the per-node entry cap (64).
	ErrMultiFileTooMany = errors.New("bmad: too many multiFileLoader entries")

	// Interactive-process sentinels (schema §5.3 / §8 / §14). Consumers test
	// wrapped errors with errors.Is for type-safe flow control.
	ErrInvalidInput       = errors.New("bmad: invalid input")
	ErrPathOutsideRepo    = errors.New("bmad: path outside repository root")
	ErrNoPendingPrompt    = errors.New("bmad: no pending prompt")
	ErrUnknownInput       = errors.New("bmad: unknown input")
	ErrStalePrompt        = errors.New("bmad: stale prompt")
	ErrInvalidRegistryRef = errors.New("bmad: invalid registry ref")
	// ErrExecNotInitialized is returned by *App bindings when bmadExecutor
	// is nil. The message carries both "bmad" and "not initialized" so
	// existing tests of the legacy RespondToQuestion shim still pass.
	ErrExecNotInitialized = errors.New("bmad executor not initialized")
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
	// Interactive process fields (schema §3). Absent on legacy processes.
	Mode        InteractionMode `json:"mode,omitempty"`
	InputSpecs  []InputSpec     `json:"inputSpecs,omitempty"`
	OutputSpecs []OutputSpec    `json:"outputSpecs,omitempty"`
	Gate        *IterationGate  `json:"gate,omitempty"`
}

// WorkflowNodeStatus tracks execution state of a single node.
type WorkflowNodeStatus string

const (
	NodePending WorkflowNodeStatus = "pending"
	NodeRunning WorkflowNodeStatus = "running"
	// NodeAwaitingInput is the suspension state for interactive processes
	// blocked on a user-supplied input (see schema §3.5).
	NodeAwaitingInput WorkflowNodeStatus = "awaiting_input"
	NodeComplete      WorkflowNodeStatus = "complete"
	NodeFailed        WorkflowNodeStatus = "failed"
	NodeSkipped       WorkflowNodeStatus = "skipped"
)

// InputSource identifies where the value for an InputSpec originates.
type InputSource string

const (
	InputFromFile     InputSource = "file"
	InputFromUpstream InputSource = "upstream"
	InputFromUser     InputSource = "user"
	InputFromEnv      InputSource = "env"
	InputFromRegistry InputSource = "registry"
)

// InputShape describes the interactive UI presentation for a user input.
type InputShape string

const (
	ShapeFree        InputShape = "free"
	ShapeChoice      InputShape = "choice"
	ShapeMultiChoice InputShape = "multi"
	ShapeApproval    InputShape = "approval"
	ShapeFile        InputShape = "file"
	ShapeJSON        InputShape = "json"
)

// InputSpec declares a single input slot for an interactive process (schema §3.1).
type InputSpec struct {
	ID             string      `json:"id"`
	Source         InputSource `json:"source"`
	Shape          InputShape  `json:"shape,omitempty"`
	Required       bool        `json:"required"`
	ArtifactName   string      `json:"artifactName,omitempty"`
	UpstreamNodeID string      `json:"upstreamNodeId,omitempty"`
	Prompt         string      `json:"prompt,omitempty"`
	Options        []string    `json:"options,omitempty"`
	OptionsRef     string      `json:"optionsRef,omitempty"`
	Default        string      `json:"default,omitempty"`
	Validation     string      `json:"validation,omitempty"`
	MaxLength      int         `json:"maxLength,omitempty"`
	HelpText       string      `json:"helpText,omitempty"`
}

// OutputTarget declares where a produced output is persisted.
type OutputTarget string

const (
	OutputToFile   OutputTarget = "file"
	OutputToMemory OutputTarget = "memory"
	OutputToBoth   OutputTarget = "both"
)

// OutputSpec declares a single output slot for an interactive process (schema §3.2).
type OutputSpec struct {
	ID           string       `json:"id"`
	Target       OutputTarget `json:"target"`
	ArtifactName string       `json:"artifactName,omitempty"`
	Description  string       `json:"description,omitempty"`
	Optional     bool         `json:"optional,omitempty"`
}

// InteractionMode selects the orchestration style for an interactive process.
type InteractionMode string

const (
	InteractAutonomous InteractionMode = "autonomous"
	InteractGuided     InteractionMode = "guided"
	InteractIterative  InteractionMode = "iterative"
	InteractParty      InteractionMode = "party"
)

// GateKind discriminates the rule used to exit an iterative loop.
type GateKind string

const (
	GateUserConfirm    GateKind = "userConfirm"
	GateArtifactExists GateKind = "artifact"
	GateExpression     GateKind = "expression"
	GateRoundLimit     GateKind = "rounds"
)

// IterationGate configures the exit condition for iterative processes (§3.3).
type IterationGate struct {
	Kind         GateKind `json:"kind"`
	MaxRounds    int      `json:"maxRounds,omitempty"`
	AcceptTokens []string `json:"acceptTokens,omitempty"`
	RejectTokens []string `json:"rejectTokens,omitempty"`
	CustomExpr   string   `json:"customExpr,omitempty"`
}

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
	// NodeTypeMultiFileLoader is a synchronous utility node that emits one
	// resolved output path per configured entry. See Story breadcrumbs-07.
	NodeTypeMultiFileLoader NodeType = "multiFileLoader"
)

// MultiFileEntry is one configured {label, path} pair for a MultiFileLoader
// node. Entries are serialized as a JSON array into node.Config["entries"].
// An empty Label falls back to positional file[N] keys at execution time.
type MultiFileEntry struct {
	Label string `json:"label"`
	Path  string `json:"path"`
}

// WorkflowNode is a process instance placed on the canvas.
type WorkflowNode struct {
	ID        string             `json:"id"`
	ProcessID string             `json:"processId"`
	Label     string             `json:"label"`
	Position  Position           `json:"position"`
	Status    WorkflowNodeStatus `json:"status"`
	// Config carries per-node configuration as a flat string map.
	//
	// Reserved keys (read by the BMAD canvas breadcrumb renderer —
	// frontend/src/lib/bmad/nodePath.ts):
	//   - "inputPath":  absolute filesystem path to the resolved input artifact,
	//                   or empty string when unresolved.
	//   - "outputPath": absolute filesystem path to the resolved output artifact,
	//                   or empty string when unresolved.
	//
	// Empty string means unresolved — the UI renders an em-dash breadcrumb.
	// Unknown keys are ignored by consumers; marshaling round-trips unchanged.
	Config     map[string]string `json:"config"`
	TmuxTarget string            `json:"tmuxTarget"`
	StartedAt  string            `json:"startedAt,omitempty"`
	StoryID    string            `json:"storyId,omitempty"`
	NodeType   NodeType          `json:"nodeType,omitempty"`
	// OutputPaths holds resolved absolute paths for each output artifact
	// (by artifact name) that existed on disk when the node transitioned
	// to NodeComplete. Unmapped or missing artifacts are omitted.
	// Populated by the executor — see Story breadcrumbs-05.
	OutputPaths map[string]string `json:"outputPaths,omitempty"`
	// InputPaths holds resolved absolute paths for input artifacts,
	// reserved for a future pass. See Story breadcrumbs-05.
	InputPaths map[string]string `json:"inputPaths,omitempty"`
	// InputSpecs optionally overrides the registry-derived InputSpecs for
	// this node. Primary use: test harnesses that construct an execState
	// without going through registry registration. When empty, the executor
	// falls back to ProcessByID(ProcessID).InputSpecs.
	InputSpecs []InputSpec `json:"inputSpecs,omitempty"`
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
	WorkflowID  string             `json:"workflowId"`
	RepoPath    string             `json:"repoPath"`
	Status      WorkflowExecStatus `json:"status"`
	Nodes       []WorkflowNode     `json:"nodes"`
	StartedAt   string             `json:"startedAt"`
	CurrentNode string             `json:"currentNode"`
	NodeOutputs map[string]string  `json:"nodeOutputs,omitempty"`
	// Interactive execution state (schema §3.5). Keyed by node ID.
	NodeRounds       map[string]int              `json:"nodeRounds,omitempty"`
	PendingPrompts   []PendingPrompt             `json:"pendingPrompts,omitempty"`
	NodeInputs       map[string]map[string]string `json:"nodeInputs,omitempty"`
	NodeInputHistory map[string][]NodeInputEntry  `json:"nodeInputHistory,omitempty"`
}

// PendingPrompt is an outstanding user-input request for a suspended node (§3.5).
type PendingPrompt struct {
	NodeID    string     `json:"nodeId"`
	InputID   string     `json:"inputId"`
	Prompt    string     `json:"prompt"`
	Shape     InputShape `json:"shape"`
	Options   []string   `json:"options,omitempty"`
	Round     int        `json:"round"`
	CreatedAt int64      `json:"createdAt"`
	PromptID  string     `json:"promptId"`
}

// NodeInputEntry is one historical user answer for a node input (§3.5).
type NodeInputEntry struct {
	InputID   string `json:"inputId"`
	Round     int    `json:"round"`
	Value     string `json:"value"`
	Timestamp int64  `json:"timestamp"`
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
	// Paths maps each found output artifact name to its resolved
	// absolute path. Populated by the executor on node completion —
	// see Story breadcrumbs-05.
	Paths map[string]string `json:"paths"`
}

// ControlFlowNodeDef describes a control flow node type for the frontend sidebar.
type ControlFlowNodeDef struct {
	Type        NodeType `json:"type"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Icon        string   `json:"icon"`
}
