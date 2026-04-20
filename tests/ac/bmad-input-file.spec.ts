/**
 * S6 — bmad-interactive-06: InputResponseModal (file shape).
 * See bmad-input-free.spec.ts for the MASHED_E2E gate rationale.
 */
import { test, expect } from '@playwright/test';

const E2E = process.env.MASHED_E2E === '1';
const APP_URL = process.env.MASHED_DEV_URL ?? 'http://localhost:34115';

test.describe('bmad-input-file (S6)', () => {
  test.skip(!E2E, 'set MASHED_E2E=1 with wails dev running to enable');

  test('file shape: invalid path triggers inline error + shake via input_invalid', async ({ page }) => {
    await page.goto(APP_URL);
    const hasHelper = await page.evaluate(() => typeof (window as any).__mashedEmitBmadEvent === 'function');
    test.skip(!hasHelper, '__mashedEmitBmadEvent helper not present.');

    await page.evaluate(() => {
      (window as any).__mashedEmitBmadEvent('bmad:node:awaiting_input', {
        execId: 'exec-e2e', nodeId: 'n1', inputId: 'path',
        prompt: 'Pick a file', shape: 'file',
        round: 1, createdAt: Date.now(), promptId: 'p1', required: true,
      });
    });
    await page.locator('[data-testid="snackbar-respond"]').click();
    await page.locator('[data-testid="file-path-input"]').fill('/etc/passwd');
    await page.locator('[data-testid="file-submit"]').click();

    await page.evaluate(() => {
      (window as any).__mashedEmitBmadEvent('bmad:node:input_invalid', {
        execId: 'exec-e2e', nodeId: 'n1', inputId: 'path',
        reason: 'path outside repository root',
      });
    });
    await expect(page.locator('[data-testid="input-modal-error"]')).toContainText('path outside repository root');
  });
});
