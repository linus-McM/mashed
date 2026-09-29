// Execution restore + backend event wiring extracted verbatim from
// WorkflowBuilder.svelte (spec R31 — pure move, no behaviour change).
//
// createExecEvents(ctx) must be called during component initialisation:
// it registers every EventsOn listener synchronously (exactly when the
// component script used to), and the returned `destroy()` must be called
// from the component's onDestroy to tear them down.

import { get } from 'svelte/store';
import { tick } from 'svelte';
import { EventsOn } from '../../../wailsjs/runtime/runtime.js';
import { GetBmadWorkflow, GetBmadCurrentExecution, GetSprintStatus,
         ListAllMashedAssets } from '../../../wailsjs/go/main/App.js';
import {
  upsertPrompt,
  resolveInput,
  setValidationError,
  updateRound,
  pushToast,
  dismissNode,
} from '../../stores/interactiveInput';
import { computeAutoFill } from '../bmad/autoFill';
import {
  snapshotCanvas,
  COMMAND_NODE_SENTINEL_PHRASE,
  COMMAND_NODE_SENTINEL_DEFAULT_BODY,
} from '../workflowSerialisation';

/** @typedef {import('../../types/workflow').CanvasNode} CanvasNode */
/** @typedef {import('../../types/workflow').CanvasEdge} CanvasEdge */
/** @typedef {import('../../types/workflow').CanvasNodeData} CanvasNodeData */
/** @typedef {import('../types/wails').Workflow} Workflow */
/** @typedef {import('../types/wails').ProcessDef} ProcessDef */
/** @typedef {import('../types/wails').SprintStatus} SprintStatus */
/** @typedef {import('../types/wails').GroupedMashedAssets} GroupedMashedAssets */
/** @typedef {import('../types/wails').MashedAssetInfo} MashedAssetInfo */
/** @typedef {GroupedMashedAssets | { localCommands: MashedAssetInfo[]; globalCommands: MashedAssetInfo[]; localSkills: MashedAssetInfo[]; globalSkills: MashedAssetInfo[] }} GroupedAssetsLike */
/** @typedef {import('../../stores/interactiveInput').PendingPrompt} PendingPrompt */
/** @typedef {import('../../types/bmadEvents').BmadAwaitingInputEvent} BmadAwaitingInputEvent */
/** @typedef {import('../../types/bmadEvents').BmadNodeStatusEvent} BmadNodeStatusEvent */
/** @typedef {import('../../types/bmadEvents').BmadNodeArtifactsEvent} BmadNodeArtifactsEvent */
/** @typedef {import('../../types/bmadEvents').BmadExecutionStatusEvent} BmadExecutionStatusEvent */
/** @typedef {import('../../types/bmadEvents').BmadSprintUpdatedEvent} BmadSprintUpdatedEvent */
/** @typedef {import('../../types/bmadEvents').BmadInteractiveEventMap} BmadInteractiveEventMap */
/** @typedef {import('../../types/bmadEvents').BmadInteractiveHandlers} BmadInteractiveHandlers */

/**
 * @typedef {object} ExecEventsContext
 * @property {import('svelte/store').Writable<CanvasNode[]>} nodes
 * @property {import('svelte/store').Writable<CanvasEdge[]>} edges
 * @property {string} repoPath
 * @property {string | null} executionId
 * @property {string} executionStatus
 * @property {SprintStatus | null} sprintStatus
 * @property {CanvasNode | null} selectedNode
 * @property {Workflow | null} currentWorkflow
 * @property {string} workflowName
 * @property {string} lastSavedSnapshot
 * @property {string} failureToastMessage
 * @property {string} terminalTarget
 * @property {boolean} showTerminalModal
 * @property {GroupedAssetsLike} groupedMashedAssets
 * @property {boolean} assetsFetching
 * @property {boolean} assetsError
 * @property {Record<string, 'created' | 'updated'>} flashedPaths
 * @property {(wf: Workflow) => void} loadNodesEdges
 * @property {() => void} updateProgress
 * @property {() => Promise<boolean>} saveWorkflow
 */

/**
 * Flatten all four mashed-asset groups into a Map<path, serialized>
 * for efficient diff detection between fetches.
 * @param {GroupedAssetsLike | null | undefined} grouped
 * @returns {Map<string, string>}
 */
function flattenAssetPaths(grouped) {
  /** @type {Map<string, string>} */
  const map = new Map();
  /** @type {Array<'localCommands' | 'globalCommands' | 'localSkills' | 'globalSkills'>} */
  const keys = ['localCommands', 'globalCommands', 'localSkills', 'globalSkills'];
  for (const key of keys) {
    for (const asset of grouped?.[key] || []) {
      map.set(asset.path, JSON.stringify(asset));
    }
  }
  return map;
}

/** @param {ExecEventsContext} ctx */
export function createExecEvents(ctx) {
  const { nodes, edges } = ctx;

  /**
   * Restore the canvas for a repo on mount or after a cross-repo switch.
   *
   * Precedence:
   *   1. A non-terminal execution exists (running or paused) → load its
   *      workflow definition and repaint per-node statuses so the live
   *      pipeline is visible immediately.
   *   2. No live execution → leave whatever is on the canvas alone. The
   *      user may have explicitly loaded a different workflow via the
   *      sidebar and we do not want to clobber their selection.
   *
   * The restored state is baselined as the clean snapshot so tryLeave()
   * does not mistakenly treat restored RUNNING nodes as unsaved edits.
   */
  /** @param {string} path */
  async function restoreForRepo(path) {
    if (!path) return;
    try {
      const exec = await GetBmadCurrentExecution(path);
      if (!exec || !exec.workflowId) return;

      // Load the workflow definition behind the execution — this gives
      // us the canonical node/edge shape. GetBmadWorkflow returns the
      // saved workflow; we then overlay the exec's live node statuses
      // on top.
      let wf = null;
      try {
        wf = await GetBmadWorkflow(exec.workflowId);
      } catch (e) {
        console.warn('restoreForRepo: workflow backing exec is missing', e);
        return;
      }
      if (!wf) return;

      ctx.currentWorkflow = wf;
      ctx.workflowName = wf.name || 'Untitled Workflow';
      ctx.loadNodesEdges(wf);

      // Overlay live statuses from the running exec so the canvas shows
      // running / complete nodes without waiting for the next event.
      const byId = new Map((exec.nodes || []).map((n) => [n.id, n]));
      nodes.set(get(nodes).map((n) => {
        const live = byId.get(n.id);
        if (!live) return n;
        const prev = n.data || {};
        return {
          ...n,
          data: {
            ...prev,
            status: live.status || prev.status,
            tmuxTarget: live.tmuxTarget || prev.tmuxTarget,
            storyId: live.storyId || prev.storyId,
          },
        };
      }));

      ctx.executionId = exec.id;
      ctx.executionStatus = exec.status || 'running';
      ctx.updateProgress();

      // Loaded + overlaid state IS the clean baseline — the user has
      // not edited anything, so tryLeave should not fire on the next
      // navigation.
      ctx.lastSavedSnapshot = snapshotCanvas(get(nodes), get(edges), ctx.workflowName);
    } catch (e) {
      console.warn('restoreForRepo: failed', e);
    }
  }

  // Listen for live node status updates (scoped by execID)
  const cancelStatusListener = EventsOn('bmad:node:status', (/** @type {BmadNodeStatusEvent} */ event) => {
    const executionId = ctx.executionId;
    if (!event?.nodeId || (executionId && event.execId !== executionId)) return;
    nodes.set(get(nodes).map(n => {
      if (n.id === event.nodeId) {
        const prev = n.data || {};
        return { ...n, data: {
          ...prev,
          status: event.status,
          tmuxTarget: event.tmuxTarget || prev.tmuxTarget,
          iterationCount: event.iteration || prev.iterationCount,
        } };
      }
      return n;
    }));
    const updatedNode = get(nodes).find(n => n.id === event.nodeId);
    // Update selected node if it matches
    const selectedNode = ctx.selectedNode;
    if (selectedNode && selectedNode.id === event.nodeId && updatedNode) {
      ctx.selectedNode = updatedNode;
    }
    // Phase 2 sentinel detection: a command node flipping to `failed` is,
    // by construction, the "command nodes not yet runnable" case — the
    // backend executor's default branch logs that phrase then calls
    // failNode. When NodeStatusEvent grows a `message` field, we prefer
    // it; until then, svelte-flow type is the authoritative signal.
    if (event.status === 'failed') {
      const eventMessage = typeof event.message === 'string' ? event.message : '';
      const isCommandFailure =
        (updatedNode && updatedNode.type === 'command') ||
        eventMessage.toLowerCase().includes(COMMAND_NODE_SENTINEL_PHRASE);
      if (isCommandFailure) {
        ctx.failureToastMessage = eventMessage || COMMAND_NODE_SENTINEL_DEFAULT_BODY;
      }
    }
    ctx.updateProgress();
  });

  /** @type {ReturnType<typeof setTimeout> | undefined} */
  let autoFillSaveTimer;
  const cancelArtifactListener = EventsOn('bmad:node:artifacts', (/** @type {BmadNodeArtifactsEvent} */ event) => {
    const executionId = ctx.executionId;
    if (!event?.nodeId || (executionId && event.execId !== executionId)) return;
    nodes.set(get(nodes).map(n => {
      if (n.id === event.nodeId) {
        return { ...n, data: {
          ...n.data,
          artifactStatus: { found: event.found || [], missing: event.missing || [] },
        } };
      }
      return n;
    }));

    // breadcrumbs-06: Downstream auto-fill of empty InputPaths from resolved
    // upstream OutputPaths. Per plan lines 138-144, fan-out is deferred — a
    // single upstream path is copied to every connected downstream slot that
    // accepts the artifact and is currently empty. Populated slots never clobber.
    if (event.paths && typeof event.paths === 'object' && Object.keys(event.paths).length > 0) {
      const { updates } = computeAutoFill(
        { execId: event.execId, nodeId: event.nodeId, paths: event.paths },
        get(nodes),
        get(edges),
      );
      if (updates.length > 0) {
        /** @type {Map<string, Array<{ nodeId: string; artifactName: string; path: string }>>} */
        const byTarget = new Map();
        for (const u of updates) {
          const bucket = byTarget.get(u.nodeId);
          if (bucket) {
            bucket.push(u);
          } else {
            byTarget.set(u.nodeId, [u]);
          }
        }
        nodes.set(get(nodes).map(n => {
          const nodeUpdates = byTarget.get(n.id);
          if (!nodeUpdates) return n;
          const prevData = n.data || {};
          /** @type {Record<string, unknown>} */
          const prevConfig = prevData.config || {};
          const prevInputPaths = /** @type {Record<string, string> | undefined} */ (
            prevConfig.inputPaths
          );
          const nextInputPaths = { ...(prevInputPaths || {}) };
          for (const u of nodeUpdates) nextInputPaths[u.artifactName] = u.path;
          return {
            ...n,
            data: {
              ...prevData,
              config: { ...prevConfig, inputPaths: nextInputPaths },
            },
          };
        }));
        clearTimeout(autoFillSaveTimer);
        autoFillSaveTimer = setTimeout(() => { ctx.saveWorkflow(); }, 500);
      }
    }

    const selectedNode = ctx.selectedNode;
    if (selectedNode && selectedNode.id === event.nodeId) {
      ctx.selectedNode = get(nodes).find(n => n.id === event.nodeId) || selectedNode;
    }
  });

  const cancelExecListener = EventsOn('bmad:execution:status', (/** @type {BmadExecutionStatusEvent} */ event) => {
    const executionId = ctx.executionId;
    if (executionId && event.execId !== executionId) return;
    if (event?.status) ctx.executionStatus = event.status;
  });

  /** @type {ReturnType<typeof setTimeout> | undefined} */
  let sprintRefreshTimer;
  const cancelSprintListener = EventsOn('bmad:sprint:updated', (/** @type {BmadSprintUpdatedEvent} */ event) => {
    // Update node storyStatus immediately
    if (event?.storyId) {
      nodes.set(get(nodes).map(n => {
        if (n.data?.storyId === event.storyId) {
          return { ...n, data: { ...n.data, storyStatus: event.status } };
        }
        return n;
      }));
    }
    // Debounce full sprint data refresh (coalesces rapid node completions)
    clearTimeout(sprintRefreshTimer);
    sprintRefreshTimer = setTimeout(async () => {
      const repoPath = ctx.repoPath;
      if (repoPath) {
        try { ctx.sprintStatus = await GetSprintStatus(repoPath); } catch {}
      }
    }, 400);
  });

  // ── Live sidebar reload on disk changes (skills-watch-02) ──
  //
  // The backend AssetWatcher (skills-watch-01) emits 'bmad:assets:changed'
  // when skill/command files change on disk. We re-fetch and diff the list
  // to update the sidebar reactively.
  let assetFetchDirty = false;
  /** @type {ReturnType<typeof setTimeout> | undefined} */
  let flashTimer;
  /** @type {ReturnType<typeof setTimeout> | undefined} */
  let errorTimer;

  async function refetchMashedAssets() {
    if (ctx.assetsFetching) {
      assetFetchDirty = true;
      return;
    }
    ctx.assetsFetching = true;
    ctx.assetsError = false;

    // Capture scroll position before refetch
    const tabContent = document.querySelector('.sidebar .tab-content');
    const prevScroll = tabContent?.scrollTop ?? 0;

    try {
      const prevPaths = flattenAssetPaths(ctx.groupedMashedAssets);
      const newGrouped = await ListAllMashedAssets(ctx.repoPath || '');

      // Compute diff: new paths → 'created', changed paths → 'updated'
      const newPaths = flattenAssetPaths(newGrouped);
      /** @type {Record<string, 'created' | 'updated'>} */
      const flashed = {};
      for (const [path, serialized] of newPaths) {
        if (!prevPaths.has(path)) {
          flashed[path] = 'created';
        } else if (prevPaths.get(path) !== serialized) {
          flashed[path] = 'updated';
        }
      }

      ctx.groupedMashedAssets = newGrouped || { localCommands: [], globalCommands: [], localSkills: [], globalSkills: [] };
      ctx.flashedPaths = flashed;

      // Clear flash classes after animation completes
      clearTimeout(flashTimer);
      if (Object.keys(flashed).length > 0) {
        flashTimer = setTimeout(() => { ctx.flashedPaths = {}; }, 400);
      }

      // Restore scroll after Svelte re-renders the keyed list
      await tick();
      if (tabContent) tabContent.scrollTop = prevScroll;

    } catch (e) {
      console.error('bmad:assets:changed refetch failed:', e);
      ctx.assetsError = true;
      clearTimeout(errorTimer);
      errorTimer = setTimeout(() => { ctx.assetsError = false; }, 2000);
    } finally {
      ctx.assetsFetching = false;
      if (assetFetchDirty) {
        assetFetchDirty = false;
        refetchMashedAssets();
      }
    }
  }

  const cancelAssetsListener = EventsOn('bmad:assets:changed', () => {
    refetchMashedAssets();
  });

  // S6: interactive input event wiring (awaiting_input / input_resolved /
  // input_invalid / round_complete / gate_satisfied / round_limit / aborted).
  /**
   * @param {BmadAwaitingInputEvent | null | undefined} payload
   * @returns {PendingPrompt | null}
   */
  function annotatePrompt(payload) {
    if (!payload || !payload.nodeId) return null;
    const repoPath = ctx.repoPath;
    const node = get(nodes).find((n) => n.id === payload.nodeId);
    const process = node?.data?.process || /** @type {Partial<ProcessDef>} */ ({});
    const specs = Array.isArray(process.inputSpecs) ? process.inputSpecs : [];
    const spec = specs.find((s) => s && s.id === payload.inputId) || /** @type {Partial<import('../types/wails').InputSpec>} */ ({});
    return {
      execId: payload.execId || ctx.executionId || '',
      nodeId: payload.nodeId,
      inputId: payload.inputId,
      prompt: payload.prompt ?? '',
      shape: payload.shape ?? '',
      options: payload.options || [],
      round: payload.round || 0,
      createdAt: payload.createdAt || Date.now(),
      promptId: payload.promptId || '',
      required: payload.required !== undefined ? payload.required : (spec.required ?? true),
      helpText: payload.helpText || spec.helpText || '',
      maxLength: payload.maxLength || spec.maxLength || 0,
      maxRounds: payload.maxRounds || process.gate?.maxRounds || 0,
      repoName: repoPath ? repoPath.split('/').pop() : '',
      repoPath,
      structured: payload.structured,
      lastOutput: payload.lastOutput,
    };
  }

  /**
   * @param {string} nodeId
   * @param {Partial<CanvasNodeData>} patch
   */
  function applyNodeDataPatch(nodeId, patch) {
    nodes.set(get(nodes).map((n) => (n.id === nodeId ? { ...n, data: { ...n.data, ...patch } } : n)));
  }

  /** @param {string} nodeId */
  function flashGate(nodeId) {
    applyNodeDataPatch(nodeId, { gateFlash: true, status: 'complete' });
    setTimeout(() => applyNodeDataPatch(nodeId, { gateFlash: false }), 400);
  }

  /** @type {BmadInteractiveHandlers} */
  const bmadEventHandlers = {
    'bmad:node:awaiting_input': (event) => {
      const executionId = ctx.executionId;
      if (executionId && event?.execId && event.execId !== executionId) return;
      const prompt = annotatePrompt(event);
      if (!prompt) return;
      upsertPrompt(prompt);
      applyNodeDataPatch(prompt.nodeId, { status: 'awaiting_input' });
      if (prompt.round > 0) updateRound(prompt.nodeId, prompt.round);
    },
    'bmad:node:input_resolved': (event) => {
      const executionId = ctx.executionId;
      if (!event?.nodeId || !event?.inputId) return;
      if (executionId && event.execId && event.execId !== executionId) return;
      resolveInput(event.nodeId, event.inputId);
    },
    'bmad:node:awaiting_dismissed': (event) => {
      const executionId = ctx.executionId;
      if (!event?.nodeId || !event?.inputId) return;
      if (executionId && event.execId && event.execId !== executionId) return;
      // Pane resumed activity mid-suspension — clear the stale prompt and
      // flip the node back to RUNNING. Backend will re-emit awaiting_input
      // with the fresh capture once claude returns to the idle prompt.
      resolveInput(event.nodeId, event.inputId);
      applyNodeDataPatch(event.nodeId, { status: 'running' });
    },
    'bmad:node:input_invalid': (event) => {
      const executionId = ctx.executionId;
      if (!event?.nodeId || !event?.inputId) return;
      if (executionId && event.execId && event.execId !== executionId) return;
      setValidationError(event.nodeId, event.inputId, event.reason || 'Invalid input');
    },
    'bmad:node:round_complete': (event) => {
      const executionId = ctx.executionId;
      if (!event?.nodeId) return;
      if (executionId && event.execId && event.execId !== executionId) return;
      updateRound(event.nodeId, event.round || 0);
      applyNodeDataPatch(event.nodeId, { nodeRound: event.round || 0 });
    },
    'bmad:node:gate_satisfied': (event) => {
      const executionId = ctx.executionId;
      if (!event?.nodeId) return;
      if (executionId && event.execId && event.execId !== executionId) return;
      dismissNode(event.nodeId);
      flashGate(event.nodeId);
      pushToast('success', event.reason || 'Gate satisfied');
    },
    'bmad:node:round_limit': (event) => {
      const executionId = ctx.executionId;
      if (!event?.nodeId) return;
      if (executionId && event.execId && event.execId !== executionId) return;
      dismissNode(event.nodeId);
      flashGate(event.nodeId);
      pushToast('warn', 'Reached round limit — process ended');
    },
    'bmad:node:aborted': (event) => {
      const executionId = ctx.executionId;
      if (!event?.nodeId) return;
      if (executionId && event.execId && event.execId !== executionId) return;
      dismissNode(event.nodeId);
      applyNodeDataPatch(event.nodeId, { status: 'failed' });
      pushToast('error', event.reason || 'Node aborted');
    },
    'bmad:node:session_dead': (event) => {
      const executionId = ctx.executionId;
      if (!event?.nodeId) return;
      if (executionId && event.execId && event.execId !== executionId) return;
      applyNodeDataPatch(event.nodeId, { tmuxTarget: '', sessionDead: true });
      // If terminal modal is open on the just-dead target, close it so the
      // user isn't left staring at a stale transcript.
      const terminalTarget = ctx.terminalTarget;
      if (terminalTarget && event.tmuxTarget && terminalTarget === event.tmuxTarget) {
        ctx.showTerminalModal = false;
      }
      pushToast('warn', 'Terminal session ended');
    },
  };

  // Object.entries() erases the key→value narrowing from BmadInteractiveHandlers,
  // so we iterate the keys explicitly to keep the handler signature aligned
  // with its event name when the bundle gets wired to EventsOn.
  const bmadListenerCancels = (
    /** @type {(keyof BmadInteractiveEventMap)[]} */ (Object.keys(bmadEventHandlers))
  ).map((name) => EventsOn(name, /** @type {(e: unknown) => void} */ (bmadEventHandlers[name])));

  if (import.meta.env.DEV && typeof window !== 'undefined') {
    window.__mashedEmitBmadEvent = (name, payload) => {
      const fn = bmadEventHandlers[name];
      if (!fn) return false;
      fn(payload);
      return true;
    };
  }

  // Skip for non-required prompts: send an empty string; backend treats as skip.
  /** @param {PendingPrompt | null | undefined} prompt */
  async function handleInputSkip(prompt) {
    if (!prompt) return;
    try {
      const { RespondToInput } = await import('../../../wailsjs/go/main/App.js');
      await RespondToInput(prompt.execId || ctx.executionId || '', prompt.nodeId, prompt.inputId, '');
    } catch (e) {
      console.warn('skip input failed', e);
    }
  }

  /** Tear down every listener/timer registered above. Call from onDestroy. */
  function destroy() {
    if (cancelStatusListener) cancelStatusListener();
    if (cancelArtifactListener) cancelArtifactListener();
    if (cancelExecListener) cancelExecListener();
    if (cancelSprintListener) cancelSprintListener();
    if (cancelAssetsListener) cancelAssetsListener();
    bmadListenerCancels.forEach((c) => c && c());
    if (import.meta.env.DEV && typeof window !== 'undefined') {
      delete window.__mashedEmitBmadEvent;
    }
    clearTimeout(flashTimer);
    clearTimeout(errorTimer);
    clearTimeout(autoFillSaveTimer);
  }

  return { restoreForRepo, refetchMashedAssets, handleInputSkip, destroy };
}
