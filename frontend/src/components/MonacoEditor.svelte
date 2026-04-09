<script>
  import { onMount, onDestroy } from 'svelte';
  import { ReadFile, ReadFileAtHead, WriteFile, ExplainDiffHunk, IsExplainAvailable } from '../../wailsjs/go/main/App.js';
  import { defineAllThemes, getEditorFont, toMonacoId } from '../lib/monacoTheme.js';
  import { currentMonoFont, currentFontSize } from '../lib/stores/font.js';
  import { allThemes, currentThemeId, builtInThemeIds } from '../lib/stores/theme.js';
  import { editorSettings } from '../lib/stores/editorSettings.js';

  export let filePath = '';
  export let repoPath = '';
  export let mode = 'source'; // 'source' or 'diff'
  export let editable = false;

  let container;
  let editor = null;
  let resizeObserver = null;
  let saveTimer = null;
  let statusTimer = null;
  let hoverDisposables = [];
  let saveStatus = ''; // '', 'saving', 'saved', 'error'
  let saving = false;
  let loading = true;
  let error = '';
  let monacoModule = null;
  let themeRegistered = false;
  let registeredThemeIds = new Set();

  // Hover-to-explain state
  let explainAvailable = false;
  let hoverTimer = null;
  let explainCache = new Map();
  let tooltipVisible = false;
  let tooltipX = 0;
  let tooltipY = 0;
  let tooltipText = '';
  let tooltipLoading = false;
  let tooltipError = '';

  $: fullPath = filePath.startsWith('/') ? filePath : repoPath + '/' + filePath;

  // Reactive block: watch props and reload
  $: if (filePath && repoPath && monacoModule) {
    loadFile(filePath, repoPath, mode);
  }

  // Reactive block: update readOnly when editable changes
  $: if (editor && mode === 'source' && editor.updateOptions) {
    editor.updateOptions({ readOnly: !editable });
  }

  function getLanguage(path) {
    const ext = path.split('.').pop()?.toLowerCase();
    const map = {
      go: 'go', js: 'javascript', ts: 'typescript', tsx: 'typescript', jsx: 'javascript',
      svelte: 'html', css: 'css', html: 'html', json: 'json', md: 'markdown',
      yaml: 'yaml', yml: 'yaml', toml: 'toml', sh: 'shell', bash: 'shell',
      py: 'python', rs: 'rust', sql: 'sql', mod: 'go',
    };
    return map[ext] || 'plaintext';
  }

  function mapSettingsToMonaco(s) {
    return {
      minimap: { enabled: s.minimapEnabled },
      scrollBeyondLastLine: s.scrollBeyondLastLine,
      renderLineHighlight: s.renderLineHighlight,
      wordWrap: s.wordWrap,
      lineNumbers: s.lineNumbers,
      renderWhitespace: s.renderWhitespace,
      tabSize: s.tabSize,
      insertSpaces: s.insertSpaces,
      cursorStyle: s.cursorStyle,
      cursorBlinking: s.cursorBlinking,
      bracketPairColorization: { enabled: s.bracketPairColorization },
      fontLigatures: s.fontLigatures,
      smoothScrolling: s.smoothScrolling,
    };
  }

  function getEditorOptions() {
    return {
      theme: toMonacoId($currentThemeId),
      fontFamily: getEditorFont(),
      fontSize: $currentFontSize,
      lineHeight: 1.5 * $currentFontSize,
      ...mapSettingsToMonaco($editorSettings),
      padding: { top: 8 },
      automaticLayout: false,
      readOnly: !editable,
      glyphMargin: false,
      folding: true,
      lineDecorationsWidth: 0,
      lineNumbersMinChars: 4,
    };
  }

  function dismissTooltip() {
    tooltipVisible = false;
    tooltipText = '';
    tooltipLoading = false;
    tooltipError = '';
    if (hoverTimer) {
      clearTimeout(hoverTimer);
      hoverTimer = null;
    }
  }

  function getHunkForLine(diffEditor, lineNumber) {
    const changes = diffEditor.getLineChanges();
    if (!changes) return null;

    for (const change of changes) {
      // modifiedEndLineNumber === 0 means pure deletion (no modified lines)
      const modStart = change.modifiedStartLineNumber;
      const modEnd = change.modifiedEndLineNumber || 0;
      // originalEndLineNumber === 0 means pure insertion (no original lines)
      const origStart = change.originalStartLineNumber;
      const origEnd = change.originalEndLineNumber || 0;

      // Check if the hovered line is in the modified range of this change
      if (modEnd > 0 && lineNumber >= modStart && lineNumber <= modEnd) {
        const origModel = diffEditor.getOriginalEditor().getModel();
        const modModel = diffEditor.getModifiedEditor().getModel();

        let hunkText = '';
        // Add removed lines (from original)
        if (origEnd > 0) {
          for (let i = origStart; i <= origEnd; i++) {
            hunkText += '-' + origModel.getLineContent(i) + '\n';
          }
        }
        // Add inserted lines (from modified)
        for (let i = modStart; i <= modEnd; i++) {
          hunkText += '+' + modModel.getLineContent(i) + '\n';
        }

        return {
          key: `${modStart}:${modEnd}`,
          text: hunkText,
        };
      }
    }
    return null;
  }

  function handleDiffLineHover(diffEditor, e) {
    if (!explainAvailable) return;
    if (!e.target?.position) return;

    const lineNumber = e.target.position.lineNumber;
    const hunk = getHunkForLine(diffEditor, lineNumber);
    if (!hunk) {
      // Not hovering over a changed line — clear any pending timer but leave visible tooltip
      if (hoverTimer) {
        clearTimeout(hoverTimer);
        hoverTimer = null;
      }
      return;
    }

    // If already showing tooltip for same hunk, do nothing
    if (tooltipVisible && explainCache.has(hunk.key) && tooltipText === explainCache.get(hunk.key)) {
      return;
    }

    if (hoverTimer) clearTimeout(hoverTimer);

    hoverTimer = setTimeout(async () => {
      const modEditor = diffEditor.getModifiedEditor();
      const pos = modEditor.getScrolledVisiblePosition({ lineNumber, column: 1 });
      if (!pos) return;

      const editorDom = modEditor.getDomNode();
      if (!editorDom) return;
      const rect = editorDom.getBoundingClientRect();

      tooltipX = rect.left + pos.left + 60;
      tooltipY = rect.top + pos.top;
      tooltipVisible = true;

      if (explainCache.has(hunk.key)) {
        tooltipText = explainCache.get(hunk.key);
        tooltipLoading = false;
        tooltipError = '';
        return;
      }

      tooltipLoading = true;
      tooltipText = '';
      tooltipError = '';

      try {
        const explanation = await ExplainDiffHunk(repoPath, filePath, hunk.text);
        explainCache.set(hunk.key, explanation);
        tooltipText = explanation;
      } catch (err) {
        tooltipError = typeof err === 'string' ? err : (err?.message || 'Failed to explain');
      } finally {
        tooltipLoading = false;
      }
    }, 400);
  }

  function setupExplainHover(diffEditor) {
    const modEditor = diffEditor.getModifiedEditor();

    hoverDisposables.push(
      modEditor.onMouseMove((e) => {
        handleDiffLineHover(diffEditor, e);
      })
    );

    hoverDisposables.push(
      modEditor.onDidScrollChange(() => {
        dismissTooltip();
      })
    );

    hoverDisposables.push(
      modEditor.onMouseLeave(() => {
        if (hoverTimer) {
          clearTimeout(hoverTimer);
          hoverTimer = null;
        }
      })
    );
  }

  function destroyEditor() {
    // Dispose hover event listeners to prevent leaks
    for (const d of hoverDisposables) {
      d.dispose();
    }
    hoverDisposables = [];

    if (editor) {
      // Dispose models to prevent memory leaks
      if (mode === 'diff' && editor.getModel) {
        const diffModel = editor.getModel();
        if (diffModel?.original) diffModel.original.dispose();
        if (diffModel?.modified) diffModel.modified.dispose();
      } else if (editor.getModel) {
        const model = editor.getModel();
        if (model?.dispose) model.dispose();
      }
      editor.dispose();
      editor = null;
    }
  }

  function setupResizeObserver() {
    if (resizeObserver) {
      resizeObserver.disconnect();
    }
    if (container) {
      resizeObserver = new ResizeObserver(() => {
        if (editor) {
          editor.layout();
        }
      });
      resizeObserver.observe(container);
    }
  }

  async function loadFile(_fp, _rp, currentMode) {
    if (!filePath || !repoPath || !monacoModule || !container) return;

    loading = true;
    error = '';
    saveStatus = '';

    // Clear explain state on file/mode change
    dismissTooltip();
    explainCache = new Map();

    // Destroy existing editor before creating a new one
    destroyEditor();

    const monaco = monacoModule;
    const lang = getLanguage(filePath);

    try {
      if (currentMode === 'diff') {
        await createDiffEditor(monaco, lang);
      } else {
        await createSourceEditor(monaco, lang);
      }
    } catch (e) {
      error = e?.message || 'Failed to load file';
    } finally {
      loading = false;
    }
  }

  async function createSourceEditor(monaco, lang) {
    let content;
    try {
      content = await ReadFile(fullPath);
    } catch (e) {
      throw new Error(e?.message || 'Failed to read file');
    }

    const model = monaco.editor.createModel(content, lang);

    editor = monaco.editor.create(container, {
      ...getEditorOptions(),
      model,
    });

    // Auto-save on content change (debounced)
    editor.onDidChangeModelContent(() => {
      scheduleSave();
    });

    // Cmd/Ctrl+S
    editor.addCommand(
      monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyS,
      () => doSave()
    );

    setupResizeObserver();
  }

  async function createDiffEditor(monaco, lang) {
    let originalContent = '';
    let modifiedContent = '';

    // Get the modified (current) content
    try {
      modifiedContent = await ReadFile(fullPath);
    } catch (e) {
      throw new Error(e?.message || 'Failed to read file');
    }

    // Get the original (HEAD) content — new files will fail, use empty string
    try {
      originalContent = await ReadFileAtHead(repoPath, filePath);
    } catch {
      originalContent = '';
    }

    const originalModel = monaco.editor.createModel(originalContent, lang);
    const modifiedModel = monaco.editor.createModel(modifiedContent, lang);

    editor = monaco.editor.createDiffEditor(container, {
      ...getEditorOptions(),
      readOnly: true, // diff mode is always read-only
      renderSideBySide: true,
      enableSplitViewResizing: true,
    });

    editor.setModel({
      original: originalModel,
      modified: modifiedModel,
    });

    // Wire up hover-to-explain on the diff editor
    if (explainAvailable) {
      setupExplainHover(editor);
    }

    setupResizeObserver();
  }

  function scheduleSave() {
    if (saveTimer) clearTimeout(saveTimer);
    saveStatus = '';
    saveTimer = setTimeout(() => doSave(), 800);
  }

  async function doSave() {
    if (!editable || saving || mode === 'diff') return;
    if (!editor || !editor.getValue) return;

    saving = true;
    saveStatus = 'saving';

    try {
      const content = editor.getValue();
      await WriteFile(fullPath, content);
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

  function switchMode(newMode) {
    if (newMode === mode) return;
    mode = newMode;
    // The reactive block will handle reloading
  }

  onMount(async () => {
    try {
      // Check explain availability early (non-blocking for editor load)
      IsExplainAvailable().then((available) => {
        explainAvailable = available;
      }).catch(() => {
        explainAvailable = false;
      });

      // Set up Monaco workers lazily — only created when Monaco requests them.
      // Uses Vite's new URL() pattern so worker code stays out of the main chunk.
      self.MonacoEnvironment = {
        getWorker(_, label) {
          if (label === 'json')
            return new Worker(new URL('monaco-editor/esm/vs/language/json/json.worker.js', import.meta.url), { type: 'module' });
          if (label === 'css' || label === 'scss' || label === 'less')
            return new Worker(new URL('monaco-editor/esm/vs/language/css/css.worker.js', import.meta.url), { type: 'module' });
          if (label === 'html' || label === 'handlebars' || label === 'razor')
            return new Worker(new URL('monaco-editor/esm/vs/language/html/html.worker.js', import.meta.url), { type: 'module' });
          if (label === 'typescript' || label === 'javascript')
            return new Worker(new URL('monaco-editor/esm/vs/language/typescript/ts.worker.js', import.meta.url), { type: 'module' });
          return new Worker(new URL('monaco-editor/esm/vs/editor/editor.worker.js', import.meta.url), { type: 'module' });
        }
      };

      // Import core editor + only the languages/features this app uses.
      // This avoids pulling ~90 language grammars we never touch.
      monacoModule = await import('monaco-editor/esm/vs/editor/editor.api.js');

      await Promise.all([
        // Basic languages (syntax highlighting only)
        import('monaco-editor/esm/vs/basic-languages/go/go.contribution.js'),
        import('monaco-editor/esm/vs/basic-languages/javascript/javascript.contribution.js'),
        import('monaco-editor/esm/vs/basic-languages/typescript/typescript.contribution.js'),
        import('monaco-editor/esm/vs/basic-languages/html/html.contribution.js'),
        import('monaco-editor/esm/vs/basic-languages/css/css.contribution.js'),
        import('monaco-editor/esm/vs/basic-languages/markdown/markdown.contribution.js'),
        import('monaco-editor/esm/vs/basic-languages/yaml/yaml.contribution.js'),
        import('monaco-editor/esm/vs/basic-languages/shell/shell.contribution.js'),
        import('monaco-editor/esm/vs/basic-languages/python/python.contribution.js'),
        import('monaco-editor/esm/vs/basic-languages/rust/rust.contribution.js'),
        import('monaco-editor/esm/vs/basic-languages/sql/sql.contribution.js'),
        import('monaco-editor/esm/vs/basic-languages/xml/xml.contribution.js'),
        // Rich languages (intellisense + validation)
        import('monaco-editor/esm/vs/language/json/monaco.contribution.js'),
        import('monaco-editor/esm/vs/language/css/monaco.contribution.js'),
        import('monaco-editor/esm/vs/language/html/monaco.contribution.js'),
        import('monaco-editor/esm/vs/language/typescript/monaco.contribution.js'),
      ]);
      if (!themeRegistered) {
        const defined = defineAllThemes(monacoModule);
        for (const id of defined) {
          registeredThemeIds.add(id);
        }
        themeRegistered = true;
      }
      // The reactive block ($: if (filePath && repoPath && monacoModule)) will
      // trigger loadFile automatically now that monacoModule is set.
    } catch (e) {
      error = 'Failed to load editor: ' + (e?.message || String(e));
      loading = false;
    }
  });

  onDestroy(() => {
    if (saveTimer) clearTimeout(saveTimer);
    if (statusTimer) clearTimeout(statusTimer);
    if (hoverTimer) clearTimeout(hoverTimer);
    if (resizeObserver) {
      resizeObserver.disconnect();
      resizeObserver = null;
    }
    destroyEditor();
  });

  // Live theme switching — register imported themes on demand, then activate
  $: if (monacoModule && $currentThemeId) {
    const theme = $allThemes[$currentThemeId];
    const monacoId = toMonacoId($currentThemeId);
    if (theme && theme.monaco && !registeredThemeIds.has($currentThemeId)) {
      try {
        monacoModule.editor.defineTheme(monacoId, theme.monaco);
        registeredThemeIds.add($currentThemeId);
      } catch (e) {
        console.error('Failed to define Monaco theme:', e);
      }
    }
    // Use sanitized ID for Monaco; fall back to first built-in if current isn't registered
    monacoModule.editor.setTheme(
      registeredThemeIds.has($currentThemeId) ? monacoId : toMonacoId(builtInThemeIds[0])
    );
  }

  // Live font switching
  $: if (editor && $currentMonoFont) {
    editor.updateOptions({ fontFamily: $currentMonoFont });
  }
  $: if (editor && $currentFontSize) {
    editor.updateOptions({ fontSize: $currentFontSize, lineHeight: 1.5 * $currentFontSize });
  }

  // Live editor settings switching
  $: if (editor && $editorSettings) {
    editor.updateOptions(mapSettingsToMonaco($editorSettings));
  }
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
        <button class:active={mode === 'source'} on:click={() => switchMode('source')}>Source</button>
        <button class:active={mode === 'diff'} on:click={() => switchMode('diff')}>Diff</button>
      </div>
    </div>
  </div>

  {#if loading && !editor}
    <div class="loading-overlay">Loading...</div>
  {/if}
  {#if error}
    <div class="error-overlay">{error}</div>
  {/if}
  <div class="editor-container" bind:this={container}></div>

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

  .editor-container {
    flex: 1;
    overflow: hidden;
  }

  .loading-overlay, .error-overlay {
    padding: 24px;
    font-size: 13px;
  }

  .loading-overlay {
    color: var(--text-dim);
  }

  .error-overlay {
    color: var(--accent-red);
  }

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
    font-size: 12px;
    line-height: 1.4;
    color: var(--text-primary);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
    z-index: 1000;
    animation: fadeIn 0.15s ease-out;
    pointer-events: auto;
  }

  .tooltip-close {
    position: absolute;
    top: 4px;
    right: 4px;
    width: 20px;
    height: 20px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: none;
    border: none;
    color: var(--text-muted);
    cursor: pointer;
    font-size: 14px;
    padding: 0;
    border-radius: var(--radius-sm);
  }

  .tooltip-close:hover {
    color: var(--text-primary);
    background: var(--bg-active);
  }

  .tooltip-loading {
    color: var(--text-dim);
    font-style: italic;
  }

  .tooltip-error {
    color: var(--accent-red);
  }

  @keyframes fadeIn {
    from { opacity: 0; transform: translateY(-100%) translateY(4px); }
    to { opacity: 1; transform: translateY(-100%); }
  }
</style>
