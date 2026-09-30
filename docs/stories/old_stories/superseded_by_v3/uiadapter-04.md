# uiadapter-04: Model-selection guard in `adapter.NewDefault`

**Status:** ready
**Domain:** backend
**Size:** S
**Depends On:** none
**Priority:** P2-medium

## Story

As a UIAdapter maintainer, I want `adapter.NewDefault` to warn when `Config.Model` is outside the vetted allowlist (unless the user explicitly opts in via `ClientConfig.AllowUnvettedModels=true`), so that teams get an early signal when they point the adapter at a community upload whose behaviour does not match the golden fixtures.

## Description

`Config.Model` is currently a free-form string. Users who set it to e.g. `gemma4:latest` receive divergent output vs. `gemma3:4b`, and all golden fixtures were captured against `gemma3:4b`. Add a soft guard (Warn-level slog, not a fatal error) with an override flag. Go-idiomatic: config-layer decisions warn, they don't panic (plan §3 "Never panic in library code"). Parallelisable with stories 1, 2, 3.

### Scope summary
- Add a package-level allowlist: `gemma3:4b`, `gemma3:4b-it-qat`, `gemma3:12b`, `gemma3:1b`.
- New field `ClientConfig.AllowUnvettedModels bool` (default `false`).
- In `NewDefault`, after reading `Config.Model`, if model ∉ allowlist and `AllowUnvettedModels == false`, emit a single `slog.Warn` line with attributes `model` and `allowlist`.
- Document the Warn behaviour in `Config`'s godoc.

### Non-goals
- Do not fail construction on unvetted models. Warn only.
- Do not add the allowlist to a runtime API — it's internal.
- Do not change any other `adapter.go` behaviour.

## Developer Notes

### Files to modify
- `internal/uiadapter/adapter.go` — add allowlist, add field to `ClientConfig`, add the Warn in `NewDefault`, update the `Config` godoc.
- `internal/uiadapter/adapter_test.go` — new tests `TestAdapter_WarnsOnUnvettedModel`, `TestAdapter_NoWarnOnAllowedModel`, `TestAdapter_NoWarnWhenAllowUnvettedModels`.

### Type/symbol inventory (exact names)
- `ClientConfig.AllowUnvettedModels bool` — default `false`; documented as experimentation-only override.
- Package-private allowlist: `var vettedModels = []string{"gemma3:4b", "gemma3:4b-it-qat", "gemma3:12b", "gemma3:1b"}` (or equivalent `map[string]struct{}` for O(1) lookup).
- Log line: `slog.Warn("uiadapter.unvetted_model", slog.String("model", cfg.Model), slog.Any("allowlist", vettedModels))`.

### Test harness for slog capture
Existing Mashed Go patterns use a test `slog.Handler` that appends records to a slice (grep existing tests under `internal/uiadapter/` or `internal/bmad/` for a reference). Use the same pattern so the tests assert exactly one Warn line with the expected attributes.

### Where to warn
- Plan §1 Story 4: "On construction, if the model isn't on the list and the flag is false, log a Warn-level line (don't fail — Go idiom prefers warnings for config-layer decisions) and continue."
- Put the check at the end of `NewDefault` (after the client is constructed) so it warns once per adapter instance, not per request.

### Risks / gotchas
- Someone legitimately needs a community model. The `AllowUnvettedModels=true` flag suppresses the warning.
- Do not log the full allowlist as a multi-line string — keep it structured (`slog.Any`) so it's queryable.
- Avoid double-warning if `NewDefault` is called multiple times in one test binary — each instance is independent, a single Warn per call is correct.
- Do not warn when `Config.Model` is empty (let the existing default-setting logic pick `gemma3:4b`, which is vetted). If the default is set AFTER the guard, move the guard below the default logic.

### Telemetry
- The Warn line's name `uiadapter.unvetted_model` is distinct from `uiadapter.translate`. This is acceptable — plan §3 only mandates that the `translate` line stays single per call. Config-layer events are free to use their own name.

### Reference files
- `internal/uiadapter/adapter.go` — `NewDefault`, `Config`, existing slog patterns.
- `internal/uiadapter/client.go` — `ClientConfig` struct.
- `internal/uiadapter/adapter_test.go` — existing slog capture idiom.
- `docs/plans/IMPLEMENTATION_PLAN.md` §1 Story 4.

## Acceptance Criteria

**AC-4.1: `TestAdapter_WarnsOnUnvettedModel` — warns when model is off-list**
- Given `Config.Model == "gemma4:latest"` and `ClientConfig.AllowUnvettedModels == false`
- When `adapter.NewDefault(cfg)` runs under a slog test handler
- Then exactly one Warn record is captured
- And that record has name `uiadapter.unvetted_model`
- And it has `model == "gemma4:latest"`

**AC-4.2: No warning for an allowed model**
- Given `Config.Model == "gemma3:4b"` and `ClientConfig.AllowUnvettedModels == false`
- When `adapter.NewDefault(cfg)` runs
- Then zero `uiadapter.unvetted_model` Warn records are captured

**AC-4.3: No warning when override flag is set**
- Given `Config.Model == "gemma4:latest"` and `ClientConfig.AllowUnvettedModels == true`
- When `adapter.NewDefault(cfg)` runs
- Then zero `uiadapter.unvetted_model` Warn records are captured

**AC-4.4: Godoc documents the Warn behaviour**
- Given the updated `Config` or `ClientConfig` godoc
- When a reader inspects the godoc
- Then the doc explicitly says unvetted models emit a Warn and lists the allowlist
- And it names the override flag `AllowUnvettedModels`

## BDD Test Scenarios

### Scenario 1: Unvetted model triggers Warn

```gherkin
Feature: Model allowlist guard

  Scenario: Unvetted model emits a single Warn
    Given ClientConfig.AllowUnvettedModels is false
    And Config.Model is "gemma4:latest"
    And a slog test handler is active
    When adapter.NewDefault runs
    Then one record with level Warn and name "uiadapter.unvetted_model" is captured
    And the record has attribute model equal to "gemma4:latest"

  Scenario: Allowed model is silent
    Given ClientConfig.AllowUnvettedModels is false
    And Config.Model is "gemma3:4b"
    When adapter.NewDefault runs
    Then no "uiadapter.unvetted_model" records are emitted

  Scenario: Override flag suppresses the warning
    Given ClientConfig.AllowUnvettedModels is true
    And Config.Model is "gemma4:latest"
    When adapter.NewDefault runs
    Then no "uiadapter.unvetted_model" records are emitted
```

### Scenario 2: Allowlist membership

```gherkin
Feature: Allowlist covers the pinned models

  Scenario Outline: Each allowed model is silent
    Given Config.Model is "<model>"
    And ClientConfig.AllowUnvettedModels is false
    When adapter.NewDefault runs
    Then no "uiadapter.unvetted_model" records are emitted

    Examples:
      | model               |
      | gemma3:4b           |
      | gemma3:4b-it-qat    |
      | gemma3:12b          |
      | gemma3:1b           |
```

## Tasks / Subtasks

- [ ] Task 1: Add allowlist and config flag (AC: 4.1, 4.2, 4.3)
  - [ ] Declare `vettedModels` in `adapter.go`
  - [ ] Add `AllowUnvettedModels bool` to `ClientConfig` with godoc
  - [ ] Decide whether to store as slice or `map[string]struct{}` — map preferred for O(1)
- [ ] Task 2: Implement the guard in `NewDefault` (AC: 4.1, 4.2, 4.3)
  - [ ] After defaulting `Config.Model`, check membership
  - [ ] Emit `slog.Warn("uiadapter.unvetted_model", ...)` when off-list and flag false
  - [ ] Continue construction unconditionally (never fail)
- [ ] Task 3: Update godoc (AC: 4.4)
  - [ ] Document the Warn behaviour on `Config` (or `ClientConfig.AllowUnvettedModels`)
  - [ ] List the allowlist values
- [ ] Task 4: Add tests (AC: 4.1, 4.2, 4.3)
  - [ ] Use the project's standard slog test handler
  - [ ] Table-drive the allowlist membership scenario
  - [ ] Assert single-Warn invariant (count == 1 when triggered)
- [ ] Task 5: Pre-flight and handoff (AC: all)
  - [ ] `go build ./...`, `go vet ./...`
  - [ ] `go test ./internal/uiadapter/... -race -short`
  - [ ] `/simplify` on `adapter.go`

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ coverage on the modified lines of `adapter.go`
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./internal/uiadapter/... -race -short` passes
- [ ] `/simplify` run on all modified code
- [ ] Pre-flight: Warn count == 1 on trigger, zero on allowed/override
- [ ] AC Validation Table complete

### AC Validation Table (fill in PR description)

| AC | Test | Status |
|----|------|--------|
| AC-4.1 | `TestAdapter_WarnsOnUnvettedModel` | |
| AC-4.2 | `TestAdapter_NoWarnOnAllowedModel` | |
| AC-4.3 | `TestAdapter_NoWarnWhenAllowUnvettedModels` | |
| AC-4.4 | Manual godoc review | |
