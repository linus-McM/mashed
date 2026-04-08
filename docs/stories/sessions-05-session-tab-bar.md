# Story 5: Session Tab Bar in AgentDetail View

**Priority:** P1-high
**Domain:** frontend
**Estimated Complexity:** L
**Depends On:** sessions-04
**Status:** done

## Description

Add a tab bar above the terminal in `AgentDetail.svelte` that shows all terminal sessions for the current repo. Users can switch between sessions (each tab reconnects the Terminal component to a different tmux pane), spawn new terminals/agents in the same repo, and kill individual sessions. This is the user-facing feature that makes persistent sessions discoverable and usable.

## Developer Notes

### Architecture

- **Modified file** `frontend/src/views/AgentDetail.svelte` — major changes to import the sessions store, render a tab bar, and manage active session switching.
- **Modified file** `frontend/src/views/NotificationFeed.svelte` — after spawn, call `addSession` to pre-populate the store before navigation.
- **No new files** — all UI is in the existing AgentDetail view.

### AgentDetail.svelte Changes

1. **Imports**: Add `import { repoSessions, refreshSessions } from '../lib/stores/sessions.js';` and `import { KillTerminalSession, SpawnTerminal, SpawnAgent } from '../../wailsjs/go/main/App.js';`

2. **On mount**: Call `refreshSessions(agent.repoPath)` to populate the store.

3. **Derive sessions**: `$: sessions = $repoSessions[agent?.repoPath] || [];`

4. **Active session tracking**: 
   ```js
   let activeSessionIdx = 0;
   $: {
       // Default to the session matching the agent's tmuxTarget
       const matchIdx = sessions.findIndex(s => s.paneTarget === agent?.tmuxTarget);
       if (matchIdx >= 0) activeSessionIdx = matchIdx;
   }
   $: activeSession = sessions[activeSessionIdx] || null;
   ```

5. **Terminal keying**: Replace the current `<Terminal paneTarget={agent?.tmuxTarget} ... />` with:
   ```svelte
   {#key activeSession?.paneTarget}
       <Terminal paneTarget={activeSession?.paneTarget || agent?.tmuxTarget || ''} repoPath={agent?.repoPath || ''} />
   {/key}
   ```
   The `{#key}` block destroys and recreates the Terminal component when switching tabs, which closes the old WebSocket and opens a new one.

6. **Tab bar markup**: Render above the terminal pane, inside `.workspace`:
   ```svelte
   <div class="session-tabs">
       {#each sessions as session, idx}
           <button
               class="session-tab"
               class:active={idx === activeSessionIdx}
               on:click={() => activeSessionIdx = idx}
           >
               <span class="tab-type">{session.sessionType === 'agent' ? 'Agent' : 'Term'}</span>
               <span class="tab-name">{session.sessionName.split('-').slice(-1)[0]}</span>
               <button class="tab-close" on:click|stopPropagation={() => killSession(session.sessionName)}>x</button>
           </button>
       {/each}
       <button class="session-tab add-tab" on:click={spawnNewTerminal}>+</button>
   </div>
   ```

7. **Kill/spawn handlers**:
   ```js
   async function killSession(sessionName) {
       await KillTerminalSession(sessionName);
       // If we killed the active tab, switch to the first remaining
       if (sessions[activeSessionIdx]?.sessionName === sessionName) {
           activeSessionIdx = 0;
       }
   }

   async function spawnNewTerminal() {
       const target = await SpawnTerminal(agent.repoPath);
       // Store is updated via the event listener in App.svelte
       // Switch to the new tab after a brief delay for the event to propagate
       setTimeout(() => {
           const newIdx = sessions.findIndex(s => s.paneTarget === target);
           if (newIdx >= 0) activeSessionIdx = newIdx;
       }, 100);
   }
   ```

### NotificationFeed.svelte Changes

In `spawnTerminalInRepo()` (line 338) and `onSessionSpawn()` (line 301), after the spawn succeeds, call `addSession` to pre-populate the sessions store so the AgentDetail view sees the session immediately on mount:

```js
import { addSession } from '../lib/stores/sessions.js';

// In spawnTerminalInRepo, after const target = await SpawnTerminal(...):
addSession(repo.path, {
    sessionName: target.replace(':0.0', ''),
    paneTarget: target,
    repoPath: repo.path,
    repoName: repo.name,
    sessionType: 'terminal',
    model: '',
    spawnedAt: new Date().toISOString(),
    isAlive: true,
});
```

### Styling

Tab bar should follow the project's design system (`DESIGN.md`):
- Background: `var(--bg-deeper)`
- Active tab: `var(--border-accent)` bottom border, `var(--fg-primary)` text
- Inactive tab: `var(--fg-secondary)` text
- Close button: `var(--fg-dim)`, hover `var(--accent-red, #ff5f57)`
- Add button: `var(--fg-dim)`, hover `var(--accent-green)`
- Height: 32px, compact, no wasted space
- Overflow: horizontal scroll when many tabs

### Technical Considerations

- **Terminal reconnection delay**: When switching tabs, the `{#key}` block destroys and recreates `Terminal.svelte`. The WebSocket reconnects and the bridge sends scrollback. This takes ~50ms. Acceptable for v1.
- **Reactivity chain**: `refreshSessions` -> store update -> `$repoSessions` -> `sessions` derived -> tab bar re-renders. All reactive, no manual refresh needed after events.
- **Tab index stability**: When a session is killed, `activeSessionIdx` may point to a deleted entry. The `killSession` handler resets to 0 as a fallback.

### Risks & Edge Cases

- **No sessions in store on mount**: If the backend hasn't been updated yet (Stories 1-3 not deployed), `ListRepoSessions` won't exist. Guard with `try/catch` in `onMount`.
- **Agent spawned from feed has no registry entry yet**: The `onSpawned` handler in `App.svelte` navigates to detail immediately. The event listener and `addSession` call in NotificationFeed pre-populate the store. `refreshSessions` on mount is the safety net.
- **Tab switch during git operation**: Git action state (`gitAction`, `gitResult`) is per-view, not per-session. Switching tabs doesn't affect in-progress git operations. The terminal output in the new tab is independent.
- **Single session (no tabs needed)**: When `sessions.length <= 1`, the tab bar can still render — shows one tab plus the `+` button. Simple and consistent.

### Reference Files

- `frontend/src/views/AgentDetail.svelte` — primary modification target
- `frontend/src/views/NotificationFeed.svelte` — add `addSession` calls
- `frontend/src/components/Terminal.svelte` — component being keyed (no changes needed)
- `frontend/src/lib/stores/sessions.js` — store from Story 4
- `DESIGN.md` — styling variables and conventions

## Acceptance Criteria

AC-1: Tab bar renders for repo sessions
- Given AgentDetail is mounted with an agent in repo "/dev/foo"
- And the backend returns 2 sessions for "/dev/foo"
- When the view renders
- Then a tab bar appears above the terminal with 2 session tabs and a "+" button

AC-2: Clicking a tab switches the terminal
- Given the tab bar shows sessions "term-foo-1" (active) and "mashed-foo-2"
- When the user clicks the "mashed-foo-2" tab
- Then the Terminal component reconnects to paneTarget "mashed-foo-2:0.0"
- And the "mashed-foo-2" tab has the active style

AC-3: Kill button removes session
- Given the tab bar shows 2 sessions
- When the user clicks the "x" button on the inactive session's tab
- Then `KillTerminalSession` is called with that session's name
- And the tab is removed from the bar
- And the active session remains unchanged

AC-4: "+" button spawns new terminal in same repo
- Given AgentDetail is showing repo "/dev/foo"
- When the user clicks the "+" button
- Then `SpawnTerminal("/dev/foo")` is called
- And a new tab appears in the tab bar
- And the new tab becomes active

AC-5: Default active tab matches agent's tmuxTarget
- Given AgentDetail receives agent with `tmuxTarget: "mashed-foo-100:0.0"`
- And the sessions store has that session
- When the view mounts
- Then the tab for "mashed-foo-100" is the active tab

AC-6: NotificationFeed pre-populates store on spawn
- Given the user clicks "Terminal" on a repo card in NotificationFeed
- When the spawn succeeds
- Then `addSession` is called with the new session data
- And navigating to AgentDetail shows the session immediately in the tab bar

## BDD Test Scenarios

### Scenario 1: Tab Bar Rendering

```gherkin
Feature: Session tab bar

  Scenario: Tabs render from sessions store
    Given the sessions store has 2 sessions for "/dev/foo"
    And AgentDetail is mounted with agent repoPath "/dev/foo"
    When the component renders
    Then 2 session tab buttons are visible
    And 1 "+" add button is visible
    And the tab matching agent.tmuxTarget has the "active" class
```

### Scenario 2: Tab Switching

```gherkin
Feature: Tab switching reconnects terminal

  Scenario: Click inactive tab switches terminal
    Given the tab bar has tabs "term-foo-1" (active) and "term-foo-2"
    When the user clicks "term-foo-2"
    Then the Terminal component's paneTarget becomes "term-foo-2:0.0"
    And "term-foo-2" tab has the "active" class
    And "term-foo-1" tab no longer has the "active" class
```

### Scenario 3: Kill Session

```gherkin
Feature: Kill session from tab

  Scenario: Kill removes tab and calls backend
    Given the tab bar has tabs "term-foo-1" (active) and "term-foo-2"
    When the user clicks "x" on "term-foo-2"
    Then KillTerminalSession is called with "term-foo-2"
    And only 1 tab remains
    And "term-foo-1" is still the active tab

  Scenario: Kill the active tab switches to first remaining
    Given the tab bar has tabs "term-foo-1" (active) and "term-foo-2"
    When the user clicks "x" on "term-foo-1"
    Then KillTerminalSession is called with "term-foo-1"
    And "term-foo-2" becomes the active tab
```

### Scenario 4: Spawn New Session

```gherkin
Feature: Spawn from tab bar

  Scenario: "+" spawns terminal and adds tab
    Given the tab bar shows 1 session for "/dev/foo"
    When the user clicks "+"
    Then SpawnTerminal("/dev/foo") is called
    And after the event propagates, 2 tabs are visible
    And the new tab is active
```

### Scenario 5: Pre-Population from Feed

```gherkin
Feature: NotificationFeed pre-populates store

  Scenario: Terminal spawn in feed populates store before navigation
    Given the user is on NotificationFeed
    When they click "Terminal" on repo "foo" at "/dev/foo"
    And SpawnTerminal succeeds with target "term-foo-123:0.0"
    Then addSession is called with repoPath "/dev/foo" and sessionName "term-foo-123"
    And navigation to AgentDetail shows the session in the tab bar immediately
```

## Tasks / Subtasks

- [ ] Task 1: Add tab bar to AgentDetail (AC: AC-1, AC-2, AC-5)
  - [ ] Import `repoSessions`, `refreshSessions` from sessions store
  - [ ] Call `refreshSessions(agent.repoPath)` in `onMount`
  - [ ] Derive `sessions` from `$repoSessions[agent?.repoPath]`
  - [ ] Add `activeSessionIdx` state with reactive default from `agent.tmuxTarget`
  - [ ] Render tab bar HTML with session tabs and active styling
  - [ ] Replace Terminal `paneTarget` with `{#key activeSession?.paneTarget}` block

- [ ] Task 2: Implement kill and spawn handlers (AC: AC-3, AC-4)
  - [ ] Import `KillTerminalSession`, `SpawnTerminal` from Wails bindings
  - [ ] Implement `killSession(sessionName)` with fallback to tab 0
  - [ ] Implement `spawnNewTerminal()` with auto-switch to new tab

- [ ] Task 3: Style tab bar (AC: AC-1)
  - [ ] Add CSS for `.session-tabs`, `.session-tab`, `.active`, `.tab-close`, `.add-tab`
  - [ ] Follow DESIGN.md variables (bg-deeper, border-accent, fg-primary, etc.)
  - [ ] Ensure horizontal scroll overflow for many tabs

- [ ] Task 4: Pre-populate store in NotificationFeed (AC: AC-6)
  - [ ] Import `addSession` from sessions store in NotificationFeed.svelte
  - [ ] Call `addSession` in `spawnTerminalInRepo()` after successful spawn
  - [ ] Call `addSession` in `onSessionSpawn()` after successful spawn

- [ ] Task 5: Manual testing (AC: AC-1 through AC-6)
  - [ ] Verify tab bar appears with sessions
  - [ ] Verify tab switch reconnects terminal
  - [ ] Verify kill removes tab
  - [ ] Verify "+" spawns and adds tab
  - [ ] Verify navigation from feed shows session immediately

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on new/modified files
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
