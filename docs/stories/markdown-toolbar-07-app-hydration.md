# Story 07: App Hydration — initMarkdownMenuSettings on mount

**Priority:** P0-critical
**Domain:** frontend
**Estimated Complexity:** S
**Depends On:** Story 01 (Wails `GetConfig` returns `markdownMenu`), Story 02 (`initMarkdownMenuSettings` exists)
**Status:** done
**UI-facing:** no

## Description

Initialise the `markdownMenuSettings` store from the persisted config during `App.svelte`'s `onMount`, alongside the existing theme/font/editorSettings hydration. Without this, the store starts at frontend DEFAULTS and the user's persisted choices don't surface until the next Settings interaction. This story completes the persistence loop.

## Developer Notes

### Architecture
- File to modify: `/Users/linus/Development/mashed/frontend/src/App.svelte`
- Existing pattern (already in the file, lines 26442 and 26482 of the corpus):
  ```ts
  import { initEditorSettings } from './lib/stores/editorSettings.js';
  // ... later in onMount ...
  try { const es = await GetEditorSettings(); initEditorSettings(es); } catch {}
  ```
- Add analogous import and init call:
  ```ts
  import { initMarkdownMenuSettings } from './lib/stores/markdownMenuSettings';
  // ... inside onMount, alongside the editor settings line ...
  initMarkdownMenuSettings(cfg.markdownMenu);
  ```

### Where to place the call
Inside the existing `onMount` async function, the code already calls `const cfg = await GetConfig();` around line 26467. Add `initMarkdownMenuSettings(cfg.markdownMenu)` near that line (i.e., use the already-fetched `cfg` rather than making a second Wails call). Wrapping in its own try/catch is preferred so a hydration failure doesn't break other init steps:

```ts
try { initMarkdownMenuSettings(cfg.markdownMenu); } catch {}
```

### Why cfg.markdownMenu and not GetMarkdownMenuSettings()
Two reasons:
1. `cfg` is already in hand from the existing `await GetConfig()` call — no extra Wails round-trip.
2. `cfg.markdownMenu` is `undefined` for pre-migration users; `initMarkdownMenuSettings` handles undefined/null gracefully (Story 02 AC-3).

If it turns out `cfg.markdownMenu` serialises inconsistently across Wails versions (e.g., `null` vs omitted), fall back to `GetMarkdownMenuSettings()` which is guaranteed to return a fully-populated struct. Not preferred but a valid fallback.

### Behavior contract
- On first launch after upgrade (no `markdownMenu` in config): store starts at DEFAULTS.
- On subsequent launches: store hydrates to the persisted values.
- If `GetConfig` throws: hydration is skipped; store retains DEFAULTS. Not a blocker for the rest of the app.

### Risks & Edge Cases
- **Wails generated models treat pointer fields as optional** — `cfg.markdownMenu` will be typed as `main.MarkdownMenuSettings | undefined`. Passing undefined to `initMarkdownMenuSettings` is explicitly supported.
- **Race with GetEditorSettings call** — the existing line does a SECOND `await` for `GetEditorSettings`. The two inits don't interact; do this story's init synchronously (no await) immediately after the already-awaited `GetConfig`.
- **Do not await unnecessarily** — `initMarkdownMenuSettings` is synchronous. Don't wrap it in `await`.

### Reference Files
- `/Users/linus/Development/mashed/frontend/src/App.svelte` — look at the existing `onMount` block, specifically the lines around the first `await GetConfig()` call.
- `/Users/linus/Development/mashed/frontend/src/lib/stores/markdownMenuSettings.ts` — import source.

## Acceptance Criteria

AC-1: Store hydrates from persisted config
- Given a config file containing `markdownMenu: {bold:false, latex:true, ...}`
- When `App.svelte` mounts and calls `initMarkdownMenuSettings(cfg.markdownMenu)`
- Then the `markdownMenuSettings` store reads `bold:false, latex:true, ...`

AC-2: Missing markdownMenu falls back to defaults
- Given a config file WITHOUT a `markdownMenu` key
- When `App.svelte` mounts
- Then `initMarkdownMenuSettings(undefined)` is called
- And the store reads DEFAULTS

AC-3: Hydration failure does not break other init steps
- Given `GetConfig` throws
- When `App.svelte` mounts
- Then `initMarkdownMenuSettings` is skipped gracefully
- And the rest of `onMount` (themes, fonts, editor settings) continues to run

AC-4: No extra Wails round-trip
- Given the existing `await GetConfig()` call
- When this story's code is added
- Then there is no additional call to `GetMarkdownMenuSettings` in `App.svelte`
- And the value is read from `cfg.markdownMenu` directly

AC-5: Init happens before any MarkdownEditor mounts
- Given `App.svelte`'s onMount sequence
- When the app transitions from 'loading' → 'feed'/'detail'
- Then `initMarkdownMenuSettings` has already completed
- And any subsequent `MarkdownEditor` mount reads correct store values via `get(markdownMenuSettings)`

## BDD Test Scenarios

```gherkin
Feature: App hydrates markdown menu settings on mount

  Scenario: Persisted settings are loaded
    Given the config file contains markdownMenu: {bold:false, italic:true, strikethrough:true, code:true, link:true, latex:true}
    When App.svelte mounts
    Then the markdownMenuSettings store reads those exact values

  Scenario: Missing markdownMenu key uses defaults
    Given the config file has no markdownMenu key
    When App.svelte mounts
    Then initMarkdownMenuSettings is called with undefined
    And the store reads defaults (bold/italic/strikethrough/code/link true, latex false)

  Scenario: GetConfig rejects — init is skipped
    Given GetConfig throws an error
    When App.svelte mounts
    Then initMarkdownMenuSettings is not called (or its try/catch swallows the error)
    And other init steps still execute

  Scenario: No duplicate Wails call
    Given App.svelte's onMount is running
    When the code path for markdown menu hydration executes
    Then GetMarkdownMenuSettings is NOT invoked
    And cfg.markdownMenu is used instead

  Scenario: Subsequent MarkdownEditor mount reads hydrated values
    Given App.svelte hydrated markdownMenuSettings with bold:false
    When the user opens a .md file and MarkdownEditor mounts
    Then buildToolbarFromSettings(get(markdownMenuSettings)) sees bold:false
```

## Tasks / Subtasks

- [ ] Task 1: Import + init (AC-1, AC-2, AC-4) — frontend
  - [ ] Add `import { initMarkdownMenuSettings } from './lib/stores/markdownMenuSettings';` at the top of the `<script>` in App.svelte.
  - [ ] Inside `onMount`, after the existing `const cfg = await GetConfig();`, add a try/catch block calling `initMarkdownMenuSettings(cfg.markdownMenu)`.
- [ ] Task 2: Integration test (AC-5) — frontend
  - [ ] Extend App.svelte unit/integration tests (or create a focused test) to:
    - mock `GetConfig` to return a config with `markdownMenu` present
    - mount App.svelte
    - assert the store reads the mocked values
- [ ] Task 3: Regression test for missing key (AC-2, AC-3) — frontend
  - [ ] Mock `GetConfig` to return a config without `markdownMenu`.
  - [ ] Assert store reads defaults after mount.
  - [ ] Mock `GetConfig` to throw; assert App.svelte still renders (no uncaught exception).

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] `svelte-check` passes with 0 warnings, 0 errors
- [ ] `/simplify` run on App.svelte
- [ ] Manual verification: set a non-default value (e.g., toggle Bold off), quit app, relaunch — Settings panel reflects Bold off, editor toolbar reflects Bold hidden
- [ ] Code review: no CRITICAL/HIGH issues
