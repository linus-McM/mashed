package bmad

// registry is the in-memory catalog of all BMAD processes.
// Initialized once in init() and never mutated afterward.
var registry []ProcessDef

// testRegistry is a secondary catalog populated only by test-code init()
// functions. ProcessByID consults it after `registry` so production code
// paths behave identically whether tests register fixtures or not.
// Production builds see an empty slice — zero overhead.
var testRegistry []ProcessDef

func init() {
	registry = []ProcessDef{
		// ── Analysis (5) ──
		{
			ID:          "bmad-brainstorming",
			Name:        "Brainstorming",
			Phase:       PhaseAnalysis,
			AgentRole:   RoleAnalyst,
			SkillName:   "bmad-brainstorming",
			Description: "Facilitate a structured brainstorming session to explore ideas and possibilities.",
			Inputs:      []string{},
			Outputs:     []string{"brainstorm-notes"},
			ModuleID:    "core",
			Version:     "1.0.0",
			// Interactive shape — schema §10.1.
			Mode:             InteractIterative,
			EnableAstAdapter: true,
			InputSpecs: []InputSpec{
				{ID: "topic", Source: InputFromUser, Shape: ShapeFree, Required: true, Prompt: "What topic do you want to brainstorm?", MaxLength: 500},
				{ID: "approach", Source: InputFromUser, Shape: ShapeChoice, Required: true, Prompt: "How should we pick techniques?", Options: []string{"user-pick", "ai-recommend", "random", "progressive"}},
				{ID: "technique", Source: InputFromRegistry, OptionsRef: "registry:brain-methods.csv#technique_name"},
				{ID: "round-response", Source: InputFromUser, Shape: ShapeJSON, Prompt: "Add ideas, pivot, or type 'done' when satisfied.", HelpText: "Type 'done' to wrap up; 'skip' to move to the next technique."},
			},
			OutputSpecs: []OutputSpec{
				{ID: "brainstorm-notes", Target: OutputToFile, ArtifactName: "brainstorm-notes", Description: "Organised brainstorm session notes"},
			},
			Gate: &IterationGate{Kind: GateUserConfirm, MaxRounds: 30, AcceptTokens: []string{"done", "wrap up", "finish"}, RejectTokens: []string{"abort", "cancel"}},
		},
		{
			ID:          "bmad-product-brief",
			Name:        "Product Brief",
			Phase:       PhaseAnalysis,
			AgentRole:   RoleAnalyst,
			SkillName:   "bmad-product-brief",
			Description: "Create a concise product brief capturing vision, goals, and target audience.",
			Inputs:      []string{},
			Outputs:     []string{"product-brief"},
			ModuleID:    "core",
			Version:     "1.0.0",
			// Interactive shape — schema §10.2.
			Mode:             InteractGuided,
			EnableAstAdapter: true,
			InputSpecs: []InputSpec{
				{ID: "mode", Source: InputFromUser, Shape: ShapeChoice, Required: true, Prompt: "How do you want to work?", Options: []string{"guided", "yolo", "autonomous"}, Default: "guided"},
				{ID: "existing-brief", Source: InputFromFile, ArtifactName: "product-brief"},
				{ID: "brainstorm-input", Source: InputFromUpstream},
				{ID: "stage-response", Source: InputFromUser, Shape: ShapeJSON, Prompt: "{{stage_prompt}}", HelpText: "Type 'skip' to move on without more detail."},
				{ID: "final-approval", Source: InputFromUser, Shape: ShapeApproval, Required: true, Prompt: "Approve this brief?"},
			},
			OutputSpecs: []OutputSpec{
				{ID: "product-brief", Target: OutputToFile, ArtifactName: "product-brief"},
			},
			Gate: &IterationGate{Kind: GateUserConfirm, MaxRounds: 10, AcceptTokens: []string{"yes"}},
		},
		{
			ID:          "bmad-domain-research",
			Name:        "Domain Research",
			Phase:       PhaseAnalysis,
			AgentRole:   RoleAnalyst,
			SkillName:   "bmad-domain-research",
			Description: "Research the problem domain to understand context, terminology, and constraints.",
			Inputs:      []string{},
			Outputs:     []string{"domain-research"},
			ModuleID:    "core",
			Version:     "1.0.0",
		},
		{
			ID:          "bmad-market-research",
			Name:        "Market Research",
			Phase:       PhaseAnalysis,
			AgentRole:   RoleAnalyst,
			SkillName:   "bmad-market-research",
			Description: "Analyze the competitive landscape and market opportunity.",
			Inputs:      []string{},
			Outputs:     []string{"market-research"},
			ModuleID:    "core",
			Version:     "1.0.0",
		},
		{
			ID:          "bmad-technical-research",
			Name:        "Technical Research",
			Phase:       PhaseAnalysis,
			AgentRole:   RoleArchitect,
			SkillName:   "bmad-technical-research",
			Description: "Investigate technical approaches, libraries, and architectural options.",
			Inputs:      []string{},
			Outputs:     []string{"tech-research"},
			ModuleID:    "core",
			Version:     "1.0.0",
		},

		// ── Planning (4) ──
		{
			ID:          "bmad-create-prd",
			Name:        "Create PRD",
			Phase:       PhasePlanning,
			AgentRole:   RolePM,
			SkillName:   "bmad-create-prd",
			Description: "Draft a product requirements document from the product brief.",
			Inputs:      []string{"product-brief"},
			Outputs:     []string{"PRD.md"},
			ModuleID:    "core",
			Version:     "1.0.0",
		},
		{
			ID:          "bmad-edit-prd",
			Name:        "Edit PRD",
			Phase:       PhasePlanning,
			AgentRole:   RolePM,
			SkillName:   "bmad-edit-prd",
			Description: "Revise and refine an existing PRD based on feedback.",
			Inputs:      []string{"PRD.md"},
			Outputs:     []string{"PRD.md"},
			ModuleID:    "core",
			Version:     "1.0.0",
		},
		{
			ID:          "bmad-validate-prd",
			Name:        "Validate PRD",
			Phase:       PhasePlanning,
			AgentRole:   RolePM,
			SkillName:   "bmad-validate-prd",
			Description: "Validate the PRD for completeness, consistency, and feasibility.",
			Inputs:      []string{"PRD.md"},
			Outputs:     []string{"prd-validation", "PRD.md"},
			ModuleID:    "core",
			Version:     "1.0.0",
		},
		{
			ID:          "bmad-create-ux-design",
			Name:        "Create UX Design",
			Phase:       PhasePlanning,
			AgentRole:   RoleUXDesigner,
			SkillName:   "bmad-create-ux-design",
			Description: "Design the user experience including flows, wireframes, and interaction patterns.",
			Inputs:      []string{"PRD.md"},
			Outputs:     []string{"ux-spec.md", "PRD.md"},
			ModuleID:    "core",
			Version:     "1.0.0",
		},

		// ── Solutioning (4) ──
		{
			ID:          "bmad-create-architecture",
			Name:        "Create Architecture",
			Phase:       PhaseSolutioning,
			AgentRole:   RoleArchitect,
			SkillName:   "bmad-create-architecture",
			Description: "Design the system architecture based on PRD requirements.",
			Inputs:      []string{"PRD.md"},
			Outputs:     []string{"architecture.md"},
			ModuleID:    "core",
			Version:     "1.0.0",
		},
		{
			ID:          "bmad-check-implementation-readiness",
			Name:        "Check Implementation Readiness",
			Phase:       PhaseSolutioning,
			AgentRole:   RoleArchitect,
			SkillName:   "bmad-check-implementation-readiness",
			Description: "Verify that all artifacts are complete and consistent before implementation begins.",
			Inputs:      []string{"architecture.md"},
			Outputs:     []string{"readiness-report", "architecture.md", "any-doc"},
			ModuleID:    "core",
			Version:     "1.0.0",
		},
		{
			ID:          "bmad-create-epics-and-stories",
			Name:        "Create Epics and Stories",
			Phase:       PhaseSolutioning,
			AgentRole:   RolePM,
			SkillName:   "bmad-create-epics-and-stories",
			Description: "Break down the PRD and architecture into epics and user stories.",
			Inputs:      []string{"PRD.md", "architecture.md"},
			Outputs:     []string{"epics/"},
			ModuleID:    "core",
			Version:     "1.0.0",
		},
		{
			ID:          "bmad-generate-project-context",
			Name:        "Generate Project Context",
			Phase:       PhaseSolutioning,
			AgentRole:   RoleArchitect,
			SkillName:   "bmad-generate-project-context",
			Description: "Generate a consolidated project context document for onboarding and reference.",
			Inputs:      []string{"PRD.md", "architecture.md"},
			Outputs:     []string{"project-context.md"},
			ModuleID:    "core",
			Version:     "1.0.0",
		},

		// ── Implementation (8) ──
		{
			ID:          "bmad-create-story",
			Name:        "Create Story",
			Phase:       PhaseImplementation,
			AgentRole:   RoleDeveloper,
			SkillName:   "bmad-create-story",
			Description: "Create a detailed implementation story from epics with tasks and acceptance criteria.",
			Inputs:      []string{"epics/"},
			Outputs:     []string{"story-*.md"},
			ModuleID:    "core",
			Version:     "1.0.0",
		},
		{
			ID:          "bmad-dev-story",
			Name:        "Develop Story",
			Phase:       PhaseImplementation,
			AgentRole:   RoleDeveloper,
			SkillName:   "bmad-dev-story",
			Description: "Implement a story by writing code and tests according to acceptance criteria.",
			Inputs:      []string{"story-*.md"},
			Outputs:     []string{"code", "tests"},
			ModuleID:    "core",
			Version:     "1.0.0",
		},
		{
			ID:          "bmad-quick-dev",
			Name:        "Quick Dev",
			Phase:       PhaseImplementation,
			AgentRole:   RoleDeveloper,
			SkillName:   "bmad-quick-dev",
			Description: "Rapid development of small features or fixes without a formal story.",
			Inputs:      []string{},
			Outputs:     []string{"code"},
			ModuleID:    "core",
			Version:     "1.0.0",
		},
		{
			ID:          "bmad-sprint-planning",
			Name:        "Sprint Planning",
			Phase:       PhaseImplementation,
			AgentRole:   RolePM,
			SkillName:   "bmad-sprint-planning",
			Description: "Plan a sprint by selecting and prioritizing stories from the backlog.",
			Inputs:      []string{"epics/"},
			Outputs:     []string{"sprint-status.yaml"},
			ModuleID:    "core",
			Version:     "1.0.0",
		},
		{
			ID:          "bmad-sprint-status",
			Name:        "Sprint Status",
			Phase:       PhaseImplementation,
			AgentRole:   RolePM,
			SkillName:   "bmad-sprint-status",
			Description: "Update and report on current sprint progress and blockers.",
			Inputs:      []string{"sprint-status.yaml"},
			Outputs:     []string{"sprint-status.yaml"},
			ModuleID:    "core",
			Version:     "1.0.0",
		},
		{
			ID:          "bmad-code-review",
			Name:        "Code Review",
			Phase:       PhaseImplementation,
			AgentRole:   RoleDeveloper,
			SkillName:   "bmad-code-review",
			Description: "Review code changes for quality, correctness, and adherence to standards.",
			Inputs:      []string{},
			Outputs:     []string{"review-report", "code"},
			ModuleID:    "core",
			Version:     "1.0.0",
		},
		{
			ID:          "bmad-qa-generate-e2e-tests",
			Name:        "QA Generate E2E Tests",
			Phase:       PhaseImplementation,
			AgentRole:   RoleQA,
			SkillName:   "bmad-qa-generate-e2e-tests",
			Description: "Generate end-to-end tests from stories and existing code.",
			Inputs:      []string{"code", "story-*.md"},
			Outputs:     []string{"tests", "any-doc"},
			ModuleID:    "core",
			Version:     "1.0.0",
		},
		{
			ID:          "bmad-retrospective",
			Name:        "Retrospective",
			Phase:       PhaseImplementation,
			AgentRole:   RolePM,
			SkillName:   "bmad-retrospective",
			Description: "Conduct a sprint retrospective to identify improvements.",
			Inputs:      []string{},
			Outputs:     []string{"retro-notes"},
			ModuleID:    "core",
			Version:     "1.0.0",
		},

		// ── Support (4) ──
		{
			ID:          "bmad-editorial-review-prose",
			Name:        "Editorial Review (Prose)",
			Phase:       PhaseSupport,
			AgentRole:   RoleTechWriter,
			SkillName:   "bmad-editorial-review-prose",
			Description: "Review document prose for clarity, tone, and readability.",
			Inputs:      []string{"any-doc"},
			Outputs:     []string{"reviewed-doc"},
			ModuleID:    "core",
			Version:     "1.0.0",
		},
		{
			ID:          "bmad-editorial-review-structure",
			Name:        "Editorial Review (Structure)",
			Phase:       PhaseSupport,
			AgentRole:   RoleTechWriter,
			SkillName:   "bmad-editorial-review-structure",
			Description: "Review document structure for organization, completeness, and consistency.",
			Inputs:      []string{"any-doc"},
			Outputs:     []string{"reviewed-doc"},
			ModuleID:    "core",
			Version:     "1.0.0",
		},
		{
			ID:          "bmad-advanced-elicitation",
			Name:        "Advanced Elicitation",
			Phase:       PhaseSupport,
			AgentRole:   RoleAnalyst,
			SkillName:   "bmad-advanced-elicitation",
			Description: "Use structured elicitation techniques to extract requirements and knowledge.",
			Inputs:      []string{},
			Outputs:     []string{"elicitation-notes"},
			ModuleID:    "core",
			Version:     "1.0.0",
			// Interactive shape — schema §10.4.
			Mode:             InteractIterative,
			EnableAstAdapter: true,
			InputSpecs: []InputSpec{
				{ID: "target-content", Source: InputFromUpstream, Required: true},
				// method recurs every round — marked Required=false so iterationInput() picks it.
				// §10.4 JSON shows required=true for schema docs; the executor treats it as the
				// per-round recurring slot per the prose under §10.4 and AC-2.
				{ID: "method", Source: InputFromUser, Shape: ShapeJSON, Prompt: "Pick a reasoning method:", OptionsRef: "registry:methods.csv?random=5", HelpText: "[r] reshuffle · [a] see all · [x] accept and proceed"},
				{ID: "apply-changes", Source: InputFromUser, Shape: ShapeApproval, Required: true, Prompt: "Apply these changes to the document?"},
			},
			OutputSpecs: []OutputSpec{
				{ID: "elicitation-notes", Target: OutputToBoth, ArtifactName: "elicitation-notes"},
			},
			Gate: &IterationGate{Kind: GateUserConfirm, MaxRounds: 20, AcceptTokens: []string{"x", "proceed", "done"}},
		},
		{
			ID:          "bmad-review-edge-case-hunter",
			Name:        "Review Edge Case Hunter",
			Phase:       PhaseSupport,
			AgentRole:   RoleQA,
			SkillName:   "bmad-review-edge-case-hunter",
			Description: "Hunt for edge cases, contradictions, and missing requirements in documents.",
			Inputs:      []string{"any-doc"},
			Outputs:     []string{"edge-case-report"},
			ModuleID:    "core",
			Version:     "1.0.0",
		},

		// ── Sprint 4: New Processes (7) ──
		{
			ID:          "bmad-party-mode",
			Name:        "Party Mode",
			Phase:       PhaseImplementation,
			AgentRole:   RoleDeveloper,
			SkillName:   "bmad-party-mode",
			Description: "Rapid prototyping with minimal ceremony — jump straight into code.",
			Inputs:      []string{},
			Outputs:     []string{"code"},
			ModuleID:    "core",
			Version:     "1.0.0",
			// Interactive shape — schema §10.3.
			Mode:             InteractParty,
			EnableAstAdapter: true,
			InputSpecs: []InputSpec{
				{ID: "topic", Source: InputFromUser, Shape: ShapeFree, Required: true, Prompt: "What do you want the team to discuss?", MaxLength: 1000},
				{ID: "message", Source: InputFromUser, Shape: ShapeJSON, Prompt: "Your turn. Type 'exit' to end the conversation."},
			},
			OutputSpecs: []OutputSpec{
				{ID: "transcript", Target: OutputToFile, ArtifactName: "retro-notes", Optional: true},
			},
			Gate: &IterationGate{Kind: GateUserConfirm, MaxRounds: 100, AcceptTokens: []string{"exit", "done", "wrap up"}},
		},
		{
			ID:          "bmad-quick-flow",
			Name:        "Quick Flow",
			Phase:       PhaseAnalysis,
			AgentRole:   RoleDeveloper,
			SkillName:   "bmad-quick-flow",
			Description: "Lightweight analysis that produces both code and a PRD in one pass.",
			Inputs:      []string{},
			Outputs:     []string{"code", "PRD.md"},
			ModuleID:    "core",
			Version:     "1.0.0",
		},
		{
			ID:          "bmad-adversarial-general",
			Name:        "Adversarial Review",
			Phase:       PhaseSupport,
			AgentRole:   RoleQA,
			SkillName:   "bmad-adversarial-general",
			Description: "Adversarial review to stress-test documents for weaknesses and blind spots.",
			Inputs:      []string{"any-doc"},
			Outputs:     []string{"adversarial-report"},
			ModuleID:    "core",
			Version:     "1.0.0",
		},
		{
			ID:          "bmad-infrastructure-devops",
			Name:        "Infrastructure & DevOps",
			Phase:       PhaseImplementation,
			AgentRole:   RoleArchitect,
			SkillName:   "bmad-infrastructure-devops",
			Description: "Design and configure infrastructure, CI/CD pipelines, and deployment automation.",
			Inputs:      []string{"architecture.md"},
			Outputs:     []string{"infra-config"},
			ModuleID:    "cis",
			Version:     "1.0.0",
		},
		{
			ID:          "bmad-document-project",
			Name:        "Document Project",
			Phase:       PhaseSupport,
			AgentRole:   RoleTechWriter,
			SkillName:   "bmad-document-project",
			Description: "Generate comprehensive project documentation from existing artifacts.",
			Inputs:      []string{"PRD.md", "architecture.md"},
			Outputs:     []string{"project-docs"},
			ModuleID:    "core",
			Version:     "1.0.0",
		},
		{
			ID:          "bmad-web-orchestrator",
			Name:        "Web Orchestrator",
			Phase:       PhaseImplementation,
			AgentRole:   RoleDeveloper,
			SkillName:   "bmad-web-orchestrator",
			Description: "Orchestrate full-stack web application development from architecture and epics.",
			Inputs:      []string{"architecture.md", "epics/"},
			Outputs:     []string{"code"},
			ModuleID:    "core",
			Version:     "1.0.0",
		},
		{
			ID:          "bmad-game-dev-studio",
			Name:        "Game Dev Studio",
			Phase:       PhaseImplementation,
			AgentRole:   RoleDeveloper,
			SkillName:   "bmad-game-dev-studio",
			Description: "Game development studio workflow for building games from a PRD.",
			Inputs:      []string{"PRD.md"},
			Outputs:     []string{"code"},
			ModuleID:    "bmgd",
			Version:     "1.0.0",
		},
		// ── Utilities ──
		{
			ID:          "util-file-loader",
			Name:        "File Loader",
			Phase:       PhaseUtilities,
			AgentRole:   "",
			SkillName:   "",
			Description: "Open a file picker dialog and output the selected file path for downstream processes to read.",
			Inputs:      []string{},
			Outputs:     []string{"file-path"},
			ModuleID:    "core",
			Version:     "1.0.0",
		},
		{
			ID:          "util-multi-file-loader",
			Name:        "Multi File Loader",
			Phase:       PhaseUtilities,
			AgentRole:   "",
			SkillName:   "",
			Description: "Emit one output path per configured {label, path} entry for downstream consumption.",
			Inputs:      []string{},
			Outputs:     []string{},
			ModuleID:    "core",
			Version:     "1.0.0",
		},
	}
}

// AllProcesses returns a copy of every process in the registry.
func AllProcesses() []ProcessDef {
	out := make([]ProcessDef, len(registry))
	copy(out, registry)
	return out
}

// ProcessesByPhase returns all processes belonging to the given phase.
func ProcessesByPhase(phase BmadPhase) []ProcessDef {
	var out []ProcessDef
	for _, p := range registry {
		if p.Phase == phase {
			out = append(out, p)
		}
	}
	return out
}

// ProcessByID looks up a single process by its unique ID.
func ProcessByID(id string) (ProcessDef, bool) {
	for _, p := range registry {
		if p.ID == id {
			return p, true
		}
	}
	for _, p := range testRegistry {
		if p.ID == id {
			return p, true
		}
	}
	return ProcessDef{}, false
}

// ProcessesByModule returns all processes belonging to the given module.
func ProcessesByModule(moduleID string) []ProcessDef {
	var out []ProcessDef
	for _, p := range registry {
		if p.ModuleID == moduleID {
			out = append(out, p)
		}
	}
	return out
}
