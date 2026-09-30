// themeConverter.ts
// Pure transformation module: VSCode color theme JSON -> Mashed unified format.
// Zero imports from other project files. Zero npm dependencies.

import type {
  MonacoBase,
  MonacoThemeDef,
  MonacoTokenRule,
  Theme,
  ThemeColorKey,
  ThemeColors,
  VSCodeTheme,
  VSCodeTokenColor,
  XtermTheme,
} from '../types/theme';

// ---------------------------------------------------------------------------
// Internal: TextMate scope -> Monaco token type mapping
// ---------------------------------------------------------------------------

interface ScopeMapping {
  scopes: string[];
  token: string;
}

const SCOPE_TO_TOKEN: ScopeMapping[] = [
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

/** Background CSS variable names that must have alpha stripped. */
const BG_VARS: ReadonlySet<ThemeColorKey> = new Set<ThemeColorKey>([
  '--bg-deepest',
  '--bg-surface',
  '--bg-elevated',
  '--bg-active',
]);

/** Required CSS variables for validation. */
const REQUIRED_CSS_VARS: readonly ThemeColorKey[] = [
  '--bg-deepest', '--bg-surface', '--bg-elevated', '--bg-active',
  '--border-subtle', '--border-emphasis',
  '--accent-green', '--accent-green-dim',
  '--accent-amber', '--accent-red', '--accent-blue', '--accent-purple', '--accent-teal',
  '--text-primary', '--text-dim', '--text-muted',
];

// ---------------------------------------------------------------------------
// Internal: Color utilities
// ---------------------------------------------------------------------------

/**
 * Normalize hex to #RRGGBB.
 * Handles #RGB (4-char), #RRGGBB (7-char), #RRGGBBAA (9-char).
 */
function normalizeHex(hex: string): string {
  if (!hex || hex[0] !== '#') return hex;
  // #RGB -> #RRGGBB
  if (hex.length === 4) {
    return '#' + hex[1] + hex[1] + hex[2] + hex[2] + hex[3] + hex[3];
  }
  // #RRGGBBAA -> #RRGGBB (strip alpha)
  if (hex.length === 9) {
    return hex.slice(0, 7);
  }
  return hex;
}

/**
 * Strip alpha channel from a hex color, returning #RRGGBB.
 * Normalizes short hex first.
 */
function stripAlpha(hex: string): string {
  return normalizeHex(hex);
}

/**
 * Dim a hex color by mixing it toward black.
 * factor: 0.0 = black, 1.0 = original color.
 * Handles #RGB, #RRGGBB, #RRGGBBAA via normalizeHex.
 */
function dimColor(hex: string, factor: number): string {
  const normalized = normalizeHex(hex);
  const r = parseInt(normalized.slice(1, 3), 16);
  const g = parseInt(normalized.slice(3, 5), 16);
  const b = parseInt(normalized.slice(5, 7), 16);
  const dr = Math.round(r * factor);
  const dg = Math.round(g * factor);
  const db = Math.round(b * factor);
  return '#' + dr.toString(16).padStart(2, '0') + dg.toString(16).padStart(2, '0') + db.toString(16).padStart(2, '0');
}

// ---------------------------------------------------------------------------
// Internal: CSS variable mapping
// ---------------------------------------------------------------------------

type RawColors = Record<string, string | null | undefined>;

/** Safely read a string color from the raw VSCode `colors` map. */
function colorString(colors: RawColors, key: string): string | undefined {
  const v = colors[key];
  return typeof v === 'string' ? v : undefined;
}

/**
 * Map VSCode editor colors to Mashed's 16 CSS custom properties.
 * Falls back to sensible defaults when a color is missing.
 */
function mapVSCodeColorsToCSSVars(colors: RawColors, isDark: boolean): ThemeColors {
  const greenFallback = isDark ? '#00e57a' : '#059a50';
  const dimFactor = isDark ? 0.4 : 0.3;

  const vars: ThemeColors = isDark
    ? {
        '--bg-deepest':       colorString(colors, 'editor.background')                                                             ?? '#07080a',
        '--bg-surface':       colorString(colors, 'editorWidget.background')                                                       ?? '#0d0f12',
        '--bg-elevated':      colorString(colors, 'sideBar.background') ?? colorString(colors, 'editorWidget.background')          ?? '#12151a',
        '--bg-active':        colorString(colors, 'list.activeSelectionBackground') ?? colorString(colors, 'editor.lineHighlightBackground') ?? '#181c23',
        '--border-subtle':    colorString(colors, 'editorWidget.border') ?? colorString(colors, 'panel.border')                    ?? '#1e2530',
        '--border-emphasis':  colorString(colors, 'focusBorder')                                                                   ?? '#2a3340',
        '--accent-green':     colorString(colors, 'terminal.ansiGreen')                                                            ?? greenFallback,
        '--accent-green-dim': dimColor(colorString(colors, 'terminal.ansiGreen') ?? greenFallback, dimFactor),
        '--accent-amber':     colorString(colors, 'terminal.ansiYellow')                                                           ?? '#f0a500',
        '--accent-red':       colorString(colors, 'terminal.ansiRed')                                                              ?? '#e84545',
        '--accent-blue':      colorString(colors, 'terminal.ansiBlue')                                                             ?? '#3d9eff',
        '--accent-purple':    colorString(colors, 'terminal.ansiMagenta')                                                          ?? '#9d6fff',
        '--accent-teal':      colorString(colors, 'terminal.ansiCyan')                                                             ?? '#00c4b3',
        '--text-primary':     colorString(colors, 'editor.foreground')                                                             ?? '#c8d4e0',
        '--text-dim':         colorString(colors, 'editorLineNumber.foreground')                                                   ?? '#4a5a6a',
        '--text-muted':       colorString(colors, 'editorWhitespace.foreground')                                                   ?? '#2e3d4d',
      }
    : {
        '--bg-deepest':       colorString(colors, 'editor.background')                                                             ?? '#f8f9fb',
        '--bg-surface':       colorString(colors, 'editorWidget.background')                                                       ?? '#ffffff',
        '--bg-elevated':      colorString(colors, 'sideBar.background') ?? colorString(colors, 'editorWidget.background')          ?? '#f0f2f5',
        '--bg-active':        colorString(colors, 'list.activeSelectionBackground') ?? colorString(colors, 'editor.lineHighlightBackground') ?? '#e8ecf0',
        '--border-subtle':    colorString(colors, 'editorWidget.border') ?? colorString(colors, 'panel.border')                    ?? '#d0d7e0',
        '--border-emphasis':  colorString(colors, 'focusBorder')                                                                   ?? '#b8c2cc',
        '--accent-green':     colorString(colors, 'terminal.ansiGreen')                                                            ?? greenFallback,
        '--accent-green-dim': dimColor(colorString(colors, 'terminal.ansiGreen') ?? greenFallback, dimFactor),
        '--accent-amber':     colorString(colors, 'terminal.ansiYellow')                                                           ?? '#c07800',
        '--accent-red':       colorString(colors, 'terminal.ansiRed')                                                              ?? '#cc3333',
        '--accent-blue':      colorString(colors, 'terminal.ansiBlue')                                                             ?? '#2270cc',
        '--accent-purple':    colorString(colors, 'terminal.ansiMagenta')                                                          ?? '#7044cc',
        '--accent-teal':      colorString(colors, 'terminal.ansiCyan')                                                             ?? '#0a8a7a',
        '--text-primary':     colorString(colors, 'editor.foreground')                                                             ?? '#1a1e24',
        '--text-dim':         colorString(colors, 'editorLineNumber.foreground')                                                   ?? '#4a5568',
        '--text-muted':       colorString(colors, 'editorWhitespace.foreground')                                                   ?? '#8896a6',
      };

  // H-6: Strip alpha from background CSS vars
  for (const key of BG_VARS) {
    vars[key] = stripAlpha(vars[key]);
  }

  return vars;
}

// ---------------------------------------------------------------------------
// Internal: Token color conversion
// ---------------------------------------------------------------------------

/**
 * Convert VSCode tokenColors into Monaco ITokenThemeRule[].
 * Processes in REVERSE order so more-specific scopes win (H-5 fix).
 */
function convertTokenColors(tokenColors: VSCodeTokenColor[] | undefined): MonacoTokenRule[] {
  if (!tokenColors || !Array.isArray(tokenColors)) return [];

  const rules: MonacoTokenRule[] = [];
  const seenTokens = new Set<string>();

  // Process in reverse so more-specific scopes override less-specific ones
  for (let i = tokenColors.length - 1; i >= 0; i--) {
    const entry = tokenColors[i];
    if (!entry.settings || !entry.scope) continue;
    if (!entry.settings.foreground && !entry.settings.fontStyle) continue;

    const scopes = Array.isArray(entry.scope)
      ? entry.scope
      : entry.scope.split(',').map((s) => s.trim());

    for (const scope of scopes) {
      for (const mapping of SCOPE_TO_TOKEN) {
        const match = mapping.scopes.some((ms) =>
          scope === ms || scope.startsWith(ms + '.')
        );

        if (match && !seenTokens.has(mapping.token)) {
          const rule: MonacoTokenRule = { token: mapping.token };

          if (entry.settings.foreground) {
            // Strip # prefix -- Monaco expects bare hex
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

// ---------------------------------------------------------------------------
// Internal: xterm theme derivation
// ---------------------------------------------------------------------------

/**
 * Derive xterm.js theme from VSCode terminal colors.
 */
function deriveXtermTheme(colors: RawColors, isDark: boolean): XtermTheme {
  return {
    background:          colorString(colors, 'terminal.background')          ?? colorString(colors, 'editor.background')     ?? (isDark ? '#07080a' : '#f8f9fb'),
    foreground:          colorString(colors, 'terminal.foreground')          ?? colorString(colors, 'editor.foreground')     ?? (isDark ? '#c8d4e0' : '#1a1e24'),
    cursor:              colorString(colors, 'terminalCursor.foreground')    ?? colorString(colors, 'editorCursor.foreground') ?? '#00e57a',
    cursorAccent:        colorString(colors, 'terminalCursor.background')    ?? colorString(colors, 'editor.background')     ?? (isDark ? '#07080a' : '#f8f9fb'),
    selectionBackground: colorString(colors, 'terminal.selectionBackground') ?? colorString(colors, 'editor.selectionBackground') ?? '#2a3340',
    black:               colorString(colors, 'terminal.ansiBlack')           ?? (isDark ? '#07080a' : '#1a1e24'),
    red:                 colorString(colors, 'terminal.ansiRed')             ?? '#e84545',
    green:               colorString(colors, 'terminal.ansiGreen')           ?? '#00e57a',
    yellow:              colorString(colors, 'terminal.ansiYellow')          ?? '#f0a500',
    blue:                colorString(colors, 'terminal.ansiBlue')            ?? '#3d9eff',
    magenta:             colorString(colors, 'terminal.ansiMagenta')         ?? '#9d6fff',
    cyan:                colorString(colors, 'terminal.ansiCyan')            ?? '#00c4b3',
    white:               colorString(colors, 'terminal.ansiWhite')           ?? (isDark ? '#c8d4e0' : '#f8f9fb'),
  };
}

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

/**
 * Map VSCode theme type to Monaco base theme identifier.
 */
export function getMonacoBase(vsTheme: VSCodeTheme): MonacoBase {
  const type = vsTheme.type || vsTheme.uiTheme || '';
  switch (type) {
    case 'light':
    case 'vs':
      return 'vs';
    case 'hc-black':
      return 'hc-black';
    case 'hc-light':
      return 'hc-light';
    default:
      return 'vs-dark';
  }
}

/**
 * Convert a VSCode color theme JSON object into Mashed's unified theme format.
 */
export function convertVSCodeTheme(vsTheme: VSCodeTheme, themeId: string): Theme {
  const base = getMonacoBase(vsTheme);
  const isDark = base !== 'vs' && base !== 'hc-light';
  const colors: RawColors = vsTheme.colors ?? {};

  const tokenRules = convertTokenColors(vsTheme.tokenColors);

  // Filter out null/undefined color values — VSCode themes use null to mean
  // "inherit from base theme" but Monaco requires string values.
  const cleanColors: Record<string, string> = {};
  for (const [k, v] of Object.entries(colors)) {
    if (typeof v === 'string') cleanColors[k] = v;
  }

  const monaco: MonacoThemeDef = {
    base,
    inherit: true,
    rules: tokenRules,
    colors: cleanColors,
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

/**
 * Validate that a converted theme has the minimum required CSS variables.
 * Accepts `unknown` so call sites reading arbitrary JSON payloads don't need
 * to pre-narrow.
 */
export function validateConvertedTheme(theme: unknown): theme is Theme {
  if (!theme || typeof theme !== 'object') return false;
  const maybe = theme as { css?: unknown };
  if (!maybe.css || typeof maybe.css !== 'object') return false;

  const css = maybe.css as Record<string, unknown>;
  for (const varName of REQUIRED_CSS_VARS) {
    if (typeof css[varName] !== 'string' || !css[varName]) return false;
  }

  return true;
}
