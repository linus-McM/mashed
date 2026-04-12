<script>
  import { createEventDispatcher, onMount, tick } from 'svelte';
  import { fly, fade, slide } from 'svelte/transition';
  import { cubicOut } from 'svelte/easing';
  import { SaveMashedAssetFrontmatter } from '../../../wailsjs/go/main/App.js';

  /** @type {import('../../../../wailsjs/go/models').bmad.MashedAssetInfo} */
  export let asset;

  const dispatch = createEventDispatcher();

  // ── Form state (shallow copy of asset fields) ──
  let name = asset?.name || '';
  let description = asset?.description || '';
  let mashedRole = asset?.role || 'command';
  let mashedCompletion = (() => {
    const c = asset?.completion || 'idle';
    if (c === 'idle' || c === 'exit') return c;
    return 'custom';
  })();
  let mashedChainable = asset?.chainable || 'single';
  let mashedSessionPinned = asset?.sessionPinned || false;
  let mashedInputs = (asset?.inputs || []).join(', ');
  let mashedOutputs = (asset?.outputs || []).join(', ');

  let saving = false;
  let saveError = '';

  $: canSave = (name || '').trim().length > 0 && !saving;

  // ── Focus trap ──
  let modalEl;
  let nameInputEl;

  onMount(async () => {
    await tick();
    if (nameInputEl) {
      nameInputEl.focus();
      nameInputEl.select();
    }
  });

  function getFocusableElements() {
    if (!modalEl) return [];
    return Array.from(
      modalEl.querySelectorAll(
        'input, textarea, select, button, [tabindex]:not([tabindex="-1"])'
      )
    ).filter(el => !el.disabled && el.offsetParent !== null);
  }

  function handleKeydown(e) {
    if (e.key === 'Escape') {
      e.preventDefault();
      if (saveError) {
        saveError = '';
        return;
      }
      handleCancel();
      return;
    }

    if (e.key === 'Tab') {
      const focusable = getFocusableElements();
      if (focusable.length === 0) return;

      const first = focusable[0];
      const last = focusable[focusable.length - 1];

      if (e.shiftKey) {
        if (document.activeElement === first) {
          e.preventDefault();
          last.focus();
        }
      } else {
        if (document.activeElement === last) {
          e.preventDefault();
          first.focus();
        }
      }
      return;
    }

    // Enter submits only from single-line inputs or the Save button
    if (e.key === 'Enter' && canSave) {
      const tag = e.target?.tagName?.toLowerCase();
      if (tag === 'textarea') return; // allow newlines
      if (tag === 'input' || tag === 'button') {
        e.preventDefault();
        handleSave();
      }
    }
  }

  function handleCancel() {
    dispatch('close');
  }

  async function handleSave() {
    if (!canSave) return;
    saving = true;
    saveError = '';

    const frontmatter = {
      name: name.trim(),
      description,
      mashedRole,
      mashedCompletion,
      mashedChainable,
      mashedSessionPinned,
      mashedInputs: mashedInputs.split(',').map(s => s.trim()).filter(Boolean),
      mashedOutputs: mashedOutputs.split(',').map(s => s.trim()).filter(Boolean),
    };

    try {
      await SaveMashedAssetFrontmatter(asset.path, frontmatter);
      dispatch('save', { path: asset.path });
    } catch (err) {
      saveError = String(err);
    } finally {
      saving = false;
    }
  }

  function handleBackdropClick() {
    handleCancel();
  }
</script>

<!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
<div
  class="overlay"
  role="presentation"
  on:click={handleBackdropClick}
  on:keydown={() => {}}
  transition:fade={{ duration: 150 }}
>
  <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
  <div
    class="modal"
    role="dialog"
    aria-labelledby="sem-title"
    bind:this={modalEl}
    on:click|stopPropagation
    on:keydown={handleKeydown}
    in:fly={{ y: 8, duration: 150, easing: cubicOut }}
    out:fly={{ y: 8, duration: 150, easing: cubicOut }}
  >
    <!-- Header -->
    <div class="modal-header">
      <div class="header-text">
        <h2 id="sem-title">Edit Skill</h2>
        <span class="file-path">{asset?.path || ''}</span>
      </div>
      <button class="close-btn" on:click={handleCancel} aria-label="Close">✕</button>
    </div>

    <!-- Form body -->
    <div class="modal-body">
      <div class="field-grid">
        <!-- Name (full width) -->
        <div class="field-group field-full">
          <label for="sem-name">Name</label>
          <input
            id="sem-name"
            type="text"
            class="input"
            bind:value={name}
            bind:this={nameInputEl}
            placeholder="e.g. simplify"
            maxlength="120"
          />
        </div>

        <!-- Description (full width) -->
        <div class="field-group field-full">
          <label for="sem-desc">Description</label>
          <textarea
            id="sem-desc"
            class="textarea"
            bind:value={description}
            placeholder="What this skill does…"
            rows="3"
          ></textarea>
        </div>

        <!-- mashedRole -->
        <div class="field-group">
          <label for="sem-role">Role</label>
          <select id="sem-role" class="select" bind:value={mashedRole}>
            <option value="command">command</option>
            <option value="skill">skill</option>
            <option value="agent">agent</option>
          </select>
        </div>

        <!-- mashedCompletion -->
        <div class="field-group">
          <label for="sem-completion">Completion</label>
          <select id="sem-completion" class="select" bind:value={mashedCompletion}>
            <option value="idle">idle</option>
            <option value="exit">exit</option>
            <option value="custom">custom</option>
          </select>
        </div>

        <!-- mashedChainable -->
        <div class="field-group">
          <label for="sem-chainable">Chainable</label>
          <select id="sem-chainable" class="select" bind:value={mashedChainable}>
            <option value="single">single</option>
            <option value="none">none</option>
            <option value="any">any</option>
          </select>
        </div>

        <!-- mashedSessionPinned -->
        <div class="field-group">
          <label for="sem-pinned">Session Pinned</label>
          <div class="checkbox-wrap">
            <input
              id="sem-pinned"
              type="checkbox"
              class="checkbox"
              bind:checked={mashedSessionPinned}
            />
            <span class="checkbox-label">{mashedSessionPinned ? 'Yes' : 'No'}</span>
          </div>
        </div>

        <!-- mashedInputs (full width) -->
        <div class="field-group field-full">
          <label for="sem-inputs">Inputs</label>
          <input
            id="sem-inputs"
            type="text"
            class="input"
            bind:value={mashedInputs}
            placeholder="comma-separated, e.g. repo, branch"
          />
          <span class="hint">Comma-separated list</span>
        </div>

        <!-- mashedOutputs (full width) -->
        <div class="field-group field-full">
          <label for="sem-outputs">Outputs</label>
          <input
            id="sem-outputs"
            type="text"
            class="input"
            bind:value={mashedOutputs}
            placeholder="comma-separated, e.g. summary, diff"
          />
          <span class="hint">Comma-separated list</span>
        </div>
      </div>
    </div>

    <!-- Footer -->
    <div class="modal-footer">
      {#if saveError}
        <div class="inline-error" transition:slide={{ duration: 150, easing: cubicOut }}>
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
          {saveError}
        </div>
      {/if}
      <div class="action-row">
        <button class="btn btn-cancel" type="button" on:click={handleCancel} disabled={saving}>Cancel</button>
        <button
          class="btn btn-save"
          class:saving
          type="button"
          on:click={handleSave}
          disabled={!canSave}
        >
          {saving ? 'Saving\u2026' : 'Save'}
        </button>
      </div>
    </div>
  </div>
</div>

<style>
  .overlay {
    position: fixed;
    inset: 0;
    background: var(--overlay-backdrop);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 400;
  }

  .modal {
    width: 520px;
    max-width: calc(100vw - 2 * var(--sp-xl));
    background: var(--bg-surface);
    border: 1px solid var(--border-emphasis);
    border-radius: var(--radius-lg);
    box-shadow: 0 12px 40px color-mix(in srgb, var(--bg-deepest) 80%, transparent);
    display: flex;
    flex-direction: column;
    max-height: calc(100vh - 2 * var(--sp-xl));
  }

  .modal-header {
    padding: var(--sp-lg) var(--sp-xl) var(--sp-md);
    border-bottom: 1px solid var(--border-subtle);
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: var(--sp-md);
  }

  .header-text {
    display: flex;
    flex-direction: column;
    gap: var(--sp-2xs);
    min-width: 0;
  }

  h2 {
    font-family: var(--font-ui);
    font-size: var(--text-section);
    font-weight: 600;
    color: var(--text-primary);
    letter-spacing: -0.01em;
    margin: 0;
  }

  .file-path {
    font-family: var(--font-mono);
    font-size: var(--text-label);
    color: var(--text-dim);
    letter-spacing: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .close-btn {
    background: none;
    border: none;
    color: var(--text-dim);
    cursor: pointer;
    padding: var(--sp-2xs);
    font-size: var(--text-body);
    line-height: 1;
    border-radius: var(--radius-sm);
    flex-shrink: 0;
    transition: color var(--duration-short) var(--ease-enter),
                background var(--duration-short) var(--ease-enter);
  }
  .close-btn:hover {
    color: var(--text-primary);
    background: var(--bg-active);
  }

  .modal-body {
    padding: var(--sp-lg) var(--sp-xl);
    overflow-y: auto;
    flex: 1;
  }

  .modal-footer {
    padding: var(--sp-md) var(--sp-xl) var(--sp-lg);
    border-top: 1px solid var(--border-subtle);
    display: flex;
    flex-direction: column;
    gap: var(--sp-sm);
  }

  .action-row {
    display: flex;
    justify-content: flex-end;
    gap: var(--sp-sm);
  }

  .inline-error {
    padding: var(--sp-sm) var(--sp-md);
    background: color-mix(in srgb, var(--accent-red) 10%, transparent);
    border-top: 1px solid color-mix(in srgb, var(--accent-red) 30%, transparent);
    border-radius: var(--radius-sm);
    color: var(--accent-red);
    font-family: var(--font-ui);
    font-size: var(--text-label);
    font-weight: 500;
    display: flex;
    align-items: center;
    gap: var(--sp-xs);
  }

  .field-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: var(--sp-md) var(--sp-lg);
  }

  .field-full {
    grid-column: 1 / -1;
  }

  .field-group {
    display: flex;
    flex-direction: column;
    gap: var(--sp-2xs);
  }

  label {
    font-family: var(--font-ui);
    font-size: var(--text-label);
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--text-dim);
  }

  .input, .textarea, .select {
    width: 100%;
    padding: var(--sp-sm) var(--sp-md);
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    color: var(--text-primary);
    font-family: var(--font-mono);
    font-size: var(--text-body);
    box-sizing: border-box;
    transition: border-color var(--duration-short) var(--ease-enter);
  }

  .input:hover, .textarea:hover, .select:hover {
    border-color: var(--border-emphasis);
  }

  .input:focus, .textarea:focus, .select:focus {
    border-color: var(--accent-amber);
    outline: none;
  }

  .textarea {
    line-height: 1.5;
    min-height: 80px;
    resize: vertical;
  }

  .select {
    appearance: none;
    background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 24 24' fill='none' stroke='%234a5a6a' stroke-width='2'%3E%3Cpath d='M6 9l6 6 6-6'/%3E%3C/svg%3E");
    background-repeat: no-repeat;
    background-position: right var(--sp-md) center;
    padding-right: var(--sp-xl);
    cursor: pointer;
  }

  .checkbox-wrap {
    display: flex;
    align-items: center;
    gap: var(--sp-sm);
    padding: var(--sp-sm) 0;
  }

  .checkbox {
    width: 16px;
    height: 16px;
    accent-color: var(--accent-green);
    cursor: pointer;
  }

  .checkbox-label {
    font-family: var(--font-mono);
    font-size: var(--text-body);
    color: var(--text-dim);
  }

  .hint {
    font-size: var(--text-label);
    color: var(--text-muted);
  }

  .btn {
    padding: var(--sp-sm) var(--sp-lg);
    border-radius: var(--radius-md);
    font-family: var(--font-ui);
    font-size: var(--text-body);
    font-weight: 500;
    cursor: pointer;
    border: 1px solid transparent;
    background: transparent;
    color: var(--text-primary);
    transition:
      background var(--duration-short) var(--ease-enter),
      border-color var(--duration-short) var(--ease-enter),
      color var(--duration-short) var(--ease-enter);
  }

  .btn-cancel {
    color: var(--text-dim);
  }
  .btn-cancel:hover {
    background: var(--bg-active);
    color: var(--text-primary);
  }

  .btn-save {
    background: var(--accent-green);
    color: var(--bg-deepest);
    font-weight: 600;
    min-width: 120px;
  }
  .btn-save:hover:not(:disabled) {
    background: color-mix(in srgb, var(--accent-green) 85%, white);
  }
  .btn-save:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
  .btn-save.saving {
    opacity: 0.8;
    cursor: not-allowed;
    background: color-mix(in srgb, var(--accent-green) 70%, transparent);
  }
  .btn-save.saving::before {
    content: '';
    display: inline-block;
    width: 10px;
    height: 10px;
    margin-right: var(--sp-xs);
    border-radius: 50%;
    border: 2px solid var(--bg-deepest);
    border-top-color: transparent;
    animation: spin 700ms linear infinite;
  }
  @keyframes spin { to { transform: rotate(360deg); } }
</style>
