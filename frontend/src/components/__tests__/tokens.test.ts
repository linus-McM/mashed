import { describe, it, expect, beforeAll, afterAll } from 'vitest';
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';

// Story uiqa-01 — Design System Token Foundation.
// Covers AC-1 (tokens exist in :root), AC-2 (tokens resolve at runtime),
// AC-3 (DESIGN.md documents each token). Maps to BDD scenarios 1-5.
// style.css is read from disk and injected as a <style> tag so jsdom's
// getComputedStyle sees the :root custom properties regardless of how
// vitest handles CSS imports.

const STYLE_CSS_PATH = resolve(__dirname, '../../style.css');
const DESIGN_MD_PATH = resolve(__dirname, '../../../../DESIGN.md');

let injectedStyle: HTMLStyleElement;

beforeAll(() => {
  injectedStyle = document.createElement('style');
  injectedStyle.textContent = readFileSync(STYLE_CSS_PATH, 'utf8');
  document.head.appendChild(injectedStyle);
});

afterAll(() => {
  injectedStyle?.remove();
});

const getToken = (name: string): string =>
  getComputedStyle(document.documentElement).getPropertyValue(name).trim();

describe('uiqa-01: design system token foundation', () => {
  describe('AC-1/AC-2: new tokens resolve from :root via getComputedStyle', () => {
    it('AC-1/AC-2: --accent-cyan resolves to #22d3ee', () => {
      expect(getToken('--accent-cyan')).toBe('#22d3ee');
    });

    it('AC-1/AC-2: --accent-orange resolves to #fb923c', () => {
      expect(getToken('--accent-orange')).toBe('#fb923c');
    });

    it('AC-1/AC-2: --overlay-backdrop resolves to rgba(0, 0, 0, 0.6)', () => {
      // jsdom normalizes rgba whitespace; compare ignoring spaces.
      expect(getToken('--overlay-backdrop').replace(/\s+/g, '')).toBe(
        'rgba(0,0,0,0.6)',
      );
    });

    it('AC-1/AC-2: --glow-spread resolves to 0 0 12px', () => {
      expect(getToken('--glow-spread')).toBe('0 0 12px');
    });
  });

  describe('AC-3: DESIGN.md documents all four tokens', () => {
    const designMd = readFileSync(DESIGN_MD_PATH, 'utf8');

    it.each([
      ['--accent-cyan'],
      ['--accent-orange'],
      ['--overlay-backdrop'],
      ['--glow-spread'],
    ])('AC-3: DESIGN.md documents %s', (tokenName) => {
      expect(designMd).toContain(tokenName);
    });

    it('AC-3: DESIGN.md shows the --glow-spread usage example', () => {
      expect(designMd).toMatch(
        /text-shadow:\s*var\(--glow-spread\)\s*var\(--accent-green\)/,
      );
    });
  });
});
