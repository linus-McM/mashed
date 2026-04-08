<script>
  import { SvelteFlowProvider } from '@xyflow/svelte';
  import '@xyflow/svelte/dist/style.css';
  import { writable } from 'svelte/store';
  import { createEventDispatcher, onMount, onDestroy } from 'svelte';
  import { ArrowLeft, Save } from 'lucide-svelte';
  import { EventsOn } from '../../wailsjs/runtime/runtime.js';
  import { GetBmadProcesses, ListBmadTemplates, ListBmadWorkflowsByRepo,
           SaveBmadWorkflow, GetBmadWorkflow, CreateFromTemplate,
           DeleteBmadWorkflow, ListBmadAgents, SaveBmadAgent, DeleteBmadAgent,
           StartBmadWorkflow, PauseBmadWorkflow, ResumeBmadWorkflow, StopBmadWorkflow,
           GetTerminalPort, GetSprintStatus } from '../../wailsjs/go/main/App.js';
  import Terminal from '../components/Terminal.svelte';
  import ProcessSidebar from '../components/bmad/ProcessSidebar.svelte';
  import CanvasPane from '../components/bmad/CanvasPane.svelte';
  import ProcessNode from '../components/bmad/ProcessNode.svelte';
  import ExecutionBar from '../components/bmad/ExecutionBar.svelte';
  import NodeConfigPanel from '../components/bmad/NodeConfigPanel.svelte';
  import AgentConfigModal from '../components/bmad/AgentConfigModal.svelte';
  import RepoContextBar from '../components/bmad/RepoContextBar.svelte';

  export let repoPath = '';

  const dispatch = createEventDispatcher();

  const nodeTypes = { bmadProcess: ProcessNode };

  const nodes = writable([]);
  const edges = writable([]);

  let processes = [];
  let templates = [];
  let savedWorkflows = [];
  let agents = [];
  let sprintStatus = null;
  let currentWorkflow = null;
  let workflowName = 'Untitled Workflow';
  let saving = false;

  // Execution state
  let executionId = null;
  let executionStatus = 'idle';
  let nodeProgress = { completed: 0, total: 0 };
  let execError = '';

  // Config panel state
  let selectedNode = null;

  // Agent modal state
  let showAgentModal = false;
  let editingAgent = null;

  // Terminal modal state
  let showTerminalModal = false;
  let terminalTarget = '';
  $: terminalRepoPath = repoPath;

  onMount(async () => {
    try {
      [processes, templates, savedWorkflows, agents] = await Promise.all([
        GetBmadProcesses(),
        ListBmadTemplates(),
        repoPath ? ListBmadWorkflowsByRepo(repoPath) : Promise.resolve([]),
        ListBmadAgents(),
      ]);
    } catch (e) {
      console.error('Failed to load BMAD data:', e);
      processes = processes || [];
      templates = templates || [];
      savedWorkflows = savedWorkflows || [];
      agents = agents || [];
    }

    if (repoPath) {
      try {
        sprintStatus = await GetSprintStatus(repoPath);
      } catch (e) {
        console.warn('No sprint status for repo:', e);
      }
    }
  });

  // Listen for live node status updates (scoped by execID)
  const cancelStatusListener = EventsOn('bmad:node:status', (event) => {
    if (!event?.nodeId || (executionId && event.execId !== executionId)) return;
    $nodes = $nodes.map(n => {
      if (n.id === event.nodeId) {
        return { ...n, data: { ...n.data, status: event.status, tmuxTarget: event.tmuxTarget || n.data.tmuxTarget } };
      }
      return n;
    });
    // Update selected node if it matches
    if (selectedNode && selectedNode.id === event.nodeId) {
      selectedNode = $nodes.find(n => n.id === event.nodeId) || selectedNode;
    }
    updateProgress();
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

  function onConnect(params) {
    const newEdge = {
      id: `edge-${Date.now()}`,
      source: params.source,
      target: params.target,
      sourceHandle: params.sourceHandle,
      targetHandle: params.targetHandle,
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
        return {
          ...e,
          source: newConnection.source,
          target: newConnection.target,
          sourceHandle: newConnection.sourceHandle,
          targetHandle: newConnection.targetHandle,
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
    // For multi-select, keep selectedNode as null (hides NodeConfigPanel)
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
      agents = await ListBmadAgents();
      showAgentModal = false;
      editingAgent = null;
    } catch (err) {
      console.error('Failed to save agent:', err);
    }
  }

  async function handleAgentDelete(e) {
    try {
      await DeleteBmadAgent(e.detail);
      agents = await ListBmadAgents();
      showAgentModal = false;
      editingAgent = null;
    } catch (err) {
      console.error('Failed to delete agent:', err);
    }
  }

  async function onOpenTerminal(e) {
    terminalTarget = e.detail;
    showTerminalModal = true;
  }

  async function saveWorkflow() {
    saving = true;
    try {
      const wf = {
        id: currentWorkflow?.id || `wf-${Date.now()}`,
        name: workflowName,
        description: '',
        repoPath: repoPath,
        nodes: $nodes.map(n => ({
          id: n.id,
          processId: n.data.processId,
          label: n.data.label,
          position: n.position,
          status: n.data.status || 'pending',
          config: n.data.config || {},
          tmuxTarget: n.data.tmuxTarget || '',
          storyId: n.data.storyId || '',
        })),
        edges: $edges.map(e => ({ id: e.id, source: e.source, target: e.target })),
        isTemplate: false,
        templateId: currentWorkflow?.templateId || '',
        createdAt: currentWorkflow?.createdAt || new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      };
      await SaveBmadWorkflow(wf);
      currentWorkflow = wf;
      savedWorkflows = repoPath ? await ListBmadWorkflowsByRepo(repoPath) : [];
    } catch (e) {
      console.error('Failed to save workflow:', e);
    }
    saving = false;
  }

  function loadNodesEdges(wf) {
    $nodes = (wf.nodes || []).map(n => ({
      id: n.id,
      type: 'bmadProcess',
      position: n.position,
      data: {
        label: n.label,
        processId: n.processId,
        process: processes.find(p => p.id === n.processId) || null,
        status: n.status || 'pending',
        config: n.config || {},
        tmuxTarget: n.tmuxTarget || '',
        storyId: n.storyId || '',
        storyStatus: n.storyStatus || '',
      },
    }));
    $edges = (wf.edges || []).map(e => ({
      id: e.id,
      source: e.source,
      target: e.target,
    }));
    selectedNode = null;
    updateProgress();
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
  }

  async function handleExecStart(e) {
    const { model } = e.detail;
    execError = '';

    // Auto-save before executing
    if (!currentWorkflow?.id) {
      await saveWorkflow();
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
      executionId = await StartBmadWorkflow(currentWorkflow.id, repoPath, model || 'claude-opus-4-6');
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
    on:use-template={useTemplate}
    on:load-workflow={loadWorkflow}
    on:delete-workflow={deleteWorkflow}
  />

  <div class="canvas-area">
    <RepoContextBar {repoPath} {sprintStatus} />
    <div class="toolbar">
      <button class="toolbar-btn back-btn" on:click={() => dispatch('back')} title="Back to feed">
        <ArrowLeft size={14} />
      </button>
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
          {onNodeClick}
          {onNodesDelete}
          {onEdgesDelete}
          {onSelectionChange}
          {onReconnect}
          {executionStatus}
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
        {agents}
        on:update={onConfigUpdate}
        on:close={() => selectedNode = null}
        on:open-terminal={onOpenTerminal}
      />
    </div>

    {#if execError}
      <div class="exec-error">
        <span class="exec-error-text">{execError}</span>
        <button class="exec-error-dismiss" on:click={() => execError = ''}>Dismiss</button>
      </div>
    {/if}

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

  {#if showTerminalModal}
    <div class="terminal-overlay" on:click={() => showTerminalModal = false} on:keydown={() => {}}>
      <div class="terminal-modal" on:click|stopPropagation on:keydown={() => {}}>
        <div class="terminal-modal-header">
          <span class="terminal-modal-title">Node Terminal: {terminalTarget}</span>
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
    padding: 6px 10px;
    background: var(--bg-surface);
    border-bottom: 1px solid var(--border-subtle);
    flex-shrink: 0;
  }

  .toolbar-btn {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 4px 8px;
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-sm);
    color: var(--text-dim);
    font-family: var(--font-mono);
    font-size: 11px;
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
  .back-btn:hover { border-color: var(--accent-blue, #58a6ff); }
  .agents-btn:hover { border-color: var(--accent-purple, #9d6fff); color: var(--accent-purple, #9d6fff); }

  .workflow-name-input {
    flex: 1;
    min-width: 120px;
    padding: 4px 8px;
    background: var(--bg-deepest);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-sm);
    color: var(--text-primary);
    font-family: var(--font-mono);
    font-size: 12px;
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
    padding: 6px 10px;
    background: rgba(248, 81, 73, 0.1);
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
    font-size: 10px;
    padding: 2px 6px;
    cursor: pointer;
    flex-shrink: 0;
  }

  .exec-error-dismiss:hover {
    background: rgba(248, 81, 73, 0.15);
  }

  /* Terminal modal */
  .terminal-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.6);
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

  .terminal-modal-close {
    background: none;
    border: none;
    color: var(--text-muted);
    font-size: 18px;
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
