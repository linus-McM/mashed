import './style.css';

import App from './App.svelte';

// ---------------------------------------------------------------------------
// DEV-only Wails runtime shim
//
// When the page loads under Vite directly (http://localhost:5173) instead of
// inside the Wails native webview, `window.runtime` and `window.go` are
// undefined and every EventsOn / Go binding call throws at module init —
// App.svelte never mounts. Stub a minimal JS-side event bus + no-op Go
// bindings so the UI boots and exposes its test seams
// (window.__mashedEmitBmadEvent, etc.) for Playwright / Claude-in-Chrome.
// ---------------------------------------------------------------------------

interface ListenerEntry {
  cb: (...data: unknown[]) => void;
  max: number;
  fired: number;
}

type EventCallback = (...data: unknown[]) => void;

/**
 * Minimum Wails runtime surface we stub in dev mode. Keep in sync with what
 * the rest of the app actually calls — not the full `window.runtime` API.
 */
interface DevRuntime {
  EventsOnMultiple(name: string, cb: EventCallback, max: number): () => void;
  EventsOn(name: string, cb: EventCallback): () => void;
  EventsOnce(name: string, cb: EventCallback): () => void;
  EventsEmit(name: string, ...data: unknown[]): void;
  EventsOff(name: string): void;
  LogInfo(): void;
  LogDebug(): void;
  LogPrint(): void;
  LogError(): void;
  LogWarning(): void;
  LogFatal(): void;
  LogTrace(): void;
  LogSetLogLevel(): void;
  BrowserOpenURL(u: string): void;
  WindowReload(): void;
  WindowSetTitle(t: string): void;
  WindowShow(): void;
  WindowHide(): void;
  WindowCenter(): void;
  WindowMinimise(): void;
  WindowUnminimise(): void;
  WindowMaximise(): void;
  WindowUnmaximise(): void;
  WindowToggleMaximise(): void;
  WindowFullscreen(): void;
  WindowUnfullscreen(): void;
  WindowGetSize(): { w: number; h: number };
  WindowSetSize(): void;
  WindowSetMinSize(): void;
  WindowSetMaxSize(): void;
  WindowGetPosition(): { x: number; y: number };
  WindowSetPosition(): void;
  WindowSetBackgroundColour(): void;
  Quit(): void;
  Environment(): Promise<{ buildType: string; platform: string; arch: string }>;
  Hide(): void;
  Show(): void;
  ClipboardSetText(t: string): Promise<void>;
  ClipboardGetText(): Promise<string>;
  ScreenGetAll(): Promise<unknown[]>;
}

// `window.runtime` / `window.go` are supplied by Wails at runtime in the
// native webview; add minimal typing so this file type-checks without a
// global `.d.ts`. The `runtime` field stays loose (`unknown`) outside the
// dev-branch because the real Wails runtime.d.ts is the source of truth.
declare global {
  interface Window {
    runtime?: DevRuntime | unknown;
    go?: unknown;
  }
}

if (import.meta.env.DEV && typeof window !== 'undefined' && !window.runtime) {
  const listeners = new Map<string, ListenerEntry[]>();

  const stub: DevRuntime = {
    EventsOnMultiple(name: string, cb: EventCallback, max: number): () => void {
      const entry: ListenerEntry = { cb, max, fired: 0 };
      const arr = listeners.get(name) ?? [];
      arr.push(entry);
      listeners.set(name, arr);
      return () => {
        const cur = listeners.get(name) ?? [];
        listeners.set(
          name,
          cur.filter((e) => e !== entry),
        );
      };
    },
    EventsOn(name: string, cb: EventCallback): () => void {
      return stub.EventsOnMultiple(name, cb, -1);
    },
    EventsOnce(name: string, cb: EventCallback): () => void {
      return stub.EventsOnMultiple(name, cb, 1);
    },
    EventsEmit(name: string, ...data: unknown[]): void {
      const arr = [...(listeners.get(name) ?? [])];
      for (const entry of arr) {
        try {
          entry.cb(...data);
        } catch (err) {
          console.warn('[dev-stub] listener threw', err);
        }
        entry.fired++;
        if (entry.max > 0 && entry.fired >= entry.max) {
          listeners.set(
            name,
            (listeners.get(name) ?? []).filter((e) => e !== entry),
          );
        }
      }
    },
    EventsOff(name: string): void {
      listeners.delete(name);
    },
    LogInfo() {},
    LogDebug() {},
    LogPrint() {},
    LogError() {},
    LogWarning() {},
    LogFatal() {},
    LogTrace() {},
    LogSetLogLevel() {},
    BrowserOpenURL(u: string): void {
      window.open(u, '_blank');
    },
    WindowReload(): void {
      location.reload();
    },
    WindowSetTitle(t: string): void {
      document.title = t;
    },
    WindowShow() {},
    WindowHide() {},
    WindowCenter() {},
    WindowMinimise() {},
    WindowUnminimise() {},
    WindowMaximise() {},
    WindowUnmaximise() {},
    WindowToggleMaximise() {},
    WindowFullscreen() {},
    WindowUnfullscreen() {},
    WindowGetSize: () => ({ w: window.innerWidth, h: window.innerHeight }),
    WindowSetSize() {},
    WindowSetMinSize() {},
    WindowSetMaxSize() {},
    WindowGetPosition: () => ({ x: 0, y: 0 }),
    WindowSetPosition() {},
    WindowSetBackgroundColour() {},
    Quit() {},
    Environment: () =>
      Promise.resolve({ buildType: 'dev', platform: 'web', arch: 'web' }),
    Hide() {},
    Show() {},
    ClipboardSetText: (t: string): Promise<void> =>
      navigator.clipboard?.writeText(t) ?? Promise.resolve(),
    ClipboardGetText: (): Promise<string> =>
      navigator.clipboard?.readText?.() ?? Promise.resolve(''),
    ScreenGetAll: (): Promise<unknown[]> => Promise.resolve([]),
  };

  window.runtime = stub;

  // `window.go` is a deeply-nested namespaced binding proxy in real Wails —
  // `window.go.main.App.MethodName(...)`. Stub every path with a Promise<null>.
  const bindingStub = new Proxy(
    {},
    {
      get: (_t, k) => (k === 'then' ? undefined : () => Promise.resolve(null)),
    },
  );
  window.go = new Proxy(
    {},
    {
      get: () =>
        new Proxy(
          {},
          {
            get: () => bindingStub,
          },
        ),
    },
  );
  console.info(
    '[mashed dev-stub] running outside Wails webview — runtime + go bindings are stubs',
  );
}

// ---------------------------------------------------------------------------
// App mount
// ---------------------------------------------------------------------------

const target = document.getElementById('app');
if (!target) {
  throw new Error('main.ts: #app mount point missing from index.html');
}

const app = new App({ target });

export default app;
