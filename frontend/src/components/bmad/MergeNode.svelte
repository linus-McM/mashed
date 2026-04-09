<script>
  import { Handle, Position } from '@xyflow/svelte';
  import { Merge } from 'lucide-svelte';

  export let data = {};
  // svelte-ignore unused-export-let
  export let id = '';
  export let selected = false;

  $: status = data.status || 'pending';
  $: label = data.label || 'Merge';
</script>

<div class="merge-node" class:selected class:running={status === 'running'}>
  <div class="diamond-wrapper">
    <div class="diamond-shape">
      <div class="phase-bar" />
      <div class="diamond-content">
        <div class="node-header">
          <span class="node-icon">
            <Merge size={13} />
          </span>
          <span class="node-label">{label}</span>
        </div>

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
    </div>
  </div>

  <Handle type="target" position={Position.Left} />
  <Handle type="source" position={Position.Right} />
</div>

<style>
  .merge-node {
    position: relative;
    font-family: var(--font-mono);
    transition: filter 150ms ease;
  }

  .diamond-wrapper {
    width: 130px;
    height: 100px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .diamond-shape {
    width: 100px;
    height: 100px;
    transform: rotate(45deg);
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md, 6px);
    overflow: hidden;
    transition: border-color 150ms ease, box-shadow 150ms ease;
  }

  .merge-node.selected .diamond-shape {
    border-color: var(--accent-green);
    box-shadow: 0 0 0 2px rgba(0, 229, 122, 0.35);
  }

  .merge-node.running .diamond-shape {
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
    background: var(--text-dim, #4a5a6a);
  }

  .diamond-content {
    transform: rotate(-45deg);
    padding: 12px 6px 6px;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: calc(100% - 4px);
    box-sizing: border-box;
  }

  .node-header {
    display: flex;
    align-items: center;
    gap: 4px;
    margin-bottom: 2px;
  }

  .node-icon {
    display: flex;
    align-items: center;
    flex-shrink: 0;
    color: var(--text-dim, #4a5a6a);
  }

  .node-label {
    font-size: 10px;
    font-weight: 600;
    color: var(--text-primary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    max-width: 60px;
  }

  .status-row {
    display: flex;
    align-items: center;
    gap: 4px;
    margin-top: 2px;
    padding-top: 2px;
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
