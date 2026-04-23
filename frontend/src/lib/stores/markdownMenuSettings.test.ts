import { describe, it, expect, beforeEach, vi } from 'vitest';
import { get } from 'svelte/store';

vi.mock('../../../wailsjs/go/main/App.js', () => ({
  SetMarkdownMenuSettings: vi.fn(() => Promise.resolve()),
}));

import * as bindings from '../../../wailsjs/go/main/App.js';
import {
  markdownMenuSettings,
  markdownMenuDirty,
  initMarkdownMenuSettings,
  updateMarkdownMenuItem,
  clearMarkdownMenuDirty,
  type MarkdownMenuSettings,
} from './markdownMenuSettings';

const mocks = bindings as unknown as {
  SetMarkdownMenuSettings: ReturnType<typeof vi.fn>;
};

const DEFAULTS: MarkdownMenuSettings = {
  bold: true,
  italic: true,
  strikethrough: true,
  code: true,
  link: true,
  latex: false,
};

beforeEach(() => {
  vi.clearAllMocks();
  mocks.SetMarkdownMenuSettings.mockImplementation(() => Promise.resolve());
  initMarkdownMenuSettings(null);
  clearMarkdownMenuDirty();
});

describe('AC-1: store defaults match backend', () => {
  it('initial value is the full DEFAULTS object', () => {
    expect(get(markdownMenuSettings)).toEqual(DEFAULTS);
  });

  it('dirty flag starts false', () => {
    expect(get(markdownMenuDirty)).toBe(false);
  });
});

describe('AC-2: init merges partial over defaults', () => {
  it('merges {bold:false, latex:true} keeping other defaults', () => {
    initMarkdownMenuSettings({ bold: false, latex: true });

    expect(get(markdownMenuSettings)).toEqual({
      bold: false,
      italic: true,
      strikethrough: true,
      code: true,
      link: true,
      latex: true,
    });
  });

  it('ignores undefined fields in the partial (falls back to defaults)', () => {
    initMarkdownMenuSettings({ bold: undefined, italic: false });

    expect(get(markdownMenuSettings)).toEqual({
      ...DEFAULTS,
      italic: false,
    });
  });

  it('ignores null-valued fields (treats as missing)', () => {
    initMarkdownMenuSettings({ code: null as unknown as boolean });

    expect(get(markdownMenuSettings).code).toBe(true);
  });
});

describe('AC-3: init(null) resets to defaults', () => {
  it('resets the store to DEFAULTS when called with null', () => {
    markdownMenuSettings.set({
      bold: false,
      italic: false,
      strikethrough: false,
      code: false,
      link: false,
      latex: true,
    });

    initMarkdownMenuSettings(null);

    expect(get(markdownMenuSettings)).toEqual(DEFAULTS);
  });

  it('does not throw on undefined input and resets to DEFAULTS', () => {
    markdownMenuSettings.set({ ...DEFAULTS, bold: false });

    expect(() => initMarkdownMenuSettings(undefined)).not.toThrow();
    expect(get(markdownMenuSettings)).toEqual(DEFAULTS);
  });
});

describe('AC-4: updateMarkdownMenuItem flips dirty and persists', () => {
  it('updates the store, sets dirty, and calls the Wails binding with the full object', async () => {
    await updateMarkdownMenuItem('bold', false);

    expect(get(markdownMenuSettings).bold).toBe(false);
    expect(get(markdownMenuDirty)).toBe(true);
    expect(mocks.SetMarkdownMenuSettings).toHaveBeenCalledTimes(1);
    expect(mocks.SetMarkdownMenuSettings).toHaveBeenCalledWith({
      ...DEFAULTS,
      bold: false,
    });
  });

  it('persists each distinct update — link=false flows through', async () => {
    await updateMarkdownMenuItem('link', false);

    expect(mocks.SetMarkdownMenuSettings).toHaveBeenCalledWith({
      ...DEFAULTS,
      link: false,
    });
  });

  it('logs and does NOT revert the store when the Wails call rejects', async () => {
    mocks.SetMarkdownMenuSettings.mockRejectedValueOnce(new Error('disk full'));
    const errSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

    await expect(updateMarkdownMenuItem('latex', true)).resolves.toBeUndefined();

    expect(get(markdownMenuSettings).latex).toBe(true);
    expect(get(markdownMenuDirty)).toBe(true);
    expect(errSpy).toHaveBeenCalledTimes(1);

    errSpy.mockRestore();
  });
});

describe('AC-5: clearMarkdownMenuDirty resets the flag only', () => {
  it('sets dirty=false without touching the store or invoking Wails', () => {
    initMarkdownMenuSettings({ latex: true });
    markdownMenuDirty.set(true);
    const before = get(markdownMenuSettings);

    clearMarkdownMenuDirty();

    expect(get(markdownMenuDirty)).toBe(false);
    expect(get(markdownMenuSettings)).toEqual(before);
    expect(get(markdownMenuSettings).latex).toBe(true);
    expect(mocks.SetMarkdownMenuSettings).not.toHaveBeenCalled();
  });
});

describe('AC-6: dirty flag flips synchronously before await', () => {
  it('markdownMenuDirty reads true immediately after invocation, before the mock resolves', async () => {
    // Deferred promise — Wails call hangs until we resolve it manually.
    let resolveWails!: () => void;
    const deferred = new Promise<void>(res => {
      resolveWails = res;
    });
    mocks.SetMarkdownMenuSettings.mockReturnValueOnce(deferred);

    // Invoke WITHOUT awaiting.
    const pending = updateMarkdownMenuItem('italic', false);

    // Synchronously (before any microtask drains) the store and dirty flag
    // must already reflect the change. This is the guarantee
    // MarkdownEditor.svelte depends on.
    expect(get(markdownMenuDirty)).toBe(true);
    expect(get(markdownMenuSettings).italic).toBe(false);
    expect(mocks.SetMarkdownMenuSettings).toHaveBeenCalledTimes(1);

    // Release the deferred call so the promise can settle.
    resolveWails();
    await pending;

    expect(get(markdownMenuDirty)).toBe(true);
  });
});
