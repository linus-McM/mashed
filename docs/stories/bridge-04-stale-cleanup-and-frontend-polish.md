# Story bridge-04: Stale Session Cleanup and Frontend Polish

**Priority:** P2-medium
**Domain:** fullstack
**Estimated Complexity:** S
**Depends On:** bridge-03
**Status:** ready

## Description

Finish the BMAD terminal bridge feature by (a) cleaning up orphaned `bmad-*` tmux sessions on app startup so the process list stays tidy across restarts, and (b) displaying the friendly, parsed session name in the View Terminal modal title and NodeConfigPanel so users see `surfseer · main · Create Story` instead of `bmad-surfseer-main-create-story-a1b2c3d4:0.0`. Also adds a release-notes entry calling out the one-way compatibility gap (sessions created with the old naming scheme cannot be reconnected — they must be restarted).

## Developer Notes

### Architecture

**Backend (cleanup)**

- **Modify file:** `internal/bmad/executor.go`
  - Add method `func (e *Executor) CleanupStaleSessions(ctx context.Context) error`
  - Lists all tmux sessions: `e.runCmd(ctx, "tmux", "list-sessions", "-F", "#{session_name}")`. An exit status > 0 with "no server running" is not an error — return nil.
  - Filters results to those with `bmad.SessionNamePrefix` prefix
  - Collects the set of session names referenced by any currently-tracked `execState` (`e.executions` map) — these are the "live" names
  - For every bmad-prefixed tmux session NOT in the live set: logs the kill decision, then `tmux kill-session -t {name}`
  - Returns the first error encountered, or nil on full success
  - **Must be safe to call before any workflow is started** (empty executions map → kills all bmad-* sessions)

- **Modify file:** `internal/bmad/executor_test.go`
  - Test: mock runner returning 3 bmad sessions, one tracked → the other two are killed
  - Test: "no server running" error from list-sessions is swallowed
  - Test: empty executions → all bmad sessions killed
  - Test: one kill fails → remaining kills still attempted; first error returned

- **Modify file:** `app.go`
  - In `OnStartup`, after `terminal.NewBridge(...)` and `a.bridge.Start(a.ctx)`, call `a.bmadExecutor.CleanupStaleSessions(a.ctx)` in a short-lived goroutine (don't block startup). Log errors. Do NOT fail startup on cleanup error.

**Frontend (polish)**

- **Modify file:** `frontend/src/components/bmad/NodeConfigPanel.svelte`
  - Currently line 133 shows the raw `tmuxTarget`. Update to display a friendly form when the name parses.
  - Add a pure helper `parseFriendlyTarget(target: string): { repo, branch, label, hash, raw } | null` in a new file `frontend/src/lib/bmadSessionName.ts`. Implements the same rules as Go `ParseSessionName`: strip `:0.0` suffix, check `bmad-` prefix, split into components, return null on mismatch.
  - Replace the button text / tooltip so when parsed: tooltip shows `repo / branch / label` and button label remains `View Terminal`.

- **Modify file:** `frontend/src/views/WorkflowBuilder.svelte`
  - Line 759 currently renders `Node Terminal: {terminalTarget}`. Replace with a computed reactive binding that uses `parseFriendlyTarget`. Format: `Terminal — {repo} · {branch} · {label}` when parsed; fall back to raw `terminalTarget` when not.

- **New file:** `frontend/src/lib/bmadSessionName.ts` — exports `parseFriendlyTarget` and a small TypeScript type.
- **New file:** `frontend/src/lib/bmadSessionName.test.ts` — vitest cases mirroring Go `ParseSessionName` tests.

**Docs**

- **Modify file:** `docs/SPECIFICATION.md` (or an existing changelog/release-notes doc; check for one first)
  - Add a short note: "BMAD tmux sessions are now named `bmad-{repo}-{branch}-{label}-{hash}`. Sessions created by older builds cannot be re-attached; they remain visible in `tmux ls` until cleaned up automatically on next app startup."

### Technical Considerations

- **Cleanup is best-effort.** Tmux's `list-sessions` returns non-zero exit status when no server is running. Detect this via stderr substring check or exit code; swallow silently.
- **Kill ordering:** kill oldest first if the order matters, but since these are orphans nobody is watching, no ordering guarantee is required.
- **Race on startup:** A user could start a BMAD workflow before `CleanupStaleSessions` finishes. This is fine because live workflows add their session name to the executions map *before* launching tmux (double-check this in `executeNode`; if not, move the map-write to happen before `tmux new-session`). If the ordering is already "tmux first, then map-write", document a 1-2 second delay before starting cleanup on startup.
  - **Safer alternative:** run cleanup synchronously *before* any workflow can start (before `a.bridge.Start`), since no executions can exist yet at `OnStartup`. Recommended approach — the executions map is guaranteed empty on fresh app launch, so the set of "live" names is trivially empty and cleanup is unambiguous.
- **Frontend pure function:** `parseFriendlyTarget` must NOT throw on malformed input; return `null`.
- **Svelte reactivity:** use `$:` reactive statements so the title updates when `terminalTarget` changes.

### Risks & Edge Cases

1. **User is running the app concurrently in another terminal:** cleanup will kill their other instance's BMAD sessions. Acceptable because `mashed` is single-user desktop; document in the release note.
2. **Non-BMAD sessions:** guarded by prefix check — never touched.
3. **Tmux not installed:** `list-sessions` errors; caller swallows and returns nil.
4. **Very old tmux without `-F` support:** unlikely on macOS/Linux targets; not mitigated.
5. **Frontend gets a raw session name without `:0.0`:** parser tolerates either form.
6. **Malformed session names:** parser returns null, UI falls back to raw string.

### Reference Files

- `internal/bmad/executor.go` — `Executor` struct with `executions` map, `runCmd` field
- `internal/bmad/session_naming.go` — `SessionNamePrefix`, `ParseSessionName` (from bridge-01)
- `frontend/src/components/bmad/NodeConfigPanel.svelte` lines 133-140, 382-390 — tmuxTarget display and View Terminal button
- `frontend/src/views/WorkflowBuilder.svelte` line 759 — terminal modal title
- `app.go` lines 140-170 — `OnStartup` wiring (bridge and bmadExecutor init)

Reference skills: `/wails`, `/simplify`, `/playwright-cli` (for the optional UI verification on the modal title), `/golang-testing`, `/golang-error-handling`.

## Acceptance Criteria

**AC-1: CleanupStaleSessions kills orphaned bmad sessions**
- Given a mock `CommandRunner` where `tmux list-sessions -F "#{session_name}"` returns `"bmad-a\nbmad-b\nuserfoo\n"`
- And the executor's executions map is empty
- When `CleanupStaleSessions(ctx)` is called
- Then `tmux kill-session -t bmad-a` is invoked
- And `tmux kill-session -t bmad-b` is invoked
- And `tmux kill-session -t userfoo` is NOT invoked
- And the returned error is nil

**AC-2: CleanupStaleSessions skips names referenced by live executions**
- Given an execution tracking a node whose TmuxTarget is `"bmad-a:0.0"`
- And `tmux list-sessions` returns `"bmad-a\nbmad-b\n"`
- When `CleanupStaleSessions(ctx)` is called
- Then only `bmad-b` is killed
- And `bmad-a` remains alive

**AC-3: CleanupStaleSessions swallows "no server running"**
- Given `tmux list-sessions` returns an error with stderr `"no server running on /private/tmp/tmux-.../default"`
- When `CleanupStaleSessions(ctx)` is called
- Then the returned error is nil
- And no kill commands are issued

**AC-4: CleanupStaleSessions continues after individual kill failure**
- Given 3 bmad sessions to kill and the first kill returns an error
- When `CleanupStaleSessions(ctx)` is called
- Then the remaining two kills are still attempted
- And the returned error wraps the first failure

**AC-5: App startup runs cleanup after bridge start**
- Given the app is starting up
- When `OnStartup` completes
- Then `CleanupStaleSessions` has been invoked once
- And an error from cleanup does NOT prevent startup

**AC-6: parseFriendlyTarget returns components for valid targets**
- Given the string `"bmad-surfseer-main-create-story-a1b2c3d4:0.0"`
- When `parseFriendlyTarget` is called
- Then it returns `{ repo: "surfseer", branch: "main", label: "create-story", hash: "a1b2c3d4", raw: "bmad-surfseer-main-create-story-a1b2c3d4:0.0" }`

**AC-7: parseFriendlyTarget returns null for invalid input**
- Given the strings `""`, `"garbage"`, and `"bmad-short"`
- When `parseFriendlyTarget` is called for each
- Then each returns `null`

**AC-8: Terminal modal title shows friendly form**
- Given a node with `tmuxTarget = "bmad-surfseer-main-create-story-a1b2c3d4:0.0"`
- When the user clicks View Terminal
- Then the modal title DOM contains the text `"Terminal — surfseer · main · create-story"`
- And when the target is unparseable, the title falls back to the raw string

**AC-9: Release notes mention the migration gap**
- Given a fresh checkout after this story merges
- When the developer reads the release notes / specification doc
- Then a note explains that sessions created by older builds cannot be re-attached and will be cleaned up on next startup

## BDD Test Scenarios

### Scenario 1: Cleanup behaviour

```gherkin
Feature: Stale BMAD session cleanup

  Scenario: Kill all orphans on startup
    Given a mock CommandRunner listing "bmad-a", "bmad-b", "bmad-c"
    And no live executions
    When CleanupStaleSessions is called
    Then kill-session is invoked for bmad-a, bmad-b, bmad-c
    And no other sessions are touched

  Scenario: Preserve tracked sessions
    Given a live execution tracking "bmad-surfseer-main-draft-prd-aaaabbbb:0.0"
    And list-sessions returns "bmad-surfseer-main-draft-prd-aaaabbbb" and "bmad-orphan"
    When CleanupStaleSessions is called
    Then only bmad-orphan is killed

  Scenario: No tmux server running
    Given list-sessions errors with "no server running"
    When CleanupStaleSessions is called
    Then it returns nil and no kills are attempted

  Scenario: Partial failure
    Given three orphans and the first kill fails
    When CleanupStaleSessions is called
    Then all three kills are attempted
    And the first error is returned
```

### Scenario 2: Frontend parser

```gherkin
Feature: parseFriendlyTarget

  Scenario: Valid target with :0.0 suffix
    Given input "bmad-surfseer-main-create-story-a1b2c3d4:0.0"
    When parseFriendlyTarget is called
    Then the result object has repo "surfseer", branch "main", label "create-story", hash "a1b2c3d4"

  Scenario: Valid target without suffix
    Given input "bmad-surfseer-main-create-story-a1b2c3d4"
    When parseFriendlyTarget is called
    Then the result is non-null with the same components

  Scenario: Invalid input
    Given input "garbage"
    When parseFriendlyTarget is called
    Then the result is null
```

### Scenario 3: Modal title

```gherkin
Feature: Terminal modal title

  Scenario: Friendly title for parseable target
    Given a running BMAD node with a parseable tmuxTarget
    When the user opens the terminal modal
    Then the modal title shows "Terminal — {repo} · {branch} · {label}"

  Scenario: Raw fallback
    Given a tmuxTarget that does not parse
    When the user opens the terminal modal
    Then the modal title shows the raw tmuxTarget string
```

## Tasks / Subtasks

- [ ] Task 1: Implement `CleanupStaleSessions` on Executor (AC: AC-1, AC-2, AC-3, AC-4)
  - [ ] Subtask 1a: Add method to `internal/bmad/executor.go`
  - [ ] Subtask 1b: Build the live-names set from `e.executions` under lock
  - [ ] Subtask 1c: Invoke `tmux list-sessions -F "#{session_name}"`; detect and swallow "no server running"
  - [ ] Subtask 1d: Filter by `SessionNamePrefix`, exclude live names, kill the rest, accumulate first error

- [ ] Task 2: Wire cleanup into app startup (AC: AC-5)
  - [ ] Subtask 2a: Call `a.bmadExecutor.CleanupStaleSessions(a.ctx)` near the end of `OnStartup`
  - [ ] Subtask 2b: Log errors but continue startup

- [ ] Task 3: Write cleanup tests in `executor_test.go` (AC: AC-1 through AC-4)
  - [ ] Subtask 3a: Mock runner asserting exact sequence of list/kill invocations
  - [ ] Subtask 3b: Case: no server running swallowed
  - [ ] Subtask 3c: Case: partial failure aggregation
  - [ ] Subtask 3d: Case: live execution preservation

- [ ] Task 4: Implement frontend `parseFriendlyTarget` helper (AC: AC-6, AC-7)
  - [ ] Subtask 4a: Create `frontend/src/lib/bmadSessionName.ts` with typed return value
  - [ ] Subtask 4b: Strip optional `:0.0` suffix, verify `bmad-` prefix, split and validate hash
  - [ ] Subtask 4c: Write vitest cases in `frontend/src/lib/bmadSessionName.test.ts`

- [ ] Task 5: Update `WorkflowBuilder.svelte` terminal modal title (AC: AC-8)
  - [ ] Subtask 5a: Import `parseFriendlyTarget`
  - [ ] Subtask 5b: Reactive computed title with raw fallback
  - [ ] Subtask 5c: Replace line 759 markup

- [ ] Task 6: Update `NodeConfigPanel.svelte` tooltip/button affordance (AC: AC-8)
  - [ ] Subtask 6a: Show parsed repo/branch/label in tooltip when available
  - [ ] Subtask 6b: Leave the raw target in `data-target` attribute for debugging

- [ ] Task 7: Add release-notes entry (AC: AC-9)
  - [ ] Subtask 7a: Locate `docs/SPECIFICATION.md` or any existing CHANGELOG / release notes file
  - [ ] Subtask 7b: Append a short paragraph about the new naming scheme and the one-way migration gap

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on `internal/bmad/executor.go` `CleanupStaleSessions` and `frontend/src/lib/bmadSessionName.ts`
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] Frontend tests (`npm test` or vitest) pass
- [ ] `/simplify` run on all modified code
- [ ] Manual verification checklist (non-blocking): run a BMAD workflow, confirm `tmux ls` shows friendly name, click View Terminal, confirm modal title shows parsed form, restart app, confirm `tmux ls` no longer lists orphaned `bmad-*` sessions
- [ ] Code review: no CRITICAL/HIGH issues
- [ ] Story status updated to `done`

## Design Brief

**Scope:** Two surgical text edits. No new components, no new CSS custom properties. Reuse the existing `.terminal-modal-header` chrome (WorkflowBuilder.svelte:972-1001) and the existing `title=` tooltip convention already used on `NodeConfigPanel.svelte:176`.

### 1. Title composition (WorkflowBuilder.svelte:759)

Replace the single `<span class="terminal-modal-title">` text node with a composed structure. The existing `.terminal-modal-title` container class is preserved — we only change the children.

```svelte
<span class="terminal-modal-title" title={terminalTarget}>
  {#if parsedTerminalTarget}
    <span class="tmt-prefix">Terminal</span>
    <span class="tmt-dash">—</span>
    <span class="tmt-repo">{parsedTerminalTarget.repo}</span>
    <span class="tmt-sep">·</span>
    <span class="tmt-branch">{parsedTerminalTarget.branch}</span>
    <span class="tmt-sep">·</span>
    <span class="tmt-label">{parsedTerminalTarget.label}</span>
  {:else}
    {terminalTarget}
  {/if}
</span>
```

Hierarchy intent: "Terminal —" is a muted overline; `repo · branch · label` carries the weight; separators fade into the background.

### 2. Typography plan (scoped CSS added next to `.terminal-modal-title`)

The container already sets `font-family: var(--font-mono); font-size: 11px;` (the `--text-label` scale). All children inherit that; only color/weight differ.

- `.tmt-prefix` — `color: var(--text-muted); font-weight: 400; letter-spacing: 0.04em; text-transform: uppercase;`
- `.tmt-dash` — `color: var(--text-muted); margin: 0 var(--sp-2xs);`
- `.tmt-repo` — `color: var(--text-primary); font-weight: 500;`
- `.tmt-branch` — `color: var(--text-dim); font-weight: 400;`
- `.tmt-label` — `color: var(--accent-green); font-weight: 500;` (the signature detail, see §8)
- `.tmt-sep` — `color: var(--text-muted); margin: 0 var(--sp-xs);` (middle-dot separators, subtle)

All spans keep `font-family: var(--font-mono)` via inheritance. No new font sizes.

### 3. Color plan

| Part | Token |
|---|---|
| "Terminal" prefix | `var(--text-muted)` |
| em-dash and `·` separators | `var(--text-muted)` |
| repo | `var(--text-primary)` |
| branch | `var(--text-dim)` |
| label (the action, e.g. `create-story`) | `var(--accent-green)` |
| raw fallback string | `var(--text-dim)` (matches current behavior) |

The neon-green on the `label` segment is the only accent in the entire header — it ties the modal title to the "this is live" semantic Mashed already uses for running nodes.

### 4. Tooltip spec for NodeConfigPanel

No tooltip primitive exists; the codebase uses the native `title=` attribute (`NodeConfigPanel.svelte:176`). Stay consistent — do not introduce a custom tooltip.

Update `NodeConfigPanel.svelte:382-387`:

```svelte
{#if hasTerminal}
  {@const parsed = parseFriendlyTarget(tmuxTarget)}
  <button
    class="terminal-btn"
    data-target={tmuxTarget}
    title={parsed ? `${parsed.repo} / ${parsed.branch} / ${parsed.label}` : tmuxTarget}
    on:click={() => dispatch('open-terminal', tmuxTarget)}
  >
    <Terminal size={13} />
    View Terminal
  </button>
{/if}
```

- Button label stays `View Terminal` — unchanged.
- `title` attribute carries the friendly slash-separated form (`surfseer / main / create-story`). Slashes (not middle-dots) because native OS tooltips render Unicode inconsistently and slashes read as a path.
- `data-target` preserves the raw string for DOM-inspector debugging. No visible styling change to the button itself.
- Fallback: when `parseFriendlyTarget` returns null, `title` is the raw target — same as today's implicit behavior but now explicit.

### 5. Fallback state (unparseable target)

Render the raw `terminalTarget` string as a single text node inside `.terminal-modal-title` with no extra spans — no color change, no weight change. Container's existing `color: var(--text-dim)` rule already makes it look like a monospace debug string, which is exactly right for a fallback. Do not add an error badge or warning color; a raw tmux name is not a failure state, just an unstructured one.

### 6. Empty/loading state

When `terminalTarget` is `undefined`/`''` the modal is not opened (guarded by `{#if showTerminalModal}` at WorkflowBuilder.svelte:754, and `hasTerminal` requires `status === 'running' && tmuxTarget` on NodeConfigPanel.svelte:134). No additional empty state is needed. Do not render a spinner inside the title.

### 7. Interaction notes

- **Hover reveal of raw string:** the outer `<span class="terminal-modal-title" title={terminalTarget}>` carries a native `title` attribute holding the full raw string including `:0.0`. Hovering the parsed title reveals the debug form. This matches the existing `title="Close"` pattern on the close button and requires no new tooltip machinery.
- **Cursor:** no change — this is text, not a button.
- **Text selection:** keep default `user-select` behavior so a developer can copy the friendly form.
- **Overflow:** `.terminal-modal-title` already has `overflow: hidden; text-overflow: ellipsis; white-space: nowrap;` — the composed spans inherit that containment. On a narrow window the label (rightmost, accent-green) truncates first, which is the right priority since it's the most recognizable segment.

### 8. Signature detail

The `label` segment in `var(--accent-green)` is the one intentional flourish. Rationale: this one word (e.g. `create-story`, `draft-prd`) is what the developer actually cares about when they glance at the modal — which BMAD action is running right now. Lighting it neon-green ties the terminal modal into the same "alive / active / go" semantic that Mashed uses for running nodes throughout the app (DESIGN.md line 27). Everything else in the header stays muted so this single hit of color reads as purposeful, not decorative. Restraint is the point: one accent per header, no gradient, no icon, no badge.

### Tokens wished-for (not blocking)

- A `--text-overline` letter-spacing token (`0.04em`) would let other overline uses share the spec. Currently inlined.
- A dedicated `--text-accent` alias pointing to `--accent-green` for semantic "this is the live thing" uses, decoupled from the literal color. Not needed for this story.
