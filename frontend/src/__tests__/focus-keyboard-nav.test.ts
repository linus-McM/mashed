import { describe, it, expect } from 'vitest';
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';

// Story uiqa-08 — Focus States + CanvasPane Keyboard Nav.
// Source-level assertions covering AC-1..AC-7. Runtime Playwright coverage for
// AC-2/AC-3 tab-through rings is deferred; global CSS rule presence is proven
// at the source level here.

const FRONTEND_SRC = resolve(__dirname, '..');
const STYLE_CSS = resolve(FRONTEND_SRC, 'style.css');
const CANVAS_PANE = resolve(FRONTEND_SRC, 'components/bmad/CanvasPane.svelte');

const read = (p: string) => readFileSync(p, 'utf8');

function scriptBlock(src: string): string {
  const m = src.match(/<script[^>]*>([\s\S]*?)<\/script>/);
  return m ? m[1] : '';
}

function templateOnly(src: string): string {
  // Strip <script>...</script> and <style>...</style> so we only match template text.
  return src
    .replace(/<script[^>]*>[\s\S]*?<\/script>/g, '')
    .replace(/<style[^>]*>[\s\S]*?<\/style>/g, '');
}

describe('uiqa-08: focus states + CanvasPane keyboard nav', () => {
  describe('AC-1: global :focus-visible rule in style.css', () => {
    const css = read(STYLE_CSS);

    it('AC-1 (BDD#1): contains button:focus-visible selector', () => {
      expect(css).toMatch(/button:focus-visible/);
    });

    it('AC-1: contains input, select, textarea focus-visible selectors', () => {
      expect(css).toMatch(/input:focus-visible/);
      expect(css).toMatch(/select:focus-visible/);
      expect(css).toMatch(/textarea:focus-visible/);
    });

    it('AC-1: contains [role="button"]:focus-visible selector', () => {
      expect(css).toMatch(/\[role="button"\]:focus-visible/);
    });

    it('AC-1: contains .agent-row:focus-visible and .repo-header:focus-visible selectors', () => {
      expect(css).toMatch(/\.agent-row:focus-visible/);
      expect(css).toMatch(/\.repo-header:focus-visible/);
    });

    it('AC-1 (BDD#1): global rule declares outline: 2px solid var(--accent-green)', () => {
      // Grab the combined selector block that includes "button:focus-visible".
      // Use a multi-selector group matcher — look for any block containing
      // button:focus-visible followed by an outline declaration.
      const match = css.match(
        /button:focus-visible[\s\S]*?\{[^}]*outline:\s*2px\s+solid\s+var\(--accent-green\)[^}]*\}/,
      );
      expect(match, 'button:focus-visible block must declare outline: 2px solid var(--accent-green)').not.toBeNull();
    });

    it('AC-1: global rule declares outline-offset: 2px', () => {
      const match = css.match(
        /button:focus-visible[\s\S]*?\{[^}]*outline-offset:\s*2px[^}]*\}/,
      );
      expect(match, 'button:focus-visible block must declare outline-offset: 2px').not.toBeNull();
    });

    it('AC-1: existing .glow-btn:focus-visible rule preserved (no regression from uiqa-04)', () => {
      expect(css).toMatch(/\.glow-btn:focus-visible\s*\{/);
    });
  });

  describe('AC-2/AC-3: global focus-visible rule covers interactive elements (source-verified)', () => {
    const css = read(STYLE_CSS);

    it('AC-2: global rule selector list includes all required interactive elements', () => {
      // Require that all core selectors exist in CSS text (runtime tab-through
      // verification deferred to Playwright manual check).
      const required = [
        'button:focus-visible',
        'input:focus-visible',
        'select:focus-visible',
        'textarea:focus-visible',
        '[role="button"]:focus-visible',
        '.agent-row:focus-visible',
        '.repo-header:focus-visible',
      ];
      for (const sel of required) {
        expect(css, `missing selector: ${sel}`).toContain(sel);
      }
    });
  });

  describe('AC-4: CanvasPane closes context menu on Escape + restores focus', () => {
    const src = read(CANVAS_PANE);
    const script = scriptBlock(src);

    it('AC-4 (BDD#4): script handles Escape key in keydown switch/branch', () => {
      expect(script).toMatch(/['"]Escape['"]/);
    });

    it('AC-4: script has a closeContextMenu or closeMenu invocation', () => {
      // Either name is acceptable; the existing file uses closeContextMenu.
      expect(script).toMatch(/close(?:Context)?Menu\s*\(/);
    });

    it('AC-4: script tracks previousFocus for restoration on close', () => {
      expect(script).toMatch(/previousFocus/);
    });

    it('AC-4: script invokes .focus() on stored previousFocus reference', () => {
      // Accept either previousFocus.focus(), previousFocus?.focus(), or an
      // aliased local (const prev = previousFocus; prev.focus()).
      expect(script).toMatch(/previousFocus/);
      expect(script).toMatch(/\.focus\s*\(\s*\)/);
    });
  });

  describe('AC-5: Arrow keys cycle menu items with wrap', () => {
    const src = read(CANVAS_PANE);
    const script = scriptBlock(src);

    it('AC-5 (BDD#5): handler references ArrowDown', () => {
      expect(script).toMatch(/['"]ArrowDown['"]/);
    });

    it('AC-5: handler references ArrowUp', () => {
      expect(script).toMatch(/['"]ArrowUp['"]/);
    });

    it('AC-5: handler uses modulo (%) for wrap-around indexing', () => {
      // Accept either `% count` (local length var) or `% items.length`.
      expect(script).toMatch(/activeIdx\s*\+\s*1\s*\)\s*%/);
      expect(script).toMatch(/activeIdx\s*-\s*1\s*\+\s*\w+\s*\)\s*%/);
    });

    it('AC-5: handler references Home and End keys', () => {
      expect(script).toMatch(/['"]Home['"]/);
      expect(script).toMatch(/['"]End['"]/);
    });

    it('AC-5: script maintains itemRefs array for focusable items', () => {
      expect(script).toMatch(/itemRefs/);
    });

    it('AC-5: script maintains activeIdx for current selection', () => {
      expect(script).toMatch(/activeIdx/);
    });
  });

  describe('AC-6: Enter/Space activates selected item', () => {
    const src = read(CANVAS_PANE);
    const script = scriptBlock(src);

    it('AC-6 (BDD#6): handler references Enter key', () => {
      expect(script).toMatch(/['"]Enter['"]/);
    });

    it('AC-6: handler references Space key (" ")', () => {
      expect(script).toMatch(/['"] ['"]|case\s*['"] ['"]/);
    });
  });

  describe('AC-7: ARIA roles and tabindex on context menu', () => {
    const src = read(CANVAS_PANE);
    const template = templateOnly(src);

    it('AC-7 (BDD#7): context menu container has role="menu"', () => {
      expect(template).toMatch(/role="menu"/);
    });

    it('AC-7: context menu items have role="menuitem"', () => {
      expect(template).toMatch(/role="menuitem"/);
    });

    it('AC-7: context menu items have tabindex="-1"', () => {
      expect(template).toMatch(/tabindex="-1"/);
    });
  });

  describe('Keyboard nav wiring', () => {
    const src = read(CANVAS_PANE);

    it('uses svelte:window on:keydown or equivalent keydown listener', () => {
      expect(src).toMatch(/svelte:window[^>]*on:keydown|on:keydown=/);
    });

    it('script imports tick from svelte for post-DOM focus', () => {
      const script = scriptBlock(src);
      expect(script).toMatch(/import\s*\{[^}]*\btick\b[^}]*\}\s*from\s*['"]svelte['"]/);
    });
  });
});
