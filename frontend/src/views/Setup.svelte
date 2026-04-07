<script>
  import { createEventDispatcher } from 'svelte';
  import { Hexagon } from 'lucide-svelte';
  import { PickDirectory, SetDevDir } from '../../wailsjs/go/main/App.js';

  const dispatch = createEventDispatcher();

  let selectedDir = '';
  let error = '';
  let loading = false;

  async function browse() {
    error = '';
    try {
      const dir = await PickDirectory();
      if (dir) {
        selectedDir = dir;
      }
    } catch (e) {
      error = 'Failed to open directory picker';
    }
  }

  async function confirm() {
    if (!selectedDir) return;
    loading = true;
    error = '';
    try {
      await SetDevDir(selectedDir);
      dispatch('ready');
    } catch (e) {
      error = e?.message || 'Failed to set directory';
      loading = false;
    }
  }

  function handleKeydown(e) {
    if (e.key === 'Enter' && selectedDir) confirm();
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<div class="setup">
  <div class="setup-card">
    <div class="icon"><Hexagon size={48} /></div>
    <h1>Mashed</h1>
    <p class="subtitle">Choose the directory where your projects live.</p>
    <p class="hint">Mashed will scan for git repos and running Claude sessions in this directory.</p>

    <div class="dir-picker">
      <button class="browse-btn" on:click={browse}>
        Choose Directory
      </button>

      {#if selectedDir}
        <div class="selected-dir">
          <span class="dir-label">Selected:</span>
          <span class="dir-path">{selectedDir}</span>
        </div>
      {/if}
    </div>

    {#if error}
      <div class="error">{error}</div>
    {/if}

    {#if selectedDir}
      <button class="confirm-btn" on:click={confirm} disabled={loading}>
        {loading ? 'Starting...' : 'Start Mashed'}
      </button>
    {/if}

    <p class="footer-hint">
      You can change this later. Config saved to ~/.mashed/config.json
    </p>
  </div>
</div>

<style>
  .setup {
    width: 100vw;
    height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--bg-deepest);
    /* Allow dragging the frameless window */
    --wails-draggable: drag;
  }

  .setup-card {
    --wails-draggable: no-drag;
    max-width: 480px;
    width: 100%;
    padding: 48px 40px;
    text-align: center;
  }

  .icon {
    font-size: 48px;
    color: var(--accent-green);
    margin-bottom: 16px;
    line-height: 1;
  }

  h1 {
    font-family: var(--font-ui);
    font-size: 20px;
    font-weight: 600;
    color: var(--text-primary);
    margin: 0 0 8px;
  }

  .subtitle {
    font-size: 14px;
    color: var(--text-primary);
    margin: 0 0 4px;
  }

  .hint {
    font-size: 12px;
    color: var(--text-dim);
    margin: 0 0 32px;
  }

  .dir-picker {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 16px;
    margin-bottom: 24px;
  }

  .browse-btn {
    font-family: var(--font-ui);
    font-size: 13px;
    font-weight: 500;
    padding: 10px 24px;
    border: 1px solid var(--border-emphasis);
    border-radius: var(--radius-md);
    background: var(--bg-elevated);
    color: var(--text-primary);
    cursor: pointer;
    transition: all 100ms ease-out;
  }

  .browse-btn:hover {
    background: var(--bg-active);
    border-color: var(--accent-green);
    color: var(--accent-green);
  }

  .selected-dir {
    display: flex;
    flex-direction: column;
    gap: 4px;
    padding: 12px 16px;
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    width: 100%;
  }

  .dir-label {
    font-size: 11px;
    color: var(--text-dim);
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .dir-path {
    font-family: var(--font-mono);
    font-size: 13px;
    color: var(--accent-green);
    word-break: break-all;
  }

  .confirm-btn {
    font-family: var(--font-ui);
    font-size: 13px;
    font-weight: 600;
    padding: 10px 32px;
    border: none;
    border-radius: var(--radius-md);
    background: var(--accent-green);
    color: var(--bg-deepest);
    cursor: pointer;
    transition: opacity 100ms ease-out;
    margin-bottom: 24px;
  }

  .confirm-btn:hover { opacity: 0.9; }
  .confirm-btn:disabled { opacity: 0.5; cursor: not-allowed; }

  .error {
    font-size: 12px;
    color: var(--accent-red);
    margin-bottom: 16px;
  }

  .footer-hint {
    font-size: 11px;
    color: var(--text-muted);
    margin: 0;
  }
</style>
