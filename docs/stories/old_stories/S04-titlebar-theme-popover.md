# Story 4: TitleBar Icons + Theme Popover

**Priority:** P1-high
**Domain:** frontend
**Estimated Complexity:** M
**Depends On:** Story 1, Story 2, Story 3
**Status:** ready

## Description

Add gear (Settings) and palette (Theme) icon buttons to the right side of the existing custom TitleBar. The palette button opens a small popover showing three theme cards with color swatches; clicking a card applies the theme instantly and persists the choice via the backend. The gear button dispatches an event to open the Settings view (handled in Story 5). This gives users a quick, always-accessible way to switch themes without leaving their current view.

## Developer Notes

### Architecture

- **Modified file: `frontend/src/components/TitleBar.svelte`** -- Add two icon buttons (gear and palette) to the right side of the titlebar, before the drag region ends. Add a ThemePopover sub-component or inline popover markup.
- **Modified file: `frontend/src/App.svelte`** -- Listen for `open-settings` event dispatched by TitleBar (handled in Story 5, but the event dispatch is wired here).

### TitleBar.svelte Layout Changes

Current structure (from reading the file):
```html
<div class="titlebar">
  <div class="traffic-lights">...</div>
  <div class="drag-region"></div>
</div>
```

New structure:
```html
<div class="titlebar">
  <div class="traffic-lights">...</div>
  <div class="drag-region" on:dblclick={WindowToggleMaximise}></div>
  <div class="titlebar-actions">
    <button class="action-btn" on:click={toggleThemePopover} title="Theme">
      <Palette size={14} />
    </button>
    <button class="action-btn" on:click={() => dispatch('open-settings')} title="Settings">
      <Settings size={14} />
    </button>
  </div>
</div>
```

The `titlebar-actions` div sits after the drag region, right-aligned, with `--wails-draggable: none` so clicks register.

### Icon Library

The project already uses `lucide-svelte` (see `App.svelte` line 9 importing `Hexagon`, and `AgentDetail.svelte` line 5 importing multiple icons). Use:
```js
import { Palette, Settings } from 'lucide-svelte';
```

### Theme Popover

The popover appears below the palette button, anchored to its position. It shows three theme cards, each with:
- Theme label (e.g., "Dark", "Light", "Midnight")
- 4-5 small color swatches showing key colors from that theme
- A checkmark or highlight on the currently active theme

Implementation approach:
```js
import { createEventDispatcher } from 'svelte';
import { Palette, Settings } from 'lucide-svelte';
import { themes } from '../lib/themes.js';
import { currentThemeId } from '../lib/stores/theme.js';
import { applyTheme } from '../lib/stores/theme.js';
import { SetTheme } from '../../wailsjs/go/main/App.js';

const dispatch = createEventDispatcher();

let showPopover = false;

function toggleThemePopover() {
  showPopover = !showPopover;
}

async function selectTheme(id) {
  applyTheme(id);
  showPopover = false;
  try {
    await SetTheme(id);
  } catch (e) {
    console.error('Failed to persist theme:', e);
  }
}

function handleClickOutside(e) {
  if (showPopover && !e.target.closest('.theme-popover') && !e.target.closest('.action-btn')) {
    showPopover = false;
  }
}
```

The popover HTML:
```html
{#if showPopover}
  <div class="theme-popover">
    {#each Object.entries(themes) as [id, theme]}
      <button
        class="theme-card"
        class:active={$currentThemeId === id}
        on:click={() => selectTheme(id)}
      >
        <span class="theme-label">{theme.label}</span>
        <div class="swatches">
          <span class="swatch" style="background: {theme.css['--bg-deepest']}"></span>
          <span class="swatch" style="background: {theme.css['--accent-green']}"></span>
          <span class="swatch" style="background: {theme.css['--accent-blue']}"></span>
          <span class="swatch" style="background: {theme.css['--text-primary']}"></span>
        </div>
      </button>
    {/each}
  </div>
{/if}
```

### Styling Guidelines

Follow existing design system patterns from `style.css`:
- Popover background: `var(--bg-elevated)`
- Popover border: `1px solid var(--border-subtle)`
- Border radius: `var(--radius-md)` (4px)
- Font: `var(--font-ui)` at `var(--text-label)` (11px) for labels
- Active indicator: `var(--accent-green)` border or checkmark
- Spacing: use `var(--sp-sm)` (8px) and `var(--sp-xs)` (4px)
- Animation: `var(--duration-short)` (100ms) fade-in with `var(--ease-enter)`
- Icon button color: `var(--text-dim)` default, `var(--text-primary)` on hover
- Swatch circles: 12px diameter, `border-radius: 50%`

The action buttons should match the traffic light button sizing pattern but use different styling -- 24px square, no background, icon only.

### Technical Considerations

- The popover must have `--wails-draggable: none` so it doesn't trigger window drag.
- Click-outside detection: attach a `svelte:window on:click` handler that checks `showPopover` and closes if the click target is not inside the popover or the palette button.
- `SetTheme(id)` is called after `applyTheme(id)` so the UI updates immediately even if the disk write is slow. Persistence failure is logged but doesn't block the UI.
- The `dispatch('open-settings')` event is caught by App.svelte's `<TitleBar on:open-settings={...}>`. This wiring is completed in Story 5.
- Popover positioning: absolute positioning relative to the titlebar-actions container. The popover appears below the palette button, right-aligned.

### Risks & Edge Cases

- If the popover is open and the user switches views (e.g., clicks on an agent), the popover should auto-close. This can be handled by watching `currentView` or simply closing on any non-popover click.
- On window resize, the popover position may shift. Using absolute positioning relative to the parent container (not fixed positioning) avoids this.
- The Settings gear button dispatches an event but the settings view doesn't exist yet (Story 5). The event simply fires with no handler until Story 5 wires it up. This is safe.

### Reference Files

- `frontend/src/components/TitleBar.svelte` -- current full file (76 lines)
- `frontend/src/App.svelte` -- lines 100-129 (template with TitleBar)
- `frontend/src/style.css` -- design system variables
- `frontend/src/views/AgentDetail.svelte` -- line 5 for lucide-svelte import pattern

## Acceptance Criteria

AC-1: TitleBar shows gear and palette icons on the right
- Given the app is running
- When the user looks at the title bar
- Then there are two icon buttons on the right side: a palette icon and a gear icon
- And the traffic lights remain on the left
- And the drag region still works between them

AC-2: Palette button toggles theme popover
- Given the app is running on the feed view
- When the user clicks the palette icon
- Then a popover appears below it showing three theme cards
- And clicking the palette icon again closes the popover

AC-3: Theme cards show label and color swatches
- Given the theme popover is open
- When inspecting a theme card
- Then it displays the theme's label (e.g., "Dark", "Light", "Midnight")
- And it shows 4 color swatches representing key colors from that theme
- And the currently active theme has a visual indicator (border or checkmark)

AC-4: Selecting a theme applies it and persists
- Given the theme popover is open and the current theme is "conductor-dark"
- When the user clicks the "Light" theme card
- Then the app UI immediately switches to light theme colors
- And `SetTheme("conductor-light")` is called to persist the choice
- And the popover closes

AC-5: Click outside closes the popover
- Given the theme popover is open
- When the user clicks anywhere outside the popover
- Then the popover closes

AC-6: Gear button dispatches open-settings event
- Given the app is running
- When the user clicks the gear icon
- Then an `open-settings` event is dispatched from the TitleBar component

## BDD Test Scenarios

### Scenario 1: TitleBar Icon Rendering

```gherkin
Feature: TitleBar Actions

  Scenario: Icons are rendered
    Given the TitleBar component is mounted
    When inspecting the DOM
    Then there is a button with title "Theme" containing a Palette icon
    And there is a button with title "Settings" containing a Settings icon
    And both buttons are in a container after the drag region

  Scenario: Icons do not interfere with window drag
    Given the TitleBar is rendered
    When the user clicks on the drag region between traffic lights and action buttons
    Then the window drag behavior activates
    And the action buttons have --wails-draggable: none
```

### Scenario 2: Theme Popover Behavior

```gherkin
Feature: Theme Popover

  Scenario: Popover opens on palette click
    Given the theme popover is closed
    When the user clicks the palette button
    Then the popover appears with 3 theme cards

  Scenario: Popover closes on second click
    Given the theme popover is open
    When the user clicks the palette button again
    Then the popover closes

  Scenario: Popover closes on outside click
    Given the theme popover is open
    When the user clicks on the feed area (outside popover)
    Then the popover closes

  Scenario: Active theme is highlighted
    Given the current theme is "conductor-dark"
    When the popover opens
    Then the "Dark" card has the active class
    And the other cards do not
```

### Scenario 3: Theme Selection

```gherkin
Feature: Theme Selection via Popover

  Scenario: Selecting a theme updates UI and persists
    Given the popover is open and current theme is "conductor-dark"
    When the user clicks the "Light" theme card
    Then applyTheme("conductor-light") is called
    And SetTheme("conductor-light") is called
    And the popover closes
    And CSS variables on :root match the light theme

  Scenario: Persistence failure does not block UI
    Given SetTheme will reject with an error
    When the user selects a new theme
    Then the UI still updates to the new theme immediately
    And the error is logged to console
```

## Tasks / Subtasks

- [ ] Task 1: Add action buttons to TitleBar (AC: AC-1, AC-6)
  - [ ] Subtask 1a: Import `Palette`, `Settings` from `lucide-svelte` and `createEventDispatcher` from `svelte`
  - [ ] Subtask 1b: Add `titlebar-actions` div after `drag-region` with two icon buttons
  - [ ] Subtask 1c: Style action buttons: 24px square, no bg, `var(--text-dim)` icons, hover to `var(--text-primary)`
  - [ ] Subtask 1d: Wire gear button to `dispatch('open-settings')`
  - [ ] Subtask 1e: Set `--wails-draggable: none` on the actions container

- [ ] Task 2: Implement theme popover (AC: AC-2, AC-3, AC-5)
  - [ ] Subtask 2a: Add popover HTML with `{#if showPopover}` block showing theme cards
  - [ ] Subtask 2b: Style popover: `var(--bg-elevated)` bg, `var(--border-subtle)` border, 100ms fade-in
  - [ ] Subtask 2c: Implement `toggleThemePopover()` for palette button
  - [ ] Subtask 2d: Implement click-outside handler via `svelte:window on:click`
  - [ ] Subtask 2e: Add color swatches (12px circles) showing 4 key colors per theme
  - [ ] Subtask 2f: Highlight active theme card with `var(--accent-green)` border

- [ ] Task 3: Wire theme selection to store and backend (AC: AC-4)
  - [ ] Subtask 3a: Import `applyTheme`, `currentThemeId` from `stores/theme.js`
  - [ ] Subtask 3b: Import `SetTheme` from Wails bindings
  - [ ] Subtask 3c: Implement `selectTheme(id)` -- calls `applyTheme`, closes popover, calls `SetTheme`
  - [ ] Subtask 3d: Handle `SetTheme` promise rejection gracefully (log, don't block UI)

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on new/modified files
- [ ] `go build ./...` passes
- [ ] `wails dev` launches with icons visible in title bar
- [ ] Theme popover opens, shows 3 themes, selecting one changes the entire UI
- [ ] Theme selection survives app restart
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
