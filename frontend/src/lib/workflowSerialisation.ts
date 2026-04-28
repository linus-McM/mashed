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

import type {
  CanvasEdge,
  CanvasNode,
  Position,
  ProcessRegistryEntry,
  Workflow,
} from '../types/workflow';

/**
 * Permissive input shape for `workflowNodesToCanvasNodes` and the edge pair.
 *
 * Accepts both the upstream Go-generated `bmad.WorkflowNode` (whose `config`
 * is `Record<string, string>`) and our own `SerialisedWorkflowNode` (wider
 * `Record<string, unknown>`). Every non-id field is optional so legacy
 * fixtures without `tmuxTarget`/`status` still flow through.
 */
export interface SavedWorkflowNode {
  id: string;
  processId?: string;
  label?: string;
  position: Position;
  status?: string;
  config?: Record<string, unknown>;
  tmuxTarget?: string;
  storyId?: string;
  storyStatus?: string;
  nodeType?: string;
}

/** Permissive input shape for `workflowEdgesToCanvasEdges`. */
export interface SavedWorkflowEdge {
  id: string;
  source: string;
  target: string;
  sourceHandle?: string;
  targetHandle?: string;
}

/**
 * Output shape of `canvasNodesToWorkflowNodes`. Aligns with Go's
 * `bmad.WorkflowNode` constructor at the persistence boundary — `config`
 * is typed loosely (`Record<string, unknown>`) because users may stash
 * arbitrary JSON there and Go unmarshalls it into `map[string]interface{}`.
 */
export interface SerialisedWorkflowNode {
  id: string;
  processId: string;
  label: string;
  position: Position;
  status: string;
  config: Record<string, unknown>;
  tmuxTarget: string;
  storyId: string;
  nodeType: string;
}

/** Output shape of `canvasEdgesToWorkflowEdges` — mirrors `bmad.WorkflowEdge`. */
export interface SerialisedWorkflowEdge {
  id: string;
  source: string;
  target: string;
  sourceHandle: string;
  targetHandle: string;
}

/**
 * Safe XY coordinate extraction — a canvas node may arrive without a
 * `position` (e.g. mid-drag) or with `undefined` components. Coerce to
 * `0` so `Math.round` doesn't produce `NaN` and the hash stays stable.
 */
function coordOrZero(v: unknown): number {
  return typeof v === 'number' && Number.isFinite(v) ? v : 0;
}

/** Coerce an arbitrary value to a string, falling back to `fallback`. */
function stringOr(v: unknown, fallback: string): string {
  return typeof v === 'string' ? v : fallback;
}

/** Coerce to a plain-object record; falls back to `{}` when missing. */
function recordOr(v: unknown): Record<string, unknown> {
  return v && typeof v === 'object' && !Array.isArray(v)
    ? (v as Record<string, unknown>)
    : {};
}

/**
 * Structural canvas snapshot used for dirty detection and save-baselining.
 *
 * Must include node `type` in the hash so a command node at a given ID is
 * distinguishable from a process node at the same ID — AC-2.
 */
export function snapshotCanvas(
  nodesArr: CanvasNode[] | null | undefined,
  edgesArr: CanvasEdge[] | null | undefined,
  name: string | null | undefined,
): string {
  return JSON.stringify({
    name: name || '',
    nodes: (nodesArr || []).map((n) => ({
      id: n.id,
      type: n.type || '',
      x: Math.round(coordOrZero(n.position?.x)),
      y: Math.round(coordOrZero(n.position?.y)),
      processId: stringOr(n.data?.processId, ''),
      label: stringOr(n.data?.label, ''),
      config: recordOr(n.data?.config),
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
 */
export function canvasNodesToWorkflowNodes(
  canvasNodes: CanvasNode[] | null | undefined,
): SerialisedWorkflowNode[] {
  // The saved workflow is a *definition*, not an execution snapshot. Runtime
  // state (status, tmuxTarget, storyId) is owned by the live exec record on
  // disk under ~/.mashed/workflows/{execID}/execution.json — persisting it
  // here causes a freshly-opened workflow to show stale `running` /
  // `complete` badges as if the pipeline were active before Run is even
  // clicked. Always emit pending + blank runtime fields so reload starts
  // clean; the restoreForRepo overlay layers live state on top when an
  // exec is actually running.
  return (canvasNodes || []).map((n) => ({
    id: n.id,
    processId: stringOr(n.data?.processId, ''),
    label: stringOr(n.data?.label, ''),
    position: n.position,
    status: 'pending',
    config: recordOr(n.data?.config),
    tmuxTarget: '',
    storyId: '',
    nodeType: stringOr(n.data?.nodeType, ''),
  }));
}

/**
 * Transform svelte-flow canvas edges into the plain-object shape the Go
 * side's `bmad.WorkflowEdge` constructor accepts. Pair of
 * `canvasNodesToWorkflowNodes`.
 */
export function canvasEdgesToWorkflowEdges(
  canvasEdges: CanvasEdge[] | null | undefined,
): SerialisedWorkflowEdge[] {
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
 */
export function workflowNodesToCanvasNodes(
  wfNodes: SavedWorkflowNode[] | null | undefined,
  processes: ProcessRegistryEntry[] | null | undefined,
): CanvasNode[] {
  const procs = processes || [];
  return (wfNodes || []).map((n) => ({
    id: n.id,
    type: n.nodeType && n.nodeType !== 'process' ? n.nodeType : 'bmadProcess',
    position: n.position,
    data: {
      label: n.label ?? '',
      processId: n.processId || '',
      nodeType: n.nodeType || '',
      process: n.processId ? procs.find((p) => p.id === n.processId) ?? null : null,
      // Runtime fields are intentionally reset on load — any persisted
      // `running` / `complete` status, tmux session, or story id is stale
      // by definition (the binary that wrote it is gone). The
      // restoreForRepo overlay in WorkflowBuilder layers live exec state
      // back on top when an actual execution is in flight.
      status: 'pending',
      config: n.config ?? {},
      tmuxTarget: '',
      storyId: '',
      storyStatus: '',
    },
  }));
}

/**
 * Transform saved `bmad.WorkflowEdge`s into svelte-flow edges. The
 * `inferEdgeLabel` callback is injected so this module stays free of
 * WorkflowBuilder-specific helpers.
 */
export function workflowEdgesToCanvasEdges(
  wfEdges: SavedWorkflowEdge[] | null | undefined,
  inferEdgeLabel: (sourceHandle: string | undefined) => string,
): CanvasEdge[] {
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

// ---------------------------------------------------------------------------
// serialise / deserialise — Workflow <-> JSON string
//
// BDD Scenario 1 contract: serialise -> deserialise -> serialise is
//   idempotent for any well-formed `Workflow`. Key order is stable because
//   we project the canonical field set explicitly, so consumers that diff
//   two serialised strings can rely on byte equality.
//
// BDD Scenario 4 contract: deserialise rejects malformed JSON (including
//   `"{}"` with missing required fields) by returning `null`. It MUST NOT
//   throw, and MUST NOT return an `any`-shaped object — the intermediate
//   parse is typed as `unknown` and narrowed through `isWorkflowShape`
//   before the typed return.
// ---------------------------------------------------------------------------

/**
 * Type guard: a parsed JSON blob satisfies the required top-level shape
 * for a `Workflow`. We only check the fields the serialiser/deserialiser
 * round trip cares about — the Wails-generated `Workflow` class handles
 * its own richer validation when a bound method is actually called.
 */
function isWorkflowShape(raw: unknown): raw is Workflow {
  if (raw === null || typeof raw !== 'object' || Array.isArray(raw)) {
    return false;
  }
  const obj = raw as Record<string, unknown>;
  if (typeof obj.id !== 'string') return false;
  if (typeof obj.name !== 'string') return false;
  if (typeof obj.description !== 'string') return false;
  if (typeof obj.createdAt !== 'string') return false;
  if (typeof obj.updatedAt !== 'string') return false;
  if (typeof obj.isTemplate !== 'boolean') return false;
  if (!Array.isArray(obj.nodes)) return false;
  if (!Array.isArray(obj.edges)) return false;
  return true;
}

/**
 * Canonicalise a `Workflow` to a JSON string. Projecting the canonical
 * field set (rather than `JSON.stringify(workflow)`) guarantees stable
 * key order across runtimes — required by the Scenario 1 idempotent
 * round-trip assertion.
 */
export function serialise(workflow: Workflow): string {
  return JSON.stringify({
    id: workflow.id,
    name: workflow.name,
    description: workflow.description,
    repoPath: workflow.repoPath ?? '',
    nodes: workflow.nodes ?? [],
    edges: workflow.edges ?? [],
    isTemplate: workflow.isTemplate,
    templateId: workflow.templateId ?? '',
    createdAt: workflow.createdAt,
    updatedAt: workflow.updatedAt,
  });
}

/**
 * Parse a JSON string back into a `Workflow`. The intermediate value is
 * typed as `unknown` and narrowed by `isWorkflowShape` — no `as Workflow`
 * without prior narrowing. Any failure (invalid JSON, wrong shape, missing
 * required fields) resolves to `null`; this function never throws.
 */
export function deserialise(input: string): Workflow | null {
  if (typeof input !== 'string' || input.length === 0) {
    return null;
  }
  let parsed: unknown;
  try {
    parsed = JSON.parse(input);
  } catch {
    return null;
  }
  if (!isWorkflowShape(parsed)) {
    return null;
  }
  return parsed;
}
