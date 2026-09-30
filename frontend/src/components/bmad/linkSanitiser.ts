// Shared URL sanitisation for AST components (§7.2 point 3-4).
//
// Reject `javascript:`, `data:`, `file:` schemes — these are the XSS/exfil
// vectors called out by the spec. Everything else (http, https, mailto, ftp,
// custom protocols) is allowed after normalisation through URL parsing.
//
// Used by MarkdownBlock link renderer AND by HintBanner / SummaryCard /
// ComparisonTable content audits (belt-and-braces — the backend validator
// already rejects these schemes, but the frontend must still refuse to render
// them if a malformed AST slips through).

const REJECTED_SCHEMES = ['javascript', 'data', 'file'] as const;
const REJECTED_PROTOCOL_RE = new RegExp(`^(?:${REJECTED_SCHEMES.join('|')}):`, 'i');

export function sanitizeUrl(href: string | null | undefined): string | null {
  if (!href) return null;
  const trimmed = href.trim();
  if (!trimmed) return null;
  try {
    const u = new URL(trimmed, 'https://invalid.local/');
    if (REJECTED_PROTOCOL_RE.test(u.protocol)) return null;
    return u.href;
  } catch {
    return null;
  }
}

// Strips rejected-scheme substrings from free-text content. Used by passive
// components that don't render links as `<a>` but may still surface a raw URL
// string. Preserves the rest of the text unchanged.
const UNSAFE_URL_IN_TEXT = new RegExp(
  `\\b(?:${REJECTED_SCHEMES.join('|')}):[^\\s)]+`,
  'gi',
);
export function scrubUnsafeUrls(text: string | null | undefined): string {
  if (!text) return '';
  return text.replace(UNSAFE_URL_IN_TEXT, '[link removed]');
}
