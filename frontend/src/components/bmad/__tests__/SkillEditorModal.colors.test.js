/**
 * AC-4: Design-system tokens only — zero hex/rgba in SkillEditorModal.svelte.
 *
 * This vitest spec reads the raw source of SkillEditorModal.svelte and
 * asserts that no hex color literals (#xxx, #xxxxxx, #xxxxxxxx) or
 * rgba(...) calls appear in its <style> block. All colors must resolve
 * through var(--…) or color-mix(in srgb, var(--…) …).
 */
import { describe, it, expect } from 'vitest';
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';

const COMPONENT_PATH = resolve(
  __dirname,
  '../SkillEditorModal.svelte',
);

function extractStyleBlock(source) {
  const match = source.match(/<style[\s>][\s\S]*?<\/style>/i);
  return match ? match[0] : '';
}

describe('SkillEditorModal — AC-4 design token compliance', () => {
  const source = readFileSync(COMPONENT_PATH, 'utf8');
  const styleBlock = extractStyleBlock(source);

  it('contains a <style> block', () => {
    expect(styleBlock.length).toBeGreaterThan(0);
  });

  it('has zero hex color literals (#xxx, #xxxxxx, #xxxxxxxx)', () => {
    // Match hex colors but exclude SVG data-URI encoded colors (stroke='%23xxx')
    // and CSS url() values that may contain encoded hex
    const lines = styleBlock.split('\n');
    const violations = [];

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i];
      // Skip lines that are inside url("data:...") — SVG chevron etc.
      if (/url\s*\(/.test(line)) continue;
      // Match standalone hex colors: #rgb, #rrggbb, #rrggbbaa
      const hexMatch = line.match(/#[0-9a-fA-F]{3,8}\b/g);
      if (hexMatch) {
        violations.push({ line: i + 1, matches: hexMatch, text: line.trim() });
      }
    }

    expect(violations).toEqual([]);
  });

  it('has zero rgba() calls with raw integer values', () => {
    // rgba(R, G, B, A) with numeric args — var(--…) inside rgba is fine
    const rgbaPattern = /rgba\s*\(\s*\d/g;
    const matches = styleBlock.match(rgbaPattern);
    expect(matches).toBeNull();
  });
});
