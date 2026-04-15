<script>
  import { createEventDispatcher } from 'svelte';
  import { fade, slide } from 'svelte/transition';
  import { cubicOut } from 'svelte/easing';
  import { Terminal, ChevronDown } from 'lucide-svelte';
  import cliConfig from '../config/claude-cli.json';

  const dispatch = createEventDispatcher();

  // uiqa-06: modal fade entry/exit. prefers-reduced-motion zeroes durations.
  const reducedMotion = typeof window !== 'undefined' &&
    window.matchMedia('(prefers-reduced-motion: reduce)').matches;
  const backdropFadeProps = reducedMotion
    ? { duration: 0 }
    : { duration: 100, easing: cubicOut };
  const cardFadeProps = reducedMotion
    ? { duration: 0, delay: 0 }
    : { duration: 100, delay: 50, easing: cubicOut };
  const slideProps = reducedMotion
    ? { duration: 0 }
    : { duration: 150, easing: cubicOut };

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

  // Section open state — Execution + Session open by default; rest collapsed.
  let openSections = {
    execution: true,
    session: true,
    tools: false,
    advanced: false,
    output: false,
  };

  function toggleSection(key) {
    openSections[key] = !openSections[key];
  }

  // Field categorization for grouping
  $: sessionFields = cliConfig.textFields.filter(
    f => f.flag === '--name' || f.flag === '--max-turns'
  );
  $: toolsFields = cliConfig.textFields.filter(
    f => f.flag === '--allowedTools' || f.flag === '--disallowedTools' || f.flag === '--mcp-config'
  );
  $: advancedTextFields = cliConfig.textFields.filter(
    f => f.flag === '--append-system-prompt'
  );

  // Svelte reactivity gotcha: `$:` only tracks variables directly referenced
  // in the block, not vars used inside called functions. Listing every
  // reactive dep via the comma operator forces re-evaluation on any change.
  $: command = (
    model,
    permissionMode,
    effort,
    outputFormat,
    toggles,
    conditionalEnabled,
    conditionalValues,
    textValues,
    repoPath,
    buildCommand()
  );

  $: isOpus = model.includes('opus');

  function buildCommand() {
    const parts = [];

    // Full statement: cd into the repo, then launch claude.
    // Paths with spaces get quoted so the preview is copy-pasteable.
    if (repoPath) {
      const safePath = /\s/.test(repoPath) ? `"${repoPath}"` : repoPath;
      parts.push(`cd ${safePath} &&`);
    }

    parts.push('claude');
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

<div class="overlay" role="presentation" on:click={cancel} on:keydown={() => {}} transition:fade={backdropFadeProps}>
  <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
  <div class="modal" role="dialog" on:click|stopPropagation on:keydown={() => {}} transition:fade={cardFadeProps}>
    <!-- Sticky header -->
    <header class="modal-header">
      <h2><Terminal size={18} /> New Session</h2>
      <p class="subtitle">Launch Claude Code in <strong>{repoName}</strong></p>
    </header>

    <!-- Scrollable body with sectioned groups -->
    <div class="modal-body">
      <!-- ── Model & Execution ─────────────────────────── -->
      <section class="group">
        <button
          class="group-header"
          aria-expanded={openSections.execution}
          on:click={() => toggleSection('execution')}
        >
          <span class="chevron"><ChevronDown size={14} /></span>
          <span class="group-title">Model &amp; Execution</span>
        </button>
        {#if openSections.execution}
          <div class="group-body" transition:slide={slideProps}>
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

            {#if isOpus}
              <div class="field" transition:slide={slideProps}>
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

            <div class="field">
              <!-- svelte-ignore a11y-label-has-associated-control -->
              <label>Permission Mode</label>
              <div class="chip-row">
                {#each cliConfig.permissionModes as p}
                  <button
                    class="chip"
                    class:selected={permissionMode === p.value}
                    on:click={() => permissionMode = p.value}
                    title={p.description}
                  >
                    {p.label}
                  </button>
                {/each}
              </div>
            </div>
          </div>
        {/if}
      </section>

      <!-- ── Session ───────────────────────────────────── -->
      <section class="group">
        <button
          class="group-header"
          aria-expanded={openSections.session}
          on:click={() => toggleSection('session')}
        >
          <span class="chevron"><ChevronDown size={14} /></span>
          <span class="group-title">Session</span>
        </button>
        {#if openSections.session}
          <div class="group-body" transition:slide={slideProps}>
            <div class="field-grid-2">
              {#each sessionFields as f}
                <div class="field">
                  <label for="text-{f.flag}">{f.label}</label>
                  <input
                    id="text-{f.flag}"
                    class="text-input"
                    type="text"
                    placeholder={f.placeholder}
                    bind:value={textValues[f.flag]}
                  />
                </div>
              {/each}
            </div>
          </div>
        {/if}
      </section>

      <!-- ── Tools & MCP ───────────────────────────────── -->
      <section class="group">
        <button
          class="group-header"
          aria-expanded={openSections.tools}
          on:click={() => toggleSection('tools')}
        >
          <span class="chevron"><ChevronDown size={14} /></span>
          <span class="group-title">Tools &amp; MCP</span>
        </button>
        {#if openSections.tools}
          <div class="group-body" transition:slide={slideProps}>
            <div class="field-grid-2">
              {#each toolsFields.filter(f => f.flag !== '--mcp-config') as f}
                <div class="field">
                  <label for="text-{f.flag}">{f.label}</label>
                  <input
                    id="text-{f.flag}"
                    class="text-input"
                    type="text"
                    placeholder={f.placeholder}
                    bind:value={textValues[f.flag]}
                  />
                </div>
              {/each}
            </div>
            {#each toolsFields.filter(f => f.flag === '--mcp-config') as f}
              <div class="field">
                <label for="text-{f.flag}">{f.label}</label>
                <input
                  id="text-{f.flag}"
                  class="text-input"
                  type="text"
                  placeholder={f.placeholder}
                  bind:value={textValues[f.flag]}
                />
              </div>
            {/each}
          </div>
        {/if}
      </section>

      <!-- ── Advanced ──────────────────────────────────── -->
      <section class="group">
        <button
          class="group-header"
          aria-expanded={openSections.advanced}
          on:click={() => toggleSection('advanced')}
        >
          <span class="chevron"><ChevronDown size={14} /></span>
          <span class="group-title">Advanced</span>
        </button>
        {#if openSections.advanced}
          <div class="group-body" transition:slide={slideProps}>
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

            <!-- System prompt (multiline) -->
            {#each advancedTextFields as f}
              <div class="field">
                <label for="text-{f.flag}">{f.label}</label>
                <textarea
                  id="text-{f.flag}"
                  class="text-input"
                  placeholder={f.placeholder}
                  bind:value={textValues[f.flag]}
                  rows="2"
                ></textarea>
              </div>
            {/each}

            <!-- Toggle flags -->
            <div class="field">
              <!-- svelte-ignore a11y-label-has-associated-control -->
              <label>Flags</label>
              <div class="toggle-grid">
                {#each cliConfig.toggleFlags as t}
                  <label class="toggle-item" title={t.description}>
                    <input type="checkbox" bind:checked={toggles[t.flag]} />
                    <span>{t.label}</span>
                  </label>
                {/each}
              </div>
            </div>
          </div>
        {/if}
      </section>

      <!-- ── Output ────────────────────────────────────── -->
      <section class="group">
        <button
          class="group-header"
          aria-expanded={openSections.output}
          on:click={() => toggleSection('output')}
        >
          <span class="chevron"><ChevronDown size={14} /></span>
          <span class="group-title">Output</span>
        </button>
        {#if openSections.output}
          <div class="group-body" transition:slide={slideProps}>
            <div class="field">
              <!-- svelte-ignore a11y-label-has-associated-control -->
              <label>Format</label>
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
          </div>
        {/if}
      </section>
    </div>

    <!-- Sticky footer: command preview + actions -->
    <footer class="modal-footer">
      <div class="command-preview-pane">
        <pre class="command-preview"><span class="command-prompt">$ </span>{command}</pre>
      </div>
      <div class="actions">
        <button class="btn-cancel" on:click={cancel}>Cancel</button>
        <button class="btn-spawn" on:click={spawn}>Launch Session</button>
      </div>
    </footer>
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
    width: 560px;
    max-height: min(720px, 85vh);
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  /* ── Header ─────────────────────────────────────────── */
  .modal-header {
    padding: 20px 24px 14px;
    border-bottom: 1px solid var(--border-subtle);
    flex-shrink: 0;
  }

  h2 {
    font-size: var(--text-section);
    font-weight: 600;
    color: var(--text-primary);
    margin: 0 0 4px;
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .subtitle {
    font-size: var(--text-body);
    color: var(--text-dim);
    margin: 0;
  }

  .subtitle strong {
    color: var(--text-primary);
    font-family: var(--font-mono);
    font-size: var(--text-label);
  }

  /* ── Scrollable body ────────────────────────────────── */
  .modal-body {
    padding: 0 24px;
    overflow-y: auto;
    flex: 1;
  }

  /* ── Section groups ─────────────────────────────────── */
  .group {
    border-top: 1px solid var(--border-subtle);
  }
  .group:first-of-type {
    border-top: none;
  }

  .group-header {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 14px 0 12px;
    background: none;
    border: none;
    width: 100%;
    cursor: pointer;
    color: var(--text-dim);
    font-family: var(--font-ui);
    font-size: 11px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    text-align: left;
    transition: color 100ms ease;
  }

  .group-header:hover {
    color: var(--text-primary);
  }

  .chevron {
    display: inline-flex;
    transition: transform 150ms var(--ease, ease);
    color: var(--text-muted);
  }

  .group-header[aria-expanded="false"] .chevron {
    transform: rotate(-90deg);
  }

  .group-body {
    padding-bottom: 14px;
  }

  /* ── Fields ─────────────────────────────────────────── */
  .field {
    margin-bottom: 12px;
  }
  .field:last-child {
    margin-bottom: 0;
  }

  .field-grid-2 {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 12px;
  }
  .field-grid-2 .field {
    margin-bottom: 0;
  }

  label {
    display: block;
    font-size: 11px;
    font-weight: 600;
    color: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: 0.4px;
    margin-bottom: var(--sp-sm);
  }

  /* ── Radio row (Model, Effort, Output Format) ───────── */
  .radio-row {
    display: flex;
    gap: var(--sp-sm);
  }

  .radio-btn {
    flex: 1;
    padding: var(--sp-xs) var(--sp-sm);
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    cursor: pointer;
    transition: all 100ms ease-out;
    font-family: var(--font-ui);
    font-size: var(--text-body);
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
    font-size: var(--text-label);
    color: var(--text-muted);
  }

  /* ── Chip row (Permission Mode) ─────────────────────── */
  .chip-row {
    display: flex;
    flex-wrap: wrap;
    gap: var(--sp-sm);
  }

  .chip {
    padding: var(--sp-sm) var(--sp-md);
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    cursor: pointer;
    font-family: var(--font-ui);
    font-size: var(--text-body);
    color: var(--text-primary);
    transition: all 100ms ease-out;
  }

  .chip:hover {
    border-color: var(--border-emphasis);
    background: var(--bg-active);
  }

  .chip.selected {
    border-color: var(--accent-green);
    background: color-mix(in srgb, var(--accent-green) 8%, transparent);
    color: var(--accent-green);
  }

  /* ── Toggle grid ────────────────────────────────────── */
  .toggle-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 8px 16px;
  }

  .toggle-item {
    display: flex;
    align-items: center;
    gap: var(--sp-sm);
    cursor: pointer;
    text-transform: none;
    letter-spacing: 0;
    font-weight: 400;
    font-size: var(--text-body);
    color: var(--text-primary);
    margin-bottom: 0;
  }

  .toggle-item input[type="checkbox"] {
    width: 13px;
    height: 13px;
    accent-color: var(--accent-green);
    cursor: pointer;
  }

  /* ── Text inputs ────────────────────────────────────── */
  .text-input {
    width: 100%;
    padding: var(--sp-xs) 10px;
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    color: var(--text-primary);
    font-family: var(--font-mono);
    font-size: var(--text-body);
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
    margin-top: var(--sp-xs);
  }

  /* ── Sticky footer ──────────────────────────────────── */
  .modal-footer {
    border-top: 1px solid var(--border-subtle);
    flex-shrink: 0;
    background: var(--bg-deepest);
  }

  .command-preview-pane {
    padding: 12px 16px 10px;
    border-bottom: 1px solid var(--border-subtle);
  }

  .command-preview {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--accent-green);
    white-space: pre-wrap;
    word-break: break-all;
    margin: 0;
    max-height: 72px;
    overflow-y: auto;
    line-height: 1.55;
  }

  .command-prompt {
    color: var(--text-muted);
    user-select: none;
  }

  .actions {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
    padding: 12px 24px;
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
    transition: all 100ms ease;
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
    transition: opacity 100ms ease;
  }

  .btn-spawn:hover { opacity: 0.9; }
</style>
