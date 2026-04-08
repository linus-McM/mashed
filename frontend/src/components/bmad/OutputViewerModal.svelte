<script>
  import { createEventDispatcher } from 'svelte';
  import { fade } from 'svelte/transition';
  import { X } from 'lucide-svelte';

  export let show = false;
  export let label = '';
  export let content = '';
  export let loading = false;

  const dispatch = createEventDispatcher();

  function close() {
    dispatch('close');
  }

  function onKeydown(e) {
    if (show && e.key === 'Escape') close();
  }
</script>

<svelte:window on:keydown={onKeydown} />

{#if show}
  <div class="output-overlay" transition:fade={{ duration: 150 }} on:click={close} on:keydown={() => {}}>
    <div class="output-modal" on:click|stopPropagation on:keydown={() => {}}>
      <div class="output-header">
        <span class="output-title">Output: {label}</span>
        <button class="output-close" on:click={close} title="Close">
          <X size={14} />
        </button>
      </div>
      <div class="output-body">
        {#if loading}
          <div class="output-loading">Loading output...</div>
        {:else if content}
          <pre class="output-content">{content}</pre>
        {:else}
          <div class="output-empty">No output captured</div>
        {/if}
      </div>
    </div>
  </div>
{/if}

<style>
  .output-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.6);
    z-index: 100;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .output-modal {
    width: 75vw;
    max-height: 80vh;
    background: var(--bg-deepest);
    border: 1px solid var(--border-emphasis);
    border-radius: var(--radius-lg, 8px);
    display: flex;
    flex-direction: column;
    overflow: hidden;
    box-shadow: 0 16px 48px rgba(0, 0, 0, 0.5);
  }

  .output-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 8px 12px;
    background: var(--bg-surface);
    border-bottom: 1px solid var(--border-subtle);
    flex-shrink: 0;
  }

  .output-title {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--text-dim);
  }

  .output-close {
    background: none;
    border: none;
    color: var(--text-muted);
    cursor: pointer;
    padding: 2px;
    display: flex;
    align-items: center;
    border-radius: var(--radius-sm);
    transition: color 100ms ease;
  }

  .output-close:hover { color: var(--text-primary); }

  .output-body {
    flex: 1;
    overflow-y: auto;
    padding: 12px;
  }

  .output-content {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--text-primary);
    line-height: 1.5;
    white-space: pre-wrap;
    word-break: break-all;
    margin: 0;
  }

  .output-loading, .output-empty {
    font-family: var(--font-mono);
    font-size: 12px;
    color: var(--text-muted);
    text-align: center;
    padding: 40px 0;
  }
</style>
