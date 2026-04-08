# Story 5: Monaco Runtime Registration, Validation & Error Recovery

**Priority:** P1-high
**Domain:** frontend
**Estimated Complexity:** S
**Depends On:** Story 3, Story 4
**Status:** ready

## Description

Make MonacoEditor.svelte reactively register imported themes with Monaco's theme engine at runtime, so switching to an imported theme instantly updates syntax highlighting and editor UI colors without re-creating the editor. This story also adds theme validation, preview thumbnails for imported themes in Settings, and robust error recovery to ensure the app never gets stuck in a broken visual state.

## Developer Notes

### Architecture

**Modified files:**
1. `frontend/src/components/MonacoEditor.svelte` -- reactive theme registration with Monaco
2. `frontend/src/views/Settings.svelte` -- add preview thumbnails for imported themes
3. `frontend/src/lib/themeConverter.js` -- ensure `validateConvertedTheme` is exported (from Story 2)

**Monaco runtime registration pattern:**

Currently, MonacoEditor.svelte has this reactive block (line 430-432):
```js
$: if (monacoModule && $currentThemeId) {
    monacoModule.editor.setTheme($currentThemeId);
}
```

This only calls `setTheme` but never calls `defineTheme` for imported themes. Monaco will silently ignore `setTheme` for an undefined theme ID.

Replace with:
```js
import { allThemes, currentThemeId } from '../lib/stores/theme.js';

let registeredThemeIds = new Set();

$: if (monacoModule && $currentThemeId) {
    const theme = $allThemes[$currentThemeId];
    if (theme && theme.monaco && !registeredThemeIds.has($currentThemeId)) {
        monacoModule.editor.defineTheme($currentThemeId, theme.monaco);
        registeredThemeIds.add($currentThemeId);
    }
    monacoModule.editor.setTheme($currentThemeId);
}
```

This subscribes to both `currentThemeId` and `allThemes`. When the user activates an imported theme:
1. Story 4 registers the theme in `allThemes` and sets `currentThemeId`
2. This reactive block fires
3. If the theme hasn't been `defineTheme`'d yet, define it first
4. Then call `setTheme` to activate it

**Built-in themes:** The existing `defineAllThemes(monacoModule)` call in `onMount` (line 407) handles built-in themes at initialization. The `registeredThemeIds` set should be pre-populated with built-in theme IDs after that call to avoid redundant `defineTheme` calls.

```js
onMount(async () => {
    monacoModule = await import('monaco-editor');
    if (!themeRegistered) {
        defineAllThemes(monacoModule);
        // Pre-populate set with built-in IDs
        for (const id of builtInThemeIds) {
            registeredThemeIds.add(id);
        }
        themeRegistered = true;
    }
});
```

**Preview thumbnails for imported themes:**

Settings.svelte currently shows imported themes as simple text buttons. Upgrade to show the same preview card style as built-in themes, using the converted theme's CSS values:

```svelte
{#each vscodiumThemes as entry}
    {@const cached = convertedCache[entry.themePath]}
    <button class="theme-option" ...>
        {#if cached}
            <div class="theme-preview" style="background: {cached.theme.css['--bg-deepest']}; border-color: {cached.theme.css['--border-subtle']}">
                <div class="preview-bar" style="background: {cached.theme.css['--bg-surface']}; border-bottom-color: {cached.theme.css['--border-subtle']}">
                    <span class="preview-dot" style="background: #ff5f57" />
                    <span class="preview-dot" style="background: #febc2e" />
                    <span class="preview-dot" style="background: #28c840" />
                </div>
                <div class="preview-body">
                    <div class="preview-line" style="background: {cached.theme.css['--accent-green']}; width: 40%" />
                    <div class="preview-line" style="background: {cached.theme.css['--text-dim']}; width: 70%" />
                    <div class="preview-line" style="background: {cached.theme.css['--accent-purple']}; width: 30%" />
                </div>
            </div>
        {:else}
            <div class="theme-preview placeholder">
                <span class="badge-text">{entry.uiTheme === 'vs-dark' ? 'D' : 'L'}</span>
            </div>
        {/if}
        <span class="theme-name">{entry.label}</span>
    </button>
{/each}
```

The `convertedCache` is exposed from `themeInit.js` so Settings can access already-converted themes for preview. Alternatively, convert themes lazily on hover or eagerly on scan.

**Error recovery pattern:**

If any step in theme activation fails (read, parse, convert, register), the app must revert:

```js
try {
    // ... activation
} catch (e) {
    console.error('Theme load failed, reverting:', e);
    applyTheme(DEFAULT_THEME);
    await SetImportedTheme('');
    await SetTheme(DEFAULT_THEME);
}
```

### Technical Considerations

- **Reactive subscription to allThemes:** By accessing `$allThemes` in the reactive block, MonacoEditor subscribes to the store. This means any store update triggers the block. Since we only `defineTheme` for new IDs (via `registeredThemeIds` set), redundant store updates are cheap (just a set lookup).
- **Memory:** Each `defineTheme` call is permanent in the Monaco instance for the session. This is fine -- themes are small objects and users won't import hundreds.
- **Theme hot-swap:** Monaco's `setTheme` immediately re-renders the editor with new colors. No editor recreation needed.
- **Preview thumbnails lazy loading:** For a directory with 30+ themes, eagerly converting all for previews would be slow. Consider converting on first activation and caching, showing the badge placeholder for unconverted themes.

### Risks & Edge Cases

- **MonacoEditor not mounted:** If no editor is open when the theme changes, `monacoModule` is null. The reactive block's `if (monacoModule && ...)` guard handles this. The theme will be registered on the next editor mount.
- **Multiple MonacoEditor instances:** If two editors are open (unlikely in current UI but possible), both will react to theme changes. The `registeredThemeIds` set is per-component instance, so `defineTheme` may be called twice. This is harmless -- Monaco's `defineTheme` is idempotent.
- **Invalid monaco theme data:** If the converter produces malformed `IStandaloneThemeData`, `defineTheme` will throw. Wrap in try/catch.

### Reference Files

- `/Users/linus/Development/mashed/frontend/src/components/MonacoEditor.svelte` -- lines 6, 396-432 (onMount and reactive theme block)
- `/Users/linus/Development/mashed/frontend/src/lib/monacoTheme.js` -- `defineAllThemes`, `defineImportedTheme` (from Story 3)
- `/Users/linus/Development/mashed/frontend/src/lib/stores/theme.js` -- `allThemes`, `currentThemeId`, `builtInThemeIds` (from Story 3)
- `/Users/linus/Development/mashed/frontend/src/views/Settings.svelte` -- imported themes section (from Story 4)

## Acceptance Criteria

AC-1: Runtime Monaco registration
- Given an imported theme is activated while a MonacoEditor is mounted
- When `currentThemeId` changes to the imported theme's ID
- Then `defineTheme` is called with the imported theme's monaco data
- And `setTheme` is called to activate it
- And the editor immediately re-renders with the new colors

AC-2: No redundant defineTheme calls
- Given a theme has already been registered via `defineTheme`
- When `currentThemeId` is set to that theme again (e.g., switching back)
- Then `defineTheme` is NOT called again
- And only `setTheme` is called

AC-3: Built-in themes pre-registered
- Given the editor mounts and `defineAllThemes` runs
- When the `registeredThemeIds` set is checked
- Then it contains all built-in theme IDs
- And switching between built-in themes does not trigger extra `defineTheme` calls

AC-4: Error recovery on failed theme load
- Given a theme file that fails to parse (corrupt JSON, conversion error)
- When the user attempts to activate it
- Then the error is caught
- And the app reverts to DEFAULT_THEME
- And the config is updated to clear the imported theme

AC-5: Preview thumbnails for converted themes
- Given an imported theme has been activated (and thus converted)
- When the Settings view shows the imported themes list
- Then the activated theme shows a color preview card (not just a badge)
- And unconverted themes show a placeholder badge

AC-6: defineTheme error handling
- Given the converter produces invalid monaco theme data
- When `defineTheme` throws
- Then the error is caught
- And the previous theme remains active
- And the user sees an error indication

## BDD Test Scenarios

### Scenario 1: Monaco runtime registration

```gherkin
Feature: Monaco runtime theme registration

  Scenario: Register imported theme on activation
    Given MonacoEditor is mounted with monacoModule initialized
    And registeredThemeIds contains only built-in theme IDs
    When currentThemeId changes to "imported-dracula"
    And allThemes contains the Dracula theme object
    Then monacoModule.editor.defineTheme is called with ("imported-dracula", draculaMonacoData)
    And monacoModule.editor.setTheme is called with "imported-dracula"
    And registeredThemeIds now contains "imported-dracula"

  Scenario: Skip defineTheme for already-registered theme
    Given registeredThemeIds contains "imported-dracula"
    When currentThemeId changes to "imported-dracula"
    Then monacoModule.editor.defineTheme is NOT called
    And monacoModule.editor.setTheme IS called

  Scenario: Handle theme switch when editor is not mounted
    Given MonacoEditor is not mounted (monacoModule is null)
    When currentThemeId changes to "imported-dracula"
    Then no defineTheme or setTheme calls are made
    And when MonacoEditor later mounts, the current theme is applied
```

### Scenario 2: Error recovery

```gherkin
Feature: Theme error recovery

  Scenario: Corrupt theme file reverts to default
    Given the user clicks on a theme with corrupt JSON
    When ReadThemeFile succeeds but JSON.parse throws
    Then the error is caught
    And applyTheme(DEFAULT_THEME) is called
    And SetImportedTheme("") clears the persisted path
    And the app displays the default theme

  Scenario: defineTheme throws for invalid data
    Given an imported theme has invalid monaco data (missing base field)
    When the reactive block calls defineTheme
    Then the error is caught
    And the previous theme remains active
    And no unhandled exception propagates

  Scenario: Network/disk error during startup restore
    Given config has importedTheme set but the file is on a disconnected volume
    When App.svelte attempts restoreImportedThemeFromConfig
    Then ReadThemeFile returns an error
    And the app falls back to the built-in theme
    And no error dialog is shown to the user
```

### Scenario 3: Preview thumbnails

```gherkin
Feature: Theme preview thumbnails

  Scenario: Show preview for converted theme
    Given imported theme "Dracula" has been activated and cached
    When Settings view renders the imported themes list
    Then "Dracula" shows a color preview card with --bg-deepest, --bg-surface, --accent-green colors
    And the preview matches the theme's actual colors

  Scenario: Show placeholder for unconverted theme
    Given imported theme "Nord" has NOT been activated yet
    When Settings view renders the imported themes list
    Then "Nord" shows a dark/light badge placeholder
    And clicking it activates and then shows the full preview on re-render
```

## Tasks / Subtasks

- [ ] Task 1: Update MonacoEditor.svelte reactive theme block (AC: AC-1, AC-2, AC-3, AC-6)
  - [ ] Import `allThemes` and `builtInThemeIds` from store
  - [ ] Add `registeredThemeIds` set, pre-populated after `defineAllThemes`
  - [ ] Replace existing reactive theme block with define-then-set pattern
  - [ ] Wrap `defineTheme` in try/catch for invalid theme data
  - [ ] Test switching between built-in and imported themes

- [ ] Task 2: Add error recovery to activation flow (AC: AC-4)
  - [ ] In `themeInit.js` `activateImportedTheme`, wrap entire flow in try/catch
  - [ ] On catch: call `applyTheme(DEFAULT_THEME)`, `SetImportedTheme('')`, `SetTheme(DEFAULT_THEME)`
  - [ ] Log the error with `console.error` for debugging
  - [ ] In `restoreImportedThemeFromConfig`, same pattern with silent fallback

- [ ] Task 3: Add preview thumbnails to Settings (AC: AC-5)
  - [ ] Expose converted theme cache from `themeInit.js` (export the cache object or a getter)
  - [ ] In Settings.svelte, check cache for each theme entry
  - [ ] Render full preview card for cached themes, badge placeholder for others
  - [ ] Re-use existing `.theme-preview`, `.preview-bar`, `.preview-body` CSS from built-in grid

- [ ] Task 4: Write tests (AC: all)
  - [ ] Test reactive block: new theme triggers defineTheme + setTheme
  - [ ] Test reactive block: known theme skips defineTheme
  - [ ] Test error recovery: bad theme reverts to default
  - [ ] Test preview rendering with cached vs uncached themes

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on modified MonacoEditor.svelte theme logic
- [ ] Monaco editor visually reflects imported theme colors immediately on switch
- [ ] No console errors during theme switching
- [ ] App never renders unstyled after a failed theme load
- [ ] Code review: no CRITICAL/HIGH issues
