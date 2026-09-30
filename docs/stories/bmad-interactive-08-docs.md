# bmad-interactive-08: Docs — mark spec implemented and update CLAUDE.md

**Status:** done
**Domain:** docs-only
**Size:** S
**Depends On:** bmad-interactive-01, bmad-interactive-02, bmad-interactive-03, bmad-interactive-04, bmad-interactive-05, bmad-interactive-06, bmad-interactive-07
**Priority:** P2-medium

## Story

As a future Mashed contributor, I want `CLAUDE.md` and `docs/bmad-interactive-process-schema.md` to reflect the shipped state of the interactive BMAD schema, so that I can orient on the new types, events, and APIs without reverse-engineering the code.

## Description

Flip the spec's `Status: Design spec — not yet implemented.` to `Status: Implemented (v1).`, add a "Shipped in stories S1-S7" note to each major section, and augment `CLAUDE.md`'s BMAD section with pointers to the new types, events, and the `RespondToInput` binding.

This story is docs-only — no Go or TS code touched. It lands after S1-S7 so the docs reflect what actually shipped, not what was planned.

### Scope summary
- Update `docs/bmad-interactive-process-schema.md` front matter + §11.2 rollout table to cross-link completed stories.
- Add a new `## BMAD Interactive Processes` section to `CLAUDE.md` covering:
  - Types: `InputSpec`, `OutputSpec`, `IterationGate`, `PendingPrompt`, `NodeInputEntry`
  - Status: `NodeAwaitingInput`
  - Event names: `bmad:node:awaiting_input`, `bmad:node:input_resolved`, `bmad:node:input_invalid`, `bmad:node:round_complete`, `bmad:node:gate_satisfied`, `bmad:node:round_limit`, `bmad:node:aborted`
  - Wails bindings: `RespondToInput`, legacy `RespondToQuestion`
  - Security invariants: path-traversal rejection on `ShapeFile`, valueHash-only events, `registry:` scheme whitelist
- Add a short reference block to `.wolf/cerebrum.md` under "Key Learnings" capturing the interactive-node invariants (the downstream-hold proof sketch from §5.4 and the "one input, one suspension" rule).
- Update `README.md` briefly if it advertises the BMAD capability.

### Non-goals
- No new spec content or design decisions — docs only reflect shipped state.
- No migration guides for external users (external API is additive).

## Developer Notes

### Files to modify
- `docs/bmad-interactive-process-schema.md` — flip status, annotate sections with story IDs.
- `CLAUDE.md` — add `## BMAD Interactive Processes` section after the existing BMAD Artifact System block.
- `.wolf/cerebrum.md` — append under `## Key Learnings` a 3-4 line block on interactive invariants.
- `README.md` — optional: add one line under features if a features list exists.

### Content for `CLAUDE.md` (add near existing BMAD section)

```markdown
## BMAD Interactive Processes

Shipped in stories `bmad-interactive-01..07`. See `docs/bmad-interactive-process-schema.md` for the full schema.

- **Types** — `internal/bmad/types.go`: `InputSpec`, `OutputSpec`, `IterationGate`, `PendingPrompt`, `NodeInputEntry`; `ProcessDef.{Mode, InputSpecs, OutputSpecs, Gate}`; `WorkflowExecution.{NodeRounds, PendingPrompts, NodeInputs, NodeInputHistory}`
- **Status** — `NodeAwaitingInput` is the single source of truth for "suspended". Downstream in-degree never decrements while a node is awaiting (see §5.4 proof).
- **Events** — `bmad:node:awaiting_input`, `:input_resolved`, `:input_invalid`, `:round_complete`, `:gate_satisfied`, `:round_limit`, `:aborted`. Payloads in `internal/bmad/events.go`. Legacy `:question` / `:idle` remain as autonomous-node fallback.
- **Wails binding** — `(*App).RespondToInput(execID, nodeID, inputID, value)`. Legacy `RespondToQuestion` stays for autonomous nodes.
- **Security** — `ShapeFile` values are rejected if not under `repoPath`; `bmad:node:input_resolved` emits `valueHash` (SHA-256), never raw values; `OptionsRef` resolver accepts only `registry:*` scheme.
```

### Content for the spec status update
- Change the front matter line `> **Status:** Design spec — not yet implemented.` to `> **Status:** Implemented (v1 shipped in stories S1-S7).`
- In §11.2, replace the Verification column phrasing with a link to each story file: `[S1: Types](stories/bmad-interactive-01-types.md)` etc.
- At the top of §3, add `> Implemented in `bmad-interactive-01-types.md`.`
- At the top of §5, add `> Routing + suspension implemented in `bmad-interactive-02`..`03`; iteration gate in `04`.`
- At the top of §7, add `> Persistence + re-emit in `bmad-interactive-05`.`
- At the top of §9, add `> Frontend shipped in `bmad-interactive-06`.`
- At the top of §10, add `> Registry entries populated in `bmad-interactive-07`.`

### Content for `.wolf/cerebrum.md`

Append under `## Key Learnings`:
```markdown
- BMAD interactive nodes: `NodeAwaitingInput` suspends via `suspendForSpec`; `activeOutEdges` only runs for `NodeComplete`, so downstream in-degree cannot decrement while upstream is awaiting. Keeps the ready-set invariant observable and correct.
- BMAD input responses: `RespondToInput(execID, nodeID, inputID, value)` validates + stores + releases a per-waiter channel. Legacy `RespondToQuestion` is retained only for autonomous nodes where claude surprises us.
- Event hygiene: `bmad:node:input_resolved` payload carries `valueHash` (SHA-256 prefix), NEVER raw value. Raw values live only in `~/.mashed/...execution.json`.
```

### Risks / gotchas
- Do NOT rewrite the spec's technical content — only add status annotations and cross-links.
- Keep the CLAUDE.md block short (< 30 lines). It is context for future agents, not a tutorial.
- Do not add a CHANGELOG unless one exists. Project convention is git log + docs/stories as the audit trail.

### Reference files
- `docs/bmad-interactive-process-schema.md` — target.
- `CLAUDE.md` — target.
- `.wolf/cerebrum.md` — target (append-only via OpenWolf protocol).
- Prior stories `bmad-interactive-01` through `bmad-interactive-07` for cross-links.

## Acceptance Criteria

**AC-1: Spec status flipped to implemented**
- Given `docs/bmad-interactive-process-schema.md` after this story
- When the file is read
- Then the front-matter `Status:` line reads `Implemented (v1 shipped in stories S1-S7).` or equivalent
- And the original `Design spec — not yet implemented.` phrasing is gone

**AC-2: Spec §11.2 rollout table cross-links to shipped story files**
- Given §11.2's rollout table
- Then each row's "Sprint story" column includes a markdown link to `stories/bmad-interactive-0N-*.md`
- And every link resolves to an existing file in `docs/stories/`

**AC-3: CLAUDE.md has a new `## BMAD Interactive Processes` section**
- Given `CLAUDE.md` after this story
- Then there is a section heading exactly `## BMAD Interactive Processes`
- And the section lists the new types (`InputSpec`, `OutputSpec`, `IterationGate`, `PendingPrompt`, `NodeInputEntry`)
- And the section lists the seven event names
- And the section mentions the `RespondToInput` binding
- And the section names the three security invariants

**AC-4: `.wolf/cerebrum.md` appended with interactive-node key learnings**
- Given `.wolf/cerebrum.md` after this story
- Then the `## Key Learnings` section contains a bullet referencing `NodeAwaitingInput` and the downstream-hold invariant
- And a bullet referencing `RespondToInput` semantics
- And a bullet referencing the valueHash event hygiene rule

**AC-5: Docs build / link-check passes (if tooling present)**
- Given any markdown linter or link checker wired into CI
- When run against the repo after this story
- Then no broken link warnings are introduced by the new cross-links

## BDD Test Scenarios

```gherkin
Feature: Docs reflect shipped interactive schema

  Scenario: Spec marked implemented
    When the design spec is opened
    Then its Status line reads "Implemented (v1 shipped in stories S1-S7)"
    And no "not yet implemented" phrase remains

  Scenario: Spec cross-links to shipped stories
    When §11.2 is read
    Then every rollout-table row links to an existing docs/stories/bmad-interactive-0N-*.md file

  Scenario: CLAUDE.md has interactive-processes section
    When CLAUDE.md is opened
    Then it contains a heading "## BMAD Interactive Processes"
    And the section names InputSpec, OutputSpec, IterationGate, PendingPrompt, NodeInputEntry
    And the section lists all seven bmad:node:* event names
    And the section mentions RespondToInput binding
    And the section lists three security invariants (path traversal, valueHash, registry: scheme)

  Scenario: Cerebrum has invariant notes
    When .wolf/cerebrum.md is opened
    Then its Key Learnings section contains a bullet about NodeAwaitingInput
    And a bullet about RespondToInput semantics
    And a bullet about valueHash event hygiene
```

## Tasks / Subtasks

- [ ] Task 1: Flip spec status + annotate sections (AC-1, AC-2)
  - [ ] Update front matter Status line
  - [ ] Add "Implemented in ..." callouts above §3, §5, §7, §9, §10
  - [ ] Convert §11.2 rollout-table story references into markdown links
- [ ] Task 2: Add CLAUDE.md section (AC-3)
  - [ ] Insert `## BMAD Interactive Processes` after existing BMAD Artifact System block
  - [ ] Keep under ~30 lines, reference `internal/bmad/*.go` by filename
- [ ] Task 3: Append cerebrum.md key learnings (AC-4)
  - [ ] Three bullets per spec above
- [ ] Task 4: Optional README nudge
  - [ ] Add one-line feature note if README has a features/highlights list; otherwise skip
- [ ] Task 5: Link-check pass (AC-5)
  - [ ] Run whatever link checker the repo has (if any); manually click every new link otherwise

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] BDD scenarios verified via manual read-through (docs-only story; no automated test harness required)
- [ ] `go build ./...` passes (sanity — no code change)
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` not applicable (docs only) — skip
- [ ] Story status flipped to `done`
- [ ] Changes committed on branch
