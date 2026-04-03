<script>
  import { createEventDispatcher, onMount } from 'svelte';
  import { ArrowLeft } from 'lucide-svelte';
  import { allThemes, themeIds, currentThemeId, applyTheme } from '../lib/stores/theme.js';
  import { GetConfig, SetTheme, SetImportedTheme, SetVSCodiumExtPath, PickDirectory, ListVSCodiumThemes, ListNerdFonts, ListLocalFonts, OpenFontsDir, SetMonoFont, SetFontSize } from '../../wailsjs/go/main/App.js';
  import { activateImportedTheme, convertedCache, makeThemeId } from '../lib/themeInit.js';
  import { currentMonoFont, currentFontSize, applyFont, registerLocalFonts } from '../lib/stores/font.js';

  const dispatch = createEventDispatcher();

  let vscodiumPath = '';
  let saveStatus = '';
  let vscodiumThemes = [];
  let loadingThemes = false;
  let themeLoadError = '';
  let activatingThemePath = '';

  let nerdFonts = [];
  let localFonts = [];
  let allFonts = [];
  let loadingFonts = false;
  let selectedFont = '';
  let selectedFontSize = 13;

  onMount(async () => {
    try {
      const cfg = await GetConfig();
      vscodiumPath = cfg.vscodiumExtPath || '';
      selectedFont = cfg.monoFont || '';
      selectedFontSize = cfg.fontSize || 13;
      if (vscodiumPath) {
        await scanThemes();
      }
      await scanFonts();
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

  async function scanFonts() {
    loadingFonts = true;
    try {
      const [local, system] = await Promise.all([ListLocalFonts(), ListNerdFonts()]);
      localFonts = local || [];
      nerdFonts = system || [];

      // Register local fonts so they're available in CSS
      registerLocalFonts(localFonts);

      // Merge: local fonts first, then system fonts not already in local
      const localNames = new Set(localFonts.map(f => f.family));
      const systemOnly = nerdFonts.filter(f => !localNames.has(f.family));
      allFonts = [
        ...localFonts.map(f => ({ family: f.family, source: 'local' })),
        ...systemOnly.map(f => ({ family: f.family, source: 'system' })),
      ];
    } catch {
      localFonts = [];
      nerdFonts = [];
      allFonts = [];
    } finally {
      loadingFonts = false;
    }
  }

  async function selectFont(family) {
    selectedFont = family;
    applyFont(family, selectedFontSize);
    try { await SetMonoFont(family); } catch {}
  }

  async function changeFontSize(size) {
    selectedFontSize = size;
    applyFont(selectedFont, size);
    try { await SetFontSize(size); } catch {}
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
      <div class="theme-list">
        {#each $themeIds as id}
          {@const theme = $allThemes[id]}
          <button
            class="theme-list-btn"
            class:active={$currentThemeId === id}
            on:click={() => selectTheme(id)}
          >
            <div class="theme-thumb" style="background: {theme.css['--bg-deepest']}; border-color: {theme.css['--border-subtle']}">
              <div class="thumb-bar" style="background: {theme.css['--bg-surface']}; border-bottom-color: {theme.css['--border-subtle']}">
                <span class="preview-dot" style="background: #ff5f57" />
                <span class="preview-dot" style="background: #febc2e" />
                <span class="preview-dot" style="background: #28c840" />
              </div>
              <div class="thumb-body">
                <div class="preview-line" style="background: {theme.css['--accent-green']}; width: 40%" />
                <div class="preview-line" style="background: {theme.css['--text-dim']}; width: 70%" />
                <div class="preview-line" style="background: {theme.css['--accent-purple']}; width: 30%" />
                <div class="preview-line" style="background: {theme.css['--text-dim']}; width: 55%" />
              </div>
            </div>
            <span class="theme-list-label">{theme.label}</span>
          </button>
        {/each}
      </div>
    </section>

    <!-- Mono Font section -->
    <section class="settings-section">
      <h2 class="section-title">Mono Font</h2>
      <p class="section-desc">Drop font files into the fonts folder, or use system-installed Nerd Fonts.</p>

      <div class="font-actions-row">
        <button class="browse-btn" on:click={() => OpenFontsDir()}>Open Fonts Folder</button>
        <button class="browse-btn" on:click={scanFonts}>Rescan</button>
      </div>

      <div class="font-size-row">
        <label>Size: {selectedFontSize}px</label>
        <input type="range" min="10" max="20" bind:value={selectedFontSize}
               on:change={() => changeFontSize(selectedFontSize)} />
      </div>

      {#if loadingFonts}
        <p class="section-desc loading-text">Scanning fonts...</p>
      {:else}
        <div class="font-list">
          <button class="font-option" class:active={!selectedFont}
                  on:click={() => selectFont('')}>
            <span class="font-preview" style="font-family: 'JetBrains Mono', monospace">Abc 0O1l</span>
            <span class="font-name">Default (JetBrains Mono)</span>
          </button>
          {#each allFonts as font}
            <button class="font-option" class:active={selectedFont === font.family}
                    on:click={() => selectFont(font.family)}>
              <span class="font-preview" style="font-family: '{font.family}', monospace">Abc 0O1l</span>
              <span class="font-name">{font.family}</span>
              <span class="font-source">{font.source}</span>
            </button>
          {/each}
          {#if allFonts.length === 0}
            <p class="section-desc" style="margin-top: var(--sp-sm)">No fonts found. Drop .ttf/.otf/.woff2 files into the fonts folder, or install <a href="https://www.nerdfonts.com/" class="nerd-link">Nerd Fonts</a> system-wide.</p>
          {/if}
        </div>
      {/if}
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
            {@const themeId = makeThemeId(entry.themePath, entry.extensionId)}
            {@const cached = convertedCache[entry.themePath]}
            <button
              class="theme-list-btn"
              class:active={$currentThemeId === themeId}
              disabled={activatingThemePath === entry.themePath}
              on:click={() => handleImportedThemeClick(entry)}
            >
              {#if cached}
                <div class="theme-thumb" style="background: {cached.theme.css['--bg-deepest']}; border-color: {cached.theme.css['--border-subtle']}">
                  <div class="thumb-bar" style="background: {cached.theme.css['--bg-surface']}; border-bottom-color: {cached.theme.css['--border-subtle']}">
                    <span class="preview-dot" style="background: #ff5f57" />
                    <span class="preview-dot" style="background: #febc2e" />
                    <span class="preview-dot" style="background: #28c840" />
                  </div>
                  <div class="thumb-body">
                    <div class="preview-line" style="background: {cached.theme.css['--accent-green']}; width: 40%" />
                    <div class="preview-line" style="background: {cached.theme.css['--text-dim']}; width: 65%" />
                    <div class="preview-line" style="background: {cached.theme.css['--accent-purple']}; width: 30%" />
                    <div class="preview-line" style="background: {cached.theme.css['--text-dim']}; width: 55%" />
                  </div>
                </div>
              {:else}
                <span class="theme-badge" class:dark={isDarkTheme(entry.uiTheme)} class:light={!isDarkTheme(entry.uiTheme)}>
                  {isDarkTheme(entry.uiTheme) ? 'D' : 'L'}
                </span>
              {/if}
              <span class="theme-list-label">
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

  /* Theme list */
  .theme-list-btn {
    display: flex;
    align-items: center;
    gap: var(--sp-md);
    padding: 6px 10px;
    background: none;
    border: 1px solid transparent;
    border-radius: var(--radius-md);
    cursor: pointer;
    text-align: left;
    transition: all 100ms ease;
  }

  .theme-list-btn:hover {
    background: var(--bg-elevated);
    border-color: var(--border-subtle);
  }

  .theme-list-btn.active {
    background: var(--bg-active);
    border-color: var(--accent-green);
  }

  .theme-list-btn:disabled {
    opacity: 0.6;
    cursor: wait;
  }

  .theme-thumb {
    width: 80px;
    height: 52px;
    border: 1px solid;
    border-radius: var(--radius-md);
    overflow: hidden;
    flex-shrink: 0;
  }

  .thumb-bar {
    display: flex;
    align-items: center;
    gap: 3px;
    padding: 3px 5px;
    border-bottom: 1px solid;
  }

  .preview-dot {
    width: 5px;
    height: 5px;
    border-radius: 50%;
  }

  .thumb-body {
    padding: 6px;
    display: flex;
    flex-direction: column;
    gap: 3px;
  }

  .preview-line {
    height: 3px;
    border-radius: 1.5px;
    opacity: 0.7;
  }

  .theme-list-label {
    font-family: var(--font-mono);
    font-size: 12px;
    color: var(--text-primary);
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

  /* Theme list shared */
  .theme-list {
    display: flex;
    flex-direction: column;
    gap: 2px;
    max-height: 400px;
    overflow-y: auto;
  }

  .theme-badge {
    width: 80px;
    height: 52px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: var(--radius-md);
    font-family: var(--font-mono);
    font-size: 14px;
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

  /* Font selection */
  .font-actions-row {
    display: flex;
    gap: var(--sp-sm);
    margin-bottom: var(--sp-md);
  }

  .font-source {
    font-family: var(--font-mono);
    font-size: 9px;
    color: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: 0.5px;
    margin-left: auto;
  }

  .font-size-row {
    display: flex;
    align-items: center;
    gap: var(--sp-sm);
    margin-bottom: var(--sp-md);
    font-family: var(--font-mono);
    font-size: 12px;
    color: var(--text-dim);
  }

  .font-size-row input[type="range"] {
    flex: 1;
    max-width: 200px;
    accent-color: var(--accent-green);
  }

  .font-list {
    display: flex;
    flex-direction: column;
    gap: 2px;
    max-height: 280px;
    overflow-y: auto;
  }

  .font-option {
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

  .font-option:hover {
    background: var(--bg-elevated);
    border-color: var(--border-subtle);
  }

  .font-option.active {
    background: var(--bg-active);
    border-color: var(--accent-green);
  }

  .font-preview {
    font-size: 16px;
    color: var(--text-primary);
    min-width: 80px;
  }

  .font-name {
    font-family: var(--font-mono);
    font-size: 12px;
    color: var(--text-dim);
  }

  .nerd-link {
    color: var(--accent-blue);
    text-decoration: none;
  }

  .nerd-link:hover {
    text-decoration: underline;
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
