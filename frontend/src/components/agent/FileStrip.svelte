<script lang="ts">
  // File list strip (Changed / All Files tabs) plus git action buttons and
  // worktree badge. Extracted from views/AgentDetail.svelte (spec R31) —
  // markup and scoped styles moved verbatim. The parent owns git/refresh
  // state and the async actions (they are shared with the commit listener and
  // modals); buttons call back through the on* props. Selection, tab and
  // strip width are shared with the parent via bind:.
  import { GitCommit as GitCommitIcon, Upload, GitPullRequest, ShieldAlert, GitBranchPlus, Download, GitMerge } from 'lucide-svelte';
  import FileTree from '../FileTree.svelte';
  import type { WorktreeInfo } from '../../lib/types/wails';

  type RepoFile = { path: string; isBinary?: boolean; added?: number; removed?: number };
  type GitActionName = 'commit' | 'push' | 'pull' | 'pr' | 'review';

  export let fileStripWidth: number;
  export let fileTab: string; // 'changed' or 'all'
  export let selectedFile: RepoFile | null;
  export let changedFiles: RepoFile[];
  export let allFiles: string[];
  export let worktree: WorktreeInfo | null;
  export let gitAction: GitActionName | null;
  export let gitResult: string | null;
  export let gitError: string | null;
  /** True while the full-width commit output panel is open. */
  export let commitPanelOpen: boolean;
  export let repoStatus: { ahead: number; behind: number; protected: boolean };
  export let onBranch: () => void;
  export let onCommit: () => void;
  export let onPull: () => void;
  export let onPush: () => void;
  export let onMerge: () => void;
  export let onPR: () => void;
  export let onReview: () => void;

  let allFilesSearch = ''; // search filter for all files tab

  $: filteredAllFiles = allFilesSearch
    ? allFiles.filter(f => f.toLowerCase().includes(allFilesSearch.toLowerCase()))
    : allFiles;

  function selectFile(file: RepoFile) {
    if (selectedFile?.path === file.path) {
      selectedFile = null; // toggle off
    } else {
      selectedFile = file;
      fileStripWidth = 350;
    }
  }

  function selectAllFile(filePath: string) {
    if (selectedFile?.path === filePath) {
      selectedFile = null;
    } else {
      selectedFile = { path: filePath, isBinary: false };
      fileStripWidth = 350;
    }
  }
</script>

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
              {#if (file.added ?? 0) > 0}<span class="added">+{file.added}</span>{/if}
              {#if (file.removed ?? 0) > 0}<span class="removed">-{file.removed}</span>{/if}
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
    {#if !commitPanelOpen}
      {#if gitResult}
        <div class="git-status git-success">{gitResult}</div>
      {:else if gitError}
        <div class="git-status git-error">{gitError}</div>
      {/if}
    {/if}
    <button class="git-btn" disabled={!!gitAction} on:click={() => onBranch()}>
      <GitBranchPlus size={14} /> Branch
    </button>
    <button class="git-btn" class:glow-btn={changedFiles.length > 0} disabled={!!gitAction} on:click={() => onCommit()}>
      <GitCommitIcon size={14} /> {gitAction === 'commit' ? 'Committing...' : 'Commit'}
    </button>
    <button class="git-btn" disabled={!!gitAction} on:click={() => onPull()}>
      <Download size={14} /> {gitAction === 'pull' ? 'Pulling...' : 'Pull'}
    </button>
    <button
      class="git-btn"
      class:glow-btn={(repoStatus.ahead || 0) > 0 && !repoStatus.protected}
      disabled={!!gitAction}
      on:click={() => onPush()}
      title={repoStatus.protected ? 'Branch is protected — push via PR' : (repoStatus.ahead || 0) > 0 ? `${repoStatus.ahead} commit(s) ahead of remote` : 'Push to origin'}
    >
      <Upload size={14} /> {gitAction === 'push' ? 'Pushing...' : 'Push'}
    </button>
    <button class="git-btn" disabled={!!gitAction} on:click={() => onMerge()}>
      <GitMerge size={14} /> Merge
    </button>
    <button class="git-btn" class:glow-btn={changedFiles.length > 0} disabled={!!gitAction} on:click={() => onPR()}>
      <GitPullRequest size={14} /> {gitAction === 'pr' ? 'Creating...' : 'PR'}
    </button>
    <button class="git-btn" disabled={!!gitAction} on:click={() => onReview()}>
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

<style>
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
    padding: var(--sp-xs) var(--sp-sm);
    background: none;
    border: none;
    border-bottom: 2px solid transparent;
    font-size: var(--text-label);
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
    padding: var(--sp-2xs) var(--sp-sm);
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
    font-size: var(--text-label);
  }

  .file-stat .added { color: var(--accent-green); }
  .file-stat .removed { color: var(--accent-red); margin-left: 2px; }
  .file-stat.binary { color: var(--text-muted); font-style: italic; }

  /* Git actions */
  .git-actions {
    padding: var(--sp-xs) var(--sp-sm);
    border-top: 1px solid var(--border-subtle);
    display: flex;
    flex-direction: column;
    gap: 2px;
    flex-shrink: 0;
  }

  .git-btn {
    display: flex;
    align-items: center;
    gap: var(--sp-sm);
    width: 100%;
    padding: var(--sp-xs) var(--sp-sm);
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

  .git-status {
    font-family: var(--font-mono);
    font-size: var(--text-label);
    padding: var(--sp-2xs) var(--sp-xs);
    border-radius: var(--radius-sm);
    margin-bottom: 2px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .git-success { color: var(--accent-green); background: color-mix(in srgb, var(--accent-green) 8%, transparent); }
  .git-error { color: var(--accent-red); background: color-mix(in srgb, var(--accent-red) 8%, transparent); }

  .worktree-mini {
    padding: var(--sp-xs) var(--sp-sm);
    border-top: 1px solid var(--border-subtle);
    display: flex;
    align-items: center;
    gap: var(--sp-xs);
    font-size: var(--text-label);
  }

  .worktree-label {
    color: var(--accent-teal);
    font-weight: 600;
  }

  .worktree-branch {
    font-family: var(--font-mono);
    color: var(--text-dim);
  }
</style>
