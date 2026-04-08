<script>
  import { BaseEdge, EdgeLabelRenderer, getBezierPath, useSvelteFlow } from '@xyflow/svelte';

  export let id;
  export let sourceX;
  export let sourceY;
  export let targetX;
  export let targetY;
  export let sourcePosition;
  export let targetPosition;
  export let style = '';
  export let markerEnd = '';
  export let selected = false;

  const { deleteElements } = useSvelteFlow();

  let hovered = false;

  $: [edgePath, labelX, labelY] = getBezierPath({
    sourceX, sourceY, targetX, targetY, sourcePosition, targetPosition,
  });

  function onDelete(e) {
    e.stopPropagation();
    deleteElements({ edges: [{ id }] });
  }
</script>

<g
  on:mouseenter={() => hovered = true}
  on:mouseleave={() => hovered = false}
>
  <BaseEdge path={edgePath} {markerEnd} {style} />
</g>

<EdgeLabelRenderer>
  <div
    class="edge-delete-btn"
    class:visible={hovered || selected}
    style="position: absolute; transform: translate(-50%, -50%) translate({labelX}px, {labelY}px); pointer-events: all;"
    on:click={onDelete}
    on:keydown={() => {}}
    on:mouseenter={() => hovered = true}
    on:mouseleave={() => hovered = false}
    role="button"
    tabindex="-1"
    title="Delete edge"
  >
    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
      <polyline points="3 6 5 6 21 6" />
      <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
    </svg>
  </div>
</EdgeLabelRenderer>

<style>
  .edge-delete-btn {
    width: 22px;
    height: 22px;
    border-radius: 50%;
    background: var(--bg-elevated, #2d333b);
    border: 1px solid var(--border-subtle, #373e47);
    color: var(--text-muted, #8b949e);
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    opacity: 0;
    transition: opacity 150ms ease, color 150ms ease, border-color 150ms ease, background 150ms ease;
    z-index: 10;
  }

  .edge-delete-btn.visible {
    opacity: 1;
  }

  .edge-delete-btn:hover {
    color: var(--accent-red, #f85149);
    border-color: var(--accent-red, #f85149);
    background: rgba(248, 81, 73, 0.12);
  }
</style>
