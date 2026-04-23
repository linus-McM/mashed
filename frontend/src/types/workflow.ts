// types/workflow.ts — frontend-facing Workflow domain types.
//
// Canonical shape lives in Go (`internal/bmad/types.go`): `WorkflowDef`,
// `WorkflowNode`, `WorkflowEdge`. Wails regenerates `wailsjs/go/models.ts`
// on `wails dev`, and `$lib/types/wails` re-exports those under friendly
// names. This module is the single import site for the frontend:
//   import type { Workflow, WorkflowNode, WorkflowEdge } from '../types/workflow';
//
// A future Go rename only needs the re-export in `$lib/types/wails` updated
// — call sites keep compiling.

import type {
  Workflow as WailsWorkflow,
  WorkflowNode as WailsWorkflowNode,
  WorkflowEdge as WailsWorkflowEdge,
  Position as WailsPosition,
  ProcessDef,
} from '../lib/types/wails';

/** XY coordinate on the workflow canvas (mirrors `bmad.Position`). */
export type Position = WailsPosition;

/** A single edge connecting two workflow nodes. */
export type WorkflowEdge = WailsWorkflowEdge;

/** A single workflow node (process, command, condition, loop, transform, …). */
export type WorkflowNode = WailsWorkflowNode;

/** A complete workflow definition — the unit that Wails `SaveWorkflow` accepts. */
export type Workflow = WailsWorkflow;

// ---------------------------------------------------------------------------
// Canvas-side shapes used by WorkflowBuilder + workflowSerialisation.
//
// These are NOT in Go — they're the svelte-flow shapes the UI manipulates
// before/after Wails round-trip. Kept here so `workflowSerialisation.ts` can
// annotate its public API without re-declaring them at each call site.
// ---------------------------------------------------------------------------

/**
 * `data` payload attached to a svelte-flow canvas node. Mirrors the
 * WorkflowBuilder `data: { … }` block — every field is optional so legacy
 * canvases (pre-command-nodes) and freshly-dropped nodes both validate.
 *
 * Additional fields (artifactStatus, iterationCount, nodeRound, gateFlash)
 * are written by the live execution event listeners — ProcessNode.svelte
 * and its siblings read them off `data` to animate the running state.
 */
export interface CanvasNodeData {
  label?: string;
  processId?: string;
  nodeType?: string;
  // `Partial<ProcessDef>` (not `ProcessDef`) because the serialiser may seed
  // this from a stripped-down `ProcessRegistryEntry` (tests pass minimal
  // fixtures). All consumers already read with fallbacks — see
  // `ProcessNode.svelte` which treats `data.process || {}` as a partial.
  process?: Partial<ProcessDef> | null;
  status?: string;
  config?: Record<string, unknown>;
  tmuxTarget?: string;
  storyId?: string;
  storyStatus?: string;
  // Live execution overlay — written by `bmad:node:status`.
  iterationCount?: number;
  // Artifact resolution from `bmad:node:artifacts`.
  artifactStatus?: { found: string[]; missing: string[] };
  // Interactive gate flash animation & current iteration round.
  gateFlash?: boolean;
  nodeRound?: number;
  // Set by `bmad:node:session_dead` when the liveness poller observes that
  // the node's tmux session has been terminated outside the executor.
  sessionDead?: boolean;
}

/** A svelte-flow canvas node as produced by WorkflowBuilder. */
export interface CanvasNode {
  id: string;
  type?: string;
  position: Position;
  data?: CanvasNodeData;
}

/** A svelte-flow canvas edge as produced by WorkflowBuilder. */
export interface CanvasEdge {
  id: string;
  source: string;
  target: string;
  sourceHandle?: string;
  targetHandle?: string;
  label?: string;
  data?: { label?: string };
}

/**
 * Minimal process-registry entry `workflowNodesToCanvasNodes` cross-references
 * to resolve the `process` backref on a restored canvas node. The full
 * `ProcessDef` is richer (see `$lib/types/wails`), but the serialiser only
 * ever reads `id`. No index signature — that would prevent structural
 * assignment from the Wails-generated `ProcessDef` class (which carries
 * fixed, non-index-compatible properties).
 */
export interface ProcessRegistryEntry {
  id: string;
}
