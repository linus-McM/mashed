<script>
  import { createEventDispatcher, onMount } from 'svelte';
  import { ArrowLeft } from 'lucide-svelte';
  import { allThemes, themeIds, currentThemeId, applyTheme } from '../lib/stores/theme.js';
  import { GetConfig, SetTheme, SetImportedTheme, SetVSCodiumExtPath, PickDirectory, ListVSCodiumThemes } from '../../wailsjs/go/main/App.js';
  import { activateImportedTheme, convertedCache } from '../lib/themeInit.js';

  const dispatch = createEventDispatcher();

  let vscodiumPath = '';
  let saveStatus = '';
  let vscodiumThemes = [];
  let loadingThemes = false;
  let themeLoadError = '';
  let activatingThemePath = '';

  onMount(async () => {
    try {
      const cfg = await GetConfig();
      vscodiumPath = cfg.vscodiumExtPath || '';
      if (vscodiumPath) {
        await scanThemes();
      }
    } catch {}
  });

  async function selectTheme(id) {
    applyTheme(id);
    try {
      await SetTheme(id);
      // AC-4: selecting a built-in theme clears the imported theme from config
      await SetImportedTheme('');
    } catch {}
  }

  async function scanThemes() {
    loadingThemes = true;
    themeLoadError = '';
    try {
      vscodiumThemes = await ListVSCodiumThemes();
    } catch (err) {
      themeLoadError = err?.message || 'Failed to scan themes';
      vscodiumThemes = [];
    } finally {
      loadingThemes = false;
    }
  }

  async function handleImportedThemeClick(entry) {
    activatingThemePath = entry.themePath;
    try {
      await activateImportedTheme(entry.themePath, entry.extensionId);
    } catch (err) {
      themeLoadError = 'Failed to activate theme: ' + (err?.message || 'unknown error');
    } finally {
      activatingThemePath = '';
    }
  }

  function isDarkTheme(uiTheme) {
    return uiTheme !== 'vs' && uiTheme !== 'vs-light';
  }

  async function browseVSCodium() {
    try {
      const dir = await PickDirectory();
      if (dir) {
        vscodiumPath = dir;
        await SetVSCodiumExtPath(dir);
        flashSave();
        // AC-7: re-scan themes after path change
        await scanThemes();
      }
    } catch {}
  }

  async function saveVSCodiumPath() {
    try {
      await SetVSCodiumExtPath(vscodiumPath);
      flashSave();
    } catch {}
  }

  function flashSave() {
    saveStatus = 'Saved';
    setTimeout(() => { saveStatus = ''; }, 2000);
  }

  function handleKeydown(e) {
    if (e.key === 'Escape') {
      dispatch('back');
    }
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<div class="settings">
  <div class="settings-header">
    <button class="back-btn" on:click={() => dispatch('back')}><ArrowLeft size={14} /> Back</button>
    <span class="settings-title">Settings</span>
  </div>

  <div class="settings-scroll">
    <!-- Theme section -->
    <section class="settings-section">
      <h2 class="section-title">Theme</h2>
      <div class="theme-grid">
        {#each $themeIds as id}
          {@const theme = $allThemes[id]}
          <button
            class="theme-option"
            class:active={$currentThemeId === id}
            on:click={() => selectTheme(id)}
          >
            <div class="theme-preview" style="background: {theme.css['--bg-deepest']}; border-color: {theme.css['--border-subtle']}">
              <div class="preview-bar" style="background: {theme.css['--bg-surface']}; border-bottom-color: {theme.css['--border-subtle']}">
                <span class="preview-dot" style="background: #ff5f57" />
                <span class="preview-dot" style="background: #febc2e" />
                <span class="preview-dot" style="background: #28c840" />
              </div>
              <div class="preview-body">
                <div class="preview-line" style="background: {theme.css['--accent-green']}; width: 40%" />
                <div class="preview-line" style="background: {theme.css['--text-dim']}; width: 70%" />
                <div class="preview-line" style="background: {theme.css['--accent-purple']}; width: 30%" />
                <div class="preview-line" style="background: {theme.css['--text-dim']}; width: 55%" />
              </div>
            </div>
            <span class="theme-name">{theme.label}</span>
          </button>
        {/each}
      </div>
    </section>

    <!-- VSCodium Extension section -->
    <section class="settings-section">
      <h2 class="section-title">VSCodium Extension</h2>
      <p class="section-desc">Path to the VSCodium extensions directory for code intelligence.</p>
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
    </section>

    <!-- Imported Themes section -->
    {#if loadingThemes}
      <section class="settings-section">
        <h2 class="section-title">Imported Themes</h2>
        <p class="section-desc loading-text">Scanning themes...</p>
      </section>
    {:else if themeLoadError}
      <section class="settings-section">
        <h2 class="section-title">Imported Themes</h2>
        <p class="section-desc error-text">{themeLoadError}</p>
      </section>
    {:else if vscodiumThemes.length > 0}
      <section class="settings-section">
        <h2 class="section-title">Imported Themes</h2>
        <p class="section-desc">Click a theme to activate it. Themes are loaded from your VSCodium extensions.</p>
        <div class="theme-list">
          {#each vscodiumThemes as entry}
            {@const themeId = 'imported-' + entry.extensionId + '-' + entry.themePath.split('/').pop().replace('.json', '')}
            {@const cached = convertedCache[entry.themePath]}
            <button
              class="imported-theme-btn"
              class:active={$currentThemeId === themeId}
              disabled={activatingThemePath === entry.themePath}
              on:click={() => handleImportedThemeClick(entry)}
            >
              {#if cached}
                <div class="theme-preview mini" style="background: {cached.theme.css['--bg-deepest']}; border-color: {cached.theme.css['--border-subtle']}">
                  <div class="preview-line" style="background: {cached.theme.css['--accent-green']}; width: 40%"></div>
                  <div class="preview-line" style="background: {cached.theme.css['--text-dim']}; width: 65%"></div>
                  <div class="preview-line" style="background: {cached.theme.css['--accent-purple']}; width: 30%"></div>
                </div>
              {:else}
                <span class="theme-badge" class:dark={isDarkTheme(entry.uiTheme)} class:light={!isDarkTheme(entry.uiTheme)}>
                  {isDarkTheme(entry.uiTheme) ? 'D' : 'L'}
                </span>
              {/if}
              <span class="imported-theme-label">
                {entry.label}
                {#if activatingThemePath === entry.themePath}
                  <span class="activating-indicator">...</span>
                {/if}
              </span>
            </button>
          {/each}
        </div>
      </section>
    {/if}
  </div>

  <div class="settings-footer">
    <kbd>Esc</kbd> Back to feed
  </div>
</div>

<style>
  .settings {
    display: flex;
    flex-direction: column;
    height: 100%;
    background: var(--bg-deepest);
  }

  .settings-header {
    display: flex;
    align-items: center;
    gap: var(--sp-sm);
    padding: 6px var(--sp-lg);
    border-bottom: 1px solid var(--border-subtle);
    background: var(--bg-surface);
    flex-shrink: 0;
  }

  .back-btn {
    background: none;
    border: none;
    color: var(--accent-green);
    font-family: var(--font-ui);
    font-size: 13px;
    cursor: pointer;
    padding: 2px 6px;
    border-radius: var(--radius-sm);
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .back-btn:hover {
    text-shadow: 0 0 10px rgba(0, 229, 122, 0.6);
  }

  .settings-title {
    font-weight: 600;
    font-size: 14px;
    color: var(--text-primary);
  }

  .settings-scroll {
    flex: 1;
    overflow-y: auto;
    padding: var(--sp-xl) var(--sp-2xl);
    max-width: 640px;
  }

  .settings-section {
    margin-bottom: var(--sp-2xl);
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

  /* Theme grid */
  .theme-grid {
    display: flex;
    gap: var(--sp-md);
  }

  .theme-option {
    background: none;
    border: 2px solid var(--border-subtle);
    border-radius: var(--radius-lg);
    cursor: pointer;
    padding: 0;
    overflow: hidden;
    transition: border-color 150ms ease, transform 100ms ease;
    width: 160px;
  }

  .theme-option:hover {
    border-color: var(--border-emphasis);
    transform: translateY(-2px);
  }

  .theme-option.active {
    border-color: var(--accent-green);
  }

  .theme-preview {
    border: none;
    border-radius: 0;
    overflow: hidden;
  }

  .preview-bar {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 4px 6px;
    border-bottom: 1px solid;
  }

  .preview-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
  }

  .preview-body {
    padding: 8px;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .preview-line {
    height: 4px;
    border-radius: 2px;
    opacity: 0.7;
  }

  .theme-name {
    display: block;
    padding: 6px;
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--text-primary);
    text-align: center;
  }

  /* VSCodium path */
  .path-input-row {
    display: flex;
    gap: var(--sp-sm);
  }

  .path-input {
    flex: 1;
    padding: 6px 10px;
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    color: var(--text-primary);
    font-family: var(--font-mono);
    font-size: 12px;
    outline: none;
  }

  .path-input:focus {
    border-color: var(--accent-green);
  }

  .path-input::placeholder {
    color: var(--text-muted);
  }

  .browse-btn {
    padding: 6px 14px;
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    color: var(--text-dim);
    font-family: var(--font-mono);
    font-size: 12px;
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
    font-size: 10px;
    color: var(--accent-green);
  }

  /* Imported themes */
  .theme-list {
    display: flex;
    flex-direction: column;
    gap: 2px;
    max-height: 320px;
    overflow-y: auto;
  }

  .imported-theme-btn {
    display: flex;
    align-items: center;
    gap: var(--sp-sm);
    padding: 6px 10px;
    background: none;
    border: 1px solid transparent;
    border-radius: var(--radius-md);
    cursor: pointer;
    text-align: left;
    transition: all 100ms ease;
  }

  .imported-theme-btn:hover {
    background: var(--bg-elevated);
    border-color: var(--border-subtle);
  }

  .imported-theme-btn.active {
    background: var(--bg-active);
    border-color: var(--accent-green);
  }

  .imported-theme-btn:disabled {
    opacity: 0.6;
    cursor: wait;
  }

  .theme-badge {
    width: 20px;
    height: 20px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: var(--radius-sm);
    font-family: var(--font-mono);
    font-size: 10px;
    font-weight: 600;
    flex-shrink: 0;
  }

  .theme-badge.dark {
    background: #1e1e2e;
    color: #cdd6f4;
    border: 1px solid #313244;
  }

  .theme-badge.light {
    background: #eff1f5;
    color: #4c4f69;
    border: 1px solid #ccd0da;
  }

  .theme-preview.mini {
    width: 32px;
    height: 24px;
    border-radius: var(--radius-sm);
    border: 1px solid;
    display: flex;
    flex-direction: column;
    justify-content: center;
    gap: 2px;
    padding: 3px 4px;
    flex-shrink: 0;
  }

  .theme-preview.mini .preview-line {
    height: 2px;
    border-radius: 1px;
  }

  .imported-theme-label {
    font-family: var(--font-mono);
    font-size: 12px;
    color: var(--text-primary);
  }

  .loading-text {
    color: var(--text-dim);
    font-style: italic;
  }

  .error-text {
    color: var(--accent-red);
  }

  .activating-indicator {
    color: var(--text-muted);
  }

  /* Footer */
  .settings-footer {
    display: flex;
    align-items: center;
    gap: var(--sp-sm);
    padding: 4px var(--sp-lg);
    border-top: 1px solid var(--border-subtle);
    background: var(--bg-surface);
    font-size: 11px;
    color: var(--text-muted);
    user-select: none;
    flex-shrink: 0;
  }

  .settings-footer :global(kbd) {
    display: inline-block;
    padding: 0 4px;
    background: var(--bg-active);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-sm);
    font-family: var(--font-mono);
    font-size: 10px;
    color: var(--text-dim);
  }
</style>
