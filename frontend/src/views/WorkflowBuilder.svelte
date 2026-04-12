<script>
  import { SvelteFlowProvider } from '@xyflow/svelte';
  import '@xyflow/svelte/dist/style.css';
  import { writable } from 'svelte/store';
  import { createEventDispatcher, onMount, onDestroy } from 'svelte';
  import { Save } from 'lucide-svelte';
  import { EventsOn } from '../../wailsjs/runtime/runtime.js';
  import { GetBmadProcesses, ListBmadTemplates, ListBmadWorkflowsByRepo,
           SaveBmadWorkflow, GetBmadWorkflow, CreateFromTemplate,
           DeleteBmadWorkflow, ListBmadAgents, ListAllAgents, SaveBmadAgent, DeleteBmadAgent,
           StartBmadWorkflow, PauseBmadWorkflow, ResumeBmadWorkflow, StopBmadWorkflow,
           GetTerminalPort, GetSprintStatus, GetNodeOutput, ListModels,
           GetBmadCurrentExecution,
           ListAllMashedAssets } from '../../wailsjs/go/main/App.js';
  // bmad.WorkflowDef / WorkflowNode / WorkflowEdge are Wails-generated
  // classes (not interfaces) with a `convertValues` method on the
  // prototype. Passing a bare object literal to SaveBmadWorkflow trips
  // a structural mismatch ("Property 'convertValues' is missing") that
  // the IDE surfaces even though the call would round-trip correctly
  // through JSON at runtime. Wrapping the literal in the generated
  // constructor produces a real class instance that satisfies the
  // type-check AND makes the field list a single source of truth with
  // the Go side of the binding.
  import { bmad } from '../../wailsjs/go/models';
  import Terminal from '../components/Terminal.svelte';
  import ProcessSidebar from '../components/bmad/ProcessSidebar.svelte';
  import CanvasPane from '../components/bmad/CanvasPane.svelte';
  import ProcessNode from '../components/bmad/ProcessNode.svelte';
  import CommandNode from '../components/bmad/CommandNode.svelte';
  import ConditionNode from '../components/bmad/ConditionNode.svelte';
  import LoopNode from '../components/bmad/LoopNode.svelte';
  import LoopUntilNode from '../components/bmad/LoopUntilNode.svelte';
  import TransformNode from '../components/bmad/TransformNode.svelte';
  import MergeNode from '../components/bmad/MergeNode.svelte';
  import ExecutionBar from '../components/bmad/ExecutionBar.svelte';
  import NodeConfigPanel from '../components/bmad/NodeConfigPanel.svelte';
  import AgentConfigModal from '../components/bmad/AgentConfigModal.svelte';
  import OutputViewerModal from '../components/bmad/OutputViewerModal.svelte';
  import ArrayEditorModal from '../components/bmad/ArrayEditorModal.svelte';
  import QuestionResponseModal from '../components/bmad/QuestionResponseModal.svelte';
  import NameWorkflowModal from '../components/bmad/NameWorkflowModal.svelte';
  import SkillEditorModal from '../components/bmad/SkillEditorModal.svelte';
  import RepoContextBar from '../components/bmad/RepoContextBar.svelte';
  import CanvasFailureToast from '../components/bmad/CanvasFailureToast.svelte';
  import { parseFriendlyTarget } from '../lib/bmadSessionName';
  import {
    snapshotCanvas,
    canvasNodesToWorkflowNodes,
    canvasEdgesToWorkflowEdges,
    workflowNodesToCanvasNodes,
    workflowEdgesToCanvasEdges,
    COMMAND_NODE_SENTINEL_PHRASE,
    COMMAND_NODE_SENTINEL_DEFAULT_BODY,
  } from '../lib/workflowSerialisation.js';

  export let repoPath = '';
  export let repoBranch = '';
  export let pendingQuestion = null;

  const dispatch = createEventDispatcher();

  let showQuestionModal = false;
  let activeQuestion = null;

  // Only open the modal on a truthy transition when not already showing.
  // Without the !showQuestionModal guard, re-setting pendingQuestion to the
  // same value (e.g. parent re-render) would silently reset activeQuestion
  // while a submit may be in-flight.
  $: if (pendingQuestion && !showQuestionModal) {
    activeQuestion = pendingQuestion;
    showQuestionModal = true;
  }

  function handleQuestionResponded() {
    const nodeId = activeQuestion?.nodeId;
    // Dispatch BEFORE clearing activeQuestion so the parent's handler
    // observes the event before any reactive chain can re-set pendingQuestion.
    dispatch('question-responded', { nodeId });
    showQuestionModal = false;
    activeQuestion = null;
  }

  function handleQuestionClose() {
    showQuestionModal = false;
    // Leave activeQuestion intact so the snackbar remains visible upstream.
  }

  const nodeTypes = {
    bmadProcess: ProcessNode,
    condition: ConditionNode,
    loop: LoopNode,
    loopUntil: LoopUntilNode,
    transform: TransformNode,
    merge: MergeNode,
    command: CommandNode,
  };

  const nodes = writable([]);
  const edges = writable([]);

  let processes = [];
  let templates = [];
  let savedWorkflows = [];
  /** @type {{ bmadAgents: any[], localAgents: any[], globalAgents: any[] }} */
  let groupedAgents = { bmadAgents: [], localAgents: [], globalAgents: [] };
  /**
   * Mashed-ready skills and commands grouped by scope × kind. Fetched
   * in onMount and re-fetched whenever the repoPath changes so the
   * sidebar always reflects the current repo's local assets.
   * @type {{ localCommands: any[], globalCommands: any[], localSkills: any[], globalSkills: any[] }}
   */
  let groupedMashedAssets = { localCommands: [], globalCommands: [], localSkills: [], globalSkills: [] };
  let sprintStatus = null;
  let currentWorkflow = null;
  let workflowName = 'Untitled Workflow';
  let saving = false;

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
  let lastSavedSnapshot = snapshotCanvas([], [], 'Untitled Workflow');
  // True when the user explicitly chose "Discard" from NameWorkflowModal —
  // suppresses re-prompting on the next reactive pass while App.svelte
  // completes the view switch.
  let discardLatch = false;

  // Leave-intercept modal state
  let showNameModal = false;
  /** @type {null | ((decision: 'save' | 'discard' | 'cancel', name?: string) => void)} */
  let nameModalResolver = null;

  // Phase 2 sentinel toast: surfaced when a command node transitions to
  // `failed`. Phase 3 will replace the fail-fast with real execution.
  let failureToastMessage = '';

  // Dirty flag re-evaluates whenever nodes/edges/name change. Discard-latch
  // short-circuits to false so the view-switch triggered by a Discard click
  // never bounces back into the modal.
  $: isDirty =
    !discardLatch &&
    snapshotCanvas($nodes, $edges, workflowName) !== lastSavedSnapshot;

  /** True when the user has built something but never saved it at all. */
  $: isUnnamedDraft = !currentWorkflow?.id && $nodes.length > 0;

  // Model registry default (loaded at startup)
  let defaultModelId = '';

  // Execution state
  let executionId = null;
  let executionStatus = 'idle';
  let nodeProgress = { completed: 0, total: 0 };
  let execError = '';

  // Config panel state
  let selectedNode = null;
  let configPanelWidth = 280;

  // Agent modal state
  let showAgentModal = false;
  let editingAgent = null;

  // Terminal modal state
  let showTerminalModal = false;
  let terminalTarget = '';
  $: terminalRepoPath = repoPath;
  $: parsedTerminalTarget = parseFriendlyTarget(terminalTarget);

  // Output viewer modal state
  let showOutputModal = false;
  let outputModalContent = '';
  let outputModalLabel = '';
  let outputLoading = false;

  // Skill editor modal state
  let editingAsset = null;

  // Array editor modal state
  let showArrayModal = false;
  let arrayModalNodeId = '';
  let arrayModalItems = [];

  onMount(async () => {
    try {
      [processes, templates, savedWorkflows, groupedAgents, groupedMashedAssets] = await Promise.all([
        GetBmadProcesses(),
        ListBmadTemplates(),
        repoPath ? ListBmadWorkflowsByRepo(repoPath) : Promise.resolve([]),
        ListAllAgents(repoPath || ''),
        ListAllMashedAssets(repoPath || ''),
      ]);
      // Load default model from registry
      const models = await ListModels();
      const def = models.find(m => m.isDefault);
      defaultModelId = def ? def.id : (models[0]?.id || '');
    } catch (e) {
      console.error('Failed to load BMAD data:', e);
      processes = processes || [];
      templates = templates || [];
      savedWorkflows = savedWorkflows || [];
      groupedAgents = groupedAgents || { bmadAgents: [], localAgents: [], globalAgents: [] };
      groupedMashedAssets = groupedMashedAssets || { localCommands: [], globalCommands: [], localSkills: [], globalSkills: [] };
    }

    if (repoPath) {
      try {
        sprintStatus = await GetSprintStatus(repoPath);
      } catch (e) {
        console.warn('No sprint status for repo:', e);
      }
      // Restore any live execution for this repo so the canvas reflects
      // the running state immediately, without the user having to
      // re-pick the workflow from the sidebar.
      await restoreForRepo(repoPath);
    }

    // Dev-only test seams for the Playwright AC spec — drive the
    // builder directly without round-tripping through the filesystem
    // or save/open UI. `import.meta.env.DEV` is a Vite static literal
    // so the block dead-code-eliminates in production.
    if (import.meta.env.DEV && typeof window !== 'undefined') {
      window.__mashed_loadWorkflowFixture = (def) => {
        try {
          loadNodesEdges(def);
          return true;
        } catch (err) {
          console.error('Failed to load workflow fixture:', err);
          return false;
        }
      };
      window.__mashed_seedMashedAssets = (grouped) => {
        groupedMashedAssets = grouped || { localCommands: [], globalCommands: [], localSkills: [], globalSkills: [] };
        return true;
      };
      // skills-cmd-03 AC-3 test seam: Playwright needs a way to assert
      // that the frontend surfaces the Phase 2 sentinel failure without
      // having to round-trip through a real backend execution (which
      // requires a valid repoPath on disk). This seam invokes the same
      // code path the EventsOn('bmad:node:status') listener would run
      // when the executor emits NodeFailed for a command node.
      window.__mashed_simulateNodeFailure = (nodeId, message) => {
        if (!$nodes.some((n) => n.id === nodeId)) return false;
        $nodes = $nodes.map((n) =>
          n.id === nodeId
            ? { ...n, data: { ...n.data, status: 'failed' } }
            : n,
        );
        failureToastMessage = message || COMMAND_NODE_SENTINEL_DEFAULT_BODY;
        updateProgress();
        return true;
      };
    }
  });

  onDestroy(() => {
    if (import.meta.env.DEV && typeof window !== 'undefined') {
      delete window.__mashed_loadWorkflowFixture;
      delete window.__mashed_seedMashedAssets;
      delete window.__mashed_simulateNodeFailure;
    }
  });

  // Remember the last repo we restored so a reactive re-fire on
  // unchanged props does not re-hit the backend on every store update.
  let lastRestoredRepoPath = '';

  // React to repoPath changes (e.g. snackbar-driven navigation to a
  // different repo). `processes` must be loaded before restoreForRepo
  // runs because loadNodesEdges looks up ProcessDef by id for node
  // rendering. The guard waits until onMount's Promise.all has populated
  // `processes` at least once.
  $: if (repoPath && processes.length > 0 && repoPath !== lastRestoredRepoPath) {
    (async () => {
      lastRestoredRepoPath = repoPath;
      // Refresh per-repo sidebar data and sprint status when the repo
      // changes mid-session. groupedMashedAssets depends on the repo's
      // local .claude/{skills,commands}/ directories, so it refreshes
      // alongside savedWorkflows even though the global half doesn't
      // change between repos.
      try {
        savedWorkflows = await ListBmadWorkflowsByRepo(repoPath);
      } catch (e) {
        savedWorkflows = [];
      }
      try {
        groupedMashedAssets = await ListAllMashedAssets(repoPath);
      } catch (e) {
        groupedMashedAssets = { localCommands: [], globalCommands: [], localSkills: [], globalSkills: [] };
      }
      try {
        sprintStatus = await GetSprintStatus(repoPath);
      } catch (e) {
        sprintStatus = null;
      }
      await restoreForRepo(repoPath);
    })();
  }

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

      currentWorkflow = wf;
      workflowName = wf.name || 'Untitled Workflow';
      loadNodesEdges(wf);

      // Overlay live statuses from the running exec so the canvas shows
      // running / complete nodes without waiting for the next event.
      const byId = new Map((exec.nodes || []).map((n) => [n.id, n]));
      $nodes = $nodes.map((n) => {
        const live = byId.get(n.id);
        if (!live) return n;
        return {
          ...n,
          data: {
            ...n.data,
            status: live.status || n.data.status,
            tmuxTarget: live.tmuxTarget || n.data.tmuxTarget,
            storyId: live.storyId || n.data.storyId,
          },
        };
      });

      executionId = exec.id;
      executionStatus = exec.status || 'running';
      updateProgress();

      // Loaded + overlaid state IS the clean baseline — the user has
      // not edited anything, so tryLeave should not fire on the next
      // navigation.
      lastSavedSnapshot = snapshotCanvas($nodes, $edges, workflowName);
    } catch (e) {
      console.warn('restoreForRepo: failed', e);
    }
  }

  // Listen for live node status updates (scoped by execID)
  const cancelStatusListener = EventsOn('bmad:node:status', (event) => {
    if (!event?.nodeId || (executionId && event.execId !== executionId)) return;
    $nodes = $nodes.map(n => {
      if (n.id === event.nodeId) {
        return { ...n, data: {
          ...n.data,
          status: event.status,
          tmuxTarget: event.tmuxTarget || n.data.tmuxTarget,
          iterationCount: event.iteration || n.data.iterationCount,
        } };
      }
      return n;
    });
    const updatedNode = $nodes.find(n => n.id === event.nodeId);
    // Update selected node if it matches
    if (selectedNode && selectedNode.id === event.nodeId && updatedNode) {
      selectedNode = updatedNode;
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
        failureToastMessage = eventMessage || COMMAND_NODE_SENTINEL_DEFAULT_BODY;
      }
    }
    updateProgress();
  });

  const cancelArtifactListener = EventsOn('bmad:node:artifacts', (event) => {
    if (!event?.nodeId || (executionId && event.execId !== executionId)) return;
    $nodes = $nodes.map(n => {
      if (n.id === event.nodeId) {
        return { ...n, data: {
          ...n.data,
          artifactStatus: { found: event.found || [], missing: event.missing || [] },
        } };
      }
      return n;
    });
    if (selectedNode && selectedNode.id === event.nodeId) {
      selectedNode = $nodes.find(n => n.id === event.nodeId) || selectedNode;
    }
  });

  const cancelExecListener = EventsOn('bmad:execution:status', (event) => {
    if (executionId && event.execId !== executionId) return;
    if (event?.status) executionStatus = event.status;
  });

  let sprintRefreshTimer = null;
  const cancelSprintListener = EventsOn('bmad:sprint:updated', (event) => {
    // Update node storyStatus immediately
    if (event?.storyId) {
      $nodes = $nodes.map(n => {
        if (n.data?.storyId === event.storyId) {
          return { ...n, data: { ...n.data, storyStatus: event.status } };
        }
        return n;
      });
    }
    // Debounce full sprint data refresh (coalesces rapid node completions)
    clearTimeout(sprintRefreshTimer);
    sprintRefreshTimer = setTimeout(async () => {
      if (repoPath) {
        try { sprintStatus = await GetSprintStatus(repoPath); } catch {}
      }
    }, 400);
  });

  onDestroy(() => {
    if (cancelStatusListener) cancelStatusListener();
    if (cancelArtifactListener) cancelArtifactListener();
    if (cancelExecListener) cancelExecListener();
    if (cancelSprintListener) cancelSprintListener();
  });

  function updateProgress() {
    const total = $nodes.length;
    const completed = $nodes.filter(n => n.data.status === 'complete').length;
    nodeProgress = { completed, total };
  }

  function isValidConnection(connection) {
    if (connection.source === connection.target) return false;
    const exists = $edges.some(e =>
      e.source === connection.source && e.target === connection.target
    );
    return !exists;
  }

  function inferEdgeLabel(sourceHandle) {
    if (sourceHandle === 'true') return 'true';
    if (sourceHandle === 'false') return 'false';
    if (sourceHandle === 'loop-body') return 'body';
    if (sourceHandle === 'loop-exit') return 'exit';
    return '';
  }

  function onConnect(params) {
    const label = inferEdgeLabel(params.sourceHandle);
    const newEdge = {
      id: `edge-${Date.now()}`,
      source: params.source,
      target: params.target,
      sourceHandle: params.sourceHandle,
      targetHandle: params.targetHandle,
      label,
      data: { label },
    };
    $edges = [...$edges, newEdge];
  }

  function onDropProcess(processId, position) {
    const process = processes.find(p => p.id === processId);
    if (!process) return;

    const newNode = {
      id: `node-${Date.now()}`,
      type: 'bmadProcess',
      position,
      data: { label: process.name, processId: process.id, process, status: 'pending', config: {} },
    };
    $nodes = [...$nodes, newNode];
  }

  function onDropStory(storyData, position) {
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
    $nodes = [...$nodes, newNode];
  }

  const controlFlowNames = { condition: 'Condition', loop: 'Loop', loopUntil: 'Loop Until', transform: 'Transform', merge: 'Merge' };

  function onDropControlFlow(nodeType, position) {
    const newNode = {
      id: `cf-${Date.now()}`,
      type: nodeType,
      position,
      data: { nodeType, label: controlFlowNames[nodeType] || nodeType, status: 'pending', config: {} },
    };
    $nodes = [...$nodes, newNode];
  }

  function onNodeClick(detail) {
    const node = detail.node;
    if (node) selectedNode = node;
  }

  function onNodesDelete(deletedNodes) {
    // Filter out running nodes — they cannot be deleted
    const protectedIds = deletedNodes
      .filter(n => n.data?.status === 'running')
      .map(n => n.id);

    if (protectedIds.length > 0) {
      // Re-add protected nodes that xyflow already removed
      const protectedNodes = deletedNodes.filter(n => protectedIds.includes(n.id));
      $nodes = [...$nodes, ...protectedNodes];
    }

    // Clear selectedNode if it was deleted
    if (selectedNode && deletedNodes.some(n => n.id === selectedNode.id) && !protectedIds.includes(selectedNode.id)) {
      selectedNode = null;
    }
    updateProgress();
  }

  function onEdgesDelete(deletedEdges) {
    // xyflow handles store removal; no additional state cleanup needed
  }

  function onReconnect(detail) {
    const { oldEdge, newConnection } = detail;
    $edges = $edges.map(e => {
      if (e.id === oldEdge.id) {
        const label = inferEdgeLabel(newConnection.sourceHandle);
        return {
          ...e,
          source: newConnection.source,
          target: newConnection.target,
          sourceHandle: newConnection.sourceHandle,
          targetHandle: newConnection.targetHandle,
          label,
          data: { label },
        };
      }
      return e;
    });
  }

  function onSelectionChange(selection) {
    if (selection.nodes.length === 1) {
      selectedNode = selection.nodes[0];
    } else if (selection.nodes.length === 0) {
      selectedNode = null;
    }
  }

  function onAddTemplate(templateId, position, connectToNodeId) {
    const tpl = templates.find(t => t.id === templateId);
    if (!tpl || !tpl.nodes?.length) return;

    const ts = Date.now();
    const idMap = {};

    // Create new nodes offset from the click position
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
          process: processes.find(p => p.id === n.processId) || null,
          status: 'pending',
          config: n.config || {},
        },
      };
    });

    // Recreate template edges with new IDs
    const newEdges = (tpl.edges || []).map((e, i) => ({
      id: `tpl-edge-${ts}-${i}`,
      source: idMap[e.source] || e.source,
      target: idMap[e.target] || e.target,
    })).filter(e => e.source && e.target);

    $nodes = [...$nodes, ...newNodes];
    $edges = [...$edges, ...newEdges];

    // Connect the source node to the first template node
    if (connectToNodeId && newNodes.length > 0) {
      $edges = [...$edges, {
        id: `connect-${ts}`,
        source: connectToNodeId,
        target: newNodes[0].id,
      }];
    }

    updateProgress();
  }

  function onConfigUpdate(e) {
    const { nodeId, config } = e.detail;
    $nodes = $nodes.map(n => {
      if (n.id === nodeId) {
        return { ...n, data: { ...n.data, config } };
      }
      return n;
    });
  }

  async function handleAgentSave(e) {
    try {
      await SaveBmadAgent(e.detail);
      groupedAgents = await ListAllAgents(repoPath || '');
      showAgentModal = false;
      editingAgent = null;
    } catch (err) {
      console.error('Failed to save agent:', err);
    }
  }

  async function handleAgentDelete(e) {
    try {
      await DeleteBmadAgent(e.detail);
      groupedAgents = await ListAllAgents(repoPath || '');
      showAgentModal = false;
      editingAgent = null;
    } catch (err) {
      console.error('Failed to delete agent:', err);
    }
  }

  function handleEditItems(e) {
    const { nodeId, items } = e.detail;
    arrayModalNodeId = nodeId;
    arrayModalItems = items || [];
    showArrayModal = true;
  }

  function handleArraySave(e) {
    const savedItems = e.detail;
    $nodes = $nodes.map(n => {
      if (n.id === arrayModalNodeId) {
        const config = { ...n.data.config, items: savedItems.length > 0 ? JSON.stringify(savedItems) : '' };
        return { ...n, data: { ...n.data, config } };
      }
      return n;
    });
    if (selectedNode && selectedNode.id === arrayModalNodeId) {
      selectedNode = $nodes.find(n => n.id === arrayModalNodeId) || selectedNode;
    }
    showArrayModal = false;
  }

  async function onOpenTerminal(e) {
    terminalTarget = e.detail;
    showTerminalModal = true;
  }

  async function handleOpenOutput(e) {
    const nodeId = e.detail;
    const node = $nodes.find(n => n.id === nodeId);
    outputModalLabel = node?.data?.label || nodeId;
    outputLoading = true;
    showOutputModal = true;
    outputModalContent = '';

    try {
      outputModalContent = await GetNodeOutput(executionId, nodeId);
    } catch (err) {
      outputModalContent = 'Error loading output: ' + err;
    }
    outputLoading = false;
  }

  /**
   * Persist the current canvas as a BMAD workflow.
   *
   * Returns `true` when the canvas was successfully written to disk,
   * `false` on any backend failure. Callers that gate navigation on a
   * successful save (see tryLeave) MUST check this return value —
   * silently proceeding after a failed save would lose the user's
   * changes without any feedback, which is exactly the class of bug
   * the leave-intercept flow is meant to prevent.
   *
   * Ordering inside the try block matters: the dirty-detector baseline
   * is updated IMMEDIATELY after SaveBmadWorkflow resolves, BEFORE the
   * (cosmetic) sidebar-list refresh. Otherwise a transient failure in
   * ListBmadWorkflowsByRepo would leave lastSavedSnapshot stale and
   * the user would see the NameWorkflowModal pop up again on their
   * next Back click even though the save itself had succeeded.
   */
  async function saveWorkflow() {
    saving = true;
    try {
      // Construct a real bmad.WorkflowDef instance (not a plain object
      // literal) so TypeScript is satisfied with the Wails binding's
      // class-based parameter type — see the import comment above.
      // The inner nodes/edges arrays do NOT need to be wrapped: Wails'
      // generated constructor copies each field by key, so a structural
      // shape is enough for everything beneath the top-level object.
      const wf = new bmad.WorkflowDef({
        id: currentWorkflow?.id || `wf-${Date.now()}`,
        name: workflowName,
        description: '',
        repoPath: repoPath,
        nodes: canvasNodesToWorkflowNodes($nodes),
        edges: canvasEdgesToWorkflowEdges($edges),
        isTemplate: false,
        templateId: currentWorkflow?.templateId || '',
        createdAt: currentWorkflow?.createdAt || new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      });
      await SaveBmadWorkflow(wf);

      // SUCCESS PATH: commit all client-side "we are now clean" state
      // synchronously, before any further awaits that could fail. If
      // anything after this point throws, the canvas is still clean
      // on disk AND in memory, and isDirty stays false.
      currentWorkflow = wf;
      lastSavedSnapshot = snapshotCanvas($nodes, $edges, workflowName);

      // Sidebar refresh is cosmetic — a failure here must NOT roll
      // back the "save succeeded" signal. Scope it in its own try so
      // the outer catch only fires on an actual SaveBmadWorkflow fail.
      if (repoPath) {
        try {
          savedWorkflows = await ListBmadWorkflowsByRepo(repoPath);
        } catch (listErr) {
          console.warn('Failed to refresh saved workflows list:', listErr);
        }
      } else {
        savedWorkflows = [];
      }
      return true;
    } catch (e) {
      console.error('Failed to save workflow:', e);
      return false;
    } finally {
      saving = false;
    }
  }

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
  // Exposed via `export` so App.svelte can call it imperatively via
  // bind:this before triggering cross-view navigation (e.g. when a
  // snackbar click wants to switch repos).
  export async function tryLeave() {
    // Nothing to protect against.
    if (!isDirty || $nodes.length === 0) return true;

    // Case 2 — named workflow, silent save. If the save fails we MUST
    // refuse the leave; otherwise the user would silently lose their
    // edits when navigating back or switching repos.
    if (currentWorkflow?.id) {
      const ok = await saveWorkflow();
      return ok;
    }

    // Case 3 — unnamed draft. Prompt the user.
    return new Promise((resolve) => {
      nameModalResolver = async (decision, name) => {
        nameModalResolver = null;
        showNameModal = false;
        if (decision === 'cancel') {
          resolve(false);
          return;
        }
        if (decision === 'discard') {
          // Blank the canvas BEFORE resolving so the reactive `isDirty`
          // check on the next tick sees an empty canvas and no longer
          // triggers the modal if the consumer re-queries.
          discardLatch = true;
          $nodes = [];
          $edges = [];
          currentWorkflow = null;
          workflowName = 'Untitled Workflow';
          lastSavedSnapshot = snapshotCanvas([], [], 'Untitled Workflow');
          // Release the latch on the next microtask so subsequent user
          // edits are tracked again.
          queueMicrotask(() => { discardLatch = false; });
          resolve(true);
          return;
        }
        // decision === 'save' — commit the name the user typed and
        // attempt the write. On backend failure we resolve false so
        // the parent back-click is aborted and the modal stays closed
        // (the user can click Back again to retry). Future polish
        // could re-open the modal with an inline error instead.
        if (name && name.length > 0) workflowName = name;
        const saved = await saveWorkflow();
        resolve(saved);
      };
      showNameModal = true;
    });
  }

  function handleNameModalSave(e) {
    if (nameModalResolver) nameModalResolver('save', e.detail?.name);
  }
  function handleNameModalDiscard() {
    if (nameModalResolver) nameModalResolver('discard');
  }
  function handleNameModalCancel() {
    if (nameModalResolver) nameModalResolver('cancel');
  }

  async function handleBackRequest() {
    const ok = await tryLeave();
    if (ok) dispatch('back');
  }

  function loadNodesEdges(wf) {
    // skills-cmd-03 AC-1: the type-mapping ternary inside
    // workflowNodesToCanvasNodes routes `nodeType === 'command'` to
    // svelte-flow `type: 'command'` (rendered by CommandNode.svelte) while
    // preserving the legacy `bmadProcess` fallback for blank / `'process'`
    // nodeTypes. Unit-tested in workflowSerialisation.test.js.
    $nodes = workflowNodesToCanvasNodes(wf.nodes, processes);
    $edges = workflowEdgesToCanvasEdges(wf.edges, inferEdgeLabel);
    selectedNode = null;
    updateProgress();
    // Loading an existing workflow is a "clean" state by definition —
    // baseline the dirty detector so the next leave is free unless the
    // user starts editing.
    lastSavedSnapshot = snapshotCanvas($nodes, $edges, workflowName);
  }

  async function loadWorkflow(e) {
    const wfId = e.detail;
    try {
      const wf = await GetBmadWorkflow(wfId);
      currentWorkflow = wf;
      workflowName = wf.name;
      loadNodesEdges(wf);
    } catch (err) {
      console.error('Failed to load workflow:', err);
    }
  }

  async function useTemplate(e) {
    const templateId = e.detail;
    // Templates CreateFromTemplate returns a brand-new saved workflow
    // rather than a local draft, so calling loadNodesEdges below already
    // baselines the dirty detector. The currentWorkflow assignment must
    // happen BEFORE loadNodesEdges so workflowName is correct when the
    // snapshot is captured.
    try {
      const wf = await CreateFromTemplate(templateId, repoPath);
      currentWorkflow = wf;
      workflowName = wf.name;
      loadNodesEdges(wf);
    } catch (err) {
      console.error('Failed to create from template:', err);
    }
  }

  async function deleteWorkflow(e) {
    const wfId = e.detail;
    try {
      await DeleteBmadWorkflow(wfId);
      savedWorkflows = repoPath ? await ListBmadWorkflowsByRepo(repoPath) : [];
      if (currentWorkflow?.id === wfId) {
        currentWorkflow = null;
        workflowName = 'Untitled Workflow';
        $nodes = [];
        $edges = [];
        selectedNode = null;
        lastSavedSnapshot = snapshotCanvas([], [], 'Untitled Workflow');
      }
    } catch (err) {
      console.error('Failed to delete workflow:', err);
    }
  }

  function newWorkflow() {
    currentWorkflow = null;
    workflowName = 'Untitled Workflow';
    executionId = null;
    executionStatus = 'idle';
    nodeProgress = { completed: 0, total: 0 };
    execError = '';
    selectedNode = null;
    showTerminalModal = false;
    $nodes = [];
    $edges = [];
    // A fresh canvas is a clean baseline — without this reset, the
    // dirty detector would still compare against whatever was previously
    // saved and treat the blank canvas as "changes to discard".
    lastSavedSnapshot = snapshotCanvas([], [], 'Untitled Workflow');
  }

  async function handleExecStart(e) {
    const { model } = e.detail;
    execError = '';
    // Clear any stale Phase 2 sentinel toast from a previous run attempt
    // so the user sees a fresh state when they click Run again.
    failureToastMessage = '';

    // Auto-save before executing. The old code inferred failure from
    // the absence of currentWorkflow.id after the call; now that
    // saveWorkflow returns a boolean we can check it directly and give
    // the user a precise reason for the abort.
    if (!currentWorkflow?.id) {
      const saved = await saveWorkflow();
      if (!saved) {
        execError = 'Save failed — check the logs before running.';
        return;
      }
    }
    if (!currentWorkflow?.id) {
      execError = 'Save the workflow before running.';
      return;
    }

    // Reset all node statuses to pending for re-execution
    $nodes = $nodes.map(n => ({
      ...n,
      data: { ...n.data, status: 'pending', tmuxTarget: '' },
    }));
    nodeProgress = { completed: 0, total: $nodes.length };

    try {
      executionId = await StartBmadWorkflow(currentWorkflow.id, repoPath, model || defaultModelId);
      executionStatus = 'running';
    } catch (err) {
      execError = String(err);
      executionStatus = 'idle';
    }
  }

  async function handleExecPause() {
    if (!executionId) return;
    try {
      await PauseBmadWorkflow(executionId);
    } catch (err) {
      execError = String(err);
    }
  }

  async function handleExecResume() {
    if (!executionId) return;
    execError = '';
    try {
      await ResumeBmadWorkflow(executionId);
    } catch (err) {
      execError = String(err);
    }
  }

  async function handleExecStop() {
    if (!executionId) return;
    try {
      await StopBmadWorkflow(executionId);
    } catch (err) {
      execError = String(err);
    }
  }
</script>

<div class="workflow-builder">
  <ProcessSidebar
    {processes}
    {templates}
    {savedWorkflows}
    {sprintStatus}
    {repoPath}
    {repoBranch}
    {groupedMashedAssets}
    on:use-template={useTemplate}
    on:load-workflow={loadWorkflow}
    on:delete-workflow={deleteWorkflow}
    on:create-custom-template={newWorkflow}
    on:branch-changed={(e) => { if (e.detail?.branch) repoBranch = e.detail.branch; }}
    on:editAsset={(e) => { editingAsset = e.detail; }}
  />

  <div class="canvas-area">
    <RepoContextBar {repoPath} {repoBranch} {sprintStatus} on:back={handleBackRequest} />
    <div class="toolbar">
      <span class="toolbar-label">File Name:</span>
      <input
        class="workflow-name-input"
        type="text"
        bind:value={workflowName}
        placeholder="Workflow name..."
      />
      <button class="toolbar-btn new-btn" on:click={newWorkflow} title="New workflow">New</button>
      <button class="toolbar-btn save-btn" on:click={saveWorkflow} disabled={saving} title="Save workflow">
        <Save size={13} />
        {saving ? 'Saving...' : 'Save'}
      </button>
      <button class="toolbar-btn agents-btn" on:click={() => { editingAgent = null; showAgentModal = true; }} title="Manage agents">
        Agents
      </button>

      <div class="toolbar-sep" />

      <ExecutionBar
        {executionStatus}
        {nodeProgress}
        {repoPath}
        on:start={handleExecStart}
        on:pause={handleExecPause}
        on:resume={handleExecResume}
        on:stop={handleExecStop}
      />
    </div>

    <div class="canvas-with-panel">
      <SvelteFlowProvider>
        <CanvasPane
          {nodes}
          {edges}
          {nodeTypes}
          {isValidConnection}
          {onConnect}
          {onDropProcess}
          {onDropStory}
          {onDropControlFlow}
          {onNodeClick}
          onPaneClick={() => selectedNode = null}
          {onNodesDelete}
          {onEdgesDelete}
          {onSelectionChange}
          {onReconnect}
          {onAddTemplate}
          {templates}
          {executionStatus}
          configPanelOpen={!!selectedNode}
          {configPanelWidth}
        >
          <div slot="empty-hint">
            {#if $nodes.length === 0 && !currentWorkflow}
              <div class="empty-hint">Drag processes from the sidebar or select a template</div>
            {/if}
          </div>
        </CanvasPane>
      </SvelteFlowProvider>

      <NodeConfigPanel
        node={selectedNode}
        {groupedAgents}
        bind:panelWidth={configPanelWidth}
        on:update={onConfigUpdate}
        on:close={() => selectedNode = null}
        on:open-terminal={onOpenTerminal}
        on:open-output={handleOpenOutput}
        on:edit-items={handleEditItems}
      />

      {#if failureToastMessage}
        <CanvasFailureToast
          message={failureToastMessage}
          onDismiss={() => (failureToastMessage = '')}
        />
      {/if}
    </div>

    {#if execError}
      <div class="exec-error">
        <span class="exec-error-text">{execError}</span>
        <button class="exec-error-dismiss" on:click={() => execError = ''}>Dismiss</button>
      </div>
    {/if}

  </div>

  {#if showTerminalModal}
    <div class="terminal-overlay" role="presentation" on:click={() => showTerminalModal = false} on:keydown={() => {}}>
      <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
      <div class="terminal-modal" role="dialog" on:click|stopPropagation on:keydown={() => {}}>
        <div class="terminal-modal-header">
          <span class="terminal-modal-title" title={terminalTarget}>
            {#if parsedTerminalTarget}
              <span class="tmt-prefix">Terminal</span>
              <span class="tmt-dash">—</span>
              <span class="tmt-repo">{parsedTerminalTarget.repo}</span>
              <span class="tmt-sep">·</span>
              <span class="tmt-branch">{parsedTerminalTarget.branch}</span>
              <span class="tmt-sep">·</span>
              <span class="tmt-label">{parsedTerminalTarget.label}</span>
            {:else}
              {terminalTarget}
            {/if}
          </span>
          <button class="terminal-modal-close" on:click={() => showTerminalModal = false}>&times;</button>
        </div>
        <div class="terminal-modal-body">
          {#key terminalTarget}
            <Terminal paneTarget={terminalTarget} repoPath={terminalRepoPath} />
          {/key}
        </div>
      </div>
    </div>
  {/if}

  {#if showAgentModal}
    <AgentConfigModal
      agent={editingAgent}
      {processes}
      on:save={handleAgentSave}
      on:delete={handleAgentDelete}
      on:close={() => { showAgentModal = false; editingAgent = null; }}
    />
  {/if}

  <OutputViewerModal
    show={showOutputModal}
    label={outputModalLabel}
    content={outputModalContent}
    loading={outputLoading}
    on:close={() => showOutputModal = false}
  />

  {#if showArrayModal}
    <ArrayEditorModal
      items={arrayModalItems}
      on:save={handleArraySave}
      on:close={() => showArrayModal = false}
    />
  {/if}

  {#if showQuestionModal && activeQuestion}
    <QuestionResponseModal
      question={activeQuestion}
      on:responded={handleQuestionResponded}
      on:close={handleQuestionClose}
    />
  {/if}

  {#if showNameModal}
    <NameWorkflowModal
      defaultName={workflowName && workflowName !== 'Untitled Workflow' ? workflowName : ''}
      nodeCount={$nodes.length}
      on:save={handleNameModalSave}
      on:discard={handleNameModalDiscard}
      on:cancel={handleNameModalCancel}
    />
  {/if}

  {#if editingAsset}
    <SkillEditorModal
      asset={editingAsset}
      on:close={() => { editingAsset = null; }}
      on:save={() => { editingAsset = null; }}
    />
  {/if}
</div>

<style>
  .workflow-builder {
    display: flex;
    flex: 1;
    overflow: hidden;
    background: var(--bg-deepest);
  }

  .canvas-area {
    flex: 1;
    display: flex;
    flex-direction: column;
    min-width: 0;
  }

  .toolbar {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: var(--sp-xs) 10px;
    background: var(--bg-surface);
    border-bottom: 1px solid var(--border-subtle);
    flex-shrink: 0;
  }

  .toolbar-btn {
    display: flex;
    align-items: center;
    gap: var(--sp-xs);
    padding: 4px 10px;
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-sm);
    color: var(--text-primary);
    font-family: var(--font-mono);
    font-size: var(--text-body);
    font-weight: 500;
    cursor: pointer;
    transition: background 100ms ease, color 100ms ease, border-color 100ms ease;
    flex-shrink: 0;
  }

  .toolbar-btn:hover {
    background: var(--bg-active);
    color: var(--text-primary);
    border-color: var(--border-emphasis);
  }

  .toolbar-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .save-btn:hover { border-color: var(--accent-green); color: var(--accent-green); }
  .agents-btn:hover { border-color: var(--accent-purple, #9d6fff); color: var(--accent-purple, #9d6fff); }

  .toolbar-label {
    font-family: var(--font-mono);
    font-size: var(--text-body);
    font-weight: 500;
    color: var(--text-primary);
    flex-shrink: 0;
    white-space: nowrap;
  }

  .toolbar-sep {
    width: 1px;
    height: 18px;
    background: var(--border-subtle);
    flex-shrink: 0;
  }

  .workflow-name-input {
    flex: 1;
    min-width: 120px;
    padding: 4px 8px;
    background: var(--bg-deepest);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-sm);
    color: var(--text-primary);
    font-family: var(--font-mono);
    font-size: var(--text-body);
    outline: none;
    transition: border-color 100ms ease;
  }

  .workflow-name-input:focus { border-color: var(--accent-green); }

  .canvas-with-panel {
    flex: 1;
    position: relative;
    overflow: hidden;
  }

  .empty-hint {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    font-family: var(--font-mono);
    font-size: 13px;
    color: var(--text-muted);
    pointer-events: none;
    z-index: 1;
    text-align: center;
  }

  /* Error banner */
  .exec-error {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: var(--sp-xs) 10px;
    background: color-mix(in srgb, var(--accent-red) 10%, transparent);
    border-top: 1px solid var(--accent-red, #f85149);
    flex-shrink: 0;
  }

  .exec-error-text {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--accent-red, #f85149);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .exec-error-dismiss {
    background: none;
    border: 1px solid var(--accent-red, #f85149);
    border-radius: var(--radius-sm);
    color: var(--accent-red, #f85149);
    font-family: var(--font-mono);
    font-size: var(--text-label);
    padding: var(--sp-2xs) var(--sp-xs);
    cursor: pointer;
    flex-shrink: 0;
  }

  .exec-error-dismiss:hover {
    background: color-mix(in srgb, var(--accent-red) 15%, transparent);
  }

  /* Terminal modal */
  .terminal-overlay {
    position: fixed;
    inset: 0;
    background: var(--overlay-backdrop);
    z-index: 100;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .terminal-modal {
    width: 80vw;
    height: 70vh;
    background: var(--bg-deepest);
    border: 1px solid var(--border-emphasis);
    border-radius: var(--radius-lg, 8px);
    display: flex;
    flex-direction: column;
    overflow: hidden;
    box-shadow: 0 16px 48px rgba(0, 0, 0, 0.5);
  }

  .terminal-modal-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 8px 12px;
    background: var(--bg-surface);
    border-bottom: 1px solid var(--border-subtle);
    flex-shrink: 0;
  }

  .terminal-modal-title {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--text-dim);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .tmt-prefix {
    color: var(--text-muted);
    font-weight: 400;
    letter-spacing: 0.04em;
    text-transform: uppercase;
  }
  .tmt-dash {
    color: var(--text-muted);
    margin: 0 var(--sp-2xs);
  }
  .tmt-repo {
    color: var(--text-primary);
    font-weight: 500;
  }
  .tmt-branch {
    color: var(--text-dim);
    font-weight: 400;
  }
  .tmt-label {
    color: var(--accent-green);
    font-weight: 500;
  }
  .tmt-sep {
    color: var(--text-muted);
    margin: 0 var(--sp-xs);
  }

  .terminal-modal-close {
    background: none;
    border: none;
    color: var(--text-muted);
    font-size: var(--text-section);
    cursor: pointer;
    padding: 0 4px;
    line-height: 1;
  }

  .terminal-modal-close:hover { color: var(--text-primary); }

  .terminal-modal-body {
    flex: 1;
    overflow: hidden;
  }
</style>
