// Phase 3a analysis-batch iterative upgrades. Same idiom as
// registry_interactive_phase2.go — see that file for design rationale.
package bmad

func init() {
	applyIterativeUpgrade(&registry[processIndex("bmad-domain-research")], IterativeUpgradeSpec{
		Prompt:    "Add findings or 'done'/'complete' when the domain notes are ready.",
		HelpText:  "Iterate on terminology, constraints; 'done' to close.",
		MaxRounds: 30,
		OutputSpecs: []OutputSpec{
			fileOutput("domain-research"),
		},
	})

	applyIterativeUpgrade(&registry[processIndex("bmad-market-research")], IterativeUpgradeSpec{
		Prompt:    "Add competitive insights or 'done'/'complete' when the market notes are ready.",
		HelpText:  "Iterate on competitors, positioning; 'done' to close.",
		MaxRounds: 30,
		OutputSpecs: []OutputSpec{
			fileOutput("market-research"),
		},
	})

	applyIterativeUpgrade(&registry[processIndex("bmad-technical-research")], IterativeUpgradeSpec{
		Prompt:    "Add technical findings or 'done'/'complete' when the tech notes are ready.",
		HelpText:  "Iterate on libraries, trade-offs; 'done' to close.",
		MaxRounds: 30,
		OutputSpecs: []OutputSpec{
			fileOutput("tech-research"),
		},
	})
}
