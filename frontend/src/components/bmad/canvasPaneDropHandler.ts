// Pure drop-handler for mashed-asset payloads. Extracted from
// CanvasPane.svelte so the contract can be unit-tested in isolation
// (no Svelte, no stores, no DOM) and so the integration cannot drift.
//
// Returns `currentNodes` unchanged on JSON.parse failure or when the
// payload's `role` is not 'command' — the sidebar already blocks skill
// drags via effectAllowed='none', but a stray payload must never crash
// the canvas or insert garbage state.

import type { CanvasNode, Position } from '../../types/workflow';

/**
 * Subset of `bmad.MashedAssetInfo` required to construct a command-node
 * drop. Typed permissively (all fields optional except `role`, which is
 * narrowed by the guard below) so a malformed payload never crashes the
 * canvas — the `role !== 'command'` check rejects anything unexpected.
 */
interface MashedAssetPayload {
  role?: string;
  name?: string;
  path?: string;
  description?: string;
}

export interface HandleMashedAssetDropArgs {
  raw: string;
  position: Position;
  currentNodes: CanvasNode[];
  now?: () => number;
}

export function handleMashedAssetDrop({
  raw,
  position,
  currentNodes,
  now = Date.now,
}: HandleMashedAssetDropArgs): CanvasNode[] {
  let asset: MashedAssetPayload;
  try {
    asset = JSON.parse(raw) as MashedAssetPayload;
  } catch {
    return currentNodes;
  }
  if (!asset || asset.role !== 'command') return currentNodes;

  const newNode: CanvasNode = {
    id: `cmd-${now()}`,
    type: 'command',
    position,
    data: {
      label: asset.name ?? '',
      nodeType: 'command',
      config: {
        commandName: asset.name ?? '',
        commandPath: asset.path ?? '',
        commandDescription: asset.description ?? '',
      },
      status: 'pending',
    },
  };
  return [...currentNodes, newNode];
}
