/**
 * ui-ast-U7 — DecisionGroup + §3.4 submit path + collapse AC suite.
 *
 * Drives the Input modal's decision_group flow end-to-end against a running
 * Wails dev server. Uses the dev harness `window.__mashedEmitBmadEvent` to
 * feed an `awaiting_input` event whose `structured` payload describes the
 * decision_group AST. `window.go.main.App.RespondToInput` is replaced with a
 * recording stub so assertions inspect the payload without round-tripping
 * through the backend.
 *
 * Gated by `MASHED_E2E=1` so CI runs without the dev server skip cleanly.
 * Run locally with:
 *
 *   MASHED_E2E=1 npx playwright test --config frontend/playwright.config.ts \
 *     tests/ac/ui-ast-decision-group.spec.ts
 */
import { test, expect, type Page } from '@playwright/test';

const E2E = process.env.MASHED_E2E === '1';
const APP_URL = process.env.MASHED_DEV_URL ?? 'http://localhost:34115';

const EXEC_ID = 'exec-ac';
const NODE_ID = 'n-ast';
const INPUT_ID = 'topic';
const PROMPT_ID = 'p-ast';

interface StructuredPromptOpts {
  structured?: object;
  shape?: string;
}

async function emitAwaitingInput(page: Page, opts: StructuredPromptOpts = {}): Promise<void> {
  const structured = opts.structured !== undefined ? JSON.stringify(opts.structured) : undefined;
  await page.evaluate(
    ({ structured: s, shape, execId, nodeId, inputId, promptId }) => {
      (window as unknown as {
        __mashedEmitBmadEvent: (name: string, payload: unknown) => boolean;
      }).__mashedEmitBmadEvent('bmad:node:awaiting_input', {
        execId,
        nodeId,
        inputId,
        prompt: 'Pick one',
        shape: shape ?? 'choice',
        round: 1,
        createdAt: Date.now(),
        promptId,
        required: true,
        structured: s,
      });
    },
    { structured, shape: opts.shape, execId: EXEC_ID, nodeId: NODE_ID, inputId: INPUT_ID, promptId: PROMPT_ID },
  );
}

async function openModal(page: Page): Promise<void> {
  await page.locator('[data-testid="snackbar-respond"]').click();
  await expect(page.locator('[data-testid="input-response-modal"]')).toBeVisible();
}

/**
 * Install the RespondToInput recording stub AFTER wails populates
 * `window.go.main.App`. The generated wrapper in `wailsjs/go/main/App.js`
 * resolves `window['go']['main']['App']['RespondToInput']` at call-time,
 * so a direct post-init assignment is captured by every subsequent call.
 */
async function installRespondStub(page: Page): Promise<void> {
  await page.waitForFunction(
    () =>
      !!(window as unknown as { go?: { main?: { App?: Record<string, unknown> } } })
        .go?.main?.App,
  );
  await page.evaluate(() => {
    const w = window as unknown as {
      __astRespondCalls: Array<[string, string, string, string]>;
      go: { main: { App: Record<string, unknown> } };
    };
    w.__astRespondCalls = [];
    w.go.main.App.RespondToInput = async (
      execId: string,
      nodeId: string,
      inputId: string,
      value: string,
    ) => {
      w.__astRespondCalls.push([execId, nodeId, inputId, value]);
      return undefined;
    };
  });
}

async function ensureHelpers(page: Page): Promise<void> {
  await page.goto(APP_URL);
  await page.waitForFunction(
    () =>
      typeof (window as unknown as { __mashed_gotoWorkflows?: unknown })
        .__mashed_gotoWorkflows === 'function',
  );
  await page.evaluate(() => {
    (window as unknown as { __mashed_gotoWorkflows: (repo?: string) => boolean })
      .__mashed_gotoWorkflows();
  });
  await page.waitForFunction(
    () =>
      typeof (window as unknown as { __mashedEmitBmadEvent?: unknown })
        .__mashedEmitBmadEvent === 'function',
  );
}

async function respondCalls(page: Page): Promise<Array<[string, string, string, string]>> {
  return page.evaluate(
    () => (window as unknown as { __astRespondCalls: Array<[string, string, string, string]> }).__astRespondCalls,
  );
}

// Polls until the recording stub has exactly `count` entries, then returns the
// array from the last successful read — one round-trip per poll iteration.
async function waitForRespond(
  page: Page,
  count: number,
): Promise<Array<[string, string, string, string]>> {
  let last: Array<[string, string, string, string]> = [];
  await expect
    .poll(async () => {
      last = await respondCalls(page);
      return last.length;
    })
    .toBe(count);
  return last;
}

/**
 * Select + commit a ChoiceWidget option. First click selects; a follow-up
 * Enter keypress on the selected option dispatches the `submit` event that
 * DecisionGroup writes into the per-modal responses store.
 */
async function pickChoice(scope: ReturnType<Page['locator']>, optionIndex: number): Promise<void> {
  const option = scope.locator(`[data-testid="choice-option-${optionIndex}"]`);
  await option.click();
  await option.press('Enter');
}

const dg = (
  key: string,
  widget: Record<string, unknown>,
  required = true,
  heading?: string,
) => ({
  type: 'decision_group',
  response_key: key,
  heading: heading ?? `DG ${key}`,
  required,
  widget,
});

test.describe('ui-ast-U7 decision_group AC suite', () => {
  test.skip(!E2E, 'set MASHED_E2E=1 with wails dev running to enable');

  test.beforeEach(async ({ page }) => {
    await ensureHelpers(page);
    await installRespondStub(page);
  });

  test('json-submit-full-map (AC-3) — Send ships JSON of all filled groups', async ({ page }) => {
    await emitAwaitingInput(page, {
      shape: 'json',
      structured: {
        version: '1',
        nodes: [
          dg('a', { type: 'choice', options: ['x', 'y'] }),
          dg('b', { type: 'choice', options: ['y', 'z'] }),
          dg('c', { type: 'choice', options: ['z', 'w'] }),
        ],
      },
    });
    await openModal(page);

    const groups = page.locator('[data-testid="decision-group"]');
    await expect(groups).toHaveCount(3);

    await pickChoice(groups.nth(0), 1);
    await pickChoice(groups.nth(1), 1);
    await pickChoice(groups.nth(2), 1);

    await page.locator('[data-testid="input-modal-send"]').click();

    const [calls] = await waitForRespond(page, 1);
    expect(calls[0]).toBe(EXEC_ID);
    expect(calls[1]).toBe(NODE_ID);
    expect(calls[2]).toBe(INPUT_ID);
    expect(calls[3]).toBe(JSON.stringify({ a: 'x', b: 'y', c: 'z' }));
  });

  test('single-group-plain-string-submit (AC-4) — one group on non-JSON shape submits plain string', async ({
    page,
  }) => {
    await emitAwaitingInput(page, {
      shape: 'free',
      structured: {
        version: '1',
        nodes: [dg('only', { type: 'choice', options: ['option-1', 'option-2'] })],
      },
    });
    await openModal(page);

    await expect(page.locator('[data-testid="collapse-banner"]')).toHaveCount(0);
    await pickChoice(page.locator('[data-testid="decision-group"]').first(), 1);
    await page.locator('[data-testid="input-modal-send"]').click();

    const [calls] = await waitForRespond(page, 1);
    expect(calls[3]).toBe('option-1');
  });

  test('collapse-banner-visible (AC-5) — non-JSON + >1 groups shows banner + single active widget', async ({
    page,
  }) => {
    await emitAwaitingInput(page, {
      shape: 'free',
      structured: {
        version: '1',
        nodes: [
          dg('a', { type: 'choice', options: ['x', 'y'] }),
          dg('b', { type: 'choice', options: ['y', 'z'] }, true),
          dg('c', { type: 'choice', options: ['z', 'w'] }, false),
        ],
      },
    });
    await openModal(page);

    const banner = page.locator('[data-testid="collapse-banner"]');
    await expect(banner).toBeVisible();
    await expect(banner).toContainText(/single answer/i);

    const enabled = page.locator('[data-testid="decision-group"]:not(.is-disabled)');
    const disabled = page.locator('[data-testid="decision-group"].is-disabled');
    await expect(enabled).toHaveCount(1);
    await expect(disabled).toHaveCount(2);
  });

  test('collapse-user-switches-active (AC-6) — click an inactive card to activate it', async ({
    page,
  }) => {
    await emitAwaitingInput(page, {
      shape: 'free',
      structured: {
        version: '1',
        nodes: [
          dg('a', { type: 'choice', options: ['x', 'y'] }, true),
          dg('b', { type: 'choice', options: ['y', 'z'] }, false),
          dg('c', { type: 'choice', options: ['z', 'w'] }, false),
        ],
      },
    });
    await openModal(page);

    await expect(page.locator('[data-testid="decision-group"][data-key="a"]'))
      .not.toHaveClass(/is-disabled/);
    await expect(page.locator('[data-testid="decision-group"][data-key="c"]'))
      .toHaveClass(/is-disabled/);

    await page.locator('[data-testid="decision-group"][data-key="c"]').click();
    await expect(page.locator('[data-testid="decision-group"][data-key="c"]'))
      .not.toHaveClass(/is-disabled/);
    await expect(page.locator('[data-testid="decision-group"][data-key="a"]'))
      .toHaveClass(/is-disabled/);

    // FIX #16 AC6 extension: Send must submit the ACTIVE group's value ('z'
    // from card 'c'), not the first-required 'a' that was the initial active.
    await pickChoice(page.locator('[data-testid="decision-group"][data-key="c"]'), 1);
    await page.locator('[data-testid="input-modal-send"]').click();

    const [calls] = await waitForRespond(page, 1);
    expect(calls[3]).toBe('z');
  });

  test('cmd-enter-submits (AC-8) — Meta+Enter fires RespondToInput with the same payload', async ({
    page,
  }) => {
    await emitAwaitingInput(page, {
      shape: 'json',
      structured: {
        version: '1',
        nodes: [dg('pick', { type: 'choice', options: ['x', 'y'] })],
      },
    });
    await openModal(page);

    await pickChoice(page.locator('[data-testid="decision-group"]').first(), 1);
    // Focus the modal card so keydown is dispatched inside it.
    await page.locator('[data-testid="input-response-modal"]').click();
    await page.keyboard.press('Meta+Enter');

    const [calls] = await waitForRespond(page, 1);
    expect(calls[3]).toBe(JSON.stringify({ pick: 'x' }));
  });

  test('zero-groups-fallback-layer1 (AC-9) — no decision_groups → Layer-1 widget visible', async ({
    page,
  }) => {
    await emitAwaitingInput(page, {
      shape: 'free',
      structured: {
        version: '1',
        nodes: [
          { type: 'markdown', content: 'Context above' },
          { type: 'hint', tone: 'info', content: 'no decision_groups here' },
        ],
      },
    });
    await openModal(page);

    await expect(page.locator('[data-testid="input-modal-send"]')).toHaveCount(0);
    await expect(page.locator('[data-testid="collapse-banner"]')).toHaveCount(0);
    await expect(page.locator('[data-testid="free-text-textarea"]')).toBeVisible();
  });
});
