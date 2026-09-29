/**
 * @vitest-environment jsdom
 */
// Tests for RawViewToggle.svelte — story ui-ast-U8. Covers AC-1..AC-4 + AC-11.
import { describe, it, expect } from 'vitest';
import { tick } from 'svelte';

import RawViewToggle from '../RawViewToggle.svelte';
import { makeMount } from './mountSvelte';

type Props = {
  raw: string;
  expandedByDefault?: boolean;
  triggered?: boolean;
};

describe('RawViewToggle', () => {
  const mount = makeMount();
  const render = (props: Props) => mount(RawViewToggle, props);

  it('AC-1 collapses by default when triggered is false', () => {
    const el = render({ raw: 'some captured output', triggered: false });
    const details = el.querySelector<HTMLDetailsElement>('details');
    expect(details, 'details element must render when raw is non-empty').not.toBeNull();
    expect(details!.hasAttribute('open')).toBe(false);
    expect(details!.open).toBe(false);
  });

  it('AC-2 toggle opens on summary click (unit portion)', async () => {
    const el = render({ raw: 'captured output body', triggered: false });
    const details = el.querySelector<HTMLDetailsElement>('details')!;
    const summary = details.querySelector<HTMLElement>('summary')!;
    expect(details.open).toBe(false);
    summary.click();
    await tick();
    expect(details.open).toBe(true);
  });

  it('AC-3 hidden-when-empty — no details element when raw is ""', () => {
    const el = render({ raw: '' });
    expect(el.querySelector('details')).toBeNull();
  });

  describe('AC-4 expanded-when-both-flags-true', () => {
    const cases: Array<{ expandedByDefault: boolean; triggered: boolean; open: boolean }> = [
      { expandedByDefault: true, triggered: true, open: true },
      { expandedByDefault: true, triggered: false, open: false },
      { expandedByDefault: false, triggered: true, open: false },
      { expandedByDefault: false, triggered: false, open: false },
    ];
    for (const c of cases) {
      it(`AC-4 expandedByDefault=${c.expandedByDefault} triggered=${c.triggered} → open=${c.open}`, () => {
        const el = render({
          raw: 'captured',
          expandedByDefault: c.expandedByDefault,
          triggered: c.triggered,
        });
        const details = el.querySelector<HTMLDetailsElement>('details')!;
        expect(details.hasAttribute('open')).toBe(c.open);
      });
    }
  });

  it('AC-11 raw content is HTML-escaped — script tag renders as literal text', () => {
    const payload = '<script>window.__rvt_pwned=1</script>hello';
    const el = render({ raw: payload, expandedByDefault: true, triggered: true });
    // No script element must make it into the DOM.
    expect(el.querySelector('script')).toBeNull();
    expect((window as unknown as { __rvt_pwned?: number }).__rvt_pwned).toBeUndefined();
    // The payload must appear verbatim inside <pre>.
    const pre = el.querySelector('pre');
    expect(pre, 'pre element renders raw content').not.toBeNull();
    expect(pre!.textContent).toContain('<script>window.__rvt_pwned=1</script>');
    expect(pre!.textContent).toContain('hello');
  });
});
