// Phase 4 guided-batch upgrades. Three single-pass staged-prompt processes
// using applyGuidedUpgrade from interactive_defaults.go.
package bmad

func init() {
	applyGuidedUpgrade(&registry[processIndex("bmad-create-prd")], GuidedUpgradeSpec{
		StagedInputs: []InputSpec{
			{
				ID: "scope", Source: InputFromUser, Shape: ShapeFree, Required: true,
				Prompt:    "What is the scope of this PRD?",
				HelpText:  "Describe the surface area in 1-3 sentences.",
				MaxLength: 1000,
			},
			{
				ID: "audience", Source: InputFromUser, Shape: ShapeFree, Required: true,
				Prompt:    "Who is the target audience?",
				HelpText:  "Personas, teams, or end-users.",
				MaxLength: 500,
			},
			{
				ID: "timeline", Source: InputFromUser, Shape: ShapeChoice, Required: true,
				Prompt:  "What is the rough timeline?",
				Options: []string{"days", "weeks", "months", "quarters"},
				Default: "weeks",
			},
		},
		FinalApproval: &InputSpec{
			ID: "approve", Source: InputFromUser, Shape: ShapeApproval, Required: true,
			Prompt: "Approve this PRD draft?",
		},
		OutputSpecs: []OutputSpec{
			fileOutput("PRD.md"),
		},
	})

	applyGuidedUpgrade(&registry[processIndex("bmad-document-project")], GuidedUpgradeSpec{
		StagedInputs: []InputSpec{
			{
				ID: "scope", Source: InputFromUser, Shape: ShapeChoice, Required: true,
				Prompt:  "What scope should the docs cover?",
				Options: []string{"public-api", "all", "selected-modules"},
				Default: "public-api",
			},
			{
				ID: "depth", Source: InputFromUser, Shape: ShapeChoice, Required: true,
				Prompt:  "How deep?",
				Options: []string{"summary", "reference", "tutorial"},
				Default: "reference",
			},
			{
				ID: "output-location", Source: InputFromUser, Shape: ShapeFree, Required: false,
				Prompt:   "Where should output land?",
				HelpText: "Repo-relative path; leave blank for default.",
				Default:  "project-docs",
			},
		},
		FinalApproval: &InputSpec{
			ID: "approve", Source: InputFromUser, Shape: ShapeApproval, Required: true,
			Prompt: "Approve this documentation plan?",
		},
		OutputSpecs: []OutputSpec{
			fileOutput("project-docs"),
		},
	})

	applyGuidedUpgrade(&registry[processIndex("bmad-generate-project-context")], GuidedUpgradeSpec{
		StagedInputs: []InputSpec{
			{
				ID: "scope", Source: InputFromUser, Shape: ShapeChoice, Required: true,
				Prompt:  "What context level do you need?",
				Options: []string{"onboarding", "deep-dive", "snapshot"},
				Default: "onboarding",
			},
			{
				ID: "sources", Source: InputFromUser, Shape: ShapeMultiChoice, Required: true,
				Prompt:  "Which sources should we include?",
				Options: []string{"PRD", "architecture", "epics", "code", "history"},
			},
		},
		FinalApproval: &InputSpec{
			ID: "approve", Source: InputFromUser, Shape: ShapeApproval, Required: true,
			Prompt: "Approve generated context?",
		},
		OutputSpecs: []OutputSpec{
			fileOutput("project-context.md"),
		},
	})
}
