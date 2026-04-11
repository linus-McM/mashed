<script>
  import { createEventDispatcher, onDestroy, onMount } from 'svelte';
  import { ChevronDown, ChevronRight, Trash2, GitBranch, Repeat, Target, Filter, GitMerge } from 'lucide-svelte';
  import SprintPanel from './SprintPanel.svelte';
  import GitPanel from './GitPanel.svelte';
  import { GetConfig, SetSidebarWidth } from '../../../wailsjs/go/main/App.js';

  export let processes = [];
  export let templates = [];
  export let savedWorkflows = [];
  export let sprintStatus = null;
  export let repoPath = '';
  export let repoBranch = '';
  /**
   * Grouped mashed-ready skills and commands. Shape matches Go's
   * bmad.GroupedMashedAssets — four arrays of MashedAssetInfo records,
   * one per {scope} × {kind}. Empty / absent is rendered as "No mashed
   * assets found — run the mashed-refactor-asset skill against your
   * .claude/skills/ or .claude/commands/ directory to populate this
   * tab".
   * @type {{ localCommands: any[], globalCommands: any[], localSkills: any[], globalSkills: any[] }}
   */
  export let groupedMashedAssets = { localCommands: [], globalCommands: [], localSkills: [], globalSkills: [] };

  const dispatch = createEventDispatcher();

  let activeTab = 'templates';

  // ── Skills/Commands tab — collapsible group state ──
  //
  // Mirrors the openPhase / controlFlowOpen pattern used by the
  // Processes tab so the visual grammar stays consistent: one group
  // expanded at a time, chevron + count badge in the header, draggable
  // rows inside. Local commands default to open because they are the
  // most likely target for a user who just refactored their repo.
  const mashedGroupKeys = ['localCommands', 'globalCommands', 'localSkills', 'globalSkills'];
  const mashedGroupLabels = {
    localCommands: 'Local Commands',
    globalCommands: 'Global Commands',
    localSkills: 'Local Skills',
    globalSkills: 'Global Skills',
  };
  const mashedGroupAccents = {
    localCommands: 'var(--accent-green, #3fb950)',
    globalCommands: 'var(--accent-blue, #58a6ff)',
    localSkills: 'var(--accent-purple, #bc8cff)',
    globalSkills: 'var(--accent-amber, #d29922)',
  };
  let openMashedGroup = 'localCommands';
  function toggleMashedGroup(key) {
    openMashedGroup = openMashedGroup === key ? null : key;
  }

  /**
   * Drag handler for a mashed asset. Only commands (bmadRole: command)
   * are actually draggable onto the canvas — the executor cannot chain
   * skills. For skill entries we still fire dragstart (to give a neutral
   * drag image) but set a role that WorkflowBuilder's drop handler
   * rejects, so the user gets a visual "nope" signal without a crash.
   */
  function onMashedAssetDragStart(e, asset) {
    e.dataTransfer.setData('application/mashed-asset', JSON.stringify({
      name: asset.name,
      path: asset.path,
      kind: asset.kind,
      source: asset.source,
      role: asset.role,
      description: asset.description || '',
    }));
    e.dataTransfer.effectAllowed = asset.role === 'command' ? 'move' : 'none';
  }

  /** Count across all four groups so the tab can show a badge when empty. */
  $: totalMashedAssets =
    (groupedMashedAssets?.localCommands?.length || 0) +
    (groupedMashedAssets?.globalCommands?.length || 0) +
    (groupedMashedAssets?.localSkills?.length || 0) +
    (groupedMashedAssets?.globalSkills?.length || 0);

  const phaseOrder = ['analysis', 'planning', 'solutioning', 'implementation', 'support', 'utilities'];
  const phaseLabels = {
    analysis: 'Analysis',
    planning: 'Planning',
    solutioning: 'Solutioning',
    implementation: 'Implementation',
    support: 'Support',
    utilities: 'Utilities',
  };
  const phaseColors = {
    analysis: 'var(--accent-blue, #58a6ff)',
    planning: 'var(--accent-green, #3fb950)',
    solutioning: 'var(--accent-purple, #bc8cff)',
    implementation: 'var(--accent-amber, #d29922)',
    support: 'var(--text-muted, #8b949e)',
    utilities: 'var(--accent-teal, #00c4b3)',
  };

  let openPhase = 'analysis';

  const controlFlowNodes = [
    { type: 'condition', name: 'Condition', description: 'If/else branch based on output', Icon: GitBranch },
    { type: 'loop', name: 'Loop', description: 'Repeat N times', Icon: Repeat },
    { type: 'loopUntil', name: 'Loop Until', description: 'Repeat until condition met', Icon: Target },
    { type: 'transform', name: 'Transform', description: 'Extract/transform data', Icon: Filter },
    { type: 'merge', name: 'Merge', description: 'Join branches', Icon: GitMerge },
  ];

  let controlFlowOpen = true;

  function onControlFlowDragStart(e, item) {
    e.dataTransfer.setData('application/bmad-controlflow', item.type);
    e.dataTransfer.effectAllowed = 'move';
  }

  function togglePhase(phase) {
    openPhase = openPhase === phase ? null : phase;
  }

  function groupByPhase(procs) {
    const groups = {};
    for (const phase of phaseOrder) {
      groups[phase] = procs.filter(p => p.phase === phase);
    }
    return groups;
  }

  $: grouped = groupByPhase(processes);

  function onDragStart(e, process) {
    e.dataTransfer.setData('application/bmad-process', process.id);
    e.dataTransfer.effectAllowed = 'move';
  }

  // Resize logic
  let sidebarWidth = 280;
  let resizing = false;

  onMount(async () => {
    try {
      const cfg = await GetConfig();
      if (cfg.sidebarWidth && cfg.sidebarWidth >= 200 && cfg.sidebarWidth <= 500) {
        sidebarWidth = cfg.sidebarWidth;
      }
    } catch {}
  });

  function onResizeStart(e) {
    e.preventDefault();
    resizing = true;
    const startX = e.clientX;
    const startWidth = sidebarWidth;

    function onMouseMove(e) {
      sidebarWidth = Math.max(200, Math.min(500, startWidth + (e.clientX - startX)));
    }

    function onMouseUp() {
      resizing = false;
      window.removeEventListener('mousemove', onMouseMove);
      window.removeEventListener('mouseup', onMouseUp);
      SetSidebarWidth(sidebarWidth).catch(() => {});
    }

    window.addEventListener('mousemove', onMouseMove);
    window.addEventListener('mouseup', onMouseUp);
  }
</script>

<div class="sidebar" style="width: {sidebarWidth}px; min-width: {sidebarWidth}px;">
  <div class="tabs">
    <button class="tab" class:active={activeTab === 'templates'} on:click={() => activeTab = 'templates'}>Templates</button>
    <button class="tab" class:active={activeTab === 'sprint'} on:click={() => activeTab = 'sprint'}>Sprint</button>
    <button class="tab" class:active={activeTab === 'processes'} on:click={() => activeTab = 'processes'}>Processes</button>
    <button class="tab" class:active={activeTab === 'skills'} on:click={() => activeTab = 'skills'}>Skills</button>
    <button class="tab" class:active={activeTab === 'saved'} on:click={() => activeTab = 'saved'}>Saved</button>
    <button class="tab" class:active={activeTab === 'git'} on:click={() => activeTab = 'git'}>Git</button>
  </div>

  <div class="tab-content">
    {#if activeTab === 'processes'}
      <div class="process-list">
        {#each phaseOrder as phase}
          {#if grouped[phase] && grouped[phase].length > 0}
            <div class="phase-group">
              <button class="phase-header" on:click={() => togglePhase(phase)}>
                <span class="phase-indicator" style="background: {phaseColors[phase]}" />
                {#if openPhase === phase}
                  <ChevronDown size={12} />
                {:else}
                  <ChevronRight size={12} />
                {/if}
                <span class="phase-label">{phaseLabels[phase]}</span>
                <span class="phase-count-badge">{grouped[phase].length}</span>
              </button>
              {#if openPhase === phase}
                <div class="phase-items">
                  {#each grouped[phase] as process}
                    <!-- svelte-ignore a11y-no-static-element-interactions -->
                    <div
                      class="process-item"
                      draggable="true"
                      on:dragstart={(e) => onDragStart(e, process)}
                      title={process.description}
                    >
                      <span class="process-dot" style="background: {phaseColors[phase]}" />
                      <span class="process-name">{process.name}</span>
                      {#if process.moduleId}
                        <span class="module-badge">{process.moduleId}</span>
                      {/if}
                    </div>
                  {/each}
                </div>
              {/if}
            </div>
          {/if}
        {/each}
        <!-- Control Flow Group -->
        <div class="phase-group control-flow-group">
          <button class="phase-header" on:click={() => controlFlowOpen = !controlFlowOpen}>
            <span class="phase-indicator" style="background: var(--accent-green)" />
            {#if controlFlowOpen}
              <ChevronDown size={12} />
            {:else}
              <ChevronRight size={12} />
            {/if}
            <span class="phase-label">Control Flow</span>
            <span class="phase-count-badge">{controlFlowNodes.length}</span>
          </button>
          {#if controlFlowOpen}
            <div class="phase-items">
              {#each controlFlowNodes as item}
                <!-- svelte-ignore a11y-no-static-element-interactions -->
                <div
                  class="process-item"
                  draggable="true"
                  on:dragstart={(e) => onControlFlowDragStart(e, item)}
                  title={item.description}
                >
                  <span class="cf-icon" style="color: var(--accent-green)">
                    <svelte:component this={item.Icon} size={12} />
                  </span>
                  <span class="process-name">{item.name}</span>
                </div>
              {/each}
            </div>
          {/if}
        </div>

        {#if processes.length === 0}
          <div class="empty-state">No processes registered</div>
        {/if}
      </div>

    {:else if activeTab === 'templates'}
      <div class="template-list">
        {#each templates as tmpl}
          <!-- svelte-ignore a11y-no-static-element-interactions -->
          <div
            class="template-card"
            draggable="true"
            on:dragstart={(e) => {
              e.dataTransfer.setData('application/bmad-template', tmpl.id);
              e.dataTransfer.effectAllowed = 'move';
            }}
            title="Drag onto canvas to add this template"
          >
            <span class="template-name">{tmpl.name}</span>
            <span class="template-meta">{tmpl.nodes?.length || 0} nodes · {tmpl.edges?.length || 0} edges</span>
            {#if tmpl.description}
              <span class="template-desc">{tmpl.description}</span>
            {/if}
          </div>
        {/each}
        {#if templates.length === 0}
          <div class="empty-state">No templates available</div>
        {/if}

        <button class="template-card custom-template" on:click={() => dispatch('create-custom-template')}>
          <span class="template-name">+ Custom Template</span>
          <span class="template-meta">Build your own workflow</span>
          <span class="template-desc">Design a bespoke pipeline by dragging processes onto a blank canvas. Arrange nodes, connect edges, and save as a reusable template for your team.</span>
        </button>
      </div>

    {:else if activeTab === 'saved'}
      <div class="saved-list">
        {#each savedWorkflows || [] as wf}
          <div class="saved-item">
            <button class="saved-name" on:click={() => dispatch('load-workflow', wf.id)}>
              {wf.name}
            </button>
            <button class="delete-btn" on:click={() => dispatch('delete-workflow', wf.id)} title="Delete">
              <Trash2 size={12} />
            </button>
          </div>
        {/each}
        {#if !savedWorkflows || savedWorkflows.length === 0}
          <div class="empty-state">No saved workflows</div>
        {/if}
      </div>

    {:else if activeTab === 'skills'}
      <div class="process-list">
        {#each mashedGroupKeys as key}
          {@const items = groupedMashedAssets?.[key] || []}
          {#if items.length > 0}
            <div class="phase-group">
              <button class="phase-header" on:click={() => toggleMashedGroup(key)}>
                <span class="phase-indicator" style="background: {mashedGroupAccents[key]}" />
                {#if openMashedGroup === key}
                  <ChevronDown size={12} />
                {:else}
                  <ChevronRight size={12} />
                {/if}
                <span class="phase-label">{mashedGroupLabels[key]}</span>
                <span class="phase-count-badge">{items.length}</span>
              </button>
              {#if openMashedGroup === key}
                <div class="phase-items">
                  {#each items as asset (asset.path)}
                    <!-- svelte-ignore a11y-no-static-element-interactions -->
                    <div
                      class="process-item"
                      class:disabled={asset.role !== 'command'}
                      draggable={asset.role === 'command'}
                      on:dragstart={(e) => onMashedAssetDragStart(e, asset)}
                      title={asset.description || asset.name}
                    >
                      <span class="process-dot" style="background: {mashedGroupAccents[key]}" />
                      <span class="process-name">{asset.name}</span>
                      {#if asset.role === 'skill'}
                        <span class="module-badge">pinned</span>
                      {/if}
                    </div>
                  {/each}
                </div>
              {/if}
            </div>
          {/if}
        {/each}

        {#if totalMashedAssets === 0}
          <div class="empty-state">
            No mashed-ready assets found.
            <br />
            Run the <code>mashed-refactor-asset</code> skill against your
            <code>.claude/skills/</code> or <code>.claude/commands/</code>
            directory to populate this tab.
          </div>
        {/if}
      </div>

    {:else if activeTab === 'sprint'}
      <SprintPanel {sprintStatus} />

    {:else if activeTab === 'git'}
      <GitPanel {repoPath} {repoBranch} on:branch-changed />
    {/if}
  </div>
  <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
  <div class="resize-handle" class:active={resizing} on:mousedown={onResizeStart} role="separator" />
</div>

<style>
  .sidebar {
    position: relative;
    display: flex;
    flex-direction: column;
    background: var(--bg-surface);
    border-right: 1px solid var(--border-subtle);
    overflow: hidden;
    flex-shrink: 0;
  }

  .resize-handle {
    position: absolute;
    top: 0;
    right: -3px;
    width: 6px;
    height: 100%;
    cursor: col-resize;
    z-index: 10;
    transition: background 150ms ease;
  }

  .resize-handle:hover,
  .resize-handle.active {
    background: var(--accent-green);
    opacity: 0.5;
  }

  .tabs {
    display: flex;
    border-bottom: 1px solid var(--border-subtle);
    flex-shrink: 0;
  }

  .tab {
    flex: 1;
    padding: 10px 0;
    background: none;
    border: none;
    border-bottom: 2px solid transparent;
    color: var(--text-primary);
    font-family: var(--font-mono);
    font-size: var(--text-body);
    font-weight: 500;
    cursor: pointer;
    transition: color 100ms ease, border-color 100ms ease, background 100ms ease;
  }

  .tab:hover {
    color: var(--text-primary);
    background: var(--bg-elevated);
  }
  .tab.active {
    color: var(--accent-green);
    font-weight: 600;
    border-bottom-color: var(--accent-green);
  }

  .tab-content {
    flex: 1;
    overflow-y: auto;
    padding: var(--sp-xs) var(--sp-sm);
  }

  /* Process list — accordion panels */
  .phase-group {
    margin-bottom: 4px;
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: 6px;
    overflow: hidden;
  }

  .phase-header {
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

  .phase-header:hover { background: var(--bg-active); }

  .phase-indicator {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    flex-shrink: 0;
  }

  .phase-label { flex: 1; text-align: left; }

  .phase-count-badge {
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

  .phase-items {
    padding: var(--sp-2xs) 0 var(--sp-xs);
    border-top: 1px solid var(--border-subtle);
    background: var(--bg-surface);
  }

  .process-item {
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

  .process-item:hover { background: var(--bg-elevated); }
  .process-item:active { cursor: grabbing; }

  /* Skill-role rows in the Skills tab are listed but not draggable
     onto the canvas (the executor cannot chain them). Visually:
     muted text, default cursor, no hover lift. */
  .process-item.disabled {
    cursor: default;
    color: var(--text-muted);
  }
  .process-item.disabled:hover { background: transparent; }
  .process-item.disabled:active { cursor: default; }

  .process-dot {
    width: 5px;
    height: 5px;
    border-radius: 50%;
    flex-shrink: 0;
  }

  .cf-icon {
    display: flex;
    align-items: center;
    flex-shrink: 0;
  }

  .process-name {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    flex: 1;
  }

  .module-badge {
    font-size: 8px;
    padding: 1px 4px;
    border-radius: 3px;
    background: var(--bg-deepest);
    color: var(--text-muted);
    border: 1px solid var(--border-subtle);
    flex-shrink: 0;
    text-transform: uppercase;
    letter-spacing: 0.3px;
  }

  /* Templates */
  .template-list {
    display: flex;
    flex-direction: column;
    gap: 4px;
    padding: var(--sp-xs) var(--sp-sm);
  }

  .template-card {
    display: flex;
    flex-direction: column;
    gap: 4px;
    padding: 10px 12px;
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: 6px;
    cursor: grab;
    user-select: none;
    transition: background 80ms ease, border-color 80ms ease;
  }

  .template-card:hover {
    background: var(--bg-active);
    border-color: var(--border-emphasis);
  }

  .template-card:active { cursor: grabbing; }

  .template-name {
    font-family: var(--font-mono);
    font-size: 13px;
    font-weight: 600;
    color: var(--text-primary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .template-meta {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--text-dim);
  }

  .template-desc {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--text-muted);
    line-height: 1.4;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .custom-template {
    margin-top: auto;
    border-style: dashed;
    border-color: var(--border-emphasis);
    background: var(--bg-surface);
    cursor: pointer;
    text-align: left;
  }

  .custom-template:hover {
    border-color: var(--accent-green);
    background: color-mix(in srgb, var(--accent-green) 5%, transparent);
  }

  .custom-template .template-name {
    color: var(--accent-green);
  }

  /* Saved workflows */
  .saved-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: var(--sp-xs) 10px;
    border-bottom: 1px solid var(--border-subtle);
  }

  .saved-name {
    background: none;
    border: none;
    color: var(--text-primary);
    font-family: var(--font-mono);
    font-size: 11px;
    cursor: pointer;
    padding: 0;
    text-align: left;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    flex: 1;
  }

  .saved-name:hover { color: var(--accent-green); }

  .delete-btn {
    background: none;
    border: none;
    color: var(--text-muted);
    cursor: pointer;
    padding: 2px;
    border-radius: var(--radius-sm);
    display: flex;
    align-items: center;
    flex-shrink: 0;
    transition: color 100ms ease;
  }

  .delete-btn:hover { color: var(--accent-red, #f85149); }

  .empty-state {
    padding: 20px 10px;
    text-align: center;
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--text-muted);
  }
</style>
