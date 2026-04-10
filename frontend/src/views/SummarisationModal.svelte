<script>
  import { onMount, onDestroy, createEventDispatcher, tick } from 'svelte';
  import { fade, fly } from 'svelte/transition';
  import { cubicOut } from 'svelte/easing';
  import { FileText, Plus, Minus, X, Sparkles, ChevronDown, FileCode, ExternalLink } from 'lucide-svelte';
  import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime.js';
  import { StreamCodeReviewSummary, ListAdviceModes, StreamAdvice, SpawnRefactorPlan, ListModels } from '../../wailsjs/go/main/App.js';

  const dispatch = createEventDispatcher();

  export let repoPath = '';

  // Summary state
  let files = [];
  let summaryLoading = true;
  let summaryError = '';
  let totalAdded = 0;
  let totalRemoved = 0;
  let currentFileIndex = 0;
  let totalFiles = 0;

  // Advice state
  let adviceModes = [];
  let selectedMode = '';
  let adviceText = '';
  let adviceLoading = false;
  let adviceError = '';

  // Model state
  let modelList = [];
  let selectedModel = '';

  // Refactor plan state
  let planPath = '';
  let planLoading = false;
  let planError = '';

  // DOM refs
  let advicePanel;
  let fileListEl;

  // Computed
  $: hasFiles = files.length > 0;
  $: showAdviceSection = hasFiles || !summaryLoading;
  $: canGetAdvice = selectedMode && !adviceLoading;
  $: canCreatePlan = adviceText && !planLoading;

  onMount(async () => {
    EventsOn('review:summary:progress', (data) => {
      if (data.repoPath !== repoPath) return;
      if (data.error) {
        summaryError = data.error;
        return;
      }
      files = [...files, data.file];
      currentFileIndex = data.index + 1;
      totalFiles = data.total;
    });

    EventsOn('review:summary:done', (data) => {
      if (data.repoPath !== repoPath) return;
      summaryLoading = false;
      if (data.error) {
        summaryError = data.error;
        return;
      }
      if (data.summary) {
        totalAdded = data.summary.totalAdded;
        totalRemoved = data.summary.totalRemoved;
      }
    });

    EventsOn('review:advice:progress', async (data) => {
      if (data.repoPath !== repoPath) return;
      if (data.error) {
        adviceError = data.error;
        adviceLoading = false;
        return;
      }
      if (data.done) {
        adviceLoading = false;
        return;
      }
      adviceText += data.text + '\n';
      await tick();
      if (advicePanel) {
        advicePanel.scrollTop = advicePanel.scrollHeight;
      }
    });

    // Load models and advice modes (non-fatal)
    try {
      const [modeList, models] = await Promise.all([
        ListAdviceModes(repoPath),
        ListModels(),
      ]);
      adviceModes = modeList;
      modelList = models || [];
      const balanced = modelList.find(m => m.tier === 'balanced');
      selectedModel = balanced ? balanced.id : (modelList[0]?.id || '');
    } catch (_) { /* non-fatal */ }

    // Start streaming summaries
    StreamCodeReviewSummary(repoPath, selectedModel);
  });

  onDestroy(() => {
    EventsOff('review:summary:progress');
    EventsOff('review:summary:done');
    EventsOff('review:advice:progress');
  });

  function close() {
    dispatch('close');
  }

  function handleKeydown(e) {
    if (e.key === 'Escape') close();
  }

  function openFile(path) {
    dispatch('open-file', { path });
  }

  function getAdvice() {
    if (!canGetAdvice) return;
    adviceText = '';
    adviceError = '';
    adviceLoading = true;
    StreamAdvice(repoPath, selectedMode, selectedModel);
  }

  async function createRefactorPlan() {
    if (!canCreatePlan) return;
    planLoading = true;
    planError = '';
    try {
      planPath = await SpawnRefactorPlan(repoPath, adviceText);
    } catch (e) {
      planError = e?.message || 'Failed to create plan';
    }
    planLoading = false;
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<div class="overlay" role="presentation" transition:fade={{ duration: 150 }} on:click={close} on:keydown={() => {}}>
  <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
  <div
    class="modal"
    role="dialog"
    aria-label="Code Review Summary"
    transition:fly={{ y: 24, duration: 250, easing: cubicOut }}
    on:click|stopPropagation
    on:keydown={() => {}}
  >
    <!-- Header -->
    <div class="modal-header">
      <div class="header-title">
        <FileText size={18} />
        <h2>Code Review Summary</h2>
      </div>
      <button class="close-btn" on:click={close} title="Close (Esc)">
        <X size={16} />
      </button>
    </div>

    <!-- Error banner -->
    {#if summaryError}
      <div class="error-banner" transition:fly={{ y: -8, duration: 200 }}>
        <span>{summaryError}</span>
      </div>
    {/if}

    <!-- Loading indicator -->
    {#if summaryLoading && !summaryError}
      <div class="loading-bar">
        <div class="loading-text">
          <Sparkles size={14} />
          <span>
            Summarising{#if totalFiles > 0} file {currentFileIndex} of {totalFiles}{/if}<span class="dots"></span>
          </span>
        </div>
        {#if totalFiles > 0}
          <div class="progress-track">
            <div class="progress-fill" style="width: {(currentFileIndex / totalFiles) * 100}%"></div>
          </div>
        {/if}
      </div>
    {/if}

    <!-- File list -->
    <div class="file-list" bind:this={fileListEl}>
      {#if files.length === 0 && !summaryLoading && !summaryError}
        <div class="empty-state">
          <FileCode size={24} />
          <span>No changes to review</span>
        </div>
      {/if}

      {#each files as file, i}
        <div
          class="file-card"
          in:fly={{ y: 12, duration: 200, delay: Math.min(i * 40, 400), easing: cubicOut }}
        >
          <div class="file-header">
            <button class="file-path" on:click={() => openFile(file.path)} title="Open {file.path}">
              <FileCode size={13} />
              <span class="path-text">{file.path}</span>
              <ExternalLink size={11} />
            </button>
            <div class="file-stats">
              {#if file.added > 0}
                <span class="stat-added"><Plus size={11} />{file.added}</span>
              {/if}
              {#if file.removed > 0}
                <span class="stat-removed"><Minus size={11} />{file.removed}</span>
              {/if}
            </div>
          </div>
          {#if file.summary}
            <p class="file-summary">{file.summary}</p>
          {/if}
        </div>
      {/each}
    </div>

    <!-- Overall totals -->
    {#if !summaryLoading && hasFiles}
      <div class="totals-bar" in:fly={{ y: 8, duration: 200, easing: cubicOut }}>
        <span class="totals-label">Overall</span>
        <div class="totals-stats">
          <span class="stat-added"><Plus size={12} />{totalAdded}</span>
          <span class="stat-removed"><Minus size={12} />{totalRemoved}</span>
        </div>
      </div>
    {/if}

    <!-- Advice section -->
    {#if showAdviceSection && hasFiles}
      <div class="divider"></div>

      <div class="advice-section">
        <div class="advice-controls">
          <label class="advice-label">Advice</label>
          <div class="select-wrap">
            <select bind:value={selectedMode} class="advice-select">
              <option value="" disabled>Select methodology</option>
              {#each adviceModes as mode}
                <option value={mode.name}>{mode.displayName}</option>
              {/each}
            </select>
            <ChevronDown size={14} />
          </div>
          <div class="select-wrap model-select-wrap">
            <select bind:value={selectedModel} class="advice-select">
              {#each modelList as m}
                <option value={m.id}>{m.displayName}</option>
              {/each}
            </select>
            <ChevronDown size={14} />
          </div>
          <button
            class="action-btn primary"
            disabled={!canGetAdvice}
            on:click={getAdvice}
          >
            <Sparkles size={13} />
            {adviceLoading ? 'Streaming...' : 'Get Advice'}
          </button>
        </div>

        {#if adviceError}
          <div class="error-banner small" transition:fly={{ y: -8, duration: 200 }}>
            <span>{adviceError}</span>
          </div>
        {/if}

        {#if adviceText || adviceLoading}
          <div class="advice-panel" bind:this={advicePanel}>
            <pre>{adviceText}{#if adviceLoading}<span class="cursor-blink">|</span>{/if}</pre>
          </div>
        {/if}

        <!-- Refactor plan -->
        {#if adviceText && !adviceLoading}
          <div class="plan-row" transition:fly={{ y: 8, duration: 200, easing: cubicOut }}>
            <button
              class="action-btn"
              disabled={!canCreatePlan}
              on:click={createRefactorPlan}
            >
              <FileText size={13} />
              {planLoading ? 'Creating plan...' : 'Create Refactor Plan'}
            </button>
            {#if planPath}
              <button class="plan-link" on:click={() => openFile(planPath)} title="Open {planPath}">
                <ExternalLink size={12} />
                <span>{planPath}</span>
              </button>
            {/if}
            {#if planError}
              <span class="plan-error">{planError}</span>
            {/if}
          </div>
        {/if}
      </div>
    {/if}
  </div>
</div>

<style>
  /* Overlay */
  .overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.6);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 100;
  }

  /* Modal body */
  .modal {
    background: var(--bg-surface);
    border: 1px solid var(--border-emphasis);
    border-radius: var(--radius-lg);
    width: 800px;
    max-width: calc(100vw - 64px);
    max-height: 85vh;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  /* Header */
  .modal-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: var(--sp-lg) var(--sp-xl);
    border-bottom: 1px solid var(--border-subtle);
    flex-shrink: 0;
  }

  .header-title {
    display: flex;
    align-items: center;
    gap: var(--sp-sm);
    color: var(--text-primary);
  }

  .header-title h2 {
    font-size: var(--text-section);
    font-weight: 600;
    line-height: 1.2;
    margin: 0;
  }

  .close-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 28px;
    height: 28px;
    border: 1px solid transparent;
    border-radius: var(--radius-md);
    background: none;
    color: var(--text-dim);
    cursor: pointer;
    transition: all var(--duration-short) var(--ease-enter);
  }

  .close-btn:hover {
    color: var(--text-primary);
    background: var(--bg-elevated);
    border-color: var(--border-subtle);
  }

  /* Error banner */
  .error-banner {
    margin: var(--sp-md) var(--sp-xl) 0;
    padding: var(--sp-sm) var(--sp-md);
    background: rgba(232, 69, 69, 0.08);
    border: 1px solid rgba(232, 69, 69, 0.2);
    border-radius: var(--radius-md);
    color: var(--accent-red);
    font-size: var(--text-label);
    font-family: var(--font-mono);
  }

  .error-banner.small {
    margin: var(--sp-sm) 0 0;
  }

  /* Loading */
  .loading-bar {
    padding: var(--sp-md) var(--sp-xl);
    flex-shrink: 0;
  }

  .loading-text {
    display: flex;
    align-items: center;
    gap: var(--sp-sm);
    color: var(--accent-blue);
    font-size: var(--text-label);
    font-family: var(--font-mono);
    margin-bottom: var(--sp-sm);
  }

  .dots::after {
    content: '';
    animation: dots 1.4s steps(4, end) infinite;
  }

  @keyframes dots {
    0%   { content: ''; }
    25%  { content: '.'; }
    50%  { content: '..'; }
    75%  { content: '...'; }
    100% { content: ''; }
  }

  .progress-track {
    height: 2px;
    background: var(--border-subtle);
    border-radius: 1px;
    overflow: hidden;
  }

  .progress-fill {
    height: 100%;
    background: var(--accent-blue);
    border-radius: 1px;
    transition: width 300ms var(--ease-enter);
  }

  /* File list */
  .file-list {
    flex: 1;
    overflow-y: auto;
    padding: var(--sp-md) var(--sp-xl);
    display: flex;
    flex-direction: column;
    gap: var(--sp-sm);
  }

  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--sp-sm);
    padding: var(--sp-3xl) 0;
    color: var(--text-dim);
    font-size: var(--text-body);
  }

  /* File card */
  .file-card {
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    padding: var(--sp-md) var(--sp-lg);
    transition: border-color var(--duration-short) var(--ease-enter);
  }

  .file-card:hover {
    border-color: var(--border-emphasis);
  }

  .file-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--sp-md);
  }

  .file-path {
    display: flex;
    align-items: center;
    gap: var(--sp-xs);
    background: none;
    border: none;
    color: var(--accent-blue);
    font-family: var(--font-mono);
    font-size: var(--text-label);
    cursor: pointer;
    transition: color var(--duration-short) var(--ease-enter);
    text-align: left;
    min-width: 0;
  }

  .file-path:hover {
    color: var(--text-primary);
  }

  .file-path:focus-visible {
    outline: 1px solid var(--accent-blue);
    outline-offset: 2px;
    border-radius: var(--radius-sm);
  }

  .path-text {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .file-stats {
    display: flex;
    align-items: center;
    gap: var(--sp-sm);
    flex-shrink: 0;
    font-family: var(--font-mono);
    font-size: var(--text-label);
    font-variant-numeric: tabular-nums;
  }

  .stat-added {
    display: flex;
    align-items: center;
    gap: 2px;
    color: var(--accent-green);
  }

  .stat-removed {
    display: flex;
    align-items: center;
    gap: 2px;
    color: var(--accent-red);
  }

  .file-summary {
    margin-top: var(--sp-sm);
    color: var(--text-dim);
    font-size: var(--text-label);
    line-height: 1.5;
  }

  /* Totals bar */
  .totals-bar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: var(--sp-sm) var(--sp-xl);
    border-top: 1px solid var(--border-subtle);
    flex-shrink: 0;
  }

  .totals-label {
    font-size: var(--text-label);
    font-weight: 600;
    color: var(--text-dim);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .totals-stats {
    display: flex;
    align-items: center;
    gap: var(--sp-md);
    font-family: var(--font-mono);
    font-size: var(--text-body);
    font-variant-numeric: tabular-nums;
  }

  /* Divider */
  .divider {
    height: 1px;
    background: var(--border-subtle);
    margin: 0 var(--sp-xl);
  }

  /* Advice section */
  .advice-section {
    padding: var(--sp-md) var(--sp-xl) var(--sp-lg);
    flex-shrink: 0;
  }

  .advice-controls {
    display: flex;
    align-items: center;
    gap: var(--sp-sm);
  }

  .advice-label {
    font-size: var(--text-label);
    font-weight: 600;
    color: var(--text-dim);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    margin-right: var(--sp-xs);
    flex-shrink: 0;
  }

  .select-wrap {
    position: relative;
    flex: 1;
    max-width: 260px;
  }

  .model-select-wrap {
    max-width: 180px;
  }

  .select-wrap :global(svg) {
    position: absolute;
    right: var(--sp-sm);
    top: 50%;
    transform: translateY(-50%);
    color: var(--text-dim);
    pointer-events: none;
  }

  .advice-select {
    width: 100%;
    padding: 6px var(--sp-xl) 6px var(--sp-md);
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    color: var(--text-primary);
    font-family: var(--font-ui);
    font-size: var(--text-label);
    appearance: none;
    cursor: pointer;
    transition: border-color var(--duration-short) var(--ease-enter);
  }

  .advice-select:hover {
    border-color: var(--border-emphasis);
  }

  .advice-select:focus-visible {
    outline: 1px solid var(--accent-blue);
    outline-offset: -1px;
  }

  /* Action buttons */
  .action-btn {
    display: flex;
    align-items: center;
    gap: var(--sp-xs);
    padding: 6px var(--sp-md);
    background: none;
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    color: var(--text-dim);
    font-family: var(--font-ui);
    font-size: var(--text-label);
    cursor: pointer;
    transition: all var(--duration-short) var(--ease-enter);
    white-space: nowrap;
    flex-shrink: 0;
  }

  .action-btn:hover:not(:disabled) {
    background: var(--bg-elevated);
    border-color: var(--border-emphasis);
    color: var(--text-primary);
  }

  .action-btn:active:not(:disabled) {
    transform: scale(0.98);
  }

  .action-btn:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }

  .action-btn:focus-visible {
    outline: 1px solid var(--accent-blue);
    outline-offset: 1px;
  }

  .action-btn.primary {
    background: rgba(0, 229, 122, 0.08);
    border-color: rgba(0, 229, 122, 0.2);
    color: var(--accent-green);
  }

  .action-btn.primary:hover:not(:disabled) {
    background: rgba(0, 229, 122, 0.14);
    border-color: var(--accent-green);
  }

  /* Advice panel */
  .advice-panel {
    margin-top: var(--sp-sm);
    background: var(--bg-deepest);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    padding: var(--sp-md);
    max-height: 300px;
    overflow-y: auto;
  }

  .advice-panel pre {
    font-family: var(--font-mono);
    font-size: var(--text-label);
    line-height: 1.6;
    color: var(--text-primary);
    white-space: pre-wrap;
    word-break: break-word;
    margin: 0;
  }

  .cursor-blink {
    animation: blink 1s step-end infinite;
    color: var(--accent-green);
  }

  @keyframes blink {
    0%, 100% { opacity: 1; }
    50% { opacity: 0; }
  }

  /* Plan row */
  .plan-row {
    display: flex;
    align-items: center;
    gap: var(--sp-sm);
    margin-top: var(--sp-sm);
    flex-wrap: wrap;
  }

  .plan-link {
    display: flex;
    align-items: center;
    gap: var(--sp-xs);
    background: none;
    border: none;
    color: var(--accent-blue);
    font-family: var(--font-mono);
    font-size: var(--text-label);
    cursor: pointer;
    transition: color var(--duration-short) var(--ease-enter);
  }

  .plan-link:hover {
    color: var(--text-primary);
  }

  .plan-link:focus-visible {
    outline: 1px solid var(--accent-blue);
    outline-offset: 2px;
    border-radius: var(--radius-sm);
  }

  .plan-link span {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    max-width: 400px;
  }

  .plan-error {
    color: var(--accent-red);
    font-size: var(--text-label);
    font-family: var(--font-mono);
  }
</style>
