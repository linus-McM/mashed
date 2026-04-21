<script>
  // Syntax-less code block with copy affordance — spec §6.2, Design Brief §5.
  //
  // Rendered on `--bg-deepest` to echo the Terminal pane depth cue. No
  // syntax highlighting in U6 — spec defers that to a later story. The copy
  // button is the one accent moment: idle text-dim → accent-green on copy
  // success for 1200ms then decays back.
  import { onDestroy } from 'svelte';
  import { Copy, Check } from 'lucide-svelte';

  export let lang = '';
  export let content = '';
  export let copyable = false;

  let copied = false;
  /** @type {ReturnType<typeof setTimeout> | null} */
  let resetTimer = null;

  async function onCopy() {
    try {
      await navigator.clipboard.writeText(content);
      copied = true;
      if (resetTimer) clearTimeout(resetTimer);
      resetTimer = setTimeout(() => {
        copied = false;
        resetTimer = null;
      }, 1200);
    } catch {
      copied = false;
    }
  }

  onDestroy(() => {
    if (resetTimer) clearTimeout(resetTimer);
  });
</script>

<div class="code-block" data-testid="code-block">
  {#if lang || copyable}
    <div class="header">
      {#if lang}
        <span class="lang">{lang}</span>
      {/if}
      {#if copyable}
        <button
          type="button"
          class="copy"
          class:copied
          data-testid="code-copy-button"
          on:click={onCopy}
        >
          <svelte:component this={copied ? Check : Copy} size={14} />
          <span class="label">{copied ? 'Copied' : 'Copy'}</span>
        </button>
      {/if}
    </div>
  {/if}
  <pre class="content" data-testid="code-block-content"><code>{content}</code></pre>
</div>

<style>
  .code-block {
    background: var(--bg-deepest);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    padding: var(--sp-sm) var(--sp-md);
    color: var(--text-primary);
    font-family: var(--font-code);
  }

  .header {
    display: flex;
    align-items: center;
    gap: var(--sp-sm);
    margin-bottom: var(--sp-xs);
  }

  .lang {
    color: var(--text-muted);
    font-family: var(--font-mono);
    font-size: var(--text-label);
    font-weight: 500;
    letter-spacing: 0.05em;
    text-transform: uppercase;
    line-height: 1;
  }

  .copy {
    margin-left: auto;
    display: inline-flex;
    align-items: center;
    gap: var(--sp-2xs);
    background: transparent;
    border: 0;
    padding: var(--sp-2xs) var(--sp-xs);
    border-radius: var(--radius-sm);
    color: var(--text-dim);
    font-family: var(--font-ui);
    font-size: var(--text-label);
    font-weight: 500;
    cursor: pointer;
    transition: color var(--duration-short) var(--ease-exit),
                background var(--duration-short) var(--ease-exit);
  }
  .copy:hover {
    color: var(--accent-green);
  }
  .copy.copied {
    color: var(--accent-green);
    background: color-mix(in srgb, var(--accent-green) 15%, transparent);
  }

  .label {
    line-height: 1;
  }

  .content {
    margin: 0;
    font-family: var(--font-code);
    font-size: var(--text-body);
    line-height: 1.55;
    font-variant-numeric: tabular-nums;
    overflow-x: auto;
    white-space: pre;
  }
  .content :global(code) {
    font-family: inherit;
  }
</style>
