// themeInit.js
// Shared module for theme activation logic used by both Settings.svelte and App.svelte.
// Handles reading, converting, registering, applying, and persisting imported VSCodium themes.

import { ReadThemeFile, SetImportedTheme, SetTheme, GetSavedThemes, SaveTheme, RemoveTheme, ListBundledThemes, ReadBundledThemeFile } from '../../wailsjs/go/main/App.js';
import { convertVSCodeTheme, validateConvertedTheme } from './themeConverter';
import { registerImportedTheme, registerSavedThemes, applyTheme, DEFAULT_THEME, allThemes } from './stores/theme.js';
import { get } from 'svelte/store';

// Cache of already-converted themes: { themePath: { id, theme } }
// Exported so Settings.svelte can render preview thumbnails for activated themes.
export const convertedCache = {};

// Activation guard: prevent double-loading on rapid clicks (M-1 fix)
let activatingPath = null;

/**
 * Extract extensionId from a theme file path.
 * Typical structure: /path/to/extensions/dracula-theme.theme-dracula-2.24.3/theme/dracula.json
 * The extension directory name is typically 2 levels up from the theme file.
 *
 * @param {string} themePath - absolute path to the theme JSON file
 * @returns {string} the extension directory name
 */
function extractExtensionId(themePath) {
  if (themePath.includes('::vsix::')) {
    const vsixPart = themePath.split('::vsix::')[0];
    const filename = vsixPart.split('/').pop();
    return filename.replace(/\.vsix$/i, '');
  }
  const parts = themePath.split('/');
  if (parts.length >= 3) {
    return parts[parts.length - 3];
  }
  return parts[parts.length - 2] || 'unknown';
}

/**
 * Generate a unique theme ID from extensionId and theme file path.
 * Includes extensionId to prevent collisions when two extensions both have e.g. dark.json (M-5 fix).
 *
 * @param {string} themePath - absolute path to the theme JSON file
 * @param {string} extensionId - extension directory name
 * @returns {string} unique theme ID
 */
export function makeThemeId(themePath, extensionId) {
  if (themePath.includes('::vsix::')) {
    const internalPath = themePath.split('::vsix::')[1];
    const filename = internalPath.split('/').pop().replace('.json', '');
    return 'imported-' + extensionId + '-' + filename;
  }
  const filename = themePath.split('/').pop().replace('.json', '');
  return 'imported-' + extensionId + '-' + filename;
}

/**
 * Load all previously-saved imported themes from ~/.mashed/themes.json
 * and register them into the allThemes store. Called once on app startup.
 *
 * @returns {Promise<void>}
 */
export async function loadSavedThemes() {
  try {
    const raw = await GetSavedThemes();
    const saved = JSON.parse(raw);
    if (saved && typeof saved === 'object' && Object.keys(saved).length > 0) {
      registerSavedThemes(saved);
    }
  } catch {
    // No saved themes or parse error — not critical
  }
}

/**
 * Remove a saved imported theme by ID. Unregisters from store and deletes from disk.
 *
 * @param {string} themeId - the theme ID to remove
 * @returns {Promise<void>}
 */
export async function removeImportedTheme(themeId) {
  const { unregisterImportedTheme } = await import('./stores/theme.js');
  unregisterImportedTheme(themeId);
  try { await RemoveTheme(themeId); } catch {}
}

/**
 * Activate an imported theme by path. Reads, converts, registers, applies, persists.
 *
 * @param {string} themePath - absolute path to the theme JSON file
 * @param {string} extensionId - extension directory name for unique ID generation
 * @returns {Promise<string>} the theme ID
 */
export async function activateImportedTheme(themePath, extensionId) {
  // Activation guard: prevent double-loading on rapid clicks
  if (activatingPath === themePath) return makeThemeId(themePath, extensionId);
  activatingPath = themePath;

  try {
    const themeId = makeThemeId(themePath, extensionId);

    // Check cache first to avoid re-converting
    if (convertedCache[themePath]) {
      const cached = convertedCache[themePath];
      registerImportedTheme(cached.id, cached.theme);
      applyTheme(cached.id);
      await Promise.all([
        SetTheme(cached.id),
        SetImportedTheme(themePath),
      ]);
      return cached.id;
    }

    // Read theme file from Go backend (strips JSONC, resolves includes)
    const raw = await ReadThemeFile(themePath);
    const vsTheme = JSON.parse(raw);

    // Convert to Mashed format
    const converted = convertVSCodeTheme(vsTheme, themeId);

    // Validate minimum CSS vars present
    if (!validateConvertedTheme(converted)) {
      throw new Error('Converted theme failed validation: missing required CSS variables');
    }

    // Cache for future use
    convertedCache[themePath] = { id: themeId, theme: converted };

    // Register in store and apply
    registerImportedTheme(themeId, converted);
    applyTheme(themeId);

    // Persist theme data + active selection
    await Promise.all([
      SaveTheme(themeId, JSON.stringify(converted)),
      SetTheme(themeId),
      SetImportedTheme(themePath),
    ]);

    return themeId;
  } catch (err) {
    console.error('Failed to activate imported theme:', err);
    // Error recovery: fall back to DEFAULT_THEME so app is never unstyled
    applyTheme(DEFAULT_THEME);
    try {
      await SetImportedTheme('');
      await SetTheme(DEFAULT_THEME);
    } catch { /* best-effort config cleanup */ }
    throw err;
  } finally {
    activatingPath = null;
  }
}

/**
 * Restore the imported theme from config on app startup.
 * Called from App.svelte onMount. If cfg.importedTheme is set, activates it.
 * Falls back to cfg.theme or DEFAULT_THEME on failure (C-5 fix).
 *
 * @param {object} cfg - the config object from GetConfig()
 */
export async function restoreImportedThemeFromConfig(cfg) {
  if (!cfg.importedTheme) {
    // No imported theme saved -- apply built-in theme
    applyTheme(cfg.theme || DEFAULT_THEME);
    return;
  }

  try {
    // Derive extensionId from the saved path
    const extensionId = extractExtensionId(cfg.importedTheme);

    await activateImportedTheme(cfg.importedTheme, extensionId);
  } catch {
    // Restoration failed (file missing, corrupt, etc.) -- fall back silently
    console.warn('Imported theme restore failed, falling back to built-in theme');
    applyTheme(cfg.theme || DEFAULT_THEME);
  }
}

/**
 * Load bundled themes from ./themes/ directory on startup.
 * Imports each theme that isn't already in the allThemes store (idempotent).
 * Individual failures are logged and skipped without aborting the batch.
 *
 * @returns {Promise<void>}
 */
export async function loadBundledThemes() {
  const entries = await ListBundledThemes();
  if (!entries || entries.length === 0) return;

  const existing = get(allThemes);
  const batch = {};

  for (const entry of entries) {
    const themeId = makeThemeId(entry.themePath, entry.extensionId);

    if (existing[themeId]) continue;

    try {
      const raw = await ReadBundledThemeFile(entry.themePath);
      const vsTheme = JSON.parse(raw);
      const converted = convertVSCodeTheme(vsTheme, themeId);

      if (!validateConvertedTheme(converted)) {
        console.warn(`Bundled theme ${entry.label} failed validation, skipping`);
        continue;
      }

      batch[themeId] = converted;
      await SaveTheme(themeId, JSON.stringify(converted));
    } catch (err) {
      console.warn(`Failed to load bundled theme ${entry.label}:`, err);
    }
  }

  if (Object.keys(batch).length > 0) {
    registerSavedThemes(batch);
  }
}
