import { describe, it, expect, vi, beforeEach } from 'vitest';
import { get } from 'svelte/store';

// Stub the Wails bindings so the test runs without a real Go backend.
const mockSet = {
  backend: vi.fn<(v: string) => Promise<void>>(),
  claudeModel: vi.fn<(v: string) => Promise<void>>(),
  cliModel: vi.fn<(v: string) => Promise<void>>(),
  routerPolicy: vi.fn<(v: string) => Promise<void>>(),
};
vi.mock('../../../wailsjs/go/main/App.js', () => ({
  GetConfig: vi.fn().mockResolvedValue({}),
  SetUIAdapterEnabled: vi.fn(),
  SetUIAdapterTimeoutMs: vi.fn(),
  SetOllamaModel: vi.fn(),
  SetOllamaEnabled: vi.fn(),
  SetUIAdapterUntrustedExpanded: vi.fn(),
  ListOllamaModels: vi.fn().mockResolvedValue([]),
  ProbeOllamaReachable: vi.fn().mockResolvedValue(true),
  SetBackend: (v: string) => mockSet.backend(v),
  SetClaudeModel: (v: string) => mockSet.claudeModel(v),
  SetCLIModel: (v: string) => mockSet.cliModel(v),
  SetRouterPolicy: (v: string) => mockSet.routerPolicy(v),
  ListBackendsAvailable: vi.fn().mockResolvedValue(['ollama', 'claude-api', 'claude-cli']),
  ListClaudeModels: vi.fn().mockResolvedValue(['claude-haiku-4-5', 'claude-sonnet-4-6']),
  ListRouterPolicies: vi.fn().mockResolvedValue([
    'local-only', 'claude-only', 'claude-first', 'ollama-first', 'cost-aware', 'privacy-strict',
  ]),
}));

import {
  backend,
  claudeModel,
  routerPolicy,
  validBackend,
  validClaudeModel,
  validRouterPolicy,
  setBackend,
  setClaudeModel,
  setRouterPolicy,
} from './uiAdapterSettings';

describe('uiAdapterSettings v3-18 selectors', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockSet.backend.mockResolvedValue(undefined);
    mockSet.claudeModel.mockResolvedValue(undefined);
    mockSet.cliModel.mockResolvedValue(undefined);
    mockSet.routerPolicy.mockResolvedValue(undefined);
    backend.set('ollama');
    claudeModel.set('claude-haiku-4-5');
    routerPolicy.set('local-only');
  });

  it('validBackend accepts enum and rejects others', () => {
    expect(validBackend('ollama')).toBe(true);
    expect(validBackend('claude-api')).toBe(true);
    expect(validBackend('claude-cli')).toBe(true);
    expect(validBackend('not-a-backend')).toBe(false);
    expect(validBackend(null)).toBe(false);
  });

  it('validClaudeModel accepts Story v3-12 allowlist', () => {
    expect(validClaudeModel('claude-haiku-4-5')).toBe(true);
    expect(validClaudeModel('claude-opus-4-6')).toBe(true);
    expect(validClaudeModel('claude-foo-9')).toBe(false);
  });

  it('validRouterPolicy covers all six plan values', () => {
    for (const p of ['local-only', 'claude-only', 'claude-first', 'ollama-first', 'cost-aware', 'privacy-strict']) {
      expect(validRouterPolicy(p)).toBe(true);
    }
    expect(validRouterPolicy('unknown-policy')).toBe(false);
  });

  it('setBackend optimistic-writes and rolls back on Wails rejection (AC-18.5)', async () => {
    expect(get(backend)).toBe('ollama');
    mockSet.backend.mockRejectedValueOnce(new Error('invalid backend'));
    const ok = await setBackend('claude-api');
    expect(ok).toBe(false);
    // Rollback — store reverts to previous value.
    expect(get(backend)).toBe('ollama');
  });

  it('setBackend happy path updates the store + calls Wails once', async () => {
    mockSet.backend.mockResolvedValueOnce(undefined);
    const ok = await setBackend('claude-api');
    expect(ok).toBe(true);
    expect(get(backend)).toBe('claude-api');
    expect(mockSet.backend).toHaveBeenCalledTimes(1);
    expect(mockSet.backend).toHaveBeenCalledWith('claude-api');
  });

  it('setBackend rejects invalid values before hitting Wails (AC-18.3)', async () => {
    const ok = await setBackend('injected-via-devtools');
    expect(ok).toBe(false);
    expect(mockSet.backend).not.toHaveBeenCalled();
    expect(get(backend)).toBe('ollama');
  });

  it('setClaudeModel rollback + validator gate', async () => {
    mockSet.claudeModel.mockRejectedValueOnce(new Error('reject'));
    claudeModel.set('claude-haiku-4-5');
    const ok = await setClaudeModel('claude-sonnet-4-6');
    expect(ok).toBe(false);
    expect(get(claudeModel)).toBe('claude-haiku-4-5');

    const bad = await setClaudeModel('claude-foo-9');
    expect(bad).toBe(false);
    expect(mockSet.claudeModel).toHaveBeenCalledTimes(1); // validator gate, 1 Wails call total
  });

  it('setRouterPolicy dispatches a single Wails call for valid enum', async () => {
    const ok = await setRouterPolicy('cost-aware');
    expect(ok).toBe(true);
    expect(get(routerPolicy)).toBe('cost-aware');
    expect(mockSet.routerPolicy).toHaveBeenCalledTimes(1);
    expect(mockSet.routerPolicy).toHaveBeenCalledWith('cost-aware');
  });
});
