/**
 * @vitest-environment jsdom
 */
// ProcessNode — S6 awaiting_input branch + round counter (AC-5, AC-6).
//
// Uses the same Handle mock as ProcessNode.test.ts to avoid needing a
// SvelteFlowProvider in unit tests.

import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest';
import { tick } from 'svelte';

vi.mock('@xyflow/svelte', () => ({
  Handle: class {
    $$ = { fragment: { c() {}, m() {}, p() {}, d() {}, l() {} }, on_mount: [], on_destroy: [], after_update: [] };
    constructor(_opts: unknown) {}
    $set(_props: unknown) {}
    $destroy() { this.$$.fragment = null as any; this.$$.on_destroy = []; }
    $on(_event: string, _fn: unknown) { return () => {}; }
  },
  Position: { Left: 'left', Right: 'right', Top: 'top', Bottom: 'bottom' },
}));

import ProcessNode from './ProcessNode.svelte';

type NodeInstance = { $destroy(): void; $set(p: object): void };
type Ctor = new (opts: { target: HTMLElement; props: object }) => NodeInstance;

function mount(container: HTMLElement, props: object): NodeInstance {
  return new (ProcessNode as unknown as Ctor)({ target: container, props });
}

describe('ProcessNode — awaiting_input (S6)', () => {
  let container: HTMLElement;
  let node: NodeInstance | null = null;

  beforeEach(() => {
    container = document.createElement('div');
    document.body.appendChild(container);
  });

  afterEach(() => {
    node?.$destroy();
    container.remove();
    node = null;
  });

  it('AC-5: renders awaiting badge when status=awaiting_input', () => {
    node = mount(container, {
      id: 'n1',
      selected: false,
      data: {
        status: 'awaiting_input',
        process: { phase: 'analysis', agentRole: 'analyst', inputs: [], outputs: [] },
      },
    });
    const badge = container.querySelector('[data-testid="awaiting-badge"]');
    expect(badge, 'awaiting badge must render').not.toBeNull();
    expect(container.querySelector('.awaiting-text')?.textContent?.trim()).toBe('awaiting');
  });

  it('AC-5: renders "N / M" round counter when nodeRound + maxRounds known', () => {
    node = mount(container, {
      id: 'n1',
      selected: false,
      data: {
        status: 'awaiting_input',
        nodeRound: 3,
        process: {
          phase: 'analysis', agentRole: 'analyst', inputs: [], outputs: [],
          gate: { kind: 'userConfirm', maxRounds: 30 },
        },
      },
    });
    const counter = container.querySelector('[data-testid="round-counter"]');
    expect(counter?.textContent?.trim()).toBe('3 / 30');
  });

  it('AC-6: updating nodeRound reflects in counter without unmount', async () => {
    node = mount(container, {
      id: 'n1',
      selected: false,
      data: {
        status: 'awaiting_input', nodeRound: 2,
        process: { phase: 'analysis', agentRole: 'analyst', inputs: [], outputs: [], gate: { kind: 'userConfirm', maxRounds: 30 } },
      },
    });
    expect(container.querySelector('[data-testid="round-counter"]')?.textContent?.trim()).toBe('2 / 30');
    node.$set({
      data: {
        status: 'awaiting_input', nodeRound: 3,
        process: { phase: 'analysis', agentRole: 'analyst', inputs: [], outputs: [], gate: { kind: 'userConfirm', maxRounds: 30 } },
      },
    });
    await tick();
    expect(container.querySelector('[data-testid="round-counter"]')?.textContent?.trim()).toBe('3 / 30');
  });

  it('falls back to "N rounds" when maxRounds is unknown', () => {
    node = mount(container, {
      id: 'n1',
      selected: false,
      data: {
        status: 'awaiting_input', nodeRound: 5,
        process: { phase: 'analysis', agentRole: 'analyst', inputs: [], outputs: [] },
      },
    });
    expect(container.querySelector('[data-testid="round-counter"]')?.textContent?.trim()).toBe('5 rounds');
  });
});
