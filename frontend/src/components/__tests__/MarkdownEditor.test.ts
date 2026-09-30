import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import {
  createDebouncedSave,
  applyToolbarAttributes,
  computeToolbarApplyTarget,
  TOOLBAR_KEYS,
  type SaveStatus,
} from '../markdownEditorUtils';
import type { MarkdownMenuSettings } from '../../lib/stores/markdownMenuSettings';

const ALL_ON: MarkdownMenuSettings = {
  bold: true,
  italic: true,
  strikethrough: true,
  code: true,
  link: true,
  latex: true,
};

const ALL_OFF: MarkdownMenuSettings = {
  bold: false,
  italic: false,
  strikethrough: false,
  code: false,
  link: false,
  latex: false,
};

describe('MarkdownEditor utilities', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });
  afterEach(() => {
    vi.useRealTimers();
  });

  describe('AC-3: Debounced auto-save (800ms)', () => {
    it('calls saveFn after delay elapses', () => {
      const saveFn = vi.fn();
      const saver = createDebouncedSave(saveFn, 800);

      saver.schedule();
      expect(saveFn).not.toHaveBeenCalled();

      vi.advanceTimersByTime(800);
      expect(saveFn).toHaveBeenCalledOnce();
    });

    it('does not call saveFn before delay elapses', () => {
      const saveFn = vi.fn();
      const saver = createDebouncedSave(saveFn, 800);

      saver.schedule();
      vi.advanceTimersByTime(799);
      expect(saveFn).not.toHaveBeenCalled();
    });

    it('resets timer on repeated schedule() calls — only one save fires', () => {
      const saveFn = vi.fn();
      const saver = createDebouncedSave(saveFn, 800);

      saver.schedule();
      vi.advanceTimersByTime(500);
      saver.schedule(); // reset
      vi.advanceTimersByTime(500);
      expect(saveFn).not.toHaveBeenCalled(); // only 500ms since last schedule

      vi.advanceTimersByTime(300);
      expect(saveFn).toHaveBeenCalledOnce();
    });

    it('does not fire again without another schedule() call', () => {
      const saveFn = vi.fn();
      const saver = createDebouncedSave(saveFn, 800);

      saver.schedule();
      vi.advanceTimersByTime(800);
      expect(saveFn).toHaveBeenCalledOnce();

      // No second call without another schedule()
      vi.advanceTimersByTime(2000);
      expect(saveFn).toHaveBeenCalledOnce();
    });
  });

  describe('AC-4: Cmd+S bypass via flush()', () => {
    it('flush() calls saveFn immediately', () => {
      const saveFn = vi.fn();
      const saver = createDebouncedSave(saveFn, 800);

      saver.flush();
      expect(saveFn).toHaveBeenCalledOnce();
    });

    it('flush() cancels any pending scheduled save', () => {
      const saveFn = vi.fn();
      const saver = createDebouncedSave(saveFn, 800);

      saver.schedule();
      vi.advanceTimersByTime(400);
      saver.flush();
      expect(saveFn).toHaveBeenCalledOnce();

      // Pending timer should be cancelled — no second call
      vi.advanceTimersByTime(800);
      expect(saveFn).toHaveBeenCalledOnce();
    });

    it('flush() works even with no pending save', () => {
      const saveFn = vi.fn();
      const saver = createDebouncedSave(saveFn, 800);

      saver.flush();
      expect(saveFn).toHaveBeenCalledOnce();
    });
  });

  describe('AC-7: File switching cleanup', () => {
    it('cancel() prevents pending save from firing', () => {
      const saveFn = vi.fn();
      const saver = createDebouncedSave(saveFn, 800);

      saver.schedule();
      vi.advanceTimersByTime(400);
      saver.cancel();

      vi.advanceTimersByTime(800);
      expect(saveFn).not.toHaveBeenCalled();
    });

    it('cancel() is a no-op when nothing is pending', () => {
      const saveFn = vi.fn();
      const saver = createDebouncedSave(saveFn, 800);

      saver.cancel();
      expect(saveFn).not.toHaveBeenCalled();
    });

    it('schedule() works normally after cancel()', () => {
      const saveFn = vi.fn();
      const saver = createDebouncedSave(saveFn, 800);

      saver.schedule();
      saver.cancel();
      saver.schedule();

      vi.advanceTimersByTime(800);
      expect(saveFn).toHaveBeenCalledOnce();
    });

    it('destroy() prevents all future saves', () => {
      const saveFn = vi.fn();
      const saver = createDebouncedSave(saveFn, 800);

      saver.schedule();
      saver.destroy();

      vi.advanceTimersByTime(800);
      expect(saveFn).not.toHaveBeenCalled();
    });

    it('schedule() is a no-op after destroy()', () => {
      const saveFn = vi.fn();
      const saver = createDebouncedSave(saveFn, 800);

      saver.destroy();
      saver.schedule();

      vi.advanceTimersByTime(2000);
      expect(saveFn).not.toHaveBeenCalled();
    });

    it('flush() is a no-op after destroy()', () => {
      const saveFn = vi.fn();
      const saver = createDebouncedSave(saveFn, 800);

      saver.destroy();
      saver.flush();

      expect(saveFn).not.toHaveBeenCalled();
    });
  });

  describe('SaveStatus type', () => {
    it.each<SaveStatus>(['idle', 'modified', 'saving', 'saved', 'error'])(
      'accepts "%s" as a valid SaveStatus',
      (status) => {
        const s: SaveStatus = status;
        expect(s).toBe(status);
      },
    );
  });
});

describe('Story 06: MarkdownEditor toolbar wiring — fallback (CSS masking)', () => {
  // applyToolbarAttributes tests use real DOM via jsdom, so real timers are
  // fine. Keep them decoupled from the fake-timer block above.

  describe('applyToolbarAttributes — pure DOM side-effect', () => {
    it('is a safe no-op when container is null (AC-1 precondition)', () => {
      // Should not throw
      expect(() => applyToolbarAttributes(null, ALL_ON)).not.toThrow();
    });

    it('AC-1: applies all six data-toolbar-* attributes to .milkdown when present', () => {
      const container = document.createElement('div');
      const milkdown = document.createElement('div');
      milkdown.classList.add('milkdown');
      container.appendChild(milkdown);

      applyToolbarAttributes(container, ALL_ON);

      expect(milkdown.getAttribute('data-toolbar-bold')).toBe('on');
      expect(milkdown.getAttribute('data-toolbar-italic')).toBe('on');
      expect(milkdown.getAttribute('data-toolbar-strikethrough')).toBe('on');
      expect(milkdown.getAttribute('data-toolbar-code')).toBe('on');
      expect(milkdown.getAttribute('data-toolbar-link')).toBe('on');
      expect(milkdown.getAttribute('data-toolbar-latex')).toBe('on');
    });

    it('falls back to the container itself when no .milkdown child exists', () => {
      const container = document.createElement('div');
      applyToolbarAttributes(container, ALL_OFF);
      expect(container.getAttribute('data-toolbar-bold')).toBe('off');
      expect(container.getAttribute('data-toolbar-latex')).toBe('off');
    });

    it('emits "off" for disabled keys and "on" for enabled keys', () => {
      const container = document.createElement('div');
      const milkdown = document.createElement('div');
      milkdown.classList.add('milkdown');
      container.appendChild(milkdown);

      const mixed: MarkdownMenuSettings = {
        bold: true,
        italic: false,
        strikethrough: true,
        code: false,
        link: true,
        latex: false,
      };
      applyToolbarAttributes(container, mixed);

      expect(milkdown.getAttribute('data-toolbar-bold')).toBe('on');
      expect(milkdown.getAttribute('data-toolbar-italic')).toBe('off');
      expect(milkdown.getAttribute('data-toolbar-strikethrough')).toBe('on');
      expect(milkdown.getAttribute('data-toolbar-code')).toBe('off');
      expect(milkdown.getAttribute('data-toolbar-link')).toBe('on');
      expect(milkdown.getAttribute('data-toolbar-latex')).toBe('off');
    });

    it('overwrites previous attribute values when called again', () => {
      const container = document.createElement('div');
      const milkdown = document.createElement('div');
      milkdown.classList.add('milkdown');
      container.appendChild(milkdown);

      applyToolbarAttributes(container, ALL_ON);
      applyToolbarAttributes(container, ALL_OFF);

      for (const key of TOOLBAR_KEYS) {
        expect(milkdown.getAttribute(`data-toolbar-${key}`)).toBe('off');
      }
    });

    it('TOOLBAR_KEYS matches all six keys of MarkdownMenuSettings', () => {
      expect(TOOLBAR_KEYS).toEqual([
        'bold',
        'italic',
        'strikethrough',
        'code',
        'link',
        'latex',
      ]);
      // Exhaustiveness check vs the settings object
      const settingsKeys = Object.keys(ALL_ON).sort();
      const toolbarKeys = [...TOOLBAR_KEYS].sort();
      expect(toolbarKeys).toEqual(settingsKeys);
    });
  });

  describe('computeToolbarApplyTarget — deferred-apply decision', () => {
    const lastApplied = JSON.stringify(ALL_ON);

    it('AC-2: returns null when dirty is true (blocks while user is toggling)', () => {
      const result = computeToolbarApplyTarget({
        dirty: true,
        mounted: true,
        loading: false,
        settings: ALL_OFF, // changed
        lastApplied,
      });
      expect(result).toBeNull();
    });

    it('AC-3: returns new JSON when dirty flips false with changed settings', () => {
      const result = computeToolbarApplyTarget({
        dirty: false,
        mounted: true,
        loading: false,
        settings: ALL_OFF,
        lastApplied,
      });
      expect(result).toBe(JSON.stringify(ALL_OFF));
    });

    it('AC-3 (idempotence): returns null on second evaluation with same settings', () => {
      const first = computeToolbarApplyTarget({
        dirty: false,
        mounted: true,
        loading: false,
        settings: ALL_OFF,
        lastApplied,
      });
      expect(first).not.toBeNull();
      // Simulate MarkdownEditor updating lastApplied to `first` after apply
      const second = computeToolbarApplyTarget({
        dirty: false,
        mounted: true,
        loading: false,
        settings: ALL_OFF,
        lastApplied: first as string,
      });
      expect(second).toBeNull();
    });

    it('AC-4: returns null when dirty flips false but settings are unchanged (reverted)', () => {
      const result = computeToolbarApplyTarget({
        dirty: false,
        mounted: true,
        loading: false,
        settings: ALL_ON, // same as lastApplied
        lastApplied,
      });
      expect(result).toBeNull();
    });

    it('blocks when editor is not mounted (crepe === null)', () => {
      const result = computeToolbarApplyTarget({
        dirty: false,
        mounted: false,
        loading: false,
        settings: ALL_OFF,
        lastApplied,
      });
      expect(result).toBeNull();
    });

    it('AC-6: returns null when loading is true (init in flight)', () => {
      const result = computeToolbarApplyTarget({
        dirty: false,
        mounted: true,
        loading: true,
        settings: ALL_OFF,
        lastApplied,
      });
      expect(result).toBeNull();
    });

    it('guard precedence: dirty check fires before mount check', () => {
      const result = computeToolbarApplyTarget({
        dirty: true,
        mounted: false,
        loading: true,
        settings: ALL_OFF,
        lastApplied,
      });
      expect(result).toBeNull();
    });
  });

  describe('AC-5 (adapted): settings-change path does NOT call saver.flush or destroy', () => {
    // In the fallback strategy the settings-change branch only calls
    // `applyToolbarAttributes`; it does NOT call saver.flush() or
    // destroyEditor(). Those remain reserved for the filePath-change branch.
    // This is a documented deviation from story 06's AC-5 (re-init path),
    // rationalised by the fact that CSS-masking does not require teardown.
    //
    // We assert the structural property by grepping the compiled component
    // source: the settings-change reactive block must NOT contain
    // `saver.flush()` or `destroyEditor()` — those tokens must only appear
    // in the filePath-change branch and onDestroy hook.
    it('MarkdownEditor.svelte settings-change reactive block has no saver.flush / destroyEditor', () => {
      const src = readFileSync(
        resolve(__dirname, '../MarkdownEditor.svelte'),
        'utf-8',
      );
      // Locate the deferred-apply reactive block by its unique marker.
      const startMarker = '// Story 06: deferred-apply';
      const startIdx = src.indexOf(startMarker);
      expect(startIdx).toBeGreaterThan(-1);
      // Block ends at the start of the next reactive/toplevel declaration.
      const endIdx = src.indexOf('// Toggle readonly', startIdx);
      expect(endIdx).toBeGreaterThan(startIdx);
      const block = src.slice(startIdx, endIdx);

      // Regex ignores comments referencing these tokens by checking for a
      // function-call shape (token followed by '(' optionally with whitespace).
      // The block's comment mentions "saver.flush()" so use a negative lookbehind
      // via line filtering: strip comment lines first.
      const code = block
        .split('\n')
        .filter((line) => !line.trim().startsWith('//'))
        .join('\n');
      expect(code).not.toMatch(/saver\.flush\s*\(/);
      expect(code).not.toMatch(/destroyEditor\s*\(/);
      // ...but the applyToolbarAttributes call must be there.
      expect(code).toMatch(/applyToolbarAttributes\s*\(/);
    });

    it('filePath-change branch is untouched and still calls saver.flush + destroyEditor', () => {
      const src = readFileSync(
        resolve(__dirname, '../MarkdownEditor.svelte'),
        'utf-8',
      );
      const fpIdx = src.indexOf('if (filePath !== prevFilePath)');
      expect(fpIdx).toBeGreaterThan(-1);
      const fpBlock = src.slice(fpIdx, fpIdx + 400);
      expect(fpBlock).toMatch(/saver\.flush\s*\(\s*\)/);
      expect(fpBlock).toMatch(/destroyEditor\s*\(\s*\)\.then/);
    });
  });

  describe('CSS contract — crepe-mashed.css hides disabled items', () => {
    const cssPath = resolve(__dirname, '../../styles/crepe-mashed.css');
    const css = readFileSync(cssPath, 'utf-8');

    it.each(TOOLBAR_KEYS)(
      'contains a hide rule for .milkdown[data-toolbar-%s="off"]',
      (key) => {
        const pattern = new RegExp(
          `\\.milkdown\\[data-toolbar-${key}="off"\\][^{]*\\{[^}]*display:\\s*none`,
        );
        expect(css).toMatch(pattern);
      },
    );

    it('declares exactly six data-toolbar-* hide rules (one per key)', () => {
      const matches = css.match(/\.milkdown\[data-toolbar-[a-z]+="off"\]/g);
      expect(matches).not.toBeNull();
      expect(matches).toHaveLength(TOOLBAR_KEYS.length);
    });
  });
});
