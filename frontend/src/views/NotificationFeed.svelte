<script>
  import { onMount, onDestroy } from 'svelte';
  import { createEventDispatcher } from 'svelte';
  import { fly, slide } from 'svelte/transition';
  import { cubicOut } from 'svelte/easing';
  import { SpawnAgent, SpawnAgentWithCommand, SpawnTerminal, KillAgent, GitCommit, GitCommitAndPush, GitCommitPushAndPR, GitCommitStreaming, GitPull, GitPush, SpawnPRReview, RepoStatus } from '../../wailsjs/go/main/App.js';
  import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime.js';
  import { GripVertical, GitBranch, Trash2, Plus, Hexagon, Circle, GitCommit as GitCommitIcon, Upload, GitPullRequest, ShieldAlert, GitBranchPlus, TerminalSquare, ChevronRight, ChevronDown, Download, GitMerge, Workflow } from 'lucide-svelte';
  import BranchModal from './BranchModal.svelte';
  import SwitchBranchModal from './SwitchBranchModal.svelte';
  import MergeModal from './MergeModal.svelte';
  import ForcePushModal from './ForcePushModal.svelte';
  import NewSessionModal from './NewSessionModal.svelte';
  import StatusBadge from '../components/StatusBadge.svelte';
  import { addSession, makeSession } from '../lib/stores/sessions.js';
  import SparkLine from '../components/SparkLine.svelte';

  const dispatch = createEventDispatcher();

  // uiqa-06: entry animations. Literal ms values mirror --duration-* tokens
  // in style.css; Svelte transition props require numeric values.
  const reducedMotion = typeof window !== 'undefined' &&
    window.matchMedia('(prefers-reduced-motion: reduce)').matches;
  const STAGGER_CAP = 400;
  const flyProps = (i) => reducedMotion
    ? { y: 0, duration: 0, delay: 0 }
    : { y: -8, duration: 150, delay: Math.min(i * 50, STAGGER_CAP), easing: cubicOut };
  const slideProps = reducedMotion
    ? { duration: 0 }
    : { duration: 150, easing: cubicOut };

  let spawningRepo = null; // repo path currently spawning

  /** @type {any[]} Notifications passed from App.svelte */
  export let notifications = [];

  // All scanned repos from backend (includes repos with no active agents)
  let allRepos = [];
  EventsOn('repos', (repos) => {
    if (repos && repos.length > 0) {
      allRepos = repos;
    }
  });

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
  let dropPosition = null; // 'above' | 'below'

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
    // Slight delay so the browser captures the drag image before we add opacity
    requestAnimationFrame(() => { dragRepo = repoName; });
  }

  function onDragOver(e, repoName) {
    e.preventDefault();
    e.dataTransfer.dropEffect = 'move';
    if (repoName === dragRepo) {
      dragOverRepo = null;
      dropPosition = null;
      return;
    }
    dragOverRepo = repoName;

    // Determine above/below based on mouse Y relative to the target element center
    const rect = e.currentTarget.getBoundingClientRect();
    const midY = rect.top + rect.height / 2;
    dropPosition = e.clientY < midY ? 'above' : 'below';
  }

  function onDragLeave(e) {
    // Only clear if we actually left the element (not entering a child)
    const related = e.relatedTarget;
    if (!e.currentTarget.contains(related)) {
      dragOverRepo = null;
      dropPosition = null;
    }
  }

  function onDrop(e, targetName) {
    e.preventDefault();
    if (!dragRepo || dragRepo === targetName) {
      dragRepo = null;
      dragOverRepo = null;
      dropPosition = null;
      return;
    }

    const names = orderedRepos.map(r => r.name);
    const fromIdx = names.indexOf(dragRepo);
    let toIdx = names.indexOf(targetName);
    if (fromIdx === -1 || toIdx === -1) return;

    // Remove from old position
    names.splice(fromIdx, 1);

    // Recalculate target index after removal
    toIdx = names.indexOf(targetName);
    if (toIdx === -1) return;

    // Insert above or below the target
    if (dropPosition === 'below') {
      names.splice(toIdx + 1, 0, dragRepo);
    } else {
      names.splice(toIdx, 0, dragRepo);
    }

    repoOrder = names;
    localStorage.setItem('mashed:repoOrder', JSON.stringify(repoOrder));

    dragRepo = null;
    dragOverRepo = null;
    dropPosition = null;
  }

  function onDragEnd() {
    dragRepo = null;
    dragOverRepo = null;
    dropPosition = null;
  }

  // Group: repo → agents → sub-agents, merged with all scanned repos
  $: repoGroups = buildRepoTree(notifications, allRepos);

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

  $: flatAgents = orderedRepos.flatMap(r => r.agents.flatMap(a => {
    const result = [a];
    if (a.subAgents && a.subAgents.length > 0 && isAgentExpanded(a.agentId)) {
      result.push(...a.subAgents);
    }
    return result;
  }));
  $: totalAgents = flatAgents.length;
  $: totalRepos = orderedRepos.length;
  $: totalTokens = notifications.reduce((sum, e) => sum + (e.tokensUsed || 0), 0);

  // uiqa-10 Sub-brief B: ambient status-bar signals.
  // - aggregateSamples concatenates the latest tokenSamples from every agent
  //   across every repo and keeps the most recent 20 values. Timestamps are
  //   intentionally NOT aligned (different agents sample at different rates) —
  //   the story calls this out as an acceptable simplification because the
  //   status sparkline is decorative ambient data, not an analytical chart.
  // - anyRunning gates the pulse dot: hidden when nothing is live, visible
  //   (and animated unless reduced-motion) whenever the fleet is active.
  $: aggregateSamples = (() => {
    const all = [];
    for (const group of orderedRepos) {
      for (const agent of (group.agents || [])) {
        if (agent.tokenSamples?.length) all.push(...agent.tokenSamples);
      }
    }
    return all.slice(-20);
  })();
  $: anyRunning = orderedRepos.some(
    (g) => (g.agents || []).some((a) => a.eventType === 'running'),
  );

  function buildRepoTree(events, scannedRepos) {
    const repoMap = new Map();

    // Seed with all scanned repos so they always show a panel
    for (const r of (scannedRepos || [])) {
      if (!repoMap.has(r.name)) {
        repoMap.set(r.name, {
          name: r.name,
          path: r.path,
          branch: r.branch,
          agents: [],
          worstStatus: 'idle',
        });
      }
    }

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

    // Sort agents within each repo: running/active on top, then by priority
    const agentPriority = { needs_response: 0, error: 1, running: 2, open: 3, started: 4, finished: 5, completed: 6 };
    for (const repo of repoMap.values()) {
      repo.agents.sort((a, b) => {
        const pa = agentPriority[a.eventType] ?? 7;
        const pb = agentPriority[b.eventType] ?? 7;
        return pa - pb;
      });
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

  // New Session modal state
  let sessionModalRepo = null; // { path, name, branch } or null

  function openSessionModal(repo) {
    sessionModalRepo = { path: repo.path, name: repo.name, branch: repo.branch };
  }

  async function onSessionSpawn(e) {
    const { command, model, repoPath } = e.detail;
    sessionModalRepo = null;
    spawningRepo = repoPath;
    try {
      const target = await SpawnAgentWithCommand(repoPath, command);
      const repoName = repoNameFromDir(repoPath);
      addSession(repoPath, makeSession(target, repoPath, repoName, 'agent', model));
      const agent = {
        agentId: `spawned-${Date.now()}`,
        agentName: model,
        model: model,
        repoName: repoName,
        repoPath: repoPath,
        repoBranch: '',
        eventType: 'running',
        tmuxTarget: target,
        tokensUsed: 0,
        tokensMax: model.includes('opus') ? 1000000 : 200000,
        summary: `New session in ${repoName}`,
      };
      dispatch('notify', agent);
      dispatch('select', agent);
    } catch (e) {
      console.error('Spawn failed:', e);
    } finally {
      spawningRepo = null;
    }
  }

  function repoNameFromDir(dir) {
    if (!dir) return 'unknown';
    const parts = dir.split('/');
    return parts[parts.length - 1] || 'unknown';
  }

  let spawningTerminal = null;

  async function spawnTerminalInRepo(repo) {
    if (spawningTerminal) return;
    spawningTerminal = repo.path;
    try {
      const target = await SpawnTerminal(repo.path);
      addSession(repo.path, makeSession(target, repo.path, repo.name, 'terminal', ''));
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
      // Tell parent to add to its notification list so it persists
      dispatch('notify', termSession);
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

  // Accordion state for sub-agents per parent agent
  let expandedAgents = new Set();

  function toggleAgentAccordion(agentId, e) {
    e.stopPropagation();
    if (expandedAgents.has(agentId)) {
      expandedAgents.delete(agentId);
    } else {
      expandedAgents.add(agentId);
    }
    expandedAgents = expandedAgents;
  }

  function isAgentExpanded(agentId) {
    return expandedAgents.has(agentId);
  }

  let killingAgents = new Set();

  async function killSession(agent, e) {
    e.stopPropagation();
    if (killingAgents.has(agent.agentId)) return;
    killingAgents.add(agent.agentId);
    killingAgents = killingAgents; // trigger reactivity
    try {
      await KillAgent(agent.agentId, agent.pid || 0, agent.tmuxTarget || '');
    } catch (err) {
      console.error('Kill failed:', err);
    }
    killingAgents.delete(agent.agentId);
    killingAgents = killingAgents;
  }

  // Listen for agent removal events from the backend — also remove sub-agents
  EventsOn('agent:removed', (agentId) => {
    notifications = notifications.filter(n =>
      n.agentId !== agentId && n.parentAgentId !== agentId
    );
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

  function isAhead(path) {
    return (repoStatuses[path]?.ahead || 0) > 0;
  }

  function isProtected(path) {
    return repoStatuses[path]?.protected || false;
  }

  // Push with conflict detection — returns true if conflict modal should open
  let forcePushRepo = null; // { path, message } or null

  async function smartPush(path) {
    repoActions[path] = { action: 'push', result: null, error: null };
    repoActions = repoActions;
    try {
      const result = await GitPush(path);
      if (result.startsWith('conflict:')) {
        forcePushRepo = { path, message: result.slice('conflict:'.length) };
        repoActions[path] = { action: null, result: null, error: null };
      } else {
        repoActions[path] = { action: null, result: 'Pushed', error: null };
        setTimeout(() => {
          if (repoActions[path] && !repoActions[path].action) {
            repoActions[path] = { action: null, result: null, error: null };
            repoActions = repoActions;
          }
        }, 5000);
      }
    } catch (err) {
      repoActions[path] = { action: null, result: null, error: err?.message || String(err) };
      setTimeout(() => {
        if (repoActions[path] && !repoActions[path].action) {
          repoActions[path] = { action: null, result: null, error: null };
          repoActions = repoActions;
        }
      }, 5000);
    }
    repoActions = repoActions;
    refreshRepoStatuses();
  }

  function onForcePushed() {
    if (forcePushRepo) {
      repoActions[forcePushRepo.path] = { action: null, result: 'Force pushed', error: null };
      repoActions = repoActions;
      setTimeout(() => {
        if (repoActions[forcePushRepo.path] && !repoActions[forcePushRepo.path].action) {
          repoActions[forcePushRepo.path] = { action: null, result: null, error: null };
          repoActions = repoActions;
        }
      }, 5000);
    }
    forcePushRepo = null;
    refreshRepoStatuses();
  }

  // Collapsed state per repo name
  let collapsedRepos = new Set();

  function toggleCollapse(name) {
    if (collapsedRepos.has(name)) {
      collapsedRepos.delete(name);
    } else {
      collapsedRepos.add(name);
    }
    collapsedRepos = collapsedRepos;
  }

  function isCollapsed(repo) {
    // Explicitly toggled takes priority
    if (collapsedRepos.has(repo.name)) return true;
    // Auto-collapse if no agents and never explicitly opened
    if (repo.agents.length === 0 && !expandedRepos.has(repo.name)) return true;
    return false;
  }

  // Track repos the user has explicitly expanded (so empty repos stay open after expand)
  let expandedRepos = new Set();

  function handleHeaderClick(repo) {
    if (isCollapsed(repo)) {
      collapsedRepos.delete(repo.name);
      expandedRepos.add(repo.name);
    } else {
      collapsedRepos.add(repo.name);
      expandedRepos.delete(repo.name);
    }
    collapsedRepos = collapsedRepos;
    expandedRepos = expandedRepos;
  }

  // Branch modal state
  let branchModalRepo = null; // { path, branch } or null
  let switchModalRepo = null; // { path, branch, color } or null
  let mergeModalRepo = null; // { path, branch } or null

  function openBranchModal(repo) {
    branchModalRepo = { path: repo.path, branch: repo.branch };
  }

  function openSwitchModal(repo) {
    switchModalRepo = { path: repo.path, branch: repo.branch, color: getRepoColor(repo.name) };
  }

  function openMergeModal(repo) {
    mergeModalRepo = { path: repo.path, branch: repo.branch };
  }

  function onMerged(e) {
    const targetBranch = e.detail?.targetBranch;
    if (targetBranch && mergeModalRepo) {
      notifications = notifications.map(n => {
        if (n.repoPath === mergeModalRepo.path) {
          return { ...n, repoBranch: targetBranch };
        }
        return n;
      });
    }
    mergeModalRepo = null;
    refreshRepoStatuses();
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

  // Streaming commit output panel: repoPath -> { lines[], error, explanation, done, visible }
  let commitPanels = {};

  function getCommitPanel(path) {
    return commitPanels[path] || null;
  }

  function startStreamingCommit(path) {
    commitPanels[path] = { lines: [], error: null, explanation: null, done: false, visible: true };
    commitPanels = commitPanels;
    repoActions[path] = { action: 'commit', result: null, error: null };
    repoActions = repoActions;
    GitCommitStreaming(path);
  }

  function closeCommitPanel(path) {
    delete commitPanels[path];
    commitPanels = commitPanels;
  }

  // Listen for streaming commit progress events
  EventsOn('git:commit:progress', (evt) => {
    const path = evt.repoPath;
    if (!commitPanels[path]) {
      commitPanels[path] = { lines: [], error: null, explanation: null, done: false, visible: true };
    }
    const panel = commitPanels[path];

    if (evt.step && !evt.error) {
      panel.lines = [...panel.lines, { step: evt.step, output: evt.output || '' }];
    }
    if (evt.error) {
      panel.error = evt.error;
      panel.explanation = evt.explanation || null;
    }
    panel.done = !!evt.done;

    if (evt.done && !evt.error) {
      repoActions[path] = { action: null, result: evt.output || 'Done', error: null };
      refreshRepoStatuses();
      // Auto-close panel after success
      setTimeout(() => {
        delete commitPanels[path];
        commitPanels = commitPanels;
      }, 1500);
    } else if (evt.done && evt.error) {
      repoActions[path] = { action: null, result: null, error: evt.error };
    }

    commitPanels = commitPanels;
    repoActions = repoActions;
  });
</script>

<svelte:window on:keydown={handleKeydown} on:click={() => colorPickerRepo = null} />

<div class="feed">
  <div class="feed-scroll" role="list">
    {#each orderedRepos as repo, i (repo.name)}
      <div
        class="repo-group"
        role="listitem"
        class:drag-over-above={dragOverRepo === repo.name && dropPosition === 'above'}
        class:drag-over-below={dragOverRepo === repo.name && dropPosition === 'below'}
        class:dragging={dragRepo === repo.name}
        style="border-color: {getRepoColor(repo.name)}"
        in:fly={flyProps(i)}
        on:dragover={(e) => onDragOver(e, repo.name)}
        on:dragleave={(e) => onDragLeave(e)}
        on:drop={(e) => onDrop(e, repo.name)}
      >
        <!-- Repo header (draggable, dblclick to toggle) -->
        <div
          class="repo-header"
          role="listitem"
          class:collapsed={isCollapsed(repo)}
          draggable="true"
          on:dragstart={(e) => onDragStart(e, repo.name)}
          on:dragend={onDragEnd}
        >
          <button class="collapse-btn" on:click|stopPropagation={() => handleHeaderClick(repo)}>
            {#if isCollapsed(repo)}
              <ChevronRight size={14} />
            {:else}
              <ChevronDown size={14} />
            {/if}
          </button>
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

        {#if !isCollapsed(repo)}
        <div class="repo-body" transition:slide={slideProps}>
          <!-- Left: Agents (75%) -->
          <div class="repo-agents">
            {#each repo.agents as agent}
              <div class="agent-accordion" class:has-children={agent.subAgents && agent.subAgents.length > 0}>
                <div
                  class="agent-row"
                  class:selected={selectedId === agent.agentId}
                  class:is-running={agent.eventType === 'running'}
                  data-agent-id={agent.agentId}
                  on:click={() => handleClick(agent)}
                  on:keydown={(e) => { if (e.key === 'Enter') handleClick(agent); }}
                  role="button"
                  tabindex="0"
                >
                  <div class="agent-content">
                    {#if agent.subAgents && agent.subAgents.length > 0}
                      <button class="accordion-toggle" on:click={(e) => toggleAgentAccordion(agent.agentId, e)}>
                        {#if isAgentExpanded(agent.agentId)}
                          <ChevronDown size={12} />
                        {:else}
                          <ChevronRight size={12} />
                        {/if}
                      </button>
                    {:else}
                      <span class="agent-indicator">
                        {#if agent.eventType === 'terminal'}
                          <TerminalSquare size={12} />
                        {:else}
                          <Circle size={8} />
                        {/if}
                      </span>
                    {/if}
                    <span class="agent-model">{agent.eventType === 'terminal' ? 'shell' : (agent.model || agent.agentName)}</span>
                    <StatusBadge status={agent.eventType} size="sm" />
                    {#if agent.subAgents && agent.subAgents.length > 0}
                      <span class="sub-count">{agent.subAgents.length} sub</span>
                    {/if}
                    <span class="agent-summary">{agent.summary}</span>
                    {#if agent.eventType !== 'terminal' && agent.tokenSamples?.length > 1}
                      <span
                        class="sparkline-wrap"
                        class:dimmed={agent.eventType !== 'running'}
                        title={`Token history: ${agent.tokenSamples[0]} \u2192 ${agent.tokenSamples[agent.tokenSamples.length - 1]} over last ${agent.tokenSamples.length} samples`}
                      >
                        <SparkLine data={agent.tokenSamples} />
                      </span>
                    {/if}
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

                {#if agent.subAgents && agent.subAgents.length > 0 && isAgentExpanded(agent.agentId)}
                  <div class="sub-agent-accordion">
                    {#each agent.subAgents as sub, i}
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
                          <span class="tree-line">{i < agent.subAgents.length - 1 ? '├─' : '└─'}</span>
                          <span class="sub-indicator" class:done={sub.subAgentStatus === 'done'}>
                            {#if sub.subAgentStatus === 'done'}
                              <Circle size={6} />
                            {:else}
                              <span class="sub-pulse" />
                            {/if}
                          </span>
                          <span class="sub-name">{sub.subAgentName || sub.agentName}</span>
                          <StatusBadge status={sub.eventType} size="sm" />
                          <span class="sub-summary" title={sub.subAgentDesc || sub.summary}>{sub.subAgentDesc || sub.summary}</span>
                          <span class="agent-elapsed mono">{formatElapsed(sub.timestamp)}</span>
                        </div>
                      </div>
                    {/each}
                  </div>
                {/if}
              </div>
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
                class:glow-btn={isDirty(repo.path)}
                disabled={!!getAction(repo.path).action}
                on:click|stopPropagation={() => startStreamingCommit(repo.path)}
                title="Stage all + AI commit message + commit"
              >
                <GitCommitIcon size={14} />
                <span>{getAction(repo.path).action === 'commit' ? 'Committing...' : 'Commit'}</span>
              </button>
              <button
                class="action-btn"
                disabled={!!getAction(repo.path).action}
                on:click|stopPropagation={() => runRepoAction(repo.path, 'pull', GitPull)}
                title="Pull remote changes"
              >
                <Download size={14} />
                <span>{getAction(repo.path).action === 'pull' ? 'Pulling...' : 'Pull'}</span>
              </button>
              <button
                class="action-btn"
                class:glow-btn={isAhead(repo.path) && !isProtected(repo.path)}
                disabled={!!getAction(repo.path).action}
                on:click|stopPropagation={() => smartPush(repo.path)}
                title={isProtected(repo.path) ? 'Branch is protected — push via PR' : isAhead(repo.path) ? `${repoStatuses[repo.path]?.ahead} commit(s) ahead of remote` : 'Push to origin'}
              >
                <Upload size={14} />
                <span>{getAction(repo.path).action === 'push' ? 'Pushing...' : 'Push'}</span>
              </button>
              <button
                class="action-btn"
                disabled={!!getAction(repo.path).action}
                on:click|stopPropagation={() => openMergeModal(repo)}
                title="Merge current branch into another"
              >
                <GitMerge size={14} />
                <span>Merge</span>
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
                class:glow-btn={hasOpenPR(repo.path)}
                disabled={!!getAction(repo.path).action}
                on:click|stopPropagation={() => runRepoAction(repo.path, 'review', SpawnPRReview)}
                title="Spawn adversarial PR review agent"
              >
                <ShieldAlert size={14} />
                <span>{getAction(repo.path).action === 'review' ? 'Spawning...' : 'Review'}</span>
              </button>
            </div>

            {#if !getCommitPanel(repo.path)}
              {#if getAction(repo.path).result}
                <div class="action-result">{getAction(repo.path).result}</div>
              {/if}
              {#if getAction(repo.path).error && !getCommitPanel(repo.path)}
                <div class="action-error">{getAction(repo.path).error}</div>
              {/if}
            {/if}
          </div>

          <!-- Commit output panel -->
          {#if getCommitPanel(repo.path)}
            {@const panel = getCommitPanel(repo.path)}
            <div class="commit-panel" style="border-color: {getRepoColor(repo.name)}">
              <div class="commit-panel-header">
                <span class="commit-panel-title">
                  {#if panel.done && !panel.error}
                    Committed
                  {:else if panel.error}
                    Commit Failed
                  {:else}
                    Committing...
                  {/if}
                </span>
                {#if panel.done}
                  <button class="commit-panel-close" on:click|stopPropagation={() => closeCommitPanel(repo.path)}>×</button>
                {/if}
              </div>
              <div class="commit-panel-body">
                {#each panel.lines as line}
                  <div class="commit-line">
                    <span class="commit-step">{line.step}</span>
                    {#if line.output}
                      <pre class="commit-output">{line.output}</pre>
                    {/if}
                  </div>
                {/each}
                {#if !panel.done && !panel.error}
                  <div class="commit-line commit-active">
                    <span class="commit-spinner" />
                  </div>
                {/if}
              </div>
              {#if panel.error}
                <div class="commit-error-section">
                  <div class="commit-error-label">Error</div>
                  <pre class="commit-error-text">{panel.error}</pre>
                  {#if panel.explanation}
                    <div class="commit-explain-label">Why this happened</div>
                    <div class="commit-explain-text">{panel.explanation}</div>
                  {:else if !panel.done}
                    <div class="commit-explain-loading">Analyzing failure...</div>
                  {/if}
                </div>
              {/if}
            </div>
          {/if}
        </div>

        {/if}

        {#if repo.path}
          <div class="spawn-row">
            <button
              class="new-session-btn"
              style="color: {getRepoColor(repo.name) !== '#1e2530' ? getRepoColor(repo.name) : ''}"
              on:click|stopPropagation={() => openSessionModal(repo)}
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
            <button
              class="new-session-btn"
              style="color: {getRepoColor(repo.name) !== '#1e2530' ? getRepoColor(repo.name) : ''}"
              on:click|stopPropagation={() => dispatch('open-workspace', repo)}
            >
              <span class="new-session-icon"><Workflow size={14} /></span>
              BMAD Workspace
            </button>
          </div>
        {/if}
      </div>
    {/each}

    {#if orderedRepos.length === 0}
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
    <!-- uiqa-10 Sub-brief B: ambient pulse dot + aggregate sparkline. Both are
         decorative; the pulse dot is gated on anyRunning (disappears when the
         fleet is idle) and the sparkline only renders when there are enough
         samples to draw a meaningful glyph. -->
    {#if anyRunning}
      <span class="status-pulse" aria-label="agents active"></span>
    {/if}
    {#if aggregateSamples.length > 1}
      <span class="status-sparkline" aria-hidden="true">
        <SparkLine data={aggregateSamples} />
      </span>
    {/if}
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

{#if sessionModalRepo}
  <NewSessionModal
    repoPath={sessionModalRepo.path}
    repoName={sessionModalRepo.name}
    on:spawn={onSessionSpawn}
    on:cancel={() => sessionModalRepo = null}
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

{#if mergeModalRepo}
  <MergeModal
    repoPath={mergeModalRepo.path}
    currentBranch={mergeModalRepo.branch}
    on:merged={onMerged}
    on:cancel={() => mergeModalRepo = null}
  />
{/if}

{#if forcePushRepo}
  <ForcePushModal
    repoPath={forcePushRepo.path}
    conflictMessage={forcePushRepo.message}
    on:pushed={onForcePushed}
    on:cancel={() => forcePushRepo = null}
  />
{/if}

<script context="module">
  const statusColors = {
    running:        'var(--accent-green)',
    open:           'var(--accent-teal)',
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
    flex: 1;
    height: 0;
    min-height: 0;
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

  .repo-group.dragging {
    opacity: 0.35;
    transform: scale(0.98);
  }

  .repo-group.drag-over-above {
    position: relative;
  }

  .repo-group.drag-over-above::before {
    content: '';
    position: absolute;
    top: -5px;
    left: 12px;
    right: 12px;
    height: 3px;
    background: var(--accent-green);
    border-radius: 2px;
    box-shadow: 0 0 8px color-mix(in srgb, var(--accent-green) 50%, transparent);
    z-index: 10;
  }

  .repo-group.drag-over-below {
    position: relative;
  }

  .repo-group.drag-over-below::after {
    content: '';
    position: absolute;
    bottom: -5px;
    left: 12px;
    right: 12px;
    height: 3px;
    background: var(--accent-green);
    border-radius: 2px;
    box-shadow: 0 0 8px color-mix(in srgb, var(--accent-green) 50%, transparent);
    z-index: 10;
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
    font-variant-numeric: tabular-nums;
    color: var(--text-dim);
    flex-shrink: 0;
  }

  /* Sparkline wrapper (uiqa-09). Text-based SparkLine sits between the
     summary and the token count so the glyph reads as the history of the
     numeric value it neighbours. flex-shrink:0 prevents long summaries
     from compressing the glyph. */
  .sparkline-wrap {
    flex-shrink: 0;
    align-self: center;
    line-height: 1;
  }

  .sparkline-wrap.dimmed :global(.sparkline) {
    color: var(--text-dim);
  }

  .agent-elapsed {
    font-size: var(--text-label);
    font-variant-numeric: tabular-nums;
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

  /* Agent accordion */
  .agent-accordion {
    border-bottom: 1px solid transparent;
  }

  .agent-accordion.has-children {
    border-bottom: 1px solid var(--border-subtle);
  }

  .agent-accordion.has-children:last-child {
    border-bottom: none;
  }

  .accordion-toggle {
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

  .accordion-toggle:hover { color: var(--accent-green); }

  .sub-count {
    font-family: var(--font-mono);
    font-size: var(--text-label);
    font-variant-numeric: tabular-nums;
    color: var(--text-muted);
    background: var(--bg-active);
    padding: 0 var(--sp-xs);
    border-radius: 8px;
    flex-shrink: 0;
    line-height: 16px;
  }

  /* Sub-agent accordion panel */
  .sub-agent-accordion {
    background: rgba(0, 0, 0, 0.15);
    border-top: 1px solid var(--border-subtle);
    padding: 2px 0;
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
    padding: var(--sp-2xs) var(--sp-lg);
    padding-left: 36px;
    flex: 1;
    min-width: 0;
  }

  .tree-line {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--text-muted);
    flex-shrink: 0;
    user-select: none;
  }

  .sub-indicator {
    color: var(--accent-green);
    font-size: 6px;
    flex-shrink: 0;
  }
  .sub-indicator.done {
    color: var(--accent-red);
  }

  .sub-pulse {
    display: inline-block;
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--accent-green);
    animation: sub-pulse-anim 1.5s ease-in-out infinite;
  }

  @keyframes sub-pulse-anim {
    0%, 100% { opacity: 0.3; }
    50% { opacity: 1; }
  }

  .sub-name {
    font-family: var(--font-mono);
    font-size: var(--text-body);
    color: var(--text-dim);
    flex-shrink: 0;
    max-width: 160px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .sub-summary {
    font-size: var(--text-body);
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
    font-size: var(--text-label);
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

  /* uiqa-10: ambient sparkline in status bar. Opacity 0.7 keeps it secondary
     to the primary counts; color override pulls the glyph into teal to
     distinguish "ambient data" (teal) from "alive/running" (green). */
  .status-sparkline {
    margin-left: var(--sp-md);
    opacity: 0.7;
  }
  .status-sparkline :global(.sparkline) {
    color: var(--accent-teal);
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
    font-size: var(--text-label);
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
    gap: var(--sp-sm);
    flex: 1;
    padding: var(--sp-xs) 20px var(--sp-xs) 20px;
    background: none;
    border: none;
    border-top: 1px dashed var(--border-subtle);
    color: var(--text-primary);
    font-family: var(--font-mono);
    font-size: var(--text-body);
    cursor: pointer;
    transition: all 100ms ease-out;
  }

  .new-session-btn:hover {
    color: var(--accent-green);
    background: color-mix(in srgb, var(--accent-green) 4%, transparent);
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
    gap: var(--sp-2xs);
  }

  .action-btn {
    display: flex;
    align-items: center;
    gap: var(--sp-sm);
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


  .action-result {
    font-family: var(--font-mono);
    font-size: var(--text-label);
    color: var(--accent-green);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    padding: 2px 0;
  }

  .action-error {
    font-family: var(--font-mono);
    font-size: var(--text-label);
    color: var(--accent-red);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    padding: 2px 0;
  }

  /* Commit output panel */
  .commit-panel {
    border-top: 2px solid var(--border-subtle);
    background: var(--bg-deepest);
    max-height: 220px;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .commit-panel-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: var(--sp-xs) var(--sp-lg);
    background: var(--bg-surface);
    border-bottom: 1px solid var(--border-subtle);
    flex-shrink: 0;
  }

  .commit-panel-title {
    font-family: var(--font-mono);
    font-size: 11px;
    font-weight: 600;
    color: var(--text-dim);
  }

  .commit-panel-close {
    background: none;
    border: none;
    color: var(--text-muted);
    font-size: 16px;
    cursor: pointer;
    padding: 0 2px;
    line-height: 1;
  }

  .commit-panel-close:hover { color: var(--text-primary); }

  .commit-panel-body {
    padding: var(--sp-xs) var(--sp-lg);
    overflow-y: auto;
    flex: 1;
    min-height: 0;
  }

  .commit-line {
    padding: 2px 0;
  }

  .commit-step {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--accent-green);
  }

  .commit-output {
    font-family: var(--font-mono);
    font-size: var(--text-label);
    color: var(--text-dim);
    margin: 2px 0 4px 0;
    padding: 4px 8px;
    background: rgba(0, 0, 0, 0.25);
    border-radius: var(--radius-sm);
    white-space: pre-wrap;
    word-break: break-word;
    max-height: 60px;
    overflow-y: auto;
  }

  .commit-active {
    display: flex;
    align-items: center;
    gap: var(--sp-sm);
  }

  .commit-spinner {
    display: inline-block;
    width: 8px;
    height: 8px;
    border: 1.5px solid var(--accent-green);
    border-top-color: transparent;
    border-radius: 50%;
    animation: commit-spin 0.6s linear infinite;
  }

  @keyframes commit-spin {
    to { transform: rotate(360deg); }
  }

  /* Error section */
  .commit-error-section {
    border-top: 1px solid color-mix(in srgb, var(--accent-red) 20%, transparent);
    padding: var(--sp-xs) var(--sp-lg);
    background: color-mix(in srgb, var(--accent-red) 4%, transparent);
    flex-shrink: 0;
  }

  .commit-error-label {
    font-family: var(--font-mono);
    font-size: 9px;
    font-weight: 600;
    color: var(--accent-red);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    margin-bottom: 4px;
  }

  .commit-error-text {
    font-family: var(--font-mono);
    font-size: var(--text-label);
    color: var(--accent-red);
    white-space: pre-wrap;
    word-break: break-word;
    margin: 0 0 8px 0;
    padding: 4px 8px;
    background: color-mix(in srgb, var(--accent-red) 6%, transparent);
    border-radius: var(--radius-sm);
    border: 1px solid color-mix(in srgb, var(--accent-red) 15%, transparent);
    max-height: 60px;
    overflow-y: auto;
  }

  .commit-explain-label {
    font-family: var(--font-mono);
    font-size: 9px;
    font-weight: 600;
    color: var(--accent-amber);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    margin-bottom: 4px;
  }

  .commit-explain-text {
    font-size: 11px;
    color: var(--text-primary);
    line-height: 1.5;
    padding: var(--sp-sm);
    background: color-mix(in srgb, var(--accent-amber) 6%, transparent);
    border-radius: var(--radius-sm);
    border: 1px solid color-mix(in srgb, var(--accent-amber) 15%, transparent);
  }

  .commit-explain-loading {
    font-family: var(--font-mono);
    font-size: var(--text-label);
    color: var(--text-muted);
    font-style: italic;
  }
</style>
