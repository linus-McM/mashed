// types/theme.ts — theme shapes used by `themeConverter.ts` and consumers.
//
// Themes are NOT Go-sourced — they come from VSCode JSON on disk and are
// converted to the Mashed unified format by `lib/themeConverter.ts`. This
// module types the inputs (VSCode JSON) and the output (`Theme`) so both
// the converter and its call sites can lose their `any`s.

// ---------------------------------------------------------------------------
// VSCode theme JSON (input)
// ---------------------------------------------------------------------------

/**
 * A single entry in a VSCode theme's `tokenColors` array. Every field is
 * optional because real VSCode themes routinely omit `scope` (globals) or
 * `settings.foreground` (fontStyle-only rules). The converter must not
 * crash on partial entries — see themeConverter.test.ts scenarios.
 */
export interface VSCodeTokenColor {
  scope?: string | string[];
  settings?: {
    foreground?: string;
    background?: string;
    fontStyle?: string;
  };
}

/**
 * A parsed VSCode color theme JSON document. Only the fields the converter
 * reads are typed — VSCode themes carry more keys (e.g. `semanticHighlighting`)
 * that we don't consume.
 */
export interface VSCodeTheme {
  name?: string;
  type?: string;
  uiTheme?: string;
  colors?: Record<string, string | null | undefined>;
  tokenColors?: VSCodeTokenColor[];
}

// ---------------------------------------------------------------------------
// Monaco theme (output)
// ---------------------------------------------------------------------------

/** Valid Monaco base theme identifiers. */
export type MonacoBase = 'vs' | 'vs-dark' | 'hc-black' | 'hc-light';

/** A single Monaco `ITokenThemeRule`. `foreground` is bare hex (no `#`). */
export interface MonacoTokenRule {
  token: string;
  foreground?: string;
  background?: string;
  fontStyle?: string;
}

/** A Monaco theme definition (`monaco.editor.IStandaloneThemeData`). */
export interface MonacoThemeDef {
  base: MonacoBase;
  inherit: boolean;
  rules: MonacoTokenRule[];
  colors: Record<string, string>;
}

// ---------------------------------------------------------------------------
// xterm.js theme (output)
// ---------------------------------------------------------------------------

/** An xterm.js `ITheme` — every field a hex color string. */
export interface XtermTheme {
  background: string;
  foreground: string;
  cursor: string;
  cursorAccent: string;
  selectionBackground: string;
  black: string;
  red: string;
  green: string;
  yellow: string;
  blue: string;
  magenta: string;
  cyan: string;
  white: string;
}

// ---------------------------------------------------------------------------
// Unified Mashed theme (output)
// ---------------------------------------------------------------------------

/** Keys of the 16 canonical Mashed CSS custom properties. */
export type ThemeColorKey =
  | '--bg-deepest'
  | '--bg-surface'
  | '--bg-elevated'
  | '--bg-active'
  | '--border-subtle'
  | '--border-emphasis'
  | '--accent-green'
  | '--accent-green-dim'
  | '--accent-amber'
  | '--accent-red'
  | '--accent-blue'
  | '--accent-purple'
  | '--accent-teal'
  | '--text-primary'
  | '--text-dim'
  | '--text-muted';

/** Map of Mashed CSS variable name -> hex color. */
export type ThemeColors = Record<ThemeColorKey, string>;

/**
 * Unified theme shape consumed by the app (Monaco + xterm + CSS vars).
 * Emitted by `convertVSCodeTheme` and validated by `validateConvertedTheme`.
 */
export interface Theme {
  label: string;
  css: ThemeColors;
  monaco: MonacoThemeDef;
  xterm: XtermTheme;
}
