# Story 2: Frontend Theme Converter (B2 Hand-Rolled)

**Priority:** P0-critical
**Domain:** frontend
**Estimated Complexity:** M
**Depends On:** none
**Status:** ready

## Description

Create a new frontend module that converts VSCode color theme JSON into Conductor's unified theme format (css, monaco, xterm). This uses the B2 approach -- a hand-rolled TextMate scope mapper with zero npm dependencies. The converter handles editor UI color mapping, token color conversion with correct specificity ordering, xterm terminal theme derivation, and hex color normalization for all VSCode color formats.

## Developer Notes

### Architecture

**New file:** `frontend/src/lib/themeConverter.js`

This module exports three public functions:
- `convertVSCodeTheme(vsTheme, themeId)` -- main entry point, returns `{ label, css, monaco, xterm }`
- `validateConvertedTheme(theme)` -- checks minimum required CSS vars are present
- `getMonacoBase(vsTheme)` -- maps VSCode theme type to Monaco base theme

**Data flow:**
Raw VSCode theme JSON (parsed object) --> `convertVSCodeTheme()` --> Conductor unified format `{ label, css: {...}, monaco: {...}, xterm: {...} }`

The converter does NOT import from any other project file. It is a pure transformation module with no side effects.

### Technical Considerations

**Token specificity (H-5 fix):** The plan's `convertTokenColors` uses a `seenTokens` set with first-match-wins, which is backwards. VSCode themes list scopes from general to specific. Fix: process `tokenColors` array in REVERSE order so more-specific scopes win. When a scope like `comment.block.documentation` maps to Monaco token `comment`, it should take priority over a broader `comment` scope that appears earlier in the array.

```js
// Process in reverse so more-specific scopes override less-specific ones
for (let i = tokenColors.length - 1; i >= 0; i--) {
    const entry = tokenColors[i];
    // ... mapping logic, seenTokens prevents less-specific from overriding
}
```

**Hex normalization (L-1 fix):** The `dimColor` function assumes 7-char `#RRGGBB` input. VSCode colors can be `#RGB` (4-char), `#RRGGBB` (7-char), or `#RRGGBBAA` (9-char). Write a `normalizeHex(hex)` function:

```js
function normalizeHex(hex) {
    if (!hex || hex[0] !== '#') return hex;
    // #RGB -> #RRGGBB
    if (hex.length === 4) {
        return '#' + hex[1]+hex[1] + hex[2]+hex[2] + hex[3]+hex[3];
    }
    // #RRGGBBAA -> #RRGGBB (strip alpha)
    if (hex.length === 9) {
        return hex.slice(0, 7);
    }
    return hex;
}
```

**High-contrast themes (L-3 fix):** Map `hc-black` and `hc-light` UITheme values to their respective Monaco bases instead of defaulting everything non-light to `vs-dark`.

```js
function getMonacoBase(vsTheme) {
    const type = vsTheme.type || vsTheme.uiTheme || '';
    switch (type) {
        case 'light': case 'vs': return 'vs';
        case 'hc-black': return 'hc-black';
        case 'hc-light': return 'hc-light';
        default: return 'vs-dark';
    }
}
```

**Alpha channel in CSS vars (H-6 note):** For background CSS vars (`--bg-deepest`, `--bg-surface`, etc.), strip alpha from 8-char hex values to prevent see-through regions. For non-background colors, alpha can be preserved.

**Scope-to-token mapping table:** Use the mapping from the plan's B2 section. The `SCOPE_TO_TOKEN` array should have more-specific scopes listed first within each mapping entry (this is for matching priority within a single mapping, separate from the reverse-iteration fix above).

```js
const SCOPE_TO_TOKEN = [
    { scopes: ['comment.line', 'comment.block', 'comment', 'punctuation.definition.comment'], token: 'comment' },
    { scopes: ['string.quoted', 'string.template', 'string', 'string.regexp'], token: 'string' },
    { scopes: ['keyword.control', 'keyword.operator.new', 'keyword', 'storage.type', 'storage.modifier', 'storage'], token: 'keyword' },
    { scopes: ['constant.numeric', 'constant.numeric.integer', 'constant.numeric.float'], token: 'number' },
    { scopes: ['constant.language', 'constant.character', 'constant'], token: 'number' },
    { scopes: ['entity.name.type', 'entity.name.class', 'entity.other.inherited-class', 'support.type', 'support.class'], token: 'type' },
    { scopes: ['entity.name.function', 'support.function', 'meta.function-call'], token: 'type.identifier' },
    { scopes: ['variable.parameter', 'variable.other', 'variable', 'meta.definition.variable'], token: 'variable' },
    { scopes: ['keyword.operator', 'keyword.operator.assignment', 'keyword.operator.arithmetic'], token: 'operator' },
    { scopes: ['entity.name.tag', 'punctuation.definition.tag'], token: 'tag' },
    { scopes: ['entity.other.attribute-name'], token: 'attribute.name' },
    { scopes: ['punctuation.bracket', 'punctuation.separator', 'punctuation', 'meta.brace'], token: 'delimiter' },
];
```

**Monaco rule format:** Monaco expects foreground colors WITHOUT the `#` prefix. The converter must strip `#` from foreground values in token rules.

```js
rule.foreground = entry.settings.foreground.replace('#', '');
```

### Risks & Edge Cases

- **Themes with no tokenColors:** Some minimal themes only define `colors`. The converter should return an empty `rules` array and rely on Monaco's base theme for syntax highlighting.
- **Themes with no colors object:** Edge case where only tokenColors are provided. CSS var mapper should use all fallback defaults.
- **Themes with scope as a string vs array:** The `scope` field in tokenColors can be a string (`"comment"`) or an array (`["comment", "comment.block"]`). Handle both.
- **Global tokenColor entries:** Some entries have no `scope` field -- these set the default foreground/background. Handle by checking for `entry.scope` existence.
- **Empty or null settings:** Some tokenColor entries have empty `settings: {}`. Skip gracefully.

### Reference Files

- `/Users/linus/Development/mashed/frontend/src/lib/themes.js` -- the target format. Each built-in theme has `{ label, css, monaco, xterm }`. The converter must produce this exact shape.
- `/Users/linus/Development/mashed/docs/vscodium-theme-loading-plan.md` lines 398-475 -- the `mapVSCodeColorsToCSSVars`, `deriveXtermTheme`, and `dimColor` functions to adapt.
- `/Users/linus/Development/mashed/docs/vscodium-theme-loading-plan.md` lines 536-645 -- the B2 scope mapper and `convertTokenColors` function to adapt (with H-5 fix).

## Acceptance Criteria

AC-1: CSS variable mapping
- Given a VSCode dark theme JSON with `colors.editor.background`, `colors.editor.foreground`, and terminal ANSI colors
- When `convertVSCodeTheme(vsTheme, 'test-theme')` is called
- Then the returned object has a `css` property with all 16 Conductor CSS custom properties populated
- And missing colors fall back to sensible dark-theme defaults

AC-2: Monaco theme generation
- Given a VSCode theme JSON with both `colors` and `tokenColors`
- When `convertVSCodeTheme()` is called
- Then the returned `monaco` property is a valid `IStandaloneThemeData` with `base`, `inherit: true`, `rules`, and `colors`
- And `base` is correctly set to `vs-dark`, `vs`, `hc-black`, or `hc-light` based on theme type

AC-3: Token color conversion with correct specificity
- Given a VSCode theme with tokenColors containing both `"comment"` (foreground: `#aaa`) and `"comment.block.documentation"` (foreground: `#bbb`)
- When `convertVSCodeTheme()` is called
- Then the Monaco rule for token `comment` has foreground from the more-specific `comment.block.documentation` scope (`bbb`)

AC-4: xterm theme derivation
- Given a VSCode theme with terminal ANSI colors defined
- When `convertVSCodeTheme()` is called
- Then the returned `xterm` property has `background`, `foreground`, `cursor`, and all 8 ANSI color fields populated

AC-5: Hex normalization
- Given colors in `#RGB`, `#RRGGBB`, and `#RRGGBBAA` formats
- When processed through the converter
- Then CSS background vars contain only `#RRGGBB` format (alpha stripped)
- And `dimColor` correctly handles all input formats

AC-6: High-contrast theme support
- Given a VSCode theme with `type: "hc-black"` or `uiTheme: "hc-black"`
- When `convertVSCodeTheme()` is called
- Then the Monaco `base` is `"hc-black"` (not `"vs-dark"`)

AC-7: Theme validation
- Given a converted theme missing `--bg-deepest` CSS variable
- When `validateConvertedTheme(theme)` is called
- Then it returns `false`

## BDD Test Scenarios

### Scenario 1: Full Dracula theme conversion

```gherkin
Feature: VSCode theme conversion

  Scenario: Convert a complete dark theme
    Given a VSCode theme JSON with:
      | field              | value                    |
      | name               | Dracula                  |
      | type               | dark                     |
      | colors.editor.background | #282a36            |
      | colors.editor.foreground | #f8f8f2            |
      | colors.terminal.ansiGreen | #50fa7b           |
    And tokenColors containing:
      | scope    | foreground | fontStyle |
      | comment  | #6272A4    | italic    |
      | string   | #F1FA8C    |           |
      | keyword  | #FF79C6    | bold      |
    When convertVSCodeTheme is called with themeId "imported-dracula"
    Then the result has label "Dracula"
    And result.css['--bg-deepest'] is "#282a36"
    And result.css['--accent-green'] is "#50fa7b"
    And result.monaco.base is "vs-dark"
    And result.monaco.rules contains { token: "comment", foreground: "6272A4", fontStyle: "italic" }
    And result.xterm.green is "#50fa7b"
```

### Scenario 2: Token specificity ordering

```gherkin
Feature: Token specificity

  Scenario: More-specific scope wins over general scope
    Given tokenColors in order:
      | scope                        | foreground |
      | comment                      | #aaaaaa    |
      | comment.block.documentation  | #bbbbbb    |
    When convertTokenColors processes these entries
    Then the rule for token "comment" has foreground "bbbbbb"

  Scenario: Scope as string vs array
    Given a tokenColor entry with scope "keyword.control, storage.type"
    And settings foreground "#FF79C6"
    When convertTokenColors processes this entry
    Then a rule for token "keyword" is generated with foreground "FF79C6"
```

### Scenario 3: Hex normalization edge cases

```gherkin
Feature: Hex color normalization

  Scenario: Normalize short hex
    Given a color value "#f8f"
    When normalizeHex is called
    Then the result is "#ff88ff"

  Scenario: Strip alpha from 8-char hex
    Given a color value "#282a36ff"
    When normalizeHex is called
    Then the result is "#282a36"

  Scenario: dimColor handles normalized input
    Given a color "#50fa7b" and factor 0.4
    When dimColor is called
    Then the result is a valid 7-char hex string
    And the RGB values are approximately 40% of the original
```

### Scenario 4: Light and high-contrast themes

```gherkin
Feature: Theme type handling

  Scenario: Light theme uses vs base
    Given a VSCode theme with type "light"
    When convertVSCodeTheme is called
    Then result.monaco.base is "vs"
    And result.css['--bg-deepest'] uses light fallback "#f8f9fb"

  Scenario: High-contrast black uses hc-black base
    Given a VSCode theme with uiTheme "hc-black"
    When convertVSCodeTheme is called
    Then result.monaco.base is "hc-black"
```

### Scenario 5: Missing data graceful handling

```gherkin
Feature: Graceful degradation

  Scenario: Theme with no tokenColors
    Given a VSCode theme with colors but no tokenColors array
    When convertVSCodeTheme is called
    Then result.monaco.rules is an empty array
    And result.monaco.inherit is true (falls back to base)

  Scenario: Theme with no colors object
    Given a VSCode theme with tokenColors but no colors object
    When convertVSCodeTheme is called
    Then all CSS vars use fallback defaults
    And result.monaco.colors is an empty object
```

## Tasks / Subtasks

- [ ] Task 1: Create themeConverter.js with utility functions (AC: AC-5, AC-6)
  - [ ] Create `/Users/linus/Development/mashed/frontend/src/lib/themeConverter.js`
  - [ ] Implement `normalizeHex(hex)` -- handles #RGB, #RRGGBB, #RRGGBBAA formats
  - [ ] Implement `dimColor(hex, factor)` -- uses normalizeHex before parsing
  - [ ] Implement `getMonacoBase(vsTheme)` -- maps vs/vs-dark/hc-black/hc-light

- [ ] Task 2: Implement CSS variable mapping (AC: AC-1)
  - [ ] Implement `mapVSCodeColorsToCSSVars(colors, isDark)` with all 16 CSS custom properties
  - [ ] Dark theme fallback defaults matching conductor-dark from themes.js
  - [ ] Light theme fallback defaults matching conductor-light from themes.js
  - [ ] Strip alpha from background CSS vars (`--bg-deepest`, `--bg-surface`, `--bg-elevated`, `--bg-active`)

- [ ] Task 3: Implement token color conversion (AC: AC-2, AC-3)
  - [ ] Define `SCOPE_TO_TOKEN` mapping array
  - [ ] Implement `convertTokenColors(tokenColors)` with reverse-iteration for specificity
  - [ ] Handle scope as string (comma-separated) and array
  - [ ] Handle entries with no scope (global defaults)
  - [ ] Strip `#` prefix from foreground values for Monaco format

- [ ] Task 4: Implement xterm theme derivation (AC: AC-4)
  - [ ] Implement `deriveXtermTheme(colors, isDark)` mapping terminal.ansi* colors
  - [ ] Fallback to editor colors when terminal colors are not defined

- [ ] Task 5: Implement main converter and validator (AC: AC-1 through AC-7)
  - [ ] Implement `convertVSCodeTheme(vsTheme, themeId)` combining all sub-functions
  - [ ] Implement `validateConvertedTheme(theme)` checking required CSS vars
  - [ ] Export all three public functions

- [ ] Task 6: Write tests (AC: all)
  - [ ] Test with a mock Dracula-like theme (dark, complete)
  - [ ] Test with a mock light theme
  - [ ] Test with hc-black theme
  - [ ] Test token specificity (general before specific in input)
  - [ ] Test hex normalization edge cases
  - [ ] Test missing colors/tokenColors graceful handling
  - [ ] Test validation for incomplete themes

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on `themeConverter.js`
- [ ] Module has zero imports from other project files (pure transformation)
- [ ] Module has zero npm dependency additions
- [ ] Code review: no CRITICAL/HIGH issues
