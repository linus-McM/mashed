import { get } from 'svelte/store';
import { allThemes } from './stores/theme.js';
import { currentMonoFont } from './stores/font.js';

/**
 * Sanitize a theme ID for Monaco — Monaco rejects names with dots,
 * uppercase, spaces, or parentheses.
 */
export function toMonacoId(id) {
  return id
    .toLowerCase()
    .replace(/[^a-z0-9-]/g, '-')
    .replace(/-+/g, '-')
    .replace(/^-|-$/g, '');
}

export function defineAllThemes(monaco) {
  const defined = [];
  for (const [id, theme] of Object.entries(get(allThemes))) {
    if (theme.monaco) {
      try {
        monaco.editor.defineTheme(toMonacoId(id), theme.monaco);
        defined.push(id);
      } catch (e) {
        console.warn('Failed to define Monaco theme', id, e);
      }
    }
  }
  return defined;
}

export function defineImportedTheme(monaco, id, monacoThemeData) {
    monaco.editor.defineTheme(id, monacoThemeData);
}

export function getEditorFont() {
  return get(currentMonoFont);
}
