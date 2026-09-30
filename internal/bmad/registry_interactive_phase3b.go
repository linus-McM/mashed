// Phase 3b planning-batch iterative upgrades. Same idiom as the phase 2/3a
// helper-call init blocks.
package bmad

func init() {
	applyIterativeUpgrade(&registry[processIndex("bmad-create-ux-design")], IterativeUpgradeSpec{
		Prompt:       "Refine the UX or 'done'/'approved'/'ship' to finalise.",
		HelpText:     "Iterate on flows/wireframes; 'approved' to ship the spec.",
		DomainAccept: []string{"approved", "ship"},
		MaxRounds:    15,
		OutputSpecs: []OutputSpec{
			fileOutput("ux-spec.md"),
			memoryOutput("PRD.md"),
		},
	})

	applyIterativeUpgrade(&registry[processIndex("bmad-create-architecture")], IterativeUpgradeSpec{
		Prompt:       "Iterate on architecture or 'done'/'approved'/'ship' to finalise.",
		HelpText:     "Discuss components, trade-offs; 'approved' or 'ship' to seal.",
		DomainAccept: []string{"approved", "ship"},
		MaxRounds:    15,
		OutputSpecs: []OutputSpec{
			fileOutput("architecture.md"),
		},
	})

	applyIterativeUpgrade(&registry[processIndex("bmad-check-implementation-readiness")], IterativeUpgradeSpec{
		Prompt:       "Provide readiness feedback or 'ready'/'done' when complete.",
		HelpText:     "Confirm artifacts coherent; 'ready' to proceed.",
		DomainAccept: []string{"ready"},
		MaxRounds:    10,
		OutputSpecs: []OutputSpec{
			fileOutput("readiness-report"),
			memoryOutput("architecture.md"),
			memoryOutput("any-doc"),
		},
	})

	applyIterativeUpgrade(&registry[processIndex("bmad-create-epics-and-stories")], IterativeUpgradeSpec{
		Prompt:    "Refine epics/stories or 'done'/'complete' when ready.",
		HelpText:  "Iterate on epic split, story carving; 'done' to finalise.",
		MaxRounds: 20,
		OutputSpecs: []OutputSpec{
			memoryOutput("epics/"),
		},
	})
}
