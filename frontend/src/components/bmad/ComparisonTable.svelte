<script>
  // Comparison table — spec §6.2, Design Brief §5.
  //
  // Zebra rows use a 50%-alpha mix of bg-elevated — at 13px density a hard
  // zebra looks harsh; the soft alpha keeps rows distinct without noise.
  import { scrubUnsafeUrls } from './linkSanitiser';

  export let heading = '';
  /** @type {string[]} */
  export let columns = [];
  /** @type {string[][]} */
  export let rows = [];

  $: safeHeading = scrubUnsafeUrls(heading);
  $: safeColumns = (columns ?? []).map(scrubUnsafeUrls);
  $: safeRows = (rows ?? []).map((r) => (r ?? []).map(scrubUnsafeUrls));
</script>

<section class="comparison-wrapper" data-testid="comparison-table">
  {#if safeHeading}
    <header class="heading">{safeHeading}</header>
  {/if}
  <table class="comparison">
    <thead>
      <tr>
        {#each safeColumns as col}
          <th>{col}</th>
        {/each}
      </tr>
    </thead>
    <tbody>
      {#each safeRows as row, i}
        <tr class:even={i % 2 === 1}>
          {#each row as cell}
            <td>{cell}</td>
          {/each}
        </tr>
      {/each}
    </tbody>
  </table>
</section>

<style>
  .comparison-wrapper {
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    overflow: hidden;
    overflow-x: auto;
    font-family: var(--font-ui);
  }

  .heading {
    padding: var(--sp-sm) var(--sp-md) var(--sp-xs);
    background: var(--bg-elevated);
    color: var(--text-primary);
    font-size: var(--text-data);
    font-weight: 600;
    line-height: 1.3;
  }

  .comparison {
    width: 100%;
    border-collapse: collapse;
    font-size: var(--text-body);
  }

  thead th {
    padding: var(--sp-xs) var(--sp-md);
    text-align: left;
    background: var(--bg-elevated);
    color: var(--text-dim);
    border-bottom: 1px solid var(--border-subtle);
    text-transform: uppercase;
    font-family: var(--font-mono);
    font-size: var(--text-label);
    font-weight: 600;
    letter-spacing: 0.05em;
    line-height: 1.4;
  }

  tbody td {
    padding: var(--sp-xs) var(--sp-md);
    color: var(--text-primary);
    border-bottom: 1px solid color-mix(in srgb, var(--border-subtle) 50%, transparent);
    line-height: 1.5;
  }

  tbody tr.even td {
    background: color-mix(in srgb, var(--bg-elevated) 50%, transparent);
  }

  tbody tr:last-child td {
    border-bottom: none;
  }
</style>
