<script>
  import { createEventDispatcher } from 'svelte';
  import { Layout } from 'lucide-svelte';

  /** @type {Array<{ id: string; name: string; description?: string; nodes?: unknown[] }>} */
  export let templates = [];

  /** @type {import('svelte').EventDispatcher<{ select: string }>} */
  const dispatch = createEventDispatcher();
</script>

<div class="template-picker">
  <div class="picker-grid">
    {#each templates as tmpl}
      <div class="template-card">
        <div class="card-icon"><Layout size={18} /></div>
        <span class="card-name">{tmpl.name}</span>
        {#if tmpl.description}
          <span class="card-desc">{tmpl.description}</span>
        {/if}
        <span class="card-meta">{tmpl.nodes?.length || 0} nodes</span>
        <button class="card-use" on:click={() => dispatch('select', tmpl.id)}>Use</button>
      </div>
    {/each}
    {#if templates.length === 0}
      <div class="empty">No templates available</div>
    {/if}
  </div>
</div>

<style>
  .template-picker {
    padding: 4px 0;
  }

  .picker-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: var(--sp-xs);
    padding: var(--sp-xs);
  }

  .template-card {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;
    padding: 10px 8px;
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md, 6px);
    transition: border-color 100ms ease;
  }

  .template-card:hover {
    border-color: var(--border-emphasis);
  }

  .card-icon {
    color: var(--accent-purple, #9d6fff);
    margin-bottom: 2px;
  }

  .card-name {
    font-family: var(--font-mono);
    font-size: 11px;
    font-weight: 600;
    color: var(--text-primary);
    text-align: center;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    width: 100%;
  }

  .card-desc {
    font-family: var(--font-mono);
    font-size: 9px;
    color: var(--text-muted);
    text-align: center;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    width: 100%;
  }

  .card-meta {
    font-family: var(--font-mono);
    font-size: 9px;
    color: var(--text-muted);
  }

  .card-use {
    padding: var(--sp-2xs) 12px;
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-sm);
    color: var(--accent-green);
    font-family: var(--font-mono);
    font-size: var(--text-label);
    cursor: pointer;
    transition: background 100ms ease, border-color 100ms ease;
    margin-top: 2px;
  }

  .card-use:hover {
    background: var(--bg-active);
    border-color: var(--accent-green);
  }

  .empty {
    grid-column: 1 / -1;
    text-align: center;
    padding: 20px;
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--text-muted);
  }
</style>
