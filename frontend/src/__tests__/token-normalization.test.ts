import { describe, it, expect } from 'vitest';
import { readFileSync, readdirSync, statSync, existsSync } from 'node:fs';
import { resolve, join } from 'node:path';
import { execSync } from 'node:child_process';

// Story uiqa-07 — Token scale normalization.
// Covers AC-1..AC-6. Walks frontend/src/{views,components} and scans *.svelte
// files for orphaned font-size and spacing literals. Excludes __tests__/ and
// style.css (token definition).

const FRONTEND_SRC = resolve(__dirname, '..');
const VIEWS_DIR = resolve(FRONTEND_SRC, 'views');
const COMPONENTS_DIR = resolve(FRONTEND_SRC, 'components');
const FRONTEND_ROOT = resolve(FRONTEND_SRC, '..');
const LINTER_SCRIPT = resolve(FRONTEND_ROOT, 'scripts/check-orphan-tokens.mjs');

const read = (p: string) => readFileSync(p, 'utf8');

function walk(dir: string, out: string[] = []): string[] {
  if (!existsSync(dir)) return out;
  for (const entry of readdirSync(dir)) {
    const p = join(dir, entry);
    const st = statSync(p);
    if (st.isDirectory()) {
      if (entry === '__tests__' || entry === 'node_modules') continue;
      walk(p, out);
    } else if (entry.endsWith('.svelte')) {
      out.push(p);
    }
  }
  return out;
}

function grepHits(pattern: RegExp, files: string[]): string[] {
  const hits: string[] = [];
  for (const f of files) {
    const src = read(f);
    const lines = src.split('\n');
    lines.forEach((line, idx) => {
      if (pattern.test(line)) {
        hits.push(`${f}:${idx + 1}  ${line.trim()}`);
      }
      // reset lastIndex on global regexes
      pattern.lastIndex = 0;
    });
  }
  return hits;
}

const SVELTE_FILES = [...walk(VIEWS_DIR), ...walk(COMPONENTS_DIR)];

describe('uiqa-07: token scale normalization', () => {
  describe('AC-1: font-size outliers eliminated', () => {
    it('AC-1a (BDD#1): zero font-size: 10px in views/ and components/', () => {
      const hits = grepHits(/font-size:\s*10px/, SVELTE_FILES);
      expect(hits, `offenders:\n${hits.join('\n')}`).toHaveLength(0);
    });

    it('AC-1b (BDD#2): zero font-size: 12px in views/ and components/', () => {
      const hits = grepHits(/font-size:\s*12px/, SVELTE_FILES);
      expect(hits, `offenders:\n${hits.join('\n')}`).toHaveLength(0);
    });

    it('AC-1c: zero font-size: 15px in views/ and components/', () => {
      const hits = grepHits(/font-size:\s*15px/, SVELTE_FILES);
      expect(hits, `offenders:\n${hits.join('\n')}`).toHaveLength(0);
    });

    it('AC-1d (BDD#3): zero font-size: 18px in views/ and components/', () => {
      const hits = grepHits(/font-size:\s*18px/, SVELTE_FILES);
      expect(hits, `offenders:\n${hits.join('\n')}`).toHaveLength(0);
    });
  });

  describe('AC-2: spacing outliers eliminated', () => {
    it('AC-2a: zero padding|margin|gap with 3px literal', () => {
      const hits = grepHits(/(padding|margin|gap)[^:]*:\s*[^;]*\b3px\b/, SVELTE_FILES);
      expect(hits, `offenders:\n${hits.join('\n')}`).toHaveLength(0);
    });

    it('AC-2b: zero padding|margin|gap with 5px literal', () => {
      const hits = grepHits(/(padding|margin|gap)[^:]*:\s*[^;]*\b5px\b/, SVELTE_FILES);
      expect(hits, `offenders:\n${hits.join('\n')}`).toHaveLength(0);
    });

    it('AC-2c: zero padding|margin|gap with 6px literal', () => {
      const hits = grepHits(/(padding|margin|gap)[^:]*:\s*[^;]*\b6px\b/, SVELTE_FILES);
      expect(hits, `offenders:\n${hits.join('\n')}`).toHaveLength(0);
    });
  });

  describe('AC-3: AgentDetail uses --text-data token', () => {
    const AGENT_DETAIL = resolve(VIEWS_DIR, 'AgentDetail.svelte');

    it('AC-3a (BDD#4): AgentDetail.svelte has zero font-size: 14px literal', () => {
      const src = read(AGENT_DETAIL);
      expect(src).not.toMatch(/font-size:\s*14px/);
    });

    it('AC-3b: AgentDetail.svelte uses font-size: var(--text-data) at least once', () => {
      const src = read(AGENT_DETAIL);
      expect(src).toMatch(/font-size:\s*var\(--text-data\)/);
    });
  });

  describe('AC-4: numeric displays use tabular-nums', () => {
    const NOTIF = resolve(VIEWS_DIR, 'NotificationFeed.svelte');
    const AGENT = resolve(VIEWS_DIR, 'AgentDetail.svelte');

    function ruleBody(src: string, selector: string): string | null {
      const esc = selector.replace(/[-/\\^$*+?.()|[\]{}]/g, '\\$&');
      // Match rule header where the selector is the LAST selector in the list
      // (so ".agent-elapsed {" matches but ".x, .agent-elapsed.foo {" also matches).
      // We grab the smallest balanced body.
      const re = new RegExp(`${esc}\\s*\\{[^}]*\\}`, 'g');
      const m = src.match(re);
      return m ? m.join('\n') : null;
    }

    it('AC-4a (BDD#5): NotificationFeed .sub-count declares tabular-nums', () => {
      const src = read(NOTIF);
      const rule = ruleBody(src, '.sub-count');
      expect(rule, '.sub-count rule not found').not.toBeNull();
      expect(rule!).toMatch(/font-variant-numeric:\s*tabular-nums/);
    });

    it('AC-4b (BDD#6): NotificationFeed .agent-tokens declares tabular-nums', () => {
      const src = read(NOTIF);
      const rule = ruleBody(src, '.agent-tokens');
      expect(rule, '.agent-tokens rule not found').not.toBeNull();
      expect(rule!).toMatch(/font-variant-numeric:\s*tabular-nums/);
    });

    it('AC-4c: NotificationFeed .agent-elapsed declares tabular-nums', () => {
      const src = read(NOTIF);
      const rule = ruleBody(src, '.agent-elapsed');
      expect(rule, '.agent-elapsed rule not found').not.toBeNull();
      expect(rule!).toMatch(/font-variant-numeric:\s*tabular-nums/);
    });

    it('AC-4d: AgentDetail .token-mini-label declares tabular-nums', () => {
      const src = read(AGENT);
      const rule = ruleBody(src, '.token-mini-label');
      expect(rule, '.token-mini-label rule not found').not.toBeNull();
      expect(rule!).toMatch(/font-variant-numeric:\s*tabular-nums/);
    });
  });

  describe('AC-6: linter script blocks regressions', () => {
    it('AC-6 (BDD#7): scripts/check-orphan-tokens.mjs exists and exits 0', () => {
      expect(existsSync(LINTER_SCRIPT), 'linter script missing').toBe(true);
      // If the script exits non-zero, execSync throws; wrap to surface message.
      let output = '';
      try {
        output = execSync(`node ${LINTER_SCRIPT}`, { cwd: FRONTEND_ROOT, stdio: 'pipe' }).toString();
      } catch (err: any) {
        throw new Error(`linter non-zero exit:\n${err.stdout?.toString() ?? ''}\n${err.stderr?.toString() ?? ''}`);
      }
      expect(output).toBeDefined();
    });
  });
});
