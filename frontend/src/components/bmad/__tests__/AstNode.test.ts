/**
 * @vitest-environment jsdom
 */
import { describe, it, expect, vi } from 'vitest';
import { writable } from 'svelte/store';

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

  it('AC1_astnode_routes_decision_group_to_DecisionGroup — routes when `responses` store provided', () => {
    const responses = writable<Record<string, string>>({});
    const node = {
      type: 'decision_group',
      response_key: 'k',
      heading: 'DG',
      widget: { type: 'choice', options: ['a', 'b'] },
    };
    const target = mount(AstNode, { node, responses });
    expect(target.querySelector('[data-testid="decision-group"]')).not.toBeNull();
    expect(target.querySelector('.choice-widget')).not.toBeNull();
  });

  it('AC1_astnode_falls_back_to_markdown_without_responses — passive callers get markdown', () => {
    const node = {
      type: 'decision_group',
      heading: 'DG',
      prompt: 'pick one',
      widget: { type: 'choice', options: ['a', 'b'] },
    };
    const target = mount(AstNode, { node });
    expect(target.querySelector('[data-testid="decision-group"]')).toBeNull();
    expect(target.querySelector('[data-testid="markdown-block"]')).not.toBeNull();
  });

  it('FIX15_astnode_applies_isGroupActive — active prop flows to DecisionGroup', () => {
    const responses = writable<Record<string, string>>({});
    const node = {
      type: 'decision_group',
      response_key: 'k',
      heading: 'DG',
      widget: { type: 'choice', options: ['a', 'b'] },
    };
    const target = mount(AstNode, {
      node,
      responses,
      isGroupActive: () => true,
    });
    expect(target.querySelector('.decision-group.is-active')).not.toBeNull();
  });

  it('AC1_astnode_applies_isGroupDisabled — disabled prop flows to DecisionGroup', () => {
    const responses = writable<Record<string, string>>({});
    const node = {
      type: 'decision_group',
      response_key: 'k',
      heading: 'DG',
      widget: { type: 'choice', options: ['a', 'b'] },
    };
    const target = mount(AstNode, {
      node,
      responses,
      isGroupDisabled: () => true,
    });
    expect(target.querySelector('.decision-group.is-disabled')).not.toBeNull();
  });
});
