import { get } from 'svelte/store';
import { allThemes } from './stores/theme.js';
import { currentMonoFont } from './stores/font.js';

/** @typedef {typeof import('monaco-editor/esm/vs/editor/editor.api')} MonacoModule */
/** @typedef {import('monaco-editor').editor.IStandaloneThemeData} IStandaloneThemeData */
/** @typedef {{ monaco?: IStandaloneThemeData }} ThemeEntry */

/**
 * Sanitize a theme ID for Monaco — Monaco rejects names with dots,
 * uppercase, spaces, or parentheses.
 * @param {string} id
 * @returns {string}
 */
export function toMonacoId(id) {
  return id
    .toLowerCase()
    .replace(/[^a-z0-9-]/g, '-')
    .replace(/-+/g, '-')
    .replace(/^-|-$/g, '');
}

/**
 * @param {MonacoModule} monaco
 * @returns {string[]}
 */
export function defineAllThemes(monaco) {
  /** @type {string[]} */
  const defined = [];
  const themes = /** @type {Record<string, ThemeEntry>} */ (/** @type {unknown} */ (get(allThemes)));
  for (const [id, theme] of Object.entries(themes)) {
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

/**
 * @param {MonacoModule} monaco
 * @param {string} id
 * @param {IStandaloneThemeData} monacoThemeData
 */
export function defineImportedTheme(monaco, id, monacoThemeData) {
    monaco.editor.defineTheme(id, monacoThemeData);
}

/** @returns {string} */
export function getEditorFont() {
  return get(currentMonoFont);
}
