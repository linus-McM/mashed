<script>
  import { onMount, createEventDispatcher } from 'svelte';
  import { X, Terminal, FileText, FolderOpen, List, Plus, Minus } from 'lucide-svelte';
  import { PickFile, ReadFile, ListModels } from '../../../wailsjs/go/main/App.js';
  import { parseFriendlyTarget } from '../../lib/bmadSessionName';
  import {
    parseEntries,
    stringifyEntries,
    hasDuplicateLabels,
  } from '../../lib/bmad/multiFileEntries';

  export let node = null;
  export let groupedAgents = { bmadAgents: [], localAgents: [], globalAgents: [] };

  const dispatch = createEventDispatcher();

  // Resize logic — exported so parent can read current width
  export let panelWidth = 280;
  let resizing = false;

  function onResizeStart(e) {
    e.preventDefault();
    resizing = true;
    const startX = e.clientX;
    const startWidth = panelWidth;

    function onMouseMove(e) {
      panelWidth = Math.max(220, Math.min(500, startWidth - (e.clientX - startX)));
    }

    function onMouseUp() {
      resizing = false;
      window.removeEventListener('mousemove', onMouseMove);
      window.removeEventListener('mouseup', onMouseUp);
    }

    window.addEventListener('mousemove', onMouseMove);
    window.addEventListener('mouseup', onMouseUp);
  }

  let models = [{ value: '', label: 'Default (inherit)' }];

  onMount(async () => {
    try {
      const modelList = await ListModels();
      models = [
        { value: '', label: 'Default (inherit)' },
        ...(modelList || []).map(m => ({ value: m.id, label: m.displayName })),
      ];
    } catch (e) {
      console.error('Failed to load models:', e);
    }
  });

  let modelOverride = '';
  let customContext = '';
  let selectedAgent = '';

  // Deduplicate agents: BMAD agents take priority, then local, then global.
  $: dedupedAgents = (() => {
    const seen = new Set();
    const bmad = (groupedAgents.bmadAgents || []).map(a => {
      seen.add(a.id);
      return a;
    });
    const local = (groupedAgents.localAgents || []).filter(a => {
      if (seen.has(a.id)) return false;
      seen.add(a.id);
      return true;
    });
    const global = (groupedAgents.globalAgents || []).filter(a => {
      if (seen.has(a.id)) return false;
      seen.add(a.id);
      return true;
    });
    return { bmad, local, global };
  })();

  $: nodeType = node?.data?.nodeType || '';
  $: isProcessNode = !nodeType || nodeType === 'process';
  $: isFileLoader = node?.data?.processId === 'util-file-loader';
  $: isMultiFileLoader = nodeType === 'multiFileLoader';
  // Reorder is deferred — ArrayEditorModal only handles string arrays and
  // the MultiFileLoader entry editor inlines add/edit/delete only. Upstream
  // story AC-2 mentions "reorder" which we are flagging as a follow-up.
  let mflEntries = [];
  $: mflDuplicates = hasDuplicateLabels(mflEntries);

  let filePath = '';
  let filePreview = '';
  let fileError = '';

  // Load preview when filePath changes
  $: if (isFileLoader && filePath) {
    ReadFile(filePath).then(content => {
      fileError = '';
      // Show first 50 lines, truncated
      const lines = content.split('\n');
      filePreview = lines.slice(0, 50).join('\n');
      if (lines.length > 50) filePreview += '\n... (' + lines.length + ' lines total)';
    }).catch(err => {
      filePreview = '';
      fileError = String(err);
    });
  } else if (isFileLoader) {
    filePreview = '';
    fileError = '';
  }


  const conditionTypeOptions = [
    { value: 'contains', label: 'Contains' },
    { value: 'notContains', label: 'Not Contains' },
    { value: 'regex', label: 'Regex' },
    { value: 'exitCode', label: 'Exit Code' },
    { value: 'fileExists', label: 'File Exists' },
    { value: 'always', label: 'Always' },
  ];

  let conditionType = 'contains';
  let conditionPattern = '';
  let sourceNode = '';
  let maxIterations = '10';
  let items = [];
  let extractType = 'regex';
  let extractPattern = '';

  $: if (node) {
    const cfg = node.data?.config || {};
    modelOverride = cfg.model || '';
    customContext = cfg.context || '';
    selectedAgent = cfg.agentId || '';
    conditionType = cfg.conditionType || 'contains';
    conditionPattern = cfg.conditionPattern || '';
    sourceNode = cfg.sourceNode || '';
    maxIterations = cfg.maxIterations || '10';
    try { items = cfg.items ? JSON.parse(cfg.items) : []; } catch { items = []; }
    extractType = cfg.extractType || 'regex';
    extractPattern = cfg.extractPattern || '';
    filePath = cfg.filePath || '';
    mflEntries = parseEntries(cfg.entries || '[]');
  }

  $: label = node?.data?.label || 'Node';
  $: status = node?.data?.status || 'pending';
  $: tmuxTarget = node?.data?.tmuxTarget || '';
  $: hasTerminal = status === 'running' && tmuxTarget;
  $: parsedTmuxTarget = parseFriendlyTarget(tmuxTarget);

  async function browseFile() {
    try {
      const path = await PickFile('Select a file');
      if (path) {
        filePath = path;
        emitUpdate();
      }
    } catch (e) {
      console.error('File picker failed:', e);
    }
  }

  function emitUpdate() {
    const config = isMultiFileLoader
      ? { entries: stringifyEntries(mflEntries) }
      : isFileLoader
      ? { filePath }
      : isProcessNode
      ? { model: modelOverride, context: customContext, agentId: selectedAgent }
      : nodeType === 'condition'
      ? { conditionType, conditionPattern, sourceNode }
      : nodeType === 'loop'
      ? { maxIterations, items: items.length > 0 ? JSON.stringify(items) : '' }
      : nodeType === 'loopUntil'
      ? { maxIterations, conditionType, conditionPattern, sourceNode, items: items.length > 0 ? JSON.stringify(items) : '' }
      : nodeType === 'transform'
      ? { extractType, extractPattern, sourceNode }
      : {};
    dispatch('update', { nodeId: node.id, config });
  }

  function addMflEntry() {
    mflEntries = [...mflEntries, { label: '', path: '' }];
    emitUpdate();
  }

  function removeMflEntry(index) {
    mflEntries = mflEntries.filter((_, i) => i !== index);
    emitUpdate();
  }

  function updateMflEntry(index, field, value) {
    mflEntries = mflEntries.map((e, i) => (i === index ? { ...e, [field]: value } : e));
    emitUpdate();
  }

  async function browseMflEntry(index) {
    try {
      const path = await PickFile('Select a file');
      if (path) updateMflEntry(index, 'path', path);
    } catch (e) {
      console.error('File picker failed:', e);
    }
  }

  function close() {
    dispatch('close');
  }
</script>

{#if node}
  <div class="config-panel" class:visible={!!node} style="width: {panelWidth}px;">
    <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
    <div class="resize-handle" class:active={resizing} on:mousedown={onResizeStart} role="separator" />
    <div class="panel-header">
      <span class="panel-title">Configure: {label}</span>
      <button class="close-btn" on:click={close} title="Close">
        <X size={14} />
      </button>
    </div>

    <div class="panel-body">
      {#if isMultiFileLoader}
        <div class="field">
          <label class="field-label">Entries</label>
          <div class="mfl-entries">
            {#each mflEntries as entry, i (i)}
              <div class="mfl-row" class:dup-row={entry.label && mflDuplicates.has(entry.label)}>
                <input
                  class="field-input mfl-label"
                  type="text"
                  value={entry.label}
                  on:input={(e) => updateMflEntry(i, 'label', e.target.value)}
                  on:blur={emitUpdate}
                  placeholder={`file[${i}]`}
                />
                <input
                  class="field-input mfl-path"
                  type="text"
                  value={entry.path}
                  on:input={(e) => updateMflEntry(i, 'path', e.target.value)}
                  on:blur={emitUpdate}
                  placeholder="/absolute/path"
                />
                <button class="mfl-btn" on:click={() => browseMflEntry(i)} title="Browse file">
                  <FolderOpen size={12} />
                </button>
                <button class="mfl-btn remove" on:click={() => removeMflEntry(i)} title="Remove entry">
                  <Minus size={12} />
                </button>
              </div>
            {/each}
          </div>
          <button class="browse-btn" on:click={addMflEntry}>
            <Plus size={13} />
            Add Entry
          </button>
        </div>
      {:else if isFileLoader}
        <div class="field">
          <label class="field-label">File Path</label>
          <div class="file-picker-row">
            <input
              class="field-input file-path-input"
              type="text"
              bind:value={filePath}
              on:blur={emitUpdate}
              placeholder="No file selected..."
              readonly
            />
          </div>
          <button class="browse-btn" on:click={browseFile}>
            <FolderOpen size={14} />
            Browse...
          </button>
        </div>
        {#if filePreview}
          <div class="field">
            <label class="field-label">Preview</label>
            <pre class="file-preview">{filePreview}</pre>
          </div>
        {/if}
        {#if fileError}
          <div class="field">
            <span class="file-error">{fileError}</span>
          </div>
        {/if}
      {:else if isProcessNode}
        <div class="field">
          <label class="field-label" for="model-override">Model</label>
          <select id="model-override" class="field-select" bind:value={modelOverride} on:change={emitUpdate}>
            {#each models as m}
              <option value={m.value}>{m.label}</option>
            {/each}
          </select>
        </div>

        <div class="field">
          <label class="field-label" for="custom-context">Context</label>
          <textarea
            id="custom-context"
            class="field-textarea"
            bind:value={customContext}
            on:blur={emitUpdate}
            placeholder="Additional context for this process..."
            rows="4"
          />
        </div>

        <div class="field">
          <label class="field-label" for="agent-select">Agent</label>
          <select id="agent-select" class="field-select" bind:value={selectedAgent} on:change={emitUpdate}>
            <option value="">Default</option>
            {#if dedupedAgents.bmad.length > 0}
              <optgroup label="BMAD Agents">
                {#each dedupedAgents.bmad as agent}
                  <option value={agent.id}>{agent.name} ({agent.role})</option>
                {/each}
              </optgroup>
            {/if}
            {#if dedupedAgents.local.length > 0}
              <optgroup label="Local Project Agents">
                {#each dedupedAgents.local as agent}
                  <option value={agent.id}>{agent.name}</option>
                {/each}
              </optgroup>
            {/if}
            {#if dedupedAgents.global.length > 0}
              <optgroup label="Global Agents">
                {#each dedupedAgents.global as agent}
                  <option value={agent.id}>{agent.name}</option>
                {/each}
              </optgroup>
            {/if}
          </select>
        </div>
      {:else if nodeType === 'condition'}
        <div class="field">
          <label class="field-label" for="cond-type">Condition Type</label>
          <select id="cond-type" class="field-select" bind:value={conditionType} on:change={emitUpdate}>
            {#each conditionTypeOptions as opt}
              <option value={opt.value}>{opt.label}</option>
            {/each}
          </select>
        </div>
        <div class="field">
          <label class="field-label" for="cond-pattern">Pattern</label>
          <input id="cond-pattern" class="field-input" type="text" bind:value={conditionPattern} on:blur={emitUpdate} placeholder="Pattern to match..." />
        </div>
        <div class="field">
          <label class="field-label" for="cond-source">Source Node</label>
          <input id="cond-source" class="field-input" type="text" bind:value={sourceNode} on:blur={emitUpdate} placeholder="Node ID..." />
        </div>
      {:else if nodeType === 'loop'}
        <div class="field">
          <label class="field-label" for="loop-max">Max Iterations</label>
          <input id="loop-max" class="field-input" type="number" bind:value={maxIterations} on:change={emitUpdate} min="1" max="100" />
        </div>
        <div class="field">
          <label class="field-label">Items Array</label>
          {#if items.length > 0}
            <div class="items-preview">{items.length} item{items.length !== 1 ? 's' : ''}</div>
          {:else}
            <div class="items-preview empty">No items — loop uses counter</div>
          {/if}
          <button class="browse-btn" on:click={() => dispatch('edit-items', { nodeId: node.id, items })}>
            <List size={14} />
            Edit Items...
          </button>
        </div>
      {:else if nodeType === 'loopUntil'}
        <div class="field">
          <label class="field-label" for="lu-max">Max Iterations</label>
          <input id="lu-max" class="field-input" type="number" bind:value={maxIterations} on:change={emitUpdate} min="1" max="100" />
        </div>
        <div class="field">
          <label class="field-label" for="lu-cond-type">Condition Type</label>
          <select id="lu-cond-type" class="field-select" bind:value={conditionType} on:change={emitUpdate}>
            {#each conditionTypeOptions as opt}
              <option value={opt.value}>{opt.label}</option>
            {/each}
          </select>
        </div>
        <div class="field">
          <label class="field-label" for="lu-pattern">Pattern</label>
          <input id="lu-pattern" class="field-input" type="text" bind:value={conditionPattern} on:blur={emitUpdate} placeholder="Pattern to match..." />
        </div>
        <div class="field">
          <label class="field-label" for="lu-source">Source Node</label>
          <input id="lu-source" class="field-input" type="text" bind:value={sourceNode} on:blur={emitUpdate} placeholder="Node ID..." />
        </div>
        <div class="field">
          <label class="field-label">Items Array</label>
          {#if items.length > 0}
            <div class="items-preview">{items.length} item{items.length !== 1 ? 's' : ''}</div>
          {:else}
            <div class="items-preview empty">No items — loop uses counter</div>
          {/if}
          <button class="browse-btn" on:click={() => dispatch('edit-items', { nodeId: node.id, items })}>
            <List size={14} />
            Edit Items...
          </button>
        </div>
      {:else if nodeType === 'transform'}
        <div class="field">
          <label class="field-label" for="tx-extract-type">Extract Type</label>
          <select id="tx-extract-type" class="field-select" bind:value={extractType} on:change={emitUpdate}>
            <option value="regex">Regex</option>
            <option value="lines">Lines</option>
          </select>
        </div>
        <div class="field">
          <label class="field-label" for="tx-extract-pattern">Extract Pattern</label>
          <input id="tx-extract-pattern" class="field-input" type="text" bind:value={extractPattern} on:blur={emitUpdate} placeholder="Regex or line range..." />
        </div>
        <div class="field">
          <label class="field-label" for="tx-source">Source Node</label>
          <input id="tx-source" class="field-input" type="text" bind:value={sourceNode} on:blur={emitUpdate} placeholder="Node ID..." />
        </div>
      {:else if nodeType === 'merge'}
        <div class="field">
          <span class="field-label" style="color: var(--text-muted)">No configuration needed</span>
        </div>
      {/if}

      {#if node?.data?.storyId}
        <div class="field">
          <span class="field-label">Linked Story</span>
          <div class="story-link">
            <span class="story-link-id">{node.data.storyId}</span>
            <span class="story-link-status">{node.data.storyStatus || 'unknown'}</span>
          </div>
        </div>
      {/if}

      {#if status === 'complete' && isProcessNode && node?.data?.artifactStatus}
        {@const artifacts = node.data.artifactStatus}
        {#if artifacts.found?.length > 0 || artifacts.missing?.length > 0}
          <div class="field">
            <span class="field-label">Artifacts</span>
            <div class="artifact-list">
              {#each artifacts.found || [] as name}
                <div class="artifact-item found">
                  <span class="artifact-badge found-badge">found</span>
                  <span class="artifact-name">{name}</span>
                </div>
              {/each}
              {#each artifacts.missing || [] as name}
                <div class="artifact-item missing">
                  <span class="artifact-badge missing-badge">missing</span>
                  <span class="artifact-name">{name}</span>
                </div>
              {/each}
            </div>
          </div>
        {/if}
      {/if}

      {#if hasTerminal}
        <button
          class="terminal-btn"
          data-target={tmuxTarget}
          title={parsedTmuxTarget ? `${parsedTmuxTarget.repo} / ${parsedTmuxTarget.branch} / ${parsedTmuxTarget.label}` : tmuxTarget}
          on:click={() => dispatch('open-terminal', tmuxTarget)}
        >
          <Terminal size={13} />
          View Terminal
        </button>
      {/if}

      {#if status === 'complete'}
        <button class="terminal-btn output-btn" on:click={() => dispatch('open-output', node.id)}>
          <FileText size={13} />
          View Output
        </button>
      {/if}
    </div>
  </div>
{/if}

<style>
  .config-panel {
    position: absolute;
    top: 0;
    right: 0;
    height: 100%;
    background: var(--bg-surface);
    border-left: 1px solid var(--border-subtle);
    display: flex;
    flex-direction: column;
    z-index: 10;
    box-shadow: -4px 0 16px rgba(0, 0, 0, 0.2);
    animation: slide-in 150ms ease-out;
  }

  .resize-handle {
    position: absolute;
    top: 0;
    left: -3px;
    width: 6px;
    height: 100%;
    cursor: col-resize;
    z-index: 11;
    transition: background 150ms ease;
  }

  .resize-handle:hover,
  .resize-handle.active {
    background: var(--accent-green);
    opacity: 0.5;
  }

  @keyframes slide-in {
    from { transform: translateX(100%); }
    to { transform: translateX(0); }
  }

  .panel-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 10px 12px;
    border-bottom: 1px solid var(--border-subtle);
    flex-shrink: 0;
  }

  .panel-title {
    font-family: var(--font-mono);
    font-size: var(--text-body);
    font-weight: 600;
    color: var(--text-primary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .close-btn {
    background: none;
    border: none;
    color: var(--text-muted);
    cursor: pointer;
    padding: 2px;
    border-radius: var(--radius-sm);
    display: flex;
    align-items: center;
    transition: color 100ms ease;
  }

  .close-btn:hover { color: var(--text-primary); }

  .panel-body {
    flex: 1;
    overflow-y: auto;
    padding: 12px;
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .field {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .field-label {
    font-family: var(--font-mono);
    font-size: var(--text-label);
    font-weight: 600;
    color: var(--text-dim);
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .field-select {
    padding: var(--sp-xs) var(--sp-sm);
    background: var(--bg-deepest);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-sm);
    color: var(--text-primary);
    font-family: var(--font-mono);
    font-size: 11px;
    outline: none;
  }

  .field-select:focus { border-color: var(--accent-green); }

  .field-select optgroup {
    font-weight: 600;
    font-size: var(--text-label);
    text-transform: uppercase;
    letter-spacing: 0.3px;
    color: var(--text-dim);
    padding-top: 4px;
  }

  .field-select optgroup option {
    font-weight: 400;
    font-size: 11px;
    text-transform: none;
    letter-spacing: normal;
    color: var(--text-primary);
  }

  .field-input {
    padding: var(--sp-xs) var(--sp-sm);
    background: var(--bg-deepest);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-sm);
    color: var(--text-primary);
    font-family: var(--font-mono);
    font-size: 11px;
    outline: none;
  }

  .field-input:focus { border-color: var(--accent-green); }

  .field-textarea {
    padding: var(--sp-xs) var(--sp-sm);
    background: var(--bg-deepest);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-sm);
    color: var(--text-primary);
    font-family: var(--font-mono);
    font-size: 11px;
    outline: none;
    resize: vertical;
    min-height: 60px;
  }

  .field-textarea:focus { border-color: var(--accent-green); }

  .field-textarea::placeholder { color: var(--text-muted); }

  .browse-btn {
    display: flex;
    align-items: center;
    gap: var(--sp-sm);
    width: 100%;
    padding: 8px 12px;
    margin-top: var(--sp-xs);
    background: var(--bg-elevated);
    border: 1px solid var(--accent-teal, #00c4b3);
    border-radius: var(--radius-sm);
    color: var(--accent-teal, #00c4b3);
    font-family: var(--font-ui);
    font-size: var(--text-body);
    font-weight: 500;
    cursor: pointer;
    transition: background 100ms ease;
  }
  .browse-btn:hover {
    background: var(--bg-active);
  }
  .file-path-input {
    font-size: 11px;
    color: var(--text-dim);
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .file-preview {
    background: var(--bg-deepest);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-sm);
    padding: var(--sp-xs) var(--sp-sm);
    font-family: var(--font-mono);
    font-size: 9px;
    line-height: 1.4;
    color: var(--text-dim);
    max-height: 200px;
    overflow-y: auto;
    white-space: pre;
    margin: 0;
  }
  .file-error {
    font-size: var(--text-label);
    color: var(--accent-red);
  }
  .items-preview {
    font-family: var(--font-mono);
    font-size: var(--text-label);
    color: var(--text-primary);
    padding: 4px 8px;
    background: var(--bg-deepest);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-sm);
  }

  .items-preview.empty {
    color: var(--text-muted);
  }

  .terminal-btn {
    display: flex;
    align-items: center;
    gap: var(--sp-sm);
    padding: var(--sp-xs) 10px;
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-sm);
    color: var(--accent-green);
    font-family: var(--font-mono);
    font-size: 11px;
    cursor: pointer;
    transition: background 100ms ease, border-color 100ms ease;
    margin-top: 4px;
  }

  .terminal-btn:hover {
    background: var(--bg-active);
    border-color: var(--accent-green);
  }

  .output-btn {
    color: var(--accent-blue, #3d9eff);
  }

  .output-btn:hover {
    border-color: var(--accent-blue, #3d9eff);
  }

  .artifact-list {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .artifact-item {
    display: flex;
    align-items: center;
    gap: var(--sp-sm);
    padding: var(--sp-2xs) var(--sp-xs);
    background: var(--bg-deepest);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-sm);
    font-family: var(--font-mono);
    font-size: var(--text-label);
  }

  .artifact-badge {
    font-size: 8px;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.3px;
    padding: 1px 4px;
    border-radius: 3px;
    flex-shrink: 0;
  }

  .found-badge {
    background: color-mix(in srgb, var(--accent-green) 15%, transparent);
    color: var(--accent-green, #00e57a);
  }

  .missing-badge {
    background: color-mix(in srgb, var(--accent-amber) 15%, transparent);
    color: var(--accent-amber, #f0a500);
  }

  .artifact-name {
    color: var(--text-primary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .story-link {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 4px var(--sp-xs);
    background: var(--bg-deepest);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-sm);
    font-family: var(--font-mono);
    font-size: var(--text-label);
  }

  .story-link-id { color: var(--text-primary); }

  .story-link-status {
    color: var(--text-muted);
    text-transform: uppercase;
    font-size: 9px;
  }

  .mfl-entries {
    display: flex;
    flex-direction: column;
    gap: var(--sp-xs);
    margin-bottom: var(--sp-xs);
  }

  .mfl-row {
    display: flex;
    gap: var(--sp-xs);
    align-items: center;
    padding: var(--sp-2xs);
    border: 1px solid transparent;
    border-radius: var(--radius-sm);
  }

  .mfl-row.dup-row {
    border-color: var(--accent-red);
  }

  .mfl-label { width: 60px; flex-shrink: 0; }
  .mfl-path { flex: 1; min-width: 0; }

  .mfl-btn {
    background: none;
    border: 1px solid var(--border-subtle);
    color: var(--text-muted);
    padding: var(--sp-2xs);
    border-radius: var(--radius-sm);
    display: flex;
    align-items: center;
    cursor: pointer;
    flex-shrink: 0;
    transition: color 100ms ease, border-color 100ms ease;
  }

  .mfl-btn:hover {
    color: var(--text-primary);
    border-color: var(--border-emphasis);
  }

  .mfl-btn.remove:hover {
    color: var(--accent-red);
    border-color: var(--accent-red);
  }
</style>
