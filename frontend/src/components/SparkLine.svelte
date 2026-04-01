<script>
  export let data = []; // array of numbers (token counts)
  export let maxVal = 0; // max value for scaling (0 = auto)

  const blocks = '\u2581\u2582\u2583\u2584\u2585\u2586\u2587\u2588';

  function render(data, maxVal) {
    if (!data.length) return '';
    const max = maxVal || Math.max(...data, 1);
    return data.map(v => {
      const idx = Math.min(Math.floor((v / max) * 7), 7);
      return blocks[idx];
    }).join('');
  }
</script>

<span class="sparkline">{render(data, maxVal)}</span>

<style>
  .sparkline {
    font-family: var(--font-mono);
    font-size: 12px;
    color: var(--accent-teal);
    letter-spacing: -0.5px;
  }
</style>
