<script>
  import { Handle, Position } from '@xyflow/svelte';
  import { Target } from 'lucide-svelte';
  import { formatConditionSummary, formatIterationDisplay } from './nodeUtils.js';

  export let data = {};
  export let id = '';
  export let selected = false;

  $: status = data.status || 'pending';
  $: label = data.label || 'Loop Until';
  $: config = data.config || {};
  $: maxIterations = config.maxIterations || 0;
  $: iterationCount = data.iterationCount;
  $: conditionSummary = formatConditionSummary(config);
  $: iterationDisplay = formatIterationDisplay(iterationCount, maxIterations);
</script>

<div class="loop-until-node" class:selected class:running={status === 'running'}>
  <div class="phase-bar" />

  <div class="node-body">
    <div class="node-header">
      <span class="node-icon">
        <Target size={13} />
      </span>
      <span class="node-label">{label}</span>
    </div>

    {#if conditionSummary}
      <div class="condition-info">{conditionSummary}</div>
    {/if}

    {#if iterationDisplay}
      <div class="iteration-info">{iterationDisplay}</div>
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

  <span class="handle-label body-label">body</span>
  <span class="handle-label exit-label">exit</span>

  <Handle type="target" position={Position.Left} />
  <Handle type="source" position={Position.Right} id="loop-body" style="top: 35%" />
  <Handle type="source" position={Position.Right} id="loop-exit" style="top: 65%" />
</div>

<style>
  .loop-until-node {
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md, 6px);
    min-width: 160px;
    max-width: 220px;
    overflow: hidden;
    font-family: var(--font-mono);
    transition: border-color 150ms ease, box-shadow 150ms ease;
    position: relative;
  }

  .loop-until-node.selected {
    border-color: var(--accent-green);
    box-shadow: 0 0 0 2px rgba(0, 229, 122, 0.35);
  }

  .loop-until-node.running {
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
    background: var(--accent-amber, #d29922);
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

  .node-icon {
    display: flex;
    align-items: center;
    flex-shrink: 0;
    color: var(--accent-amber, #d29922);
  }

  .node-label {
    font-size: 11px;
    font-weight: 600;
    color: var(--text-primary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .condition-info {
    font-size: 9px;
    color: var(--text-muted);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    margin-bottom: 2px;
  }

  .iteration-info {
    font-size: 9px;
    color: var(--text-dim);
    margin-bottom: 2px;
    font-variant-numeric: tabular-nums;
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

  .handle-label {
    position: absolute;
    font-size: 8px;
    font-weight: 600;
    font-family: var(--font-mono);
    color: var(--text-dim);
    right: 12px;
    pointer-events: none;
    text-transform: uppercase;
    letter-spacing: 0.3px;
  }

  .body-label {
    top: 30%;
  }

  .exit-label {
    top: 60%;
  }
</style>
