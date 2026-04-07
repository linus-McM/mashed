# Svelte 4 Upgrade & xyflow Installation

**Story ID**: bmad-01
**Status**: ready
**Priority**: P0
**Depends On**: none

## Description

Upgrade the frontend from Svelte 3.49 to Svelte 4.2+ and install `@xyflow/svelte` as a dependency. This is a hard prerequisite for every other BMAD story because `@xyflow/svelte` requires Svelte 4+. Svelte 4 is backward-compatible with all patterns currently used in the codebase (export props, createEventDispatcher, `$:` reactivity, `on:event` directives). The `@sveltejs/vite-plugin-svelte` must also be upgraded to a version compatible with Svelte 4.

## Developer Notes

### Architecture

This story touches only `frontend/` -- no Go changes.

**Files to modify:**

- `frontend/package.json` -- Upgrade `svelte` from `^3.49.0` to `^4.2.0`, upgrade `@sveltejs/vite-plugin-svelte` from `^1.0.1` to `^3.0.0` (Svelte 4 compatible), add `@xyflow/svelte` as a dependency.
- `frontend/vite.config.js` -- Likely no changes needed, but verify after install. The current config is minimal (just `plugins: [svelte()]` and a chunk warning limit).

**Existing `devDependencies` to upgrade:**
```json
{
  "@sveltejs/vite-plugin-svelte": "^3.0.0",
  "svelte": "^4.2.0",
  "vite": "^5.0.0"
}
```

**New dependency to add:**
```json
{
  "@xyflow/svelte": "^1.0.0"
}
```

Note: `vite` itself may need to go to v5 for the newer vite-plugin-svelte. Check compatibility.

### Technical Considerations

- `lucide-svelte@0.577.0` is already installed and supports Svelte 4.
- `monaco-editor` is framework-agnostic, unaffected.
- `@xterm/xterm`, `@xterm/addon-fit`, `@xterm/addon-canvas` are framework-agnostic, unaffected.
- The frontend uses no SvelteKit, no `+page.svelte`, no `$app/` imports -- this is a pure Svelte 3/4 SPA compiled by Vite for Wails.
- Svelte 4 dropped IE11 support and changed some internal compiler output, but the public API used in this project (`export let`, `createEventDispatcher`, `$:` reactive statements, `{#if}` / `{#each}` blocks, `on:click`, `bind:`) is fully compatible.

### Risks & Edge Cases

- **vite-plugin-svelte version mismatch**: `@sveltejs/vite-plugin-svelte@1.x` does not support Svelte 4. Must upgrade to 3.x. If vite 3.x is incompatible with vite-plugin-svelte 3.x, vite must also be upgraded.
- **Peer dependency conflicts**: Run `npm install` and check for peer dependency warnings. Resolve any conflicts before proceeding.
- **Build regression**: The build must produce the same functional output. Run `wails build` (or `cd frontend && npm run build`) to verify zero errors.
- **Runtime regression**: After build, launch the app and verify existing views still render (NotificationFeed, AgentDetail, Settings, SpawnAgent, Setup).

### Reference Files

- `frontend/package.json` -- current dependencies
- `frontend/vite.config.js` -- current vite config
- `frontend/src/App.svelte` -- main entry, uses standard Svelte 3 patterns
- `frontend/src/views/NotificationFeed.svelte` -- representative view component
- `frontend/src/components/TitleBar.svelte` -- representative component

## Acceptance Criteria

- [ ] AC1: Given the frontend `package.json`, When the dependency versions are read, Then `svelte` is `^4.2.0` or higher, `@sveltejs/vite-plugin-svelte` is `^3.0.0` or higher, and `@xyflow/svelte` is listed as a dependency.
- [ ] AC2: Given the upgraded dependencies, When `cd frontend && npm install && npm run build` is run, Then the build completes with zero errors and zero Svelte-related warnings.
- [ ] AC3: Given the built application, When `wails build` is run, Then it produces a working binary with no errors.
- [ ] AC4: Given the running application after upgrade, When the user navigates through existing views (feed, detail, settings, spawn), Then all views render correctly with no regressions.
- [ ] AC5: Given `@xyflow/svelte` is installed, When a test Svelte file imports `import { SvelteFlow } from '@xyflow/svelte'`, Then the import resolves without errors during build.

## BDD Test Scenarios

### Scenario 1: Dependency versions are correct after upgrade

```gherkin
Feature: Svelte 4 Upgrade

  Scenario: Package versions are upgraded correctly
    Given the file frontend/package.json exists
    When the file is parsed as JSON
    Then devDependencies.svelte matches "^4.2.0" or higher
    And devDependencies["@sveltejs/vite-plugin-svelte"] matches "^3.0.0" or higher
    And dependencies["@xyflow/svelte"] is present

  Scenario: Frontend build succeeds
    Given all npm dependencies are installed via npm install
    When npm run build is executed in the frontend directory
    Then the exit code is 0
    And no error output is produced

  Scenario: Wails build succeeds
    Given the frontend build has completed
    When wails build is run from the project root
    Then the exit code is 0
    And a binary is produced in build/bin/

  Scenario: xyflow import resolves
    Given @xyflow/svelte is installed
    When a Svelte component contains "import { SvelteFlow } from '@xyflow/svelte'"
    Then vite resolves the import without errors during build
```

## Tasks / Subtasks

- [ ] Task 1: Upgrade package.json dependencies (AC: AC1)
  - [ ] Subtask 1a: Update `svelte` to `^4.2.0`
  - [ ] Subtask 1b: Update `@sveltejs/vite-plugin-svelte` to `^3.0.0`
  - [ ] Subtask 1c: Update `vite` to `^5.0.0` if required by vite-plugin-svelte 3.x
  - [ ] Subtask 1d: Add `@xyflow/svelte` to dependencies
- [ ] Task 2: Install and resolve dependency conflicts (AC: AC1, AC2)
  - [ ] Subtask 2a: Run `npm install` and resolve any peer dependency conflicts
  - [ ] Subtask 2b: Verify `node_modules/@xyflow/svelte` exists and contains expected exports
- [ ] Task 3: Verify vite config compatibility (AC: AC2)
  - [ ] Subtask 3a: Check `frontend/vite.config.js` works with upgraded plugin
  - [ ] Subtask 3b: Adjust chunk size warning limit if @xyflow/svelte adds significant bundle size
- [ ] Task 4: Build verification (AC: AC2, AC3, AC5)
  - [ ] Subtask 4a: Run `cd frontend && npm run build` -- zero errors
  - [ ] Subtask 4b: Run `wails build` -- zero errors
  - [ ] Subtask 4c: Create a minimal test import of SvelteFlow to verify resolution
- [ ] Task 5: Runtime regression check (AC: AC4)
  - [ ] Subtask 5a: Launch app and verify all existing views render
  - [ ] Subtask 5b: Verify Terminal.svelte still connects via xterm
  - [ ] Subtask 5c: Verify Monaco editor still loads in AgentDetail/CodeEditor

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] `cd frontend && npm install && npm run build` succeeds
- [ ] `wails build` succeeds
- [ ] All existing views render without regression
- [ ] No peer dependency warnings in npm install output
- [ ] /simplify run on all modified files
