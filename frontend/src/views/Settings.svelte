<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';
  import { ArrowLeft } from 'lucide-svelte';
  import { allThemes, themeIds, currentThemeId, applyTheme } from '../lib/stores/theme.js';
  import { GetConfig, SetTheme, SetImportedTheme, SetVSCodiumExtPath, PickDirectory, ListVSCodiumThemes, ListLocalFonts, SetMonoFont, SetFontSize, SetSidebarWidth } from '../../wailsjs/go/main/App.js';
  import { activateImportedTheme, removeImportedTheme, makeThemeId } from '../lib/themeInit.js';
  import { builtInThemeIds } from '../lib/stores/theme.js';
  import { applyFont, registerLocalFonts } from '../lib/stores/font.js';
  import { editorSettings, updateEditorSetting } from '../lib/stores/editorSettings.js';
  import {
    uiAdapterEnabled,
    uiAdapterTimeoutMs,
    ollamaModel,
    ollamaEnabled,
    uiAdapterUntrustedExpanded,
    ollamaReachable,
    ollamaModels,
    hydrate as hydrateUIAdapter,
    refreshModels as refreshUIAdapterModels,
    setEnabled as setUIAdapterEnabled,
    setTimeoutMs as setUIAdapterTimeoutMs,
    setModel as setUIAdapterModel,
    setOllamaEnabled as setUIAdapterOllamaEnabled,
    setUntrustedExpanded as setUIAdapterUntrustedExpanded,
    validOllamaModelName,
  } from '../lib/stores/uiAdapterSettings';
  import { BrowserOpenURL } from '../../wailsjs/runtime/runtime.js';
  import { RefreshCw, TriangleAlert } from 'lucide-svelte';
  import { errorMessage } from '../lib/errorMessage';
  import type { LocalFontFamily, VSCodeThemeEntry } from '../lib/types/wails';

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

  let vscodiumPath = '';
  let saveStatus = '';
  let vscodiumThemes: VSCodeThemeEntry[] = [];
  let loadingThemes = false;
  let themeLoadError = '';
  let activatingThemePath = '';

  let localFonts: LocalFontFamily[] = [];
  let allFonts: FontOption[] = [];
  let loadingFonts = false;
  let selectedFont = '';
  let selectedFontSize = 13;
  let selectedSidebarWidth = 280;

  // UI AST adapter — local form state mirrors the store; binds to inputs so
  // blur-validated edits don't push partial keystrokes to disk.
  let timeoutInput = 3000;
  let timeoutError = '';
  let modelSelectValue = 'gemma3:4b';
  let customModelInput = '';
  let customModelError = '';
  let refreshingModels = false;

  const OLLAMA_INSTALL_URL = 'https://ollama.com/download';
  const CUSTOM_MODEL_SENTINEL = '__custom__';

  $: if ($uiAdapterTimeoutMs !== undefined) timeoutInput = $uiAdapterTimeoutMs;
  $: if ($ollamaModel) {
    modelSelectValue = resolveModelSelectValue($ollamaModel, $ollamaModels);
    if (modelSelectValue === CUSTOM_MODEL_SENTINEL && !customModelInput) {
      customModelInput = $ollamaModel;
    }
  }

  function resolveModelSelectValue(model: string, availableModels: string[]): string {
    if (!model) return CUSTOM_MODEL_SENTINEL;
    if (availableModels && availableModels.includes(model)) return model;
    return CUSTOM_MODEL_SENTINEL;
  }

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
      hasVscodiumPath ? scanThemes() : Promise.resolve(),
      scanFonts(),
      hydrateUIAdapter(),
    ]);
  });

  async function toggleUIAdapter() {
    const ok = await setUIAdapterEnabled(!$uiAdapterEnabled);
    if (ok) {
      flashSave();
      if ($uiAdapterEnabled) await refreshUIAdapterModels();
    }
  }

  async function toggleOllamaEnabled() {
    const ok = await setUIAdapterOllamaEnabled(!$ollamaEnabled);
    if (ok) flashSave();
  }

  async function toggleUntrustedExpanded() {
    const ok = await setUIAdapterUntrustedExpanded(!$uiAdapterUntrustedExpanded);
    if (ok) flashSave();
  }

  async function commitTimeout() {
    const ms = Number(timeoutInput);
    if (!Number.isFinite(ms) || ms < 500 || ms > 30000) {
      timeoutError = 'must be 500-30000';
      return;
    }
    timeoutError = '';
    if (ms === $uiAdapterTimeoutMs) return;
    const ok = await setUIAdapterTimeoutMs(ms);
    if (ok) flashSave();
    else timeoutError = 'save failed';
  }

  async function onModelSelect(value: string): Promise<void> {
    modelSelectValue = value;
    if (value === CUSTOM_MODEL_SENTINEL) {
      customModelInput = $ollamaModel;
      return;
    }
    customModelError = '';
    const ok = await setUIAdapterModel(value);
    if (ok) flashSave();
  }

  async function commitCustomModel() {
    if (!validOllamaModelName(customModelInput)) {
      customModelError = 'only [a-zA-Z0-9._:-], 1-64 chars';
      return;
    }
    customModelError = '';
    if (customModelInput === $ollamaModel) return;
    const ok = await setUIAdapterModel(customModelInput);
    if (ok) flashSave();
    else customModelError = 'save failed';
  }

  async function manualRefreshModels() {
    refreshingModels = true;
    try {
      await refreshUIAdapterModels();
    } finally {
      refreshingModels = false;
    }
  }

  function openOllamaDocs() {
    BrowserOpenURL(OLLAMA_INSTALL_URL);
  }

  $: sortedOllamaModels = [...($ollamaModels || [])].sort((a, b) => a.localeCompare(b));
  $: configuredModelMissing =
    $ollamaModel && !sortedOllamaModels.includes($ollamaModel);
  $: showOfflineBanner = $uiAdapterEnabled && $ollamaReachable === false;

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

  async function scanThemes(): Promise<void> {
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

  async function handleRemoveTheme(id: string): Promise<void> {
    await removeImportedTheme(id);
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
        flashSave();
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
      dispatch('back');
    }
  }

  /**
   * Narrow an `<input>` / `<select>` change-event target to an `HTMLInputElement`.
   * svelte-check reports `'e.target' is possibly null` + `Property 'value' does
   * not exist on type 'EventTarget'`; routing every handler through this helper
   * keeps the setting row templates tidy.
   */
  function fieldValue(e: Event): string {
    const t = e.currentTarget as HTMLInputElement | HTMLSelectElement | null;
    return t?.value ?? '';
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<div class="settings">
  <div class="settings-header">
    <button class="back-btn" on:click={() => dispatch('back')}><ArrowLeft size={14} /> Back</button>
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
      <!-- Mono Font -->
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

      <!-- Sidebar Width -->
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

      <!-- Editor -->
      <section class="settings-section">
        <h2 class="section-title">Editor</h2>

        <!-- Cursor -->
        <h3 class="subsection-title">Cursor</h3>
        <div class="setting-row">
          <label class="setting-label" for="cursorStyle">Cursor Style</label>
          <select id="cursorStyle" class="setting-select" value={$editorSettings.cursorStyle}
            on:change={(e) => updateEditorSetting('cursorStyle', fieldValue(e))}>
            <option value="line">line</option>
            <option value="line-thin">line-thin</option>
            <option value="block">block</option>
            <option value="block-outline">block-outline</option>
            <option value="underline">underline</option>
            <option value="underline-thin">underline-thin</option>
          </select>
        </div>
        <div class="setting-row">
          <label class="setting-label" for="cursorBlinking">Cursor Blinking</label>
          <select id="cursorBlinking" class="setting-select" value={$editorSettings.cursorBlinking}
            on:change={(e) => updateEditorSetting('cursorBlinking', fieldValue(e))}>
            <option value="blink">blink</option>
            <option value="smooth">smooth</option>
            <option value="phase">phase</option>
            <option value="expand">expand</option>
            <option value="solid">solid</option>
          </select>
        </div>

        <!-- Display -->
        <h3 class="subsection-title">Display</h3>
        <div class="setting-row">
          <label class="setting-label" for="wordWrap">Word Wrap</label>
          <select id="wordWrap" class="setting-select" value={$editorSettings.wordWrap}
            on:change={(e) => updateEditorSetting('wordWrap', fieldValue(e))}>
            <option value="off">off</option>
            <option value="on">on</option>
            <option value="wordWrapColumn">wordWrapColumn</option>
            <option value="bounded">bounded</option>
          </select>
        </div>
        <div class="setting-row">
          <label class="setting-label" for="lineNumbers">Line Numbers</label>
          <select id="lineNumbers" class="setting-select" value={$editorSettings.lineNumbers}
            on:change={(e) => updateEditorSetting('lineNumbers', fieldValue(e))}>
            <option value="on">on</option>
            <option value="off">off</option>
            <option value="relative">relative</option>
            <option value="interval">interval</option>
          </select>
        </div>
        <div class="setting-row">
          <label class="setting-label" for="renderLineHighlight">Line Highlight</label>
          <select id="renderLineHighlight" class="setting-select" value={$editorSettings.renderLineHighlight}
            on:change={(e) => updateEditorSetting('renderLineHighlight', fieldValue(e))}>
            <option value="none">none</option>
            <option value="gutter">gutter</option>
            <option value="line">line</option>
            <option value="all">all</option>
          </select>
        </div>
        <div class="setting-row">
          <label class="setting-label" for="renderWhitespace">Whitespace</label>
          <select id="renderWhitespace" class="setting-select" value={$editorSettings.renderWhitespace}
            on:change={(e) => updateEditorSetting('renderWhitespace', fieldValue(e))}>
            <option value="none">none</option>
            <option value="boundary">boundary</option>
            <option value="selection">selection</option>
            <option value="trailing">trailing</option>
            <option value="all">all</option>
          </select>
        </div>
        <div class="setting-row">
          <label class="setting-label" for="minimapEnabled">Minimap</label>
          <button id="minimapEnabled" class="setting-toggle" class:active={$editorSettings.minimapEnabled}
            on:click={() => updateEditorSetting('minimapEnabled', !$editorSettings.minimapEnabled)}>
            {$editorSettings.minimapEnabled ? 'On' : 'Off'}
          </button>
        </div>

        <!-- Editing -->
        <h3 class="subsection-title">Editing</h3>
        <div class="setting-row">
          <label class="setting-label" for="tabSize">Tab Size</label>
          <input id="tabSize" class="setting-number" type="number" min="2" max="8"
            value={$editorSettings.tabSize}
            on:change={(e) => updateEditorSetting('tabSize', Math.min(8, Math.max(2, parseInt(fieldValue(e)) || 2)))} />
        </div>
        <div class="setting-row">
          <label class="setting-label" for="insertSpaces">Insert Spaces</label>
          <button id="insertSpaces" class="setting-toggle" class:active={$editorSettings.insertSpaces}
            on:click={() => updateEditorSetting('insertSpaces', !$editorSettings.insertSpaces)}>
            {$editorSettings.insertSpaces ? 'On' : 'Off'}
          </button>
        </div>
        <div class="setting-row">
          <label class="setting-label" for="bracketPairColorization">Bracket Colors</label>
          <button id="bracketPairColorization" class="setting-toggle" class:active={$editorSettings.bracketPairColorization}
            on:click={() => updateEditorSetting('bracketPairColorization', !$editorSettings.bracketPairColorization)}>
            {$editorSettings.bracketPairColorization ? 'On' : 'Off'}
          </button>
        </div>

        <!-- Behavior -->
        <h3 class="subsection-title">Behavior</h3>
        <div class="setting-row">
          <label class="setting-label" for="fontLigatures">Font Ligatures</label>
          <button id="fontLigatures" class="setting-toggle" class:active={$editorSettings.fontLigatures}
            on:click={() => updateEditorSetting('fontLigatures', !$editorSettings.fontLigatures)}>
            {$editorSettings.fontLigatures ? 'On' : 'Off'}
          </button>
        </div>
        <div class="setting-row">
          <label class="setting-label" for="scrollBeyondLastLine">Scroll Beyond End</label>
          <button id="scrollBeyondLastLine" class="setting-toggle" class:active={$editorSettings.scrollBeyondLastLine}
            on:click={() => updateEditorSetting('scrollBeyondLastLine', !$editorSettings.scrollBeyondLastLine)}>
            {$editorSettings.scrollBeyondLastLine ? 'On' : 'Off'}
          </button>
        </div>
        <div class="setting-row">
          <label class="setting-label" for="smoothScrolling">Smooth Scrolling</label>
          <button id="smoothScrolling" class="setting-toggle" class:active={$editorSettings.smoothScrolling}
            on:click={() => updateEditorSetting('smoothScrolling', !$editorSettings.smoothScrolling)}>
            {$editorSettings.smoothScrolling ? 'On' : 'Off'}
          </button>
        </div>
      </section>

      <!-- VSCodium Extension path -->
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

      <!-- UI AST adapter -->
      <section class="settings-section" data-testid="ui-adapter-section">
        <h2 class="section-title">UI AST adapter</h2>
        <p class="section-desc">
          Translates Claude's round output into typed widgets. Requires Ollama running locally.
          Changes take effect on next app launch.
        </p>

        <div class="setting-row">
          <label class="setting-label" for="uiAdapterEnabled">Enable UI AST adapter</label>
          <button
            id="uiAdapterEnabled"
            class="setting-toggle"
            class:active={$uiAdapterEnabled}
            data-testid="ui-adapter-toggle"
            on:click={toggleUIAdapter}
          >
            {$uiAdapterEnabled ? 'On' : 'Off'}
          </button>
        </div>

        <div class="setting-row">
          <label class="setting-label" for="ollamaEnabled">Ollama enabled</label>
          <button
            id="ollamaEnabled"
            class="setting-toggle"
            class:active={$ollamaEnabled}
            on:click={toggleOllamaEnabled}
          >
            {$ollamaEnabled ? 'On' : 'Off'}
          </button>
        </div>

        <div class="setting-row">
          <label class="setting-label" for="uiAdapterTimeout">Timeout (ms)</label>
          <input
            id="uiAdapterTimeout"
            class="setting-number"
            class:invalid={timeoutError}
            type="number"
            min="500"
            max="30000"
            step="100"
            bind:value={timeoutInput}
            on:blur={commitTimeout}
          />
        </div>
        {#if timeoutError}
          <div class="setting-error" role="alert">{timeoutError}</div>
        {/if}

        <div class="setting-row">
          <label class="setting-label" for="ollamaModel">Ollama model</label>
          <div class="model-row-controls">
            <select
              id="ollamaModel"
              class="setting-select"
              data-testid="ui-adapter-model-select"
              value={modelSelectValue}
              on:change={(e) => onModelSelect(e.currentTarget.value)}
            >
              {#each sortedOllamaModels as m}
                <option value={m}>{m}</option>
                  {/each}
              {#if configuredModelMissing}
                <option value={$ollamaModel}
                  >{$ollamaModel}{sortedOllamaModels.length > 0 ? ' (not pulled)' : ''}</option
                >
              {/if}
              <option value={CUSTOM_MODEL_SENTINEL}>Custom…</option>
            </select>
            <button
              class="icon-btn"
              type="button"
              aria-label="Refresh model list"
              title="Refresh model list"
              data-testid="ui-adapter-refresh"
              disabled={refreshingModels}
              on:click={manualRefreshModels}
            >
              <RefreshCw size={12} />
            </button>
          </div>
        </div>

        {#if modelSelectValue === '__custom__'}
          <div class="setting-row custom-model-row">
            <label class="setting-label" for="customOllamaModel">Custom model</label>
            <input
              id="customOllamaModel"
              class="path-input custom-model-input"
              class:invalid={customModelError}
              type="text"
              placeholder="e.g. llama3.2:3b"
              bind:value={customModelInput}
              on:blur={commitCustomModel}
              data-testid="ui-adapter-custom-model"
            />
          </div>
          {#if customModelError}
            <div class="setting-error" role="alert">{customModelError}</div>
          {/if}
        {/if}

        <div class="setting-row">
          <label class="setting-label" for="untrustedExpanded">
            Expand "View raw" by default when a translation is flagged as untrusted
          </label>
          <button
            id="untrustedExpanded"
            class="setting-toggle"
            class:active={$uiAdapterUntrustedExpanded}
            on:click={toggleUntrustedExpanded}
          >
            {$uiAdapterUntrustedExpanded ? 'On' : 'Off'}
          </button>
        </div>

        {#if showOfflineBanner}
          <aside class="offline-banner" role="status" data-testid="ui-adapter-offline-banner">
            <span class="offline-banner-icon" aria-hidden="true">
              <TriangleAlert size={14} />
            </span>
            <div class="offline-banner-body">
              <div class="offline-banner-title">Ollama not reachable at localhost:11434.</div>
              <div class="offline-banner-text">
                Install Ollama and pull the model to enable this.
              </div>
              <a
                class="offline-banner-link"
                href={OLLAMA_INSTALL_URL}
                on:click|preventDefault={openOllamaDocs}
              >
                Install docs
              </a>
            </div>
          </aside>
        {/if}
      </section>
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

  /* Right column: other settings */
  .col-settings {
    flex: 1;
    overflow-y: auto;
    padding: var(--sp-lg) var(--sp-xl);
    min-width: 0;
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

  .loading-text {
    color: var(--text-dim);
    font-style: italic;
  }

  .error-text {
    color: var(--accent-red);
    font-size: 11px;
  }

  /* Editor settings: subsections, rows, controls */
  .subsection-title {
    font-family: var(--font-mono);
    font-size: var(--text-label);
    font-weight: 600;
    color: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    margin: var(--sp-lg) 0 var(--sp-sm);
  }

  .subsection-title:first-of-type {
    margin-top: 0;
  }

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

  .setting-select {
    padding: 4px 8px;
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-sm);
    color: var(--text-primary);
    font-family: var(--font-mono);
    font-size: 11px;
    outline: none;
    cursor: pointer;
    min-width: 120px;
    transition: border-color 100ms ease;
  }

  .setting-select:focus {
    border-color: var(--accent-green);
  }

  .setting-select:hover {
    border-color: var(--border-emphasis);
  }

  .setting-number {
    width: 60px;
    padding: 4px 8px;
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-sm);
    color: var(--text-primary);
    font-family: var(--font-mono);
    font-size: 11px;
    font-variant-numeric: tabular-nums;
    outline: none;
    text-align: center;
    transition: border-color 100ms ease;
  }

  .setting-number:focus {
    border-color: var(--accent-green);
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

  /* UI AST adapter section */
  .setting-number.invalid,
  .path-input.invalid {
    border-color: var(--accent-red);
  }

  .setting-error {
    padding-left: var(--sp-xs);
    margin-top: var(--sp-2xs);
    font-family: var(--font-ui);
    font-size: var(--text-label);
    font-weight: 500;
    color: var(--accent-red);
  }

  .model-row-controls {
    display: flex;
    align-items: center;
    gap: var(--sp-xs);
  }

  .icon-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 24px;
    height: 24px;
    padding: 0;
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-sm);
    color: var(--text-dim);
    cursor: pointer;
    transition: color var(--duration-short) var(--ease-move),
      border-color var(--duration-short) var(--ease-move);
  }

  .icon-btn:hover:not(:disabled) {
    color: var(--text-primary);
    border-color: var(--border-emphasis);
  }

  .icon-btn:disabled {
    opacity: 0.5;
    cursor: wait;
  }

  .custom-model-row {
    align-items: stretch;
  }

  .custom-model-input {
    flex: 1;
    max-width: 260px;
  }

  .offline-banner {
    display: flex;
    align-items: flex-start;
    gap: var(--sp-sm);
    margin-top: var(--sp-md);
    padding: var(--sp-sm) var(--sp-md);
    background: color-mix(in srgb, var(--accent-amber) 10%, transparent);
    border-left: 3px solid var(--accent-amber);
    border-radius: 0 var(--radius-md) var(--radius-md) 0;
  }

  .offline-banner-icon {
    display: inline-flex;
    align-items: center;
    color: var(--accent-amber);
    margin-top: 2px;
    flex-shrink: 0;
  }

  .offline-banner-body {
    display: flex;
    flex-direction: column;
    gap: var(--sp-2xs);
    min-width: 0;
  }

  .offline-banner-title {
    font-family: var(--font-ui);
    font-size: var(--text-body);
    font-weight: 600;
    color: var(--accent-amber);
  }

  .offline-banner-text {
    font-family: var(--font-ui);
    font-size: var(--text-label);
    color: var(--text-dim);
  }

  .offline-banner-link {
    align-self: flex-start;
    margin-top: var(--sp-2xs);
    padding: 0;
    background: none;
    border: none;
    font-family: var(--font-ui);
    font-size: var(--text-label);
    font-weight: 500;
    color: var(--accent-green);
    cursor: pointer;
    text-decoration: underline;
    text-underline-offset: 2px;
  }

  .offline-banner-link:hover {
    text-shadow: var(--glow-spread) color-mix(in srgb, var(--accent-green) 60%, transparent);
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
