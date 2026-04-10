<script>
  import { createEventDispatcher, onMount, onDestroy } from 'svelte';
  import { fly } from 'svelte/transition';
  import { flip } from 'svelte/animate';
  import { cubicOut, cubicIn } from 'svelte/easing';
  import { MessageCircleQuestion } from 'lucide-svelte';
  import {
    truncate,
    getBorderColor,
    timeAgo,
    partitionForDisplay,
  } from './questionSnackbarUtils';

  /** @type {import('./questionSnackbarUtils').QuestionEventLike[]} */
  export let questions = [];

  const dispatch = createEventDispatcher();

  // Tick a store every 30s so timeAgo() output re-renders without each card
  // having its own setInterval.
  let now = Date.now();
  /** @type {ReturnType<typeof setInterval> | undefined} */
  let tickHandle;
  onMount(() => {
    tickHandle = setInterval(() => { now = Date.now(); }, 30_000);
  });
  onDestroy(() => {
    if (tickHandle) clearInterval(tickHandle);
  });

  $: ({ visible, overflow } = partitionForDisplay(questions));

  function handleClick(q) {
    dispatch('navigate', { repoPath: q.repoPath, question: q });
  }

  function handleKeydown(e, q) {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      handleClick(q);
    }
  }

  function displayName(q) {
    return q.repoName && q.repoName.length > 0 ? q.repoName : 'Unknown Repo';
  }
</script>

<div class="snackbar-stack" aria-live="polite" role="status">
  {#each visible as q (q.nodeId)}
    <div
      class="snackbar-card"
      role="button"
      tabindex="0"
      on:click={() => handleClick(q)}
      on:keydown={(e) => handleKeydown(e, q)}
      in:fly={{ x: 320, duration: 200, easing: cubicOut }}
      out:fly={{ x: 320, duration: 200, easing: cubicIn }}
      animate:flip={{ duration: 200 }}
    >
      <span
        class="accent-stripe"
        style:background-color={getBorderColor(q.repoName)}
        aria-hidden="true"
      ></span>
      <div class="card-content">
        <div class="card-top">
          <span class="question-icon" aria-hidden="true">
            <MessageCircleQuestion size={14} />
          </span>
          <span class="repo-name">{displayName(q)}</span>
          {#if q.timestamp}
            <span class="timestamp">{timeAgo(q.timestamp, now)}</span>
          {/if}
        </div>
        <div class="question-text">{truncate(q.question || '', 80)}</div>
      </div>
    </div>
  {/each}

  {#if overflow > 0}
    <div class="overflow-indicator" aria-label={`${overflow} more questions`}>
      +{overflow} more
    </div>
  {/if}
</div>

<style>
  .snackbar-stack {
    position: fixed;
    top: var(--sp-lg);
    right: var(--sp-lg);
    z-index: 300;
    display: flex;
    flex-direction: column;
    gap: var(--sp-sm);
    width: 320px;
    pointer-events: none;
  }

  .snackbar-card {
    pointer-events: auto;
    display: flex;
    align-items: stretch;
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    cursor: pointer;
    overflow: hidden;
    box-shadow: 0 2px 8px color-mix(in srgb, var(--bg-deepest) 60%, transparent);
    transition:
      background var(--duration-short) var(--ease-enter),
      border-color var(--duration-short) var(--ease-enter),
      transform var(--duration-micro) var(--ease-enter);
  }

  .snackbar-card:hover {
    background: var(--bg-active);
    border-color: var(--border-emphasis);
  }

  .snackbar-card:focus-visible {
    outline: none;
    border-color: var(--accent-amber);
    box-shadow: 0 0 0 2px color-mix(in srgb, var(--accent-amber) 40%, transparent);
  }

  .snackbar-card:active {
    transform: scale(0.98);
  }

  .accent-stripe {
    width: 3px;
    flex-shrink: 0;
    border-radius: var(--radius-sm) 0 0 var(--radius-sm);
  }

  .card-content {
    display: flex;
    flex-direction: column;
    padding: var(--sp-sm) var(--sp-md);
    gap: var(--sp-2xs);
    flex: 1;
    min-width: 0;
  }

  .card-top {
    display: flex;
    align-items: baseline;
    gap: var(--sp-xs);
  }

  .question-icon {
    display: inline-flex;
    align-items: center;
    color: var(--accent-amber);
    flex-shrink: 0;
    /* Slight nudge so the icon aligns with the baseline of text */
    transform: translateY(2px);
  }

  .repo-name {
    flex: 1;
    font-weight: 600;
    font-size: var(--text-body);
    color: var(--text-primary);
    line-height: 1.4;
    letter-spacing: -0.01em;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .timestamp {
    font-family: var(--font-mono);
    font-size: var(--text-label);
    line-height: 1.4;
    color: var(--text-muted);
    font-variant-numeric: tabular-nums;
    white-space: nowrap;
    flex-shrink: 0;
  }

  .question-text {
    font-size: var(--text-body);
    color: var(--text-dim);
    line-height: 1.4;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .overflow-indicator {
    pointer-events: none;
    font-family: var(--font-mono);
    font-size: var(--text-label);
    font-weight: 500;
    color: var(--text-dim);
    text-align: center;
    padding: var(--sp-xs) 0;
    letter-spacing: 0.02em;
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
  }
</style>
