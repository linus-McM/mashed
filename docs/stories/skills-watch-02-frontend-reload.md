# skills-watch-02: Frontend reactive reload on `bmad:assets:changed`

**Status:** ready
**Domain:** frontend
**Size:** S
**Depends on:** skills-watch-01
**Phase:** 4b

## Description

Subscribe to the `bmad:assets:changed` Wails event in the frontend and re-fetch `ListAllMashedAssets` to refresh `groupedMashedAssets`. The sidebar updates reactively without a manual reload. This is the frontend half of Phase 4b.

**This story is UI-facing — the refresh animation/indicator needs ui-architect review if one is added.**

## Developer Notes

- **Files to create/modify:**
  - `frontend/src/views/WorkflowBuilder.svelte` (or wherever `groupedMashedAssets` store is owned):
    - Subscribe to `EventsOn('bmad:assets:changed', handler)` in `onMount`
    - Handler re-invokes `ListAllMashedAssets` and updates the store
    - `EventsOff('bmad:assets:changed')` in `onDestroy`
  - Optional: `frontend/src/lib/stores/mashedAssets.ts` — if the logic is duplicated across components, lift the fetch+store into a dedicated Svelte store file.
- **Behavioural requirements:**
  - Debounce on the frontend is NOT needed — the backend already debounces 500 ms.
  - If `ListAllMashedAssets` fails, log to console and keep the previous store value (don't clear the sidebar on transient errors).
  - Rapid successive events (if debounce slips) should be handled — concurrent fetches must not race. Use a simple "in-flight" guard: if a fetch is already pending, mark "dirty" and re-fetch when it completes.
- **Optional UX indicator (behind ui-architect review):**
  - Show a small pulse/badge near the sidebar title during the re-fetch so the user knows the list is updating.
- **Risks / gotchas:**
  - Wails `EventsOn` must be unsubscribed in `onDestroy` or the handler fires on the next mount too.
  - A user editing a file in a rapid-save editor shouldn't cause the sidebar to flicker. The 500ms backend debounce plus the in-flight guard here should handle that.
  - Preserve scroll position across reloads — if the user has scrolled the sidebar, the new list should keep the same scroll offset when possible.
- **Prerequisites already in place:**
  - skills-watch-01 ships the backend event `bmad:assets:changed`.
  - `ListAllMashedAssets` binding exists from Phase 1.
  - Svelte store pattern already established in the frontend.

## Acceptance Criteria

**AC-1: Backend event triggers a refetch**
- Given the frontend is mounted and listening
- When the backend emits `bmad:assets:changed`
- Then `ListAllMashedAssets` is invoked
- And the resulting list replaces the current store value

**AC-2: Sidebar re-renders with the new list**
- Given the sidebar shows 3 commands
- When a new command file is created on disk and the event fires
- Then the sidebar renders 4 commands after the refetch completes

**AC-3: Transient fetch failure leaves old list in place**
- Given the backend event fires
- When `ListAllMashedAssets` rejects with an error
- Then the store value is unchanged (still shows the previous list)
- And a console error is logged

**AC-4: Concurrent events are coalesced**
- Given two events fire within 50 ms of each other while a previous fetch is still in flight
- When the in-flight fetch completes
- Then exactly one additional fetch runs
- And the sidebar reflects the final state (not an intermediate one)

**AC-5: Event listener is unsubscribed on component destroy**
- Given the component is mounted and subscribed
- When the component is destroyed and remounted
- Then only one subscription is active
- And events fire the handler exactly once per event

## BDD Test Scenarios (Gherkin)

```gherkin
Feature: Reactive sidebar reload

  Scenario: Backend event triggers refetch
    Given the sidebar is mounted
    When bmad:assets:changed is emitted
    Then ListAllMashedAssets is invoked
    And the store updates with the new list

  Scenario: New file appears after save
    Given the sidebar shows 3 commands
    When a new command file is created on disk
    And the debounced event fires
    Then the sidebar shows 4 commands

  Scenario: Fetch error preserves previous state
    Given the sidebar shows the current list
    When the refetch rejects
    Then the sidebar still shows the previous list
    And an error is logged to the console

  Scenario: Two rapid events trigger one extra fetch
    Given a refetch is in flight
    When two more events fire
    Then exactly one additional fetch runs after the in-flight one completes

  Scenario: Remount does not double-subscribe
    Given the component is unmounted and remounted
    When an event fires
    Then the handler runs exactly once
```

## Tasks / Subtasks

- [ ] Task 1 — Subscribe/unsubscribe lifecycle (AC-1, AC-5)
  - [ ] `EventsOn('bmad:assets:changed', handler)` in `onMount`
  - [ ] `EventsOff('bmad:assets:changed')` in `onDestroy`
- [ ] Task 2 — Refetch with in-flight guard (AC-1, AC-2, AC-4)
  - [ ] Boolean `fetching` flag
  - [ ] Dirty flag for events that arrive during a fetch
  - [ ] Re-run after completion if dirty
- [ ] Task 3 — Error handling (AC-3)
  - [ ] Console error on failure
  - [ ] Preserve previous store value
- [ ] Task 4 — Tests (AC-1 through AC-5)
  - [ ] Playwright: simulate a file change (via skills-watch-01 backend test harness) and assert sidebar updates
  - [ ] Unit test: in-flight guard coalesces rapid events
  - [ ] Unit test: error path preserves previous value
  - [ ] Unit test: unsubscribe on destroy

## Definition of Done

- [ ] All ACs verified by an automated test
- [ ] Coverage ≥ 80% on modified files
- [ ] `go build ./... && go vet ./...` clean
- [ ] `go test ./... -race -short` clean
- [ ] `/simplify` run before sign-off
- [ ] No hardcoded paths, magic numbers, or color literals added
- [ ] Existing tests still pass

## Design Brief

### Layout composition
Minimal layout impact — this is a lifecycle story, not a new surface. Three tiny visual slots added to `ProcessSidebar.svelte`:

1. **Sync indicator** next to the active "Skills" tab label when a refetch is in flight — a 6px pulsing dot in the tab strip, aligned right of the label. Slot: `display: inline-flex; align-items: center; gap: var(--sp-xs);` inside the `.tab.active` label.
2. **Row update flash** on any asset row whose identity (path) changed in the new list — reuses the `row-flash` keyframe pattern established in skills-editor-02 for commit acknowledgement.
3. **Empty state replacement** — if the refetch returns zero assets (e.g. `.claude/skills/` deleted on disk), the existing `.empty-state` div is rendered with the same copy it already uses. No new empty state needed.

### Typography plan
No new typography introduced. Reuses:
- **Tab label** (`.tab` — `font-family: var(--font-mono)`, `font-size: var(--text-body)`, `font-weight: 500`; active state `font-weight: 600`).
- **Empty state** (existing `.empty-state` — inherits `font-family: var(--font-mono)`, `font-size: 11px`, `color: var(--text-muted)`).

### Color strategy
- **Sync dot**: `var(--accent-green)` with `box-shadow: var(--glow-spread) color-mix(in srgb, var(--accent-green) 80%, transparent);` — identical to `.status-pulse` in `style.css`. Reusing the global class is the right move.
- **Row flash on update**: `color-mix(in srgb, var(--accent-green) 15%, transparent)` → `transparent` over 400ms. Same keyframe as skills-editor-02's `row-flash`.
- **Row flash on create** (new asset appears): `color-mix(in srgb, var(--accent-blue) 15%, transparent)` → `transparent`. Blue because create is informational, not a user-initiated commit; the color carries semantic weight (DESIGN.md §Color: blue = info).
- **Row flash on delete** (not rendered — the row simply unmounts, no flash needed). The disappearance is the feedback.
- **Error state** (refetch failed): tab label turns `var(--accent-amber)` for 2 seconds, then reverts. No banner, no toast — this is a transient backend hiccup and the DESIGN.md principle is minimal-functional.

### Interaction model
- **Passive feedback only** — user does not click anything for this flow. The sync dot is ambient, the row flash is transient.
- **Sync dot lifecycle**: appears when `fetching === true`, disappears within `var(--duration-medium)` (150ms fade) after fetch completes. If the fetch is very fast (<50ms), delay showing the dot by 50ms to avoid flicker.
- **Row diff detection**: compare `prevList` vs `newList` by `path`. New paths get `just-created` class for 400ms; changed-same-path rows (e.g. mashedRole changed) get `just-saved` class for 400ms. Removed rows unmount.
- **Scroll preservation**: capture `scrollTop` on the `.tab-content` element before refetch; restore after refetch completes and list has rendered (via `tick()` then assignment).
- **Reduced motion**: the flash animations are gated under `@media (prefers-reduced-motion: no-preference)`. With reduced motion, rows appear/update without flash — diff still works, just without the color cue.
- **No focus changes** — this is a background refresh and must not steal focus from any interactive element.

### Component specs
```
.tab.active .sync-indicator {
  display: inline-block;
  width: 6px;
  height: 6px;
  margin-left: var(--sp-xs);
  border-radius: 50%;
  background: var(--accent-green);
  box-shadow: var(--glow-spread) color-mix(in srgb, var(--accent-green) 80%, transparent);
  vertical-align: middle;
}
.process-item.just-created {
  animation: row-create-flash 400ms var(--ease-enter);
}
@keyframes row-create-flash {
  0%   { background: color-mix(in srgb, var(--accent-blue) 15%, transparent); }
  100% { background: transparent; }
}
@media (prefers-reduced-motion: no-preference) {
  .tab.active .sync-indicator {
    animation: status-pulse 1500ms var(--ease-move) infinite;
  }
}
.tab.error-state {
  color: var(--accent-amber);
  transition: color var(--duration-medium) var(--ease-enter);
}
```
The `status-pulse` keyframe already exists in `style.css` — reuse it verbatim. `row-flash` from skills-editor-02 is also reused for the update path.

### Signature elements
- **Ambient sync dot** in the tab strip (vs a spinner overlay or loading bar) respects the "calm command center" mood from `DESIGN.md`. The app never looks busy; it looks aware.
- **Semantic flash colors** — green = committed by user, blue = came from elsewhere, amber = something went wrong. This three-color language scales beyond this story: it's a cross-app pattern for "something in the UI changed, and here's why".
- **Non-harmonic pulse cadence** (1500ms sync dot vs 1000ms terminal cursor vs 2000ms agent-row) — the existing `style.css` comment explicitly calls this out, and reusing `status-pulse` inherits that discipline.

### Distinguishing from existing patterns
- **vs skills-editor-02 save flow**: save flow is a modal → toast → row-flash-green sequence (user-initiated). This story is a silent backend event → sync-dot → row-flash-blue sequence (externally-initiated). The color split (green vs blue) keeps them distinct even when the underlying keyframe is identical.
- **vs `.status-pulse` usage on the agent status bar**: the agent pulse says "something is running forever"; the sync dot says "something is fetching briefly". Both use the same visual token — that's correct; cadence and duration differentiate them (persistent vs ~300ms burst).
- **Badge/toast rule** (cross-story): skills-validate-01 will introduce warning badges. Those are PERSISTENT and indicate asset quality. The sync dot is TRANSIENT and indicates data freshness. Never put both at the same DOM slot; the sync dot lives in the tab strip, validation badges live inside the row.
