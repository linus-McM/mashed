<script lang="ts">
  // Full-width streaming commit output panel (steps, spinner, error + AI
  // explanation). Extracted from views/AgentDetail.svelte (spec R31) — markup
  // and scoped styles moved verbatim. The parent owns the panel state and the
  // git:commit:progress listener. Distinct from components/feed/CommitOutputPanel
  // (different layout/styles).
  type CommitPanelState = {
    lines: Array<{ step: string; output: string }>;
    error: string | null;
    explanation: string | null;
    done: boolean;
  };

  export let panel: CommitPanelState;
  export let onClose: () => void;
</script>

<div class="ws-commit-panel">
  <div class="ws-commit-header">
    <span class="ws-commit-title">
      {#if panel.done && !panel.error}
        Committed
      {:else if panel.error}
        Commit Failed
      {:else}
        Committing...
      {/if}
    </span>
    {#if panel.done}
      <button class="ws-commit-close" on:click={() => onClose()}>×</button>
    {/if}
  </div>
  <div class="ws-commit-body">
    {#each panel.lines as line}
      <div class="ws-commit-line">
        <span class="ws-commit-step">{line.step}</span>
        {#if line.output}
          <pre class="ws-commit-output">{line.output}</pre>
        {/if}
      </div>
    {/each}
    {#if !panel.done && !panel.error}
      <div class="ws-commit-line ws-commit-active">
        <span class="ws-commit-spinner" />
      </div>
    {/if}
  </div>
  {#if panel.error}
    <div class="ws-commit-error-section">
      <div class="ws-commit-error-label">Error</div>
      <pre class="ws-commit-error-text">{panel.error}</pre>
      {#if panel.explanation}
        <div class="ws-commit-explain-label">Why this happened</div>
        <div class="ws-commit-explain-text">{panel.explanation}</div>
      {:else if !panel.done}
        <div class="ws-commit-explain-loading">Analyzing failure...</div>
      {/if}
    </div>
  {/if}
</div>

<style>
  /* Full-width commit output panel — between workspace and bottom bar */
  .ws-commit-panel {
    border-top: 2px solid var(--accent-green);
    background: var(--bg-deepest);
    display: flex;
    flex-direction: column;
    max-height: 240px;
    overflow: hidden;
    flex-shrink: 0;
    width: 100%;
  }

  .ws-commit-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 4px 8px;
    background: var(--bg-surface);
    border-bottom: 1px solid var(--border-subtle);
    flex-shrink: 0;
  }

  .ws-commit-title {
    font-family: var(--font-mono);
    font-size: var(--text-label);
    font-weight: 600;
    color: var(--text-dim);
  }

  .ws-commit-close {
    background: none;
    border: none;
    color: var(--text-muted);
    font-size: var(--text-data);
    cursor: pointer;
    padding: 0 2px;
    line-height: 1;
  }

  .ws-commit-close:hover { color: var(--text-primary); }

  .ws-commit-body {
    padding: 4px 8px;
    overflow-y: auto;
    flex: 1;
    min-height: 0;
  }

  .ws-commit-line {
    padding: 1px 0;
  }

  .ws-commit-step {
    font-family: var(--font-mono);
    font-size: var(--text-label);
    color: var(--accent-green);
  }

  .ws-commit-output {
    font-family: var(--font-mono);
    font-size: 9px;
    color: var(--text-dim);
    margin: var(--sp-2xs) 0 var(--sp-2xs) 0;
    padding: var(--sp-2xs) var(--sp-xs);
    background: rgba(0, 0, 0, 0.3);
    border-radius: var(--radius-sm);
    white-space: pre-wrap;
    word-break: break-word;
    max-height: 50px;
    overflow-y: auto;
  }

  .ws-commit-active {
    display: flex;
    align-items: center;
    gap: var(--sp-sm);
  }

  .ws-commit-spinner {
    display: inline-block;
    width: 7px;
    height: 7px;
    border: 1.5px solid var(--accent-green);
    border-top-color: transparent;
    border-radius: 50%;
    animation: ws-commit-spin 0.6s linear infinite;
  }

  @keyframes ws-commit-spin {
    to { transform: rotate(360deg); }
  }

  .ws-commit-error-section {
    border-top: 1px solid color-mix(in srgb, var(--accent-red) 20%, transparent);
    padding: 4px 8px;
    background: color-mix(in srgb, var(--accent-red) 4%, transparent);
    flex-shrink: 0;
  }

  .ws-commit-error-label {
    font-family: var(--font-mono);
    font-size: 8px;
    font-weight: 600;
    color: var(--accent-red);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    margin-bottom: var(--sp-2xs);
  }

  .ws-commit-error-text {
    font-family: var(--font-mono);
    font-size: 9px;
    color: var(--accent-red);
    white-space: pre-wrap;
    word-break: break-word;
    margin: 0 0 var(--sp-xs) 0;
    padding: var(--sp-2xs) var(--sp-xs);
    background: color-mix(in srgb, var(--accent-red) 6%, transparent);
    border-radius: var(--radius-sm);
    border: 1px solid color-mix(in srgb, var(--accent-red) 15%, transparent);
    max-height: 50px;
    overflow-y: auto;
  }

  .ws-commit-explain-label {
    font-family: var(--font-mono);
    font-size: 8px;
    font-weight: 600;
    color: var(--accent-amber);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    margin-bottom: var(--sp-2xs);
  }

  .ws-commit-explain-text {
    font-size: var(--text-label);
    color: var(--text-primary);
    line-height: 1.4;
    padding: var(--sp-xs);
    background: color-mix(in srgb, var(--accent-amber) 6%, transparent);
    border-radius: var(--radius-sm);
    border: 1px solid color-mix(in srgb, var(--accent-amber) 15%, transparent);
  }

  .ws-commit-explain-loading {
    font-family: var(--font-mono);
    font-size: 9px;
    color: var(--text-muted);
    font-style: italic;
  }
</style>
