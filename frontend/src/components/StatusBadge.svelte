<script>
  /** @type {'running' | 'error' | 'completed' | 'needs_response' | 'started'} */
  export let status = 'running';

  const colorMap = {
    running:        '#f0a500',
    error:          '#e84545',
    completed:      '#3d9eff',
    needs_response: '#00e57a',
    started:        '#9d6fff',
  };

  const labelMap = {
    running:        'RUNNING',
    error:          'ERROR',
    completed:      'COMPLETED',
    needs_response: 'NEEDS RESPONSE',
    started:        'STARTED',
  };

  $: color = colorMap[status] || '#4a5a6a';
  $: label = labelMap[status] || status.toUpperCase();
</script>

<span
  class="badge"
  style="background: {color}26; color: {color};"
>
  {label}
</span>

<style>
  .badge {
    display: inline-flex;
    align-items: center;
    padding: 1px 8px;
    border-radius: 999px;
    font-family: var(--font-ui);
    font-size: var(--text-label);
    font-weight: 600;
    letter-spacing: 0.04em;
    line-height: 1.6;
    white-space: nowrap;
    user-select: none;
  }
</style>
