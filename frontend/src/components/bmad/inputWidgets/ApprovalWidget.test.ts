/**
 * @vitest-environment jsdom
 */
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { tick } from 'svelte';
import ApprovalWidget from './ApprovalWidget.svelte';
import type { PendingPrompt } from '../../../stores/interactiveInput';

function makePrompt(overrides: Partial<PendingPrompt> = {}): PendingPrompt {
  return {
    nodeId: 'n1',
    inputId: 'ok',
    prompt: 'Proceed?',
    shape: 'approval',
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
  const Ctor = ApprovalWidget as unknown as new (opts: { target: HTMLElement; props: object }) => any;
  const component = new Ctor({ target, props: { prompt: makePrompt(), ...props } });
  return { component, target };
}

describe('ApprovalWidget', () => {
  let mounted: Mount | null = null;
  beforeEach(() => { mounted = null; });
  afterEach(() => { mounted?.component?.$destroy(); mounted?.target?.remove(); });

  it('renders Yes/No buttons', () => {
    mounted = mount();
    expect(mounted.target.querySelector('[data-testid="approval-yes"]')).not.toBeNull();
    expect(mounted.target.querySelector('[data-testid="approval-no"]')).not.toBeNull();
  });

  it('clicking Yes emits value "yes"', async () => {
    mounted = mount();
    const handler = vi.fn();
    mounted.component.$on('submit', handler);
    mounted.target.querySelector<HTMLButtonElement>('[data-testid="approval-yes"]')!.click();
    await tick();
    expect(handler.mock.calls[0][0].detail.value).toBe('yes');
  });

  it('clicking No emits value "no"', async () => {
    mounted = mount();
    const handler = vi.fn();
    mounted.component.$on('submit', handler);
    mounted.target.querySelector<HTMLButtonElement>('[data-testid="approval-no"]')!.click();
    await tick();
    expect(handler.mock.calls[0][0].detail.value).toBe('no');
  });
});
