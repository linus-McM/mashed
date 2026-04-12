// Minimal Playwright config for `tests/ac/*.spec.ts`.
//
// Located in frontend/ so it resolves @playwright/test from
// frontend/node_modules without a root package.json. Successive AC stories
// may extend this with projects, retries, or richer reporters.
//
// Prerequisite: `wails dev` must be running on http://localhost:34115 (or
// $MASHED_DEV_URL). The config does NOT spawn the dev server — that's an
// out-of-band action so test runs don't fight a live editor session.
//
// Run from the repo root with:
//   npx playwright test --config frontend/playwright.config.ts
import { defineConfig } from '@playwright/test';

export default defineConfig({
  testDir: '../tests/ac',
  testMatch: /.*\.spec\.ts$/,
  fullyParallel: false,
  retries: 0,
  reporter: [['list']],
  use: {
    baseURL: process.env.MASHED_DEV_URL ?? 'http://localhost:34115',
    trace: 'retain-on-failure',
  },
});
