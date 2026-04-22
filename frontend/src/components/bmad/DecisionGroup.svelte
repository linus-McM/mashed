<script>
  import { createEventDispatcher } from 'svelte';
  import ChoiceWidget from './inputWidgets/ChoiceWidget.svelte';
  import MultiChoiceWidget from './inputWidgets/MultiChoiceWidget.svelte';
  import ApprovalWidget from './inputWidgets/ApprovalWidget.svelte';
  import FreeTextWidget from './inputWidgets/FreeTextWidget.svelte';
  import FileInputWidget from './inputWidgets/FileInputWidget.svelte';
  import JsonInputWidget from './inputWidgets/JsonInputWidget.svelte';

  /** @type {import('../../types/uiAst').UINode} */
  export let node;
  /** @type {import('svelte/store').Writable<Record<string,string>>} */
  export let responses;
  export let disabled = false;
  // Under the collapse rule, the owning modal marks exactly one group `active`
  // so the accent-green border fires before the user focuses into the card.
  // Pre-focus state is the signature indicator per Design Brief §3 + §6.
  export let active = false;

  const dispatch = createEventDispatcher();

  $: widget = node.widget;
  $: responseKey = node.response_key ?? '';
  $: labelId = `dg-${responseKey}`;
  $: widgetPrompt = {
    prompt: node.prompt ?? node.heading ?? '',
    options: widget?.options ?? [],
    required: !!node.required,
    maxLength: widget?.maxLength,
    helpText: node.help ?? '',
  };

  /** @param {string} v */
  function onValue(v) {
    if (!responseKey) return;
    responses.update((r) => ({ ...r, [responseKey]: v }));
  }

  /** @param {CustomEvent<{ value: string }>} e */
  function onWidgetSubmit(e) {
    onValue(e.detail?.value ?? '');
  }

  function onCardActivate() {
    if (disabled) dispatch('activate', { key: responseKey });
  }

  /** @param {KeyboardEvent} e */
  function onCardKey(e) {
    if (!disabled) return;
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      dispatch('activate', { key: responseKey });
    }
  }
</script>

<!-- svelte-ignore a11y-click-events-have-key-events -->
<!-- svelte-ignore a11y-no-noninteractive-tabindex -->
<!--
  Spec §6.4 prescribes tabindex=-1 on inactive cards; we deviate to -1/0
  (active/inactive) so keyboard-only users can Tab to an inactive card and
  press Enter/Space to activate it. The spec's -1 would strand them with
  no way to switch groups without a pointer.
-->
<section
  class="decision-group"
  class:is-disabled={disabled}
  class:is-active={active}
  data-testid="decision-group"
  data-key={responseKey}
  aria-labelledby={labelId}
  role={disabled ? 'button' : undefined}
  tabindex={disabled ? 0 : -1}
  on:click={onCardActivate}
  on:keydown={onCardKey}
>
  <header class="dg-heading">
    <h3 id={labelId}>{node.heading ?? ''}</h3>
    {#if node.required}
      <span class="required-pill" data-testid="required-pill">Required</span>
    {/if}
  </header>
  {#if node.prompt}
    <p class="prompt">{node.prompt}</p>
  {/if}
  {#if node.help}
    <p class="help">{node.help}</p>
  {/if}

  <div class="widget-wrap">
    {#if widget?.type === 'choice'}
      <ChoiceWidget prompt={widgetPrompt} {disabled} on:submit={onWidgetSubmit} />
    {:else if widget?.type === 'multi'}
      <MultiChoiceWidget prompt={widgetPrompt} {disabled} on:submit={onWidgetSubmit} />
    {:else if widget?.type === 'approval'}
      <ApprovalWidget prompt={widgetPrompt} {disabled} on:submit={onWidgetSubmit} />
    {:else if widget?.type === 'free'}
      <FreeTextWidget prompt={widgetPrompt} {disabled} on:submit={onWidgetSubmit} />
    {:else if widget?.type === 'file'}
      <FileInputWidget prompt={widgetPrompt} {disabled} on:submit={onWidgetSubmit} />
    {:else if widget?.type === 'json'}
      <JsonInputWidget prompt={widgetPrompt} {disabled} on:submit={onWidgetSubmit} />
    {/if}
  </div>
</section>

<style>
  .decision-group {
    display: flex;
    flex-direction: column;
    gap: var(--sp-xs);
    padding: var(--sp-md);
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    transition:
      opacity var(--duration-short) var(--ease-move),
      border-color var(--duration-short) var(--ease-move);
  }

  .decision-group.is-active,
  .decision-group:focus-within:not(.is-disabled) {
    border-color: var(--accent-green);
  }

  .decision-group.is-disabled {
    opacity: 0.45;
    cursor: pointer;
  }

  .decision-group.is-disabled:hover {
    opacity: 0.7;
  }

  .decision-group.is-disabled:focus-visible {
    outline: 1px solid var(--accent-green);
    outline-offset: 2px;
  }

  .decision-group.is-disabled .widget-wrap {
    pointer-events: none;
  }

  .dg-heading {
    display: flex;
    align-items: center;
    gap: var(--sp-sm);
  }

  .dg-heading h3 {
    margin: 0;
    font-family: var(--font-ui);
    font-size: var(--text-data);
    font-weight: 600;
    color: var(--text-primary);
  }

  .required-pill {
    background: color-mix(in srgb, var(--accent-green) 15%, transparent);
    color: var(--accent-green);
    padding: 0 var(--sp-xs);
    border-radius: var(--radius-sm);
    font-family: var(--font-mono);
    font-size: var(--text-label);
    font-weight: 600;
    letter-spacing: 0.05em;
    text-transform: uppercase;
  }

  .prompt {
    margin: 0;
    margin-top: var(--sp-2xs);
    font-family: var(--font-ui);
    font-size: var(--text-body);
    color: var(--text-primary);
    line-height: 1.5;
  }

  .help {
    margin: 0;
    margin-top: var(--sp-2xs);
    font-family: var(--font-ui);
    font-size: var(--text-label);
    color: var(--text-dim);
    font-style: italic;
  }

  .widget-wrap {
    margin-top: var(--sp-md);
  }
</style>
