/**
 * @vitest-environment jsdom
 */
// EditorRouter (R31): MarkdownEditor is code-split behind a dynamic import so
// it stays out of the entry chunk. Until the import resolves an aria-busy
// placeholder renders; afterwards the editor mounts with the routed props.

import { describe, it, expect, afterEach, vi } from 'vitest';
import { tick } from 'svelte';

vi.mock('../MarkdownEditor.svelte', async () => ({
  default: (await import('./fixtures/LazyProbe.svelte')).default,
}));

import EditorRouter from '../EditorRouter.svelte';

type Instance = { $destroy(): void };
type Ctor = new (opts: { target: HTMLElement; props: object }) => Instance;

async function flush(): Promise<void> {
  for (let i = 0; i < 5; i++) {
    await Promise.resolve();
    await tick();
  }
}

describe('EditorRouter — lazy MarkdownEditor', () => {
  let container: HTMLElement;
  let router: Instance | null = null;

  afterEach(() => {
    router?.$destroy();
    router = null;
    container?.remove();
  });

  it('renders an aria-busy placeholder, then mounts MarkdownEditor with its props', async () => {
    container = document.createElement('div');
    document.body.appendChild(container);
    router = new (EditorRouter as unknown as Ctor)({
      target: container,
      props: { filePath: 'docs/README.md', repoPath: '/repo', editable: true },
    });

    // Synchronously after mount the dynamic import cannot have resolved yet.
    const busy = container.querySelector('[aria-busy="true"]');
    expect(busy).not.toBeNull();
    expect(busy?.getAttribute('role')).toBe('status');
    expect(container.querySelector('[data-testid="lazy-probe"]')).toBeNull();

    // The mocked module resolves through vitest's async module graph, which
    // takes an unknown number of ticks; poll rather than count microtasks.
    await vi.waitFor(() => {
      expect(container.querySelector('[data-testid="lazy-probe"]')).not.toBeNull();
    });
    await flush();

    expect(container.querySelector('[aria-busy="true"]')).toBeNull();
    const editor = container.querySelector('[data-testid="lazy-probe"]') as HTMLElement;
    expect(editor).not.toBeNull();
    expect(JSON.parse(editor.dataset.props ?? '{}')).toEqual({
      filePath: 'docs/README.md',
      repoPath: '/repo',
      editable: true,
    });
  });
});
