// frontend/src/lib/bmad/__tests__/autoFill.test.ts
// Story breadcrumbs-06: Downstream File Loader empty-only auto-fill
//
// RED Phase: These tests define the contract for computeAutoFill().
// They MUST FAIL until the ui-engineer implements frontend/src/lib/bmad/autoFill.ts

import { describe, it, expect } from 'vitest';
import {
  computeAutoFill,
  type ArtifactEvent,
  type NodeCanvas,
  type EdgeCanvas,
  type AutoFillResult,
} from '../autoFill';

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function makeEvent(nodeId: string, paths: Record<string, string>): ArtifactEvent {
  return { execId: 'exec-1', nodeId, paths };
}

function makeNode(id: string, inputs: string[], inputPaths: Record<string, string> = {}): NodeCanvas {
  return {
    id,
    data: {
      process: { inputs },
      config: { inputPaths },
    },
  };
}

function makeEdge(source: string, target: string, opts: { sourceHandle?: string; targetHandle?: string } = {}): EdgeCanvas {
  return { source, target, ...opts };
}

// ---------------------------------------------------------------------------
// AC-1: Empty downstream slot → produces update entry
// ---------------------------------------------------------------------------

describe('computeAutoFill — AC-1: empty slot fills', () => {
  it('TestStory6_AC1_EmptySlot_ProducesUpdate', () => {
    const event = makeEvent('nodeA', { 'PRD.md': '/abs/PRD.md' });
    const nodes: NodeCanvas[] = [
      makeNode('nodeA', [], {}),
      makeNode('nodeB', ['PRD.md'], {}), // empty slot
    ];
    const edges: EdgeCanvas[] = [makeEdge('nodeA', 'nodeB')];

    const result: AutoFillResult = computeAutoFill(event, nodes, edges);

    expect(result.updates).toHaveLength(1);
    expect(result.updates[0]).toEqual({
      nodeId: 'nodeB',
      artifactName: 'PRD.md',
      path: '/abs/PRD.md',
    });
    expect(result.skipped).toHaveLength(0);
  });

  it('TestStory6_AC1_UndefinedSlot_TreatedAsEmpty', () => {
    // Node with no inputPaths object at all
    const event = makeEvent('nodeA', { 'PRD.md': '/abs/PRD.md' });
    const nodes: NodeCanvas[] = [
      makeNode('nodeA', [], {}),
      { id: 'nodeB', data: { process: { inputs: ['PRD.md'] }, config: {} } },
    ];
    const edges: EdgeCanvas[] = [makeEdge('nodeA', 'nodeB')];

    const result = computeAutoFill(event, nodes, edges);

    expect(result.updates).toHaveLength(1);
    expect(result.updates[0].artifactName).toBe('PRD.md');
    expect(result.updates[0].path).toBe('/abs/PRD.md');
  });
});

// ---------------------------------------------------------------------------
// AC-2: Non-empty slot → skipped:populated, NOT updated
// ---------------------------------------------------------------------------

describe('computeAutoFill — AC-2: populated slot preserved', () => {
  it('TestStory6_AC2_PopulatedSlot_ProducesSkipped', () => {
    const event = makeEvent('nodeA', { 'PRD.md': '/auto/PRD.md' });
    const nodes: NodeCanvas[] = [
      makeNode('nodeA', [], {}),
      makeNode('nodeB', ['PRD.md'], { 'PRD.md': '/user/custom.md' }),
    ];
    const edges: EdgeCanvas[] = [makeEdge('nodeA', 'nodeB')];

    const result = computeAutoFill(event, nodes, edges);

    expect(result.updates).toHaveLength(0);
    expect(result.skipped).toHaveLength(1);
    expect(result.skipped[0]).toEqual({
      nodeId: 'nodeB',
      artifactName: 'PRD.md',
      reason: 'populated',
    });
  });

  it('TestStory6_AC2_WhitespaceOnlyPath_TreatedAsPopulated', () => {
    // Whitespace-only is populated per story task 2 notes
    const event = makeEvent('nodeA', { 'PRD.md': '/auto/PRD.md' });
    const nodes: NodeCanvas[] = [
      makeNode('nodeA', [], {}),
      makeNode('nodeB', ['PRD.md'], { 'PRD.md': '   ' }),
    ];
    const edges: EdgeCanvas[] = [makeEdge('nodeA', 'nodeB')];

    const result = computeAutoFill(event, nodes, edges);

    expect(result.updates).toHaveLength(0);
    expect(result.skipped[0]?.reason).toBe('populated');
  });
});

// ---------------------------------------------------------------------------
// AC-3: Artifact not in paths map → no update, no skipped
// ---------------------------------------------------------------------------

describe('computeAutoFill — AC-3: unmapped artifact produces nothing', () => {
  it('TestStory6_AC3_ArtifactNotInPaths_NoOutput', () => {
    // paths has 'architecture.md' but downstream declares only 'PRD.md' —
    // target is traversed but correctly classified as not-an-input (Rule 4).
    // No update emitted; skipped records the diagnostic.
    const event = makeEvent('nodeA', { 'architecture.md': '/abs/arch.md' });
    const nodes: NodeCanvas[] = [
      makeNode('nodeA', [], {}),
      makeNode('nodeB', ['PRD.md'], {}),
    ];
    const edges: EdgeCanvas[] = [makeEdge('nodeA', 'nodeB')];

    const result = computeAutoFill(event, nodes, edges);

    expect(result.updates).toHaveLength(0);
    expect(result.skipped).toHaveLength(1);
    expect(result.skipped[0].reason).toBe('not-an-input');
  });

  it('TestStory6_AC3_EmptyPathsMap_ZeroOutput', () => {
    const event = makeEvent('nodeA', {});
    const nodes: NodeCanvas[] = [
      makeNode('nodeA', [], {}),
      makeNode('nodeB', ['PRD.md'], {}),
    ];
    const edges: EdgeCanvas[] = [makeEdge('nodeA', 'nodeB')];

    const result = computeAutoFill(event, nodes, edges);

    expect(result.updates).toHaveLength(0);
    expect(result.skipped).toHaveLength(0);
  });
});

// ---------------------------------------------------------------------------
// AC-5: Idempotency — same path already set → skipped:populated
// ---------------------------------------------------------------------------

describe('computeAutoFill — AC-5: idempotency', () => {
  it('TestStory6_AC5_SamePath_SkippedNotUpdated', () => {
    const path = '/abs/PRD.md';
    const event = makeEvent('nodeA', { 'PRD.md': path });
    const nodes: NodeCanvas[] = [
      makeNode('nodeA', [], {}),
      makeNode('nodeB', ['PRD.md'], { 'PRD.md': path }), // already set to the same value
    ];
    const edges: EdgeCanvas[] = [makeEdge('nodeA', 'nodeB')];

    const result = computeAutoFill(event, nodes, edges);

    expect(result.updates).toHaveLength(0);
    expect(result.skipped).toHaveLength(1);
    expect(result.skipped[0].reason).toBe('populated');
  });
});

// ---------------------------------------------------------------------------
// Fan-out: two downstream nodes both accept the artifact and both empty
// ---------------------------------------------------------------------------

describe('computeAutoFill — fan-out', () => {
  it('TestStory6_FanOut_TwoEmptyDownstreams_TwoUpdates', () => {
    const event = makeEvent('nodeA', { 'PRD.md': '/abs/PRD.md' });
    const nodes: NodeCanvas[] = [
      makeNode('nodeA', [], {}),
      makeNode('nodeB', ['PRD.md'], {}),
      makeNode('nodeC', ['PRD.md'], {}),
    ];
    const edges: EdgeCanvas[] = [
      makeEdge('nodeA', 'nodeB'),
      makeEdge('nodeA', 'nodeC'),
    ];

    const result = computeAutoFill(event, nodes, edges);

    expect(result.updates).toHaveLength(2);
    const ids = result.updates.map((u) => u.nodeId).sort();
    expect(ids).toEqual(['nodeB', 'nodeC']);
    result.updates.forEach((u) => {
      expect(u.artifactName).toBe('PRD.md');
      expect(u.path).toBe('/abs/PRD.md');
    });
    expect(result.skipped).toHaveLength(0);
  });
});

// ---------------------------------------------------------------------------
// Target node has NO matching input name → skipped:not-an-input
// ---------------------------------------------------------------------------

describe('computeAutoFill — not-an-input', () => {
  it('TestStory6_NoMatchingInput_SkippedNotAnInput', () => {
    const event = makeEvent('nodeA', { 'PRD.md': '/abs/PRD.md' });
    const nodes: NodeCanvas[] = [
      makeNode('nodeA', [], {}),
      makeNode('nodeB', ['architecture.md'], {}), // 'PRD.md' not in its inputs
    ];
    const edges: EdgeCanvas[] = [makeEdge('nodeA', 'nodeB')];

    const result = computeAutoFill(event, nodes, edges);

    expect(result.updates).toHaveLength(0);
    expect(result.skipped).toHaveLength(1);
    expect(result.skipped[0]).toEqual({
      nodeId: 'nodeB',
      artifactName: 'PRD.md',
      reason: 'not-an-input',
    });
  });

  it('TestStory6_NoInputsDeclared_SkippedNotAnInput', () => {
    const event = makeEvent('nodeA', { 'PRD.md': '/abs/PRD.md' });
    const nodes: NodeCanvas[] = [
      makeNode('nodeA', [], {}),
      { id: 'nodeB', data: { config: {} } }, // no process.inputs at all
    ];
    const edges: EdgeCanvas[] = [makeEdge('nodeA', 'nodeB')];

    const result = computeAutoFill(event, nodes, edges);

    expect(result.updates).toHaveLength(0);
    expect(result.skipped).toHaveLength(1);
    expect(result.skipped[0].reason).toBe('not-an-input');
  });
});

// ---------------------------------------------------------------------------
// Edge with sourceHandle set — breadcrumbs-08 contract: when non-empty,
// sourceHandle MUST equal the artifact key or the edge is skipped for that
// artifact. Matching-by-name sourceHandle still routes normally.
// ---------------------------------------------------------------------------

describe('computeAutoFill — sourceHandle gates by exact artifact name (breadcrumbs-08)', () => {
  it('TestStory6_SourceHandleSet_MatchesByArtifactName', () => {
    const event = makeEvent('nodeA', { 'PRD.md': '/abs/PRD.md' });
    const nodes: NodeCanvas[] = [
      makeNode('nodeA', [], {}),
      makeNode('nodeB', ['PRD.md'], {}),
    ];
    // sourceHandle = artifact name — edge routes the artifact as expected.
    const edges: EdgeCanvas[] = [makeEdge('nodeA', 'nodeB', { sourceHandle: 'PRD.md' })];

    const result = computeAutoFill(event, nodes, edges);

    expect(result.updates).toHaveLength(1);
    expect(result.updates[0].artifactName).toBe('PRD.md');
  });

  it('TestStory6_TargetHandleSet_MatchesByArtifactName', () => {
    const event = makeEvent('nodeA', { 'PRD.md': '/abs/PRD.md' });
    const nodes: NodeCanvas[] = [
      makeNode('nodeA', [], {}),
      makeNode('nodeB', ['PRD.md'], {}),
    ];
    const edges: EdgeCanvas[] = [makeEdge('nodeA', 'nodeB', { targetHandle: 'in-PRD.md' })];

    const result = computeAutoFill(event, nodes, edges);

    expect(result.updates).toHaveLength(1);
    expect(result.updates[0].artifactName).toBe('PRD.md');
  });
});

// ---------------------------------------------------------------------------
// Multiple artifacts in one event — each routed independently
// ---------------------------------------------------------------------------

describe('computeAutoFill — multiple artifacts per event', () => {
  it('TestStory6_MultipleArtifacts_EachRoutedIndependently', () => {
    const event = makeEvent('nodeA', {
      'PRD.md': '/abs/PRD.md',
      'architecture.md': '/abs/arch.md',
    });
    const nodes: NodeCanvas[] = [
      makeNode('nodeA', [], {}),
      // nodeB accepts PRD.md (empty) and architecture.md (populated)
      makeNode('nodeB', ['PRD.md', 'architecture.md'], {
        'PRD.md': '',
        'architecture.md': '/user/arch.md',
      }),
    ];
    const edges: EdgeCanvas[] = [makeEdge('nodeA', 'nodeB')];

    const result = computeAutoFill(event, nodes, edges);

    expect(result.updates).toHaveLength(1);
    expect(result.updates[0]).toEqual({
      nodeId: 'nodeB',
      artifactName: 'PRD.md',
      path: '/abs/PRD.md',
    });

    expect(result.skipped).toHaveLength(1);
    expect(result.skipped[0]).toEqual({
      nodeId: 'nodeB',
      artifactName: 'architecture.md',
      reason: 'populated',
    });
  });

  it('TestStory6_MultipleArtifacts_BothEmpty_TwoUpdates', () => {
    const event = makeEvent('nodeA', {
      'PRD.md': '/abs/PRD.md',
      'architecture.md': '/abs/arch.md',
    });
    const nodes: NodeCanvas[] = [
      makeNode('nodeA', [], {}),
      makeNode('nodeB', ['PRD.md', 'architecture.md'], {}),
    ];
    const edges: EdgeCanvas[] = [makeEdge('nodeA', 'nodeB')];

    const result = computeAutoFill(event, nodes, edges);

    expect(result.updates).toHaveLength(2);
    const names = result.updates.map((u) => u.artifactName).sort();
    expect(names).toEqual(['PRD.md', 'architecture.md']);
    expect(result.skipped).toHaveLength(0);
  });
});

// ---------------------------------------------------------------------------
// No edges from source node — no output
// ---------------------------------------------------------------------------

describe('computeAutoFill — no outgoing edges', () => {
  it('TestStory6_NoEdges_ZeroOutput', () => {
    const event = makeEvent('nodeA', { 'PRD.md': '/abs/PRD.md' });
    const nodes: NodeCanvas[] = [
      makeNode('nodeA', [], {}),
      makeNode('nodeB', ['PRD.md'], {}),
    ];
    const edges: EdgeCanvas[] = []; // no edges at all

    const result = computeAutoFill(event, nodes, edges);

    expect(result.updates).toHaveLength(0);
    expect(result.skipped).toHaveLength(0);
  });
});
