<script>
  import { createEventDispatcher, onDestroy } from 'svelte';
  import { EventsOn } from '../../wailsjs/runtime/runtime.js';
  import { CreateRepo } from '../../wailsjs/go/main/App.js';

  const dispatch = createEventDispatcher();

  let name = '';
  let isPublic = false;
  let installBmad = true;
  let creating = false;
  let error = '';
  let steps = [];

  const cancelProgress = EventsOn('repo:create:progress', (event) => {
    if (event.error && event.done) {
      error = event.error;
      creating = false;
      return;
    }
    if (event.message) {
      steps = [...steps, event.message];
    }
    if (event.done && !event.error) {
      creating = false;
      dispatch('created', { name });
    }
  });

  onDestroy(() => {
    if (cancelProgress) cancelProgress();
  });

  $: slug = name.toLowerCase().replace(/[^a-z0-9-]/g, '-').replace(/-+/g, '-').replace(/^-|-$/g, '');
  $: valid = slug.length > 0 && slug.length <= 100;

  function submit() {
    if (!valid || creating) return;
    creating = true;
    error = '';
    steps = [];
    CreateRepo(slug, isPublic, installBmad);
  }

  function cancel() {
    if (!creating) dispatch('cancel');
  }

  function handleKeydown(e) {
    if (e.key === 'Escape') cancel();
    if (e.key === 'Enter' && valid && !creating) submit();
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<div class="overlay" role="presentation" on:click={cancel} on:keydown={() => {}}>
  <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
  <div class="modal" role="dialog" on:click|stopPropagation on:keydown={() => {}}>
    <h2>New Repository</h2>
    <p class="subtitle">Create a new project in your development directory</p>

    <div class="field">
      <label for="repo-name">Repository Name</label>
      <!-- svelte-ignore a11y-autofocus -->
      <input
        id="repo-name"
        type="text"
        class="name-input"
        bind:value={name}
        placeholder="my-project"
        disabled={creating}
        autofocus
      />
      {#if name && slug !== name}
        <span class="slug-hint">Will be created as: {slug}</span>
      {/if}
    </div>

    <div class="field">
      <!-- svelte-ignore a11y-label-has-associated-control -->
      <label>Visibility</label>
      <div class="toggle-row">
        <button
          class="toggle-option"
          class:selected={!isPublic}
          on:click={() => isPublic = false}
          disabled={creating}
        >
          Private
        </button>
        <button
          class="toggle-option"
          class:selected={isPublic}
          on:click={() => isPublic = true}
          disabled={creating}
        >
          Public
        </button>
      </div>
    </div>

    <div class="field">
      <label class="checkbox-label">
        <input type="checkbox" bind:checked={installBmad} disabled={creating} />
        Install BMAD Method
      </label>
      <span class="field-hint">Scaffolds agents, workflows, and sprint tracking via bmad-method</span>
    </div>

    {#if steps.length > 0}
      <div class="progress-log">
        {#each steps as step}
          <div class="progress-step">{step}</div>
        {/each}
        {#if creating}
          <div class="progress-step active">Working...</div>
        {/if}
      </div>
    {/if}

    {#if error}
      <div class="error">{error}</div>
    {/if}

    <div class="actions">
      <button class="btn-cancel" on:click={cancel} disabled={creating}>Cancel</button>
      <button
        class="btn-create"
        on:click={submit}
        disabled={!valid || creating}
      >
        {creating ? 'Creating...' : 'Create Repository'}
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
    z-index: 100;
  }

  .modal {
    background: var(--bg-surface);
    border: 1px solid var(--border-emphasis);
    border-radius: var(--radius-lg);
    padding: 32px;
    width: 420px;
    max-height: 80vh;
    overflow-y: auto;
  }

  h2 {
    font-size: var(--text-section);
    font-weight: 600;
    color: var(--text-primary);
    margin: 0 0 4px;
  }

  .subtitle {
    font-size: var(--text-body);
    color: var(--text-dim);
    margin: 0 0 24px;
  }

  .field {
    margin-bottom: 20px;
  }

  label {
    display: block;
    font-size: 11px;
    font-weight: 600;
    color: var(--text-dim);
    text-transform: uppercase;
    letter-spacing: 0.5px;
    margin-bottom: 8px;
  }

  .name-input {
    width: 100%;
    padding: 10px 12px;
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    color: var(--text-primary);
    font-family: var(--font-mono);
    font-size: 14px;
    outline: none;
    transition: border-color 100ms ease;
    box-sizing: border-box;
  }

  .name-input:focus { border-color: var(--accent-green); }
  .name-input:disabled { opacity: 0.5; }

  .slug-hint {
    display: block;
    font-size: var(--text-label);
    color: var(--text-muted);
    margin-top: 4px;
    font-family: var(--font-mono);
  }

  .toggle-row {
    display: flex;
    gap: 8px;
  }

  .toggle-option {
    flex: 1;
    padding: 8px;
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    cursor: pointer;
    transition: all 100ms ease-out;
    font-family: var(--font-mono);
    font-size: var(--text-body);
    color: var(--text-primary);
  }

  .toggle-option:hover:not(:disabled) {
    border-color: var(--border-emphasis);
  }

  .toggle-option.selected {
    border-color: var(--accent-green);
    background: color-mix(in srgb, var(--accent-green) 8%, transparent);
  }

  .toggle-option:disabled { opacity: 0.5; cursor: not-allowed; }

  .checkbox-label {
    display: flex;
    align-items: center;
    gap: 8px;
    text-transform: none;
    font-size: 13px;
    font-weight: 500;
    color: var(--text-primary);
    cursor: pointer;
  }

  .checkbox-label input[type="checkbox"] {
    accent-color: var(--accent-green);
    width: 16px;
    height: 16px;
  }

  .field-hint {
    display: block;
    font-size: var(--text-label);
    color: var(--text-muted);
    margin-top: 4px;
  }

  .progress-log {
    margin-bottom: 16px;
    padding: 8px 10px;
    background: var(--bg-deepest);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    max-height: 120px;
    overflow-y: auto;
  }

  .progress-step {
    font-family: var(--font-mono);
    font-size: var(--text-label);
    color: var(--text-dim);
    padding: 2px 0;
  }

  .progress-step.active {
    color: var(--accent-green);
  }

  .error {
    font-size: var(--text-body);
    color: var(--accent-red);
    margin-bottom: 16px;
    padding: 8px 10px;
    background: rgba(248, 81, 73, 0.08);
    border: 1px solid rgba(248, 81, 73, 0.2);
    border-radius: var(--radius-md);
  }

  .actions {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
    margin-top: 8px;
  }

  .btn-cancel {
    padding: 8px 16px;
    background: none;
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    color: var(--text-dim);
    font-family: var(--font-mono);
    font-size: 13px;
    cursor: pointer;
    transition: all 100ms ease;
  }

  .btn-cancel:hover:not(:disabled) {
    border-color: var(--border-emphasis);
    color: var(--text-primary);
  }

  .btn-cancel:disabled { opacity: 0.5; cursor: not-allowed; }

  .btn-create {
    padding: 8px 20px;
    background: var(--accent-green);
    border: none;
    border-radius: var(--radius-md);
    color: var(--bg-deepest);
    font-family: var(--font-mono);
    font-size: 13px;
    font-weight: 600;
    cursor: pointer;
    transition: opacity 100ms ease;
  }

  .btn-create:hover:not(:disabled) { opacity: 0.9; }
  .btn-create:disabled { opacity: 0.4; cursor: not-allowed; }
</style>
