/**
 * @vitest-environment jsdom
 */
// Story breadcrumbs-04: Multi-input breadcrumb rendering
// One .breadcrumb-row per declared input/output.
//
// Mock pattern mirrors ProcessNode.status.test.ts.

import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest';

import { mountComponent, type MountedComponent } from './mountSvelte';

// `vi.mock` factory is hoisted but may await a dynamic import — keeps the
// xyflow stub centralised in `mountSvelte.ts`.
vi.mock('@xyflow/svelte', async () => (await import('./mountSvelte')).xyflowHandleStub());

import ProcessNode from '../ProcessNode.svelte';

const EM_DASH = '—';

// Test-local subset — see ProcessNode.status.test.ts for rationale.
interface TestProcess {
  name?: string;
  phase?: string;
  agentRole?: string;
  inputs?: string[];
  outputs?: string[];
}

interface TestNodeData {
  process?: TestProcess | null;
  config?: Record<string, unknown>;
  status?: string;
  label?: string;
}

interface ProcessNodeProps {
  data: TestNodeData;
  id: string;
  selected: boolean;
}

interface MultiInputOverrides {
  inputs?: string[];
  outputs?: string[];
  config?: Record<string, unknown>;
  status?: string;
}

function makeProps(o: MultiInputOverrides = {}): ProcessNodeProps {
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
function collectBreadcrumbRows(container: HTMLElement): HTMLElement[] {
  return Array.from(container.querySelectorAll<HTMLElement>('.breadcrumb-row'));
}

describe('ProcessNode — multi-input breadcrumb rendering (breadcrumbs-04)', () => {
  let container: HTMLDivElement;
  let node: MountedComponent | null = null;

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
    node = mountComponent(
      ProcessNode,
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
    node = mountComponent(
      ProcessNode,
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
    node = mountComponent(
      ProcessNode,
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
    node = mountComponent(
      ProcessNode,
      container,
      makeProps({
        inputs: ['doc'],
        config: { inputPaths: { doc: fullPath } },
      }),
    );

    const rows = collectBreadcrumbRows(container);
    const match = rows.find((r) => r.getAttribute('title') === fullPath);
    expect(match, 'a .breadcrumb-row must carry title={fullPath} for the doc input').toBeDefined();
  });

  // AC-1 mirror: outputs render one row per declared output, in order
  it('TestStory4_AC1_OutputsMirrorInputs', () => {
    node = mountComponent(
      ProcessNode,
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
    node = mountComponent(
      ProcessNode,
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
    expect(rows.length, 'with zero inputs, only OUT breadcrumbs exist').toBe(draftMatches.length);
  });
});
