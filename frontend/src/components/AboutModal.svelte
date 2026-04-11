<script>
  import { createEventDispatcher } from 'svelte';
  import { Hexagon } from 'lucide-svelte';

  const dispatch = createEventDispatcher();

  function close() {
    dispatch('close');
  }

  function handleKeydown(e) {
    if (e.key === 'Escape') close();
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<div class="overlay" role="presentation" on:click={close} on:keydown={() => {}}>
  <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
  <div class="modal" role="dialog" on:click|stopPropagation on:keydown={() => {}}>
    <div class="icon"><Hexagon size={32} /></div>
    <h2>mashed</h2>
    <p class="tagline">Notification-first IDE for multi-agent development</p>

    <div class="info">
      <div class="info-row">
        <span class="info-label">Version</span>
        <span class="info-value">0.1.0</span>
      </div>
      <div class="info-row">
        <span class="info-label">Runtime</span>
        <span class="info-value">Wails v2</span>
      </div>
      <div class="info-row">
        <span class="info-label">Frontend</span>
        <span class="info-value">Svelte</span>
      </div>
    </div>

    <div class="actions">
      <button class="btn-close" on:click={close}>Close</button>
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
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-lg);
    padding: 32px;
    width: 320px;
    text-align: center;
  }

  .icon {
    color: var(--accent-green);
    margin-bottom: 12px;
  }

  h2 {
    font-size: 20px;
    font-weight: 700;
    color: var(--text-primary);
    margin: 0 0 4px;
    letter-spacing: -0.5px;
  }

  .tagline {
    font-size: var(--text-body);
    color: var(--text-dim);
    margin: 0 0 24px;
    line-height: 1.4;
  }

  .info {
    background: var(--bg-deepest);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    padding: 12px 16px;
    margin-bottom: 24px;
  }

  .info-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 4px 0;
  }

  .info-row + .info-row {
    border-top: 1px solid var(--border-subtle);
    margin-top: 4px;
    padding-top: 8px;
  }

  .info-label {
    font-size: 11px;
    font-weight: 600;
    color: var(--text-dim);
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .info-value {
    font-size: var(--text-body);
    color: var(--text-primary);
    font-family: var(--font-mono);
  }

  .actions {
    display: flex;
    justify-content: center;
  }

  .btn-close {
    padding: 8px 24px;
    background: none;
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    color: var(--text-dim);
    font-family: var(--font-mono);
    font-size: 13px;
    cursor: pointer;
    transition: all 100ms ease;
  }

  .btn-close:hover {
    border-color: var(--accent-green);
    color: var(--text-primary);
  }
</style>
