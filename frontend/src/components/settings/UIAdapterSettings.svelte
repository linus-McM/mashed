<script lang="ts">
  // UI AST adapter panel (enable toggles, timeout, Ollama model picker,
  // offline banner). Extracted from views/Settings.svelte (spec R31) —
  // markup and scoped styles moved verbatim; the shared panel/section and
  // setting-row/control styles are duplicated here because Svelte scopes
  // them per component. Store hydration stays in the parent's onMount.
  import {
    uiAdapterEnabled,
    uiAdapterTimeoutMs,
    ollamaModel,
    ollamaEnabled,
    uiAdapterUntrustedExpanded,
    ollamaReachable,
    ollamaModels,
    refreshModels as refreshUIAdapterModels,
    setEnabled as setUIAdapterEnabled,
    setTimeoutMs as setUIAdapterTimeoutMs,
    setModel as setUIAdapterModel,
    setOllamaEnabled as setUIAdapterOllamaEnabled,
    setUntrustedExpanded as setUIAdapterUntrustedExpanded,
    validOllamaModelName,
  } from '../../lib/stores/uiAdapterSettings';
  import { BrowserOpenURL } from '../../../wailsjs/runtime/runtime.js';
  import { RefreshCw, TriangleAlert } from 'lucide-svelte';

  /** Called after a successful save so the parent can flash "Saved". */
  export let onSaved: () => void = () => {};

  // UI AST adapter — local form state mirrors the store; binds to inputs so
  // blur-validated edits don't push partial keystrokes to disk.
  let timeoutInput = 3000;
  let timeoutError = '';
  let modelSelectValue = 'gemma3:4b';
  let customModelInput = '';
  let customModelError = '';
  let refreshingModels = false;

  const OLLAMA_INSTALL_URL = 'https://ollama.com/download';
  const CUSTOM_MODEL_SENTINEL = '__custom__';

  $: if ($uiAdapterTimeoutMs !== undefined) timeoutInput = $uiAdapterTimeoutMs;
  $: if ($ollamaModel) {
    modelSelectValue = resolveModelSelectValue($ollamaModel, $ollamaModels);
    if (modelSelectValue === CUSTOM_MODEL_SENTINEL && !customModelInput) {
      customModelInput = $ollamaModel;
    }
  }

  function resolveModelSelectValue(model: string, availableModels: string[]): string {
    if (!model) return CUSTOM_MODEL_SENTINEL;
    if (availableModels && availableModels.includes(model)) return model;
    return CUSTOM_MODEL_SENTINEL;
  }

  async function toggleUIAdapter() {
    const ok = await setUIAdapterEnabled(!$uiAdapterEnabled);
    if (ok) {
      onSaved();
      if ($uiAdapterEnabled) await refreshUIAdapterModels();
    }
  }

  async function toggleOllamaEnabled() {
    const ok = await setUIAdapterOllamaEnabled(!$ollamaEnabled);
    if (ok) onSaved();
  }

  async function toggleUntrustedExpanded() {
    const ok = await setUIAdapterUntrustedExpanded(!$uiAdapterUntrustedExpanded);
    if (ok) onSaved();
  }

  async function commitTimeout() {
    const ms = Number(timeoutInput);
    if (!Number.isFinite(ms) || ms < 500 || ms > 30000) {
      timeoutError = 'must be 500-30000';
      return;
    }
    timeoutError = '';
    if (ms === $uiAdapterTimeoutMs) return;
    const ok = await setUIAdapterTimeoutMs(ms);
    if (ok) onSaved();
    else timeoutError = 'save failed';
  }

  async function onModelSelect(value: string): Promise<void> {
    modelSelectValue = value;
    if (value === CUSTOM_MODEL_SENTINEL) {
      customModelInput = $ollamaModel;
      return;
    }
    customModelError = '';
    const ok = await setUIAdapterModel(value);
    if (ok) onSaved();
  }

  async function commitCustomModel() {
    if (!validOllamaModelName(customModelInput)) {
      customModelError = 'only [a-zA-Z0-9._:-], 1-64 chars';
      return;
    }
    customModelError = '';
    if (customModelInput === $ollamaModel) return;
    const ok = await setUIAdapterModel(customModelInput);
    if (ok) onSaved();
    else customModelError = 'save failed';
  }

  async function manualRefreshModels() {
    refreshingModels = true;
    try {
      await refreshUIAdapterModels();
    } finally {
      refreshingModels = false;
    }
  }

  function openOllamaDocs() {
    BrowserOpenURL(OLLAMA_INSTALL_URL);
  }

  $: sortedOllamaModels = [...($ollamaModels || [])].sort((a, b) => a.localeCompare(b));
  $: configuredModelMissing =
    $ollamaModel && !sortedOllamaModels.includes($ollamaModel);
  $: showOfflineBanner = $uiAdapterEnabled && $ollamaReachable === false;
</script>

<!-- UI AST adapter -->
<div class="settings-panel">
<section class="settings-section" data-testid="ui-adapter-section">
  <h2 class="section-title">UI AST adapter</h2>
  <p class="section-desc">
    Translates Claude's round output into typed widgets. Requires Ollama running locally.
    Changes take effect on next app launch.
  </p>

  <div class="setting-row">
    <label class="setting-label" for="uiAdapterEnabled">Enable UI AST adapter</label>
    <button
      id="uiAdapterEnabled"
      class="setting-toggle"
      class:active={$uiAdapterEnabled}
      data-testid="ui-adapter-toggle"
      on:click={toggleUIAdapter}
    >
      {$uiAdapterEnabled ? 'On' : 'Off'}
    </button>
  </div>

  <div class="setting-row">
    <label class="setting-label" for="ollamaEnabled">Ollama enabled</label>
    <button
      id="ollamaEnabled"
      class="setting-toggle"
      class:active={$ollamaEnabled}
      on:click={toggleOllamaEnabled}
    >
      {$ollamaEnabled ? 'On' : 'Off'}
    </button>
  </div>

  <div class="setting-row">
    <label class="setting-label" for="uiAdapterTimeout">Timeout (ms)</label>
    <input
      id="uiAdapterTimeout"
      class="setting-number"
      class:invalid={timeoutError}
      type="number"
      min="500"
      max="30000"
      step="100"
      bind:value={timeoutInput}
      on:blur={commitTimeout}
    />
  </div>
  {#if timeoutError}
    <div class="setting-error" role="alert">{timeoutError}</div>
  {/if}

  <div class="setting-row">
    <label class="setting-label" for="ollamaModel">Ollama model</label>
    <div class="model-row-controls">
      <select
        id="ollamaModel"
        class="setting-select"
        data-testid="ui-adapter-model-select"
        value={modelSelectValue}
        on:change={(e) => onModelSelect(e.currentTarget.value)}
      >
        {#each sortedOllamaModels as m}
          <option value={m}>{m}</option>
            {/each}
        {#if configuredModelMissing}
          <option value={$ollamaModel}
            >{$ollamaModel}{sortedOllamaModels.length > 0 ? ' (not pulled)' : ''}</option
          >
        {/if}
        <option value={CUSTOM_MODEL_SENTINEL}>Custom…</option>
      </select>
      <button
        class="icon-btn"
        type="button"
        aria-label="Refresh model list"
        title="Refresh model list"
        data-testid="ui-adapter-refresh"
        disabled={refreshingModels}
        on:click={manualRefreshModels}
      >
        <RefreshCw size={12} />
      </button>
    </div>
  </div>

  {#if modelSelectValue === '__custom__'}
    <div class="setting-row custom-model-row">
      <label class="setting-label" for="customOllamaModel">Custom model</label>
      <input
        id="customOllamaModel"
        class="path-input custom-model-input"
        class:invalid={customModelError}
        type="text"
        placeholder="e.g. llama3.2:3b"
        bind:value={customModelInput}
        on:blur={commitCustomModel}
        data-testid="ui-adapter-custom-model"
      />
    </div>
    {#if customModelError}
      <div class="setting-error" role="alert">{customModelError}</div>
    {/if}
  {/if}

  <div class="setting-row">
    <label class="setting-label" for="untrustedExpanded">
      Expand "View raw" by default when a translation is flagged as untrusted
    </label>
    <button
      id="untrustedExpanded"
      class="setting-toggle"
      class:active={$uiAdapterUntrustedExpanded}
      on:click={toggleUntrustedExpanded}
    >
      {$uiAdapterUntrustedExpanded ? 'On' : 'Off'}
    </button>
  </div>

  {#if showOfflineBanner}
    <aside class="offline-banner" role="status" data-testid="ui-adapter-offline-banner">
      <span class="offline-banner-icon" aria-hidden="true">
        <TriangleAlert size={14} />
      </span>
      <div class="offline-banner-body">
        <div class="offline-banner-title">Ollama not reachable at localhost:11434.</div>
        <div class="offline-banner-text">
          Install Ollama and pull the model to enable this.
        </div>
        <a
          class="offline-banner-link"
          href={OLLAMA_INSTALL_URL}
          on:click|preventDefault={openOllamaDocs}
        >
          Install docs
        </a>
      </div>
    </aside>
  {/if}
</section>
</div>

<style>
  .settings-panel {
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    padding: var(--sp-lg);
  }

  /* Panel padding owns bottom spacing; neutralise the legacy section margin. */
  .settings-panel .settings-section {
    margin-bottom: 0;
  }

  .section-title {
    font-family: var(--font-mono);
    font-size: var(--text-data);
    font-weight: 600;
    color: var(--text-primary);
    margin-bottom: var(--sp-md);
  }

  .section-desc {
    font-size: var(--text-body);
    color: var(--text-dim);
    margin-bottom: var(--sp-md);
  }

  .settings-section {
    margin-bottom: var(--sp-2xl);
  }

  .setting-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: var(--sp-xs) 0;
    gap: var(--sp-md);
  }

  .setting-label {
    font-family: var(--font-mono);
    font-size: var(--text-body);
    color: var(--text-dim);
    white-space: nowrap;
  }

  .setting-select {
    padding: 4px 8px;
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-sm);
    color: var(--text-primary);
    font-family: var(--font-mono);
    font-size: 11px;
    outline: none;
    cursor: pointer;
    min-width: 120px;
    transition: border-color 100ms ease;
  }

  .setting-select:focus {
    border-color: var(--accent-green);
  }

  .setting-select:hover {
    border-color: var(--border-emphasis);
  }

  .setting-number {
    width: 60px;
    padding: 4px 8px;
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-sm);
    color: var(--text-primary);
    font-family: var(--font-mono);
    font-size: 11px;
    font-variant-numeric: tabular-nums;
    outline: none;
    text-align: center;
    transition: border-color 100ms ease;
  }

  .setting-number:focus {
    border-color: var(--accent-green);
  }

  .setting-toggle {
    padding: 4px 12px;
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-sm);
    color: var(--text-muted);
    font-family: var(--font-mono);
    font-size: 11px;
    cursor: pointer;
    min-width: 48px;
    text-align: center;
    transition: all 100ms ease;
  }

  .setting-toggle:hover {
    border-color: var(--border-emphasis);
    color: var(--text-dim);
  }

  .setting-toggle.active {
    background: var(--bg-active);
    border-color: var(--accent-green);
    color: var(--accent-green);
  }

  .path-input {
    flex: 1;
    padding: var(--sp-xs) 10px;
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    color: var(--text-primary);
    font-family: var(--font-mono);
    font-size: var(--text-body);
    outline: none;
  }

  .path-input:focus {
    border-color: var(--accent-green);
  }

  .path-input::placeholder {
    color: var(--text-muted);
  }

  /* UI AST adapter section */
  .setting-number.invalid,
  .path-input.invalid {
    border-color: var(--accent-red);
  }

  .setting-error {
    padding-left: var(--sp-xs);
    margin-top: var(--sp-2xs);
    font-family: var(--font-ui);
    font-size: var(--text-label);
    font-weight: 500;
    color: var(--accent-red);
  }

  .model-row-controls {
    display: flex;
    align-items: center;
    gap: var(--sp-xs);
  }

  .icon-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 24px;
    height: 24px;
    padding: 0;
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-sm);
    color: var(--text-dim);
    cursor: pointer;
    transition: color var(--duration-short) var(--ease-move),
      border-color var(--duration-short) var(--ease-move);
  }

  .icon-btn:hover:not(:disabled) {
    color: var(--text-primary);
    border-color: var(--border-emphasis);
  }

  .icon-btn:disabled {
    opacity: 0.5;
    cursor: wait;
  }

  .custom-model-row {
    align-items: stretch;
  }

  .custom-model-input {
    flex: 1;
    max-width: 260px;
  }

  .offline-banner {
    display: flex;
    align-items: flex-start;
    gap: var(--sp-sm);
    margin-top: var(--sp-md);
    padding: var(--sp-sm) var(--sp-md);
    background: color-mix(in srgb, var(--accent-amber) 10%, transparent);
    border-left: 3px solid var(--accent-amber);
    border-radius: 0 var(--radius-md) var(--radius-md) 0;
  }

  .offline-banner-icon {
    display: inline-flex;
    align-items: center;
    color: var(--accent-amber);
    margin-top: 2px;
    flex-shrink: 0;
  }

  .offline-banner-body {
    display: flex;
    flex-direction: column;
    gap: var(--sp-2xs);
    min-width: 0;
  }

  .offline-banner-title {
    font-family: var(--font-ui);
    font-size: var(--text-body);
    font-weight: 600;
    color: var(--accent-amber);
  }

  .offline-banner-text {
    font-family: var(--font-ui);
    font-size: var(--text-label);
    color: var(--text-dim);
  }

  .offline-banner-link {
    align-self: flex-start;
    margin-top: var(--sp-2xs);
    padding: 0;
    background: none;
    border: none;
    font-family: var(--font-ui);
    font-size: var(--text-label);
    font-weight: 500;
    color: var(--accent-green);
    cursor: pointer;
    text-decoration: underline;
    text-underline-offset: 2px;
  }

  .offline-banner-link:hover {
    text-shadow: var(--glow-spread) color-mix(in srgb, var(--accent-green) 60%, transparent);
  }
</style>
