<script>
  import { onMount, createEventDispatcher } from 'svelte';
  import { Play, Pause, Square, RotateCcw } from 'lucide-svelte';
  import { ListModels } from '../../../wailsjs/go/main/App.js';

  export let executionStatus = 'idle';
  export let nodeProgress = { completed: 0, total: 0 };
  export let repoPath = '';

  const dispatch = createEventDispatcher();

  let models = [];
  let selectedModel = '';

  onMount(async () => {
    try {
      const modelList = await ListModels();
      models = modelList || [];
      const defaultModel = models.find(m => m.isDefault);
      selectedModel = defaultModel ? defaultModel.id : (models[0]?.id || '');
    } catch (e) {
      console.error('Failed to load models:', e);
    }
  });

  $: isIdle = executionStatus === 'idle';
  $: isRunning = executionStatus === 'running';
  $: isPaused = executionStatus === 'paused';
  $: isDone = executionStatus === 'complete' || executionStatus === 'failed';

  $: runEnabled = (isIdle || isPaused || isDone) && repoPath;
  $: pauseEnabled = isRunning;
  $: stopEnabled = isRunning || isPaused;

  function handleRun() {
    if (isPaused) {
      dispatch('resume');
    } else {
      dispatch('start', { model: selectedModel });
    }
  }
</script>

<div class="execution-bar">
  <div class="selectors">
    <select class="bar-select" bind:value={selectedModel} disabled={isRunning || isPaused}>
      {#each models as m}
        <option value={m.id}>{m.displayName}</option>
      {/each}
    </select>
  </div>

  <div class="controls">
    <button
      class="ctrl-btn run"
      class:active={isRunning}
      on:click={handleRun}
      disabled={!runEnabled}
      title={isPaused ? 'Resume' : isDone ? 'Restart' : 'Run'}
    >
      {#if isDone}
        <RotateCcw size={14} />
      {:else}
        <Play size={14} />
      {/if}
      <span>{isPaused ? 'Resume' : isDone ? 'Restart' : 'Run'}</span>
    </button>

    <button
      class="ctrl-btn pause"
      class:active={isPaused}
      on:click={() => dispatch('pause')}
      disabled={!pauseEnabled}
      title="Pause"
    >
      <Pause size={14} />
      <span>Pause</span>
    </button>

    <button
      class="ctrl-btn stop"
      class:active={isDone && executionStatus === 'failed'}
      on:click={() => dispatch('stop')}
      disabled={!stopEnabled}
      title="Stop"
    >
      <Square size={14} />
      <span>Stop</span>
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
    gap: 8px;
    flex-shrink: 0;
  }

  .selectors {
    display: flex;
    gap: 6px;
  }

  .bar-select {
    padding: 4px 8px;
    background: var(--bg-deepest);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-sm);
    color: var(--text-primary);
    font-family: var(--font-mono);
    font-size: 11px;
    outline: none;
    max-width: 180px;
  }

  .bar-select:focus { border-color: var(--accent-green); }
  .bar-select:disabled { opacity: 0.5; }

  .controls {
    display: flex;
    gap: 4px;
  }

  .ctrl-btn {
    display: flex;
    align-items: center;
    gap: 5px;
    padding: 4px 10px;
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-sm);
    font-family: var(--font-mono);
    font-size: 11px;
    font-weight: 500;
    cursor: pointer;
    transition: background 100ms ease, color 100ms ease, border-color 100ms ease, box-shadow 100ms ease;
  }

  .ctrl-btn:disabled {
    opacity: 0.3;
    cursor: not-allowed;
  }

  /* Run: cyan idle, green when running */
  .ctrl-btn.run {
    color: var(--accent-cyan);
    border-color: color-mix(in srgb, var(--accent-cyan) 30%, transparent);
  }
  .ctrl-btn.run:hover:not(:disabled) {
    background: color-mix(in srgb, var(--accent-cyan) 10%, transparent);
    border-color: var(--accent-cyan);
    box-shadow: 0 0 8px color-mix(in srgb, var(--accent-cyan) 25%, transparent);
  }
  .ctrl-btn.run.active {
    color: var(--accent-green, #00e57a);
    border-color: var(--accent-green, #00e57a);
    background: rgba(0, 229, 122, 0.1);
    box-shadow: 0 0 8px rgba(0, 229, 122, 0.3);
  }

  /* Pause: amber idle, amber glow when paused */
  .ctrl-btn.pause {
    color: var(--accent-amber, #f0a500);
    border-color: rgba(240, 165, 0, 0.3);
  }
  .ctrl-btn.pause:hover:not(:disabled) {
    background: rgba(240, 165, 0, 0.1);
    border-color: var(--accent-amber, #f0a500);
    box-shadow: 0 0 8px rgba(240, 165, 0, 0.25);
  }
  .ctrl-btn.pause.active {
    color: var(--accent-amber, #f0a500);
    border-color: var(--accent-amber, #f0a500);
    background: rgba(240, 165, 0, 0.15);
    box-shadow: 0 0 8px rgba(240, 165, 0, 0.4);
  }

  /* Stop: orange idle, red when stopped/failed */
  .ctrl-btn.stop {
    color: var(--accent-orange);
    border-color: color-mix(in srgb, var(--accent-orange) 30%, transparent);
  }
  .ctrl-btn.stop:hover:not(:disabled) {
    background: color-mix(in srgb, var(--accent-orange) 10%, transparent);
    border-color: var(--accent-orange);
    box-shadow: 0 0 8px color-mix(in srgb, var(--accent-orange) 25%, transparent);
  }
  .ctrl-btn.stop.active {
    color: var(--accent-red, #f85149);
    border-color: var(--accent-red, #f85149);
    background: rgba(248, 81, 73, 0.1);
    box-shadow: 0 0 8px rgba(248, 81, 73, 0.3);
  }

  .progress {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-left: auto;
  }

  .progress-text {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--text-primary);
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
