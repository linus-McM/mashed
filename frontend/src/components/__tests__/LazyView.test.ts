/**
 * @vitest-environment jsdom
 */
// LazyView (R31): renders an accessible aria-busy placeholder until the
// dynamic import resolves, then mounts the loaded component with the
// pass-through props.

import { describe, it, expect, afterEach } from 'vitest';
import { tick } from 'svelte';
import LazyView from '../LazyView.svelte';
import LazyProbe from './fixtures/LazyProbe.svelte';

type Instance = { $destroy(): void; $set(p: object): void };
type Ctor = new (opts: { target: HTMLElement; props: object }) => Instance;

function deferred<T>() {
  let resolve!: (v: T) => void;
  let reject!: (e: unknown) => void;
  const promise = new Promise<T>((res, rej) => {
    resolve = res;
    reject = rej;
  });
  return { promise, resolve, reject };
}

async function flush(): Promise<void> {
  for (let i = 0; i < 5; i++) {
    await Promise.resolve();
    await tick();
  }
}

describe('LazyView', () => {
  let container: HTMLElement;
  let view: Instance | null = null;

  function mount(props: object): Instance {
    container = document.createElement('div');
    document.body.appendChild(container);
    view = new (LazyView as unknown as Ctor)({ target: container, props });
    return view;
  }

  afterEach(() => {
    view?.$destroy();
    view = null;
    container?.remove();
  });

  it('renders an aria-busy status placeholder until the loader resolves', async () => {
    const d = deferred<{ default: typeof LazyProbe }>();
    mount({ loader: () => d.promise });
    await flush();

    const busy = container.querySelector('[aria-busy="true"]');
    expect(busy).not.toBeNull();
    expect(busy?.getAttribute('role')).toBe('status');
    expect(container.querySelector('[data-testid="lazy-probe"]')).toBeNull();

    d.resolve({ default: LazyProbe });
    await flush();

    expect(container.querySelector('[aria-busy="true"]')).toBeNull();
    expect(container.querySelector('[data-testid="lazy-probe"]')).not.toBeNull();
  });

  it('passes the remaining props through to the loaded component', async () => {
    const instance = mount({
      loader: () => Promise.resolve({ default: LazyProbe }),
      repoPath: '/repo',
      count: 3,
    });
    await flush();

    const probe = container.querySelector('[data-testid="lazy-probe"]') as HTMLElement;
    expect(probe).not.toBeNull();
    const received = JSON.parse(probe.dataset.props ?? '{}');
    expect(received).toEqual({ repoPath: '/repo', count: 3 });

    instance.$set({ repoPath: '/other' });
    await flush();
    expect(JSON.parse(probe.dataset.props ?? '{}').repoPath).toBe('/other');
  });

  it('shows an alert instead of spinning forever when the import fails', async () => {
    mount({ loader: () => Promise.reject(new Error('chunk missing')) });
    await flush();

    expect(container.querySelector('[aria-busy="true"]')).toBeNull();
    const alert = container.querySelector('[role="alert"]');
    expect(alert?.textContent).toContain('chunk missing');
  });

  it('renders the loaded component synchronously on remount (module cache)', async () => {
    const loader = () => Promise.resolve({ default: LazyProbe });
    mount({ loader });
    await flush();
    view?.$destroy();
    container.remove();

    mount({ loader });
    expect(container.querySelector('[aria-busy="true"]')).toBeNull();
    expect(container.querySelector('[data-testid="lazy-probe"]')).not.toBeNull();
  });
});
