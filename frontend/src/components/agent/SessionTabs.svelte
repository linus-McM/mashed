<script lang="ts">
  // Terminal session tab bar. Extracted from views/AgentDetail.svelte
  // (spec R31) — markup and scoped styles moved verbatim. The parent owns the
  // session list and kill/spawn calls; the active index is shared via bind:.
  import type { Session } from '../../types/session';

  export let sessions: Session[];
  export let activeSessionIdx: number;
  export let onKill: (sessionName: string) => void;
  export let onAdd: () => void;
</script>

<div class="session-tabs">
  {#each sessions as session, idx}
    <button
      class="session-tab"
      class:active={idx === activeSessionIdx}
      on:click={() => activeSessionIdx = idx}
    >
      <span class="tab-type">{session.sessionType === 'agent' ? 'Agent' : 'Term'}</span>
      <span class="tab-name">{session.sessionName.split('-').slice(-1)[0]}</span>
      <button class="tab-close" on:click|stopPropagation={() => onKill(session.sessionName)}>×</button>
    </button>
  {/each}
  <button class="session-tab add-tab" on:click={onAdd}>+</button>
</div>

<style>
  /* Session tab bar */
  .session-tabs {
    display: flex;
    align-items: center;
    gap: 2px;
    padding: 0 8px;
    height: 32px;
    background: var(--bg-deeper);
    border-bottom: 1px solid var(--border-subtle);
    overflow-x: auto;
    flex-shrink: 0;
  }
  .session-tab {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 4px 8px;
    border: none;
    background: transparent;
    color: var(--fg-secondary);
    font-size: 11px;
    cursor: pointer;
    border-bottom: 2px solid transparent;
    white-space: nowrap;
  }
  .session-tab.active {
    color: var(--fg-primary);
    border-bottom-color: var(--border-accent);
  }
  .session-tab:hover { color: var(--fg-primary); }
  .tab-close {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 16px;
    height: 16px;
    border: none;
    background: transparent;
    color: var(--fg-dim);
    font-size: var(--text-body);
    cursor: pointer;
    border-radius: 3px;
    padding: 0;
  }
  .tab-close:hover { background: rgba(255, 95, 87, 0.2); color: var(--accent-red); }
  .add-tab { color: var(--fg-dim); font-size: var(--text-data); }
  .add-tab:hover { color: var(--accent-green, #50fa7b); }
</style>
