<script>
  import { createEventDispatcher } from 'svelte';
  import { SpawnAgent } from '../../wailsjs/go/main/App.js';
  import StatusBadge from '../components/StatusBadge.svelte';
  import SparkLine from '../components/SparkLine.svelte';

  const dispatch = createEventDispatcher();

  let spawningRepo = null; // repo path currently spawning

  /** @type {any[]} Notifications passed from App.svelte */
  export let notifications = [];

  let selectedId = null;

  // Group: repo → agents → sub-agents
  $: repoGroups = buildRepoTree(notifications);
  $: flatAgents = repoGroups.flatMap(r => r.agents);
  $: totalAgents = flatAgents.length;
  $: totalRepos = repoGroups.length;
  $: totalTokens = notifications.reduce((sum, e) => sum + (e.tokensUsed || 0), 0);

  function buildRepoTree(events) {
    const repoMap = new Map();

    for (const evt of events) {
      const repoKey = evt.repoName || 'unknown';
      if (!repoMap.has(repoKey)) {
        repoMap.set(repoKey, {
          name: repoKey,
          path: evt.repoPath || '',
          branch: evt.repoBranch || '',
          agents: [],
          worstStatus: 'running',
        });
      }
      const repo = repoMap.get(repoKey);

      // Check if this is a sub-agent (ID contains "-sub-")
      const isSubAgent = evt.agentId && evt.agentId.includes('-sub-');

      if (isSubAgent) {
        // Find parent agent and nest under it
        const parentId = evt.agentId.split('-sub-')[0];
        let parent = repo.agents.find(a => a.agentId === parentId);
        if (!parent) {
          // Parent not found, show as top-level
          repo.agents.push({ ...evt, subAgents: [] });
        } else {
          if (!parent.subAgents) parent.subAgents = [];
          parent.subAgents.push(evt);
        }
      } else {
        // Top-level agent
        const existing = repo.agents.find(a => a.agentId === evt.agentId);
        if (existing) {
          Object.assign(existing, evt);
        } else {
          repo.agents.push({ ...evt, subAgents: [] });
        }
      }

      // Track worst status for repo header
      if (evt.eventType === 'needs_response' || evt.eventType === 'error') {
        repo.worstStatus = evt.eventType;
      }
    }

    // Sort repos: repos with attention-needed first, then alphabetical
    const statusPriority = { needs_response: 0, error: 1, running: 2, started: 3, completed: 4 };
    return Array.from(repoMap.values()).sort((a, b) => {
      const pa = statusPriority[a.worstStatus] ?? 5;
      const pb = statusPriority[b.worstStatus] ?? 5;
      if (pa !== pb) return pa - pb;
      return a.name.localeCompare(b.name);
    });
  }

  function repoTokens(repo) {
    let sum = 0;
    for (const a of repo.agents) {
      sum += a.tokensUsed || 0;
      if (a.subAgents) {
        for (const s of a.subAgents) sum += s.tokensUsed || 0;
      }
    }
    return sum;
  }

  function repoStatusColor(repo) {
    return statusColor(repo.worstStatus);
  }

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

  async function spawnInRepo(repo) {
    if (spawningRepo) return;
    spawningRepo = repo.path;
    try {
      console.log('Spawning in repo:', repo.path);
      const target = await SpawnAgent(repo.path, 'claude-opus-4-6');
      console.log('SpawnAgent returned target:', target);
      const agent = {
        agentId: `spawned-${Date.now()}`,
        agentName: 'claude-opus-4-6',
        model: 'claude-opus-4-6',
        repoName: repo.name,
        repoPath: repo.path,
        repoBranch: repo.branch,
        eventType: 'running',
        tmuxTarget: target,
        tokensUsed: 0,
        tokensMax: 1000000,
        summary: `New session in ${repo.name}`,
      };
      console.log('Dispatching select with agent:', JSON.stringify(agent));
      dispatch('select', agent);
    } catch (e) {
      console.error('Spawn failed:', e);
    } finally {
      spawningRepo = null;
    }
  }

  function handleClick(evt) {
    selectedId = evt.agentId;
    dispatch('select', evt);
  }

  function handleKeydown(e) {
    const allIds = flatAgents.map(a => a.agentId);
    const currentIdx = allIds.indexOf(selectedId);

    if (e.key === 'j' || e.key === 'ArrowDown') {
      e.preventDefault();
      const next = Math.min(currentIdx + 1, allIds.length - 1);
      selectedId = allIds[next];
      scrollIntoView(selectedId);
    } else if (e.key === 'k' || e.key === 'ArrowUp') {
      e.preventDefault();
      const prev = Math.max(currentIdx - 1, 0);
      selectedId = allIds[prev];
      scrollIntoView(selectedId);
    } else if (e.key === 'Enter' && selectedId) {
      e.preventDefault();
      const evt = flatAgents.find(a => a.agentId === selectedId);
      if (evt) dispatch('select', evt);
    }
  }

  function scrollIntoView(id) {
    const el = document.querySelector(`[data-agent-id="${id}"]`);
    if (el) el.scrollIntoView({ block: 'nearest' });
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<div class="feed">
  <div class="feed-scroll">
    {#each repoGroups as repo}
      <div class="repo-group">
        <!-- Repo header -->
        <div class="repo-header">
          <div class="repo-accent" style="background: {repoStatusColor(repo)}" />
          <span class="repo-name">{repo.name}</span>
          {#if repo.branch}
            <span class="repo-branch">⎇ {repo.branch}</span>
          {/if}
          <span class="repo-stats mono">
            {repo.agents.length} agent{repo.agents.length !== 1 ? 's' : ''} · {formatTokens(repoTokens(repo))}
          </span>
        </div>

        <!-- Agents -->
        {#each repo.agents as agent}
          <div
            class="agent-row"
            class:selected={selectedId === agent.agentId}
            data-agent-id={agent.agentId}
            on:click={() => handleClick(agent)}
            on:keydown={(e) => { if (e.key === 'Enter') handleClick(agent); }}
            role="button"
            tabindex="0"
          >
            <div class="agent-stripe" style="background: {statusColor(agent.eventType)}" />
            <div class="agent-content">
              <span class="agent-indicator">●</span>
              <span class="agent-model">{agent.model || agent.agentName}</span>
              <StatusBadge status={agent.eventType} size="sm" />
              <span class="agent-summary">{agent.summary}</span>
              <span class="agent-tokens mono">{formatTokens(agent.tokensUsed || 0)}</span>
              <span class="agent-elapsed mono">{formatElapsed(agent.timestamp)}</span>
            </div>
          </div>

          <!-- Sub-agents (indented) -->
          {#if agent.subAgents && agent.subAgents.length > 0}
            {#each agent.subAgents as sub}
              <div
                class="sub-agent-row"
                class:selected={selectedId === sub.agentId}
                data-agent-id={sub.agentId}
                on:click={() => handleClick(sub)}
                on:keydown={(e) => { if (e.key === 'Enter') handleClick(sub); }}
                role="button"
                tabindex="0"
              >
                <div class="sub-stripe" style="background: {statusColor(sub.eventType)}" />
                <div class="sub-content">
                  <span class="tree-line">├─</span>
                  <span class="sub-indicator">·</span>
                  <span class="sub-name">{sub.agentName}</span>
                  <StatusBadge status={sub.eventType} size="sm" />
                  <span class="sub-summary">{sub.summary}</span>
                  <span class="agent-elapsed mono">{formatElapsed(sub.timestamp)}</span>
                </div>
              </div>
            {/each}
          {/if}
        {/each}

        <!-- New Session button per repo -->
        {#if repo.path}
          <button
            class="new-session-btn"
            on:click|stopPropagation={() => spawnInRepo(repo)}
            disabled={spawningRepo === repo.path}
          >
            <span class="new-session-icon">+</span>
            {spawningRepo === repo.path ? 'Spawning...' : 'New Session'}
          </button>
        {/if}
      </div>
    {/each}

    {#if notifications.length === 0}
      <div class="empty">
        <div class="empty-icon">⬡</div>
        <div class="empty-text">No active agents</div>
        <div class="empty-sub">Press <kbd>⌘N</kbd> to spawn a new agent, or start a Claude Code session</div>
      </div>
    {/if}
  </div>

  <div class="status-bar">
    <span>{totalAgents} agent{totalAgents !== 1 ? 's' : ''}</span>
    <span class="sep">·</span>
    <span>{totalRepos} repo{totalRepos !== 1 ? 's' : ''}</span>
    <span class="sep">·</span>
    <span class="mono">{formatTokens(totalTokens)} tokens</span>
    <span class="keys">
      <kbd>j</kbd>/<kbd>k</kbd> navigate · <kbd>Enter</kbd> open ·
      <button class="spawn-btn" on:click={() => dispatch('spawn')}>
        <kbd>⌘N</kbd> Spawn Agent
      </button>
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

  /* Repo group */
  .repo-group {
    margin-bottom: 2px;
  }

  .repo-header {
    display: flex;
    align-items: center;
    gap: var(--sp-sm);
    padding: var(--sp-sm) var(--sp-lg);
    background: var(--bg-surface);
    border-bottom: 1px solid var(--border-subtle);
    user-select: none;
  }

  .repo-accent {
    width: 3px;
    height: 16px;
    border-radius: 2px;
    flex-shrink: 0;
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

  /* Agent row */
  .agent-row {
    display: flex;
    align-items: stretch;
    cursor: pointer;
    transition: background 100ms ease-out;
  }

  .agent-row:hover { background: var(--bg-surface); }
  .agent-row.selected { background: var(--bg-elevated); }

  .agent-stripe {
    width: 3px;
    flex-shrink: 0;
  }

  .agent-content {
    display: flex;
    align-items: center;
    gap: var(--sp-sm);
    padding: var(--sp-xs) var(--sp-lg);
    padding-left: 20px;
    flex: 1;
    min-width: 0;
  }

  .agent-indicator {
    color: var(--accent-green);
    font-size: 8px;
    flex-shrink: 0;
  }

  .agent-model {
    font-family: var(--font-mono);
    font-size: var(--text-body);
    color: var(--text-primary);
    flex-shrink: 0;
  }

  .agent-summary {
    font-size: var(--text-body);
    color: var(--text-dim);
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .agent-tokens {
    font-size: var(--text-label);
    color: var(--text-dim);
    flex-shrink: 0;
  }

  .agent-elapsed {
    font-size: var(--text-label);
    color: var(--text-muted);
    flex-shrink: 0;
    min-width: 32px;
    text-align: right;
  }

  /* Sub-agent row */
  .sub-agent-row {
    display: flex;
    align-items: stretch;
    cursor: pointer;
    transition: background 100ms ease-out;
  }

  .sub-agent-row:hover { background: var(--bg-surface); }
  .sub-agent-row.selected { background: var(--bg-elevated); }

  .sub-stripe {
    width: 3px;
    flex-shrink: 0;
    opacity: 0.5;
  }

  .sub-content {
    display: flex;
    align-items: center;
    gap: var(--sp-xs);
    padding: 2px var(--sp-lg);
    padding-left: 32px;
    flex: 1;
    min-width: 0;
  }

  .tree-line {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--text-muted);
    flex-shrink: 0;
  }

  .sub-indicator {
    color: var(--text-dim);
    font-size: 8px;
    flex-shrink: 0;
  }

  .sub-name {
    font-family: var(--font-mono);
    font-size: 12px;
    color: var(--text-dim);
    flex-shrink: 0;
  }

  .sub-summary {
    font-size: 12px;
    color: var(--text-muted);
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
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
    font-size: 40px;
    color: var(--accent-green);
    margin-bottom: var(--sp-lg);
    opacity: 0.3;
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

  .empty-sub :global(kbd) {
    display: inline-block;
    padding: 0 4px;
    background: var(--bg-active);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-sm);
    font-family: var(--font-mono);
    font-size: 10px;
    color: var(--text-dim);
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

  .sep { color: var(--text-muted); }
  .mono { font-family: var(--font-mono); }

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

  .spawn-btn {
    background: none;
    border: none;
    color: var(--accent-green);
    font-family: var(--font-ui);
    font-size: 11px;
    cursor: pointer;
    padding: 0;
    display: inline-flex;
    align-items: center;
    gap: 4px;
  }

  .spawn-btn:hover { opacity: 0.8; }

  .new-session-btn {
    display: flex;
    align-items: center;
    gap: 6px;
    width: 100%;
    padding: 6px 20px 6px 20px;
    background: none;
    border: none;
    border-top: 1px dashed var(--border-subtle);
    color: var(--text-muted);
    font-family: var(--font-mono);
    font-size: 12px;
    cursor: pointer;
    transition: all 100ms ease-out;
  }

  .new-session-btn:hover {
    color: var(--accent-green);
    background: rgba(0, 229, 122, 0.04);
  }

  .new-session-btn:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }

  .new-session-icon {
    font-size: 14px;
    font-weight: 300;
  }
</style>
