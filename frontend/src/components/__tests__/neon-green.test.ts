import { describe, it, expect, beforeAll, afterAll } from 'vitest';
import { readFileSync, readdirSync, statSync } from 'node:fs';
import { resolve, join } from 'node:path';

// Story uiqa-02 — Eradicate the legacy neon green hex.
// Covers AC-1..AC-6 and the 7 BDD scenarios. Tests read source files from
// disk so they exercise the actual shipped code, and inject style.css as a
// <style> tag so jsdom's getComputedStyle sees the :root tokens.
//
// IMPORTANT: This file must NOT contain the forbidden hex literal anywhere
// in its source — otherwise the AC-1 walker would flag this file and the
// repo-wide grep gate would fail. The hex is assembled from char codes.

const FORBIDDEN_HEX = String.fromCharCode(0x33, 0x39, 0x66, 0x66, 0x31, 0x34); // six chars
const FORBIDDEN_RE = new RegExp('#?' + FORBIDDEN_HEX, 'i');
const FORBIDDEN_RGB_RE = /rgba\(\s*57\s*,\s*255\s*,\s*20/;

const FRONTEND_SRC = resolve(__dirname, '../..');
const STYLE_CSS_PATH = resolve(FRONTEND_SRC, 'style.css');

const STATUS_BADGE = resolve(FRONTEND_SRC, 'components/StatusBadge.svelte');
const NOTIFICATION_FEED = resolve(FRONTEND_SRC, 'views/NotificationFeed.svelte');
// R31: NotificationFeed was split into components/feed/* children (and
// lib/feed/repoTree.ts). Scoped markup/CSS moved with the child that renders
// it, so feed assertions read the parent plus its children as one source.
const FEED_FILES = [
  NOTIFICATION_FEED,
  ...['RepoHeader', 'AgentList', 'RepoActions', 'CommitOutputPanel'].map((n) =>
    resolve(FRONTEND_SRC, `components/feed/${n}.svelte`),
  ),
  resolve(FRONTEND_SRC, 'lib/feed/repoTree.ts'),
];
const AGENT_DETAIL = resolve(FRONTEND_SRC, 'views/AgentDetail.svelte');
const REPO_CONTEXT_BAR = resolve(
  FRONTEND_SRC,
  'components/bmad/RepoContextBar.svelte',
);
const TITLE_BAR = resolve(FRONTEND_SRC, 'components/TitleBar.svelte');
const SETTINGS = resolve(FRONTEND_SRC, 'views/Settings.svelte');

const read = (p: string) => readFileSync(p, 'utf8');
const readFeed = () => FEED_FILES.map(read).join('\n');

function walk(dir: string, exts: Set<string>): string[] {
  const out: string[] = [];
  for (const entry of readdirSync(dir)) {
    if (entry === 'node_modules' || entry === '__tests__') continue;
    const full = join(dir, entry);
    const st = statSync(full);
    if (st.isDirectory()) {
      out.push(...walk(full, exts));
    } else {
      const dot = entry.lastIndexOf('.');
      if (dot >= 0 && exts.has(entry.slice(dot))) out.push(full);
    }
  }
  return out;
}

let injectedStyle: HTMLStyleElement;

beforeAll(() => {
  injectedStyle = document.createElement('style');
  injectedStyle.textContent = read(STYLE_CSS_PATH);
  document.head.appendChild(injectedStyle);
});

afterAll(() => {
  injectedStyle?.remove();
});

const getToken = (name: string): string =>
  getComputedStyle(document.documentElement).getPropertyValue(name).trim();

// Mirror StatusBadge's binding (style="--status-color: X") and read the
// custom property back. jsdom does not resolve var() in computed `color`,
// so we verify the binding chain via getPropertyValue('--status-color')
// plus the token check (--accent-green / --accent-teal) on :root.
function withStatusSpan(statusColorValue: string, fn: (span: HTMLSpanElement) => void) {
  const span = document.createElement('span');
  span.setAttribute('style', `--status-color: ${statusColorValue};`);
  document.body.appendChild(span);
  try {
    fn(span);
  } finally {
    span.remove();
  }
}

const statusCustomProp = (el: Element) =>
  getComputedStyle(el).getPropertyValue('--status-color').trim();

function statusMapValue(source: string, key: 'running' | 'open'): string {
  const match = source.match(new RegExp(`${key}:\\s*'([^']+)'`));
  expect(match, `${key} must be defined`).not.toBeNull();
  return match![1];
}

describe('uiqa-02: eradicate legacy neon green hex', () => {
  describe('AC-1: zero forbidden hex matches in frontend/src', () => {
    it('AC-1: no .svelte/.ts/.js file under frontend/src contains the forbidden hex', () => {
      const files = walk(FRONTEND_SRC, new Set(['.svelte', '.ts', '.js']));
      const offenders: string[] = [];
      for (const f of files) {
        if (FORBIDDEN_RE.test(read(f))) offenders.push(f);
      }
      expect(offenders, `offenders: ${offenders.join(', ')}`).toEqual([]);
    });
  });

  describe('AC-2: JS statusColors maps use CSS variable strings', () => {
    it("AC-2: StatusBadge.svelte has running: 'var(--accent-green)'", () => {
      const src = read(STATUS_BADGE);
      expect(src).toMatch(/running:\s*'var\(--accent-green\)'/);
      expect(src).not.toMatch(new RegExp("running:\\s*'#" + FORBIDDEN_HEX + "'"));
    });

    it("AC-2: StatusBadge.svelte has open: 'var(--accent-teal)'", () => {
      const src = read(STATUS_BADGE);
      expect(src).toMatch(/open:\s*'var\(--accent-teal\)'/);
      expect(src).not.toMatch(/open:\s*'#00c4b3'/);
    });

    it('AC-2: NotificationFeed statusColors uses CSS var strings', () => {
      const src = readFeed();
      expect(src).toMatch(/running:\s*'var\(--accent-green\)'/);
      expect(src).toMatch(/open:\s*'var\(--accent-teal\)'/);
    });
  });

  describe('AC-3: badge --status-color resolves through CSS vars', () => {
    it('AC-3: running maps to var(--accent-green); --accent-green token === #00e57a', () => {
      const value = statusMapValue(read(STATUS_BADGE), 'running');
      expect(value).toBe('var(--accent-green)');
      // .badge CSS uses color: var(--status-color) — verify the binding target.
      expect(read(STATUS_BADGE)).toMatch(/color:\s*var\(--status-color\)/);
      expect(getToken('--accent-green').toLowerCase()).toBe('#00e57a');
      // jsdom keeps custom properties literal — confirm the token reference flows in.
      withStatusSpan(value, (span) => {
        expect(statusCustomProp(span)).toBe('var(--accent-green)');
      });
    });

    it('AC-3: open maps to var(--accent-teal); --accent-teal token === #00c4b3', () => {
      const value = statusMapValue(read(STATUS_BADGE), 'open');
      expect(value).toBe('var(--accent-teal)');
      expect(getToken('--accent-teal').toLowerCase()).toBe('#00c4b3');
      withStatusSpan(value, (span) => {
        expect(statusCustomProp(span)).toBe('var(--accent-teal)');
      });
    });
  });

  describe('AC-4: theme override propagates to all affected sites', () => {
    it('AC-4: overriding --accent-green on :root changes the resolved token', () => {
      expect(getToken('--accent-green').toLowerCase()).toBe('#00e57a');
      const override = document.createElement('style');
      override.textContent = ':root { --accent-green: #ff00ff; }';
      document.head.appendChild(override);
      try {
        // The token resolves to the override — this is the root of the cascade
        // that StatusBadge / .action-hot / .back-btn / .git-hot / RepoContextBar
        // all consume via var(--accent-green).
        expect(getToken('--accent-green').toLowerCase()).toBe('#ff00ff');
      } finally {
        override.remove();
      }
      expect(getToken('--accent-green').toLowerCase()).toBe('#00e57a');
    });

    it('AC-4: NotificationFeed hot buttons use var(--accent-green) via .glow-btn', () => {
      const src = readFeed();
      // uiqa-04 replaced .action-hot with the shared .glow-btn utility.
      // NotificationFeed must no longer contain #39ff14, and its template
      // must bind class:glow-btn (hot-button equivalent). The .glow-btn
      // rule body (verified in glow-btn.test.ts) references --accent-green.
      expect(FORBIDDEN_RE.test(src)).toBe(false);
      expect(src).toMatch(/class:glow-btn/);
    });

    it('AC-4: AgentDetail .back-btn / .git-hot use var(--accent-green)', () => {
      const src = read(AGENT_DETAIL);
      const backBlocks = src.match(/\.back-btn[^{]*\{[^}]*\}/g) ?? [];
      expect(backBlocks.length).toBeGreaterThan(0);
      for (const block of backBlocks) {
        expect(FORBIDDEN_RE.test(block)).toBe(false);
        // .back-btn glow must use color-mix, not the wrong-RGB raw rgba.
        expect(FORBIDDEN_RGB_RE.test(block)).toBe(false);
      }
      expect(
        backBlocks.some(
          (b) => /color-mix/.test(b) && /var\(--accent-green\)/.test(b),
        ),
        'back-btn must use color-mix with --accent-green',
      ).toBe(true);

      // Note: .git-hot was removed in uiqa-04 and replaced with the shared
      // .glow-btn utility (see glow-btn.test.ts for equivalent assertions).
      // The AgentDetail hot-button color is now enforced via style.css's
      // .glow-btn rule, not a per-component .git-hot rule.
    });

    it('AC-4: RepoContextBar .back-btn uses var(--accent-green)', () => {
      const src = read(REPO_CONTEXT_BAR);
      expect(FORBIDDEN_RE.test(src)).toBe(false);
      expect(src).toMatch(/var\(--accent-green\)/);
    });
  });

  describe('AC-5: .color-swatch.active uses --text-primary', () => {
    it('AC-5: NotificationFeed .color-swatch.active border-color references --text-primary', () => {
      const src = readFeed();
      const match = src.match(/\.color-swatch\.active\s*\{[^}]*\}/);
      expect(match, '.color-swatch.active rule must exist').not.toBeNull();
      const block = match![0];
      expect(block).toMatch(/border-color:\s*var\(--text-primary\)/);
      expect(block).not.toMatch(/border-color:\s*#fff\b/);
    });
  });

  describe('AC-6: acceptable exceptions preserved', () => {
    it.each([
      ['#ff5f57'],
      ['#febc2e'],
      ['#28c840'],
    ])('AC-6: TitleBar.svelte still contains %s', (hex) => {
      expect(read(TITLE_BAR)).toContain(hex);
    });

    it.each([
      ['#ff5f57'],
      ['#febc2e'],
      ['#28c840'],
    ])('AC-6: Settings.svelte still contains %s', (hex) => {
      expect(read(SETTINGS)).toContain(hex);
    });
  });
});
