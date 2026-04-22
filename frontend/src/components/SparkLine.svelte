<script lang="ts">
  /** SparkLine props — per cerebrum (Phase 4c story): ONLY `data` + `maxVal`. */
  export let data: number[] = []; // array of numbers (token counts)
  export let maxVal: number = 0;   // max value for scaling (0 = auto)

  const blocks = '▁▂▃▄▅▆▇█';

  function render(data: number[], maxVal: number): string {
    if (!data.length) return '';
    const max = maxVal || Math.max(...data, 1);
    return data.map((v: number) => {
      const idx = Math.min(Math.floor((v / max) * 7), 7);
      return blocks[idx];
    }).join('');
  }
</script>

<span class="sparkline">{render(data, maxVal)}</span>

<style>
  .sparkline {
    font-family: var(--font-mono);
    font-size: var(--text-body);
    color: var(--accent-teal);
    letter-spacing: -0.5px;
  }
</style>
