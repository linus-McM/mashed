import { writable, get } from 'svelte/store';
import {
  GetConfig,
  SetUIAdapterEnabled,
  SetUIAdapterTimeoutMs,
  SetOllamaModel,
  SetOllamaEnabled,
  SetUIAdapterUntrustedExpanded,
  ListOllamaModels,
  ProbeOllamaReachable,
  SetBackend,
  SetClaudeModel,
  SetCLIModel,
  SetRouterPolicy,
  ListBackendsAvailable,
  ListClaudeModels,
  ListRouterPolicies,
} from '../../../wailsjs/go/main/App.js';

const DEFAULT_TIMEOUT_MS = 3000;
const DEFAULT_MODEL = 'gemma3:4b';

// Mirrors validOllamaModelName in app.go — reject at the UI boundary before
// hitting Wails so bad keystrokes never cross the bridge (spec §4.4 / AC-9).
const MODEL_NAME_RE = /^[a-zA-Z0-9._:-]{1,64}$/;

export const uiAdapterEnabled = writable<boolean>(true);
export const uiAdapterTimeoutMs = writable<number>(DEFAULT_TIMEOUT_MS);
export const ollamaModel = writable<string>(DEFAULT_MODEL);
export const ollamaEnabled = writable<boolean>(true);
export const uiAdapterUntrustedExpanded = writable<boolean>(false);
export const ollamaReachable = writable<boolean | null>(null);
export const ollamaModels = writable<string[]>([]);

// Plan v3 Story 18 — backend / router selectors.
export const backend = writable<string>('ollama');
export const claudeModel = writable<string>('claude-haiku-4-5');
export const cliModel = writable<string>('claude-haiku-4-5');
export const routerPolicy = writable<string>('local-only');
export const backendsAvailable = writable<string[]>([]);
export const claudeModelList = writable<string[]>([]);
export const routerPolicyList = writable<string[]>([]);
export const claudeApiReachable = writable<boolean | null>(null);
export const claudeCliReachable = writable<boolean | null>(null);

const BACKENDS_ENUM = ['ollama', 'claude-api', 'claude-cli'] as const;
const ROUTER_POLICIES_ENUM = [
  'local-only',
  'claude-only',
  'claude-first',
  'ollama-first',
  'cost-aware',
  'privacy-strict',
] as const;
const CLAUDE_MODEL_ENUM = ['claude-haiku-4-5', 'claude-sonnet-4-6', 'claude-opus-4-6'] as const;

export function validOllamaModelName(s: unknown): boolean {
  return typeof s === 'string' && MODEL_NAME_RE.test(s);
}
export function validBackend(s: unknown): boolean {
  return typeof s === 'string' && (BACKENDS_ENUM as readonly string[]).includes(s);
}
export function validRouterPolicy(s: unknown): boolean {
  return typeof s === 'string' && (ROUTER_POLICIES_ENUM as readonly string[]).includes(s);
}
export function validClaudeModel(s: unknown): boolean {
  return typeof s === 'string' && (CLAUDE_MODEL_ENUM as readonly string[]).includes(s);
}

async function refreshProbeAndModels(callList: boolean): Promise<void> {
  const [probe, models] = await Promise.allSettled([
    ProbeOllamaReachable(),
    callList ? ListOllamaModels() : Promise.resolve<string[]>([]),
  ]);
  ollamaReachable.set(probe.status === 'fulfilled' ? probe.value : false);
  ollamaModels.set(
    models.status === 'fulfilled' && Array.isArray(models.value) ? models.value : [],
  );
}

async function refreshBackendOptions(): Promise<void> {
  const [b, c, p] = await Promise.allSettled([
    ListBackendsAvailable(),
    ListClaudeModels(),
    ListRouterPolicies(),
  ]);
  if (b.status === 'fulfilled' && Array.isArray(b.value)) backendsAvailable.set(b.value);
  if (c.status === 'fulfilled' && Array.isArray(c.value)) claudeModelList.set(c.value);
  if (p.status === 'fulfilled' && Array.isArray(p.value)) routerPolicyList.set(p.value);
}

export async function hydrate(): Promise<void> {
  try {
    const cfg: any = await GetConfig();
    if (cfg) {
      uiAdapterEnabled.set(!!cfg.uiAdapterEnabled);
      uiAdapterTimeoutMs.set(cfg.uiAdapterTimeoutMs || DEFAULT_TIMEOUT_MS);
      ollamaModel.set(cfg.ollamaModel || DEFAULT_MODEL);
      ollamaEnabled.set(!!cfg.ollamaEnabled);
      uiAdapterUntrustedExpanded.set(!!cfg.uiAdapterUntrustedExpanded);
      // v3 Story 18 — backend / router fields.
      if (cfg.backend && validBackend(cfg.backend)) backend.set(cfg.backend);
      if (cfg.claudeModel && validClaudeModel(cfg.claudeModel)) claudeModel.set(cfg.claudeModel);
      if (cfg.cliModel && validClaudeModel(cfg.cliModel)) cliModel.set(cfg.cliModel);
      if (cfg.routerPolicy && validRouterPolicy(cfg.routerPolicy)) routerPolicy.set(cfg.routerPolicy);
    }
  } catch {
    // keep current defaults — hydrate must never throw at the view boundary
  }
  await refreshProbeAndModels(get(uiAdapterEnabled));
  await refreshBackendOptions();
}

export async function refreshModels(): Promise<void> {
  await refreshProbeAndModels(true);
}

export async function setEnabled(v: boolean): Promise<boolean> {
  try {
    await SetUIAdapterEnabled(v);
    uiAdapterEnabled.set(v);
    return true;
  } catch {
    return false;
  }
}

export async function setTimeoutMs(v: number): Promise<boolean> {
  try {
    await SetUIAdapterTimeoutMs(v);
    uiAdapterTimeoutMs.set(v);
    return true;
  } catch {
    return false;
  }
}

export async function setModel(v: string): Promise<boolean> {
  if (!validOllamaModelName(v)) return false;
  try {
    await SetOllamaModel(v);
    ollamaModel.set(v);
    return true;
  } catch {
    return false;
  }
}

export async function setOllamaEnabled(v: boolean): Promise<boolean> {
  try {
    await SetOllamaEnabled(v);
    ollamaEnabled.set(v);
    return true;
  } catch {
    return false;
  }
}

// Plan v3 Story 18 — optimistic-write setters with rollback on Wails
// rejection (AC-18.5). Client-side enum validation first; on Go-side
// rejection, the writable reverts to its previous value.
export async function setBackend(v: string): Promise<boolean> {
  if (!validBackend(v)) return false;
  const prev = get(backend);
  backend.set(v);
  try {
    await SetBackend(v);
    return true;
  } catch {
    backend.set(prev);
    return false;
  }
}

export async function setClaudeModel(v: string): Promise<boolean> {
  if (!validClaudeModel(v)) return false;
  const prev = get(claudeModel);
  claudeModel.set(v);
  try {
    await SetClaudeModel(v);
    return true;
  } catch {
    claudeModel.set(prev);
    return false;
  }
}

export async function setCliModel(v: string): Promise<boolean> {
  if (!validClaudeModel(v)) return false;
  const prev = get(cliModel);
  cliModel.set(v);
  try {
    await SetCLIModel(v);
    return true;
  } catch {
    cliModel.set(prev);
    return false;
  }
}

export async function setRouterPolicy(v: string): Promise<boolean> {
  if (!validRouterPolicy(v)) return false;
  const prev = get(routerPolicy);
  routerPolicy.set(v);
  try {
    await SetRouterPolicy(v);
    return true;
  } catch {
    routerPolicy.set(prev);
    return false;
  }
}

export async function setUntrustedExpanded(v: boolean): Promise<boolean> {
  try {
    await SetUIAdapterUntrustedExpanded(v);
    uiAdapterUntrustedExpanded.set(v);
    return true;
  } catch {
    return false;
  }
}

// Dev-only test seam — Playwright seeds the Q7 flag to exercise the modal's
// "expand on flagged" branch without going through the backend config cycle.
// `import.meta.env.DEV` is a Vite static literal so this block is dead code
// in production bundles.
if (import.meta.env.DEV && typeof window !== 'undefined') {
  (
    window as unknown as {
      __mashed_setUntrustedExpandedForTests?: (v: boolean) => void;
    }
  ).__mashed_setUntrustedExpandedForTests = (v) => uiAdapterUntrustedExpanded.set(v);
}
