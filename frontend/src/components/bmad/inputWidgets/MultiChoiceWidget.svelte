<script>
  import { createEventDispatcher, onMount, tick } from 'svelte';
  import { Check, Send } from 'lucide-svelte';

  /** @type {import('../../../stores/interactiveInput').PendingPrompt} */
  export let prompt;
  export let disabled = false;
  export let validationError = '';

  const dispatch = createEventDispatcher();

  /** @type {Set<number>} */
  let selected = new Set();
  /** @type {HTMLDivElement | null} */
  let groupEl = null;

  $: options = prompt?.options ?? [];
  $: canSubmit = !disabled && selected.size > 0;

  onMount(async () => {
    await tick();
    const first = groupEl?.querySelector('input[type="checkbox"]');
    if (first instanceof HTMLElement) first.focus();
  });

  /** @param {number} index */
  function toggle(index) {
    if (disabled) return;
    const next = new Set(selected);
    if (next.has(index)) next.delete(index);
    else next.add(index);
    selected = next;
  }

  function submit() {
    if (!canSubmit) return;
    const values = [...selected].sort((a, b) => a - b).map((i) => options[i]);
    dispatch('submit', { value: values.join(',') });
  }
</script>

<!-- svelte-ignore a11y-role-supports-aria-props -->
<div
  class="multi-widget"
  role="group"
  aria-label={prompt?.prompt ?? 'Multi-choice'}
  aria-required={prompt?.required ? 'true' : 'false'}
  aria-invalid={!!validationError}
  bind:this={groupEl}
>
  {#each options as option, i}
    <label class="checkbox-row">
      <input
        type="checkbox"
        class="native-checkbox"
        data-testid={`multi-checkbox-${i + 1}`}
        checked={selected.has(i)}
        {disabled}
        on:change={() => toggle(i)}
      />
      <span class="custom-box" aria-hidden="true">
        {#if selected.has(i)}
          <Check size={10} />
        {/if}
      </span>
      <span class="option-label">{option}</span>
    </label>
  {/each}

  <div class="widget-actions">
    <button
      type="button"
      class="btn-submit"
      data-testid="multi-submit"
      disabled={!canSubmit}
      on:click={submit}
    >
      <Send size={13} aria-hidden="true" />
      Submit
    </button>
  </div>
</div>

<style>
  .multi-widget {
    display: flex;
    flex-direction: column;
    gap: var(--sp-sm);
  }

  .checkbox-row {
    display: flex;
    align-items: center;
    gap: var(--sp-sm);
    padding: var(--sp-2xs) var(--sp-xs);
    border-radius: var(--radius-sm);
    cursor: pointer;
    font-family: var(--font-ui);
    font-size: var(--text-body);
    color: var(--text-primary);
  }

  .checkbox-row:hover { background: var(--bg-elevated); }

  .native-checkbox {
    position: absolute;
    opacity: 0;
    width: 0;
    height: 0;
    pointer-events: none;
  }

  .custom-box {
    width: 14px;
    height: 14px;
    border: 1px solid var(--border-emphasis);
    border-radius: var(--radius-sm);
    background: var(--bg-deepest);
    display: inline-flex;
    align-items: center;
    justify-content: center;
    color: var(--accent-green);
    flex-shrink: 0;
    transition:
      background var(--duration-short) var(--ease-enter),
      border-color var(--duration-short) var(--ease-enter);
  }

  .native-checkbox:checked + .custom-box {
    background: color-mix(in srgb, var(--accent-green) 15%, transparent);
    border-color: var(--accent-green);
  }

  .native-checkbox:focus-visible + .custom-box {
    outline: 2px solid var(--accent-green);
    outline-offset: 1px;
  }

  .option-label { flex: 1; min-width: 0; }

  .widget-actions {
    display: flex;
    justify-content: flex-end;
    margin-top: var(--sp-xs);
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
