// Pure utilities for QuestionSnackbarStack.svelte.
//
// Extracted so they can be unit-tested without mounting a Svelte component.
// Kept framework-free: no Svelte imports, no DOM APIs beyond localStorage.

/** Default left-stripe colour when a repo has no saved border colour. */
export const DEFAULT_BORDER_COLOR = '#1e2530';

/** Maximum number of snackbar cards shown at once before overflow indicator kicks in. */
export const MAX_VISIBLE = 5;

/** localStorage key storing per-repo border colours as JSON `{ [repoName]: hex }`. */
export const REPO_COLORS_KEY = 'mashed:repoBorderColors';

/**
 * Truncate a string to `max` characters. When truncated, a single ellipsis
 * character is appended, so the returned length is `max + 1` codepoints.
 *
 * Uses spread-counted length rather than `.length` so tests that inspect
 * codepoint counts (e.g. `[...result].length`) get the expected value.
 */
export function truncate(text: string, max = 80): string {
  if (typeof text !== 'string') return '';
  const codepoints = [...text];
  if (codepoints.length <= max) return text;
  return codepoints.slice(0, max).join('') + '…';
}

/**
 * Look up the saved border colour for a repo from localStorage.
 *
 * Returns `DEFAULT_BORDER_COLOR` when:
 * - `repoName` is empty or not a string
 * - localStorage has no entry
 * - the stored JSON is malformed
 * - the stored value isn't an object
 * - the repo is not in the map
 * - localStorage itself throws (e.g. jsdom without storage, SSR)
 */
export function getBorderColor(repoName: string): string {
  if (!repoName || typeof repoName !== 'string') return DEFAULT_BORDER_COLOR;
  try {
    const raw = localStorage.getItem(REPO_COLORS_KEY);
    if (!raw) return DEFAULT_BORDER_COLOR;
    const parsed = JSON.parse(raw);
    if (!parsed || typeof parsed !== 'object') return DEFAULT_BORDER_COLOR;
    const value = (parsed as Record<string, unknown>)[repoName];
    return typeof value === 'string' && value.length > 0 ? value : DEFAULT_BORDER_COLOR;
  } catch {
    return DEFAULT_BORDER_COLOR;
  }
}

/**
 * Format a relative timestamp as a short human-readable string.
 *
 * - `< 60s`  -> "just now"
 * - `< 60m`  -> "Nm ago"
 * - `< 24h`  -> "Nh ago"
 * - `>= 24h` -> "Nd ago"
 *
 * `nowMs` is injectable for deterministic tests.
 */
export function timeAgo(unixMillis: number, nowMs: number = Date.now()): string {
  if (!Number.isFinite(unixMillis)) return '';
  const diff = Math.max(0, nowMs - unixMillis);
  const seconds = Math.floor(diff / 1000);
  if (seconds < 60) return 'just now';
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  return `${days}d ago`;
}

/**
 * Describes the minimal shape QuestionSnackbarStack consumes from a
 * QuestionEvent. Matches `internal/bmad.QuestionEvent`.
 */
export interface QuestionEventLike {
  execId?: string;
  nodeId: string;
  repoPath: string;
  repoName?: string;
  question: string;
  options?: string[];
  tmuxTarget?: string;
  timestamp?: number;
  questionId: string;
}

/**
 * Append (or replace) an event in the queue, keeping at most one entry per
 * `nodeId`. A new event for an existing `nodeId` replaces the old one rather
 * than duplicating it. Returns a new array — never mutates the input.
 */
export function upsertQuestion(
  queue: QuestionEventLike[],
  event: QuestionEventLike,
): QuestionEventLike[] {
  const filtered = queue.filter((q) => q.nodeId !== event.nodeId);
  return [...filtered, event];
}

/**
 * Remove any queued event matching `nodeId`. Returns a new array — never
 * mutates the input.
 */
export function dismissQuestion(
  queue: QuestionEventLike[],
  nodeId: string,
): QuestionEventLike[] {
  return queue.filter((q) => q.nodeId !== nodeId);
}

/**
 * Split a queue into the visible slice and an overflow count, enforcing the
 * `MAX_VISIBLE` cap.
 */
export function partitionForDisplay(
  queue: QuestionEventLike[],
): { visible: QuestionEventLike[]; overflow: number } {
  const visible = queue.slice(0, MAX_VISIBLE);
  const overflow = Math.max(0, queue.length - MAX_VISIBLE);
  return { visible, overflow };
}
