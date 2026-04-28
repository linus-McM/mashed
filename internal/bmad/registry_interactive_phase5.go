// Phase 5 party-batch upgrades. Three multi-agent free-form processes using
// applyPartyUpgrade from interactive_defaults.go.
package bmad

func init() {
	applyPartyUpgrade(&registry[processIndex("bmad-retrospective")], PartyUpgradeSpec{
		Topic: InputSpec{
			ID: "topic", Source: InputFromUser, Shape: ShapeFree, Required: true,
			Prompt:    "What do you want the retrospective to focus on?",
			HelpText:  "Sprint scope, blockers to discuss, or 'general'.",
			MaxLength: 1000,
		},
		RoundPrompt: "Your turn. Type 'done' or 'exit' to end the retrospective.",
		HelpText:    "[exit] to wrap up · [done] to close the round.",
		OutputSpecs: []OutputSpec{
			{ID: "retro-notes", Target: OutputToFile, ArtifactName: "retro-notes", Optional: true},
		},
	})

	applyPartyUpgrade(&registry[processIndex("bmad-web-orchestrator")], PartyUpgradeSpec{
		Topic: InputSpec{
			ID: "topic", Source: InputFromUser, Shape: ShapeFree, Required: true,
			Prompt:    "What feature should the web team build?",
			HelpText:  "Reference architecture/epics; team will decompose.",
			MaxLength: 1500,
		},
		RoundPrompt: "Your turn. Type 'done' or 'exit' to end the session.",
		HelpText:    "[exit] to wrap up · [done] to close the round.",
		OutputSpecs: []OutputSpec{
			memoryOutput("code"),
		},
	})

	applyPartyUpgrade(&registry[processIndex("bmad-game-dev-studio")], PartyUpgradeSpec{
		Topic: InputSpec{
			ID: "topic", Source: InputFromUser, Shape: ShapeFree, Required: true,
			Prompt:    "What game / feature is the studio building?",
			HelpText:  "Reference the PRD; the studio will branch into specialists.",
			MaxLength: 1500,
		},
		RoundPrompt: "Your turn. Type 'done' or 'exit' to end the session.",
		HelpText:    "[exit] to wrap up · [done] to close the round.",
		OutputSpecs: []OutputSpec{
			memoryOutput("code"),
		},
	})
}
