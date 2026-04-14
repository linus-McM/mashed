/**
 * @vitest-environment jsdom
 */
// frontend/src/components/bmad/__tests__/CommandNode.breadcrumb.test.js
// Story breadcrumbs-03: Breadcrumb row on CommandNode
// Task RED phase: Failing tests for .breadcrumb-row rendering (AC-1, AC-2)
//
// RED Phase: These tests MUST FAIL until the ui-engineer adds .breadcrumb-row
// markup to CommandNode.svelte — those elements do not exist yet.
//
// Note: @testing-library/svelte is not installed. Using the Svelte 4 component
// constructor API directly. vitest.config.js provides jsdom + @sveltejs/vite-plugin-svelte.

import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest';

// Mock @xyflow/svelte Handle — requires <SvelteFlowProvider /> context at runtime
// which is unavailable in unit tests. Only the three fields Svelte 4's
// mount_component/destroy_component actually read are provided.
vi.mock('@xyflow/svelte', () => ({
  Handle: class {
    $$ = {
      fragment: { c() {}, m() {}, p() {}, d() {}, l() {} },
      on_mount: [],
      on_destroy: [],
      after_update: [],
    };
    constructor(_opts) {}
    $set(_props) {}
    $destroy() { this.$$.fragment = null; this.$$.on_destroy = []; }
    $on(_event, _fn) { return () => {}; }
  },
  Position: { Left: 'left', Right: 'right', Top: 'top', Bottom: 'bottom' },
}));

import CommandNode from '../CommandNode.svelte';

const EM_DASH = '\u2014';

function mountCommandNode(container, props) {
  return new CommandNode({ target: container, props });
}

/** Minimal valid props shape — preserves data.config structure CommandNode reads */
function makeNodeProps(overrides = {}) {
  return {
    data: {
      config: {
        commandName: 'test',
        commandDescription: '',
        inputPath: overrides.inputPath ?? '',
        outputPath: overrides.outputPath ?? '',
        ...(overrides.config ?? {}),
      },
      status: overrides.status ?? 'pending',
      processId: 'some-id',
    },
    id: 'cmd-1',
    selected: false,
  };
}

describe('CommandNode — breadcrumb row (breadcrumbs-03)', () => {
  let container;
  let node = null;

  beforeEach(() => {
    container = document.createElement('div');
    document.body.appendChild(container);
  });

  afterEach(() => {
    node?.$destroy();
    container.remove();
    node = null;
  });

  // AC-1 + AC-2: Resolved outputPath renders .../basename with full title
  it('TestStory3_AC2_ResolvedOutputPath_ShowsBasename', () => {
    const outputPath = '/Users/x/notes/brief.md';
    node = mountCommandNode(container, makeNodeProps({ outputPath }));

    const rows = container.querySelectorAll('.breadcrumb-row');
    expect(rows.length, '.breadcrumb-row must exist when outputPath is set').toBeGreaterThan(0);

    const outRow = Array.from(rows).find(
      (el) => el.textContent?.trim() === '.../brief.md',
    );
    expect(outRow, 'OUT breadcrumb text must be .../brief.md').not.toBeUndefined();
    expect(outRow?.getAttribute('title')).toBe(outputPath);
  });

  // AC-2: Empty/missing outputPath renders em-dash with .unresolved class and title="unresolved"
  it('TestStory3_AC2_EmptyOutputPath_RendersEmDash', () => {
    node = mountCommandNode(container, makeNodeProps({ outputPath: '' }));

    const unresolvedRow = container.querySelector('.breadcrumb-row.unresolved');
    expect(
      unresolvedRow,
      '.breadcrumb-row.unresolved must exist when outputPath is empty',
    ).not.toBeNull();
    expect(unresolvedRow?.textContent?.trim()).toBe(EM_DASH);
    expect(unresolvedRow?.getAttribute('title')).toBe('unresolved');
  });

  // AC-1: Resolved inputPath renders .../basename breadcrumb row
  it('TestStory3_AC1_ResolvedInputPath_ShowsBasename', () => {
    const inputPath = '/a/b/c.txt';
    node = mountCommandNode(container, makeNodeProps({ inputPath }));

    const rows = container.querySelectorAll('.breadcrumb-row');
    expect(rows.length, '.breadcrumb-row must exist when inputPath is set').toBeGreaterThan(0);

    const inRow = Array.from(rows).find(
      (el) => el.textContent?.trim() === '.../c.txt',
    );
    expect(inRow, 'IN breadcrumb text must be .../c.txt').not.toBeUndefined();
    expect(inRow?.getAttribute('title')).toBe(inputPath);
  });

  // Regression seed: class name is `.breadcrumb-row`, NOT `.command-breadcrumb-row`
  // Parity contract: CommandNode must use the identical class as ProcessNode.
  it('TestStory3_Regression_BreadcrumbRowClass_NoDivergence', () => {
    const outputPath = '/some/path/output.md';
    node = mountCommandNode(container, makeNodeProps({ outputPath }));

    const withClass = container.querySelector('.breadcrumb-row');
    const withWrongClass = container.querySelector('.command-breadcrumb-row');

    expect(withClass, '.breadcrumb-row must exist (shared class across all node types)').not.toBeNull();
    expect(withWrongClass, '.command-breadcrumb-row must NOT exist — use shared .breadcrumb-row').toBeNull();
  });
});
