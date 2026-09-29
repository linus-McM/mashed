<script>
  // skills-cmd-03: bottom-anchored, non-blocking sentinel toast surfaced
  // when a command node transitions to `failed` during execution. This is
  // the Phase 2 "command nodes not yet runnable" signal — Phase 3 replaces
  // it with real execution feedback.
  //
  // Design Brief (story skills-cmd-03 — Design Brief section is source of
  // truth) mandates:
  //   - bottom-anchored, translate(-50%, 0) centered
  //   - red-tinted elevated surface via color-mix(--accent-red)
  //   - uppercase red label prefix in var(--text-label)
  //   - var(--font-ui) / var(--text-body) body copy
  //   - auto-dismiss at 3000ms; Escape dismisses immediately; click does
  //     NOT dismiss (user may want to copy the text)
  //   - slide + fade entry at 150ms var(--ease-enter)
  //
  // Cerebrum Do-Not-Repeat 2026-04-10: ZERO hex literals, ZERO rgb/rgba,
  // ZERO toxic neon green. Every color must resolve through `var(--token)`
  // — even fallback arguments are forbidden because that's the exact hole
  // the toxic neon crept through last time.
  //
  // TODO: the box-shadow literal below is hand-rolled because `style.css`
  // has no `--shadow-*` tokens yet. Promote to `--shadow-lifted` once a
  // second component (modal, dropdown, etc.) needs the same elevation.

  import { onMount, onDestroy } from 'svelte';
  import { fly, fade } from 'svelte/transition';

  export let message = '';
  /** @type {() => void} */
  export let onDismiss = () => {};

  const AUTO_DISMISS_MS = 3000;
  const LABEL_PREFIX = 'Command nodes not yet runnable';

  /** @type {ReturnType<typeof setTimeout> | null} */
  let dismissTimer = null;

  onMount(() => {
    dismissTimer = setTimeout(() => {
      onDismiss();
    }, AUTO_DISMISS_MS);
    window.addEventListener('keydown', handleKeydown);
  });

  onDestroy(() => {
    if (dismissTimer) clearTimeout(dismissTimer);
    window.removeEventListener('keydown', handleKeydown);
  });

  /** @param {KeyboardEvent} event */
  function handleKeydown(event) {
    if (event.key === 'Escape') {
      onDismiss();
    }
  }
</script>

<div
  class="canvas-failure-toast"
  role="status"
  aria-live="polite"
  data-testid="canvas-failure-toast"
  in:fly={{ y: 16, duration: 150 }}
  out:fade={{ duration: 150 }}
>
  <span class="label">{LABEL_PREFIX}</span>
  <span class="body">{message}</span>
</div>

<style>
  .canvas-failure-toast {
    position: absolute;
    bottom: var(--sp-lg);
    left: 50%;
    transform: translateX(-50%);
    display: flex;
    align-items: center;
    gap: var(--sp-sm);
    padding: var(--sp-sm) var(--sp-md);
    background: color-mix(in srgb, var(--accent-red) 12%, var(--bg-elevated));
    border: 1px solid color-mix(in srgb, var(--accent-red) 40%, transparent);
    border-radius: var(--radius-md);
    color: var(--text-primary);
    font-family: var(--font-ui);
    font-size: var(--text-body);
    /* TODO(skills-cmd-03): promote to `--shadow-lifted` when a second
       component needs the same elevation. Literal stays until then. */
    box-shadow: 0 8px 24px
      color-mix(in srgb, var(--bg-deepest) 70%, transparent);
    z-index: 50;
    max-width: min(480px, 80%);
    pointer-events: auto;
  }

  .label {
    color: var(--accent-red);
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    font-size: var(--text-label);
    flex-shrink: 0;
  }

  .body {
    color: var(--text-primary);
    font-weight: 500;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
</style>
