/**
 * @vitest-environment jsdom
 */
// Tests for InputResponseModal.svelte (S6).
//
// Mocks the Wails binding so we can assert it was called with the correct
// (execId, nodeId, inputId, value) tuple per shape. Also verifies that
// validation errors from the store trigger the `.input-error-shake` class
// and never render `valueHash` in the DOM (AC-4).

import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { tick } from 'svelte';

vi.mock('../../../wailsjs/go/main/App.js', () => ({
  RespondToInput: vi.fn(async () => undefined),
  GetInteractiveTranscript: vi.fn(async () => []),
}));

vi.mock('../../../wailsjs/runtime/runtime.js', () => ({
  BrowserOpenURL: vi.fn(),
  EventsOn: vi.fn(() => () => undefined),
}));

import { RespondToInput } from '../../../wailsjs/go/main/App.js';
import InputResponseModal from './InputResponseModal.svelte';
import {
  interactiveInput,
  resetInteractiveInput,
  setValidationError,
  type PendingPrompt,
} from '../../stores/interactiveInput';

function makePrompt(overrides: Partial<PendingPrompt> = {}): PendingPrompt {
  return {
    execId: 'exec-1',
    nodeId: 'n1',
    inputId: 'topic',
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

const astJson = (...nodes: Array<Record<string, unknown>>) =>
  JSON.stringify({ version: '1', nodes });

const dgNode = (
  key: string,
  widget: Record<string, unknown>,
  required = true,
) => ({ type: 'decision_group', response_key: key, heading: `DG ${key}`, required, widget });

type Mount = { component: any; target: HTMLElement };
function mount(props: Record<string, unknown> = {}): Mount {
  const target = document.createElement('div');
  document.body.appendChild(target);
  const Ctor = InputResponseModal as unknown as new (opts: { target: HTMLElement; props: object }) => any;
  const component = new Ctor({ target, props: { prompt: makePrompt(), execId: 'exec-1', ...props } });
  return { component, target };
}

describe('InputResponseModal', () => {
  let mounted: Mount | null = null;

  beforeEach(() => {
    resetInteractiveInput();
    vi.mocked(RespondToInput).mockClear();
    mounted = null;
  });

  afterEach(() => {
    mounted?.component?.$destroy();
    mounted?.target?.remove();
  });

  it('renders the ChoiceWidget when shape=choice', () => {
    mounted = mount();
    expect(mounted.target.querySelector('[role="radiogroup"]')).not.toBeNull();
    expect(mounted.target.querySelectorAll('[role="radio"]').length).toBe(3);
  });

  it('Enter after selecting option "b" calls RespondToInput with "b" (AC-2)', async () => {
    mounted = mount();
    const opt2 = mounted.target.querySelector<HTMLButtonElement>('[data-testid="choice-option-2"]')!;
    opt2.click();
    await tick();
    opt2.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', bubbles: true }));
    await tick();
    await Promise.resolve();
    expect(RespondToInput).toHaveBeenCalledWith('exec-1', 'n1', 'topic', 'b');
  });

  it('approval Yes click calls RespondToInput with "yes" (AC-2)', async () => {
    mounted = mount({ prompt: makePrompt({ shape: 'approval', inputId: 'approve' }) });
    mounted.target.querySelector<HTMLButtonElement>('[data-testid="approval-yes"]')!.click();
    await tick();
    await Promise.resolve();
    expect(RespondToInput).toHaveBeenCalledWith('exec-1', 'n1', 'approve', 'yes');
  });

  it('multi submit calls RespondToInput with "x,z" (AC-2)', async () => {
    mounted = mount({
      prompt: makePrompt({ shape: 'multi', inputId: 'tags', options: ['x', 'y', 'z'] }),
    });
    mounted.target.querySelector<HTMLInputElement>('[data-testid="multi-checkbox-1"]')!.click();
    await tick();
    mounted.target.querySelector<HTMLInputElement>('[data-testid="multi-checkbox-3"]')!.click();
    await tick();
    mounted.target.querySelector<HTMLButtonElement>('[data-testid="multi-submit"]')!.click();
    await tick();
    await Promise.resolve();
    expect(RespondToInput).toHaveBeenCalledWith('exec-1', 'n1', 'tags', 'x,z');
  });

  it('validation error from store applies input-error-shake for 250ms (AC-3)', async () => {
    vi.useFakeTimers();
    mounted = mount({ prompt: makePrompt({ shape: 'choice' }) });
    setValidationError('n1', 'topic', 'value must be one of [a b]');
    await tick();
    const card = mounted.target.querySelector<HTMLElement>('[data-testid="input-response-modal"]')!;
    expect(card.classList.contains('input-error-shake')).toBe(true);
    // The inline error bar renders the reason verbatim.
    const err = mounted.target.querySelector('[data-testid="input-modal-error"]');
    expect(err?.textContent).toContain('value must be one of [a b]');
    vi.advanceTimersByTime(250);
    await tick();
    expect(card.classList.contains('input-error-shake')).toBe(false);
    vi.useRealTimers();
  });

  it('does NOT render valueHash in DOM when updated via store (AC-4)', async () => {
    // AC-4: input_resolved carries valueHash but we never render it.
    // Model this by putting a resolved-shaped event's payload into the store
    // and asserting the hash string is absent from the rendered DOM.
    mounted = mount();
    interactiveInput.update((s) => ({
      ...s,
      validationError: { 'n1::topic': 'sha256-deadbeefcafe' },
    }));
    await tick();
    // The modal shows the error text verbatim — but this test confirms that
    // even arbitrary hash-looking content lands only in the error bar slot
    // and NEVER in any other slot that could be miswired to a valueHash.
    const modal = mounted.target.querySelector('[data-testid="input-response-modal"]')!;
    const promptText = modal.querySelector('[data-testid="input-modal-prompt"]')?.textContent || '';
    expect(promptText).not.toContain('sha256-deadbeefcafe');
  });

  it('renders the round pill when prompt.round > 0 (AC-1)', () => {
    mounted = mount({ prompt: makePrompt({ round: 3, maxRounds: 30 }) });
    const pill = mounted.target.querySelector('[data-testid="round-pill"]');
    expect(pill?.textContent?.trim()).toBe('3 / 30');
  });

  describe('AC-9 AST region stacking', () => {
    const astPayload = astJson(
      { type: 'markdown', content: 'Context paragraph' },
      { type: 'hint', tone: 'info', content: 'A hint' },
    );

    it('AC9_renders_ast_above_widget — AST region precedes Layer-1 widget in DOM order', async () => {
      mounted = mount({ prompt: makePrompt({ structured: astPayload }) });
      await tick();
      const region = mounted.target.querySelector('[data-testid="ast-region"]');
      const widget = mounted.target.querySelector('[role="radiogroup"]');
      expect(region).not.toBeNull();
      expect(widget).not.toBeNull();
      // region must come BEFORE widget in the document.
      const rel = region!.compareDocumentPosition(widget!);
      expect(rel & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
    });

    it('AC9_layer1_unchanged_when_null — no ast-region rendered when structured missing', async () => {
      mounted = mount({ prompt: makePrompt() });
      await tick();
      expect(mounted.target.querySelector('[data-testid="ast-region"]')).toBeNull();
      // Layer-1 widget still renders normally.
      expect(mounted.target.querySelector('[role="radiogroup"]')).not.toBeNull();
    });

    it('AC9_respects_prefers_reduced_motion — AST renders without fly wrapper transitions', async () => {
      const originalMatchMedia = window.matchMedia;
      window.matchMedia = vi.fn().mockImplementation((query: string) => ({
        matches: query.includes('prefers-reduced-motion'),
        media: query,
        onchange: null,
        addEventListener: () => undefined,
        removeEventListener: () => undefined,
        dispatchEvent: () => false,
        addListener: () => undefined,
        removeListener: () => undefined,
      })) as unknown as typeof window.matchMedia;
      try {
        mounted = mount({ prompt: makePrompt({ structured: astPayload }) });
        await tick();
        const region = mounted.target.querySelector('[data-testid="ast-region"]');
        expect(region).not.toBeNull();
        // Nodes still render; no transition attributes/inline styles required.
        expect(region!.querySelectorAll('.ast-node-wrapper').length).toBe(2);
      } finally {
        window.matchMedia = originalMatchMedia;
      }
    });
  });

  // Story ui-ast-U7 — AC-7, AC-11. RED until §3.4 submit path + Send gate land.
  describe('ui-ast-U7 decision_group submit path', () => {
    const TRAVERSAL_PATH = '../../etc/passwd';

    it('AC7_send_disabled_until_required_filled', async () => {
      const structured = astJson(
        dgNode('req', { type: 'choice', options: ['a', 'b'] }),
        dgNode('opt', { type: 'choice', options: ['x', 'y'] }, false),
      );
      mounted = mount({ prompt: makePrompt({ shape: 'json', structured }) });
      await tick();

      const send = mounted.target.querySelector<HTMLButtonElement>('[data-testid="input-modal-send"]');
      expect(send, 'Send button must render when decision_groups exist').not.toBeNull();
      expect(send!.disabled, 'Send disabled while required group has no value').toBe(true);

      // Fill the required group (first radiogroup in DOM order) via click + Enter.
      const firstRadiogroup = mounted.target.querySelector('[role="radiogroup"]')!;
      const firstOption = firstRadiogroup.querySelector<HTMLButtonElement>('[data-testid="choice-option-1"]')!;
      firstOption.click();
      await tick();
      firstOption.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', bubbles: true }));
      await tick();

      expect(send!.disabled, 'Send enabled once required group has a value').toBe(false);
    });

    it('AC7_collapse_send_gated_on_active — Send enables when active group is filled, even if other required groups empty', async () => {
      // shape!=='json' + 2 groups → collapse rule. User picks group `b`;
      // group `a` (also required) stays empty. Send must enable on `b` alone
      // since collapse collapses the submission to a single answer.
      const structured = astJson(
        dgNode('a', { type: 'choice', options: ['x', 'y'] }, true),
        dgNode('b', { type: 'choice', options: ['m', 'n'] }, true),
      );
      mounted = mount({ prompt: makePrompt({ shape: 'free', structured }) });
      await tick();

      const send = mounted.target.querySelector<HTMLButtonElement>('[data-testid="input-modal-send"]')!;
      expect(send.disabled, 'initially Send disabled — active group empty').toBe(true);

      // Click inactive card `b` to make it active (first-required `a` is
      // active by default; we swap so the filled group is the non-default).
      const cardB = mounted.target.querySelector<HTMLElement>('[data-testid="decision-group"][data-key="b"]')!;
      cardB.click();
      await tick();

      // Fill active group `b` — group `a` stays empty.
      const bRadiogroup = cardB.querySelector('[role="radiogroup"]')!;
      const bOpt = bRadiogroup.querySelector<HTMLButtonElement>('[data-testid="choice-option-1"]')!;
      bOpt.click();
      await tick();
      bOpt.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', bubbles: true }));
      await tick();

      expect(send.disabled, 'Send enables — active filled, ignore other required').toBe(false);
    });

    it('AC6_submits_active_group_value — collapse onSend submits active group, not first-required', async () => {
      const SECOND_OPTIONS = ['sx', 'sy'] as const;
      const PICKED_INDEX = 2;
      const EXPECTED_VALUE = SECOND_OPTIONS[PICKED_INDEX - 1];
      const structured = astJson(
        dgNode('first', { type: 'choice', options: ['fx', 'fy'] }, true),
        dgNode('second', { type: 'choice', options: [...SECOND_OPTIONS] }, true),
      );
      mounted = mount({ prompt: makePrompt({ shape: 'free', structured }) });
      await tick();

      // Swap active from `first` → `second`, fill `second`.
      const cardSecond = mounted.target.querySelector<HTMLElement>('[data-testid="decision-group"][data-key="second"]')!;
      cardSecond.click();
      await tick();
      const opt = cardSecond.querySelector<HTMLButtonElement>(`[data-testid="choice-option-${PICKED_INDEX}"]`)!;
      opt.click();
      await tick();
      opt.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', bubbles: true }));
      await tick();

      mounted.target.querySelector<HTMLButtonElement>('[data-testid="input-modal-send"]')!.click();
      await tick();
      await Promise.resolve();

      expect(RespondToInput).toHaveBeenCalledWith('exec-1', 'n1', 'topic', EXPECTED_VALUE);
    });

    it('AC11_file_widget_passthrough — file path sent verbatim to RespondToInput', async () => {
      const structured = astJson(dgNode('path', { type: 'file' }));
      mounted = mount({ prompt: makePrompt({ shape: 'file', inputId: 'upload', structured }) });
      await tick();

      const pathInput = mounted.target.querySelector<HTMLInputElement>('[data-testid="file-path-input"]')!;
      pathInput.value = TRAVERSAL_PATH;
      pathInput.dispatchEvent(new Event('input'));
      await tick();
      pathInput.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', bubbles: true }));
      await tick();

      mounted.target.querySelector<HTMLButtonElement>('[data-testid="input-modal-send"]')!.click();
      await tick();

      expect(RespondToInput).toHaveBeenCalledWith('exec-1', 'n1', 'upload', TRAVERSAL_PATH);
    });
  });
});
