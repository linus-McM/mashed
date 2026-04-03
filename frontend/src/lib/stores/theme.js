import { writable, derived } from 'svelte/store';
import { themes, DEFAULT_THEME } from '../themes.js';

export { DEFAULT_THEME, themes };
export const themeIds = Object.keys(themes);
export const currentThemeId = writable(DEFAULT_THEME);
export const currentTheme = derived(currentThemeId, ($id) => themes[$id] || themes[DEFAULT_THEME]);

export function applyTheme(id) {
  const theme = themes[id];
  if (!theme) return;
  for (const [prop, value] of Object.entries(theme.css)) {
    document.documentElement.style.setProperty(prop, value);
  }
  currentThemeId.set(id);
}
