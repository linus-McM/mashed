# VSCode / VSCodium Theme Loading — Implementation Plan

This document covers three approaches for loading VSCode/VSCodium color themes into Conductor's Monaco editor, ordered by complexity. Each approach builds on the previous one and can be adopted incrementally.

All approaches share the same Go backend (Phase 1) and Settings UI (Phase 4). The difference is in how thoroughly the theme data is converted on the frontend.

---

## Architecture Overview

```
┌────────────────────────────────────────────────────────────────┐
│  Settings.svelte                                               │
│  ┌──────────────┐  ┌─────────────────────────────────────────┐ │
│  │ Built-in     │  │ VSCodium Themes (new section)           │ │
│  │ Dark/Light/  │  │ ┌────────┐ ┌────────┐ ┌────────┐       │ │
│  │ Midnight     │  │ │Dracula │ │ Nord   │ │One Dark│ ...   │ │
│  │              │  │ └────────┘ └────────┘ └────────┘       │ │
│  └──────────────┘  └─────────────────────────────────────────┘ │
└────────────────────────────────────────────────────────────────┘
         │                          │
         ▼                          ▼
┌─────────────────┐    ┌──────────────────────────────┐
│ themes.js       │    │ themeConverter.js (new)       │
│ (built-in defs) │    │ VSCode JSON → Conductor fmt   │
└────────┬────────┘    └──────────────┬───────────────┘
         │                            │
         ▼                            ▼
┌──────────────────────────────────────────────┐
│ stores/theme.js                              │
│ currentThemeId, applyTheme(), importedThemes │
└──────────────────┬───────────────────────────┘
                   │
         ┌─────────┴──────────┐
         ▼                    ▼
┌────────────────┐   ┌───────────────┐
│ Monaco editor  │   │ xterm.js      │
│ defineTheme()  │   │ terminal.opts │
│ setTheme()     │   │               │
└────────────────┘   └───────────────┘
```

The Go backend handles filesystem access (scanning the extensions directory, reading theme JSON files). The Svelte frontend handles conversion and registration.

---

## Phase 1: Go Backend — Extension Scanner

This phase is shared by all three approaches. It adds two new Wails-bound functions to `app.go`.

### 1.1 Data Structures

Add to `app.go` (or a new `internal/themes/scanner.go` if you prefer):

```go
// VSCodeThemeEntry represents a single theme discovered in the extensions directory.
type VSCodeThemeEntry struct {
    // Label is the human-readable name from package.json contributes.themes[].label
    Label string `json:"label"`
    // ExtensionID is the directory name, e.g. "dracula-theme.theme-dracula-2.24.3"
    ExtensionID string `json:"extensionId"`
    // ThemePath is the absolute path to the theme JSON file
    ThemePath string `json:"themePath"`
    // UITheme is "vs-dark", "vs", or "hc-black" from the package.json declaration
    UITheme string `json:"uiTheme"`
}
```

### 1.2 ListVSCodiumThemes

This function scans the configured VSCodium extensions directory and returns all discovered themes.

```go
// ListVSCodiumThemes scans the VSCodium extensions directory and returns all
// color themes found in installed extensions.
func (a *App) ListVSCodiumThemes() ([]VSCodeThemeEntry, error) {
    cfg := loadConfig()
    extDir := cfg.VSCodiumExtPath
    if extDir == "" {
        return nil, fmt.Errorf("VSCodium extensions path not configured")
    }

    // Expand ~ if present
    if strings.HasPrefix(extDir, "~") {
        home, _ := os.UserHomeDir()
        extDir = filepath.Join(home, extDir[1:])
    }

    entries, err := os.ReadDir(extDir)
    if err != nil {
        return nil, fmt.Errorf("reading extensions directory: %w", err)
    }

    var themes []VSCodeThemeEntry

    for _, entry := range entries {
        if !entry.IsDir() {
            continue
        }

        pkgPath := filepath.Join(extDir, entry.Name(), "package.json")
        pkgData, err := os.ReadFile(pkgPath)
        if err != nil {
            continue // Not a valid extension, skip
        }

        // Parse package.json to find contributes.themes
        var pkg struct {
            Contributes struct {
                Themes []struct {
                    Label   string `json:"label"`
                    UITheme string `json:"uiTheme"`
                    Path    string `json:"path"`
                } `json:"themes"`
            } `json:"contributes"`
        }
        if err := json.Unmarshal(pkgData, &pkg); err != nil {
            continue
        }

        for _, t := range pkg.Contributes.Themes {
            if t.Path == "" {
                continue
            }

            // Resolve the theme file path relative to the extension directory
            themePath := filepath.Join(extDir, entry.Name(), t.Path)

            // Verify the file exists
            if _, err := os.Stat(themePath); err != nil {
                continue
            }

            label := t.Label
            if label == "" {
                label = entry.Name()
            }

            uiTheme := t.UITheme
            if uiTheme == "" {
                uiTheme = "vs-dark" // Default assumption
            }

            themes = append(themes, VSCodeThemeEntry{
                Label:       label,
                ExtensionID: entry.Name(),
                ThemePath:   themePath,
                UITheme:     uiTheme,
            })
        }
    }

    sort.Slice(themes, func(i, j int) bool {
        return themes[i].Label < themes[j].Label
    })

    return themes, nil
}
```

### 1.3 ReadThemeFile

This function reads a theme JSON file and returns its raw contents. The frontend will handle parsing and conversion.

```go
// ReadThemeFile reads a VSCode color theme JSON file and returns its contents.
// Only allows reading files within the configured extensions directory for safety.
func (a *App) ReadThemeFile(themePath string) (string, error) {
    if themePath == "" {
        return "", fmt.Errorf("empty theme path")
    }

    // Security: verify the path is within the configured extensions directory
    cfg := loadConfig()
    extDir := cfg.VSCodiumExtPath
    if extDir == "" {
        return "", fmt.Errorf("extensions path not configured")
    }
    if strings.HasPrefix(extDir, "~") {
        home, _ := os.UserHomeDir()
        extDir = filepath.Join(home, extDir[1:])
    }

    absTheme, _ := filepath.Abs(themePath)
    absExt, _ := filepath.Abs(extDir)
    if !strings.HasPrefix(absTheme, absExt) {
        return "", fmt.Errorf("theme path is outside extensions directory")
    }

    data, err := os.ReadFile(themePath)
    if err != nil {
        return "", fmt.Errorf("reading theme file: %w", err)
    }

    // Cap at 512KB — theme files are typically 20-60KB
    if len(data) > 512*1024 {
        return "", fmt.Errorf("theme file too large (>512KB)")
    }

    return string(data), nil
}
```

### 1.4 Persist Selected Imported Theme

Add a field to `conductorConfig` so the user's imported theme selection survives app restart:

```go
type conductorConfig struct {
    DevDir          string `json:"devDir"`
    Theme           string `json:"theme,omitempty"`
    VSCodiumExtPath string `json:"vscodiumExtPath,omitempty"`
    ImportedTheme   string `json:"importedTheme,omitempty"` // NEW: path to active imported theme
}
```

And a setter:

```go
// SetImportedTheme persists the path to the active imported VSCodium theme.
// Pass an empty string to clear (revert to built-in theme).
func (a *App) SetImportedTheme(themePath string) error {
    cfg := loadConfig()
    cfg.ImportedTheme = themePath
    return saveConfig(cfg)
}
```

### 1.5 Wails Bindings

These functions are automatically exposed to the frontend by Wails since they are public methods on `*App`. The TypeScript bindings will be auto-generated at build time into `frontend/wailsjs/go/main/App.js`.

After adding the functions, run:

```bash
wails generate module
```

This creates the frontend bindings you can import:

```js
import {
    ListVSCodiumThemes,
    ReadThemeFile,
    SetImportedTheme
} from '../../wailsjs/go/main/App.js';
```

### 1.6 VSCode Extension Directory Structures

For reference, here is what the scanner will encounter. VSCodium stores installed extensions as unzipped directories:

```
~/.vscode-oss/extensions/
├── dracula-theme.theme-dracula-2.24.3/
│   ├── package.json
│   └── theme/
│       └── dracula.json                    ← Color theme file
├── catppuccin.catppuccin-vsc-3.15.0/
│   ├── package.json
│   └── themes/
│       ├── catppuccin-latte.json
│       ├── catppuccin-frappe.json
│       ├── catppuccin-macchiato.json
│       └── catppuccin-mocha.json           ← Multiple themes per extension
├── github.github-vscode-theme-6.3.5/
│   ├── package.json
│   └── themes/
│       ├── dark.json
│       ├── light.json
│       ├── dimmed.json
│       └── dark-default.json
└── zhuangtongfa.material-theme-3.17.5/     ← One Dark Pro
    ├── package.json
    └── themes/
        └── OneDark-Pro.json
```

The relevant `package.json` section:

```json
{
  "contributes": {
    "themes": [
      {
        "label": "Dracula",
        "uiTheme": "vs-dark",
        "path": "./theme/dracula.json"
      }
    ]
  }
}
```

Some extensions use `.tmTheme` (TextMate XML) instead of JSON. The scanner should filter for `.json` files only — `.tmTheme` conversion requires additional parsing and is not worth the complexity for the initial implementation.

---

## Phase 2: Frontend Theme Converter

This is where the three approaches diverge.

### VSCode Theme JSON Structure (Reference)

A VSCode color theme file looks like this:

```json
{
  "name": "Dracula",
  "type": "dark",
  "colors": {
    "editor.background": "#282a36",
    "editor.foreground": "#f8f8f2",
    "editorCursor.foreground": "#f8f8f2",
    "editor.selectionBackground": "#44475a",
    "editor.lineHighlightBackground": "#44475a75",
    "editorLineNumber.foreground": "#6272a4",
    "editorLineNumber.activeForeground": "#f8f8f2",
    "editorWidget.background": "#21222c",
    "editorWidget.border": "#191a21"
  },
  "tokenColors": [
    {
      "scope": ["comment", "punctuation.definition.comment"],
      "settings": {
        "foreground": "#6272A4",
        "fontStyle": "italic"
      }
    },
    {
      "scope": ["string", "string.template"],
      "settings": {
        "foreground": "#F1FA8C"
      }
    },
    {
      "scope": ["keyword", "storage.type", "storage.modifier"],
      "settings": {
        "foreground": "#FF79C6",
        "fontStyle": "bold"
      }
    }
  ]
}
```

Conductor's unified theme format (from `themes.js`) has three parts: `css`, `monaco`, and `xterm`. We need to convert the VSCode JSON into all three.

---

### Approach A: Color Palette Only (Simplest)

**Effort:** ~1-2 hours | **Fidelity:** Editor UI colors accurate, syntax highlighting uses Monaco defaults

**What you get:** The background, cursor, selection, scrollbar, gutter, widget colors all match the imported theme. Syntax highlighting uses Monaco's built-in tokenizer colors inherited from the `vs-dark` or `vs` base theme.

**What you miss:** Custom syntax highlighting colors (strings, keywords, comments, etc. will use Monaco's defaults, not the theme's).

#### File: `frontend/src/lib/themeConverter.js`

```js
/**
 * Approach A: Extract editor colors only, inherit syntax from base theme.
 *
 * @param {object} vsTheme  – Parsed VSCode color theme JSON
 * @param {string} themeId  – Unique ID for this theme (e.g. "imported-dracula")
 * @returns {object}        – Conductor unified theme format { label, css, monaco, xterm }
 */
export function convertVSCodeThemeSimple(vsTheme, themeId) {
  const isDark = vsTheme.type !== 'light';
  const colors = vsTheme.colors || {};

  // ─── Monaco theme (editor colors only, inherit syntax rules) ───
  const monaco = {
    base: isDark ? 'vs-dark' : 'vs',
    inherit: true,
    rules: [],  // Empty — inherit all token rules from base
    colors: { ...colors },  // Pass through all VSCode editor color keys
  };

  // ─── CSS custom properties (map VSCode colors → Conductor variables) ───
  const css = mapVSCodeColorsToCSSVars(colors, isDark);

  // ─── xterm.js theme (derive from editor colors or ANSI palette) ───
  const xterm = deriveXtermTheme(colors, isDark);

  return {
    label: vsTheme.name || themeId,
    css,
    monaco,
    xterm,
  };
}

/**
 * Map VSCode editor colors to Conductor's 16 CSS custom properties.
 * Falls back to sensible defaults when a color is missing.
 */
function mapVSCodeColorsToCSSVars(colors, isDark) {
  if (isDark) {
    return {
      '--bg-deepest':       colors['editor.background']            || '#07080a',
      '--bg-surface':       colors['editorWidget.background']      || '#0d0f12',
      '--bg-elevated':      colors['sideBar.background']           || colors['editorWidget.background'] || '#12151a',
      '--bg-active':        colors['list.activeSelectionBackground'] || colors['editor.lineHighlightBackground'] || '#181c23',
      '--border-subtle':    colors['editorWidget.border']          || colors['panel.border'] || '#1e2530',
      '--border-emphasis':  colors['focusBorder']                  || '#2a3340',
      '--accent-green':     colors['terminal.ansiGreen']           || '#00e57a',
      '--accent-green-dim': dimColor(colors['terminal.ansiGreen']  || '#00e57a', 0.4),
      '--accent-amber':     colors['terminal.ansiYellow']          || '#f0a500',
      '--accent-red':       colors['terminal.ansiRed']             || '#e84545',
      '--accent-blue':      colors['terminal.ansiBlue']            || '#3d9eff',
      '--accent-purple':    colors['terminal.ansiMagenta']         || '#9d6fff',
      '--accent-teal':      colors['terminal.ansiCyan']            || '#00c4b3',
      '--text-primary':     colors['editor.foreground']            || '#c8d4e0',
      '--text-dim':         colors['editorLineNumber.foreground']  || '#4a5a6a',
      '--text-muted':       colors['editorWhitespace.foreground']  || '#2e3d4d',
    };
  }

  // Light theme mapping
  return {
    '--bg-deepest':       colors['editor.background']            || '#f8f9fb',
    '--bg-surface':       colors['editorWidget.background']      || '#ffffff',
    '--bg-elevated':      colors['sideBar.background']           || '#f0f2f5',
    '--bg-active':        colors['list.activeSelectionBackground'] || '#e8ecf0',
    '--border-subtle':    colors['editorWidget.border']          || '#d0d7e0',
    '--border-emphasis':  colors['focusBorder']                  || '#b8c2cc',
    '--accent-green':     colors['terminal.ansiGreen']           || '#059a50',
    '--accent-green-dim': dimColor(colors['terminal.ansiGreen']  || '#059a50', 0.3),
    '--accent-amber':     colors['terminal.ansiYellow']          || '#c07800',
    '--accent-red':       colors['terminal.ansiRed']             || '#cc3333',
    '--accent-blue':      colors['terminal.ansiBlue']            || '#2270cc',
    '--accent-purple':    colors['terminal.ansiMagenta']         || '#7044cc',
    '--accent-teal':      colors['terminal.ansiCyan']            || '#0a8a7a',
    '--text-primary':     colors['editor.foreground']            || '#1a1e24',
    '--text-dim':         colors['editorLineNumber.foreground']  || '#4a5568',
    '--text-muted':       colors['editorWhitespace.foreground']  || '#8896a6',
  };
}

/**
 * Derive xterm.js theme from VSCode terminal colors.
 */
function deriveXtermTheme(colors, isDark) {
  return {
    background:          colors['terminal.background']        || colors['editor.background'] || (isDark ? '#07080a' : '#f8f9fb'),
    foreground:          colors['terminal.foreground']        || colors['editor.foreground'] || (isDark ? '#c8d4e0' : '#1a1e24'),
    cursor:              colors['terminalCursor.foreground']  || colors['editorCursor.foreground'] || '#00e57a',
    cursorAccent:        colors['terminalCursor.background']  || colors['editor.background'] || (isDark ? '#07080a' : '#f8f9fb'),
    selectionBackground: colors['terminal.selectionBackground'] || colors['editor.selectionBackground'] || '#2a3340',
    black:               colors['terminal.ansiBlack']         || (isDark ? '#07080a' : '#1a1e24'),
    red:                 colors['terminal.ansiRed']           || '#e84545',
    green:               colors['terminal.ansiGreen']         || '#00e57a',
    yellow:              colors['terminal.ansiYellow']        || '#f0a500',
    blue:                colors['terminal.ansiBlue']          || '#3d9eff',
    magenta:             colors['terminal.ansiMagenta']       || '#9d6fff',
    cyan:                colors['terminal.ansiCyan']          || '#00c4b3',
    white:               colors['terminal.ansiWhite']         || (isDark ? '#c8d4e0' : '#f8f9fb'),
  };
}

/**
 * Dim a hex color by mixing it toward black.
 * factor: 0.0 = black, 1.0 = original color
 */
function dimColor(hex, factor) {
  const r = parseInt(hex.slice(1, 3), 16);
  const g = parseInt(hex.slice(3, 5), 16);
  const b = parseInt(hex.slice(5, 7), 16);
  const dr = Math.round(r * factor);
  const dg = Math.round(g * factor);
  const db = Math.round(b * factor);
  return `#${dr.toString(16).padStart(2, '0')}${dg.toString(16).padStart(2, '0')}${db.toString(16).padStart(2, '0')}`;
}
```

#### Why this works

Monaco's `colors` object accepts the same keys as VSCode's `colors` object — they share the API surface. Setting `inherit: true` with `base: 'vs-dark'` means Monaco's built-in syntax highlighting rules still apply, so code is readable even without token color conversion.

---

### Approach B: Token Color Conversion (Recommended)

**Effort:** ~3-4 hours | **Fidelity:** Good syntax highlighting for common token types

**What you get:** Everything from Approach A, plus syntax highlighting colors for comments, strings, keywords, numbers, types, functions, variables, and operators.

**What you miss:** Deeply nested TextMate scopes (e.g. `meta.function.parameters.js`, `entity.other.inherited-class.python`) won't map. Language-specific scopes fall back to the base theme.

There are two sub-options here:

#### Option B1: Use `monaco-vscode-textmate-theme-converter` (npm package)

```bash
cd frontend
npm install monaco-vscode-textmate-theme-converter
```

```js
import { convertTheme } from 'monaco-vscode-textmate-theme-converter';

/**
 * Approach B1: Use the converter package for token mapping.
 */
export function convertVSCodeThemeWithTokens(vsTheme, themeId) {
  const isDark = vsTheme.type !== 'light';
  const colors = vsTheme.colors || {};

  // The converter handles both colors and tokenColors → IStandaloneThemeData
  const monacoTheme = convertTheme(vsTheme);

  const css = mapVSCodeColorsToCSSVars(colors, isDark);
  const xterm = deriveXtermTheme(colors, isDark);

  return {
    label: vsTheme.name || themeId,
    css,
    monaco: monacoTheme,
    xterm,
  };
}
```

**Caveat:** This package was last updated in 2023. Test it with your Monaco 0.55.1 — it should work since the `IStandaloneThemeData` interface hasn't changed, but verify.

#### Option B2: Hand-rolled scope mapper (no dependency)

If you'd rather avoid the npm dependency, here's a manual mapper. This is more work but gives you full control.

```js
/**
 * TextMate scope → Monaco token type mapping.
 * Order matters — more specific scopes should come first.
 */
const SCOPE_TO_TOKEN = [
  // Comments
  { scopes: ['comment.line', 'comment.block', 'comment', 'punctuation.definition.comment'], token: 'comment' },

  // Strings
  { scopes: ['string.quoted', 'string.template', 'string', 'string.regexp'], token: 'string' },

  // Keywords
  { scopes: ['keyword.control', 'keyword.operator.new', 'keyword', 'storage.type', 'storage.modifier', 'storage'], token: 'keyword' },

  // Numbers
  { scopes: ['constant.numeric', 'constant.numeric.integer', 'constant.numeric.float'], token: 'number' },

  // Constants (booleans, null, etc.)
  { scopes: ['constant.language', 'constant.character', 'constant'], token: 'number' },

  // Types and classes
  { scopes: ['entity.name.type', 'entity.name.class', 'entity.other.inherited-class', 'support.type', 'support.class'], token: 'type' },

  // Functions
  { scopes: ['entity.name.function', 'support.function', 'meta.function-call'], token: 'type.identifier' },

  // Variables and parameters
  { scopes: ['variable.parameter', 'variable.other', 'variable', 'meta.definition.variable'], token: 'variable' },

  // Operators
  { scopes: ['keyword.operator', 'keyword.operator.assignment', 'keyword.operator.arithmetic'], token: 'operator' },

  // Tags (HTML/JSX)
  { scopes: ['entity.name.tag', 'punctuation.definition.tag'], token: 'tag' },

  // Attributes
  { scopes: ['entity.other.attribute-name'], token: 'attribute.name' },

  // Delimiters (brackets, punctuation)
  { scopes: ['punctuation.bracket', 'punctuation.separator', 'punctuation', 'meta.brace'], token: 'delimiter' },
];

/**
 * Convert VSCode tokenColors into Monaco ITokenThemeRule[].
 */
function convertTokenColors(tokenColors) {
  if (!tokenColors || !Array.isArray(tokenColors)) return [];

  const rules = [];
  const seenTokens = new Set();

  for (const entry of tokenColors) {
    if (!entry.settings || !entry.scope) continue;

    const scopes = Array.isArray(entry.scope)
      ? entry.scope
      : entry.scope.split(',').map(s => s.trim());

    for (const scope of scopes) {
      // Find the best matching Monaco token type
      for (const mapping of SCOPE_TO_TOKEN) {
        const match = mapping.scopes.some(ms =>
          scope === ms || scope.startsWith(ms + '.')
        );

        if (match && !seenTokens.has(mapping.token)) {
          const rule = { token: mapping.token };

          if (entry.settings.foreground) {
            // Strip # prefix — Monaco expects bare hex
            rule.foreground = entry.settings.foreground.replace('#', '');
          }
          if (entry.settings.fontStyle) {
            rule.fontStyle = entry.settings.fontStyle;
          }

          rules.push(rule);
          seenTokens.add(mapping.token);
          break;
        }
      }
    }
  }

  return rules;
}

/**
 * Approach B2: Full conversion with hand-rolled scope mapper.
 */
export function convertVSCodeThemeFull(vsTheme, themeId) {
  const isDark = vsTheme.type !== 'light';
  const colors = vsTheme.colors || {};

  const tokenRules = convertTokenColors(vsTheme.tokenColors);

  const monaco = {
    base: isDark ? 'vs-dark' : 'vs',
    inherit: true,
    rules: tokenRules,
    colors: { ...colors },
  };

  const css = mapVSCodeColorsToCSSVars(colors, isDark);
  const xterm = deriveXtermTheme(colors, isDark);

  return {
    label: vsTheme.name || themeId,
    css,
    monaco,
    xterm,
  };
}
```

**Note:** Monaco's tokenizer uses simple token names like `comment`, `string`, `keyword`. The mapping above covers the ~12 most common token types. This provides good visual fidelity for all mainstream themes.

---

### Approach C: Full TextMate via Shiki (Maximum Fidelity)

**Effort:** ~1-2 days | **Fidelity:** Pixel-perfect VSCode syntax highlighting

**What you get:** Exact VSCode rendering. Shiki uses the same Oniguruma regex engine and TextMate grammars as VSCode.

**What it costs:** ~1-2MB WASM payload, more complex initialization, grammar files for each language.

#### Install Dependencies

```bash
cd frontend
npm install shiki @shikijs/monaco
```

#### Integration Points

Shiki replaces Monaco's built-in tokenizer entirely. Instead of `defineTheme` + token rules, you wire Shiki as the tokenization provider.

```js
// frontend/src/lib/shikiSetup.js

import { createHighlighter } from 'shiki';
import { shikiToMonaco } from '@shikijs/monaco';

let highlighterInstance = null;

/**
 * Initialize the Shiki highlighter with the languages your editor needs.
 * Call this once during app startup, before creating any Monaco editors.
 */
export async function initShiki() {
  if (highlighterInstance) return highlighterInstance;

  highlighterInstance = await createHighlighter({
    // Start with the languages Conductor's getLanguage() function supports
    langs: [
      'go', 'javascript', 'typescript', 'html', 'css', 'json',
      'markdown', 'yaml', 'toml', 'shellscript', 'python', 'rust', 'sql',
    ],
    // Load your built-in themes plus any default popular themes
    themes: ['vitesse-dark', 'vitesse-light'],
  });

  return highlighterInstance;
}

/**
 * Register Shiki as Monaco's tokenization provider.
 * Call this after both Shiki and Monaco are initialized.
 */
export function wireShikiToMonaco(monaco) {
  if (!highlighterInstance) {
    throw new Error('Call initShiki() before wireShikiToMonaco()');
  }
  shikiToMonaco(highlighterInstance, monaco);
}

/**
 * Load a VSCode theme into Shiki and make it available to Monaco.
 * @param {object} vsTheme - Parsed VSCode theme JSON
 * @param {string} themeId - Unique theme identifier
 */
export async function loadThemeIntoShiki(vsTheme, themeId) {
  if (!highlighterInstance) {
    throw new Error('Call initShiki() first');
  }

  // Shiki can load VS Code themes directly — same format
  await highlighterInstance.loadTheme({
    ...vsTheme,
    name: themeId,
  });
}

/**
 * Load additional languages on demand.
 * Call this when a user opens a file in a language not yet loaded.
 */
export async function loadLanguage(langId) {
  if (!highlighterInstance) return;

  const loaded = highlighterInstance.getLoadedLanguages();
  if (loaded.includes(langId)) return;

  try {
    await highlighterInstance.loadLanguage(langId);
  } catch {
    // Language not available in Shiki's grammar registry — fall back to Monaco's built-in
    console.warn(`Shiki: grammar not found for "${langId}", using Monaco default`);
  }
}
```

#### Modified MonacoEditor.svelte (key changes)

```js
// In MonacoEditor.svelte's onMount:
import { initShiki, wireShikiToMonaco, loadLanguage } from '../lib/shikiSetup.js';

onMount(async () => {
  // Initialize Shiki FIRST (loads WASM + grammars)
  await initShiki();

  // Then load Monaco
  monacoModule = await import('monaco-editor');

  // Wire Shiki as the tokenization provider
  wireShikiToMonaco(monacoModule);

  // Register built-in themes for editor colors
  if (!themeRegistered) {
    defineAllThemes(monacoModule);
    themeRegistered = true;
  }
});

// In loadFile(), before creating the editor:
async function loadFile(_fp, _rp, currentMode) {
  // ...existing code...
  const lang = getLanguage(filePath);

  // Ensure the language grammar is loaded in Shiki
  await loadLanguage(lang);

  // ...create editor as before...
}
```

#### How Shiki Themes Work with Monaco

When Shiki is wired as the tokenization provider:

1. `shikiToMonaco()` intercepts Monaco's tokenization requests
2. For each line, Shiki tokenizes using the real TextMate grammar + Oniguruma
3. The token colors come from the active Shiki theme
4. Monaco just renders — it doesn't need `defineTheme` rules for syntax

You still use `defineTheme` for **editor UI colors** (background, cursor, selection, etc.) — Shiki only handles syntax tokens.

```js
// For imported themes, you still define the editor colors:
monaco.editor.defineTheme('imported-dracula', {
  base: 'vs-dark',
  inherit: false,        // Don't inherit — Shiki handles tokens
  rules: [],             // Empty — Shiki provides all token colors
  colors: vsTheme.colors // Editor UI colors pass through
});
```

#### Performance Considerations

Shiki's WASM module adds ~1.5MB to the initial load. Mitigation strategies:

1. **Lazy load** — Don't block app startup. Load Shiki in the background after the feed view renders. The editor view can show a loading state until Shiki is ready.

2. **Cache the WASM** — The browser caches the WASM binary after the first load. Subsequent launches are fast.

3. **Limit initial grammars** — Only load grammars for Go, JS, TS, and JSON on startup. Load others on demand when the user opens a file.

```js
// Lazy loading pattern for Conductor
let shikiReady = false;

// Start loading in the background during app init
initShiki().then(() => { shikiReady = true; });

// In MonacoEditor, wait for it only when needed
async function ensureShiki() {
  if (!shikiReady) {
    await initShiki();
    shikiReady = true;
  }
}
```

---

## Phase 3: Theme Store Updates

The theme store needs to handle dynamically imported themes alongside the built-in ones.

### Modified `stores/theme.js`

```js
import { writable, derived, get } from 'svelte/store';
import { themes as builtInThemes, DEFAULT_THEME } from '../themes.js';

export { DEFAULT_THEME };

// All themes: built-in + imported
export const allThemes = writable({ ...builtInThemes });

// Re-derive themeIds whenever allThemes changes
export const themeIds = derived(allThemes, ($all) => Object.keys($all));
export const builtInThemeIds = Object.keys(builtInThemes);

export const currentThemeId = writable(DEFAULT_THEME);
export const currentTheme = derived(
  [currentThemeId, allThemes],
  ([$id, $all]) => $all[$id] || $all[DEFAULT_THEME]
);

// For backward compatibility
export const themes = builtInThemes;

/**
 * Register a dynamically imported theme.
 * @param {string} id    - Unique theme ID (e.g. "imported-dracula")
 * @param {object} theme - Conductor unified theme format { label, css, monaco, xterm }
 */
export function registerImportedTheme(id, theme) {
  allThemes.update(all => ({ ...all, [id]: theme }));
}

/**
 * Remove an imported theme. Falls back to default if it was active.
 */
export function unregisterImportedTheme(id) {
  if (builtInThemes[id]) return; // Never remove built-in themes

  allThemes.update(all => {
    const next = { ...all };
    delete next[id];
    return next;
  });

  if (get(currentThemeId) === id) {
    applyTheme(DEFAULT_THEME);
  }
}

/**
 * Apply a theme (built-in or imported).
 */
export function applyTheme(id) {
  const all = get(allThemes);
  const theme = all[id];
  if (!theme) return;

  for (const [prop, value] of Object.entries(theme.css)) {
    document.documentElement.style.setProperty(prop, value);
  }
  currentThemeId.set(id);
}
```

### Modified `monacoTheme.js`

```js
import { themes as builtInThemes } from './themes.js';

export function defineAllThemes(monaco) {
  for (const [id, theme] of Object.entries(builtInThemes)) {
    monaco.editor.defineTheme(id, theme.monaco);
  }
}

/**
 * Register a single imported theme with Monaco at runtime.
 * Call this after converting the VSCode theme.
 */
export function defineImportedTheme(monaco, id, monacoThemeData) {
  monaco.editor.defineTheme(id, monacoThemeData);
}

export const EDITOR_FONT = "'Geist Mono', 'JetBrains Mono', monospace";
```

---

## Phase 4: Settings UI — Theme Picker

Add a new section to `Settings.svelte` below the existing Theme section, showing themes discovered from the VSCodium extensions directory.

### Modified `Settings.svelte`

Add these imports and state:

```svelte
<script>
  import { createEventDispatcher, onMount } from 'svelte';
  import { ArrowLeft } from 'lucide-svelte';
  import {
    builtInThemeIds, allThemes, currentThemeId, applyTheme, registerImportedTheme
  } from '../lib/stores/theme.js';
  import { defineImportedTheme } from '../lib/monacoTheme.js';
  import { convertVSCodeThemeFull } from '../lib/themeConverter.js';
  import {
    GetConfig, SetTheme, SetVSCodiumExtPath, SetImportedTheme,
    PickDirectory, ListVSCodiumThemes, ReadThemeFile
  } from '../../wailsjs/go/main/App.js';

  const dispatch = createEventDispatcher();

  let vscodiumPath = '';
  let saveStatus = '';
  let vscodiumThemes = [];       // List of VSCodeThemeEntry from Go
  let loadingThemes = false;
  let themeLoadError = '';
  let importedThemeMap = {};     // { themePath: convertedTheme }

  onMount(async () => {
    try {
      const cfg = await GetConfig();
      vscodiumPath = cfg.vscodiumExtPath || '';

      // If a path is configured, scan for themes
      if (vscodiumPath) {
        await scanThemes();
      }

      // If there was a previously selected imported theme, re-apply it
      if (cfg.importedTheme) {
        await activateImportedTheme(cfg.importedTheme);
      }
    } catch {}
  });

  async function scanThemes() {
    loadingThemes = true;
    themeLoadError = '';
    try {
      vscodiumThemes = await ListVSCodiumThemes();
    } catch (e) {
      themeLoadError = e?.message || 'Failed to scan themes';
      vscodiumThemes = [];
    } finally {
      loadingThemes = false;
    }
  }

  async function activateImportedTheme(themePath) {
    try {
      // Check if already converted
      if (!importedThemeMap[themePath]) {
        const raw = await ReadThemeFile(themePath);
        const vsTheme = JSON.parse(raw);
        const themeId = 'imported-' + themePath.split('/').pop().replace('.json', '');
        const converted = convertVSCodeThemeFull(vsTheme, themeId);
        importedThemeMap[themePath] = { id: themeId, theme: converted };

        // Register with the store and Monaco
        registerImportedTheme(themeId, converted);

        // Register with Monaco (need access to the module)
        // This is handled reactively — see MonacoEditor.svelte integration below
      }

      const { id } = importedThemeMap[themePath];
      applyTheme(id);
      await SetTheme(id);
      await SetImportedTheme(themePath);
    } catch (e) {
      console.error('Failed to load theme:', e);
    }
  }

  async function selectBuiltInTheme(id) {
    applyTheme(id);
    try {
      await SetTheme(id);
      await SetImportedTheme(''); // Clear imported theme
    } catch {}
  }

  // ...existing browseVSCodium, saveVSCodiumPath, handleKeydown functions...

  // After changing the path, re-scan
  async function browseVSCodium() {
    try {
      const dir = await PickDirectory();
      if (dir) {
        vscodiumPath = dir;
        await SetVSCodiumExtPath(dir);
        flashSave();
        await scanThemes();
      }
    } catch {}
  }
</script>
```

Add this section to the template, after the VSCodium Extension section:

```svelte
<!-- Imported Themes section (shown when themes are found) -->
{#if vscodiumThemes.length > 0}
  <section class="settings-section">
    <h2 class="section-title">Imported Themes</h2>
    <p class="section-desc">
      Found {vscodiumThemes.length} theme{vscodiumThemes.length === 1 ? '' : 's'} in your VSCodium extensions.
    </p>
    <div class="theme-list">
      {#each vscodiumThemes as entry}
        <button
          class="imported-theme-btn"
          class:active={importedThemeMap[entry.themePath]?.id === $currentThemeId}
          on:click={() => activateImportedTheme(entry.themePath)}
        >
          <span class="theme-badge" class:dark={entry.uiTheme === 'vs-dark'} class:light={entry.uiTheme === 'vs'}>
            {entry.uiTheme === 'vs-dark' ? 'D' : 'L'}
          </span>
          <span class="imported-theme-label">{entry.label}</span>
        </button>
      {/each}
    </div>
  </section>
{:else if loadingThemes}
  <section class="settings-section">
    <p class="section-desc">Scanning for themes...</p>
  </section>
{:else if themeLoadError}
  <section class="settings-section">
    <p class="section-desc error">{themeLoadError}</p>
  </section>
{/if}
```

Add styles for the new section:

```css
/* Imported themes */
.theme-list {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.imported-theme-btn {
  display: flex;
  align-items: center;
  gap: var(--sp-sm);
  padding: 6px 10px;
  background: none;
  border: 1px solid transparent;
  border-radius: var(--radius-md);
  cursor: pointer;
  text-align: left;
  transition: all 100ms ease;
}

.imported-theme-btn:hover {
  background: var(--bg-elevated);
  border-color: var(--border-subtle);
}

.imported-theme-btn.active {
  background: var(--bg-active);
  border-color: var(--accent-green);
}

.theme-badge {
  width: 20px;
  height: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-sm);
  font-family: var(--font-mono);
  font-size: 10px;
  font-weight: 600;
  flex-shrink: 0;
}

.theme-badge.dark {
  background: #1e1e2e;
  color: #cdd6f4;
  border: 1px solid #313244;
}

.theme-badge.light {
  background: #eff1f5;
  color: #4c4f69;
  border: 1px solid #ccd0da;
}

.imported-theme-label {
  font-family: var(--font-mono);
  font-size: 12px;
  color: var(--text-primary);
}
```

---

## Phase 5: Monaco Runtime Registration

When an imported theme is activated, it needs to be registered with the live Monaco instance. The cleanest approach is to make `MonacoEditor.svelte` react to new themes in the store.

### In `MonacoEditor.svelte`, add:

```js
import { allThemes, currentThemeId } from '../lib/stores/theme.js';

// Keep track of which theme IDs we've registered with Monaco
let registeredThemeIds = new Set();

// Reactive: when currentThemeId changes, ensure the theme is registered
$: if (monacoModule && $currentThemeId) {
  const all = $allThemes; // Subscribe to allThemes too
  const theme = all[$currentThemeId];

  if (theme && !registeredThemeIds.has($currentThemeId)) {
    monacoModule.editor.defineTheme($currentThemeId, theme.monaco);
    registeredThemeIds.add($currentThemeId);
  }

  monacoModule.editor.setTheme($currentThemeId);
}
```

This replaces the existing reactive block that only called `setTheme`.

---

## Phase 6: Edge Cases and Robustness

### 6.1 Themes with `include` (Theme Inheritance)

Some VSCode themes use `"include": "./base-theme.json"` to inherit from another file. The Go backend should resolve these:

```go
// In ReadThemeFile, after reading the JSON:
// Check for "include" field and merge base theme
type rawTheme struct {
    Include     string                 `json:"include"`
    Name        string                 `json:"name"`
    Type        string                 `json:"type"`
    Colors      map[string]string      `json:"colors"`
    TokenColors []map[string]interface{} `json:"tokenColors"`
}

func (a *App) ReadThemeFileResolved(themePath string) (string, error) {
    data, err := a.ReadThemeFile(themePath)
    if err != nil {
        return "", err
    }

    var theme rawTheme
    if err := json.Unmarshal([]byte(data), &theme); err != nil {
        return data, nil // Return raw if parse fails
    }

    if theme.Include != "" {
        // Resolve relative to the theme file's directory
        baseDir := filepath.Dir(themePath)
        includePath := filepath.Join(baseDir, theme.Include)
        baseData, err := a.ReadThemeFile(includePath)
        if err == nil {
            var baseTheme rawTheme
            if err := json.Unmarshal([]byte(baseData), &baseTheme); err == nil {
                // Merge: base colors first, then overlay current theme's colors
                merged := baseTheme.Colors
                for k, v := range theme.Colors {
                    merged[k] = v
                }
                theme.Colors = merged

                // Prepend base tokenColors, then current (current wins on conflict)
                theme.TokenColors = append(baseTheme.TokenColors, theme.TokenColors...)
            }
        }
    }

    resolved, err := json.Marshal(theme)
    if err != nil {
        return data, nil
    }
    return string(resolved), nil
}
```

### 6.2 Handling Missing Colors

VSCode themes don't always define every color key. The CSS variable mapper already provides fallbacks, but some themes are minimal. Add a validation step:

```js
/**
 * Validate that a converted theme has the minimum required colors.
 * Returns true if usable, false if too many colors are missing.
 */
export function validateConvertedTheme(theme) {
  const required = ['--bg-deepest', '--bg-surface', '--text-primary', '--accent-green'];
  const missing = required.filter(key => !theme.css[key]);
  return missing.length === 0;
}
```

### 6.3 Theme Preview Thumbnails

For the built-in themes, you generate preview thumbnails inline. For imported themes, you can do the same using the converted CSS values:

```svelte
{#each vscodiumThemes as entry}
  {@const preview = importedThemeMap[entry.themePath]?.theme}
  <button class="theme-option" ...>
    {#if preview}
      <!-- Same preview card structure as built-in themes -->
      <div class="theme-preview" style="background: {preview.css['--bg-deepest']}; ...">
        <!-- ... -->
      </div>
    {:else}
      <!-- Placeholder before the theme is loaded -->
      <div class="theme-preview placeholder">
        <span>{entry.uiTheme === 'vs-dark' ? '🌙' : '☀️'}</span>
      </div>
    {/if}
    <span class="theme-name">{entry.label}</span>
  </button>
{/each}
```

### 6.4 Error Recovery

If a theme fails to load or convert, the app should never get stuck in a broken state:

```js
async function activateImportedTheme(themePath) {
  try {
    // ... conversion logic ...
    applyTheme(id);
  } catch (e) {
    console.error('Theme load failed, reverting to default:', e);
    applyTheme(DEFAULT_THEME);
    await SetImportedTheme('');
  }
}
```

---

## Decision Matrix

| Criteria                     | A: Colors Only | B: Token Converter | C: Shiki (Full) |
|------------------------------|:--------------:|:------------------:|:----------------:|
| Implementation time          | ~1 hour        | ~3-4 hours         | ~1-2 days        |
| Bundle size impact           | 0 KB           | ~5 KB              | ~1.5 MB (WASM)  |
| Editor UI color fidelity     | Exact          | Exact              | Exact            |
| Syntax highlighting fidelity | Defaults only  | ~80% match         | ~99% match       |
| Maintenance burden           | None           | Low                | Medium           |
| Dependencies added           | 0              | 0-1                | 2 (shiki + WASM) |
| Works offline                | Yes            | Yes                | Yes              |
| Startup performance impact   | None           | None               | ~200-500ms       |

**Recommendation:** Start with Approach B2 (hand-rolled mapper, no extra dependency). It gives you good syntax highlighting fidelity, zero new dependencies, and you can always upgrade to Shiki later if users want pixel-perfect rendering.

---

## Implementation Order

1. **Phase 1** — Go backend scanner (`ListVSCodiumThemes`, `ReadThemeFile`, `SetImportedTheme`)
2. **Phase 2** — Frontend converter (`themeConverter.js` with your chosen approach)
3. **Phase 3** — Theme store updates (`stores/theme.js` to support imported themes)
4. **Phase 4** — Settings UI (theme list, activation, persistence)
5. **Phase 5** — Monaco runtime registration (reactive defineTheme in MonacoEditor)
6. **Phase 6** — Edge cases (include resolution, validation, error recovery)

Each phase can be built and tested independently. Phase 1 can be verified by calling `ListVSCodiumThemes()` from the browser console via Wails. Phase 2 can be tested by hardcoding a theme JSON. Phases 3-5 wire everything together.

---

## File Changes Summary

| File | Type | Description |
|------|------|-------------|
| `app.go` | Modify | Add `VSCodeThemeEntry`, `ListVSCodiumThemes()`, `ReadThemeFile()`, `SetImportedTheme()`, update `conductorConfig` |
| `frontend/src/lib/themeConverter.js` | New | VSCode → Conductor theme conversion (approach A, B, or C) |
| `frontend/src/lib/themes.js` | No change | Built-in themes stay as-is |
| `frontend/src/lib/monacoTheme.js` | Modify | Add `defineImportedTheme()` helper |
| `frontend/src/lib/stores/theme.js` | Modify | Add `allThemes`, `registerImportedTheme()`, `unregisterImportedTheme()` |
| `frontend/src/views/Settings.svelte` | Modify | Add imported themes section, scan/activate logic |
| `frontend/src/components/MonacoEditor.svelte` | Modify | React to dynamically registered themes |
| `frontend/src/lib/shikiSetup.js` | New (C only) | Shiki initialization and Monaco wiring |
| `frontend/package.json` | Modify (B1/C) | Add `monaco-vscode-textmate-theme-converter` or `shiki` + `@shikijs/monaco` |
| `frontend/src/components/TitleBar.svelte` | Modify | Update to use `$themeIds` store and `$allThemes` instead of static arrays |
| `frontend/src/App.svelte` | Modify | Add imported theme restoration on startup |

---

## Addendum: Review Fixes (from adversarial review 2026-04-03)

The following issues were identified during spec review and MUST be addressed during implementation. See `docs/vscodium-theme-loading-review.md` for the full review.

### Critical Fixes (non-negotiable)

**Fix 1: Path traversal via symlinks (C-1)**
`ReadThemeFile` must use `filepath.EvalSymlinks` instead of `filepath.Abs`, and append `os.PathSeparator` to the prefix check to prevent `/ext-other/` matching `/ext/`. All `os.UserHomeDir()` errors must be checked, not discarded.

**Fix 2: TitleBar.svelte breakage (C-2)**
The store refactor changes `themeIds` from a plain array to a Svelte derived store. TitleBar.svelte must be updated to use `$themeIds` and `$allThemes` instead of the static `themes` and `themeIds` exports. Add TitleBar.svelte to the file changes list.

**Fix 3: JSONC theme files (C-3)**
Many VSCode themes use JSONC format (JSON with `//` comments and trailing commas). Strip comments in the Go backend before returning the string — add a JSONC-stripping pass in `ReadThemeFile` (or `ReadThemeFileResolved`). This avoids adding a frontend dependency.

**Fix 4: Size check before read (C-4)**
Move the 512KB size check to use `os.Stat` BEFORE `os.ReadFile` to prevent memory exhaustion from large files.

**Fix 5: Imported theme startup restore (C-5)**
Move imported theme restoration logic from Settings.svelte `onMount` to App.svelte (or a shared `initThemeFromConfig()` in `stores/theme.js`). Without this, imported themes reset on every app restart.

### High-Priority Fixes

**Fix 6: Wire ReadThemeFileResolved (H-1)**
Either merge include resolution into `ReadThemeFile` directly, or update `activateImportedTheme` in Phase 4 to call `ReadThemeFileResolved` instead of `ReadThemeFile`.

**Fix 7: Config race condition (H-2)**
Add mutex protection around all config load-modify-save operations to prevent concurrent Wails calls from clobbering each other.

**Fix 8: Tilde expansion (H-3)**
Change `strings.HasPrefix(extDir, "~")` to `strings.HasPrefix(extDir, "~/") || extDir == "~"` and check `os.UserHomeDir()` error.

**Fix 9: .tmTheme filter (H-7)**
Add `strings.HasSuffix(strings.ToLower(t.Path), ".json")` filter in the scanner loop to skip non-JSON theme files.

**Fix 10: Token specificity (H-5)**
Process `tokenColors` in reverse order so more-specific scopes win, or remove the `seenTokens` set entirely.

**Fix 11: Theme ID collisions (M-5)**
Include the extension ID in theme ID generation: `'imported-' + entry.extensionId + '-' + filename`.

### Implementation Notes

- **Approach B1 dropped** — The npm package `monaco-vscode-textmate-theme-converter` may not exist. Use B2 (hand-rolled) only.
- **`wails generate module` → `wails generate`** — The `module` subcommand doesn't exist in Wails v2.
- **High-contrast themes** — Map `hc-black` and `hc-light` UITheme values to their respective Monaco bases instead of defaulting to `vs-dark`.
- **`dimColor` hex normalization** — Handle 4-char (`#RGB`), 7-char (`#RRGGBB`), and 9-char (`#RRGGBBAA`) hex formats.
