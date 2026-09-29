<script lang="ts">
  // Plan v3 Story 18 — reusable pulldown for Backend / Model / Policy.
  // Mirrors the existing .theme-picker-wrap popover pattern in TitleBar.
  export let label: string = '';
  export let value: string = '';
  export let options: Array<{ value: string; label?: string; disabled?: boolean; hint?: string }> = [];
  export let disabled = false;
  export let glyph: string = '';
  export let onSelect: (v: string) => void;

  let open = false;

  export function close(): void {
    open = false;
  }

  function toggle(e: MouseEvent) {
    e.stopPropagation();
    if (disabled) return;
    open = !open;
  }

  function handleKey(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      open = false;
    }
    if ((e.key === 'Enter' || e.key === ' ') && !open) {
      e.preventDefault();
      open = true;
    }
  }

  function select(v: string, rowDisabled?: boolean) {
    if (rowDisabled) return;
    onSelect(v);
    open = false;
  }

  $: visibleLabel = truncate((options.find((o) => o.value === value)?.label || value), 14);

  function truncate(s: string, n: number) {
    if (!s) return '';
    return s.length <= n ? s : s.slice(0, n - 1) + '…';
  }
</script>

<div class="titlebar-selector-wrap">
  <button
    class="titlebar-btn titlebar-selector"
    class:open
    class:disabled
    on:click={toggle}
    on:keydown={handleKey}
    {disabled}
    title={label + ': ' + value}
    aria-haspopup="menu"
    aria-expanded={open}
  >
    {#if glyph}
      <span class="glyph">{glyph}</span>
    {/if}
    <span class="label">{visibleLabel}</span>
  </button>

  {#if open}
    <!-- svelte-ignore a11y-no-static-element-interactions -->
    <div class="selector-popover" role="menu" tabindex="0" on:click|stopPropagation on:keydown={handleKey}>
      <div class="sel-caption">{label}</div>
      {#each options as opt (opt.value)}
        <button
          class="sel-item"
          class:active={opt.value === value}
          class:disabled={opt.disabled}
          on:click={() => select(opt.value, opt.disabled)}
          role="menuitem"
          tabindex="0"
        >
          <span class="sel-label">{opt.label || opt.value}</span>
          {#if opt.disabled}
            <span class="sel-badge">offline</span>
          {:else if opt.value === value}
            <span class="sel-tick">✓</span>
          {/if}
          {#if opt.hint}
            <span class="sel-hint">{opt.hint}</span>
          {/if}
        </button>
      {/each}
    </div>
  {/if}
</div>

<style>
  .titlebar-selector-wrap {
    position: relative;
    display: inline-flex;
  }

  .titlebar-btn.titlebar-selector {
    display: inline-flex;
    align-items: center;
    gap: var(--sp-xs, 4px);
    padding: 0 var(--sp-sm, 8px);
    font-family: var(--font-code, 'JetBrains Mono', monospace);
    font-size: var(--text-label, 11px);
    font-weight: 500;
    color: var(--text-muted, #a0a0a0);
    background: transparent;
    border: 1px solid transparent;
    border-radius: var(--radius-sm, 4px);
    height: 24px;
    cursor: pointer;
    transition: background var(--duration-short, 120ms) var(--ease-enter, ease-out),
                color var(--duration-short, 120ms) var(--ease-enter, ease-out),
                border-color var(--duration-short, 120ms) var(--ease-enter, ease-out);
  }

  .titlebar-btn.titlebar-selector:hover {
    color: var(--text-primary, #e0e0e0);
    background: var(--bg-elevated, #2a2a2a);
    border-color: var(--border-subtle, #333);
  }

  .titlebar-btn.titlebar-selector.open {
    color: var(--text-primary);
    background: var(--bg-active, #353535);
    border-color: var(--border-emphasis, #555);
  }

  .titlebar-btn.titlebar-selector.disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .glyph {
    font-size: 13px;
    line-height: 1;
    color: var(--accent-green, #6bcf7f);
  }

  .label {
    max-width: 14ch;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .selector-popover {
    position: absolute;
    top: calc(100% + var(--sp-sm, 8px));
    right: 0;
    min-width: 180px;
    padding: var(--sp-xs, 4px);
    background: var(--bg-elevated, #2a2a2a);
    border: 1px solid var(--border-emphasis, #555);
    border-radius: var(--radius-lg, 8px);
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
    z-index: 200;
    display: flex;
    flex-direction: column;
    gap: 0;
  }

  .sel-caption {
    padding: var(--sp-xs, 4px) var(--sp-sm, 8px);
    font-family: var(--font-code);
    font-size: var(--text-xs, 10px);
    font-weight: 600;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--text-dim, #888);
  }

  .sel-item {
    display: flex;
    align-items: center;
    gap: var(--sp-xs, 4px);
    padding: var(--sp-xs, 4px) var(--sp-sm, 8px);
    font-family: var(--font-code);
    font-size: var(--text-label, 11px);
    color: var(--text-primary);
    background: transparent;
    border: none;
    border-radius: var(--radius-md, 6px);
    text-align: left;
    cursor: pointer;
    transition: background var(--duration-short, 120ms) var(--ease-enter, ease-out);
  }

  .sel-item:hover:not(.disabled) {
    background: var(--bg-active, #353535);
  }

  .sel-item.active {
    color: var(--accent-green, #6bcf7f);
  }

  .sel-item.disabled {
    color: var(--text-dim, #888);
    cursor: not-allowed;
  }

  .sel-label {
    flex: 1;
  }

  .sel-tick {
    color: var(--accent-green);
  }

  .sel-badge {
    padding: 0 var(--sp-sm, 8px);
    font-size: 9px;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--accent-amber, #ffaa44);
    border: 1px solid currentColor;
    border-radius: var(--radius-sm, 4px);
  }

  .sel-hint {
    display: block;
    max-height: var(--sel-tooltip-max, 48px);
    overflow: hidden;
    font-size: var(--text-xs, 10px);
    color: var(--text-dim);
  }
</style>
