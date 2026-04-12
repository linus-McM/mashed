// skills-cmd-03: pure serialisation helpers extracted from WorkflowBuilder.svelte.
//
// These functions were inlined inside the Svelte component, which made the
// AC-1 (save+reload round-trip), AC-2 (snapshot drift detection) and AC-4
// (restoreForRepo type-agnostic position preservation) acceptance criteria
// hard to cover with unit tests — the canvas<->WorkflowDef transforms were
// closed over component state.
//
// Extracting them here serves two goals:
//   1. Lock the transforms with vitest so a future refactor cannot silently
//      drop command nodes through a type-filtering shortcut.
//   2. Keep the component thin — the store-mutating call sites in
//      WorkflowBuilder.svelte now delegate to these pure functions instead
//      of re-implementing the mapping inline.
//
// Contract: every function here is pure. No store access, no side effects,
// no async. Inputs go in, outputs come out, and the tests cover every
// type+config combination Phase 2 cares about.

/**
 * Structural canvas snapshot used for dirty detection and save-baselining.
 *
 * Must include node `type` in the hash so a command node at a given ID is
 * distinguishable from a process node at the same ID — AC-2.
 *
 * @param {Array<object>} nodesArr svelte-flow nodes
 * @param {Array<object>} edgesArr svelte-flow edges
 * @param {string} name workflow name
 * @returns {string} JSON-stringified snapshot (stable key order)
 */
export function snapshotCanvas(nodesArr, edgesArr, name) {
  return JSON.stringify({
    name: name || '',
    nodes: (nodesArr || []).map((n) => ({
      id: n.id,
      type: n.type || '',
      x: Math.round(n.position?.x || 0),
      y: Math.round(n.position?.y || 0),
      processId: n.data?.processId || '',
      label: n.data?.label || '',
      config: n.data?.config || {},
    })),
    edges: (edgesArr || []).map((e) => ({
      id: e.id,
      source: e.source,
      target: e.target,
      sourceHandle: e.sourceHandle || '',
      targetHandle: e.targetHandle || '',
    })),
  });
}

/**
 * Transform svelte-flow canvas nodes into the plain-object shape the Go
 * side's `bmad.WorkflowNode` constructor accepts.
 *
 * Type-agnostic: command, bmadProcess, condition, loop, etc. all flow
 * through the same mapping. The `nodeType` field carries the discriminator
 * so `loadNodesEdges` can reconstruct the svelte-flow type on reload.
 *
 * @param {Array<object>} canvasNodes
 * @returns {Array<object>}
 */
export function canvasNodesToWorkflowNodes(canvasNodes) {
  return (canvasNodes || []).map((n) => ({
    id: n.id,
    processId: n.data?.processId || '',
    label: n.data?.label || '',
    position: n.position,
    status: n.data?.status || 'pending',
    config: n.data?.config || {},
    tmuxTarget: n.data?.tmuxTarget || '',
    storyId: n.data?.storyId || '',
    nodeType: n.data?.nodeType || '',
  }));
}

/**
 * Transform svelte-flow canvas edges into the plain-object shape the Go
 * side's `bmad.WorkflowEdge` constructor accepts. Pair of
 * `canvasNodesToWorkflowNodes`.
 *
 * @param {Array<object>} canvasEdges
 * @returns {Array<object>}
 */
export function canvasEdgesToWorkflowEdges(canvasEdges) {
  return (canvasEdges || []).map((e) => ({
    id: e.id,
    source: e.source,
    target: e.target,
    sourceHandle: e.sourceHandle || '',
    targetHandle: e.targetHandle || '',
  }));
}

/**
 * Phase 2 sentinel emitted by `internal/bmad/executor.go` when a command
 * node is dispatched. Lives here so the frontend listener, the toast
 * component, and the tests all pull from a single source of truth.
 */
export const COMMAND_NODE_SENTINEL_PHRASE = 'command nodes not yet runnable';
export const COMMAND_NODE_SENTINEL_DEFAULT_BODY =
  'Phase 3 will add real execution.';

/**
 * Transform saved `bmad.WorkflowNode`s back into svelte-flow canvas nodes.
 *
 * AC-1: a node saved with `nodeType === 'command'` MUST round-trip to
 * svelte-flow type `'command'` (and thus render as `CommandNode.svelte`).
 *
 * Any non-empty `nodeType` that is NOT the legacy `'process'` string is
 * used verbatim as the svelte-flow type. Legacy workflows (`nodeType` blank
 * or `'process'`) fall back to `'bmadProcess'`.
 *
 * AC-4: positions are copied through untouched regardless of type. No
 * type-based filtering occurs; the function operates on node IDs only.
 *
 * @param {Array<object>} wfNodes
 * @param {Array<object>} processes process registry for the `process` backref
 * @returns {Array<object>}
 */
export function workflowNodesToCanvasNodes(wfNodes, processes) {
  const procs = processes || [];
  return (wfNodes || []).map((n) => ({
    id: n.id,
    type: n.nodeType && n.nodeType !== 'process' ? n.nodeType : 'bmadProcess',
    position: n.position,
    data: {
      label: n.label,
      processId: n.processId || '',
      nodeType: n.nodeType || '',
      process: n.processId ? procs.find((p) => p.id === n.processId) || null : null,
      status: n.status || 'pending',
      config: n.config || {},
      tmuxTarget: n.tmuxTarget || '',
      storyId: n.storyId || '',
      storyStatus: n.storyStatus || '',
    },
  }));
}

/**
 * Transform saved `bmad.WorkflowEdge`s into svelte-flow edges. The
 * `inferEdgeLabel` callback is injected so this module stays free of
 * WorkflowBuilder-specific helpers.
 *
 * @param {Array<object>} wfEdges
 * @param {(sourceHandle: string|undefined) => string} inferEdgeLabel
 * @returns {Array<object>}
 */
export function workflowEdgesToCanvasEdges(wfEdges, inferEdgeLabel) {
  return (wfEdges || []).map((e) => {
    const label = inferEdgeLabel(e.sourceHandle);
    return {
      id: e.id,
      source: e.source,
      target: e.target,
      sourceHandle: e.sourceHandle || undefined,
      targetHandle: e.targetHandle || undefined,
      label,
      data: { label },
    };
  });
}
