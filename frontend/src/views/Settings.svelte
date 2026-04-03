<script>
  import { createEventDispatcher, onMount } from 'svelte';
  import { ArrowLeft } from 'lucide-svelte';
  import { themes, themeIds, currentThemeId, applyTheme } from '../lib/stores/theme.js';
  import { GetConfig, SetTheme, SetVSCodiumExtPath, PickDirectory } from '../../wailsjs/go/main/App.js';

  const dispatch = createEventDispatcher();

  let vscodiumPath = '';
  let saveStatus = '';

  onMount(async () => {
    try {
      const cfg = await GetConfig();
      vscodiumPath = cfg.vscodiumExtPath || '';
    } catch {}
  });

  async function selectTheme(id) {
    applyTheme(id);
    try { await SetTheme(id); } catch {}
  }

  async function browseVSCodium() {
    try {
      const dir = await PickDirectory();
      if (dir) {
        vscodiumPath = dir;
        await SetVSCodiumExtPath(dir);
        flashSave();
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
        {#each themeIds as id}
          {@const theme = themes[id]}
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
