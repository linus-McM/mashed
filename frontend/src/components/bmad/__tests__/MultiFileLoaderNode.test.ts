/**
 * @vitest-environment jsdom
 */
// Story breadcrumbs-08 tests for MultiFileLoaderNode.svelte.

import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest';

import type { CanvasNodeData } from '../../../types/workflow';
import { mountComponent, type MountedComponent } from './mountSvelte';

// `vi.mock` factory is hoisted but may await a dynamic import — keeps the
// xyflow stub centralised in `mountSvelte.ts`.
vi.mock('@xyflow/svelte', async () => (await import('./mountSvelte')).xyflowHandleStub());

import MultiFileLoaderNode from '../MultiFileLoaderNode.svelte';

const EM_DASH = '—';

interface Entry {
  label: string;
  path: string;
}

interface MultiFileLoaderProps {
  data: CanvasNodeData;
  id: string;
  selected: boolean;
}

interface MultiFileLoaderOverrides {
  status?: string;
}

function makeProps(
  entries: Entry[] | string | undefined,
  overrides: MultiFileLoaderOverrides = {},
): MultiFileLoaderProps {
  return {
    data: {
      label: 'Multi File Loader',
      nodeType: 'multiFileLoader',
      config: {
        entries: typeof entries === 'string' ? entries : JSON.stringify(entries ?? []),
      },
      status: overrides.status ?? 'pending',
    },
    id: 'mfl-1',
    selected: false,
  };
}

describe('MultiFileLoaderNode — breadcrumb rendering (breadcrumbs-08)', () => {
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

  // AC-3a: two configured entries → exactly two breadcrumb rows with the
  // correct prefix labels (story example: brief + file[1]).
  it('TestStory8_AC3a_TwoEntries_TwoBreadcrumbsWithPrefixLabels', () => {
    node = mountComponent(
      MultiFileLoaderNode,
      container,
      makeProps([
        { label: 'brief', path: '/x/a.md' },
        { label: '', path: '/x/b.md' },
      ]),
    );

    const rows = container.querySelectorAll('.breadcrumb-row');
    expect(rows.length, 'expected exactly two .breadcrumb-row entries').toBe(2);

    const html = container.innerHTML;
    expect(html).toContain('brief');
    expect(html).toContain('file[1]');
  });

  // AC-3b: each row text is .../basename and title is the absolute path.
  it('TestStory8_AC3b_BreadcrumbBasename_AndTitleAbsolute', () => {
    node = mountComponent(
      MultiFileLoaderNode,
      container,
      makeProps([
        { label: 'brief', path: '/x/a.md' },
        { label: '', path: '/x/b.md' },
      ]),
    );

    const rows = Array.from(container.querySelectorAll<HTMLElement>('.breadcrumb-row'));
    const aRow = rows.find((el) => el.textContent?.includes('.../a.md'));
    const bRow = rows.find((el) => el.textContent?.includes('.../b.md'));

    expect(aRow, 'row for a.md').not.toBeUndefined();
    expect(bRow, 'row for b.md').not.toBeUndefined();
    expect(aRow?.getAttribute('title')).toBe('/x/a.md');
    expect(bRow?.getAttribute('title')).toBe('/x/b.md');
  });

  // AC-3c: empty entries list → ghost row with em-dash + "(no files)" hint.
  it('TestStory8_AC3c_NoEntries_RendersGhostPlaceholder', () => {
    node = mountComponent(MultiFileLoaderNode, container, makeProps([]));

    const ghost = container.querySelector('.path-list .empty-state');
    expect(ghost, '.path-list .empty-state must render when zero entries').not.toBeNull();
    expect(ghost?.textContent ?? '').toContain(EM_DASH);
    expect(ghost?.textContent ?? '').toMatch(/no files/i);
  });

  // AC-3d: footer shows count badge "N files".
  it('TestStory8_AC3d_FooterCountBadge_NFiles', () => {
    node = mountComponent(
      MultiFileLoaderNode,
      container,
      makeProps([
        { label: 'a', path: '/x/1' },
        { label: 'b', path: '/x/2' },
        { label: 'c', path: '/x/3' },
      ]),
    );

    const badge = container.querySelector('.count-badge');
    expect(badge, '.count-badge must render in the footer').not.toBeNull();
    expect(badge?.textContent?.trim()).toBe('3 files');
  });

  // AC-3e: malformed JSON in config.entries → no throw, falls back to empty
  // state. The constructor itself succeeds because parseEntries guards.
  it('TestStory8_AC3e_MalformedEntriesJSON_FallsBackToEmpty', () => {
    expect(() => {
      node = mountComponent(MultiFileLoaderNode, container, makeProps('not-json'));
    }).not.toThrow();
    const ghost = container.querySelector('.path-list .empty-state');
    expect(ghost, 'malformed JSON must fall back to ghost state').not.toBeNull();
  });

  // AC-3f: positional fallback applies only when label is empty string.
  // A non-empty label (even whitespace) wins over file[N].
  it('TestStory8_AC3f_NonEmptyLabel_BeatsPositionalFallback', () => {
    node = mountComponent(
      MultiFileLoaderNode,
      container,
      makeProps([
        { label: '', path: '/x/a' }, // file[0]
        { label: 'plan', path: '/x/b' }, // plan, NOT file[1]
        { label: '', path: '/x/c' }, // file[2]
      ]),
    );

    const html = container.innerHTML;
    expect(html).toContain('file[0]');
    expect(html).toContain('plan');
    expect(html).toContain('file[2]');
    expect(html).not.toContain('file[1]'); // positional skipped because index 1 is labeled
  });

  // Class parity: the breadcrumb class is `.breadcrumb-row` (shared, no
  // divergent .multifile-breadcrumb-row).
  it('TestStory8_BreadcrumbRowClass_NoDivergence', () => {
    node = mountComponent(MultiFileLoaderNode, container, makeProps([{ label: 'a', path: '/x' }]));
    expect(container.querySelector('.breadcrumb-row')).not.toBeNull();
    expect(container.querySelector('.multifile-breadcrumb-row')).toBeNull();
  });
});
