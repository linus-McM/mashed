// frontend/src/lib/bmad/nodePath.ts
// Story breadcrumbs-01, AC-3: safe accessor for node input/output paths.
// Story breadcrumbs-04: optional `artifactName` selects a map-valued entry
// (config.inputPaths / outputPaths) with legacy single-path fallback.

export type PathDirection = 'in' | 'out';

interface NodeShape {
  data?: {
    process?: {
      inputs?: unknown;
      outputs?: unknown;
    };
    config?: Record<string, unknown>;
  };
}

function readConfig(node: unknown): Record<string, unknown> | null {
  if (typeof node !== 'object' || node === null) return null;
  const data = (node as NodeShape).data;
  if (typeof data !== 'object' || data === null) return null;
  const config = (data as { config?: unknown }).config;
  if (typeof config !== 'object' || config === null) return null;
  return config as Record<string, unknown>;
}

function declaredCount(node: unknown, dir: PathDirection): number {
  if (typeof node !== 'object' || node === null) return 0;
  const process = (node as NodeShape).data?.process;
  if (!process) return 0;
  const list = dir === 'in' ? process.inputs : process.outputs;
  return Array.isArray(list) ? list.length : 0;
}

export function getNodePath(
  node: unknown,
  dir: PathDirection,
  artifactName?: string,
): string {
  const config = readConfig(node);
  if (!config) return '';
  const mapKey = dir === 'in' ? 'inputPaths' : 'outputPaths';
  const legacyKey = dir === 'in' ? 'inputPath' : 'outputPath';

  if (typeof artifactName === 'string') {
    const map = config[mapKey];
    if (typeof map === 'object' && map !== null) {
      const mapped = (map as Record<string, unknown>)[artifactName];
      if (typeof mapped === 'string' && mapped !== '') return mapped;
      return '';
    }
    const legacy = config[legacyKey];
    if (typeof legacy === 'string' && legacy !== '' && declaredCount(node, dir) === 1) {
      return legacy;
    }
    return '';
  }

  const value = config[legacyKey];
  return typeof value === 'string' ? value : '';
}

export function formatBreadcrumb(path: unknown): string {
  if (typeof path !== 'string' || path === '') return '\u2014';
  const base = path.split('/').pop();
  if (!base) return '\u2014';
  return '.../' + base;
}
