// frontend/src/lib/bmad/autoFill.ts
// Story breadcrumbs-06: Downstream File Loader empty-only auto-fill.
//
// Pure helper — NO Svelte or Wails imports. When an upstream node emits
// resolved artifact OutputPaths, compute which downstream (input-accepting)
// nodes should have those paths written into their empty InputPaths slots,
// and which should be skipped.

export interface ArtifactEvent {
  execId?: string;
  nodeId: string;
  paths: Record<string, string>;
}

export interface NodeCanvas {
  id: string;
  // Permissive shape — the caller's data object may carry arbitrary extra
  // keys beyond `process` / `config`. Every field is optional so the
  // WorkflowBuilder `CanvasNode` (and test fixtures) flow through without
  // a cast.
  data?: {
    process?: { inputs?: string[] } | null;
    config?: Record<string, unknown>;
  };
}

export interface EdgeCanvas {
  source: string;
  target: string;
  sourceHandle?: string;
  targetHandle?: string;
}

export interface AutoFillResult {
  updates: Array<{ nodeId: string; artifactName: string; path: string }>;
  skipped: Array<{ nodeId: string; artifactName: string; reason: 'populated' | 'not-an-input' }>;
}

export function computeAutoFill(
  event: ArtifactEvent,
  nodes: NodeCanvas[],
  edges: EdgeCanvas[],
): AutoFillResult {
  const result: AutoFillResult = { updates: [], skipped: [] };

  const paths = event?.paths;
  if (!paths || typeof paths !== 'object') return result;

  const outgoing = edges.filter((e) => e.source === event.nodeId);
  if (outgoing.length === 0) return result;

  const nodeById = new Map<string, NodeCanvas>();
  for (const n of nodes) nodeById.set(n.id, n);

  for (const [artifactName, incomingPath] of Object.entries(paths)) {
    for (const edge of outgoing) {
      // breadcrumbs-08: when an edge carries a non-empty sourceHandle,
      // it addresses a single named output on the upstream node — route
      // only that artifact through this edge. Legacy process-to-process
      // edges omit sourceHandle and keep name-based matching.
      if (typeof edge.sourceHandle === 'string' && edge.sourceHandle !== '') {
        if (artifactName !== edge.sourceHandle) continue;
      }

      const target = nodeById.get(edge.target);
      if (!target) continue;

      const inputs = target.data?.process?.inputs;
      if (!Array.isArray(inputs) || !inputs.includes(artifactName)) {
        result.skipped.push({
          nodeId: target.id,
          artifactName,
          reason: 'not-an-input',
        });
        continue;
      }

      const config = target.data?.config;
      const inputPaths =
        config && typeof config === 'object'
          ? (config as { inputPaths?: Record<string, string> }).inputPaths
          : undefined;
      const existing = inputPaths?.[artifactName];
      const isEmpty = existing === undefined || existing === null || existing === '';

      if (isEmpty) {
        result.updates.push({
          nodeId: target.id,
          artifactName,
          path: incomingPath,
        });
      } else {
        result.skipped.push({
          nodeId: target.id,
          artifactName,
          reason: 'populated',
        });
      }
    }
  }

  return result;
}
