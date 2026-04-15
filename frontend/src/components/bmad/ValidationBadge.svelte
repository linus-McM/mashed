<script>
  import { onDestroy, tick } from 'svelte';
  import { AlertTriangle, Info } from 'lucide-svelte';

  /** @type {{ field: string, message: string, severity: string }[]} */
  export let issues = [];

  /** Unique ID for aria-describedby linking. */
  export let assetId = '';

  let showTooltip = false;
  let badgeEl;
  let tooltipEl;
  let hoverTimeout;
  let leaveTimeout;

  // Highest severity wins: warn > info.
  $: severity = issues.some(i => i.severity === 'warn') ? 'warn' : 'info';
  $: tooltipId = `asset-${assetId}-issues`;

  async function openTooltip() {
    clearTimeout(leaveTimeout);
    showTooltip = true;
    await tick();
    positionTooltip();
  }

  function closeTooltip() {
    clearTimeout(hoverTimeout);
    showTooltip = false;
  }

  function handleMouseEnter() {
    clearTimeout(leaveTimeout);
    hoverTimeout = setTimeout(openTooltip, 200);
  }

  function handleMouseLeave() {
    clearTimeout(hoverTimeout);
    leaveTimeout = setTimeout(closeTooltip, 100);
  }

  function handleClick(e) {
    e.stopPropagation();
    if (showTooltip) {
      closeTooltip();
    } else {
      openTooltip();
    }
  }

  function handleKeydown(e) {
    if (e.key === 'Escape') {
      closeTooltip();
    } else if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      handleClick(e);
    }
  }

  function positionTooltip() {
    if (!badgeEl || !tooltipEl) return;
    const rect = badgeEl.getBoundingClientRect();
    const tooltipWidth = 280;
    const viewportWidth = window.innerWidth;

    // Position right of badge by default, left if clipped.
    if (rect.right + tooltipWidth + 8 < viewportWidth) {
      tooltipEl.style.left = `${rect.width + 4}px`;
      tooltipEl.style.right = 'auto';
    } else {
      tooltipEl.style.right = `${rect.width + 4}px`;
      tooltipEl.style.left = 'auto';
    }
    tooltipEl.style.top = '0';
  }

  onDestroy(() => {
    clearTimeout(hoverTimeout);
    clearTimeout(leaveTimeout);
  });
</script>

<!-- svelte-ignore a11y-no-static-element-interactions -->
<button
  class="validation-badge {severity}"
  bind:this={badgeEl}
  on:mouseenter={handleMouseEnter}
  on:mouseleave={handleMouseLeave}
  on:click={handleClick}
  on:keydown={handleKeydown}
  on:focus={openTooltip}
  on:blur={closeTooltip}
  aria-describedby={showTooltip ? tooltipId : undefined}
  tabindex="0"
  type="button"
>
  {#if severity === 'warn'}
    <AlertTriangle size={10} />
  {:else}
    <Info size={10} />
  {/if}

  {#if showTooltip}
    <div
      class="validation-tooltip"
      id={tooltipId}
      bind:this={tooltipEl}
      role="tooltip"
      on:mouseenter={() => clearTimeout(leaveTimeout)}
      on:mouseleave={handleMouseLeave}
    >
      {#each issues as issue}
        <div class="issue">
          <span class="issue-dot {issue.severity}" />
          <span class="issue-text">
            <span class="field">{issue.field}:</span> {issue.message}
          </span>
        </div>
      {/each}
    </div>
  {/if}
</button>

<style>
  .validation-badge {
    position: relative;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 14px;
    height: 14px;
    margin-left: var(--sp-xs);
    padding: 0;
    border: none;
    background: transparent;
    border-radius: 50%;
    cursor: pointer;
    flex-shrink: 0;
    transition: background var(--duration-short) var(--ease-enter);
  }
  .validation-badge.warn {
    background: color-mix(in srgb, var(--accent-amber) 15%, transparent);
    color: var(--accent-amber);
  }
  .validation-badge.info {
    background: color-mix(in srgb, var(--accent-blue) 15%, transparent);
    color: var(--accent-blue);
  }
  .validation-badge:hover.warn {
    background: color-mix(in srgb, var(--accent-amber) 25%, transparent);
  }
  .validation-badge:hover.info {
    background: color-mix(in srgb, var(--accent-blue) 25%, transparent);
  }

  .validation-tooltip {
    position: absolute;
    z-index: 200;
    max-width: 280px;
    padding: var(--sp-sm) var(--sp-md);
    background: var(--bg-elevated);
    border: 1px solid var(--border-emphasis);
    border-radius: var(--radius-md);
    box-shadow: 0 8px 24px color-mix(in srgb, var(--bg-deepest) 80%, transparent);
    display: flex;
    flex-direction: column;
    gap: var(--sp-xs);
    font-family: var(--font-ui);
    font-size: var(--text-label);
    color: var(--text-primary);
    line-height: 1.4;
    white-space: normal;
    text-align: left;
    pointer-events: auto;
  }
  .issue {
    display: flex;
    align-items: flex-start;
    gap: var(--sp-xs);
  }
  .issue-dot {
    width: 4px;
    height: 4px;
    border-radius: 50%;
    margin-top: var(--sp-sm);
    flex-shrink: 0;
  }
  .issue-dot.warn { background: var(--accent-amber); }
  .issue-dot.info { background: var(--accent-blue); }
  .field {
    font-family: var(--font-mono);
    color: var(--text-dim);
  }
</style>
