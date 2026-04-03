import { writable, get } from 'svelte/store';

export const DEFAULT_MONO_FONT = "'JetBrains Mono', monospace";
export const DEFAULT_FONT_SIZE = 13;

export const currentMonoFont = writable(DEFAULT_MONO_FONT);
export const currentFontSize = writable(DEFAULT_FONT_SIZE);

/**
 * Apply a font to all consumers. This is the single source of truth —
 * CSS variables, Terminal, and Monaco all derive from these stores.
 * Always called on startup, even with no user selection.
 */
export function applyFont(fontFamily, fontSize) {
  const family = fontFamily
    ? `'${fontFamily}', 'JetBrains Mono', monospace`
    : DEFAULT_MONO_FONT;
  const size = fontSize || DEFAULT_FONT_SIZE;

  document.documentElement.style.setProperty('--font-mono', family);
  document.documentElement.style.setProperty('--font-terminal', family);

  currentMonoFont.set(family);
  currentFontSize.set(size);
}

/**
 * Register local fonts by injecting @font-face rules into the document.
 * Must be called before applyFont() so local fonts are available to CSS.
 * Safe to call multiple times (replaces previous registration).
 */
export function registerLocalFonts(localFamilies) {
  const prev = document.getElementById('local-fonts');
  if (prev) prev.remove();

  if (!localFamilies || localFamilies.length === 0) return;

  const rules = [];
  for (const family of localFamilies) {
    for (const file of family.files) {
      rules.push(`@font-face {
  font-family: '${family.family}';
  src: url(data:font/${file.format};base64,${file.base64}) format('${file.format}');
  font-weight: ${file.weight};
  font-style: ${file.style};
}`);
    }
  }

  const style = document.createElement('style');
  style.id = 'local-fonts';
  style.textContent = rules.join('\n');
  document.head.appendChild(style);
}
