<script>
  /** @type {string} */
  export let raw;
  export let expandedByDefault = false;
  export let triggered = false;

  // Snapshot on instantiation so later prop flips can't override the user's
  // expand/collapse interaction. The modal unmounts on close, so each new
  // suspension gets a fresh component — matches Q7 "no persistence" intent.
  let shouldBeOpen = expandedByDefault && triggered;
</script>

{#if raw !== ''}
  <details
    class="raw-view-toggle"
    data-testid="raw-view-toggle"
    open={shouldBeOpen}
  >
    <summary>
      <span class="chevron" aria-hidden="true">▸</span>
      <span class="label">View raw</span>
    </summary>
    <pre class="raw-content">{raw}</pre>
  </details>
{/if}

<style>
  .raw-view-toggle {
    border: 1px dashed var(--border-subtle);
    border-radius: var(--radius-md);
    background: transparent;
    overflow: hidden;
  }

  .raw-view-toggle[open] {
    border: 1px solid var(--border-emphasis);
    background: var(--bg-surface);
  }

  .raw-view-toggle:not([open]):hover {
    border-color: var(--border-emphasis);
  }

  .raw-view-toggle:not([open]):hover summary {
    color: var(--text-primary);
  }

  summary {
    display: flex;
    align-items: center;
    gap: var(--sp-xs);
    padding: var(--sp-xs) var(--sp-sm);
    cursor: pointer;
    user-select: none;
    color: var(--text-dim);
    font-family: var(--font-mono);
    font-size: var(--text-body);
    font-weight: 500;
    list-style: none;
  }

  .raw-view-toggle[open] summary {
    color: var(--text-primary);
  }

  summary::-webkit-details-marker {
    display: none;
  }

  summary:focus-visible {
    outline: 1px solid var(--accent-green);
    outline-offset: 2px;
  }

  .chevron {
    display: inline-flex;
    width: var(--sp-md);
    height: var(--sp-md);
    align-items: center;
    justify-content: center;
    color: var(--text-muted);
    transition: transform var(--duration-medium) var(--ease-move);
  }

  .raw-view-toggle[open] .chevron {
    transform: rotate(90deg);
  }

  .raw-content {
    margin: 0;
    padding: var(--sp-sm) var(--sp-md);
    background: var(--bg-deepest);
    border-top: 1px solid var(--border-subtle);
    font-family: var(--font-code);
    font-size: var(--text-body);
    line-height: 1.5;
    white-space: pre;
    overflow: auto;
    max-height: var(--raw-view-max-height);
    color: var(--text-primary);
  }

  @media (prefers-reduced-motion: reduce) {
    .chevron {
      transition: none;
    }
  }
</style>
