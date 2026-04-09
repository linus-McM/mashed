# Story 6: Screenshot-to-Claude-Code -- Frontend Injection

**Priority:** P1-high
**Domain:** frontend
**Estimated Complexity:** S
**Depends On:** Story 5 (meditor-05)
**Status:** ready

## Description

Wire the frontend to receive `screenshot:inject` Wails events and inject the screenshot file path into the correct terminal session via the existing WebSocket bridge. Add `SetActiveContext` calls in `AgentDetail.svelte` when the view mounts or the active session changes, and clear context in `App.svelte` when navigating away. The screenshot path is sent as bracketed paste + Enter, causing Claude Code CLI to read the image.

## Developer Notes

### Architecture
- **Modified files:**
  - `frontend/src/components/Terminal.svelte` -- add scoped `screenshot:inject` event listener
  - `frontend/src/views/AgentDetail.svelte` -- add `SetActiveContext` call on mount/session change
  - `frontend/src/App.svelte` -- clear context on `goBack()`
- **Data flow:**
  1. Go emits `screenshot:inject` event with `{ path, paneTarget }`
  2. Every mounted Terminal.svelte receives the event
  3. Only the terminal whose `paneTarget` prop matches `data.paneTarget` responds
  4. Matching terminal sends the path via `ws.send()` as bracketed paste + Enter
  5. Go bridge receives binary data, writes to the session's pty
  6. Claude Code CLI receives the path as typed input and reads the image

### Technical Considerations
- **Event listener lifecycle:** Subscribe in `onMount` (after WebSocket opens), unsubscribe in `onDestroy`. Use Wails runtime `EventsOn` which returns an unsubscribe function.
- **Bracketed paste format:** Same as existing `pasteToTerminal` function:
  ```javascript
  const encoder = new TextEncoder();
  ws.send(encoder.encode('\x1b[200~' + data.path + '\x1b[201~'));
  ws.send(encoder.encode('\r')); // Enter to submit
  ```
- **WebSocket readiness guard:** Check `ws && ws.readyState === WebSocket.OPEN` before sending. If not ready, log warning and skip.
- **paneTarget matching:** The `paneTarget` prop on Terminal.svelte is the session identifier. Compare `data.paneTarget` with the component's `paneTarget` prop.
- **SetActiveContext timing:** Use a reactive statement `$: if (agent?.repoPath && activeSession?.paneTarget)` so it fires on every session tab switch. Import `SetActiveContext` from the Wails bindings.
- **Clear on goBack:** In `App.svelte`, call `SetActiveContext('', '')` in the `goBack()` function (or wherever the view transitions from agent detail to feed).

### Risks & Edge Cases
- **Multiple terminals mounted:** Multiple Terminal.svelte instances may be mounted simultaneously (e.g., in different tabs). The paneTarget matching ensures only one responds.
- **Claude Code not at prompt:** If Claude Code is mid-output when the screenshot path arrives, the injected text may be garbled. Mitigation: `screencapture -i` takes focus, so Claude is unlikely to be streaming.
- **WebSocket not yet open:** If the terminal just mounted and the WebSocket handshake is in progress, the event arrives before `ws.onopen`. Guard with readyState check.
- **Session with no WebSocket:** Log-only terminals (no `paneTarget`) don't have a WebSocket. They won't match the event, so no action needed.

### Reference Files
- `frontend/src/components/Terminal.svelte:109-115` -- existing `pasteToTerminal` function (copy this pattern)
- `frontend/src/components/Terminal.svelte:164-229` -- WebSocket setup (subscribe after `ws.onopen`)
- `frontend/src/views/AgentDetail.svelte:85-102` -- session tab state (`activeSession`, `sessions`)
- `frontend/src/App.svelte:92-96` -- existing `screenshot:taken` listener (co-locate new logic)
- `docs/feasibility-multi-editor.md` -- Sections 8.5, 8.6

## Acceptance Criteria

AC-1: Screenshot path injection
- Given a Terminal.svelte instance with `paneTarget="myrepo:agent-1"` and an open WebSocket
- When a `screenshot:inject` event fires with `{ path: "/repo/.screenshots/shot.png", paneTarget: "myrepo:agent-1" }`
- Then the terminal sends the path as bracketed paste followed by Enter via the WebSocket
- And Claude Code CLI receives the path as typed input

AC-2: Scoped event matching
- Given two Terminal instances with paneTargets `"myrepo:agent-1"` and `"myrepo:agent-2"`
- When a `screenshot:inject` event fires with `paneTarget: "myrepo:agent-1"`
- Then only the terminal with matching paneTarget sends the path
- And the other terminal ignores the event

AC-3: SetActiveContext on mount/session change
- Given the user is viewing AgentDetail for repo "myrepo"
- When the active session tab changes to `paneTarget="myrepo:agent-2"`
- Then `SetActiveContext("myrepo", "myrepo:agent-2")` is called via the Wails binding

AC-4: Clear context on navigation
- Given the user is viewing AgentDetail
- When the user navigates back to the feed (goBack)
- Then `SetActiveContext("", "")` is called
- And subsequent screenshots will be skipped (no active context)

AC-5: Event listener cleanup
- Given a Terminal.svelte instance has subscribed to `screenshot:inject`
- When the component is destroyed
- Then the event listener is unsubscribed
- And no memory leaks or stale listeners remain

AC-6: WebSocket not ready guard
- Given a Terminal.svelte instance where the WebSocket is not yet open
- When a `screenshot:inject` event fires with a matching paneTarget
- Then no data is sent
- And a warning is logged to the console

## BDD Test Scenarios

### Scenario 1: Screenshot injection via WebSocket
```gherkin
Feature: Screenshot path injection into terminal

  Scenario: Inject screenshot path into matching terminal
    Given Terminal.svelte is mounted with paneTarget="myrepo:agent-1"
    And the WebSocket is open
    When "screenshot:inject" event fires with { path: "/repo/.screenshots/shot.png", paneTarget: "myrepo:agent-1" }
    Then ws.send is called with bracketed paste bytes for "/repo/.screenshots/shot.png"
    And ws.send is called again with carriage return bytes

  Scenario: Ignore non-matching paneTarget
    Given Terminal.svelte is mounted with paneTarget="myrepo:agent-1"
    When "screenshot:inject" event fires with { paneTarget: "myrepo:agent-2" }
    Then ws.send is not called

  Scenario: Skip injection when WebSocket not ready
    Given Terminal.svelte is mounted but WebSocket is in CONNECTING state
    When "screenshot:inject" event fires with matching paneTarget
    Then ws.send is not called
    And a console.warn is logged
```

### Scenario 2: Active context tracking
```gherkin
Feature: Frontend pushes active context to Go

  Scenario: Set context on AgentDetail mount
    Given the user navigates to AgentDetail for repo "myrepo"
    And the active session has paneTarget "myrepo:agent-1"
    When AgentDetail mounts
    Then SetActiveContext("myrepo", "myrepo:agent-1") is called

  Scenario: Update context on session tab switch
    Given AgentDetail is mounted with sessions for "myrepo"
    When the user clicks the tab for "myrepo:agent-2"
    Then SetActiveContext is called with ("myrepo", "myrepo:agent-2")

  Scenario: Clear context on goBack
    Given the user is viewing AgentDetail
    When the user clicks back to the feed
    Then SetActiveContext("", "") is called
```

### Scenario 3: Cleanup
```gherkin
Feature: Event listener cleanup on destroy

  Scenario: Unsubscribe on destroy
    Given Terminal.svelte has subscribed to "screenshot:inject"
    When the component is destroyed (onDestroy fires)
    Then the EventsOn unsubscribe function is called
    And no screenshot:inject listeners remain for this component
```

## Tasks / Subtasks

- [ ] Task 1: Add screenshot:inject listener to Terminal.svelte (AC: 1, 2, 5, 6)
  - [ ] Subtask 1a: Import `EventsOn` from Wails runtime (already imported in Terminal.svelte)
  - [ ] Subtask 1b: Subscribe to `screenshot:inject` inside `onMount`, after WebSocket setup
  - [ ] Subtask 1c: Implement paneTarget matching guard (`if (data.paneTarget !== paneTarget) return`)
  - [ ] Subtask 1d: Implement WebSocket readiness guard (`if (!ws || ws.readyState !== WebSocket.OPEN) return`)
  - [ ] Subtask 1e: Send path as bracketed paste + Enter via `ws.send(encoder.encode(...))`
  - [ ] Subtask 1f: Store unsubscribe function and call it in `onDestroy`

- [ ] Task 2: Add SetActiveContext calls to AgentDetail.svelte (AC: 3)
  - [ ] Subtask 2a: Import `SetActiveContext` from Wails bindings
  - [ ] Subtask 2b: Add reactive statement: `$: if (agent?.repoPath && activeSession?.paneTarget) SetActiveContext(agent.repoPath, activeSession.paneTarget)`

- [ ] Task 3: Clear context on navigation in App.svelte (AC: 4)
  - [ ] Subtask 3a: Import `SetActiveContext` from Wails bindings
  - [ ] Subtask 3b: Add `SetActiveContext('', '')` call in `goBack()` function

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on new/modified files
- [ ] Screenshot inject works end-to-end: Cmd+Shift+S -> path appears in terminal
- [ ] Only the matching terminal receives the injected path
- [ ] Event listener properly cleaned up on component destroy
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
