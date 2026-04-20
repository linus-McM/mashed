<script>
  import { Handle, Position } from '@xyflow/svelte';
  import { Search, Briefcase, Palette, Building2, Code, FileText, TestTube, MessageCircleQuestion } from 'lucide-svelte';
  import { getNodePath, formatBreadcrumb } from '../../lib/bmad/nodePath';

  export let data = {};
  // svelte-ignore unused-export-let
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
  $: nodeRound = data.nodeRound || 0;
  $: maxRounds = process.gate?.maxRounds || data.maxRounds || 0;
  $: gateFlash = !!data.gateFlash;
  $: roundText = (() => {
    if (!nodeRound || nodeRound <= 0) return '';
    if (maxRounds > 0) return `${nodeRound} / ${maxRounds}`;
    return `${nodeRound} rounds`;
  })();

  import { storyStatusColors } from '../../lib/sprintColors.js';
</script>

<div
  class="process-node"
  class:selected
  class:running={status === 'running'}
  class:awaiting={status === 'awaiting_input'}
  class:gate-flash={gateFlash}
>
  <div class="phase-bar" style="background: {phaseColor}" />

  <div class="node-body">
    <div class="node-header">
      <span class="role-icon" style="color: {phaseColor}">
        <svelte:component this={RoleIcon} size={13} />
      </span>
      <span class="node-label">{label}</span>
    </div>

    {#each inputs as name}
      <div class="artifacts">
        <span class="artifact-label">in:</span>
        <span class="artifact-list">{name}</span>
      </div>
      {@const inPath = getNodePath({ data }, 'in', name)}
      <div
        class="breadcrumb-row"
        class:unresolved={!inPath}
        title={inPath || 'unresolved'}
      >
        {formatBreadcrumb(inPath)}
      </div>
    {/each}

    {#each outputs as name, i}
      <div class="artifacts" class:out-first={i === 0 && inputs.length > 0}>
        <span class="artifact-label">out:</span>
        <span class="artifact-list">{name}</span>
      </div>
      {@const outPath = getNodePath({ data }, 'out', name)}
      <div
        class="breadcrumb-row"
        class:unresolved={!outPath}
        title={outPath || 'unresolved'}
      >
        {formatBreadcrumb(outPath)}
      </div>
    {/each}

    {#if roundText}
      <div class="round-counter" data-testid="round-counter">{roundText}</div>
    {/if}

    <div class="status-row">
      {#if status === 'pending'}
        <span class="status-dot pending" />
        <span class="status-text">pending</span>
      {:else if status === 'running'}
        <span class="status-dot running-dot" />
        <span class="status-text running-text">running</span>
      {:else if status === 'awaiting_input'}
        <span class="awaiting-badge" aria-label="Awaiting input" data-testid="awaiting-badge">
          <MessageCircleQuestion size={11} />
        </span>
        <span class="status-text awaiting-text">awaiting</span>
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
    box-shadow: 0 0 0 2px color-mix(in srgb, var(--accent-green) 35%, transparent);
  }

  .process-node.running {
    border-color: var(--accent-green, #00e57a);
    animation: node-pulse 2s ease-in-out infinite;
  }

  @keyframes node-pulse {
    0%, 100% { box-shadow: 0 0 0 0 color-mix(in srgb, var(--accent-green) 0%, transparent); }
    50% { box-shadow: 0 0 0 3px color-mix(in srgb, var(--accent-green) 25%, transparent); }
  }

  .phase-bar {
    height: 4px;
    width: 100%;
  }

  .node-body {
    padding: 8px 10px var(--sp-xs);
  }

  .node-header {
    display: flex;
    align-items: center;
    gap: var(--sp-sm);
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
    font-size: var(--text-label);
    color: var(--text-primary);
    line-height: 1.3;
    margin-bottom: 2px;
  }

  .artifact-label {
    color: var(--text-secondary);
    flex-shrink: 0;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.3px;
  }

  .artifact-list {
    color: var(--text-primary);
    font-weight: 600;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .status-row {
    display: flex;
    align-items: center;
    gap: var(--sp-xs);
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
    font-size: var(--text-label);
    color: var(--accent-green, #00e57a);
    line-height: 1;
  }

  .complete-text { color: var(--accent-green, #00e57a); }

  .status-x {
    font-size: var(--text-label);
    color: var(--accent-red, #f85149);
    line-height: 1;
  }

  .failed-text { color: var(--accent-red, #f85149); }

  .status-dash {
    font-size: var(--text-label);
    color: var(--text-muted);
    line-height: 1;
  }

  .artifact-indicators {
    display: flex;
    gap: var(--sp-2xs);
    margin-top: var(--sp-2xs);
    padding-top: var(--sp-2xs);
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
    background: color-mix(in srgb, var(--accent-green) 15%, transparent);
    color: var(--accent-green, #00e57a);
  }

  .artifact-icon.missing {
    background: color-mix(in srgb, var(--accent-amber) 15%, transparent);
    color: var(--accent-amber, #f0a500);
  }

  .story-badge {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: var(--sp-2xs) 0 0;
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

  .breadcrumb-row {
    display: flex;
    gap: var(--sp-xs);
    padding: 0 0 0 calc(var(--sp-sm) + 4px);
    margin-bottom: var(--sp-2xs);
    min-height: 12px;
    font-family: var(--font-mono);
    font-size: 9px;
    font-weight: 400;
    line-height: 1.3;
    font-variant-numeric: tabular-nums;
    color: var(--text-secondary);
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
    max-width: 100%;
    transition: color var(--duration-short) var(--ease-enter);
    user-select: text;
  }

  .breadcrumb-row.unresolved {
    color: var(--text-dim);
    opacity: 1;
  }

  .artifacts.out-first {
    margin-top: var(--sp-2xs);
  }

  /* S6: awaiting_input badge + round counter */
  .awaiting-badge {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 14px;
    height: 14px;
    border-radius: 50%;
    background: color-mix(in srgb, var(--accent-amber) 15%, transparent);
    color: var(--accent-amber);
    flex-shrink: 0;
  }

  .awaiting-text { color: var(--accent-amber); }

  .round-counter {
    font-family: var(--font-mono);
    font-size: 9px;
    font-variant-numeric: tabular-nums;
    color: var(--accent-amber);
    letter-spacing: 0.04em;
    padding-top: var(--sp-2xs);
    margin-top: var(--sp-2xs);
    border-top: 1px solid var(--border-subtle);
  }

  @media (prefers-reduced-motion: no-preference) {
    .process-node.awaiting .awaiting-badge {
      animation: awaiting-pulse 1800ms var(--ease-move) infinite;
    }
    .process-node.awaiting {
      animation: awaiting-node-pulse 1800ms var(--ease-move) infinite;
    }
  }

  @keyframes awaiting-pulse {
    0%, 100% {
      box-shadow: 0 0 0 0 color-mix(in srgb, var(--accent-amber) 0%, transparent);
      background: color-mix(in srgb, var(--accent-amber) 15%, transparent);
    }
    50% {
      box-shadow: var(--glow-spread) color-mix(in srgb, var(--accent-amber) 60%, transparent);
      background: color-mix(in srgb, var(--accent-amber) 30%, transparent);
    }
  }

  @keyframes awaiting-node-pulse {
    0%, 100% { box-shadow: 0 0 0 0 color-mix(in srgb, var(--accent-amber) 0%, transparent); }
    50% { box-shadow: 0 0 0 3px color-mix(in srgb, var(--accent-amber) 25%, transparent); }
  }

  .process-node.gate-flash {
    animation: gate-flash var(--duration-flash) var(--ease-exit) 1;
  }

  @keyframes gate-flash {
    0% { background: color-mix(in srgb, var(--accent-amber) 30%, transparent); }
    50% {
      background: color-mix(in srgb, var(--accent-green) 40%, transparent);
      box-shadow: 0 0 0 3px color-mix(in srgb, var(--accent-green) 35%, transparent);
    }
    100% { background: var(--bg-elevated); }
  }
</style>
