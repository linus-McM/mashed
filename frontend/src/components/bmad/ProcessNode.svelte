<script>
  import { Handle, Position } from '@xyflow/svelte';
  import { Search, Briefcase, Palette, Building2, Code, FileText, TestTube } from 'lucide-svelte';

  export let data = {};
  export let id = '';
  export let selected = false;

  const phaseColors = {
    analysis: 'var(--accent-blue, #3d9eff)',
    planning: 'var(--accent-green, #00e57a)',
    solutioning: 'var(--accent-purple, #9d6fff)',
    implementation: 'var(--accent-amber, #f0a500)',
    support: 'var(--text-dim, #4a5a6a)',
  };

  const roleIcons = {
    analyst: Search,
    pm: Briefcase,
    'ux-designer': Palette,
    architect: Building2,
    developer: Code,
    'tech-writer': FileText,
    qa: TestTube,
  };

  $: process = data.process || {};
  $: phase = process.phase || 'support';
  $: phaseColor = phaseColors[phase] || phaseColors.support;
  $: agentRole = process.agentRole || 'developer';
  $: RoleIcon = roleIcons[agentRole] || Code;
  $: inputs = process.inputs || [];
  $: outputs = process.outputs || [];
  $: status = data.status || 'pending';
  $: label = data.label || process.name || 'Process';
  $: storyId = data.storyId || '';
  $: storyStatus = data.storyStatus || '';
  $: artifactStatus = data.artifactStatus || null;
  $: hasArtifacts = artifactStatus && (artifactStatus.found?.length > 0 || artifactStatus.missing?.length > 0);

  import { storyStatusColors } from '../../lib/sprintColors.js';
</script>

<div class="process-node" class:selected class:running={status === 'running'}>
  <div class="phase-bar" style="background: {phaseColor}" />

  <div class="node-body">
    <div class="node-header">
      <span class="role-icon" style="color: {phaseColor}">
        <svelte:component this={RoleIcon} size={13} />
      </span>
      <span class="node-label">{label}</span>
    </div>

    {#if inputs.length > 0}
      <div class="artifacts">
        <span class="artifact-label">in:</span>
        <span class="artifact-list">{inputs.join(', ')}</span>
      </div>
    {/if}

    {#if outputs.length > 0}
      <div class="artifacts">
        <span class="artifact-label">out:</span>
        <span class="artifact-list">{outputs.join(', ')}</span>
      </div>
    {/if}

    <div class="status-row">
      {#if status === 'pending'}
        <span class="status-dot pending" />
        <span class="status-text">pending</span>
      {:else if status === 'running'}
        <span class="status-dot running-dot" />
        <span class="status-text running-text">running</span>
      {:else if status === 'complete'}
        <span class="status-check">&#10003;</span>
        <span class="status-text complete-text">complete</span>
      {:else if status === 'failed'}
        <span class="status-x">&#10005;</span>
        <span class="status-text failed-text">failed</span>
      {:else if status === 'skipped'}
        <span class="status-dash">&mdash;</span>
        <span class="status-text">skipped</span>
      {/if}
    </div>

    {#if status === 'complete' && hasArtifacts}
      <div class="artifact-indicators">
        {#each artifactStatus.found as name}
          <span class="artifact-icon found" title="{name} found">&#10003;</span>
        {/each}
        {#each artifactStatus.missing as name}
          <span class="artifact-icon missing" title="{name} missing">!</span>
        {/each}
      </div>
    {/if}

    {#if storyId}
      <div class="story-badge">
        <span class="story-badge-dot" style="background: {storyStatusColors[storyStatus] || storyStatusColors.backlog}" />
        <span class="story-badge-id">{storyId}</span>
      </div>
    {/if}
  </div>

  <Handle type="target" position={Position.Left} />
  <Handle type="source" position={Position.Right} />
</div>

<style>
  .process-node {
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md, 6px);
    min-width: 160px;
    max-width: 220px;
    overflow: hidden;
    font-family: var(--font-mono);
    transition: border-color 150ms ease, box-shadow 150ms ease;
  }

  .process-node.selected {
    border-color: var(--accent-green);
    box-shadow: 0 0 0 2px rgba(0, 229, 122, 0.35);
  }

  .process-node.running {
    border-color: var(--accent-green, #00e57a);
    animation: node-pulse 2s ease-in-out infinite;
  }

  @keyframes node-pulse {
    0%, 100% { box-shadow: 0 0 0 0 rgba(0, 229, 122, 0); }
    50% { box-shadow: 0 0 0 3px rgba(0, 229, 122, 0.25); }
  }

  .phase-bar {
    height: 4px;
    width: 100%;
  }

  .node-body {
    padding: 8px 10px 6px;
  }

  .node-header {
    display: flex;
    align-items: center;
    gap: 6px;
    margin-bottom: 4px;
  }

  .role-icon {
    display: flex;
    align-items: center;
    flex-shrink: 0;
  }

  .node-label {
    font-size: 11px;
    font-weight: 600;
    color: var(--text-primary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .artifacts {
    display: flex;
    gap: 4px;
    font-size: 9px;
    color: var(--text-muted);
    line-height: 1.3;
    margin-bottom: 2px;
  }

  .artifact-label {
    color: var(--text-dim);
    flex-shrink: 0;
    font-weight: 600;
  }

  .artifact-list {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .status-row {
    display: flex;
    align-items: center;
    gap: 5px;
    margin-top: 4px;
    padding-top: 4px;
    border-top: 1px solid var(--border-subtle);
  }

  .status-text {
    font-size: 9px;
    color: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: 0.3px;
  }

  .status-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    flex-shrink: 0;
  }

  .status-dot.pending {
    background: var(--text-muted, #8b949e);
  }

  .status-dot.running-dot {
    background: var(--accent-green, #00e57a);
    animation: dot-pulse 1.5s ease-in-out infinite;
  }

  @keyframes dot-pulse {
    0%, 100% { opacity: 0.4; }
    50% { opacity: 1; }
  }

  .running-text { color: var(--accent-green, #00e57a); }

  .status-check {
    font-size: 10px;
    color: var(--accent-green, #00e57a);
    line-height: 1;
  }

  .complete-text { color: var(--accent-green, #00e57a); }

  .status-x {
    font-size: 10px;
    color: var(--accent-red, #f85149);
    line-height: 1;
  }

  .failed-text { color: var(--accent-red, #f85149); }

  .status-dash {
    font-size: 10px;
    color: var(--text-muted);
    line-height: 1;
  }

  .artifact-indicators {
    display: flex;
    gap: 3px;
    margin-top: 3px;
    padding-top: 3px;
    border-top: 1px solid var(--border-subtle);
    flex-wrap: wrap;
  }

  .artifact-icon {
    width: 14px;
    height: 14px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 8px;
    font-weight: 700;
    line-height: 1;
    cursor: default;
  }

  .artifact-icon.found {
    background: rgba(0, 229, 122, 0.15);
    color: var(--accent-green, #00e57a);
  }

  .artifact-icon.missing {
    background: rgba(240, 165, 0, 0.15);
    color: var(--accent-amber, #f0a500);
  }

  .story-badge {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 3px 0 0;
    border-top: 1px solid var(--border-subtle);
    margin-top: 2px;
  }

  .story-badge-dot {
    width: 5px;
    height: 5px;
    border-radius: 50%;
    flex-shrink: 0;
  }

  .story-badge-id {
    font-size: 8px;
    color: var(--text-muted);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
</style>
