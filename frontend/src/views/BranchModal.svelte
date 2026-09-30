<script>
  import { createEventDispatcher } from 'svelte';
  import { fade } from 'svelte/transition';
  import { cubicOut } from 'svelte/easing';
  import { GitBranch } from 'lucide-svelte';
  import { GitCreateBranch } from '../../wailsjs/go/main/App.js';
  import { errorMessage } from '../lib/errorMessage';

  /** @type {import('svelte').EventDispatcher<{ created: { branch: string }; cancel: void }>} */
  const dispatch = createEventDispatcher();

  // uiqa-06: modal fade entry/exit. prefers-reduced-motion zeroes durations.
  const reducedMotion = typeof window !== 'undefined' &&
    window.matchMedia('(prefers-reduced-motion: reduce)').matches;
  const backdropFadeProps = reducedMotion
    ? { duration: 0 }
    : { duration: 100, easing: cubicOut };
  const cardFadeProps = reducedMotion
    ? { duration: 0, delay: 0 }
    : { duration: 100, delay: 50, easing: cubicOut };

  export let repoPath = '';
  export let repoBranch = '';

  const prefixes = [
    { value: '', label: 'None' },
    { value: 'feature', label: 'feature/' },
    { value: 'hotfix', label: 'hotfix/' },
    { value: 'bugfix', label: 'bugfix/' },
    { value: 'release', label: 'release/' },
    { value: 'chore', label: 'chore/' },
    { value: 'docs', label: 'docs/' },
    { value: 'refactor', label: 'refactor/' },
    { value: 'test', label: 'test/' },
  ];

  let prefix = 'feature';
  let name = '';
  let autoCommit = false;
  let creating = false;
  let error = '';

  // Sanitize branch name as the user types
  /** @param {string} raw */
  function sanitize(raw) {
    return raw
      .toLowerCase()
      .replace(/\s+/g, '-')        // spaces to hyphens
      .replace(/\.{2,}/g, '.')     // no consecutive dots
      .replace(/[~^:?*\[\]\\]/g, '') // forbidden git chars
      .replace(/\/\//g, '/')       // no double slashes
      .replace(/^[.\-\/]+/, '')    // no leading dot/dash/slash
      .replace(/\.lock$/i, '')     // no .lock suffix
      .replace(/[.\-\/]+$/, '');   // no trailing dot/dash/slash
  }

  /** @param {Event} e */
  function handleInput(e) {
    const target = /** @type {HTMLInputElement} */ (e.target);
    name = sanitize(target.value);
  }

  $: fullName = prefix ? prefix + '/' + name : name;
  $: valid = name.length > 0 && name.length <= 100;

  async function create() {
    if (!valid) return;
    creating = true;
    error = '';
    try {
      await GitCreateBranch(repoPath, prefix, name, autoCommit);
      dispatch('created', { branch: fullName });
    } catch (e) {
      error = errorMessage(e) || 'Failed to create branch';
      creating = false;
    }
  }

  function cancel() {
    dispatch('cancel');
  }

  /** @param {KeyboardEvent} e */
  function handleKeydown(e) {
    if (e.key === 'Escape') cancel();
    if (e.key === 'Enter' && valid && !creating) create();
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<div class="overlay" role="presentation" on:click={cancel} on:keydown={() => {}} transition:fade={backdropFadeProps}>
  <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
  <div class="modal" role="dialog" on:click|stopPropagation on:keydown={() => {}} transition:fade={cardFadeProps}>
    <h2><GitBranch size={18} /> New Branch</h2>
    <p class="subtitle">Create a new branch from <strong>{repoBranch || 'HEAD'}</strong></p>

    <!-- Prefix selector -->
    <div class="field">
      <!-- svelte-ignore a11y-label-has-associated-control -->
      <label>Prefix</label>
      <div class="prefix-list">
        {#each prefixes as p}
          <button
            class="prefix-option"
            class:selected={prefix === p.value}
            on:click={() => prefix = p.value}
          >
            {p.label}
          </button>
        {/each}
      </div>
    </div>

    <!-- Branch name -->
    <div class="field">
      <label for="branch-name">Branch Name</label>
      <!-- svelte-ignore a11y-autofocus -->
      <input
        id="branch-name"
        type="text"
        class="branch-input"
        placeholder="my-branch-name"
        value={name}
        on:input={handleInput}
        autofocus
      />
      {#if name}
        <div class="preview mono">{fullName}</div>
      {/if}
    </div>

    <!-- Auto-commit toggle -->
    <div class="field toggle-field">
      <label class="toggle-label">
        <input type="checkbox" bind:checked={autoCommit} />
        <span>Auto-commit current changes before branching</span>
      </label>
    </div>

    {#if error}
      <div class="error">{error}</div>
    {/if}

    <div class="actions">
      <button class="btn-cancel" on:click={cancel}>Cancel</button>
      <button
        class="btn-create"
        on:click={create}
        disabled={!valid || creating}
      >
        {creating ? 'Creating...' : 'Create Branch'}
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
    width: 440px;
    max-height: 80vh;
    overflow-y: auto;
  }

  h2 {
    font-size: var(--text-section);
    font-weight: 600;
    color: var(--text-primary);
    margin: 0 0 4px;
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .subtitle {
    font-size: var(--text-body);
    color: var(--text-dim);
    margin: 0 0 24px;
  }

  .subtitle strong {
    color: var(--text-primary);
    font-family: var(--font-mono);
    font-size: 11px;
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

  .prefix-list {
    display: flex;
    flex-wrap: wrap;
    gap: var(--sp-sm);
  }

  .prefix-option {
    padding: 4px 10px;
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    cursor: pointer;
    transition: all 100ms ease-out;
    font-family: var(--font-mono);
    font-size: var(--text-body);
    color: var(--text-primary);
  }

  .prefix-option:hover {
    border-color: var(--border-emphasis);
    background: var(--bg-active);
  }

  .prefix-option.selected {
    border-color: var(--accent-green);
    background: color-mix(in srgb, var(--accent-green) 8%, transparent);
  }

  .branch-input {
    width: 100%;
    padding: 8px 12px;
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    color: var(--text-primary);
    font-family: var(--font-mono);
    font-size: 13px;
    outline: none;
    transition: border-color 100ms ease;
    box-sizing: border-box;
  }

  .branch-input:focus {
    border-color: var(--accent-green);
  }

  .branch-input::placeholder {
    color: var(--text-muted);
  }

  .preview {
    margin-top: var(--sp-xs);
    font-size: var(--text-body);
    color: var(--accent-green);
    padding: 4px 0;
  }

  .toggle-field {
    margin-bottom: 16px;
  }

  .toggle-label {
    display: flex;
    align-items: center;
    gap: 8px;
    cursor: pointer;
    text-transform: none;
    letter-spacing: 0;
    font-weight: 400;
    font-size: 13px;
    color: var(--text-primary);
  }

  .toggle-label input[type="checkbox"] {
    width: 14px;
    height: 14px;
    accent-color: var(--accent-green);
    cursor: pointer;
  }

  .error {
    font-size: var(--text-body);
    color: var(--accent-red);
    margin-bottom: 16px;
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
    font-family: var(--font-ui);
    font-size: 13px;
    cursor: pointer;
  }

  .btn-cancel:hover {
    border-color: var(--border-emphasis);
    color: var(--text-primary);
  }

  .btn-create {
    padding: 8px 20px;
    background: var(--accent-green);
    border: none;
    border-radius: var(--radius-md);
    color: var(--bg-deepest);
    font-family: var(--font-ui);
    font-size: 13px;
    font-weight: 600;
    cursor: pointer;
  }

  .btn-create:hover { opacity: 0.9; }
  .btn-create:disabled { opacity: 0.4; cursor: not-allowed; }

  .mono { font-family: var(--font-mono); }
</style>
