<script lang="ts">
  // Agent rows (with sparkline + kill button) and the sub-agent accordion for
  // one repo group. Extracted from views/NotificationFeed.svelte (R31).
  // Selection, kill state and keyboard nav stay in the parent; accordion
  // expansion state is two-way bound so the parent's flatAgents sees it.
  import { createEventDispatcher } from 'svelte';
  import { Trash2, Circle, TerminalSquare, ChevronRight, ChevronDown } from 'lucide-svelte';
  import StatusBadge from '../StatusBadge.svelte';
  import SparkLine from '../SparkLine.svelte';
  import { formatTokens, formatElapsed, type NotificationEntry } from '../../lib/feed/repoTree';

  export let agents: NotificationEntry[];
  export let selectedId: string | null;
  export let killingAgents: Set<string>;
  /** Accordion state for sub-agents per parent agent (bind: from parent). */
  export let expandedAgents: Set<string>;

  const dispatch = createEventDispatcher<{
    select: NotificationEntry;
    kill: { agent: NotificationEntry; event: Event };
  }>();

  function handleClick(evt: NotificationEntry): void {
    dispatch('select', evt);
  }

  function toggleAgentAccordion(agentId: string, e: Event): void {
    e.stopPropagation();
    if (expandedAgents.has(agentId)) {
      expandedAgents.delete(agentId);
    } else {
      expandedAgents.add(agentId);
    }
    expandedAgents = expandedAgents;
  }

  function isAgentExpanded(agentId: string): boolean {
    return expandedAgents.has(agentId);
  }

  /**
   * Narrows an optional `tokenSamples` array to a defined one for template
   * use. The caller guards on `agent.tokenSamples?.length > 1` immediately
   * before the `{@const}` that invokes this helper, so the input is never
   * `undefined` in practice — the assertion is confined to this one spot.
   */
  function requireSamples(s: number[] | undefined): number[] {
    return s as number[];
  }
</script>

<!-- Left: Agents (75%) -->
<div class="repo-agents">
  {#each agents as agent}
    <div class="agent-accordion" class:has-children={agent.subAgents && agent.subAgents.length > 0}>
      <div
        class="agent-row"
        class:selected={selectedId === agent.agentId}
        class:is-running={agent.eventType === 'running'}
        data-agent-id={agent.agentId}
        on:click={() => handleClick(agent)}
        on:keydown={(e) => { if (e.key === 'Enter') handleClick(agent); }}
        role="button"
        tabindex="0"
      >
        <div class="agent-content">
          {#if agent.subAgents && agent.subAgents.length > 0}
            <button class="accordion-toggle" on:click={(e) => toggleAgentAccordion(agent.agentId, e)}>
              {#if isAgentExpanded(agent.agentId)}
                <ChevronDown size={12} />
              {:else}
                <ChevronRight size={12} />
              {/if}
            </button>
          {:else}
            <span class="agent-indicator">
              {#if agent.eventType === 'terminal'}
                <TerminalSquare size={12} />
              {:else}
                <Circle size={8} />
              {/if}
            </span>
          {/if}
          <span class="agent-model">{agent.eventType === 'terminal' ? 'shell' : (agent.model || agent.agentName)}</span>
          <StatusBadge status={agent.eventType} size="sm" />
          {#if agent.subAgents && agent.subAgents.length > 0}
            <span class="sub-count">{agent.subAgents.length} sub</span>
          {/if}
          <span class="agent-summary">{agent.summary}</span>
          {#if agent.eventType !== 'terminal' && agent.tokenSamples && agent.tokenSamples?.length > 1}
            {@const samples = requireSamples(agent.tokenSamples)}
            <span
              class="sparkline-wrap"
              class:dimmed={agent.eventType !== 'running'}
              title={`Token history: ${samples[0]} → ${samples[samples.length - 1]} over last ${samples.length} samples`}
            >
              <SparkLine data={agent.tokenSamples} />
            </span>
          {/if}
          {#if agent.eventType !== 'terminal'}
            <span class="agent-tokens mono">{formatTokens(agent.tokensUsed || 0)}</span>
          {/if}
          <span class="agent-elapsed mono">{formatElapsed(agent.timestamp)}</span>
          <button
            class="kill-btn"
            title="Kill session"
            disabled={killingAgents.has(agent.agentId)}
            on:click={(e) => dispatch('kill', { agent, event: e })}
          ><Trash2 size={12} /></button>
        </div>
      </div>

      {#if agent.subAgents && agent.subAgents.length > 0 && isAgentExpanded(agent.agentId)}
        <div class="sub-agent-accordion">
          {#each agent.subAgents as sub, i}
            <div
              class="sub-agent-row"
              class:selected={selectedId === sub.agentId}
              data-agent-id={sub.agentId}
              on:click={() => handleClick(sub)}
              on:keydown={(e) => { if (e.key === 'Enter') handleClick(sub); }}
              role="button"
              tabindex="0"
            >
              <div class="sub-content">
                <span class="tree-line">{i < agent.subAgents.length - 1 ? '├─' : '└─'}</span>
                <span class="sub-indicator" class:done={sub.subAgentStatus === 'done'}>
                  {#if sub.subAgentStatus === 'done'}
                    <Circle size={6} />
                  {:else}
                    <span class="sub-pulse" />
                  {/if}
                </span>
                <span class="sub-name">{sub.subAgentName || sub.agentName}</span>
                <StatusBadge status={sub.eventType} size="sm" />
                <span class="sub-summary" title={sub.subAgentDesc || sub.summary}>{sub.subAgentDesc || sub.summary}</span>
                <span class="agent-elapsed mono">{formatElapsed(sub.timestamp)}</span>
              </div>
            </div>
          {/each}
        </div>
      {/if}
    </div>
  {/each}

</div>

<style>
  .repo-agents {
    flex: 3;
    min-width: 0;
    border-right: 1px solid var(--border-subtle);
  }

  /* Agent row */
  .agent-row {
    display: flex;
    align-items: stretch;
    cursor: pointer;
    transition: background 100ms ease-out;
  }

  .agent-row:hover { background: var(--bg-surface); }
  .agent-row.selected { background: var(--bg-elevated); }

  .agent-content {
    display: flex;
    align-items: center;
    gap: var(--sp-sm);
    padding: var(--sp-xs) var(--sp-lg);
    padding-left: 20px;
    flex: 1;
    min-width: 0;
  }

  .agent-indicator {
    color: var(--accent-green);
    font-size: 8px;
    flex-shrink: 0;
  }

  .agent-model {
    font-family: var(--font-mono);
    font-size: var(--text-body);
    color: var(--text-primary);
    flex-shrink: 0;
  }

  .agent-summary {
    font-size: var(--text-body);
    color: var(--text-dim);
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .agent-tokens {
    font-size: var(--text-label);
    font-variant-numeric: tabular-nums;
    color: var(--text-dim);
    flex-shrink: 0;
  }

  /* Sparkline wrapper (uiqa-09). Text-based SparkLine sits between the
     summary and the token count so the glyph reads as the history of the
     numeric value it neighbours. flex-shrink:0 prevents long summaries
     from compressing the glyph. */
  .sparkline-wrap {
    flex-shrink: 0;
    align-self: center;
    line-height: 1;
  }

  .sparkline-wrap.dimmed :global(.sparkline) {
    color: var(--text-dim);
  }

  .agent-elapsed {
    font-size: var(--text-label);
    font-variant-numeric: tabular-nums;
    color: var(--text-muted);
    flex-shrink: 0;
    min-width: 32px;
    text-align: right;
  }

  .kill-btn {
    background: none;
    border: none;
    color: var(--text-dim);
    font-size: 11px;
    cursor: pointer;
    padding: 2px 4px;
    border-radius: var(--radius-sm);
    transition: color 100ms ease;
    flex-shrink: 0;
    line-height: 1;
  }

  .kill-btn:hover { color: var(--accent-red); }

  .kill-btn:disabled {
    opacity: 0.3;
    cursor: not-allowed;
  }

  /* Agent accordion */
  .agent-accordion {
    border-bottom: 1px solid transparent;
  }

  .agent-accordion.has-children {
    border-bottom: 1px solid var(--border-subtle);
  }

  .agent-accordion.has-children:last-child {
    border-bottom: none;
  }

  .accordion-toggle {
    background: none;
    border: none;
    color: var(--text-muted);
    cursor: pointer;
    padding: 0;
    display: flex;
    align-items: center;
    flex-shrink: 0;
    transition: color 100ms ease;
  }

  .accordion-toggle:hover { color: var(--accent-green); }

  .sub-count {
    font-family: var(--font-mono);
    font-size: var(--text-label);
    font-variant-numeric: tabular-nums;
    color: var(--text-muted);
    background: var(--bg-active);
    padding: 0 var(--sp-xs);
    border-radius: 8px;
    flex-shrink: 0;
    line-height: 16px;
  }

  /* Sub-agent accordion panel */
  .sub-agent-accordion {
    background: rgba(0, 0, 0, 0.15);
    border-top: 1px solid var(--border-subtle);
    padding: 2px 0;
  }

  /* Sub-agent row */
  .sub-agent-row {
    display: flex;
    align-items: stretch;
    cursor: pointer;
    transition: background 100ms ease-out;
  }

  .sub-agent-row:hover { background: var(--bg-surface); }
  .sub-agent-row.selected { background: var(--bg-elevated); }

  .sub-content {
    display: flex;
    align-items: center;
    gap: var(--sp-xs);
    padding: var(--sp-2xs) var(--sp-lg);
    padding-left: 36px;
    flex: 1;
    min-width: 0;
  }

  .tree-line {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--text-muted);
    flex-shrink: 0;
    user-select: none;
  }

  .sub-indicator {
    color: var(--accent-green);
    font-size: 6px;
    flex-shrink: 0;
  }
  .sub-indicator.done {
    color: var(--accent-red);
  }

  .sub-pulse {
    display: inline-block;
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--accent-green);
    animation: sub-pulse-anim 1.5s ease-in-out infinite;
  }

  @keyframes sub-pulse-anim {
    0%, 100% { opacity: 0.3; }
    50% { opacity: 1; }
  }

  .sub-name {
    font-family: var(--font-mono);
    font-size: var(--text-body);
    color: var(--text-dim);
    flex-shrink: 0;
    max-width: 160px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .sub-summary {
    font-size: var(--text-body);
    color: var(--text-muted);
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .mono { font-family: var(--font-mono); }
</style>
