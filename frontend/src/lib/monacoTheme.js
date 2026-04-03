import { themes } from './themes.js';

export function defineAllThemes(monaco) {
  for (const [id, theme] of Object.entries(themes)) {
    monaco.editor.defineTheme(id, theme.monaco);
  }
}

export const EDITOR_FONT = "'Geist Mono', 'JetBrains Mono', monospace";
