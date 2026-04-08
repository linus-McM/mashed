<script>
  import { ChevronDown, ChevronRight } from 'lucide-svelte';
  import { storyStatusColors as statusColors, storyStatusLabels as statusLabels } from '../../lib/sprintColors.js';

  export let sprintStatus = null;

  let expandedEpics = {};

  $: if (sprintStatus?.epics) {
    for (const epic of sprintStatus.epics) {
      if (!(epic.id in expandedEpics)) {
        expandedEpics[epic.id] = true;
      }
    }
  }

  function toggleEpic(epicId) {
    expandedEpics[epicId] = !expandedEpics[epicId];
    expandedEpics = expandedEpics;
  }

  function epicProgress(epic) {
    const done = epic.stories.filter(s => s.status === 'done').length;
    return `${done}/${epic.stories.length}`;
  }

  function onDragStart(e, story) {
    e.dataTransfer.setData('application/bmad-story', JSON.stringify({
      storyId: story.id,
      epicId: story.epicId,
      status: story.status,
    }));
    e.dataTransfer.effectAllowed = 'move';
  }
</script>

<div class="sprint-panel">
  {#if !sprintStatus || !sprintStatus.epics || sprintStatus.epics.length === 0}
    <div class="empty-state">No sprint data found</div>
  {:else}
    {#each sprintStatus.epics as epic}
      <div class="epic-group">
        <button class="epic-header" on:click={() => toggleEpic(epic.id)}>
          <span class="epic-indicator" style="background: {statusColors[epic.status] || statusColors.backlog}" />
          {#if expandedEpics[epic.id]}
            <ChevronDown size={12} />
          {:else}
            <ChevronRight size={12} />
          {/if}
          <span class="epic-label">{epic.id}</span>
          <span class="epic-progress">{epicProgress(epic)}</span>
        </button>
        {#if expandedEpics[epic.id]}
          <div class="story-items">
            {#each epic.stories as story}
              <div
                class="story-item"
                draggable="true"
                on:dragstart={(e) => onDragStart(e, story)}
                title="Drag to canvas to create a workflow node"
              >
                <span class="story-dot" style="background: {statusColors[story.status] || statusColors.backlog}" />
                <span class="story-name">{story.id}</span>
                <span class="story-status" style="color: {statusColors[story.status] || statusColors.backlog}">
                  {statusLabels[story.status] || story.status}
                </span>
              </div>
            {/each}
          </div>
        {/if}
      </div>
    {/each}
  {/if}
</div>

<style>
  .sprint-panel {
    padding: 4px 0;
  }

  .epic-group {
    margin-bottom: 2px;
  }

  .epic-header {
    display: flex;
    align-items: center;
    gap: 6px;
    width: 100%;
    padding: 6px 10px;
    background: none;
    border: none;
    color: var(--text-dim);
    font-family: var(--font-mono);
    font-size: 11px;
    font-weight: 600;
    cursor: pointer;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .epic-header:hover { background: var(--bg-elevated); }

  .epic-indicator {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    flex-shrink: 0;
  }

  .epic-label {
    flex: 1;
    text-align: left;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .epic-progress {
    font-size: 10px;
    color: var(--text-muted);
    font-weight: 400;
  }

  .story-items {
    padding: 0 0 4px;
  }

  .story-item {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 5px 10px 5px 24px;
    cursor: grab;
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--text-primary);
    transition: background 80ms ease;
    user-select: none;
  }

  .story-item:hover { background: var(--bg-elevated); }
  .story-item:active { cursor: grabbing; }

  .story-dot {
    width: 5px;
    height: 5px;
    border-radius: 50%;
    flex-shrink: 0;
  }

  .story-name {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    flex: 1;
  }

  .story-status {
    font-size: 8px;
    padding: 1px 4px;
    border-radius: 3px;
    background: var(--bg-deepest);
    border: 1px solid var(--border-subtle);
    flex-shrink: 0;
    text-transform: uppercase;
    letter-spacing: 0.3px;
    font-weight: 600;
  }

  .empty-state {
    padding: 20px 10px;
    text-align: center;
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--text-muted);
  }
</style>
