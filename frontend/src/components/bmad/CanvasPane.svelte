<script>
  import { SvelteFlow, Controls, MiniMap, Background, useSvelteFlow } from '@xyflow/svelte';

  export let nodes;
  export let edges;
  export let isValidConnection;
  export let onConnect;
  export let onDropProcess;
  export let onDropStory = null;
  export let nodeTypes = {};
  export let onNodeClick = null;
  export let onNodesDelete = null;
  export let onEdgesDelete = null;
  export let onSelectionChange = null;
  export let onReconnect = null;
  export let executionStatus = 'idle';

  const { screenToFlowPosition } = useSvelteFlow();

  $: deleteKeys = executionStatus === 'idle' ? ['Delete', 'Backspace'] : [];

  function onDrop(e) {
    e.preventDefault();
    const processId = e.dataTransfer.getData('application/bmad-process');
    if (processId) {
      const position = screenToFlowPosition({ x: e.clientX, y: e.clientY });
      onDropProcess(processId, position);
      return;
    }
    const storyJson = e.dataTransfer.getData('application/bmad-story');
    if (storyJson && onDropStory) {
      try {
        const storyData = JSON.parse(storyJson);
        const position = screenToFlowPosition({ x: e.clientX, y: e.clientY });
        onDropStory(storyData, position);
      } catch { /* invalid JSON — ignore */ }
    }
  }

  function onDragOver(e) {
    e.preventDefault();
    e.dataTransfer.dropEffect = 'move';
  }

  function handleConnect(e) {
    onConnect(e.detail);
  }

  function handleNodeClick(e) {
    if (onNodeClick) onNodeClick(e.detail);
  }

  function handleNodesDelete(e) {
    if (onNodesDelete) onNodesDelete(e.detail);
  }

  function handleEdgesDelete(e) {
    if (onEdgesDelete) onEdgesDelete(e.detail);
  }

  function handleSelectionChange(e) {
    if (onSelectionChange) onSelectionChange(e.detail);
  }

  function handleReconnect(e) {
    if (onReconnect) onReconnect(e.detail);
  }
</script>

<div class="flow-wrap" on:drop={onDrop} on:dragover={onDragOver}>
  <slot name="empty-hint" />
  <SvelteFlow
    {nodes}
    {edges}
    {nodeTypes}
    {isValidConnection}
    on:connect={handleConnect}
    on:nodeclick={handleNodeClick}
    on:nodesdelete={handleNodesDelete}
    on:edgesdelete={handleEdgesDelete}
    on:selectionchange={handleSelectionChange}
    on:reconnect={handleReconnect}
    deleteKeyCode={deleteKeys}
    selectionKeyCode="Shift"
    multiSelectionKeyCode="Meta"
    edgesReconnectable
    fitView
  >
    <Controls position="bottom-right" />
    <MiniMap
      pannable
      zoomable
      style="background: var(--bg-surface); border: 1px solid var(--border-subtle);"
    />
    <Background gap={16} size={1} />
  </SvelteFlow>
</div>

<style>
  .flow-wrap {
    flex: 1;
    position: relative;
    width: 100%;
    height: 100%;
  }

  /* SvelteFlow theme overrides */
  .flow-wrap :global(.svelte-flow) {
    background: var(--bg-deepest);
  }

  .flow-wrap :global(.svelte-flow__node) {
    background: transparent;
    border: none;
    padding: 0;
  }

  .flow-wrap :global(.svelte-flow__edge-path) {
    stroke: var(--border-emphasis, #484f58);
    stroke-width: 2;
  }

  .flow-wrap :global(.svelte-flow__handle) {
    width: 8px;
    height: 8px;
    background: var(--accent-green);
    border: 2px solid var(--bg-elevated);
  }

  .flow-wrap :global(.svelte-flow__controls) {
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md, 6px);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
  }

  .flow-wrap :global(.svelte-flow__controls button) {
    background: var(--bg-surface);
    color: var(--text-dim);
    border-bottom: 1px solid var(--border-subtle);
  }

  .flow-wrap :global(.svelte-flow__controls button:hover) {
    background: var(--bg-elevated);
    color: var(--text-primary);
  }

  .flow-wrap :global(.svelte-flow__controls button svg) {
    fill: currentColor;
  }

  .flow-wrap :global(.svelte-flow__minimap) {
    border-radius: var(--radius-md, 6px);
    overflow: hidden;
  }

  .flow-wrap :global(.svelte-flow__background pattern line) {
    stroke: var(--border-subtle);
  }

  /* Selected edge styling */
  .flow-wrap :global(.svelte-flow__edge.selected .svelte-flow__edge-path) {
    stroke: var(--accent-green);
    stroke-width: 3;
  }

  .flow-wrap :global(.svelte-flow__edge-path:hover) {
    stroke: var(--accent-green);
    cursor: pointer;
  }

  /* Edge reconnection handle */
  .flow-wrap :global(.svelte-flow__edgeupdater) {
    cursor: grab;
  }

  /* Multi-select rectangle */
  .flow-wrap :global(.svelte-flow__selection) {
    background: rgba(0, 229, 122, 0.08);
    border: 1px dashed var(--accent-green);
  }
</style>
