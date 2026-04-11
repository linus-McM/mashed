<script>
  export let files = []; // Array of {path, added, removed, isBinary, isNew}

  function formatStat(file) {
    if (file.isBinary) return 'binary';
    const parts = [];
    if (file.added > 0) parts.push(`+${file.added}`);
    if (file.removed > 0) parts.push(`-${file.removed}`);
    return parts.join(' ') || '~';
  }
</script>

{#if files.length === 0}
  <div class="empty">No changes detected</div>
{:else}
  <div class="diff-list">
    {#each files as file}
      <div class="diff-row">
        <span class="filename" class:new={file.isNew} class:binary={file.isBinary}>
          {file.isNew ? '+ ' : ''}{file.path}
        </span>
        <span class="stats">
          {#if file.added > 0}
            <span class="added">+{file.added}</span>
          {/if}
          {#if file.removed > 0}
            <span class="removed">-{file.removed}</span>
          {/if}
          {#if file.isBinary}
            <span class="binary-label">binary</span>
          {/if}
        </span>
      </div>
    {/each}
  </div>
{/if}

<style>
  .diff-list {
    font-family: var(--font-mono);
    font-size: var(--text-body);
  }
  .diff-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 4px 8px;
    border-bottom: 1px solid var(--border-subtle);
  }
  .diff-row:hover {
    background: var(--bg-elevated);
  }
  .filename {
    color: var(--text-primary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .filename.new { color: var(--accent-green); }
  .filename.binary { color: var(--text-dim); }
  .stats {
    flex-shrink: 0;
    margin-left: 12px;
  }
  .added { color: var(--accent-green); margin-right: 4px; }
  .removed { color: var(--accent-red); }
  .binary-label { color: var(--text-dim); font-style: italic; }
  .empty {
    color: var(--text-dim);
    padding: 16px;
    text-align: center;
    font-style: italic;
  }
</style>
