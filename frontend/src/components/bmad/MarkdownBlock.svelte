<script context="module" lang="ts">
  // Module-scope so every MarkdownBlock instance shares one parser + rule
  // override. The override closes over stateless helpers — safe to share.
  //
  // Security invariants:
  //  - `html: false` — never render raw HTML from AST content.
  //  - `linkify: false` — autolinker would emit <a>s bypassing our sanitiser.
  //  - every <a> href runs through sanitizeUrl; rejected schemes degrade to
  //    plain text (the <a> is emitted without an href attribute).
  import MarkdownIt from 'markdown-it';
  import { sanitizeUrl } from './linkSanitiser';

  type RenderRule = NonNullable<MarkdownIt['renderer']['rules'][string]>;

  const md = new MarkdownIt({ html: false, linkify: false, breaks: false });

  const defaultLinkOpen: RenderRule =
    md.renderer.rules.link_open ||
    ((tokens, idx, options, _env, self) => self.renderToken(tokens, idx, options));

  md.renderer.rules.link_open = (tokens, idx, options, env, self) => {
    const token = tokens[idx];
    const hrefIdx = token.attrIndex('href');
    if (hrefIdx >= 0 && token.attrs) {
      const href = token.attrs[hrefIdx][1];
      const safe = sanitizeUrl(href);
      if (!safe) {
        token.attrs.splice(hrefIdx, 1);
      } else {
        token.attrs[hrefIdx][1] = safe;
        token.attrSet('data-sanitised-href', safe);
      }
    }
    return defaultLinkOpen(tokens, idx, options, env, self);
  };
</script>

<script lang="ts">
  import { BrowserOpenURL } from '../../../wailsjs/runtime/runtime.js';

  export let content = '';
  $: rendered = md.render(content ?? '');

  // §7.2: intercept every <a> click inside the rendered prose, show a confirm
  // dialog with the fully-resolved URL, and open via the Wails runtime so the
  // OS browser takes over (never the app's own webview).
  function onRootClick(event: MouseEvent): void {
    const anchor = event.target instanceof Element ? event.target.closest('a[href]') : null;
    if (!anchor) return;
    event.preventDefault();
    const safe = anchor.getAttribute('data-sanitised-href') ?? sanitizeUrl(anchor.getAttribute('href'));
    if (!safe) return;
    if (window.confirm(`Open ${safe}?`)) {
      BrowserOpenURL(safe);
    }
  }
</script>

<div
  class="markdown-block"
  data-testid="markdown-block"
  on:click={onRootClick}
  role="presentation"
>
  {@html rendered}
</div>

<style>
  .markdown-block {
    color: var(--text-primary);
    font-family: var(--font-ui);
    font-size: var(--text-body);
    line-height: 1.55;
  }

  .markdown-block :global(p) {
    margin: 0;
  }
  .markdown-block :global(p + p) {
    margin-top: var(--sp-sm);
  }

  .markdown-block :global(h1),
  .markdown-block :global(h2),
  .markdown-block :global(h3) {
    font-family: var(--font-ui);
    font-weight: 600;
    line-height: 1.4;
    margin-top: var(--sp-md);
    margin-bottom: var(--sp-xs);
  }
  .markdown-block :global(h1) { font-size: var(--text-section); }
  .markdown-block :global(h2) { font-size: var(--text-data); }
  .markdown-block :global(h3) { font-size: var(--text-body); }
  .markdown-block :global(:first-child) { margin-top: 0; }

  .markdown-block :global(code) {
    font-family: var(--font-code);
    font-weight: 500;
    padding: 0 var(--sp-2xs);
    background: var(--bg-elevated);
    border-radius: var(--radius-sm);
  }

  .markdown-block :global(ul),
  .markdown-block :global(ol) {
    padding-left: var(--sp-lg);
    margin: 0;
  }
  .markdown-block :global(li + li) {
    margin-top: var(--sp-2xs);
  }

  .markdown-block :global(a) {
    color: var(--accent-green);
    text-decoration: none;
  }
  .markdown-block :global(a:hover) {
    text-decoration: underline;
    text-decoration-thickness: 1.5px;
    text-underline-offset: 2px;
  }
</style>
