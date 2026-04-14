// frontend/src/lib/bmad/multiFileEntries.ts
// Story breadcrumbs-08: pure helpers for MultiFileLoader entry lists.
// Entries serialise to/from a JSON string in node.data.config.entries.

export interface MultiFileEntry {
  label: string;
  path: string;
}

export function parseEntries(json: string): MultiFileEntry[] {
  if (typeof json !== 'string' || json === '') return [];
  let raw: unknown;
  try {
    raw = JSON.parse(json);
  } catch {
    return [];
  }
  if (!Array.isArray(raw)) return [];
  const out: MultiFileEntry[] = [];
  for (const item of raw) {
    if (typeof item !== 'object' || item === null) continue;
    const rec = item as Record<string, unknown>;
    const label = typeof rec.label === 'string' ? rec.label : '';
    const path = typeof rec.path === 'string' ? rec.path : '';
    out.push({ label, path });
  }
  return out;
}

export function stringifyEntries(entries: MultiFileEntry[]): string {
  return JSON.stringify(entries);
}

export function entryLabelFor(entry: MultiFileEntry, index: number): string {
  if (entry && typeof entry.label === 'string' && entry.label !== '') {
    return entry.label;
  }
  return `file[${index}]`;
}

export function hasDuplicateLabels(entries: MultiFileEntry[]): Set<string> {
  const counts = new Map<string, number>();
  for (const e of entries) {
    const label = e?.label;
    if (typeof label !== 'string' || label === '') continue;
    counts.set(label, (counts.get(label) ?? 0) + 1);
  }
  const dups = new Set<string>();
  for (const [label, count] of counts) {
    if (count >= 2) dups.add(label);
  }
  return dups;
}
