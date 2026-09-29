import { writable, get } from 'svelte/store';
import { SetEditorSettings } from '../../../wailsjs/go/main/App.js';

/** @typedef {import('../types/wails').EditorSettings} EditorSettings */

/** @type {EditorSettings} */
const defaults = {
  minimapEnabled: false,
  wordWrap: "off",
  lineNumbers: "on",
  renderWhitespace: "none",
  tabSize: 2,
  insertSpaces: true,
  cursorStyle: "line",
  cursorBlinking: "blink",
  bracketPairColorization: true,
  renderLineHighlight: "line",
  fontLigatures: false,
  scrollBeyondLastLine: false,
  smoothScrolling: false,
};

export const editorSettings = writable(/** @type {EditorSettings} */ ({ ...defaults }));

/**
 * Merge backend settings into the store, falling back to defaults per field.
 * @param {Partial<EditorSettings> | null | undefined} settings
 */
export function initEditorSettings(settings) {
  /** @type {EditorSettings} */
  const merged = { ...defaults };
  if (settings) {
    const source = /** @type {Record<string, unknown>} */ (/** @type {unknown} */ (settings));
    const target = /** @type {Record<string, unknown>} */ (/** @type {unknown} */ (merged));
    for (const key of /** @type {(keyof EditorSettings)[]} */ (Object.keys(defaults))) {
      const value = source[key];
      if (value !== undefined && value !== null) {
        target[key] = value;
      }
    }
  }
  editorSettings.set(merged);
}

/**
 * Update a single setting, persist the full object to the Go backend.
 * @template {keyof EditorSettings} K
 * @param {K} key
 * @param {EditorSettings[K]} value
 */
export async function updateEditorSetting(key, value) {
  editorSettings.update(current => ({ ...current, [key]: value }));
  await SetEditorSettings(get(editorSettings));
}
