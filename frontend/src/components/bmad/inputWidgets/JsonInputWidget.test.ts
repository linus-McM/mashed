/**
 * @vitest-environment jsdom
 */
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { tick } from 'svelte';
import JsonInputWidget from './JsonInputWidget.svelte';
import type { PendingPrompt } from '../../../stores/interactiveInput';

function makePrompt(overrides: Partial<PendingPrompt> = {}): PendingPrompt {
  return {
    nodeId: 'n1',
    inputId: 'payload',
    prompt: 'Send JSON',
    shape: 'json',
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
  const Ctor = JsonInputWidget as unknown as new (opts: { target: HTMLElement; props: object }) => any;
  const component = new Ctor({ target, props: { prompt: makePrompt(), ...props } });
  return { component, target };
}

describe('JsonInputWidget', () => {
  let mounted: Mount | null = null;
  beforeEach(() => { mounted = null; });
  afterEach(() => { mounted?.component?.$destroy(); mounted?.target?.remove(); });

  it('renders a monospace textarea', () => {
    mounted = mount();
    const ta = mounted.target.querySelector<HTMLTextAreaElement>('[data-testid="json-textarea"]');
    expect(ta).not.toBeNull();
  });

  it('invalid JSON keeps submit disabled and does NOT dispatch', async () => {
    mounted = mount();
    const handler = vi.fn();
    mounted.component.$on('submit', handler);
    const ta = mounted.target.querySelector<HTMLTextAreaElement>('[data-testid="json-textarea"]')!;
    ta.value = '{not-json}';
    ta.dispatchEvent(new Event('input'));
    await tick();
    const submit = mounted.target.querySelector<HTMLButtonElement>('[data-testid="json-submit"]')!;
    expect(submit.disabled).toBe(true);
    submit.click();
    await tick();
    expect(handler).not.toHaveBeenCalled();
    expect(mounted.target.querySelector('[data-testid="json-parse-status"] .err')).not.toBeNull();
  });

  it('valid JSON enables submit and dispatches the raw text', async () => {
    mounted = mount();
    const handler = vi.fn();
    mounted.component.$on('submit', handler);
    const ta = mounted.target.querySelector<HTMLTextAreaElement>('[data-testid="json-textarea"]')!;
    ta.value = '{"ok":true}';
    ta.dispatchEvent(new Event('input'));
    await tick();
    const submit = mounted.target.querySelector<HTMLButtonElement>('[data-testid="json-submit"]')!;
    expect(submit.disabled).toBe(false);
    submit.click();
    await tick();
    expect(handler).toHaveBeenCalled();
    expect(handler.mock.calls[0][0].detail.value).toBe('{"ok":true}');
  });
});
