import { themes } from './themes.js';

export function defineAllThemes(monaco) {
  for (const [id, theme] of Object.entries(themes)) {
    monaco.editor.defineTheme(id, theme.monaco);
  }
}

export function defineImportedTheme(monaco, id, monacoThemeData) {
    monaco.editor.defineTheme(id, monacoThemeData);
}

import { get } from 'svelte/store';
import { currentMonoFont } from './stores/font.js';

export function getEditorFont() {
  return get(currentMonoFont);
}
