/**
 * @vitest-environment jsdom
 */
import { describe, it, expect } from 'vitest';

import SummaryCard from '../SummaryCard.svelte';
import { makeMount } from './mountSvelte';

describe('SummaryCard', () => {
  const mount = makeMount();
  const render = (heading: string, bullets: string[]) =>
    mount(SummaryCard, { heading, bullets });

  it('renders the heading', () => {
    expect(render('Next steps', ['ship it']).textContent).toContain('Next steps');
  });

  it('renders every bullet from the bullets prop', () => {
    const el = render('Checklist', ['alpha', 'beta', 'gamma']);
    const items = Array.from(el.querySelectorAll<HTMLElement>('li')).map((li) => li.textContent);
    expect(items.length).toBe(3);
    expect(items.some((t) => t?.includes('alpha'))).toBe(true);
    expect(items.some((t) => t?.includes('beta'))).toBe(true);
    expect(items.some((t) => t?.includes('gamma'))).toBe(true);
  });

  it('renders an empty bullet list without crashing', () => {
    const el = render('Empty', []);
    expect(el.querySelectorAll('li').length).toBe(0);
    expect(el.textContent).toContain('Empty');
  });
});
