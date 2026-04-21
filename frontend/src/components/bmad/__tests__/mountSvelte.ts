// Minimal Svelte-4 mount helper for component tests. @testing-library/svelte
// is not a project dep — this wraps the raw constructor API with a fresh
// target element and auto-cleanup registered via afterEach.

import { afterEach } from 'vitest';

export type MountedComponent = { $destroy(): void };
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
