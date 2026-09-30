/**
 * @vitest-environment jsdom
 *
 * Story ui-ast-U7 — AC-1 + AC-10.
 * RED phase: `DecisionGroup.svelte` does not exist yet; these tests must fail
 * until the component is created and wired per spec §6.2.
 */
import { describe, it, expect, vi } from 'vitest';
import { writable } from 'svelte/store';

vi.mock('../../../../wailsjs/runtime/runtime.js', () => ({
  BrowserOpenURL: vi.fn(),
}));

import { tick } from 'svelte';
import { get } from 'svelte/store';
import DecisionGroup from '../DecisionGroup.svelte';
import { makeMount } from './mountSvelte';

const CASES = [
  { kind: 'choice', sel: '.choice-widget', options: ['a', 'b'] },
  { kind: 'multi', sel: '.multi-widget', options: ['a', 'b'] },
  { kind: 'approval', sel: '.approval-widget' },
  { kind: 'free', sel: '.free-text-widget' },
  { kind: 'file', sel: '.file-widget' },
  { kind: 'json', sel: '.json-widget' },
] as const;

describe('DecisionGroup dispatcher', () => {
  const mount = makeMount();
  const render = (node: Record<string, unknown>) =>
    mount(DecisionGroup, { node, responses: writable<Record<string, string>>({}) });

  for (const { kind, sel, options } of CASES) {
    it(`AC1_dispatches_six_widget_types — widget.type="${kind}" mounts matching widget`, () => {
      const widget: Record<string, unknown> = { type: kind };
      if (options) widget.options = options;
      const node = { type: 'decision_group', response_key: 'k', heading: 'H', widget };
      expect(render(node).querySelector(sel)).not.toBeNull();
    });
  }

  it('AC10_aria_labelledby_wired — heading id matches widget-region aria-labelledby', () => {
    const node = {
      type: 'decision_group',
      response_key: 'output-sink',
      heading: 'Output sink',
      widget: { type: 'choice', options: ['a', 'b'] },
    };
    const target = render(node);

    const heading = target.querySelector('#dg-output-sink');
    expect(heading, 'heading must render with id="dg-output-sink"').not.toBeNull();
    expect(heading?.tagName.toLowerCase()).toBe('h3');

    expect(
      target.querySelector('[aria-labelledby="dg-output-sink"]'),
      'a widget-region element must carry aria-labelledby="dg-output-sink"',
    ).not.toBeNull();
  });

  it('writes submitted value into responses store keyed by response_key', async () => {
    const responses = writable<Record<string, string>>({});
    const target = document.createElement('div');
    document.body.appendChild(target);
    const Ctor = DecisionGroup as unknown as new (o: { target: HTMLElement; props: object }) => { $destroy(): void };
    const instance = new Ctor({
      target,
      props: {
        node: {
          type: 'decision_group',
          response_key: 'pick',
          heading: 'Pick',
          widget: { type: 'choice', options: ['a', 'b'] },
        },
        responses,
      },
    });

    const option = target.querySelector<HTMLButtonElement>('[data-testid="choice-option-2"]')!;
    option.click();
    await tick();
    option.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', bubbles: true }));
    await tick();

    expect(get(responses)).toEqual({ pick: 'b' });
    instance.$destroy();
    target.remove();
  });

  it('renders prompt, help, and Required pill when provided', () => {
    const target = render({
      type: 'decision_group',
      response_key: 'x',
      heading: 'H',
      prompt: 'Prompt copy',
      help: 'Help copy',
      required: true,
      widget: { type: 'choice', options: ['a'] },
    });
    expect(target.querySelector('.prompt')?.textContent).toBe('Prompt copy');
    expect(target.querySelector('.help')?.textContent).toBe('Help copy');
    expect(target.querySelector('[data-testid="required-pill"]')).not.toBeNull();
  });

  it('emits "activate" event when a disabled card is clicked or Enter pressed', async () => {
    const responses = writable<Record<string, string>>({});
    const target = document.createElement('div');
    document.body.appendChild(target);
    const Ctor = DecisionGroup as unknown as new (o: { target: HTMLElement; props: object }) => {
      $destroy(): void;
      $on(ev: string, cb: (e: CustomEvent) => void): () => void;
    };
    const instance = new Ctor({
      target,
      props: {
        node: {
          type: 'decision_group',
          response_key: 'q2',
          heading: 'Q2',
          widget: { type: 'free' },
        },
        responses,
        disabled: true,
      },
    });
    const calls: string[] = [];
    instance.$on('activate', (e) => calls.push((e.detail as { key: string }).key));

    const card = target.querySelector<HTMLElement>('[data-testid="decision-group"]')!;
    card.click();
    await tick();
    card.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', bubbles: true }));
    await tick();
    card.dispatchEvent(new KeyboardEvent('keydown', { key: ' ', bubbles: true }));
    await tick();

    expect(calls).toEqual(['q2', 'q2', 'q2']);
    instance.$destroy();
    target.remove();
  });

  it('FIX15_active_prop_applies_is_active_class — active=true → .is-active (green border pre-focus)', () => {
    const responses = writable<Record<string, string>>({});
    const node = {
      type: 'decision_group',
      response_key: 'k',
      heading: 'H',
      widget: { type: 'choice', options: ['a', 'b'] },
    };
    const target = mount(DecisionGroup, { node, responses, active: true });
    expect(target.querySelector('.decision-group.is-active')).not.toBeNull();
  });

  it('does not emit activate when enabled and keyboard is other key', async () => {
    const responses = writable<Record<string, string>>({});
    const target = document.createElement('div');
    document.body.appendChild(target);
    const Ctor = DecisionGroup as unknown as new (o: { target: HTMLElement; props: object }) => {
      $destroy(): void;
      $on(ev: string, cb: (e: CustomEvent) => void): () => void;
    };
    const instance = new Ctor({
      target,
      props: {
        node: {
          type: 'decision_group',
          response_key: 'q3',
          heading: 'Q3',
          widget: { type: 'free' },
        },
        responses,
        disabled: false,
      },
    });
    const calls: string[] = [];
    instance.$on('activate', (e) => calls.push((e.detail as { key: string }).key));

    const card = target.querySelector<HTMLElement>('[data-testid="decision-group"]')!;
    card.click();
    card.dispatchEvent(new KeyboardEvent('keydown', { key: 'Tab', bubbles: true }));
    await tick();

    expect(calls).toEqual([]);
    instance.$destroy();
    target.remove();
  });
});
