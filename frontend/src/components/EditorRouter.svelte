<script lang="ts">
  import { getEditorType } from './editorUtils';
  import MonacoEditor from './MonacoEditor.svelte';
  import ImageViewer from './ImageViewer.svelte';
  import MarkdownEditor from './MarkdownEditor.svelte';

  type EditorMode = 'source' | 'diff';

  export let filePath = '';
  export let repoPath = '';
  export let mode: EditorMode = 'source';
  export let editable = false;

  $: editorType = getEditorType(filePath);
</script>

{#if editorType === 'markdown'}
  <MarkdownEditor {filePath} {repoPath} {editable} />
{:else if editorType === 'image'}
  <ImageViewer {filePath} {repoPath} />
{:else}
  <MonacoEditor {filePath} {repoPath} {mode} {editable} />
{/if}

