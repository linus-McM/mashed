package bmad

// Test-only registration of interactive ProcessDefs referenced by
// executor_interactive_test.go. Populates testRegistry (a secondary table
// consulted by ProcessByID AFTER the main registry) so that registry-count
// assertions in production tests remain unaffected.
func init() {
	testRegistry = append(testRegistry,
		ProcessDef{
			ID:          "test-interact-guided",
			Name:        "Test Interactive Guided",
			Mode:        InteractGuided,
			SkillName:   "test-skill",
			InputSpecs:  []InputSpec{},
			OutputSpecs: []OutputSpec{{ID: "out1", Target: OutputToMemory}},
		},
		ProcessDef{
			ID:          "test-interact-iterative",
			Name:        "Test Interactive Iterative",
			Mode:        InteractIterative,
			SkillName:   "test-skill",
			InputSpecs:  []InputSpec{},
			OutputSpecs: []OutputSpec{{ID: "out1", Target: OutputToMemory}},
		},
		ProcessDef{
			ID:          "test-interact-party",
			Name:        "Test Interactive Party",
			Mode:        InteractParty,
			SkillName:   "test-skill",
			InputSpecs:  []InputSpec{},
			OutputSpecs: []OutputSpec{{ID: "out1", Target: OutputToMemory}},
		},
	)
}
