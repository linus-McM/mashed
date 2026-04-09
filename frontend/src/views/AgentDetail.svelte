<script>
  import { onMount, onDestroy, createEventDispatcher } from 'svelte';
  import { GetScopedDiff, GetWorktrees, ListRepoFiles, GitCommit, GitCommitAndPush, GitCommitPushAndPR, GitCommitStreaming, GitPull, GitPush, SpawnPRReview, RepoStatus, RepoMtimes, KillTerminalSession, SpawnTerminal, SetActiveContext } from '../../wailsjs/go/main/App.js';
  import { repoSessions, refreshSessions } from '../lib/stores/sessions.js';
  import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime.js';
  import { ArrowLeft, GitBranch, GripVertical, GitCommit as GitCommitIcon, Upload, GitPullRequest, ShieldAlert, GitBranchPlus, Download, GitMerge, AlertTriangle } from 'lucide-svelte';
  import StatusBadge from '../components/StatusBadge.svelte';
  import Terminal from '../components/Terminal.svelte';
  import EditorRouter from '../components/EditorRouter.svelte';
  import FileTree from '../components/FileTree.svelte';
  import BranchModal from './BranchModal.svelte';
  import MergeModal from './MergeModal.svelte';
  import ForcePushModal from './ForcePushModal.svelte';

  const dispatch = createEventDispatcher();

  export let agent;

  let changedFiles = [];
  let allFiles = []; // all repo files for the "All" tab
  let worktree = null;
  let selectedFile = null; // when set, shows code editor on right
  let fileTab = 'changed'; // 'changed' or 'all'
  let allFilesSearch = ''; // search filter for all files tab

  // Resizable pane widths
  let workspaceEl;
  let fileStripWidth = 180;
  let editorFraction = 0.5; // fraction of remaining space for editor
  let dragging = null; // 'file-strip' or 'editor'

  function onMouseDown(pane) {
    return (e) => {
      e.preventDefault();
      dragging = pane;
      document.addEventListener('mousemove', onMouseMove);
      document.addEventListener('mouseup', onMouseUp);
      document.body.style.cursor = 'col-resize';
      document.body.style.userSelect = 'none';
    };
  }

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
      const availableForEditor = totalW - fileStripWidth;
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

  function formatTokens(n) {
    if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + 'M';
    if (n >= 1_000) return (n / 1_000).toFixed(1) + 'K';
    return String(n);
  }

  // Session tab state
  let activeSessionIdx = 0;
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

  async function killSession(sessionName) {
    await KillTerminalSession(sessionName);
    if (sessions[activeSessionIdx]?.sessionName === sessionName) {
      activeSessionIdx = 0;
    }
  }

  async function spawnNewTerminal() {
    const target = await SpawnTerminal(agent.repoPath);
    pendingPaneTarget = target;
  }

  // Git action state
  let gitAction = null; // 'commit' | 'push' | 'pr' | 'review' | null
  let gitResult = null;
  let gitError = null;
  let showBranchModal = false;
  let showMergeModal = false;
  let forcePushState = null; // { message } or null

  // Repo status for push highlighting
  let repoStatus = { ahead: 0, behind: 0, protected: false };
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
      gitError = err?.message || String(err);
    }
    gitAction = null;
    refreshRepoStatus();
    setTimeout(() => { gitResult = null; gitError = null; }, 5000);
  }

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
      gitError = err?.message || String(err);
    }
    gitAction = null;
    // Clear result/error after 5s
    setTimeout(() => { gitResult = null; gitError = null; }, 5000);
  }

  // Streaming commit panel
  let commitPanel = null; // { lines[], error, explanation, done }

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

  let commitEventCancel;

  function setupCommitListener() {
    commitEventCancel = EventsOn('git:commit:progress', (evt) => {
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

  $: filteredAllFiles = allFilesSearch
    ? allFiles.filter(f => f.toLowerCase().includes(allFilesSearch.toLowerCase()))
    : allFiles;

  function selectFile(file) {
    if (selectedFile?.path === file.path) {
      selectedFile = null; // toggle off
    } else {
      selectedFile = file;
      fileStripWidth = 350;
    }
  }

  function selectAllFile(filePath) {
    if (selectedFile?.path === filePath) {
      selectedFile = null;
    } else {
      selectedFile = { path: filePath, isBinary: false };
      fileStripWidth = 350;
    }
  }

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
      } catch (e) {
        console.warn('Failed to get worktrees:', e);
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
    <div class="sub-agent-panel">
      {#if agent.subAgentDesc}
        <div class="sub-panel-desc">{agent.subAgentDesc}</div>
      {/if}
      {#if agent.subAgentResult}
        <div class="sub-panel-result">
          <span class="sub-panel-label">Result</span>
          <pre class="sub-panel-text">{agent.subAgentResult}</pre>
        </div>
      {/if}
      {#if agent.subAgentLogLines && agent.subAgentLogLines.length > 0}
        <div class="sub-panel-logs">
          <span class="sub-panel-label">Activity ({agent.subAgentLogLines.length} events)</span>
          <div class="sub-panel-log-list">
            {#each agent.subAgentLogLines as line}
              <div class="sub-log-line" class:log-ok={line.kind === 'ok'} class:log-err={line.kind === 'err'} class:log-dim={line.kind === 'dim'} class:log-system={line.kind === 'system'}>
                <span class="sub-log-text">{line.text}</span>
              </div>
            {/each}
          </div>
        </div>
      {/if}
    </div>
  {/if}

  <!-- Main workspace -->
  <div class="workspace" bind:this={workspaceEl}>
    <!-- Terminal pane -->
    <div
      class="terminal-pane"
      style={selectedFile ? `flex: 0 0 calc(100% - ${fileStripWidth}px - ${editorFraction * 100}%)` : `flex: 1 1 0; width: 0`}
    >
      {#if sessions.length > 0}
      <div class="session-tabs">
        {#each sessions as session, idx}
          <button
            class="session-tab"
            class:active={idx === activeSessionIdx}
            on:click={() => activeSessionIdx = idx}
          >
            <span class="tab-type">{session.sessionType === 'agent' ? 'Agent' : 'Term'}</span>
            <span class="tab-name">{session.sessionName.split('-').slice(-1)[0]}</span>
            <button class="tab-close" on:click|stopPropagation={() => killSession(session.sessionName)}>×</button>
          </button>
        {/each}
        <button class="session-tab add-tab" on:click={spawnNewTerminal}>+</button>
      </div>
      {/if}
      {#key activeSession?.paneTarget}
        <Terminal paneTarget={activeSession?.paneTarget || agent?.tmuxTarget || ''} repoPath={agent?.repoPath || ''} />
      {/key}
    </div>

    <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
    <!-- Resize handle: terminal | file strip -->
    <div class="resize-handle" role="separator" on:mousedown={onMouseDown('file-strip')}>
      <div class="resize-grip"><GripVertical size={10} /></div>
    </div>

    <!-- File list strip -->
    <div class="file-strip visible" style="width: {fileStripWidth}px">
      <div class="file-strip-header">
        <div class="file-tabs">
          <button
            class="file-tab"
            class:active={fileTab === 'changed'}
            on:click={() => fileTab = 'changed'}
          >
            Changed
            {#if changedFiles.length > 0}
              <span class="file-tab-count">{changedFiles.length}</span>
            {/if}
          </button>
          <button
            class="file-tab"
            class:active={fileTab === 'all'}
            on:click={() => fileTab = 'all'}
          >
            All Files
          </button>
        </div>
      </div>

      {#if fileTab === 'changed'}
        <div class="file-strip-list">
          {#each changedFiles as file}
            <button
              class="file-item"
              class:active={selectedFile?.path === file.path}
              class:binary={file.isBinary}
              on:click={() => selectFile(file)}
              title={file.path}
            >
              <span class="file-name">{file.path.split('/').pop()}</span>
              {#if file.isBinary}
                <span class="file-stat binary">bin</span>
              {:else}
                <span class="file-stat">
                  {#if file.added > 0}<span class="added">+{file.added}</span>{/if}
                  {#if file.removed > 0}<span class="removed">-{file.removed}</span>{/if}
                </span>
              {/if}
            </button>
          {/each}
          {#if changedFiles.length === 0}
            <div class="file-strip-empty">No changes</div>
          {/if}
        </div>
      {:else}
        <FileTree
          files={allFiles}
          selectedPath={selectedFile?.path || ''}
          changedPaths={new Set(changedFiles.map(f => f.path))}
          on:select={e => selectAllFile(e.detail.path)}
        />
      {/if}

      <!-- Git actions -->
      <div class="git-actions">
        {#if !commitPanel}
          {#if gitResult}
            <div class="git-status git-success">{gitResult}</div>
          {:else if gitError}
            <div class="git-status git-error">{gitError}</div>
          {/if}
        {/if}
        <button class="git-btn" disabled={!!gitAction} on:click={() => showBranchModal = true}>
          <GitBranchPlus size={14} /> Branch
        </button>
        <button class="git-btn" class:git-hot={changedFiles.length > 0} disabled={!!gitAction} on:click={() => startStreamingCommit()}>
          <GitCommitIcon size={14} /> {gitAction === 'commit' ? 'Committing...' : 'Commit'}
        </button>
        <button class="git-btn" disabled={!!gitAction} on:click={() => runGitAction('pull', GitPull)}>
          <Download size={14} /> {gitAction === 'pull' ? 'Pulling...' : 'Pull'}
        </button>
        <button
          class="git-btn"
          class:git-hot={(repoStatus.ahead || 0) > 0 && !repoStatus.protected}
          disabled={!!gitAction}
          on:click={() => smartPush()}
          title={repoStatus.protected ? 'Branch is protected — push via PR' : (repoStatus.ahead || 0) > 0 ? `${repoStatus.ahead} commit(s) ahead of remote` : 'Push to origin'}
        >
          <Upload size={14} /> {gitAction === 'push' ? 'Pushing...' : 'Push'}
        </button>
        <button class="git-btn" disabled={!!gitAction} on:click={() => showMergeModal = true}>
          <GitMerge size={14} /> Merge
        </button>
        <button class="git-btn" class:git-hot={changedFiles.length > 0} disabled={!!gitAction} on:click={() => runGitAction('pr', GitCommitPushAndPR)}>
          <GitPullRequest size={14} /> {gitAction === 'pr' ? 'Creating...' : 'PR'}
        </button>
        <button class="git-btn" disabled={!!gitAction} on:click={() => runGitAction('review', SpawnPRReview)}>
          <ShieldAlert size={14} /> {gitAction === 'review' ? 'Reviewing...' : 'Review'}
        </button>
      </div>

      {#if worktree}
        <div class="worktree-mini">
          <span class="worktree-label">WT</span>
          <span class="worktree-branch">{worktree.branch}</span>
        </div>
      {/if}
    </div>

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
    <div class="ws-commit-panel">
      <div class="ws-commit-header">
        <span class="ws-commit-title">
          {#if commitPanel.done && !commitPanel.error}
            Committed
          {:else if commitPanel.error}
            Commit Failed
          {:else}
            Committing...
          {/if}
        </span>
        {#if commitPanel.done}
          <button class="ws-commit-close" on:click={() => closeCommitPanel()}>×</button>
        {/if}
      </div>
      <div class="ws-commit-body">
        {#each commitPanel.lines as line}
          <div class="ws-commit-line">
            <span class="ws-commit-step">{line.step}</span>
            {#if line.output}
              <pre class="ws-commit-output">{line.output}</pre>
            {/if}
          </div>
        {/each}
        {#if !commitPanel.done && !commitPanel.error}
          <div class="ws-commit-line ws-commit-active">
            <span class="ws-commit-spinner" />
          </div>
        {/if}
      </div>
      {#if commitPanel.error}
        <div class="ws-commit-error-section">
          <div class="ws-commit-error-label">Error</div>
          <pre class="ws-commit-error-text">{commitPanel.error}</pre>
          {#if commitPanel.explanation}
            <div class="ws-commit-explain-label">Why this happened</div>
            <div class="ws-commit-explain-text">{commitPanel.explanation}</div>
          {:else if !commitPanel.done}
            <div class="ws-commit-explain-loading">Analyzing failure...</div>
          {/if}
        </div>
      {/if}
    </div>
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
    padding: 6px var(--sp-lg);
    border-bottom: 1px solid var(--border-subtle);
    background: var(--bg-surface);
    flex-shrink: 0;
  }

  .back-btn {
    background: none;
    border: none;
    color: #39ff14;
    font-family: var(--font-ui);
    font-size: 13px;
    cursor: pointer;
    padding: 2px 6px;
    border-radius: var(--radius-sm);
    text-shadow: 0 0 6px rgba(57, 255, 20, 0.6);
  }

  .back-btn:hover {
    color: #39ff14;
    text-shadow: 0 0 10px rgba(57, 255, 20, 0.9), 0 0 20px rgba(57, 255, 20, 0.4);
  }

  .header-repo {
    font-weight: 600;
    font-size: 14px;
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
    gap: 6px;
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
    font-size: 10px;
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

  /* Session tab bar */
  .session-tabs {
    display: flex;
    align-items: center;
    gap: 2px;
    padding: 0 8px;
    height: 32px;
    background: var(--bg-deeper);
    border-bottom: 1px solid var(--border-subtle);
    overflow-x: auto;
    flex-shrink: 0;
  }
  .session-tab {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 4px 8px;
    border: none;
    background: transparent;
    color: var(--fg-secondary);
    font-size: 11px;
    cursor: pointer;
    border-bottom: 2px solid transparent;
    white-space: nowrap;
  }
  .session-tab.active {
    color: var(--fg-primary);
    border-bottom-color: var(--border-accent);
  }
  .session-tab:hover { color: var(--fg-primary); }
  .tab-close {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 16px;
    height: 16px;
    border: none;
    background: transparent;
    color: var(--fg-dim);
    font-size: 12px;
    cursor: pointer;
    border-radius: 3px;
    padding: 0;
  }
  .tab-close:hover { background: rgba(255, 95, 87, 0.2); color: #ff5f57; }
  .add-tab { color: var(--fg-dim); font-size: 14px; }
  .add-tab:hover { color: var(--accent-green, #50fa7b); }

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

  /* File strip */
  .file-strip {
    flex-shrink: 0;
    overflow: hidden;
    background: var(--bg-surface);
    display: flex;
    flex-direction: column;
    min-width: 120px;
  }

  .file-strip-header {
    display: flex;
    align-items: center;
    padding: 0;
    border-bottom: 1px solid var(--border-subtle);
    flex-shrink: 0;
  }

  .file-tabs {
    display: flex;
    width: 100%;
  }

  .file-tab {
    flex: 1;
    padding: 6px 8px;
    background: none;
    border: none;
    border-bottom: 2px solid transparent;
    font-size: 10px;
    font-weight: 600;
    color: var(--text-muted);
    letter-spacing: 0.04em;
    cursor: pointer;
    font-family: var(--font-ui);
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 4px;
  }

  .file-tab:hover { color: var(--text-dim); }

  .file-tab.active {
    color: var(--accent-green);
    border-bottom-color: var(--accent-green);
  }

  .file-tab-count {
    font-family: var(--font-mono);
    font-size: 9px;
    background: var(--bg-active);
    padding: 0 4px;
    border-radius: 8px;
    color: var(--accent-green);
  }

  .file-strip-empty {
    padding: 12px 8px;
    font-size: 11px;
    color: var(--text-muted);
    text-align: center;
  }


  .file-strip-list {
    flex: 1;
    overflow-y: auto;
  }

  .file-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    width: 100%;
    padding: 3px 8px;
    background: none;
    border: none;
    border-bottom: 1px solid var(--border-subtle);
    cursor: pointer;
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--text-primary);
    text-align: left;
    transition: background 100ms ease-out;
  }

  .file-item:hover { background: var(--bg-elevated); }

  .file-item.active {
    background: var(--bg-active);
    border-left: 2px solid var(--accent-green);
  }

  .file-item.binary { color: var(--text-dim); }

  .file-name {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    flex: 1;
  }

  .file-stat {
    flex-shrink: 0;
    margin-left: 4px;
    font-size: 10px;
  }

  .file-stat .added { color: var(--accent-green); }
  .file-stat .removed { color: var(--accent-red); margin-left: 2px; }
  .file-stat.binary { color: var(--text-muted); font-style: italic; }

  /* Git actions */
  .git-actions {
    padding: 6px 8px;
    border-top: 1px solid var(--border-subtle);
    display: flex;
    flex-direction: column;
    gap: 2px;
    flex-shrink: 0;
  }

  .git-btn {
    display: flex;
    align-items: center;
    gap: 6px;
    width: 100%;
    padding: 5px 8px;
    background: none;
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-sm);
    color: var(--text-dim);
    font-family: var(--font-ui);
    font-size: 11px;
    cursor: pointer;
    text-align: left;
  }

  .git-btn:hover:not(:disabled) {
    background: var(--bg-elevated);
    color: var(--text-primary);
    border-color: var(--text-muted);
  }

  .git-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .git-btn.git-hot {
    color: #39ff14;
    border-color: rgba(57, 255, 20, 0.3);
    text-shadow: 0 0 6px rgba(57, 255, 20, 0.4);
  }

  .git-btn.git-hot:hover:not(:disabled) {
    color: #39ff14;
    border-color: #39ff14;
    background: rgba(57, 255, 20, 0.08);
    text-shadow: 0 0 10px rgba(57, 255, 20, 0.6);
  }

  .git-status {
    font-family: var(--font-mono);
    font-size: 10px;
    padding: 3px 6px;
    border-radius: var(--radius-sm);
    margin-bottom: 2px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .git-success { color: var(--accent-green); background: rgba(0, 229, 122, 0.08); }
  .git-error { color: var(--accent-red); background: rgba(232, 69, 69, 0.08); }

  .worktree-mini {
    padding: 6px 8px;
    border-top: 1px solid var(--border-subtle);
    display: flex;
    align-items: center;
    gap: 4px;
    font-size: 10px;
  }

  .worktree-label {
    color: var(--accent-teal);
    font-weight: 600;
  }

  .worktree-branch {
    font-family: var(--font-mono);
    color: var(--text-dim);
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
    font-size: 10px;
    color: var(--text-dim);
  }

  /* Sub-agent header badge */
  .header-sub-name {
    font-family: var(--font-mono);
    font-size: 12px;
    color: var(--accent-teal);
  }

  .sub-agent-badge {
    font-family: var(--font-mono);
    font-size: 10px;
    padding: 1px 8px;
    border-radius: 8px;
    background: rgba(0, 196, 179, 0.12);
    color: var(--accent-teal);
    border: 1px solid rgba(0, 196, 179, 0.25);
  }

  /* Sub-agent info panel */
  .sub-agent-panel {
    flex-shrink: 0;
    border-bottom: 1px solid var(--border-subtle);
    background: var(--bg-surface);
    max-height: 200px;
    overflow-y: auto;
  }

  .sub-panel-desc {
    padding: 6px var(--sp-lg);
    font-size: 12px;
    color: var(--text-dim);
    border-bottom: 1px solid var(--border-subtle);
  }

  .sub-panel-result {
    padding: 6px var(--sp-lg);
    border-bottom: 1px solid var(--border-subtle);
  }

  .sub-panel-label {
    display: block;
    font-family: var(--font-mono);
    font-size: 9px;
    color: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    margin-bottom: 4px;
  }

  .sub-panel-text {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--text-primary);
    white-space: pre-wrap;
    word-break: break-word;
    margin: 0;
    max-height: 80px;
    overflow-y: auto;
  }

  .sub-panel-logs {
    padding: 6px var(--sp-lg);
  }

  .sub-panel-log-list {
    max-height: 100px;
    overflow-y: auto;
  }

  .sub-log-line {
    font-family: var(--font-mono);
    font-size: 10px;
    color: var(--text-dim);
    padding: 1px 0;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .sub-log-line.log-ok { color: var(--accent-green); }
  .sub-log-line.log-err { color: var(--accent-red); }
  .sub-log-line.log-dim { color: var(--text-muted); }
  .sub-log-line.log-system { color: var(--accent-purple); }

  /* Full-width commit output panel — between workspace and bottom bar */
  .ws-commit-panel {
    border-top: 2px solid var(--accent-green);
    background: var(--bg-deepest);
    display: flex;
    flex-direction: column;
    max-height: 240px;
    overflow: hidden;
    flex-shrink: 0;
    width: 100%;
  }

  .ws-commit-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 4px 8px;
    background: var(--bg-surface);
    border-bottom: 1px solid var(--border-subtle);
    flex-shrink: 0;
  }

  .ws-commit-title {
    font-family: var(--font-mono);
    font-size: 10px;
    font-weight: 600;
    color: var(--text-dim);
  }

  .ws-commit-close {
    background: none;
    border: none;
    color: var(--text-muted);
    font-size: 14px;
    cursor: pointer;
    padding: 0 2px;
    line-height: 1;
  }

  .ws-commit-close:hover { color: var(--text-primary); }

  .ws-commit-body {
    padding: 4px 8px;
    overflow-y: auto;
    flex: 1;
    min-height: 0;
  }

  .ws-commit-line {
    padding: 1px 0;
  }

  .ws-commit-step {
    font-family: var(--font-mono);
    font-size: 10px;
    color: var(--accent-green);
  }

  .ws-commit-output {
    font-family: var(--font-mono);
    font-size: 9px;
    color: var(--text-dim);
    margin: 2px 0 3px 0;
    padding: 3px 6px;
    background: rgba(0, 0, 0, 0.3);
    border-radius: var(--radius-sm);
    white-space: pre-wrap;
    word-break: break-word;
    max-height: 50px;
    overflow-y: auto;
  }

  .ws-commit-active {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .ws-commit-spinner {
    display: inline-block;
    width: 7px;
    height: 7px;
    border: 1.5px solid var(--accent-green);
    border-top-color: transparent;
    border-radius: 50%;
    animation: ws-commit-spin 0.6s linear infinite;
  }

  @keyframes ws-commit-spin {
    to { transform: rotate(360deg); }
  }

  .ws-commit-error-section {
    border-top: 1px solid rgba(232, 69, 69, 0.2);
    padding: 4px 8px;
    background: rgba(232, 69, 69, 0.04);
    flex-shrink: 0;
  }

  .ws-commit-error-label {
    font-family: var(--font-mono);
    font-size: 8px;
    font-weight: 600;
    color: var(--accent-red);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    margin-bottom: 3px;
  }

  .ws-commit-error-text {
    font-family: var(--font-mono);
    font-size: 9px;
    color: var(--accent-red);
    white-space: pre-wrap;
    word-break: break-word;
    margin: 0 0 6px 0;
    padding: 3px 6px;
    background: rgba(232, 69, 69, 0.06);
    border-radius: var(--radius-sm);
    border: 1px solid rgba(232, 69, 69, 0.15);
    max-height: 50px;
    overflow-y: auto;
  }

  .ws-commit-explain-label {
    font-family: var(--font-mono);
    font-size: 8px;
    font-weight: 600;
    color: var(--accent-amber);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    margin-bottom: 3px;
  }

  .ws-commit-explain-text {
    font-size: 10px;
    color: var(--text-primary);
    line-height: 1.4;
    padding: 4px 6px;
    background: rgba(240, 165, 0, 0.06);
    border-radius: var(--radius-sm);
    border: 1px solid rgba(240, 165, 0, 0.15);
  }

  .ws-commit-explain-loading {
    font-family: var(--font-mono);
    font-size: 9px;
    color: var(--text-muted);
    font-style: italic;
  }
</style>
