<script>
  import { onMount, onDestroy, createEventDispatcher } from 'svelte';
  import { GetScopedDiff, GetWorktrees } from '../../wailsjs/go/main/App.js';
  import StatusBadge from '../components/StatusBadge.svelte';
  import Terminal from '../components/Terminal.svelte';
  import CodeEditor from '../components/CodeEditor.svelte';

  const dispatch = createEventDispatcher();

  export let agent;

  let changedFiles = [];
  let worktree = null;
  let selectedFile = null; // when set, shows code editor on right

  $: tokenPct = agent ? Math.min(((agent.tokensUsed || 0) / (agent.tokensMax || 1)) * 100, 100) : 0;
  $: tokenLabel = agent ? formatTokens(agent.tokensUsed || 0) + ' / ' + formatTokens(agent.tokensMax || 0) : '';

  function formatTokens(n) {
    if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + 'M';
    if (n >= 1_000) return (n / 1_000).toFixed(1) + 'K';
    return String(n);
  }

  function selectFile(file) {
    if (selectedFile?.path === file.path) {
      selectedFile = null; // toggle off
    } else {
      selectedFile = file;
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
    }
  });

  onDestroy(() => {
    window.removeEventListener('keydown', handleKeydown);
  });
</script>

<div class="detail">
  <!-- Compact header bar -->
  <div class="header">
    <button class="back-btn" on:click={() => dispatch('back')}>&#8592; Back</button>
    <span class="header-repo">{agent.repoName}</span>
    <span class="header-sep">/</span>
    <span class="header-agent">{agent.model || agent.agentName}</span>
    <StatusBadge status={agent.eventType} size="sm" />
    {#if agent.repoBranch}
      <span class="header-branch">⎇ {agent.repoBranch}</span>
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
  <div class="workspace">
    <!-- Left half: Terminal (always visible) -->
    <div class="terminal-pane" class:half={selectedFile}>
      <Terminal paneTarget={agent?.tmuxTarget || ''} repoPath={agent?.repoPath || ''} />
    </div>

    <!-- File list strip (between terminal and editor) -->
    <div class="file-strip" class:visible={changedFiles.length > 0}>
      <div class="file-strip-header">
        <span class="file-strip-title">FILES</span>
        <span class="file-strip-count">{changedFiles.length}</span>
      </div>
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
      </div>
      {#if worktree}
        <div class="worktree-mini">
          <span class="worktree-label">WT</span>
          <span class="worktree-branch">{worktree.branch}</span>
        </div>
      {/if}
    </div>

    <!-- Right half: Code editor (visible when file selected) -->
    {#if selectedFile}
      <div class="editor-pane">
        <CodeEditor
          filePath={selectedFile.path}
          repoPath={agent.repoPath}
          mode={selectedFile.isBinary ? 'source' : 'diff'}
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
    color: var(--text-dim);
    font-family: var(--font-ui);
    font-size: 13px;
    cursor: pointer;
    padding: 2px 6px;
    border-radius: var(--radius-sm);
  }

  .back-btn:hover { color: var(--accent-green); }

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
  }

  .terminal-pane {
    flex: 1;
    display: flex;
    flex-direction: column;
    min-width: 0;
    transition: flex 150ms ease-out;
  }

  .terminal-pane.half {
    flex: 0 0 50%;
  }

  /* File strip */
  .file-strip {
    width: 0;
    overflow: hidden;
    border-left: 1px solid var(--border-subtle);
    background: var(--bg-surface);
    display: flex;
    flex-direction: column;
    transition: width 150ms ease-out;
  }

  .file-strip.visible {
    width: 180px;
  }

  .file-strip-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 6px 8px;
    border-bottom: 1px solid var(--border-subtle);
  }

  .file-strip-title {
    font-size: 10px;
    font-weight: 600;
    color: var(--text-dim);
    letter-spacing: 0.06em;
  }

  .file-strip-count {
    font-family: var(--font-mono);
    font-size: 10px;
    color: var(--text-muted);
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
    flex: 0 0 50%;
    display: flex;
    flex-direction: column;
    border-left: 1px solid var(--border-subtle);
    min-width: 0;
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
