<script lang="ts">
  import { getEditorType } from './editorUtils';
  import MonacoEditor from './MonacoEditor.svelte';
  import ImageViewer from './ImageViewer.svelte';
  import LazyView from './LazyView.svelte';

  // MarkdownEditor is code-split (R31) so it stays out of the entry chunk.
  // Module-level loader keeps its identity stable for LazyView's cache.
  const loadMarkdownEditor = () => import('./MarkdownEditor.svelte');

  type EditorMode = 'source' | 'diff';

  export let filePath = '';
  export let repoPath = '';
  export let mode: EditorMode = 'source';
  export let editable = false;

  $: editorType = getEditorType(filePath);
</script>

{#if editorType === 'markdown'}
  <LazyView loader={loadMarkdownEditor} {filePath} {repoPath} {editable} />
{:else if editorType === 'image'}
  <ImageViewer {filePath} {repoPath} />
{:else}
  <MonacoEditor {filePath} {repoPath} {mode} {editable} />
{/if}

