<script>
  // Workflow builder toolbar (file name, New / Save / Agents, execution
  // controls). Extracted from WorkflowBuilder.svelte (spec R31) — markup
  // and scoped styles moved verbatim; ExecutionBar events are forwarded.
  import { Save } from 'lucide-svelte';
  import ExecutionBar from './ExecutionBar.svelte';

  /** @type {string} */
  export let workflowName = '';
  export let saving = false;
  export let executionStatus = 'idle';
  export let nodeProgress = { completed: 0, total: 0 };
  /** @type {string} */
  export let repoPath = '';
  /** @type {() => void} */
  export let onNew = () => {};
  /** @type {() => void} */
  export let onSave = () => {};
  /** @type {() => void} */
  export let onAgents = () => {};
</script>

<div class="toolbar">
  <span class="toolbar-label">File Name:</span>
  <input
    class="workflow-name-input"
    type="text"
    bind:value={workflowName}
    placeholder="Workflow name..."
  />
  <button class="toolbar-btn new-btn" on:click={onNew} title="New workflow">New</button>
  <button class="toolbar-btn save-btn" on:click={onSave} disabled={saving} title="Save workflow">
    <Save size={13} />
    {saving ? 'Saving...' : 'Save'}
  </button>
  <button class="toolbar-btn agents-btn" on:click={onAgents} title="Manage agents">
    Agents
  </button>

  <div class="toolbar-sep" />

  <ExecutionBar
    {executionStatus}
    {nodeProgress}
    {repoPath}
    on:start
    on:pause
    on:resume
    on:stop
  />
</div>

<style>
  .toolbar {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: var(--sp-xs) 10px;
    background: var(--bg-surface);
    border-bottom: 1px solid var(--border-subtle);
    flex-shrink: 0;
  }

  .toolbar-btn {
    display: flex;
    align-items: center;
    gap: var(--sp-xs);
    padding: 4px 10px;
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-sm);
    color: var(--text-primary);
    font-family: var(--font-mono);
    font-size: var(--text-body);
    font-weight: 500;
    cursor: pointer;
    transition: background 100ms ease, color 100ms ease, border-color 100ms ease;
    flex-shrink: 0;
  }

  .toolbar-btn:hover {
    background: var(--bg-active);
    color: var(--text-primary);
    border-color: var(--border-emphasis);
  }

  .toolbar-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .save-btn:hover { border-color: var(--accent-green); color: var(--accent-green); }
  .agents-btn:hover { border-color: var(--accent-purple, #9d6fff); color: var(--accent-purple, #9d6fff); }

  .toolbar-label {
    font-family: var(--font-mono);
    font-size: var(--text-body);
    font-weight: 500;
    color: var(--text-primary);
    flex-shrink: 0;
    white-space: nowrap;
  }

  .toolbar-sep {
    width: 1px;
    height: 18px;
    background: var(--border-subtle);
    flex-shrink: 0;
  }

  .workflow-name-input {
    flex: 1;
    min-width: 120px;
    padding: 4px 8px;
    background: var(--bg-deepest);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-sm);
    color: var(--text-primary);
    font-family: var(--font-mono);
    font-size: var(--text-body);
    outline: none;
    transition: border-color 100ms ease;
  }

  .workflow-name-input:focus { border-color: var(--accent-green); }
</style>
