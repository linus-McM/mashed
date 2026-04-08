package main

import (
	"fmt"
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
	exec, err := a.bmadExecutor.StartWorkflow(workflowID, repoPath, model)
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

// GetBmadExecution returns the current state of an execution.
func (a *App) GetBmadExecution(execID string) (*bmad.WorkflowExecution, error) {
	if a.bmadExecutor == nil {
		return nil, fmt.Errorf("bmad executor not initialized")
	}
	return a.bmadExecutor.GetExecution(execID)
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
