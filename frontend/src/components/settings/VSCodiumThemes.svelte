<script lang="ts">
  // VSCodium theme-extension panel (extension path, scan & import list).
  // Extracted from views/Settings.svelte (spec R31) — markup and scoped
  // styles moved verbatim; the shared panel/section styles are duplicated
  // here because Svelte scopes them per component.
  import { allThemes } from '../../lib/stores/theme.js';
  import { SetVSCodiumExtPath, PickDirectory, ListVSCodiumThemes } from '../../../wailsjs/go/main/App.js';
  import { activateImportedTheme, makeThemeId } from '../../lib/themeInit.js';
  import { errorMessage } from '../../lib/errorMessage';
  import type { VSCodeThemeEntry } from '../../lib/types/wails';

  /** Structural shape of an `allThemes` entry (see views/Settings.svelte). */
  type ThemeMap = Record<string, { label: string; css: Record<string, string> }>;

  /** Extension directory; bound by the parent, which loads it from config. */
  export let vscodiumPath = '';
  /** Transient "Saved" flash text, owned by the parent. */
  export let saveStatus = '';
  /** Called after a successful save so the parent can flash `saveStatus`. */
  export let onSaved: () => void = () => {};

  let vscodiumThemes: VSCodeThemeEntry[] = [];
  let loadingThemes = false;
  let themeLoadError = '';
  let activatingThemePath = '';

  $: themes = $allThemes as unknown as ThemeMap;

  /** Called by the parent's onMount when a path is configured. */
  export async function scanThemes(): Promise<void> {
    loadingThemes = true;
    themeLoadError = '';
    try {
      vscodiumThemes = await ListVSCodiumThemes();
    } catch (err) {
      themeLoadError = errorMessage(err) || 'Failed to scan themes';
      vscodiumThemes = [];
    } finally {
      loadingThemes = false;
    }
  }

  async function handleImportedThemeClick(entry: VSCodeThemeEntry): Promise<void> {
    activatingThemePath = entry.themePath;
    try {
      await activateImportedTheme(entry.themePath, entry.extensionId);
    } catch (err) {
      themeLoadError = 'Failed to activate theme: ' + (errorMessage(err) || 'unknown error');
    } finally {
      activatingThemePath = '';
    }
  }

  function isDarkTheme(uiTheme: string): boolean {
    return uiTheme !== 'vs' && uiTheme !== 'vs-light';
  }

  async function browseVSCodium() {
    try {
      const dir = await PickDirectory();
      if (dir) {
        vscodiumPath = dir;
        await SetVSCodiumExtPath(dir);
        onSaved();
        await scanThemes();
      }
    } catch {}
  }

  async function saveVSCodiumPath() {
    try {
      await SetVSCodiumExtPath(vscodiumPath);
      onSaved();
    } catch {}
  }
</script>

<!-- VSCodium Extension path -->
<div class="settings-panel">
<section class="settings-section">
  <h2 class="section-title">Theme Extensions</h2>
  <p class="section-desc">Path to .vsix theme files.</p>
  <div class="path-input-row">
    <input
      class="path-input"
      type="text"
      bind:value={vscodiumPath}
      placeholder="e.g. ~/.vscode-oss/extensions"
      on:blur={saveVSCodiumPath}
    />
    <button class="browse-btn" on:click={browseVSCodium}>Browse</button>
  </div>
  {#if saveStatus}
    <span class="save-status">{saveStatus}</span>
  {/if}

  {#if vscodiumPath}
    <div class="import-row">
      <button class="browse-btn" on:click={scanThemes} disabled={loadingThemes}>
        {loadingThemes ? 'Scanning...' : 'Scan & Import'}
      </button>
      {#if themeLoadError}
        <span class="error-text">{themeLoadError}</span>
      {/if}
    </div>
    {#if vscodiumThemes.length > 0}
      <div class="import-list">
        {#each vscodiumThemes as entry}
          {@const themeId = makeThemeId(entry.themePath, entry.extensionId)}
          {@const alreadyImported = !!themes[themeId]}
          <button
            class="import-item"
            class:imported={alreadyImported}
            disabled={activatingThemePath === entry.themePath}
            on:click={() => handleImportedThemeClick(entry)}
          >
            <span class="import-indicator" class:dark={isDarkTheme(entry.uiTheme)} class:light={!isDarkTheme(entry.uiTheme)} />
            <span class="import-label">{entry.label}</span>
            {#if activatingThemePath === entry.themePath}
              <span class="activating-indicator">...</span>
            {:else if alreadyImported}
              <span class="imported-check">&#10003;</span>
            {/if}
          </button>
        {/each}
      </div>
    {/if}
  {/if}
</section>
</div>

<style>
  .settings-panel {
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    padding: var(--sp-lg);
  }

  /* Panel padding owns bottom spacing; neutralise the legacy section margin. */
  .settings-panel .settings-section {
    margin-bottom: 0;
  }

  .section-title {
    font-family: var(--font-mono);
    font-size: var(--text-data);
    font-weight: 600;
    color: var(--text-primary);
    margin-bottom: var(--sp-md);
  }

  .section-desc {
    font-size: var(--text-body);
    color: var(--text-dim);
    margin-bottom: var(--sp-md);
  }

  .settings-section {
    margin-bottom: var(--sp-2xl);
  }

  /* VSCodium / import */
  .path-input-row {
    display: flex;
    gap: var(--sp-sm);
  }

  .path-input {
    flex: 1;
    padding: var(--sp-xs) 10px;
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    color: var(--text-primary);
    font-family: var(--font-mono);
    font-size: var(--text-body);
    outline: none;
  }

  .path-input:focus {
    border-color: var(--accent-green);
  }

  .path-input::placeholder {
    color: var(--text-muted);
  }

  .browse-btn {
    padding: var(--sp-xs) 14px;
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    color: var(--text-dim);
    font-family: var(--font-mono);
    font-size: var(--text-body);
    cursor: pointer;
    transition: all 100ms ease;
    white-space: nowrap;
  }

  .browse-btn:hover {
    color: var(--text-primary);
    border-color: var(--border-emphasis);
    background: var(--bg-active);
  }

  .save-status {
    display: inline-block;
    margin-top: var(--sp-xs);
    font-family: var(--font-mono);
    font-size: var(--text-label);
    color: var(--accent-green);
  }

  .import-row {
    display: flex;
    align-items: center;
    gap: var(--sp-sm);
    margin-top: var(--sp-md);
  }

  .import-list {
    display: flex;
    flex-direction: column;
    gap: 1px;
    margin-top: var(--sp-sm);
    max-height: 200px;
    overflow-y: auto;
  }

  .import-item {
    display: flex;
    align-items: center;
    gap: var(--sp-sm);
    padding: 4px 8px;
    background: none;
    border: 1px solid transparent;
    border-radius: var(--radius-sm);
    cursor: pointer;
    text-align: left;
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--text-dim);
    transition: all 100ms ease;
  }

  .import-item:hover {
    background: var(--bg-elevated);
    color: var(--text-primary);
  }

  .import-item.imported {
    color: var(--text-muted);
  }

  .import-item:disabled {
    opacity: 0.6;
    cursor: wait;
  }

  .import-indicator {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    flex-shrink: 0;
  }

  .import-indicator.dark {
    background: var(--text-muted);
  }

  .import-indicator.light {
    background: var(--text-dim);
  }

  .import-label {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .activating-indicator {
    color: var(--text-muted);
    margin-left: auto;
  }

  .imported-check {
    color: var(--accent-green);
    margin-left: auto;
    font-size: var(--text-label);
  }

  .error-text {
    color: var(--accent-red);
    font-size: 11px;
  }
</style>
