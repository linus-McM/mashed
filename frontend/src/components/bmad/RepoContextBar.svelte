<script>
  import { GitBranch } from 'lucide-svelte';

  export let repoPath = '';
  export let sprintStatus = null;

  $: repoName = repoPath ? repoPath.split('/').pop() : 'No repo';
  $: totalStories = sprintStatus?.epics?.reduce((sum, e) => sum + (e.stories?.length || 0), 0) || 0;
  $: doneStories = sprintStatus?.epics?.reduce((sum, e) => sum + (e.stories?.filter(s => s.status === 'done').length || 0), 0) || 0;
  $: progressPct = totalStories > 0 ? (doneStories / totalStories) * 100 : 0;
  $: hasSprint = sprintStatus && sprintStatus.epics && sprintStatus.epics.length > 0;
</script>

<div class="context-bar">
  <div class="repo-info" title={repoPath}>
    <GitBranch size={12} />
    <span class="repo-name">{repoName}</span>
  </div>

  <div class="divider" />

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
    height: 28px;
    background: var(--bg-deepest);
    border-bottom: 1px solid var(--border-subtle);
    flex-shrink: 0;
    font-family: var(--font-mono);
    font-size: 10px;
    color: var(--text-dim);
  }

  .repo-info {
    display: flex;
    align-items: center;
    gap: 6px;
    color: var(--text-primary);
    max-width: 200px;
    overflow: hidden;
  }

  .repo-name {
    font-weight: 600;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .divider {
    width: 1px;
    height: 14px;
    background: var(--border-subtle);
    flex-shrink: 0;
  }

  .sprint-info {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .sprint-text {
    color: var(--text-dim);
    white-space: nowrap;
  }

  .sprint-text.dim {
    color: var(--text-muted);
  }

  .sprint-bar {
    width: 60px;
    height: 4px;
    background: var(--bg-surface);
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
