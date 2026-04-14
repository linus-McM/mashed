// Story breadcrumbs-08 RED tests for the MultiFileEntries pure helper.
// The module `frontend/src/lib/bmad/multiFileEntries.ts` does NOT exist yet;
// these tests MUST fail with import / "not implemented" errors until the
// GREEN phase ships the implementation.

import { describe, it, expect } from 'vitest';
import {
  parseEntries,
  stringifyEntries,
  entryLabelFor,
  hasDuplicateLabels,
  type MultiFileEntry,
} from '../multiFileEntries';

describe('parseEntries — JSON tolerant decode', () => {
  it('returns [] for malformed JSON', () => {
    expect(parseEntries('not-json')).toEqual([]);
  });

  it('returns [] for empty string', () => {
    expect(parseEntries('')).toEqual([]);
  });

  it('returns [] for undefined / null', () => {
    expect(parseEntries(undefined as unknown as string)).toEqual([]);
    expect(parseEntries(null as unknown as string)).toEqual([]);
  });

  it('returns [] for non-array JSON (object, number)', () => {
    expect(parseEntries('{"label":"x","path":"/a"}')).toEqual([]);
    expect(parseEntries('42')).toEqual([]);
  });

  it('parses a valid entries array', () => {
    const json = '[{"label":"brief","path":"/x/a.md"},{"label":"","path":"/x/b.md"}]';
    expect(parseEntries(json)).toEqual([
      { label: 'brief', path: '/x/a.md' },
      { label: '', path: '/x/b.md' },
    ]);
  });

  it('coerces missing fields to empty strings', () => {
    const json = '[{"path":"/x/a.md"},{"label":"only-label"}]';
    expect(parseEntries(json)).toEqual([
      { label: '', path: '/x/a.md' },
      { label: 'only-label', path: '' },
    ]);
  });

  it('drops non-object entries inside the array', () => {
    const json = '[{"label":"a","path":"/x"}, 42, null, "skip"]';
    expect(parseEntries(json)).toEqual([{ label: 'a', path: '/x' }]);
  });
});

describe('stringifyEntries — round trip', () => {
  it('serialises and re-parses identically', () => {
    const entries: MultiFileEntry[] = [
      { label: 'brief', path: '/x/a.md' },
      { label: '', path: '/x/b.md' },
    ];
    expect(parseEntries(stringifyEntries(entries))).toEqual(entries);
  });

  it('serialises an empty array', () => {
    expect(stringifyEntries([])).toBe('[]');
  });
});

describe('entryLabelFor — positional fallback', () => {
  it('returns label when non-empty', () => {
    expect(entryLabelFor({ label: 'brief', path: '/x' }, 0)).toBe('brief');
    expect(entryLabelFor({ label: 'brief', path: '/x' }, 7)).toBe('brief');
  });

  it('returns file[N] when label is empty', () => {
    expect(entryLabelFor({ label: '', path: '/x' }, 0)).toBe('file[0]');
    expect(entryLabelFor({ label: '', path: '/x' }, 1)).toBe('file[1]');
    expect(entryLabelFor({ label: '', path: '/x' }, 5)).toBe('file[5]');
  });

  it('treats whitespace-only labels as non-empty (preserves user input)', () => {
    // Whitespace is a user-typed value; backend rejects only by exact-match
    // duplicates, so the UI must not silently drop trimmed-empty labels.
    expect(entryLabelFor({ label: ' ', path: '/x' }, 0)).toBe(' ');
  });
});

describe('hasDuplicateLabels — visual warning helper', () => {
  it('returns empty set when all labels unique', () => {
    const dups = hasDuplicateLabels([
      { label: 'a', path: '/x' },
      { label: 'b', path: '/y' },
      { label: '', path: '/z' },
    ]);
    expect(dups.size).toBe(0);
  });

  it('returns labels that appear twice or more', () => {
    const dups = hasDuplicateLabels([
      { label: 'brief', path: '/a' },
      { label: 'brief', path: '/b' },
      { label: 'plan', path: '/c' },
    ]);
    expect(dups.has('brief')).toBe(true);
    expect(dups.has('plan')).toBe(false);
    expect(dups.size).toBe(1);
  });

  it('never flags empty labels even when many entries are unlabeled', () => {
    const dups = hasDuplicateLabels([
      { label: '', path: '/a' },
      { label: '', path: '/b' },
      { label: '', path: '/c' },
    ]);
    expect(dups.size).toBe(0);
  });

  it('handles three-way duplicates as a single set entry', () => {
    const dups = hasDuplicateLabels([
      { label: 'x', path: '/a' },
      { label: 'x', path: '/b' },
      { label: 'x', path: '/c' },
    ]);
    expect(dups.size).toBe(1);
    expect(dups.has('x')).toBe(true);
  });
});
