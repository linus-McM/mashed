// Package bmad — interactive_defaults.go provides three idempotent helpers
// that upgrade an autonomous ProcessDef into one of the four-pillar interactive
// shapes (iterative, guided, party).
//
// Idempotency contract: if def.Mode is already populated the helper is a no-op.
package bmad

// commonAcceptTokens are the baseline accept-tokens prepended to every
// iterative process's IterationGate. Domain-specific tokens supplied via
// IterativeUpgradeSpec.DomainAccept are appended after these (dedup against
// the baseline preserves baseline ordering — see applyIterativeUpgrade).
//
// Note: this baseline says "complete", whereas the legacy brainstorming
// reference (registry.go) uses "finish". The idempotency guard preserves
// brainstorming's "finish" untouched; new iterative processes upgraded via
// these helpers honour "complete" instead, per the rollout plan baseline.
var commonAcceptTokens = []string{"done", "wrap up", "complete"}

// commonRejectTokens are the baseline reject-tokens used as-is on every
// iterative process. Reject tokens are not domain-extensible in this rollout.
var commonRejectTokens = []string{"abort", "cancel"}

// defaultGuidedAccept and defaultPartyAccept are the per-mode accept-token
// fallbacks used when the spec leaves AcceptTokens unset.
var (
	defaultGuidedAccept = []string{"yes"}
	defaultPartyAccept  = []string{"exit", "done", "wrap up"}
)

// Default round caps per interaction mode (used when the spec leaves
// MaxRounds == 0). Centralised here so the three helpers stay aligned with
// the story spec without sprinkling magic numbers.
const (
	defaultIterativeMaxRounds = 30
	defaultGuidedMaxRounds    = 10
	defaultPartyMaxRounds     = 100
)

// roundResponseMaxLength caps the per-round user input on auto-generated
// recurring slots (round-response, party message). Tighter than the
// 64 KiB JSON default in validate.go — keeps modal payloads modest while
// allowing several paragraphs of free-form feedback per round.
const roundResponseMaxLength = 4000

// IterativeUpgradeSpec carries the domain-specific knobs callers need to
// supply when upgrading a ProcessDef into an iterative shape. The helper
// applies sensible defaults (MaxRounds=30, baseline accept/reject tokens,
// ShapeJSON round-response slot) so callers only override what they care
// about.
type IterativeUpgradeSpec struct {
	// Prompt is rendered on the recurring round-response InputSpec and
	// surfaces in the InputResponseModal between rounds.
	Prompt string
	// HelpText is the secondary copy beneath the round-response prompt.
	HelpText string
	// DomainAccept appends process-specific accept tokens after the baseline.
	// Duplicates (against baseline or within the slice) are dropped while
	// preserving first-seen order.
	DomainAccept []string
	// MaxRounds caps the iterative loop. Zero ⇒ defaults to 30.
	MaxRounds int
	// ArtifactInputs are upstream/file/registry slots rendered before the
	// recurring round-response slot. They never satisfy iterationInput()
	// (Required=true OR Source != InputFromUser) so the helper's last-slot
	// ordering keeps round-response discoverable.
	ArtifactInputs []InputSpec
	// OutputSpecs declares the artifacts produced by the process.
	OutputSpecs []OutputSpec
}

// GuidedUpgradeSpec describes a single-pass guided process: a fixed sequence
// of staged inputs followed by an optional final approval step. There is NO
// recurring slot — iterationInput() must return (_, false) post-upgrade.
type GuidedUpgradeSpec struct {
	// StagedInputs are rendered in declared order — the executor advances one
	// stage at a time.
	StagedInputs []InputSpec
	// OutputSpecs declares the artifacts produced by the process.
	OutputSpecs []OutputSpec
	// FinalApproval optionally appends a ShapeApproval terminator. Nil ⇒ omit.
	FinalApproval *InputSpec
	// MaxRounds caps the guided loop. Zero ⇒ defaults to 10.
	MaxRounds int
	// AcceptTokens overrides the gate accept tokens. Zero-length ⇒ {"yes"}.
	AcceptTokens []string
}

// PartyUpgradeSpec describes a free-form multi-turn process: a one-shot topic
// prompt followed by a recurring per-message slot.
type PartyUpgradeSpec struct {
	// Topic is the one-shot opener (typically Required=true so it does not
	// match the iterationInput predicate).
	Topic InputSpec
	// RoundPrompt is rendered on the recurring "message" slot.
	RoundPrompt string
	// HelpText is the secondary copy beneath the recurring message prompt.
	HelpText string
	// OutputSpecs declares the artifacts produced by the process.
	OutputSpecs []OutputSpec
	// MaxRounds caps the party loop. Zero ⇒ defaults to 100.
	MaxRounds int
	// AcceptTokens overrides the gate accept tokens. Zero-length ⇒
	// {"exit", "done", "wrap up"}.
	AcceptTokens []string
}

// applyIterativeUpgrade upgrades def into the iterative interactive shape.
// No-op when def.Mode != "".
func applyIterativeUpgrade(def *ProcessDef, spec IterativeUpgradeSpec) {
	if def.Mode != "" {
		return
	}

	def.Mode = InteractIterative
	def.EnableAstAdapter = true

	// ArtifactInputs first, then the recurring round-response slot last so
	// iterationInput()'s linear scan resolves to it (schema §5.2).
	inputs := make([]InputSpec, 0, len(spec.ArtifactInputs)+1)
	inputs = append(inputs, spec.ArtifactInputs...)
	inputs = append(inputs, InputSpec{
		ID:        RoundResponseInputID,
		Source:    InputFromUser,
		Shape:     ShapeJSON,
		Required:  false,
		Default:   "",
		MaxLength: roundResponseMaxLength,
		Prompt:    spec.Prompt,
		HelpText:  spec.HelpText,
	})
	def.InputSpecs = inputs
	def.OutputSpecs = spec.OutputSpecs

	def.Gate = buildGate(
		defaultMaxRounds(spec.MaxRounds, defaultIterativeMaxRounds),
		mergeAcceptTokens(commonAcceptTokens, spec.DomainAccept),
	)
}

// applyGuidedUpgrade upgrades def into the guided interactive shape: a
// declarative sequence of StagedInputs (+ optional FinalApproval) with a
// single-pass gate. No recurring slot is added — iterationInput() returns
// (_, false). No-op when def.Mode != "".
func applyGuidedUpgrade(def *ProcessDef, spec GuidedUpgradeSpec) {
	if def.Mode != "" {
		return
	}

	def.Mode = InteractGuided
	def.EnableAstAdapter = true

	total := len(spec.StagedInputs)
	if spec.FinalApproval != nil {
		total++
	}
	inputs := make([]InputSpec, 0, total)
	inputs = append(inputs, spec.StagedInputs...)
	if spec.FinalApproval != nil {
		inputs = append(inputs, *spec.FinalApproval)
	}
	def.InputSpecs = inputs
	def.OutputSpecs = spec.OutputSpecs

	accept := spec.AcceptTokens
	if len(accept) == 0 {
		accept = defaultGuidedAccept
	}
	def.Gate = buildGate(
		defaultMaxRounds(spec.MaxRounds, defaultGuidedMaxRounds),
		accept,
	)
}

// applyPartyUpgrade upgrades def into the party interactive shape: a one-shot
// Topic input followed by a recurring message slot that satisfies the
// iterationInput predicate. No-op when def.Mode != "".
func applyPartyUpgrade(def *ProcessDef, spec PartyUpgradeSpec) {
	if def.Mode != "" {
		return
	}

	def.Mode = InteractParty
	def.EnableAstAdapter = true

	def.InputSpecs = []InputSpec{
		spec.Topic,
		{
			ID:        PartyMessageInputID,
			Source:    InputFromUser,
			Shape:     ShapeJSON,
			Required:  false,
			Default:   "",
			MaxLength: roundResponseMaxLength,
			Prompt:    spec.RoundPrompt,
			HelpText:  spec.HelpText,
		},
	}
	def.OutputSpecs = spec.OutputSpecs

	accept := spec.AcceptTokens
	if len(accept) == 0 {
		accept = defaultPartyAccept
	}
	def.Gate = buildGate(
		defaultMaxRounds(spec.MaxRounds, defaultPartyMaxRounds),
		accept,
	)
}

// mergeAcceptTokens concatenates baseline and domain tokens, dropping
// duplicates while preserving first-seen order. The returned slice is a
// fresh allocation so callers cannot mutate the package-level baseline by
// appending into it.
func mergeAcceptTokens(baseline, domain []string) []string {
	seen := make(map[string]struct{}, len(baseline)+len(domain))
	out := make([]string, 0, len(baseline)+len(domain))
	for _, src := range [][]string{baseline, domain} {
		for _, t := range src {
			if _, dup := seen[t]; dup {
				continue
			}
			seen[t] = struct{}{}
			out = append(out, t)
		}
	}
	return out
}

// buildGate constructs an IterationGate. Accept tokens are aliased — caller
// must pass either a freshly allocated slice (mergeAcceptTokens result) or a
// package-level default that nothing mutates. Reject tokens are copied so
// the package-level baseline can never be mutated by a downstream append.
func buildGate(maxRounds int, accept []string) *IterationGate {
	return &IterationGate{
		Kind:         GateUserConfirm,
		MaxRounds:    maxRounds,
		AcceptTokens: accept,
		RejectTokens: append([]string(nil), commonRejectTokens...),
	}
}

func defaultMaxRounds(specVal, fallback int) int {
	if specVal == 0 {
		return fallback
	}
	return specVal
}

// processIndex returns the slot of id in the package-level registry slice.
// Panics when id is unknown — fail loud at startup, not at runtime. The
// panic message names the missing id so a typo in a rollout helper is caught
// the moment the package is imported.
func processIndex(id string) int {
	for i := range registry {
		if registry[i].ID == id {
			return i
		}
	}
	panic("bmad: unknown process id in interactive rollout: " + id)
}

// memoryOutput constructs an OutputSpec that does NOT persist to disk.
// ArtifactName is empty per ResolveArtifactPath's "unmapped" semantics.
func memoryOutput(id string) OutputSpec {
	return OutputSpec{ID: id, Target: OutputToMemory}
}

// fileOutput constructs an OutputSpec that persists under the given artifact
// name. ID equals ArtifactName since callers always reference the file by its
// canonical artifact name.
func fileOutput(name string) OutputSpec {
	return OutputSpec{ID: name, Target: OutputToFile, ArtifactName: name}
}
