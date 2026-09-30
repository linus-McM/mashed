/**
 * @vitest-environment jsdom
 */
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { tick } from 'svelte';
import FreeTextWidget from './FreeTextWidget.svelte';
import type { PendingPrompt } from '../../../stores/interactiveInput';

function makePrompt(overrides: Partial<PendingPrompt> = {}): PendingPrompt {
  return {
    nodeId: 'n1',
    inputId: 'topic',
    prompt: 'What topic?',
    shape: 'free',
    options: [],
    round: 1,
    createdAt: 0,
    promptId: 'p',
    required: true,
    maxLength: 0,
    ...overrides,
  };
}

type Mount = { component: any; target: HTMLElement };
function mount(props: Record<string, unknown> = {}): Mount {
  const target = document.createElement('div');
  document.body.appendChild(target);
  const Ctor = FreeTextWidget as unknown as new (opts: { target: HTMLElement; props: object }) => any;
  const component = new Ctor({ target, props: { prompt: makePrompt(), ...props } });
  return { component, target };
}

describe('FreeTextWidget', () => {
  let mounted: Mount | null = null;
  beforeEach(() => { mounted = null; });
  afterEach(() => { mounted?.component?.$destroy(); mounted?.target?.remove(); });

  it('renders a textarea with the prompt as aria-label', () => {
    mounted = mount();
    const ta = mounted.target.querySelector<HTMLTextAreaElement>('[data-testid="free-text-textarea"]');
    expect(ta).not.toBeNull();
    expect(ta?.getAttribute('aria-label')).toBe('What topic?');
  });

  it('submits typed value via submit event', async () => {
    mounted = mount();
    const handler = vi.fn();
    mounted.component.$on('submit', handler);
    const ta = mounted.target.querySelector<HTMLTextAreaElement>('[data-testid="free-text-textarea"]')!;
    ta.value = 'hello world';
    ta.dispatchEvent(new Event('input'));
    await tick();
    const btn = mounted.target.querySelector<HTMLButtonElement>('[data-testid="free-text-submit"]')!;
    btn.click();
    await tick();
    expect(handler).toHaveBeenCalled();
    expect(handler.mock.calls[0][0].detail.value).toBe('hello world');
  });

  it('Cmd+Enter submits', async () => {
    mounted = mount();
    const handler = vi.fn();
    mounted.component.$on('submit', handler);
    const ta = mounted.target.querySelector<HTMLTextAreaElement>('[data-testid="free-text-textarea"]')!;
    ta.value = 'x';
    ta.dispatchEvent(new Event('input'));
    await tick();
    ta.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', metaKey: true, bubbles: true }));
    await tick();
    expect(handler).toHaveBeenCalled();
  });

  it('submit button disabled when textarea is empty', async () => {
    mounted = mount();
    await tick();
    const btn = mounted.target.querySelector<HTMLButtonElement>('[data-testid="free-text-submit"]')!;
    expect(btn.disabled).toBe(true);
  });
});
