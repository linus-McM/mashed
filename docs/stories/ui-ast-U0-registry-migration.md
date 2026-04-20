# ui-ast-U0: Registry migration to ShapeJSON + per-process EnableAstAdapter opt-in

**Status:** ready
**Domain:** backend
**Size:** M
**Depends on:** none
**Priority:** P0-critical

## Story

As the BMAD registry owner, I want every interactive process to flip its iterative `InputSpec.Shape` to `ShapeJSON` behind a per-process opt-in flag (`ProcessDef.EnableAstAdapter`), so that the downstream adapter wiring in U4 can land the flatten-on-receipt map path without disturbing autonomous or not-yet-migrated processes.

## Description

Implements §9 Phase 0 of `docs/mashed-ui-ast-schema.md` (the "U0" row). Adds a boolean field `ProcessDef.EnableAstAdapter` (default `false`, `omitempty`) and migrates the iterative `InputSpec.Shape` for every interactive process in `internal/bmad/registry.go` whose process opts in. Non-opted-in processes keep their existing `ShapeFree`/`ShapeChoice`/etc. and are guaranteed byte-equivalent JSON output. This story ships no adapter wiring — §5.2 and §5.3.1 of the spec remain unimplemented until U4. The sole observable change today is that opted-in processes' iteration specs now accept a JSON map at `RespondToInput` time; the modal still submits a plain string until Phase C lands, so the `validateInput` path for `ShapeJSON` (64 KiB cap in `internal/bmad/validate.go:13`) must accept the current plain-string submissions as a no-op regression.

### Scope summary

- Add `ProcessDef.EnableAstAdapter bool `json:"enableAstAdapter,omitempty"`` in `internal/bmad/types.go`.
- Flip the iteration `InputSpec.Shape` to `ShapeJSON` on the Party Mode / Brainstorming / Product-Brief / Advanced-Elicitation entries in `internal/bmad/registry.go` **and** set `EnableAstAdapter: true` on those four processes.
- Regression snapshot of `registry.go` JSON output for every non-migrated process.
- Document the field in `docs/bmad-interactive-process-schema.md` under §11 (additive changes).

### Non-goals

- No `internal/uiadapter` package (U2).
- No `PendingPrompt.Structured` field (U4 — §5.1).
- No frontend changes (U6/U7).
- No `suspendForSpec` changes (U4 — §5.2).

## Developer Notes

### Files to modify

- `internal/bmad/types.go` — extend `ProcessDef` with `EnableAstAdapter bool` (omitempty), placed adjacent to `Mode InteractionMode` per the §3.4 (interactive schema) layout convention.
- `internal/bmad/registry.go` — flip iteration `InputSpec.Shape` to `ShapeJSON` on the four interactive processes; set `EnableAstAdapter: true` on each.
- `internal/bmad/registry_test.go` — golden-JSON regression: unchanged processes serialise byte-identical to a committed fixture; migrated processes match the new fixture.
- `internal/bmad/types_test.go` — round-trip test for `ProcessDef.EnableAstAdapter` with both `true` and absent (zero) values.
- `docs/bmad-interactive-process-schema.md` — one-paragraph addendum under §11.

### Processes to migrate (spec §9 + interactive schema §7)

Each of these has an iterative `InputSpec` (`Shape: ShapeFree` today) that drives the round loop — flip to `ShapeJSON` and set `EnableAstAdapter: true`:

1. `bmad-brainstorming`
2. `bmad-product-brief`
3. `bmad-party-mode`
4. `bmad-advanced-elicitation`

Look for the `Shape:` tokens on each process's iteration `InputSpec` (the one whose `Required == true` and is consumed inside the iteration gate — see `internal/bmad/registry.go`). Do NOT migrate non-iterative specs (e.g. `bmad-brainstorming`'s method-choice spec stays `ShapeChoice`).

### Regression contract

Unchanged non-interactive and non-migrated processes (every process whose `Mode == ""` or that isn't in the four above) MUST produce byte-identical JSON after this story. The test:

```go
// internal/bmad/registry_test.go
func TestRegistry_NonMigratedProcessesUnchanged(t *testing.T) {
    for _, id := range nonMigratedIDs {
        got, _ := json.Marshal(ProcessByIDOrFail(t, id))
        want := loadGolden(t, "testdata/registry/"+id+".json")
        if !bytes.Equal(got, want) { ... }
    }
}
```

Commit `testdata/registry/*.json` fixtures captured from `main` before the migration. This is the sole guard that prevents "accidentally changed an unrelated process" regressions.

### `EnableAstAdapter` semantics

- Default `false` on every `ProcessDef` literal that does not explicitly set it.
- Consumed only by U4 (`suspendForSpec` adapter-call gate) and U5 (settings toggle read). Presence in U0 is declarative only — the field lands with no reader today.
- JSON key `enableAstAdapter`; `omitempty` so the non-migrated golden snapshots stay byte-identical.

### Risks / gotchas

- **Silent `ShapeJSON` breakage.** `validateInput` (`internal/bmad/validate.go:13`) accepts any well-formed JSON ≤ 64 KiB for `ShapeJSON`. Current frontend submits a plain string — that is NOT well-formed JSON unless quoted. Action: add a compatibility branch in `validateInput` that accepts a bare-string submission for `ShapeJSON` when the string is not valid JSON (document the rule: "JSON shape accepts either a JSON value or a bare string for backward compat until U4"). Without this, U0 would break existing modal submissions before U7 lands. Verify via `TestValidateInput_ShapeJSON_BareStringAccepted`.
- **`EnableAstAdapter: true` with no reader.** Expected — the field is inert until U4. Add a `//nolint:unused` or a brief code comment noting "wired by U4" so reviewers do not delete.
- **Registry `IsTemplate: true` entries.** Skip — templates do not carry iteration gates.
- **Snapshot round-trip.** `WorkflowDef.Nodes[i].ProcessDef` is not persisted — only `ProcessID` is. No snapshot migration needed.

### Reference files

- `internal/bmad/registry.go` — the four process literals to mutate.
- `internal/bmad/validate.go:13` — `validateInput` JSON-shape handling.
- `internal/bmad/types.go` — `ProcessDef` struct (see `docs/stories/bmad-interactive-01-types.md` for field layout).
- `docs/mashed-ui-ast-schema.md` §3.4 (collapse rule), §9 (rollout).

## Acceptance Criteria

**AC-1: `ProcessDef.EnableAstAdapter` compiles and round-trips**
- Given a `ProcessDef` with `EnableAstAdapter: true`
- When it is marshalled to JSON and unmarshalled back
- Then the decoded struct has `EnableAstAdapter == true`
- And the JSON contains `"enableAstAdapter":true`
- And a `ProcessDef` with `EnableAstAdapter: false` marshals without the `enableAstAdapter` key (omitempty)
- Verified by `TestProcessDef_EnableAstAdapter_RoundTrip`

**AC-2: The four interactive processes migrate to `ShapeJSON` + opt in**
- Given `ProcessByID("bmad-brainstorming")`, `bmad-product-brief`, `bmad-party-mode`, `bmad-advanced-elicitation`
- When the caller reads the process's iteration `InputSpec`
- Then its `Shape == ShapeJSON`
- And the `ProcessDef.EnableAstAdapter == true`
- Verified by `TestRegistry_MigratedProcesses_ShapeJSON`

**AC-3: Non-migrated processes are byte-identical to pre-U0 JSON**
- Given every `ProcessDef` in the registry whose ID is not in the migrated-four set
- When marshalled to JSON
- Then the output is byte-equal to the committed `testdata/registry/<id>.json` golden
- Verified by `TestRegistry_NonMigratedProcessesUnchanged`

**AC-4: `validateInput` accepts bare-string submissions for `ShapeJSON` (back-compat until U7)**
- Given an `InputSpec{Shape: ShapeJSON}`
- When `validateInput(spec, "pick option 2")` is called with a plain string that is not valid JSON
- Then validation returns nil (accepts the string as a legacy submission)
- And given `validateInput(spec, "{\"foo\":\"bar\"}")` with valid JSON
- Then validation also returns nil
- Verified by `TestValidateInput_ShapeJSON_BareStringAccepted`

**AC-5: §11 Q6 resolved — `UIAdapterEnabled` defaults to TRUE (user decision 2026-04-21)**
- Given `EnableAstAdapter: true` in a process and the resolved default `UIAdapterEnabled: true` (overriding spec §4.4's `false` default)
- When a workflow runs that process today (with no U4 wiring in place)
- Then the round loop behaves exactly as pre-U0 — the U0 changes are declarative only until U4
- And when U4 lands, an unreachable Ollama degrades to `fallback:unreachable` (raw markdown) per §4.8; there is no auto-disable
- Verified by `TestRegistry_MigratedProcessesEnableAstAdapter` — asserts `EnableAstAdapter: true` on the four migrated processes. Carries a `// NOTE(U4): Q6 resolved — enabled by default; unreachable Ollama degrades to fallback:unreachable per spec §4.8` comment

## BDD Test Scenarios

```gherkin
Feature: Registry migration to ShapeJSON + EnableAstAdapter opt-in

  Scenario: EnableAstAdapter round-trips through JSON
    Given a ProcessDef with EnableAstAdapter true
    When it is marshalled and unmarshalled
    Then the decoded EnableAstAdapter equals true
    And the JSON contains "enableAstAdapter":true

  Scenario: Default-zero EnableAstAdapter omits the key
    Given a ProcessDef with EnableAstAdapter false
    When marshalled to JSON
    Then the output has no "enableAstAdapter" key

  Scenario: Four interactive processes carry ShapeJSON + opt-in
    Given ProcessByID for bmad-brainstorming
    Then its iteration InputSpec Shape equals ShapeJSON
    And its EnableAstAdapter equals true

  Scenario: Non-migrated processes are byte-identical to the golden
    Given any non-migrated process in the registry
    When marshalled to JSON
    Then the output equals the committed testdata/registry/<id>.json byte-for-byte

  Scenario: ShapeJSON accepts a bare string for backward compatibility
    Given an InputSpec with Shape ShapeJSON
    When validateInput is called with "pick option 2"
    Then validation returns nil

  Scenario: ShapeJSON still accepts valid JSON maps
    Given an InputSpec with Shape ShapeJSON
    When validateInput is called with {"confirm":"done"}
    Then validation returns nil

  Scenario: Q6 resolved — migrated processes opted in, defaults documented
    Given the four migrated processes
    Then each asserts EnableAstAdapter true
    And the test file documents that UIAdapterEnabled defaults to true (user decision 2026-04-21)
    And unreachable Ollama degrades per spec §4.8 (no auto-disable)
```

## Tasks / Subtasks

- [ ] Task 1: Add `ProcessDef.EnableAstAdapter` type field (AC-1)
  - [ ] RED: write `TestProcessDef_EnableAstAdapter_RoundTrip` asserting both `true` and zero-value behaviour
  - [ ] GREEN: add `EnableAstAdapter bool `json:"enableAstAdapter,omitempty"`` to `ProcessDef`
  - [ ] REFACTOR: run `/simplify` on `internal/bmad/types.go`
- [ ] Task 2: Capture pre-migration golden JSON for every registry process (AC-3)
  - [ ] Write `internal/bmad/testdata/registry/<id>.json` for every process currently in `ProcessByID`
  - [ ] Commit goldens before any migration edits so the regression is load-bearing
- [ ] Task 3: Migrate the four interactive processes (AC-2)
  - [ ] RED: write `TestRegistry_MigratedProcesses_ShapeJSON` (fails before edit)
  - [ ] GREEN: flip iteration `Shape` to `ShapeJSON` + set `EnableAstAdapter: true` for each of the four
  - [ ] Refresh the goldens for the four migrated IDs
  - [ ] REFACTOR: run `/simplify` on `internal/bmad/registry.go`
- [ ] Task 4: Lock non-migrated processes to golden JSON (AC-3)
  - [ ] RED: write `TestRegistry_NonMigratedProcessesUnchanged` iterating the non-migrated ID set
  - [ ] GREEN: confirm the test passes; no code change expected
- [ ] Task 5: Extend `validateInput` `ShapeJSON` back-compat (AC-4)
  - [ ] RED: write `TestValidateInput_ShapeJSON_BareStringAccepted`
  - [ ] GREEN: add the "accept bare string when not valid JSON" branch in `internal/bmad/validate.go`
  - [ ] REFACTOR: `/simplify`
- [ ] Task 6: Confirm §11 Q6 resolution in tests + schema (AC-5)
  - [ ] Add `TestRegistry_MigratedProcessesEnableAstAdapter` with the "Q6 resolved — enabled by default" comment
  - [ ] Append a one-paragraph §11 addendum to `docs/bmad-interactive-process-schema.md` recording the 2026-04-21 decision: `UIAdapterEnabled` default TRUE; unreachable Ollama → `fallback:unreachable` per spec §4.8

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ coverage on modified code in `internal/bmad/types.go`, `registry.go`, `validate.go`
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on every modified Go file; no CRITICAL/HIGH findings
- [ ] Story status flipped to `done`
- [ ] Changes committed on branch
