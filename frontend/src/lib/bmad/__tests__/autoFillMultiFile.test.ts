// Story breadcrumbs-08 RED tests for sourceHandle-aware auto-fill.
// MultiFileLoader's outgoing edges carry sourceHandle = <label or file[N]>.
// computeAutoFill must restrict each artifact to the edges whose
// sourceHandle matches that artifact name (so a brief edge does not pump
// the wrong path into a downstream node).
//
// These tests will fail until computeAutoFill in autoFill.ts honours
// edge.sourceHandle. The current GREEN code from breadcrumbs-06 ignores it.

import { describe, it, expect } from 'vitest';
import { computeAutoFill, type ArtifactEvent } from '../autoFill';

const event = (paths: Record<string, string>): ArtifactEvent => ({
  execId: 'exec-1',
  nodeId: 'mfl-1',
  paths,
});

describe('computeAutoFill — sourceHandle gating (breadcrumbs-08 AC-5)', () => {
  it('routes artifact to ONLY the edges whose sourceHandle matches', () => {
    const nodes = [
      { id: 'mfl-1', data: {} },
      {
        id: 'down-A',
        data: {
          process: { inputs: ['brief'] },
          config: { inputPaths: {} },
        },
      },
      {
        id: 'down-B',
        data: {
          process: { inputs: ['brief'] },
          config: { inputPaths: {} },
        },
      },
    ];
    const edges = [
      { source: 'mfl-1', target: 'down-A', sourceHandle: 'brief' },
      { source: 'mfl-1', target: 'down-B', sourceHandle: 'plan' },
    ];

    const result = computeAutoFill(
      event({ brief: '/x/a.md', plan: '/x/p.md' }),
      nodes,
      edges,
    );

    // Only down-A receives "brief"; down-B's edge is sourceHandle="plan",
    // and down-B doesn't even declare "plan" as an input, so it's skipped.
    const updateTargets = result.updates.map((u) => ({ id: u.nodeId, name: u.artifactName }));
    expect(updateTargets).toContainEqual({ id: 'down-A', name: 'brief' });
    expect(updateTargets).not.toContainEqual({ id: 'down-B', name: 'brief' });
  });

  it('falls back to name-only matching when sourceHandle is absent (legacy)', () => {
    // Pre-MultiFileLoader process-to-process edges have no sourceHandle;
    // the breadcrumbs-06 contract must keep working unchanged. Source id
    // matches event.nodeId ('mfl-1') so the edge is in 'outgoing'.
    const nodes = [
      { id: 'mfl-1', data: {} },
      {
        id: 'dst',
        data: { process: { inputs: ['PRD.md'] }, config: { inputPaths: {} } },
      },
    ];
    const edges = [{ source: 'mfl-1', target: 'dst' }];
    const result = computeAutoFill(event({ 'PRD.md': '/abs/PRD.md' }), nodes, edges);
    expect(result.updates).toContainEqual({
      nodeId: 'dst',
      artifactName: 'PRD.md',
      path: '/abs/PRD.md',
    });
  });

  it('ignores artifact when sourceHandle present but does not match any path key', () => {
    const nodes = [
      { id: 'mfl-1', data: {} },
      { id: 'dst', data: { process: { inputs: ['brief'] }, config: { inputPaths: {} } } },
    ];
    const edges = [{ source: 'mfl-1', target: 'dst', sourceHandle: 'unrelated' }];

    const result = computeAutoFill(event({ brief: '/x/a.md' }), nodes, edges);
    // No update — the edge's handle doesn't match the artifact key.
    expect(result.updates).toHaveLength(0);
  });
});
