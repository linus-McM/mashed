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
  kind?: 'question';
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
 * Describes the minimal shape QuestionSnackbarStack consumes from an
 * IdleEvent. Matches `internal/bmad.IdleEvent`.
 *
 * Idle entries represent "pane is at the Claude CLI prompt waiting for
 * user input" — there is no question text because the signal is
 * pane-level quiescence, not a structured prompt. Rendered with a
 * distinct icon/label so the user can tell it apart from a real question.
 */
export interface IdleEventLike {
  kind: 'idle';
  execId?: string;
  nodeId: string;
  repoPath: string;
  repoName?: string;
  tmuxTarget?: string;
  timestamp?: number;
}

/** Union of everything the snackbar stack can render. */
export type SnackbarEntry = QuestionEventLike | IdleEventLike;

/**
 * Type guard: true when the entry is a structured question. Splits the
 * union so render paths can safely read `.question` / `.options` without
 * optional-chaining gymnastics.
 */
export function isQuestionEntry(
  entry: SnackbarEntry,
): entry is QuestionEventLike {
  return entry.kind !== 'idle';
}

/**
 * Append (or replace) a question event in the queue, keeping at most one
 * entry per `nodeId`. A new event for an existing `nodeId` replaces the
 * old one — including the case where the existing entry is an idle event,
 * because structured questions carry more information and should win.
 * Returns a new array — never mutates the input.
 */
export function upsertQuestion(
  queue: SnackbarEntry[],
  event: QuestionEventLike,
): SnackbarEntry[] {
  const normalised: QuestionEventLike = { ...event, kind: 'question' };
  const filtered = queue.filter((q) => q.nodeId !== event.nodeId);
  return [...filtered, normalised];
}

/**
 * Append (or replace) an idle event in the queue — but only when the
 * existing entry for `nodeId` is NOT a structured question. Questions
 * always win because they carry richer context; an idle signal for a
 * node that is already showing a question is redundant.
 *
 * This conservative policy is important because the Claude CLI input
 * prompt (`❯`) appears BELOW structured questions too, so both detectors
 * can fire for the same pane state — we do not want an idle event to
 * clobber a real question that already lit up the snackbar.
 */
export function upsertIdle(
  queue: SnackbarEntry[],
  event: IdleEventLike,
): SnackbarEntry[] {
  const existing = queue.find((q) => q.nodeId === event.nodeId);
  if (existing && isQuestionEntry(existing)) {
    return queue;
  }
  const normalised: IdleEventLike = { ...event, kind: 'idle' };
  const filtered = queue.filter((q) => q.nodeId !== event.nodeId);
  return [...filtered, normalised];
}

/**
 * Remove any queued event matching `nodeId`. Returns a new array — never
 * mutates the input. Used by both question and idle dismissal paths —
 * dismissal is by nodeId, not by kind.
 */
export function dismissQuestion(
  queue: SnackbarEntry[],
  nodeId: string,
): SnackbarEntry[] {
  return queue.filter((q) => q.nodeId !== nodeId);
}

/**
 * Remove any queued IDLE entry matching `nodeId`, leaving question entries
 * untouched. Used when an EventIdleDismissed arrives: we must not clobber
 * a question that may have taken over this node's slot since the idle
 * event was queued.
 */
export function dismissIdle(
  queue: SnackbarEntry[],
  nodeId: string,
): SnackbarEntry[] {
  return queue.filter((q) => !(q.nodeId === nodeId && !isQuestionEntry(q)));
}

/**
 * Split a queue into the visible slice and an overflow count, enforcing the
 * `MAX_VISIBLE` cap.
 */
export function partitionForDisplay(
  queue: SnackbarEntry[],
): { visible: SnackbarEntry[]; overflow: number } {
  const visible = queue.slice(0, MAX_VISIBLE);
  const overflow = Math.max(0, queue.length - MAX_VISIBLE);
  return { visible, overflow };
}
