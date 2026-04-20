<script>
  import { createEventDispatcher, onMount, tick } from 'svelte';
  import { Send } from 'lucide-svelte';

  /** @type {import('../../../stores/interactiveInput').PendingPrompt} */
  export let prompt;
  export let disabled = false;
  export let validationError = '';

  const dispatch = createEventDispatcher();

  let value = '';
  /** @type {HTMLTextAreaElement | null} */
  let textareaEl = null;

  $: maxLength = prompt?.maxLength && prompt.maxLength > 0 ? prompt.maxLength : 5000;
  $: counterColor =
    value.length >= maxLength
      ? 'var(--accent-red)'
      : value.length >= Math.floor(maxLength * 0.9)
      ? 'var(--accent-amber)'
      : 'var(--text-muted)';
  $: canSubmit = !disabled && value.trim().length > 0;

  onMount(async () => {
    await tick();
    textareaEl?.focus();
  });

  function submit() {
    if (!canSubmit) return;
    dispatch('submit', { value });
  }

  /** @param {KeyboardEvent} e */
  function onKeydown(e) {
    if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
      e.preventDefault();
      submit();
    }
  }
</script>

<div class="free-text-widget">
  <label class="field">
    <textarea
      bind:this={textareaEl}
      bind:value
      class="response-textarea"
      class:invalid={!!validationError}
      data-testid="free-text-textarea"
      placeholder="Type your response..."
      aria-label={prompt?.prompt ?? 'Response'}
      aria-required={prompt?.required ? 'true' : 'false'}
      aria-invalid={!!validationError}
      maxlength={maxLength}
      {disabled}
      on:keydown={onKeydown}
    ></textarea>
    <div class="field-footer">
      <span class="keyboard-hint">Cmd+Enter to send</span>
      {#if prompt?.maxLength}
        <span class="char-count" style:color={counterColor}>
          {value.length} / {maxLength}
        </span>
      {/if}
    </div>
  </label>

  <div class="widget-actions">
    <button
      type="button"
      class="btn-submit"
      data-testid="free-text-submit"
      disabled={!canSubmit}
      on:click={submit}
    >
      <Send size={13} aria-hidden="true" />
      Send
    </button>
  </div>
</div>

<style>
  .free-text-widget {
    display: flex;
    flex-direction: column;
    gap: var(--sp-md);
  }

  .field {
    display: flex;
    flex-direction: column;
    gap: var(--sp-2xs);
  }

  .response-textarea {
    width: 100%;
    min-height: 72px;
    max-height: 200px;
    resize: vertical;
    box-sizing: border-box;
    background: var(--bg-deepest);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    padding: var(--sp-sm) var(--sp-md);
    font-family: var(--font-ui);
    font-size: var(--text-body);
    color: var(--text-primary);
    line-height: 1.5;
    outline: none;
    transition: border-color var(--duration-short) var(--ease-enter);
  }

  .response-textarea::placeholder { color: var(--text-muted); }
  .response-textarea:focus { border-color: var(--accent-green); }
  .response-textarea.invalid { border-color: var(--accent-red); }
  .response-textarea:disabled { opacity: 0.6; cursor: not-allowed; }

  .field-footer {
    display: flex;
    justify-content: space-between;
    gap: var(--sp-sm);
  }

  .keyboard-hint,
  .char-count {
    font-family: var(--font-mono);
    font-size: var(--text-label);
    color: var(--text-muted);
  }

  .char-count { font-variant-numeric: tabular-nums; }

  .widget-actions {
    display: flex;
    justify-content: flex-end;
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
      background var(--duration-short) var(--ease-enter),
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
