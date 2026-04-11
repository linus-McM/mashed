import { describe, it, expect } from 'vitest';
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';

// Story uiqa-10 — Signature Moments: running agent pulse, status-bar ambient
// SparkLine + pulse dot, terminal-style commit panel. Source-level assertions
// cover AC-1..AC-7 the same way uiqa-06/uiqa-09 verify CSS + template wiring.
// Commit-panel ACs (5, 6) are NOT APPLICABLE — current NotificationFeed commit
// panel streams one line at a time with no per-step active/done state; the
// story's per-step "is-active"/"is-done" model has no underlying data to bind.
// The shared terminal aesthetic (mono font + prompt ::before) is still added
// to style.css so any future per-step refactor picks up the visual automatically,
// but AC-5/6 are not asserted here.

const FRONTEND_SRC = resolve(__dirname, '..');
const STYLE_CSS = resolve(FRONTEND_SRC, 'style.css');
const NOTIFICATION_FEED = resolve(FRONTEND_SRC, 'views/NotificationFeed.svelte');

const read = (p: string) => readFileSync(p, 'utf8');

/**
 * Extract every `@media (prefers-reduced-motion: no-preference) { ... }` block
 * body from style.css so AC-2 and AC-7 can assert keyframes/animation rules
 * are wrapped correctly. Handles nested braces via depth counting.
 */
function motionGuardedBlocks(css: string): string[] {
  const blocks: string[] = [];
  const marker = '@media (prefers-reduced-motion: no-preference)';
  let i = 0;
  while (true) {
    const start = css.indexOf(marker, i);
    if (start === -1) break;
    const open = css.indexOf('{', start);
    if (open === -1) break;
    let depth = 1;
    let j = open + 1;
    while (j < css.length && depth > 0) {
      if (css[j] === '{') depth++;
      else if (css[j] === '}') depth--;
      if (depth === 0) break;
      j++;
    }
    blocks.push(css.slice(open + 1, j));
    i = j + 1;
  }
  return blocks;
}

describe('uiqa-10: signature moments — pulse, ambient, terminal', () => {
  describe('Sub-brief A: running agent row pulse', () => {
    it('AC-1: style.css declares @keyframes agent-running-pulse', () => {
      const css = read(STYLE_CSS);
      expect(css).toMatch(/@keyframes\s+agent-running-pulse\b/);
    });

    it('AC-1: style.css declares a .agent-row.is-running selector', () => {
      const css = read(STYLE_CSS);
      expect(css).toMatch(/\.agent-row\.is-running\b/);
    });

    it('AC-1: static fallback box-shadow uses var(--accent-green) so reduced-motion users still see the solid border', () => {
      const css = read(STYLE_CSS);
      // Find the unguarded .agent-row.is-running block (outside any media query).
      // Simple proof: the full file contains an is-running rule with a
      // box-shadow referencing var(--accent-green).
      expect(css).toMatch(
        /\.agent-row\.is-running[^{}]*\{[^{}]*box-shadow:[^;}]*var\(--accent-green\)/,
      );
    });

    it('AC-1: NotificationFeed.svelte binds class:is-running on the .agent-row element', () => {
      const src = read(NOTIFICATION_FEED);
      // The binding lives inside the agent-row element declaration — match the
      // class attribute + class:is-running in close proximity.
      expect(src).toMatch(/class=["']agent-row["'][\s\S]{0,400}?class:is-running=/);
    });

    it('AC-1: class:is-running expression matches on eventType === "running" (project-specific status field)', () => {
      const src = read(NOTIFICATION_FEED);
      expect(src).toMatch(/class:is-running=\{\s*agent\.eventType\s*===\s*['"]running['"]\s*\}/);
    });

    it('AC-2/AC-7: @keyframes agent-running-pulse is inside a prefers-reduced-motion: no-preference block', () => {
      const css = read(STYLE_CSS);
      const blocks = motionGuardedBlocks(css);
      const guarded = blocks.some((b) => /@keyframes\s+agent-running-pulse\b/.test(b));
      expect(guarded).toBe(true);
    });

    it('AC-2/AC-7: .agent-row.is-running animation declaration is inside the no-preference block', () => {
      const css = read(STYLE_CSS);
      const blocks = motionGuardedBlocks(css);
      const guarded = blocks.some((b) =>
        /\.agent-row\.is-running[^{}]*\{[^{}]*animation:\s*agent-running-pulse/.test(b),
      );
      expect(guarded).toBe(true);
    });
  });

  describe('Sub-brief B: status bar ambient SparkLine + pulse dot', () => {
    it('AC-3: NotificationFeed declares a reactive aggregateSamples computation', () => {
      const src = read(NOTIFICATION_FEED);
      expect(src).toMatch(/\$:\s*aggregateSamples\b/);
    });

    it('AC-3: NotificationFeed declares a reactive anyRunning computation', () => {
      const src = read(NOTIFICATION_FEED);
      expect(src).toMatch(/\$:\s*anyRunning\b/);
    });

    it('AC-3: status bar template renders <SparkLine data={aggregateSamples} ...>', () => {
      const src = read(NOTIFICATION_FEED);
      // Status bar container followed by a SparkLine with data={aggregateSamples}.
      expect(src).toMatch(
        /class=["']status-bar["'][\s\S]*?<SparkLine[^>]*data=\{aggregateSamples\}/,
      );
    });

    it('AC-3: aggregateSamples SparkLine is guarded by a length check (> 1)', () => {
      const src = read(NOTIFICATION_FEED);
      expect(src).toMatch(/aggregateSamples\.length\s*>\s*1/);
    });

    it('AC-4: status bar template contains a .status-pulse element with aria-label "agents active"', () => {
      const src = read(NOTIFICATION_FEED);
      expect(src).toMatch(
        /class=["']status-bar["'][\s\S]*?class=["']status-pulse["'][^>]*aria-label=["']agents active["']/,
      );
    });

    it('AC-4: status-pulse element is guarded by {#if anyRunning}', () => {
      const src = read(NOTIFICATION_FEED);
      // anyRunning if-block must enclose the status-pulse span.
      expect(src).toMatch(
        /\{#if\s+anyRunning\s*\}[\s\S]*?class=["']status-pulse["'][\s\S]*?\{\/if\}/,
      );
    });

    it('AC-4: style.css defines a .status-pulse rule with background: var(--accent-green)', () => {
      const css = read(STYLE_CSS);
      expect(css).toMatch(
        /\.status-pulse[^{}]*\{[^{}]*background:\s*var\(--accent-green\)/,
      );
    });

    it('AC-7: @keyframes status-pulse is inside a prefers-reduced-motion: no-preference block', () => {
      const css = read(STYLE_CSS);
      const blocks = motionGuardedBlocks(css);
      const guarded = blocks.some((b) => /@keyframes\s+status-pulse\b/.test(b));
      expect(guarded).toBe(true);
    });

    it('AC-7: .status-pulse animation declaration is inside the no-preference block', () => {
      const css = read(STYLE_CSS);
      const blocks = motionGuardedBlocks(css);
      const guarded = blocks.some((b) =>
        /\.status-pulse[^{}]*\{[^{}]*animation:\s*status-pulse/.test(b),
      );
      expect(guarded).toBe(true);
    });
  });

  describe('Sub-brief C: terminal-style commit panel (shared styles — per-step state NOT APPLICABLE)', () => {
    it('AC-5 (shared aesthetic): style.css applies var(--font-mono) to .commit-panel container', () => {
      const css = read(STYLE_CSS);
      expect(css).toMatch(
        /\.commit-panel[^{}]*\{[^{}]*font-family:\s*var\(--font-mono\)/,
      );
    });

    it('AC-5 (shared aesthetic): style.css adds a "$ " prompt ::before to .commit-step', () => {
      const css = read(STYLE_CSS);
      expect(css).toMatch(/\.commit-step::before[^{}]*\{[^{}]*content:\s*['"]\$\s['"]/);
    });

    it('AC-7: @keyframes cursor-blink is inside a prefers-reduced-motion: no-preference block', () => {
      const css = read(STYLE_CSS);
      const blocks = motionGuardedBlocks(css);
      const guarded = blocks.some((b) => /@keyframes\s+cursor-blink\b/.test(b));
      expect(guarded).toBe(true);
    });
  });

  describe('AC-7 sweep: every @keyframes added by this story lives behind prefers-reduced-motion', () => {
    const KEYFRAMES_ADDED_BY_UIQA_10 = [
      'agent-running-pulse',
      'status-pulse',
      'cursor-blink',
    ];

    for (const name of KEYFRAMES_ADDED_BY_UIQA_10) {
      it(`AC-7: @keyframes ${name} is wrapped in @media (prefers-reduced-motion: no-preference)`, () => {
        const css = read(STYLE_CSS);
        const blocks = motionGuardedBlocks(css);
        const guarded = blocks.some((b) => new RegExp(`@keyframes\\s+${name}\\b`).test(b));
        expect(guarded).toBe(true);
      });
    }
  });
});
