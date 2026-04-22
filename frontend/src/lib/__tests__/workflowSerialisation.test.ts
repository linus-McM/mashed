// skills-cmd-03 AC-1 / AC-2 / AC-4: canvas <-> WorkflowDef serialisation contracts.
//
// These tests lock the three "no code change expected but MUST be guarded
// by tests" checkpoints from the story's Developer Notes:
//
//   AC-1: save + reload preserves command nodes (type AND config keys)
//   AC-2: snapshotCanvas distinguishes command from process at the same ID
//   AC-4: workflowNodesToCanvasNodes is type-agnostic — positions for both
//         node kinds survive a round trip without any filtering.
//
// The pure serialisation module was extracted from WorkflowBuilder.svelte
// explicitly to make these assertions testable without mounting the Svelte
// component. See workflowSerialisation.ts for the extraction rationale.

import { describe, it, expect } from 'vitest';
import {
  snapshotCanvas,
  canvasNodesToWorkflowNodes,
  workflowNodesToCanvasNodes,
  workflowEdgesToCanvasEdges,
  serialise,
  deserialise,
  type SavedWorkflowNode,
} from '../workflowSerialisation';
import type { CanvasEdge, CanvasNode, Workflow } from '../../types/workflow';

/** Build a realistic mixed canvas with one process node, one command node,
 *  and a single edge wiring them together. Matches the shape WorkflowBuilder
 *  produces after a process drop + a command drop + a manual connect. */
function makeMixedCanvas(): { nodes: CanvasNode[]; edges: CanvasEdge[] } {
  const processNode: CanvasNode = {
    id: 'proc-A',
    type: 'bmadProcess',
    position: { x: 100, y: 150 },
    data: {
      label: 'Do Thing',
      processId: 'bmad-dev-story',
      nodeType: '',
      status: 'pending',
      config: { storyId: 'S-1' },
      tmuxTarget: '',
      storyId: 'S-1',
    },
  };
  const commandNode: CanvasNode = {
    id: 'cmd-B',
    type: 'command',
    position: { x: 400, y: 250 },
    data: {
      label: 'simplify',
      processId: '',
      nodeType: 'command',
      status: 'pending',
      config: {
        commandName: 'simplify',
        commandPath: '/tmp/simplify.md',
        commandDescription: 'Review recent changes',
      },
      tmuxTarget: '',
      storyId: '',
    },
  };
  const edge: CanvasEdge = {
    id: 'edge-1',
    source: 'proc-A',
    target: 'cmd-B',
    sourceHandle: undefined,
    targetHandle: undefined,
  };
  return { nodes: [processNode, commandNode], edges: [edge] };
}

const noopInferEdgeLabel = (): string => '';

describe('skills-cmd-03 AC-1: save + reload preserves command nodes', () => {
  it('round-trips a mixed canvas through canvas->workflow->canvas with types intact', () => {
    const { nodes, edges } = makeMixedCanvas();

    // Canvas -> WorkflowDef shape (what saveWorkflow hands to Wails).
    const wfNodes = canvasNodesToWorkflowNodes(nodes);
    const wfEdges = edges.map((e) => ({
      id: e.id,
      source: e.source,
      target: e.target,
      sourceHandle: e.sourceHandle || '',
      targetHandle: e.targetHandle || '',
    }));

    // The saved WorkflowNode for the command must carry nodeType === 'command'.
    const savedCmd = wfNodes.find((n) => n.id === 'cmd-B');
    if (!savedCmd) throw new Error('expected saved command node');
    expect(savedCmd.nodeType).toBe('command');
    expect(savedCmd.config.commandName).toBe('simplify');
    expect(savedCmd.config.commandPath).toBe('/tmp/simplify.md');
    expect(savedCmd.config.commandDescription).toBe('Review recent changes');

    // WorkflowDef -> canvas (what loadNodesEdges produces on reload).
    const processes = [{ id: 'bmad-dev-story', name: 'Do Thing' }];
    const restored = workflowNodesToCanvasNodes(wfNodes, processes);

    const restoredProc = restored.find((n) => n.id === 'proc-A');
    const restoredCmd = restored.find((n) => n.id === 'cmd-B');
    if (!restoredProc) throw new Error('expected restored process node');
    if (!restoredCmd) throw new Error('expected restored command node');

    // AC-1: both node types reappear with their original svelte-flow types.
    expect(restoredProc.type).toBe('bmadProcess');
    expect(restoredCmd.type).toBe('command');

    // AC-1: all three command config keys survive untouched.
    expect(restoredCmd.data?.config).toEqual({
      commandName: 'simplify',
      commandPath: '/tmp/simplify.md',
      commandDescription: 'Review recent changes',
    });

    // Positions preserved for both node kinds.
    expect(restoredProc.position).toEqual({ x: 100, y: 150 });
    expect(restoredCmd.position).toEqual({ x: 400, y: 250 });

    // Edge round-trips too — the edge connecting a process to a command
    // is the smallest mixed-workflow case and MUST not be dropped.
    const restoredEdges = workflowEdgesToCanvasEdges(wfEdges, noopInferEdgeLabel);
    expect(restoredEdges).toHaveLength(1);
    expect(restoredEdges[0].source).toBe('proc-A');
    expect(restoredEdges[0].target).toBe('cmd-B');
  });

  it('legacy workflows with blank nodeType still fall back to bmadProcess', () => {
    // Regression guard: the ternary in workflowNodesToCanvasNodes MUST
    // keep legacy workflows rendering correctly. An empty string (or the
    // legacy literal `'process'`) means "old-style process node".
    const legacy: SavedWorkflowNode[] = [
      { id: 'x', nodeType: '', position: { x: 0, y: 0 }, label: 'L' },
      { id: 'y', nodeType: 'process', position: { x: 10, y: 20 }, label: 'M' },
    ];
    const restored = workflowNodesToCanvasNodes(legacy, []);
    expect(restored[0].type).toBe('bmadProcess');
    expect(restored[1].type).toBe('bmadProcess');
  });
});

describe('skills-cmd-03 AC-2: snapshotCanvas detects node-type drift', () => {
  it('distinguishes a command node from a process node at the same ID', () => {
    const commandAtN1: CanvasNode[] = [
      {
        id: 'n1',
        type: 'command',
        position: { x: 200, y: 200 },
        data: { label: 'simplify', nodeType: 'command', config: { commandName: 'simplify' } },
      },
    ];
    const processAtN1: CanvasNode[] = [
      {
        id: 'n1',
        type: 'bmadProcess',
        position: { x: 200, y: 200 },
        data: { label: 'Do Thing', processId: 'p1', config: {} },
      },
    ];

    const h1 = snapshotCanvas(commandAtN1, [], 'same-name');
    const h2 = snapshotCanvas(processAtN1, [], 'same-name');

    // Core contract: swapping node type at the same ID/position MUST
    // produce a different hash. Without this, the dirty detector would
    // silently miss a command<->process swap and a save-on-leave would
    // write the wrong type to disk.
    expect(h1).not.toEqual(h2);
  });

  it('produces a stable hash for an identical canvas', () => {
    const canvas = makeMixedCanvas();
    const h1 = snapshotCanvas(canvas.nodes, canvas.edges, 'wf');
    const h2 = snapshotCanvas(canvas.nodes, canvas.edges, 'wf');
    expect(h1).toEqual(h2);
  });
});

describe('skills-cmd-03 AC-4: workflowNodesToCanvasNodes is type-agnostic', () => {
  it('preserves positions for mixed process + command nodes without filtering', () => {
    const wfNodes: SavedWorkflowNode[] = [
      {
        id: 'proc-A',
        nodeType: '',
        processId: 'bmad-dev-story',
        label: 'Do Thing',
        position: { x: 111, y: 222 },
        status: 'complete',
        config: {},
      },
      {
        id: 'cmd-B',
        nodeType: 'command',
        processId: '',
        label: 'simplify',
        position: { x: 333, y: 444 },
        status: 'pending',
        config: { commandName: 'simplify' },
      },
      {
        id: 'cond-C',
        nodeType: 'condition',
        processId: '',
        label: 'if',
        position: { x: 555, y: 666 },
        status: 'pending',
        config: {},
      },
    ];

    const restored = workflowNodesToCanvasNodes(wfNodes, [
      { id: 'bmad-dev-story', name: 'Do Thing' },
    ]);

    // AC-4: no filtering — all three nodes must survive regardless of type.
    expect(restored).toHaveLength(3);

    // Positions preserved verbatim for every type.
    expect(restored[0].position).toEqual({ x: 111, y: 222 });
    expect(restored[1].position).toEqual({ x: 333, y: 444 });
    expect(restored[2].position).toEqual({ x: 555, y: 666 });

    // Types survive the discriminator ternary for ALL non-process values.
    expect(restored[0].type).toBe('bmadProcess');
    expect(restored[1].type).toBe('command');
    expect(restored[2].type).toBe('condition');

    // Status and config both types pass through the same code path.
    expect(restored[0].data?.status).toBe('complete');
    expect(restored[1].data?.status).toBe('pending');
    expect(restored[1].data?.config?.commandName).toBe('simplify');
  });
});

/**
 * Canonical workflow fixture used by the Scenario 1 round-trip assertion.
 *
 * Using a single `as Workflow` at the boundary is intentional — the upstream
 * `bmad.WorkflowDef` is a Wails-generated class, and we want a plain object
 * literal for JSON round-trip semantics. There is no narrowing to do here
 * (the shape is constructed in-place), so the cast is safe.
 */
function makeWorkflowFixture(): Workflow {
  const plain = {
    id: 'wf-1',
    name: 'Fixture Workflow',
    description: 'A canonical workflow used for serialise/deserialise tests.',
    repoPath: '/tmp/repo',
    nodes: [
      {
        id: 'proc-A',
        processId: 'bmad-dev-story',
        label: 'Do Thing',
        position: { x: 100, y: 150 },
        status: 'pending',
        config: { storyId: 'S-1' },
        tmuxTarget: '',
        storyId: 'S-1',
        nodeType: '',
      },
    ],
    edges: [],
    isTemplate: false,
    templateId: '',
    createdAt: '2026-04-22T00:00:00Z',
    updatedAt: '2026-04-22T00:00:00Z',
  };
  return plain as unknown as Workflow;
}

describe('BDD Scenario 1: serialise ↔ deserialise round-trip is idempotent', () => {
  it('serialise(fixture) → deserialise(...) → serialise(...) is byte-identical', () => {
    const fixture = makeWorkflowFixture();
    const first = serialise(fixture);
    const parsed = deserialise(first);
    expect(parsed).not.toBeNull();
    if (parsed === null) throw new Error('expected deserialise to succeed');
    const second = serialise(parsed);
    expect(second).toEqual(first);
  });

  it('intermediate deserialise result satisfies the Workflow type shape', () => {
    const fixture = makeWorkflowFixture();
    const parsed = deserialise(serialise(fixture));
    expect(parsed).not.toBeNull();
    if (parsed === null) throw new Error('expected deserialise to succeed');
    // The type narrowing in deserialise guarantees these fields exist.
    expect(typeof parsed.id).toBe('string');
    expect(typeof parsed.name).toBe('string');
    expect(typeof parsed.isTemplate).toBe('boolean');
    expect(Array.isArray(parsed.nodes)).toBe(true);
    expect(Array.isArray(parsed.edges)).toBe(true);
  });
});

describe('BDD Scenario 4: deserialise rejects malformed input without throwing', () => {
  it('returns null for "{}" (missing required fields)', () => {
    expect(() => deserialise('{}')).not.toThrow();
    expect(deserialise('{}')).toBeNull();
  });

  it('returns null for invalid JSON without throwing', () => {
    expect(() => deserialise('not-json')).not.toThrow();
    expect(deserialise('not-json')).toBeNull();
  });

  it('returns null for an empty string', () => {
    expect(deserialise('')).toBeNull();
  });

  it('returns null when a required field has the wrong type', () => {
    const bad = JSON.stringify({
      id: 123, // should be string
      name: 'x',
      description: '',
      isTemplate: false,
      nodes: [],
      edges: [],
      createdAt: '',
      updatedAt: '',
    });
    expect(deserialise(bad)).toBeNull();
  });

  it('returns null for a JSON array at the top level', () => {
    expect(deserialise('[]')).toBeNull();
  });

  it('returns null for a JSON null', () => {
    expect(deserialise('null')).toBeNull();
  });
});
