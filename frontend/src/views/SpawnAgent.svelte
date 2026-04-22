<script>
  import { onMount, createEventDispatcher } from 'svelte';
  import { fade } from 'svelte/transition';
  import { cubicOut } from 'svelte/easing';
  import { ListRepoChoices, SpawnAgent, ListModels } from '../../wailsjs/go/main/App.js';
  import { errorMessage } from '../lib/errorMessage';

  /** @type {import('svelte').EventDispatcher<{ spawned: { target: string; repo: unknown; model: string }; cancel: void }>} */
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

  let repos = [];
  let selectedRepo = null;
  let model = '';
  let spawning = false;
  let error = '';

  let models = [];

  onMount(async () => {
    try {
      const [repoList, modelList] = await Promise.all([
        ListRepoChoices(),
        ListModels(),
      ]);
      repos = repoList || [];
      models = (modelList || []).map(m => ({ value: m.id, label: m.displayName }));
      const defaultModel = modelList?.find(m => m.isDefault);
      model = defaultModel ? defaultModel.id : (models[0]?.value || '');
    } catch {
      error = 'Failed to load repos';
    }
  });

  async function spawn() {
    if (!selectedRepo) return;
    spawning = true;
    error = '';
    try {
      const target = await SpawnAgent(selectedRepo.path, model);
      dispatch('spawned', { target, repo: selectedRepo, model });
    } catch (e) {
      error = errorMessage(e) || 'Failed to spawn agent';
      spawning = false;
    }
  }

  function cancel() {
    dispatch('cancel');
  }

  /** @param {KeyboardEvent} e */
  function handleKeydown(e) {
    if (e.key === 'Escape') cancel();
    if (e.key === 'Enter' && selectedRepo) spawn();
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<div class="overlay" role="presentation" on:click={cancel} on:keydown={() => {}} transition:fade={backdropFadeProps}>
  <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
  <div class="modal" role="dialog" on:click|stopPropagation on:keydown={() => {}} transition:fade={cardFadeProps}>
    <h2>Spawn Agent</h2>
    <p class="subtitle">Start a new Claude session in a tmux terminal</p>

    <!-- Repo selector -->
    <div class="field">
      <!-- svelte-ignore a11y-label-has-associated-control -->
      <label>Repository</label>
      <div class="repo-list">
        {#each repos as repo}
          <button
            class="repo-option"
            class:selected={selectedRepo?.path === repo.path}
            on:click={() => selectedRepo = repo}
          >
            <span class="repo-name">{repo.name}</span>
            <span class="repo-branch">{repo.branch}</span>
          </button>
        {/each}
        {#if repos.length === 0}
          <div class="empty">No repos found in dev directory</div>
        {/if}
      </div>
    </div>

    <!-- Model selector -->
    <div class="field">
      <!-- svelte-ignore a11y-label-has-associated-control -->
      <label>Model</label>
      <div class="model-list">
        {#each models as m}
          <button
            class="model-option"
            class:selected={model === m.value}
            on:click={() => model = m.value}
          >
            {m.label}
          </button>
        {/each}
      </div>
    </div>

    {#if error}
      <div class="error">{error}</div>
    {/if}

    <div class="actions">
      <button class="btn-cancel" on:click={cancel}>Cancel</button>
      <button
        class="btn-spawn"
        on:click={spawn}
        disabled={!selectedRepo || spawning}
      >
        {spawning ? 'Spawning...' : 'Spawn Agent'}
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

  .repo-list {
    display: flex;
    flex-direction: column;
    gap: 4px;
    max-height: 200px;
    overflow-y: auto;
  }

  .repo-option {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 8px 12px;
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

  .repo-option:hover {
    border-color: var(--border-emphasis);
    background: var(--bg-active);
  }

  .repo-option.selected {
    border-color: var(--accent-green);
    background: color-mix(in srgb, var(--accent-green) 8%, transparent);
  }

  .repo-name {
    font-family: var(--font-mono);
    font-weight: 500;
  }

  .repo-branch {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--text-dim);
  }

  .model-list {
    display: flex;
    gap: 8px;
  }

  .model-option {
    flex: 1;
    padding: 8px;
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    cursor: pointer;
    transition: all 100ms ease-out;
    font-family: var(--font-ui);
    font-size: var(--text-body);
    color: var(--text-primary);
  }

  .model-option:hover {
    border-color: var(--border-emphasis);
  }

  .model-option.selected {
    border-color: var(--accent-green);
    background: color-mix(in srgb, var(--accent-green) 8%, transparent);
  }

  .error {
    font-size: var(--text-body);
    color: var(--accent-red);
    margin-bottom: 16px;
  }

  .empty {
    color: var(--text-dim);
    font-size: var(--text-body);
    padding: 12px;
    text-align: center;
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

  .btn-spawn {
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

  .btn-spawn:hover { opacity: 0.9; }
  .btn-spawn:disabled { opacity: 0.4; cursor: not-allowed; }
</style>
