/**
 * @vitest-environment jsdom
 */
// Unit tests for the shared sanitisation helper (§7.2 guards). The surface-
// level §7.2 ACs (javascript/data/file stripping in MarkdownBlock) live in
// MarkdownBlock.test.ts as part of T6; these are the inner-function tests.

import { describe, it, expect } from 'vitest';
import { sanitizeUrl, scrubUnsafeUrls } from '../linkSanitiser';

describe('sanitizeUrl', () => {
  it('returns null for null/undefined/empty input', () => {
    expect(sanitizeUrl(null)).toBeNull();
    expect(sanitizeUrl(undefined)).toBeNull();
    expect(sanitizeUrl('')).toBeNull();
    expect(sanitizeUrl('   ')).toBeNull();
  });

  it('rejects javascript: scheme (case-insensitive)', () => {
    expect(sanitizeUrl('javascript:alert(1)')).toBeNull();
    expect(sanitizeUrl('JavaScript:alert(1)')).toBeNull();
    expect(sanitizeUrl('JAVASCRIPT:alert(1)')).toBeNull();
  });

  it('rejects data: scheme', () => {
    expect(sanitizeUrl('data:text/html,<script>alert(1)</script>')).toBeNull();
  });

  it('rejects file: scheme', () => {
    expect(sanitizeUrl('file:///etc/passwd')).toBeNull();
  });

  it('accepts https URLs and returns normalised href', () => {
    expect(sanitizeUrl('https://example.com')).toBe('https://example.com/');
    expect(sanitizeUrl('  https://example.com/path?q=1  ')).toBe('https://example.com/path?q=1');
  });

  it('accepts http URLs', () => {
    expect(sanitizeUrl('http://example.com/')).toBe('http://example.com/');
  });

  it('accepts mailto URLs', () => {
    expect(sanitizeUrl('mailto:test@example.com')).toBe('mailto:test@example.com');
  });

  it('treats relative paths as relative to a dummy base and returns absolute', () => {
    const out = sanitizeUrl('/docs/intro');
    expect(out).toContain('/docs/intro');
  });
});

describe('scrubUnsafeUrls', () => {
  it('returns empty string for null/undefined', () => {
    expect(scrubUnsafeUrls(null)).toBe('');
    expect(scrubUnsafeUrls(undefined)).toBe('');
    expect(scrubUnsafeUrls('')).toBe('');
  });

  it('leaves safe text unchanged', () => {
    expect(scrubUnsafeUrls('Hello world')).toBe('Hello world');
    expect(scrubUnsafeUrls('See https://example.com for docs')).toBe(
      'See https://example.com for docs',
    );
  });

  it('strips javascript: URL substrings', () => {
    const out = scrubUnsafeUrls('click javascript:alert(1) here');
    expect(out).not.toContain('javascript:');
    expect(out).toContain('[link removed]');
  });

  it('strips data: URL substrings', () => {
    const out = scrubUnsafeUrls('see data:text/html,<x> now');
    expect(out).not.toContain('data:text/html');
  });

  it('strips file: URL substrings', () => {
    const out = scrubUnsafeUrls('open file:///etc/passwd please');
    expect(out).not.toContain('file:///');
  });
});
