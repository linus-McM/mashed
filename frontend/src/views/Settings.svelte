<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';
  import { ArrowLeft } from 'lucide-svelte';
  import { allThemes, themeIds, currentThemeId, applyTheme } from '../lib/stores/theme.js';
  import { GetConfig, SetTheme, SetImportedTheme, ListLocalFonts, SetMonoFont, SetFontSize, SetSidebarWidth } from '../../wailsjs/go/main/App.js';
  import { removeImportedTheme } from '../lib/themeInit.js';
  import { builtInThemeIds } from '../lib/stores/theme.js';
  import { applyFont, registerLocalFonts } from '../lib/stores/font.js';
  import {
    markdownMenuSettings,
    updateMarkdownMenuItem,
    clearMarkdownMenuDirty,
  } from '../lib/stores/markdownMenuSettings';
  import { hydrate as hydrateUIAdapter } from '../lib/stores/uiAdapterSettings';
  import type { LocalFontFamily } from '../lib/types/wails';
  import EditorSettings from '../components/settings/EditorSettings.svelte';
  import VSCodiumThemes from '../components/settings/VSCodiumThemes.svelte';
  import UIAdapterSettings from '../components/settings/UIAdapterSettings.svelte';

  /** A theme entry as stored in `allThemes`. Mirrors the JSDoc-typed
   * `ThemeEntry` in `lib/stores/theme.js` — each imported VSCodium theme
   * has the same structural shape. */
  type ThemeEntry = {
    label: string;
    css: Record<string, string>;
    monaco?: unknown;
    xterm?: import('@xterm/xterm').ITheme;
  };
  type ThemeMap = Record<string, ThemeEntry>;

  /** Entry used by the local-font option list in the font picker. */
  type FontOption = { family: string; source: 'bundled' };

  const dispatch = createEventDispatcher<{ back: void }>();

  const TOOLBAR_ITEMS = [
    { key: 'bold',          label: 'Bold' },
    { key: 'italic',        label: 'Italic' },
    { key: 'strikethrough', label: 'Strikethrough' },
    { key: 'code',          label: 'Code' },
    { key: 'link',          label: 'Link' },
    { key: 'latex',         label: 'LaTeX' },
  ] as const;

  /** Dispatch `back` after clearing the markdown-menu dirty flag so
   * MarkdownEditor's re-init guard (story 06) sees a fresh state.
   * Call order is load-bearing — unit tests assert it explicitly. */
  function goBack(): void {
    clearMarkdownMenuDirty();
    dispatch('back');
  }

  let vscodiumPath = '';
  let saveStatus = '';
  let vscodiumThemesPanel: VSCodiumThemes;

  let localFonts: LocalFontFamily[] = [];
  let allFonts: FontOption[] = [];
  let loadingFonts = false;
  let selectedFont = '';
  let selectedFontSize = 13;
  let selectedSidebarWidth = 280;

  onMount(async () => {
    let hasVscodiumPath = false;
    try {
      const cfg = await GetConfig();
      vscodiumPath = cfg.vscodiumExtPath || '';
      selectedFont = cfg.monoFont || '';
      selectedFontSize = cfg.fontSize || 13;
      selectedSidebarWidth = cfg.sidebarWidth || 280;
      hasVscodiumPath = !!vscodiumPath;
    } catch {}
    await Promise.all([
      hasVscodiumPath ? vscodiumThemesPanel.scanThemes() : Promise.resolve(),
      scanFonts(),
      hydrateUIAdapter(),
    ]);
  });

  // Narrow the store value to the friendly ThemeMap shape — `allThemes` is
  // authored in JS as a loose object, but every entry structurally matches.
  $: themes = $allThemes as unknown as ThemeMap;

  async function selectTheme(id: string): Promise<void> {
    applyTheme(id);
    try {
      await SetTheme(id);
      await SetImportedTheme('');
    } catch {}
  }

  async function handleRemoveTheme(id: string): Promise<void> {
    await removeImportedTheme(id);
  }

  function flashSave() {
    saveStatus = 'Saved';
    setTimeout(() => { saveStatus = ''; }, 2000);
  }

  async function scanFonts(): Promise<void> {
    loadingFonts = true;
    try {
      localFonts = (await ListLocalFonts()) || [];
      registerLocalFonts(localFonts);
      allFonts = localFonts.map((f) => ({ family: f.family, source: 'bundled' as const }));
    } catch {
      localFonts = [];
      allFonts = [];
    } finally {
      loadingFonts = false;
    }
  }

  async function selectFont(family: string): Promise<void> {
    selectedFont = family;
    applyFont(family, selectedFontSize);
    try { await SetMonoFont(family); } catch {}
  }

  async function changeFontSize(size: number): Promise<void> {
    selectedFontSize = size;
    applyFont(selectedFont, size);
    try { await SetFontSize(size); } catch {}
  }

  async function changeSidebarWidth(width: number): Promise<void> {
    selectedSidebarWidth = width;
    try { await SetSidebarWidth(width); } catch {}
  }

  function handleKeydown(e: KeyboardEvent): void {
    if (e.key === 'Escape') {
      goBack();
    }
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<div class="settings">
  <div class="settings-header">
    <button class="back-btn" on:click={goBack}><ArrowLeft size={14} /> Back</button>
    <span class="settings-title">Settings</span>
  </div>

  <div class="settings-body">
    <!-- Left column: themes -->
    <div class="col-themes">
      <h2 class="section-title">Themes</h2>
      <div class="theme-list">
        {#each $themeIds as id}
          {@const theme = themes[id]}
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
            {#if !builtInThemeIds.includes(id)}
              <button class="remove-theme-btn" on:click|stopPropagation={() => handleRemoveTheme(id)} title="Remove theme">&times;</button>
            {/if}
          </button>
        {/each}
      </div>
    </div>

    <!-- Right column: font, extensions, import -->
    <div class="col-settings">
      <div class="settings-col settings-col-1">
      <!-- Mono Font -->
      <div class="settings-panel">
      <section class="settings-section">
        <h2 class="section-title">Font</h2>

        <!-- Live code preview -->
        <div class="font-code-preview" style="font-size: {selectedFontSize}px">
          <span class="preview-keyword">const</span> <span class="preview-fn">render</span> = (<span class="preview-param">items</span>) =&gt; {'{'}<br/>
          &nbsp;&nbsp;<span class="preview-keyword">return</span> items.<span class="preview-fn">filter</span>(x =&gt; x !== <span class="preview-str">""</span>)<br/>
          &nbsp;&nbsp;&nbsp;&nbsp;.<span class="preview-fn">map</span>((v, i) =&gt; <span class="preview-str">`${'{'}<span class="preview-param">i</span>{'}'}: ${'{'}<span class="preview-param">v</span>{'}'}`</span>)  <span class="preview-comment">// 0O 1lI</span><br/>
          {'}'}
        </div>

        <!-- Size control -->
        <div class="font-size-control">
          <button class="size-btn" on:click={() => changeFontSize(Math.max(10, selectedFontSize - 1))} disabled={selectedFontSize <= 10}>-</button>
          <span class="size-value">{selectedFontSize}px</span>
          <button class="size-btn" on:click={() => changeFontSize(Math.min(20, selectedFontSize + 1))} disabled={selectedFontSize >= 20}>+</button>
          <input type="range" min="10" max="20" bind:value={selectedFontSize}
                 on:input={() => changeFontSize(selectedFontSize)} class="size-slider" />
        </div>

        <!-- Font list -->
        {#if loadingFonts}
          <p class="section-desc loading-text">Loading fonts...</p>
        {:else}
          <div class="font-list">
            {#each allFonts as font}
              <button
                class="font-option"
                class:active={selectedFont === font.family}
                on:click={() => selectFont(font.family)}
              >
                <span class="font-sample" style="font-family: '{font.family}', monospace; font-size: {Math.max(selectedFontSize, 14)}px">
                  AaBb 0123
                </span>
                <span class="font-meta">
                  <span class="font-name">{font.family}</span>
                  <span class="font-glyphs" style="font-family: '{font.family}', monospace">      </span>
                </span>
              </button>
            {/each}
          </div>
        {/if}
      </section>
      </div>

      <!-- Sidebar Width -->
      <div class="settings-panel">
      <section class="settings-section">
        <h2 class="section-title">Sidebar Width</h2>
        <p class="section-desc">Default width for the workflow process sidebar.</p>
        <div class="font-size-control">
          <button class="size-btn" on:click={() => changeSidebarWidth(Math.max(200, selectedSidebarWidth - 10))} disabled={selectedSidebarWidth <= 200}>-</button>
          <span class="size-value">{selectedSidebarWidth}px</span>
          <button class="size-btn" on:click={() => changeSidebarWidth(Math.min(500, selectedSidebarWidth + 10))} disabled={selectedSidebarWidth >= 500}>+</button>
          <input type="range" min="200" max="500" step="10" bind:value={selectedSidebarWidth}
                 on:input={() => changeSidebarWidth(selectedSidebarWidth)} class="size-slider" />
        </div>
      </section>
      </div>

      <EditorSettings />
      </div>

      <div class="settings-col settings-col-2">
      <VSCodiumThemes bind:this={vscodiumThemesPanel} bind:vscodiumPath {saveStatus} onSaved={flashSave} />

      <UIAdapterSettings onSaved={flashSave} />

      <!-- Markdown Editor — selection toolbar toggles (story 05) -->
      <div class="settings-panel">
      <section class="settings-section" data-testid="markdown-menu-section">
        <h2 class="section-title">Markdown Editor</h2>
        <p class="section-desc">
          Selection toolbar items. Changes apply when you close Settings.
        </p>

        {#each TOOLBAR_ITEMS as item}
          <div class="setting-row">
            <span class="setting-label">{item.label}</span>
            <button
              class="setting-toggle"
              class:active={$markdownMenuSettings[item.key]}
              aria-pressed={$markdownMenuSettings[item.key]}
              data-testid={`toolbar-toggle-${item.key}`}
              on:click={() => updateMarkdownMenuItem(item.key, !$markdownMenuSettings[item.key])}
            >
              {$markdownMenuSettings[item.key] ? 'On' : 'Off'}
            </button>
          </div>
        {/each}
      </section>
      </div>
      </div>
    </div>
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
    padding: var(--sp-xs) var(--sp-lg);
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
    padding: var(--sp-2xs) var(--sp-xs);
    border-radius: var(--radius-sm);
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .back-btn:hover {
    text-shadow: var(--glow-spread) color-mix(in srgb, var(--accent-green) 60%, transparent);
  }

  .settings-title {
    font-weight: 600;
    font-size: 14px;
    color: var(--text-primary);
  }

  /* Two-column body */
  .settings-body {
    flex: 1;
    display: flex;
    overflow: hidden;
  }

  /* Left column: theme list */
  .col-themes {
    width: 280px;
    flex-shrink: 0;
    border-right: 1px solid var(--border-subtle);
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .col-themes .section-title {
    padding: var(--sp-lg) var(--sp-lg) var(--sp-sm);
    margin: 0;
  }

  .theme-list {
    flex: 1;
    overflow-y: auto;
    padding: 0 var(--sp-sm) var(--sp-lg);
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .theme-list-btn {
    display: flex;
    align-items: center;
    gap: var(--sp-sm);
    padding: var(--sp-xs) var(--sp-sm);
    background: none;
    border: 1px solid transparent;
    border-radius: var(--radius-md);
    cursor: pointer;
    text-align: left;
    transition: all 100ms ease;
    width: 100%;
  }

  .theme-list-btn:hover {
    background: var(--bg-elevated);
    border-color: var(--border-subtle);
  }

  .theme-list-btn.active {
    background: var(--bg-active);
    border-color: var(--accent-green);
  }

  .theme-thumb {
    width: 56px;
    height: 36px;
    border: 1px solid;
    border-radius: var(--radius-sm);
    overflow: hidden;
    flex-shrink: 0;
  }

  .thumb-bar {
    display: flex;
    align-items: center;
    gap: 2px;
    padding: 2px 4px;
    border-bottom: 1px solid;
  }

  .preview-dot {
    width: 4px;
    height: 4px;
    border-radius: 50%;
  }

  .thumb-body {
    padding: 4px;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .preview-line {
    height: 2px;
    border-radius: 1px;
    opacity: 0.7;
  }

  .theme-list-label {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--text-primary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .remove-theme-btn {
    margin-left: auto;
    background: none;
    border: none;
    color: var(--text-muted);
    font-size: 14px;
    cursor: pointer;
    padding: 0 2px;
    line-height: 1;
    border-radius: var(--radius-sm);
    transition: color 100ms ease;
    flex-shrink: 0;
    opacity: 0;
  }

  .theme-list-btn:hover .remove-theme-btn {
    opacity: 1;
  }

  .remove-theme-btn:hover {
    color: var(--accent-red);
  }

  /* Right column: other settings — 2-column panel grid (story 04) */
  .col-settings {
    flex: 1;
    overflow-y: auto;
    padding: var(--sp-lg) var(--sp-xl);
    min-width: 0;
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: var(--sp-lg);
    align-items: start;
  }

  .settings-col {
    display: flex;
    flex-direction: column;
    gap: var(--sp-lg);
    min-width: 0;
  }

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

  @media (max-width: 1100px) {
    .col-settings {
      grid-template-columns: 1fr;
    }
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

  /* Font: live code preview */
  .font-code-preview {
    font-family: var(--font-mono);
    line-height: 1.6;
    padding: 12px 16px;
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    margin-bottom: var(--sp-md);
    color: var(--text-primary);
    overflow-x: auto;
    white-space: nowrap;
  }

  .preview-keyword { color: var(--accent-purple); font-weight: 600; }
  .preview-fn { color: var(--accent-blue); }
  .preview-param { color: var(--accent-teal); }
  .preview-str { color: var(--accent-green); }
  .preview-comment { color: var(--text-muted); font-style: italic; }

  /* Font: size control */
  .font-size-control {
    display: flex;
    align-items: center;
    gap: var(--sp-sm);
    margin-bottom: var(--sp-md);
  }

  .size-btn {
    width: 28px;
    height: 28px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-sm);
    color: var(--text-dim);
    font-family: var(--font-mono);
    font-size: 14px;
    font-weight: 600;
    cursor: pointer;
    transition: all 100ms ease;
  }

  .size-btn:hover:not(:disabled) {
    color: var(--text-primary);
    border-color: var(--border-emphasis);
  }

  .size-btn:disabled {
    opacity: 0.3;
    cursor: default;
  }

  .size-value {
    font-family: var(--font-mono);
    font-size: var(--text-body);
    color: var(--text-primary);
    min-width: 36px;
    text-align: center;
    font-weight: 600;
  }

  .size-slider {
    flex: 1;
    max-width: 160px;
    accent-color: var(--accent-green);
    margin-left: var(--sp-sm);
  }

  /* Font: list */
  .font-list {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .font-option {
    display: flex;
    align-items: center;
    gap: var(--sp-md);
    padding: 8px 12px;
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

  .font-sample {
    color: var(--text-primary);
    min-width: 120px;
    white-space: nowrap;
  }

  .font-meta {
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-width: 0;
  }

  .font-name {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--text-dim);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .font-glyphs {
    font-size: var(--text-body);
    color: var(--text-muted);
    letter-spacing: 2px;
  }

  .loading-text {
    color: var(--text-dim);
    font-style: italic;
  }

  /* Setting rows (Markdown Editor toggles; children carry their own copies) */
  .setting-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: var(--sp-xs) 0;
    gap: var(--sp-md);
  }

  .setting-label {
    font-family: var(--font-mono);
    font-size: var(--text-body);
    color: var(--text-dim);
    white-space: nowrap;
  }

  .setting-toggle {
    padding: 4px 12px;
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-sm);
    color: var(--text-muted);
    font-family: var(--font-mono);
    font-size: 11px;
    cursor: pointer;
    min-width: 48px;
    text-align: center;
    transition: all 100ms ease;
  }

  .setting-toggle:hover {
    border-color: var(--border-emphasis);
    color: var(--text-dim);
  }

  .setting-toggle.active {
    background: var(--bg-active);
    border-color: var(--accent-green);
    color: var(--accent-green);
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
    font-size: var(--text-label);
    color: var(--text-dim);
  }
</style>
