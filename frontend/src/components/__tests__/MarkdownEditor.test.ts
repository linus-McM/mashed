import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { createDebouncedSave, type SaveStatus } from '../markdownEditorUtils';

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
