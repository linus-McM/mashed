// themeConverter.js
// Pure transformation module: VSCode color theme JSON -> Mashed unified format.
// Zero imports from other project files. Zero npm dependencies.

// ---------------------------------------------------------------------------
// Internal: TextMate scope -> Monaco token type mapping
// ---------------------------------------------------------------------------

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

/** Background CSS variable names that must have alpha stripped. */
const BG_VARS = new Set(['--bg-deepest', '--bg-surface', '--bg-elevated', '--bg-active']);

/** Required CSS variables for validation. */
const REQUIRED_CSS_VARS = [
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
function normalizeHex(hex) {
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
function stripAlpha(hex) {
  return normalizeHex(hex);
}

/**
 * Dim a hex color by mixing it toward black.
 * factor: 0.0 = black, 1.0 = original color.
 * Handles #RGB, #RRGGBB, #RRGGBBAA via normalizeHex.
 */
function dimColor(hex, factor) {
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

/**
 * Map VSCode editor colors to Mashed's 16 CSS custom properties.
 * Falls back to sensible defaults when a color is missing.
 */
function mapVSCodeColorsToCSSVars(colors, isDark) {
  let vars;
  if (isDark) {
    vars = {
      '--bg-deepest':       colors['editor.background']                                                    || '#07080a',
      '--bg-surface':       colors['editorWidget.background']                                              || '#0d0f12',
      '--bg-elevated':      colors['sideBar.background'] || colors['editorWidget.background']              || '#12151a',
      '--bg-active':        colors['list.activeSelectionBackground'] || colors['editor.lineHighlightBackground'] || '#181c23',
      '--border-subtle':    colors['editorWidget.border'] || colors['panel.border']                        || '#1e2530',
      '--border-emphasis':  colors['focusBorder']                                                          || '#2a3340',
      '--accent-green':     colors['terminal.ansiGreen']                                                   || '#00e57a',
      '--accent-green-dim': dimColor(colors['terminal.ansiGreen'] || '#00e57a', 0.4),
      '--accent-amber':     colors['terminal.ansiYellow']                                                  || '#f0a500',
      '--accent-red':       colors['terminal.ansiRed']                                                     || '#e84545',
      '--accent-blue':      colors['terminal.ansiBlue']                                                    || '#3d9eff',
      '--accent-purple':    colors['terminal.ansiMagenta']                                                 || '#9d6fff',
      '--accent-teal':      colors['terminal.ansiCyan']                                                    || '#00c4b3',
      '--text-primary':     colors['editor.foreground']                                                    || '#c8d4e0',
      '--text-dim':         colors['editorLineNumber.foreground']                                          || '#4a5a6a',
      '--text-muted':       colors['editorWhitespace.foreground']                                          || '#2e3d4d',
    };
  } else {
    vars = {
      '--bg-deepest':       colors['editor.background']                                                    || '#f8f9fb',
      '--bg-surface':       colors['editorWidget.background']                                              || '#ffffff',
      '--bg-elevated':      colors['sideBar.background'] || colors['editorWidget.background']              || '#f0f2f5',
      '--bg-active':        colors['list.activeSelectionBackground'] || colors['editor.lineHighlightBackground'] || '#e8ecf0',
      '--border-subtle':    colors['editorWidget.border'] || colors['panel.border']                        || '#d0d7e0',
      '--border-emphasis':  colors['focusBorder']                                                          || '#b8c2cc',
      '--accent-green':     colors['terminal.ansiGreen']                                                   || '#059a50',
      '--accent-green-dim': dimColor(colors['terminal.ansiGreen'] || '#059a50', 0.3),
      '--accent-amber':     colors['terminal.ansiYellow']                                                  || '#c07800',
      '--accent-red':       colors['terminal.ansiRed']                                                     || '#cc3333',
      '--accent-blue':      colors['terminal.ansiBlue']                                                    || '#2270cc',
      '--accent-purple':    colors['terminal.ansiMagenta']                                                 || '#7044cc',
      '--accent-teal':      colors['terminal.ansiCyan']                                                    || '#0a8a7a',
      '--text-primary':     colors['editor.foreground']                                                    || '#1a1e24',
      '--text-dim':         colors['editorLineNumber.foreground']                                          || '#4a5568',
      '--text-muted':       colors['editorWhitespace.foreground']                                          || '#8896a6',
    };
  }

  // H-6: Strip alpha from background CSS vars
  for (const key of BG_VARS) {
    if (vars[key]) {
      vars[key] = stripAlpha(vars[key]);
    }
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
function convertTokenColors(tokenColors) {
  if (!tokenColors || !Array.isArray(tokenColors)) return [];

  const rules = [];
  const seenTokens = new Set();

  // Process in reverse so more-specific scopes override less-specific ones
  for (let i = tokenColors.length - 1; i >= 0; i--) {
    const entry = tokenColors[i];
    if (!entry.settings || !entry.scope) continue;
    if (!entry.settings.foreground && !entry.settings.fontStyle) continue;

    const scopes = Array.isArray(entry.scope)
      ? entry.scope
      : entry.scope.split(',').map(s => s.trim());

    for (const scope of scopes) {
      for (const mapping of SCOPE_TO_TOKEN) {
        const match = mapping.scopes.some(ms =>
          scope === ms || scope.startsWith(ms + '.')
        );

        if (match && !seenTokens.has(mapping.token)) {
          const rule = { token: mapping.token };

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
function deriveXtermTheme(colors, isDark) {
  return {
    background:          colors['terminal.background']          || colors['editor.background']     || (isDark ? '#07080a' : '#f8f9fb'),
    foreground:          colors['terminal.foreground']          || colors['editor.foreground']     || (isDark ? '#c8d4e0' : '#1a1e24'),
    cursor:              colors['terminalCursor.foreground']    || colors['editorCursor.foreground'] || '#00e57a',
    cursorAccent:        colors['terminalCursor.background']    || colors['editor.background']     || (isDark ? '#07080a' : '#f8f9fb'),
    selectionBackground: colors['terminal.selectionBackground'] || colors['editor.selectionBackground'] || '#2a3340',
    black:               colors['terminal.ansiBlack']           || (isDark ? '#07080a' : '#1a1e24'),
    red:                 colors['terminal.ansiRed']             || '#e84545',
    green:               colors['terminal.ansiGreen']           || '#00e57a',
    yellow:              colors['terminal.ansiYellow']          || '#f0a500',
    blue:                colors['terminal.ansiBlue']            || '#3d9eff',
    magenta:             colors['terminal.ansiMagenta']         || '#9d6fff',
    cyan:                colors['terminal.ansiCyan']            || '#00c4b3',
    white:               colors['terminal.ansiWhite']           || (isDark ? '#c8d4e0' : '#f8f9fb'),
  };
}

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

/**
 * Map VSCode theme type to Monaco base theme identifier.
 * @param {object} vsTheme - VSCode theme object with `type` or `uiTheme`
 * @returns {'vs' | 'vs-dark' | 'hc-black' | 'hc-light'}
 */
export function getMonacoBase(vsTheme) {
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
 *
 * @param {object} vsTheme - Parsed VSCode theme JSON (name, type/uiTheme, colors, tokenColors)
 * @param {string} themeId - Identifier for the theme (used as label fallback)
 * @returns {{ label: string, css: object, monaco: object, xterm: object }}
 */
export function convertVSCodeTheme(vsTheme, themeId) {
  const base = getMonacoBase(vsTheme);
  const isDark = base !== 'vs' && base !== 'hc-light';
  const colors = vsTheme.colors || {};

  const tokenRules = convertTokenColors(vsTheme.tokenColors);

  // Filter out null/undefined color values — VSCode themes use null to mean
  // "inherit from base theme" but Monaco requires string values.
  const cleanColors = {};
  for (const [k, v] of Object.entries(colors)) {
    if (typeof v === 'string') cleanColors[k] = v;
  }

  const monaco = {
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
 *
 * @param {object} theme - A converted theme object
 * @returns {boolean} true if valid
 */
export function validateConvertedTheme(theme) {
  if (!theme || typeof theme !== 'object') return false;
  if (!theme.css || typeof theme.css !== 'object') return false;

  for (const varName of REQUIRED_CSS_VARS) {
    if (!theme.css[varName]) return false;
  }

  return true;
}
