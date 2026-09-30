/**
 * @vitest-environment jsdom
 */
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { tick } from 'svelte';
import MultiChoiceWidget from './MultiChoiceWidget.svelte';
import type { PendingPrompt } from '../../../stores/interactiveInput';

function makePrompt(overrides: Partial<PendingPrompt> = {}): PendingPrompt {
  return {
    nodeId: 'n1',
    inputId: 'multi',
    prompt: 'Pick many',
    shape: 'multi',
    options: ['x', 'y', 'z'],
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
  const Ctor = MultiChoiceWidget as unknown as new (opts: { target: HTMLElement; props: object }) => any;
  const component = new Ctor({ target, props: { prompt: makePrompt(), ...props } });
  return { component, target };
}

describe('MultiChoiceWidget', () => {
  let mounted: Mount | null = null;
  beforeEach(() => { mounted = null; });
  afterEach(() => { mounted?.component?.$destroy(); mounted?.target?.remove(); });

  it('renders one checkbox per option', () => {
    mounted = mount();
    const boxes = mounted.target.querySelectorAll('input[type="checkbox"]');
    expect(boxes.length).toBe(3);
  });

  it('submit emits comma-joined value of checked boxes (AC-2)', async () => {
    mounted = mount();
    const handler = vi.fn();
    mounted.component.$on('submit', handler);
    const x = mounted.target.querySelector<HTMLInputElement>('[data-testid="multi-checkbox-1"]')!;
    const z = mounted.target.querySelector<HTMLInputElement>('[data-testid="multi-checkbox-3"]')!;
    x.click(); await tick();
    z.click(); await tick();
    const submit = mounted.target.querySelector<HTMLButtonElement>('[data-testid="multi-submit"]')!;
    submit.click();
    await tick();
    expect(handler).toHaveBeenCalled();
    expect(handler.mock.calls[0][0].detail.value).toBe('x,z');
  });

  it('submit button disabled when nothing is checked', async () => {
    mounted = mount();
    await tick();
    const submit = mounted.target.querySelector<HTMLButtonElement>('[data-testid="multi-submit"]')!;
    expect(submit.disabled).toBe(true);
  });
});
