<script>
  import { onMount, onDestroy } from 'svelte';
  import { ReadFile, ReadFileDiff, WriteFile } from '../../wailsjs/go/main/App.js';

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

  $: if (filePath && repoPath) loadFile(filePath, repoPath, mode);
  $: fullPath = filePath.startsWith('/') ? filePath : repoPath + '/' + filePath;

  async function loadFile(fp, rp, m) {
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

  onDestroy(() => {
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

  <div class="editor-content">
    {#if loading}
      <div class="loading">Loading...</div>
    {:else if error}
      <div class="error">{error}</div>
    {:else if mode === 'diff' && content.startsWith('diff ')}
      <pre class="code diff">{#each content.split('\n') as line, i}<span class="line {isDiffLine(line)}"><span class="line-num">{i + 1}</span>{line}
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
    padding: 6px 12px;
    background: var(--bg-surface);
    border-bottom: 1px solid var(--border-subtle);
    flex-shrink: 0;
  }

  .file-path {
    font-family: var(--font-mono);
    font-size: 12px;
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
    font-size: 10px;
    padding: 1px 6px;
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
    font-size: 12px;
  }

  /* Diff colors */
  .line.added {
    background: rgba(0, 229, 122, 0.08);
    color: var(--accent-green);
  }

  .line.removed {
    background: rgba(232, 69, 69, 0.08);
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
</style>
