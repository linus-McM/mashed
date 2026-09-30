// Phase 2 high-traffic iterative upgrades: registry mutations applied via
// applyIterativeUpgrade after registry.go's var-init phase populates the
// `registry` slice literal. See docs/stories/bmad-rollout-02-* for the
// per-process spec table.
package bmad

func init() {
	applyIterativeUpgrade(&registry[processIndex("bmad-dev-story")], IterativeUpgradeSpec{
		Prompt:       "Provide feedback or type 'done' when the implementation is ready.",
		HelpText:     "Reply with feedback for the next round; 'done' to ship.",
		DomainAccept: []string{"ship", "approved"},
		MaxRounds:    10,
		OutputSpecs: []OutputSpec{
			memoryOutput("code"),
			memoryOutput("tests"),
		},
	})

	applyIterativeUpgrade(&registry[processIndex("bmad-code-review")], IterativeUpgradeSpec{
		Prompt:       "Reply with notes, or 'approved'/'done' to finish the review.",
		HelpText:     "Type 'approved' or 'done' when the changes look right.",
		DomainAccept: []string{"ship", "approved"},
		MaxRounds:    10,
		OutputSpecs: []OutputSpec{
			fileOutput("review-report"),
			memoryOutput("code"),
		},
	})

	applyIterativeUpgrade(&registry[processIndex("bmad-create-story")], IterativeUpgradeSpec{
		Prompt:    "Refine the story or type 'done' when the draft is ready.",
		HelpText:  "Iterate on tasks/AC; 'done' to finalise.",
		MaxRounds: 15,
		OutputSpecs: []OutputSpec{
			memoryOutput("story-*.md"),
		},
	})

	applyIterativeUpgrade(&registry[processIndex("bmad-validate-prd")], IterativeUpgradeSpec{
		Prompt:       "Provide validation feedback or type 'pass'/'approved' when done.",
		HelpText:     "Use 'pass' or 'approved' to finish validation.",
		DomainAccept: []string{"pass", "approved"},
		MaxRounds:    10,
		OutputSpecs: []OutputSpec{
			fileOutput("prd-validation"),
			// PRD.md is OutputToMemory not OutputToFile — the executor reads
			// the live PRD.md from disk on resume rather than persisting a
			// per-round memory copy keyed by artifact name.
			memoryOutput("PRD.md"),
		},
	})

	applyIterativeUpgrade(&registry[processIndex("bmad-edit-prd")], IterativeUpgradeSpec{
		Prompt:       "Edit instructions or 'done'/'apply'/'ship' to commit changes.",
		HelpText:     "Send edit notes; 'apply' or 'ship' to finalise.",
		DomainAccept: []string{"apply", "ship"},
		MaxRounds:    15,
		OutputSpecs: []OutputSpec{
			fileOutput("PRD.md"),
		},
	})

	applyIterativeUpgrade(&registry[processIndex("bmad-quick-dev")], IterativeUpgradeSpec{
		Prompt:       "Iterate on the change or 'done'/'ship' to finish.",
		HelpText:     "Lightweight loop; 'ship' to finalise.",
		DomainAccept: []string{"ship"},
		MaxRounds:    10,
		OutputSpecs: []OutputSpec{
			memoryOutput("code"),
		},
	})
}
