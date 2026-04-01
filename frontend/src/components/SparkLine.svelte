<script>
  /** @type {number[]} Ring buffer data (token burn values) */
  export let data = [];

  /** @type {number} Max value for normalization (0 = auto) */
  export let max = 0;

  const blocks = '▁▂▃▄▅▆▇█';

  $: effectiveMax = max > 0 ? max : Math.max(...data, 1);
  $: sparkline = data
    .map(v => {
      const idx = Math.round((v / effectiveMax) * (blocks.length - 1));
      return blocks[Math.min(Math.max(idx, 0), blocks.length - 1)];
    })
    .join('');
</script>

<span class="sparkline" title="Token burn rate">{sparkline || '▁'.repeat(20)}</span>

<style>
  .sparkline {
    font-family: var(--font-mono);
    font-size: var(--text-body);
    color: var(--teal);
    letter-spacing: -0.5px;
    line-height: 1;
    user-select: none;
  }
</style>
