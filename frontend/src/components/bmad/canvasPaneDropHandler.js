// Pure drop-handler for mashed-asset payloads. Extracted from
// CanvasPane.svelte so the contract can be unit-tested in isolation
// (no Svelte, no stores, no DOM) and so the integration cannot drift.
//
// Returns `currentNodes` unchanged on JSON.parse failure or when the
// payload's `role` is not 'command' — the sidebar already blocks skill
// drags via effectAllowed='none', but a stray payload must never crash
// the canvas or insert garbage state.

export function handleMashedAssetDrop({ raw, position, currentNodes, now = Date.now }) {
  let asset;
  try {
    asset = JSON.parse(raw);
  } catch {
    return currentNodes;
  }
  if (!asset || asset.role !== 'command') return currentNodes;

  const newNode = {
    id: `cmd-${now()}`,
    type: 'command',
    position,
    data: {
      label: asset.name,
      nodeType: 'command',
      config: {
        commandName: asset.name,
        commandPath: asset.path,
        commandDescription: asset.description || '',
      },
      status: 'pending',
    },
  };
  return [...currentNodes, newNode];
}
