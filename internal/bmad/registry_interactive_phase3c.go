// Phase 3c implementation-batch iterative upgrade. Same idiom as the phase
// 2/3a/3b helper-call init blocks.
package bmad

func init() {
	applyIterativeUpgrade(&registry[processIndex("bmad-qa-generate-e2e-tests")], IterativeUpgradeSpec{
		Prompt:    "Iterate on the test suite or 'done'/'complete' when satisfied.",
		HelpText:  "Add cases, fix flakes; 'done' to finalise.",
		MaxRounds: 15,
		OutputSpecs: []OutputSpec{
			memoryOutput("tests"),
			memoryOutput("any-doc"),
		},
	})
}
