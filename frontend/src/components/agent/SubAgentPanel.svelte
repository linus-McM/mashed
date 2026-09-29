<script lang="ts">
  // Sub-agent info panel (description, result, activity log). Extracted from
  // views/AgentDetail.svelte (spec R31) — markup and scoped styles moved
  // verbatim. The parent renders this only when the agent is a sub-agent.
  import type { Notification } from '../../lib/types/wails';

  export let agent: Notification;
</script>

<div class="sub-agent-panel">
  {#if agent.subAgentDesc}
    <div class="sub-panel-desc">{agent.subAgentDesc}</div>
  {/if}
  {#if agent.subAgentResult}
    <div class="sub-panel-result">
      <span class="sub-panel-label">Result</span>
      <pre class="sub-panel-text">{agent.subAgentResult}</pre>
    </div>
  {/if}
  {#if agent.subAgentLogLines && agent.subAgentLogLines.length > 0}
    <div class="sub-panel-logs">
      <span class="sub-panel-label">Activity ({agent.subAgentLogLines.length} events)</span>
      <div class="sub-panel-log-list">
        {#each agent.subAgentLogLines as line}
          <div class="sub-log-line" class:log-ok={line.kind === 'ok'} class:log-err={line.kind === 'err'} class:log-dim={line.kind === 'dim'} class:log-system={line.kind === 'system'}>
            <span class="sub-log-text">{line.text}</span>
          </div>
        {/each}
      </div>
    </div>
  {/if}
</div>

<style>
  /* Sub-agent info panel */
  .sub-agent-panel {
    flex-shrink: 0;
    border-bottom: 1px solid var(--border-subtle);
    background: var(--bg-surface);
    max-height: 200px;
    overflow-y: auto;
  }

  .sub-panel-desc {
    padding: var(--sp-xs) var(--sp-lg);
    font-size: var(--text-body);
    color: var(--text-dim);
    border-bottom: 1px solid var(--border-subtle);
  }

  .sub-panel-result {
    padding: var(--sp-xs) var(--sp-lg);
    border-bottom: 1px solid var(--border-subtle);
  }

  .sub-panel-label {
    display: block;
    font-family: var(--font-mono);
    font-size: 9px;
    color: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    margin-bottom: 4px;
  }

  .sub-panel-text {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--text-primary);
    white-space: pre-wrap;
    word-break: break-word;
    margin: 0;
    max-height: 80px;
    overflow-y: auto;
  }

  .sub-panel-logs {
    padding: var(--sp-xs) var(--sp-lg);
  }

  .sub-panel-log-list {
    max-height: 100px;
    overflow-y: auto;
  }

  .sub-log-line {
    font-family: var(--font-mono);
    font-size: var(--text-label);
    color: var(--text-dim);
    padding: 1px 0;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .sub-log-line.log-ok { color: var(--accent-green); }
  .sub-log-line.log-err { color: var(--accent-red); }
  .sub-log-line.log-dim { color: var(--text-muted); }
  .sub-log-line.log-system { color: var(--accent-purple); }
</style>
