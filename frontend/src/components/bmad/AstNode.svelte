<script>
  // Recursive dispatcher for UIAST nodes — spec §6.2.
  //
  // Unknown types fall through to MarkdownBlock per §3.3 rule-1.
  // `decision_group` routes to DecisionGroup only when a `responses` store is
  // supplied by an interactive caller (the modal); passive callers pass no
  // store and get the markdown fallback so read-only views never render
  // interactive widgets.
  import MarkdownBlock from './MarkdownBlock.svelte';
  import HintBanner from './HintBanner.svelte';
  import SummaryCard from './SummaryCard.svelte';
  import CodeBlock from './CodeBlock.svelte';
  import ComparisonTable from './ComparisonTable.svelte';
  import DecisionGroup from './DecisionGroup.svelte';

  /** @type {import('../../types/uiAst').UINode} */
  export let node;
  /** @type {import('svelte/store').Writable<Record<string, string>> | null} */
  export let responses = null;
  /** @type {((node: import('../../types/uiAst').UINode) => boolean) | null} */
  export let isGroupDisabled = null;
  /** @type {((node: import('../../types/uiAst').UINode) => boolean) | null} */
  export let isGroupActive = null;
</script>

{#if node.type === 'markdown'}
  <MarkdownBlock content={node.content ?? ''} />
{:else if node.type === 'hint'}
  <HintBanner tone={node.tone ?? 'info'} content={node.content ?? ''} />
{:else if node.type === 'summary'}
  <SummaryCard heading={node.heading ?? ''} bullets={node.bullets ?? []} />
{:else if node.type === 'code'}
  <CodeBlock
    lang={node.lang ?? ''}
    content={node.content ?? ''}
    copyable={node.copyable ?? false}
  />
{:else if node.type === 'table'}
  <ComparisonTable
    heading={node.heading ?? ''}
    columns={node.columns ?? []}
    rows={node.rows ?? []}
  />
{:else if node.type === 'decision_group' && responses}
  <DecisionGroup
    {node}
    {responses}
    disabled={isGroupDisabled ? isGroupDisabled(node) : false}
    active={isGroupActive ? isGroupActive(node) : false}
    on:activate
  />
{:else if node.type === 'decision_group'}
  <MarkdownBlock content={node.prompt ?? node.heading ?? JSON.stringify(node)} />
{:else}
  <MarkdownBlock content={node.content ?? JSON.stringify(node)} />
{/if}
