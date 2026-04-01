<script>
  import { onMount } from 'svelte';
  import { ReadFile, ReadFileDiff } from '../../wailsjs/go/main/App.js';

  export let filePath = '';
  export let repoPath = '';
  export let mode = 'source'; // 'source' or 'diff'

  let content = '';
  let loading = true;
  let error = '';

  $: if (filePath && repoPath) loadFile(filePath, repoPath, mode);

  async function loadFile(fp, rp, m) {
    loading = true;
    error = '';
    content = '';
    try {
      if (m === 'diff') {
        content = await ReadFileDiff(rp, fp);
        if (!content) {
          // No diff, show source instead
          const fullPath = fp.startsWith('/') ? fp : rp + '/' + fp;
          content = await ReadFile(fullPath);
        }
      } else {
        const fullPath = fp.startsWith('/') ? fp : rp + '/' + fp;
        content = await ReadFile(fullPath);
      }
    } catch (e) {
      error = e?.message || 'Failed to load file';
    } finally {
      loading = false;
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
</script>

<div class="editor">
  <div class="editor-header">
    <span class="file-path">{filePath}</span>
    <div class="mode-toggle">
      <button class:active={mode === 'source'} on:click={() => mode = 'source'}>Source</button>
      <button class:active={mode === 'diff'} on:click={() => mode = 'diff'}>Diff</button>
    </div>
  </div>

  <div class="editor-content">
    {#if loading}
      <div class="loading">Loading...</div>
    {:else if error}
      <div class="error">{error}</div>
    {:else if mode === 'diff' && content.startsWith('diff ')}
      <pre class="code diff">{#each content.split('\n') as line, i}<span class="line {isDiffLine(line)}"><span class="line-num">{i + 1}</span>{line}</span>
{/each}</pre>
    {:else}
      <pre class="code"><code class="lang-{getLanguage(filePath)}">{#each content.split('\n') as line, i}<span class="line"><span class="line-num">{i + 1}</span>{line}</span>
{/each}</code></pre>
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
