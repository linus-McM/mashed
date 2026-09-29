<script lang="ts">
  // Repo group header: collapse toggle, drag handle, name/branch/stats and the
  // border colour picker. Extracted from views/NotificationFeed.svelte (R31).
  // Colour + collapse state live in the parent; this component renders them
  // and bubbles intents. dragstart/dragend are forwarded to the parent, which
  // owns drag-reorder.
  import { createEventDispatcher } from 'svelte';
  import { GripVertical, GitBranch, ChevronRight, ChevronDown } from 'lucide-svelte';
  import { REPO_BORDER_PALETTE, REPO_BORDER_NONE } from '../../lib/repoPalette';
  import { formatTokens, repoTokens, type RepoGroup } from '../../lib/feed/repoTree';

  export let repo: RepoGroup;
  export let collapsed: boolean;
  /** Current border colour for this repo (parent's getRepoColor(repo.name)). */
  export let repoColor: string;
  /** Whether this repo's colour-picker popover is open. */
  export let pickerOpen: boolean;

  const dispatch = createEventDispatcher<{
    toggle: void;
    togglecolorpicker: void;
    setcolor: string;
  }>();

  const borderPalette: readonly string[] = REPO_BORDER_PALETTE;
</script>

<!-- Repo header (draggable, dblclick to toggle) -->
<div
  class="repo-header"
  role="listitem"
  class:collapsed={collapsed}
  draggable="true"
  on:dragstart
  on:dragend
>
  <button class="collapse-btn" on:click|stopPropagation={() => dispatch('toggle')}>
    {#if collapsed}
      <ChevronRight size={14} />
    {:else}
      <ChevronDown size={14} />
    {/if}
  </button>
  <span class="drag-handle" style="color: {repoColor !== REPO_BORDER_NONE ? repoColor : ''}"><GripVertical size={14} /></span>
  <span class="repo-name">{repo.name}</span>
  {#if repo.branch}
    <span class="repo-branch" style="color: {repoColor !== REPO_BORDER_NONE ? repoColor : ''}"><GitBranch size={12} /> {repo.branch}</span>
  {/if}
  <span class="repo-stats mono">
    {repo.agents.length} agent{repo.agents.length !== 1 ? 's' : ''} · {formatTokens(repoTokens(repo))}
  </span>
  <div class="color-picker-wrap">
    <button
      class="color-picker-btn"
      style="background: {repoColor}"
      on:click|stopPropagation={() => dispatch('togglecolorpicker')}
      title="Change border color"
    />
    {#if pickerOpen}
      <div class="color-picker-popover">
        {#each borderPalette as color}
          <button
            class="color-swatch"
            class:active={repoColor === color}
            style="background: {color}"
            on:click|stopPropagation={() => dispatch('setcolor', color)}
          />
        {/each}
      </div>
    {/if}
  </div>
</div>

<style>
  .repo-header {
    display: flex;
    align-items: center;
    gap: var(--sp-sm);
    padding: var(--sp-sm) var(--sp-lg);
    background: var(--bg-surface);
    border-bottom: 1px solid var(--border-subtle);
    user-select: none;
    cursor: grab;
  }

  .repo-header:active { cursor: grabbing; }

  .repo-header.collapsed {
    border-bottom: none;
  }

  .collapse-btn {
    background: none;
    border: none;
    color: var(--text-muted);
    cursor: pointer;
    padding: 0;
    display: flex;
    align-items: center;
    flex-shrink: 0;
    transition: color 100ms ease;
  }

  .collapse-btn:hover { color: var(--text-dim); }

  .drag-handle {
    color: var(--text-muted);
    font-size: 14px;
    line-height: 1;
    flex-shrink: 0;
    transition: color 120ms ease;
  }

  .repo-header:hover .drag-handle { color: var(--text-dim); }

  /* Color picker */
  .color-picker-wrap {
    position: relative;
    margin-left: var(--sp-xs);
  }

  .color-picker-btn {
    width: 14px;
    height: 14px;
    border-radius: 50%;
    border: 1.5px solid rgba(255, 255, 255, 0.15);
    cursor: pointer;
    transition: transform 120ms ease, box-shadow 120ms ease;
  }

  .color-picker-btn:hover {
    transform: scale(1.2);
    box-shadow: 0 0 6px rgba(255, 255, 255, 0.15);
  }

  .color-picker-popover {
    position: absolute;
    top: calc(100% + 6px);
    right: 0;
    display: flex;
    gap: var(--sp-sm);
    padding: 8px;
    background: var(--bg-elevated);
    border: 1px solid var(--border-emphasis);
    border-radius: var(--radius-md);
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.5);
    z-index: 100;
    flex-wrap: wrap;
    width: 160px;
  }

  .color-swatch {
    width: 18px;
    height: 18px;
    border-radius: 50%;
    border: 1.5px solid transparent;
    cursor: pointer;
    transition: transform 100ms ease, border-color 100ms ease;
  }

  .color-swatch:hover {
    transform: scale(1.25);
  }

  .color-swatch.active {
    border-color: var(--text-primary);
  }

  .repo-name {
    font-family: var(--font-mono);
    font-size: var(--text-data);
    font-weight: 600;
    color: var(--text-primary);
  }

  .repo-branch {
    font-family: var(--font-mono);
    font-size: var(--text-label);
    color: var(--text-dim);
  }

  .repo-stats {
    margin-left: auto;
    font-size: var(--text-label);
    color: var(--text-dim);
  }
  .mono { font-family: var(--font-mono); }
</style>
