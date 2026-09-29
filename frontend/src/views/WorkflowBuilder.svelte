<script>
  import { SvelteFlowProvider } from '@xyflow/svelte';
  import '@xyflow/svelte/dist/style.css';
  import { writable } from 'svelte/store';
  import { createEventDispatcher, onMount, onDestroy } from 'svelte';
  import { fly } from 'svelte/transition';
  import { GetBmadProcesses, ListBmadTemplates, ListBmadWorkflowsByRepo,
           SaveBmadWorkflow, GetBmadWorkflow, CreateFromTemplate,
           DeleteBmadWorkflow, ListBmadAgents, ListAllAgents, SaveBmadAgent, DeleteBmadAgent,
           StartBmadWorkflow, PauseBmadWorkflow, ResumeBmadWorkflow, StopBmadWorkflow,
           GetTerminalPort, GetSprintStatus, ListModels,
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
  import ProcessSidebar from '../components/bmad/ProcessSidebar.svelte';
  import CanvasPane from '../components/bmad/CanvasPane.svelte';
  import ProcessNode from '../components/bmad/ProcessNode.svelte';
  import CommandNode from '../components/bmad/CommandNode.svelte';
  import ConditionNode from '../components/bmad/ConditionNode.svelte';
  import LoopNode from '../components/bmad/LoopNode.svelte';
  import LoopUntilNode from '../components/bmad/LoopUntilNode.svelte';
  import TransformNode from '../components/bmad/TransformNode.svelte';
  import MergeNode from '../components/bmad/MergeNode.svelte';
  import MultiFileLoaderNode from '../components/bmad/MultiFileLoaderNode.svelte';
  import BuilderToolbar from '../components/bmad/BuilderToolbar.svelte';
  import TerminalModal from '../components/bmad/TerminalModal.svelte';
  import NodeConfigPanel from '../components/bmad/NodeConfigPanel.svelte';
  import AgentConfigModal from '../components/bmad/AgentConfigModal.svelte';
  import OutputViewerModal from '../components/bmad/OutputViewerModal.svelte';
  import ArrayEditorModal from '../components/bmad/ArrayEditorModal.svelte';
  import QuestionResponseModal from '../components/bmad/QuestionResponseModal.svelte';
  import InputResponseModal from '../components/bmad/InputResponseModal.svelte';
  import NodeInputSnackbarStack from '../components/bmad/NodeInputSnackbarStack.svelte';
  import NameWorkflowModal from '../components/bmad/NameWorkflowModal.svelte';
  import {
    interactiveInput,
    openModal as openInteractiveModal,
    closeModal as closeInteractiveModal,
    openModalForNode,
    dismissNode,
  } from '../stores/interactiveInput';
  import SkillEditorModal from '../components/bmad/SkillEditorModal.svelte';
  import RepoContextBar from '../components/bmad/RepoContextBar.svelte';
  import CanvasFailureToast from '../components/bmad/CanvasFailureToast.svelte';
  import {
    snapshotCanvas,
    canvasNodesToWorkflowNodes,
    canvasEdgesToWorkflowEdges,
    workflowNodesToCanvasNodes,
    workflowEdgesToCanvasEdges,
    COMMAND_NODE_SENTINEL_DEFAULT_BODY,
  } from '../lib/workflowSerialisation';
  // Extracted logic (spec R31): canvas handlers, execution restore + event
  // wiring, and the leave intercept live in plain modules that mutate this
  // component's state through the accessor context built below.
  import { createCanvasHandlers } from '../lib/workflowBuilder/canvasHandlers';
  import { createExecEvents } from '../lib/workflowBuilder/execEvents';
  import { createLeaveIntercept, computeIsDirty } from '../lib/workflowBuilder/leaveIntercept';

  // ── Type aliases: reusable JSDoc shortcuts to keep the annotations
  // below terse and to centralise the re-export surface so a future
  // rename only needs to touch this block.
  /** @typedef {import('../types/workflow').CanvasNode} CanvasNode */
  /** @typedef {import('../types/workflow').CanvasEdge} CanvasEdge */
  /** @typedef {import('../lib/types/wails').Workflow} Workflow */
  /** @typedef {import('../lib/types/wails').ProcessDef} ProcessDef */
  /** @typedef {import('../lib/types/wails').BmadAgentConfig} BmadAgentConfig */
  /** @typedef {import('../lib/types/wails').GroupedAgents} GroupedAgents */
  /** @typedef {import('../lib/types/wails').GroupedMashedAssets} GroupedMashedAssets */
  /** @typedef {import('../lib/types/wails').MashedAssetInfo} MashedAssetInfo */
  /** @typedef {import('../lib/types/wails').SprintStatus} SprintStatus */
  /** @typedef {import('../lib/types/wails').SprintStory} SprintStory */

  /** @type {string} */
  export let repoPath = '';
  /** @type {string} */
  export let repoBranch = '';
  /** @type {import('../components/bmad/questionSnackbarUtils').QuestionEventLike | null} */
  export let pendingQuestion = null;
  /** @type {string} */
  export let pendingTmuxTarget = '';

  /** @type {import('svelte').EventDispatcher<{ 'question-responded': { nodeId: string | undefined }; back: void; 'tmux-opened': void }>} */
  const dispatch = createEventDispatcher();

  let showQuestionModal = false;
  /** @type {import('../components/bmad/questionSnackbarUtils').QuestionEventLike | null} */
  let activeQuestion = null;

  // Only open the modal on a truthy transition when not already showing.
  // Without the !showQuestionModal guard, re-setting pendingQuestion to the
  // same value (e.g. parent re-render) would silently reset activeQuestion
  // while a submit may be in-flight.
  $: if (pendingQuestion && !showQuestionModal) {
    activeQuestion = pendingQuestion;
    showQuestionModal = true;
  }

  // Idle snackbar click arrives as pendingTmuxTarget — open the terminal
  // modal for that tmux session and immediately clear the prop so the
  // reactive block cannot re-fire on unrelated parent re-renders.
  $: if (pendingTmuxTarget) {
    terminalTarget = pendingTmuxTarget;
    showTerminalModal = true;
    dispatch('tmux-opened');
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
    multiFileLoader: MultiFileLoaderNode,
  };

  /** @type {import('svelte/store').Writable<CanvasNode[]>} */
  const nodes = writable(/** @type {CanvasNode[]} */ ([]));
  /** @type {import('svelte/store').Writable<CanvasEdge[]>} */
  const edges = writable(/** @type {CanvasEdge[]} */ ([]));

  /** @type {ProcessDef[]} */
  let processes = [];
  /** @type {Workflow[]} */
  let templates = [];
  /** @type {Workflow[]} */
  let savedWorkflows = [];
  /** @type {GroupedAgents | { bmadAgents: BmadAgentConfig[]; localAgents: import('../lib/types/wails').AgentInfo[]; globalAgents: import('../lib/types/wails').AgentInfo[] }} */
  let groupedAgents = { bmadAgents: [], localAgents: [], globalAgents: [] };
  /**
   * Mashed-ready skills and commands grouped by scope × kind. Fetched
   * in onMount and re-fetched whenever the repoPath changes so the
   * sidebar always reflects the current repo's local assets.
   * @type {GroupedMashedAssets | { localCommands: MashedAssetInfo[]; globalCommands: MashedAssetInfo[]; localSkills: MashedAssetInfo[]; globalSkills: MashedAssetInfo[] }}
   */
  let groupedMashedAssets = { localCommands: [], globalCommands: [], localSkills: [], globalSkills: [] };
  /** @type {SprintStatus | null} */
  let sprintStatus = null;
  /** @type {Workflow | null} */
  let currentWorkflow = null;
  let workflowName = 'Untitled Workflow';
  let saving = false;

  // ── Unsaved-changes tracking (see lib/workflowBuilder/leaveIntercept.js) ──
  // lastSavedSnapshot is the canvas shape at the most recent successful
  // save (or load); `isDirty` compares a fresh snapshot against it.
  let lastSavedSnapshot = snapshotCanvas([], [], 'Untitled Workflow');
  // True when the user explicitly chose "Discard" from NameWorkflowModal —
  // suppresses re-prompting on the next reactive pass while App.svelte
  // completes the view switch.
  let discardLatch = false;

  // Leave-intercept modal state
  let showNameModal = false;

  // Phase 2 sentinel toast: surfaced when a command node transitions to
  // `failed`. Phase 3 will replace the fail-fast with real execution.
  let failureToastMessage = '';

  // Dirty flag re-evaluates whenever nodes/edges/name change. Discard-latch
  // short-circuits to false so the view-switch triggered by a Discard click
  // never bounces back into the modal.
  $: isDirty = computeIsDirty(discardLatch, $nodes, $edges, workflowName, lastSavedSnapshot);

  /** True when the user has built something but never saved it at all. */
  $: isUnnamedDraft = !currentWorkflow?.id && $nodes.length > 0;

  // Model registry default (loaded at startup)
  let defaultModelId = '';

  // Execution state
  /** @type {string | null} */
  let executionId = null;
  let executionStatus = 'idle';
  let nodeProgress = { completed: 0, total: 0 };
  let execError = '';

  // Config panel state
  /** @type {CanvasNode | null} */
  let selectedNode = null;
  let configPanelWidth = 360;

  // Agent modal state
  let showAgentModal = false;
  /** @type {BmadAgentConfig | null} */
  let editingAgent = null;

  // Terminal modal state
  let showTerminalModal = false;
  let terminalTarget = '';

  // Output viewer modal state
  let showOutputModal = false;
  let outputModalContent = '';
  let outputModalLabel = '';
  let outputLoading = false;

  // Skill editor modal state
  /** @type {MashedAssetInfo | null} */
  let editingAsset = null;
  let lastSavedPath = '';
  let showSaveToast = false;
  let saveToastPath = '';

  // Array editor modal state
  let showArrayModal = false;
  let arrayModalNodeId = '';
  /** @type {string[]} */
  let arrayModalItems = [];

  // Live sidebar reload state (skills-watch-02) — driven by
  // refetchMashedAssets in lib/workflowBuilder/execEvents.js.
  let assetsFetching = false;
  let assetsError = false;
  /** @type {Record<string, 'created' | 'updated'>} */
  let flashedPaths = {};

  // Accessor context shared by the extracted modules. Every setter is a
  // plain assignment inside this component, so Svelte invalidation (and
  // therefore re-render ordering) is identical to the inline code.
  const ctx = {
    nodes,
    edges,
    get repoPath() { return repoPath; },
    get processes() { return processes; },
    get templates() { return templates; },
    get isDirty() { return isDirty; },
    get sprintStatus() { return sprintStatus; }, set sprintStatus(v) { sprintStatus = v; },
    get currentWorkflow() { return currentWorkflow; }, set currentWorkflow(v) { currentWorkflow = v; },
    get workflowName() { return workflowName; }, set workflowName(v) { workflowName = v; },
    get lastSavedSnapshot() { return lastSavedSnapshot; }, set lastSavedSnapshot(v) { lastSavedSnapshot = v; },
    get discardLatch() { return discardLatch; }, set discardLatch(v) { discardLatch = v; },
    get showNameModal() { return showNameModal; }, set showNameModal(v) { showNameModal = v; },
    get failureToastMessage() { return failureToastMessage; }, set failureToastMessage(v) { failureToastMessage = v; },
    get executionId() { return executionId; }, set executionId(v) { executionId = v; },
    get executionStatus() { return executionStatus; }, set executionStatus(v) { executionStatus = v; },
    get selectedNode() { return selectedNode; }, set selectedNode(v) { selectedNode = v; },
    get terminalTarget() { return terminalTarget; }, set terminalTarget(v) { terminalTarget = v; },
    get arrayModalNodeId() { return arrayModalNodeId; }, set arrayModalNodeId(v) { arrayModalNodeId = v; },
    get arrayModalItems() { return arrayModalItems; }, set arrayModalItems(v) { arrayModalItems = v; },
    get showArrayModal() { return showArrayModal; }, set showArrayModal(v) { showArrayModal = v; },
    get showOutputModal() { return showOutputModal; }, set showOutputModal(v) { showOutputModal = v; },
    get outputModalContent() { return outputModalContent; }, set outputModalContent(v) { outputModalContent = v; },
    get outputModalLabel() { return outputModalLabel; }, set outputModalLabel(v) { outputModalLabel = v; },
    get outputLoading() { return outputLoading; }, set outputLoading(v) { outputLoading = v; },
    get editingAsset() { return editingAsset; }, set editingAsset(v) { editingAsset = v; },
    get lastSavedPath() { return lastSavedPath; }, set lastSavedPath(v) { lastSavedPath = v; },
    get saveToastPath() { return saveToastPath; }, set saveToastPath(v) { saveToastPath = v; },
    get showSaveToast() { return showSaveToast; }, set showSaveToast(v) { showSaveToast = v; },
    get showTerminalModal() { return showTerminalModal; }, set showTerminalModal(v) { showTerminalModal = v; },
    get groupedMashedAssets() { return groupedMashedAssets; }, set groupedMashedAssets(v) { groupedMashedAssets = v; },
    get assetsFetching() { return assetsFetching; }, set assetsFetching(v) { assetsFetching = v; },
    get assetsError() { return assetsError; }, set assetsError(v) { assetsError = v; },
    get flashedPaths() { return flashedPaths; }, set flashedPaths(v) { flashedPaths = v; },
    loadNodesEdges,
    updateProgress,
    saveWorkflow,
    onLeave: () => dispatch('back'),
  };

  const {
    isValidConnection, inferEdgeLabel, onConnect, onDropProcess, onDropStory,
    onDropControlFlow, onNodeClick, onNodesDelete, onEdgesDelete, onReconnect,
    onSelectionChange, onAddTemplate, onConfigUpdate,
    handleEditItems, handleArraySave, onOpenTerminal, handleOpenOutput,
  } = createCanvasHandlers(ctx);
  // Registers every EventsOn listener now (component init), exactly as the
  // inline listeners did; torn down via execEvents.destroy() in onDestroy.
  const execEvents = createExecEvents(ctx);
  const { restoreForRepo, refetchMashedAssets, handleInputSkip } = execEvents;
  const leave = createLeaveIntercept(ctx);
  const {
    handleNameModalSave, handleNameModalDiscard, handleNameModalCancel, handleBackRequest, handleAssetSaved,
  } = leave;

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
      window.__mashed_simulateAssetsChanged = () => {
        refetchMashedAssets();
        return true;
      };
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
      delete window.__mashed_simulateAssetsChanged;
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

  // Sync rounds store -> node data so ProcessNode re-renders round counter.
  $: if ($interactiveInput.rounds) {
    const patches = $interactiveInput.rounds;
    let changed = false;
    const next = $nodes.map((n) => {
      const r = patches[n.id];
      if (typeof r === 'number' && n.data?.nodeRound !== r) {
        changed = true;
        return { ...n, data: { ...n.data, nodeRound: r } };
      }
      return n;
    });
    if (changed) $nodes = next;
  }

  onDestroy(() => {
    execEvents.destroy();
  });

  function updateProgress() {
    const total = $nodes.length;
    const completed = $nodes.filter(n => n.data?.status === 'complete').length;
    nodeProgress = { completed, total };
  }

  /** @param {CustomEvent<BmadAgentConfig>} e */
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

  /** @param {CustomEvent<string>} e */
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

  // Leave intercept — see lib/workflowBuilder/leaveIntercept.js for the
  // decision tree. Exposed via `export` so App.svelte can call it
  // imperatively via bind:this before triggering cross-view navigation
  // (e.g. when a snackbar click wants to switch repos).
  export function tryLeave() {
    return leave.tryLeave();
  }

  /** @param {Workflow} wf */
  function loadNodesEdges(wf) {
    // skills-cmd-03 AC-1: the type-mapping ternary inside
    // workflowNodesToCanvasNodes routes `nodeType === 'command'` to
    // svelte-flow `type: 'command'` (rendered by CommandNode.svelte) while
    // preserving the legacy `bmadProcess` fallback for blank / `'process'`
    // nodeTypes. Unit-tested in workflowSerialisation.test.ts.
    $nodes = workflowNodesToCanvasNodes(wf.nodes, processes);
    $edges = workflowEdgesToCanvasEdges(wf.edges, inferEdgeLabel);
    selectedNode = null;
    updateProgress();
    // Loading an existing workflow is a "clean" state by definition —
    // baseline the dirty detector so the next leave is free unless the
    // user starts editing.
    lastSavedSnapshot = snapshotCanvas($nodes, $edges, workflowName);
  }

  /** @param {CustomEvent<string>} e */
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

  /** @param {CustomEvent<string>} e */
  async function useTemplate(e) {
    const templateId = e.detail;
    // Templates CreateFromTemplate returns a brand-new saved workflow
    // rather than a local draft, so calling loadNodesEdges below already
    // baselines the dirty detector. The currentWorkflow assignment must
    // happen BEFORE loadNodesEdges so workflowName is correct when the
    // snapshot is captured.
    try {
      const wf = await CreateFromTemplate(repoPath, templateId);
      currentWorkflow = wf;
      workflowName = wf.name;
      loadNodesEdges(wf);
    } catch (err) {
      console.error('Failed to create from template:', err);
    }
  }

  /** @param {CustomEvent<string>} e */
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

  /** @param {CustomEvent<{ model?: string }>} e */
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
      executionId = await StartBmadWorkflow(repoPath, currentWorkflow.id, model || defaultModelId);
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
    {assetsFetching}
    {assetsError}
    {flashedPaths}
    on:use-template={useTemplate}
    on:load-workflow={loadWorkflow}
    on:delete-workflow={deleteWorkflow}
    on:create-custom-template={newWorkflow}
    on:branch-changed={(e) => { if (e.detail?.branch) repoBranch = e.detail.branch; }}
    on:editAsset={(e) => { editingAsset = e.detail; }}
  />

  <div class="canvas-area">
    <RepoContextBar {repoPath} {repoBranch} {sprintStatus} on:back={handleBackRequest} />
    <BuilderToolbar
      bind:workflowName
      {saving}
      {executionStatus}
      {nodeProgress}
      {repoPath}
      onNew={newWorkflow}
      onSave={saveWorkflow}
      onAgents={() => { editingAgent = null; showAgentModal = true; }}
      on:start={handleExecStart}
      on:pause={handleExecPause}
      on:resume={handleExecResume}
      on:stop={handleExecStop}
    />

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
        on:open-prompt={(e) => openModalForNode(e.detail)}
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
    <TerminalModal
      {terminalTarget}
      {repoPath}
      onClose={() => showTerminalModal = false}
    />
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

  <!-- S6: Interactive input snackbar + modal. Snackbar surfaces all
       PendingPrompts in the store; clicking a Respond button opens the modal. -->
  <NodeInputSnackbarStack
    pendingPrompts={$interactiveInput.pendingPrompts}
    on:respond={(e) => openInteractiveModal(e.detail.prompt)}
    on:skip={(e) => handleInputSkip(e.detail.prompt)}
    on:dismiss={(e) => {
      const nodeId = e.detail?.entry?.nodeId;
      if (nodeId) dismissNode(nodeId);
    }}
  />

  {#if $interactiveInput.activeModal}
    <InputResponseModal
      prompt={$interactiveInput.activeModal}
      execId={executionId || $interactiveInput.activeModal.execId || ''}
      repoName={repoPath ? repoPath.split('/').pop() : ''}
      on:close={closeInteractiveModal}
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
      on:save={handleAssetSaved}
    />
  {/if}

  {#if showSaveToast}
    <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
    <div class="save-toast" on:click={() => { showSaveToast = false; }} transition:fly={{ y: -8, duration: 150 }}>
      <svg class="toast-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"></polyline></svg>
      <div class="toast-body">
        <span class="toast-text">Asset saved</span>
        <span class="toast-path">{saveToastPath}</span>
      </div>
    </div>
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

  /* ── Save toast ── */
  .save-toast {
    position: fixed;
    top: var(--sp-lg);
    right: var(--sp-lg);
    padding: var(--sp-sm) var(--sp-md);
    background: var(--bg-elevated);
    border: 1px solid color-mix(in srgb, var(--accent-green) 40%, transparent);
    border-radius: var(--radius-md);
    box-shadow: 0 12px 40px color-mix(in srgb, var(--bg-deepest) 80%, transparent);
    display: flex;
    align-items: center;
    gap: var(--sp-sm);
    z-index: 500;
    cursor: pointer;
  }
  .toast-icon { color: var(--accent-green); }
  .toast-body { display: flex; flex-direction: column; gap: 2px; }
  .toast-text {
    font-family: var(--font-ui);
    font-size: var(--text-body);
    color: var(--text-primary);
  }
  .toast-path {
    font-family: var(--font-mono);
    font-size: var(--text-label);
    color: var(--text-dim);
  }
</style>
