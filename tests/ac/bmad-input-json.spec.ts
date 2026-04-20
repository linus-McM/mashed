/**
 * S6 — bmad-interactive-06: InputResponseModal (json shape).
 * See bmad-input-free.spec.ts for the MASHED_E2E gate rationale.
 */
import { test, expect } from '@playwright/test';

const E2E = process.env.MASHED_E2E === '1';
const APP_URL = process.env.MASHED_DEV_URL ?? 'http://localhost:34115';

test.describe('bmad-input-json (S6)', () => {
  test.skip(!E2E, 'set MASHED_E2E=1 with wails dev running to enable');

  test('json shape: invalid JSON keeps submit disabled and shows parse error', async ({ page }) => {
    await page.goto(APP_URL);
    const hasHelper = await page.evaluate(() => typeof (window as any).__mashedEmitBmadEvent === 'function');
    test.skip(!hasHelper, '__mashedEmitBmadEvent helper not present.');

    await page.evaluate(() => {
      (window as any).__mashedEmitBmadEvent('bmad:node:awaiting_input', {
        execId: 'exec-e2e', nodeId: 'n1', inputId: 'payload',
        prompt: 'Send JSON', shape: 'json',
        round: 1, createdAt: Date.now(), promptId: 'p1', required: true,
      });
    });
    await page.locator('[data-testid="snackbar-respond"]').click();
    await page.locator('[data-testid="json-textarea"]').fill('{not-json}');
    await expect(page.locator('[data-testid="json-submit"]')).toBeDisabled();
  });
});
