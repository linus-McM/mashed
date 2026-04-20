import './style.css'

// DEV-only shim: when the page loads under Vite directly (http://localhost:5173)
// instead of inside the Wails native webview, `window.runtime` and `window.go`
// are undefined and every EventsOn / Go binding call throws at module init —
// App.svelte never mounts. Stub a minimal JS-side event bus + no-op Go
// bindings so the UI boots and exposes its test seams
// (window.__mashedEmitBmadEvent, etc.) for Playwright / Claude-in-Chrome.
if (import.meta.env.DEV && typeof window !== 'undefined' && !window.runtime) {
  const listeners = new Map()
  window.runtime = {
    EventsOnMultiple(name, cb, max) {
      const entry = { cb, max, fired: 0 }
      const arr = listeners.get(name) || []
      arr.push(entry)
      listeners.set(name, arr)
      return () => {
        const cur = listeners.get(name) || []
        listeners.set(name, cur.filter((e) => e !== entry))
      }
    },
    EventsOn(name, cb) { return window.runtime.EventsOnMultiple(name, cb, -1) },
    EventsOnce(name, cb) { return window.runtime.EventsOnMultiple(name, cb, 1) },
    EventsEmit(name, ...data) {
      const arr = [...(listeners.get(name) || [])]
      for (const entry of arr) {
        try { entry.cb(...data) } catch (err) { console.warn('[dev-stub] listener threw', err) }
        entry.fired++
        if (entry.max > 0 && entry.fired >= entry.max) {
          listeners.set(name, (listeners.get(name) || []).filter((e) => e !== entry))
        }
      }
    },
    EventsOff(name) { listeners.delete(name) },
    LogInfo() {}, LogDebug() {}, LogPrint() {}, LogError() {}, LogWarning() {}, LogFatal() {}, LogTrace() {}, LogSetLogLevel() {},
    BrowserOpenURL(u) { window.open(u, '_blank') },
    WindowReload() { location.reload() },
    WindowSetTitle(t) { document.title = t },
    WindowShow() {}, WindowHide() {}, WindowCenter() {},
    WindowMinimise() {}, WindowUnminimise() {}, WindowMaximise() {}, WindowUnmaximise() {}, WindowToggleMaximise() {},
    WindowFullscreen() {}, WindowUnfullscreen() {},
    WindowGetSize: () => ({ w: window.innerWidth, h: window.innerHeight }),
    WindowSetSize() {}, WindowSetMinSize() {}, WindowSetMaxSize() {},
    WindowGetPosition: () => ({ x: 0, y: 0 }),
    WindowSetPosition() {},
    WindowSetBackgroundColour() {},
    Quit() {},
    Environment: () => Promise.resolve({ buildType: 'dev', platform: 'web', arch: 'web' }),
    Hide() {}, Show() {},
    ClipboardSetText(t) { return navigator.clipboard?.writeText(t) || Promise.resolve() },
    ClipboardGetText: () => navigator.clipboard?.readText?.() || Promise.resolve(''),
    ScreenGetAll: () => Promise.resolve([]),
  }
  const bindingStub = new Proxy({}, {
    get: (_t, k) => (k === 'then' ? undefined : () => Promise.resolve(null)),
  })
  window.go = new Proxy({}, { get: () => new Proxy({}, { get: () => bindingStub }) })
  console.info('[mashed dev-stub] running outside Wails webview — runtime + go bindings are stubs')
}

import App from './App.svelte'

const app = new App({
  target: document.getElementById('app')
})

export default app
