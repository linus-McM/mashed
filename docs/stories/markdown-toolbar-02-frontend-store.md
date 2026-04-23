# Story 02: Frontend Store — markdownMenuSettings

**Priority:** P0-critical
**Domain:** frontend
**Estimated Complexity:** S
**Depends On:** Story 01 (Wails bindings must exist)
**Status:** ready
**UI-facing:** no

## Description

Create a Svelte store that mirrors the backend `MarkdownMenuSettings` plus a dirty flag. The store is the single source of truth that Settings toggles, the toolbar builder reads, and the `MarkdownEditor` watches to decide when to re-init. The dirty flag is what lets the editor defer re-init until the user closes Settings.

## Developer Notes

### Architecture
- New file: `/Users/linus/Development/mashed/frontend/src/lib/stores/markdownMenuSettings.ts`
- Mirror the shape of `/Users/linus/Development/mashed/frontend/src/lib/stores/editorSettings.js` (pattern reference). Use TypeScript (not .js) — the plan specifies `.ts`.
- Exported API (exact names):

```ts
import { writable, get } from 'svelte/store';
import { SetMarkdownMenuSettings } from '../../../wailsjs/go/main/App.js';
import type { main } from '../../../wailsjs/go/models';

export type MarkdownMenuSettings = main.MarkdownMenuSettings;

const DEFAULTS: MarkdownMenuSettings = {
  bold: true, italic: true, strikethrough: true,
  code: true, link: true, latex: false,
};

export const markdownMenuSettings = writable<MarkdownMenuSettings>({ ...DEFAULTS });
export const markdownMenuDirty = writable<boolean>(false);

export function initMarkdownMenuSettings(s: Partial<MarkdownMenuSettings> | null): void;
export async function updateMarkdownMenuItem<K extends keyof MarkdownMenuSettings>(key: K, value: MarkdownMenuSettings[K]): Promise<void>;
export function clearMarkdownMenuDirty(): void;
```

### Behavior contract
- `initMarkdownMenuSettings(null)` — set store to DEFAULTS (no fields overridden).
- `initMarkdownMenuSettings(partial)` — merge partial over DEFAULTS so missing keys use defaults. This matches `editorSettings.initEditorSettings`.
- `updateMarkdownMenuItem(key, value)` — update store, set dirty=true, then `await SetMarkdownMenuSettings(get(markdownMenuSettings))`. Order matters: store update before the Wails call so synchronous subscribers (the Settings panel) see the new value immediately; dirty flag flips before the `await` so the `MarkdownEditor.svelte` re-init guard is tripped before the persistence round-trip completes.
- `clearMarkdownMenuDirty()` — just `markdownMenuDirty.set(false)`. Does NOT persist; persistence already happened in `updateMarkdownMenuItem`.

### Defaults authority
Frontend DEFAULTS exist as a safety net (if the Wails call fails during hydration). They MUST match the Go `DefaultMarkdownMenuSettings` values exactly. If they drift, Story 01 is authoritative — update this file.

### Technical Considerations
- TypeScript strict mode — the repo uses svelte-check strict; the generic `<K extends keyof MarkdownMenuSettings>` must type-check cleanly. `value: MarkdownMenuSettings[K]` is necessary.
- Do not re-export `main.MarkdownMenuSettings` as a type without care — the type re-export keeps consumers from having to import Wails-generated paths in every component.

### Risks & Edge Cases
- **Partial init from GetConfig** — `cfg.markdownMenu` may be undefined (pre-migration users). `initMarkdownMenuSettings(cfg.markdownMenu)` must NOT throw on null/undefined input and MUST fall back to DEFAULTS. Mirror `initEditorSettings` exactly here.
- **Rapid toggle double-click** — two `updateMarkdownMenuItem` calls in 20ms. Both issue `SetMarkdownMenuSettings`. Last-writer-wins is acceptable (backend `saveConfig` serialises).
- **Wails call fails mid-update** — if `SetMarkdownMenuSettings` rejects, the store value is already updated locally. This is acceptable for MVP — user sees the UI reflect the toggle, disk persistence fails silently. Log via `console.error` but do not revert the store. Re-opening the app will re-read from disk, reverting the failed change. Document this tradeoff in a code comment.

### Reference Files
- `/Users/linus/Development/mashed/frontend/src/lib/stores/editorSettings.js` — copy pattern: init with merge, update with persist, sync with writable store.
- `/Users/linus/Development/mashed/frontend/src/lib/bmadSessionName.ts` — TypeScript module style already in the repo.

## Acceptance Criteria

AC-1: Store defaults match backend
- Given the module is imported fresh
- When I subscribe to `markdownMenuSettings` without calling init
- Then the initial value is `{bold:true, italic:true, strikethrough:true, code:true, link:true, latex:false}`

AC-2: init merges partial over defaults
- Given `initMarkdownMenuSettings({bold: false, latex: true})` is called
- When I read the store via `get`
- Then the value is `{bold:false, italic:true, strikethrough:true, code:true, link:true, latex:true}`

AC-3: init(null) resets to defaults
- Given the store was previously modified
- When `initMarkdownMenuSettings(null)` is called
- Then the store equals DEFAULTS

AC-4: updateMarkdownMenuItem flips dirty and persists
- Given a mocked `SetMarkdownMenuSettings`
- When `updateMarkdownMenuItem('bold', false)` is awaited
- Then the store's `bold` field is `false`
- And `markdownMenuDirty` is `true`
- And `SetMarkdownMenuSettings` was called once with the full current store value

AC-5: clearMarkdownMenuDirty resets the flag only
- Given `markdownMenuDirty` is `true` and the store has custom values
- When `clearMarkdownMenuDirty()` is called
- Then `markdownMenuDirty` is `false`
- And the settings store value is unchanged
- And `SetMarkdownMenuSettings` was not called

AC-6: update dirty flag flips BEFORE the await resolves
- Given `SetMarkdownMenuSettings` is mocked to resolve after a tick
- When `updateMarkdownMenuItem('italic', false)` is invoked (not awaited)
- Then synchronously after invocation `markdownMenuDirty` reads `true`

## BDD Test Scenarios

```gherkin
Feature: markdownMenuSettings store

  Scenario: Module default values
    Given the store module is freshly imported
    When I read markdownMenuSettings synchronously
    Then bold, italic, strikethrough, code, link are true
    And latex is false

  Scenario: initMarkdownMenuSettings merges partial over defaults
    Given initMarkdownMenuSettings is called with {bold: false}
    When the store is read
    Then bold is false and italic, strikethrough, code, link are true and latex is false

  Scenario: initMarkdownMenuSettings(null) resets to defaults
    Given the store was previously set to all-false values
    When initMarkdownMenuSettings(null) is called
    Then the store equals defaults

  Scenario: updateMarkdownMenuItem persists and flips dirty
    Given SetMarkdownMenuSettings is mocked to resolve
    When updateMarkdownMenuItem('link', false) is awaited
    Then the store's link field is false
    And markdownMenuDirty is true
    And SetMarkdownMenuSettings was called with the full settings object containing link=false

  Scenario: clearMarkdownMenuDirty resets only the flag
    Given markdownMenuDirty is true and the store has latex=true
    When clearMarkdownMenuDirty is called
    Then markdownMenuDirty is false
    And latex is still true
    And SetMarkdownMenuSettings was not invoked

  Scenario: Dirty flag flips synchronously
    Given SetMarkdownMenuSettings is mocked with a deferred promise
    When updateMarkdownMenuItem('bold', false) is invoked without awaiting
    Then markdownMenuDirty reads true before the next microtask
```

## Tasks / Subtasks

- [ ] Task 1: Create store module (AC-1, AC-2, AC-3) — frontend
  - [ ] Create `frontend/src/lib/stores/markdownMenuSettings.ts`.
  - [ ] Declare DEFAULTS constant and writable stores.
  - [ ] Implement `initMarkdownMenuSettings` with merge-over-defaults logic.
  - [ ] Re-export `MarkdownMenuSettings` type from `wailsjs/go/models`.
- [ ] Task 2: Implement update and clear (AC-4, AC-5, AC-6) — frontend
  - [ ] Implement `updateMarkdownMenuItem` — update store, set dirty, await Wails call.
  - [ ] Implement `clearMarkdownMenuDirty`.
  - [ ] Add code comment documenting the "store updates before Wails rejection" tradeoff.
- [ ] Task 3: Unit tests (all ACs) — frontend
  - [ ] Create `frontend/src/lib/stores/__tests__/markdownMenuSettings.test.ts` (or next to the source per repo convention).
  - [ ] Mock `SetMarkdownMenuSettings` via `vi.mock('../../../wailsjs/go/main/App.js', ...)`.
  - [ ] Cover all six BDD scenarios above.

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ line coverage on `markdownMenuSettings.ts`
- [ ] `svelte-check` passes with 0 warnings, 0 errors
- [ ] Vitest unit tests pass
- [ ] `/simplify` run on the new file
- [ ] Code review: no CRITICAL/HIGH issues
