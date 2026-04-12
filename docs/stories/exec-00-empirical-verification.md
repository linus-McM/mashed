# exec-00: Empirical verification gate — slash-injection, idle stability, crash survival

**Status:** ready
**Domain:** backend
**Size:** M
**Depends on:** none
**Phase:** 3

## Description

**This story is a BLOCKING GATE for every other Phase 3 story.** Before a single line of `executeCommandNode` or `waitForIdleCompletion` is written, three empirical invariants about Claude Code CLI + tmux behaviour must be verified by recorded fixture tests. If any one invariant fails, the Phase 3 design changes and subsequent stories must be re-scoped.

The three invariants (plan §Phase 3 "Before you write ANY code, verify three empirical invariants"):

1. **Slash-command injection via `tmux send-keys -H`** — firing `/simplify\n` as hex bytes into a live claude pane causes claude to recognise and execute the slash command.
2. **Pane hash stability at idle** — three consecutive `tmux capture-pane -p` invocations on an idle claude pane produce identical SHA256 hashes. The `pollForIdle` state machine's "two stable polls" guard depends on this.
3. **Session survival after `/exit`** — wrapping claude in `bash -c 'claude ...; exec bash'` leaves the tmux pane alive after claude exits, so downstream command nodes can reuse the session.

Each invariant is captured as a recorded fixture (output bytes + pane metadata) so the Go test suite can replay it via the existing `CommandRunner` mock without depending on a live claude binary. Outcomes go into the eventual Phase 3 commit message.

## Developer Notes

- **Files to create/modify:**
  - `internal/bmad/exec_verify_test.go` — NEW. Hosts the three replayable fixture tests.
  - `internal/bmad/testdata/exec_verify/` — NEW directory.
    - `verify1_slash_injection.txt` — raw pane capture bytes after `/simplify\n` injection.
    - `verify2_idle_stable_pane.txt` — pane capture bytes for three consecutive polls (identical).
    - `verify3_session_survives_exit_panes.txt` — output of `tmux list-panes -F '#{pane_dead} #{pane_current_command}'` after claude exits.
  - `docs/plans/skills-and-session-reuse.md` — APPEND a short "Verification outcomes" subsection under Phase 3 capturing results (pass/fail, escape hatch used if any).
- **Types/symbols introduced:**
  - `ErrIdleTimeoutNoStart` sentinel (exported) — wired later in exec-01 but declared here so the verification test can reference the contract.
  - Test helpers `captureFixtureSlashInjection`, `captureFixtureIdleStability`, `captureFixtureSessionSurvival` — only used inside `exec_verify_test.go`.
- **How to actually run the real tmux steps (one-time manual dry run before recording fixtures):**
  - Verification 1:
    ```bash
    tmux new-session -d -s verify-1 'claude --dangerously-skip-permissions'
    sleep 8
    tmux send-keys -H -t verify-1:0.0 2f 73 69 6d 70 6c 69 66 79 0d
    tmux capture-pane -t verify-1:0.0 -p > internal/bmad/testdata/exec_verify/verify1_slash_injection.txt
    tmux kill-session -t verify-1
    ```
  - Verification 2:
    ```bash
    tmux new-session -d -s verify-2 'claude --dangerously-skip-permissions'
    sleep 10
    for i in 1 2 3; do tmux capture-pane -t verify-2:0.0 -p >> internal/bmad/testdata/exec_verify/verify2_idle_stable_pane.txt; sleep 3; done
    tmux kill-session -t verify-2
    ```
  - Verification 3:
    ```bash
    tmux new-session -d -s verify-3 -c "$HOME" 'bash -c "claude --dangerously-skip-permissions; exec bash"'
    sleep 10
    tmux send-keys -t verify-3:0.0 '/exit' Enter
    sleep 3
    tmux list-panes -t verify-3 -F '#{pane_dead} #{pane_current_command}' > internal/bmad/testdata/exec_verify/verify3_session_survives_exit_panes.txt
    tmux kill-session -t verify-3
    ```
- **Risks / gotchas:**
  - **Plan §Phase 3 "Verification 1" FAIL MODES:** if claude doesn't recognise the piped hex bytes, escape hatches are (a) byte-at-a-time with 50ms delay, (b) `send-keys -l <name>` + separate `Enter`. Document which path was used in the commit.
  - **Plan §Phase 3 "Verification 2" FAIL MODE:** hashes differ due to animated spinner / status bar. Escape hatch: strip status bar before hashing, OR require three stable polls instead of two. Document the choice.
  - **Plan §Phase 3 "Verification 3" FAIL MODE:** user's `$SHELL` isn't bash or errexit kills the wrap. Escape hatch: `bash --norc -c '... ; exec bash'` or `$SHELL` literal. Document the choice.
  - **Test must be able to run CI-offline.** Do NOT gate CI on a live claude binary. Record once, replay forever. The fixture is the test, not a smoke script.
  - `tmux` is unavailable on some CI runners. Tests that replay fixtures must not shell out to tmux at all — they parse the recorded bytes directly.
- **Prerequisites already in place:**
  - `internal/bmad/question.go`'s `detectIdlePrompt` + `hashCapturedOutput` already exist.
  - `internal/terminal/tmux_adapter.go:524+` already emits `send-keys -H` with correct hex. Verification 1's fixture test must assert the exact argv pattern.
  - `fixture_verify_test.go` in `internal/bmad` is an existing precedent for the replay-fixture pattern.

## Acceptance Criteria

**AC-1: Slash-command injection fixture replays cleanly**
- Given the recorded pane capture at `testdata/exec_verify/verify1_slash_injection.txt`
- When the fixture bytes are fed through `detectIdlePrompt` and a "slash command recognised" detector
- Then the test asserts claude echoed `/simplify` and produced subsequent output (i.e. the slash command was recognised, not left in the input buffer)
- And the argv produced by `tmux_adapter.SendInput([]byte("/simplify\n"))` matches the exact byte sequence `send-keys -H -t <target> 2f 73 69 6d 70 6c 69 66 79 0a` (or `0d` — document which was chosen)

**AC-2: Idle-prompt stability fixture holds across three polls**
- Given the three-capture fixture at `testdata/exec_verify/verify2_idle_stable_pane.txt`
- When each capture is passed through `hashCapturedOutput`
- Then all three hashes are identical
- And `detectIdlePrompt` returns true on each capture

**AC-3: Session survival fixture confirms pane alive post-exit**
- Given the fixture at `testdata/exec_verify/verify3_session_survives_exit_panes.txt`
- When the fixture output is parsed for `#{pane_dead} #{pane_current_command}`
- Then `pane_dead == 0`
- And `pane_current_command` is a shell (bash/zsh/etc.), NOT `claude`

**AC-4: Verification outcomes documented**
- Given the three fixture tests pass
- When the Phase 3 plan document is inspected
- Then a "Verification outcomes" subsection exists under Phase 3 with date, pass/fail status, and any escape hatches used
- And the outcomes are also present in the git commit message body

## BDD Test Scenarios (Gherkin)

```gherkin
Feature: Phase 3 empirical verification gate

  Scenario: Slash injection fixture proves recognition
    Given the recorded slash-injection pane capture fixture
    When the test replays it through the idle detector
    Then the detector confirms claude processed the /simplify command
    And the exact tmux send-keys -H hex byte sequence is verified

  Scenario: Idle-stable fixture holds across three polls
    Given three consecutive pane captures taken three seconds apart on an idle claude pane
    When each capture is hashed with hashCapturedOutput
    Then all three hashes are byte-identical
    And detectIdlePrompt returns true for every capture

  Scenario: bash -c exec bash wrapper preserves pane after claude exit
    Given a fixture capture of "tmux list-panes -F #{pane_dead} #{pane_current_command}" after /exit
    When the fixture is parsed
    Then pane_dead is 0
    And pane_current_command is a shell (not "claude")

  Scenario: Verification outcomes are recorded in the plan document
    Given the three fixture tests have run and passed
    When the plan document is read
    Then a "Verification outcomes" subsection exists under Phase 3
    And it names each verification and records pass/fail
```

## Tasks / Subtasks

- [ ] Task 1 — Record the three fixtures manually (AC-1, AC-2, AC-3)
  - [ ] Run the Verification 1 script against a live claude session; commit the capture file
  - [ ] Run the Verification 2 script; commit the multi-capture file
  - [ ] Run the Verification 3 script; commit the list-panes output
  - [ ] If any fails, document the escape hatch in plan §Phase 3 "Verification outcomes" and re-record
- [ ] Task 2 — Write the three replay tests (AC-1, AC-2, AC-3)
  - [ ] `TestVerifySlashInjectionFixture` — parses the capture, asserts recognition, asserts expected `send-keys -H` argv from `tmux_adapter.SendInput`
  - [ ] `TestVerifyIdleStabilityFixture` — hashes all three captures and asserts equality
  - [ ] `TestVerifySessionSurvivalFixture` — parses list-panes output, asserts `pane_dead=0` and shell current command
- [ ] Task 3 — Declare `ErrIdleTimeoutNoStart` sentinel (for exec-01 handoff)
  - [ ] Add `var ErrIdleTimeoutNoStart = errors.New("bmad: idle wait timed out before claude produced output")` in `internal/bmad/executor.go`
- [ ] Task 4 — Document verification outcomes (AC-4)
  - [ ] Append a "Verification outcomes" subsection to `docs/plans/skills-and-session-reuse.md` with date, results, escape hatches
  - [ ] Ensure the eventual Phase 3 commit message includes the same table

## Definition of Done

- [ ] All ACs verified by an automated test (Go table-driven replay tests; no live tmux; no "manually verified")
- [ ] Coverage ≥ 80% on modified files (`go test -coverprofile=cover.out ./internal/bmad/... && go tool cover -func=cover.out`)
- [ ] `go build ./... && go vet ./...` clean
- [ ] `go test ./... -race -short` clean, offline, no tmux required
- [ ] `/simplify` run before sign-off
- [ ] No hardcoded paths or magic numbers added (fixture paths via `testdata/` convention only)
- [ ] Plan doc "Verification outcomes" subsection written
- [ ] Existing tests still pass
