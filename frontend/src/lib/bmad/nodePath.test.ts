// frontend/src/lib/bmad/nodePath.test.ts
// Story breadcrumbs-01: Discovery + node-path data model
// Story breadcrumbs-02: Breadcrumb row on ProcessNode
//
// Task 1 RED phase (breadcrumbs-01): Tests for getNodePath helper (AC-3)
// Task 1 RED phase (breadcrumbs-02): Tests for formatBreadcrumb formatter (AC-1, AC-2)
//
// formatBreadcrumb tests MUST FAIL until the frontend engineer exports
// formatBreadcrumb from ./nodePath.ts — it does not exist yet.

import { describe, it, expect, beforeAll } from 'vitest';
import { getNodePath } from './nodePath';

describe('getNodePath — AC-3: safe defaults', () => {
  // Story breadcrumbs-01, AC-3 + BDD: empty config → '' for both dirs
  it('TestStory1_AC3_EmptyConfigReturnsEmpty', () => {
    const node = { data: { config: {} } };
    expect(getNodePath(node, 'in')).toBe('');
    expect(getNodePath(node, 'out')).toBe('');
  });

  // Story breadcrumbs-01, AC-3: node with no data property → ''
  it('TestStory1_AC3_MissingDataReturnsEmpty', () => {
    const node = {};
    expect(getNodePath(node, 'in')).toBe('');
    expect(getNodePath(node, 'out')).toBe('');
  });

  // Story breadcrumbs-01, AC-3: node with data but no config → ''
  it('TestStory1_AC3_MissingConfigReturnsEmpty', () => {
    const node = { data: {} };
    expect(getNodePath(node, 'in')).toBe('');
    expect(getNodePath(node, 'out')).toBe('');
  });

  // Story breadcrumbs-01, AC-3: config.inputPath is null → ''
  it('TestStory1_AC3_NullValueReturnsEmpty', () => {
    const node = { data: { config: { inputPath: null } } };
    expect(getNodePath(node, 'in')).toBe('');
  });

  // Story breadcrumbs-01, AC-3: config.inputPath is a number → ''
  it('TestStory1_AC3_NonStringValueReturnsEmpty', () => {
    const node = { data: { config: { inputPath: 42 } } };
    expect(getNodePath(node, 'in')).toBe('');
  });

  // Story breadcrumbs-01, AC-3 + BDD: populated inputPath is returned as-is
  it('TestStory1_AC3_PopulatedInReturnsValue', () => {
    const node = { data: { config: { inputPath: '/abs/path' } } };
    expect(getNodePath(node, 'in')).toBe('/abs/path');
  });

  // Story breadcrumbs-01, AC-3 + BDD: populated outputPath is returned as-is
  it('TestStory1_AC3_PopulatedOutReturnsValue', () => {
    const node = { data: { config: { outputPath: '/abs/other' } } };
    expect(getNodePath(node, 'out')).toBe('/abs/other');
  });
});

// ─── breadcrumbs-02, AC-1 + AC-2: formatBreadcrumb formatter ─────────────────
// RED Phase: formatBreadcrumb is not yet exported from ./nodePath.ts.
// Dynamic import in beforeAll keeps existing getNodePath tests GREEN.

const EM_DASH = '\u2014';

describe('formatBreadcrumb — AC-2: path formatting', () => {
  let formatBreadcrumb: (path: unknown) => string;

  beforeAll(async () => {
    const mod = await import('./nodePath') as Record<string, unknown>;
    formatBreadcrumb = typeof mod.formatBreadcrumb === 'function'
      ? mod.formatBreadcrumb as (path: unknown) => string
      : () => { throw new Error('formatBreadcrumb not yet exported from nodePath.ts'); };
  });

  it('TestStory2_AC2_FormatBreadcrumb_EmptyReturnsEmDash', () => {
    expect(formatBreadcrumb('')).toBe(EM_DASH);
  });

  it('TestStory2_AC2_FormatBreadcrumb_NullReturnsEmDash', () => {
    expect(formatBreadcrumb(null)).toBe(EM_DASH);
  });

  it('TestStory2_AC2_FormatBreadcrumb_UndefinedReturnsEmDash', () => {
    expect(formatBreadcrumb(undefined)).toBe(EM_DASH);
  });

  it('TestStory2_AC2_FormatBreadcrumb_NonStringReturnsEmDash', () => {
    expect(formatBreadcrumb(42)).toBe(EM_DASH);
  });

  it('TestStory2_AC2_FormatBreadcrumb_AbsolutePathReturnsDotDotBasename', () => {
    expect(formatBreadcrumb('/abs/path/PRD.md')).toBe('.../PRD.md');
  });

  it('TestStory2_AC2_FormatBreadcrumb_SpacesInBasename', () => {
    expect(formatBreadcrumb('/abs/path/file with spaces.md')).toBe('.../file with spaces.md');
  });

  it('TestStory2_AC2_FormatBreadcrumb_NoSlashes', () => {
    expect(formatBreadcrumb('single')).toBe('.../single');
  });

  // Trailing slash → split('/').pop() returns '' (empty basename) → em-dash
  it('TestStory2_AC2_FormatBreadcrumb_TrailingSlashReturnsEmDash', () => {
    expect(formatBreadcrumb('/abs/path/')).toBe(EM_DASH);
  });
});
