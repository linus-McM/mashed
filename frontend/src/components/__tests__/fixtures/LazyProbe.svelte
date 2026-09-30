<script lang="ts">
  // Test double for lazily-loaded views: echoes every prop it receives so
  // tests can assert pass-through, and dispatches `ping` on click so event
  // wiring across the lazy boundary can be asserted.
  import { createEventDispatcher } from 'svelte';

  const dispatch = createEventDispatcher<{ ping: string }>();

  $: received = JSON.stringify($$props);
</script>

<button type="button" data-testid="lazy-probe" data-props={received} on:click={() => dispatch('ping', 'pong')}>
  probe
</button>
