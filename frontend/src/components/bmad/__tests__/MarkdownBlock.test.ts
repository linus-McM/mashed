/**
 * @vitest-environment jsdom
 */
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

vi.mock('../../../../wailsjs/runtime/runtime.js', () => ({
  BrowserOpenURL: vi.fn(),
}));

import MarkdownBlock from '../MarkdownBlock.svelte';
import { BrowserOpenURL } from '../../../../wailsjs/runtime/runtime.js';
import { makeMount } from './mountSvelte';

describe('MarkdownBlock', () => {
  const mount = makeMount();
  const render = (content: string) => mount(MarkdownBlock, { content });

  it('renders paragraph text', () => {
    const el = render('Hello world');
    expect(el.textContent).toContain('Hello world');
    expect(el.querySelector('p')).not.toBeNull();
  });

  it('renders headings (h1/h2/h3)', () => {
    const el = render('# H1\n\n## H2\n\n### H3');
    expect(el.querySelector('h1')?.textContent).toContain('H1');
    expect(el.querySelector('h2')?.textContent).toContain('H2');
    expect(el.querySelector('h3')?.textContent).toContain('H3');
  });

  it('renders inline code', () => {
    const el = render('use `npm test`');
    const code = el.querySelector('code');
    expect(code?.textContent).toBe('npm test');
  });

  it('renders bullet lists', () => {
    const el = render('- a\n- b\n- c');
    expect(el.querySelectorAll('ul li').length).toBe(3);
  });

  it('renders ordered lists', () => {
    const el = render('1. a\n2. b');
    expect(el.querySelectorAll('ol li').length).toBe(2);
  });

  // XSS guard: the markdown renderer must NOT enable html: true.
  it('escapes raw HTML instead of rendering it', () => {
    const el = render('<script>window.__pwned=1</script>hello');
    expect(el.querySelector('script')).toBeNull();
    expect((window as unknown as { __pwned?: number }).__pwned).toBeUndefined();
    expect(el.textContent).toContain('hello');
  });

  describe('§7.2 link sanitisation', () => {
    beforeEach(() => {
      vi.mocked(BrowserOpenURL).mockClear();
    });

    it('AC6_javascript_href_stripped — javascript: links degrade to text', () => {
      const el = render('[click me](javascript:alert(1))');
      expect(el.textContent).toContain('click me');
      const anchor = el.querySelector('a');
      // Anchor tag may render but must not carry the javascript: href.
      if (anchor) {
        expect(anchor.getAttribute('href')).not.toMatch(/javascript:/i);
      }
      const hrefs = Array.from(el.querySelectorAll('a')).map((a) => a.getAttribute('href') ?? '');
      for (const h of hrefs) {
        expect(h).not.toMatch(/javascript:/i);
      }
    });

    it('AC7_data_and_file_schemes_rejected — data: / file: degrade to text', () => {
      const el = render('[x](data:text/html,<x>) and [y](file:///etc/passwd)');
      expect(el.textContent).toContain('x');
      expect(el.textContent).toContain('y');
      const hrefs = Array.from(el.querySelectorAll('a')).map((a) => a.getAttribute('href') ?? '');
      for (const h of hrefs) {
        expect(h).not.toMatch(/^data:/i);
        expect(h).not.toMatch(/^file:/i);
      }
    });

    it('AC8_https_link_routes_through_confirm — confirm=true invokes BrowserOpenURL', () => {
      const confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(true);
      const el = render('[docs](https://example.com)');
      const anchor = el.querySelector('a');
      expect(anchor).not.toBeNull();
      anchor!.click();
      expect(confirmSpy).toHaveBeenCalledTimes(1);
      expect(confirmSpy.mock.calls[0][0]).toContain('https://example.com');
      expect(BrowserOpenURL).toHaveBeenCalledTimes(1);
      expect(vi.mocked(BrowserOpenURL).mock.calls[0][0]).toContain('https://example.com');
      confirmSpy.mockRestore();
    });

    it('AC8_https_link_routes_through_confirm — confirm=false does NOT invoke BrowserOpenURL', () => {
      const confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(false);
      const el = render('[docs](https://example.com)');
      const anchor = el.querySelector('a');
      anchor!.click();
      expect(confirmSpy).toHaveBeenCalledTimes(1);
      expect(BrowserOpenURL).not.toHaveBeenCalled();
      confirmSpy.mockRestore();
    });
  });
});
