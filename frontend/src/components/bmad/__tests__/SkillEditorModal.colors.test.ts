/**
 * AC-4: Design-system tokens only — zero hex/rgba in SkillEditorModal.svelte.
 *
 * Reads the raw source of SkillEditorModal.svelte and asserts that no hex
 * color literals (#xxx, #xxxxxx, #xxxxxxxx) or rgba(…) calls appear in its
 * <style> block. All colors must resolve through var(--…) or
 * color-mix(in srgb, var(--…) …).
 */
import { describe, it, expect } from 'vitest';
import { resolve, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';

import { auditTokens } from './tokenAudit';

const __dirname = dirname(fileURLToPath(import.meta.url));
const COMPONENT_PATH = resolve(__dirname, '../SkillEditorModal.svelte');

describe('SkillEditorModal — AC-4 design token compliance', () => {
  const audit = auditTokens(COMPONENT_PATH, 'style');

  it('contains a <style> block', () => {
    expect(audit.styleBlock.length).toBeGreaterThan(0);
  });

  it('has zero hex color literals (#xxx, #xxxxxx, #xxxxxxxx)', () => {
    expect(audit.hexLiterals).toEqual([]);
  });

  it('has zero rgba() calls with raw integer values', () => {
    expect(audit.rgbLiterals.length).toBe(0);
  });
});
