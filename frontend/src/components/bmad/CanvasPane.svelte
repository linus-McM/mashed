<script>
  import { tick } from 'svelte';
  import { SvelteFlow, Controls, MiniMap, Background, useSvelteFlow } from '@xyflow/svelte';
  import { Trash2, LayoutTemplate, ChevronRight, Maximize2 } from 'lucide-svelte';
  import DeletableEdge from './DeletableEdge.svelte';
  import { handleMashedAssetDrop } from './canvasPaneDropHandler';
  import { MASHED_ASSET_MIME } from './dragMimeTypes.js';

  /** @typedef {import('../../types/workflow').CanvasNode} CanvasNode */
  /** @typedef {import('../../types/workflow').CanvasEdge} CanvasEdge */
  /** @typedef {import('../../types/workflow').Position} Position */
  /** @typedef {import('../../lib/types/wails').Workflow} Workflow */
  /**
   * @typedef {Object} ContextMenuState
   * @property {number} x
   * @property {number} y
   * @property {string | null} [nodeId]
   * @property {string | null} [nodeStatus]
   * @property {Position} flowPosition
   */
  /**
   * @typedef {CustomEvent<{ node?: CanvasNode; event: MouseEvent | TouchEvent }>} XyflowNodeCtxEvent
   * @typedef {CustomEvent<{ event: MouseEvent | TouchEvent }>} XyflowPaneCtxEvent
   * @typedef {CustomEvent<{ source: string; target: string; sourceHandle?: string | null; targetHandle?: string | null }>} XyflowConnectEvent
   * @typedef {CustomEvent<{ oldEdge: CanvasEdge; newConnection: { source: string; target: string; sourceHandle?: string | null; targetHandle?: string | null } }>} XyflowReconnectEvent
   * @typedef {CustomEvent<{ node?: CanvasNode }>} XyflowNodeClickEvent
   * @typedef {CustomEvent<CanvasNode[]>} XyflowNodesDeleteEvent
   * @typedef {CustomEvent<CanvasEdge[]>} XyflowEdgesDeleteEvent
   * @typedef {CustomEvent<{ nodes: CanvasNode[]; edges: CanvasEdge[] }>} XyflowSelectionChangeEvent
   */

  /** @type {import('svelte/store').Writable<CanvasNode[]>} */
  export let nodes;
  /** @type {import('svelte/store').Writable<CanvasEdge[]>} */
  export let edges;
  /** @type {(connection: { source: string; target: string; sourceHandle?: string | null; targetHandle?: string | null }) => boolean} */
  export let isValidConnection;
  /** @type {(params: { source: string; target: string; sourceHandle?: string | null; targetHandle?: string | null }) => void} */
  export let onConnect;
  /** @type {(processId: string, position: Position) => void} */
  export let onDropProcess;
  /** @type {((storyData: { storyId: string; status?: string }, position: Position) => void) | null} */
  export let onDropStory = null;
  /** @type {((nodeType: 'condition' | 'loop' | 'loopUntil' | 'transform' | 'merge', position: Position) => void) | null} */
  export let onDropControlFlow = null;
  /** @type {Record<string, import('svelte').ComponentType>} */
  export let nodeTypes = {};
  /** @type {((detail: { node?: CanvasNode }) => void) | null} */
  export let onNodeClick = null;
  /** @type {(() => void) | null} */
  export let onPaneClick = null;
  /** @type {((deletedNodes: CanvasNode[]) => void) | null} */
  export let onNodesDelete = null;
  /** @type {((deletedEdges: CanvasEdge[]) => void) | null} */
  export let onEdgesDelete = null;
  /** @type {((selection: { nodes: CanvasNode[]; edges: CanvasEdge[] }) => void) | null} */
  export let onSelectionChange = null;
  /** @type {((detail: { oldEdge: CanvasEdge; newConnection: { source: string; target: string; sourceHandle?: string | null; targetHandle?: string | null } }) => void) | null} */
  export let onReconnect = null;
  /** @type {((templateId: string, position: Position, connectToNodeId?: string) => void) | null} */
  export let onAddTemplate = null;
  /** @type {Workflow[]} */
  export let templates = [];
  export let executionStatus = 'idle';
  export let configPanelOpen = false;
  export let configPanelWidth = 280;

  const { screenToFlowPosition, deleteElements, fitView } = useSvelteFlow();

  /** @type {import('@xyflow/svelte').EdgeTypes} */
  const edgeTypes = /** @type {import('@xyflow/svelte').EdgeTypes} */ (/** @type {unknown} */ ({ default: DeletableEdge }));

  /** @type {import('svelte/store').Writable<import('@xyflow/svelte').Node[]>} */
  const nodesAny = /** @type {import('svelte/store').Writable<import('@xyflow/svelte').Node[]>} */ (/** @type {unknown} */ (nodes));
  /** @type {import('svelte/store').Writable<import('@xyflow/svelte').Edge[]>} */
  const edgesAny = /** @type {import('svelte/store').Writable<import('@xyflow/svelte').Edge[]>} */ (/** @type {unknown} */ (edges));

  $: deleteKeys = executionStatus === 'idle' ? ['Delete', 'Backspace'] : [];

  // Context menu state
  /** @type {ContextMenuState | null} */
  let contextMenu = null;
  let showTemplateSub = false;

  // uiqa-08: keyboard nav state for context menu
  // itemRefs holds bound <button> nodes in render order; activeIdx is the
  // currently-focused index; previousFocus restores focus on close.
  /** @type {HTMLButtonElement[]} */
  let itemRefs = [];
  let activeIdx = 0;
  /** @type {HTMLElement | null} */
  let previousFocus = null;

  // Top-level menu items in render order. Rebuilt reactively so ArrowDown/Up
  // wrap lengths match the DOM. The "Add Template" entry is always present;
  // "Delete Node" only when the menu opened on a node.
  $: menuItemCount = contextMenu ? (contextMenu.nodeId ? 2 : 1) : 0;

  async function openContextMenuFocus() {
    // Capture the previously focused element (may be body/canvas) so we can
    // restore it on close. Reset index, clear stale refs, then focus item 0
    // after the menu has rendered.
    previousFocus = (typeof document !== 'undefined' ? /** @type {HTMLElement | null} */ (document.activeElement) : null);
    activeIdx = 0;
    itemRefs = [];
    await tick();
    itemRefs[0]?.focus();
  }

  /** @param {MouseEvent | TouchEvent} ev */
  function pointerOf(ev) {
    if ('touches' in ev && ev.touches.length > 0) {
      return { x: ev.touches[0].clientX, y: ev.touches[0].clientY };
    }
    const me = /** @type {MouseEvent} */ (ev);
    return { x: me.clientX, y: me.clientY };
  }

  /** @param {XyflowNodeCtxEvent} e */
  function handleNodeContextMenu(e) {
    const node = e.detail.node;
    if (!node) return;
    e.detail.event.preventDefault();
    showTemplateSub = false;
    const { x, y } = pointerOf(e.detail.event);
    contextMenu = {
      x,
      y,
      nodeId: node.id,
      nodeStatus: node.data?.status || 'pending',
      flowPosition: screenToFlowPosition({ x, y }),
    };
    openContextMenuFocus();
  }

  /** @param {XyflowPaneCtxEvent} e */
  function handlePaneContextMenu(e) {
    e.detail.event.preventDefault();
    showTemplateSub = false;
    const { x, y } = pointerOf(e.detail.event);
    contextMenu = {
      x,
      y,
      nodeId: null,
      nodeStatus: null,
      flowPosition: screenToFlowPosition({ x, y }),
    };
    openContextMenuFocus();
  }

  function closeContextMenu() {
    if (!contextMenu) return;
    contextMenu = null;
    showTemplateSub = false;
    itemRefs = [];
    activeIdx = 0;
    // Restore focus to the element that opened the menu.
    const prev = previousFocus;
    previousFocus = null;
    if (prev && typeof prev.focus === 'function') {
      try { prev.focus(); } catch { /* ignore */ }
    }
  }

  /** @param {KeyboardEvent} e */
  function handleMenuKeydown(e) {
    if (!contextMenu) return;
    const count = menuItemCount;
    if (count === 0) return;
    switch (e.key) {
      case 'Escape':
        e.preventDefault();
        closeContextMenu();
        break;
      case 'ArrowDown':
        e.preventDefault();
        activeIdx = (activeIdx + 1) % count;
        itemRefs[activeIdx]?.focus();
        break;
      case 'ArrowUp':
        e.preventDefault();
        activeIdx = (activeIdx - 1 + count) % count;
        itemRefs[activeIdx]?.focus();
        break;
      case 'Home':
        e.preventDefault();
        activeIdx = 0;
        itemRefs[0]?.focus();
        break;
      case 'End':
        e.preventDefault();
        activeIdx = count - 1;
        itemRefs[activeIdx]?.focus();
        break;
      case 'Enter':
      case ' ':
        e.preventDefault();
        itemRefs[activeIdx]?.click();
        break;
      case 'Tab':
        // Tab closes the menu (standard desktop menu behavior).
        closeContextMenu();
        break;
    }
  }

  function contextDeleteNode() {
    if (!contextMenu || !contextMenu.nodeId) return;
    if (contextMenu.nodeStatus === 'running') {
      closeContextMenu();
      return;
    }
    const nodeId = contextMenu.nodeId;
    deleteElements({ nodes: [{ id: nodeId }] });
    if (onNodesDelete) {
      const deleted = $nodes.filter((n) => n.id === nodeId);
      onNodesDelete(deleted);
    }
    closeContextMenu();
  }

  /** @param {string} templateId */
  function contextAddTemplate(templateId) {
    if (!contextMenu || !onAddTemplate) return;
    onAddTemplate(templateId, contextMenu.flowPosition, contextMenu.nodeId ?? undefined);
    closeContextMenu();
  }

  /** @param {DragEvent} e */
  function onDrop(e) {
    e.preventDefault();
    if (!e.dataTransfer) return;
    const processId = e.dataTransfer.getData('application/bmad-process');
    if (processId) {
      const position = screenToFlowPosition({ x: e.clientX, y: e.clientY });
      onDropProcess(processId, position);
      return;
    }
    const templateId = e.dataTransfer.getData('application/bmad-template');
    if (templateId && onAddTemplate) {
      const position = screenToFlowPosition({ x: e.clientX, y: e.clientY });
      onAddTemplate(templateId, position, undefined);
      return;
    }
    const controlFlowType = e.dataTransfer.getData('application/bmad-controlflow');
    if (controlFlowType && onDropControlFlow) {
      const position = screenToFlowPosition({ x: e.clientX, y: e.clientY });
      onDropControlFlow(/** @type {'condition' | 'loop' | 'loopUntil' | 'transform' | 'merge'} */ (controlFlowType), position);
      return;
    }
    const storyJson = e.dataTransfer.getData('application/bmad-story');
    if (storyJson && onDropStory) {
      try {
        const storyData = JSON.parse(storyJson);
        const position = screenToFlowPosition({ x: e.clientX, y: e.clientY });
        onDropStory(storyData, position);
      } catch { /* invalid JSON — ignore */ }
      return;
    }
    // Mashed-ready commands. Delegated to the pure handler so the
    // unit-test contract and the integration cannot drift.
    const mashedAssetRaw = e.dataTransfer.getData(MASHED_ASSET_MIME);
    if (mashedAssetRaw) {
      const position = screenToFlowPosition({ x: e.clientX, y: e.clientY });
      $nodes = handleMashedAssetDrop({
        raw: mashedAssetRaw,
        position,
        currentNodes: $nodes,
      });
    }
  }

  /** @param {DragEvent} e */
  function onDragOver(e) {
    e.preventDefault();
    if (e.dataTransfer) e.dataTransfer.dropEffect = 'move';
  }

  /** @param {XyflowConnectEvent} e */
  function handleConnect(e) {
    onConnect(e.detail);
  }

  /** @param {XyflowNodeClickEvent} e */
  function handleNodeClick(e) {
    if (onNodeClick) onNodeClick(e.detail);
  }

  function handlePaneClick() {
    if (onPaneClick) onPaneClick();
  }

  /** @param {XyflowNodesDeleteEvent} e */
  function handleNodesDelete(e) {
    if (onNodesDelete) onNodesDelete(e.detail);
  }

  /** @param {XyflowEdgesDeleteEvent} e */
  function handleEdgesDelete(e) {
    if (onEdgesDelete) onEdgesDelete(e.detail);
  }

  /** @param {XyflowSelectionChangeEvent} e */
  function handleSelectionChange(e) {
    if (onSelectionChange) onSelectionChange(e.detail);
  }

  /** @param {XyflowReconnectEvent} e */
  function handleReconnect(e) {
    if (onReconnect) onReconnect(e.detail);
  }
</script>

<svelte:window on:click={closeContextMenu} on:keydown={handleMenuKeydown} />

<!-- svelte-ignore a11y-no-static-element-interactions -->
<div class="flow-wrap" on:drop={onDrop} on:dragover={onDragOver} role="application">
  <slot name="empty-hint" />
  <SvelteFlow
    nodes={nodesAny}
    edges={edgesAny}
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
    deleteKey={deleteKeys}
    selectionKey="Shift"
    multiSelectionKey="Meta"
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
    <!-- svelte-ignore a11y-no-static-element-interactions -->
    <div
      class="context-menu"
      style="left: {contextMenu.x}px; top: {contextMenu.y}px;"
      on:click|stopPropagation
      on:keydown|stopPropagation={handleMenuKeydown}
      role="menu"
    >
      {#if contextMenu.nodeId}
        <button
          class="context-item"
          class:disabled={contextMenu.nodeStatus === 'running'}
          on:click={contextDeleteNode}
          bind:this={itemRefs[0]}
          role="menuitem"
          tabindex="-1"
        >
          <Trash2 size={13} />
          <span>Delete Node</span>
        </button>
        <div class="context-divider" />
      {/if}

      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div class="context-submenu-wrap"
        on:mouseenter={() => showTemplateSub = true}
        on:mouseleave={() => showTemplateSub = false}
      >
        <button
          class="context-item"
          on:click={() => showTemplateSub = !showTemplateSub}
          bind:this={itemRefs[contextMenu?.nodeId ? 1 : 0]}
          role="menuitem"
          tabindex="-1"
        >
          <LayoutTemplate size={13} />
          <span>Add Template</span>
          <ChevronRight size={12} class="submenu-arrow" />
        </button>

        {#if showTemplateSub && templates.length > 0}
          <div class="context-submenu" role="menu">
            {#each templates as tpl}
              <button
                class="context-item"
                on:click={() => contextAddTemplate(tpl.id)}
                role="menuitem"
                tabindex="-1"
              >
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
    background: color-mix(in srgb, var(--accent-green) 8%, transparent);
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
    padding: var(--sp-xs) 10px;
    background: none;
    border: none;
    border-radius: var(--radius-sm);
    color: var(--text-primary);
    font-family: var(--font-mono);
    font-size: var(--text-body);
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
    margin: var(--sp-2xs) var(--sp-xs);
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
    font-size: var(--text-label);
    color: var(--text-muted);
  }
</style>
