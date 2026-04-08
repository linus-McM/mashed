<script>
  import { createEventDispatcher } from 'svelte';
  import { ChevronDown, ChevronRight, Trash2 } from 'lucide-svelte';
  import SprintPanel from './SprintPanel.svelte';

  export let processes = [];
  export let templates = [];
  export let savedWorkflows = [];
  export let sprintStatus = null;

  const dispatch = createEventDispatcher();

  let activeTab = 'processes';

  const phaseOrder = ['analysis', 'planning', 'solutioning', 'implementation', 'support'];
  const phaseLabels = {
    analysis: 'Analysis',
    planning: 'Planning',
    solutioning: 'Solutioning',
    implementation: 'Implementation',
    support: 'Support',
  };
  const phaseColors = {
    analysis: 'var(--accent-blue, #58a6ff)',
    planning: 'var(--accent-green, #3fb950)',
    solutioning: 'var(--accent-purple, #bc8cff)',
    implementation: 'var(--accent-amber, #d29922)',
    support: 'var(--text-muted, #8b949e)',
  };

  let expandedPhases = { analysis: true, planning: true, solutioning: true, implementation: true, support: true };

  function togglePhase(phase) {
    expandedPhases[phase] = !expandedPhases[phase];
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
</script>

<div class="sidebar">
  <div class="tabs">
    <button class="tab" class:active={activeTab === 'processes'} on:click={() => activeTab = 'processes'}>Processes</button>
    <button class="tab" class:active={activeTab === 'templates'} on:click={() => activeTab = 'templates'}>Templates</button>
    <button class="tab" class:active={activeTab === 'saved'} on:click={() => activeTab = 'saved'}>Saved</button>
    <button class="tab" class:active={activeTab === 'sprint'} on:click={() => activeTab = 'sprint'}>Sprint</button>
  </div>

  <div class="tab-content">
    {#if activeTab === 'processes'}
      <div class="process-list">
        {#each phaseOrder as phase}
          {#if grouped[phase] && grouped[phase].length > 0}
            <div class="phase-group">
              <button class="phase-header" on:click={() => togglePhase(phase)}>
                <span class="phase-indicator" style="background: {phaseColors[phase]}" />
                {#if expandedPhases[phase]}
                  <ChevronDown size={12} />
                {:else}
                  <ChevronRight size={12} />
                {/if}
                <span class="phase-label">{phaseLabels[phase]}</span>
                <span class="phase-count">{grouped[phase].length}</span>
              </button>
              {#if expandedPhases[phase]}
                <div class="phase-items">
                  {#each grouped[phase] as process}
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
        {#if processes.length === 0}
          <div class="empty-state">No processes registered</div>
        {/if}
      </div>

    {:else if activeTab === 'templates'}
      <div class="template-list">
        {#each templates as tmpl}
          <div class="template-card">
            <div class="template-info">
              <span class="template-name">{tmpl.name}</span>
              <span class="template-meta">{tmpl.nodes?.length || 0} nodes</span>
            </div>
            <button class="use-btn" on:click={() => dispatch('use-template', tmpl.id)}>Use</button>
          </div>
        {/each}
        {#if templates.length === 0}
          <div class="empty-state">No templates available</div>
        {/if}
      </div>

    {:else if activeTab === 'saved'}
      <div class="saved-list">
        {#each savedWorkflows as wf}
          <div class="saved-item">
            <button class="saved-name" on:click={() => dispatch('load-workflow', wf.id)}>
              {wf.name}
            </button>
            <button class="delete-btn" on:click={() => dispatch('delete-workflow', wf.id)} title="Delete">
              <Trash2 size={12} />
            </button>
          </div>
        {/each}
        {#if savedWorkflows.length === 0}
          <div class="empty-state">No saved workflows</div>
        {/if}
      </div>

    {:else if activeTab === 'sprint'}
      <SprintPanel {sprintStatus} />
    {/if}
  </div>
</div>

<style>
  .sidebar {
    width: 240px;
    min-width: 240px;
    display: flex;
    flex-direction: column;
    background: var(--bg-surface);
    border-right: 1px solid var(--border-subtle);
    overflow: hidden;
  }

  .tabs {
    display: flex;
    border-bottom: 1px solid var(--border-subtle);
    flex-shrink: 0;
  }

  .tab {
    flex: 1;
    padding: 8px 0;
    background: none;
    border: none;
    border-bottom: 2px solid transparent;
    color: var(--text-muted);
    font-family: var(--font-mono);
    font-size: 11px;
    cursor: pointer;
    transition: color 100ms ease, border-color 100ms ease;
  }

  .tab:hover { color: var(--text-dim); }
  .tab.active {
    color: var(--text-primary);
    border-bottom-color: var(--accent-green);
  }

  .tab-content {
    flex: 1;
    overflow-y: auto;
    padding: 4px 0;
  }

  /* Process list */
  .phase-group {
    margin-bottom: 2px;
  }

  .phase-header {
    display: flex;
    align-items: center;
    gap: 6px;
    width: 100%;
    padding: 6px 10px;
    background: none;
    border: none;
    color: var(--text-dim);
    font-family: var(--font-mono);
    font-size: 11px;
    font-weight: 600;
    cursor: pointer;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .phase-header:hover { background: var(--bg-elevated); }

  .phase-indicator {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    flex-shrink: 0;
  }

  .phase-label { flex: 1; text-align: left; }

  .phase-count {
    font-size: 10px;
    color: var(--text-muted);
    font-weight: 400;
  }

  .phase-items {
    padding: 0 0 4px;
  }

  .process-item {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 5px 10px 5px 24px;
    cursor: grab;
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--text-primary);
    transition: background 80ms ease;
    user-select: none;
  }

  .process-item:hover { background: var(--bg-elevated); }
  .process-item:active { cursor: grabbing; }

  .process-dot {
    width: 5px;
    height: 5px;
    border-radius: 50%;
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
  .template-card {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 8px 10px;
    border-bottom: 1px solid var(--border-subtle);
  }

  .template-info {
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-width: 0;
  }

  .template-name {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--text-primary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .template-meta {
    font-family: var(--font-mono);
    font-size: 10px;
    color: var(--text-muted);
  }

  .use-btn {
    padding: 3px 8px;
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-sm);
    color: var(--accent-green);
    font-family: var(--font-mono);
    font-size: 10px;
    cursor: pointer;
    flex-shrink: 0;
    transition: background 100ms ease;
  }

  .use-btn:hover {
    background: var(--bg-active);
    border-color: var(--accent-green);
  }

  /* Saved workflows */
  .saved-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 6px 10px;
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
