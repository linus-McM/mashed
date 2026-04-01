<script>
  export let status = 'running';
  export let size = 'md'; // 'sm' or 'md'

  const colors = {
    running: 'var(--accent-amber)',
    error: 'var(--accent-red)',
    completed: 'var(--accent-blue)',
    needs_response: 'var(--accent-green)',
    started: 'var(--accent-purple)',
    blocked: 'var(--accent-amber)',
    done: 'var(--accent-blue)',
  };

  const labels = {
    running: 'RUNNING',
    error: 'ERROR',
    completed: 'DONE',
    needs_response: 'NEEDS RESPONSE',
    started: 'STARTED',
    blocked: 'BLOCKED',
    done: 'DONE',
  };
</script>

<span class="badge {size}" style="--status-color: {colors[status] || 'var(--text-dim)'}">
  {labels[status] || status.toUpperCase()}
</span>

<style>
  .badge {
    display: inline-block;
    padding: 2px 8px;
    border-radius: var(--radius-md);
    background: color-mix(in srgb, var(--status-color) 15%, transparent);
    color: var(--status-color);
    font-family: var(--font-ui);
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }
  .badge.sm { font-size: 10px; padding: 1px 6px; }
  .badge.md { font-size: 11px; }
</style>
