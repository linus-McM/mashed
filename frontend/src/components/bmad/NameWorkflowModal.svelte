<script>
  // NameWorkflowModal — first-time save gate for an unnamed BMAD workflow.
  //
  // Appears when the user tries to leave the WorkflowBuilder with a canvas
  // that has nodes AND has never been saved (no currentWorkflow.id yet).
  // Gives the user three explicit choices so an unnamed draft is never
  // silently persisted OR silently lost:
  //
  //   • Save   — commit with the typed name; navigation proceeds.
  //   • Discard — throw the canvas away; navigation proceeds.
  //   • Cancel  — keep the canvas as-is; navigation is aborted.
  //
  // The parent awaits `save` / `discard` / `cancel` via bind:this+an async
  // method on WorkflowBuilder, so this component is purely presentational —
  // it owns no persistence logic of its own.

  import { createEventDispatcher, onMount } from 'svelte';

  /** Suggested default name, typically derived from the repo or a timestamp. */
  export let defaultName = 'Untitled Workflow';
  /** How many nodes the user is about to save/discard, used for copy. */
  export let nodeCount = 0;

  const dispatch = createEventDispatcher();

  let name = defaultName;
  let inputEl;

  $: trimmed = (name || '').trim();
  $: canSave = trimmed.length > 0 && trimmed.length <= 120;

  onMount(() => {
    // Select-all on open so the user can just start typing to replace the
    // default. Wrapped in a microtask-ish setTimeout because the input is
    // created synchronously in the same tick as the modal mounts.
    setTimeout(() => {
      if (inputEl) {
        inputEl.focus();
        inputEl.select();
      }
    }, 0);
  });

  function handleSave() {
    if (!canSave) return;
    dispatch('save', { name: trimmed });
  }

  function handleDiscard() {
    dispatch('discard');
  }

  function handleCancel() {
    dispatch('cancel');
  }

  function handleKeydown(e) {
    if (e.key === 'Escape') {
      e.preventDefault();
      handleCancel();
      return;
    }
    if (e.key === 'Enter' && canSave) {
      e.preventDefault();
      handleSave();
    }
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<div class="overlay" role="presentation" on:click={handleCancel} on:keydown={() => {}}>
  <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
  <div class="modal" role="dialog" aria-labelledby="nwm-title" on:click|stopPropagation on:keydown={() => {}}>
    <h2 id="nwm-title">Name this workflow</h2>
    <p class="subtitle">
      You're leaving the workspace with
      {#if nodeCount === 1}
        1 unsaved node
      {:else}
        {nodeCount} unsaved nodes
      {/if}
      on the canvas. Give the workflow a name so it can be saved, or discard
      it to throw the canvas away.
    </p>

    <div class="field">
      <label for="nwm-name">Workflow name</label>
      <input
        id="nwm-name"
        type="text"
        class="name-input"
        bind:value={name}
        bind:this={inputEl}
        placeholder="e.g. Sprint triage"
        maxlength="120"
      />
      {#if !canSave && trimmed.length === 0}
        <span class="hint">A name is required to save.</span>
      {/if}
    </div>

    <div class="actions">
      <button class="btn btn-cancel" type="button" on:click={handleCancel}>Cancel</button>
      <button class="btn btn-discard" type="button" on:click={handleDiscard}>Discard</button>
      <button
        class="btn btn-save"
        type="button"
        on:click={handleSave}
        disabled={!canSave}
      >
        Save &amp; leave
      </button>
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
    background: var(--bg-surface);
    border: 1px solid var(--border-emphasis);
    border-radius: var(--radius-lg);
    padding: var(--sp-xl);
    width: 420px;
    max-width: calc(100vw - 2 * var(--sp-lg));
    box-shadow: 0 12px 40px color-mix(in srgb, var(--bg-deepest) 80%, transparent);
  }

  h2 {
    font-size: var(--text-section);
    font-weight: 600;
    color: var(--text-primary);
    margin: 0 0 var(--sp-2xs);
    letter-spacing: -0.01em;
  }

  .subtitle {
    font-size: var(--text-body);
    color: var(--text-dim);
    margin: 0 0 var(--sp-lg);
    line-height: 1.45;
  }

  .field {
    margin-bottom: var(--sp-lg);
  }

  label {
    display: block;
    font-size: var(--text-label);
    font-weight: 600;
    color: var(--text-dim);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    margin-bottom: var(--sp-xs);
  }

  .name-input {
    width: 100%;
    padding: 10px 12px;
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    color: var(--text-primary);
    font-family: var(--font-mono);
    font-size: var(--text-body);
    outline: none;
    box-sizing: border-box;
    transition: border-color var(--duration-short) var(--ease-enter);
  }

  .name-input:focus {
    border-color: var(--accent-amber);
  }

  .hint {
    display: block;
    margin-top: var(--sp-2xs);
    font-size: var(--text-label);
    color: var(--accent-red);
  }

  .actions {
    display: flex;
    justify-content: flex-end;
    gap: var(--sp-sm);
  }

  .btn {
    padding: 8px 16px;
    border-radius: var(--radius-md);
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

  .btn-discard {
    color: var(--accent-red);
    border-color: color-mix(in srgb, var(--accent-red) 40%, transparent);
  }
  .btn-discard:hover {
    background: color-mix(in srgb, var(--accent-red) 12%, transparent);
    border-color: var(--accent-red);
  }

  .btn-save {
    background: var(--accent-amber);
    color: var(--bg-deepest);
    font-weight: 600;
  }
  .btn-save:hover:not(:disabled) {
    background: color-mix(in srgb, var(--accent-amber) 85%, white);
  }
  .btn-save:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
</style>
