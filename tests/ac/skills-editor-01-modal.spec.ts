/**
 * skills-editor-01 — Skill editor modal shell + sidebar right-click.
 *
 * Playwright spec covering:
 *   AC-1: Right-click opens the modal with asset pre-loaded
 *   AC-2: Escape and backdrop close the modal
 *   AC-3: Modal traps focus while open
 *
 * Prerequisites
 * -------------
 *   1. `wails dev` MUST be running on http://localhost:34115 (or
 *      $MASHED_DEV_URL) before invoking this spec.
 *   2. The WorkflowBuilder view must be accessible and have at least
 *      one mashed-asset row in the Skills tab.
 */
import { test, expect } from '@playwright/test';

const APP_URL = process.env.MASHED_DEV_URL ?? 'http://localhost:34115';

test.describe('skills-editor-01 — SkillEditorModal (AC-1, AC-2, AC-3)', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(APP_URL, { waitUntil: 'domcontentloaded' });
    // Navigate to a repo context so the sidebar populates
    await page.waitForTimeout(1500);
  });

  test('AC-1: right-click on mashed-asset row opens modal with fields pre-populated', async ({ page }) => {
    // Switch to Skills tab
    const skillsTab = page.locator('button.tab', { hasText: 'Skills' });
    await skillsTab.click();
    await page.waitForTimeout(500);

    // Find a mashed-asset row (any)
    const assetRow = page.locator('[data-testid^="mashed-asset-"]').first();
    await expect(assetRow).toBeVisible({ timeout: 5000 });

    // Right-click to open the modal
    await assetRow.click({ button: 'right' });

    // Modal should appear
    const modal = page.locator('[role="dialog"][aria-labelledby="sem-title"]');
    await expect(modal).toBeVisible({ timeout: 3000 });

    // Name field should be pre-populated (non-empty)
    const nameInput = modal.locator('#sem-name');
    await expect(nameInput).toHaveValue(/.+/);

    // File path subtitle should be visible
    const filePath = modal.locator('.file-path');
    await expect(filePath).toBeVisible();
    await expect(filePath).not.toBeEmpty();
  });

  test('AC-2a: Escape closes the modal', async ({ page }) => {
    // Open the modal via right-click
    const skillsTab = page.locator('button.tab', { hasText: 'Skills' });
    await skillsTab.click();
    await page.waitForTimeout(500);

    const assetRow = page.locator('[data-testid^="mashed-asset-"]').first();
    await expect(assetRow).toBeVisible({ timeout: 5000 });
    await assetRow.click({ button: 'right' });

    const modal = page.locator('[role="dialog"][aria-labelledby="sem-title"]');
    await expect(modal).toBeVisible({ timeout: 3000 });

    // Press Escape
    await page.keyboard.press('Escape');

    // Modal should be gone
    await expect(modal).not.toBeVisible({ timeout: 2000 });
  });

  test('AC-2b: clicking backdrop closes the modal', async ({ page }) => {
    // Open the modal
    const skillsTab = page.locator('button.tab', { hasText: 'Skills' });
    await skillsTab.click();
    await page.waitForTimeout(500);

    const assetRow = page.locator('[data-testid^="mashed-asset-"]').first();
    await expect(assetRow).toBeVisible({ timeout: 5000 });
    await assetRow.click({ button: 'right' });

    const modal = page.locator('[role="dialog"][aria-labelledby="sem-title"]');
    await expect(modal).toBeVisible({ timeout: 3000 });

    // Click the overlay (backdrop) — click at the edge, outside the card
    const overlay = page.locator('.overlay');
    await overlay.click({ position: { x: 10, y: 10 } });

    // Modal should be gone
    await expect(modal).not.toBeVisible({ timeout: 2000 });
  });

  test('AC-3: Tab cycles focus within the modal (focus trap)', async ({ page }) => {
    // Open the modal
    const skillsTab = page.locator('button.tab', { hasText: 'Skills' });
    await skillsTab.click();
    await page.waitForTimeout(500);

    const assetRow = page.locator('[data-testid^="mashed-asset-"]').first();
    await expect(assetRow).toBeVisible({ timeout: 5000 });
    await assetRow.click({ button: 'right' });

    const modal = page.locator('[role="dialog"][aria-labelledby="sem-title"]');
    await expect(modal).toBeVisible({ timeout: 3000 });

    // Name input should have initial focus
    const nameInput = modal.locator('#sem-name');
    await expect(nameInput).toBeFocused();

    // Count all focusable elements inside the modal
    const focusableCount = await modal.locator(
      'input, textarea, select, button'
    ).count();

    // Tab through all elements — after cycling through all, focus should
    // remain inside the modal (not escape to the backdrop or body)
    for (let i = 0; i < focusableCount + 1; i++) {
      await page.keyboard.press('Tab');
    }

    // After a full cycle + 1, focus should be back inside the modal
    const focused = page.locator(':focus');
    const isInsideModal = await focused.evaluate((el) => {
      return !!el.closest('[role="dialog"]');
    });
    expect(isInsideModal).toBe(true);
  });
});
