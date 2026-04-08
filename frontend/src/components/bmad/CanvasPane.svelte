<script>
  import { SvelteFlow, Controls, MiniMap, Background, useSvelteFlow } from '@xyflow/svelte';
  import { Trash2, LayoutTemplate, ChevronRight, Maximize2 } from 'lucide-svelte';
  import DeletableEdge from './DeletableEdge.svelte';

  export let nodes;
  export let edges;
  export let isValidConnection;
  export let onConnect;
  export let onDropProcess;
  export let onDropStory = null;
  export let nodeTypes = {};
  export let onNodeClick = null;
  export let onPaneClick = null;
  export let onNodesDelete = null;
  export let onEdgesDelete = null;
  export let onSelectionChange = null;
  export let onReconnect = null;
  export let onAddTemplate = null;
  export let templates = [];
  export let executionStatus = 'idle';
  export let configPanelOpen = false;
  export let configPanelWidth = 280;

  const { screenToFlowPosition, deleteElements, fitView } = useSvelteFlow();

  const edgeTypes = { default: DeletableEdge };

  $: deleteKeys = executionStatus === 'idle' ? ['Delete', 'Backspace'] : [];

  // Context menu state
  let contextMenu = null; // { x, y, nodeId?, nodeStatus?, flowPosition }
  let showTemplateSub = false;

  function handleNodeContextMenu(e) {
    const node = e.detail.node;
    if (!node) return;
    e.detail.event.preventDefault();
    showTemplateSub = false;
    contextMenu = {
      x: e.detail.event.clientX,
      y: e.detail.event.clientY,
      nodeId: node.id,
      nodeStatus: node.data?.status || 'pending',
      flowPosition: screenToFlowPosition({ x: e.detail.event.clientX, y: e.detail.event.clientY }),
    };
  }

  function handlePaneContextMenu(e) {
    e.detail.event.preventDefault();
    showTemplateSub = false;
    contextMenu = {
      x: e.detail.event.clientX,
      y: e.detail.event.clientY,
      nodeId: null,
      nodeStatus: null,
      flowPosition: screenToFlowPosition({ x: e.detail.event.clientX, y: e.detail.event.clientY }),
    };
  }

  function closeContextMenu() {
    contextMenu = null;
    showTemplateSub = false;
  }

  function contextDeleteNode() {
    if (!contextMenu || !contextMenu.nodeId) return;
    if (contextMenu.nodeStatus === 'running') {
      closeContextMenu();
      return;
    }
    deleteElements({ nodes: [{ id: contextMenu.nodeId }] });
    if (onNodesDelete) {
      const deleted = $nodes.filter(n => n.id === contextMenu.nodeId);
      onNodesDelete(deleted);
    }
    closeContextMenu();
  }

  function contextAddTemplate(templateId) {
    if (!contextMenu || !onAddTemplate) return;
    onAddTemplate(templateId, contextMenu.flowPosition, contextMenu.nodeId);
    closeContextMenu();
  }

  function onDrop(e) {
    e.preventDefault();
    const processId = e.dataTransfer.getData('application/bmad-process');
    if (processId) {
      const position = screenToFlowPosition({ x: e.clientX, y: e.clientY });
      onDropProcess(processId, position);
      return;
    }
    const templateId = e.dataTransfer.getData('application/bmad-template');
    if (templateId && onAddTemplate) {
      const position = screenToFlowPosition({ x: e.clientX, y: e.clientY });
      onAddTemplate(templateId, position, null);
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

  function handlePaneClick() {
    if (onPaneClick) onPaneClick();
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

<svelte:window on:click={closeContextMenu} />

<div class="flow-wrap" on:drop={onDrop} on:dragover={onDragOver}>
  <slot name="empty-hint" />
  <SvelteFlow
    {nodes}
    {edges}
    {nodeTypes}
    {edgeTypes}
    {isValidConnection}
    on:connect={handleConnect}
    on:nodeclick={handleNodeClick}
    on:paneclick={handlePaneClick}
    on:nodecontextmenu={handleNodeContextMenu}
    on:panecontextmenu={handlePaneContextMenu}
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
    <Controls position="bottom-right" style={configPanelOpen ? `right: ${configPanelWidth + 10}px;` : ''} />
    <MiniMap
      pannable
      zoomable
      style="background: var(--bg-surface); border: 1px solid var(--border-subtle); {configPanelOpen ? `right: ${configPanelWidth + 10}px;` : ''}"
    />
    <Background gap={16} size={1} />
  </SvelteFlow>

  <button
    class="fit-view-btn"
    style={configPanelOpen ? `right: ${configPanelWidth + 10}px;` : ''}
    on:click={() => fitView({ padding: 0.15, duration: 300 })}
    title="Fit all nodes in view"
  >
    <Maximize2 size={13} />
  </button>

  {#if contextMenu}
    <div
      class="context-menu"
      style="left: {contextMenu.x}px; top: {contextMenu.y}px;"
      on:click|stopPropagation
      on:keydown={() => {}}
    >
      {#if contextMenu.nodeId}
        <button
          class="context-item"
          class:disabled={contextMenu.nodeStatus === 'running'}
          on:click={contextDeleteNode}
        >
          <Trash2 size={13} />
          <span>Delete Node</span>
        </button>
        <div class="context-divider" />
      {/if}

      <div class="context-submenu-wrap"
        on:mouseenter={() => showTemplateSub = true}
        on:mouseleave={() => showTemplateSub = false}
      >
        <button class="context-item">
          <LayoutTemplate size={13} />
          <span>Add Template</span>
          <ChevronRight size={12} class="submenu-arrow" />
        </button>

        {#if showTemplateSub && templates.length > 0}
          <div class="context-submenu">
            {#each templates as tpl}
              <button class="context-item" on:click={() => contextAddTemplate(tpl.id)}>
                <span>{tpl.name}</span>
                <span class="tpl-meta">{tpl.nodes?.length || 0} nodes</span>
              </button>
            {/each}
          </div>
        {/if}
      </div>
    </div>
  {/if}
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

  /* Fit-view button above minimap */
  .fit-view-btn {
    position: absolute;
    bottom: 130px;
    right: 10px;
    width: 28px;
    height: 28px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md, 6px);
    color: var(--text-dim);
    cursor: pointer;
    z-index: 5;
    transition: color 100ms ease, border-color 100ms ease, background 100ms ease;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
  }

  .fit-view-btn:hover {
    color: var(--accent-green);
    border-color: var(--accent-green);
    background: var(--bg-elevated);
  }

  /* Right-click context menu */
  .context-menu {
    position: fixed;
    z-index: 50;
    min-width: 160px;
    padding: 4px;
    background: var(--bg-elevated);
    border: 1px solid var(--border-emphasis);
    border-radius: var(--radius-md, 6px);
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.5);
    font-family: var(--font-mono);
  }

  .context-item {
    display: flex;
    align-items: center;
    gap: 8px;
    width: 100%;
    padding: 6px 10px;
    background: none;
    border: none;
    border-radius: var(--radius-sm);
    color: var(--text-primary);
    font-family: var(--font-mono);
    font-size: 12px;
    cursor: pointer;
    transition: background 80ms ease, color 80ms ease;
    text-align: left;
  }

  .context-item:hover {
    background: var(--bg-active);
    color: var(--accent-red, #f85149);
  }

  .context-item.disabled {
    opacity: 0.35;
    cursor: not-allowed;
  }

  .context-item.disabled:hover {
    background: none;
    color: var(--text-primary);
  }

  .context-divider {
    height: 1px;
    background: var(--border-subtle);
    margin: 2px 6px;
  }

  .context-submenu-wrap {
    position: relative;
  }

  .context-item :global(.submenu-arrow) {
    margin-left: auto;
    color: var(--text-muted);
  }

  .context-submenu {
    position: absolute;
    left: 100%;
    top: -4px;
    min-width: 200px;
    padding: 4px;
    background: var(--bg-elevated);
    border: 1px solid var(--border-emphasis);
    border-radius: var(--radius-md, 6px);
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.5);
  }

  .context-submenu .context-item:hover {
    color: var(--accent-green);
  }

  .tpl-meta {
    margin-left: auto;
    font-size: 10px;
    color: var(--text-muted);
  }
</style>
