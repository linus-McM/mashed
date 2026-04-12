/**
 * skills-cmd-02 — Frontend CommandNode + drop handler.
 *
 * RED-phase Playwright spec covering AC-1 (drop creates command node) and
 * AC-3 (saved fixture renders via CommandNode.svelte).
 *
 * Prerequisites
 * -------------
 *   1. `wails dev` MUST be running on http://localhost:34115 (or
 *      $MASHED_DEV_URL) before invoking this spec — the script does NOT
 *      spawn the dev server.
 *   2. `@playwright/test` and a minimal `playwright.config.*` must exist.
 *      The engineer in GREEN is responsible for ensuring both are present
 *      if they are not already wired up at the repo root.
 *
 * RED expectation
 * ---------------
 * Both tests should fail at the assertion step (selector not found / class
 * missing) — NOT at module-import time. The failure proves CommandNode and
 * the drop-handler branch are missing, which is exactly what GREEN delivers.
 */
import { test, expect } from '@playwright/test';
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';

const APP_URL = process.env.MASHED_DEV_URL ?? 'http://localhost:34115';

const FIXTURE_PATH = resolve(
  __dirname,
  'fixtures/skills-cmd-02-command-node.json',
);
const FIXTURE = JSON.parse(readFileSync(FIXTURE_PATH, 'utf8'));
const COMMAND_NAME = FIXTURE.nodes[0].config.commandName;

test.describe('skills-cmd-02 — CommandNode drop + render (RED)', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(APP_URL, { waitUntil: 'domcontentloaded' });
    // GREEN-phase navigation: jump straight to the WorkflowBuilder view
    // via the dev-only hook exposed in App.svelte. This avoids having
    // to click through the feed/repo-picker flow and keeps the spec
    // deterministic across fresh and primed environments.
    await page.waitForFunction(() => typeof (window as unknown as { __mashed_gotoWorkflows?: unknown }).__mashed_gotoWorkflows === 'function');
    await page.evaluate(() => (window as unknown as { __mashed_gotoWorkflows: (p?: string) => boolean }).__mashed_gotoWorkflows(''));
    // Wait for WorkflowBuilder to mount its own dev hooks before the
    // individual tests reach for them.
    await page.waitForFunction(() => typeof (window as unknown as { __mashed_seedMashedAssets?: unknown }).__mashed_seedMashedAssets === 'function');
  });

  test(`AC-1: dragging "${COMMAND_NAME}" from the sidebar onto the canvas at (300, 200) renders a command node with the visible label "${COMMAND_NAME}"`, async ({
    page,
  }) => {
    // Seed the sidebar with a mashed-ready command via the dev hook —
    // the fresh test env has no real `.claude/commands/simplify.md` on
    // disk, so we inject one directly into the store. This still
    // exercises the real drag/drop pipeline end-to-end (sidebar row,
    // dragstart, onDrop, handleMashedAssetDrop, $nodes append).
    await page.evaluate((name) => {
      const seed = (window as unknown as {
        __mashed_seedMashedAssets: (g: unknown) => boolean;
      }).__mashed_seedMashedAssets;
      seed({
        localCommands: [
          {
            name,
            path: `/tmp/${name}.md`,
            kind: 'command',
            source: 'local',
            role: 'command',
            description: 'Review recent changes',
          },
        ],
        globalCommands: [],
        localSkills: [],
        globalSkills: [],
      });
    }, COMMAND_NAME);

    // Switch to the "Skills" sidebar tab — that's where mashed-ready
    // commands and skills are grouped. The localCommands accordion is
    // open by default so the row renders immediately after the tab
    // click.
    await page.getByRole('button', { name: 'Skills', exact: true }).click();

    const sidebarRow = page.locator(
      `[data-testid="mashed-asset-${COMMAND_NAME}"]`,
    );
    await expect(sidebarRow).toBeVisible();

    const canvas = page.locator('.svelte-flow__pane').first();
    await expect(canvas).toBeVisible();

    await sidebarRow.dragTo(canvas, { targetPosition: { x: 300, y: 200 } });

    const commandNode = page.locator('[data-node-type="command"]').first();
    await expect(commandNode).toBeVisible();
    await expect(commandNode).toContainText(COMMAND_NAME);
  });

  test('AC-3: a saved workflow containing one command node renders via CommandNode.svelte (`.command-node` class) with label matching data.config.commandName', async ({
    page,
  }) => {
    // GREEN may wire AC-3 in either of two ways:
    //   (a) expose a dev-only `window.__mashed_loadWorkflowFixture(def)`
    //       guarded by `import.meta.env.DEV`, or
    //   (b) use the existing Wails binding `SaveBmadWorkflow` via
    //       `window.go.main.App.SaveBmadWorkflow(JSON.stringify(def))` and
    //       drive the existing "Open Workflow" UI.
    // The RED spec exercises (a). If GREEN picks (b), update both branches
    // here in the same commit.
    const loaded = await page.evaluate((wf) => {
      const w = window as unknown as {
        __mashed_loadWorkflowFixture?: (def: unknown) => boolean;
      };
      return typeof w.__mashed_loadWorkflowFixture === 'function'
        ? w.__mashed_loadWorkflowFixture(wf)
        : false;
    }, FIXTURE);

    expect(loaded).toBe(true);

    const node = page.locator('.command-node').first();
    await expect(node).toBeVisible();
    await expect(node).toContainText(COMMAND_NAME);
  });
});
