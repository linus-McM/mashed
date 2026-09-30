/**
 * @vitest-environment jsdom
 */
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { tick } from 'svelte';
import FileInputWidget from './FileInputWidget.svelte';
import type { PendingPrompt } from '../../../stores/interactiveInput';

function makePrompt(overrides: Partial<PendingPrompt> = {}): PendingPrompt {
  return {
    nodeId: 'n1',
    inputId: 'path',
    prompt: 'Pick a file',
    shape: 'file',
    options: [],
    round: 1,
    createdAt: 0,
    promptId: 'p',
    required: true,
    ...overrides,
  };
}

type Mount = { component: any; target: HTMLElement };
function mount(props: Record<string, unknown> = {}): Mount {
  const target = document.createElement('div');
  document.body.appendChild(target);
  const Ctor = FileInputWidget as unknown as new (opts: { target: HTMLElement; props: object }) => any;
  const component = new Ctor({ target, props: { prompt: makePrompt(), ...props } });
  return { component, target };
}

describe('FileInputWidget', () => {
  let mounted: Mount | null = null;
  beforeEach(() => { mounted = null; });
  afterEach(() => { mounted?.component?.$destroy(); mounted?.target?.remove(); });

  it('renders a drop-zone and a path text input', () => {
    mounted = mount();
    expect(mounted.target.querySelector('[data-testid="file-path-input"]')).not.toBeNull();
    expect(mounted.target.querySelector('.drop-zone')).not.toBeNull();
  });

  it('pressing Enter in path input submits the raw path', async () => {
    mounted = mount();
    const handler = vi.fn();
    mounted.component.$on('submit', handler);
    const input = mounted.target.querySelector<HTMLInputElement>('[data-testid="file-path-input"]')!;
    input.value = '/etc/passwd';
    input.dispatchEvent(new Event('input'));
    await tick();
    input.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', bubbles: true }));
    await tick();
    expect(handler).toHaveBeenCalled();
    expect(handler.mock.calls[0][0].detail.value).toBe('/etc/passwd');
  });

  it('submit button disabled until path is non-empty', async () => {
    mounted = mount();
    await tick();
    const btn = mounted.target.querySelector<HTMLButtonElement>('[data-testid="file-submit"]')!;
    expect(btn.disabled).toBe(true);
  });
});
