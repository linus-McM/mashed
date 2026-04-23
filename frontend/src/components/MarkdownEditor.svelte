<script>
  import { onMount, onDestroy } from 'svelte';
  import { get } from 'svelte/store';
  import { ReadFile, WriteFile } from '../../wailsjs/go/main/App.js';
  import {
    createDebouncedSave,
    applyToolbarAttributes,
    computeToolbarApplyTarget,
  } from './markdownEditorUtils';
  import { errorMessage } from '../lib/errorMessage';
  import { markdownMenuSettings, markdownMenuDirty } from '../lib/stores/markdownMenuSettings';

  /** @typedef {import('@milkdown/crepe').Crepe} Crepe */
  /** @typedef {typeof import('@milkdown/crepe')} CrepeModuleType */
  /** @typedef {import('@milkdown/plugin-listener').ListenerManager} ListenerManager */
  /** @typedef {import('../lib/stores/markdownMenuSettings').MarkdownMenuSettings} MarkdownMenuSettings */

  /** @type {string} */
  export let filePath = '';
  /** @type {string} */
  export let repoPath = '';
  /** @type {boolean} */
  export let editable = false;

  /** @type {HTMLDivElement | null} */
  let container = null;
  /** @type {Crepe | null} */
  let crepe = null;
  /** @type {'idle' | 'modified' | 'saving' | 'saved' | 'error'} */
  let saveStatus = 'idle';
  /** @type {ReturnType<typeof setTimeout> | null} */
  let statusTimer = null;
  let loading = true;
  let error = '';

  $: fullPath = filePath.startsWith('/') ? filePath : repoPath + '/' + filePath;

  // Module-level cache for lazy import
  /** @type {CrepeModuleType | null} */
  let CrepeModule = null;

  /** @returns {Promise<CrepeModuleType>} */
  async function loadCrepeModule() {
    if (CrepeModule) return CrepeModule;
    const [mod] = await Promise.all([
      import('@milkdown/crepe'),
      import('@milkdown/crepe/theme/classic-dark.css'),
      import('../styles/crepe-mashed.css'),
    ]);
    CrepeModule = mod;
    return mod;
  }

  const saver = createDebouncedSave(async () => {
    if (!crepe || !editable) return;
    saveStatus = 'saving';
    try {
      const content = crepe.getMarkdown();
      await WriteFile(fullPath, content);
      saveStatus = 'saved';
      if (statusTimer) clearTimeout(statusTimer);
      statusTimer = setTimeout(() => { saveStatus = 'idle'; }, 2000);
    } catch (e) {
      saveStatus = 'error';
      console.error('Auto-save failed:', e);
    }
  }, 800);

  // Generation counter prevents interleaved initEditor calls on rapid filePath changes
  let initGeneration = 0;

  async function initEditor() {
    const gen = ++initGeneration;
    loading = true;
    error = '';
    if (statusTimer) { clearTimeout(statusTimer); statusTimer = null; }

    try {
      const markdown = await ReadFile(fullPath);
      if (gen !== initGeneration) return;
      const { Crepe } = await loadCrepeModule();
      if (gen !== initGeneration) return;

      crepe = new Crepe({
        root: container,
        defaultValue: markdown,
      });

      crepe.on((/** @type {ListenerManager} */ api) => {
        api.markdownUpdated((_ctx, md, prevMd) => {
          if (md !== prevMd && editable) {
            saveStatus = 'modified';
            saver.schedule();
          }
        });
      });

      await crepe.create();
      if (gen !== initGeneration) { crepe.destroy().catch(() => {}); crepe = null; return; }
      crepe.setReadonly(!editable);

      // Story 06: apply data-toolbar-<key> attrs on .milkdown root for CSS-masking
      // fallback. See crepe-mashed.css for the hide rules. Must run AFTER
      // crepe.create() so the .milkdown element exists in `container`.
      applyToolbarAttributes(container, get(markdownMenuSettings));
    } catch (e) {
      if (gen !== initGeneration) return;
      error = errorMessage(e) || 'Failed to load editor';
    } finally {
      if (gen === initGeneration) loading = false;
    }
  }

  async function destroyEditor() {
    if (crepe) {
      const ref = crepe;
      crepe = null;
      try { await ref.destroy(); } catch { /* ignore — DOM may already be detached */ }
    }
  }

  /** @param {KeyboardEvent} e */
  function handleKeydown(e) {
    if ((e.metaKey || e.ctrlKey) && e.key === 's') {
      e.preventDefault();
      if (editable) saver.flush();
    }
  }

  onMount(() => {
    initEditor();
    window.addEventListener('keydown', handleKeydown);
  });

  onDestroy(() => {
    window.removeEventListener('keydown', handleKeydown);
    saver.flush();
    saver.destroy();
    if (statusTimer) { clearTimeout(statusTimer); statusTimer = null; }
    destroyEditor();
  });

  // Re-init on filePath change
  let prevFilePath = filePath;
  $: if (filePath !== prevFilePath) {
    prevFilePath = filePath;
    saver.flush();
    destroyEditor().then(() => initEditor());
  }

  // Story 06: deferred-apply of toolbar settings. Reactive block re-evaluates
  // when `$markdownMenuDirty` or `$markdownMenuSettings` change. Guards:
  //   - !dirty  → user is not mid-toggle inside Settings (apply is deferred
  //               until they click Back / Esc and clearMarkdownMenuDirty fires)
  //   - crepe   → editor is mounted
  //   - !loading → no re-init in flight
  // JSON stringify comparison makes the apply idempotent per change-set
  // (AC-3 / AC-4 / AC-6).
  //
  // Fallback strategy: CSS-masking (see crepe-mashed.css + markdownToolbarBuilder
  // header) makes a full destroy/init cycle unnecessary for toolbar-settings
  // changes — we only swap data attributes, preserving cursor + scroll. The
  // deferred-apply semantics from the story are preserved via the dirty guard.
  // The filePath-change branch ABOVE still uses the full saver.flush() →
  // destroy → init pattern because reading a new file requires a full remount.
  let lastAppliedSettings = JSON.stringify(get(markdownMenuSettings));
  $: {
    const next = computeToolbarApplyTarget({
      dirty: $markdownMenuDirty,
      mounted: crepe !== null,
      loading,
      settings: $markdownMenuSettings,
      lastApplied: lastAppliedSettings,
    });
    if (next !== null) {
      lastAppliedSettings = next;
      applyToolbarAttributes(container, $markdownMenuSettings);
      // No re-init: CSS masking handles visibility without teardown.
    }
  }

  // Toggle readonly when editable changes
  /** @type {boolean | null} */
  let lastReadonly = null;
  $: if (crepe && !loading) {
    const ro = !editable;
    if (ro !== lastReadonly) { lastReadonly = ro; crepe.setReadonly(ro); }
  }
</script>

<div class="markdown-editor">
  <div class="editor-header">
    <span class="file-path">{filePath}</span>
    <div class="header-right">
      {#if saveStatus === 'saving'}
        <span class="save-status saving">Saving...</span>
      {:else if saveStatus === 'saved'}
        <span class="save-status saved">Saved</span>
      {:else if saveStatus === 'modified'}
        <span class="save-status modified">Modified</span>
      {:else if saveStatus === 'error'}
        <span class="save-status error">Save failed</span>
      {/if}
    </div>
  </div>

  {#if loading && !crepe}
    <div class="loading-overlay">Loading editor...</div>
  {/if}
  {#if error}
    <div class="error-overlay">{error}</div>
  {/if}

  <div class="crepe-container" bind:this={container}></div>
</div>

<style>
  .markdown-editor {
    display: flex;
    flex-direction: column;
    height: 100%;
    background: var(--bg-deepest);
  }

  .editor-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 4px 12px;
    background: var(--bg-surface);
    border-bottom: 1px solid var(--border-default);
    min-height: 32px;
    flex-shrink: 0;
  }

  .file-path {
    font-family: var(--font-mono);
    font-size: var(--text-body);
    color: var(--text-dim);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .save-status {
    font-size: 11px;
    transition: color 100ms ease;
  }

  .save-status.saving { color: var(--text-muted); }
  .save-status.saved { color: var(--accent-green); }
  .save-status.modified { color: var(--accent-amber, #f0a500); }
  .save-status.error { color: var(--accent-red, #e84545); }

  .crepe-container {
    flex: 1;
    overflow: auto;
    min-height: 0;
  }

  /* Ensure Crepe fills the container */
  .crepe-container :global(.milkdown) {
    min-height: 100%;
    padding: 16px 24px;
  }

  .loading-overlay,
  .error-overlay {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 24px;
    font-size: 13px;
  }

  .loading-overlay { color: var(--text-dim); }
  .error-overlay { color: var(--accent-red, #e84545); }
</style>
