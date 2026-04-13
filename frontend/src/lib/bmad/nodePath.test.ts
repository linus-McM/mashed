// frontend/src/lib/bmad/nodePath.test.ts
// Story breadcrumbs-01: Discovery + node-path data model
// Task 1 (RED phase): Tests for getNodePath helper (AC-3)
//
// RED Phase: These tests define expected behavior and MUST FAIL until
// the frontend engineer implements getNodePath in ./nodePath.ts

import { describe, it, expect } from 'vitest';
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
