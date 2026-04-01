<script>
  import { createEventDispatcher } from 'svelte';
  import StatusBadge from '../components/StatusBadge.svelte';
  import SparkLine from '../components/SparkLine.svelte';

  const dispatch = createEventDispatcher();

  /** @type {any[]} Notifications passed from App.svelte */
  export let notifications = [];

  let selectedIndex = 0;

  // Group order: NEEDS RESPONSE (green), ERRORS (red), COMPLETED (blue), RUNNING (amber)
  const groupOrder = ['needs_response', 'error', 'completed', 'running', 'started'];
  const groupLabels = {
    needs_response: 'NEEDS RESPONSE',
    error: 'ERRORS',
    completed: 'COMPLETED',
    running: 'RUNNING',
    started: 'STARTED',
  };

  $: grouped = groupOrder
    .map(type => ({
      type,
      label: groupLabels[type],
      items: notifications.filter(e => e.eventType === type),
    }))
    .filter(g => g.items.length > 0);

  $: flatEvents = grouped.flatMap(g => g.items);
  $: totalAgents = new Set(notifications.map(e => e.agentId)).size;
  $: totalRepos = new Set(notifications.map(e => e.repoName)).size;
  $: totalTokens = notifications.reduce((sum, e) => sum + (e.tokensUsed || 0), 0);

  function formatTokens(n) {
    if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + 'M';
    if (n >= 1_000) return (n / 1_000).toFixed(1) + 'K';
    return String(n);
  }

  function formatElapsed(ts) {
    if (!ts) return '';
    const diff = Date.now() - new Date(ts).getTime();
    const secs = Math.floor(diff / 1000);
    if (secs < 60) return secs + 's';
    const mins = Math.floor(secs / 60);
    if (mins < 60) return mins + 'm';
    const hrs = Math.floor(mins / 60);
    return hrs + 'h ' + (mins % 60) + 'm';
  }

  function handleKeydown(e) {
    if (e.key === 'j' || e.key === 'ArrowDown') {
      e.preventDefault();
      selectedIndex = Math.min(selectedIndex + 1, flatEvents.length - 1);
      scrollSelectedIntoView();
    } else if (e.key === 'k' || e.key === 'ArrowUp') {
      e.preventDefault();
      selectedIndex = Math.max(selectedIndex - 1, 0);
      scrollSelectedIntoView();
    } else if (e.key === 'Enter') {
      e.preventDefault();
      const ev = flatEvents[selectedIndex];
      if (ev) dispatch('select', ev);
    }
  }

  function scrollSelectedIntoView() {
    const el = document.querySelector(`[data-index="${selectedIndex}"]`);
    if (el) el.scrollIntoView({ block: 'nearest' });
  }

  function handleClick(ev, idx) {
    selectedIndex = idx;
    dispatch('select', ev);
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<div class="feed">
  <div class="feed-scroll">
    {#each grouped as group}
      <div class="group">
        <div class="group-header">{group.label}</div>
        {#each group.items as event, i}
          {@const globalIdx = flatEvents.indexOf(event)}
          <div
            class="row"
            class:selected={globalIdx === selectedIndex}
            data-index={globalIdx}
            on:click={() => handleClick(event, globalIdx)}
            on:keydown={(e) => { if (e.key === 'Enter') handleClick(event, globalIdx); }}
            role="button"
            tabindex="0"
          >
            <div class="accent-stripe" style="background: {statusColor(event.eventType)}" />
            <div class="row-content">
              <span class="repo mono">{event.repoName}</span>
              <span class="agent">{event.agentName}</span>
              <span class="summary">{event.summary}</span>
              {#if event.tokenBurn && event.tokenBurn.length > 0}
                <SparkLine data={event.tokenBurn} />
              {/if}
              <span class="elapsed mono">{formatElapsed(event.timestamp)}</span>
            </div>
          </div>
        {/each}
      </div>
    {/each}

    {#if notifications.length === 0}
      <div class="empty">
        <div class="empty-icon">&#9671;</div>
        <div class="empty-text">No active agents</div>
        <div class="empty-sub">Start a Claude Code session to see events here</div>
      </div>
    {/if}
  </div>

  <div class="status-bar">
    <span>{totalAgents} agent{totalAgents !== 1 ? 's' : ''}</span>
    <span class="sep">&middot;</span>
    <span>{totalRepos} repo{totalRepos !== 1 ? 's' : ''}</span>
    <span class="sep">&middot;</span>
    <span class="mono">{formatTokens(totalTokens)} tokens</span>
    <span class="keys">
      <kbd>j</kbd>/<kbd>k</kbd> navigate &middot; <kbd>Enter</kbd> open
    </span>
  </div>
</div>

<script context="module">
  const statusColors = {
    running:        'var(--accent-amber)',
    error:          'var(--accent-red)',
    completed:      'var(--accent-blue)',
    needs_response: 'var(--accent-green)',
    started:        'var(--accent-purple)',
  };

  function statusColor(type) {
    return statusColors[type] || 'var(--text-dim)';
  }
</script>

<style>
  .feed {
    display: flex;
    flex-direction: column;
    height: 100%;
    background: var(--bg-deepest);
  }

  .feed-scroll {
    flex: 1;
    overflow-y: auto;
    padding: var(--sp-sm) 0;
  }

  /* Group */
  .group {
    margin-bottom: var(--sp-sm);
  }

  .group-header {
    padding: var(--sp-xs) var(--sp-lg);
    font-size: var(--text-label);
    font-weight: 600;
    color: var(--text-dim);
    letter-spacing: 0.06em;
    user-select: none;
  }

  /* Row */
  .row {
    display: flex;
    align-items: stretch;
    cursor: pointer;
    transition: background var(--duration-short) var(--ease-enter);
  }

  .row:hover {
    background: var(--bg-surface);
  }

  .row.selected {
    background: var(--bg-elevated);
  }

  .accent-stripe {
    width: 4px;
    flex-shrink: 0;
    border-radius: 0 var(--radius-sm) var(--radius-sm) 0;
  }

  .row-content {
    display: flex;
    align-items: center;
    gap: var(--sp-md);
    padding: var(--sp-sm) var(--sp-lg);
    flex: 1;
    min-width: 0;
  }

  .repo {
    font-size: var(--text-data);
    font-weight: 500;
    color: var(--text-primary);
    flex-shrink: 0;
  }

  .agent {
    font-size: var(--text-body);
    color: var(--text-dim);
    flex-shrink: 0;
  }

  .summary {
    font-size: var(--text-body);
    color: var(--text-primary);
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .elapsed {
    font-size: var(--text-label);
    color: var(--text-dim);
    flex-shrink: 0;
  }

  /* Empty state */
  .empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 60vh;
    user-select: none;
  }

  .empty-icon {
    font-size: 32px;
    color: var(--text-muted);
    margin-bottom: var(--sp-lg);
  }

  .empty-text {
    font-size: var(--text-section);
    color: var(--text-dim);
    margin-bottom: var(--sp-xs);
  }

  .empty-sub {
    font-size: var(--text-body);
    color: var(--text-muted);
  }

  /* Status bar */
  .status-bar {
    display: flex;
    align-items: center;
    gap: var(--sp-sm);
    padding: var(--sp-sm) var(--sp-lg);
    border-top: 1px solid var(--border-subtle);
    background: var(--bg-surface);
    font-size: var(--text-label);
    color: var(--text-dim);
    user-select: none;
  }

  .sep {
    color: var(--text-muted);
  }

  .keys {
    margin-left: auto;
    color: var(--text-muted);
    font-size: var(--text-label);
  }

  .keys :global(kbd) {
    display: inline-block;
    padding: 0 4px;
    background: var(--bg-active);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-sm);
    font-family: var(--font-mono);
    font-size: 10px;
    color: var(--text-dim);
  }
</style>
