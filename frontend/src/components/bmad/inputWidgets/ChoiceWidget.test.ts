/**
 * @vitest-environment jsdom
 */
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { tick } from 'svelte';
import ChoiceWidget from './ChoiceWidget.svelte';
import type { PendingPrompt } from '../../../stores/interactiveInput';

function makePrompt(overrides: Partial<PendingPrompt> = {}): PendingPrompt {
  return {
    nodeId: 'n1',
    inputId: 'pick',
    prompt: 'Pick one',
    shape: 'choice',
    options: ['a', 'b', 'c'],
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
  const Ctor = ChoiceWidget as unknown as new (opts: { target: HTMLElement; props: object }) => any;
  const component = new Ctor({ target, props: { prompt: makePrompt(), ...props } });
  return { component, target };
}

describe('ChoiceWidget', () => {
  let mounted: Mount | null = null;
  beforeEach(() => { mounted = null; });
  afterEach(() => { mounted?.component?.$destroy(); mounted?.target?.remove(); });

  it('renders radiogroup with all options', () => {
    mounted = mount();
    const group = mounted.target.querySelector('[role="radiogroup"]');
    expect(group).not.toBeNull();
    const radios = mounted.target.querySelectorAll('[role="radio"]');
    expect(radios.length).toBe(3);
  });

  it('clicking an option then pressing Enter submits that value', async () => {
    mounted = mount();
    const handler = vi.fn();
    mounted.component.$on('submit', handler);
    const second = mounted.target.querySelector<HTMLButtonElement>('[data-testid="choice-option-2"]')!;
    second.click();
    await tick();
    expect(second.getAttribute('aria-checked')).toBe('true');
    second.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', bubbles: true }));
    await tick();
    expect(handler).toHaveBeenCalled();
    expect(handler.mock.calls[0][0].detail.value).toBe('b');
  });

  it('ArrowDown moves focus to the next radio', async () => {
    mounted = mount();
    const first = mounted.target.querySelector<HTMLButtonElement>('[data-testid="choice-option-1"]')!;
    first.focus();
    first.dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowDown', bubbles: true }));
    await tick();
    expect(document.activeElement).toBe(mounted.target.querySelector('[data-testid="choice-option-2"]'));
  });
});
