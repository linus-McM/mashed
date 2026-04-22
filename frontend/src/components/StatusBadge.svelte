<script lang="ts">
  import type { StatusToken } from '../types/status';

  // `status` arrives from parents that still pass `string` (Wails event
  // payloads, snapshot data). The template narrows at the single lookup site
  // below so callers do not need to pre-cast.
  export let status: StatusToken | string = 'running';
  export let size: 'sm' | 'md' = 'md';

  // Exhaustive `Record<StatusToken, …>` — svelte-check will fail if a new
  // token lands in `types/status.ts` without a colour/label entry here.
  const colors: Record<StatusToken, string> = {
    running:        'var(--accent-green)',
    open:           'var(--accent-teal)',
    finished:       'var(--accent-red)',
    needs_response: 'var(--accent-red)',
    waiting:        'var(--accent-red)',
    error:          'var(--accent-red)',
    completed:      'var(--accent-red)',
    started:        'var(--accent-purple)',
    blocked:        'var(--accent-red)',
    done:           'var(--accent-blue)',
    queued:         'var(--text-dim)',
    terminal:       'var(--text-dim)',
  };

  const labels: Record<StatusToken, string> = {
    running:        'RUNNING',
    open:           'OPEN',
    finished:       'FINISHED',
    needs_response: 'WAITING',
    waiting:        'WAITING',
    error:          'ERROR',
    completed:      'DONE',
    started:        'STARTED',
    blocked:        'WAITING',
    done:           'DONE',
    queued:         'QUEUED',
    terminal:       'TERMINAL',
  };

  // Narrow once — after this the rest of the template is token-safe.
  $: token = (status as StatusToken);
  $: color = colors[token] ?? 'var(--text-dim)';
  $: label = labels[token] ?? String(status).toUpperCase();
</script>

<span class="badge {size}" style="--status-color: {color}">
  {label}
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
