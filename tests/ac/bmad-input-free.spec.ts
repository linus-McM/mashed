/**
 * S6 — bmad-interactive-06: InputResponseModal (free shape)
 *
 * These Playwright specs require `wails dev` running on
 * http://localhost:34115 and an active workflow execution that emits a
 * `bmad:node:awaiting_input` event. Gated behind `MASHED_E2E=1` so CI
 * does not block on a missing dev server. When running locally:
 *
 *   MASHED_E2E=1 npx playwright test --config frontend/playwright.config.ts \
 *     tests/ac/bmad-input-free.spec.ts
 *
 * The spec uses the window-scoped test helper injected by the frontend in
 * DEV mode to dispatch an awaiting_input event:
 *
 *   window.__mashedEmitBmadEvent('bmad:node:awaiting_input', {...})
 *
 * If that helper is absent the test skips with an explanatory message.
 */
import { test, expect } from '@playwright/test';

const E2E = process.env.MASHED_E2E === '1';
const APP_URL = process.env.MASHED_DEV_URL ?? 'http://localhost:34115';

test.describe('bmad-input-free (S6 AC-1, AC-2)', () => {
  test.skip(!E2E, 'set MASHED_E2E=1 with wails dev running to enable');

  test('awaiting_input for shape=free opens modal with textarea and submits value', async ({ page }) => {
    await page.goto(APP_URL);

    const hasHelper = await page.evaluate(() => typeof (window as any).__mashedEmitBmadEvent === 'function');
    test.skip(!hasHelper, '__mashedEmitBmadEvent helper not present — add a DEV-only event emitter to enable this spec.');

    await page.evaluate(() => {
      (window as any).__mashedEmitBmadEvent('bmad:node:awaiting_input', {
        execId: 'exec-e2e',
        nodeId: 'n1',
        inputId: 'topic',
        prompt: 'What topic?',
        shape: 'free',
        round: 1,
        createdAt: Date.now(),
        promptId: 'p1',
        required: true,
      });
    });

    await page.locator('[data-testid="snackbar-respond"]').click();
    await expect(page.locator('[data-testid="input-response-modal"]')).toBeVisible();
    await page.locator('[data-testid="free-text-textarea"]').fill('ai assistants');
    await page.locator('[data-testid="free-text-submit"]').click();
    // Backend RespondToInput would normally emit input_resolved to close.
  });
});
