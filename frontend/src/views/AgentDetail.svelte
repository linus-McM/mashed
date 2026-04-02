<script>
  import { onMount, onDestroy, createEventDispatcher } from 'svelte';
  import { GetScopedDiff, GetWorktrees, ListRepoFiles } from '../../wailsjs/go/main/App.js';
  import { ArrowLeft, GitBranch, GripVertical } from 'lucide-svelte';
  import StatusBadge from '../components/StatusBadge.svelte';
  import Terminal from '../components/Terminal.svelte';
  import CodeEditor from '../components/CodeEditor.svelte';

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

  $: tokenPct = agent ? Math.min(((agent.tokensUsed || 0) / (agent.tokensMax || 1)) * 100, 100) : 0;
  $: tokenLabel = agent ? formatTokens(agent.tokensUsed || 0) + ' / ' + formatTokens(agent.tokensMax || 0) : '';

  function formatTokens(n) {
    if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + 'M';
    if (n >= 1_000) return (n / 1_000).toFixed(1) + 'K';
    return String(n);
  }

  $: filteredAllFiles = allFilesSearch
    ? allFiles.filter(f => f.toLowerCase().includes(allFilesSearch.toLowerCase()))
    : allFiles;

  function selectFile(file) {
    if (selectedFile?.path === file.path) {
      selectedFile = null; // toggle off
    } else {
      selectedFile = file;
    }
  }

  function selectAllFile(filePath) {
    if (selectedFile?.path === filePath) {
      selectedFile = null;
    } else {
      selectedFile = { path: filePath, isBinary: false };
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

  onMount(async () => {
    window.addEventListener('keydown', handleKeydown);

    if (agent?.repoPath) {
      try {
        const diff = await GetScopedDiff(agent.repoPath);
        if (diff && diff.files) changedFiles = diff.files;
      } catch (e) {
        console.warn('Failed to get scoped diff:', e);
      }

      try {
        const wts = await GetWorktrees(agent.repoPath);
        if (wts && wts.length > 0) worktree = wts[0];
      } catch (e) {
        console.warn('Failed to get worktrees:', e);
      }

      try {
        const files = await ListRepoFiles(agent.repoPath);
        if (files) allFiles = files;
      } catch (e) {
        console.warn('Failed to list repo files:', e);
      }
    }
  });

  onDestroy(() => {
    window.removeEventListener('keydown', handleKeydown);
  });
</script>

<div class="detail">
  <!-- Compact header bar -->
  <div class="header">
    <button class="back-btn" on:click={() => dispatch('back')}><ArrowLeft size={14} /> Back</button>
    <span class="header-repo">{agent.repoName}</span>
    <span class="header-sep">/</span>
    <span class="header-agent">{agent.model || agent.agentName}</span>
    <StatusBadge status={agent.eventType} size="sm" />
    {#if agent.repoBranch}
      <span class="header-branch"><GitBranch size={12} /> {agent.repoBranch}</span>
    {/if}
    <div class="header-right">
      <div class="token-mini">
        <div class="token-mini-bar">
          <div class="token-mini-fill" style="width: {tokenPct}%" />
        </div>
        <span class="token-mini-label">{tokenLabel}</span>
      </div>
    </div>
  </div>

  <!-- Main workspace -->
  <div class="workspace" bind:this={workspaceEl}>
    <!-- Terminal pane -->
    <div
      class="terminal-pane"
      style={selectedFile ? `flex: 0 0 calc(100% - ${fileStripWidth}px - ${editorFraction * 100}%)` : `flex: 1 1 0; width: 0`}
    >
      <Terminal paneTarget={agent?.tmuxTarget || ''} repoPath={agent?.repoPath || ''} />
    </div>

    <!-- Resize handle: terminal | file strip -->
    <div class="resize-handle" on:mousedown={onMouseDown('file-strip')}>
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
        <div class="file-search">
          <input
            type="text"
            placeholder="Filter files..."
            bind:value={allFilesSearch}
            class="file-search-input"
          />
        </div>
        <div class="file-strip-list">
          {#each filteredAllFiles as filePath}
            <button
              class="file-item"
              class:active={selectedFile?.path === filePath}
              on:click={() => selectAllFile(filePath)}
              title={filePath}
            >
              <span class="file-name">{filePath.split('/').pop()}</span>
              <span class="file-dir">{filePath.includes('/') ? filePath.substring(0, filePath.lastIndexOf('/')) : ''}</span>
            </button>
          {/each}
          {#if filteredAllFiles.length === 0}
            <div class="file-strip-empty">No files found</div>
          {/if}
        </div>
      {/if}

      {#if worktree}
        <div class="worktree-mini">
          <span class="worktree-label">WT</span>
          <span class="worktree-branch">{worktree.branch}</span>
        </div>
      {/if}
    </div>

    <!-- Resize handle: file strip | editor -->
    {#if selectedFile}
      <div class="resize-handle" on:mousedown={onMouseDown('editor')}>
        <div class="resize-grip"><GripVertical size={10} /></div>
      </div>
    {/if}

    <!-- Code editor (visible when file selected) -->
    {#if selectedFile}
      <div class="editor-pane" style="flex: 0 0 {editorFraction * 100}%">
        <CodeEditor
          filePath={selectedFile.path}
          repoPath={agent.repoPath}
          mode={selectedFile.isBinary ? 'source' : (fileTab === 'all' ? 'source' : 'diff')}
          editable={!selectedFile.isBinary}
        />
      </div>
    {/if}
  </div>

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

  .file-search {
    padding: 4px 6px;
    border-bottom: 1px solid var(--border-subtle);
    flex-shrink: 0;
  }

  .file-search-input {
    width: 100%;
    padding: 3px 6px;
    background: var(--bg-deepest);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-sm);
    color: var(--text-primary);
    font-family: var(--font-mono);
    font-size: 11px;
    outline: none;
  }

  .file-search-input:focus {
    border-color: var(--accent-green);
  }

  .file-search-input::placeholder {
    color: var(--text-muted);
  }

  .file-strip-empty {
    padding: 12px 8px;
    font-size: 11px;
    color: var(--text-muted);
    text-align: center;
  }

  .file-dir {
    font-size: 9px;
    color: var(--text-muted);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    flex-shrink: 1;
    min-width: 0;
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
</style>
