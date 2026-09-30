// skills-cmd-03: CanvasFailureToast.svelte uses ONLY design-system tokens.
//
// Zero hex literals, zero rgb/rgba literals, and (pinned by Cerebrum
// Do-Not-Repeat 2026-04-10) zero `#39ff14`. The same regex contract as
// CommandNode.colors.test.ts — this file exists so a future drift in the
// toast's color strategy fails loudly before it reaches a user's screen.

import { describe, it, expect, beforeAll } from 'vitest';
import { resolve, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';

import { auditTokens, type TokenAuditResult } from './tokenAudit';

const __dirname = dirname(fileURLToPath(import.meta.url));
const TOAST = resolve(__dirname, '..', 'CanvasFailureToast.svelte');

describe('skills-cmd-03: CanvasFailureToast.svelte uses only design-system tokens', () => {
  let audit: TokenAuditResult;

  beforeAll(() => {
    audit = auditTokens(TOAST, 'full');
  });

  it('CanvasFailureToast.svelte exists', () => {
    expect(audit.exists).toBe(true);
  });

  it('contains zero hex color literals', () => {
    if (!audit.source) return; // gated by existence test above
    expect(
      audit.hexLiterals,
      `found hex literals: ${[...new Set(audit.hexLiterals)].join(', ')}`,
    ).toEqual([]);
  });

  it('contains zero rgb/rgba literals', () => {
    if (!audit.source) return;
    expect(audit.rgbLiterals.length).toBe(0);
  });

  it('contains zero occurrences of the toxic neon hex (Cerebrum Do-Not-Repeat 2026-04-10)', () => {
    if (!audit.source) return;
    expect(audit.containsToxicNeon).toBe(false);
  });
});
