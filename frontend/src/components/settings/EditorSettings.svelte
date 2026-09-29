<script lang="ts">
  // Editor settings panel (Monaco cursor / display / editing / behaviour
  // toggles). Extracted from views/Settings.svelte (spec R31) — markup and
  // scoped styles moved verbatim; the shared panel/section styles are
  // duplicated here because Svelte scopes them per component.
  import { editorSettings, updateEditorSetting } from '../../lib/stores/editorSettings.js';

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

<!-- Editor -->
<div class="settings-panel">
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

  .settings-section {
    margin-bottom: var(--sp-2xl);
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
</style>
