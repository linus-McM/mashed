<script>
  import { createEventDispatcher, onMount, tick } from 'svelte';

  /** @type {import('../../../stores/interactiveInput').PendingPrompt} */
  export let prompt;
  export let disabled = false;
  export let validationError = '';

  /** @type {import('svelte').EventDispatcher<{ submit: { value: string } }>} */
  const dispatch = createEventDispatcher();

  let selected = -1;
  /** @type {HTMLDivElement | null} */
  let groupEl = null;

  $: options = prompt?.options ?? [];

  onMount(async () => {
    await tick();
    focusOption(0);
  });

  /** @param {number} index */
  function focusOption(index) {
    const buttons = groupEl?.querySelectorAll('[role="radio"]');
    if (!buttons || buttons.length === 0) return;
    const clamped = Math.max(0, Math.min(index, buttons.length - 1));
    /** @type {HTMLElement} */ (buttons[clamped]).focus();
  }

  /** @param {number} index */
  function choose(index) {
    if (disabled) return;
    selected = index;
  }

  function submit() {
    if (disabled || selected < 0) return;
    dispatch('submit', { value: options[selected] });
  }

  /** @param {KeyboardEvent} e @param {number} index */
  function onKeydown(e, index) {
    if (disabled) return;
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      if (selected === -1) {
        // First Enter selects; second Enter submits.
        selected = index;
        return;
      }
      if (selected === index) {
        submit();
      } else {
        selected = index;
      }
      return;
    }
    if (e.key === 'ArrowDown' || e.key === 'j') {
      e.preventDefault();
      focusOption(index + 1);
    } else if (e.key === 'ArrowUp' || e.key === 'k') {
      e.preventDefault();
      focusOption(index - 1);
    }
  }
</script>

<div
  class="choice-widget"
  role="radiogroup"
  aria-label={prompt?.prompt ?? 'Choice'}
  aria-required={prompt?.required ? 'true' : 'false'}
  aria-invalid={!!validationError}
  bind:this={groupEl}
>
  {#each options as option, i}
    <button
      type="button"
      role="radio"
      aria-checked={selected === i}
      class="option-btn"
      class:selected={selected === i}
      data-testid={`choice-option-${i + 1}`}
      tabindex={selected === i || (selected === -1 && i === 0) ? 0 : -1}
      {disabled}
      on:click={() => choose(i)}
      on:keydown={(e) => onKeydown(e, i)}
    >
      <span class="option-num">{i + 1}.</span>
      {#if selected === i}
        <span class="option-chevron" aria-hidden="true">›</span>
      {/if}
      <span class="option-label">{option}</span>
    </button>
  {/each}
</div>

<style>
  .choice-widget {
    display: flex;
    flex-direction: column;
    gap: var(--sp-sm);
  }

  .option-btn {
    display: flex;
    align-items: center;
    gap: var(--sp-md);
    padding: var(--sp-sm) var(--sp-md);
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    cursor: pointer;
    font-family: var(--font-ui);
    font-size: var(--text-body);
    font-weight: 500;
    color: var(--text-primary);
    text-align: left;
    transition:
      background var(--duration-short) var(--ease-enter),
      border-color var(--duration-short) var(--ease-enter),
      transform var(--duration-micro) var(--ease-enter);
  }

  .option-btn:hover:not(:disabled) {
    background: var(--bg-active);
    border-color: var(--border-emphasis);
  }

  .option-btn:active:not(:disabled) { transform: scale(0.98); }

  .option-btn:focus-visible {
    outline: 2px solid var(--accent-green);
    outline-offset: -1px;
  }

  .option-btn.selected {
    background: color-mix(in srgb, var(--accent-green) 15%, transparent);
    border-color: var(--accent-green);
  }

  .option-btn:disabled { opacity: 0.6; cursor: not-allowed; }

  .option-num {
    font-family: var(--font-mono);
    font-size: var(--text-label);
    color: var(--text-dim);
    flex-shrink: 0;
    width: 20px;
  }

  .option-chevron {
    color: var(--accent-green);
    font-weight: 700;
    font-size: var(--text-data);
    line-height: 1;
    flex-shrink: 0;
  }

  .option-label { flex: 1; min-width: 0; }
</style>
