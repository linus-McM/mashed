/**
 * ui-ast-U6 — UIAST rendering AC suite.
 *
 * Exercises the Input modal's AST region end-to-end against a running
 * Wails dev server. The dev harness `window.__mashedEmitBmadEvent` (set up in
 * WorkflowBuilder.svelte under `import.meta.env.DEV`) drives an
 * `awaiting_input` event carrying a `structured` UIAST v1 JSON blob; the
 * modal then opens and renders the AST above the Layer-1 widget.
 *
 * Gated by `MASHED_E2E=1` so CI runs without the dev server skip cleanly.
 * Run locally with:
 *
 *   MASHED_E2E=1 npx playwright test --config frontend/playwright.config.ts \
 *     tests/ac/ui-ast-rendering.spec.ts
 */
import { test, expect, type Page } from '@playwright/test';

const E2E = process.env.MASHED_E2E === '1';
const APP_URL = process.env.MASHED_DEV_URL ?? 'http://localhost:34115';

interface StructuredPromptOpts {
  structured?: object;
  shape?: string;
  options?: string[];
  nodeId?: string;
  inputId?: string;
}

async function emitAwaitingInput(page: Page, opts: StructuredPromptOpts = {}): Promise<void> {
  const structured = opts.structured !== undefined ? JSON.stringify(opts.structured) : undefined;
  await page.evaluate(
    ({ structured: s, shape, options, nodeId, inputId }) => {
      (window as unknown as {
        __mashedEmitBmadEvent: (name: string, payload: unknown) => boolean;
      }).__mashedEmitBmadEvent('bmad:node:awaiting_input', {
        execId: 'exec-ac',
        nodeId: nodeId ?? 'n-ast',
        inputId: inputId ?? 'topic',
        prompt: 'Pick one',
        shape: shape ?? 'choice',
        options: options ?? ['a', 'b', 'c'],
        round: 1,
        createdAt: Date.now(),
        promptId: 'p-ast',
        required: true,
        structured: s,
      });
    },
    { structured, shape: opts.shape, options: opts.options, nodeId: opts.nodeId, inputId: opts.inputId },
  );
}

async function openModal(page: Page): Promise<void> {
  await page.locator('[data-testid="snackbar-respond"]').click();
  await expect(page.locator('[data-testid="input-response-modal"]')).toBeVisible();
}

async function ensureHelpers(page: Page): Promise<void> {
  await page.goto(APP_URL);
  // Navigate into WorkflowBuilder so it mounts __mashedEmitBmadEvent.
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

test.describe('ui-ast-U6 rendering AC suite', () => {
  test.skip(!E2E, 'set MASHED_E2E=1 with wails dev running to enable');

  test('six-node-shapes (AC-5) — all passive types render via AstNode dispatcher', async ({ page }) => {
    await ensureHelpers(page);
    await emitAwaitingInput(page, {
      structured: {
        version: '1',
        nodes: [
          { type: 'markdown', content: 'A markdown paragraph' },
          { type: 'hint', tone: 'info', content: 'An info hint' },
          { type: 'summary', heading: 'Summary', bullets: ['point a', 'point b'] },
          { type: 'code', lang: 'sh', content: 'echo ok', copyable: true },
          {
            type: 'table',
            heading: 'Cmp',
            columns: ['col1', 'col2'],
            rows: [['r1c1', 'r1c2']],
          },
          { type: 'unknown_future_type', content: 'degrades gracefully' },
        ],
      },
    });
    await openModal(page);

    const region = page.locator('[data-testid="ast-region"]');
    await expect(region).toBeVisible();
    await expect(region.locator('[data-testid="markdown-block"]').first()).toBeVisible();
    await expect(region.locator('[data-testid="hint-banner"]')).toBeVisible();
    await expect(region.locator('[data-testid="summary-card"]')).toBeVisible();
    await expect(region.locator('[data-testid="code-block"]')).toBeVisible();
    await expect(region.locator('[data-testid="comparison-table"]')).toBeVisible();
    // Unknown type falls back to MarkdownBlock — at least 2 markdown blocks now
    // (the fallback + the explicit markdown node).
    await expect(region.locator('[data-testid="markdown-block"]')).toHaveCount(2);
  });

  test('javascript-href-stripped-dom (AC-6) — no javascript: anchors in rendered markdown', async ({ page }) => {
    await ensureHelpers(page);
    await emitAwaitingInput(page, {
      structured: {
        version: '1',
        nodes: [{ type: 'markdown', content: '[click me](javascript:alert(1))' }],
      },
    });
    await openModal(page);
    const region = page.locator('[data-testid="ast-region"]');
    await expect(region).toBeVisible();
    // Text survives, href is stripped.
    await expect(region).toContainText('click me');
    await expect(page.locator('[data-testid="ast-region"] a[href^="javascript:" i]')).toHaveCount(0);
  });

  test('data-and-file-schemes-rejected-dom (AC-7) — data: and file: hrefs absent', async ({ page }) => {
    await ensureHelpers(page);
    await emitAwaitingInput(page, {
      structured: {
        version: '1',
        nodes: [
          {
            type: 'markdown',
            content: '[x](data:text/html,<x>) and [y](file:///etc/passwd)',
          },
        ],
      },
    });
    await openModal(page);
    const region = page.locator('[data-testid="ast-region"]');
    await expect(region).toBeVisible();
    await expect(region).toContainText('x');
    await expect(region).toContainText('y');
    await expect(page.locator('[data-testid="ast-region"] a[href^="data:" i]')).toHaveCount(0);
    await expect(page.locator('[data-testid="ast-region"] a[href^="file:" i]')).toHaveCount(0);
  });

  test('https-link-confirm-dialog (AC-8) — confirm gates BrowserOpenURL', async ({ page }) => {
    // Stub confirm + BrowserOpenURL before any app code runs. Init script is
    // re-run on every navigation, so the stub survives page.goto inside
    // ensureHelpers.
    await page.addInitScript(() => {
      const w = window as unknown as {
        __astConfirmCalls: string[];
        __astOpenedUrls: string[];
        __astConfirmResult: boolean;
        confirm: (message?: string) => boolean;
        runtime?: Record<string, unknown>;
      };
      w.__astConfirmCalls = [];
      w.__astOpenedUrls = [];
      w.__astConfirmResult = true;
      w.confirm = (message?: string) => {
        w.__astConfirmCalls.push(String(message ?? ''));
        return w.__astConfirmResult;
      };
      // Wails assigns window.runtime late — intercept via setter so our stub
      // survives any subsequent assignment.
      let _runtime: Record<string, unknown> | undefined;
      const stubBrowserOpenURL = (url: string) => {
        w.__astOpenedUrls.push(url);
      };
      const install = (obj: Record<string, unknown> | undefined) => {
        if (obj && typeof obj === 'object') {
          obj.BrowserOpenURL = stubBrowserOpenURL;
        }
      };
      Object.defineProperty(window, 'runtime', {
        configurable: true,
        get() {
          return _runtime;
        },
        set(v) {
          _runtime = v;
          install(_runtime);
        },
      });
      // If runtime already exists (unlikely given init-script ordering), patch now.
      install((window as unknown as { runtime?: Record<string, unknown> }).runtime);
    });
    await ensureHelpers(page);

    await emitAwaitingInput(page, {
      structured: {
        version: '1',
        nodes: [{ type: 'markdown', content: '[docs](https://example.com)' }],
      },
    });
    await openModal(page);

    // Confirm=true path → BrowserOpenURL called.
    await page.locator('[data-testid="ast-region"] a').first().click();
    const confirmCalls = await page.evaluate(
      () => (window as unknown as { __astConfirmCalls: string[] }).__astConfirmCalls,
    );
    expect(confirmCalls.length).toBe(1);
    expect(confirmCalls[0]).toContain('https://example.com');
    expect(confirmCalls[0]).toMatch(/^Open https:\/\/example\.com\/?\?$/);
    const opened = await page.evaluate(
      () => (window as unknown as { __astOpenedUrls: string[] }).__astOpenedUrls,
    );
    expect(opened.length).toBe(1);
    expect(opened[0]).toContain('https://example.com');

    // Flip to confirm=false and click again → no new BrowserOpenURL call.
    await page.evaluate(() => {
      (window as unknown as { __astConfirmResult: boolean }).__astConfirmResult = false;
    });
    await page.locator('[data-testid="ast-region"] a').first().click();
    const openedAfter = await page.evaluate(
      () => (window as unknown as { __astOpenedUrls: string[] }).__astOpenedUrls,
    );
    expect(openedAfter.length).toBe(1); // unchanged
  });

  test('modal-renders-ast-above-widget (AC-9) — ast-region precedes Layer-1 widget', async ({ page }) => {
    await ensureHelpers(page);
    await emitAwaitingInput(page, {
      shape: 'choice',
      options: ['a', 'b', 'c'],
      structured: {
        version: '1',
        nodes: [{ type: 'markdown', content: 'Context above the widget' }],
      },
    });
    await openModal(page);
    const region = page.locator('[data-testid="ast-region"]');
    const widget = page.locator('[role="radiogroup"]');
    await expect(region).toBeVisible();
    await expect(widget).toBeVisible();
    const rel = await region.evaluate((regionEl, widgetEl) => {
      const mask = regionEl.compareDocumentPosition(widgetEl as Node);
      return Boolean(mask & Node.DOCUMENT_POSITION_FOLLOWING);
    }, await widget.elementHandle());
    expect(rel).toBe(true);
  });

  test('modal-layer1-unchanged-when-null (AC-9) — no ast-region when structured missing', async ({ page }) => {
    await ensureHelpers(page);
    await emitAwaitingInput(page, {
      shape: 'choice',
      options: ['a', 'b', 'c'],
      // no structured → $pendingAst === null
    });
    await openModal(page);
    await expect(page.locator('[data-testid="ast-region"]')).toHaveCount(0);
    await expect(page.locator('[role="radiogroup"]')).toBeVisible();
  });

  test('code-copy-button-works (AC-11) — clipboard receives content, button flips to "Copied"', async ({ page, context }) => {
    await context.grantPermissions(['clipboard-read', 'clipboard-write']);
    await ensureHelpers(page);
    await emitAwaitingInput(page, {
      structured: {
        version: '1',
        nodes: [{ type: 'code', lang: 'sh', content: 'rm -rf ~', copyable: true }],
      },
    });
    await openModal(page);
    const copyBtn = page.locator('[data-testid="code-copy-button"]');
    await expect(copyBtn).toBeVisible();
    await copyBtn.click();
    await expect(copyBtn).toContainText('Copied');
    const clipboardText = await page.evaluate(() => navigator.clipboard.readText());
    expect(clipboardText).toBe('rm -rf ~');
  });
});
