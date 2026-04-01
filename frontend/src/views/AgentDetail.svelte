<script>
  import { onMount, onDestroy, createEventDispatcher } from 'svelte';
  import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime.js';
  import StatusBadge from '../components/StatusBadge.svelte';
  import SparkLine from '../components/SparkLine.svelte';

  const dispatch = createEventDispatcher();

  /** @type {import('../types').NotificationEvent} */
  export let event;

  /** @type {import('../types').Agent|null} */
  let agent = null;

  /** @type {import('../types').DiffFileStat[]} */
  let changedFiles = [];

  /** @type {import('../types').WorktreeInfo|null} */
  let worktree = null;

  /** @type {number[]} Token burn ring buffer */
  let tokenBurn = [];

  $: tokenPct = agent ? Math.min((agent.tokensUsed / agent.tokensMax) * 100, 100) : 0;
  $: tokenLabel = agent ? formatTokens(agent.tokensUsed) + ' / ' + formatTokens(agent.tokensMax) : '';

  function formatTokens(n) {
    if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + 'M';
    if (n >= 1_000) return (n / 1_000).toFixed(1) + 'K';
    return String(n);
  }

  function handleKeydown(e) {
    if (e.key === 'Escape') {
      e.preventDefault();
      dispatch('back');
    }
  }

  onMount(() => {
    // Subscribe to agent-specific updates
    EventsOn('agent:update', (data) => {
      if (data.id === event.agentId) {
        agent = data;
      }
    });

    EventsOn('agent:diff', (data) => {
      if (data.agentId === event.agentId) {
        changedFiles = data.files || [];
      }
    });

    EventsOn('agent:worktree', (data) => {
      if (data.agentId === event.agentId) {
        worktree = data;
      }
    });

    EventsOn('agent:tokenburn', (data) => {
      if (data.agentId === event.agentId) {
        tokenBurn = data.history || [];
      }
    });

    window.addEventListener('keydown', handleKeydown);
  });

  onDestroy(() => {
    EventsOff('agent:update');
    EventsOff('agent:diff');
    EventsOff('agent:worktree');
    EventsOff('agent:tokenburn');
    window.removeEventListener('keydown', handleKeydown);
  });
</script>

<div class="detail">
  <!-- Header -->
  <div class="header">
    <button class="back-btn" on:click={() => dispatch('back')}>← Back</button>
    <span class="header-repo">{event.repoName}</span>
    <span class="header-sep">/</span>
    <span class="header-agent">{event.agentName}</span>
    <StatusBadge status={event.eventType} />
  </div>

  <div class="split">
    <!-- Left panel: agent info -->
    <div class="panel-left">
      <!-- Agent info -->
      <section class="info-section">
        <h3 class="section-title">Agent</h3>
        <div class="info-grid">
          <span class="info-label">Name</span>
          <span class="info-value">{agent?.name || event.agentName}</span>
          <span class="info-label">Repo</span>
          <span class="info-value mono">{event.repoName}</span>
          <span class="info-label">Branch</span>
          <span class="info-value mono">{event.repoBranch || '—'}</span>
          <span class="info-label">Model</span>
          <span class="info-value mono">{agent?.model || '—'}</span>
          <span class="info-label">Status</span>
          <span class="info-value"><StatusBadge status={event.eventType} /></span>
        </div>
      </section>

      <!-- Token progress -->
      <section class="info-section">
        <h3 class="section-title">Token Usage</h3>
        <div class="token-bar-container">
          <div class="token-bar">
            <div class="token-bar-fill" style="width: {tokenPct}%" />
          </div>
          <span class="token-label mono">{tokenLabel}</span>
        </div>
        {#if tokenBurn.length > 0}
          <div class="sparkline-row">
            <span class="info-label">Burn rate</span>
            <SparkLine data={tokenBurn} />
          </div>
        {/if}
      </section>

      <!-- Changed files -->
      <section class="info-section">
        <h3 class="section-title">Changed Files</h3>
        {#if changedFiles.length > 0}
          <div class="file-list">
            {#each changedFiles as file}
              <div class="file-row">
                <span class="file-path mono">{file.path}</span>
                <span class="file-stats">
                  {#if file.added > 0}<span class="stat-add">+{file.added}</span>{/if}
                  {#if file.removed > 0}<span class="stat-remove">-{file.removed}</span>{/if}
                  {#if file.isNew}<span class="stat-new">new</span>{/if}
                </span>
              </div>
            {/each}
          </div>
        {:else}
          <div class="no-data">No changes yet</div>
        {/if}
      </section>

      <!-- Worktree -->
      {#if worktree}
        <section class="info-section">
          <h3 class="section-title">Worktree</h3>
          <div class="worktree-info">
            <span class="info-label">Branch</span>
            <span class="info-value mono">{worktree.branch}</span>
          </div>
          <div class="worktree-actions">
            <button class="action-link">Merge</button>
            <button class="action-link">View Diff</button>
            <button class="action-link">Open Terminal</button>
          </div>
        </section>
      {/if}
    </div>

    <!-- Right panel: terminal placeholder -->
    <div class="panel-right">
      <div class="terminal-placeholder">
        <!-- Terminal component will be added by terminal-agent -->
        <div class="placeholder-text">Terminal output will appear here</div>
      </div>
    </div>
  </div>
</div>

<style>
  .detail {
    display: flex;
    flex-direction: column;
    height: 100%;
    background: var(--bg-deepest);
  }

  /* ── Header ── */
  .header {
    display: flex;
    align-items: center;
    gap: var(--sp-sm);
    padding: var(--sp-sm) var(--sp-lg);
    border-bottom: 1px solid var(--border-subtle);
    background: var(--bg-surface);
  }

  .back-btn {
    background: none;
    border: none;
    color: var(--text-dim);
    font-family: var(--font-ui);
    font-size: var(--text-body);
    cursor: pointer;
    padding: var(--sp-xs) var(--sp-sm);
    border-radius: var(--radius-md);
    transition: color var(--duration-short) var(--ease-enter);
  }

  .back-btn:hover {
    color: var(--accent);
  }

  .header-repo {
    font-weight: 600;
    font-size: var(--text-data);
    color: var(--text-primary);
  }

  .header-sep {
    color: var(--text-muted);
  }

  .header-agent {
    font-size: var(--text-data);
    color: var(--text-dim);
  }

  /* ── Split layout ── */
  .split {
    flex: 1;
    display: flex;
    overflow: hidden;
  }

  .panel-left {
    width: 360px;
    flex-shrink: 0;
    overflow-y: auto;
    border-right: 1px solid var(--border-subtle);
    padding: var(--sp-lg);
  }

  .panel-right {
    flex: 1;
    display: flex;
    flex-direction: column;
  }

  /* ── Info sections ── */
  .info-section {
    margin-bottom: var(--sp-xl);
  }

  .section-title {
    font-size: var(--text-label);
    font-weight: 600;
    color: var(--text-dim);
    letter-spacing: 0.06em;
    text-transform: uppercase;
    margin-bottom: var(--sp-sm);
  }

  .info-grid {
    display: grid;
    grid-template-columns: auto 1fr;
    gap: var(--sp-xs) var(--sp-md);
    align-items: center;
  }

  .info-label {
    font-size: var(--text-label);
    color: var(--text-dim);
  }

  .info-value {
    font-size: var(--text-body);
    color: var(--text-primary);
  }

  .mono {
    font-family: var(--font-mono);
  }

  /* ── Token bar ── */
  .token-bar-container {
    display: flex;
    align-items: center;
    gap: var(--sp-sm);
  }

  .token-bar {
    flex: 1;
    height: 6px;
    background: var(--bg-active);
    border-radius: 3px;
    overflow: hidden;
  }

  .token-bar-fill {
    height: 100%;
    background: var(--teal);
    border-radius: 3px;
    transition: width var(--duration-medium) var(--ease-move);
  }

  .token-label {
    font-size: var(--text-label);
    color: var(--text-dim);
    white-space: nowrap;
  }

  .sparkline-row {
    display: flex;
    align-items: center;
    gap: var(--sp-sm);
    margin-top: var(--sp-sm);
  }

  /* ── Changed files ── */
  .file-list {
    display: flex;
    flex-direction: column;
    gap: var(--sp-2xs);
  }

  .file-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: var(--sp-2xs) 0;
  }

  .file-path {
    font-size: var(--text-body);
    color: var(--text-primary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .file-stats {
    display: flex;
    gap: var(--sp-xs);
    flex-shrink: 0;
    font-family: var(--font-mono);
    font-size: var(--text-label);
  }

  .stat-add {
    color: var(--status-needs-response);
  }

  .stat-remove {
    color: var(--status-error);
  }

  .stat-new {
    color: var(--status-completed);
  }

  .no-data {
    font-size: var(--text-body);
    color: var(--text-muted);
  }

  /* ── Worktree ── */
  .worktree-info {
    display: flex;
    gap: var(--sp-sm);
    align-items: center;
    margin-bottom: var(--sp-sm);
  }

  .worktree-actions {
    display: flex;
    gap: var(--sp-md);
  }

  .action-link {
    background: none;
    border: none;
    color: var(--teal);
    font-family: var(--font-ui);
    font-size: var(--text-body);
    cursor: pointer;
    padding: 0;
    text-decoration: none;
  }

  .action-link:hover {
    color: var(--accent);
  }

  /* ── Terminal placeholder ── */
  .terminal-placeholder {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--bg-deepest);
    font-family: var(--font-terminal);
    font-size: var(--text-body);
  }

  .placeholder-text {
    color: var(--text-muted);
    user-select: none;
  }
</style>
