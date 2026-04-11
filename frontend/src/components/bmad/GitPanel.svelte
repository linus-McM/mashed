<script>
  import { onMount, onDestroy, createEventDispatcher } from 'svelte';
  import { EventsOn, EventsOff } from '../../../wailsjs/runtime/runtime.js';
  import { GitCommitStreaming, GitPull, GitPush, GitCommitPushAndPR, SpawnPRReview, RepoStatus } from '../../../wailsjs/go/main/App.js';
  import { GitBranch, GitCommit as GitCommitIcon, Upload, Download, GitMerge, GitPullRequest, ShieldAlert, GitBranchPlus, AlertCircle, CheckCircle2, FileSearch } from 'lucide-svelte';
  import BranchModal from '../../views/BranchModal.svelte';
  import SwitchBranchModal from '../../views/SwitchBranchModal.svelte';
  import MergeModal from '../../views/MergeModal.svelte';
  import ForcePushModal from '../../views/ForcePushModal.svelte';
  import SummarisationModal from '../../views/SummarisationModal.svelte';

  export let repoPath = '';
  export let repoBranch = '';

  const dispatch = createEventDispatcher();

  // Repo status
  let status = { dirty: false, openPRs: 0, ahead: 0, behind: 0, protected: false };
  let statusInterval;

  async function refreshStatus() {
    if (!repoPath) return;
    try {
      status = await RepoStatus(repoPath);
    } catch (_) {}
  }

  onMount(() => {
    refreshStatus();
    statusInterval = setInterval(refreshStatus, 10000);
  });

  onDestroy(() => {
    if (statusInterval) clearInterval(statusInterval);
    EventsOff('git:commit:progress');
  });

  $: if (repoPath) refreshStatus();

  // Action states
  let actionState = { action: null, result: null, error: null };

  function clearActionAfterDelay() {
    setTimeout(() => {
      if (!actionState.action) {
        actionState = { action: null, result: null, error: null };
      }
    }, 5000);
  }

  async function runAction(actionName, fn) {
    actionState = { action: actionName, result: null, error: null };
    try {
      const result = await fn(repoPath);
      actionState = { action: null, result: result || 'Done', error: null };
    } catch (err) {
      actionState = { action: null, result: null, error: err?.message || String(err) };
    }
    refreshStatus();
    clearActionAfterDelay();
  }

  // Push with conflict detection
  let forcePushRepo = null;

  async function smartPush() {
    actionState = { action: 'push', result: null, error: null };
    try {
      const result = await GitPush(repoPath);
      if (result.startsWith('conflict:')) {
        forcePushRepo = { path: repoPath, message: result.slice('conflict:'.length) };
        actionState = { action: null, result: null, error: null };
      } else {
        actionState = { action: null, result: 'Pushed', error: null };
        clearActionAfterDelay();
      }
    } catch (err) {
      actionState = { action: null, result: null, error: err?.message || String(err) };
      clearActionAfterDelay();
    }
    refreshStatus();
  }

  function onForcePushed() {
    actionState = { action: null, result: 'Force pushed', error: null };
    forcePushRepo = null;
    refreshStatus();
    clearActionAfterDelay();
  }

  // Streaming commit
  let commitPanel = null;

  function startStreamingCommit() {
    commitPanel = { lines: [], error: null, explanation: null, done: false };
    actionState = { action: 'commit', result: null, error: null };
    GitCommitStreaming(repoPath);
  }

  function closeCommitPanel() {
    commitPanel = null;
  }

  EventsOn('git:commit:progress', (evt) => {
    if (evt.repoPath !== repoPath) return;
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

    if (evt.done && !evt.error) {
      actionState = { action: null, result: evt.output || 'Committed', error: null };
      refreshStatus();
      setTimeout(() => { commitPanel = null; }, 1500);
    } else if (evt.done && evt.error) {
      actionState = { action: null, result: null, error: evt.error };
    }

    commitPanel = commitPanel;
  });

  // Modal state
  let branchModalRepo = null;
  let switchModalRepo = null;
  let mergeModalRepo = null;
  let summariseModalOpen = false;

  function onBranchCreated(e) {
    const newBranch = e.detail?.branch;
    if (newBranch) {
      repoBranch = newBranch;
      dispatch('branch-changed', { branch: newBranch });
    }
    branchModalRepo = null;
    refreshStatus();
  }

  function onBranchSwitched(e) {
    const newBranch = e.detail?.branch;
    if (newBranch) {
      repoBranch = newBranch;
      dispatch('branch-changed', { branch: newBranch });
    }
    switchModalRepo = null;
    refreshStatus();
  }

  function onMerged() {
    mergeModalRepo = null;
    refreshStatus();
  }
</script>

<div class="git-panel">
  <!-- Branch header -->
  <button
    class="branch-header"
    on:click={() => switchModalRepo = { path: repoPath, branch: repoBranch, color: 'var(--accent-green)' }}
    title="Switch branch"
  >
    <GitBranch size={14} />
    <span class="branch-name">{repoBranch || 'detached'}</span>
  </button>

  <!-- Status badges -->
  <div class="status-row">
    {#if status.dirty}
      <span class="status-badge dirty">
        <AlertCircle size={10} /> dirty
      </span>
    {:else}
      <span class="status-badge clean">
        <CheckCircle2 size={10} /> clean
      </span>
    {/if}
    {#if status.ahead > 0}
      <span class="status-badge ahead">{status.ahead} ahead</span>
    {/if}
    {#if status.behind > 0}
      <span class="status-badge behind">{status.behind} behind</span>
    {/if}
    {#if status.openPRs > 0}
      <span class="status-badge pr">{status.openPRs} PR{status.openPRs > 1 ? 's' : ''}</span>
    {/if}
    {#if status.protected}
      <span class="status-badge protected">protected</span>
    {/if}
  </div>

  <!-- Action buttons -->
  <div class="actions">
    <button
      class="action-btn"
      on:click={() => branchModalRepo = { path: repoPath, branch: repoBranch }}
      title="Create a new branch"
    >
      <GitBranchPlus size={14} />
      <span>Branch</span>
    </button>
    <button
      class="action-btn"
      class:hot={status.dirty}
      disabled={!!actionState.action}
      on:click={startStreamingCommit}
      title="Stage all + AI commit message + commit"
    >
      <GitCommitIcon size={14} />
      <span>{actionState.action === 'commit' ? 'Committing...' : 'Commit'}</span>
    </button>
    <button
      class="action-btn"
      disabled={!!actionState.action}
      on:click={() => runAction('pull', GitPull)}
      title="Pull remote changes"
    >
      <Download size={14} />
      <span>{actionState.action === 'pull' ? 'Pulling...' : 'Pull'}</span>
    </button>
    <button
      class="action-btn"
      class:hot={status.ahead > 0 && !status.protected}
      disabled={!!actionState.action}
      on:click={smartPush}
      title={status.protected ? 'Branch is protected — push via PR' : status.ahead > 0 ? `${status.ahead} commit(s) ahead` : 'Push to origin'}
    >
      <Upload size={14} />
      <span>{actionState.action === 'push' ? 'Pushing...' : 'Push'}</span>
    </button>
    <button
      class="action-btn"
      disabled={!!actionState.action}
      on:click={() => mergeModalRepo = { path: repoPath, branch: repoBranch }}
      title="Merge current branch into another"
    >
      <GitMerge size={14} />
      <span>Merge</span>
    </button>
    <button
      class="action-btn"
      disabled={!!actionState.action}
      on:click={() => runAction('pr', GitCommitPushAndPR)}
      title="Commit + push + create PR"
    >
      <GitPullRequest size={14} />
      <span>{actionState.action === 'pr' ? 'Creating PR...' : 'PR'}</span>
    </button>
    <button
      class="action-btn review"
      class:hot={status.openPRs > 0}
      disabled={!!actionState.action}
      on:click={() => runAction('review', SpawnPRReview)}
      title="Spawn adversarial PR review agent"
    >
      <ShieldAlert size={14} />
      <span>{actionState.action === 'review' ? 'Spawning...' : 'Review'}</span>
    </button>
    <button
      class="action-btn"
      class:hot={status.dirty}
      disabled={!!actionState.action}
      on:click={() => summariseModalOpen = true}
      title="AI-powered code review summary"
    >
      <FileSearch size={14} />
      <span>Summarise</span>
    </button>
  </div>

  <!-- Action result / error -->
  {#if !commitPanel}
    {#if actionState.result}
      <div class="action-result">{actionState.result}</div>
    {/if}
    {#if actionState.error}
      <div class="action-error">{actionState.error}</div>
    {/if}
  {/if}

  <!-- Streaming commit panel -->
  {#if commitPanel}
    <div class="commit-panel">
      <div class="commit-panel-header">
        <span class="commit-panel-title">
          {#if commitPanel.done && !commitPanel.error}
            Committed
          {:else if commitPanel.error}
            Commit Failed
          {:else}
            Committing...
          {/if}
        </span>
        {#if commitPanel.done}
          <button class="commit-panel-close" on:click={closeCommitPanel}>&times;</button>
        {/if}
      </div>
      <div class="commit-panel-body">
        {#each commitPanel.lines as line}
          <div class="commit-line">
            <span class="commit-step">{line.step}</span>
            {#if line.output}
              <pre class="commit-output">{line.output}</pre>
            {/if}
          </div>
        {/each}
        {#if !commitPanel.done && !commitPanel.error}
          <div class="commit-line commit-active">
            <span class="commit-spinner" />
          </div>
        {/if}
      </div>
      {#if commitPanel.error}
        <div class="commit-error-section">
          <div class="commit-error-label">Error</div>
          <pre class="commit-error-text">{commitPanel.error}</pre>
          {#if commitPanel.explanation}
            <div class="commit-explain-label">Why this happened</div>
            <div class="commit-explain-text">{commitPanel.explanation}</div>
          {:else if !commitPanel.done}
            <div class="commit-explain-loading">Analyzing failure...</div>
          {/if}
        </div>
      {/if}
    </div>
  {/if}
</div>

<!-- Modals -->
{#if branchModalRepo}
  <BranchModal
    repoPath={branchModalRepo.path}
    repoBranch={branchModalRepo.branch}
    on:created={onBranchCreated}
    on:cancel={() => branchModalRepo = null}
  />
{/if}

{#if switchModalRepo}
  <SwitchBranchModal
    repoPath={switchModalRepo.path}
    currentBranch={switchModalRepo.branch}
    repoColor={switchModalRepo.color}
    on:switched={onBranchSwitched}
    on:cancel={() => switchModalRepo = null}
  />
{/if}

{#if mergeModalRepo}
  <MergeModal
    repoPath={mergeModalRepo.path}
    currentBranch={mergeModalRepo.branch}
    on:merged={onMerged}
    on:cancel={() => mergeModalRepo = null}
  />
{/if}

{#if forcePushRepo}
  <ForcePushModal
    repoPath={forcePushRepo.path}
    conflictMessage={forcePushRepo.message}
    on:pushed={onForcePushed}
    on:cancel={() => forcePushRepo = null}
  />
{/if}

{#if summariseModalOpen}
  <SummarisationModal
    {repoPath}
    on:close={() => summariseModalOpen = false}
    on:open-file
  />
{/if}

<style>
  .git-panel {
    display: flex;
    flex-direction: column;
    gap: var(--sp-md);
    padding: var(--sp-md);
    height: 100%;
    overflow-y: auto;
  }

  .branch-header {
    display: flex;
    align-items: center;
    gap: var(--sp-sm);
    padding: 8px 12px;
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    color: var(--accent-green);
    font-family: var(--font-mono);
    font-size: 13px;
    font-weight: 600;
    cursor: pointer;
    transition: all 100ms ease;
  }

  .branch-header:hover {
    border-color: var(--accent-green);
    background: var(--bg-active);
  }

  .branch-name {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  /* Status badges */
  .status-row {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
  }

  .status-badge {
    display: inline-flex;
    align-items: center;
    gap: var(--sp-2xs);
    padding: 2px 8px;
    font-family: var(--font-mono);
    font-size: var(--text-label);
    font-weight: 500;
    border-radius: 10px;
    border: 1px solid var(--border-subtle);
    background: var(--bg-elevated);
    color: var(--text-dim);
  }

  .status-badge.dirty { color: var(--accent-amber); border-color: var(--accent-amber); }
  .status-badge.clean { color: var(--accent-green); border-color: var(--accent-green); opacity: 0.6; }
  .status-badge.ahead { color: var(--accent-blue); border-color: var(--accent-blue); }
  .status-badge.behind { color: var(--accent-purple); border-color: var(--accent-purple); }
  .status-badge.pr { color: var(--accent-teal); border-color: var(--accent-teal); }
  .status-badge.protected { color: var(--accent-red); border-color: var(--accent-red); }

  /* Action buttons */
  .actions {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .action-btn {
    display: flex;
    align-items: center;
    gap: var(--sp-sm);
    padding: 8px 12px;
    background: none;
    border: 1px solid transparent;
    border-radius: var(--radius-md);
    color: var(--text-dim);
    font-family: var(--font-mono);
    font-size: var(--text-body);
    cursor: pointer;
    transition: all 100ms ease;
    text-align: left;
  }

  .action-btn:hover:not(:disabled) {
    background: var(--bg-elevated);
    border-color: var(--border-subtle);
    color: var(--text-primary);
  }

  .action-btn:disabled {
    opacity: 0.4;
    cursor: default;
  }

  .action-btn.hot {
    color: var(--accent-green);
    background: color-mix(in srgb, var(--accent-green) 5%, transparent);
  }

  .action-btn.hot:hover:not(:disabled) {
    background: color-mix(in srgb, var(--accent-green) 10%, transparent);
    border-color: var(--accent-green);
  }

  .action-btn.review.hot {
    color: var(--accent-amber);
    background: rgba(210, 153, 34, 0.05);
  }

  .action-btn.review.hot:hover:not(:disabled) {
    background: rgba(210, 153, 34, 0.1);
    border-color: var(--accent-amber);
  }

  /* Result / error */
  .action-result {
    padding: var(--sp-xs) 10px;
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--accent-green);
    background: color-mix(in srgb, var(--accent-green) 6%, transparent);
    border-radius: var(--radius-sm);
    word-break: break-word;
  }

  .action-error {
    padding: var(--sp-xs) 10px;
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--accent-red);
    background: rgba(255, 95, 87, 0.06);
    border-radius: var(--radius-sm);
    word-break: break-word;
  }

  /* Commit panel */
  .commit-panel {
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    overflow: hidden;
    background: var(--bg-elevated);
  }

  .commit-panel-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: var(--sp-xs) 10px;
    background: var(--bg-surface);
    border-bottom: 1px solid var(--border-subtle);
  }

  .commit-panel-title {
    font-family: var(--font-mono);
    font-size: 11px;
    font-weight: 600;
    color: var(--text-primary);
  }

  .commit-panel-close {
    background: none;
    border: none;
    color: var(--text-muted);
    cursor: pointer;
    font-size: 16px;
    line-height: 1;
    padding: 0 2px;
  }

  .commit-panel-close:hover {
    color: var(--text-primary);
  }

  .commit-panel-body {
    padding: var(--sp-xs) 10px;
    max-height: 200px;
    overflow-y: auto;
  }

  .commit-line {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--text-dim);
    padding: 2px 0;
  }

  .commit-step {
    color: var(--accent-green);
    font-weight: 500;
  }

  .commit-output {
    font-family: var(--font-mono);
    font-size: var(--text-label);
    color: var(--text-muted);
    margin: 2px 0 0;
    white-space: pre-wrap;
    word-break: break-all;
  }

  .commit-active {
    display: flex;
    align-items: center;
    gap: var(--sp-sm);
  }

  .commit-spinner {
    width: 8px;
    height: 8px;
    border: 1.5px solid var(--accent-green);
    border-top-color: transparent;
    border-radius: 50%;
    animation: spin 600ms linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .commit-error-section {
    padding: 8px 10px;
    border-top: 1px solid var(--border-subtle);
    background: rgba(255, 95, 87, 0.04);
  }

  .commit-error-label,
  .commit-explain-label {
    font-family: var(--font-mono);
    font-size: var(--text-label);
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    margin-bottom: 4px;
  }

  .commit-error-label { color: var(--accent-red); }
  .commit-explain-label { color: var(--accent-amber); margin-top: 8px; }

  .commit-error-text {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--accent-red);
    white-space: pre-wrap;
    word-break: break-all;
    margin: 0;
  }

  .commit-explain-text {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--text-dim);
    line-height: 1.5;
  }

  .commit-explain-loading {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--text-muted);
    font-style: italic;
  }
</style>
