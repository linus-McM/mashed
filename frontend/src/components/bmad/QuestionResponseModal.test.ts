// Story 4: Question Response Modal & E2E Integration
//
// RED phase tests for QuestionResponseModal.svelte. These tests MUST fail
// initially because the component does not exist yet.
//
// Contract the ui-engineer must satisfy:
//   - File:  frontend/src/components/bmad/QuestionResponseModal.svelte
//   - Props: question (QuestionEventLike)
//   - Events: 'responded', 'close', 'error'
//   - Calls:  RespondToQuestion(execId, nodeId, answer) from Wails bindings
//   - data-testid attributes:
//       question-modal        -- root modal card
//       question-text         -- element containing the full question string
//       answer-textarea       -- freeform textarea
//       option-button-1..N    -- menu option buttons (1-indexed)
//       send-button           -- primary Send button
//       cancel-button         -- Cancel button
//       error-message         -- error bar element (only when error is set)
//
// Testing approach:
//   - Mount the component directly with `new QuestionResponseModal({ target, props })`
//   - Use `component.$set(...)` to update props, `component.$on(...)` for events
//   - Query the DOM via `target.querySelector('[data-testid="..."]')`
//   - Mock `RespondToQuestion` from the Wails bindings with `vi.mock`

import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { tick } from 'svelte';
import type { QuestionEventLike } from './questionSnackbarUtils';

// Mock the Wails binding BEFORE importing the component. The mock factory
// must not reference out-of-scope variables per Vitest's hoisting rules.
vi.mock('../../../wailsjs/go/main/App.js', () => ({
  RespondToQuestion: vi.fn(),
}));

// Import the mocked binding so tests can control its behaviour.
import { RespondToQuestion } from '../../../wailsjs/go/main/App.js';

// Component under test. This import path MUST resolve once the ui-engineer
// creates QuestionResponseModal.svelte. During RED, Vitest will fail to
// resolve the module -- this is the proof the test is red.
// eslint-disable-next-line @typescript-eslint/ban-ts-comment
// @ts-ignore -- component does not exist yet (RED phase)
import QuestionResponseModal from './QuestionResponseModal.svelte';

// ────────────────────────────────────────────────────────────────
// Fixtures
// ────────────────────────────────────────────────────────────────

function mockQuestion(overrides: Partial<QuestionEventLike> = {}): QuestionEventLike {
  return {
    execId: 'exec-1',
    nodeId: 'n1',
    repoPath: '/dev/test',
    repoName: 'test',
    question: 'What file?',
    options: [],
    tmuxTarget: 'bmad-n1-123:0.0',
    timestamp: Date.now(),
    questionId: 'hash-1',
    ...overrides,
  };
}

// Some jsdom configs lack a full localStorage — provide a minimal shim so
// utils that touch it (via shared code) do not blow up the test run.
const memoryStore: Record<string, string> = {};
const fakeStorage: Storage = {
  get length() {
    return Object.keys(memoryStore).length;
  },
  clear() {
    for (const key of Object.keys(memoryStore)) delete memoryStore[key];
  },
  getItem(key: string) {
    return Object.prototype.hasOwnProperty.call(memoryStore, key) ? memoryStore[key] : null;
  },
  setItem(key: string, value: string) {
    memoryStore[key] = String(value);
  },
  removeItem(key: string) {
    delete memoryStore[key];
  },
  key(index: number) {
    return Object.keys(memoryStore)[index] ?? null;
  },
};
Object.defineProperty(globalThis, 'localStorage', {
  value: fakeStorage,
  writable: true,
  configurable: true,
});

// ────────────────────────────────────────────────────────────────
// Mount helpers
// ────────────────────────────────────────────────────────────────

type MountResult = {
  component: any;
  target: HTMLElement;
  destroy: () => void;
};

function mount(props: Record<string, unknown>): MountResult {
  const target = document.createElement('div');
  document.body.appendChild(target);
  const component = new QuestionResponseModal({ target, props });
  return {
    component,
    target,
    destroy: () => {
      component.$destroy?.();
      target.remove();
    },
  };
}

function $(target: HTMLElement, testId: string): HTMLElement | null {
  return target.querySelector(`[data-testid="${testId}"]`);
}

function $all(target: HTMLElement, testIdPrefix: string): HTMLElement[] {
  return Array.from(target.querySelectorAll(`[data-testid^="${testIdPrefix}"]`));
}

// ────────────────────────────────────────────────────────────────
// Tests
// ────────────────────────────────────────────────────────────────

describe('QuestionResponseModal', () => {
  let mounted: MountResult | null = null;

  beforeEach(() => {
    vi.mocked(RespondToQuestion).mockReset();
    memoryStore && Object.keys(memoryStore).forEach((k) => delete memoryStore[k]);
  });

  afterEach(() => {
    mounted?.destroy();
    mounted = null;
  });

  // ──────────────────────────────────────────────────────────────
  // AC-1: Freeform mode renders textarea
  // ──────────────────────────────────────────────────────────────
  describe('AC-1: freeform mode renders textarea', () => {
    it('renders a textarea when question.options is empty', async () => {
      mounted = mount({ question: mockQuestion({ options: [] }) });
      await tick();

      const textarea = $(mounted.target, 'answer-textarea');
      expect(textarea).not.toBeNull();
      expect(textarea?.tagName).toBe('TEXTAREA');
    });

    it('renders no option buttons in freeform mode', async () => {
      mounted = mount({ question: mockQuestion({ options: [] }) });
      await tick();

      const options = $all(mounted.target, 'option-button-');
      expect(options).toHaveLength(0);
    });

    it('textarea has a placeholder prompting the user to respond', async () => {
      mounted = mount({ question: mockQuestion({ options: [] }) });
      await tick();

      const textarea = $(mounted.target, 'answer-textarea') as HTMLTextAreaElement | null;
      expect(textarea).not.toBeNull();
      expect((textarea?.placeholder ?? '').toLowerCase()).toContain('type your response');
    });

    it('renders the full question text in the body', async () => {
      const q = mockQuestion({ question: 'What file should I create?' });
      mounted = mount({ question: q });
      await tick();

      const qText = $(mounted.target, 'question-text');
      expect(qText).not.toBeNull();
      expect(qText?.textContent ?? '').toContain('What file should I create?');
    });
  });

  // ──────────────────────────────────────────────────────────────
  // AC-2: Menu mode renders option buttons
  // ──────────────────────────────────────────────────────────────
  describe('AC-2: menu mode renders option buttons', () => {
    it('renders one button per option with option text visible', async () => {
      mounted = mount({
        question: mockQuestion({ options: ['First', 'Second', 'Third'] }),
      });
      await tick();

      const options = $all(mounted.target, 'option-button-');
      expect(options).toHaveLength(3);
      expect(options[0].textContent ?? '').toContain('First');
      expect(options[1].textContent ?? '').toContain('Second');
      expect(options[2].textContent ?? '').toContain('Third');
    });

    it('does not render a textarea in menu mode', async () => {
      mounted = mount({
        question: mockQuestion({ options: ['A', 'B', 'C'] }),
      });
      await tick();

      const textarea = $(mounted.target, 'answer-textarea');
      expect(textarea).toBeNull();
    });

    it('renders 4 buttons when 4 options are provided', async () => {
      mounted = mount({
        question: mockQuestion({ options: ['A', 'B', 'C', 'D'] }),
      });
      await tick();

      const options = $all(mounted.target, 'option-button-');
      expect(options).toHaveLength(4);
    });
  });

  // ──────────────────────────────────────────────────────────────
  // AC-3: Successful response dispatches 'responded'
  // ──────────────────────────────────────────────────────────────
  describe("AC-3: successful response dispatches 'responded'", () => {
    it('calls RespondToQuestion with (execId, nodeId, answer) and dispatches responded', async () => {
      vi.mocked(RespondToQuestion).mockResolvedValue(undefined as never);

      mounted = mount({
        question: mockQuestion({
          execId: 'exec-42',
          nodeId: 'node-A',
          options: [],
        }),
      });
      await tick();

      const respondedHandler = vi.fn();
      mounted.component.$on('responded', respondedHandler);

      // Type into the textarea and dispatch an input event so bind:value updates.
      const textarea = $(mounted.target, 'answer-textarea') as HTMLTextAreaElement;
      textarea.value = 'src/utils/helper.go';
      textarea.dispatchEvent(new Event('input', { bubbles: true }));
      await tick();

      const sendBtn = $(mounted.target, 'send-button') as HTMLButtonElement;
      sendBtn.click();
      // Flush microtasks so the awaited RespondToQuestion promise settles.
      await tick();
      await Promise.resolve();
      await tick();

      expect(RespondToQuestion).toHaveBeenCalledTimes(1);
      expect(RespondToQuestion).toHaveBeenCalledWith(
        'exec-42',
        'node-A',
        'src/utils/helper.go',
      );
      expect(respondedHandler).toHaveBeenCalledTimes(1);
    });
  });

  // ──────────────────────────────────────────────────────────────
  // AC-4: Failed response shows error message
  // ──────────────────────────────────────────────────────────────
  describe('AC-4: failed response surfaces error, keeps modal open', () => {
    it('shows the backend error text in the DOM', async () => {
      vi.mocked(RespondToQuestion).mockRejectedValue(new Error('pane is dead'));

      mounted = mount({ question: mockQuestion({ options: [] }) });
      await tick();

      const respondedHandler = vi.fn();
      mounted.component.$on('responded', respondedHandler);

      const textarea = $(mounted.target, 'answer-textarea') as HTMLTextAreaElement;
      textarea.value = 'some answer';
      textarea.dispatchEvent(new Event('input', { bubbles: true }));
      await tick();

      const sendBtn = $(mounted.target, 'send-button') as HTMLButtonElement;
      sendBtn.click();
      await tick();
      await Promise.resolve();
      await Promise.resolve();
      await tick();

      const errorEl = $(mounted.target, 'error-message');
      expect(errorEl).not.toBeNull();
      expect(errorEl?.textContent ?? '').toContain('pane is dead');
    });

    it('does NOT dispatch responded on failure', async () => {
      vi.mocked(RespondToQuestion).mockRejectedValue(new Error('pane is dead'));

      mounted = mount({ question: mockQuestion({ options: [] }) });
      await tick();

      const respondedHandler = vi.fn();
      mounted.component.$on('responded', respondedHandler);

      const textarea = $(mounted.target, 'answer-textarea') as HTMLTextAreaElement;
      textarea.value = 'x';
      textarea.dispatchEvent(new Event('input', { bubbles: true }));
      await tick();

      ($(mounted.target, 'send-button') as HTMLButtonElement).click();
      await tick();
      await Promise.resolve();
      await Promise.resolve();
      await tick();

      expect(respondedHandler).not.toHaveBeenCalled();
    });

    it('re-enables the Send button after a failed send', async () => {
      vi.mocked(RespondToQuestion).mockRejectedValue(new Error('pane is dead'));

      mounted = mount({ question: mockQuestion({ options: [] }) });
      await tick();

      const textarea = $(mounted.target, 'answer-textarea') as HTMLTextAreaElement;
      textarea.value = 'x';
      textarea.dispatchEvent(new Event('input', { bubbles: true }));
      await tick();

      const sendBtn = $(mounted.target, 'send-button') as HTMLButtonElement;
      sendBtn.click();
      await tick();
      await Promise.resolve();
      await Promise.resolve();
      await tick();

      expect(sendBtn.disabled).toBe(false);
    });
  });

  // ──────────────────────────────────────────────────────────────
  // AC-6: Cancel dispatches 'close' (keeps snackbar)
  // ──────────────────────────────────────────────────────────────
  describe("AC-6: cancel dispatches 'close'", () => {
    it('clicking the Cancel button dispatches close', async () => {
      mounted = mount({ question: mockQuestion({ options: [] }) });
      await tick();

      const closeHandler = vi.fn();
      mounted.component.$on('close', closeHandler);

      const cancelBtn = $(mounted.target, 'cancel-button') as HTMLButtonElement;
      expect(cancelBtn).not.toBeNull();
      cancelBtn.click();
      await tick();

      expect(closeHandler).toHaveBeenCalledTimes(1);
    });

    it('cancel does NOT call RespondToQuestion', async () => {
      mounted = mount({ question: mockQuestion({ options: [] }) });
      await tick();

      const cancelBtn = $(mounted.target, 'cancel-button') as HTMLButtonElement;
      cancelBtn.click();
      await tick();

      expect(RespondToQuestion).not.toHaveBeenCalled();
    });
  });

  // ──────────────────────────────────────────────────────────────
  // AC-7: Menu option click immediately submits
  // ──────────────────────────────────────────────────────────────
  describe('AC-7: menu option click immediately submits', () => {
    it('clicking the second option submits answer "2"', async () => {
      vi.mocked(RespondToQuestion).mockResolvedValue(undefined as never);

      mounted = mount({
        question: mockQuestion({
          execId: 'exec-99',
          nodeId: 'node-menu',
          options: ['New file', 'Existing file', 'Cancel'],
        }),
      });
      await tick();

      const respondedHandler = vi.fn();
      mounted.component.$on('responded', respondedHandler);

      const options = $all(mounted.target, 'option-button-');
      expect(options).toHaveLength(3);

      (options[1] as HTMLButtonElement).click();
      await tick();
      await Promise.resolve();
      await tick();

      expect(RespondToQuestion).toHaveBeenCalledTimes(1);
      expect(RespondToQuestion).toHaveBeenCalledWith('exec-99', 'node-menu', '2');
      expect(respondedHandler).toHaveBeenCalledTimes(1);
    });

    it('clicking the first option submits answer "1"', async () => {
      vi.mocked(RespondToQuestion).mockResolvedValue(undefined as never);

      mounted = mount({
        question: mockQuestion({
          execId: 'e',
          nodeId: 'n',
          options: ['First', 'Second'],
        }),
      });
      await tick();

      const options = $all(mounted.target, 'option-button-');
      (options[0] as HTMLButtonElement).click();
      await tick();
      await Promise.resolve();
      await tick();

      expect(RespondToQuestion).toHaveBeenCalledWith('e', 'n', '1');
    });
  });

  // ──────────────────────────────────────────────────────────────
  // Additional: Send disabled when textarea empty (BDD Scenario 1)
  // ──────────────────────────────────────────────────────────────
  describe('freeform Send button is disabled when answer is empty', () => {
    it('Send is disabled on initial freeform render', async () => {
      mounted = mount({ question: mockQuestion({ options: [] }) });
      await tick();

      const sendBtn = $(mounted.target, 'send-button') as HTMLButtonElement;
      expect(sendBtn).not.toBeNull();
      expect(sendBtn.disabled).toBe(true);
    });

    it('Send is disabled when textarea contains only whitespace', async () => {
      mounted = mount({ question: mockQuestion({ options: [] }) });
      await tick();

      const textarea = $(mounted.target, 'answer-textarea') as HTMLTextAreaElement;
      textarea.value = '   \n\t  ';
      textarea.dispatchEvent(new Event('input', { bubbles: true }));
      await tick();

      const sendBtn = $(mounted.target, 'send-button') as HTMLButtonElement;
      expect(sendBtn.disabled).toBe(true);
    });

    it('Send becomes enabled when user types a non-empty answer', async () => {
      mounted = mount({ question: mockQuestion({ options: [] }) });
      await tick();

      const textarea = $(mounted.target, 'answer-textarea') as HTMLTextAreaElement;
      textarea.value = 'hello';
      textarea.dispatchEvent(new Event('input', { bubbles: true }));
      await tick();

      const sendBtn = $(mounted.target, 'send-button') as HTMLButtonElement;
      expect(sendBtn.disabled).toBe(false);
    });
  });
});
