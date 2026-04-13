# Story breadcrumbs-09: Ollama settings UI + lifecycle (DEFERRED)

**Priority:** P3-low
**Domain:** fullstack
**Estimated Complexity:** L
**Depends On:** breadcrumbs-06
**Status:** deferred

> DEFERRED — Phase 3 per plan lines 202, 212-217. Only release after Phase 2 surfaces evidence of ambiguity that deterministic code cannot handle. Requires explicit human decision to move to `ready`.

## Description

Phase 3 Task 3.0 (plan lines 146-161). Adds an Ollama section to `Settings.svelte` implementing the state machine `not-installed → installed-stopped → starting → running → model-selected`. Surfaces Wails bindings for install detection, server start/stop, status polling, model listing, and model selection. Persists `OllamaModel` and `OllamaEnabled` in `mashedConfig`. Manages graceful shutdown: if Mashed started Ollama, Mashed stops it; if Ollama was pre-running, leave it alone.

## Developer Notes

### Architecture

State machine (plan lines 148-153):
- `not-installed` — detect via `which ollama`; show link to https://ollama.com.
- `installed-stopped` — show Start button; click → spawn `ollama serve`; poll `GET http://localhost:11434/api/version`.
- `running` — `GET /api/tags` populates model dropdown (name + param count + size).
- `model-selected` — persist to `mashedConfig.OllamaModel` + `OllamaEnabled`; show Stop button.

New bindings in `app.go` (plan line 154):
- `IsOllamaInstalled() bool`
- `StartOllamaServer() error`
- `StopOllamaServer() error`
- `GetOllamaStatus() (OllamaStatus, error)` — running, pid, port
- `ListOllamaModels() ([]OllamaModel, error)` — parses `/api/tags`
- `SetOllamaModel(name string) error`

`mashedConfig` lives around corpus line 30121 (in `app.go` / main). Add `OllamaModel string` + `OllamaEnabled bool` fields with `json:",omitempty"`.

### Technical Considerations

- **Subprocess tracking**: `app.go` holds a `*exec.Cmd` + `startedByUs bool`. On app-shutdown, if `startedByUs`, send SIGTERM and wait up to 5s.
- **Health polling**: bounded timeout (default 10s) from Start click to first 200. Fail state on timeout.
- **Port hardcoded**: 11434. Document in a constant; no user override this story.
- **Cross-platform**: macOS `which ollama` is fine; design keeps the door open for Windows (`where ollama`) but this story stays macOS.

### Wails binding requirements

Register all six bindings in `app.go` and regenerate `frontend/wailsjs/go/main/App.d.ts`.

### Risks & Edge Cases

- User quits Mashed mid-start: ensure the context used by the spawner cancels and the subprocess is cleaned up.
- Ollama takes long to boot on first run (model weight load): the poll loop honors a user-visible "starting" state — do NOT let it look frozen.
- `/api/tags` returns empty: UI shows a "No models — run `ollama pull <model>` in terminal" hint.
- Multiple Mashed windows: `startedByUs` is a process-global flag — single-instance assumption matches current Mashed design.

### Reference Files

- Plan: `docs/plans/node-path-breadcrumbs.md` lines 146-161, 193.
- Corpus: `frontend/src/views/Settings.svelte` (14909), `app.go` (mashedConfig around 30121).
- Skill: `/wails` for bindings + event lifecycle.

## Acceptance Criteria

AC-1: State machine transitions correctly
- Given Ollama is not installed
- When the Settings view mounts
- Then the Ollama section shows "Not installed" and a link to https://ollama.com

AC-2: Start / stop lifecycle
- Given Ollama is installed but not running
- When the user clicks Start
- Then the app spawns `ollama serve`, polls for health, and transitions to "running" within 10s (or fails visibly)

AC-3: Model selection persists
- Given the user is in the `running` state with ≥1 model
- When the user picks a model
- Then `mashedConfig.OllamaModel` and `OllamaEnabled = true` persist to disk
- And the view transitions to `model-selected`

AC-4: Graceful shutdown
- Given Mashed started Ollama this session
- When the app quits
- Then the Ollama subprocess is terminated (SIGTERM within 5s, SIGKILL after)

AC-5: Pre-existing Ollama untouched
- Given Ollama was already running when Mashed started
- When Mashed quits
- Then the Ollama subprocess continues running

## BDD Test Scenarios

```gherkin
Feature: Ollama settings lifecycle

  Scenario: Not installed
    Given `which ollama` returns non-zero
    When Settings mounts
    Then the section shows "Not installed" with a download link

  Scenario: Start-to-running
    Given Ollama is installed, not running
    When user clicks Start
    Then the Wails binding StartOllamaServer() is called
    And within 10s the view shows "running"

  Scenario: Model persistence
    Given state is running, models include "llama3:8b"
    When user picks "llama3:8b"
    Then mashedConfig.OllamaModel = "llama3:8b" and OllamaEnabled = true
    And config is saved to disk

  Scenario: Mashed-started subprocess killed on quit
    Given Mashed started Ollama (startedByUs = true)
    When the app quits
    Then SIGTERM is sent; if still alive after 5s SIGKILL is sent

  Scenario: Pre-existing server untouched
    Given startedByUs = false
    When the app quits
    Then no kill signal is sent
```

## Tasks / Subtasks

- [ ] Task 1: Go bindings (AC-1, AC-2, AC-3, AC-4)
  - [ ] New file `internal/ollama/lifecycle.go`: Install detection, spawn, health poll, stop
  - [ ] Add bindings to `app.go`; expose types `OllamaStatus`, `OllamaModel`
  - [ ] Wire shutdown hook to `OnShutdown` / `OnBeforeClose` Wails hook
- [ ] Task 2: Config fields (AC-3)
  - [ ] Extend `mashedConfig` with `OllamaModel` + `OllamaEnabled` (json:",omitempty")
  - [ ] Round-trip test
- [ ] Task 3: Settings UI (AC-1, AC-2, AC-3)
  - [ ] State machine in `Settings.svelte` with conditional rendering per state
  - [ ] Model dropdown fed by `ListOllamaModels()`
- [ ] Task 4: Integration test for subprocess lifecycle (AC-4, AC-5)
  - [ ] Mockable spawner; assert `startedByUs` behavior
- [ ] Task 5: Regenerate bindings
  - [ ] Update `frontend/wailsjs/go/main/App.d.ts` (ensure in PR)

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on new/modified files
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
