/**
 * @vitest-environment jsdom
 */
// frontend/src/components/bmad/ProcessNode.test.ts
// Story breadcrumbs-02: Breadcrumb row on ProcessNode
// Task 1 (RED phase): Component tests for .breadcrumb-row rendering (AC-1, AC-2, AC-4)
//
// RED Phase: These tests MUST FAIL until the frontend engineer adds .breadcrumb-row
// markup to ProcessNode.svelte — the element does not exist yet.
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
      on_mount: [] as Array<() => unknown>,
      on_destroy: [] as Array<() => void>,
      after_update: [] as Array<() => void>,
    };
    constructor(_opts: unknown) {}
    $set(_props: unknown) {}
    $destroy() { this.$$.fragment = null as unknown as typeof this.$$.fragment; this.$$.on_destroy = []; }
    $on(_event: string, _fn: unknown) { return () => {}; }
  },
  Position: { Left: 'left', Right: 'right', Top: 'top', Bottom: 'bottom' },
}));

import ProcessNode from './ProcessNode.svelte';

const EM_DASH = '\u2014';

type NodeInstance = { $destroy(): void };
type SvelteInit = new (opts: { target: HTMLElement; props: object }) => NodeInstance;

function mountProcessNode(container: HTMLElement, props: object): NodeInstance {
  return new (ProcessNode as unknown as SvelteInit)({ target: container, props });
}

function makeNodeProps(overrides: {
  inputs?: string[];
  outputs?: string[];
  config?: Record<string, unknown>;
  status?: string;
}) {
  return {
    data: {
      process: {
        inputs: overrides.inputs ?? [],
        outputs: overrides.outputs ?? [],
      },
      config: overrides.config ?? {},
      status: overrides.status ?? 'pending',
    },
    id: 'node-test',
    selected: false,
  };
}

describe('ProcessNode — breadcrumb row (breadcrumbs-02)', () => {
  let container: HTMLElement;
  let node: NodeInstance | null = null;

  beforeEach(() => {
    container = document.createElement('div');
    document.body.appendChild(container);
  });

  afterEach(() => {
    node?.$destroy();
    container.remove();
    node = null;
  });

  // Story breadcrumbs-02, AC-1 + BDD "Unresolved path renders em-dash"
  it('TestStory2_AC1_UnresolvedInputRendersEmDash', () => {
    node = mountProcessNode(container, makeNodeProps({ inputs: ['brainstorm-notes'] }));

    const breadcrumbRow = container.querySelector('.breadcrumb-row');
    expect(breadcrumbRow, '.breadcrumb-row must exist for unresolved IN').not.toBeNull();
    expect(breadcrumbRow?.textContent?.trim()).toBe(EM_DASH);
    expect(breadcrumbRow?.getAttribute('title')).toBe('unresolved');
  });

  // Story breadcrumbs-02, AC-2 + BDD "Resolved path renders basename"
  it('TestStory2_AC2_OutputBreadcrumbShowsBasename', () => {
    const outputPath = '/abs/_bmad-output/planning-artifacts/PRD.md';
    node = mountProcessNode(container, makeNodeProps({ outputs: ['prd'], config: { outputPath } }));

    const breadcrumbRows = container.querySelectorAll('.breadcrumb-row');
    expect(breadcrumbRows.length, 'at least one .breadcrumb-row must exist').toBeGreaterThan(0);

    const outRow = Array.from(breadcrumbRows).find((el) => el.textContent?.trim() === '.../PRD.md');
    expect(outRow, 'OUT breadcrumb text must be .../PRD.md').not.toBeUndefined();
    expect(outRow?.getAttribute('title')).toBe(outputPath);
  });

  // Story breadcrumbs-02, AC-2: unresolved row title="unresolved" (Design Brief line 139)
  it('TestStory2_AC2_UnresolvedTitleAttr', () => {
    node = mountProcessNode(container, makeNodeProps({ inputs: ['brainstorm-notes'], outputs: ['prd'] }));

    const unresolvedRow = Array.from(container.querySelectorAll('.breadcrumb-row')).find(
      (el) => el.getAttribute('title') === 'unresolved'
    );
    expect(unresolvedRow, 'at least one .breadcrumb-row must have title="unresolved"').not.toBeUndefined();
  });

  // Story breadcrumbs-02, AC-4 + BDD "Symbolic rows unchanged"
  it('TestStory2_AC4_SymbolicRowsUnchanged', () => {
    node = mountProcessNode(container, makeNodeProps({ inputs: ['brainstorm-notes'], outputs: ['prd'] }));

    const labelTexts = Array.from(container.querySelectorAll('.artifact-label')).map((el) => el.textContent?.trim());
    const listTexts = Array.from(container.querySelectorAll('.artifact-list')).map((el) => el.textContent?.trim());

    expect(labelTexts).toContain('in:');
    expect(labelTexts).toContain('out:');
    expect(listTexts).toContain('brainstorm-notes');
    expect(listTexts).toContain('prd');
  });
});
