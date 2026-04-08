<script>
  import { onMount, createEventDispatcher } from 'svelte';
  import { fade, fly } from 'svelte/transition';
  import { cubicOut } from 'svelte/easing';
  import { ListRepoChoices } from '../../../wailsjs/go/main/App.js';

  const dispatch = createEventDispatcher();

  let repos = [];
  let loading = true;
  let error = '';

  onMount(async () => {
    try {
      repos = await ListRepoChoices() || [];
    } catch (e) {
      error = 'Failed to load repositories';
    }
    loading = false;
  });

  function selectRepo(repo) {
    dispatch('select', { path: repo.path, name: repo.name, branch: repo.branch });
  }

  function cancel() {
    dispatch('cancel');
  }

  function handleKeydown(e) {
    if (e.key === 'Escape') cancel();
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<div class="overlay" on:click={cancel} on:keydown={() => {}} transition:fade={{ duration: 150 }}>
  <div
    class="modal"
    on:click|stopPropagation
    on:keydown={() => {}}
    transition:fly={{ y: 16, duration: 250, easing: cubicOut }}
  >
    <h2>Select Repository</h2>
    <p class="subtitle">Choose a repo to scope this workflow session</p>

    <div class="repo-list">
      {#if loading}
        <div class="loading-state">Loading repositories...</div>
      {:else if error}
        <div class="error-state">{error}</div>
      {:else if repos.length === 0}
        <div class="empty-state">No repositories found</div>
      {:else}
        {#each repos as repo, i}
          <button
            class="repo-item"
            on:click={() => selectRepo(repo)}
            transition:fly={{ y: 12, duration: 200, delay: i * 40, easing: cubicOut }}
          >
            <div class="repo-info">
              <span class="repo-name">{repo.name}</span>
              <span class="repo-branch">{repo.branch}</span>
            </div>
            <span class="repo-path">{repo.path}</span>
          </button>
        {/each}
      {/if}
    </div>

    <div class="actions">
      <button class="btn-cancel" on:click={cancel}>Cancel</button>
    </div>
  </div>
</div>

<style>
  .overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.6);
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
    width: 480px;
    max-height: 80vh;
    overflow-y: auto;
  }

  h2 {
    font-size: 18px;
    font-weight: 600;
    color: var(--text-primary);
    margin: 0 0 4px;
    font-family: var(--font-mono);
  }

  .subtitle {
    font-size: 12px;
    color: var(--text-dim);
    margin: 0 0 24px;
    font-family: var(--font-mono);
  }

  .repo-list {
    display: flex;
    flex-direction: column;
    gap: 4px;
    max-height: 360px;
    overflow-y: auto;
    margin-bottom: 20px;
  }

  .repo-item {
    display: flex;
    flex-direction: column;
    gap: 4px;
    padding: 10px 12px;
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    cursor: pointer;
    transition: background 100ms ease-out, border-color 100ms ease-out;
    text-align: left;
    width: 100%;
  }

  .repo-item:hover {
    border-color: var(--accent-green);
    background: rgba(0, 229, 122, 0.06);
  }

  .repo-item:focus-visible {
    outline: 2px solid var(--accent-green);
    outline-offset: -2px;
  }

  .repo-info {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .repo-name {
    font-family: var(--font-mono);
    font-weight: 600;
    font-size: 13px;
    color: var(--text-primary);
  }

  .repo-branch {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--text-dim);
  }

  .repo-path {
    font-family: var(--font-mono);
    font-size: 10px;
    color: var(--text-muted);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .loading-state,
  .error-state,
  .empty-state {
    font-family: var(--font-mono);
    font-size: 12px;
    color: var(--text-dim);
    padding: 24px;
    text-align: center;
  }

  .error-state {
    color: var(--accent-red, #f85149);
  }

  .actions {
    display: flex;
    justify-content: flex-end;
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
    transition: border-color 100ms ease-out, color 100ms ease-out;
  }

  .btn-cancel:hover {
    border-color: var(--border-emphasis);
    color: var(--text-primary);
  }

  .btn-cancel:focus-visible {
    outline: 2px solid var(--accent-green);
    outline-offset: -2px;
  }
</style>
