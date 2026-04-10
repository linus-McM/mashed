<script>
  import { createEventDispatcher, onMount, tick } from 'svelte';
  import { fade, fly, slide } from 'svelte/transition';
  import { cubicOut, cubicIn } from 'svelte/easing';
  import { X, MessageCircleQuestion, Send } from 'lucide-svelte';
  import { RespondToQuestion } from '../../../wailsjs/go/main/App.js';

  /** @type {import('./questionSnackbarUtils').QuestionEventLike} */
  export let question;

  const dispatch = createEventDispatcher();

  let answer = '';
  let sending = false;
  let error = '';

  /** @type {HTMLTextAreaElement | null} */
  let textareaEl = null;
  /** @type {HTMLElement | null} */
  let modalEl = null;

  $: isMenu = Array.isArray(question?.options) && question.options.length > 0;
  $: canSubmit = !sending && (isMenu || answer.trim().length > 0);

  onMount(async () => {
    await tick();
    requestAnimationFrame(() => {
      if (isMenu) {
        const firstOption = modalEl?.querySelector('[data-testid="option-button-1"]');
        if (firstOption instanceof HTMLElement) firstOption.focus();
      } else {
        textareaEl?.focus();
      }
    });
  });

  function repoLabel() {
    const name = question?.repoName;
    return name && name.length > 0 ? name : 'Unknown Repo';
  }

  /** @param {string} value */
  async function submit(value) {
    if (sending) return;
    sending = true;
    error = '';
    try {
      await RespondToQuestion(question.execId ?? '', question.nodeId, value);
      dispatch('responded');
    } catch (e) {
      const msg = e instanceof Error ? e.message : String(e);
      error = msg;
      sending = false;
      dispatch('error', { message: msg });
    }
  }

  function handleSend() {
    if (!canSubmit) return;
    submit(answer);
  }

  /** @param {number} index */
  function handleOption(index) {
    if (sending) return;
    submit(String(index + 1));
  }

  function handleCancel() {
    dispatch('close');
  }

  function handleOverlayClick() {
    dispatch('close');
  }

  /** @param {KeyboardEvent} e */
  function handleKeydown(e) {
    if (e.key === 'Escape') {
      // stopPropagation prevents App.svelte's global Escape handler from
      // also triggering goBack() and unmounting the WorkflowBuilder view.
      e.preventDefault();
      e.stopPropagation();
      dispatch('close');
      return;
    }
    if (!isMenu && (e.ctrlKey || e.metaKey) && e.key === 'Enter') {
      e.preventDefault();
      if (canSubmit) handleSend();
      return;
    }
    // Menu mode: arrow keys (and j/k) move focus between option buttons.
    if (isMenu && (e.key === 'ArrowDown' || e.key === 'ArrowUp' || e.key === 'j' || e.key === 'k')) {
      const buttons = modalEl?.querySelectorAll('[data-testid^="option-button-"]');
      if (!buttons || buttons.length === 0) return;
      const current = document.activeElement;
      const idx = Array.from(buttons).indexOf(/** @type {Element} */ (current));
      const forward = e.key === 'ArrowDown' || e.key === 'j';
      const next = idx < 0
        ? (forward ? 0 : buttons.length - 1)
        : (idx + (forward ? 1 : -1) + buttons.length) % buttons.length;
      e.preventDefault();
      /** @type {HTMLElement} */ (buttons[next]).focus();
    }
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<!-- svelte-ignore a11y-click-events-have-key-events -->
<div
  class="modal-overlay"
  role="presentation"
  on:click|self={handleOverlayClick}
  transition:fade={{ duration: 150 }}
>
  <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
  <div
    class="modal-card"
    role="dialog"
    aria-modal="true"
    aria-labelledby="question-modal-title"
    data-testid="question-modal"
    bind:this={modalEl}
    in:fly={{ y: 20, duration: 200, easing: cubicOut }}
    out:fly={{ y: 20, duration: 150, easing: cubicIn }}
    on:click|stopPropagation
  >
    <header class="modal-header">
      <MessageCircleQuestion size={16} class="header-icon" aria-hidden="true" />
      <span id="question-modal-title" class="title">
        Question from <span class="repo-name">{repoLabel()}</span>
      </span>
      <button
        type="button"
        class="close-btn"
        aria-label="Close"
        on:click={handleCancel}
      >
        <X size={16} aria-hidden="true" />
      </button>
    </header>

    <div class="modal-body">
      <blockquote class="question-block" data-testid="question-text">
        {question?.question ?? ''}
      </blockquote>

      {#if isMenu}
        <div class="option-list">
          {#each question.options ?? [] as option, i}
            <button
              type="button"
              class="option-btn"
              data-testid={`option-button-${i + 1}`}
              disabled={sending}
              on:click={() => handleOption(i)}
            >
              <span class="option-num">{i + 1}.</span>
              <span class="option-label">{option}</span>
            </button>
          {/each}
        </div>
      {:else}
        <label class="freeform-field">
          <textarea
            bind:this={textareaEl}
            bind:value={answer}
            class="response-textarea"
            data-testid="answer-textarea"
            placeholder="Type your response..."
            rows="3"
            disabled={sending}
          ></textarea>
          <span class="keyboard-hint">Ctrl+Enter to send</span>
        </label>
      {/if}
    </div>

    {#if error}
      <div class="error-bar" data-testid="error-message" transition:slide={{ duration: 150 }}>
        <span class="error-text">{error}</span>
        <button
          type="button"
          class="error-dismiss"
          data-testid="error-dismiss-button"
          aria-label="Dismiss error and close"
          on:click={handleCancel}
        >
          Dismiss
        </button>
      </div>
    {/if}

    <footer class="modal-footer">
      <button
        type="button"
        class="btn-cancel"
        data-testid="cancel-button"
        on:click={handleCancel}
      >
        Cancel
      </button>
      {#if !isMenu}
        <button
          type="button"
          class="btn-send"
          data-testid="send-button"
          disabled={!canSubmit}
          on:click={handleSend}
        >
          <Send size={13} aria-hidden="true" />
          {sending ? 'Sending...' : 'Send'}
        </button>
      {/if}
    </footer>
  </div>
</div>

<style>
  .modal-overlay {
    position: fixed;
    inset: 0;
    background: color-mix(in srgb, var(--bg-deepest) 70%, transparent);
    z-index: 400;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: var(--sp-2xl);
  }

  .modal-card {
    max-width: 560px;
    width: 100%;
    max-height: 80vh;
    background: var(--bg-surface);
    border: 1px solid var(--border-emphasis);
    border-radius: var(--radius-lg);
    display: flex;
    flex-direction: column;
    overflow: hidden;
    box-shadow: 0 16px 48px color-mix(in srgb, var(--bg-deepest) 80%, transparent);
  }

  .modal-header {
    display: flex;
    align-items: center;
    gap: var(--sp-sm);
    padding: var(--sp-md) var(--sp-lg);
    border-bottom: 1px solid var(--border-subtle);
    flex-shrink: 0;
  }

  :global(.modal-header .header-icon) {
    color: var(--accent-amber);
    flex-shrink: 0;
  }

  .modal-header .title {
    flex: 1;
    font-family: var(--font-ui);
    font-size: var(--text-body);
    font-weight: 600;
    color: var(--text-primary);
    line-height: 1.4;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .repo-name {
    font-family: var(--font-mono);
    font-weight: 500;
    color: var(--accent-green);
  }

  .close-btn {
    background: none;
    border: none;
    color: var(--text-muted);
    cursor: pointer;
    padding: var(--sp-2xs);
    border-radius: var(--radius-sm);
    display: inline-flex;
    align-items: center;
    justify-content: center;
    transition: color var(--duration-short) var(--ease-enter);
  }

  .close-btn:hover { color: var(--text-dim); }
  .close-btn:focus-visible {
    outline: 1px solid var(--accent-green);
    outline-offset: 2px;
  }

  .modal-body {
    padding: var(--sp-lg);
    display: flex;
    flex-direction: column;
    gap: var(--sp-lg);
    overflow-y: auto;
    flex: 1;
  }

  .question-block {
    margin: 0;
    background: var(--bg-deepest);
    border-left: 3px solid var(--accent-green);
    border-radius: 0 var(--radius-md) var(--radius-md) 0;
    padding: var(--sp-md) var(--sp-lg);
    font-family: var(--font-ui);
    font-size: var(--text-data);
    color: var(--text-primary);
    line-height: 1.6;
    white-space: pre-wrap;
    word-break: break-word;
  }

  /* Freeform mode */
  .freeform-field {
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
  .response-textarea:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .keyboard-hint {
    font-family: var(--font-mono);
    font-size: var(--text-label);
    color: var(--text-muted);
    text-align: right;
  }

  /* Menu mode */
  .option-list {
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

  .option-btn:active:not(:disabled) {
    transform: scale(0.98);
  }

  .option-btn:focus-visible {
    outline: 1px solid var(--accent-green);
    outline-offset: -1px;
  }

  .option-btn:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .option-num {
    font-family: var(--font-mono);
    font-size: var(--text-label);
    color: var(--text-dim);
    flex-shrink: 0;
    width: 20px;
  }

  .option-label {
    flex: 1;
    min-width: 0;
  }

  /* Error bar */
  .error-bar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: var(--sp-sm);
    background: color-mix(in srgb, var(--accent-red) 10%, transparent);
    border-top: 1px solid color-mix(in srgb, var(--accent-red) 30%, transparent);
    padding: var(--sp-sm) var(--sp-lg);
    flex-shrink: 0;
  }

  .error-text {
    flex: 1;
    min-width: 0;
    font-family: var(--font-ui);
    font-size: var(--text-body);
    font-weight: 500;
    color: var(--accent-red);
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .error-dismiss {
    flex-shrink: 0;
    background: transparent;
    border: 1px solid color-mix(in srgb, var(--accent-red) 40%, transparent);
    border-radius: var(--radius-sm);
    padding: var(--sp-2xs) var(--sp-sm);
    font-family: var(--font-ui);
    font-size: var(--text-label);
    font-weight: 500;
    color: var(--accent-red);
    cursor: pointer;
    transition: background var(--duration-short) var(--ease-enter);
  }

  .error-dismiss:hover {
    background: color-mix(in srgb, var(--accent-red) 15%, transparent);
  }

  .error-dismiss:focus-visible {
    outline: 1px solid var(--accent-red);
    outline-offset: 1px;
  }

  /* Footer */
  .modal-footer {
    display: flex;
    justify-content: flex-end;
    align-items: center;
    gap: var(--sp-sm);
    padding: var(--sp-md) var(--sp-lg);
    border-top: 1px solid var(--border-subtle);
    flex-shrink: 0;
  }

  .btn-cancel {
    background: transparent;
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    padding: var(--sp-xs) var(--sp-lg);
    font-family: var(--font-ui);
    font-size: var(--text-body);
    font-weight: 500;
    color: var(--text-dim);
    cursor: pointer;
    transition:
      background var(--duration-short) var(--ease-enter),
      color var(--duration-short) var(--ease-enter);
  }

  .btn-cancel:hover {
    background: var(--bg-elevated);
    color: var(--text-primary);
  }

  .btn-cancel:focus-visible {
    outline: 1px solid var(--accent-green);
    outline-offset: 2px;
  }

  .btn-send {
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
      color var(--duration-short) var(--ease-enter),
      transform var(--duration-micro) var(--ease-enter);
  }

  .btn-send:hover:not(:disabled) { filter: brightness(1.1); }
  .btn-send:active:not(:disabled) { transform: scale(0.97); }

  .btn-send:disabled {
    background: var(--accent-green-dim);
    color: var(--text-muted);
    cursor: not-allowed;
  }

  .btn-send:focus-visible {
    outline: 1px solid var(--accent-green);
    outline-offset: 2px;
  }
</style>
