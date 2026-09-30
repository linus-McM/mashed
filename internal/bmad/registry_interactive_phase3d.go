// Phase 3d support-batch iterative upgrades. Same idiom as the phase
// 2/3a/3b/3c helper-call init blocks.
package bmad

func init() {
	applyIterativeUpgrade(&registry[processIndex("bmad-editorial-review-prose")], IterativeUpgradeSpec{
		Prompt:       "Add prose feedback or 'approved'/'done' when satisfied.",
		HelpText:     "Iterate on tone/clarity; 'approved' to ship.",
		DomainAccept: []string{"approved"},
		MaxRounds:    15,
		OutputSpecs: []OutputSpec{
			fileOutput("reviewed-doc"),
		},
	})

	applyIterativeUpgrade(&registry[processIndex("bmad-editorial-review-structure")], IterativeUpgradeSpec{
		Prompt:       "Add structural feedback or 'approved'/'done' when satisfied.",
		HelpText:     "Iterate on organisation/sections; 'approved' to ship.",
		DomainAccept: []string{"approved"},
		MaxRounds:    15,
		OutputSpecs: []OutputSpec{
			fileOutput("reviewed-doc"),
		},
	})

	applyIterativeUpgrade(&registry[processIndex("bmad-review-edge-case-hunter")], IterativeUpgradeSpec{
		Prompt:       "Add edge cases or 'done'/'approved' when the report is complete.",
		HelpText:     "Iterate on contradictions/missing reqs; 'done' to close.",
		DomainAccept: []string{"approved"},
		MaxRounds:    15,
		OutputSpecs: []OutputSpec{
			fileOutput("edge-case-report"),
		},
	})

	applyIterativeUpgrade(&registry[processIndex("bmad-quick-flow")], IterativeUpgradeSpec{
		Prompt:       "Iterate on the flow or 'done'/'ship' when ready.",
		HelpText:     "Lightweight loop; 'ship' to finalise.",
		DomainAccept: []string{"ship"},
		MaxRounds:    10,
		OutputSpecs: []OutputSpec{
			memoryOutput("code"),
			fileOutput("PRD.md"),
		},
	})

	applyIterativeUpgrade(&registry[processIndex("bmad-adversarial-general")], IterativeUpgradeSpec{
		Prompt:    "Add adversarial findings or 'done'/'complete' when the review is closed.",
		HelpText:  "Iterate on weaknesses/blind spots; 'done' to close.",
		MaxRounds: 15,
		OutputSpecs: []OutputSpec{
			fileOutput("adversarial-report"),
		},
	})

	applyIterativeUpgrade(&registry[processIndex("bmad-infrastructure-devops")], IterativeUpgradeSpec{
		Prompt:       "Iterate on infra/CI or 'done'/'ship' when ready.",
		HelpText:     "Discuss pipeline/deploy; 'ship' to finalise.",
		DomainAccept: []string{"ship"},
		MaxRounds:    15,
		OutputSpecs: []OutputSpec{
			fileOutput("infra-config"),
		},
	})
}
