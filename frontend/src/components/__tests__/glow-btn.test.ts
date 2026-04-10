import { describe, it, expect, beforeAll, afterAll } from 'vitest';
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';

// Story uiqa-04 — Shared .glow-btn class. Covers AC-1..AC-6 and all 7 BDD
// scenarios. Reads source files from disk and injects style.css into jsdom
// so the class applies on mounted elements. Pattern mirrors neon-green.test.ts.

const FRONTEND_SRC = resolve(__dirname, '../..');
const STYLE_CSS = resolve(FRONTEND_SRC, 'style.css');
const NOTIFICATION_FEED = resolve(FRONTEND_SRC, 'views/NotificationFeed.svelte');
const AGENT_DETAIL = resolve(FRONTEND_SRC, 'views/AgentDetail.svelte');

const read = (p: string) => readFileSync(p, 'utf8');

// Isolate a component's <style> block so we don't accidentally match class
// names that appear in the template's `class:X={}` bindings.
function styleBlock(src: string): string {
  const m = src.match(/<style[^>]*>([\s\S]*?)<\/style>/);
  return m ? m[1] : '';
}

function templateOnly(src: string): string {
  const block = styleBlock(src);
  return block ? src.replace(block, '') : src;
}

let injectedStyle: HTMLStyleElement;
beforeAll(() => {
  injectedStyle = document.createElement('style');
  injectedStyle.textContent = read(STYLE_CSS);
  document.head.appendChild(injectedStyle);
});
afterAll(() => {
  injectedStyle?.remove();
});

const getToken = (name: string): string =>
  getComputedStyle(document.documentElement).getPropertyValue(name).trim();

function mountButton(): HTMLButtonElement {
  const btn = document.createElement('button');
  btn.className = 'glow-btn';
  document.body.appendChild(btn);
  return btn;
}

describe('uiqa-04: shared .glow-btn class', () => {
  describe('AC-1: .glow-btn declared in style.css with accent-green + glow-spread', () => {
    it('AC-1 (BDD#1): style.css contains a rule matching /^\\.glow-btn\\s*\\{/m', () => {
      expect(read(STYLE_CSS)).toMatch(/^\.glow-btn\s*\{/m);
    });

    it('AC-1 (BDD#1): .glow-btn rule body references var(--accent-green)', () => {
      const match = read(STYLE_CSS).match(/\.glow-btn\s*\{([^}]*)\}/);
      expect(match, '.glow-btn rule must exist').not.toBeNull();
      expect(match![1]).toMatch(/var\(--accent-green\)/);
    });

    it('AC-1 (BDD#1): .glow-btn rule (or hover) references var(--glow-spread)', () => {
      // token may appear in the base rule or in :hover — check combined blocks.
      const combined =
        read(STYLE_CSS).match(/\.glow-btn[^{]*\{[^}]*\}/g)?.join('\n') ?? '';
      expect(combined).toMatch(/var\(--glow-spread\)/);
    });

    it('AC-1 (BDD#2): .glow-btn:hover rule exists', () => {
      expect(read(STYLE_CSS)).toMatch(/\.glow-btn:hover\s*\{/);
    });

    it('AC-1 (BDD#2): .glow-btn:focus-visible rule exists', () => {
      expect(read(STYLE_CSS)).toMatch(/\.glow-btn:focus-visible\s*\{/);
    });
  });

  describe('AC-2: .action-hot removed from NotificationFeed + glow-btn binding present', () => {
    it('AC-2 (BDD#3): NotificationFeed <style> block has zero ".action-hot" selectors', () => {
      expect(styleBlock(read(NOTIFICATION_FEED))).not.toMatch(/\.action-hot/);
    });

    it('AC-2 (BDD#3): NotificationFeed template has at least one class:glow-btn binding', () => {
      expect(templateOnly(read(NOTIFICATION_FEED))).toMatch(/class:glow-btn/);
    });
  });

  describe('AC-3: .git-hot removed from AgentDetail + glow-btn binding present', () => {
    it('AC-3 (BDD#4): AgentDetail <style> block has zero ".git-hot" selectors', () => {
      expect(styleBlock(read(AGENT_DETAIL))).not.toMatch(/\.git-hot/);
    });

    it('AC-3 (BDD#4): AgentDetail template has at least one class:glow-btn binding', () => {
      expect(templateOnly(read(AGENT_DETAIL))).toMatch(/class:glow-btn/);
    });
  });

  describe('AC-4: mounted .glow-btn resolves accent color + non-empty glow', () => {
    it('AC-4 (BDD#5): computed color on mounted <button class="glow-btn"> resolves via accent-green', () => {
      const btn = mountButton();
      try {
        const color = getComputedStyle(btn).color;
        // jsdom may return the literal "var(--accent-green)" OR the resolved rgb.
        // Either proves the rule applied. An unstyled button returns '' or the
        // inherited default — neither would contain accent-green.
        expect(color).not.toBe('');
        const accentRgb = color.replace(/\s+/g, '');
        const isAccent =
          /var\(--accent-green\)/.test(color) || accentRgb === 'rgb(0,229,122)';
        expect(isAccent, `computed color was: ${color}`).toBe(true);
      } finally {
        btn.remove();
      }
    });

    it('AC-4 (BDD#5): mounted .glow-btn has a non-empty text-shadow (glow present)', () => {
      const btn = mountButton();
      try {
        const shadow = getComputedStyle(btn).textShadow;
        expect(shadow).not.toBe('');
        expect(shadow).not.toBe('none');
      } finally {
        btn.remove();
      }
    });
  });

  describe('AC-5: theme override propagates to .glow-btn via --accent-green', () => {
    it('AC-5 (BDD#6): overriding --accent-green on :root changes the cascaded token', () => {
      expect(getToken('--accent-green').toLowerCase()).toBe('#00e57a');
      const override = document.createElement('style');
      override.textContent = ':root { --accent-green: #ff00ff; }';
      document.head.appendChild(override);
      try {
        expect(getToken('--accent-green').toLowerCase()).toBe('#ff00ff');
        // .glow-btn references var(--accent-green), so the cascade is proven
        // by the token flip + the rule-body assertion in AC-1.
        expect(read(STYLE_CSS)).toMatch(/\.glow-btn[^}]*var\(--accent-green\)/);
      } finally {
        override.remove();
      }
      expect(getToken('--accent-green').toLowerCase()).toBe('#00e57a');
    });
  });

  describe('AC-6: hot-button count conserved (6 pre → 6 post)', () => {
    it('AC-6: NotificationFeed has exactly 3 class:glow-btn bindings (was 3 class:action-hot)', () => {
      const matches = templateOnly(read(NOTIFICATION_FEED)).match(/class:glow-btn/g) ?? [];
      expect(matches.length).toBe(3);
    });

    it('AC-6: AgentDetail has exactly 3 class:glow-btn bindings (was 3 class:git-hot)', () => {
      const matches = templateOnly(read(AGENT_DETAIL)).match(/class:glow-btn/g) ?? [];
      expect(matches.length).toBe(3);
    });

    it('AC-6: total glow-btn count (6) matches pre-refactor hot-button count (6)', () => {
      const nf = (templateOnly(read(NOTIFICATION_FEED)).match(/class:glow-btn/g) ?? []).length;
      const ad = (templateOnly(read(AGENT_DETAIL)).match(/class:glow-btn/g) ?? []).length;
      expect(nf + ad).toBe(6);
    });
  });

  describe('BDD#7: focus ring (focus-visible outline) defined', () => {
    it('BDD#7: .glow-btn:focus-visible declares a non-zero outline width', () => {
      const match = read(STYLE_CSS).match(/\.glow-btn:focus-visible\s*\{([^}]*)\}/);
      expect(match, '.glow-btn:focus-visible rule must exist').not.toBeNull();
      // outline shorthand or outline-width — require a digit 1-9 indicating thickness.
      expect(match![1]).toMatch(/outline(?:-width)?:\s*[^;]*[1-9]/);
    });
  });
});
