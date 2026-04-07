<script>
  import { SvelteFlow, Controls, MiniMap, Background, useSvelteFlow } from '@xyflow/svelte';

  export let nodes;
  export let edges;
  export let isValidConnection;
  export let onConnect;
  export let onDropProcess;
  export let nodeTypes = {};
  export let onNodeClick = null;

  const { screenToFlowPosition } = useSvelteFlow();

  function onDrop(e) {
    e.preventDefault();
    const processId = e.dataTransfer.getData('application/bmad-process');
    if (!processId) return;
    const position = screenToFlowPosition({ x: e.clientX, y: e.clientY });
    onDropProcess(processId, position);
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
</style>
