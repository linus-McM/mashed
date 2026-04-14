<script>
  // Svelte-Flow `command` node type. Mirrors ProcessNode envelope so
  // mixed canvases stay visually coherent.
  //
  // Cerebrum Do-Not-Repeat 2026-04-10: ZERO hex / rgb / rgba literals.
  // Every color must be `var(--token)` with NO fallback hex argument
  // (`var(--accent-green, #...)` is forbidden — fallbacks are how the
  // toxic neon leaked back in last time).
  import { Handle, Position } from '@xyflow/svelte';
  import { Terminal } from 'lucide-svelte';
  import { getNodePath, formatBreadcrumb } from '../../lib/bmad/nodePath';

  export let data = {};
  // svelte-ignore unused-export-let
  export let id = '';
  export let selected = false;

  $: config = data.config || {};
  $: status = data.status || 'pending';
  $: commandName = config.commandName || data.label || 'command';
  $: commandDescription = config.commandDescription || '';

  $: inputPath = getNodePath({ data }, 'in');
  $: outputPath = getNodePath({ data }, 'out');
  $: inputBreadcrumb = formatBreadcrumb(inputPath);
  $: outputBreadcrumb = formatBreadcrumb(outputPath);
</script>

<div
  class="command-node"
  class:selected
  class:running={status === 'running'}
  data-node-type="command"
>
  <div class="accent-stripe" />

  <div class="body">
    <div class="header">
      <span class="icon">
        <Terminal size={13} />
      </span>
      <span class="name">{commandName}</span>
    </div>
    {#if commandDescription}
      <div class="description">{commandDescription}</div>
    {/if}

    {#if inputPath}
      <div class="breadcrumb-row" title={inputPath}>
        {inputBreadcrumb}
      </div>
    {/if}
    <div
      class="breadcrumb-row"
      class:unresolved={!outputPath}
      title={outputPath || 'unresolved'}
    >
      {outputBreadcrumb}
    </div>

    <div class="status-row">
      {#if status === 'pending'}
        <span class="status-dot pending" />
        <span class="status-text">pending</span>
      {:else if status === 'running'}
        <span class="status-dot running-dot" />
        <span class="status-text running-text">running</span>
      {:else if status === 'complete'}
        <span class="status-check">✓</span>
        <span class="status-text complete-text">complete</span>
      {:else if status === 'failed'}
        <span class="status-x">✕</span>
        <span class="status-text failed-text">failed</span>
      {:else if status === 'skipped'}
        <span class="status-dash">—</span>
        <span class="status-text">skipped</span>
      {/if}
    </div>
  </div>

  <Handle type="target" position={Position.Left} />
  <Handle type="source" position={Position.Right} />
</div>

<style>
  .command-node {
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    min-width: 160px;
    max-width: 220px;
    overflow: hidden;
    font-family: var(--font-mono);
    transition:
      border-color var(--duration-medium) var(--ease-enter),
      box-shadow var(--duration-medium) var(--ease-enter);
  }

  .command-node:hover {
    border-color: var(--border-emphasis);
  }

  .command-node.selected {
    border-color: var(--accent-green);
    box-shadow: 0 0 0 2px color-mix(in srgb, var(--accent-green) 35%, transparent);
  }

  .command-node.running {
    border-color: var(--accent-green);
    animation: node-pulse 2s var(--ease-move) infinite;
  }

  @keyframes node-pulse {
    0%, 100% { box-shadow: 0 0 0 0 color-mix(in srgb, var(--accent-green) 0%, transparent); }
    50%      { box-shadow: 0 0 0 3px color-mix(in srgb, var(--accent-green) 25%, transparent); }
  }

  .accent-stripe {
    height: 4px;
    width: 100%;
    background: var(--accent-green);
  }

  .body {
    padding: var(--sp-sm) var(--sp-md) var(--sp-xs);
  }

  .header {
    display: flex;
    align-items: center;
    gap: var(--sp-sm);
  }

  .icon {
    display: flex;
    align-items: center;
    flex-shrink: 0;
    color: var(--accent-green);
  }

  .name {
    font-family: var(--font-ui);
    font-size: var(--text-label);
    font-weight: 600;
    color: var(--text-primary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .description {
    margin-top: var(--sp-2xs);
    font-family: var(--font-mono);
    font-size: 9px;
    color: var(--text-muted);
    line-height: 1.3;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .status-row {
    display: flex;
    align-items: center;
    gap: var(--sp-xs);
    margin-top: var(--sp-xs);
    padding-top: var(--sp-xs);
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
    background: var(--text-muted);
  }

  .status-dot.running-dot {
    background: var(--accent-green);
    animation: dot-pulse 1.5s ease-in-out infinite;
  }

  @keyframes dot-pulse {
    0%, 100% { opacity: 0.4; }
    50%      { opacity: 1; }
  }

  .running-text { color: var(--accent-green); }

  .status-check {
    font-size: var(--text-label);
    color: var(--accent-green);
    line-height: 1;
  }

  .complete-text { color: var(--accent-green); }

  .status-x {
    font-size: var(--text-label);
    color: var(--accent-red);
    line-height: 1;
  }

  .failed-text { color: var(--accent-red); }

  .status-dash {
    font-size: var(--text-label);
    color: var(--text-muted);
    line-height: 1;
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
    color: var(--text-muted);
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
    max-width: 100%;
    transition: color var(--duration-short) var(--ease-enter);
    user-select: text;
  }

  .breadcrumb-row.unresolved {
    opacity: 0.7;
  }
</style>
