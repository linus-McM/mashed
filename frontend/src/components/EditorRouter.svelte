<script lang="ts">
  import { getEditorType } from './editorUtils';
  import MonacoEditor from './MonacoEditor.svelte';

  export let filePath = '';
  export let repoPath = '';
  export let mode = 'source';
  export let editable = false;

  $: editorType = getEditorType(filePath);
</script>

{#if editorType === 'markdown'}
  <div class="placeholder">
    <p>Markdown Editor</p>
    <p class="file-path">{filePath}</p>
    <p class="coming-soon">Coming soon</p>
  </div>
{:else if editorType === 'image'}
  <div class="placeholder">
    <p>Image Viewer</p>
    <p class="file-path">{filePath}</p>
    <p class="coming-soon">Coming soon</p>
  </div>
{:else}
  <MonacoEditor {filePath} {repoPath} {mode} {editable} />
{/if}

<style>
  .placeholder {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    background: var(--bg-deepest);
    gap: 0.5rem;
  }

  .placeholder .file-path {
    font-family: monospace;
    font-size: 0.85rem;
  }

  .placeholder .coming-soon {
    color: var(--text-secondary);
    font-size: 0.85rem;
  }
</style>
