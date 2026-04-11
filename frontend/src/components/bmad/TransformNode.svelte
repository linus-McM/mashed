<script>
  import { Handle, Position } from '@xyflow/svelte';
  import { Wand2 } from 'lucide-svelte';

  export let data = {};
  // svelte-ignore unused-export-let
  export let id = '';
  export let selected = false;

  $: status = data.status || 'pending';
  $: label = data.label || 'Transform';
  $: config = data.config || {};
  $: extractType = config.extractType || '';
</script>

<div class="transform-node" class:selected class:running={status === 'running'}>
  <div class="phase-bar" />

  <div class="node-body">
    <div class="node-header">
      <span class="node-icon">
        <Wand2 size={12} />
      </span>
      <span class="node-label">{label}</span>
    </div>

    {#if extractType}
      <div class="extract-type">{extractType}</div>
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
  </div>

  <Handle type="target" position={Position.Left} />
  <Handle type="source" position={Position.Right} />
</div>

<style>
  .transform-node {
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: 20px;
    min-width: 100px;
    max-width: 160px;
    overflow: hidden;
    font-family: var(--font-mono);
    transition: border-color 150ms ease, box-shadow 150ms ease;
  }

  .transform-node.selected {
    border-color: var(--accent-green);
    box-shadow: 0 0 0 2px color-mix(in srgb, var(--accent-green) 35%, transparent);
  }

  .transform-node.running {
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
    background: var(--accent-blue, #58a6ff);
    border-radius: 20px 20px 0 0;
  }

  .node-body {
    padding: var(--sp-xs) 10px;
  }

  .node-header {
    display: flex;
    align-items: center;
    gap: var(--sp-xs);
    margin-bottom: 2px;
  }

  .node-icon {
    display: flex;
    align-items: center;
    flex-shrink: 0;
    color: var(--accent-blue, #58a6ff);
  }

  .node-label {
    font-size: var(--text-label);
    font-weight: 600;
    color: var(--text-primary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .extract-type {
    font-size: 9px;
    color: var(--text-muted);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    margin-bottom: 2px;
  }

  .status-row {
    display: flex;
    align-items: center;
    gap: 4px;
    margin-top: var(--sp-2xs);
    padding-top: var(--sp-2xs);
    border-top: 1px solid var(--border-subtle);
  }

  .status-text {
    font-size: 8px;
    color: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: 0.3px;
  }

  .status-dot {
    width: 5px;
    height: 5px;
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
    font-size: 9px;
    color: var(--accent-green, #00e57a);
    line-height: 1;
  }

  .complete-text { color: var(--accent-green, #00e57a); }

  .status-x {
    font-size: 9px;
    color: var(--accent-red, #f85149);
    line-height: 1;
  }

  .failed-text { color: var(--accent-red, #f85149); }

  .status-dash {
    font-size: 9px;
    color: var(--text-muted);
    line-height: 1;
  }
</style>
