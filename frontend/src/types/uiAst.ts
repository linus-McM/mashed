// TypeScript mirror of the UIAST schema (version 1) — spec §3.
//
// The backend validator (internal/uiast) is the source of truth; this module
// only describes the wire shape so the Svelte components can consume it.
// Fields are intentionally permissive — the dispatcher in AstNode.svelte
// falls back gracefully when a field is missing (§3.3 rule-1).

export type KnownUINodeType =
  | 'markdown'
  | 'hint'
  | 'summary'
  | 'code'
  | 'table'
  | 'decision_group';

// `(string & {})` preserves IDE autocomplete for KnownUINodeType while still
// accepting forward-compat types the backend may emit (§3.3 rule-1).
export type UINodeType = KnownUINodeType | (string & {});

export type HintTone = 'info' | 'warn' | 'error' | 'success';

export interface UINode {
  type: UINodeType;
  content?: string;
  tone?: HintTone;
  heading?: string;
  bullets?: string[];
  lang?: string;
  copyable?: boolean;
  columns?: string[];
  rows?: string[][];
  prompt?: string;
}

export interface UIAST {
  version: '1';
  nodes: UINode[];
}
