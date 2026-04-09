# Story 6: Frontend Terminal and Session Cleanup

**Priority:** P1-high
**Domain:** frontend
**Estimated Complexity:** S
**Depends On:** pty-04
**Status:** ready

## Description

Remove tmux-specific workarounds from the Svelte frontend: strip the `stripControlSequences` regex from `Terminal.svelte` (no longer needed since xterm.js connects directly to the managed PTY, not through tmux's alternate screen), update the connection message, and remove the `:0.0` suffix stripping from the sessions store's `makeSession` helper. These are small but necessary changes to complete the tmux removal from the user-facing pipeline.

## Developer Notes

### Architecture
- **Modify file:** `frontend/src/components/Terminal.svelte` (278 lines)
  - **Remove:** `stripRe` regex definition (lines 123-128) and `stripControlSequences` function (lines 129-131)
  - **Remove:** Call to `stripControlSequences(raw)` in `ws.onmessage` handler (line 185). Replace with just `raw`:
    ```javascript
    ws.onmessage = (evt) => {
      const raw = evt.data instanceof ArrayBuffer
        ? new TextDecoder().decode(evt.data)
        : evt.data;
      term.write(raw);
    };
    ```
  - **Remove:** Comment block about stripping escape sequences (lines 117-122)
  - **Change:** Line 166: `'Connecting to tmux session...'` -> `'Connecting...'`
  - **Keep unchanged:** Wails clipboard API (ClipboardGetText/ClipboardSetText), `--wails-draggable: no-drag`, `user-select: none`, FitAddon, resize handler, keyboard handler, theme/font reactivity

- **Modify file:** `frontend/src/lib/stores/sessions.js` (48 lines)
  - **Change:** Line 39: `sessionName: target.replace(':0.0', '')` -> `sessionName: target`
  - **Change:** Line 40: `paneTarget: target` stays the same (target is now already the session name)
  - The `makeSession` function is called from `AgentDetail.svelte` after spawn methods return. Since spawn methods now return session names without `:0.0`, the `.replace()` is no longer needed.

- **No changes to:** `frontend/src/views/AgentDetail.svelte` (already uses `paneTarget` prop which works with session names), `frontend/src/views/WorkflowBuilder.svelte` (passes `terminalTarget` which is a pane target / session name)

### Technical Considerations
- **xterm.js behavior improvement:** Without `stripControlSequences`, xterm.js will receive the raw PTY output directly. Since the managed PTY runs the shell/claude process directly (not through tmux), there are no alternate screen sequences to strip. Mouse tracking, scrolling, and text selection should work natively.
- **Binary vs text WebSocket messages:** The bridge sends `BinaryMessage` type. The `onmessage` handler correctly handles `ArrayBuffer` -> `TextDecoder`. No change needed.
- **No risk of regression:** The stripping was a workaround for tmux interference. Without tmux, stripping would only accidentally remove legitimate application escape sequences (e.g., if claude outputs mouse-tracking sequences for a TUI).

### Risks & Edge Cases
- **Programs that use alternate screen (e.g., vim, less):** With tmux removed, these programs' escape sequences flow directly to xterm.js, which handles them natively. This is the correct behavior -- xterm.js is a full terminal emulator.
- **Bracketed paste mode:** The strip regex also removed bracketed paste sequences. Without tmux, the shell itself sends these, and xterm.js handles them correctly. The Wails clipboard paste handler already wraps text in bracketed paste sequences (`\x1b[200~` + text + `\x1b[201~`), which is the standard mechanism.

### Reference Files
- `frontend/src/components/Terminal.svelte` (full file, 278 lines)
- `frontend/src/lib/stores/sessions.js` (full file, 48 lines)
- `frontend/src/views/AgentDetail.svelte` (lines 88-113 for spawn/session handling)

## Acceptance Criteria

AC-1: stripControlSequences is completely removed
- Given the updated `Terminal.svelte`
- When the file is searched for "stripControlSequences" or "stripRe"
- Then zero matches are found
- And the ws.onmessage handler writes raw data directly to term.write()

AC-2: Connection message no longer mentions tmux
- Given a terminal component connecting to a session
- When the WebSocket connection is being established
- Then the user sees "Connecting..." (not "Connecting to tmux session...")

AC-3: makeSession no longer strips :0.0 suffix
- Given the sessions store
- When `makeSession("mashed-repo-123", ...)` is called
- Then `sessionName` is "mashed-repo-123" (unchanged, no .replace())

AC-4: Terminal still works with direct PTY output
- Given a Terminal component connected to a managed session via WebSocket
- When the session produces output containing ANSI color codes and cursor movement
- Then xterm.js renders them correctly without any stripping

AC-5: Clipboard integration unchanged
- Given a Terminal component with an active session
- When the user presses Cmd+C with text selected
- Then the text is copied via `ClipboardSetText`
- And Cmd+V pastes via `ClipboardGetText` with bracketed paste wrapping

## BDD Test Scenarios

### Scenario 1: Raw Output Passthrough

```gherkin
Feature: Terminal raw output

  Scenario: PTY output rendered without stripping
    Given a Terminal component connected to a managed session
    When the session sends "\x1b[32mgreen text\x1b[0m"
    Then xterm.js renders "green text" in green
    And no escape sequences are stripped

  Scenario: Alternate screen sequences pass through
    Given a Terminal component connected to a managed session
    When the session runs a program that uses alternate screen
    Then xterm.js enters alternate screen mode natively
    And the user can scroll in normal buffer after the program exits
```

### Scenario 2: Connection Message

```gherkin
Feature: Connection status message

  Scenario: Connecting message is generic
    Given a Terminal component with a paneTarget
    When the component mounts
    Then the terminal shows "Connecting..." in dim gray
    And the message does not contain "tmux"
```

### Scenario 3: Session Name Handling

```gherkin
Feature: Session name in store

  Scenario: makeSession preserves session name as-is
    Given a spawn method returns "mashed-repo-123"
    When makeSession("mashed-repo-123", "/tmp/repo", "repo", "agent", "opus") is called
    Then sessionName is "mashed-repo-123"
    And paneTarget is "mashed-repo-123"
```

## Tasks / Subtasks

- [ ] Task 1: Remove stripControlSequences from Terminal.svelte (AC: AC-1, AC-2, AC-4)
  - [ ] Subtask 1a: Delete the `stripRe` regex definition (lines 123-128)
  - [ ] Subtask 1b: Delete the `stripControlSequences` function (lines 129-131)
  - [ ] Subtask 1c: Delete the comment block about stripping (lines 117-122)
  - [ ] Subtask 1d: Update `ws.onmessage` to pass `raw` directly to `term.write(raw)` instead of `term.write(stripControlSequences(raw))`
  - [ ] Subtask 1e: Change "Connecting to tmux session..." to "Connecting..." on line 166

- [ ] Task 2: Update sessions store (AC: AC-3)
  - [ ] Subtask 2a: Change `sessionName: target.replace(':0.0', '')` to `sessionName: target` in `makeSession`
  - [ ] Subtask 2b: Verify `paneTarget: target` is still correct (it is -- target is now session name)

- [ ] Task 3: Manual verification (AC: AC-4, AC-5)
  - [ ] Subtask 3a: Verify xterm.js renders colored output correctly with direct PTY
  - [ ] Subtask 3b: Verify mouse scroll works in the terminal
  - [ ] Subtask 3c: Verify text selection works with click-drag
  - [ ] Subtask 3d: Verify Cmd+C/Cmd+V clipboard operations work

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios verified (manual or automated)
- [ ] No "tmux" references remain in Terminal.svelte (except possibly in comments about the migration)
- [ ] No "stripControlSequences" references remain anywhere in `frontend/src/`
- [ ] `wails dev` compiles the frontend without errors
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
