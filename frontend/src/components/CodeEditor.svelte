<script>
  import { onMount, onDestroy } from 'svelte';
  import { ReadFile, ReadFileDiff, WriteFile, ExplainDiffHunk, IsExplainAvailable } from '../../wailsjs/go/main/App.js';

  export let filePath = '';
  export let repoPath = '';
  export let mode = 'source'; // 'source' or 'diff'
  export let editable = false;

  let content = '';
  let editContent = '';
  let loading = true;
  let error = '';
  let saving = false;
  let saveStatus = ''; // '', 'saving', 'saved'
  let saveTimer = null;
  let statusTimer = null;
  let isEditing = false;
  let tooltipVisible = false;
  let tooltipX = 0;
  let tooltipY = 0;
  let tooltipText = '';
  let tooltipLoading = false;
  let tooltipError = '';
  let hoverTimer = null;
  let explainAvailable = false;
  let explainCache = new Map();

  $: if (filePath && repoPath) loadFile(filePath, repoPath, mode);
  $: fullPath = filePath.startsWith('/') ? filePath : repoPath + '/' + filePath;

  async function loadFile(fp, rp, m) {
    dismissTooltip();
    explainCache = new Map();
    loading = true;
    error = '';
    content = '';
    editContent = '';
    isEditing = false;
    saveStatus = '';
    try {
      if (m === 'diff') {
        content = await ReadFileDiff(rp, fp);
        if (!content) {
          const p = fp.startsWith('/') ? fp : rp + '/' + fp;
          content = await ReadFile(p);
        }
      } else {
        const p = fp.startsWith('/') ? fp : rp + '/' + fp;
        content = await ReadFile(p);
      }
      editContent = content;
    } catch (e) {
      error = e?.message || 'Failed to load file';
    } finally {
      loading = false;
    }
  }

  function startEditing() {
    if (!editable || mode === 'diff') return;
    isEditing = true;
    editContent = content;
  }

  function handleInput(e) {
    editContent = e.target.value;
    scheduleSave();
  }

  function scheduleSave() {
    if (saveTimer) clearTimeout(saveTimer);
    saveStatus = '';
    saveTimer = setTimeout(() => doSave(), 800);
  }

  async function doSave() {
    if (!editable || saving) return;
    saving = true;
    saveStatus = 'saving';
    try {
      await WriteFile(fullPath, editContent);
      content = editContent;
      saveStatus = 'saved';
      if (statusTimer) clearTimeout(statusTimer);
      statusTimer = setTimeout(() => { saveStatus = ''; }, 2000);
    } catch (e) {
      saveStatus = 'error';
      console.error('Auto-save failed:', e);
    } finally {
      saving = false;
    }
  }

  function handleKeydown(e) {
    // Handle Tab key for indentation
    if (e.key === 'Tab') {
      e.preventDefault();
      const ta = e.target;
      const start = ta.selectionStart;
      const end = ta.selectionEnd;
      editContent = editContent.substring(0, start) + '\t' + editContent.substring(end);
      // Restore cursor position after Svelte updates the textarea
      requestAnimationFrame(() => {
        ta.selectionStart = ta.selectionEnd = start + 1;
      });
      scheduleSave();
    }
    // Cmd/Ctrl+S to save immediately
    if ((e.metaKey || e.ctrlKey) && e.key === 's') {
      e.preventDefault();
      if (saveTimer) clearTimeout(saveTimer);
      doSave();
    }
  }

  function getLanguage(path) {
    const ext = path.split('.').pop()?.toLowerCase();
    const map = {
      go: 'go', js: 'javascript', ts: 'typescript', tsx: 'tsx', jsx: 'jsx',
      svelte: 'svelte', css: 'css', html: 'html', json: 'json', md: 'markdown',
      yaml: 'yaml', yml: 'yaml', toml: 'toml', sh: 'bash', bash: 'bash',
      py: 'python', rs: 'rust', sql: 'sql', mod: 'go',
    };
    return map[ext] || 'text';
  }

  function isDiffLine(line) {
    if (line.startsWith('+') && !line.startsWith('+++')) return 'added';
    if (line.startsWith('-') && !line.startsWith('---')) return 'removed';
    if (line.startsWith('@@')) return 'hunk';
    if (line.startsWith('diff ') || line.startsWith('index ')) return 'meta';
    return '';
  }

  $: hunks = (mode === 'diff' && content.startsWith('diff ')) ? parseHunks(content) : [];

  function parseHunks(text) {
    const lines = text.split('\n');
    const result = [];
    let current = null;
    for (let i = 0; i < lines.length; i++) {
      if (lines[i].startsWith('@@')) {
        if (current) result.push(current);
        current = { startLine: i, endLine: i, text: lines[i] + '\n' };
      } else if (current) {
        current.endLine = i;
        current.text += lines[i] + '\n';
      }
    }
    if (current) result.push(current);
    return result;
  }

  function getHunkForLine(lineIndex) {
    return hunks.find(h => lineIndex >= h.startLine && lineIndex <= h.endLine);
  }

  function handleDiffLineEnter(e, lineIndex) {
    if (!explainAvailable) return;
    const line = content.split('\n')[lineIndex];
    const type = isDiffLine(line);
    if (type !== 'added' && type !== 'removed') return;
    const hunk = getHunkForLine(lineIndex);
    if (!hunk) return;
    if (hoverTimer) clearTimeout(hoverTimer);
    hoverTimer = setTimeout(async () => {
      const hunkKey = hunk.startLine + ':' + hunk.endLine;
      const rect = e.target.getBoundingClientRect();
      tooltipX = rect.left + 60;
      tooltipY = rect.top;
      tooltipVisible = true;
      if (explainCache.has(hunkKey)) {
        tooltipText = explainCache.get(hunkKey);
        tooltipLoading = false;
        tooltipError = '';
        return;
      }
      tooltipLoading = true;
      tooltipText = '';
      tooltipError = '';
      try {
        const explanation = await ExplainDiffHunk(repoPath, filePath, hunk.text);
        explainCache.set(hunkKey, explanation);
        tooltipText = explanation;
      } catch (err) {
        console.error('ExplainDiffHunk error:', err);
        tooltipError = typeof err === 'string' ? err : (err?.message || 'Failed to explain');
      } finally {
        tooltipLoading = false;
      }
    }, 400);
  }

  function handleDiffLineLeave() {
    if (hoverTimer) clearTimeout(hoverTimer);
  }

  function dismissTooltip() {
    tooltipVisible = false;
    tooltipText = '';
    tooltipError = '';
    tooltipLoading = false;
  }

  onMount(async () => {
    try { explainAvailable = await IsExplainAvailable(); } catch {}
  });

  onDestroy(() => {
    if (hoverTimer) clearTimeout(hoverTimer);
    if (saveTimer) clearTimeout(saveTimer);
    if (statusTimer) clearTimeout(statusTimer);
  });
</script>

<div class="editor">
  <div class="editor-header">
    <span class="file-path">{filePath}</span>
    <div class="header-right">
      {#if saveStatus === 'saving'}
        <span class="save-status saving">Saving...</span>
      {:else if saveStatus === 'saved'}
        <span class="save-status saved">Saved</span>
      {:else if saveStatus === 'error'}
        <span class="save-status error">Save failed</span>
      {/if}
      <div class="mode-toggle">
        <button class:active={mode === 'source'} on:click={() => mode = 'source'}>Source</button>
        <button class:active={mode === 'diff'} on:click={() => mode = 'diff'}>Diff</button>
      </div>
    </div>
  </div>

  <div class="editor-content" on:scroll={dismissTooltip}>
    {#if loading}
      <div class="loading">Loading...</div>
    {:else if error}
      <div class="error">{error}</div>
    {:else if mode === 'diff' && content.startsWith('diff ')}
      <pre class="code diff">{#each content.split('\n') as line, i}<span class="line {isDiffLine(line)}" class:hoverable={explainAvailable && (isDiffLine(line) === 'added' || isDiffLine(line) === 'removed')} on:mouseenter={(e) => handleDiffLineEnter(e, i)} on:mouseleave={handleDiffLineLeave}><span class="line-num">{i + 1}</span>{line}
</span>{/each}</pre>
    {:else if editable && isEditing}
      <textarea
        class="code-textarea"
        value={editContent}
        on:input={handleInput}
        on:keydown={handleKeydown}
        spellcheck="false"
      ></textarea>
    {:else}
      <!-- svelte-ignore a11y-click-events-have-key-events -->
      <pre
        class="code"
        class:editable
        on:click={editable ? startEditing : undefined}
        role={editable ? 'textbox' : undefined}
        tabindex={editable ? 0 : undefined}
      ><code class="lang-{getLanguage(filePath)}">{#each content.split('\n') as line, i}<span class="line"><span class="line-num">{i + 1}</span>{line}
</span>{/each}</code></pre>
    {/if}
  </div>
  {#if tooltipVisible}
    <div class="explain-tooltip" style="left: {tooltipX}px; top: {tooltipY}px;">
      <button class="tooltip-close" on:click={dismissTooltip}>&times;</button>
      {#if tooltipLoading}
        <span class="tooltip-loading">Thinking...</span>
      {:else if tooltipError}
        <span class="tooltip-error">{tooltipError}</span>
      {:else}
        <span class="tooltip-text">{tooltipText}</span>
      {/if}
    </div>
  {/if}
</div>

<style>
  .editor {
    display: flex;
    flex-direction: column;
    height: 100%;
    background: var(--bg-deepest);
  }

  .editor-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: var(--sp-xs) 12px;
    background: var(--bg-surface);
    border-bottom: 1px solid var(--border-subtle);
    flex-shrink: 0;
  }

  .file-path {
    font-family: var(--font-mono);
    font-size: var(--text-body);
    color: var(--text-primary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-shrink: 0;
  }

  .save-status {
    font-family: var(--font-mono);
    font-size: var(--text-label);
    padding: 1px var(--sp-xs);
    border-radius: var(--radius-sm);
  }

  .save-status.saving { color: var(--text-muted); }
  .save-status.saved { color: var(--accent-green); }
  .save-status.error { color: var(--accent-red); }

  .mode-toggle {
    display: flex;
    gap: 2px;
    flex-shrink: 0;
  }

  .mode-toggle button {
    padding: 2px 8px;
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    color: var(--text-dim);
    font-family: var(--font-ui);
    font-size: 11px;
    cursor: pointer;
    border-radius: var(--radius-sm);
  }

  .mode-toggle button:first-child { border-radius: var(--radius-sm) 0 0 var(--radius-sm); }
  .mode-toggle button:last-child { border-radius: 0 var(--radius-sm) var(--radius-sm) 0; }

  .mode-toggle button.active {
    background: var(--bg-active);
    color: var(--accent-green);
    border-color: var(--accent-green);
  }

  .editor-content {
    flex: 1;
    overflow: auto;
    padding: 0;
  }

  .code {
    margin: 0;
    padding: 8px 0;
    font-family: var(--font-mono);
    font-size: 13px;
    line-height: 1.5;
    color: var(--text-primary);
    white-space: pre;
    tab-size: 4;
  }

  .code.editable {
    cursor: text;
  }

  .code.editable:hover {
    background: rgba(255, 255, 255, 0.01);
  }

  .code-textarea {
    width: 100%;
    height: 100%;
    margin: 0;
    padding: 8px 12px;
    font-family: var(--font-mono);
    font-size: 13px;
    line-height: 1.5;
    color: var(--text-primary);
    background: var(--bg-deepest);
    border: none;
    outline: none;
    resize: none;
    white-space: pre;
    tab-size: 4;
    overflow: auto;
  }

  .line {
    display: block;
    padding: 0 12px 0 0;
  }

  .line:hover {
    background: var(--bg-elevated);
  }

  .line-num {
    display: inline-block;
    width: 48px;
    padding-right: 12px;
    text-align: right;
    color: var(--text-muted);
    user-select: none;
    font-size: var(--text-body);
  }

  /* Diff colors */
  .line.added {
    background: color-mix(in srgb, var(--accent-green) 8%, transparent);
    color: var(--accent-green);
  }

  .line.removed {
    background: color-mix(in srgb, var(--accent-red) 8%, transparent);
    color: var(--accent-red);
  }

  .line.hunk {
    color: var(--accent-blue);
    background: rgba(61, 158, 255, 0.05);
  }

  .line.meta {
    color: var(--text-muted);
  }

  .loading, .error {
    padding: 24px;
    color: var(--text-dim);
    font-size: 13px;
  }

  .error { color: var(--accent-red); }

  .line.hoverable { cursor: help; }
  .explain-tooltip {
    position: fixed;
    transform: translateY(-100%);
    max-width: 400px;
    padding: 8px 12px;
    padding-right: 28px;
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-sm);
    font-family: var(--font-ui);
    font-size: var(--text-body);
    line-height: 1.4;
    color: var(--text-primary);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
    z-index: 1000;
    animation: fadeIn 0.15s ease-out;
    pointer-events: auto;
  }
  .tooltip-close {
    position: absolute; top: 4px; right: 4px;
    width: 20px; height: 20px;
    display: flex; align-items: center; justify-content: center;
    background: none; border: none;
    color: var(--text-muted); cursor: pointer;
    font-size: 14px; padding: 0;
    border-radius: var(--radius-sm);
  }
  .tooltip-close:hover { color: var(--text-primary); background: var(--bg-active); }
  .tooltip-loading { color: var(--text-dim); font-style: italic; }
  .tooltip-error { color: var(--accent-red); }
  @keyframes fadeIn {
    from { opacity: 0; transform: translateY(-100%) translateY(4px); }
    to { opacity: 1; transform: translateY(-100%); }
  }
</style>
