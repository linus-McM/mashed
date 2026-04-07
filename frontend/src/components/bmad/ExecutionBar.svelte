<script>
  import { createEventDispatcher, onMount } from 'svelte';
  import { Play, Pause, Square, RotateCcw } from 'lucide-svelte';
  import { ListRepoChoices } from '../../../wailsjs/go/main/App.js';

  export let executionStatus = 'idle';
  export let nodeProgress = { completed: 0, total: 0 };

  const dispatch = createEventDispatcher();

  const models = [
    'claude-opus-4-6',
    'claude-sonnet-4-20250514',
    'claude-haiku-3.5',
  ];

  let repoChoices = [];
  let selectedRepo = '';
  let selectedModel = models[0];

  onMount(async () => {
    try {
      repoChoices = await ListRepoChoices();
      if (repoChoices.length > 0) {
        selectedRepo = repoChoices[0].path || repoChoices[0];
      }
    } catch {}
  });

  $: isIdle = executionStatus === 'idle';
  $: isRunning = executionStatus === 'running';
  $: isPaused = executionStatus === 'paused';
  $: isDone = executionStatus === 'complete' || executionStatus === 'failed';

  $: runEnabled = isIdle || isPaused || isDone;
  $: pauseEnabled = isRunning;
  $: stopEnabled = isRunning || isPaused;

  function handleRun() {
    if (isPaused) {
      dispatch('resume');
    } else {
      dispatch('start', { repoPath: selectedRepo, model: selectedModel });
    }
  }
</script>

<div class="execution-bar">
  <div class="selectors">
    <select class="bar-select" bind:value={selectedRepo} disabled={isRunning || isPaused}>
      {#each repoChoices as repo}
        <option value={repo.path || repo}>{repo.name || repo}</option>
      {/each}
      {#if repoChoices.length === 0}
        <option value="">No repos</option>
      {/if}
    </select>

    <select class="bar-select" bind:value={selectedModel} disabled={isRunning || isPaused}>
      {#each models as m}
        <option value={m}>{m}</option>
      {/each}
    </select>
  </div>

  <div class="controls">
    <button
      class="ctrl-btn run"
      on:click={handleRun}
      disabled={!runEnabled}
      title={isPaused ? 'Resume' : isDone ? 'Restart' : 'Run'}
    >
      {#if isDone}
        <RotateCcw size={13} />
      {:else}
        <Play size={13} />
      {/if}
      <span>{isPaused ? 'Resume' : isDone ? 'Restart' : 'Run'}</span>
    </button>

    <button
      class="ctrl-btn pause"
      on:click={() => dispatch('pause')}
      disabled={!pauseEnabled}
      title="Pause"
    >
      <Pause size={13} />
    </button>

    <button
      class="ctrl-btn stop"
      on:click={() => dispatch('stop')}
      disabled={!stopEnabled}
      title="Stop"
    >
      <Square size={13} />
    </button>
  </div>

  <div class="progress">
    {#if nodeProgress.total > 0}
      <span class="progress-text">
        {nodeProgress.completed}/{nodeProgress.total} done
      </span>
      <div class="progress-bar">
        <div
          class="progress-fill"
          class:failed={executionStatus === 'failed'}
          style="width: {(nodeProgress.completed / nodeProgress.total) * 100}%"
        />
      </div>
    {/if}
  </div>
</div>

<style>
  .execution-bar {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 6px 10px;
    background: var(--bg-surface);
    border-top: 1px solid var(--border-subtle);
    flex-shrink: 0;
    height: 38px;
  }

  .selectors {
    display: flex;
    gap: 6px;
  }

  .bar-select {
    padding: 3px 6px;
    background: var(--bg-deepest);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-sm);
    color: var(--text-primary);
    font-family: var(--font-mono);
    font-size: 10px;
    outline: none;
    max-width: 160px;
  }

  .bar-select:focus { border-color: var(--accent-green); }
  .bar-select:disabled { opacity: 0.5; }

  .controls {
    display: flex;
    gap: 4px;
    padding: 0 8px;
    border-left: 1px solid var(--border-subtle);
    border-right: 1px solid var(--border-subtle);
  }

  .ctrl-btn {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 3px 8px;
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-sm);
    color: var(--text-dim);
    font-family: var(--font-mono);
    font-size: 10px;
    cursor: pointer;
    transition: background 100ms ease, color 100ms ease, border-color 100ms ease;
  }

  .ctrl-btn:hover:not(:disabled) {
    background: var(--bg-active);
    color: var(--text-primary);
  }

  .ctrl-btn:disabled {
    opacity: 0.35;
    cursor: not-allowed;
  }

  .ctrl-btn.run:hover:not(:disabled) {
    border-color: var(--accent-green);
    color: var(--accent-green);
  }

  .ctrl-btn.stop:hover:not(:disabled) {
    border-color: var(--accent-red, #f85149);
    color: var(--accent-red, #f85149);
  }

  .progress {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-left: auto;
  }

  .progress-text {
    font-family: var(--font-mono);
    font-size: 10px;
    color: var(--text-muted);
    white-space: nowrap;
  }

  .progress-bar {
    width: 80px;
    height: 4px;
    background: var(--bg-deepest);
    border-radius: 2px;
    overflow: hidden;
  }

  .progress-fill {
    height: 100%;
    background: var(--accent-green);
    border-radius: 2px;
    transition: width 300ms ease;
  }

  .progress-fill.failed {
    background: var(--accent-red, #f85149);
  }
</style>
