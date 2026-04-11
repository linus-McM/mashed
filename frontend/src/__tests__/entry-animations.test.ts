import { describe, it, expect } from 'vitest';
import { readFileSync, existsSync } from 'node:fs';
import { resolve } from 'node:path';

// Story uiqa-06 — Entry Animations: Repo Groups & Modals.
// Source-level assertions covering AC-1..AC-5 and AC-8.
// Runtime-only ACs (AC-6 staggered opacity rise, AC-7 modal fade visible) are
// source-verified here via transition-prop literal inspection; Playwright
// runtime verification is deferred to manual check (see sign-off table in
// story notes).

const FRONTEND_SRC = resolve(__dirname, '..');

const NOTIFICATION_FEED = resolve(FRONTEND_SRC, 'views/NotificationFeed.svelte');
const NEW_SESSION_MODAL = resolve(FRONTEND_SRC, 'views/NewSessionModal.svelte');
const SPAWN_AGENT = resolve(FRONTEND_SRC, 'views/SpawnAgent.svelte');
const BRANCH_MODAL = resolve(FRONTEND_SRC, 'views/BranchModal.svelte');

const read = (p: string) => readFileSync(p, 'utf8');

function scriptBlock(src: string): string {
  // Extract the first <script>...</script> block so we don't confuse template
  // text with script-level imports.
  const m = src.match(/<script[^>]*>([\s\S]*?)<\/script>/);
  return m ? m[1] : '';
}

describe('uiqa-06: entry animations — repo groups & modals', () => {
  describe('AC-1: NotificationFeed imports fly and slide from svelte/transition', () => {
    it('AC-1 (BDD#1): script block contains fly and slide imports from svelte/transition', () => {
      const src = read(NOTIFICATION_FEED);
      const script = scriptBlock(src);
      // Allow combined or split imports; both names must resolve to svelte/transition.
      const re =
        /import\s*\{[^}]*\bfly\b[^}]*\bslide\b[^}]*\}\s*from\s*['"]svelte\/transition['"]|import\s*\{[^}]*\bslide\b[^}]*\bfly\b[^}]*\}\s*from\s*['"]svelte\/transition['"]/;
      expect(script).toMatch(re);
    });

    it('AC-1: NotificationFeed imports an easing function from svelte/easing', () => {
      const src = read(NOTIFICATION_FEED);
      const script = scriptBlock(src);
      expect(script).toMatch(
        /import\s*\{[^}]*cubic(Out|In|InOut)[^}]*\}\s*from\s*['"]svelte\/easing['"]/,
      );
    });
  });

  describe('AC-2: Repo group has transition:fly with staggered delay', () => {
    it('AC-2 (BDD#2): repo-group element has transition:fly or in:fly directive', () => {
      const src = read(NOTIFICATION_FEED);
      // Look for fly directive near the repo-group class.
      expect(src).toMatch(/class=["']repo-group["'][\s\S]{0,600}?(transition|in):fly=/);
    });

    it('AC-2: fly directive references the each-block index multiplied by 50 (stagger)', () => {
      const src = read(NOTIFICATION_FEED);
      // Accept either a literal "i * 50" expression or a helper that embeds it.
      const hasLiteralStagger = /i\s*\*\s*50/.test(src);
      expect(hasLiteralStagger).toBe(true);
    });

    it('AC-2: fly directive references duration 150 (either inline or via helper constant)', () => {
      const src = read(NOTIFICATION_FEED);
      const script = scriptBlock(src);
      // The helper pattern assigns a DUR_MEDIUM-style constant to 150.
      const hasDur150 =
        /duration:\s*150/.test(src) ||
        /=\s*reducedMotion\s*\?\s*0\s*:\s*150/.test(script) ||
        /:\s*150[^0-9]/.test(script);
      expect(hasDur150).toBe(true);
    });
  });

  describe('AC-3: Repo body uses transition:slide on expand/collapse', () => {
    it('AC-3 (BDD#3): an element inside the repo each block has transition:slide', () => {
      const src = read(NOTIFICATION_FEED);
      expect(src).toMatch(/transition:slide=/);
    });
  });

  describe('AC-4: Modals import and apply fade', () => {
    const MODAL_FILES = [
      { path: NEW_SESSION_MODAL, name: 'NewSessionModal.svelte' },
      { path: SPAWN_AGENT, name: 'SpawnAgent.svelte' },
      { path: BRANCH_MODAL, name: 'BranchModal.svelte' },
    ];

    for (const { path, name } of MODAL_FILES) {
      it(`AC-4 (BDD#4): ${name} imports fade from svelte/transition`, () => {
        expect(existsSync(path), `${name} must exist`).toBe(true);
        const src = read(path);
        const script = scriptBlock(src);
        expect(script).toMatch(
          /import\s*\{[^}]*\bfade\b[^}]*\}\s*from\s*['"]svelte\/transition['"]/,
        );
      });

      it(`AC-4: ${name} has transition:fade on the backdrop/overlay element`, () => {
        const src = read(path);
        // The class is "overlay" in all three modals — match the overlay rule.
        expect(src).toMatch(/class=["']overlay["'][\s\S]{0,600}?transition:fade=/);
      });

      it(`AC-4: ${name} has transition:fade on the modal card element`, () => {
        const src = read(path);
        expect(src).toMatch(/class=["']modal["'][\s\S]{0,600}?transition:fade=/);
      });

      it(`AC-4: ${name} modal card fade uses a 50ms delay (or reducedMotion 0)`, () => {
        const src = read(path);
        const script = scriptBlock(src);
        // Accept either an inline delay literal on the modal element OR a
        // helper constant in the script block that resolves to 50ms delay.
        const inlineDelay =
          /class=["']modal["'][\s\S]{0,600}?transition:fade=\{[^}]*delay:\s*(reducedMotion\s*\?\s*0\s*:\s*50|50)\b/.test(
            src,
          );
        const helperDelay =
          /cardFadeProps[\s\S]*?delay:\s*50\b/.test(script) ||
          /delay:\s*50\s*,\s*easing/.test(script);
        expect(inlineDelay || helperDelay).toBe(true);
      });
    }
  });

  describe('AC-5: Reduced-motion honored', () => {
    const FILES = [NOTIFICATION_FEED, NEW_SESSION_MODAL, SPAWN_AGENT, BRANCH_MODAL];

    for (const path of FILES) {
      const name = path.split('/').pop();
      it(`AC-5: ${name} references prefers-reduced-motion matchMedia query`, () => {
        const src = read(path);
        const script = scriptBlock(src);
        expect(script).toMatch(/prefers-reduced-motion/);
        expect(script).toMatch(/matchMedia/);
      });

      it(`AC-5: ${name} defines a reducedMotion variable`, () => {
        const src = read(path);
        const script = scriptBlock(src);
        expect(script).toMatch(/\breducedMotion\b/);
      });
    }
  });

  describe('AC-6 / AC-7: transition prop literals prove staggered + fade behavior (source-verified)', () => {
    it('AC-6: NotificationFeed stagger expression produces observable cadence (i * 50)', () => {
      const src = read(NOTIFICATION_FEED);
      // Proof of observable stagger: literal 50ms per step > 1 animation frame (16.6ms).
      expect(src).toMatch(/i\s*\*\s*50/);
    });

    it('AC-7: NewSessionModal backdrop + card fades use 100ms duration (non-instant)', () => {
      const src = read(NEW_SESSION_MODAL);
      const script = scriptBlock(src);
      // The overlay must have a fade directive, and the script must define a
      // helper (or inline value) with duration 100 (non-instant for default motion).
      expect(src).toMatch(/class=["']overlay["'][\s\S]{0,600}?transition:fade=/);
      expect(script).toMatch(/backdropFadeProps[\s\S]*?duration:\s*100\b|duration:\s*100\s*,\s*easing/);
    });
  });

  describe('AC-8: NotificationFeed each block uses a stable key (not index)', () => {
    it('AC-8 (BDD#7): the repo each block key is not a bare index', () => {
      const src = read(NOTIFICATION_FEED);
      // Grab every each-block with a key and assert none are literally (i) or (index).
      const eachRe = /\{#each\s+[^}]+?\s+as\s+([A-Za-z_$][\w$]*)(?:\s*,\s*([A-Za-z_$][\w$]*))?\s*\(([^)]+)\)\}/g;
      let m: RegExpExecArray | null;
      let sawRepoEach = false;
      while ((m = eachRe.exec(src)) !== null) {
        const [, itemName, , keyExpr] = m;
        // The repo each block binds to "repo".
        if (itemName === 'repo') {
          sawRepoEach = true;
          // key must reference a stable identifier (e.g. repo.name, repo.id),
          // not a bare index variable like "i" or "index".
          expect(keyExpr.trim()).not.toBe('i');
          expect(keyExpr.trim()).not.toBe('index');
          expect(keyExpr).toMatch(/repo\.[A-Za-z_$][\w$]*/);
        }
      }
      expect(sawRepoEach, 'must find a keyed each block over repos').toBe(true);
    });
  });
});
