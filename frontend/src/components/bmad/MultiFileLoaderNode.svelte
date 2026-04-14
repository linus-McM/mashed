<script>
  import { Handle, Position } from '@xyflow/svelte';
  import { FolderTree } from 'lucide-svelte';
  import { formatBreadcrumb } from '../../lib/bmad/nodePath';
  import { parseEntries, entryLabelFor } from '../../lib/bmad/multiFileEntries';

  export let data = {};
  // svelte-ignore unused-export-let
  export let id = '';
  export let selected = false;

  $: entries = parseEntries(data?.config?.entries ?? '');
  $: status = data?.status || 'pending';
  $: label = data?.label || 'Multi File Loader';
</script>

<div class="process-node multi-file-loader" class:selected class:running={status === 'running'}>
  <div class="phase-bar" />

  <div class="node-body">
    <div class="node-header">
      <span class="role-icon">
        <FolderTree size={13} />
      </span>
      <span class="node-label">{label}</span>
    </div>

    <div class="path-list">
      {#if entries.length === 0}
        <div class="empty-state">&#8212; (no files)</div>
      {:else}
        {#each entries as entry, i (i)}
          {@const prefix = entryLabelFor(entry, i)}
          <div class="breadcrumb-item">
            <div class="artifacts">
              <span class="artifact-label" title="output: {prefix}">{prefix}</span>
            </div>
            <div
              class="breadcrumb-row"
              class:unresolved={!entry.path}
              title={entry.path || 'unresolved'}
            >
              {formatBreadcrumb(entry.path)}
            </div>
            <Handle type="source" position={Position.Right} id={prefix} />
          </div>
        {/each}
      {/if}
    </div>

    <div class="node-footer">
      <span class="count-badge">{entries.length} files</span>
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

<style>
  .process-node {
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    min-width: 180px;
    max-width: 240px;
    overflow: hidden;
    font-family: var(--font-mono);
    transition: border-color 150ms ease, box-shadow 150ms ease;
  }

  .process-node.selected {
    border-color: var(--accent-green);
    box-shadow: 0 0 0 2px color-mix(in srgb, var(--accent-green) 35%, transparent);
  }

  .process-node.running {
    border-color: var(--accent-green);
    animation: node-pulse 2s ease-in-out infinite;
  }

  @keyframes node-pulse {
    0%, 100% { box-shadow: 0 0 0 0 color-mix(in srgb, var(--accent-green) 0%, transparent); }
    50% { box-shadow: 0 0 0 3px color-mix(in srgb, var(--accent-green) 25%, transparent); }
  }

  .phase-bar {
    height: 4px;
    width: 100%;
    background: var(--text-dim);
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
    color: var(--text-dim);
  }

  .node-label {
    font-size: 11px;
    font-weight: 600;
    color: var(--text-primary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .path-list {
    display: flex;
    flex-direction: column;
    gap: 2px;
    margin-top: var(--sp-2xs);
    max-height: 96px;
    overflow-y: auto;
  }

  .breadcrumb-item {
    display: flex;
    flex-direction: column;
    gap: 0;
    position: relative;
  }

  .empty-state {
    font-size: 9px;
    color: var(--text-muted);
    opacity: 0.5;
    padding: 2px 0;
    font-family: var(--font-mono);
  }

  .artifacts {
    display: flex;
    gap: var(--sp-xs);
    color: var(--text-primary);
    line-height: 1.3;
    margin-bottom: var(--sp-2xs);
    min-height: 12px;
  }

  .artifact-label {
    color: var(--text-dim);
    font-size: 9px;
    font-weight: 600;
    flex-shrink: 0;
    font-family: var(--font-mono);
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
    user-select: text;
  }

  .breadcrumb-row.unresolved {
    opacity: 0.7;
  }

  .node-footer {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin: var(--sp-xs) calc(-1 * 10px) 0;
    padding: var(--sp-xs) 10px 0;
    border-top: 1px solid var(--border-subtle);
  }

  .count-badge {
    font-size: 9px;
    font-family: var(--font-mono);
    font-weight: 400;
    color: var(--text-muted);
    letter-spacing: 0.3px;
  }

  .status-row {
    display: flex;
    align-items: center;
    gap: var(--sp-xs);
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
    50% { opacity: 1; }
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
</style>
