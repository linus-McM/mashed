<script>
  import { createEventDispatcher } from 'svelte';
  import { AlertTriangle } from 'lucide-svelte';
  import { GitForcePush } from '../../wailsjs/go/main/App.js';

  const dispatch = createEventDispatcher();

  export let repoPath = '';
  export let conflictMessage = '';

  let pushing = false;
  let error = '';

  async function forcePush() {
    pushing = true;
    error = '';
    try {
      await GitForcePush(repoPath);
      dispatch('pushed', { force: true });
    } catch (e) {
      error = e?.message || 'Force push failed';
      pushing = false;
    }
  }

  function cancel() {
    dispatch('cancel');
  }

  function handleKeydown(e) {
    if (e.key === 'Escape') cancel();
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<div class="overlay" role="presentation" on:click={cancel} on:keydown={() => {}}>
  <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
  <div class="modal" role="dialog" on:click|stopPropagation on:keydown={() => {}}>
    <h2><AlertTriangle size={18} /> Push Conflict</h2>
    <p class="subtitle">The remote branch has diverged from your local branch.</p>

    <div class="conflict-detail">
      <pre>{conflictMessage}</pre>
    </div>

    <div class="warning">
      <AlertTriangle size={14} />
      <span>Force push will overwrite the remote branch. This uses <code>--force-with-lease</code> for safety.</span>
    </div>

    {#if error}
      <div class="error">{error}</div>
    {/if}

    <div class="actions">
      <button class="btn-cancel" on:click={cancel}>Cancel</button>
      <button
        class="btn-force"
        on:click={forcePush}
        disabled={pushing}
      >
        {pushing ? 'Force Pushing...' : 'Force Push'}
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
    width: 460px;
    max-height: 80vh;
    overflow-y: auto;
  }

  h2 {
    font-size: 18px;
    font-weight: 600;
    color: var(--accent-red);
    margin: 0 0 4px;
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .subtitle {
    font-size: 12px;
    color: var(--text-dim);
    margin: 0 0 20px;
  }

  .conflict-detail {
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    padding: 12px;
    margin-bottom: 16px;
    max-height: 150px;
    overflow-y: auto;
  }

  .conflict-detail pre {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--text-dim);
    margin: 0;
    white-space: pre-wrap;
    word-break: break-word;
  }

  .warning {
    display: flex;
    align-items: flex-start;
    gap: 8px;
    padding: 8px 12px;
    background: color-mix(in srgb, var(--accent-amber) 8%, transparent);
    border: 1px solid color-mix(in srgb, var(--accent-amber) 20%, transparent);
    border-radius: var(--radius-md);
    color: var(--accent-amber);
    font-size: 12px;
    margin-bottom: 16px;
  }

  .warning code {
    font-family: var(--font-mono);
    font-size: 11px;
    background: color-mix(in srgb, var(--accent-amber) 12%, transparent);
    padding: 1px 4px;
    border-radius: 3px;
  }

  .error {
    font-size: 12px;
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

  .btn-force {
    padding: 8px 20px;
    background: var(--accent-red);
    border: none;
    border-radius: var(--radius-md);
    color: #fff;
    font-family: var(--font-ui);
    font-size: 13px;
    font-weight: 600;
    cursor: pointer;
  }

  .btn-force:hover { opacity: 0.9; }
  .btn-force:disabled { opacity: 0.4; cursor: not-allowed; }
</style>
