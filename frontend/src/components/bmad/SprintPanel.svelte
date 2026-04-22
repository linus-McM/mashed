<script>
  /** @typedef {import('../../lib/types/wails').SprintStatus} SprintStatus */
  /** @typedef {import('../../lib/types/wails').SprintEpic} SprintEpic */
  /** @typedef {import('../../lib/types/wails').SprintStory} SprintStory */

  import { ChevronDown, ChevronRight } from 'lucide-svelte';
  import { storyStatusColors as statusColors, storyStatusLabels as statusLabels } from '../../lib/sprintColors.js';

  /** @type {SprintStatus | null} */
  export let sprintStatus = null;

  /** @type {string | null} */
  let openEpic = null;

  /** @param {string} key */
  const colorFor = (key) => /** @type {Record<string, string>} */ (statusColors)[key] || /** @type {Record<string, string>} */ (statusColors).backlog;
  /** @param {string} key */
  const labelFor = (key) => /** @type {Record<string, string>} */ (statusLabels)[key] || key;

  // Auto-open the first epic that has a backlog story
  $: if (sprintStatus?.epics?.length && openEpic === null) {
    const backlogEpic = sprintStatus.epics.find((/** @type {SprintEpic} */ e) =>
      e.stories.some((/** @type {SprintStory} */ s) => s.status === 'backlog')
    );
    openEpic = backlogEpic ? backlogEpic.id : sprintStatus.epics[0].id;
  }

  /** @param {string} epicId */
  function toggleEpic(epicId) {
    openEpic = openEpic === epicId ? null : epicId;
  }

  /** @param {SprintEpic} epic */
  function epicDone(epic) {
    return epic.stories.filter((/** @type {SprintStory} */ s) => s.status === 'done').length;
  }

  // The first epic with a backlog story — gets green highlight
  /** @type {string | null} */
  $: activeEpicId = sprintStatus?.epics?.find((/** @type {SprintEpic} */ e) => e.stories.some((/** @type {SprintStory} */ s) => s.status === 'backlog'))?.id || null;

  /**
   * @param {DragEvent} e
   * @param {SprintStory} story
   */
  function onDragStart(e, story) {
    if (!e.dataTransfer) return;
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
      {@const done = epicDone(epic)}
      {@const total = epic.stories.length}
      {@const isActive = epic.id === activeEpicId}
      {@const isComplete = done === total && total > 0}
      <div class="epic-group" class:active-epic={isActive}>
        <button class="epic-header" on:click={() => toggleEpic(epic.id)}>
          <span class="epic-indicator" style="background: {colorFor(epic.status)}" />
          {#if openEpic === epic.id}
            <ChevronDown size={12} />
          {:else}
            <ChevronRight size={12} />
          {/if}
          <span class="epic-label">{epic.id}</span>
          <span class="epic-progress-badge" class:complete={isComplete} class:active={isActive}>
            {done}/{total}
          </span>
        </button>
        {#if openEpic === epic.id}
          <div class="story-items">
            {#each epic.stories as story}
              <!-- svelte-ignore a11y-no-static-element-interactions -->
              <div
                class="story-item"
                draggable="true"
                on:dragstart={(e) => onDragStart(e, story)}
                title="Drag to canvas to create a workflow node"
                role="listitem"
              >
                <span class="story-dot" style="background: {colorFor(story.status)}" />
                <span class="story-name">{story.id}</span>
                <span class="story-status" style="color: {colorFor(story.status)}">
                  {labelFor(story.status)}
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
    padding: var(--sp-xs) var(--sp-sm);
  }

  .epic-group {
    margin-bottom: 4px;
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: 6px;
    overflow: hidden;
  }

  .epic-header {
    display: flex;
    align-items: center;
    gap: 8px;
    width: 100%;
    padding: 8px 10px;
    background: var(--bg-elevated);
    border: none;
    color: var(--text-primary);
    font-family: var(--font-mono);
    font-size: var(--text-body);
    font-weight: 600;
    cursor: pointer;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    transition: background 80ms ease;
  }

  .epic-header:hover { background: var(--bg-active); }

  .epic-indicator {
    width: 8px;
    height: 8px;
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

  .epic-progress-badge {
    font-family: var(--font-mono);
    font-size: 11px;
    font-weight: 600;
    color: var(--text-primary);
    padding: 1px 8px;
    border-radius: 10px;
    background: var(--bg-deepest);
    border: 1px solid var(--border-subtle);
    flex-shrink: 0;
  }

  .epic-progress-badge.complete {
    color: var(--accent-green, #00e57a);
    border-color: color-mix(in srgb, var(--accent-green) 30%, transparent);
    background: color-mix(in srgb, var(--accent-green) 10%, transparent);
  }

  .epic-progress-badge.active {
    color: var(--accent-green, #00e57a);
    border-color: var(--accent-green, #00e57a);
    background: color-mix(in srgb, var(--accent-green) 12%, transparent);
  }

  .active-epic {
    border-color: var(--accent-green, #00e57a);
  }

  .story-items {
    padding: var(--sp-2xs) 0 var(--sp-xs);
    border-top: 1px solid var(--border-subtle);
    background: var(--bg-surface);
  }

  .story-item {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: var(--sp-xs) 10px var(--sp-xs) 24px;
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
