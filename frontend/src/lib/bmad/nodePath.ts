// frontend/src/lib/bmad/nodePath.ts
// Story breadcrumbs-01, AC-3: safe accessor for node input/output paths.

export type PathDirection = 'in' | 'out';

export function getNodePath(node: unknown, dir: PathDirection): string {
  if (typeof node !== 'object' || node === null) return '';
  const data = (node as { data?: unknown }).data;
  if (typeof data !== 'object' || data === null) return '';
  const config = (data as { config?: unknown }).config;
  if (typeof config !== 'object' || config === null) return '';
  const key = dir === 'in' ? 'inputPath' : 'outputPath';
  const value = (config as Record<string, unknown>)[key];
  return typeof value === 'string' ? value : '';
}

export function formatBreadcrumb(path: unknown): string {
  if (typeof path !== 'string' || path === '') return '\u2014';
  const base = path.split('/').pop();
  if (!base) return '\u2014';
  return '.../' + base;
}
