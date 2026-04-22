/**
 * @vitest-environment jsdom
 */
// Story breadcrumbs-03: Breadcrumb row on CommandNode
// Covers AC-1 (IN) / AC-2 (OUT + em-dash unresolved) + regression class parity.
//
// @testing-library/svelte is not a project dep — we use the Svelte 4
// component constructor via the typed `mountComponent` helper.

import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest';

import type { CanvasNodeData } from '../../../types/workflow';
import { mountComponent, type MountedComponent } from './mountSvelte';

// Stub `@xyflow/svelte` — `<Handle>` requires a `<SvelteFlowProvider />` at
// runtime. `vi.mock` factory is hoisted but may await a dynamic import, so
// the stub stays centralised in `mountSvelte.ts`.
vi.mock('@xyflow/svelte', async () => (await import('./mountSvelte')).xyflowHandleStub());

import CommandNode from '../CommandNode.svelte';

const EM_DASH = '—';

interface CommandNodeProps {
  data: CanvasNodeData;
  id: string;
  selected: boolean;
}

interface BreadcrumbOverrides {
  inputPath?: string;
  outputPath?: string;
  status?: string;
  config?: Record<string, unknown>;
}

/** Minimal valid props shape — preserves data.config structure CommandNode reads. */
function makeNodeProps(overrides: BreadcrumbOverrides = {}): CommandNodeProps {
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

  // AC-1 + AC-2: Resolved outputPath renders .../basename with full title
  it('TestStory3_AC2_ResolvedOutputPath_ShowsBasename', () => {
    const outputPath = '/Users/x/notes/brief.md';
    node = mountComponent(CommandNode, container, makeNodeProps({ outputPath }));

    const rows = container.querySelectorAll<HTMLElement>('.breadcrumb-row');
    expect(rows.length, '.breadcrumb-row must exist when outputPath is set').toBeGreaterThan(0);

    const outRow = Array.from(rows).find((el) => el.textContent?.trim() === '.../brief.md');
    expect(outRow, 'OUT breadcrumb text must be .../brief.md').not.toBeUndefined();
    expect(outRow?.getAttribute('title')).toBe(outputPath);
  });

  // AC-2: Empty/missing outputPath renders em-dash with .unresolved class and title="unresolved"
  it('TestStory3_AC2_EmptyOutputPath_RendersEmDash', () => {
    node = mountComponent(CommandNode, container, makeNodeProps({ outputPath: '' }));

    const unresolvedRow = container.querySelector<HTMLElement>('.breadcrumb-row.unresolved');
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
    node = mountComponent(CommandNode, container, makeNodeProps({ inputPath }));

    const rows = container.querySelectorAll<HTMLElement>('.breadcrumb-row');
    expect(rows.length, '.breadcrumb-row must exist when inputPath is set').toBeGreaterThan(0);

    const inRow = Array.from(rows).find((el) => el.textContent?.trim() === '.../c.txt');
    expect(inRow, 'IN breadcrumb text must be .../c.txt').not.toBeUndefined();
    expect(inRow?.getAttribute('title')).toBe(inputPath);
  });

  // Regression seed: class name is `.breadcrumb-row`, NOT `.command-breadcrumb-row`
  // Parity contract: CommandNode must use the identical class as ProcessNode.
  it('TestStory3_Regression_BreadcrumbRowClass_NoDivergence', () => {
    const outputPath = '/some/path/output.md';
    node = mountComponent(CommandNode, container, makeNodeProps({ outputPath }));

    const withClass = container.querySelector('.breadcrumb-row');
    const withWrongClass = container.querySelector('.command-breadcrumb-row');

    expect(
      withClass,
      '.breadcrumb-row must exist (shared class across all node types)',
    ).not.toBeNull();
    expect(
      withWrongClass,
      '.command-breadcrumb-row must NOT exist — use shared .breadcrumb-row',
    ).toBeNull();
  });
});
