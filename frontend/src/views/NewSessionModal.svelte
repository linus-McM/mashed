<script>
  import { createEventDispatcher } from 'svelte';
  import { Terminal } from 'lucide-svelte';
  import cliConfig from '../config/claude-cli.json';

  const dispatch = createEventDispatcher();

  export let repoPath = '';
  export let repoName = '';

  // State from config defaults
  let model = cliConfig.models.find(m => m.default)?.value || cliConfig.models[0].value;
  let permissionMode = cliConfig.permissionModes.find(p => p.default)?.value || 'default';
  let effort = cliConfig.effortLevels.find(e => e.default)?.value || 'high';
  let outputFormat = cliConfig.outputFormats.find(o => o.default)?.value || 'text';

  // Toggle flags
  let toggles = {};
  for (const t of cliConfig.toggleFlags) {
    toggles[t.flag] = t.default || false;
  }

  // Conditional flags (toggle + text value)
  let conditionalEnabled = {};
  let conditionalValues = {};
  for (const c of (cliConfig.conditionalFlags || [])) {
    conditionalEnabled[c.flag] = c.default || false;
    conditionalValues[c.flag] = '';
  }

  // Text fields
  let textValues = {};
  for (const f of cliConfig.textFields) {
    textValues[f.flag] = f.default || '';
  }

  $: command = buildCommand();

  $: isOpus = model.includes('opus');

  function buildCommand() {
    const parts = ['claude'];

    parts.push('--model', model);

    // Permission mode
    if (permissionMode === 'dontAsk') {
      parts.push('--dangerously-skip-permissions');
    } else if (permissionMode !== 'default') {
      parts.push('--permission-mode', permissionMode);
    }

    // Effort only for Opus
    if (isOpus) {
      parts.push('--effort', effort);
    }

    if (outputFormat !== 'text') {
      parts.push('--output-format', outputFormat);
    }

    // Toggle flags
    for (const t of cliConfig.toggleFlags) {
      if (toggles[t.flag]) {
        parts.push(t.flag);
      }
    }

    // Conditional flags (enabled + optional value)
    for (const c of (cliConfig.conditionalFlags || [])) {
      if (conditionalEnabled[c.flag]) {
        const val = (conditionalValues[c.flag] || '').trim();
        if (val) {
          parts.push(c.flag, val);
        } else {
          parts.push(c.flag);
        }
      }
    }

    // Text fields
    for (const f of cliConfig.textFields) {
      const val = (textValues[f.flag] || '').trim();
      if (val) {
        if (f.flag === '--allowedTools' || f.flag === '--disallowedTools') {
          for (const tool of val.split(',').map(s => s.trim()).filter(Boolean)) {
            parts.push(f.flag, `"${tool}"`);
          }
        } else if (f.multiline) {
          parts.push(f.flag, `"${val.replace(/"/g, '\\"')}"`);
        } else {
          parts.push(f.flag, val);
        }
      }
    }

    return parts.join(' ');
  }

  function spawn() {
    dispatch('spawn', { command, model, repoPath });
  }

  function cancel() {
    dispatch('cancel');
  }

  function handleKeydown(e) {
    if (e.key === 'Escape') cancel();
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<div class="overlay" role="presentation" on:click={cancel} on:keydown={() => {}}>
  <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
  <div class="modal" role="dialog" on:click|stopPropagation on:keydown={() => {}}>
    <h2><Terminal size={18} /> New Session</h2>
    <p class="subtitle">Launch Claude Code in <strong>{repoName}</strong></p>

    <!-- Model -->
    <div class="field">
      <!-- svelte-ignore a11y-label-has-associated-control -->
      <label>Model</label>
      <div class="radio-row">
        {#each cliConfig.models as m}
          <button
            class="radio-btn"
            class:selected={model === m.value}
            on:click={() => model = m.value}
          >
            <span class="radio-label">{m.label}</span>
            <span class="radio-meta">{m.context}</span>
          </button>
        {/each}
      </div>
    </div>

    <!-- Permission Mode -->
    <div class="field">
      <!-- svelte-ignore a11y-label-has-associated-control -->
      <label>Permission Mode</label>
      <div class="radio-row">
        {#each cliConfig.permissionModes as p}
          <button
            class="radio-btn"
            class:selected={permissionMode === p.value}
            on:click={() => permissionMode = p.value}
            title={p.description}
          >
            {p.label}
          </button>
        {/each}
      </div>
    </div>

    <!-- Effort (Opus only) -->
    {#if isOpus}
    <div class="field">
      <!-- svelte-ignore a11y-label-has-associated-control -->
      <label>Effort</label>
      <div class="radio-row">
        {#each cliConfig.effortLevels as e}
          <button
            class="radio-btn"
            class:selected={effort === e.value}
            on:click={() => effort = e.value}
            title={e.description}
          >
            {e.label}
          </button>
        {/each}
      </div>
    </div>
    {/if}

    <!-- Output Format -->
    <div class="field">
      <!-- svelte-ignore a11y-label-has-associated-control -->
      <label>Output Format</label>
      <div class="radio-row">
        {#each cliConfig.outputFormats as o}
          <button
            class="radio-btn"
            class:selected={outputFormat === o.value}
            on:click={() => outputFormat = o.value}
          >
            {o.label}
          </button>
        {/each}
      </div>
    </div>

    <!-- Toggle flags -->
    <div class="field">
      <!-- svelte-ignore a11y-label-has-associated-control -->
      <label>Options</label>
      <div class="toggle-row">
        {#each cliConfig.toggleFlags as t}
          <label class="toggle-item" title={t.description}>
            <input type="checkbox" bind:checked={toggles[t.flag]} />
            <span>{t.label}</span>
          </label>
        {/each}
      </div>
    </div>

    <!-- Conditional flags (toggle + text input) -->
    {#each (cliConfig.conditionalFlags || []) as c}
      <div class="field">
        <label class="toggle-item" title={c.description}>
          <input type="checkbox" bind:checked={conditionalEnabled[c.flag]} />
          <span>{c.label}</span>
        </label>
        {#if conditionalEnabled[c.flag]}
          <input
            class="text-input conditional-input"
            type="text"
            placeholder={c.placeholder}
            bind:value={conditionalValues[c.flag]}
          />
        {/if}
      </div>
    {/each}

    <!-- Text fields -->
    {#each cliConfig.textFields as f}
      <div class="field">
        <label for="text-{f.flag}">{f.label}</label>
        {#if f.multiline}
          <textarea
            id="text-{f.flag}"
            class="text-input"
            placeholder={f.placeholder}
            bind:value={textValues[f.flag]}
            rows="2"
          ></textarea>
        {:else}
          <input
            id="text-{f.flag}"
            class="text-input"
            type="text"
            placeholder={f.placeholder}
            bind:value={textValues[f.flag]}
          />
        {/if}
      </div>
    {/each}

    <!-- Command preview -->
    <div class="field">
      <!-- svelte-ignore a11y-label-has-associated-control -->
      <label>Command Preview</label>
      <pre class="command-preview">{command}</pre>
    </div>

    <div class="actions">
      <button class="btn-cancel" on:click={cancel}>Cancel</button>
      <button class="btn-spawn" on:click={spawn}>Launch Session</button>
    </div>
  </div>
</div>

<style>
  .overlay {
    position: fixed;
    inset: 0;
    background: var(--overlay-backdrop);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 100;
  }

  .modal {
    background: var(--bg-surface);
    border: 1px solid var(--border-emphasis);
    border-radius: var(--radius-lg);
    padding: 28px;
    width: 480px;
    max-height: 85vh;
    overflow-y: auto;
  }

  h2 {
    font-size: 18px;
    font-weight: 600;
    color: var(--text-primary);
    margin: 0 0 4px;
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .subtitle {
    font-size: 12px;
    color: var(--text-dim);
    margin: 0 0 20px;
  }

  .subtitle strong {
    color: var(--text-primary);
    font-family: var(--font-mono);
    font-size: 11px;
  }

  .field {
    margin-bottom: 14px;
  }

  label {
    display: block;
    font-size: 11px;
    font-weight: 600;
    color: var(--text-dim);
    text-transform: uppercase;
    letter-spacing: 0.5px;
    margin-bottom: 6px;
  }

  .radio-row {
    display: flex;
    gap: 6px;
  }

  .radio-btn {
    flex: 1;
    padding: 6px 8px;
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    cursor: pointer;
    transition: all 100ms ease-out;
    font-family: var(--font-ui);
    font-size: 12px;
    color: var(--text-primary);
    text-align: center;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 2px;
  }

  .radio-btn:hover {
    border-color: var(--border-emphasis);
    background: var(--bg-active);
  }

  .radio-btn.selected {
    border-color: var(--accent-green);
    background: color-mix(in srgb, var(--accent-green) 8%, transparent);
  }

  .radio-meta {
    font-family: var(--font-mono);
    font-size: 10px;
    color: var(--text-muted);
  }

  .toggle-row {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
  }

  .toggle-item {
    display: flex;
    align-items: center;
    gap: 6px;
    cursor: pointer;
    text-transform: none;
    letter-spacing: 0;
    font-weight: 400;
    font-size: 12px;
    color: var(--text-primary);
    margin-bottom: 0;
  }

  .toggle-item input[type="checkbox"] {
    width: 13px;
    height: 13px;
    accent-color: var(--accent-green);
    cursor: pointer;
  }

  .text-input {
    width: 100%;
    padding: 6px 10px;
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    color: var(--text-primary);
    font-family: var(--font-mono);
    font-size: 12px;
    outline: none;
    transition: border-color 100ms ease;
    box-sizing: border-box;
    resize: vertical;
  }

  .text-input:focus {
    border-color: var(--accent-green);
  }

  .text-input::placeholder {
    color: var(--text-muted);
  }

  .conditional-input {
    margin-top: 6px;
  }

  .command-preview {
    background: var(--bg-deepest);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    padding: 8px 10px;
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--accent-green);
    white-space: pre-wrap;
    word-break: break-all;
    margin: 0;
    max-height: 60px;
    overflow-y: auto;
  }

  .actions {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
    margin-top: 16px;
  }

  .btn-cancel {
    padding: 8px 16px;
    background: none;
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    color: var(--text-dim);
    font-family: var(--font-ui);
    font-size: 13px;
    cursor: pointer;
  }

  .btn-cancel:hover {
    border-color: var(--border-emphasis);
    color: var(--text-primary);
  }

  .btn-spawn {
    padding: 8px 20px;
    background: var(--accent-green);
    border: none;
    border-radius: var(--radius-md);
    color: var(--bg-deepest);
    font-family: var(--font-ui);
    font-size: 13px;
    font-weight: 600;
    cursor: pointer;
  }

  .btn-spawn:hover { opacity: 0.9; }
</style>
