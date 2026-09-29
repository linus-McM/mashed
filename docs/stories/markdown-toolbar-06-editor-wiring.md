# Story 06: MarkdownEditor Wiring — featureConfigs + deferred re-init

**Priority:** P0-critical
**Domain:** frontend
**Estimated Complexity:** M
**Depends On:** Story 02 (store), Story 03 (builder)
**Status:** done
**UI-facing:** YES (behaviour-facing — no new visual design, but toolbar appearance changes based on settings)

## Description

Wire `MarkdownEditor.svelte` to read the store, pass `buildToolbar` into Crepe's constructor via `featureConfigs`, and re-init the editor only when the user closes Settings (dirty flag flips to false) AND the applied settings actually differ. The re-init branch MUST call `saver.flush()` before `destroyEditor()` to preserve unsaved edits — this is the same safety pattern used in the existing filePath-change branch.

## Developer Notes

### Architecture
- File to modify: `/Users/linus/Development/mashed/frontend/src/components/MarkdownEditor.svelte`
- New imports (add to the existing `<script>` block):

```ts
import { get } from 'svelte/store';
import { markdownMenuSettings, markdownMenuDirty } from '../lib/stores/markdownMenuSettings';
import { buildToolbarFromSettings } from './markdownToolbarBuilder';
```

- Inside `initEditor`, replace the current Crepe constructor:

```ts
crepe = new Crepe({
  root: container,
  defaultValue: markdown,
  featureConfigs: {
    [Crepe.Feature.Toolbar]: {
      buildToolbar: buildToolbarFromSettings(get(markdownMenuSettings)),
    },
  },
});
```

- Add a tracking variable and reactive re-init block at the bottom of `<script>`, AFTER the existing filePath-change block (don't interleave — the execution order of reactive blocks matters for diff clarity):

```ts
let lastAppliedSettings = JSON.stringify(get(markdownMenuSettings));

$: {
  if (!$markdownMenuDirty && crepe && !loading) {
    const current = JSON.stringify($markdownMenuSettings);
    if (current !== lastAppliedSettings) {
      lastAppliedSettings = current;
      saver.flush();
      destroyEditor().then(() => initEditor());
    }
  }
}
```

### The saver.flush() placement (resolved plan decision)
The plan explicitly notes: `MarkdownEditor.svelte` re-init branch must call `saver.flush()` before `destroyEditor()`, mirroring the existing filePath-change handler. Reload reads from disk, so unsaved in-memory edits would otherwise be lost. This is non-negotiable — the story fails review without it.

The existing filePath-change branch (already in the file) is:
```ts
$: if (filePath !== prevFilePath) {
  prevFilePath = filePath;
  saver.flush();
  destroyEditor().then(() => initEditor());
}
```
The new settings-change branch mirrors this pattern exactly.

### Reactivity semantics
The reactive block must read BOTH `$markdownMenuDirty` AND `$markdownMenuSettings` so Svelte subscribes to both. The guards (`!dirty && crepe && !loading`) keep it from running:
- **While user is typing in Settings** — dirty=true blocks.
- **During initial load** — crepe is null until `initEditor` completes.
- **During another re-init in flight** — loading=true blocks.

When `clearMarkdownMenuDirty()` fires (Settings back button), dirty flips to false, the block re-evaluates, the two JSON strings differ, and the editor re-mounts exactly once.

### Why JSON stringify
Settings is a small flat object with primitive values. `JSON.stringify` is the simplest stable comparator. Deep-equal libraries would work but are overkill.

### Behavior contract
- On initial mount: editor uses the current store value. No re-init fires because `lastAppliedSettings` was seeded with the same value.
- User opens Settings, toggles three items. Each `updateMarkdownMenuItem` sets dirty=true, which BLOCKS re-init. Store values update, but the editor behind Settings keeps its cursor and scroll.
- User clicks Back. `clearMarkdownMenuDirty` fires. Dirty flips to false. The reactive block evaluates; current JSON differs from lastApplied; `saver.flush()` runs (saves any in-flight typing), then destroy→init runs once.
- User returns to editor with new toolbar configuration.

### Risks & Edge Cases
- **saver.flush() is async** — it returns a Promise. Do we await it? Look at the filePath-change branch: it doesn't await. For consistency, this story shouldn't either. Flush schedules the save; destroy can proceed without waiting — WriteFile is just disk I/O and the old buffer is already captured.
  - Caveat: if WriteFile rejects, the user loses that save attempt. Same risk as the existing branch. Accept as-is for MVP.
- **Very rapid Settings open/close** — user toggles off Bold, clicks Back, immediately re-opens Settings, toggles Bold back on, clicks Back. lastAppliedSettings tracks correctly (two JSON diffs, two re-inits). Acceptable; re-init takes <100ms.
- **Setting change while editor is destroyed** — e.g., user navigates away from a .md file, then changes settings. `crepe` is null; guard prevents re-init. When user re-opens a .md file, `initEditor` uses the current store value. Correct.
- **Initial JSON seed uses get(), not a subscription** — the seed captures the store value at mount time. If that changes in the same tick the block re-runs, but the current != lastApplied check catches it.
- **featureConfigs shape** — Crepe's typing may require exact key names. `Crepe.Feature.Toolbar` is the enum; verify this import path in the existing Crepe integration. If the enum namespace differs, import directly: `import { Feature } from '@milkdown/crepe'` then use `[Feature.Toolbar]`.

### Reference Files
- `/Users/linus/Development/mashed/frontend/src/components/MarkdownEditor.svelte` — current file, see the filePath-change block as the pattern.
- `/Users/linus/Development/mashed/frontend/src/components/markdownEditorUtils.ts` — `createDebouncedSave` implementation (already imports `saver.flush()`).
- Milkdown Crepe docs on `featureConfigs` — verify the `Toolbar` key name.

## Acceptance Criteria

AC-1: Editor uses store values on initial mount
- Given the store has `{bold:true, italic:false, ...}` when the editor mounts
- When `initEditor` runs
- Then `buildToolbarFromSettings(store)` is invoked once during Crepe construction
- And the Crepe instance is created with `featureConfigs.Toolbar.buildToolbar` populated

AC-2: No re-init while dirty flag is true
- Given the editor is mounted
- When `markdownMenuSettings` changes while `markdownMenuDirty === true`
- Then `destroyEditor` is NOT called
- And the Crepe instance reference remains the same

AC-3: Re-init fires exactly once when dirty flips false with changed settings
- Given the editor is mounted and `lastAppliedSettings` matches the initial store
- And the user has toggled one setting (making current != lastApplied) while dirty=true
- When `clearMarkdownMenuDirty()` sets dirty=false
- Then `saver.flush()` is called once
- And `destroyEditor().then(initEditor)` is called once
- And `lastAppliedSettings` updates to the new current JSON

AC-4: No re-init when dirty flips false but settings didn't change
- Given dirty was true but the user reverted all toggles to their original state before clicking Back
- When dirty flips to false
- Then current JSON equals lastAppliedSettings
- And no re-init fires

AC-5: saver.flush() runs before destroyEditor
- Given a re-init is triggered
- When the reactive block executes
- Then `saver.flush` is invoked strictly before `destroyEditor`
- Verified by call-order assertion in the test

AC-6: Guards block re-init during loading
- Given `loading === true` (another init in flight)
- When a dirty→clean transition occurs with changed settings
- Then no re-init is triggered (the current init finishes first)

## BDD Test Scenarios

```gherkin
Feature: MarkdownEditor responds to toolbar settings

  Scenario: Initial mount wires featureConfigs.Toolbar.buildToolbar
    Given the markdownMenuSettings store has defaults
    When MarkdownEditor mounts
    Then Crepe constructor is invoked with featureConfigs containing a Toolbar.buildToolbar function
    And buildToolbarFromSettings was called once with the default store value

  Scenario: Dirty store blocks re-init
    Given the editor is mounted and rendering
    When markdownMenuSettings.bold flips to false and markdownMenuDirty is true
    Then destroyEditor is not called
    And the Crepe instance is unchanged

  Scenario: Clearing dirty with changed settings triggers one re-init
    Given the store has changed (bold:false) while dirty=true
    When clearMarkdownMenuDirty is called and dirty becomes false
    Then saver.flush is invoked
    And then destroyEditor is invoked
    And then initEditor is invoked with the new settings
    And lastAppliedSettings now matches the current store JSON

  Scenario: Reverted toggle — no re-init
    Given the editor is mounted with defaults
    And the user toggled bold off then back on while dirty=true
    When dirty becomes false
    Then no re-init fires
    And the Crepe instance reference is unchanged

  Scenario: saver.flush runs before destroyEditor
    Given a re-init is about to fire
    When the reactive block executes
    Then the call order is: saver.flush, then destroyEditor, then initEditor

  Scenario: Loading state blocks re-init
    Given initEditor is currently in flight (loading=true)
    When dirty goes false with changed settings
    Then destroyEditor is not called
    And the guard skips the re-init
```

## Tasks / Subtasks

- [ ] Task 1: Wire featureConfigs in initEditor (AC-1) — frontend
  - [ ] Add store + builder imports at top of `<script>`.
  - [ ] Replace the Crepe constructor call with the version including `featureConfigs.Toolbar.buildToolbar`.
  - [ ] Use `get(markdownMenuSettings)` to read the snapshot once per init.
- [ ] Task 2: Deferred re-init reactive block (AC-2, AC-3, AC-4, AC-6) — frontend
  - [ ] Declare `lastAppliedSettings` seeded from `get(markdownMenuSettings)`.
  - [ ] Add the reactive `$:` block with guards on dirty, crepe, loading.
  - [ ] Within the block: compare JSON, early-return if equal, else update lastApplied.
- [ ] Task 3: saver.flush before destroy (AC-5) — frontend
  - [ ] Inside the re-init branch, call `saver.flush()` before `destroyEditor()`.
  - [ ] Verify the existing filePath-change branch still reads the same pattern (no regression).
- [ ] Task 4: Tests (all ACs) — frontend
  - [ ] Create `frontend/src/components/__tests__/MarkdownEditor.test.ts` or extend the existing one (story corpus shows the file already exists at `frontend/src/components/__tests__/MarkdownEditor.test.ts`).
  - [ ] Mock Crepe: replace the dynamic import `loadCrepeModule` with a spy that returns a fake `Crepe` class. Capture constructor args to assert featureConfigs.
  - [ ] Test: dirty=true blocks re-init.
  - [ ] Test: dirty flip with changed settings triggers saver.flush then destroy then init, in order.
  - [ ] Test: dirty flip with unchanged settings (reverted) does NOT re-init.
  - [ ] Test: loading=true blocks re-init.

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ line coverage on the `<script>` additions (reactive block + store glue)
- [ ] `svelte-check` passes with 0 warnings, 0 errors
- [ ] `/simplify` run on MarkdownEditor.svelte
- [ ] Manual verification: open a .md file, toggle Bold off in Settings, close Settings, selection toolbar no longer shows Bold, cursor position and scroll preserved through the round-trip
- [ ] Code review: no CRITICAL/HIGH issues — especially around the saver.flush / destroy order

## Design Brief

### Intent
This story has no new visual components, but it owns the single most user-visible side-effect of the entire feature: **the moment the selection toolbar's shape changes**. The user's mental model is "I toggled Bold off, I closed Settings, now the toolbar doesn't have Bold." The deferred re-init is a behavioural design choice — it protects the typing session inside Settings — and the design brief's job is to define what "this feels right" means for that moment and enforce the invisible contract around it.

The signature UX property of this story is **deterministic, invisible, once**. The user should never see a toolbar flicker, never lose their cursor position mid-edit, and never notice the re-init as anything other than "the toolbar is just different now."

### 1. Layout composition (behavioural layout, not visual)
No markup changes to the editor surface. The "layout" here is the **temporal layout** of the transition:

```
  Settings open ─────────────────────────┐
       │                                 │
       │ user toggles Bold off          │  dirty = true
       │ user toggles Italic off        │  (re-init BLOCKED)
       │ user toggles LaTeX on          │
       │                                 │
  Settings close (Back/Esc) ─────────────┤
       │                                 │
       │ clearMarkdownMenuDirty()       │  dirty = false
       │ reactive block re-evaluates    │
       │ saver.flush()                   │  (unsaved edits written)
       │ destroyEditor()                 │  (no flicker: Settings still covers editor)
       │ initEditor()                    │  (new toolbar config)
       │                                 │
  Back dispatch completes ───────────────┘
       │
       │ Settings unmounts
       │ editor revealed WITH new toolbar already mounted
```

**Design invariant:** the destroy→init must complete *before or during* the Settings unmount, not after. If Settings closes first and the user sees the editor re-mount, that IS a flicker. This is why `clearMarkdownMenuDirty()` must fire *before* `dispatch('back')` in story 05 — so the reactive block runs while Settings still occludes the editor.

### 2. Typography plan
No typography changes. The selection toolbar uses Crepe's stock typography (which inherits `--font-ui` / `--font-mono` via CSS variables at the app root). Verify toolbar buttons still render in Geist after re-init.

### 3. Color strategy
No color changes. The selection toolbar's accent, hover, and active states are Crepe-internal and unaffected by this wiring. The test to pass: after re-init, toolbar button colors match pre-init exactly (no theme desync from unmounting).

### 4. Interaction model — the "feels right" moments
Five observable behaviours define acceptance:

1. **Toggle clicks inside Settings are instant.** `.setting-toggle` flips color in 100ms (existing transition). No "applying..." spinner. The dirty flag is invisible to the user.
2. **The editor behind Settings is frozen.** While dirty=true, any typing state, cursor position, scroll offset, and selection in the editor must remain exactly as it was when Settings opened. No Crepe re-render. No text reflow. If the user toggles 6 times, the editor DOM underneath Settings must be byte-identical to when they started.
3. **Back/Esc feels like closing a drawer, not reloading a page.** The transition must be: (a) Settings slides/fades away, (b) editor revealed with the new toolbar configuration already in place. The user should NOT see: old toolbar → flash → new toolbar.
4. **Selecting text after close shows the new toolbar shape.** First text selection post-close produces the new toolbar config. No second attempt needed.
5. **Cursor position preservation** — from the user's viewpoint, cursor was at line 10:5 before Settings opened, still at line 10:5 when Settings closed. The re-init technically resets cursor to document start (Crepe limitation), BUT because the re-init happens during Settings occlusion + before the next interaction, the user experiences it as "cursor is where I left it." If the user didn't click before opening Settings, this holds. If they did, the cursor resets — call this out as a known caveat but NOT a blocker.

**Transition durations:**
- Dirty flag flip: instant (synchronous store update).
- `saver.flush()` → `destroyEditor()`: no artificial delay, back-to-back in the same tick.
- `destroyEditor()`: Crepe native teardown time (~10–50ms).
- `initEditor()`: Crepe native mount time (~50–150ms).
- Total from Back-click to new-toolbar-ready: < `var(--duration-medium)` (150ms) target. Acceptable ceiling: 250ms. If it's longer, the user sees flicker.

### 5. Component specs
No CSS. The design contract lives in the reactive block and its guards:

```ts
// Seed at mount — must happen BEFORE any reactive block runs
let lastAppliedSettings = JSON.stringify(get(markdownMenuSettings));

// The single source of truth for "is a re-init warranted right now?"
$: {
  if (!$markdownMenuDirty && crepe && !loading) {
    const current = JSON.stringify($markdownMenuSettings);
    if (current !== lastAppliedSettings) {
      lastAppliedSettings = current;
      saver.flush();
      destroyEditor().then(() => initEditor());
    }
  }
}
```

**Invariants encoded above:**
- Re-init is *gated*, not triggered — reacts to dirty-clean transition, not to settings changes directly.
- Re-init is *idempotent per-change-set* — the JSON comparison prevents running on every reactive tick.
- Re-init is *atomic* — `saver.flush()` → `destroyEditor()` → `initEditor()` in one reactive evaluation.
- Re-init is *side-effect-safe* — `loading` guard prevents concurrent inits.

**Crepe constructor must be invoked exactly once per (file, settings) tuple:**
```ts
crepe = new Crepe({
  root: container,
  defaultValue: markdown,
  featureConfigs: {
    [Crepe.Feature.Toolbar]: {
      buildToolbar: buildToolbarFromSettings(get(markdownMenuSettings)),
    },
  },
});
```

`get()` (not subscription) is deliberate — we want a snapshot at init time, not a live-updating toolbar that rebuilds mid-keystroke.

### 6. Signature elements — what makes this "unmistakably Mashed" behaviourally
- **Deferred re-init is a Linear-grade detail.** Consumer apps would just re-render on every toggle; Mashed respects the user's in-progress typing session. This is the invisible polish that makes the app feel fast.
- **Exactly-once semantics.** Revert a toggle twice, click Back: zero re-inits. Toggle once, click Back: one re-init. This determinism is dev-tool discipline — surprising behaviour is a bug.
- **saver.flush() before destroy.** Never lose a keystroke. This is data-safety discipline, not aesthetics, but it's what makes the command-center feel trustworthy.
- **`clearMarkdownMenuDirty()` in the back handler** (story 05) is the *handoff*. This story's re-init block is the *receiver*. The two are designed as a pair — document the pair in code comments so future engineers don't break it.
- **No user-facing indicator of the re-init.** No "Settings applied" toast, no spinner, no color flash on the toolbar. Silent, instantaneous, deterministic. That silence IS the signature.

### 7. Anti-patterns to avoid
- **No toolbar flicker.** If `destroyEditor()` can complete before Settings unmounts, the user briefly sees a naked editor surface. Test: Settings must still occlude the editor when destroy+init runs. The `clearMarkdownMenuDirty()` before `dispatch('back')` ordering in story 05 enforces this.
- **No cursor jump mid-edit.** Never call `destroyEditor()` while `dirty === true`. Never call it during active typing. Guards enforce this.
- **No re-init on every toggle.** Six rapid clicks inside Settings must produce zero re-inits while dirty=true, then exactly one re-init when dirty clears. If you see six re-inits logged, the reactive block is missing the dirty guard.
- **No re-init when settings reverted.** If user toggles Bold off then Bold on and clicks Back, `current === lastAppliedSettings` must be true; no re-init. This is AC-4 and is design-critical — the user didn't change anything, so nothing should rebuild.
- **Do NOT `await saver.flush()` before destroy.** The existing `filePath-change` branch doesn't await; this branch must mirror exactly. Awaiting adds async delay that moves re-init past Settings unmount → flicker.
- **Do NOT subscribe to settings with a live `$:` that rebuilds toolbar.** The builder runs inside Crepe's constructor, not as a reactive output. A live-rebuild would flicker the toolbar per keystroke in Settings.
- **Do NOT add a user-visible "reloading editor" state.** Silence is the design. If re-init is slow enough to need a spinner, the bug is the slowness, not the missing spinner.
- **Do NOT put the reactive block BEFORE the filePath-change block.** Order matters for predictability — filePath changes are orthogonal and higher-priority.
- **Do NOT mutate `lastAppliedSettings` outside the reactive block.** That variable is the re-init's memory; external writes break AC-4.
- **Do NOT emit a toast, notification, or log line visible in the UI on re-init.** The design is silent.
