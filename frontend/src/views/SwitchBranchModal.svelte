<script>
  import { onMount, createEventDispatcher } from 'svelte';
  import { GitBranch, AlertTriangle, Search } from 'lucide-svelte';
  import { GitListBranches, GitSwitchBranch, RepoStatus } from '../../wailsjs/go/main/App.js';
  import { errorMessage } from '../lib/errorMessage';

  /** @type {import('svelte').EventDispatcher<{ switched: { branch: string }; cancel: void }>} */
  const dispatch = createEventDispatcher();

  export let repoPath = '';
  export let currentBranch = '';
  export let repoColor = '';

  /** @type {Array<{ name: string; current: boolean }>} */
  let branches = [];
  let selectedBranch = '';
  let dirty = false;
  let autoCommit = false;
  let switching = false;
  let error = '';
  let filter = '';

  $: displayColor = repoColor && repoColor !== '#1e2530' ? repoColor : 'var(--accent-green)';
  $: filtered = branches.filter(b =>
    !b.current && b.name.toLowerCase().includes(filter.toLowerCase())
  );
  $: canSwitch = selectedBranch && (!dirty || autoCommit);

  onMount(async () => {
    try {
      const [branchList, status] = await Promise.all([
        GitListBranches(repoPath),
        RepoStatus(repoPath),
      ]);
      branches = branchList || [];
      dirty = status?.dirty || false;
    } catch {
      error = 'Failed to load branches';
    }
  });

  async function switchBranch() {
    if (!canSwitch) return;
    switching = true;
    error = '';
    try {
      await GitSwitchBranch(repoPath, selectedBranch, autoCommit);
      dispatch('switched', { branch: selectedBranch });
    } catch (e) {
      error = errorMessage(e) || 'Failed to switch branch';
      switching = false;
    }
  }

  function cancel() {
    dispatch('cancel');
  }

  /** @param {KeyboardEvent} e */
  function handleKeydown(e) {
    if (e.key === 'Escape') cancel();
    if (e.key === 'Enter' && canSwitch && !switching) switchBranch();
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<div class="overlay" role="presentation" on:click={cancel} on:keydown={() => {}}>
  <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
  <div class="modal" role="dialog" on:click|stopPropagation on:keydown={() => {}}>
    <h2><GitBranch size={18} /> Switch Branch</h2>
    <p class="subtitle">Current: <strong style="color: {displayColor}">{currentBranch || 'HEAD'}</strong></p>

    {#if dirty}
      <div class="warning">
        <AlertTriangle size={14} />
        <span>You have uncommitted changes</span>
      </div>
    {/if}

    <!-- Search filter -->
    <div class="search-wrap">
      <Search size={14} />
      <!-- svelte-ignore a11y-autofocus -->
      <input
        type="text"
        class="search-input"
        placeholder="Filter branches..."
        bind:value={filter}
        autofocus
      />
    </div>

    <!-- Branch list -->
    <div class="branch-list">
      {#each filtered as branch}
        <button
          class="branch-option"
          class:selected={selectedBranch === branch.name}
          on:click={() => selectedBranch = branch.name}
        >
          <GitBranch size={12} />
          <span class="branch-name">{branch.name}</span>
        </button>
      {/each}
      {#if filtered.length === 0}
        <div class="empty">No other branches found</div>
      {/if}
    </div>

    {#if dirty}
      <div class="field toggle-field">
        <label class="toggle-label">
          <input type="checkbox" bind:checked={autoCommit} />
          <span>Auto-commit changes before switching</span>
        </label>
      </div>
    {/if}

    {#if error}
      <div class="error">{error}</div>
    {/if}

    <div class="actions">
      <button class="btn-cancel" on:click={cancel}>Cancel</button>
      {#if !dirty || autoCommit}
        <button
          class="btn-switch"
          on:click={switchBranch}
          disabled={!canSwitch || switching}
        >
          {switching ? 'Switching...' : 'Switch Branch'}
        </button>
      {/if}
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
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .subtitle {
    font-size: var(--text-body);
    color: var(--text-dim);
    margin: 0 0 20px;
  }

  .subtitle strong {
    font-family: var(--font-mono);
    font-size: 11px;
  }

  .warning {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 12px;
    background: color-mix(in srgb, var(--accent-amber) 8%, transparent);
    border: 1px solid color-mix(in srgb, var(--accent-amber) 20%, transparent);
    border-radius: var(--radius-md);
    color: var(--accent-amber);
    font-size: var(--text-body);
    margin-bottom: 16px;
  }

  .search-wrap {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: var(--sp-sm) 12px;
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    margin-bottom: 12px;
    color: var(--text-dim);
  }

  .search-input {
    flex: 1;
    background: none;
    border: none;
    color: var(--text-primary);
    font-family: var(--font-mono);
    font-size: 13px;
    outline: none;
  }

  .search-input::placeholder {
    color: var(--text-muted);
  }

  .branch-list {
    display: flex;
    flex-direction: column;
    gap: var(--sp-2xs);
    max-height: 240px;
    overflow-y: auto;
    margin-bottom: 16px;
  }

  .branch-option {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 7px 12px;
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    cursor: pointer;
    transition: all 100ms ease-out;
    font-family: var(--font-ui);
    color: var(--text-primary);
    font-size: 13px;
    text-align: left;
  }

  .branch-option:hover {
    border-color: var(--border-emphasis);
    background: var(--bg-active);
  }

  .branch-option.selected {
    border-color: var(--accent-green);
    background: color-mix(in srgb, var(--accent-green) 8%, transparent);
  }

  .branch-name {
    font-family: var(--font-mono);
    font-size: var(--text-body);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .empty {
    color: var(--text-dim);
    font-size: var(--text-body);
    padding: 12px;
    text-align: center;
  }

  .field { margin-bottom: 16px; }

  .toggle-field { margin-bottom: 12px; }

  .toggle-label {
    display: flex;
    align-items: center;
    gap: 8px;
    cursor: pointer;
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

  .btn-switch {
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

  .btn-switch:hover { opacity: 0.9; }
  .btn-switch:disabled { opacity: 0.4; cursor: not-allowed; }
</style>
