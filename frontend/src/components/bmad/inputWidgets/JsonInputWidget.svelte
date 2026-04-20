<script>
  import { createEventDispatcher, onMount, tick } from 'svelte';
  import { Send } from 'lucide-svelte';

  /** @type {import('../../../stores/interactiveInput').PendingPrompt} */
  export let prompt;
  export let disabled = false;
  export let validationError = '';

  const dispatch = createEventDispatcher();

  let value = '';
  let parseError = '';
  /** @type {HTMLTextAreaElement | null} */
  let textareaEl = null;

  $: isParseValid = computeParseStatus(value);
  $: canSubmit = !disabled && value.trim().length > 0 && isParseValid;

  /** @param {string} raw */
  function computeParseStatus(raw) {
    if (!raw.trim()) {
      parseError = '';
      return false;
    }
    try {
      JSON.parse(raw);
      parseError = '';
      return true;
    } catch (e) {
      parseError = e instanceof Error ? e.message : String(e);
      return false;
    }
  }

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

<div class="json-widget">
  <label class="field">
    <textarea
      bind:this={textareaEl}
      bind:value
      class="json-textarea"
      class:invalid={!!validationError || (!!value && !isParseValid)}
      data-testid="json-textarea"
      placeholder="{`{ "key": "value" }`}"
      aria-label={prompt?.prompt ?? 'JSON input'}
      aria-required={prompt?.required ? 'true' : 'false'}
      aria-invalid={!!validationError || (!!value && !isParseValid)}
      spellcheck="false"
      {disabled}
      on:keydown={onKeydown}
    ></textarea>
    <div class="parse-status" data-testid="json-parse-status">
      {#if value.trim() === ''}
        <span class="muted">Enter JSON payload</span>
      {:else if isParseValid}
        <span class="ok">&#10003; valid JSON</span>
      {:else}
        <span class="err">{parseError}</span>
      {/if}
    </div>
  </label>

  <div class="widget-actions">
    <span class="keyboard-hint">Cmd+Enter to send</span>
    <button
      type="button"
      class="btn-submit"
      data-testid="json-submit"
      disabled={!canSubmit}
      on:click={submit}
    >
      <Send size={13} aria-hidden="true" />
      Submit
    </button>
  </div>
</div>

<style>
  .json-widget {
    display: flex;
    flex-direction: column;
    gap: var(--sp-sm);
  }

  .field {
    display: flex;
    flex-direction: column;
    gap: var(--sp-2xs);
  }

  .json-textarea {
    width: 100%;
    min-height: 320px;
    resize: vertical;
    box-sizing: border-box;
    background: var(--bg-deepest);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    padding: var(--sp-sm) var(--sp-md);
    font-family: var(--font-code);
    font-size: var(--text-body);
    line-height: 1.5;
    tab-size: 2;
    color: var(--text-primary);
    white-space: pre;
    overflow: auto;
    outline: none;
    transition: border-color var(--duration-short) var(--ease-enter);
  }

  .json-textarea::placeholder { color: var(--text-muted); }
  .json-textarea:focus { border-color: var(--accent-green); }
  .json-textarea.invalid { border-color: var(--accent-red); }
  .json-textarea:disabled { opacity: 0.6; cursor: not-allowed; }

  .parse-status {
    font-family: var(--font-mono);
    font-size: var(--text-label);
  }

  .parse-status .muted { color: var(--text-muted); }
  .parse-status .ok { color: var(--accent-green); }
  .parse-status .err { color: var(--accent-red); }

  .widget-actions {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: var(--sp-md);
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
