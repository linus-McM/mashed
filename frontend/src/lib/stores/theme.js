import { writable, derived, get } from 'svelte/store';
import { themes as builtInThemes, DEFAULT_THEME } from '../themes.js';

export { DEFAULT_THEME };

/**
 * @typedef {import('@xterm/xterm').ITheme} XtermTheme
 * @typedef {{
 *   label: string;
 *   css: Record<string, string>;
 *   monaco?: unknown;
 *   xterm?: XtermTheme;
 * }} ThemeEntry
 * @typedef {Record<string, ThemeEntry>} ThemeMap
 */

// Static backward-compat exports
export const themes = builtInThemes;
export const builtInThemeIds = Object.keys(builtInThemes);

// Reactive map of ALL themes (built-in + imported). Typed as a loose
// `ThemeMap` so imported VSCodium themes (dynamically keyed) coexist with
// the built-in set without hitting "string can't index {...}" errors.
/** @type {import('svelte/store').Writable<ThemeMap>} */
export const allThemes = writable({ ...builtInThemes });

// Derived store of theme IDs — reactively updates when allThemes changes
export const themeIds = derived(allThemes, ($all) => Object.keys($all));

export const currentThemeId = writable(DEFAULT_THEME);

export const currentTheme = derived(
  [currentThemeId, allThemes],
  ([$id, $all]) => $all[$id] || $all[DEFAULT_THEME]
);

/** @param {string} id */
export function applyTheme(id) {
  const all = get(allThemes);
  const theme = all[id];
  if (!theme) return;
  for (const [prop, value] of Object.entries(theme.css)) {
    document.documentElement.style.setProperty(prop, value);
  }
  currentThemeId.set(id);
}

/**
 * @param {string} id
 * @param {ThemeEntry} theme
 */
export function registerImportedTheme(id, theme) {
  allThemes.update(($all) => ({ ...$all, [id]: theme }));
}

/** @param {ThemeMap} themesMap */
export function registerSavedThemes(themesMap) {
  allThemes.update(($all) => ({ ...$all, ...themesMap }));
}

/** @param {string} id */
export function unregisterImportedTheme(id) {
  // Prevent unregistering built-in themes
  if (builtInThemeIds.includes(id)) return;

  allThemes.update(($all) => {
    const next = { ...$all };
    delete next[id];
    return next;
  });

  // If the removed theme was active, fall back to default
  if (get(currentThemeId) === id) {
    applyTheme(DEFAULT_THEME);
  }
}
