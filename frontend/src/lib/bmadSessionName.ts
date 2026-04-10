// bmadSessionName.ts — parses BMAD tmux session names produced by
// internal/bmad.BuildSessionName into their component parts for display.
// Mirrors the Go ParseSessionName logic.

export interface FriendlyTarget {
  repo: string;
  branch: string;
  label: string;
  hash: string;
  raw: string;
}

const SESSION_PREFIX = 'bmad-';
const HASH_REGEX = /^[0-9a-f]{8}$/;

/**
 * parseFriendlyTarget decomposes a BMAD tmux target into its component
 * parts (repo, branch, label, 8-char hex hash). Accepts either form:
 *   - "bmad-repo-branch-label-abcd1234"
 *   - "bmad-repo-branch-label-abcd1234:0.0"
 * Returns null for any input that doesn't match the expected shape.
 */
export function parseFriendlyTarget(target: string): FriendlyTarget | null {
  if (!target) return null;
  const raw = target;

  // Strip optional ":window.pane" suffix.
  const colonIdx = target.indexOf(':');
  const name = colonIdx === -1 ? target : target.slice(0, colonIdx);

  if (!name.startsWith(SESSION_PREFIX)) return null;

  const body = name.slice(SESSION_PREFIX.length);
  const parts = body.split('-');
  // Need at least: repo, branch, label (>=1 part), hash
  if (parts.length < 4) return null;

  const hash = parts[parts.length - 1];
  if (!HASH_REGEX.test(hash)) return null;

  const repo = parts[0];
  const branch = parts[1];
  const label = parts.slice(2, -1).join('-');

  if (!repo || !branch || !label) return null;

  return { repo, branch, label, hash, raw };
}
