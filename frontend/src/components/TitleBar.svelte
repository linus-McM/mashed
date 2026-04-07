<script>
  import { createEventDispatcher } from 'svelte';
  import { Quit, WindowMinimise, WindowToggleMaximise } from '../../wailsjs/runtime/runtime.js';
  import { Settings, Palette, Workflow } from 'lucide-svelte';
  import { allThemes, themeIds, currentThemeId, applyTheme } from '../lib/stores/theme.js';
  import { SetTheme } from '../../wailsjs/go/main/App.js';

  const dispatch = createEventDispatcher();

  let showThemePicker = false;

  function toggleThemePicker() {
    showThemePicker = !showThemePicker;
  }

  async function selectTheme(id) {
    applyTheme(id);
    showThemePicker = false;
    try { await SetTheme(id); } catch {}
  }

  function openSettings() {
    showThemePicker = false;
    dispatch('open-settings');
  }

  function openWorkflows() {
    showThemePicker = false;
    dispatch('open-workflows');
  }
</script>

<svelte:window on:click={() => showThemePicker = false} />

<div class="titlebar">
  <div class="traffic-lights">
    <button class="tl-btn close" on:click={Quit} title="Close">
      <svg viewBox="0 0 12 12"><path d="M3.5 3.5l5 5M8.5 3.5l-5 5" stroke="currentColor" stroke-width="1.2" fill="none"/></svg>
    </button>
    <button class="tl-btn minimize" on:click={WindowMinimise} title="Minimize">
      <svg viewBox="0 0 12 12"><path d="M2.5 6h7" stroke="currentColor" stroke-width="1.2" fill="none"/></svg>
    </button>
    <button class="tl-btn fullscreen" on:click={WindowToggleMaximise} title="Fullscreen">
      <svg viewBox="0 0 12 12"><path d="M3 8.5L6 4l3 4.5H3z" fill="currentColor"/></svg>
    </button>
  </div>
  <div class="drag-region" on:dblclick={WindowToggleMaximise}></div>
  <div class="titlebar-actions">
    <div class="theme-picker-wrap">
      <button class="titlebar-btn" on:click|stopPropagation={toggleThemePicker} title="Switch theme">
        <Palette size={14} />
      </button>
      {#if showThemePicker}
        <div class="theme-popover" on:click|stopPropagation>
          {#each $themeIds as id}
            {@const theme = $allThemes[id]}
            <button
              class="theme-card"
              class:active={$currentThemeId === id}
              on:click={() => selectTheme(id)}
            >
              <div class="theme-swatches">
                <span class="swatch" style="background: {theme.css['--bg-deepest']}" />
                <span class="swatch" style="background: {theme.css['--accent-green']}" />
                <span class="swatch" style="background: {theme.css['--accent-purple']}" />
                <span class="swatch" style="background: {theme.css['--text-primary']}" />
              </div>
              <span class="theme-label">{theme.label}</span>
            </button>
          {/each}
        </div>
      {/if}
    </div>
    <button class="titlebar-btn" on:click={openWorkflows} title="Workflows">
      <Workflow size={14} />
    </button>
    <button class="titlebar-btn" on:click={openSettings} title="Settings">
      <Settings size={14} />
    </button>
  </div>
</div>

<style>
  .titlebar {
    display: flex;
    align-items: center;
    height: 38px;
    padding: 0 12px;
    background: var(--bg-surface);
    border-bottom: 1px solid var(--border-subtle);
    flex-shrink: 0;
    --wails-draggable: drag;
    user-select: none;
    -webkit-user-select: none;
  }

  .drag-region {
    flex: 1;
    height: 100%;
    --wails-draggable: drag;
  }

  .traffic-lights {
    display: flex;
    gap: 8px;
    align-items: center;
    --wails-draggable: none;
  }

  .tl-btn {
    width: 12px;
    height: 12px;
    border-radius: 50%;
    border: none;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 0;
    transition: filter 120ms ease;
  }

  .tl-btn svg {
    width: 8px;
    height: 8px;
    color: transparent;
    transition: color 120ms ease;
  }

  .tl-btn.close { background: #ff5f57; }
  .tl-btn.minimize { background: #febc2e; }
  .tl-btn.fullscreen { background: #28c840; }

  .tl-btn:hover svg { color: rgba(0, 0, 0, 0.5); }
  .tl-btn:active { filter: brightness(0.8); }

  .traffic-lights:hover .tl-btn svg { color: rgba(0, 0, 0, 0.5); }

  /* Right-side actions */
  .titlebar-actions {
    display: flex;
    align-items: center;
    gap: 4px;
    --wails-draggable: none;
  }

  .titlebar-btn {
    background: none;
    border: none;
    color: var(--text-muted);
    cursor: pointer;
    padding: 4px 6px;
    border-radius: var(--radius-sm);
    display: flex;
    align-items: center;
    transition: color 100ms ease, background 100ms ease;
  }

  .titlebar-btn:hover {
    color: var(--text-dim);
    background: var(--bg-elevated);
  }

  /* Theme picker popover */
  .theme-picker-wrap {
    position: relative;
  }

  .theme-popover {
    position: absolute;
    top: calc(100% + 8px);
    right: 0;
    display: flex;
    flex-direction: column;
    gap: 4px;
    padding: 8px;
    background: var(--bg-elevated);
    border: 1px solid var(--border-emphasis);
    border-radius: var(--radius-lg);
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.5);
    z-index: 200;
    min-width: 140px;
  }

  .theme-card {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 6px 8px;
    background: none;
    border: 1px solid transparent;
    border-radius: var(--radius-md);
    cursor: pointer;
    transition: all 100ms ease;
  }

  .theme-card:hover {
    background: var(--bg-active);
    border-color: var(--border-subtle);
  }

  .theme-card.active {
    border-color: var(--accent-green);
    background: var(--bg-active);
  }

  .theme-swatches {
    display: flex;
    gap: 3px;
  }

  .swatch {
    width: 12px;
    height: 12px;
    border-radius: 50%;
    border: 1px solid rgba(128, 128, 128, 0.2);
  }

  .theme-label {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--text-primary);
  }
</style>
