<script lang="ts">
  // Right-hand actions sidebar of a repo group (branch switch + git action
  // buttons + transient result/error). Extracted from
  // views/NotificationFeed.svelte (R31). The parent owns repo status/action/
  // commit-panel state and runs the actions; this component reads that state
  // and bubbles button intents.
  import { createEventDispatcher } from 'svelte';
  import { GitBranch, GitCommit as GitCommitIcon, Upload, GitPullRequest, ShieldAlert, GitBranchPlus, Download, GitMerge } from 'lucide-svelte';
  import { REPO_BORDER_NONE } from '../../lib/repoPalette';
  import {
    IDLE_ACTION,
    type CommitPanel,
    type RepoAction,
    type RepoGroup,
    type RepoStatusSnapshot,
  } from '../../lib/feed/repoTree';

  export let repo: RepoGroup;
  /** Current border colour for this repo (parent's getRepoColor(repo.name)). */
  export let repoColor: string;
  export let repoStatuses: Record<string, RepoStatusSnapshot>;
  export let repoActions: Record<string, RepoAction>;
  export let commitPanels: Record<string, CommitPanel>;

  const dispatch = createEventDispatcher<{
    switch: void;
    branch: void;
    commit: void;
    pull: void;
    push: void;
    merge: void;
    pr: void;
    review: void;
  }>();

  function isDirty(path: string): boolean {
    return repoStatuses[path]?.dirty || false;
  }

  function hasOpenPR(path: string): boolean {
    return (repoStatuses[path]?.openPRs || 0) > 0;
  }

  function isAhead(path: string): boolean {
    return (repoStatuses[path]?.ahead || 0) > 0;
  }

  function isProtected(path: string): boolean {
    return repoStatuses[path]?.protected || false;
  }

  function getAction(path: string): RepoAction {
    return repoActions[path] || IDLE_ACTION;
  }

  function getCommitPanel(path: string): CommitPanel | null {
    return commitPanels[path] || null;
  }
</script>

<!-- Right: Actions sidebar (25%) -->
<div class="repo-actions">
  <button
    class="actions-branch"
    style="color: {repoColor !== REPO_BORDER_NONE ? repoColor : 'var(--text-dim)'}"
    on:click|stopPropagation={() => dispatch('switch')}
    title="Switch branch"
  >
    <GitBranch size={12} />
    <span class="actions-branch-name">{repo.branch || 'detached'}</span>
  </button>

  <div class="actions-buttons">
    <button
      class="action-btn"
      on:click|stopPropagation={() => dispatch('branch')}
      title="Create a new branch"
    >
      <GitBranchPlus size={14} />
      <span>Branch</span>
    </button>
    <button
      class="action-btn"
      class:glow-btn={isDirty(repo.path)}
      disabled={!!getAction(repo.path).action}
      on:click|stopPropagation={() => dispatch('commit')}
      title="Stage all + AI commit message + commit"
    >
      <GitCommitIcon size={14} />
      <span>{getAction(repo.path).action === 'commit' ? 'Committing...' : 'Commit'}</span>
    </button>
    <button
      class="action-btn"
      disabled={!!getAction(repo.path).action}
      on:click|stopPropagation={() => dispatch('pull')}
      title="Pull remote changes"
    >
      <Download size={14} />
      <span>{getAction(repo.path).action === 'pull' ? 'Pulling...' : 'Pull'}</span>
    </button>
    <button
      class="action-btn"
      class:glow-btn={isAhead(repo.path) && !isProtected(repo.path)}
      disabled={!!getAction(repo.path).action}
      on:click|stopPropagation={() => dispatch('push')}
      title={isProtected(repo.path) ? 'Branch is protected — push via PR' : isAhead(repo.path) ? `${repoStatuses[repo.path]?.ahead} commit(s) ahead of remote` : 'Push to origin'}
    >
      <Upload size={14} />
      <span>{getAction(repo.path).action === 'push' ? 'Pushing...' : 'Push'}</span>
    </button>
    <button
      class="action-btn"
      disabled={!!getAction(repo.path).action}
      on:click|stopPropagation={() => dispatch('merge')}
      title="Merge current branch into another"
    >
      <GitMerge size={14} />
      <span>Merge</span>
    </button>
    <button
      class="action-btn"
      disabled={!!getAction(repo.path).action}
      on:click|stopPropagation={() => dispatch('pr')}
      title="Commit + push + create PR"
    >
      <GitPullRequest size={14} />
      <span>{getAction(repo.path).action === 'pr' ? 'Creating PR...' : 'PR'}</span>
    </button>
    <button
      class="action-btn action-review"
      class:glow-btn={hasOpenPR(repo.path)}
      disabled={!!getAction(repo.path).action}
      on:click|stopPropagation={() => dispatch('review')}
      title="Spawn adversarial PR review agent"
    >
      <ShieldAlert size={14} />
      <span>{getAction(repo.path).action === 'review' ? 'Spawning...' : 'Review'}</span>
    </button>
  </div>

  {#if !getCommitPanel(repo.path)}
    {#if getAction(repo.path).result}
      <div class="action-result">{getAction(repo.path).result}</div>
    {/if}
    {#if getAction(repo.path).error && !getCommitPanel(repo.path)}
      <div class="action-error">{getAction(repo.path).error}</div>
    {/if}
  {/if}
</div>

<style>
  .repo-actions {
    flex: 1;
    display: flex;
    flex-direction: column;
    padding: var(--sp-sm);
    gap: var(--sp-xs);
    min-width: 140px;
    max-width: 200px;
  }

  .actions-branch {
    display: flex;
    align-items: center;
    gap: var(--sp-xs);
    color: var(--text-dim);
    font-family: var(--font-mono);
    font-size: var(--text-label);
    padding: var(--sp-2xs) 0 var(--sp-xs);
    border: none;
    border-bottom: 1px solid var(--border-subtle);
    margin-bottom: var(--sp-2xs);
    background: none;
    cursor: pointer;
    transition: opacity 100ms ease;
    width: 100%;
    text-align: left;
  }

  .actions-branch:hover { opacity: 0.8; }

  .actions-branch-name {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .actions-buttons {
    display: flex;
    flex-direction: column;
    gap: var(--sp-2xs);
  }

  .action-btn {
    display: flex;
    align-items: center;
    gap: var(--sp-sm);
    width: 100%;
    padding: 4px 8px;
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-sm);
    color: var(--text-dim);
    font-family: var(--font-mono);
    font-size: 11px;
    cursor: pointer;
    transition: all 100ms ease;
  }

  .action-btn:hover {
    color: var(--text-primary);
    border-color: var(--border-emphasis);
    background: var(--bg-active);
  }

  .action-btn:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }


  .action-result {
    font-family: var(--font-mono);
    font-size: var(--text-label);
    color: var(--accent-green);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    padding: 2px 0;
  }

  .action-error {
    font-family: var(--font-mono);
    font-size: var(--text-label);
    color: var(--accent-red);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    padding: 2px 0;
  }
</style>
