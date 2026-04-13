# Story breadcrumbs-11: Ollama router integration (DEFERRED)

**Priority:** P3-low
**Domain:** backend
**Estimated Complexity:** M
**Depends On:** breadcrumbs-10
**Status:** deferred

> DEFERRED — Phase 3 per plan lines 202, 212-217. Requires explicit human decision to move to `ready`.

## Description

Phase 3 Task 3.2 (plan lines 169-173). Subscribe to `bmad:node:artifacts` + `bmad:node:status` on the Go side. When downstream fan-out is ambiguous (multiple targets accept the same artifact) OR when output content matches a semantic trigger regex, call Ollama with a tight prompt. Parse the JSON decision into one of `{populate-path, toast, noop}` and apply it.

## Developer Notes

### Architecture

A new subscriber in Go that wraps the emit bus used by the executor. On each event, run:

1. **Ambiguity check** — count downstream edges from the source node whose target accepts the same artifact name AND whose `inputPaths[name]` is empty. If count > 1, ask Ollama for a `target` decision.
2. **Semantic trigger** — run output snippet (the content of the artifact file, first 4 KB) against a small regex set (e.g. "blocker", "needs clarification", "regression"). On hit, ask Ollama whether to fire a toast.
3. Parse the response:

```json
{
  "action": "populate-path" | "toast" | "noop",
  "target": "<nodeId>",
  "toastLevel": "info"|"warn"|"error",
  "toastText": "..."
}
```

4. Apply:
   - `populate-path`: emit a targeted `bmad:node:artifacts-assign` event with `{nodeId, artifactName, path}` that the frontend (breadcrumbs-06) uses to bypass the empty-only rule for this one write.
   - `toast`: emit `app:toast` with level + text.
   - `noop`: log only.

### Technical Considerations

- **Opt-in**: `IsOllamaAvailable()` false → skip all routing logic entirely. Deterministic behavior unchanged (plan line 182).
- **Determinism-within-session**: cache `(prompt hash) → decision` in-memory. Same input → same decision until model or prompt changes (plan line 183 acceptance).
- **Hard timeout**: 2s total per call (inherits from client). If exceeded, fall through to deterministic.

### Risks & Edge Cases

- Ollama hallucinates a `target` nodeId that doesn't exist: validate against current graph; on mismatch → noop + log.
- Routes the same path to multiple targets: caller re-runs per target; normally decision returns one target.
- User disables Ollama mid-run: next event skips the router.

### Reference Files

- Plan: `docs/plans/node-path-breadcrumbs.md` lines 169-173, 182-184.
- Skills: `/golang-testing`, `/golang-error-handling`.

## Acceptance Criteria

AC-1: Ambiguity triggers Ollama
- Given node A emits `PRD.md` with two downstream consumers both empty
- When Ollama is enabled
- Then the router queries Ollama with the ambiguity prompt
- And the decision applies to exactly one target

AC-2: Semantic trigger fires toast
- Given A's output contains the word "blocker"
- When the router processes the event
- Then it queries Ollama for a toast decision
- And on action=toast the frontend receives an `app:toast` event

AC-3: Disabled → deterministic passthrough
- Given Ollama disabled
- When a fan-out event fires
- Then no HTTP request is made
- And Phase 2 auto-fill (breadcrumbs-06) runs unchanged

AC-4: Hallucinated target validated
- Given Ollama returns `target: "node-does-not-exist"`
- When the router applies the decision
- Then no edge is modified
- And the error is logged

AC-5: Cache keyed by prompt hash
- Given the same event fires twice within a session
- When the router processes the second event
- Then no new Ollama call is made
- And the cached decision is applied

## BDD Test Scenarios

```gherkin
Feature: Ollama router

  Scenario: Fan-out ambiguity resolved
    Given a fan-out event with targets B and C both empty
    And Ollama returns {action:"populate-path", target:"B"}
    When the router applies
    Then B.inputPaths["PRD.md"] fills; C remains empty

  Scenario: Semantic toast
    Given output content contains "blocker"
    And Ollama returns {action:"toast", toastLevel:"warn", toastText:"BMAD flagged"}
    When the router applies
    Then an app:toast event fires with the warn level

  Scenario: Disabled passthrough
    Given Ollama disabled
    When any event fires
    Then no HTTP request is made
    And deterministic auto-fill still runs

  Scenario: Cache hit on repeat
    Given identical prompt fires twice
    When the router processes event 2
    Then no HTTP request is made
    And the cached decision applies
```

## Tasks / Subtasks

- [ ] Task 1: Event subscriber wiring (AC-1, AC-3)
  - [ ] New `internal/ollama/router.go` subscribing to the executor emit bus
  - [ ] Bail when `IsOllamaAvailable()` false
- [ ] Task 2: Ambiguity detection (AC-1)
  - [ ] Compute candidate targets per artifact; query only when count > 1
- [ ] Task 3: Semantic trigger regex set (AC-2)
  - [ ] Config-driven list in `mashedConfig.OllamaTriggers`
- [ ] Task 4: Apply decision (AC-1, AC-2, AC-4)
  - [ ] Validate targets against current graph
  - [ ] Emit events for populate-path / toast
- [ ] Task 5: Decision cache (AC-5)
  - [ ] In-memory map keyed by sha256(prompt); session-scoped
- [ ] Task 6: Integration tests (all ACs)

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on new/modified files
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
