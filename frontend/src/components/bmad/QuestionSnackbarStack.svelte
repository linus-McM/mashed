<!--
  Legacy alias. The stack was renamed to NodeInputSnackbarStack in S6 so it
  can render both legacy QuestionEvents and new interactive PendingPrompts
  in a single stack. Keep this shim for one release to avoid breaking
  existing imports across the codebase.
-->
<script>
  import Stack from './NodeInputSnackbarStack.svelte';
  /** @type {import('./questionSnackbarUtils').SnackbarEntry[]} */
  export let questions = [];
</script>

<Stack {questions} on:navigate on:dismiss on:respond on:skip />
