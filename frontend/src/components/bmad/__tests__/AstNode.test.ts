/**
 * @vitest-environment jsdom
 */
import { describe, it, expect, vi } from 'vitest';

vi.mock('../../../../wailsjs/runtime/runtime.js', () => ({
  BrowserOpenURL: vi.fn(),
}));

import AstNode from '../AstNode.svelte';
import { makeMount } from './mountSvelte';

type UINode = {
  type: string;
  content?: string;
  tone?: string;
  heading?: string;
  bullets?: string[];
  lang?: string;
  copyable?: boolean;
  columns?: string[];
  rows?: string[][];
};

const FIXTURES: Array<{ label: string; node: UINode; expectTestId: string }> = [
  { label: 'markdown', node: { type: 'markdown', content: '# Hi' }, expectTestId: 'markdown-block' },
  { label: 'hint', node: { type: 'hint', tone: 'warn', content: 'careful' }, expectTestId: 'hint-banner' },
  {
    label: 'summary',
    node: { type: 'summary', heading: 'Done', bullets: ['a', 'b'] },
    expectTestId: 'summary-card',
  },
  {
    label: 'code',
    node: { type: 'code', lang: 'sh', content: 'ls', copyable: true },
    expectTestId: 'code-block',
  },
  {
    label: 'table',
    node: { type: 'table', heading: 'T', columns: ['x'], rows: [['1']] },
    expectTestId: 'comparison-table',
  },
  // Unknown type → MarkdownBlock fallback per spec §3.3 rule-1 / AC-10.
  { label: 'unknown', node: { type: 'snarkfish', content: 'hello' }, expectTestId: 'markdown-block' },
];

describe('AstNode dispatcher', () => {
  const mount = makeMount();
  const render = (node: UINode) => mount(AstNode, { node });

  for (const { label, node, expectTestId } of FIXTURES) {
    it(`AC5_six_node_shapes — type="${label}" renders [data-testid="${expectTestId}"]`, () => {
      expect(render(node).querySelector(`[data-testid="${expectTestId}"]`)).not.toBeNull();
    });
  }

  it('AC10_unknown_type_markdown_fallback — snarkfish node renders MarkdownBlock with its content', () => {
    const md = render({ type: 'snarkfish', content: 'hello' }).querySelector('[data-testid="markdown-block"]');
    expect(md).not.toBeNull();
    expect(md?.textContent).toContain('hello');
  });

  it('AC10_unknown_type_does_not_throw — rendering an unknown type must not throw', () => {
    expect(() => render({ type: 'snarkfish', content: 'hello' })).not.toThrow();
  });
});
