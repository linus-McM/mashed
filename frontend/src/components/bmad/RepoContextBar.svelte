<script>
  import { createEventDispatcher } from 'svelte';
  import { ArrowLeft, GitBranch } from 'lucide-svelte';

  export let repoPath = '';
  export let repoBranch = '';
  export let sprintStatus = null;

  const dispatch = createEventDispatcher();

  $: repoName = repoPath ? repoPath.split('/').pop() : 'No repo';
  $: totalStories = sprintStatus?.epics?.reduce((sum, e) => sum + (e.stories?.length || 0), 0) || 0;
  $: doneStories = sprintStatus?.epics?.reduce((sum, e) => sum + (e.stories?.filter(s => s.status === 'done').length || 0), 0) || 0;
  $: progressPct = totalStories > 0 ? (doneStories / totalStories) * 100 : 0;
  $: hasSprint = sprintStatus && sprintStatus.epics && sprintStatus.epics.length > 0;
</script>

<div class="context-bar">
  <button class="back-btn" on:click={() => dispatch('back')}><ArrowLeft size={14} /> Back</button>

  <div class="divider" />

  <div class="repo-info" title={repoPath}>
    <GitBranch size={12} />
    <span class="repo-name">{repoName}</span>
    {#if repoBranch}
      <span class="branch-name">{repoBranch}</span>
    {/if}
  </div>

  <div class="sprint-info">
    {#if hasSprint}
      <span class="sprint-text">{doneStories}/{totalStories} done</span>
      <div class="sprint-bar">
        <div class="sprint-fill" style="width: {progressPct}%" />
      </div>
    {:else}
      <span class="sprint-text dim">No sprint data</span>
    {/if}
  </div>
</div>

<style>
  .context-bar {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 0 10px;
    height: 36px;
    background: var(--bg-surface);
    border-bottom: 1px solid var(--border-subtle);
    flex-shrink: 0;
    font-family: var(--font-mono);
    font-size: 12px;
    color: var(--text-primary);
  }

  .back-btn {
    background: none;
    border: none;
    color: var(--accent-green);
    font-family: var(--font-mono);
    font-size: 13px;
    font-weight: 600;
    cursor: pointer;
    padding: 3px 8px;
    border-radius: var(--radius-sm);
    display: flex;
    align-items: center;
    gap: 5px;
    flex-shrink: 0;
    transition: opacity 100ms ease;
  }

  .back-btn:hover {
    opacity: 0.8;
  }

  .divider {
    width: 1px;
    height: 18px;
    background: var(--border-subtle);
    flex-shrink: 0;
  }

  .repo-info {
    display: flex;
    align-items: center;
    gap: 8px;
    color: var(--text-primary);
    max-width: 300px;
    overflow: hidden;
  }

  .repo-name {
    font-size: 13px;
    font-weight: 600;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .branch-name {
    color: var(--accent-blue, #3d9eff);
    font-size: 12px;
    font-weight: 500;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .sprint-info {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-left: auto;
  }

  .sprint-text {
    color: var(--text-primary);
    font-size: 12px;
    white-space: nowrap;
  }

  .sprint-text.dim {
    color: var(--text-muted);
  }

  .sprint-bar {
    width: 60px;
    height: 4px;
    background: var(--bg-deepest);
    border-radius: 2px;
    overflow: hidden;
  }

  .sprint-fill {
    height: 100%;
    background: var(--accent-green);
    border-radius: 2px;
    transition: width 300ms ease;
  }
</style>
