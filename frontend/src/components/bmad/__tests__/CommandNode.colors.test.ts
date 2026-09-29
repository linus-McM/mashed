// skills-cmd-02 AC-4: CommandNode.svelte uses ONLY design-system tokens.
//
// Zero hex literals, zero rgb/rgba literals, and (pinned by Cerebrum
// Do-Not-Repeat 2026-04-10) zero `#39ff14` — the brand green is
// `var(--accent-green: #00e57a)`, never the toxic neon.

import { describe, it, expect, beforeAll } from 'vitest';
import { resolve, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';

import { auditTokens, type TokenAuditResult } from './tokenAudit';

const __dirname = dirname(fileURLToPath(import.meta.url));
const COMMAND_NODE = resolve(__dirname, '..', 'CommandNode.svelte');

describe('skills-cmd-02 AC-4: CommandNode.svelte uses only design-system tokens', () => {
  let audit: TokenAuditResult;

  beforeAll(() => {
    audit = auditTokens(COMMAND_NODE, 'full');
  });

  it('CommandNode.svelte exists', () => {
    expect(audit.exists, `expected ${COMMAND_NODE} (RED until GREEN ships)`).toBe(true);
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
