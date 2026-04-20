<script>
  import { createEventDispatcher, onMount, onDestroy, tick } from 'svelte';
  import { fade, fly } from 'svelte/transition';
  import { cubicOut, cubicIn } from 'svelte/easing';
  import { X, MessageCircleQuestion } from 'lucide-svelte';
  import { RespondToInput } from '../../../wailsjs/go/main/App.js';
  import {
    interactiveInput,
    validationKey,
  } from '../../stores/interactiveInput';
  import FreeTextWidget from './inputWidgets/FreeTextWidget.svelte';
  import ChoiceWidget from './inputWidgets/ChoiceWidget.svelte';
  import MultiChoiceWidget from './inputWidgets/MultiChoiceWidget.svelte';
  import ApprovalWidget from './inputWidgets/ApprovalWidget.svelte';
  import FileInputWidget from './inputWidgets/FileInputWidget.svelte';
  import JsonInputWidget from './inputWidgets/JsonInputWidget.svelte';

  /** @type {import('../../stores/interactiveInput').PendingPrompt | null} */
  export let prompt = null;
  export let execId = '';
  export let repoName = '';

  const dispatch = createEventDispatcher();

  /** @type {HTMLDivElement | null} */
  let cardEl = null;
  let shaking = false;
  let sending = false;
  let lastError = '';

  $: key = prompt ? validationKey(prompt.nodeId, prompt.inputId) : '';
  $: storeError = key ? $interactiveInput.validationError[key] || '' : '';
  $: effectiveError = lastError || storeError;

  // When a new validation error arrives from the store, trigger the shake.
  let lastShakeFor = '';
  $: if (storeError && key && storeError !== lastShakeFor) {
    lastShakeFor = storeError;
    triggerShake();
  }

  function triggerShake() {
    shaking = true;
    setTimeout(() => { shaking = false; }, 250);
  }

  onMount(async () => {
    await tick();
  });

  onDestroy(() => {
    shaking = false;
  });

  /** @param {{ detail: { value: string } }} e */
  async function onWidgetSubmit(e) {
    if (!prompt || sending) return;
    sending = true;
    lastError = '';
    try {
      await RespondToInput(execId || prompt.execId || '', prompt.nodeId, prompt.inputId, e.detail.value);
      dispatch('responded', { nodeId: prompt.nodeId, inputId: prompt.inputId });
      // Backend will emit input_resolved which closes via the store.
    } catch (err) {
      lastError = err instanceof Error ? err.message : String(err);
      triggerShake();
    } finally {
      sending = false;
    }
  }

  function close() {
    dispatch('close');
  }

  /** @param {KeyboardEvent} e */
  function onKeydown(e) {
    if (e.key === 'Escape') {
      e.preventDefault();
      e.stopPropagation();
      close();
    }
  }

  function onOverlayClick() {
    close();
  }

  $: shape = prompt?.shape || '';
  $: isWideShape = shape === 'file' || shape === 'json';
  $: hasFooter = shape !== 'approval';
  $: roundPillText = (() => {
    if (!prompt || !prompt.round || prompt.round <= 0) return '';
    if (prompt.maxRounds && prompt.maxRounds > 0) return `${prompt.round} / ${prompt.maxRounds}`;
    return `Round ${prompt.round}`;
  })();
</script>

<svelte:window on:keydown={onKeydown} />

{#if prompt}
  <!-- svelte-ignore a11y-click-events-have-key-events -->
  <div
    class="modal-overlay"
    role="presentation"
    on:click|self={onOverlayClick}
    transition:fade={{ duration: 150 }}
  >
    <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
    <div
      class="modal-card"
      class:wide={isWideShape}
      class:input-error-shake={shaking}
      role="dialog"
      aria-modal="true"
      aria-labelledby="input-modal-title"
      data-testid="input-response-modal"
      data-shape={shape}
      bind:this={cardEl}
      in:fly={{ y: 20, duration: 200, easing: cubicOut }}
      out:fly={{ y: 20, duration: 150, easing: cubicIn }}
      on:click|stopPropagation
    >
      <header class="modal-header">
        <span class="header-icon" aria-hidden="true">
          <MessageCircleQuestion size={16} />
        </span>
        <span id="input-modal-title" class="title">
          Input from <span class="repo-name">{repoName || prompt.repoName || 'workflow'}</span>
        </span>
        {#if roundPillText}
          <span class="round-pill" data-testid="round-pill">{roundPillText}</span>
        {/if}
        <button
          type="button"
          class="close-btn"
          aria-label="Close"
          data-testid="input-modal-close"
          on:click={close}
        >
          <X size={16} aria-hidden="true" />
        </button>
      </header>

      <div class="modal-body" class:editor={shape === 'json'}>
        <blockquote class="prompt-block" data-testid="input-modal-prompt">
          {prompt.prompt}
        </blockquote>

        {#if shape === 'free'}
          <FreeTextWidget
            {prompt}
            disabled={sending}
            validationError={effectiveError}
            on:submit={onWidgetSubmit}
          />
        {:else if shape === 'choice'}
          <ChoiceWidget
            {prompt}
            disabled={sending}
            validationError={effectiveError}
            on:submit={onWidgetSubmit}
          />
        {:else if shape === 'multi'}
          <MultiChoiceWidget
            {prompt}
            disabled={sending}
            validationError={effectiveError}
            on:submit={onWidgetSubmit}
          />
        {:else if shape === 'approval'}
          <ApprovalWidget
            {prompt}
            disabled={sending}
            on:submit={onWidgetSubmit}
          />
        {:else if shape === 'file'}
          <FileInputWidget
            {prompt}
            disabled={sending}
            validationError={effectiveError}
            on:submit={onWidgetSubmit}
          />
        {:else if shape === 'json'}
          <JsonInputWidget
            {prompt}
            disabled={sending}
            validationError={effectiveError}
            on:submit={onWidgetSubmit}
          />
        {/if}

        {#if prompt.helpText}
          <div class="help-text" data-testid="input-modal-help">{prompt.helpText}</div>
        {/if}
      </div>

      {#if effectiveError}
        <div class="error-bar" data-testid="input-modal-error">
          <span class="error-text">{effectiveError}</span>
        </div>
      {/if}

      {#if hasFooter}
        <footer class="modal-footer">
          <button type="button" class="btn-cancel" data-testid="input-modal-cancel" on:click={close}>
            Cancel
          </button>
        </footer>
      {/if}
    </div>
  </div>
{/if}

<style>
  .modal-overlay {
    position: fixed;
    inset: 0;
    background: color-mix(in srgb, var(--bg-deepest) 70%, transparent);
    z-index: var(--z-modal);
    display: flex;
    align-items: center;
    justify-content: center;
    padding: var(--sp-2xl);
  }

  .modal-card {
    width: 100%;
    max-width: 560px;
    min-width: var(--modal-width-min);
    max-height: 80vh;
    background: var(--bg-surface);
    border: 1px solid var(--border-emphasis);
    border-radius: var(--radius-lg);
    display: flex;
    flex-direction: column;
    overflow: hidden;
    box-shadow: 0 16px 48px color-mix(in srgb, var(--bg-deepest) 80%, transparent);
  }

  .modal-card.wide { max-width: var(--modal-width-wide); }

  .modal-header {
    display: flex;
    align-items: center;
    gap: var(--sp-sm);
    padding: var(--sp-md) var(--sp-lg);
    border-bottom: 1px solid var(--border-subtle);
    flex-shrink: 0;
  }

  .header-icon {
    display: inline-flex;
    color: var(--accent-amber);
    flex-shrink: 0;
  }

  .title {
    flex: 1;
    font-family: var(--font-ui);
    font-size: var(--text-body);
    font-weight: 600;
    color: var(--text-primary);
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

  .round-pill {
    background: color-mix(in srgb, var(--accent-amber) 15%, transparent);
    color: var(--accent-amber);
    padding: 0 var(--sp-xs);
    border-radius: var(--radius-sm);
    font-family: var(--font-mono);
    font-size: var(--text-label);
    font-weight: 600;
    font-variant-numeric: tabular-nums;
    letter-spacing: 0.04em;
    flex-shrink: 0;
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

  .modal-body.editor { min-height: var(--modal-height-editor); }

  .prompt-block {
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

  .help-text {
    font-family: var(--font-ui);
    font-size: var(--text-label);
    color: var(--text-dim);
    font-style: italic;
  }

  .error-bar {
    padding: var(--sp-sm) var(--sp-lg);
    background: color-mix(in srgb, var(--accent-red) 10%, transparent);
    border-top: 1px solid color-mix(in srgb, var(--accent-red) 30%, transparent);
    flex-shrink: 0;
  }

  .error-text {
    font-family: var(--font-ui);
    font-size: var(--text-body);
    font-weight: 500;
    color: var(--accent-red);
  }

  .modal-footer {
    display: flex;
    justify-content: flex-end;
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

  @media (prefers-reduced-motion: no-preference) {
    .input-error-shake {
      animation: input-error-shake var(--duration-shake) var(--ease-move) 1;
    }
  }

  @keyframes input-error-shake {
    0% { transform: translateX(0); }
    20% { transform: translateX(-6px); }
    40% { transform: translateX(6px); }
    60% { transform: translateX(-4px); }
    80% { transform: translateX(4px); }
    100% { transform: translateX(0); }
  }
</style>
