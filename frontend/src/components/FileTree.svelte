<script>
  import { createEventDispatcher } from 'svelte';
  import { File, Folder, FolderOpen, ChevronRight } from 'lucide-svelte';
  import { flatPathsToTree } from '../lib/fileTree.js';

  export let files = [];
  export let selectedPath = '';
  export let changedPaths = new Set();

  const dispatch = createEventDispatcher();

  let expandedDirs = new Set();
  let searchQuery = '';

  $: filteredFiles = searchQuery
    ? files.filter(f => f.toLowerCase().includes(searchQuery.toLowerCase()))
    : files;

  $: tree = flatPathsToTree(filteredFiles);

  $: flatNodes = flattenVisible(tree, 0, expandedDirs);

  /**
   * Flatten the tree into a renderable list, only including children
   * of expanded directories. Each entry carries its depth for indentation.
   */
  function flattenVisible(nodes, depth, expanded) {
    let result = [];
    for (const node of nodes) {
      result.push({ node, depth });
      if (node.type === 'dir' && expanded.has(node.path) && node.children) {
        result = result.concat(flattenVisible(node.children, depth + 1, expanded));
      }
    }
    return result;
  }

  /**
   * Toggle directory expansion or dispatch file selection.
   * @param {import('../lib/fileTree.js').TreeNode} node
   */
  function handleClick(node) {
    if (node.type === 'dir') {
      const next = new Set(expandedDirs);
      if (next.has(node.path)) {
        next.delete(node.path);
      } else {
        next.add(node.path);
      }
      expandedDirs = next;
    } else {
      dispatch('select', { path: node.path });
    }
  }
</script>

<div class="file-tree">
  <div class="tree-search">
    <input
      type="text"
      placeholder="Filter files..."
      bind:value={searchQuery}
      class="tree-search-input"
    />
  </div>
  <div class="tree-list">
    {#if flatNodes.length === 0}
      <div class="tree-empty">
        {#if searchQuery}
          No files match "{searchQuery}"
        {:else}
          No files
        {/if}
      </div>
    {/if}
    {#each flatNodes as { node, depth }}
      <button
        class="tree-item"
        class:dir={node.type === 'dir'}
        class:file={node.type === 'file'}
        class:active={node.type === 'file' && selectedPath === node.path}
        class:changed={changedPaths.has(node.path)}
        style="padding-left: {8 + depth * 16}px"
        on:click={() => handleClick(node)}
      >
        <span class="tree-icon">
          {#if node.type === 'dir'}
            <svelte:component this={expandedDirs.has(node.path) ? FolderOpen : Folder} size={14} />
          {:else}
            <File size={14} />
          {/if}
        </span>
        <span class="tree-name">{node.name}</span>
        {#if node.type === 'dir'}
          <span class="tree-chevron" class:expanded={expandedDirs.has(node.path)}>
            <ChevronRight size={12} />
          </span>
        {/if}
      </button>
    {/each}
  </div>
</div>

<style>
  .file-tree {
    display: flex;
    flex-direction: column;
    height: 100%;
    overflow: hidden;
  }

  .tree-search {
    padding: 6px 8px;
    border-bottom: 1px solid var(--border-subtle);
  }

  .tree-search-input {
    width: 100%;
    padding: 4px 8px;
    background: var(--bg-deepest);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-sm);
    color: var(--text-primary);
    font-family: var(--font-mono);
    font-size: 11px;
    outline: none;
    box-sizing: border-box;
  }

  .tree-search-input:focus {
    border-color: var(--accent-green);
  }

  .tree-list {
    flex: 1;
    overflow-y: auto;
    padding: 4px 0;
  }

  .tree-item {
    display: flex;
    align-items: center;
    gap: 4px;
    width: 100%;
    padding: 3px 8px;
    background: none;
    border: none;
    color: var(--text-dim);
    font-family: var(--font-mono);
    font-size: 11px;
    cursor: pointer;
    text-align: left;
    white-space: nowrap;
    overflow: hidden;
  }

  .tree-item:hover {
    background: var(--bg-elevated);
    color: var(--text-primary);
  }

  .tree-item.active {
    background: var(--bg-active);
    color: var(--accent-green);
  }

  .tree-item.dir {
    color: var(--text-secondary);
  }

  .tree-item.changed .tree-name {
    color: var(--accent-yellow);
  }

  .tree-icon {
    flex-shrink: 0;
    display: flex;
    align-items: center;
    color: var(--text-muted);
  }

  .tree-item.dir .tree-icon {
    color: var(--accent-blue);
  }

  .tree-name {
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .tree-chevron {
    margin-left: auto;
    display: flex;
    align-items: center;
    color: var(--text-muted);
    transition: transform 0.15s ease;
  }

  .tree-chevron.expanded {
    transform: rotate(90deg);
  }

  .tree-empty {
    padding: 12px 8px;
    color: var(--text-muted);
    font-family: var(--font-mono);
    font-size: 11px;
    text-align: center;
  }
</style>
