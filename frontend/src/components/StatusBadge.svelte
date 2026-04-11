<script>
  export let status = 'running';
  export let size = 'md'; // 'sm' or 'md'

  const colors = {
    running: 'var(--accent-green)',
    open: 'var(--accent-teal)',
    finished: 'var(--accent-red)',
    needs_response: 'var(--accent-red)',
    waiting: 'var(--accent-red)',
    error: 'var(--accent-red)',
    completed: 'var(--accent-red)',
    started: 'var(--accent-purple)',
    blocked: 'var(--accent-red)',
    done: 'var(--accent-blue)',
    terminal: 'var(--text-dim)',
  };

  const labels = {
    running: 'RUNNING',
    open: 'OPEN',
    finished: 'FINISHED',
    needs_response: 'WAITING',
    waiting: 'WAITING',
    error: 'ERROR',
    completed: 'DONE',
    started: 'STARTED',
    blocked: 'WAITING',
    done: 'DONE',
    terminal: 'TERMINAL',
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
  .badge.sm { font-size: var(--text-label); padding: 1px var(--sp-xs); }
  .badge.md { font-size: 11px; }
</style>
