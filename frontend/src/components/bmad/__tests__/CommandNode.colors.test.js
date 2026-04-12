// skills-cmd-02 AC-4: CommandNode.svelte uses ONLY design-system tokens.
//
// Zero hex literals, zero rgb/rgba literals, and (pinned by Cerebrum
// Do-Not-Repeat 2026-04-10) zero `#39ff14` — the brand green is
// `var(--accent-green: #00e57a)`, never the toxic neon.
//
// RED state: `CommandNode.svelte` does not exist yet. The first assertion
// gates on existence so the regex assertions trip ONLY post-GREEN drift.

import { describe, it, expect, beforeAll } from 'vitest';
import { existsSync, readFileSync } from 'node:fs';
import { resolve, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const COMMAND_NODE = resolve(__dirname, '..', 'CommandNode.svelte');

// Obfuscated so this file does not self-trip the future repo-wide hex
// walker (mirrors the precedent in neon-green.test.ts).
const TOXIC_NEON =
  '#' + String.fromCharCode(51, 57, 102, 102, 49, 52); // #39ff14

describe('skills-cmd-02 AC-4: CommandNode.svelte uses only design-system tokens', () => {
  let source = '';

  beforeAll(() => {
    if (existsSync(COMMAND_NODE)) {
      source = readFileSync(COMMAND_NODE, 'utf8');
    }
  });

  it('CommandNode.svelte exists', () => {
    expect(
      existsSync(COMMAND_NODE),
      `expected ${COMMAND_NODE} (RED until GREEN ships)`,
    ).toBe(true);
  });

  it('contains zero hex color literals', () => {
    if (!source) return; // gated by existence test above
    const hits = source.match(/#[0-9a-fA-F]{3,8}\b/g) || [];
    expect(hits, `found hex literals: ${[...new Set(hits)].join(', ')}`).toEqual([]);
  });

  it('contains zero rgb/rgba literals', () => {
    if (!source) return;
    const hits = source.match(/\brgba?\s*\(/gi) || [];
    expect(hits.length).toBe(0);
  });

  it(`contains zero occurrences of the toxic neon hex (Cerebrum Do-Not-Repeat 2026-04-10)`, () => {
    if (!source) return;
    expect(source.includes(TOXIC_NEON)).toBe(false);
  });
});
