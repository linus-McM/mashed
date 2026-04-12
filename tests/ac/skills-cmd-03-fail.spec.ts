/**
 * skills-cmd-03 AC-3 (frontend slice) — Running a workflow with a command
 * node surfaces the Phase 2 sentinel toast and flips the node to `failed`.
 *
 * Prerequisites
 * -------------
 *   1. `wails dev` MUST be running on http://localhost:34115 (or
 *      $MASHED_DEV_URL). The script does NOT spawn the dev server — the
 *      engineer or the sprint orchestrator is responsible for that.
 *   2. The dev-only hooks `__mashed_loadWorkflowFixture` and
 *      `__mashed_simulateNodeFailure` must be present on window (they are,
 *      per the `import.meta.env.DEV` block in WorkflowBuilder.svelte).
 *
 * Why simulate instead of clicking Run end-to-end?
 * -----------------------------------------------
 * AC-3 is split across two agents: go-engineer owns the backend test that
 * drives StartBmadWorkflow -> failNode for a command node. This frontend
 * spec covers the UI-side wiring: canvas renders the failed badge and the
 * failure toast. Using a real backend execution would require a valid
 * repoPath on disk, a saved workflow, and the tmux binary — far more
 * brittle than the narrowly-scoped frontend contract we actually want to
 * lock. The simulate hook exercises the EXACT code paths a real backend
 * emission would take (status listener state update + sentinel detection).
 */
import { test, expect } from '@playwright/test';
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';

const APP_URL = process.env.MASHED_DEV_URL ?? 'http://localhost:34115';

const FIXTURE_PATH = resolve(
  __dirname,
  'fixtures/skills-cmd-03-command-node-fail.json',
);
const FIXTURE = JSON.parse(readFileSync(FIXTURE_PATH, 'utf8'));
const COMMAND_NODE_ID = FIXTURE.nodes[0].id;
const COMMAND_NAME = FIXTURE.nodes[0].config.commandName;

test.describe('skills-cmd-03 AC-3 — Command node fails cleanly with toast', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(APP_URL, { waitUntil: 'domcontentloaded' });
    await page.waitForFunction(
      () =>
        typeof (window as unknown as { __mashed_gotoWorkflows?: unknown })
          .__mashed_gotoWorkflows === 'function',
    );
    await page.evaluate(() =>
      (window as unknown as { __mashed_gotoWorkflows: (p?: string) => boolean })
        .__mashed_gotoWorkflows(''),
    );
    await page.waitForFunction(
      () =>
        typeof (window as unknown as { __mashed_loadWorkflowFixture?: unknown })
          .__mashed_loadWorkflowFixture === 'function',
    );
  });

  test(`AC-3: command node transitions to failed and toast surfaces "command nodes not yet runnable"`, async ({
    page,
  }) => {
    // Step 1 — drop a one-command-node workflow onto the canvas via the
    // dev-only fixture hook. Mirrors skills-cmd-02's AC-3 test seam.
    const loaded = await page.evaluate((wf) => {
      const w = window as unknown as {
        __mashed_loadWorkflowFixture?: (def: unknown) => boolean;
      };
      return typeof w.__mashed_loadWorkflowFixture === 'function'
        ? w.__mashed_loadWorkflowFixture(wf)
        : false;
    }, FIXTURE);
    expect(loaded).toBe(true);

    // Confirm the command node rendered BEFORE we simulate the failure —
    // otherwise a hook-less regression would look like a pass.
    const commandNode = page.locator('[data-node-type="command"]').first();
    await expect(commandNode).toBeVisible();
    await expect(commandNode).toContainText(COMMAND_NAME);

    // Step 2 — wait for the simulate-failure hook (it is installed in
    // the same DEV block as loadWorkflowFixture but arrives a tick later
    // on slow mounts).
    await page.waitForFunction(
      () =>
        typeof (
          window as unknown as { __mashed_simulateNodeFailure?: unknown }
        ).__mashed_simulateNodeFailure === 'function',
    );

    // Step 3 — drive the same code path the `bmad:node:status` listener
    // runs when the backend emits NodeFailed for a command node. This is
    // the hand-off point where go-engineer's backend slice meets ours.
    const triggered = await page.evaluate((nodeId) => {
      const w = window as unknown as {
        __mashed_simulateNodeFailure: (
          id: string,
          msg?: string,
        ) => boolean;
      };
      return w.__mashed_simulateNodeFailure(nodeId);
    }, COMMAND_NODE_ID);
    expect(triggered).toBe(true);

    // AC-3a — the failure toast appears within the 5s budget the story spec
    // calls out.
    const toast = page.getByTestId('canvas-failure-toast');
    await expect(toast).toBeVisible({ timeout: 5000 });
    await expect(toast).toContainText(/command nodes not yet runnable/i);

    // AC-3b — the command node itself shows the failed badge (shared
    // `.failed-text` class from CommandNode.svelte / ProcessNode.svelte).
    const failedBadge = commandNode.locator('.failed-text');
    await expect(failedBadge).toBeVisible();

    // AC-3c — the toast auto-dismisses after 3000ms without blocking the
    // canvas. Wait a hair over the timeout to avoid flakiness.
    await expect(toast).toBeHidden({ timeout: 4000 });
  });
});
