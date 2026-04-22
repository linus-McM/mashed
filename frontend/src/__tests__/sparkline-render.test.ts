import { describe, it, expect } from 'vitest';
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';

// Story uiqa-09 — Render SparkLine in NotificationFeed.
// Grep-based contract tests that ensure SparkLine is actually wired into
// the agent row template with the correct guard and props. These tests run
// without mounting Svelte so they're resilient to any reactive runtime
// differences between dev, build, and vitest environments.

const FRONTEND_SRC = resolve(__dirname, '..');
const NOTIFICATION_FEED = resolve(FRONTEND_SRC, 'views/NotificationFeed.svelte');
const SPARKLINE = resolve(FRONTEND_SRC, 'components/SparkLine.svelte');

const read = (p: string) => readFileSync(p, 'utf8');

describe('uiqa-09 SparkLine render in NotificationFeed', () => {
  const feed = read(NOTIFICATION_FEED);

  it('AC-7: keeps SparkLine import and uses it at least once', () => {
    // Import site
    expect(feed).toMatch(/import\s+SparkLine\s+from\s+['"][^'"]*SparkLine\.svelte['"]/);
    // At least one SparkLine tag in the template
    expect(feed).toMatch(/<SparkLine\b/);
  });

  it('AC-4: passes tokenSamples array as the data prop', () => {
    // The component call site must bind data={agent.tokenSamples}
    expect(feed).toMatch(/<SparkLine[^>]*data=\{agent\.tokenSamples\}/);
  });

  it('AC-4/AC-5: guards rendering on tokenSamples?.length > 1', () => {
    // The `{#if}` block gating the SparkLine must require more than one sample
    // so brand-new sessions with 0 or 1 points hide the sparkline cleanly.
    expect(feed).toMatch(/tokenSamples\?\.length\s*>\s*1/);
  });

  it('does NOT pass width or height props (SparkLine is a text component)', () => {
    // Defensive: the earlier spec wrote width/height which SparkLine ignores.
    // Fail the test if those reappear — they'd be dead props and a misleading API.
    const sparklineTag = feed.match(/<SparkLine\b[^>]*>/g) ?? [];
    for (const tag of sparklineTag) {
      expect(tag).not.toMatch(/\bwidth=/);
      expect(tag).not.toMatch(/\bheight=/);
    }
  });

  it('SparkLine component still exposes only the text-based props API', () => {
    // Guard against accidental refactor: props must be exactly `data` and `maxVal`.
    // Regex accepts optional TypeScript type annotations post-retyping (Phase 4c).
    const spark = read(SPARKLINE);
    expect(spark).toMatch(/export\s+let\s+data(?:\s*:\s*number\[\])?\s*=\s*\[\]/);
    expect(spark).toMatch(/export\s+let\s+maxVal(?:\s*:\s*number)?\s*=\s*0/);
    // Ensure it is still a span, not an SVG.
    expect(spark).toMatch(/<span\s+class="sparkline"/);
  });
});
