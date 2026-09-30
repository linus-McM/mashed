<script context="module" lang="ts">
  import type { ComponentType } from 'svelte';

  export type LazyLoader = () => Promise<{ default: ComponentType }>;

  // Resolved components keyed by loader identity, so remounting a lazy view
  // (e.g. navigating back to it) renders immediately with no placeholder
  // flash. Callers should pass a stable, module-level loader.
  const resolved = new WeakMap<LazyLoader, ComponentType>();
</script>

<script lang="ts">
  // LazyView (R31) — code-splits a heavy view behind a dynamic import() so it
  // stays out of the entry chunk. Renders an accessible aria-busy placeholder
  // until the import resolves, then either:
  //   - renders the default slot with `let:component` (use this when the
  //     parent needs on:event / bind: on the loaded component), or
  //   - mounts the component itself with every other prop passed through.
  export let loader: LazyLoader;
  export let loadingLabel = 'Loading…';

  let component: ComponentType | null = resolved.get(loader) ?? null;
  let error = '';
  let requested: LazyLoader | null = component ? loader : null;

  $: if (loader !== requested) load(loader);

  function load(l: LazyLoader): void {
    requested = l;
    const cached = resolved.get(l);
    if (cached) {
      component = cached;
      error = '';
      return;
    }
    component = null;
    error = '';
    l().then(
      (mod) => {
        resolved.set(l, mod.default);
        if (requested === l) component = mod.default;
      },
      (err: unknown) => {
        if (requested === l) error = err instanceof Error ? err.message : String(err);
      },
    );
  }
</script>

{#if component}
  {#if $$slots.default}
    <slot {component} />
  {:else}
    <svelte:component this={component} {...$$restProps} />
  {/if}
{:else if error}
  <div class="lazy-view lazy-view-error" role="alert">Failed to load view: {error}</div>
{:else}
  <div class="lazy-view" role="status" aria-busy="true" aria-live="polite">{loadingLabel}</div>
{/if}

<style>
  .lazy-view {
    display: flex;
    flex: 1;
    align-items: center;
    justify-content: center;
    padding: var(--sp-xl);
    color: var(--text-secondary);
    font-size: var(--text-body);
  }

  .lazy-view-error {
    color: var(--accent-red);
  }
</style>
