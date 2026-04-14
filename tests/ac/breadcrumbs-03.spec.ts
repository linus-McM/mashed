/**
 * breadcrumbs-03 — Playwright smoke: every process/command node renders
 * at least one `.breadcrumb-row`.
 *
 * AC-1: CommandNode renders `.breadcrumb-row` when inputPath/outputPath set.
 * AC-3 regression guard: ProcessNode continues to render `.breadcrumb-row`
 *      when its process has inputs/outputs.
 *
 * Prerequisites
 * -------------
 *   1. `wails dev` MUST be running on http://localhost:34115 (or
 *      $MASHED_DEV_URL) — the spec does NOT spawn the dev server.
 *   2. The WorkflowBuilder dev-only seam `__mashed_seedCanvasNodes` must
 *      be present (added alongside this spec in breadcrumbs-03 GREEN).
 *
 * Assertions use `page.$$eval` so the per-node breadcrumb count is
 * computed atomically in one JS round-trip (robust against layout thrash).
 */
import { test, expect } from '@playwright/test';
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';

const APP_URL = process.env.MASHED_DEV_URL ?? 'http://localhost:34115';

const FIXTURE_PATH = resolve(
  __dirname,
  'fixtures/breadcrumbs-03-canvas-nodes.json',
);
const CANVAS_NODES = JSON.parse(readFileSync(FIXTURE_PATH, 'utf8'));

test.describe('breadcrumbs-03 — per-node breadcrumb row smoke', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(APP_URL, { waitUntil: 'domcontentloaded' });
    await page.waitForFunction(
      () =>
        typeof (
          window as unknown as { __mashed_gotoWorkflows?: unknown }
        ).__mashed_gotoWorkflows === 'function',
    );
    await page.evaluate(() =>
      (
        window as unknown as {
          __mashed_gotoWorkflows: (p?: string) => boolean;
        }
      ).__mashed_gotoWorkflows(''),
    );
    await page.waitForFunction(
      () =>
        typeof (
          window as unknown as { __mashed_seedCanvasNodes?: unknown }
        ).__mashed_seedCanvasNodes === 'function',
    );
  });

  test('AC-1 + AC-3: every .process-node and .command-node has >= 1 .breadcrumb-row', async ({
    page,
  }) => {
    const seeded = await page.evaluate((nodesArr) => {
      const w = window as unknown as {
        __mashed_seedCanvasNodes?: (n: unknown) => boolean;
      };
      return typeof w.__mashed_seedCanvasNodes === 'function'
        ? w.__mashed_seedCanvasNodes(nodesArr)
        : false;
    }, CANVAS_NODES);
    expect(seeded, 'seed seam must be registered').toBe(true);

    await expect(page.locator('.process-node').first()).toBeVisible();
    await expect(page.locator('.command-node').first()).toBeVisible();

    const counts = await page.$$eval(
      '.process-node, .command-node',
      (els) =>
        els.map((el) => ({
          nodeType: el.classList.contains('command-node')
            ? 'command'
            : 'process',
          breadcrumbRows: el.querySelectorAll('.breadcrumb-row').length,
        })),
    );

    expect(counts.length, 'fixture must render 2 nodes').toBeGreaterThanOrEqual(
      2,
    );
    expect(
      counts.some((c) => c.nodeType === 'process'),
      'at least one ProcessNode must be present',
    ).toBe(true);
    expect(
      counts.some((c) => c.nodeType === 'command'),
      'at least one CommandNode must be present',
    ).toBe(true);

    for (const c of counts) {
      expect(
        c.breadcrumbRows,
        `${c.nodeType} node must have >= 1 .breadcrumb-row`,
      ).toBeGreaterThanOrEqual(1);
    }
  });

  test('AC-3 regression: ProcessNode breadcrumb text mirrors configured path', async ({
    page,
  }) => {
    await page.evaluate((nodesArr) => {
      const w = window as unknown as {
        __mashed_seedCanvasNodes: (n: unknown) => boolean;
      };
      w.__mashed_seedCanvasNodes(nodesArr);
    }, CANVAS_NODES);

    await expect(page.locator('.process-node').first()).toBeVisible();

    const processCrumbs = await page.$$eval(
      '.process-node .breadcrumb-row',
      (els) => els.map((el) => el.textContent?.trim()).filter(Boolean),
    );

    expect(processCrumbs).toContain('.../brief.md');
    expect(processCrumbs).toContain('.../analysis.md');
  });

  // AC-2 (File Loader): when a File Loader process node is present
  // (processId === 'util-file-loader'), its `.breadcrumb-row` must carry
  // a `title` attribute — either the resolved path or the literal
  // 'unresolved' sentinel. The assertion is conditional: if the fixture
  // did not include a File Loader, the test skips rather than fails.
  test('AC-2: File Loader node breadcrumb carries a title attribute', async ({
    page,
  }) => {
    const fileLoaderNodes = CANVAS_NODES.filter(
      (n: { data?: { processId?: string } }) =>
        n.data?.processId === 'util-file-loader',
    );

    test.skip(
      fileLoaderNodes.length === 0,
      'fixture has no util-file-loader node — skipped',
    );

    await page.evaluate((nodesArr) => {
      const w = window as unknown as {
        __mashed_seedCanvasNodes: (n: unknown) => boolean;
      };
      w.__mashed_seedCanvasNodes(nodesArr);
    }, CANVAS_NODES);

    await expect(page.locator('.process-node').first()).toBeVisible();

    const titles = await page.$$eval(
      '.process-node .breadcrumb-row',
      (els) => els.map((el) => el.getAttribute('title')),
    );

    expect(titles.length).toBeGreaterThan(0);
    for (const title of titles) {
      expect(
        title,
        'File Loader breadcrumb must have title (path or "unresolved")',
      ).toBeTruthy();
    }
  });
});
