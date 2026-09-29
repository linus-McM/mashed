<script>
  // Full-screen tmux terminal modal. Extracted from WorkflowBuilder.svelte
  // (spec R31) — markup and scoped styles moved verbatim. The parent keeps
  // the `{#if showTerminalModal}` guard and owns the open/close state.
  import Terminal from '../Terminal.svelte';
  import { parseFriendlyTarget } from '../../lib/bmadSessionName';

  /** @type {string} */
  export let terminalTarget = '';
  /** @type {string} */
  export let repoPath = '';
  /** @type {() => void} */
  export let onClose = () => {};

  $: parsedTerminalTarget = parseFriendlyTarget(terminalTarget);
</script>

<div class="terminal-overlay" role="presentation" on:click={() => onClose()} on:keydown={() => {}}>
  <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
  <div class="terminal-modal" role="dialog" on:click|stopPropagation on:keydown={() => {}}>
    <div class="terminal-modal-header">
      <span class="terminal-modal-title" title={terminalTarget}>
        {#if parsedTerminalTarget}
          <span class="tmt-prefix">Terminal</span>
          <span class="tmt-dash">—</span>
          <span class="tmt-repo">{parsedTerminalTarget.repo}</span>
          <span class="tmt-sep">·</span>
          <span class="tmt-branch">{parsedTerminalTarget.branch}</span>
          <span class="tmt-sep">·</span>
          <span class="tmt-label">{parsedTerminalTarget.label}</span>
        {:else}
          {terminalTarget}
        {/if}
      </span>
      <button class="terminal-modal-close" on:click={() => onClose()}>&times;</button>
    </div>
    <div class="terminal-modal-body">
      {#key terminalTarget}
        <Terminal paneTarget={terminalTarget} {repoPath} />
      {/key}
    </div>
  </div>
</div>

<style>
  /* Terminal modal */
  .terminal-overlay {
    position: fixed;
    inset: 0;
    background: var(--overlay-backdrop);
    z-index: 100;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .terminal-modal {
    width: 80vw;
    height: 70vh;
    background: var(--bg-deepest);
    border: 1px solid var(--border-emphasis);
    border-radius: var(--radius-lg, 8px);
    display: flex;
    flex-direction: column;
    overflow: hidden;
    box-shadow: 0 16px 48px rgba(0, 0, 0, 0.5);
  }

  .terminal-modal-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 8px 12px;
    background: var(--bg-surface);
    border-bottom: 1px solid var(--border-subtle);
    flex-shrink: 0;
  }

  .terminal-modal-title {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--text-dim);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .tmt-prefix {
    color: var(--text-muted);
    font-weight: 400;
    letter-spacing: 0.04em;
    text-transform: uppercase;
  }
  .tmt-dash {
    color: var(--text-muted);
    margin: 0 var(--sp-2xs);
  }
  .tmt-repo {
    color: var(--text-primary);
    font-weight: 500;
  }
  .tmt-branch {
    color: var(--text-dim);
    font-weight: 400;
  }
  .tmt-label {
    color: var(--accent-green);
    font-weight: 500;
  }
  .tmt-sep {
    color: var(--text-muted);
    margin: 0 var(--sp-xs);
  }

  .terminal-modal-close {
    background: none;
    border: none;
    color: var(--text-muted);
    font-size: var(--text-section);
    cursor: pointer;
    padding: 0 4px;
    line-height: 1;
  }

  .terminal-modal-close:hover { color: var(--text-primary); }

  .terminal-modal-body {
    flex: 1;
    overflow: hidden;
  }
</style>
