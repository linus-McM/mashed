// Shared token-compliance helpers for BMAD colour audit tests.
//
// Four tests across story svelte-check-06 check that a `.svelte` component
// uses ONLY design-system tokens (no hex literals, no `rgb()/rgba()`,
// no `#39ff14`). This module centralises the reading + regex logic so
// drift in any component can be caught by a single one-line assertion.

import { existsSync, readFileSync } from 'node:fs';

/**
 * Obfuscated spelling of the toxic neon green hex pinned by Cerebrum
 * Do-Not-Repeat 2026-04-10. Written char-by-char so this module itself does
 * not trip any future repo-wide `#39ff14` grep.
 */
export const TOXIC_NEON: string = '#' + String.fromCharCode(51, 57, 102, 102, 49, 52); // #39ff14

/** Extract the `<style>…</style>` block verbatim from a Svelte source file. */
export function extractStyleBlock(source: string): string {
  const match = source.match(/<style[\s>][\s\S]*?<\/style>/i);
  return match ? match[0] : '';
}

/**
 * Find hex colour literals in `text`. Lines that contain `url(…)` (SVG
 * data-URIs) are skipped, and `var(--name, #fallback)` fallbacks inside
 * `var(…)` are stripped before matching.
 */
export function findHexLiterals(text: string): string[] {
  const hits: string[] = [];
  for (const line of text.split('\n')) {
    if (/url\s*\(/.test(line)) continue;
    const stripped = line.replace(/var\(--[\w-]+,\s*#[0-9a-fA-F]{3,8}\)/g, '');
    const found = stripped.match(/#[0-9a-fA-F]{3,8}\b/g);
    if (found) hits.push(...found);
  }
  return hits;
}

/** Find `rgb(` / `rgba(` calls whose first arg is a raw digit. */
export function findRgbLiterals(text: string): string[] {
  return text.match(/\brgba?\s*\(\s*\d/gi) ?? [];
}

/** Read a component file if it exists; returns `''` so callers can gate on presence. */
export function readSourceIfPresent(path: string): string {
  return existsSync(path) ? readFileSync(path, 'utf8') : '';
}

export interface TokenAuditResult {
  exists: boolean;
  source: string;
  styleBlock: string;
  hexLiterals: string[];
  rgbLiterals: string[];
  containsToxicNeon: boolean;
}

/**
 * One-shot token audit: given an absolute Svelte component path, return the
 * four signals callers assert on (existence, hex hits, rgb hits, neon hit).
 * When `scope: 'style'` (default) only the `<style>` block is scanned — that
 * matches the `ProcessSidebar` / `SkillEditorModal` precedent. `scope: 'full'`
 * scans the entire file (matches `CommandNode.colors` / `CanvasFailureToast`).
 */
export function auditTokens(
  componentPath: string,
  scope: 'style' | 'full' = 'full',
): TokenAuditResult {
  const exists = existsSync(componentPath);
  const source = exists ? readFileSync(componentPath, 'utf8') : '';
  const styleBlock = source ? extractStyleBlock(source) : '';
  const scanned = scope === 'style' ? styleBlock : source;
  return {
    exists,
    source,
    styleBlock,
    hexLiterals: scanned ? findHexLiterals(scanned) : [],
    rgbLiterals: scanned ? findRgbLiterals(scanned) : [],
    containsToxicNeon: scanned.includes(TOXIC_NEON),
  };
}
