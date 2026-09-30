// Unsaved-changes tracking + leave intercept extracted verbatim from
// WorkflowBuilder.svelte (spec R31 — pure move, no behaviour change).
//
// ── Unsaved-changes tracking ────────────────────────────────────────
//
// lastSavedSnapshot captures the canvas shape at the most recent
// successful save (or load). The reactive `isDirty` check compares a
// fresh snapshot against this baseline so tryLeave() can decide whether
// navigation is free, silent-autosave, or needs the NameWorkflowModal.
//
// Dirty detection is deliberately structural (nodes + edges + name)
// rather than change-counting: it's stable across undo, reorder, and
// config edits without bookkeeping inside each mutation path. The
// downside is an O(nodes+edges) JSON.stringify on every leave, which
// is negligible for canvases with dozens of nodes.

import { get } from 'svelte/store';
import { ListAllMashedAssets } from '../../../wailsjs/go/main/App.js';
import { snapshotCanvas } from '../workflowSerialisation';

/** @typedef {import('../../types/workflow').CanvasNode} CanvasNode */
/** @typedef {import('../../types/workflow').CanvasEdge} CanvasEdge */
/** @typedef {import('../types/wails').Workflow} Workflow */
/** @typedef {import('../types/wails').GroupedMashedAssets} GroupedMashedAssets */
/** @typedef {import('../types/wails').MashedAssetInfo} MashedAssetInfo */

/**
 * @typedef {object} LeaveInterceptContext
 * @property {import('svelte/store').Writable<CanvasNode[]>} nodes
 * @property {import('svelte/store').Writable<CanvasEdge[]>} edges
 * @property {boolean} isDirty
 * @property {boolean} discardLatch
 * @property {boolean} showNameModal
 * @property {Workflow | null} currentWorkflow
 * @property {string} workflowName
 * @property {string} lastSavedSnapshot
 * @property {() => Promise<boolean>} saveWorkflow
 * @property {() => void} onLeave  Called when a Back request is allowed to proceed.
 * @property {string} repoPath
 * @property {MashedAssetInfo | null} editingAsset
 * @property {string} lastSavedPath
 * @property {string} saveToastPath
 * @property {boolean} showSaveToast
 * @property {GroupedMashedAssets | { localCommands: MashedAssetInfo[]; globalCommands: MashedAssetInfo[]; localSkills: MashedAssetInfo[]; globalSkills: MashedAssetInfo[] }} groupedMashedAssets
 */

/**
 * Dirty flag re-evaluates whenever nodes/edges/name change. Discard-latch
 * short-circuits to false so the view-switch triggered by a Discard click
 * never bounces back into the modal.
 * @param {boolean} discardLatch
 * @param {CanvasNode[]} nodes
 * @param {CanvasEdge[]} edges
 * @param {string} workflowName
 * @param {string} lastSavedSnapshot
 */
export function computeIsDirty(discardLatch, nodes, edges, workflowName, lastSavedSnapshot) {
  return (
    !discardLatch &&
    snapshotCanvas(nodes, edges, workflowName) !== lastSavedSnapshot
  );
}

/** @param {LeaveInterceptContext} ctx */
export function createLeaveIntercept(ctx) {
  const { nodes, edges } = ctx;

  /** @type {null | ((decision: 'save' | 'discard' | 'cancel', name?: string) => void)} */
  let nameModalResolver = null;

  // ── Leave intercept ─────────────────────────────────────────────────
  //
  // tryLeave returns true when navigation should proceed, false when it
  // should be cancelled (e.g. user clicked Cancel in NameWorkflowModal).
  //
  // Decision tree:
  //   1. Not dirty (or canvas empty)         → proceed
  //   2. Dirty + has saved ID                → silent save + proceed
  //   3. Dirty + unnamed (no ID yet)         → show modal, await choice
  //         • save     → name & save, proceed
  //         • discard  → blank canvas, proceed
  //         • cancel   → stay put, return false
  //
  // Exposed via `export` on the component so App.svelte can call it
  // imperatively via bind:this before triggering cross-view navigation
  // (e.g. when a snackbar click wants to switch repos).
  /** @returns {Promise<boolean>} */
  async function tryLeave() {
    // Nothing to protect against.
    if (!ctx.isDirty || get(nodes).length === 0) return true;

    // Case 2 — named workflow, silent save. If the save fails we MUST
    // refuse the leave; otherwise the user would silently lose their
    // edits when navigating back or switching repos.
    if (ctx.currentWorkflow?.id) {
      const ok = await ctx.saveWorkflow();
      return ok;
    }

    // Case 3 — unnamed draft. Prompt the user.
    return new Promise((resolve) => {
      nameModalResolver = async (decision, name) => {
        nameModalResolver = null;
        ctx.showNameModal = false;
        if (decision === 'cancel') {
          resolve(false);
          return;
        }
        if (decision === 'discard') {
          // Blank the canvas BEFORE resolving so the reactive `isDirty`
          // check on the next tick sees an empty canvas and no longer
          // triggers the modal if the consumer re-queries.
          ctx.discardLatch = true;
          nodes.set([]);
          edges.set([]);
          ctx.currentWorkflow = null;
          ctx.workflowName = 'Untitled Workflow';
          ctx.lastSavedSnapshot = snapshotCanvas([], [], 'Untitled Workflow');
          // Release the latch on the next microtask so subsequent user
          // edits are tracked again.
          queueMicrotask(() => { ctx.discardLatch = false; });
          resolve(true);
          return;
        }
        // decision === 'save' — commit the name the user typed and
        // attempt the write. On backend failure we resolve false so
        // the parent back-click is aborted and the modal stays closed
        // (the user can click Back again to retry). Future polish
        // could re-open the modal with an inline error instead.
        if (name && name.length > 0) ctx.workflowName = name;
        const saved = await ctx.saveWorkflow();
        resolve(saved);
      };
      ctx.showNameModal = true;
    });
  }

  /** @param {CustomEvent<{ name: string }>} e */
  function handleNameModalSave(e) {
    if (nameModalResolver) nameModalResolver('save', e.detail?.name);
  }
  function handleNameModalDiscard() {
    if (nameModalResolver) nameModalResolver('discard');
  }
  function handleNameModalCancel() {
    if (nameModalResolver) nameModalResolver('cancel');
  }

  /** @param {CustomEvent<{ path?: string }>} e */
  async function handleAssetSaved(e) {
    const savedPath = e.detail?.path || '';
    ctx.editingAsset = null;
    ctx.lastSavedPath = savedPath;

    // Refresh sidebar data.
    try {
      ctx.groupedMashedAssets = await ListAllMashedAssets(ctx.repoPath || '');
    } catch {
      ctx.groupedMashedAssets = { localCommands: [], globalCommands: [], localSkills: [], globalSkills: [] };
    }

    // Show toast and auto-dismiss after 2500ms.
    ctx.saveToastPath = savedPath;
    ctx.showSaveToast = true;
    setTimeout(() => { ctx.showSaveToast = false; }, 2500);

    // Clear row flash highlight after 400ms.
    setTimeout(() => { ctx.lastSavedPath = ''; }, 400);
  }

  async function handleBackRequest() {
    const ok = await tryLeave();
    if (ok) ctx.onLeave();
  }

  return {
    handleAssetSaved,
    tryLeave,
    handleNameModalSave,
    handleNameModalDiscard,
    handleNameModalCancel,
    handleBackRequest,
  };
}
