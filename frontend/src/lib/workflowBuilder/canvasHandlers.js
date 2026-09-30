// Canvas interaction handlers extracted verbatim from WorkflowBuilder.svelte
// (spec R31 — pure move, no behaviour change). The handlers mutate
// component state, so they are produced by a factory that closes over a
// context object of stores + getters/setters built by the component.
// `$nodes = x` in the component becomes `ctx.nodes.set(x)` here and a
// `$nodes` read becomes `get(ctx.nodes)`; both are synchronous, so the
// ordering of every store update is preserved.

import { get } from 'svelte/store';
import { GetNodeOutput } from '../../../wailsjs/go/main/App.js';
import { openModalForNode } from '../../stores/interactiveInput';

/** @typedef {import('../../types/workflow').CanvasNode} CanvasNode */
/** @typedef {import('../../types/workflow').CanvasEdge} CanvasEdge */
/** @typedef {import('../../types/workflow').Position} Position */
/** @typedef {import('../types/wails').Workflow} Workflow */
/** @typedef {import('../types/wails').ProcessDef} ProcessDef */
/** @typedef {{ source: string; target: string; sourceHandle?: string | null; targetHandle?: string | null }} Connection */

/**
 * @typedef {object} CanvasHandlersContext
 * @property {import('svelte/store').Writable<CanvasNode[]>} nodes
 * @property {import('svelte/store').Writable<CanvasEdge[]>} edges
 * @property {ProcessDef[]} processes
 * @property {Workflow[]} templates
 * @property {CanvasNode | null} selectedNode
 * @property {() => void} updateProgress
 * @property {string | null} executionId
 * @property {string} arrayModalNodeId
 * @property {string[]} arrayModalItems
 * @property {boolean} showArrayModal
 * @property {string} terminalTarget
 * @property {boolean} showTerminalModal
 * @property {string} outputModalLabel
 * @property {string} outputModalContent
 * @property {boolean} outputLoading
 * @property {boolean} showOutputModal
 */

/** @param {string | null | undefined} sourceHandle */
export function inferEdgeLabel(sourceHandle) {
  if (sourceHandle === 'true') return 'true';
  if (sourceHandle === 'false') return 'false';
  if (sourceHandle === 'loop-body') return 'body';
  if (sourceHandle === 'loop-exit') return 'exit';
  return '';
}

/** @type {Record<'condition' | 'loop' | 'loopUntil' | 'transform' | 'merge', string>} */
const controlFlowNames = { condition: 'Condition', loop: 'Loop', loopUntil: 'Loop Until', transform: 'Transform', merge: 'Merge' };

/** @param {CanvasHandlersContext} ctx */
export function createCanvasHandlers(ctx) {
  const { nodes, edges } = ctx;

  /** @param {Connection} connection */
  function isValidConnection(connection) {
    if (connection.source === connection.target) return false;
    const exists = get(edges).some(e =>
      e.source === connection.source && e.target === connection.target
    );
    return !exists;
  }

  /** @param {Connection} params */
  function onConnect(params) {
    const label = inferEdgeLabel(params.sourceHandle);
    /** @type {CanvasEdge} */
    const newEdge = {
      id: `edge-${Date.now()}`,
      source: params.source,
      target: params.target,
      sourceHandle: params.sourceHandle ?? undefined,
      targetHandle: params.targetHandle ?? undefined,
      label,
      data: { label },
    };
    edges.set([...get(edges), newEdge]);
  }

  /**
   * @param {string} processId
   * @param {Position} position
   */
  function onDropProcess(processId, position) {
    if (processId === 'util-multi-file-loader') {
      const newNode = {
        id: `node-${Date.now()}`,
        type: 'multiFileLoader',
        position,
        data: {
          label: 'Multi File Loader',
          processId,
          nodeType: 'multiFileLoader',
          status: 'pending',
          config: { entries: '[]' },
        },
      };
      nodes.set([...get(nodes), newNode]);
      return;
    }

    const process = ctx.processes.find(p => p.id === processId);
    if (!process) return;

    const newNode = {
      id: `node-${Date.now()}`,
      type: 'bmadProcess',
      position,
      data: { label: process.name, processId: process.id, process, status: 'pending', config: {} },
    };
    nodes.set([...get(nodes), newNode]);
  }

  /**
   * @param {{ storyId: string; status?: string }} storyData
   * @param {Position} position
   */
  function onDropStory(storyData, position) {
    /** @type {CanvasNode} */
    const newNode = {
      id: `story-node-${Date.now()}`,
      type: 'bmadProcess',
      position,
      data: {
        label: storyData.storyId,
        processId: 'bmad-dev-story',
        storyId: storyData.storyId,
        storyStatus: storyData.status,
        status: 'pending',
        config: { storyId: storyData.storyId },
      },
    };
    nodes.set([...get(nodes), newNode]);
  }

  /**
   * @param {'condition' | 'loop' | 'loopUntil' | 'transform' | 'merge'} nodeType
   * @param {Position} position
   */
  function onDropControlFlow(nodeType, position) {
    /** @type {CanvasNode} */
    const newNode = {
      id: `cf-${Date.now()}`,
      type: nodeType,
      position,
      data: { nodeType, label: controlFlowNames[nodeType] || nodeType, status: 'pending', config: {} },
    };
    nodes.set([...get(nodes), newNode]);
  }

  /** @param {{ node?: CanvasNode }} detail */
  function onNodeClick(detail) {
    const node = detail.node;
    if (node) {
      ctx.selectedNode = node;
      if (node.data?.status === 'awaiting_input') {
        openModalForNode(node.id);
      }
    }
  }

  /** @param {CanvasNode[]} deletedNodes */
  function onNodesDelete(deletedNodes) {
    // Filter out running nodes — they cannot be deleted
    const protectedIds = deletedNodes
      .filter(n => n.data?.status === 'running')
      .map(n => n.id);

    if (protectedIds.length > 0) {
      // Re-add protected nodes that xyflow already removed
      const protectedNodes = deletedNodes.filter(n => protectedIds.includes(n.id));
      nodes.set([...get(nodes), ...protectedNodes]);
    }

    // Clear selectedNode if it was deleted
    const currentlySelected = ctx.selectedNode;
    if (currentlySelected && deletedNodes.some(n => n.id === currentlySelected.id) && !protectedIds.includes(currentlySelected.id)) {
      ctx.selectedNode = null;
    }
    ctx.updateProgress();
  }

  /** @param {CanvasEdge[]} _deletedEdges */
  function onEdgesDelete(_deletedEdges) {
    // xyflow handles store removal; no additional state cleanup needed
  }

  /** @param {{ oldEdge: CanvasEdge; newConnection: Connection }} detail */
  function onReconnect(detail) {
    const { oldEdge, newConnection } = detail;
    edges.set(get(edges).map(e => {
      if (e.id === oldEdge.id) {
        const label = inferEdgeLabel(newConnection.sourceHandle);
        return {
          ...e,
          source: newConnection.source,
          target: newConnection.target,
          sourceHandle: newConnection.sourceHandle ?? undefined,
          targetHandle: newConnection.targetHandle ?? undefined,
          label,
          data: { label },
        };
      }
      return e;
    }));
  }

  /** @param {{ nodes: CanvasNode[]; edges: CanvasEdge[] }} selection */
  function onSelectionChange(selection) {
    if (selection.nodes.length === 1) {
      ctx.selectedNode = selection.nodes[0];
    } else if (selection.nodes.length === 0) {
      ctx.selectedNode = null;
    }
  }

  /**
   * @param {string} templateId
   * @param {Position} position
   * @param {string} [connectToNodeId]
   */
  function onAddTemplate(templateId, position, connectToNodeId) {
    const tpl = ctx.templates.find(t => t.id === templateId);
    if (!tpl || !tpl.nodes?.length) return;

    const ts = Date.now();
    /** @type {Record<string, string>} */
    const idMap = {};

    // Create new nodes offset from the click position
    /** @type {CanvasNode[]} */
    const newNodes = tpl.nodes.map((n, i) => {
      const newId = `tpl-${ts}-${i}`;
      idMap[n.id] = newId;
      return {
        id: newId,
        type: 'bmadProcess',
        position: { x: position.x + (n.position?.x || 0), y: position.y + (n.position?.y || 0) },
        data: {
          label: n.label,
          processId: n.processId,
          process: ctx.processes.find(p => p.id === n.processId) || null,
          status: 'pending',
          config: n.config || {},
        },
      };
    });

    // Recreate template edges with new IDs
    /** @type {CanvasEdge[]} */
    const newEdges = (tpl.edges || []).map((e, i) => ({
      id: `tpl-edge-${ts}-${i}`,
      source: idMap[e.source] || e.source,
      target: idMap[e.target] || e.target,
    })).filter(e => e.source && e.target);

    nodes.set([...get(nodes), ...newNodes]);
    edges.set([...get(edges), ...newEdges]);

    // Connect the source node to the first template node
    if (connectToNodeId && newNodes.length > 0) {
      edges.set([...get(edges), {
        id: `connect-${ts}`,
        source: connectToNodeId,
        target: newNodes[0].id,
      }]);
    }

    ctx.updateProgress();
  }

  /** @param {CustomEvent<{ nodeId: string; config: Record<string, unknown> }>} e */
  function onConfigUpdate(e) {
    const { nodeId, config } = e.detail;
    nodes.set(get(nodes).map(n => {
      if (n.id === nodeId) {
        return { ...n, data: { ...n.data, config } };
      }
      return n;
    }));
  }

  // ── NodeConfigPanel handlers (array editor, terminal, output viewer) ──

  /** @param {CustomEvent<{ nodeId: string; items: string[] }>} e */
  function handleEditItems(e) {
    const { nodeId, items } = e.detail;
    ctx.arrayModalNodeId = nodeId;
    ctx.arrayModalItems = items || [];
    ctx.showArrayModal = true;
  }

  /** @param {CustomEvent<string[]>} e */
  function handleArraySave(e) {
    const savedItems = e.detail;
    const arrayModalNodeId = ctx.arrayModalNodeId;
    nodes.set(get(nodes).map(n => {
      if (n.id === arrayModalNodeId) {
        const prevConfig = n.data?.config || {};
        const config = { ...prevConfig, items: savedItems.length > 0 ? JSON.stringify(savedItems) : '' };
        return { ...n, data: { ...n.data, config } };
      }
      return n;
    }));
    const selectedNode = ctx.selectedNode;
    if (selectedNode && selectedNode.id === arrayModalNodeId) {
      ctx.selectedNode = get(nodes).find(n => n.id === arrayModalNodeId) || selectedNode;
    }
    ctx.showArrayModal = false;
  }

  /** @param {CustomEvent<string>} e */
  async function onOpenTerminal(e) {
    ctx.terminalTarget = e.detail;
    ctx.showTerminalModal = true;
  }

  /** @param {CustomEvent<string>} e */
  async function handleOpenOutput(e) {
    const nodeId = e.detail;
    const node = get(nodes).find(n => n.id === nodeId);
    ctx.outputModalLabel = node?.data?.label || nodeId;
    ctx.outputLoading = true;
    ctx.showOutputModal = true;
    ctx.outputModalContent = '';

    try {
      ctx.outputModalContent = await GetNodeOutput(ctx.executionId || '', nodeId);
    } catch (err) {
      ctx.outputModalContent = 'Error loading output: ' + err;
    }
    ctx.outputLoading = false;
  }

  return {
    isValidConnection,
    inferEdgeLabel,
    onConnect,
    onDropProcess,
    onDropStory,
    onDropControlFlow,
    onNodeClick,
    onNodesDelete,
    onEdgesDelete,
    onReconnect,
    onSelectionChange,
    onAddTemplate,
    onConfigUpdate,
    handleEditItems,
    handleArraySave,
    onOpenTerminal,
    handleOpenOutput,
  };
}
