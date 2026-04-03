// themeInit.js
// Shared module for theme activation logic used by both Settings.svelte and App.svelte.
// Handles reading, converting, registering, applying, and persisting imported VSCodium themes.

import { ReadThemeFile, SetImportedTheme, SetTheme } from '../../wailsjs/go/main/App.js';
import { convertVSCodeTheme, validateConvertedTheme } from './themeConverter.js';
import { registerImportedTheme, applyTheme, DEFAULT_THEME } from './stores/theme.js';

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
  const parts = themePath.split('/');
  // The extension dir is typically the grandparent directory of the theme file
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
function makeThemeId(themePath, extensionId) {
  const filename = themePath.split('/').pop().replace('.json', '');
  return 'imported-' + extensionId + '-' + filename;
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

    // Convert to Conductor format
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

    // Persist to config
    await Promise.all([
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
