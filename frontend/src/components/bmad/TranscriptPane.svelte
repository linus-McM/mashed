<script>
  import { afterUpdate, tick } from 'svelte';

  /**
   * @typedef {Object} Turn
   * @property {number} round
   * @property {'claude'|'user'} role
   * @property {string} inputId
   * @property {string} content
   * @property {number} timestamp
   */

  /** @type {Turn[]} */
  export let turns = [];
  export let loading = false;
  export let error = '';

  /** @type {HTMLDivElement | null} */
  let scrollEl = null;
  let lastTurnCount = 0;
  let userPinnedToBottom = true;

  function isAtBottom() {
    if (!scrollEl) return true;
    const threshold = 48;
    return scrollEl.scrollHeight - scrollEl.scrollTop - scrollEl.clientHeight < threshold;
  }

  function onScroll() {
    userPinnedToBottom = isAtBottom();
  }

  afterUpdate(async () => {
    if (!scrollEl) return;
    if (turns.length === lastTurnCount) return;
    const grew = turns.length > lastTurnCount;
    lastTurnCount = turns.length;
    if (grew && userPinnedToBottom) {
      await tick();
      scrollEl.scrollTop = scrollEl.scrollHeight;
    }
  });

  function roleLabel(role) {
    return role === 'claude' ? 'Claude' : 'You';
  }
</script>

<div
  class="transcript"
  bind:this={scrollEl}
  on:scroll={onScroll}
  data-testid="transcript-pane"
  role="log"
  aria-live="polite"
  aria-relevant="additions"
>
  {#if loading && turns.length === 0}
    <div class="transcript-empty" data-testid="transcript-loading">Loading conversation…</div>
  {:else if error}
    <div class="transcript-error" role="alert">{error}</div>
  {:else if turns.length === 0}
    <div class="transcript-empty" data-testid="transcript-empty">
      No conversation yet — Claude hasn't finished its first turn.
    </div>
  {:else}
    {#each turns as turn, i (i + '-' + turn.round + '-' + turn.role)}
      <article
        class="turn"
        class:claude={turn.role === 'claude'}
        class:user={turn.role === 'user'}
        data-testid="turn"
        data-role={turn.role}
        data-round={turn.round}
      >
        <header class="turn-head">
          <span class="turn-round">Round {turn.round}</span>
          <span class="turn-role">{roleLabel(turn.role)}</span>
          {#if turn.role === 'user' && turn.inputId}
            <span class="turn-input-id">{turn.inputId}</span>
          {/if}
        </header>
        <pre class="turn-body">{turn.content}</pre>
      </article>
    {/each}
  {/if}
</div>

<style>
  .transcript {
    display: flex;
    flex-direction: column;
    gap: var(--sp-sm);
    max-height: 420px;
    overflow-y: auto;
    padding: var(--sp-sm) var(--sp-md) var(--sp-md) var(--sp-md);
    background: var(--bg-deepest);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    scrollbar-width: thin;
    scrollbar-color: var(--text-dim) transparent;
  }

  .transcript::-webkit-scrollbar {
    width: 8px;
  }
  .transcript::-webkit-scrollbar-thumb {
    background: var(--text-dim);
    border-radius: var(--radius-sm);
  }

  .transcript-empty,
  .transcript-error {
    padding: var(--sp-md);
    color: var(--text-secondary);
    font-size: var(--text-label);
    font-style: italic;
  }
  .transcript-error {
    color: var(--accent-amber);
    font-style: normal;
  }

  .turn {
    display: flex;
    flex-direction: column;
    gap: var(--sp-2xs);
    padding: var(--sp-sm) var(--sp-md);
    border-radius: var(--radius-md);
    border-left: 3px solid transparent;
    background: var(--bg-panel);
  }

  .turn.claude {
    border-left-color: var(--accent-green);
  }

  .turn.user {
    border-left-color: var(--accent-cyan);
    background: color-mix(in oklab, var(--bg-panel) 82%, var(--accent-cyan) 18%);
  }

  .turn-head {
    display: flex;
    align-items: center;
    gap: var(--sp-sm);
    font-family: var(--font-mono);
    font-size: var(--text-label);
    text-transform: uppercase;
    letter-spacing: 0.4px;
  }

  .turn-round {
    color: var(--text-secondary);
    font-weight: 600;
  }

  .turn-role {
    color: var(--text-primary);
    font-weight: 700;
  }

  .turn.claude .turn-role {
    color: var(--accent-green);
  }
  .turn.user .turn-role {
    color: var(--accent-cyan);
  }

  .turn-input-id {
    color: var(--text-secondary);
    font-weight: 500;
  }

  .turn-body {
    margin: 0;
    font-family: var(--font-code, var(--font-mono));
    font-size: var(--text-label);
    line-height: 1.55;
    color: var(--text-primary);
    white-space: pre-wrap;
    word-break: break-word;
    max-height: 240px;
    overflow-y: auto;
    overflow-x: hidden;
  }

  .turn.user .turn-body {
    color: var(--text-primary);
    font-weight: 500;
  }
</style>
