<script lang="ts">
  // Streaming commit output panel (steps, spinner, error + AI explanation).
  // Extracted from views/NotificationFeed.svelte (R31). The parent owns the
  // panel state and the git:commit:progress listener.
  import { createEventDispatcher } from 'svelte';
  import type { CommitPanel } from '../../lib/feed/repoTree';

  export let panel: CommitPanel;
  /** Current border colour for this repo (parent's getRepoColor(repo.name)). */
  export let repoColor: string;

  const dispatch = createEventDispatcher<{ close: void }>();
</script>

<div class="commit-panel" style="border-color: {repoColor}">
  <div class="commit-panel-header">
    <span class="commit-panel-title">
      {#if panel.done && !panel.error}
        Committed
      {:else if panel.error}
        Commit Failed
      {:else}
        Committing...
      {/if}
    </span>
    {#if panel.done}
      <button class="commit-panel-close" on:click|stopPropagation={() => dispatch('close')}>×</button>
    {/if}
  </div>
  <div class="commit-panel-body">
    {#each panel.lines as line}
      <div class="commit-line">
        <span class="commit-step">{line.step}</span>
        {#if line.output}
          <pre class="commit-output">{line.output}</pre>
        {/if}
      </div>
    {/each}
    {#if !panel.done && !panel.error}
      <div class="commit-line commit-active">
        <span class="commit-spinner" />
      </div>
    {/if}
  </div>
  {#if panel.error}
    <div class="commit-error-section">
      <div class="commit-error-label">Error</div>
      <pre class="commit-error-text">{panel.error}</pre>
      {#if panel.explanation}
        <div class="commit-explain-label">Why this happened</div>
        <div class="commit-explain-text">{panel.explanation}</div>
      {:else if !panel.done}
        <div class="commit-explain-loading">Analyzing failure...</div>
      {/if}
    </div>
  {/if}
</div>

<style>
  /* Commit output panel */
  .commit-panel {
    border-top: 2px solid var(--border-subtle);
    background: var(--bg-deepest);
    max-height: 220px;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .commit-panel-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: var(--sp-xs) var(--sp-lg);
    background: var(--bg-surface);
    border-bottom: 1px solid var(--border-subtle);
    flex-shrink: 0;
  }

  .commit-panel-title {
    font-family: var(--font-mono);
    font-size: 11px;
    font-weight: 600;
    color: var(--text-dim);
  }

  .commit-panel-close {
    background: none;
    border: none;
    color: var(--text-muted);
    font-size: 16px;
    cursor: pointer;
    padding: 0 2px;
    line-height: 1;
  }

  .commit-panel-close:hover { color: var(--text-primary); }

  .commit-panel-body {
    padding: var(--sp-xs) var(--sp-lg);
    overflow-y: auto;
    flex: 1;
    min-height: 0;
  }

  .commit-line {
    padding: 2px 0;
  }

  .commit-step {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--accent-green);
  }

  .commit-output {
    font-family: var(--font-mono);
    font-size: var(--text-label);
    color: var(--text-dim);
    margin: 2px 0 4px 0;
    padding: 4px 8px;
    background: rgba(0, 0, 0, 0.25);
    border-radius: var(--radius-sm);
    white-space: pre-wrap;
    word-break: break-word;
    max-height: 60px;
    overflow-y: auto;
  }

  .commit-active {
    display: flex;
    align-items: center;
    gap: var(--sp-sm);
  }

  .commit-spinner {
    display: inline-block;
    width: 8px;
    height: 8px;
    border: 1.5px solid var(--accent-green);
    border-top-color: transparent;
    border-radius: 50%;
    animation: commit-spin 0.6s linear infinite;
  }

  @keyframes commit-spin {
    to { transform: rotate(360deg); }
  }

  /* Error section */
  .commit-error-section {
    border-top: 1px solid color-mix(in srgb, var(--accent-red) 20%, transparent);
    padding: var(--sp-xs) var(--sp-lg);
    background: color-mix(in srgb, var(--accent-red) 4%, transparent);
    flex-shrink: 0;
  }

  .commit-error-label {
    font-family: var(--font-mono);
    font-size: 9px;
    font-weight: 600;
    color: var(--accent-red);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    margin-bottom: 4px;
  }

  .commit-error-text {
    font-family: var(--font-mono);
    font-size: var(--text-label);
    color: var(--accent-red);
    white-space: pre-wrap;
    word-break: break-word;
    margin: 0 0 8px 0;
    padding: 4px 8px;
    background: color-mix(in srgb, var(--accent-red) 6%, transparent);
    border-radius: var(--radius-sm);
    border: 1px solid color-mix(in srgb, var(--accent-red) 15%, transparent);
    max-height: 60px;
    overflow-y: auto;
  }

  .commit-explain-label {
    font-family: var(--font-mono);
    font-size: 9px;
    font-weight: 600;
    color: var(--accent-amber);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    margin-bottom: 4px;
  }

  .commit-explain-text {
    font-size: 11px;
    color: var(--text-primary);
    line-height: 1.5;
    padding: var(--sp-sm);
    background: color-mix(in srgb, var(--accent-amber) 6%, transparent);
    border-radius: var(--radius-sm);
    border: 1px solid color-mix(in srgb, var(--accent-amber) 15%, transparent);
  }

  .commit-explain-loading {
    font-family: var(--font-mono);
    font-size: var(--text-label);
    color: var(--text-muted);
    font-style: italic;
  }
</style>
