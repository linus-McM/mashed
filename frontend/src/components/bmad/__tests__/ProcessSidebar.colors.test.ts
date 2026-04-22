/**
 * skills-watch-02: Design-system token compliance for ProcessSidebar.svelte.
 *
 * Asserts that no hex color literals (#xxx, #xxxxxx) or rgba() calls with
 * raw integer values appear in the <style> block. All colors must use
 * var(--…) or color-mix(in srgb, var(--…) …).
 */
import { describe, it, expect } from 'vitest';
import { resolve, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';

import { auditTokens } from './tokenAudit';

const __dirname = dirname(fileURLToPath(import.meta.url));
const COMPONENT_PATH = resolve(__dirname, '../ProcessSidebar.svelte');

describe('ProcessSidebar — design token compliance', () => {
  const audit = auditTokens(COMPONENT_PATH, 'style');

  it('contains a <style> block', () => {
    expect(audit.styleBlock.length).toBeGreaterThan(0);
  });

  it('has zero hex color literals (#xxx, #xxxxxx, #xxxxxxxx) outside var() fallbacks', () => {
    expect(audit.hexLiterals).toEqual([]);
  });

  it('has zero rgba() calls with raw integer values', () => {
    // `auditTokens` already strips `rgba(var(--…), …)` — only raw-digit calls remain.
    expect(audit.rgbLiterals.length).toBe(0);
  });
});
