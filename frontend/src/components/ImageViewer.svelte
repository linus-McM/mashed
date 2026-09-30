<script lang="ts">
  import { onDestroy } from 'svelte';
  import { buildImagePath } from './imageViewerUtils';
  import { ReadFileBase64 } from '../../wailsjs/go/main/App.js';
  import { errorMessage } from '../lib/errorMessage';

  /** Discriminated view-model for the image viewer's three render states. */
  type ViewerState =
    | { status: 'loading' }
    | { status: 'loaded'; dataUri: string }
    | { status: 'error'; message: string };

  /** Minimal structural shape of a panzoom instance — we only touch `dispose()`. */
  interface PanzoomInstance {
    dispose(): void;
  }

  export let filePath = '';
  export let repoPath = '';

  let state: ViewerState = { status: 'loading' };
  let imgElement: HTMLImageElement | undefined;
  let panzoomInstance: PanzoomInstance | null = null;
  let loadGeneration = 0;

  function disposePanzoom(): void {
    if (panzoomInstance) {
      panzoomInstance.dispose();
      panzoomInstance = null;
    }
  }

  async function loadImage(fp: string, rp: string): Promise<void> {
    if (!fp) return;
    const gen = ++loadGeneration;
    state = { status: 'loading' };
    disposePanzoom();
    try {
      const fullPath = buildImagePath(rp, fp);
      const dataUri = await ReadFileBase64(fullPath);
      if (gen !== loadGeneration) return;
      state = { status: 'loaded', dataUri };
    } catch (err) {
      if (gen !== loadGeneration) return;
      state = { status: 'error', message: `Failed to load ${fp}: ${errorMessage(err)}` };
    }
  }

  $: loadImage(filePath, repoPath);

  async function attachPanzoom(): Promise<void> {
    if (!imgElement || panzoomInstance) return;
    const gen = loadGeneration;
    const panzoom = (await import('panzoom')).default;
    if (gen !== loadGeneration || !imgElement) return;
    panzoomInstance = panzoom(imgElement, {
      smoothScroll: false,
      zoomDoubleClickSpeed: 1,
    });
  }

  onDestroy(disposePanzoom);
</script>

{#if state.status === 'loading'}
  <div class="viewer-container loading">
    <div class="spinner">Loading...</div>
  </div>
{:else if state.status === 'error'}
  <div class="viewer-container error">
    <p class="error-message">{state.message}</p>
  </div>
{:else}
  <div class="viewer-container">
    <img
      bind:this={imgElement}
      src={state.dataUri}
      alt={filePath}
      on:load={attachPanzoom}
    />
  </div>
{/if}

<style>
  .viewer-container {
    height: 100%;
    background: var(--bg-deepest);
    display: flex;
    align-items: center;
    justify-content: center;
    overflow: hidden;
  }

  .error-message {
    color: var(--text-secondary);
  }

  img {
    max-width: 100%;
    max-height: 100%;
    object-fit: contain;
  }
</style>
