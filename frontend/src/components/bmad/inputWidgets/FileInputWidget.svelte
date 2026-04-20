<script>
  import { createEventDispatcher, onMount, tick } from 'svelte';
  import { Upload, Send } from 'lucide-svelte';

  /** @type {import('../../../stores/interactiveInput').PendingPrompt} */
  export let prompt;
  export let disabled = false;
  export let validationError = '';

  const dispatch = createEventDispatcher();

  let path = '';
  let isDragOver = false;
  /** @type {HTMLInputElement | null} */
  let pathInput = null;

  $: canSubmit = !disabled && path.trim().length > 0;

  onMount(async () => {
    await tick();
    pathInput?.focus();
  });

  function submit() {
    if (!canSubmit) return;
    dispatch('submit', { value: path.trim() });
  }

  /** @param {KeyboardEvent} e */
  function onKeydown(e) {
    if (e.key === 'Enter') {
      e.preventDefault();
      submit();
    }
  }

  /** @param {DragEvent} e */
  function onDrop(e) {
    e.preventDefault();
    isDragOver = false;
    const file = e.dataTransfer?.files?.[0];
    if (file && 'path' in file) {
      // Electron/Wails drag-and-drop exposes the absolute path.
      path = /** @type {any} */ (file).path || file.name;
    } else if (file) {
      path = file.name;
    }
  }

  /** @param {DragEvent} e */
  function onDragOver(e) {
    e.preventDefault();
    isDragOver = true;
  }

  function onDragLeave() {
    isDragOver = false;
  }

  function onZoneKeydown(/** @type {KeyboardEvent} */ e) {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      pathInput?.focus();
    }
  }
</script>

<div class="file-widget">
  <!-- svelte-ignore a11y-interactive-supports-focus -->
  <div
    class="drop-zone"
    class:drag-over={isDragOver}
    class:invalid={!!validationError}
    role="button"
    tabindex="0"
    aria-label="Drop a file or press Enter to focus path input"
    on:drop={onDrop}
    on:dragover={onDragOver}
    on:dragleave={onDragLeave}
    on:keydown={onZoneKeydown}
  >
    <Upload size={20} aria-hidden="true" />
    <span class="zone-title">Drop a file or enter a path below</span>
    <span class="path-hint">Paths outside the repo root will be rejected</span>
  </div>

  <label class="path-field">
    <input
      bind:this={pathInput}
      bind:value={path}
      type="text"
      class="path-input"
      class:invalid={!!validationError}
      data-testid="file-path-input"
      placeholder="/absolute/or/repo-relative/path"
      aria-label={prompt?.prompt ?? 'File path'}
      aria-required={prompt?.required ? 'true' : 'false'}
      aria-invalid={!!validationError}
      {disabled}
      on:keydown={onKeydown}
    />
  </label>

  <div class="widget-actions">
    <button
      type="button"
      class="btn-submit"
      data-testid="file-submit"
      disabled={!canSubmit}
      on:click={submit}
    >
      <Send size={13} aria-hidden="true" />
      Submit
    </button>
  </div>
</div>

<style>
  .file-widget {
    display: flex;
    flex-direction: column;
    gap: var(--sp-md);
  }

  .drop-zone {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--sp-sm);
    padding: var(--sp-xl) var(--sp-lg);
    border: 1px dashed var(--border-emphasis);
    border-radius: var(--radius-md);
    background: var(--bg-deepest);
    color: var(--text-dim);
    text-align: center;
    cursor: pointer;
    transition:
      border-color var(--duration-short) var(--ease-enter),
      background var(--duration-short) var(--ease-enter);
  }

  .drop-zone:hover,
  .drop-zone:focus-visible {
    border-color: var(--accent-green);
    background: color-mix(in srgb, var(--accent-green) 5%, transparent);
    outline: none;
  }

  .drop-zone.drag-over {
    border-style: solid;
    border-color: var(--accent-green);
    background: color-mix(in srgb, var(--accent-green) 10%, transparent);
  }

  .drop-zone.invalid { border-color: var(--accent-red); }

  .zone-title {
    font-family: var(--font-ui);
    font-size: var(--text-body);
    color: var(--text-primary);
  }

  .path-hint {
    font-family: var(--font-mono);
    font-size: var(--text-label);
    color: var(--text-muted);
  }

  .path-field { display: flex; flex-direction: column; }

  .path-input {
    width: 100%;
    box-sizing: border-box;
    background: var(--bg-deepest);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    padding: var(--sp-sm) var(--sp-md);
    font-family: var(--font-mono);
    font-size: var(--text-body);
    color: var(--text-primary);
    outline: none;
    transition: border-color var(--duration-short) var(--ease-enter);
  }

  .path-input::placeholder { color: var(--text-muted); }
  .path-input:focus { border-color: var(--accent-green); }
  .path-input.invalid { border-color: var(--accent-red); }
  .path-input:disabled { opacity: 0.6; cursor: not-allowed; }

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
