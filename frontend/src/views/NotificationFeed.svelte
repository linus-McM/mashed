<script>
  import { onMount, onDestroy } from 'svelte';
  import { createEventDispatcher } from 'svelte';
  import { SpawnAgent, SpawnTerminal, KillAgent, GitCommit, GitCommitAndPush, GitCommitPushAndPR, SpawnPRReview, RepoStatus } from '../../wailsjs/go/main/App.js';
  import { EventsOn } from '../../wailsjs/runtime/runtime.js';
  import { GripVertical, GitBranch, Trash2, Plus, Hexagon, Circle, GitCommit as GitCommitIcon, Upload, GitPullRequest, ShieldAlert, GitBranchPlus, TerminalSquare } from 'lucide-svelte';
  import BranchModal from './BranchModal.svelte';
  import SwitchBranchModal from './SwitchBranchModal.svelte';
  import StatusBadge from '../components/StatusBadge.svelte';
  import SparkLine from '../components/SparkLine.svelte';

  const dispatch = createEventDispatcher();

  let spawningRepo = null; // repo path currently spawning

  /** @type {any[]} Notifications passed from App.svelte */
  export let notifications = [];

  let selectedId = null;
  let colorPickerRepo = null; // repo name with open color picker

  const borderPalette = [
    '#ff5f57', '#febc2e', '#28c840', '#3d9eff', '#9d6fff',
    '#00c4b3', '#f0a500', '#e84545', '#00e57a', '#ff6ec7',
    '#1e2530', // "none" — matches border-subtle
  ];

  // Load saved border colors from localStorage
  let repoBorderColors = {};
  try {
    const saved = localStorage.getItem('mashed:repoBorderColors');
    if (saved) repoBorderColors = JSON.parse(saved);
  } catch (_) {}

  function getRepoColor(name) {
    return repoBorderColors[name] || '#1e2530';
  }

  function setRepoColor(name, color) {
    repoBorderColors[name] = color;
    repoBorderColors = repoBorderColors; // trigger reactivity
    localStorage.setItem('mashed:repoBorderColors', JSON.stringify(repoBorderColors));
    colorPickerRepo = null;
  }

  function toggleColorPicker(name) {
    colorPickerRepo = colorPickerRepo === name ? null : name;
  }

  // Drag and drop reordering
  let dragRepo = null;
  let dragOverRepo = null;

  // Load saved repo order from localStorage
  let repoOrder = [];
  try {
    const saved = localStorage.getItem('mashed:repoOrder');
    if (saved) repoOrder = JSON.parse(saved);
  } catch (_) {}

  function onDragStart(e, repoName) {
    dragRepo = repoName;
    e.dataTransfer.effectAllowed = 'move';
    e.dataTransfer.setData('text/plain', repoName);
  }

  function onDragOver(e, repoName) {
    e.preventDefault();
    e.dataTransfer.dropEffect = 'move';
    if (repoName !== dragRepo) {
      dragOverRepo = repoName;
    }
  }

  function onDragLeave() {
    dragOverRepo = null;
  }

  function onDrop(e, targetName) {
    e.preventDefault();
    if (!dragRepo || dragRepo === targetName) {
      dragRepo = null;
      dragOverRepo = null;
      return;
    }

    // Build current display order names
    const names = orderedRepos.map(r => r.name);
    const fromIdx = names.indexOf(dragRepo);
    const toIdx = names.indexOf(targetName);
    if (fromIdx === -1 || toIdx === -1) return;

    // Reorder
    names.splice(fromIdx, 1);
    names.splice(toIdx, 0, dragRepo);

    repoOrder = names;
    localStorage.setItem('mashed:repoOrder', JSON.stringify(repoOrder));

    dragRepo = null;
    dragOverRepo = null;
  }

  function onDragEnd() {
    dragRepo = null;
    dragOverRepo = null;
  }

  // Group: repo → agents → sub-agents
  $: repoGroups = buildRepoTree(notifications);

  // Apply manual order on top of the default sort
  $: orderedRepos = applyRepoOrder(repoGroups, repoOrder);

  function applyRepoOrder(groups, order) {
    if (!order || order.length === 0) return groups;
    const byName = new Map(groups.map(r => [r.name, r]));
    const result = [];
    // Add repos in saved order first
    for (const name of order) {
      if (byName.has(name)) {
        result.push(byName.get(name));
        byName.delete(name);
      }
    }
    // Append any new repos not in saved order
    for (const r of byName.values()) {
      result.push(r);
    }
    return result;
  }

  $: flatAgents = orderedRepos.flatMap(r => r.agents);
  $: totalAgents = flatAgents.length;
  $: totalRepos = orderedRepos.length;
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

  let spawningTerminal = null;

  async function spawnTerminalInRepo(repo) {
    if (spawningTerminal) return;
    spawningTerminal = repo.path;
    try {
      const target = await SpawnTerminal(repo.path);
      const termSession = {
        agentId: `term-${Date.now()}`,
        agentName: 'terminal',
        model: 'terminal',
        repoName: repo.name,
        repoPath: repo.path,
        repoBranch: repo.branch,
        eventType: 'terminal',
        tmuxTarget: target,
        tokensUsed: 0,
        tokensMax: 0,
        summary: 'Shell session',
        priority: 10,
      };
      // Add to notifications so it shows in the repo panel
      notifications = [...notifications, termSession];
      dispatch('select', termSession);
    } catch (e) {
      console.error('Terminal spawn failed:', e);
    } finally {
      spawningTerminal = null;
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

  let killingAgents = new Set();

  async function killSession(agent, e) {
    e.stopPropagation();
    if (killingAgents.has(agent.agentId)) return;
    killingAgents.add(agent.agentId);
    killingAgents = killingAgents; // trigger reactivity
    try {
      await KillAgent(agent.agentId, agent.pid || 0);
    } catch (err) {
      console.error('Kill failed:', err);
    }
    killingAgents.delete(agent.agentId);
    killingAgents = killingAgents;
  }

  // Listen for agent removal events from the backend
  EventsOn('agent:removed', (agentId) => {
    notifications = notifications.filter(n => n.agentId !== agentId);
  });

  // Repo git status: repoPath -> { dirty: bool, openPRs: number }
  let repoStatuses = {};
  let statusInterval;

  async function refreshRepoStatuses() {
    for (const repo of orderedRepos) {
      if (!repo.path) continue;
      try {
        const status = await RepoStatus(repo.path);
        repoStatuses[repo.path] = status;
      } catch (_) {}
    }
    repoStatuses = repoStatuses;
  }

  onMount(() => {
    refreshRepoStatuses();
    statusInterval = setInterval(refreshRepoStatuses, 10000);
  });

  onDestroy(() => {
    if (statusInterval) clearInterval(statusInterval);
  });

  function isDirty(path) {
    return repoStatuses[path]?.dirty || false;
  }

  function hasOpenPR(path) {
    return (repoStatuses[path]?.openPRs || 0) > 0;
  }

  // Branch modal state
  let branchModalRepo = null; // { path, branch } or null
  let switchModalRepo = null; // { path, branch, color } or null

  function openBranchModal(repo) {
    branchModalRepo = { path: repo.path, branch: repo.branch };
  }

  function openSwitchModal(repo) {
    switchModalRepo = { path: repo.path, branch: repo.branch, color: getRepoColor(repo.name) };
  }

  function onBranchSwitched(e) {
    const newBranch = e.detail?.branch;
    if (newBranch && switchModalRepo) {
      notifications = notifications.map(n => {
        if (n.repoPath === switchModalRepo.path) {
          return { ...n, repoBranch: newBranch };
        }
        return n;
      });
    }
    switchModalRepo = null;
    refreshRepoStatuses();
  }

  function onBranchCreated(e) {
    const newBranch = e.detail?.branch;
    if (newBranch) {
      // Update branch in all notifications for this repo so it shows immediately
      notifications = notifications.map(n => {
        if (n.repoPath === branchModalRepo.path) {
          return { ...n, repoBranch: newBranch };
        }
        return n;
      });
    }
    branchModalRepo = null;
    refreshRepoStatuses();
  }

  // Repo action states: repoPath -> { action: string, result: string, error: string }
  let repoActions = {};

  function getAction(path) {
    return repoActions[path] || { action: null, result: null, error: null };
  }

  async function runRepoAction(path, actionName, fn) {
    repoActions[path] = { action: actionName, result: null, error: null };
    repoActions = repoActions;
    try {
      const result = await fn(path);
      repoActions[path] = { action: null, result: result || 'Done', error: null };
    } catch (err) {
      repoActions[path] = { action: null, result: null, error: err?.message || String(err) };
    }
    repoActions = repoActions;
    refreshRepoStatuses();
    // Clear result/error after 5s
    setTimeout(() => {
      if (repoActions[path] && !repoActions[path].action) {
        repoActions[path] = { action: null, result: null, error: null };
        repoActions = repoActions;
      }
    }, 5000);
  }
</script>

<svelte:window on:keydown={handleKeydown} on:click={() => colorPickerRepo = null} />

<div class="feed">
  <div class="feed-scroll">
    {#each orderedRepos as repo (repo.name)}
      <div
        class="repo-group"
        class:drag-over={dragOverRepo === repo.name}
        style="border-color: {getRepoColor(repo.name)}"
      >
        <!-- Repo header (draggable) -->
        <div
          class="repo-header"
          draggable="true"
          on:dragstart={(e) => onDragStart(e, repo.name)}
          on:dragover={(e) => onDragOver(e, repo.name)}
          on:dragleave={onDragLeave}
          on:drop={(e) => onDrop(e, repo.name)}
          on:dragend={onDragEnd}
        >
          <span class="drag-handle" style="color: {getRepoColor(repo.name) !== '#1e2530' ? getRepoColor(repo.name) : ''}"><GripVertical size={14} /></span>
          <span class="repo-name">{repo.name}</span>
          {#if repo.branch}
            <span class="repo-branch" style="color: {getRepoColor(repo.name) !== '#1e2530' ? getRepoColor(repo.name) : ''}"><GitBranch size={12} /> {repo.branch}</span>
          {/if}
          <span class="repo-stats mono">
            {repo.agents.length} agent{repo.agents.length !== 1 ? 's' : ''} · {formatTokens(repoTokens(repo))}
          </span>
          <div class="color-picker-wrap">
            <button
              class="color-picker-btn"
              style="background: {getRepoColor(repo.name)}"
              on:click|stopPropagation={() => toggleColorPicker(repo.name)}
              title="Change border color"
            />
            {#if colorPickerRepo === repo.name}
              <div class="color-picker-popover">
                {#each borderPalette as color}
                  <button
                    class="color-swatch"
                    class:active={getRepoColor(repo.name) === color}
                    style="background: {color}"
                    on:click|stopPropagation={() => setRepoColor(repo.name, color)}
                  />
                {/each}
              </div>
            {/if}
          </div>
        </div>

        <div class="repo-body">
          <!-- Left: Agents (75%) -->
          <div class="repo-agents">
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
                <div class="agent-content">
                  <span class="agent-indicator">
                    {#if agent.eventType === 'terminal'}
                      <TerminalSquare size={12} />
                    {:else}
                      <Circle size={8} />
                    {/if}
                  </span>
                  <span class="agent-model">{agent.eventType === 'terminal' ? 'shell' : (agent.model || agent.agentName)}</span>
                  <StatusBadge status={agent.eventType} size="sm" />
                  <span class="agent-summary">{agent.summary}</span>
                  {#if agent.eventType !== 'terminal'}
                    <span class="agent-tokens mono">{formatTokens(agent.tokensUsed || 0)}</span>
                  {/if}
                  <span class="agent-elapsed mono">{formatElapsed(agent.timestamp)}</span>
                  <button
                    class="kill-btn"
                    title="Kill session"
                    disabled={killingAgents.has(agent.agentId)}
                    on:click={(e) => killSession(agent, e)}
                  ><Trash2 size={12} /></button>
                </div>
              </div>

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
                    <div class="sub-content">
                      <span class="tree-line">├─</span>
                      <span class="sub-indicator">·</span>
                      <span class="sub-name">{sub.agentName}</span>
                      <StatusBadge status={sub.eventType} size="sm" />
                      <span class="sub-summary">{sub.summary}</span>
                      <span class="agent-elapsed mono">{formatElapsed(sub.timestamp)}</span>
                      <button
                        class="kill-btn"
                        title="Kill session"
                        disabled={killingAgents.has(sub.agentId)}
                        on:click={(e) => killSession(sub, e)}
                      ><Trash2 size={12} /></button>
                    </div>
                  </div>
                {/each}
              {/if}
            {/each}

          </div>

          <!-- Right: Actions sidebar (25%) -->
          <div class="repo-actions">
            <button
              class="actions-branch"
              style="color: {getRepoColor(repo.name) !== '#1e2530' ? getRepoColor(repo.name) : 'var(--text-dim)'}"
              on:click|stopPropagation={() => openSwitchModal(repo)}
              title="Switch branch"
            >
              <GitBranch size={12} />
              <span class="actions-branch-name">{repo.branch || 'detached'}</span>
            </button>

            <div class="actions-buttons">
              <button
                class="action-btn"
                on:click|stopPropagation={() => openBranchModal(repo)}
                title="Create a new branch"
              >
                <GitBranchPlus size={14} />
                <span>Branch</span>
              </button>
              <button
                class="action-btn"
                class:action-hot={isDirty(repo.path)}
                disabled={!!getAction(repo.path).action}
                on:click|stopPropagation={() => runRepoAction(repo.path, 'commit', GitCommit)}
                title="Stage all + AI commit message + commit"
              >
                <GitCommitIcon size={14} />
                <span>{getAction(repo.path).action === 'commit' ? 'Committing...' : 'Commit'}</span>
              </button>
              <button
                class="action-btn"
                disabled={!!getAction(repo.path).action}
                on:click|stopPropagation={() => runRepoAction(repo.path, 'push', GitCommitAndPush)}
                title="Commit + push to origin"
              >
                <Upload size={14} />
                <span>{getAction(repo.path).action === 'push' ? 'Pushing...' : 'Push'}</span>
              </button>
              <button
                class="action-btn"
                disabled={!!getAction(repo.path).action}
                on:click|stopPropagation={() => runRepoAction(repo.path, 'pr', GitCommitPushAndPR)}
                title="Commit + push + create PR"
              >
                <GitPullRequest size={14} />
                <span>{getAction(repo.path).action === 'pr' ? 'Creating PR...' : 'PR'}</span>
              </button>
              <button
                class="action-btn action-review"
                class:action-hot={hasOpenPR(repo.path)}
                disabled={!!getAction(repo.path).action}
                on:click|stopPropagation={() => runRepoAction(repo.path, 'review', SpawnPRReview)}
                title="Spawn adversarial PR review agent"
              >
                <ShieldAlert size={14} />
                <span>{getAction(repo.path).action === 'review' ? 'Spawning...' : 'Review'}</span>
              </button>
            </div>

            {#if getAction(repo.path).result}
              <div class="action-result">{getAction(repo.path).result}</div>
            {/if}
            {#if getAction(repo.path).error}
              <div class="action-error">{getAction(repo.path).error}</div>
            {/if}
          </div>
        </div>

        {#if repo.path}
          <div class="spawn-row">
            <button
              class="new-session-btn"
              style="color: {getRepoColor(repo.name) !== '#1e2530' ? getRepoColor(repo.name) : ''}"
              on:click|stopPropagation={() => spawnInRepo(repo)}
              disabled={spawningRepo === repo.path}
            >
              <span class="new-session-icon"><Plus size={14} /></span>
              {spawningRepo === repo.path ? 'Spawning...' : 'New Session'}
            </button>
            <button
              class="new-session-btn"
              style="color: {getRepoColor(repo.name) !== '#1e2530' ? getRepoColor(repo.name) : ''}"
              on:click|stopPropagation={() => spawnTerminalInRepo(repo)}
              disabled={spawningTerminal === repo.path}
            >
              <span class="new-session-icon"><TerminalSquare size={14} /></span>
              {spawningTerminal === repo.path ? 'Opening...' : 'Terminal'}
            </button>
          </div>
        {/if}
      </div>
    {/each}

    {#if notifications.length === 0}
      <div class="empty">
        <div class="empty-icon"><Hexagon size={40} /></div>
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

{#if branchModalRepo}
  <BranchModal
    repoPath={branchModalRepo.path}
    repoBranch={branchModalRepo.branch}
    on:created={onBranchCreated}
    on:cancel={() => branchModalRepo = null}
  />
{/if}

{#if switchModalRepo}
  <SwitchBranchModal
    repoPath={switchModalRepo.path}
    currentBranch={switchModalRepo.branch}
    repoColor={switchModalRepo.color}
    on:switched={onBranchSwitched}
    on:cancel={() => switchModalRepo = null}
  />
{/if}

<script context="module">
  const statusColors = {
    running:        '#39ff14',
    open:           '#00c4b3',
    finished:       'var(--accent-amber)',
    needs_response: 'var(--accent-red)',
    waiting:        'var(--accent-red)',
    error:          'var(--accent-red)',
    completed:      'var(--accent-blue)',
    started:        'var(--accent-purple)',
    terminal:       'var(--text-dim)',
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
    margin: var(--sp-sm) var(--sp-lg);
    border: 1.5px solid var(--border-subtle);
    border-radius: var(--radius-lg);
    overflow: hidden;
    transition: border-color 200ms ease, transform 150ms ease, opacity 150ms ease;
  }

  .repo-group.drag-over {
    border-top: 2px solid var(--accent-green);
  }

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
    gap: 6px;
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
    border-color: #fff;
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

  .kill-btn {
    background: none;
    border: none;
    color: var(--text-dim);
    font-size: 11px;
    cursor: pointer;
    padding: 2px 4px;
    border-radius: var(--radius-sm);
    transition: color 100ms ease;
    flex-shrink: 0;
    line-height: 1;
  }

  .kill-btn:hover { color: var(--accent-red); }

  .kill-btn:disabled {
    opacity: 0.3;
    cursor: not-allowed;
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

  .spawn-row {
    display: flex;
    border-top: 1px dashed var(--border-subtle);
  }

  .spawn-row .new-session-btn {
    border-top: none;
  }

  .spawn-row .new-session-btn + .new-session-btn {
    border-left: 1px dashed var(--border-subtle);
  }

  .new-session-btn {
    display: flex;
    align-items: center;
    gap: 6px;
    flex: 1;
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

  /* Repo body: 75/25 split */
  .repo-body {
    display: flex;
    min-height: 48px;
  }

  .repo-agents {
    flex: 3;
    min-width: 0;
    border-right: 1px solid var(--border-subtle);
  }

  .repo-actions {
    flex: 1;
    display: flex;
    flex-direction: column;
    padding: var(--sp-sm);
    gap: var(--sp-xs);
    min-width: 140px;
    max-width: 200px;
  }

  .actions-branch {
    display: flex;
    align-items: center;
    gap: var(--sp-xs);
    color: var(--text-dim);
    font-family: var(--font-mono);
    font-size: var(--text-label);
    padding: var(--sp-2xs) 0 var(--sp-xs);
    border: none;
    border-bottom: 1px solid var(--border-subtle);
    margin-bottom: var(--sp-2xs);
    background: none;
    cursor: pointer;
    transition: opacity 100ms ease;
    width: 100%;
    text-align: left;
  }

  .actions-branch:hover { opacity: 0.8; }

  .actions-branch-name {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .actions-buttons {
    display: flex;
    flex-direction: column;
    gap: 3px;
  }

  .action-btn {
    display: flex;
    align-items: center;
    gap: 6px;
    width: 100%;
    padding: 4px 8px;
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-sm);
    color: var(--text-dim);
    font-family: var(--font-mono);
    font-size: 11px;
    cursor: pointer;
    transition: all 100ms ease;
  }

  .action-btn:hover {
    color: var(--text-primary);
    border-color: var(--border-emphasis);
    background: var(--bg-active);
  }

  .action-btn:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }

  .action-btn.action-hot {
    color: #39ff14;
    border-color: rgba(57, 255, 20, 0.3);
    text-shadow: 0 0 6px rgba(57, 255, 20, 0.4);
  }

  .action-btn.action-hot:hover {
    color: #39ff14;
    border-color: #39ff14;
    background: rgba(57, 255, 20, 0.08);
    text-shadow: 0 0 10px rgba(57, 255, 20, 0.6);
  }

  .action-btn.action-review.action-hot {
    color: #39ff14;
    border-color: rgba(57, 255, 20, 0.3);
    text-shadow: 0 0 6px rgba(57, 255, 20, 0.4);
  }

  .action-btn.action-review.action-hot:hover {
    color: #39ff14;
    border-color: #39ff14;
    background: rgba(57, 255, 20, 0.08);
    text-shadow: 0 0 10px rgba(57, 255, 20, 0.6);
  }

  .action-result {
    font-family: var(--font-mono);
    font-size: 10px;
    color: var(--accent-green);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    padding: 2px 0;
  }

  .action-error {
    font-family: var(--font-mono);
    font-size: 10px;
    color: var(--accent-red);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    padding: 2px 0;
  }
</style>
