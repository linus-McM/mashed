<script>
  import { onDestroy } from 'svelte';
  import { buildImagePath } from './imageViewerUtils';
  import { ReadFileBase64 } from '../../wailsjs/go/main/App.js';

  export let filePath = '';
  export let repoPath = '';

  let state = { status: 'loading' };
  let imgElement;
  let panzoomInstance = null;
  let loadGeneration = 0;

  function disposePanzoom() {
    if (panzoomInstance) {
      panzoomInstance.dispose();
      panzoomInstance = null;
    }
  }

  async function loadImage(fp, rp) {
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
      state = { status: 'error', message: `Failed to load ${fp}: ${err}` };
    }
  }

  $: loadImage(filePath, repoPath);

  async function attachPanzoom() {
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
