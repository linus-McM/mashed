# Story breadcrumbs-12: Ollama observability logs + feed (DEFERRED)

**Priority:** P3-low
**Domain:** fullstack
**Estimated Complexity:** S
**Depends On:** breadcrumbs-11
**Status:** deferred

> DEFERRED — Phase 3 per plan lines 202, 212-217. Requires explicit human decision to move to `ready`.

## Description

Phase 3 Task 3.3 (plan lines 175-177). Every Ollama call + decision is logged to `~/.mashed/logs/ollama-router.jsonl` as a structured JSONL line. Latency + token count surface in the existing `NotificationFeed` UI so users can see what the router is deciding in real time. Per plan line 184 acceptance: "All LLM decisions logged with the prompt, the response, and the action taken."

## Developer Notes

### Architecture

- **Log sink**: append-only JSONL at `~/.mashed/logs/ollama-router.jsonl`. One line per decision with `{ts, eventType, nodeId, prompt, response, latencyMs, tokenCount, action, target}`.
- **File rotation**: keep it simple — cap at 10 MB, rename to `.1` on rollover. Follow existing rotate logic if mashed has one (grep for `log-rotate`); otherwise implement inline.
- **UI feed**: `NotificationFeed.svelte` already exists (grep frontend/src). Add a filter chip "Ollama router" so users can focus on router events.

### Technical Considerations

- Log writes must not block the event loop — async via buffered channel + dedicated goroutine.
- Serialize the prompt and response with a size cap (truncate at 8 KB, append `[truncated]`).
- On log write failure, emit a one-time warning toast; never crash.

### Wails binding requirements

- `GetOllamaRouterLog(limit int) ([]OllamaRouterEntry, error)` for the NotificationFeed pull path.

### Risks & Edge Cases

- `~/.mashed/logs/` doesn't exist on first run: create with 0700.
- Concurrent router decisions (parallel nodes): single writer goroutine with a `chan OllamaRouterEntry`.

### Reference Files

- Plan: `docs/plans/node-path-breadcrumbs.md` lines 175-177, 182-184.
- Corpus: NotificationFeed component (locate via anatomy / files.md grep `NotificationFeed`).

## Acceptance Criteria

AC-1: Every decision logs a JSONL line
- Given Ollama router decides on an event
- When the decision applies
- Then a line is appended to `~/.mashed/logs/ollama-router.jsonl` with prompt, response, and action

AC-2: Latency + token count visible in feed
- Given the NotificationFeed is open
- When a router decision fires
- Then a feed entry renders with `latencyMs` and `tokenCount` columns

AC-3: Rotation at 10 MB
- Given the JSONL file exceeds 10 MB
- When a new line is written
- Then the current file is renamed to `.1` and a fresh file is started

AC-4: Truncation caps entries
- Given a prompt or response larger than 8 KB
- When logging
- Then the stored field is capped at 8 KB with `[truncated]` appended

## BDD Test Scenarios

```gherkin
Feature: Ollama observability

  Scenario: Decision logged
    Given Ollama router decides on an event
    When the decision applies
    Then ollama-router.jsonl contains a line with keys [ts, action, prompt, response]

  Scenario: Feed entry
    Given the NotificationFeed is open
    When a router decision fires with latency 120ms
    Then a feed entry shows "Ollama router • 120ms"

  Scenario: Rotation
    Given ollama-router.jsonl is 10.1 MB
    When a new line is written
    Then ollama-router.jsonl.1 exists; ollama-router.jsonl is fresh

  Scenario: Truncation
    Given a prompt of 12 KB
    When logging
    Then the logged "prompt" field is 8 KB + "[truncated]"
```

## Tasks / Subtasks

- [ ] Task 1: JSONL writer with rotation (AC-1, AC-3)
  - [ ] `internal/ollama/log.go` with single writer goroutine + `chan OllamaRouterEntry`
- [ ] Task 2: Truncation (AC-4)
  - [ ] Helper capped at 8 KB
- [ ] Task 3: Router wires log emit on every decision (AC-1)
- [ ] Task 4: NotificationFeed filter chip + entry renderer (AC-2)
- [ ] Task 5: Wails binding GetOllamaRouterLog (AC-2)

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on new/modified files
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
