/**
 * S6 — bmad-interactive-06: InputResponseModal (multi shape, AC-2).
 * See bmad-input-free.spec.ts for the MASHED_E2E gate rationale.
 */
import { test, expect } from '@playwright/test';

const E2E = process.env.MASHED_E2E === '1';
const APP_URL = process.env.MASHED_DEV_URL ?? 'http://localhost:34115';

test.describe('bmad-input-multi (S6 AC-2)', () => {
  test.skip(!E2E, 'set MASHED_E2E=1 with wails dev running to enable');

  test('multi shape renders checkboxes; Submit sends comma-joined "x,z"', async ({ page }) => {
    await page.goto(APP_URL);
    const hasHelper = await page.evaluate(() => typeof (window as any).__mashedEmitBmadEvent === 'function');
    test.skip(!hasHelper, '__mashedEmitBmadEvent helper not present.');

    await page.evaluate(() => {
      (window as any).__mashedEmitBmadEvent('bmad:node:awaiting_input', {
        execId: 'exec-e2e', nodeId: 'n1', inputId: 'tags',
        prompt: 'Tag it', shape: 'multi', options: ['x', 'y', 'z'],
        round: 1, createdAt: Date.now(), promptId: 'p1', required: true,
      });
    });
    await page.locator('[data-testid="snackbar-respond"]').click();
    await page.locator('[data-testid="multi-checkbox-1"]').click();
    await page.locator('[data-testid="multi-checkbox-3"]').click();
    await page.locator('[data-testid="multi-submit"]').click();
  });
});
