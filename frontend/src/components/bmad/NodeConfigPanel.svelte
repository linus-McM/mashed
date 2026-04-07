<script>
  import { createEventDispatcher } from 'svelte';
  import { X, Terminal } from 'lucide-svelte';

  export let node = null;
  export let agents = [];

  const dispatch = createEventDispatcher();

  const models = [
    { value: '', label: 'Default (inherit)' },
    { value: 'claude-opus-4-6', label: 'claude-opus-4-6' },
    { value: 'claude-sonnet-4-20250514', label: 'claude-sonnet-4-20250514' },
    { value: 'claude-haiku-3.5', label: 'claude-haiku-3.5' },
  ];

  let modelOverride = '';
  let customContext = '';
  let selectedAgent = '';

  $: if (node) {
    const cfg = node.data?.config || {};
    modelOverride = cfg.model || '';
    customContext = cfg.context || '';
    selectedAgent = cfg.agentId || '';
  }

  $: label = node?.data?.label || 'Node';
  $: status = node?.data?.status || 'pending';
  $: tmuxTarget = node?.data?.tmuxTarget || '';
  $: hasTerminal = status === 'running' && tmuxTarget;

  function emitUpdate() {
    dispatch('update', {
      nodeId: node.id,
      config: {
        model: modelOverride,
        context: customContext,
        agentId: selectedAgent,
      },
    });
  }

  function close() {
    dispatch('close');
  }
</script>

{#if node}
  <div class="config-panel" class:visible={!!node}>
    <div class="panel-header">
      <span class="panel-title">Configure: {label}</span>
      <button class="close-btn" on:click={close} title="Close">
        <X size={14} />
      </button>
    </div>

    <div class="panel-body">
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
          {#each agents as agent}
            <option value={agent.id}>{agent.name} ({agent.role})</option>
          {/each}
        </select>
      </div>

      {#if hasTerminal}
        <button class="terminal-btn" on:click={() => dispatch('open-terminal', tmuxTarget)}>
          <Terminal size={13} />
          View Terminal
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
    width: 260px;
    height: 100%;
    background: var(--bg-surface);
    border-left: 1px solid var(--border-subtle);
    display: flex;
    flex-direction: column;
    z-index: 10;
    box-shadow: -4px 0 16px rgba(0, 0, 0, 0.2);
    animation: slide-in 150ms ease-out;
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
    font-size: 12px;
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
    font-size: 10px;
    font-weight: 600;
    color: var(--text-dim);
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .field-select {
    padding: 5px 8px;
    background: var(--bg-deepest);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-sm);
    color: var(--text-primary);
    font-family: var(--font-mono);
    font-size: 11px;
    outline: none;
  }

  .field-select:focus { border-color: var(--accent-green); }

  .field-textarea {
    padding: 6px 8px;
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

  .terminal-btn {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 6px 10px;
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
</style>
