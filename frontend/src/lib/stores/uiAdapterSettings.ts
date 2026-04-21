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

export function validOllamaModelName(s: unknown): boolean {
  return typeof s === 'string' && MODEL_NAME_RE.test(s);
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

export async function hydrate(): Promise<void> {
  try {
    const cfg: any = await GetConfig();
    if (cfg) {
      uiAdapterEnabled.set(!!cfg.uiAdapterEnabled);
      uiAdapterTimeoutMs.set(cfg.uiAdapterTimeoutMs || DEFAULT_TIMEOUT_MS);
      ollamaModel.set(cfg.ollamaModel || DEFAULT_MODEL);
      ollamaEnabled.set(!!cfg.ollamaEnabled);
      uiAdapterUntrustedExpanded.set(!!cfg.uiAdapterUntrustedExpanded);
    }
  } catch {
    // keep current defaults — hydrate must never throw at the view boundary
  }
  await refreshProbeAndModels(get(uiAdapterEnabled));
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

export async function setUntrustedExpanded(v: boolean): Promise<boolean> {
  try {
    await SetUIAdapterUntrustedExpanded(v);
    uiAdapterUntrustedExpanded.set(v);
    return true;
  } catch {
    return false;
  }
}
