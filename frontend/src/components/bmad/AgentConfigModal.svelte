<script>
  import { createEventDispatcher } from 'svelte';
  import { X } from 'lucide-svelte';

  export let agent = null;
  export let processes = [];

  const dispatch = createEventDispatcher();

  const roles = [
    { value: 'analyst', label: 'Analyst' },
    { value: 'pm', label: 'PM' },
    { value: 'ux-designer', label: 'UX Designer' },
    { value: 'architect', label: 'Architect' },
    { value: 'developer', label: 'Developer' },
    { value: 'tech-writer', label: 'Tech Writer' },
    { value: 'qa', label: 'QA' },
  ];

  const models = [
    'claude-opus-4-6',
    'claude-sonnet-4-20250514',
    'claude-haiku-3.5',
  ];

  let name = '';
  let role = 'developer';
  let model = models[1];
  let persona = '';
  let selectedSkills = {};

  $: if (agent) {
    name = agent.name || '';
    role = agent.role || 'developer';
    model = agent.model || models[1];
    persona = agent.persona || '';
    selectedSkills = {};
    for (const s of (agent.skills || [])) {
      selectedSkills[s] = true;
    }
  } else {
    name = '';
    role = 'developer';
    model = models[1];
    persona = '';
    selectedSkills = {};
  }

  $: isEditing = !!agent?.id;

  function handleSave() {
    const skills = Object.entries(selectedSkills)
      .filter(([, v]) => v)
      .map(([k]) => k);

    dispatch('save', {
      id: agent?.id || `agent-${Date.now()}`,
      name,
      role,
      model,
      persona,
      skills,
      createdAt: agent?.createdAt || new Date().toISOString(),
    });
  }

  function handleDelete() {
    if (agent?.id) {
      dispatch('delete', agent.id);
    }
  }

  function handleKeydown(e) {
    if (e.key === 'Escape') dispatch('close');
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<div class="overlay" on:click={() => dispatch('close')} on:keydown={() => {}}>
  <div class="modal" on:click|stopPropagation on:keydown={() => {}}>
    <div class="modal-header">
      <h3>{isEditing ? 'Edit Agent' : 'Create Custom Agent'}</h3>
      <button class="close-btn" on:click={() => dispatch('close')}><X size={16} /></button>
    </div>

    <div class="modal-body">
      <div class="field">
        <label class="field-label" for="agent-name">Name</label>
        <input id="agent-name" class="field-input" type="text" bind:value={name} placeholder="Agent name..." />
      </div>

      <div class="field">
        <label class="field-label" for="agent-role">Role</label>
        <select id="agent-role" class="field-select" bind:value={role}>
          {#each roles as r}
            <option value={r.value}>{r.label}</option>
          {/each}
        </select>
      </div>

      <div class="field">
        <label class="field-label" for="agent-model">Model</label>
        <select id="agent-model" class="field-select" bind:value={model}>
          {#each models as m}
            <option value={m}>{m}</option>
          {/each}
        </select>
      </div>

      <div class="field">
        <label class="field-label" for="agent-persona">Persona</label>
        <textarea
          id="agent-persona"
          class="field-textarea"
          bind:value={persona}
          placeholder="Custom system prompt additions..."
          rows="4"
        />
      </div>

      <div class="field">
        <label class="field-label">Allowed Skills</label>
        <div class="skill-list">
          {#each processes as proc}
            <label class="skill-checkbox">
              <input type="checkbox" bind:checked={selectedSkills[proc.id]} />
              <span class="skill-name">{proc.name}</span>
            </label>
          {/each}
        </div>
        <span class="skill-hint">Leave all unchecked to allow any process</span>
      </div>
    </div>

    <div class="modal-footer">
      {#if isEditing}
        <button class="btn delete-btn" on:click={handleDelete}>Delete</button>
      {/if}
      <button class="btn save-btn" on:click={handleSave} disabled={!name.trim()}>
        {isEditing ? 'Update Agent' : 'Save Agent'}
      </button>
    </div>
  </div>
</div>

<style>
  .overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.6);
    z-index: 100;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .modal {
    width: 420px;
    max-height: 80vh;
    background: var(--bg-surface);
    border: 1px solid var(--border-emphasis);
    border-radius: var(--radius-lg, 8px);
    display: flex;
    flex-direction: column;
    overflow: hidden;
    box-shadow: 0 16px 48px rgba(0, 0, 0, 0.5);
  }

  .modal-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 16px;
    border-bottom: 1px solid var(--border-subtle);
    flex-shrink: 0;
  }

  .modal-header h3 {
    font-family: var(--font-mono);
    font-size: 13px;
    font-weight: 600;
    color: var(--text-primary);
    margin: 0;
  }

  .close-btn {
    background: none;
    border: none;
    color: var(--text-muted);
    cursor: pointer;
    padding: 2px;
    display: flex;
    align-items: center;
  }

  .close-btn:hover { color: var(--text-primary); }

  .modal-body {
    flex: 1;
    overflow-y: auto;
    padding: 16px;
    display: flex;
    flex-direction: column;
    gap: 14px;
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

  .field-input, .field-select {
    padding: 6px 8px;
    background: var(--bg-deepest);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-sm);
    color: var(--text-primary);
    font-family: var(--font-mono);
    font-size: 12px;
    outline: none;
  }

  .field-input:focus, .field-select:focus { border-color: var(--accent-green); }

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

  .skill-list {
    max-height: 160px;
    overflow-y: auto;
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-sm);
    padding: 4px;
    background: var(--bg-deepest);
  }

  .skill-checkbox {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 3px 6px;
    cursor: pointer;
    border-radius: var(--radius-sm);
    transition: background 80ms ease;
  }

  .skill-checkbox:hover { background: var(--bg-elevated); }

  .skill-checkbox input[type="checkbox"] {
    accent-color: var(--accent-green);
    cursor: pointer;
  }

  .skill-name {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--text-primary);
  }

  .skill-hint {
    font-family: var(--font-mono);
    font-size: 9px;
    color: var(--text-muted);
  }

  .modal-footer {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 8px;
    padding: 12px 16px;
    border-top: 1px solid var(--border-subtle);
    flex-shrink: 0;
  }

  .btn {
    padding: 6px 14px;
    border-radius: var(--radius-sm);
    font-family: var(--font-mono);
    font-size: 11px;
    cursor: pointer;
    transition: background 100ms ease, color 100ms ease, border-color 100ms ease;
  }

  .save-btn {
    background: var(--accent-green, #00e57a);
    border: 1px solid var(--accent-green, #00e57a);
    color: var(--bg-deepest);
    font-weight: 600;
  }

  .save-btn:hover { filter: brightness(1.1); }
  .save-btn:disabled { opacity: 0.4; cursor: not-allowed; }

  .delete-btn {
    background: none;
    border: 1px solid var(--accent-red, #f85149);
    color: var(--accent-red, #f85149);
    margin-right: auto;
  }

  .delete-btn:hover { background: rgba(248, 81, 73, 0.1); }
</style>
