<script lang="ts">
  import { onMount, onDestroy, createEventDispatcher } from 'svelte';
  import { fly, slide } from 'svelte/transition';
  import { cubicOut } from 'svelte/easing';
  import { SpawnAgentWithCommand, SpawnTerminal, KillAgent, GitCommitPushAndPR, GitCommitStreaming, GitPull, GitPush, SpawnPRReview, RepoStatus } from '../../wailsjs/go/main/App.js';
  import { EventsOn } from '../../wailsjs/runtime/runtime.js';
  import { Plus, Hexagon, TerminalSquare, Workflow } from 'lucide-svelte';
  import BranchModal from './BranchModal.svelte';
  import SwitchBranchModal from './SwitchBranchModal.svelte';
  import MergeModal from './MergeModal.svelte';
  import ForcePushModal from './ForcePushModal.svelte';
  import NewSessionModal from './NewSessionModal.svelte';
  import RepoHeader from '../components/feed/RepoHeader.svelte';
  import AgentList from '../components/feed/AgentList.svelte';
  import RepoActions from '../components/feed/RepoActions.svelte';
  import CommitOutputPanel from '../components/feed/CommitOutputPanel.svelte';
  import { addSession, makeSession } from '../lib/stores/sessions';
  import { estimatePtySize } from '../lib/ptySize';
  import SparkLine from '../components/SparkLine.svelte';
  import { REPO_BORDER_NONE } from '../lib/repoPalette';
  import { buildRepoTree, applyRepoOrder, formatTokens, IDLE_ACTION, STATUS_COLORS } from '../lib/feed/repoTree';
  import type { NotificationEntry, RepoDescriptor, RepoGroup, RepoStatusSnapshot, RepoAction, CommitPanel,
    CommitProgressEvent, SessionModalRepo, BranchModalRepo, SwitchModalRepo, MergeModalRepo, ForcePushRepo,
    SessionSpawnDetail, BranchEventDetail, MergeEventDetail } from '../lib/feed/repoTree';

  // Feed entry / repo-tree types and pure tree helpers live in
  // lib/feed/repoTree.ts; repo header, agent rows, actions sidebar and commit
  // panel are components/feed/* (spec R31 split).

  const dispatch = createEventDispatcher<{
    notify: NotificationEntry;
    select: NotificationEntry;
    'open-workspace': RepoGroup;
    spawn: void;
  }>();

  // uiqa-06: entry animations. Literal ms values mirror --duration-* tokens
  // in style.css; Svelte transition props require numeric values.
  const reducedMotion = typeof window !== 'undefined' &&
    window.matchMedia('(prefers-reduced-motion: reduce)').matches;
  const STAGGER_CAP = 400;
  const flyProps = (i: number) => reducedMotion
    ? { y: 0, duration: 0, delay: 0 }
    : { y: -8, duration: 150, delay: Math.min(i * 50, STAGGER_CAP), easing: cubicOut };
  const slideProps = reducedMotion
    ? { duration: 0 }
    : { duration: 150, easing: cubicOut };

  let spawningRepo: string | null = null; // repo path currently spawning

  /** Notifications passed from App.svelte. */
  export let notifications: NotificationEntry[] = [];

  // All scanned repos from backend (includes repos with no active agents)
  let allRepos: RepoDescriptor[] = [];
  EventsOn('repos', (repos: RepoDescriptor[] | undefined) => {
    if (repos && repos.length > 0) {
      allRepos = repos;
    }
  });

  let selectedId: string | null = null;
  let colorPickerRepo: string | null = null; // repo name with open color picker

  // Load saved border colors from localStorage
  let repoBorderColors: Record<string, string> = {};
  try {
    const saved = localStorage.getItem('mashed:repoBorderColors');
    if (saved) repoBorderColors = JSON.parse(saved) as Record<string, string>;
  } catch (_) {}

  function getRepoColor(name: string): string {
    return repoBorderColors[name] || REPO_BORDER_NONE;
  }

  function setRepoColor(name: string, color: string): void {
    repoBorderColors[name] = color;
    repoBorderColors = repoBorderColors; // trigger reactivity
    localStorage.setItem('mashed:repoBorderColors', JSON.stringify(repoBorderColors));
    colorPickerRepo = null;
  }

  function toggleColorPicker(name: string): void {
    colorPickerRepo = colorPickerRepo === name ? null : name;
  }

  // Drag and drop reordering
  let dragRepo: string | null = null;
  let dragOverRepo: string | null = null;
  let dropPosition: 'above' | 'below' | null = null;

  // Load saved repo order from localStorage
  let repoOrder: string[] = [];
  try {
    const saved = localStorage.getItem('mashed:repoOrder');
    if (saved) repoOrder = JSON.parse(saved) as string[];
  } catch (_) {}

  function onDragStart(e: DragEvent, repoName: string): void {
    dragRepo = repoName;
    if (e.dataTransfer) {
      e.dataTransfer.effectAllowed = 'move';
      e.dataTransfer.setData('text/plain', repoName);
    }
    // Slight delay so the browser captures the drag image before we add opacity
    requestAnimationFrame(() => { dragRepo = repoName; });
  }

  function onDragOver(e: DragEvent, repoName: string): void {
    e.preventDefault();
    if (e.dataTransfer) e.dataTransfer.dropEffect = 'move';
    if (repoName === dragRepo) {
      dragOverRepo = null;
      dropPosition = null;
      return;
    }
    dragOverRepo = repoName;

    // Determine above/below based on mouse Y relative to the target element center
    const target = e.currentTarget as HTMLElement | null;
    if (!target) return;
    const rect = target.getBoundingClientRect();
    const midY = rect.top + rect.height / 2;
    dropPosition = e.clientY < midY ? 'above' : 'below';
  }

  function onDragLeave(e: DragEvent): void {
    // Only clear if we actually left the element (not entering a child)
    const related = e.relatedTarget as Node | null;
    const target = e.currentTarget as HTMLElement | null;
    if (target && !target.contains(related)) {
      dragOverRepo = null;
      dropPosition = null;
    }
  }

  function onDrop(e: DragEvent, targetName: string): void {
    e.preventDefault();
    if (!dragRepo || dragRepo === targetName) {
      dragRepo = null;
      dragOverRepo = null;
      dropPosition = null;
      return;
    }

    const names: string[] = orderedRepos.map(r => r.name);
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

  function onDragEnd(): void {
    dragRepo = null;
    dragOverRepo = null;
    dropPosition = null;
  }

  // Group: repo → agents → sub-agents, merged with all scanned repos
  $: repoGroups = buildRepoTree(notifications, allRepos);

  // Apply manual order on top of the default sort
  $: orderedRepos = applyRepoOrder(repoGroups, repoOrder);

  $: flatAgents = orderedRepos.flatMap((r): NotificationEntry[] => r.agents.flatMap((a): NotificationEntry[] => {
    const result: NotificationEntry[] = [a];
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
    const all: number[] = [];
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

  // New Session modal state
  let sessionModalRepo: SessionModalRepo | null = null;

  function openSessionModal(repo: RepoGroup): void {
    sessionModalRepo = { path: repo.path, name: repo.name, branch: repo.branch };
  }

  async function onSessionSpawn(e: CustomEvent<SessionSpawnDetail>): Promise<void> {
    const { command, model, repoPath } = e.detail;
    sessionModalRepo = null;
    spawningRepo = repoPath;
    try {
      const { cols, rows } = estimatePtySize();
      const target = await SpawnAgentWithCommand(repoPath, command, cols, rows);
      const repoName = repoNameFromDir(repoPath);
      addSession(repoPath, makeSession(target, repoPath, repoName, 'agent', model));
      const agent: NotificationEntry = {
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
    } catch (err) {
      console.error('Spawn failed:', err);
    } finally {
      spawningRepo = null;
    }
  }

  function repoNameFromDir(dir: string | undefined): string {
    if (!dir) return 'unknown';
    const parts = dir.split('/');
    return parts[parts.length - 1] || 'unknown';
  }

  let spawningTerminal: string | null = null;

  async function spawnTerminalInRepo(repo: RepoGroup): Promise<void> {
    if (spawningTerminal) return;
    spawningTerminal = repo.path;
    try {
      const { cols, rows } = estimatePtySize();
      const target = await SpawnTerminal(repo.path, cols, rows);
      addSession(repo.path, makeSession(target, repo.path, repo.name, 'terminal', ''));
      const termSession: NotificationEntry = {
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
    } catch (err) {
      console.error('Terminal spawn failed:', err);
    } finally {
      spawningTerminal = null;
    }
  }

  function handleClick(evt: NotificationEntry): void {
    selectedId = evt.agentId;
    dispatch('select', evt);
  }

  function handleKeydown(e: KeyboardEvent): void {
    const allIds = flatAgents.map(a => a.agentId);
    const currentIdx = selectedId == null ? -1 : allIds.indexOf(selectedId);

    if (e.key === 'j' || e.key === 'ArrowDown') {
      e.preventDefault();
      const next = Math.min(currentIdx + 1, allIds.length - 1);
      selectedId = allIds[next] ?? null;
      if (selectedId) scrollIntoView(selectedId);
    } else if (e.key === 'k' || e.key === 'ArrowUp') {
      e.preventDefault();
      const prev = Math.max(currentIdx - 1, 0);
      selectedId = allIds[prev] ?? null;
      if (selectedId) scrollIntoView(selectedId);
    } else if (e.key === 'Enter' && selectedId) {
      e.preventDefault();
      const evt = flatAgents.find(a => a.agentId === selectedId);
      if (evt) dispatch('select', evt);
    }
  }

  function scrollIntoView(id: string): void {
    const el = document.querySelector(`[data-agent-id="${id}"]`);
    if (el) el.scrollIntoView({ block: 'nearest' });
  }

  // Accordion state for sub-agents per parent agent
  let expandedAgents = new Set<string>();

  function isAgentExpanded(agentId: string): boolean {
    return expandedAgents.has(agentId);
  }

  let killingAgents = new Set<string>();

  async function killSession(agent: NotificationEntry, e: Event): Promise<void> {
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
  EventsOn('agent:removed', (agentId: string) => {
    notifications = notifications.filter(n =>
      n.agentId !== agentId && n.parentAgentId !== agentId
    );
  });

  // Repo git status: repoPath -> RepoStatusSnapshot
  let repoStatuses: Record<string, RepoStatusSnapshot> = {};
  let statusInterval: ReturnType<typeof setInterval> | undefined;

  async function refreshRepoStatuses(): Promise<void> {
    for (const repo of orderedRepos) {
      if (!repo.path) continue;
      try {
        const status = await RepoStatus(repo.path);
        repoStatuses[repo.path] = status as RepoStatusSnapshot;
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

  // Push with conflict detection — returns true if conflict modal should open
  let forcePushRepo: ForcePushRepo | null = null;


  function clearActionAfterDelay(path: string, ms = 5000): void {
    setTimeout(() => {
      if (repoActions[path] && !repoActions[path].action) {
        repoActions[path] = { ...IDLE_ACTION };
        repoActions = repoActions;
      }
    }, ms);
  }

  async function smartPush(path: string): Promise<void> {
    repoActions[path] = { action: 'push', result: null, error: null };
    repoActions = repoActions;
    try {
      const result = await GitPush(path);
      if (result.startsWith('conflict:')) {
        forcePushRepo = { path, message: result.slice('conflict:'.length) };
        repoActions[path] = { ...IDLE_ACTION };
      } else {
        repoActions[path] = { action: null, result: 'Pushed', error: null };
        clearActionAfterDelay(path);
      }
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : String(err);
      repoActions[path] = { action: null, result: null, error: message };
      clearActionAfterDelay(path);
    }
    repoActions = repoActions;
    refreshRepoStatuses();
  }

  function onForcePushed(): void {
    if (forcePushRepo) {
      const pushedPath = forcePushRepo.path;
      repoActions[pushedPath] = { action: null, result: 'Force pushed', error: null };
      repoActions = repoActions;
      clearActionAfterDelay(pushedPath);
    }
    forcePushRepo = null;
    refreshRepoStatuses();
  }

  // Collapsed state per repo name
  let collapsedRepos = new Set<string>();

  function isCollapsed(repo: RepoGroup): boolean {
    // Explicitly toggled takes priority
    if (collapsedRepos.has(repo.name)) return true;
    // Auto-collapse if no agents and never explicitly opened
    if (repo.agents.length === 0 && !expandedRepos.has(repo.name)) return true;
    return false;
  }

  // Track repos the user has explicitly expanded (so empty repos stay open after expand)
  let expandedRepos = new Set<string>();

  function handleHeaderClick(repo: RepoGroup): void {
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
  let branchModalRepo: BranchModalRepo | null = null;
  let switchModalRepo: SwitchModalRepo | null = null;
  let mergeModalRepo: MergeModalRepo | null = null;

  function openBranchModal(repo: RepoGroup): void {
    branchModalRepo = { path: repo.path, branch: repo.branch };
  }

  function openSwitchModal(repo: RepoGroup): void {
    switchModalRepo = { path: repo.path, branch: repo.branch, color: getRepoColor(repo.name) };
  }

  function openMergeModal(repo: RepoGroup): void {
    mergeModalRepo = { path: repo.path, branch: repo.branch };
  }

  function onMerged(e: CustomEvent<MergeEventDetail>): void {
    const targetBranch = e.detail?.targetBranch;
    const targetPath = mergeModalRepo?.path;
    if (targetBranch && targetPath) {
      notifications = notifications.map(n =>
        n.repoPath === targetPath ? { ...n, repoBranch: targetBranch } : n,
      );
    }
    mergeModalRepo = null;
    refreshRepoStatuses();
  }

  function onBranchSwitched(e: CustomEvent<BranchEventDetail>): void {
    const newBranch = e.detail?.branch;
    const targetPath = switchModalRepo?.path;
    if (newBranch && targetPath) {
      notifications = notifications.map(n =>
        n.repoPath === targetPath ? { ...n, repoBranch: newBranch } : n,
      );
    }
    switchModalRepo = null;
    refreshRepoStatuses();
  }

  function onBranchCreated(e: CustomEvent<BranchEventDetail>): void {
    const newBranch = e.detail?.branch;
    const targetPath = branchModalRepo?.path;
    if (newBranch && targetPath) {
      // Update branch in all notifications for this repo so it shows immediately
      notifications = notifications.map(n =>
        n.repoPath === targetPath ? { ...n, repoBranch: newBranch } : n,
      );
    }
    branchModalRepo = null;
    refreshRepoStatuses();
  }

  // Repo action states: repoPath -> RepoAction
  let repoActions: Record<string, RepoAction> = {};

  async function runRepoAction(
    path: string,
    actionName: string,
    fn: (path: string) => Promise<string>,
  ): Promise<void> {
    repoActions[path] = { action: actionName, result: null, error: null };
    repoActions = repoActions;
    try {
      const result = await fn(path);
      repoActions[path] = { action: null, result: result || 'Done', error: null };
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : String(err);
      repoActions[path] = { action: null, result: null, error: message };
    }
    repoActions = repoActions;
    refreshRepoStatuses();
    // Clear result/error after 5s
    clearActionAfterDelay(path);
  }

  // Streaming commit output panel: repoPath -> CommitPanel
  let commitPanels: Record<string, CommitPanel> = {};

  /**
   * Single-element list wrapper used by the commit-panel `{#each}` in the
   * template. `{@const}` cannot be the immediate child of a mid-markup
   * position, and TypeScript does not narrow through a call-then-call pattern
   * (`{#if f()}{@const p = f()!}`), so we iterate over a filtered list
   * instead: the compiler narrows `panel` to `CommitPanel` for us.
   */
  function commitPanelList(path: string): CommitPanel[] {
    const p = commitPanels[path];
    return p ? [p] : [];
  }

  function newCommitPanel(): CommitPanel {
    return { lines: [], error: null, explanation: null, done: false, visible: true };
  }

  function startStreamingCommit(path: string): void {
    commitPanels[path] = newCommitPanel();
    commitPanels = commitPanels;
    repoActions[path] = { action: 'commit', result: null, error: null };
    repoActions = repoActions;
    GitCommitStreaming(path);
  }

  function closeCommitPanel(path: string): void {
    delete commitPanels[path];
    commitPanels = commitPanels;
  }

  // Listen for streaming commit progress events
  EventsOn('git:commit:progress', (evt: CommitProgressEvent) => {
    const path = evt.repoPath;
    if (!commitPanels[path]) {
      commitPanels[path] = newCommitPanel();
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
        <RepoHeader
          {repo}
          collapsed={isCollapsed(repo)}
          repoColor={getRepoColor(repo.name)}
          pickerOpen={colorPickerRepo === repo.name}
          on:dragstart={(e) => onDragStart(e, repo.name)}
          on:dragend={onDragEnd}
          on:toggle={() => handleHeaderClick(repo)}
          on:togglecolorpicker={() => toggleColorPicker(repo.name)}
          on:setcolor={(e) => setRepoColor(repo.name, e.detail)}
        />

        {#if !isCollapsed(repo)}
        <div class="repo-body" transition:slide={slideProps}>
          <AgentList
            agents={repo.agents}
            {selectedId}
            {killingAgents}
            bind:expandedAgents
            on:select={(e) => handleClick(e.detail)}
            on:kill={(e) => killSession(e.detail.agent, e.detail.event)}
          />

          <RepoActions
            {repo}
            repoColor={getRepoColor(repo.name)}
            {repoStatuses}
            {repoActions}
            {commitPanels}
            on:switch={() => openSwitchModal(repo)}
            on:branch={() => openBranchModal(repo)}
            on:commit={() => startStreamingCommit(repo.path)}
            on:pull={() => runRepoAction(repo.path, 'pull', GitPull)}
            on:push={() => smartPush(repo.path)}
            on:merge={() => openMergeModal(repo)}
            on:pr={() => runRepoAction(repo.path, 'pr', GitCommitPushAndPR)}
            on:review={() => runRepoAction(repo.path, 'review', SpawnPRReview)}
          />

          <!-- Commit output panel — `{#each}` over a filtered 0-or-1-element
               list narrows `panel` to CommitPanel without a type assertion. -->
          {#each commitPanelList(repo.path) as panel}
            <CommitOutputPanel
              {panel}
              repoColor={getRepoColor(repo.name)}
              on:close={() => closeCommitPanel(repo.path)}
            />
          {/each}
        </div>

        {/if}

        {#if repo.path}
          <div class="spawn-row">
            <button
              class="new-session-btn"
              style="color: {getRepoColor(repo.name) !== REPO_BORDER_NONE ? getRepoColor(repo.name) : ''}"
              on:click|stopPropagation={() => openSessionModal(repo)}
              disabled={spawningRepo === repo.path}
            >
              <span class="new-session-icon"><Plus size={14} /></span>
              {spawningRepo === repo.path ? 'Spawning...' : 'New Session'}
            </button>
            <button
              class="new-session-btn"
              style="color: {getRepoColor(repo.name) !== REPO_BORDER_NONE ? getRepoColor(repo.name) : ''}"
              on:click|stopPropagation={() => spawnTerminalInRepo(repo)}
              disabled={spawningTerminal === repo.path}
            >
              <span class="new-session-icon"><TerminalSquare size={14} /></span>
              {spawningTerminal === repo.path ? 'Opening...' : 'Terminal'}
            </button>
            <button
              class="new-session-btn"
              style="color: {getRepoColor(repo.name) !== REPO_BORDER_NONE ? getRepoColor(repo.name) : ''}"
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
      <span class="status-sparkline" aria-hidden="true" style="--ambient-color: {STATUS_COLORS.open}">
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
    /* Colour fed by inline --ambient-color from STATUS_COLORS.open so the
       status→token mapping lives in one place (the script table). Falls
       back to accent-teal to match the long-standing ambient hue. */
    color: var(--ambient-color, var(--accent-teal));
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
</style>
