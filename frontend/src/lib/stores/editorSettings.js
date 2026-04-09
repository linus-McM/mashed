import { writable, get } from 'svelte/store';
import { SetEditorSettings } from '../../../wailsjs/go/main/App.js';

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

export const editorSettings = writable({ ...defaults });

/**
 * Merge backend settings into the store, falling back to defaults per field.
 */
export function initEditorSettings(settings) {
  const merged = { ...defaults };
  if (settings) {
    for (const key of Object.keys(defaults)) {
      if (settings[key] !== undefined && settings[key] !== null) {
        merged[key] = settings[key];
      }
    }
  }
  editorSettings.set(merged);
}

/**
 * Update a single setting, persist the full object to the Go backend.
 */
export async function updateEditorSetting(key, value) {
  editorSettings.update(current => ({ ...current, [key]: value }));
  await SetEditorSettings(get(editorSettings));
}
