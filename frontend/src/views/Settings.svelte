<script>
  import { createEventDispatcher, onMount } from 'svelte';
  import { ArrowLeft } from 'lucide-svelte';
  import { allThemes, themeIds, currentThemeId, applyTheme } from '../lib/stores/theme.js';
  import { GetConfig, SetTheme, SetImportedTheme, SetVSCodiumExtPath, PickDirectory, ListVSCodiumThemes, ListLocalFonts, SetMonoFont, SetFontSize, SetSidebarWidth } from '../../wailsjs/go/main/App.js';
  import { activateImportedTheme, removeImportedTheme, convertedCache, makeThemeId } from '../lib/themeInit.js';
  import { builtInThemeIds } from '../lib/stores/theme.js';
  import { currentMonoFont, currentFontSize, applyFont, registerLocalFonts } from '../lib/stores/font.js';
  import { editorSettings, updateEditorSetting } from '../lib/stores/editorSettings.js';

  const dispatch = createEventDispatcher();

  let vscodiumPath = '';
  let saveStatus = '';
  let vscodiumThemes = [];
  let loadingThemes = false;
  let themeLoadError = '';
  let activatingThemePath = '';

  let localFonts = [];
  let allFonts = [];
  let loadingFonts = false;
  let selectedFont = '';
  let selectedFontSize = 13;
  let selectedSidebarWidth = 280;

  onMount(async () => {
    try {
      const cfg = await GetConfig();
      vscodiumPath = cfg.vscodiumExtPath || '';
      selectedFont = cfg.monoFont || '';
      selectedFontSize = cfg.fontSize || 13;
      selectedSidebarWidth = cfg.sidebarWidth || 280;
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

  async function handleRemoveTheme(id) {
    await removeImportedTheme(id);
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
      localFonts = await ListLocalFonts() || [];
      registerLocalFonts(localFonts);
      allFonts = localFonts.map(f => ({ family: f.family, source: 'bundled' }));
    } catch {
      localFonts = [];
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

  async function changeSidebarWidth(width) {
    selectedSidebarWidth = width;
    try { await SetSidebarWidth(width); } catch {}
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

  <div class="settings-body">
    <!-- Left column: themes -->
    <div class="col-themes">
      <h2 class="section-title">Themes</h2>
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
            on:change={(e) => updateEditorSetting('cursorStyle', e.target.value)}>
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
            on:change={(e) => updateEditorSetting('cursorBlinking', e.target.value)}>
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
            on:change={(e) => updateEditorSetting('wordWrap', e.target.value)}>
            <option value="off">off</option>
            <option value="on">on</option>
            <option value="wordWrapColumn">wordWrapColumn</option>
            <option value="bounded">bounded</option>
          </select>
        </div>
        <div class="setting-row">
          <label class="setting-label" for="lineNumbers">Line Numbers</label>
          <select id="lineNumbers" class="setting-select" value={$editorSettings.lineNumbers}
            on:change={(e) => updateEditorSetting('lineNumbers', e.target.value)}>
            <option value="on">on</option>
            <option value="off">off</option>
            <option value="relative">relative</option>
            <option value="interval">interval</option>
          </select>
        </div>
        <div class="setting-row">
          <label class="setting-label" for="renderLineHighlight">Line Highlight</label>
          <select id="renderLineHighlight" class="setting-select" value={$editorSettings.renderLineHighlight}
            on:change={(e) => updateEditorSetting('renderLineHighlight', e.target.value)}>
            <option value="none">none</option>
            <option value="gutter">gutter</option>
            <option value="line">line</option>
            <option value="all">all</option>
          </select>
        </div>
        <div class="setting-row">
          <label class="setting-label" for="renderWhitespace">Whitespace</label>
          <select id="renderWhitespace" class="setting-select" value={$editorSettings.renderWhitespace}
            on:change={(e) => updateEditorSetting('renderWhitespace', e.target.value)}>
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
            on:change={(e) => updateEditorSetting('tabSize', Math.min(8, Math.max(2, parseInt(e.target.value) || 2)))} />
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
                {@const alreadyImported = !!$allThemes[themeId]}
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
    padding: 5px 8px;
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
    font-size: 12px;
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
    font-size: 12px;
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
    font-size: 10px;
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
    font-size: 10px;
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
    padding: 5px 0;
    gap: var(--sp-md);
  }

  .setting-label {
    font-family: var(--font-mono);
    font-size: 12px;
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
