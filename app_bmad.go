package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"mashed/internal/bmad"
)

// ── BMAD Workflow CRUD ──

// ListBmadWorkflows returns all user-saved workflows.
func (a *App) ListBmadWorkflows() ([]bmad.WorkflowDef, error) {
	if a.bmadStorage == nil {
		return nil, fmt.Errorf("bmad storage not initialized")
	}
	return a.bmadStorage.ListWorkflows()
}

// GetBmadWorkflow loads a single workflow by ID.
func (a *App) GetBmadWorkflow(id string) (bmad.WorkflowDef, error) {
	if a.bmadStorage == nil {
		return bmad.WorkflowDef{}, fmt.Errorf("bmad storage not initialized")
	}
	return a.bmadStorage.LoadWorkflow(id)
}

// SaveBmadWorkflow persists a workflow definition.
func (a *App) SaveBmadWorkflow(wf bmad.WorkflowDef) error {
	if a.bmadStorage == nil {
		return fmt.Errorf("bmad storage not initialized")
	}
	return a.bmadStorage.SaveWorkflow(wf)
}

// DeleteBmadWorkflow removes a workflow by ID.
func (a *App) DeleteBmadWorkflow(id string) error {
	if a.bmadStorage == nil {
		return fmt.Errorf("bmad storage not initialized")
	}
	return a.bmadStorage.DeleteWorkflow(id)
}

// ── BMAD Process Registry ──

// GetBmadProcesses returns the full process catalog.
func (a *App) GetBmadProcesses() []bmad.ProcessDef {
	return bmad.AllProcesses()
}

// GetBmadProcessesByPhase returns processes filtered by lifecycle phase.
func (a *App) GetBmadProcessesByPhase(phase string) []bmad.ProcessDef {
	return bmad.ProcessesByPhase(bmad.BmadPhase(phase))
}

// ListBmadWorkflowsByRepo returns workflows scoped to a specific repository path.
func (a *App) ListBmadWorkflowsByRepo(repoPath string) ([]bmad.WorkflowDef, error) {
	if a.bmadStorage == nil {
		return nil, fmt.Errorf("bmad storage not initialized")
	}
	return a.bmadStorage.ListWorkflowsByRepo(repoPath)
}

// ── BMAD Templates ──

// ListBmadTemplates returns the 6 built-in workflow templates.
func (a *App) ListBmadTemplates() []bmad.WorkflowDef {
	return bmad.BuiltinTemplates()
}

// CreateFromTemplate deep-copies a built-in template into a user workflow.
func (a *App) CreateFromTemplate(templateID, repoPath string) (bmad.WorkflowDef, error) {
	if a.bmadStorage == nil {
		return bmad.WorkflowDef{}, fmt.Errorf("bmad storage not initialized")
	}

	var tpl bmad.WorkflowDef
	var found bool
	for _, t := range bmad.BuiltinTemplates() {
		if t.ID == templateID {
			tpl = t
			found = true
			break
		}
	}
	if !found {
		return bmad.WorkflowDef{}, fmt.Errorf("template %q not found", templateID)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	wf := bmad.WorkflowDef{
		ID:          fmt.Sprintf("wf-%s-%d", templateID, time.Now().UnixMilli()),
		Name:        "Copy of " + tpl.Name,
		Description: tpl.Description,
		Nodes:       make([]bmad.WorkflowNode, len(tpl.Nodes)),
		Edges:       make([]bmad.WorkflowEdge, len(tpl.Edges)),
		IsTemplate:  false,
		TemplateID:  templateID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	copy(wf.Nodes, tpl.Nodes)
	copy(wf.Edges, tpl.Edges)
	wf.RepoPath = repoPath

	if err := a.bmadStorage.SaveWorkflow(wf); err != nil {
		return bmad.WorkflowDef{}, fmt.Errorf("saving workflow from template: %w", err)
	}
	return wf, nil
}

// ── BMAD Sprint Status ──

// GetSprintStatus reads and parses the sprint-status.yaml for the given repo.
func (a *App) GetSprintStatus(repoPath string) (bmad.SprintStatus, error) {
	return bmad.ParseSprintStatus(repoPath)
}

// UpdateStoryStatus modifies a story's status in sprint-status.yaml.
func (a *App) UpdateStoryStatus(repoPath, storyID, newStatus string) error {
	return bmad.UpdateStoryStatus(repoPath, storyID, newStatus)
}

// ── BMAD Execution ──

// StartBmadWorkflow begins executing a workflow and returns the execution ID.
func (a *App) StartBmadWorkflow(workflowID, repoPath, model string) (string, error) {
	if a.bmadExecutor == nil {
		return "", fmt.Errorf("bmad executor not initialized")
	}
	exec, err := a.bmadExecutor.StartWorkflow(a.ctx, workflowID, repoPath, model)
	if err != nil {
		return "", err
	}
	return exec.ID, nil
}

// PauseBmadWorkflow pauses a running execution.
func (a *App) PauseBmadWorkflow(execID string) error {
	if a.bmadExecutor == nil {
		return fmt.Errorf("bmad executor not initialized")
	}
	return a.bmadExecutor.PauseWorkflow(execID)
}

// ResumeBmadWorkflow resumes a paused execution.
func (a *App) ResumeBmadWorkflow(execID string) error {
	if a.bmadExecutor == nil {
		return fmt.Errorf("bmad executor not initialized")
	}
	return a.bmadExecutor.ResumeWorkflow(execID)
}

// StopBmadWorkflow cancels a running execution.
func (a *App) StopBmadWorkflow(execID string) error {
	if a.bmadExecutor == nil {
		return fmt.Errorf("bmad executor not initialized")
	}
	return a.bmadExecutor.StopWorkflow(execID)
}

// RespondToQuestion injects a user's answer into the Claude CLI tmux pane for
// the given node and dispatches Enter. Used by the frontend Question modal to
// unblock a workflow that is waiting on interactive input.
func (a *App) RespondToQuestion(execID, nodeID, answer string) error {
	if a.bmadExecutor == nil {
		return fmt.Errorf("bmad executor not initialized")
	}
	return a.bmadExecutor.RespondToQuestion(execID, nodeID, answer)
}

// GetBmadExecution returns the current state of an execution.
func (a *App) GetBmadExecution(execID string) (*bmad.WorkflowExecution, error) {
	if a.bmadExecutor == nil {
		return nil, fmt.Errorf("bmad executor not initialized")
	}
	return a.bmadExecutor.GetExecution(execID)
}

// GetBmadCurrentExecution returns a deep copy of the most recently
// started non-terminal execution (running or paused) whose RepoPath
// matches the given path, or nil when no such execution exists.
//
// This is the restore-on-mount hook the WorkflowBuilder calls when a
// user re-enters a repo's workspace: if a workflow is mid-run, the
// frontend loads the associated workflow definition and repaints per-
// node statuses so the canvas shows the live execution state. Returning
// (nil, nil) for "no match" is intentional — the frontend falls through
// to its autosave-draft / blank-canvas paths without treating absence
// as an error.
func (a *App) GetBmadCurrentExecution(repoPath string) (*bmad.WorkflowExecution, error) {
	if a.bmadExecutor == nil {
		return nil, fmt.Errorf("bmad executor not initialized")
	}
	return a.bmadExecutor.GetCurrentExecution(repoPath)
}

// ── BMAD Agent Management ──

// ListBmadAgents returns all custom agent configurations.
func (a *App) ListBmadAgents() ([]bmad.BmadAgentConfig, error) {
	if a.bmadStorage == nil {
		return nil, fmt.Errorf("bmad storage not initialized")
	}
	return a.bmadStorage.ListAgents()
}

// SaveBmadAgent persists a custom agent configuration.
func (a *App) SaveBmadAgent(agent bmad.BmadAgentConfig) error {
	if a.bmadStorage == nil {
		return fmt.Errorf("bmad storage not initialized")
	}
	return a.bmadStorage.SaveAgent(agent)
}

// DeleteBmadAgent removes a custom agent by ID.
func (a *App) DeleteBmadAgent(id string) error {
	if a.bmadStorage == nil {
		return fmt.Errorf("bmad storage not initialized")
	}
	return a.bmadStorage.DeleteAgent(id)
}

// ListAllAgents returns agents grouped by source: BMAD agents, local project
// agents (from repoPath/.claude/agents/), and global agents (~/.claude/agents/).
func (a *App) ListAllAgents(repoPath string) (bmad.GroupedAgents, error) {
	var result bmad.GroupedAgents

	// BMAD agents from ~/.mashed/bmad-agents/
	if a.bmadStorage != nil {
		agents, err := a.bmadStorage.ListAgents()
		if err != nil {
			return result, err
		}
		result.BmadAgents = agents
	}

	// Local project agents from {repoPath}/.claude/agents/
	if repoPath != "" {
		result.LocalAgents = bmad.ListClaudeAgents(filepath.Join(repoPath, ".claude", "agents"))
	}

	// Global agents from ~/.claude/agents/
	home, err := os.UserHomeDir()
	if err == nil {
		result.GlobalAgents = bmad.ListClaudeAgents(filepath.Join(home, ".claude", "agents"))
	}

	return result, nil
}

// ListAllMashedAssets returns every "mashed-ready" skill and command
// discovered under the conventional Claude Code asset directories,
// grouped by {scope} × {kind}. "Mashed-ready" means the asset's YAML
// frontmatter contains a `mashedRole` field — assets without it are
// silently skipped so the sidebar only surfaces things that actually
// work in the Mashed workspace model.
//
// Four directories are scanned (nonexistent ones are not errors):
//
//	{repoPath}/.claude/skills/     → LocalSkills    (directory-per-skill)
//	{repoPath}/.claude/commands/   → LocalCommands  (flat .md files)
//	~/.claude/skills/              → GlobalSkills   (directory-per-skill)
//	~/.claude/commands/            → GlobalCommands (flat .md files)
//
// Per-file parse errors (malformed YAML, unknown role) are logged at the
// loader level and the bad file is skipped — one broken file must not
// hide every other asset in the same directory. Only I/O errors on the
// top-level directory reads propagate up as a caller-visible error, and
// even those are swallowed for nonexistent directories via the loader's
// own os.IsNotExist short-circuit.
func (a *App) ListAllMashedAssets(repoPath string) (bmad.GroupedMashedAssets, error) {
	var result bmad.GroupedMashedAssets

	if repoPath != "" {
		localSkillsDir := filepath.Join(repoPath, ".claude", "skills")
		localCommandsDir := filepath.Join(repoPath, ".claude", "commands")
		if v, err := bmad.LoadMashedAssetsFromDir(localSkillsDir, bmad.MashedKindSkill, bmad.MashedSourceLocal); err != nil {
			return result, err
		} else {
			result.LocalSkills = v
		}
		if v, err := bmad.LoadMashedAssetsFromDir(localCommandsDir, bmad.MashedKindCommand, bmad.MashedSourceLocal); err != nil {
			return result, err
		} else {
			result.LocalCommands = v
		}
	}

	home, err := os.UserHomeDir()
	if err == nil {
		globalSkillsDir := filepath.Join(home, ".claude", "skills")
		globalCommandsDir := filepath.Join(home, ".claude", "commands")
		if v, lerr := bmad.LoadMashedAssetsFromDir(globalSkillsDir, bmad.MashedKindSkill, bmad.MashedSourceGlobal); lerr != nil {
			return result, lerr
		} else {
			result.GlobalSkills = v
		}
		if v, lerr := bmad.LoadMashedAssetsFromDir(globalCommandsDir, bmad.MashedKindCommand, bmad.MashedSourceGlobal); lerr != nil {
			return result, lerr
		} else {
			result.GlobalCommands = v
		}
	}

	return result, nil
}

// ── BMAD Modules ──

// GetBmadModules returns all available BMAD modules.
func (a *App) GetBmadModules() []bmad.ModuleDef {
	return bmad.GetModules()
}

// ── BMAD Control Flow ──

// GetNodeOutput retrieves captured terminal output for a specific node in an execution.
func (a *App) GetNodeOutput(execID, nodeID string) (string, error) {
	if a.bmadExecutor == nil {
		return "", fmt.Errorf("bmad executor not initialized")
	}
	exec, err := a.bmadExecutor.GetExecution(execID)
	if err != nil {
		return "", err
	}
	return exec.NodeOutputs[nodeID], nil
}

// GetArtifactStatus checks whether a BMAD artifact exists at its expected path.
// Returns (exists, resolvedPath, error). Unmapped artifacts return (false, "", nil).
func (a *App) GetArtifactStatus(repoPath, artifactName string) (bool, string, error) {
	return bmad.GetArtifactStatus(repoPath, artifactName)
}

// GetControlFlowNodes returns the list of available control flow node types for the sidebar.
func (a *App) GetControlFlowNodes() []bmad.ControlFlowNodeDef {
	return []bmad.ControlFlowNodeDef{
		{Type: bmad.NodeTypeCondition, Name: "Condition", Description: "If/else branch based on output", Icon: "GitBranch"},
		{Type: bmad.NodeTypeLoop, Name: "Loop", Description: "Repeat N times", Icon: "Repeat"},
		{Type: bmad.NodeTypeLoopUntil, Name: "Loop Until", Description: "Repeat until condition met", Icon: "Target"},
		{Type: bmad.NodeTypeTransform, Name: "Transform", Description: "Extract/transform data", Icon: "Filter"},
		{Type: bmad.NodeTypeMerge, Name: "Merge", Description: "Join branches", Icon: "GitMerge"},
	}
}
