<script>
  import { createEventDispatcher, onMount, tick } from 'svelte';
  import { Send } from 'lucide-svelte';

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

  /**
   * Options arrive in two shapes depending on the source:
   *   - Layer-1 path (`prompt.options`) → array of strings
   *   - AST decision_group path → array of `{value, label}` objects
   * Normalise both ends so the widget renders the human label and submits
   * the canonical value (Anthropic's "value" field, falling back to label).
   *
   * @param {string | {value?: string, label?: string}} opt
   */
  function optionLabel(opt) {
    if (typeof opt === 'string') return opt;
    if (opt && typeof opt === 'object') return opt.label ?? opt.value ?? '';
    return String(opt ?? '');
  }
  /** @param {string | {value?: string, label?: string}} opt */
  function optionValue(opt) {
    if (typeof opt === 'string') return opt;
    if (opt && typeof opt === 'object') return opt.value ?? opt.label ?? '';
    return String(opt ?? '');
  }

  function submit() {
    if (disabled || selected < 0) return;
    dispatch('submit', { value: optionValue(options[selected]) });
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
      <span class="option-label">{optionLabel(option)}</span>
    </button>
  {/each}
</div>

<div class="widget-actions">
  <span class="keyboard-hint">Enter to send</span>
  <button
    type="button"
    class="btn-submit"
    data-testid="choice-submit"
    disabled={disabled || selected < 0}
    on:click={submit}
  >
    <Send size={13} aria-hidden="true" />
    Send
  </button>
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

  .widget-actions {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: var(--sp-md);
    margin-top: var(--sp-sm);
  }

  .keyboard-hint {
    font-family: var(--font-mono);
    font-size: var(--text-label);
    color: var(--text-muted);
  }

  .btn-submit {
    display: inline-flex;
    align-items: center;
    gap: var(--sp-xs);
    background: var(--accent-green);
    border: none;
    border-radius: var(--radius-md);
    padding: var(--sp-xs) var(--sp-xl);
    font-family: var(--font-ui);
    font-size: var(--text-body);
    font-weight: 600;
    color: var(--bg-deepest);
    cursor: pointer;
    transition:
      filter var(--duration-short) var(--ease-enter),
      transform var(--duration-micro) var(--ease-enter);
  }

  .btn-submit:hover:not(:disabled) { filter: brightness(1.1); }
  .btn-submit:active:not(:disabled) { transform: scale(0.97); }

  .btn-submit:disabled {
    background: var(--accent-green-dim);
    color: var(--text-muted);
    cursor: not-allowed;
  }

  .btn-submit:focus-visible {
    outline: 1px solid var(--accent-green);
    outline-offset: 2px;
  }
</style>
