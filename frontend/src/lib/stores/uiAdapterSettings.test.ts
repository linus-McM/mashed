import { describe, it, expect, beforeEach, vi } from 'vitest';
import { get } from 'svelte/store';

vi.mock('../../../wailsjs/go/main/App.js', () => ({
  GetConfig: vi.fn(),
  SetUIAdapterEnabled: vi.fn(),
  SetUIAdapterTimeoutMs: vi.fn(),
  SetOllamaModel: vi.fn(),
  SetOllamaEnabled: vi.fn(),
  SetUIAdapterUntrustedExpanded: vi.fn(),
  ListOllamaModels: vi.fn(),
  ProbeOllamaReachable: vi.fn(),
}));

import * as bindings from '../../../wailsjs/go/main/App.js';
import {
  uiAdapterEnabled,
  uiAdapterTimeoutMs,
  ollamaModel,
  ollamaEnabled,
  uiAdapterUntrustedExpanded,
  ollamaReachable,
  ollamaModels,
  hydrate,
  refreshModels,
  setEnabled,
  setTimeoutMs,
  setModel,
  setOllamaEnabled,
  setUntrustedExpanded,
  validOllamaModelName,
} from './uiAdapterSettings';

const mocks = bindings as unknown as Record<string, ReturnType<typeof vi.fn>>;

function resetStores() {
  uiAdapterEnabled.set(true);
  uiAdapterTimeoutMs.set(3000);
  ollamaModel.set('gemma3:4b');
  ollamaEnabled.set(true);
  uiAdapterUntrustedExpanded.set(false);
  ollamaReachable.set(null);
  ollamaModels.set([]);
}

beforeEach(() => {
  vi.clearAllMocks();
  resetStores();
});

describe('validOllamaModelName (AC-9 client-side guard)', () => {
  it('accepts canonical Ollama tags', () => {
    expect(validOllamaModelName('gemma3:4b')).toBe(true);
    expect(validOllamaModelName('qwen2.5:3b')).toBe(true);
    expect(validOllamaModelName('llama3.2:3b')).toBe(true);
    expect(validOllamaModelName('mistral')).toBe(true);
    expect(validOllamaModelName('a')).toBe(true);
  });

  it('rejects empty strings', () => {
    expect(validOllamaModelName('')).toBe(false);
  });

  it('rejects path traversal attempts', () => {
    expect(validOllamaModelName('../../etc/passwd')).toBe(false);
    expect(validOllamaModelName('./foo')).toBe(false);
    expect(validOllamaModelName('foo/bar')).toBe(false);
    expect(validOllamaModelName('\\windows\\foo')).toBe(false);
  });

  it('rejects names longer than 64 characters', () => {
    const sixtyFour = 'a'.repeat(64);
    const sixtyFive = 'a'.repeat(65);
    expect(validOllamaModelName(sixtyFour)).toBe(true);
    expect(validOllamaModelName(sixtyFive)).toBe(false);
  });

  it('rejects characters outside [a-zA-Z0-9._:-]', () => {
    expect(validOllamaModelName('foo bar')).toBe(false);
    expect(validOllamaModelName('foo$bar')).toBe(false);
    expect(validOllamaModelName('foo;rm -rf')).toBe(false);
    expect(validOllamaModelName('foo\nbar')).toBe(false);
  });

  it('rejects non-string input defensively', () => {
    // @ts-expect-error — deliberate misuse to assert runtime guard
    expect(validOllamaModelName(undefined)).toBe(false);
    // @ts-expect-error
    expect(validOllamaModelName(null)).toBe(false);
    // @ts-expect-error
    expect(validOllamaModelName(42)).toBe(false);
  });
});

describe('hydrate (AC-5, AC-11)', () => {
  it('populates every store from GetConfig', async () => {
    mocks.GetConfig.mockResolvedValue({
      uiAdapterEnabled: false,
      uiAdapterTimeoutMs: 5000,
      ollamaModel: 'qwen2.5:3b',
      ollamaEnabled: false,
      uiAdapterUntrustedExpanded: true,
    });
    mocks.ProbeOllamaReachable.mockResolvedValue(true);
    mocks.ListOllamaModels.mockResolvedValue(['gemma3:4b']);

    await hydrate();

    expect(get(uiAdapterEnabled)).toBe(false);
    expect(get(uiAdapterTimeoutMs)).toBe(5000);
    expect(get(ollamaModel)).toBe('qwen2.5:3b');
    expect(get(ollamaEnabled)).toBe(false);
    expect(get(uiAdapterUntrustedExpanded)).toBe(true);
    expect(get(ollamaReachable)).toBe(true);
  });

  it('applies defaults when config fields are missing', async () => {
    mocks.GetConfig.mockResolvedValue({});
    mocks.ProbeOllamaReachable.mockResolvedValue(false);
    mocks.ListOllamaModels.mockResolvedValue([]);

    await hydrate();

    expect(get(uiAdapterTimeoutMs)).toBe(3000);
    expect(get(ollamaModel)).toBe('gemma3:4b');
    expect(get(uiAdapterUntrustedExpanded)).toBe(false);
  });

  it('skips ListOllamaModels when adapter is disabled', async () => {
    mocks.GetConfig.mockResolvedValue({ uiAdapterEnabled: false });
    mocks.ProbeOllamaReachable.mockResolvedValue(true);
    mocks.ListOllamaModels.mockResolvedValue(['should-not-see']);

    await hydrate();

    expect(mocks.ListOllamaModels).not.toHaveBeenCalled();
    expect(get(ollamaModels)).toEqual([]);
  });

  it('AC-11: graceful offline — ListOllamaModels error leaves models empty', async () => {
    mocks.GetConfig.mockResolvedValue({ uiAdapterEnabled: true });
    mocks.ProbeOllamaReachable.mockResolvedValue(false);
    mocks.ListOllamaModels.mockRejectedValue(new Error('ErrOllamaUnreachable'));

    await expect(hydrate()).resolves.toBeUndefined();

    expect(get(ollamaReachable)).toBe(false);
    expect(get(ollamaModels)).toEqual([]);
  });

  it('survives a ProbeOllamaReachable rejection without throwing', async () => {
    mocks.GetConfig.mockResolvedValue({ uiAdapterEnabled: true });
    mocks.ProbeOllamaReachable.mockRejectedValue(new Error('boom'));
    mocks.ListOllamaModels.mockResolvedValue(['gemma3:4b']);

    await expect(hydrate()).resolves.toBeUndefined();

    expect(get(ollamaReachable)).toBe(false);
  });

  it('survives a GetConfig rejection without throwing', async () => {
    mocks.GetConfig.mockRejectedValue(new Error('config io'));
    mocks.ProbeOllamaReachable.mockResolvedValue(false);
    mocks.ListOllamaModels.mockResolvedValue([]);

    await expect(hydrate()).resolves.toBeUndefined();
    expect(get(uiAdapterEnabled)).toBe(true);
  });
});

describe('refreshModels (AC-12)', () => {
  it('re-invokes ListOllamaModels and ProbeOllamaReachable', async () => {
    mocks.ProbeOllamaReachable.mockResolvedValue(true);
    mocks.ListOllamaModels.mockResolvedValue(['gemma3:4b', 'qwen2.5:3b']);

    await refreshModels();

    expect(mocks.ProbeOllamaReachable).toHaveBeenCalledTimes(1);
    expect(mocks.ListOllamaModels).toHaveBeenCalledTimes(1);
    expect(get(ollamaReachable)).toBe(true);
    expect(get(ollamaModels)).toEqual(['gemma3:4b', 'qwen2.5:3b']);
  });

  it('falls back to [] on ListOllamaModels error', async () => {
    mocks.ProbeOllamaReachable.mockResolvedValue(false);
    mocks.ListOllamaModels.mockRejectedValue(new Error('unreachable'));

    await expect(refreshModels()).resolves.toBeUndefined();

    expect(get(ollamaReachable)).toBe(false);
    expect(get(ollamaModels)).toEqual([]);
  });
});

describe('setters update stores on success', () => {
  it('setEnabled forwards to binding and mirrors the store', async () => {
    mocks.SetUIAdapterEnabled.mockResolvedValue(undefined);

    const ok = await setEnabled(false);

    expect(mocks.SetUIAdapterEnabled).toHaveBeenCalledWith(false);
    expect(get(uiAdapterEnabled)).toBe(false);
    expect(ok).toBe(true);
  });

  it('setTimeoutMs forwards to binding and mirrors the store', async () => {
    mocks.SetUIAdapterTimeoutMs.mockResolvedValue(undefined);

    const ok = await setTimeoutMs(5000);

    expect(mocks.SetUIAdapterTimeoutMs).toHaveBeenCalledWith(5000);
    expect(get(uiAdapterTimeoutMs)).toBe(5000);
    expect(ok).toBe(true);
  });

  it('setModel forwards valid names to binding and mirrors the store', async () => {
    mocks.SetOllamaModel.mockResolvedValue(undefined);

    const ok = await setModel('qwen2.5:3b');

    expect(mocks.SetOllamaModel).toHaveBeenCalledWith('qwen2.5:3b');
    expect(get(ollamaModel)).toBe('qwen2.5:3b');
    expect(ok).toBe(true);
  });

  it('setOllamaEnabled forwards to binding and mirrors the store', async () => {
    mocks.SetOllamaEnabled.mockResolvedValue(undefined);

    const ok = await setOllamaEnabled(false);

    expect(mocks.SetOllamaEnabled).toHaveBeenCalledWith(false);
    expect(get(ollamaEnabled)).toBe(false);
    expect(ok).toBe(true);
  });

  it('setUntrustedExpanded forwards to binding and mirrors the store', async () => {
    mocks.SetUIAdapterUntrustedExpanded.mockResolvedValue(undefined);

    const ok = await setUntrustedExpanded(true);

    expect(mocks.SetUIAdapterUntrustedExpanded).toHaveBeenCalledWith(true);
    expect(get(uiAdapterUntrustedExpanded)).toBe(true);
    expect(ok).toBe(true);
  });
});

describe('setters swallow binding errors', () => {
  it('setEnabled returns false and leaves the store untouched on error', async () => {
    uiAdapterEnabled.set(true);
    mocks.SetUIAdapterEnabled.mockRejectedValue(new Error('disk full'));

    const ok = await setEnabled(false);

    expect(ok).toBe(false);
    expect(get(uiAdapterEnabled)).toBe(true);
  });

  it('setTimeoutMs returns false and leaves the store untouched on error', async () => {
    uiAdapterTimeoutMs.set(3000);
    mocks.SetUIAdapterTimeoutMs.mockRejectedValue(new Error('out of range'));

    const ok = await setTimeoutMs(250);

    expect(ok).toBe(false);
    expect(get(uiAdapterTimeoutMs)).toBe(3000);
  });

  it('setModel rejects invalid names client-side without calling the binding', async () => {
    ollamaModel.set('gemma3:4b');

    const ok = await setModel('../../etc/passwd');

    expect(ok).toBe(false);
    expect(mocks.SetOllamaModel).not.toHaveBeenCalled();
    expect(get(ollamaModel)).toBe('gemma3:4b');
  });

  it('setModel returns false when the binding rejects a syntactically valid name', async () => {
    ollamaModel.set('gemma3:4b');
    mocks.SetOllamaModel.mockRejectedValue(new Error('backend rejected'));

    const ok = await setModel('qwen2.5:3b');

    expect(ok).toBe(false);
    expect(get(ollamaModel)).toBe('gemma3:4b');
  });

  it('setOllamaEnabled returns false and leaves the store untouched on error', async () => {
    ollamaEnabled.set(true);
    mocks.SetOllamaEnabled.mockRejectedValue(new Error('boom'));

    const ok = await setOllamaEnabled(false);

    expect(ok).toBe(false);
    expect(get(ollamaEnabled)).toBe(true);
  });

  it('setUntrustedExpanded returns false and leaves the store untouched on error', async () => {
    uiAdapterUntrustedExpanded.set(false);
    mocks.SetUIAdapterUntrustedExpanded.mockRejectedValue(new Error('boom'));

    const ok = await setUntrustedExpanded(true);

    expect(ok).toBe(false);
    expect(get(uiAdapterUntrustedExpanded)).toBe(false);
  });
});
