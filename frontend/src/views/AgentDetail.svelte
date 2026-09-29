<script>
  import { onMount, onDestroy, createEventDispatcher } from 'svelte';
  import { GetScopedDiff, GetWorktrees, ListRepoFiles, GitCommit, GitCommitAndPush, GitCommitPushAndPR, GitCommitStreaming, GitPull, GitPush, SpawnPRReview, RepoStatus, RepoMtimes, KillTerminalSession, SpawnTerminal, SetActiveContext } from '../../wailsjs/go/main/App.js';
  import { repoSessions, refreshSessions } from '../lib/stores/sessions';
  import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime.js';
  import { ArrowLeft, GitBranch, GripVertical } from 'lucide-svelte';
  import StatusBadge from '../components/StatusBadge.svelte';
  import Terminal from '../components/Terminal.svelte';
  import EditorRouter from '../components/EditorRouter.svelte';
  import FileStrip from '../components/agent/FileStrip.svelte';
  import CommitOutputPanel from '../components/agent/CommitOutputPanel.svelte';
  import SubAgentPanel from '../components/agent/SubAgentPanel.svelte';
  import SessionTabs from '../components/agent/SessionTabs.svelte';
  import BranchModal from './BranchModal.svelte';
  import MergeModal from './MergeModal.svelte';
  import ForcePushModal from './ForcePushModal.svelte';
  import { errorMessage } from '../lib/errorMessage';
  import { estimatePtySize } from '../lib/ptySize';

  /** @type {import('svelte').EventDispatcher<{ back: void }>} */
  const dispatch = createEventDispatcher();

  /** @typedef {import('../lib/types/wails').Notification} AgentLike */
  /** @typedef {{ path: string; isBinary?: boolean; added?: number; removed?: number }} RepoFile */
  /** @typedef {'file-strip' | 'editor'} DragPane */
  /** @typedef {'commit' | 'push' | 'pull' | 'pr' | 'review'} GitActionName */

  /** @type {AgentLike} */
  export let agent;

  /** @type {RepoFile[]} */
  let changedFiles = [];
  /** @type {string[]} */
  let allFiles = []; // all repo files for the "All" tab
  /** @type {import('../lib/types/wails').WorktreeInfo | null} */
  let worktree = null;
  /** @type {RepoFile | null} */
  let selectedFile = null; // when set, shows code editor on right
  let fileTab = 'changed'; // 'changed' or 'all'

  // Resizable pane widths
  /** @type {HTMLElement | undefined} */
  let workspaceEl;
  let fileStripWidth = 180;
  let editorFraction = 0.5; // fraction of remaining space for editor
  /** @type {DragPane | null} */
  let dragging = null; // 'file-strip' or 'editor'

  /** @param {DragPane} pane */
  function onMouseDown(pane) {
    /** @param {MouseEvent} e */
    return (e) => {
      e.preventDefault();
      dragging = pane;
      document.addEventListener('mousemove', onMouseMove);
      document.addEventListener('mouseup', onMouseUp);
      document.body.style.cursor = 'col-resize';
      document.body.style.userSelect = 'none';
    };
  }

  /** @param {MouseEvent} e */
  function onMouseMove(e) {
    if (!dragging || !workspaceEl) return;
    const rect = workspaceEl.getBoundingClientRect();
    const x = e.clientX - rect.left;
    const totalW = rect.width;

    if (dragging === 'file-strip') {
      // Dragging the handle between terminal and file strip
      // file strip starts at x and goes to either editor or right edge
      if (selectedFile) {
        const editorW = totalW * editorFraction;
        const newStripW = totalW - x - editorW;
        fileStripWidth = Math.max(120, Math.min(350, newStripW));
      } else {
        const newStripW = totalW - x;
        fileStripWidth = Math.max(120, Math.min(350, newStripW));
      }
    } else if (dragging === 'editor') {
      // Dragging the handle between file strip and editor
      const editorW = totalW - x;
      editorFraction = Math.max(0.2, Math.min(0.8, editorW / totalW));
    }
  }

  function onMouseUp() {
    dragging = null;
    document.removeEventListener('mousemove', onMouseMove);
    document.removeEventListener('mouseup', onMouseUp);
    document.body.style.cursor = '';
    document.body.style.userSelect = '';
  }

  $: isSubAgent = agent?.isSubAgent || false;
  $: tokenPct = agent ? Math.min(((agent.tokensUsed || 0) / (agent.tokensMax || 1)) * 100, 100) : 0;
  $: tokenLabel = agent ? formatTokens(agent.tokensUsed || 0) + ' / ' + formatTokens(agent.tokensMax || 0) : '';

  /** @param {number} n */
  function formatTokens(n) {
    if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + 'M';
    if (n >= 1_000) return (n / 1_000).toFixed(1) + 'K';
    return String(n);
  }

  // Session tab state
  let activeSessionIdx = 0;
  /** @type {string | null} */
  let pendingPaneTarget = null;
  $: sessions = $repoSessions[agent?.repoPath] || [];
  $: {
    if (pendingPaneTarget) {
      const newIdx = sessions.findIndex(s => s.paneTarget === pendingPaneTarget);
      if (newIdx >= 0) {
        activeSessionIdx = newIdx;
        pendingPaneTarget = null;
      }
    } else if (activeSessionIdx === 0 && sessions.length > 0) {
      const matchIdx = sessions.findIndex(s => s.paneTarget === agent?.tmuxTarget);
      if (matchIdx >= 0) activeSessionIdx = matchIdx;
    }
  }
  $: activeSession = sessions[activeSessionIdx] || null;

  // Keep backend aware of which terminal is active (used for screenshot routing)
  $: if (agent?.repoPath && activeSession?.paneTarget) {
    SetActiveContext(agent.repoPath, activeSession.paneTarget);
  }

  /** @param {string} sessionName */
  async function killSession(sessionName) {
    await KillTerminalSession(sessionName);
    if (sessions[activeSessionIdx]?.sessionName === sessionName) {
      activeSessionIdx = 0;
    }
  }

  async function spawnNewTerminal() {
    const { cols, rows } = estimatePtySize();
    const target = await SpawnTerminal(agent.repoPath, cols, rows);
    pendingPaneTarget = target;
  }

  // Git action state
  /** @type {GitActionName | null} */
  let gitAction = null;
  /** @type {string | null} */
  let gitResult = null;
  /** @type {string | null} */
  let gitError = null;
  let showBranchModal = false;
  let showMergeModal = false;
  /** @type {{ message: string } | null} */
  let forcePushState = null;

  // Repo status for push highlighting
  let repoStatus = { ahead: 0, behind: 0, protected: false };
  /** @type {ReturnType<typeof setInterval> | undefined} */
  let statusInterval;

  async function refreshRepoStatus() {
    if (!agent?.repoPath) return;
    try {
      repoStatus = await RepoStatus(agent.repoPath);
    } catch (_) {}
  }

  $: if (agent?.repoPath) refreshRepoStatus();

  async function smartPush() {
    if (!agent?.repoPath) return;
    gitAction = 'push';
    gitResult = null;
    gitError = null;
    try {
      const result = await GitPush(agent.repoPath);
      if (result.startsWith('conflict:')) {
        forcePushState = { message: result.slice('conflict:'.length) };
        gitAction = null;
        return;
      }
      gitResult = 'Pushed';
      refreshChangedFiles();
      refreshAllFiles();
    } catch (err) {
      gitError = errorMessage(err);
    }
    gitAction = null;
    refreshRepoStatus();
    setTimeout(() => { gitResult = null; gitError = null; }, 5000);
  }

  /**
   * @param {GitActionName} actionName
   * @param {(repoPath: string) => Promise<string>} fn
   */
  async function runGitAction(actionName, fn) {
    if (!agent?.repoPath) return;
    gitAction = actionName;
    gitResult = null;
    gitError = null;
    try {
      const result = await fn(agent.repoPath);
      gitResult = result || 'Done';
      refreshChangedFiles();
      refreshAllFiles();
    } catch (err) {
      gitError = errorMessage(err);
    }
    gitAction = null;
    // Clear result/error after 5s
    setTimeout(() => { gitResult = null; gitError = null; }, 5000);
  }

  /** @typedef {{ lines: Array<{ step: string; output: string }>; error: string | null; explanation: string | null; done: boolean }} CommitPanelState */
  /** @type {CommitPanelState | null} */
  let commitPanel = null;

  function startStreamingCommit() {
    if (!agent?.repoPath) return;
    commitPanel = { lines: [], error: null, explanation: null, done: false };
    gitAction = 'commit';
    gitResult = null;
    gitError = null;
    GitCommitStreaming(agent.repoPath);
  }

  function closeCommitPanel() {
    commitPanel = null;
  }

  /** @type {(() => void) | undefined} */
  let commitEventCancel;

  /** @typedef {{ repoPath?: string; step?: string; output?: string; error?: string; explanation?: string; done?: boolean }} CommitProgressEvent */

  function setupCommitListener() {
    commitEventCancel = EventsOn('git:commit:progress', (/** @type {CommitProgressEvent} */ evt) => {
      if (evt.repoPath !== agent?.repoPath) return;
      if (!commitPanel) {
        commitPanel = { lines: [], error: null, explanation: null, done: false };
      }

      if (evt.step && !evt.error) {
        commitPanel.lines = [...commitPanel.lines, { step: evt.step, output: evt.output || '' }];
      }
      if (evt.error) {
        commitPanel.error = evt.error;
        commitPanel.explanation = evt.explanation || null;
      }
      commitPanel.done = !!evt.done;
      commitPanel = commitPanel; // trigger reactivity

      if (evt.done) {
        gitAction = null;
        if (!evt.error) {
          gitResult = evt.output || 'Done';
          refreshChangedFiles();
          refreshAllFiles();
          // Auto-close panel after success
          setTimeout(() => { commitPanel = null; }, 1500);
        } else {
          gitError = evt.error;
        }
      }
    });
  }

  /** @param {KeyboardEvent} e */
  function handleKeydown(e) {
    if (e.key === 'Escape') {
      if (selectedFile) {
        e.preventDefault();
        selectedFile = null;
      } else {
        e.preventDefault();
        dispatch('back');
      }
    }
  }

  // Mtime-based file change detection
  let lastIndexMtime = 0;
  let lastRootMtime = 0;
  /** @type {ReturnType<typeof setInterval> | undefined} */
  let mtimeInterval;

  async function refreshChangedFiles() {
    if (!agent?.repoPath) return;
    try {
      const diff = await GetScopedDiff(agent.repoPath);
      if (diff && diff.files) changedFiles = diff.files;
    } catch {}
  }

  async function refreshAllFiles() {
    if (!agent?.repoPath) return;
    try {
      const files = await ListRepoFiles(agent.repoPath);
      if (files) allFiles = files;
    } catch {}
  }

  async function checkMtimes() {
    if (!agent?.repoPath) return;
    try {
      const mt = await RepoMtimes(agent.repoPath);
      if (!mt) return;

      const indexChanged = mt.index !== lastIndexMtime && lastIndexMtime !== 0;
      const rootChanged = mt.root !== lastRootMtime && lastRootMtime !== 0;

      lastIndexMtime = mt.index;
      lastRootMtime = mt.root;

      if (indexChanged) {
        // Stage/commit/checkout — refresh both tabs
        refreshChangedFiles();
        refreshAllFiles();
      } else if (rootChanged) {
        // Unstaged edits — refresh changed tab only
        refreshChangedFiles();
      }
    } catch {}
  }

  onMount(async () => {
    window.addEventListener('keydown', handleKeydown);
    setupCommitListener();

    if (agent?.repoPath) {
      // Initial loads
      refreshSessions(agent.repoPath);
      refreshChangedFiles();
      refreshAllFiles();

      try {
        const wts = await GetWorktrees(agent.repoPath);
        if (wts && wts.length > 0) worktree = wts[0];
      } catch (err) {
        console.warn('Failed to get worktrees:', err);
      }

      // Seed mtimes then start polling
      try {
        const mt = await RepoMtimes(agent.repoPath);
        if (mt) {
          lastIndexMtime = mt.index;
          lastRootMtime = mt.root;
        }
      } catch {}
      mtimeInterval = setInterval(checkMtimes, 5000);

      // Poll repo status for push highlighting
      statusInterval = setInterval(refreshRepoStatus, 10000);
    }
  });

  onDestroy(() => {
    window.removeEventListener('keydown', handleKeydown);
    if (commitEventCancel) commitEventCancel();
    if (mtimeInterval) clearInterval(mtimeInterval);
    if (statusInterval) clearInterval(statusInterval);
  });
</script>

<div class="detail">
  <!-- Compact header bar -->
  <div class="header">
    <button class="back-btn" on:click={() => dispatch('back')}><ArrowLeft size={14} /> Back</button>
    <span class="header-repo">{agent.repoName}</span>
    <span class="header-sep">/</span>
    <span class="header-agent">{agent.model || agent.agentName}</span>
    {#if isSubAgent}
      <span class="header-sep">/</span>
      <span class="header-sub-name">{agent.subAgentName || agent.agentName}</span>
    {/if}
    <StatusBadge status={agent.eventType} size="sm" />
    {#if agent.repoBranch}
      <span class="header-branch"><GitBranch size={12} /> {agent.repoBranch}</span>
    {/if}
    <div class="header-right">
      {#if !isSubAgent}
        <div class="token-mini">
          <div class="token-mini-bar">
            <div class="token-mini-fill" style="width: {tokenPct}%" />
          </div>
          <span class="token-mini-label">{tokenLabel}</span>
        </div>
      {:else}
        <span class="sub-agent-badge">{agent.subAgentStatus === 'done' ? 'completed' : 'running'}</span>
      {/if}
    </div>
  </div>

  <!-- Sub-agent info panel -->
  {#if isSubAgent}
    <SubAgentPanel {agent} />
  {/if}

  <!-- Main workspace -->
  <div class="workspace" bind:this={workspaceEl}>
    <!-- Terminal pane -->
    <div
      class="terminal-pane"
      style={selectedFile ? `flex: 0 0 calc(100% - ${fileStripWidth}px - ${editorFraction * 100}%)` : `flex: 1 1 0; width: 0`}
    >
      {#if sessions.length > 0}
        <SessionTabs {sessions} bind:activeSessionIdx onKill={killSession} onAdd={spawnNewTerminal} />
      {/if}
      {#key activeSession?.paneTarget}
        {#if (activeSession?.paneTarget || agent?.tmuxTarget) && agent?.eventType !== 'completed' && agent?.eventType !== 'finished'}
          <Terminal paneTarget={activeSession?.paneTarget || agent?.tmuxTarget || ''} repoPath={agent?.repoPath || ''} />
        {:else}
          <div class="session-ended">
            <span class="session-ended-icon">&#x25CB;</span>
            <span>Session ended</span>
          </div>
        {/if}
      {/key}
    </div>

    <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
    <!-- Resize handle: terminal | file strip -->
    <div class="resize-handle" role="separator" on:mousedown={onMouseDown('file-strip')}>
      <div class="resize-grip"><GripVertical size={10} /></div>
    </div>

    <!-- File list strip -->
    <FileStrip
      bind:fileStripWidth
      bind:fileTab
      bind:selectedFile
      {changedFiles}
      {allFiles}
      {worktree}
      {gitAction}
      {gitResult}
      {gitError}
      commitPanelOpen={!!commitPanel}
      {repoStatus}
      onBranch={() => showBranchModal = true}
      onCommit={() => startStreamingCommit()}
      onPull={() => runGitAction('pull', GitPull)}
      onPush={() => smartPush()}
      onMerge={() => showMergeModal = true}
      onPR={() => runGitAction('pr', GitCommitPushAndPR)}
      onReview={() => runGitAction('review', SpawnPRReview)}
    />

    <!-- Resize handle: file strip | editor -->
    {#if selectedFile}
      <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
      <div class="resize-handle" role="separator" on:mousedown={onMouseDown('editor')}>
        <div class="resize-grip"><GripVertical size={10} /></div>
      </div>
    {/if}

    <!-- Code editor (visible when file selected) -->
    {#if selectedFile}
      <div class="editor-pane" style="flex: 0 0 {editorFraction * 100}%">
        <EditorRouter
          filePath={selectedFile.path}
          repoPath={agent.repoPath}
          mode={selectedFile.isBinary ? 'source' : (fileTab === 'all' ? 'source' : 'diff')}
          editable={!selectedFile.isBinary}
        />
      </div>
    {/if}
  </div>

  <!-- Full-width commit output panel -->
  {#if commitPanel}
    <CommitOutputPanel panel={commitPanel} onClose={closeCommitPanel} />
  {/if}

  <!-- Bottom bar -->
  <div class="bottom-bar">
    <kbd>Esc</kbd>
    {#if selectedFile}
      Close editor
    {:else}
      Back to feed
    {/if}
  </div>
</div>

{#if showBranchModal}
  <BranchModal
    repoPath={agent.repoPath}
    repoBranch={agent.repoBranch || 'main'}
    on:close={() => showBranchModal = false}
  />
{/if}

{#if showMergeModal}
  <MergeModal
    repoPath={agent.repoPath}
    currentBranch={agent.repoBranch || 'main'}
    on:merged={(e) => { showMergeModal = false; gitResult = e.detail?.result || 'Merged'; refreshChangedFiles(); }}
    on:cancel={() => showMergeModal = false}
  />
{/if}

{#if forcePushState}
  <ForcePushModal
    repoPath={agent.repoPath}
    conflictMessage={forcePushState.message}
    on:pushed={() => { forcePushState = null; gitResult = 'Force pushed'; refreshRepoStatus(); refreshChangedFiles(); setTimeout(() => { gitResult = null; }, 5000); }}
    on:cancel={() => forcePushState = null}
  />
{/if}

<style>
  .detail {
    display: flex;
    flex-direction: column;
    height: 100%;
    background: var(--bg-deepest);
  }

  /* Compact header */
  .header {
    display: flex;
    align-items: center;
    gap: var(--sp-sm);
    padding: var(--sp-xs) var(--sp-lg);
    border-bottom: 1px solid var(--border-subtle);
    background: var(--bg-surface);
    flex-shrink: 0;
  }

  .back-btn {
    background: none;
    border: none;
    color: var(--accent-green);
    font-family: var(--font-ui);
    font-size: 13px;
    cursor: pointer;
    padding: var(--sp-2xs) var(--sp-xs);
    border-radius: var(--radius-sm);
    text-shadow: 0 0 6px color-mix(in srgb, var(--accent-green) 60%, transparent);
  }

  .back-btn:hover {
    color: var(--accent-green);
    text-shadow:
      0 0 10px color-mix(in srgb, var(--accent-green) 90%, transparent),
      0 0 20px color-mix(in srgb, var(--accent-green) 40%, transparent);
  }

  .header-repo {
    font-weight: 600;
    font-size: var(--text-data);
    color: var(--text-primary);
  }

  .header-sep { color: var(--text-muted); }

  .header-agent {
    font-family: var(--font-mono);
    font-size: 13px;
    color: var(--text-dim);
  }

  .header-branch {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--text-muted);
  }

  .header-right {
    margin-left: auto;
    display: flex;
    align-items: center;
    gap: var(--sp-md);
  }

  .token-mini {
    display: flex;
    align-items: center;
    gap: var(--sp-sm);
  }

  .token-mini-bar {
    width: 60px;
    height: 4px;
    background: var(--bg-active);
    border-radius: 2px;
    overflow: hidden;
  }

  .token-mini-fill {
    height: 100%;
    background: var(--accent-teal);
    border-radius: 2px;
  }

  .token-mini-label {
    font-family: var(--font-mono);
    font-size: var(--text-label);
    font-variant-numeric: tabular-nums;
    color: var(--text-muted);
    white-space: nowrap;
  }

  /* Workspace: terminal + file strip + editor */
  .workspace {
    flex: 1;
    display: flex;
    overflow: hidden;
    max-width: 100%;
    width: 100%;
  }

  .terminal-pane {
    display: flex;
    flex-direction: column;
    min-width: 100px;
    overflow: hidden;
  }
  .session-ended {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 8px;
    color: var(--text-muted);
    font-size: 13px;
  }
  .session-ended-icon {
    font-size: 24px;
    opacity: 0.4;
  }

  /* Resize handle */
  .resize-handle {
    width: 6px;
    flex-shrink: 0;
    cursor: col-resize;
    background: var(--border-subtle);
    display: flex;
    align-items: center;
    justify-content: center;
    transition: background 100ms ease;
    position: relative;
    z-index: 2;
  }

  .resize-handle:hover, .resize-handle:active {
    background: var(--accent-green);
  }

  .resize-grip {
    color: var(--text-muted);
    opacity: 0;
    transition: opacity 100ms ease;
  }

  .resize-handle:hover .resize-grip {
    opacity: 1;
    color: var(--bg-deepest);
  }

  /* Editor pane */
  .editor-pane {
    display: flex;
    flex-direction: column;
    min-width: 150px;
    overflow: hidden;
  }

  /* Bottom bar */
  .bottom-bar {
    display: flex;
    align-items: center;
    gap: var(--sp-sm);
    padding: 4px var(--sp-lg);
    border-top: 1px solid var(--border-subtle);
    background: var(--bg-surface);
    font-size: 11px;
    color: var(--text-muted);
    user-select: none;
    flex-shrink: 0;
  }

  .bottom-bar :global(kbd) {
    display: inline-block;
    padding: 0 4px;
    background: var(--bg-active);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-sm);
    font-family: var(--font-mono);
    font-size: var(--text-label);
    color: var(--text-dim);
  }

  /* Sub-agent header badge */
  .header-sub-name {
    font-family: var(--font-mono);
    font-size: var(--text-label);
    color: var(--accent-teal);
  }

  .sub-agent-badge {
    font-family: var(--font-mono);
    font-size: var(--text-label);
    padding: 1px 8px;
    border-radius: 8px;
    background: rgba(0, 196, 179, 0.12);
    color: var(--accent-teal);
    border: 1px solid rgba(0, 196, 179, 0.25);
  }
</style>
