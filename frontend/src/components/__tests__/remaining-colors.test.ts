import { describe, it, expect, beforeAll, afterAll } from 'vitest';
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';

// Story uiqa-03 — Remaining hardcoded colors.
// Covers AC-1..AC-6 and the 7 BDD scenarios.
//
// Approach: source-text assertions for the structural ACs (the file content
// is the contract), plus jsdom-style + custom-property checks for the
// theme-override AC. jsdom does NOT resolve var() inside computed colors,
// so the theme-override scenario verifies the binding chain through
// getPropertyValue (same approach as uiqa-02 commit 885f377).

const FRONTEND_SRC = resolve(__dirname, '../..');

const EXECUTION_BAR = resolve(FRONTEND_SRC, 'components/bmad/ExecutionBar.svelte');
const AGENT_DETAIL = resolve(FRONTEND_SRC, 'views/AgentDetail.svelte');
const SETTINGS = resolve(FRONTEND_SRC, 'views/Settings.svelte');
const PROCESS_NODE = resolve(FRONTEND_SRC, 'components/bmad/ProcessNode.svelte');
const WORKFLOW_BUILDER = resolve(FRONTEND_SRC, 'views/WorkflowBuilder.svelte');
const TITLE_BAR = resolve(FRONTEND_SRC, 'components/TitleBar.svelte');
const STYLE_CSS_PATH = resolve(FRONTEND_SRC, 'style.css');

const read = (p: string) => readFileSync(p, 'utf8');

// Extract the contiguous { ... } block that follows a selector. Handles single
// nesting depth (sufficient for plain CSS rules in <style> blocks).
function extractRule(source: string, selectorRegex: RegExp): string | null {
  const m = source.match(selectorRegex);
  if (!m || m.index === undefined) return null;
  const startBrace = source.indexOf('{', m.index);
  if (startBrace < 0) return null;
  let depth = 0;
  for (let i = startBrace; i < source.length; i++) {
    const ch = source[i];
    if (ch === '{') depth++;
    else if (ch === '}') {
      depth--;
      if (depth === 0) return source.slice(startBrace, i + 1);
    }
  }
  return null;
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

describe('uiqa-03: remaining hardcoded colors', () => {
  describe('AC-1: ExecutionBar uses --accent-cyan and --accent-orange', () => {
    it('AC-1: ExecutionBar.svelte contains zero #22d3ee literals', () => {
      const src = read(EXECUTION_BAR);
      expect(src).not.toMatch(/#22d3ee/i);
    });

    it('AC-1: ExecutionBar.svelte contains zero #fb923c literals', () => {
      const src = read(EXECUTION_BAR);
      expect(src).not.toMatch(/#fb923c/i);
    });

    it('AC-1: ExecutionBar.svelte contains zero rgba(34, 211, 238, ...) literals', () => {
      const src = read(EXECUTION_BAR);
      expect(src).not.toMatch(/rgba\(\s*34\s*,\s*211\s*,\s*238/);
    });

    it('AC-1: ExecutionBar.svelte contains zero rgba(251, 146, 60, ...) literals', () => {
      const src = read(EXECUTION_BAR);
      expect(src).not.toMatch(/rgba\(\s*251\s*,\s*146\s*,\s*60/);
    });

    it('AC-1: .ctrl-btn.run rule references var(--accent-cyan)', () => {
      const src = read(EXECUTION_BAR);
      const block = extractRule(src, /\.ctrl-btn\.run\s*\{/);
      expect(block, '.ctrl-btn.run rule must exist').not.toBeNull();
      expect(block!).toMatch(/var\(--accent-cyan\)/);
    });

    it('AC-1: .ctrl-btn.run:hover rule references var(--accent-cyan) (color-mix or var)', () => {
      const src = read(EXECUTION_BAR);
      const block = extractRule(src, /\.ctrl-btn\.run:hover[^{]*\{/);
      expect(block, '.ctrl-btn.run:hover rule must exist').not.toBeNull();
      expect(block!).toMatch(/var\(--accent-cyan\)/);
    });

    it('AC-1: .ctrl-btn.stop rule references var(--accent-orange)', () => {
      const src = read(EXECUTION_BAR);
      const block = extractRule(src, /\.ctrl-btn\.stop\s*\{/);
      expect(block, '.ctrl-btn.stop rule must exist').not.toBeNull();
      expect(block!).toMatch(/var\(--accent-orange\)/);
    });

    it('AC-1: .ctrl-btn.stop:hover rule references var(--accent-orange)', () => {
      const src = read(EXECUTION_BAR);
      const block = extractRule(src, /\.ctrl-btn\.stop:hover[^{]*\{/);
      expect(block, '.ctrl-btn.stop:hover rule must exist').not.toBeNull();
      expect(block!).toMatch(/var\(--accent-orange\)/);
    });
  });

  describe('AC-2: AgentDetail tab-close hover uses --accent-red', () => {
    it('AC-2: AgentDetail.svelte contains zero #ff5f57 literals', () => {
      const src = read(AGENT_DETAIL);
      expect(src).not.toMatch(/#ff5f57/i);
    });

    it('AC-2: .tab-close:hover declares color: var(--accent-red)', () => {
      const src = read(AGENT_DETAIL);
      const block = extractRule(src, /\.tab-close:hover[^{]*\{/);
      expect(block, '.tab-close:hover rule must exist').not.toBeNull();
      expect(block!).toMatch(/color:\s*var\(--accent-red\)/);
    });
  });

  describe('AC-3: Settings indicators and text-shadow use tokens', () => {
    it('AC-3: Settings.svelte contains zero #565670 literals', () => {
      const src = read(SETTINGS);
      expect(src).not.toMatch(/#565670/i);
    });

    it('AC-3: Settings.svelte contains zero #c0c0d0 literals', () => {
      const src = read(SETTINGS);
      expect(src).not.toMatch(/#c0c0d0/i);
    });

    it('AC-3: Settings.svelte contains zero rgba(0, 229, 122, ...) literals', () => {
      const src = read(SETTINGS);
      expect(src).not.toMatch(/rgba\(\s*0\s*,\s*229\s*,\s*122/);
    });

    it('AC-3: .import-indicator.dark background references var(--text-muted)', () => {
      const src = read(SETTINGS);
      const block = extractRule(src, /\.import-indicator\.dark\s*\{/);
      expect(block, '.import-indicator.dark rule must exist').not.toBeNull();
      expect(block!).toMatch(/background:\s*var\(--text-muted\)/);
    });

    it('AC-3: .import-indicator.light background references var(--text-dim)', () => {
      const src = read(SETTINGS);
      const block = extractRule(src, /\.import-indicator\.light\s*\{/);
      expect(block, '.import-indicator.light rule must exist').not.toBeNull();
      expect(block!).toMatch(/background:\s*var\(--text-dim\)/);
    });

    it('AC-3: .back-btn:hover text-shadow uses var(--glow-spread) and color-mix with --accent-green', () => {
      const src = read(SETTINGS);
      const block = extractRule(src, /\.back-btn:hover\s*\{/);
      expect(block, '.back-btn:hover rule must exist').not.toBeNull();
      expect(block!).toMatch(/text-shadow:[^;]*var\(--glow-spread\)/);
      expect(block!).toMatch(/color-mix\([^)]*var\(--accent-green\)/);
    });
  });

  describe('AC-4: ProcessNode and WorkflowBuilder use color-mix with tokens', () => {
    it('AC-4: ProcessNode.svelte contains zero rgba(0, 229, 122, 0.35) literals', () => {
      const src = read(PROCESS_NODE);
      expect(src).not.toMatch(/rgba\(\s*0\s*,\s*229\s*,\s*122\s*,\s*0\.35\s*\)/);
    });

    it('AC-4: .process-node.selected box-shadow uses color-mix with --accent-green', () => {
      const src = read(PROCESS_NODE);
      const block = extractRule(src, /\.process-node\.selected\s*\{/);
      expect(block, '.process-node.selected rule must exist').not.toBeNull();
      expect(block!).toMatch(/box-shadow:[^;]*color-mix\([^)]*var\(--accent-green\)[^)]*35%/);
    });

    it('AC-4: WorkflowBuilder.svelte contains zero rgba(248, 81, 73, ...) literals', () => {
      const src = read(WORKFLOW_BUILDER);
      expect(src).not.toMatch(/rgba\(\s*248\s*,\s*81\s*,\s*73/);
    });

    it('AC-4: .exec-error background uses color-mix with --accent-red at 10%', () => {
      const src = read(WORKFLOW_BUILDER);
      const block = extractRule(src, /\.exec-error\s*\{/);
      expect(block, '.exec-error rule must exist').not.toBeNull();
      expect(block!).toMatch(/background:[^;]*color-mix\([^)]*var\(--accent-red\)[^)]*10%/);
    });

    it('AC-4: .exec-error-dismiss:hover background uses color-mix with --accent-red at 15%', () => {
      const src = read(WORKFLOW_BUILDER);
      const block = extractRule(src, /\.exec-error-dismiss:hover\s*\{/);
      expect(block, '.exec-error-dismiss:hover rule must exist').not.toBeNull();
      expect(block!).toMatch(/background:[^;]*color-mix\([^)]*var\(--accent-red\)[^)]*15%/);
    });
  });

  describe('AC-5: theme override propagates', () => {
    // jsdom does not resolve var() in computed `color`, so we follow the
    // uiqa-02 pattern: verify the underlying token resolves to its override
    // and that the source binds the rule to the token. The cascade is then
    // observed in the live browser via /playwright-cli during QG.

    it('AC-5: --accent-cyan token defaults to #22d3ee and accepts overrides', () => {
      expect(getToken('--accent-cyan').toLowerCase()).toBe('#22d3ee');
      const override = document.createElement('style');
      override.textContent = ':root { --accent-cyan: #00ffff; }';
      document.head.appendChild(override);
      try {
        expect(getToken('--accent-cyan').toLowerCase()).toBe('#00ffff');
      } finally {
        override.remove();
      }
      expect(getToken('--accent-cyan').toLowerCase()).toBe('#22d3ee');
    });

    it('AC-5: --accent-orange token defaults to #fb923c and accepts overrides', () => {
      expect(getToken('--accent-orange').toLowerCase()).toBe('#fb923c');
      const override = document.createElement('style');
      override.textContent = ':root { --accent-orange: #ffa500; }';
      document.head.appendChild(override);
      try {
        expect(getToken('--accent-orange').toLowerCase()).toBe('#ffa500');
      } finally {
        override.remove();
      }
      expect(getToken('--accent-orange').toLowerCase()).toBe('#fb923c');
    });

    it('AC-5: --accent-red token defaults to #e84545 and accepts overrides', () => {
      expect(getToken('--accent-red').toLowerCase()).toBe('#e84545');
      const override = document.createElement('style');
      override.textContent = ':root { --accent-red: #ff0000; }';
      document.head.appendChild(override);
      try {
        expect(getToken('--accent-red').toLowerCase()).toBe('#ff0000');
      } finally {
        override.remove();
      }
      expect(getToken('--accent-red').toLowerCase()).toBe('#e84545');
    });

    it('AC-5: ExecutionBar binds .ctrl-btn.run color to var(--accent-cyan)', () => {
      const src = read(EXECUTION_BAR);
      const block = extractRule(src, /\.ctrl-btn\.run\s*\{/);
      expect(block!).toMatch(/color:\s*var\(--accent-cyan\)/);
    });

    it('AC-5: ExecutionBar binds .ctrl-btn.stop color to var(--accent-orange)', () => {
      const src = read(EXECUTION_BAR);
      const block = extractRule(src, /\.ctrl-btn\.stop\s*\{/);
      expect(block!).toMatch(/color:\s*var\(--accent-orange\)/);
    });
  });

  describe('AC-6: acceptable macOS chrome exceptions still present', () => {
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
    ])('AC-6: Settings.svelte theme preview still contains %s', (hex) => {
      expect(read(SETTINGS)).toContain(hex);
    });
  });
});
