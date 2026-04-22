<script>
  import { createEventDispatcher, onMount, onDestroy, tick } from 'svelte';
  import { get } from 'svelte/store';
  import { fade, fly } from 'svelte/transition';
  import { cubicOut, cubicIn } from 'svelte/easing';
  import { X, MessageCircleQuestion, Send } from 'lucide-svelte';
  import { RespondToInput, GetInteractiveTranscript } from '../../../wailsjs/go/main/App.js';
  import {
    interactiveInput,
    validationKey,
    pendingPrompt,
    parseStructuredAst,
    pushToast,
  } from '../../stores/interactiveInput';
  import {
    uiAdapterUntrustedExpanded,
    ollamaReachable,
    uiAdapterEnabled,
    ollamaEnabled,
  } from '../../lib/stores/uiAdapterSettings';
  import FreeTextWidget from './inputWidgets/FreeTextWidget.svelte';
  import ChoiceWidget from './inputWidgets/ChoiceWidget.svelte';
  import MultiChoiceWidget from './inputWidgets/MultiChoiceWidget.svelte';
  import ApprovalWidget from './inputWidgets/ApprovalWidget.svelte';
  import FileInputWidget from './inputWidgets/FileInputWidget.svelte';
  import JsonInputWidget from './inputWidgets/JsonInputWidget.svelte';
  import TranscriptPane from './TranscriptPane.svelte';
  import AstNode from './AstNode.svelte';
  import HintBanner from './HintBanner.svelte';
  import DiagnosticsChip from './DiagnosticsChip.svelte';
  import RawViewToggle from './RawViewToggle.svelte';
  import { makeAstResponses } from '../../stores/astResponses';
  import { EventsOn } from '../../../wailsjs/runtime/runtime.js';
  import { errorMessage } from '../../lib/errorMessage';
  import { normaliseInteractiveTurn } from '../../types/transcript';

  // Keep value in sync with style.css `--duration-medium` — the Svelte
  // transition API takes a number, not a CSS variable.
  const AST_FLY_DURATION_MS = 150;
  const AST_FLY_STAGGER_MS = 40;

  const reduceMotion =
    typeof window !== 'undefined' && typeof window.matchMedia === 'function'
      ? window.matchMedia('(prefers-reduced-motion: reduce)').matches
      : false;

  /** @type {import('../../stores/interactiveInput').PendingPrompt | null} */
  export let prompt = null;
  export let execId = '';
  export let repoName = '';

  /** @type {import('svelte').EventDispatcher<{ responded: { nodeId: string; inputId: string }; close: void }>} */
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

  // Per-modal response map — reset when `structured` flips so state never
  // leaks between suspensions (spec §6.3).
  const responses = makeAstResponses();
  /** @type {string | undefined} */
  let lastStructuredRef = undefined;
  /** @type {string} */
  let activeGroupKey = '';
  $: if (prompt?.structured !== lastStructuredRef) {
    lastStructuredRef = prompt?.structured;
    responses.set({});
    activeGroupKey = '';
  }

  // Derived from the `prompt` prop (not `$pendingAst`) so the AST lands in
  // the same reactive cycle as the incoming prompt — routing through the
  // store adds a tick of lag that hid the Send button in synchronous tests.
  $: parsedAst = parseStructuredAst(prompt?.structured);
  $: diagnostics = parsedAst?.diagnostics ?? null;
  $: generatedBy = parsedAst?.generated_by ?? '';
  $: isFallback = typeof generatedBy === 'string' && generatedBy.startsWith('fallback:');
  $: rawAutoExpand =
    diagnostics?.untrusted === true ||
    (diagnostics?.fallback_reasons?.length ?? 0) > 0 ||
    isFallback;
  $: showOllamaOffline =
    $uiAdapterEnabled && $ollamaEnabled && $ollamaReachable === false;
  $: decisionGroups = parsedAst?.nodes.filter((n) => n.type === 'decision_group') ?? [];
  $: requiredGroupKeys = decisionGroups
    .filter((g) => g.required && !!g.response_key)
    .map((g) => /** @type {string} */ (g.response_key));
  $: showCollapseBanner = shape !== 'json' && decisionGroups.length > 1;
  $: showModalSend = decisionGroups.length > 0;
  $: hideLayer1Widget = decisionGroups.length > 0;
  // Under collapse the user picks ONE group — gating on every required key
  // would deadlock Send when the active group is non-required but other
  // required groups sit unfilled (which is the normal case).
  $: sendDisabled =
    sending ||
    (showCollapseBanner
      ? !($responses[activeGroupKey] ?? '')
      : requiredGroupKeys.some((k) => !($responses[k] ?? '')));

  // Under collapse rule, activate the first required group (else first group)
  // on initial render. User clicks on another card swap activation.
  $: if (showCollapseBanner && !activeGroupKey) {
    const first = decisionGroups.find((g) => g.required) ?? decisionGroups[0];
    activeGroupKey = first?.response_key ?? '';
  }

  // Reactive closures — rebuilt whenever `showCollapseBanner` or
  // `activeGroupKey` change, so AstNode's prop-identity check fires and the
  // derived `disabled` / `active` flow into each DecisionGroup card.
  /** @type {(group: import('../../types/uiAst').UINode) => boolean} */
  $: isGroupDisabled = (group) => {
    if (!showCollapseBanner) return false;
    return group.response_key !== activeGroupKey;
  };
  /** @type {(group: import('../../types/uiAst').UINode) => boolean} */
  $: isGroupActive = (group) => {
    if (!showCollapseBanner) return false;
    return group.response_key === activeGroupKey;
  };

  /** @param {CustomEvent<{ key: string }>} e */
  function onGroupActivate(e) {
    if (!showCollapseBanner) return;
    activeGroupKey = e.detail.key;
  }

  async function onSend() {
    if (!prompt || sending) return;
    if (sendDisabled) {
      triggerShake();
      return;
    }
    sending = true;
    lastError = '';
    try {
      const r = get(responses);
      const eid = execId || prompt.execId || '';
      let value;
      if (shape === 'json') {
        value = JSON.stringify(r);
      } else if (decisionGroups.length === 1) {
        const k = decisionGroups[0].response_key ?? '';
        value = r[k] ?? '';
      } else {
        // Collapse rule: submit the user's currently-selected group, not the
        // first-required. activeGroupKey is the source of truth; fall back
        // only when it's somehow empty (shouldn't happen post-initial-render).
        const active =
          decisionGroups.find((g) => g.response_key === activeGroupKey) ?? decisionGroups[0];
        const k = active?.response_key ?? '';
        value = r[k] ?? '';
        pushToast(
          'info',
          'Only your active answer was sent — this process accepts a single answer.',
        );
      }
      await RespondToInput(eid, prompt.nodeId, prompt.inputId, value);
      dispatch('responded', { nodeId: prompt.nodeId, inputId: prompt.inputId });
    } catch (err) {
      lastError = errorMessage(err);
      triggerShake();
    } finally {
      sending = false;
    }
  }

  /** @type {import('../../types/transcript').Turn[]} */
  let transcript = [];
  let transcriptLoading = false;
  let transcriptError = '';
  /** @type {(() => void) | null} */
  let cancelRoundListener = null;

  async function loadTranscript() {
    if (!prompt) return;
    const eid = execId || prompt.execId || '';
    if (!eid || !prompt.nodeId) return;
    transcriptLoading = true;
    transcriptError = '';
    try {
      const turns = await GetInteractiveTranscript(eid, prompt.nodeId);
      transcript = Array.isArray(turns) ? turns.map(normaliseInteractiveTurn) : [];
    } catch (err) {
      transcriptError = errorMessage(err);
    } finally {
      transcriptLoading = false;
    }
  }

  onMount(async () => {
    await tick();
    loadTranscript();
    // Refresh the transcript every time Claude finishes a round for this
    // node — round_complete event carries {execId, nodeId, round}.
    cancelRoundListener = EventsOn('bmad:node:round_complete', (event) => {
      if (!prompt || !event) return;
      if (event.nodeId !== prompt.nodeId) return;
      loadTranscript();
    });
  });

  // Reload when the active prompt flips to a new node/exec — happens when
  // the user closes one modal and a queued prompt opens another.
  $: if (prompt && prompt.nodeId) {
    loadTranscript();
  }

  // §6.1: mirror the `prompt` prop into the singular `pendingPrompt` store so
  // `parsedAst` can derive from the structured payload. Cleared on destroy
  // so a stale AST never leaks to the next modal open.
  //
  // Identity-guarded — unrelated reactivity (transcript reloads, shape
  // recompute) must not re-trigger the derived JSON.parse.
  /** @type {import('../../stores/interactiveInput').PendingPrompt | null} */
  let lastPromptRef = null;
  $: if (prompt !== lastPromptRef) {
    lastPromptRef = prompt;
    pendingPrompt.set(prompt);
  }

  onDestroy(() => {
    shaking = false;
    if (cancelRoundListener) cancelRoundListener();
    pendingPrompt.set(null);
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
      lastError = errorMessage(err);
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
      return;
    }
    if (showModalSend && (e.metaKey || e.ctrlKey) && e.key === 'Enter') {
      e.preventDefault();
      if (sendDisabled) triggerShake();
      else onSend();
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
        <TranscriptPane
          turns={transcript}
          loading={transcriptLoading}
          error={transcriptError}
        />
        <blockquote class="prompt-block" data-testid="input-modal-prompt">
          {prompt.prompt}
        </blockquote>

        {#if showCollapseBanner}
          <div data-testid="collapse-banner">
            <HintBanner
              tone="warn"
              content="This process accepts a single answer — Pick one decision to submit. Others are hidden."
            />
          </div>
        {/if}

        {#if parsedAst}
          <div class="ast-region" data-testid="ast-region">
            {#each parsedAst.nodes as node, i (i)}
              <div
                class="ast-node-wrapper"
                in:fly={{
                  y: reduceMotion ? 0 : 8,
                  duration: reduceMotion ? 0 : AST_FLY_DURATION_MS,
                  delay: reduceMotion ? 0 : i * AST_FLY_STAGGER_MS,
                  easing: cubicOut,
                }}
              >
                <AstNode
                  {node}
                  {responses}
                  {isGroupDisabled}
                  {isGroupActive}
                  on:activate={onGroupActivate}
                />
              </div>
            {/each}
          </div>
          <DiagnosticsChip {diagnostics} {generatedBy} />
          {#if showOllamaOffline}
            <div class="ollama-offline-notice" role="status" data-testid="ollama-offline-notice">
              Ollama not reachable at localhost:11434 — structured UI is rendered from fallback path.
            </div>
          {/if}
          <RawViewToggle
            raw={prompt.lastOutput ?? ''}
            triggered={rawAutoExpand}
            expandedByDefault={$uiAdapterUntrustedExpanded || isFallback}
          />
        {/if}

        {#if !hideLayer1Widget}
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

      {#if hasFooter || showModalSend}
        <footer class="modal-footer">
          {#if showModalSend}
            <span class="footer-hint">Cmd+Enter to send</span>
          {/if}
          <button type="button" class="btn-cancel" data-testid="input-modal-cancel" on:click={close}>
            Cancel
          </button>
          {#if showModalSend}
            <button
              type="button"
              class="btn-send"
              data-testid="input-modal-send"
              disabled={sendDisabled}
              on:click={onSend}
            >
              <Send size={13} aria-hidden="true" />
              Send all responses
            </button>
          {/if}
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
    max-width: 880px;
    min-width: var(--modal-width-min);
    max-height: 82vh;
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

  .ast-region {
    display: flex;
    flex-direction: column;
    gap: var(--sp-md);
    padding: 0;
    width: 100%;
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
    align-items: center;
    justify-content: flex-end;
    gap: var(--sp-md);
    padding: var(--sp-md) var(--sp-lg);
    border-top: 1px solid var(--border-subtle);
    flex-shrink: 0;
  }

  .footer-hint {
    margin-right: auto;
    font-family: var(--font-mono);
    font-size: var(--text-label);
    color: var(--text-muted);
  }

  .btn-send {
    display: inline-flex;
    align-items: center;
    gap: var(--sp-xs);
    min-width: var(--btn-min-wide);
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

  .ollama-offline-notice {
    margin-top: var(--sp-xs);
    padding: var(--sp-xs) var(--sp-sm);
    border: 1px solid color-mix(in srgb, var(--accent-amber) 30%, transparent);
    border-radius: var(--radius-sm);
    background: color-mix(in srgb, var(--accent-amber) 10%, transparent);
    color: var(--accent-amber);
    font-family: var(--font-mono);
    font-size: var(--text-body);
  }
</style>
