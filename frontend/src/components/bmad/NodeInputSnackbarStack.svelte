<script>
  import { createEventDispatcher, onMount, onDestroy } from 'svelte';
  import { fly } from 'svelte/transition';
  import { flip } from 'svelte/animate';
  import { cubicOut, cubicIn } from 'svelte/easing';
  import { MessageCircleQuestion, Keyboard } from 'lucide-svelte';
  import {
    truncate,
    getBorderColor,
    timeAgo,
    partitionForDisplay,
    isQuestionEntry,
  } from './questionSnackbarUtils';

  /** @type {import('./questionSnackbarUtils').SnackbarEntry[]} */
  export let questions = [];
  /** @type {import('../../stores/interactiveInput').PendingPrompt[]} */
  export let pendingPrompts = [];

  const dispatch = createEventDispatcher();

  let now = Date.now();
  /** @type {ReturnType<typeof setInterval> | undefined} */
  let tickHandle;
  onMount(() => {
    tickHandle = setInterval(() => { now = Date.now(); }, 30_000);
  });
  onDestroy(() => {
    if (tickHandle) clearInterval(tickHandle);
  });

  $: merged = mergeEntries(questions, pendingPrompts);
  $: ({ visible, overflow } = partitionForDisplay(merged));

  /**
   * Merge legacy question/idle entries with new interactive PendingPrompts.
   * Prompt entries carry kind:'prompt' so the render path can branch on it.
   */
  function mergeEntries(qs, prompts) {
    const promptEntries = (prompts || []).map((p) => ({
      kind: 'prompt',
      nodeId: p.nodeId,
      inputId: p.inputId,
      repoPath: p.repoPath || '',
      repoName: p.repoName || '',
      question: p.prompt,
      options: p.options || [],
      timestamp: p.createdAt,
      round: p.round,
      maxRounds: p.maxRounds,
      required: p.required !== false,
      shape: p.shape,
      prompt: p,
    }));
    // Keep legacy queue first (oldest on top), then new interactive prompts.
    return [...(qs || []), ...promptEntries];
  }

  function isPromptEntry(q) { return q && q.kind === 'prompt'; }

  /** @param {{round?: number, maxRounds?: number}} entry */
  function roundLabel(entry) {
    if (!entry || !entry.round || entry.round <= 0) return '';
    if (entry.maxRounds && entry.maxRounds > 0) return `${entry.round} / ${entry.maxRounds}`;
    return `Round ${entry.round}`;
  }

  function handleNavigate(entry) {
    if (isPromptEntry(entry)) {
      dispatch('respond', { prompt: entry.prompt });
      return;
    }
    dispatch('navigate', {
      repoPath: entry.repoPath,
      question: isQuestionEntry(entry) ? entry : undefined,
      entry,
    });
  }

  function handleSkip(entry) {
    if (!isPromptEntry(entry)) return;
    dispatch('skip', { prompt: entry.prompt });
  }

  function handleClick(entry) { handleNavigate(entry); }

  function handleKeydown(e, entry) {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      handleNavigate(entry);
    }
  }

  function displayName(entry) {
    return entry.repoName && entry.repoName.length > 0 ? entry.repoName : 'Unknown Repo';
  }

  function bodyText(entry) {
    if (isPromptEntry(entry)) return truncate(entry.question || '', 80);
    if (isQuestionEntry(entry)) return truncate(entry.question || '', 80);
    return 'Waiting for input — open terminal to reply';
  }
</script>

<div class="snackbar-stack" aria-live="polite" role="status">
  {#each visible as q (q.kind === 'prompt' ? `p:${q.nodeId}:${q.inputId}` : q.nodeId)}
    <div
      class="snackbar-card"
      class:idle={!isQuestionEntry(q) && !isPromptEntry(q)}
      class:prompt={isPromptEntry(q)}
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
            {#if isPromptEntry(q) || isQuestionEntry(q)}
              <MessageCircleQuestion size={14} />
            {:else}
              <Keyboard size={14} />
            {/if}
          </span>
          <span class="repo-name">{displayName(q)}</span>
          {#if isPromptEntry(q)}
            {#if roundLabel(q)}
              <span class="round-pill" data-testid="snackbar-round-pill">{roundLabel(q)}</span>
            {/if}
          {/if}
          {#if q.timestamp}
            <span class="timestamp">{timeAgo(q.timestamp, now)}</span>
          {/if}
        </div>
        <div class="question-text">{bodyText(q)}</div>
        {#if isPromptEntry(q)}
          <div class="card-actions">
            <button
              type="button"
              class="respond-btn"
              data-testid="snackbar-respond"
              on:click|stopPropagation={() => handleNavigate(q)}
            >
              Respond
            </button>
            {#if q.required === false}
              <button
                type="button"
                class="skip-btn"
                data-testid="snackbar-skip"
                on:click|stopPropagation={() => handleSkip(q)}
              >
                Skip
              </button>
            {/if}
          </div>
        {/if}
      </div>
    </div>
  {/each}

  {#if overflow > 0}
    <div class="overflow-indicator" aria-label={`${overflow} more prompts`}>
      +{overflow} more
    </div>
  {/if}
</div>

<style>
  .snackbar-stack {
    position: fixed;
    top: var(--sp-lg);
    right: var(--sp-lg);
    z-index: var(--z-snackbar);
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

  .snackbar-card:active { transform: scale(0.98); }

  .snackbar-card.prompt { border-color: color-mix(in srgb, var(--accent-amber) 40%, transparent); }

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
    transform: translateY(2px);
  }

  .snackbar-card.idle .question-icon { color: var(--accent-cyan, var(--text-dim)); }
  .snackbar-card.idle .question-text { color: var(--text-dim); font-style: italic; }

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

  .round-pill {
    background: color-mix(in srgb, var(--accent-amber) 15%, transparent);
    color: var(--accent-amber);
    padding: 0 var(--sp-xs);
    border-radius: var(--radius-sm);
    font-family: var(--font-mono);
    font-size: var(--text-label);
    font-weight: 600;
    font-variant-numeric: tabular-nums;
    letter-spacing: 0.04em;
    flex-shrink: 0;
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
    color: var(--text-primary);
    line-height: 1.4;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .card-actions {
    display: flex;
    gap: var(--sp-xs);
    padding-top: var(--sp-2xs);
  }

  .respond-btn {
    background: color-mix(in srgb, var(--accent-amber) 15%, transparent);
    color: var(--accent-amber);
    border: 1px solid color-mix(in srgb, var(--accent-amber) 40%, transparent);
    border-radius: var(--radius-sm);
    padding: var(--sp-2xs) var(--sp-sm);
    font-family: var(--font-ui);
    font-size: var(--text-label);
    font-weight: 600;
    cursor: pointer;
    transition:
      border-color var(--duration-short) var(--ease-enter),
      filter var(--duration-short) var(--ease-enter);
  }

  .respond-btn:hover { border-color: var(--accent-amber); filter: brightness(1.1); }
  .respond-btn:focus-visible {
    outline: 1px solid var(--accent-amber);
    outline-offset: 1px;
  }

  .skip-btn {
    background: transparent;
    color: var(--text-dim);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-sm);
    padding: var(--sp-2xs) var(--sp-sm);
    font-family: var(--font-ui);
    font-size: var(--text-label);
    font-weight: 500;
    cursor: pointer;
    transition:
      background var(--duration-short) var(--ease-enter),
      color var(--duration-short) var(--ease-enter);
  }

  .skip-btn:hover { background: var(--bg-elevated); color: var(--text-primary); }
  .skip-btn:focus-visible {
    outline: 1px solid var(--text-dim);
    outline-offset: 1px;
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
