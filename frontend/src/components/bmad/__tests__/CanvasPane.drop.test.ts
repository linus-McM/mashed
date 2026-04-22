// skills-cmd-02 AC-2: drop handler MUST ignore skill payloads.
//
// Contract — GREEN phase MUST export `handleMashedAssetDrop` from
// `frontend/src/components/bmad/canvasPaneDropHandler.js`. Signature:
//   handleMashedAssetDrop({ raw, position, currentNodes, now }) => Node[]
// Returns `currentNodes` unchanged when JSON.parse fails OR role !== 'command'.
// Otherwise appends a command node and returns the new array.
// CanvasPane.svelte's onDrop branch MUST delegate to this same function so
// the integration cannot drift from the unit contract.

import { describe, it, expect } from 'vitest';
import { handleMashedAssetDrop } from '../canvasPaneDropHandler.js';

import type { CanvasNode, Position } from '../../../types/workflow';

/**
 * Local mirror of the shape `handleMashedAssetDrop` emits. The JS source
 * writes these fields verbatim; asserting against an interface keeps the
 * contract visible here in TypeScript land.
 */
interface DroppedCommandNode extends CanvasNode {
  id: string;
  type: 'command';
  position: Position;
  data: {
    label: string;
    nodeType: 'command';
    status: 'pending';
    config: {
      commandName: string;
      commandPath: string;
      commandDescription: string;
    };
  };
}

const COMMAND_NAME = 'simplify';
const COMMAND_PATH = '/tmp/simplify.md';
const COMMAND_DESCRIPTION = 'Review recent changes';

describe('skills-cmd-02 AC-2: drop handler ignores skill payloads', () => {
  it('returns nodes unchanged when role === "skill"', () => {
    const before: CanvasNode[] = [
      { id: 'p-1', type: 'bmadProcess', position: { x: 0, y: 0 }, data: {} },
    ];
    const skillPayload = JSON.stringify({
      name: 'long-running-skill',
      path: '/tmp/skill.md',
      kind: 'skill',
      source: 'local',
      role: 'skill',
      description: 'should never be appended',
    });

    const after = handleMashedAssetDrop({
      raw: skillPayload,
      position: { x: 200, y: 100 },
      currentNodes: before,
      now: () => 1234567890,
    }) as CanvasNode[];

    expect(after).toEqual(before);
  });

  it('returns nodes unchanged when JSON.parse fails (defensive guard)', () => {
    const before: CanvasNode[] = [];
    const after = handleMashedAssetDrop({
      raw: '{not valid json',
      position: { x: 50, y: 50 },
      currentNodes: before,
      now: () => 1,
    }) as CanvasNode[];
    expect(after).toEqual(before);
  });

  it('appends a command node when role === "command" (positive control so AC-2 cannot pass by accident)', () => {
    const commandPayload = JSON.stringify({
      name: COMMAND_NAME,
      path: COMMAND_PATH,
      kind: 'command',
      source: 'local',
      role: 'command',
      description: COMMAND_DESCRIPTION,
    });

    const after = handleMashedAssetDrop({
      raw: commandPayload,
      position: { x: 300, y: 200 },
      currentNodes: [] as CanvasNode[],
      now: () => 9876543210,
    }) as DroppedCommandNode[];

    expect(after).toHaveLength(1);
    const node = after[0];
    expect(node.id).toMatch(/^cmd-/); // story Developer Notes: id = `cmd-${Date.now()}`
    expect(node.type).toBe('command');
    expect(node.position).toEqual({ x: 300, y: 200 });
    expect(node.data.label).toBe(COMMAND_NAME);
    expect(node.data.nodeType).toBe('command');
    expect(node.data.status).toBe('pending');
    expect(node.data.config).toEqual({
      commandName: COMMAND_NAME,
      commandPath: COMMAND_PATH,
      commandDescription: COMMAND_DESCRIPTION,
    });
  });
});
