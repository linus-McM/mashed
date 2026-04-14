/**
 * @vitest-environment jsdom
 */
// frontend/src/components/bmad/__tests__/ProcessNode.multiInput.test.js
// Story breadcrumbs-04: Multi-input breadcrumb rendering
// Task 2 RED phase: one .breadcrumb-row per declared input/output
//
// RED Phase: These tests MUST FAIL until the ui-engineer replaces the single
// .breadcrumb-row block in ProcessNode.svelte with an {#each} loop that emits
// one row per symbolic input (and mirrors for outputs).
//
// Mock pattern mirrors frontend/src/components/bmad/__tests__/CommandNode.breadcrumb.test.js.

import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest';

// Stub @xyflow/svelte so mount works without <SvelteFlowProvider />.
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

import ProcessNode from '../ProcessNode.svelte';

const EM_DASH = '\u2014';

function mountProcessNode(container, props) {
  return new ProcessNode({ target: container, props });
}

/**
 * Build props for ProcessNode:
 * @param {{inputs?:string[], outputs?:string[], config?:object, status?:string}} o
 */
function makeProps(o = {}) {
  return {
    data: {
      process: {
        name: 'Test',
        phase: 'implementation',
        agentRole: 'developer',
        inputs: o.inputs ?? [],
        outputs: o.outputs ?? [],
      },
      config: o.config ?? {},
      status: o.status ?? 'pending',
      label: 'Test Node',
    },
    id: 'proc-1',
    selected: false,
  };
}

/** Collect breadcrumb rows belonging to the IN group (positioned before any OUT row). */
function collectBreadcrumbRows(container) {
  return Array.from(container.querySelectorAll('.breadcrumb-row'));
}

describe('ProcessNode — multi-input breadcrumb rendering (breadcrumbs-04)', () => {
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

  // AC-1 + AC-2: two symbolic inputs → two IN breadcrumb rows in declared order
  it('TestStory4_AC1_TwoInputs_RendersTwoBreadcrumbRows_InDeclaredOrder', () => {
    node = mountProcessNode(
      container,
      makeProps({
        inputs: ['prd', 'sprint-status'],
        config: {
          inputPaths: {
            prd: '/a/PRD.md',
            'sprint-status': '/a/sprint.yaml',
          },
        },
      }),
    );

    const rows = collectBreadcrumbRows(container);
    expect(rows.length, 'one .breadcrumb-row per declared input').toBeGreaterThanOrEqual(2);

    const texts = rows.map((r) => r.textContent?.trim());
    const prdIdx = texts.indexOf('.../PRD.md');
    const sprintIdx = texts.indexOf('.../sprint.yaml');

    expect(prdIdx, 'prd breadcrumb must render').toBeGreaterThanOrEqual(0);
    expect(sprintIdx, 'sprint-status breadcrumb must render').toBeGreaterThanOrEqual(0);
    expect(prdIdx, 'prd row must precede sprint-status row (declared order)').toBeLessThan(
      sprintIdx,
    );
  });

  // AC-2 partial mapping: unmapped input renders em-dash
  it('TestStory4_AC2_PartialMapping_ShowsEmDashForUnmapped', () => {
    node = mountProcessNode(
      container,
      makeProps({
        inputs: ['prd', 'sprint-status'],
        config: { inputPaths: { prd: '/a/PRD.md' } },
      }),
    );

    const rows = collectBreadcrumbRows(container);
    const texts = rows.map((r) => r.textContent?.trim());
    expect(texts).toContain('.../PRD.md');
    expect(texts, 'unmapped sprint-status row must render em-dash').toContain(EM_DASH);
  });

  // AC-3: legacy fallback — single input + legacy config.inputPath (no map)
  it('TestStory4_AC3_LegacyFallback_SingleInput_UsesInputPath', () => {
    node = mountProcessNode(
      container,
      makeProps({
        inputs: ['brainstorm-notes'],
        config: { inputPath: '/a/brain.md' },
      }),
    );

    const rows = collectBreadcrumbRows(container);
    const texts = rows.map((r) => r.textContent?.trim());
    expect(texts, 'legacy single-input fallback must render .../brain.md').toContain(
      '.../brain.md',
    );
  });

  // AC-4: title attribute holds full absolute path (not basename)
  it('TestStory4_AC4_TitleAttribute_HoldsFullPath', () => {
    const fullPath = '/long/nested/dir/file.md';
    node = mountProcessNode(
      container,
      makeProps({
        inputs: ['doc'],
        config: { inputPaths: { doc: fullPath } },
      }),
    );

    const rows = collectBreadcrumbRows(container);
    const match = rows.find((r) => r.getAttribute('title') === fullPath);
    expect(
      match,
      'a .breadcrumb-row must carry title={fullPath} for the doc input',
    ).toBeDefined();
  });

  // AC-1 mirror: outputs render one row per declared output, in order
  it('TestStory4_AC1_OutputsMirrorInputs', () => {
    node = mountProcessNode(
      container,
      makeProps({
        outputs: ['draft', 'final'],
        config: {
          outputPaths: {
            draft: '/d.md',
            final: '/f.md',
          },
        },
      }),
    );

    const rows = collectBreadcrumbRows(container);
    const texts = rows.map((r) => r.textContent?.trim());
    const draftIdx = texts.indexOf('.../d.md');
    const finalIdx = texts.indexOf('.../f.md');

    expect(draftIdx, 'draft output breadcrumb must render').toBeGreaterThanOrEqual(0);
    expect(finalIdx, 'final output breadcrumb must render').toBeGreaterThanOrEqual(0);
    expect(draftIdx).toBeLessThan(finalIdx);
  });

  // AC-1 edge: zero inputs → zero IN breadcrumb rows
  it('TestStory4_AC1_ZeroInputs_RendersNoInBreadcrumbRows', () => {
    node = mountProcessNode(
      container,
      makeProps({
        inputs: [],
        outputs: ['draft'],
        config: { outputPaths: { draft: '/d.md' } },
      }),
    );

    // Only the OUT row should exist. If the loop is missing or renders a
    // placeholder IN row, this assertion catches it.
    const rows = collectBreadcrumbRows(container);
    const draftMatches = rows.filter((r) => r.textContent?.trim() === '.../d.md');
    expect(draftMatches.length, 'OUT row must still render').toBeGreaterThanOrEqual(1);
    expect(rows.length, 'with zero inputs, only OUT breadcrumbs exist').toBe(
      draftMatches.length,
    );
  });
});
