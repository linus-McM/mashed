<script>
  // Summary card — spec §6.2, Design Brief §3 + §5.
  //
  // One accent moment per card: the `›` bullet marker in accent-green. The
  // rest stays on bg-elevated/border-subtle/text-primary so the card reads
  // as a calm grouped summary, not a CTA.
  import { scrubUnsafeUrls } from './linkSanitiser';

  export let heading = '';
  /** @type {string[]} */
  export let bullets = [];

  $: safeHeading = scrubUnsafeUrls(heading);
  $: safeBullets = (bullets ?? []).map(scrubUnsafeUrls);
</script>

<section class="summary-card" data-testid="summary-card">
  {#if safeHeading}
    <h3 class="heading">{safeHeading}</h3>
  {/if}
  {#if safeBullets.length > 0}
    <ul class="bullets">
      {#each safeBullets as bullet}
        <li><span class="marker" aria-hidden="true">›</span><span class="text">{bullet}</span></li>
      {/each}
    </ul>
  {/if}
</section>

<style>
  .summary-card {
    padding: var(--sp-md);
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    color: var(--text-primary);
    font-family: var(--font-ui);
  }

  .heading {
    margin: 0 0 var(--sp-sm);
    font-size: var(--text-data);
    font-weight: 600;
    line-height: 1.3;
    color: var(--text-primary);
  }

  .bullets {
    list-style: none;
    padding: 0;
    margin: 0;
    display: flex;
    flex-direction: column;
    gap: var(--sp-xs);
    font-size: var(--text-body);
    line-height: 1.5;
  }

  .bullets li {
    display: flex;
    gap: var(--sp-sm);
    align-items: baseline;
  }

  .marker {
    color: var(--accent-green);
    font-weight: 700;
    flex: 0 0 auto;
  }

  .text {
    flex: 1 1 auto;
    min-width: 0;
  }
</style>
