# uiadapter-v3-15: ClaudeCodeCLIBackend

**Status:** ready
**Domain:** backend
**Size:** L
**Depends On:** B, C, v3-06
**Priority:** P1-high

## Story

As a user committed to CLI workflows, I want an opt-in `ClaudeCodeCLIBackend` that wraps `claude -p --output-format stream-json`, reuses session IDs for internal prompt caching, and cleans up subprocesses deterministically, so that CLI users get the same UIAST pipeline without maintaining a separate code path.

## Description

Plan §3 Phase 4 "Story 15 — `ClaudeCodeCLIBackend`" (lines 601–627). Subprocess wrapper for `claude -p`. Softer parse-rate target (≥0.95) because there's no `tool_use` hook from outside the CLI — contract is "fenced JSON in the final assistant message."

Transport: `os/exec.CommandContext` with arguments:
```
claude -p
  --output-format stream-json
  --include-partial-messages
  --append-system-prompt "<static prefix>"
  --disallowedTools "*"
  --permission-mode bypassPermissions
  --input-format text
```

Session reuse: extract session ID from first call's `init` event, pass `--resume <session-id>` on subsequent calls. One session per backend instance; rotate on error. Lifecycle: always `cmd.Wait()`; `cmd.Cancel` sends SIGTERM then SIGKILL on context cancel. No zombies.

## Developer Notes

- **Files (new):**
  - `internal/uiadapter/backend/claudecli/backend.go`
  - `internal/uiadapter/backend/claudecli/process.go` (exec wrapper)
  - `internal/uiadapter/backend/claudecli/session.go` (session-id reuse)
- **Types / API surface:**
  ```go
  type backend struct { binary string; extraFlags []string; sessionID atomic.Value /* string */ }
  func New(cfg Config) (*backend, error)
  func init() { backend.Register("claude-cli", New) }
  ```
- **Parsing:** `json.Decoder` over stdout; one JSON object per line. Consume `assistant` events; concatenate text content blocks; extract the fenced JSON block at end-of-turn. Validate against per-kind schema.
- **Health:** `claude --version` at WarmUp; missing binary surfaces useful error.
- **Sharp edges:** cold subprocess start 300–800ms; `--resume` mitigates but still pays fork. Opt-in. Network-connected host should prefer `ClaudeAPIBackend`.
- **Risks:** §8 "CLI fragility" — softer eval target; opt-in.
- **Dependencies:** `os/exec` stdlib only.
- **Use-repo-code directive:** use `use-repo-code` to locate existing claude-CLI wrappers (`app_review.go` — cerebrum §35) for best practices (`--bare`, stdin for content, `--system-prompt`).

## Acceptance Criteria

AC-15.1: `TestClaudeCLI_VersionCheck` — missing `claude` binary returns a useful error from `WarmUp`.

AC-15.2: `TestClaudeCLI_SessionResume` — second call uses `--resume <id>` captured from first call's init event.

AC-15.3: `TestClaudeCLI_ParsingHandlesShapes` — parses: (a) clean fenced JSON, (b) JSON with trailing prose, (c) no JSON at all (falls to fallback tier).

AC-15.4: `TestClaudeCLI_ProcessCancel` — subprocess killed on ctx cancellation; no zombies under 1000-iteration stress.

AC-15.5: (CLI) `parse_rate ≥ 0.95` on the eval corpus.

## BDD Test Scenarios

```gherkin
Feature: ClaudeCodeCLIBackend

  Scenario: AC-15.1 — Missing binary surfaces error
    Given cfg.ClaudeCLIBinary="/no/such/claude"
    When WarmUp runs
    Then it returns an error naming the missing path

  Scenario: AC-15.2 — Session resume
    Given the first call's stdout includes `{"type":"init","session_id":"abc"}`
    When the second call is made
    Then its argv includes `--resume abc`

  Scenario: AC-15.3 — Three parse shapes
    Given three stdout fixtures: clean fenced JSON, JSON+trailing prose, no JSON
    When the parser runs on each
    Then the first two return valid UIASTs
    And the third falls through to the fallback tier

  Scenario: AC-15.4 — No zombies
    Given 1000 iterations of Translate with ctx cancellation mid-flight
    When the test ends
    Then `ps` (or /proc) shows no lingering claude subprocesses

  Scenario: AC-15.5 — Parse rate on corpus
    Given the eval corpus
    When the CLI backend processes all entries
    Then parse_rate ≥ 0.95
```

## Tasks / Subtasks

- [ ] Task 1 — Exec wrapper + argument assembly (maps to AC-15.1)
- [ ] Task 2 — Stream-JSON parser (maps to AC-15.3)
- [ ] Task 3 — Session-id extraction + reuse (maps to AC-15.2)
- [ ] Task 4 — Deterministic lifecycle (SIGTERM→SIGKILL) (maps to AC-15.4)
- [ ] Task 5 — `claude --version` in WarmUp (maps to AC-15.1)
- [ ] Task 6 — Tests (maps to AC-15.1–15.5)

## Definition of Done

- [ ] All ACs verified with PASS evidence
- [ ] 80%+ line coverage on modified Go files
- [ ] `/simplify` skill run
- [ ] `go build ./... && go vet ./...` clean
- [ ] `go test ./... -race -count=1 -short` clean
- [ ] Story file `Status: ready` → `done`; all `- [ ]` → `- [x]`

## Design Brief

_(backend story — no design brief)_
