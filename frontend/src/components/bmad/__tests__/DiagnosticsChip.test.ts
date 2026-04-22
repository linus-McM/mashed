/**
 * @vitest-environment jsdom
 */
// Tests for DiagnosticsChip.svelte — story ui-ast-U8. Covers AC-5, AC-6, AC-7,
// plus design-brief §5 overflow cap and the null-diagnostics silent case.
import { describe, it, expect } from 'vitest';

import DiagnosticsChip from '../DiagnosticsChip.svelte';
import type { Diagnostics } from '../../../types/uiAst';
import { makeMount } from './mountSvelte';

const mount = makeMount();
const render = (diagnostics: Diagnostics | null) => mount(DiagnosticsChip, { diagnostics });

const findUntrusted = (el: HTMLElement) =>
  Array.from(el.querySelectorAll<HTMLElement>('.chip')).find(
    (c) => c.textContent?.trim() === 'Untrusted',
  );

const findNotesChip = (el: HTMLElement) =>
  Array.from(el.querySelectorAll<HTMLElement>('.chip')).find((c) =>
    /\badapter notes?\b/.test(c.textContent ?? ''),
  );

describe('DiagnosticsChip', () => {
  describe('AC-5 untrusted-chip-visible', () => {
    it('AC-5 shows "Untrusted" chip with tooltip when diagnostics.untrusted is true', () => {
      const el = render({ untrusted: true });
      const chip = findUntrusted(el);
      expect(chip, 'Untrusted chip must render').toBeDefined();
      const title = chip!.getAttribute('title') ?? '';
      expect(title.length).toBeGreaterThan(0);
      expect(title.toLowerCase()).toContain('raw');
    });
  });

  describe('AC-6 note-count-pluralisation', () => {
    it('AC-6 two reasons → "2 adapter notes" with both reasons in the title', () => {
      const reasons = ['empty_options', 'dup_key'];
      const el = render({ fallback_reasons: reasons });
      const chip = findNotesChip(el);
      expect(chip, 'notes chip must render when fallback_reasons is non-empty').toBeDefined();
      expect(chip!.textContent?.trim()).toBe('2 adapter notes');
      const title = chip!.getAttribute('title') ?? '';
      for (const r of reasons) {
        expect(title).toContain(r);
      }
    });

    it('AC-6 one reason → "1 adapter note" (singular)', () => {
      const el = render({ fallback_reasons: ['only_one'] });
      const chip = findNotesChip(el);
      expect(chip, 'notes chip must render for one reason').toBeDefined();
      expect(chip!.textContent?.trim()).toBe('1 adapter note');
      expect(chip!.getAttribute('title') ?? '').toContain('only_one');
    });
  });

  describe('AC-7 empty-diagnostics-silent', () => {
    it('AC-7 renders nothing when untrusted=false and fallback_reasons=[]', () => {
      const el = render({ untrusted: false, fallback_reasons: [] });
      expect(el.querySelector('.chip')).toBeNull();
      expect(el.querySelector('[role="status"]')).toBeNull();
    });

    it('AC-7 renders nothing when diagnostics is null', () => {
      const el = render(null);
      expect(el.querySelector('.chip')).toBeNull();
      expect(el.querySelector('[role="status"]')).toBeNull();
    });
  });

  describe('design-brief §5 overflow cap', () => {
    const mk = (n: number): string[] =>
      Array.from({ length: n }, (_, i) => `reason_${i + 1}`);

    it('boundary 8 — title lists all reasons, no "+N more" suffix', () => {
      const reasons = mk(8);
      const el = render({ fallback_reasons: reasons });
      const chip = findNotesChip(el)!;
      expect(chip.textContent?.trim()).toBe('8 adapter notes');
      const title = chip.getAttribute('title') ?? '';
      for (const r of reasons) expect(title).toContain(r);
      expect(title).not.toMatch(/\+\s*\d+\s*more/i);
    });

    it('boundary 9 — chip shows total count, title caps reasons and appends "+1 more"', () => {
      const reasons = mk(9);
      const el = render({ fallback_reasons: reasons });
      const chip = findNotesChip(el)!;
      expect(chip.textContent?.trim()).toBe('9 adapter notes');
      const title = chip.getAttribute('title') ?? '';
      // First 8 reasons visible in the tooltip.
      for (const r of reasons.slice(0, 8)) expect(title).toContain(r);
      // The 9th reason is NOT listed verbatim — it's folded into "+1 more".
      expect(title).not.toContain('reason_9');
      expect(title).toMatch(/\+\s*1\s*more/i);
    });
  });
});
