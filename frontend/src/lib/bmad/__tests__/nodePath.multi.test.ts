// frontend/src/lib/bmad/__tests__/nodePath.multi.test.ts
// Story breadcrumbs-04: Multi-input breadcrumb rendering
// Task 1 RED phase: getNodePath(node, dir, artifactName?) — per-artifact path lookup
//
// RED Phase: These tests MUST FAIL until the frontend engineer extends
// getNodePath with the optional 3rd `artifactName` parameter that reads from
// data.config.inputPaths / outputPaths maps.

import { describe, it, expect } from 'vitest';
import { getNodePath } from '../nodePath';

// getNodePath is typed (node: unknown, dir: PathDirection). The 3rd arg doesn't
// exist yet, so TS will complain — cast through a permissive signature to let
// the test compile while still binding to the real export.
const getNodePathX = getNodePath as unknown as (
  node: unknown,
  dir: 'in' | 'out',
  artifactName?: string,
) => string;

describe('getNodePath — breadcrumbs-04: per-artifact path lookup', () => {
  // AC-2: map present, key present → return mapped value
  it('TestStory4_AC2_WithArtifactName_ReadsFromMap', () => {
    const node = {
      data: {
        process: { inputs: ['prd', 'sprint-status'] },
        config: {
          inputPaths: {
            prd: '/a/PRD.md',
            'sprint-status': '/a/sprint.yaml',
          },
        },
      },
    };
    expect(getNodePathX(node, 'in', 'prd')).toBe('/a/PRD.md');
    expect(getNodePathX(node, 'in', 'sprint-status')).toBe('/a/sprint.yaml');
  });

  // AC-2 edge: map present, key absent → return empty (no accidental fallback)
  it('TestStory4_AC2_WithArtifactName_MissingKey_ReturnsEmpty', () => {
    const node = {
      data: {
        process: { inputs: ['prd', 'sprint-status'] },
        config: {
          inputPaths: { prd: '/a/PRD.md' },
        },
      },
    };
    expect(getNodePathX(node, 'in', 'sprint-status')).toBe('');
  });

  // AC-3: legacy singleton — no map, single declared input, legacy inputPath set,
  // caller passes artifactName → helper returns legacy value.
  it('TestStory4_AC3_LegacyFallback_SingleInput_UsesLegacyKey', () => {
    const node = {
      data: {
        process: { inputs: ['only'] },
        config: { inputPath: '/x.md' },
      },
    };
    expect(getNodePathX(node, 'in', 'only')).toBe('/x.md');
  });

  // AC-3 safety: no map, legacy set, but MULTIPLE declared inputs — do NOT use
  // legacy value (would be ambiguous which input it belongs to). Return empty.
  it('TestStory4_AC3_NoMapNoLegacyFallbackWithMultipleInputs_ReturnsEmpty', () => {
    const node = {
      data: {
        process: { inputs: ['prd', 'sprint-status'] },
        config: { inputPath: '/ambiguous.md' }, // legacy, but 2 inputs declared
      },
    };
    expect(getNodePathX(node, 'in', 'prd')).toBe('');
    expect(getNodePathX(node, 'in', 'sprint-status')).toBe('');
  });

  // Symmetry: outputPaths map must behave identically to inputPaths.
  it('TestStory4_AC2_OutputPaths_SymmetricBehavior', () => {
    const node = {
      data: {
        process: { outputs: ['draft', 'final'] },
        config: {
          outputPaths: {
            draft: '/d.md',
            final: '/f.md',
          },
        },
      },
    };
    expect(getNodePathX(node, 'out', 'draft')).toBe('/d.md');
    expect(getNodePathX(node, 'out', 'final')).toBe('/f.md');
  });

  // Output legacy fallback, symmetrical to input legacy fallback.
  it('TestStory4_AC3_OutputLegacyFallback_SingleOutput_UsesLegacyKey', () => {
    const node = {
      data: {
        process: { outputs: ['result'] },
        config: { outputPath: '/out.md' },
      },
    };
    expect(getNodePathX(node, 'out', 'result')).toBe('/out.md');
  });

  // AC-3 backwards compat: calling without artifactName preserves existing
  // single-path reader behavior (required by breadcrumbs-01 AC-3 tests).
  it('TestStory4_AC3_NoArtifactName_PreservesLegacyReader_In', () => {
    const node = { data: { config: { inputPath: '/legacy-in.md' } } };
    expect(getNodePathX(node, 'in')).toBe('/legacy-in.md');
  });

  it('TestStory4_AC3_NoArtifactName_PreservesLegacyReader_Out', () => {
    const node = { data: { config: { outputPath: '/legacy-out.md' } } };
    expect(getNodePathX(node, 'out')).toBe('/legacy-out.md');
  });

  // Defensive: non-string map value must not be returned as a path.
  it('TestStory4_MapNonStringValue_ReturnsEmpty', () => {
    const node = {
      data: {
        process: { inputs: ['prd'] },
        config: { inputPaths: { prd: 42 } },
      },
    };
    expect(getNodePathX(node, 'in', 'prd')).toBe('');
  });

  // Defensive: map is not an object (e.g., null) → empty string, no throw.
  it('TestStory4_MapNull_ReturnsEmpty', () => {
    const node = {
      data: {
        process: { inputs: ['prd'] },
        config: { inputPaths: null },
      },
    };
    expect(getNodePathX(node, 'in', 'prd')).toBe('');
  });
});
