// Shared typed test helpers for BMAD component tests.
//
// - `makeMount()` / `mountComponent()` wrap the Svelte 4 constructor API with a
//   fresh target element and auto-cleanup registered via `afterEach`.
//   `@testing-library/svelte` is not a project dep.
// - `xyflowHandleStub()` returns the `Handle` + `Position` replacement that the
//   `vi.mock('@xyflow/svelte', …)` factories reach for when a node component is
//   mounted outside a `<SvelteFlowProvider />`.
// - `createAppMock()` produces a fully-typed stub for the Wails-generated
//   `wailsjs/go/main/App` module — callers pass per-method overrides that are
//   checked against the generated `App.d.ts` signatures (story svelte-check-06).

import { afterEach, vi } from 'vitest';

import type * as WailsApp from '../../../wailsjs/go/main/App';

// ---------------------------------------------------------------------------
// Svelte 4 mount helpers
// ---------------------------------------------------------------------------

export type MountedComponent = { $destroy(): void; $set?(props: object): void };
type SvelteInit = new (opts: { target: HTMLElement; props: object }) => MountedComponent;

export function makeMount() {
  let instance: MountedComponent | null = null;
  let target: HTMLElement | null = null;

  afterEach(() => {
    instance?.$destroy();
    target?.remove();
    instance = null;
    target = null;
  });

  return function mount<P extends object>(Ctor: unknown, props: P): HTMLElement {
    target = document.createElement('div');
    document.body.appendChild(target);
    instance = new (Ctor as SvelteInit)({ target, props });
    return target;
  };
}

/**
 * Mount a Svelte 4 component into a caller-owned container. Returns the live
 * instance so tests can call `$set()` / `$destroy()` themselves — useful for
 * reactive-update assertions where `makeMount`'s auto-cleanup is too coarse.
 */
export function mountComponent<P extends object>(
  Ctor: unknown,
  target: HTMLElement,
  props: P,
): MountedComponent {
  return new (Ctor as SvelteInit)({ target, props });
}

// ---------------------------------------------------------------------------
// @xyflow/svelte <Handle> stub
// ---------------------------------------------------------------------------

/**
 * Minimal replacement for `@xyflow/svelte`'s `Handle` + `Position` exports.
 * The real `Handle` requires a `<SvelteFlowProvider />` context that is
 * unavailable in jsdom unit tests; only the three `$$` fields Svelte 4's
 * `mount_component` / `destroy_component` actually read are provided.
 *
 * Usage:
 * ```ts
 * vi.mock('@xyflow/svelte', async () =>
 *   (await import('./mountSvelte')).xyflowHandleStub(),
 * );
 * ```
 */
export function xyflowHandleStub() {
  type Frag = { c(): void; m(): void; p(): void; d(): void; l(): void };
  type Internal = {
    fragment: Frag | null;
    on_mount: unknown[];
    on_destroy: unknown[];
    after_update: unknown[];
  };
  class HandleStub {
    $$: Internal = {
      fragment: {
        c() {},
        m() {},
        p() {},
        d() {},
        l() {},
      },
      on_mount: [],
      on_destroy: [],
      after_update: [],
    };
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    constructor(_opts: unknown) {}
    $set(_props: unknown) {
      void _props;
    }
    $destroy() {
      this.$$.fragment = null;
      this.$$.on_destroy = [];
    }
    $on(_event: string, _fn: unknown) {
      void _event;
      void _fn;
      return () => {};
    }
  }
  return {
    Handle: HandleStub,
    Position: { Left: 'left', Right: 'right', Top: 'top', Bottom: 'bottom' } as const,
  };
}

// ---------------------------------------------------------------------------
// Wails App mock factory (AC-2)
// ---------------------------------------------------------------------------

/**
 * Shape of the Wails-generated App module (`wailsjs/go/main/App`). Every
 * exported binding is a function, so a full mock is a record of
 * method-name → function.
 */
export type WailsAppModule = typeof WailsApp;

/**
 * Partial override map. Any key must match the Wails binding's exact generated
 * signature — misuse surfaces as a svelte-check error at the call site.
 */
export type WailsAppOverrides = Partial<WailsAppModule>;

/**
 * Build a fully-typed stub of the Wails App bindings suitable for
 * `vi.mock('../../../wailsjs/go/main/App', () => createAppMock({ … }))`.
 *
 * Each method defaults to a `vi.fn()` that returns `Promise.resolve(undefined)`
 * — which type-matches every Wails binding because they all return a
 * `Promise<T>` and `undefined` is assignable to `T | undefined` at runtime.
 * Overrides replace specific methods; unspecified methods keep the default
 * no-op and are type-safe to call (but will resolve `undefined`).
 */
export function createAppMock(overrides: WailsAppOverrides = {}): WailsAppModule {
  // The generated module is a bag of named function exports. We build a
  // Proxy-free object by enumerating known method names from the generated
  // module *at runtime* — but the import is type-only here (vitest hoists
  // `vi.mock` before real imports resolve), so we fall back to "provide what
  // the caller uses and Proxy the rest with vi.fn()".
  //
  // A Proxy keeps `typeof module` inference intact: unknown accesses return
  // `vi.fn()` that resolves `undefined`, which the caller can still override
  // per-test via `vi.mocked(App.Foo).mockResolvedValue(...)`.
  const target = { ...overrides } as Record<string, unknown>;
  const handler: ProxyHandler<Record<string, unknown>> = {
    get(t, key: string) {
      if (key in t) return t[key];
      const fn = vi.fn().mockResolvedValue(undefined);
      t[key] = fn;
      return fn;
    },
    has() {
      return true;
    },
  };
  return new Proxy(target, handler) as unknown as WailsAppModule;
}
