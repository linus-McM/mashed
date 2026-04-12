# Story 3: Frontend Testing Infrastructure with Vitest

**Priority:** P0-critical
**Domain:** frontend
**Estimated Complexity:** M
**Depends On:** none
**Status:** done

## Description

Set up Vitest as the frontend test framework, configure it for Svelte 4 component testing with jsdom, and migrate the existing `themeConverter.test.js` from `node:test`/`node:assert` to Vitest's `describe/it/expect` API. This establishes the testing foundation that the lefthook `frontend-test` hook (Story 4) depends on and enables future frontend test coverage.

## Developer Notes

### Architecture
- **New file:** `frontend/vitest.config.js` -- Vitest configuration with Svelte plugin and jsdom environment
- **Modified files:**
  - `frontend/package.json` -- add devDependencies (`vitest`, `@testing-library/svelte`, `jsdom`), add `test` and `test:watch` scripts
  - `frontend/src/lib/themeConverter.test.js` -- migrate from `node:test` + `node:assert/strict` to `vitest` `describe/it/expect`
- **Data flow:** `npx vitest run` discovers `*.test.js` files, runs them with jsdom environment, reports results

### Technical Considerations
- **Vitest config structure:**
  ```js
  import { defineConfig } from 'vitest/config';
  import { svelte } from '@sveltejs/vite-plugin-svelte';

  export default defineConfig({
    plugins: [svelte({ hot: !process.env.VITEST })],
    test: {
      environment: 'jsdom',
      globals: true,
      coverage: {
        provider: 'v8',
      },
    },
  });
  ```
- **Svelte 4 compatibility:** The `@sveltejs/vite-plugin-svelte` version `^3.0.0` already in `package.json` works with Vitest. The `hot: !process.env.VITEST` flag disables HMR during test runs.
- **Separate vitest.config.js:** Keep this separate from `vite.config.js` (which handles the Wails build). Vitest automatically finds `vitest.config.js` and merges with `vite.config.js` defaults, but having a separate file keeps test configuration explicit.
- **Migration pattern (themeConverter.test.js):**
  - Replace `import { describe, it } from 'node:test'` with `import { describe, it, expect } from 'vitest'`
  - Remove `import assert from 'node:assert/strict'`
  - Replace `assert.equal(a, b)` with `expect(a).toBe(b)`
  - Replace `assert.deepEqual(a, b)` with `expect(a).toEqual(b)`
  - Replace `assert.ok(value, msg)` with `expect(value).toBeTruthy()`
  - Replace `assert.ok(!value.startsWith('#'), msg)` with `expect(value.startsWith('#')).toBe(false)`
  - This is a mechanical find-replace -- test logic and fixtures stay identical
- **Package.json scripts:**
  ```json
  "test": "vitest run",
  "test:watch": "vitest",
  "check": "svelte-check"
  ```
  Note: `svelte-check` requires `svelte-check` package. If not already installed, add it to devDependencies. Check first.
- **@testing-library/svelte:** Install for future component testing even though `themeConverter.test.js` doesn't need it. It's a devDependency with zero impact on production bundle.

### Risks & Edge Cases
- **Import path for themeConverter:** The test imports from `'./themeConverter.js'`. Vitest with the Svelte plugin should resolve this correctly, but verify the import works. If not, may need to adjust to `'./themeConverter'` (without extension).
- **jsdom limitations:** jsdom doesn't support WebSocket, Canvas, or other browser APIs used by xterm.js and Monaco. This is fine -- unit tests for utility functions don't need them. Component tests for xterm/Monaco components will need mocking (future stories).
- **Vitest version:** Use latest stable (^3.x as of 2026). Check compatibility with Vite 5.
- **`svelte-check` dependency:** The plan mentions adding a `check` script. Verify if `svelte-check` is already in devDependencies. If not, install `svelte-check` and `typescript` (peer dep).
- **Existing vite.config.js interaction:** Vitest reads `vite.config.js` as a fallback. The existing `chunkSizeWarningLimit` setting is build-only and won't affect tests.

### Reference Files
- `frontend/vite.config.js` -- existing Vite config (lines 1-12), shows Svelte plugin usage
- `frontend/package.json` -- current dependencies and scripts (no test script exists)
- `frontend/src/lib/themeConverter.test.js` -- the test file to migrate (565 lines, 10 describe blocks, ~40 test cases)
- `frontend/src/lib/themeConverter.js` -- the module under test (referenced by import)

## Acceptance Criteria

AC-1: Vitest is installed and configured
- Given the frontend directory has no test framework
- When a developer runs `cd frontend && npm install`
- Then `vitest`, `@testing-library/svelte`, and `jsdom` are installed as devDependencies
- And `frontend/vitest.config.js` exists with Svelte plugin, jsdom environment, and v8 coverage provider

AC-2: `npm test` runs Vitest
- Given `frontend/package.json` has `"test": "vitest run"` in scripts
- When a developer runs `cd frontend && npm test`
- Then Vitest discovers and runs all `*.test.js` files
- And reports test results with pass/fail counts

AC-3: themeConverter tests pass under Vitest
- Given `themeConverter.test.js` has been migrated from `node:test` to Vitest
- When a developer runs `cd frontend && npm test`
- Then all ~40 test cases pass
- And no `node:test` or `node:assert` imports remain in the file

AC-4: Test watch mode works
- Given `frontend/package.json` has `"test:watch": "vitest"` in scripts
- When a developer runs `cd frontend && npm run test:watch`
- Then Vitest starts in watch mode
- And re-runs affected tests when files change

AC-5: Migration preserves test coverage
- Given the original test file has ~40 test cases across 10 describe blocks
- When comparing the migrated file to the original
- Then every original test case has a corresponding Vitest test case
- And all fixtures (DRACULA_THEME, LIGHT_THEME, HC_BLACK_THEME, HC_LIGHT_THEME) are unchanged

## BDD Test Scenarios

### Scenario 1: Fresh install and test run
```gherkin
Feature: Frontend testing infrastructure

  Scenario: Developer sets up and runs tests
    Given the developer has cloned the repo
    And run "cd frontend && npm install"
    When they run "npm test"
    Then Vitest runs and discovers themeConverter.test.js
    And all test cases pass
    And the output shows "Tests: X passed" with X >= 40
```

### Scenario 2: Migrated test assertions work correctly
```gherkin
  Scenario: Vitest assertions match original node:assert behavior
    Given themeConverter.test.js uses vitest expect() assertions
    When the "Dracula dark theme" describe block runs
    Then expect(result.label).toBe('Dracula') passes
    And expect(result.css['--bg-deepest']).toBe('#282a36') passes
    And expect(result.monaco.base).toBe('vs-dark') passes
    And expect(Object.keys(result.css).length).toBe(16) passes
```

### Scenario 3: No node:test imports remain
```gherkin
  Scenario: Migration is complete
    Given the developer opens themeConverter.test.js
    When they search for "node:test" or "node:assert"
    Then no matches are found
    And all imports come from 'vitest' and local modules
```

### Scenario 4: Vitest config is correct for Svelte
```gherkin
  Scenario: Vitest handles Svelte files
    Given vitest.config.js includes the svelte() plugin
    And the test environment is set to "jsdom"
    When a future test imports a .svelte component
    Then Vitest can compile and render it
    And the jsdom environment provides DOM APIs
```

## Tasks / Subtasks

- [ ] Task 1: Install Vitest and dependencies (AC: AC-1)
  - [ ] Subtask 1a: Run `cd frontend && npm install -D vitest @testing-library/svelte jsdom`
  - [ ] Subtask 1b: Verify packages appear in `frontend/package.json` devDependencies
  - [ ] Subtask 1c: Check if `svelte-check` is needed for the `check` script; install if missing

- [ ] Task 2: Create Vitest configuration (AC: AC-1, AC-4)
  - [ ] Subtask 2a: Create `frontend/vitest.config.js` with Svelte plugin, jsdom environment, v8 coverage
  - [ ] Subtask 2b: Add `"test": "vitest run"` and `"test:watch": "vitest"` to `frontend/package.json` scripts

- [ ] Task 3: Migrate themeConverter.test.js to Vitest (AC: AC-3, AC-5)
  - [ ] Subtask 3a: Replace `import { describe, it } from 'node:test'` with `import { describe, it, expect } from 'vitest'`
  - [ ] Subtask 3b: Remove `import assert from 'node:assert/strict'`
  - [ ] Subtask 3c: Replace all `assert.equal()` with `expect().toBe()`
  - [ ] Subtask 3d: Replace all `assert.deepEqual()` with `expect().toEqual()`, `assert.ok()` with `expect().toBeTruthy()`
  - [ ] Subtask 3e: Verify all ~40 tests pass with `cd frontend && npm test`

- [ ] Task 4: Verify end-to-end (AC: AC-2, AC-3, AC-4)
  - [ ] Subtask 4a: Run `cd frontend && npm test` and confirm all tests pass
  - [ ] Subtask 4b: Run `cd frontend && npm run test:watch` and confirm watch mode starts

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] `cd frontend && npm test` passes with all tests green
- [ ] `vitest.config.js` exists with correct Svelte + jsdom configuration
- [ ] No `node:test` or `node:assert` imports remain in any test file
- [ ] All original test cases are preserved in the migrated file
- [ ] `/simplify` run on `vitest.config.js` and `themeConverter.test.js`
- [ ] `cd frontend && npx vite build --mode development` still builds successfully
