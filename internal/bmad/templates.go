package bmad

// BuiltinTemplates returns the 8 built-in workflow templates.
// Each template is a WorkflowDef with IsTemplate: true and pre-placed
// nodes connected by edges in a left-to-right layout.
func BuiltinTemplates() []WorkflowDef {
	return []WorkflowDef{
		fullProductLifecycle(),
		quickSprint(),
		architectureReview(),
		storyDevelopment(),
		prdPipeline(),
		qaAndPolish(),
		rapidPrototype(),
		fullInfra(),
	}
}

// TemplateByName returns the built-in template with the given name, if any.
func TemplateByName(name string) (WorkflowDef, bool) {
	for _, t := range BuiltinTemplates() {
		if t.Name == name {
			return t, true
		}
	}
	return WorkflowDef{}, false
}

func node(id, processID, label string, x float64) WorkflowNode {
	return WorkflowNode{
		ID:        id,
		ProcessID: processID,
		Label:     label,
		Position:  Position{X: x, Y: 200},
		Status:    NodePending,
		Config:    map[string]string{},
	}
}

func edge(id, source, target string) WorkflowEdge {
	return WorkflowEdge{ID: id, Source: source, Target: target}
}

func chain(prefix string, nodes []WorkflowNode) []WorkflowEdge {
	edges := make([]WorkflowEdge, 0, len(nodes)-1)
	for i := 0; i < len(nodes)-1; i++ {
		edges = append(edges, edge(
			prefix+"-e"+itoa(i+1),
			nodes[i].ID,
			nodes[i+1].ID,
		))
	}
	return edges
}

func itoa(n int) string {
	const digits = "0123456789"
	if n < 10 {
		return string(digits[n])
	}
	return itoa(n/10) + string(digits[n%10])
}

func fullProductLifecycle() WorkflowDef {
	nodes := []WorkflowNode{
		node("tpl-fpl-1", "bmad-brainstorming", "Brainstorming", 0),
		node("tpl-fpl-2", "bmad-product-brief", "Product Brief", 250),
		node("tpl-fpl-3", "bmad-create-prd", "Create PRD", 500),
		node("tpl-fpl-4", "bmad-validate-prd", "Validate PRD", 750),
		node("tpl-fpl-5", "bmad-create-ux-design", "UX Design", 1000),
		node("tpl-fpl-6", "bmad-create-architecture", "Architecture", 1250),
		node("tpl-fpl-7", "bmad-check-implementation-readiness", "Readiness Check", 1500),
		node("tpl-fpl-8", "bmad-create-epics-and-stories", "Epics & Stories", 1750),
		node("tpl-fpl-9", "bmad-create-story", "Create Story", 2000),
		node("tpl-fpl-10", "bmad-dev-story", "Develop Story", 2250),
		node("tpl-fpl-11", "bmad-code-review", "Code Review", 2500),
		node("tpl-fpl-12", "bmad-qa-generate-e2e-tests", "E2E Tests", 2750),
		node("tpl-fpl-13", "bmad-retrospective", "Retrospective", 3000),
	}
	return WorkflowDef{
		ID:          "tpl-full-product-lifecycle",
		Name:        "Full Product Lifecycle",
		Description: "Complete BMAD lifecycle from brainstorming through retrospective.",
		Nodes:       nodes,
		Edges:       chain("tpl-fpl", nodes),
		IsTemplate:  true,
	}
}

func quickSprint() WorkflowDef {
	nodes := []WorkflowNode{
		node("tpl-qs-1", "bmad-create-story", "Create Story", 0),
		node("tpl-qs-2", "bmad-dev-story", "Develop Story", 250),
		node("tpl-qs-3", "bmad-code-review", "Code Review", 500),
		node("tpl-qs-4", "bmad-qa-generate-e2e-tests", "E2E Tests", 750),
	}
	return WorkflowDef{
		ID:          "tpl-quick-sprint",
		Name:        "Quick Sprint",
		Description: "Fast iteration: story creation through QA testing.",
		Nodes:       nodes,
		Edges:       chain("tpl-qs", nodes),
		IsTemplate:  true,
	}
}

func architectureReview() WorkflowDef {
	nodes := []WorkflowNode{
		node("tpl-ar-1", "bmad-create-architecture", "Architecture", 0),
		node("tpl-ar-2", "bmad-check-implementation-readiness", "Readiness Check", 250),
		node("tpl-ar-3", "bmad-review-edge-case-hunter", "Edge Cases", 500),
	}
	return WorkflowDef{
		ID:          "tpl-architecture-review",
		Name:        "Architecture Review",
		Description: "Design and validate system architecture.",
		Nodes:       nodes,
		Edges:       chain("tpl-ar", nodes),
		IsTemplate:  true,
	}
}

func storyDevelopment() WorkflowDef {
	nodes := []WorkflowNode{
		node("tpl-sd-1", "bmad-create-story", "Create Story", 0),
		node("tpl-sd-2", "bmad-dev-story", "Develop Story", 250),
		node("tpl-sd-3", "bmad-code-review", "Code Review", 500),
	}
	return WorkflowDef{
		ID:          "tpl-story-development",
		Name:        "Story Development",
		Description: "Single story from creation through code review.",
		Nodes:       nodes,
		Edges:       chain("tpl-sd", nodes),
		IsTemplate:  true,
	}
}

func prdPipeline() WorkflowDef {
	nodes := []WorkflowNode{
		node("tpl-pp-1", "bmad-brainstorming", "Brainstorming", 0),
		node("tpl-pp-2", "bmad-product-brief", "Product Brief", 250),
		node("tpl-pp-3", "bmad-create-prd", "Create PRD", 500),
		node("tpl-pp-4", "bmad-validate-prd", "Validate PRD", 750),
	}
	return WorkflowDef{
		ID:          "tpl-prd-pipeline",
		Name:        "PRD Pipeline",
		Description: "From brainstorming to a validated PRD.",
		Nodes:       nodes,
		Edges:       chain("tpl-pp", nodes),
		IsTemplate:  true,
	}
}

func qaAndPolish() WorkflowDef {
	nodes := []WorkflowNode{
		node("tpl-qp-1", "bmad-code-review", "Code Review", 0),
		node("tpl-qp-2", "bmad-qa-generate-e2e-tests", "E2E Tests", 250),
		node("tpl-qp-3", "bmad-editorial-review-prose", "Editorial Review", 500),
	}
	return WorkflowDef{
		ID:          "tpl-qa-and-polish",
		Name:        "QA & Polish",
		Description: "Code review, end-to-end testing, and editorial polish.",
		Nodes:       nodes,
		Edges:       chain("tpl-qp", nodes),
		IsTemplate:  true,
	}
}

func rapidPrototype() WorkflowDef {
	nodes := []WorkflowNode{
		node("tpl-rp-1", "bmad-party-mode", "Party Mode", 0),
		node("tpl-rp-2", "bmad-code-review", "Code Review", 250),
	}
	return WorkflowDef{
		ID:          "tpl-rapid-prototype",
		Name:        "Rapid Prototype",
		Description: "Quick prototype with party mode followed by code review.",
		Nodes:       nodes,
		Edges:       chain("tpl-rp", nodes),
		IsTemplate:  true,
	}
}

func fullInfra() WorkflowDef {
	nodes := []WorkflowNode{
		node("tpl-fi-1", "bmad-create-architecture", "Architecture", 0),
		node("tpl-fi-2", "bmad-infrastructure-devops", "Infrastructure & DevOps", 250),
		node("tpl-fi-3", "bmad-code-review", "Code Review", 500),
	}
	return WorkflowDef{
		ID:          "tpl-full-infra",
		Name:        "Full Infrastructure",
		Description: "Architecture design through infrastructure setup and code review.",
		Nodes:       nodes,
		Edges:       chain("tpl-fi", nodes),
		IsTemplate:  true,
	}
}
