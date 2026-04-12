/**
 * skills-watch-02: Design-system token compliance for ProcessSidebar.svelte.
 *
 * Asserts that no hex color literals (#xxx, #xxxxxx) or rgba() calls with
 * raw integer values appear in the <style> block. All colors must use
 * var(--…) or color-mix(in srgb, var(--…) …).
 */
import { describe, it, expect } from 'vitest';
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';

const COMPONENT_PATH = resolve(__dirname, '../ProcessSidebar.svelte');

function extractStyleBlock(source) {
  const match = source.match(/<style[\s>][\s\S]*?<\/style>/i);
  return match ? match[0] : '';
}

describe('ProcessSidebar — design token compliance', () => {
  const source = readFileSync(COMPONENT_PATH, 'utf8');
  const styleBlock = extractStyleBlock(source);

  it('contains a <style> block', () => {
    expect(styleBlock.length).toBeGreaterThan(0);
  });

  it('has zero hex color literals (#xxx, #xxxxxx, #xxxxxxxx) outside var() fallbacks', () => {
    const lines = styleBlock.split('\n');
    const violations = [];
    for (let i = 0; i < lines.length; i++) {
      const line = lines[i];
      // Skip lines inside url("data:...") — SVG chevron etc.
      if (/url\s*\(/.test(line)) continue;
      // Strip var(--…, #fallback) before scanning — fallbacks are acceptable
      const stripped = line.replace(/var\(--[\w-]+,\s*#[0-9a-fA-F]{3,8}\)/g, '');
      const hexMatch = stripped.match(/#[0-9a-fA-F]{3,8}\b/g);
      if (hexMatch) {
        violations.push({ line: i + 1, matches: hexMatch, text: line.trim() });
      }
    }
    expect(violations).toEqual([]);
  });

  it('has zero rgba() calls with raw integer values', () => {
    const rgbaPattern = /rgba\s*\(\s*\d/g;
    const matches = styleBlock.match(rgbaPattern);
    expect(matches).toBeNull();
  });
});
