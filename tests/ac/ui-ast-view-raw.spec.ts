/**
 * ui-ast-U8 — View raw + Diagnostics AC suite.
 *
 * Drives the Input modal's U8 surface end-to-end against a running Wails dev
 * server. Feeds an `awaiting_input` event whose `structured` payload carries
 * the diagnostics block, then toggles the `uiAdapterUntrustedExpanded` store
 * to exercise Q7 config paths.
 *
 * Covers AC-2 (toggle-opens-on-click), AC-8 (Q7 default collapsed), AC-9
 * (Q7 expand-when-flagged), and the AC-11 Playwright leg (raw content is
 * text-only; <script> renders as literal characters).
 *
 * Gated by `MASHED_E2E=1`. Run locally with:
 *
 *   MASHED_E2E=1 npx playwright test --config frontend/playwright.config.ts \
 *     tests/ac/ui-ast-view-raw.spec.ts
 */
import { test, expect, type Page } from '@playwright/test';

const E2E = process.env.MASHED_E2E === '1';
const APP_URL = process.env.MASHED_DEV_URL ?? 'http://localhost:34115';

const EXEC_ID = 'exec-u8';
const NODE_ID = 'n-u8';
const INPUT_ID = 'topic';
const PROMPT_ID = 'p-u8';

interface SeedOpts {
  diagnostics?: { untrusted?: boolean; fallback_reasons?: string[] };
  lastOutput?: string;
}

async function emitAwaitingInput(page: Page, opts: SeedOpts = {}): Promise<void> {
  const structured = JSON.stringify({
    version: '1',
    nodes: [{ type: 'markdown', content: 'Seeded context' }],
    ...(opts.diagnostics ? { diagnostics: opts.diagnostics } : {}),
  });
  await page.evaluate(
    ({ structured: s, lastOutput, execId, nodeId, inputId, promptId }) => {
      (
        window as unknown as {
          __mashedEmitBmadEvent: (name: string, payload: unknown) => boolean;
        }
      ).__mashedEmitBmadEvent('bmad:node:awaiting_input', {
        execId,
        nodeId,
        inputId,
        prompt: 'Pick one',
        shape: 'free',
        round: 1,
        createdAt: Date.now(),
        promptId,
        required: true,
        structured: s,
        lastOutput,
      });
    },
    {
      structured,
      lastOutput: opts.lastOutput ?? '',
      execId: EXEC_ID,
      nodeId: NODE_ID,
      inputId: INPUT_ID,
      promptId: PROMPT_ID,
    },
  );
}

async function setUntrustedExpanded(page: Page, value: boolean): Promise<void> {
  await page.evaluate((v) => {
    (
      window as unknown as {
        __mashed_setUntrustedExpandedForTests: (x: boolean) => void;
      }
    ).__mashed_setUntrustedExpandedForTests(v);
  }, value);
}

async function openModal(page: Page): Promise<void> {
  await page.locator('[data-testid="snackbar-respond"]').click();
  await expect(page.locator('[data-testid="input-response-modal"]')).toBeVisible();
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
  await page.waitForFunction(
    () =>
      typeof (
        window as unknown as { __mashed_setUntrustedExpandedForTests?: unknown }
      ).__mashed_setUntrustedExpandedForTests === 'function',
  );
}

test.describe('ui-ast-U8 view-raw AC suite', () => {
  test.skip(!E2E, 'set MASHED_E2E=1 with wails dev running to enable');

  test.beforeEach(async ({ page }) => {
    await ensureHelpers(page);
  });

  test('AC-2 toggle-opens-on-click — <pre> contains raw after clicking summary', async ({ page }) => {
    await setUntrustedExpanded(page, false);
    await emitAwaitingInput(page, {
      diagnostics: { untrusted: true },
      lastOutput: 'RAW CAPTURE',
    });
    await openModal(page);

    const toggle = page.locator('[data-testid="raw-view-toggle"]');
    await expect(toggle).toBeVisible();
    await expect(toggle).not.toHaveAttribute('open', /.*/);

    await toggle.locator('summary').click();
    await expect(toggle).toHaveAttribute('open', /.*/);
    await expect(toggle.locator('pre')).toContainText('RAW CAPTURE');
  });

  test('AC-8 q7-default-collapsed — untrusted + Q7=false keeps panel collapsed', async ({ page }) => {
    await setUntrustedExpanded(page, false);
    await emitAwaitingInput(page, {
      diagnostics: { untrusted: true },
      lastOutput: 'hidden body',
    });
    await openModal(page);

    const toggle = page.locator('[data-testid="raw-view-toggle"]');
    await expect(toggle).toBeVisible();
    await expect(toggle).not.toHaveAttribute('open', /.*/);
  });

  test('AC-9 q7-expand-when-flagged — untrusted + Q7=true opens on initial render', async ({ page }) => {
    await setUntrustedExpanded(page, true);
    await emitAwaitingInput(page, {
      diagnostics: { untrusted: true },
      lastOutput: 'visible body',
    });
    await openModal(page);

    const toggle = page.locator('[data-testid="raw-view-toggle"]');
    await expect(toggle).toBeVisible();
    await expect(toggle).toHaveAttribute('open', /.*/);
    await expect(toggle.locator('pre')).toContainText('visible body');
  });

  test('AC-11 raw content is text-only — <script> renders literally and does not execute', async ({ page }) => {
    await setUntrustedExpanded(page, true);
    await emitAwaitingInput(page, {
      diagnostics: { untrusted: true },
      lastOutput: '<script>window.__evil = 1;</script>',
    });
    await openModal(page);

    const toggle = page.locator('[data-testid="raw-view-toggle"]');
    await expect(toggle).toBeVisible();
    await expect(toggle.locator('pre')).toContainText('<script>window.__evil = 1;</script>');

    const evil = await page.evaluate(
      () => (window as unknown as { __evil?: number }).__evil,
    );
    expect(evil).toBeUndefined();
  });
});
