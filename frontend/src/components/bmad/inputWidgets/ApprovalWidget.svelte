<script>
  import { createEventDispatcher, onMount, tick } from 'svelte';

  /** @type {import('../../../stores/interactiveInput').PendingPrompt} */
  // svelte-ignore unused-export-let
  export let prompt;
  export let disabled = false;

  const dispatch = createEventDispatcher();

  /** @type {HTMLButtonElement | null} */
  let yesBtn = null;

  onMount(async () => {
    await tick();
    yesBtn?.focus();
  });

  /** @param {'yes' | 'no'} value */
  function pick(value) {
    if (disabled) return;
    dispatch('submit', { value });
  }
</script>

<div class="approval-widget">
  <button
    bind:this={yesBtn}
    type="button"
    class="btn-yes"
    data-testid="approval-yes"
    {disabled}
    on:click={() => pick('yes')}
  >
    Yes
  </button>
  <button
    type="button"
    class="btn-no"
    data-testid="approval-no"
    {disabled}
    on:click={() => pick('no')}
  >
    No
  </button>
</div>

<style>
  .approval-widget {
    display: flex;
    gap: var(--sp-md);
    justify-content: flex-end;
  }

  .btn-yes,
  .btn-no {
    min-width: var(--btn-min-wide);
    padding: var(--sp-xs) var(--sp-xl);
    border-radius: var(--radius-md);
    font-family: var(--font-ui);
    font-size: var(--text-body);
    font-weight: 600;
    cursor: pointer;
    transition:
      background var(--duration-short) var(--ease-enter),
      color var(--duration-short) var(--ease-enter),
      transform var(--duration-micro) var(--ease-enter);
  }

  .btn-yes {
    background: var(--accent-green);
    border: none;
    color: var(--bg-deepest);
  }
  .btn-yes:hover:not(:disabled) { filter: brightness(1.1); }
  .btn-yes:active:not(:disabled) { transform: scale(0.97); }
  .btn-yes:focus-visible {
    outline: 1px solid var(--accent-green);
    outline-offset: 2px;
  }

  .btn-no {
    background: transparent;
    border: 1px solid var(--border-emphasis);
    color: var(--text-dim);
    font-weight: 500;
  }
  .btn-no:hover:not(:disabled) {
    color: var(--text-primary);
    background: var(--bg-elevated);
  }
  .btn-no:active:not(:disabled) { transform: scale(0.97); }
  .btn-no:focus-visible {
    outline: 1px solid var(--text-dim);
    outline-offset: 2px;
  }

  .btn-yes:disabled,
  .btn-no:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }
</style>
