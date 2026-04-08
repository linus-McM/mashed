# Story 4: Svelte Sessions Store & Event Subscription

**Priority:** P1-high
**Domain:** frontend
**Estimated Complexity:** S
**Depends On:** sessions-01, sessions-03
**Status:** ready

## Description

Create a Svelte writable store (`sessions.js`) that caches terminal session data per repo, provides `refreshSessions()` to query the backend, and supports optimistic add/remove updates. Wire `App.svelte` to subscribe to `terminal:session:added` and `terminal:session:removed` Wails events so the store stays current without polling. This is the frontend foundation for the session tab bar in Story 5.

## Developer Notes

### Architecture

- **New file** `frontend/src/lib/stores/sessions.js` — joins existing stores `theme.js` and `font.js` in the same directory.
- **Modified file** `frontend/src/App.svelte` — add event subscriptions for `terminal:session:added` and `terminal:session:removed`.
- The store uses `writable({})` where the value is `{ [repoPath]: TerminalSession[] }`.
- `refreshSessions(repoPath)` calls the Wails-generated `ListRepoSessions` binding.
- `addSession(repoPath, session)` and `removeSession(repoPath, sessionName)` provide optimistic updates so the UI reacts immediately before the backend confirms.

### Store API

```js
// frontend/src/lib/stores/sessions.js
import { writable } from 'svelte/store';
import { ListRepoSessions } from '../../../wailsjs/go/main/App.js';

export const repoSessions = writable({});

export async function refreshSessions(repoPath) {
    const sessions = await ListRepoSessions(repoPath);
    repoSessions.update(cur => ({ ...cur, [repoPath]: sessions || [] }));
    return sessions || [];
}

export function addSession(repoPath, session) {
    repoSessions.update(cur => {
        const list = cur[repoPath] || [];
        // Avoid duplicates by sessionName
        if (list.some(s => s.sessionName === session.sessionName)) return cur;
        return { ...cur, [repoPath]: [...list, session] };
    });
}

export function removeSession(repoPath, sessionName) {
    repoSessions.update(cur => {
        const list = (cur[repoPath] || []).filter(s => s.sessionName !== sessionName);
        return { ...cur, [repoPath]: list };
    });
}
```

### App.svelte Event Wiring

Add in the `<script>` block after the existing `EventsOn('agent:removed', ...)`:

```js
import { refreshSessions, addSession, removeSession } from './lib/stores/sessions.js';

EventsOn('terminal:session:added', (session) => {
    addSession(session.repoPath, session);
});

EventsOn('terminal:session:removed', (sessionName) => {
    // Need to find which repoPath this session belongs to, so just refresh all known repos
    // or iterate the store. Simple approach: iterate store keys and remove.
    repoSessions.update(cur => {
        const updated = { ...cur };
        for (const rp of Object.keys(updated)) {
            updated[rp] = updated[rp].filter(s => s.sessionName !== sessionName);
        }
        return updated;
    });
});
```

### Technical Considerations

- **Wails binding import**: `ListRepoSessions` will be auto-generated at `frontend/wailsjs/go/main/App.js` after `wails generate` or `wails dev`. The engineer must run this after Story 1's backend changes are compiled.
- **Store reactivity**: Svelte's `$repoSessions` auto-subscription works in components. Derive per-repo sessions with `$: sessions = $repoSessions[repoPath] || []`.
- **No polling**: Events keep the store current. `refreshSessions` is called on-demand (e.g., on component mount) as a fallback.

### Risks & Edge Cases

- **Wails binding not generated yet**: If `ListRepoSessions` doesn't exist in the generated bindings, the import will fail at dev time. Ensure `wails dev` is restarted after backend changes.
- **Event before store subscription**: If a `terminal:session:added` event fires before `App.svelte` mounts, it will be missed. This is fine — `refreshSessions` on component mount catches up.
- **Stale data after app sleep**: If the user puts the machine to sleep, sessions may die without events. `refreshSessions` on view mount handles this.

### Reference Files

- `frontend/src/lib/stores/theme.js` — existing store pattern to follow
- `frontend/src/lib/stores/font.js` — another existing store
- `frontend/src/App.svelte` — where to add event subscriptions (lines 51-65)
- `frontend/wailsjs/go/main/App.js` — auto-generated bindings (after backend compile)

## Acceptance Criteria

AC-1: Sessions store exports correct API
- Given the store file exists at `frontend/src/lib/stores/sessions.js`
- When imported by a component
- Then `repoSessions`, `refreshSessions`, `addSession`, and `removeSession` are available

AC-2: refreshSessions populates store from backend
- Given the backend `ListRepoSessions` returns 2 sessions for "/dev/repo"
- When `refreshSessions("/dev/repo")` is called
- Then `$repoSessions["/dev/repo"]` contains 2 sessions
- And the function returns the sessions array

AC-3: addSession performs optimistic insert without duplicates
- Given `$repoSessions["/dev/repo"]` has 1 session
- When `addSession("/dev/repo", newSession)` is called
- Then `$repoSessions["/dev/repo"]` has 2 sessions
- And calling `addSession` again with the same sessionName does not create a duplicate

AC-4: removeSession removes by sessionName
- Given `$repoSessions["/dev/repo"]` has sessions ["a", "b"]
- When `removeSession("/dev/repo", "a")` is called
- Then `$repoSessions["/dev/repo"]` has only session "b"

AC-5: App.svelte subscribes to session events
- Given the app is running
- When a `terminal:session:added` event fires with session data
- Then `addSession` is called with the session's repoPath and data
- And when a `terminal:session:removed` event fires with a sessionName
- Then the session is removed from all repo entries in the store

## BDD Test Scenarios

### Scenario 1: Store Initialization

```gherkin
Feature: Sessions store

  Scenario: Store starts empty
    Given the sessions store is imported
    When no actions have been taken
    Then repoSessions value is an empty object {}
```

### Scenario 2: Refresh from Backend

```gherkin
Feature: refreshSessions

  Scenario: Fetch and populate sessions for a repo
    Given ListRepoSessions("/dev/repo") returns [{ sessionName: "term-repo-1", ... }, { sessionName: "term-repo-2", ... }]
    When refreshSessions("/dev/repo") is called
    Then repoSessions["/dev/repo"] has length 2
    And the returned value has length 2

  Scenario: Backend returns null/empty
    Given ListRepoSessions("/dev/empty") returns null
    When refreshSessions("/dev/empty") is called
    Then repoSessions["/dev/empty"] is an empty array
```

### Scenario 3: Optimistic Updates

```gherkin
Feature: Optimistic add/remove

  Scenario: addSession inserts a new session
    Given repoSessions["/dev/repo"] is empty
    When addSession("/dev/repo", { sessionName: "term-1", paneTarget: "term-1:0.0" }) is called
    Then repoSessions["/dev/repo"] has 1 entry with sessionName "term-1"

  Scenario: addSession rejects duplicate sessionName
    Given repoSessions["/dev/repo"] has session "term-1"
    When addSession("/dev/repo", { sessionName: "term-1", paneTarget: "term-1:0.0" }) is called
    Then repoSessions["/dev/repo"] still has exactly 1 entry

  Scenario: removeSession filters by name
    Given repoSessions["/dev/repo"] has sessions "term-1" and "term-2"
    When removeSession("/dev/repo", "term-1") is called
    Then repoSessions["/dev/repo"] has 1 entry with sessionName "term-2"
```

### Scenario 4: Event-Driven Updates in App.svelte

```gherkin
Feature: Wails event subscription

  Scenario: terminal:session:added updates store
    Given App.svelte is mounted
    When a "terminal:session:added" event fires with { sessionName: "term-foo-1", repoPath: "/dev/foo" }
    Then repoSessions["/dev/foo"] contains session "term-foo-1"

  Scenario: terminal:session:removed cleans all repos
    Given repoSessions has "/dev/foo" with session "term-foo-1"
    When a "terminal:session:removed" event fires with "term-foo-1"
    Then repoSessions["/dev/foo"] no longer contains "term-foo-1"
```

## Tasks / Subtasks

- [ ] Task 1: Create sessions store (AC: AC-1, AC-2, AC-3, AC-4)
  - [ ] Create `frontend/src/lib/stores/sessions.js`
  - [ ] Export `repoSessions` writable store
  - [ ] Implement `refreshSessions(repoPath)` calling `ListRepoSessions`
  - [ ] Implement `addSession(repoPath, session)` with duplicate guard
  - [ ] Implement `removeSession(repoPath, sessionName)`

- [ ] Task 2: Wire events in App.svelte (AC: AC-5)
  - [ ] Import `addSession`, `removeSession`, `repoSessions` from sessions store
  - [ ] Add `EventsOn('terminal:session:added', ...)` handler
  - [ ] Add `EventsOn('terminal:session:removed', ...)` handler

- [ ] Task 3: Verify Wails bindings (AC: AC-2)
  - [ ] Run `wails dev` or `wails generate` to confirm `ListRepoSessions` appears in `frontend/wailsjs/go/main/App.js`
  - [ ] Verify the generated TypeScript types match `TerminalSession`

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on new/modified files
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
