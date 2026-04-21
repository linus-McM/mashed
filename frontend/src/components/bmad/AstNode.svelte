<script>
  // Recursive dispatcher for UIAST nodes — spec §6.2.
  //
  // Unknown types fall through to MarkdownBlock per §3.3 rule-1 (graceful
  // fallback rather than hide-or-throw). decision_group is handled in U7;
  // until then it also renders as markdown so partial deployments degrade
  // cleanly.
  import MarkdownBlock from './MarkdownBlock.svelte';
  import HintBanner from './HintBanner.svelte';
  import SummaryCard from './SummaryCard.svelte';
  import CodeBlock from './CodeBlock.svelte';
  import ComparisonTable from './ComparisonTable.svelte';

  export let node;
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
{:else if node.type === 'decision_group'}
  <MarkdownBlock content={node.prompt ?? node.heading ?? JSON.stringify(node)} />
{:else}
  <MarkdownBlock content={node.content ?? JSON.stringify(node)} />
{/if}
