// types/reviewEvents.ts — typed event payload interfaces for `review:*`
// events emitted by app_review.go / app_review_scoped.go.
//
// Story svelte-check-04c (Phase 4c) — the SummarisationModal subscribes
// to progress + done streams.  Keeping the shapes pinned here prevents
// drift between the Go emitter and the Svelte handler at compile time.

/**
 * Per-file entry inside a `review:summary:progress` / `review:summary:done`
 * event.  Mirrors `main.FileSummary` in `app_review.go`.
 */
export interface ReviewFileSummary {
  path: string;
  added: number;
  removed: number;
  summary: string;
  isBinary?: boolean;
}

/** Aggregate payload of a completed review. Mirrors `main.ReviewSummary`. */
export interface ReviewSummary {
  files: ReviewFileSummary[];
  totalAdded: number;
  totalRemoved: number;
}

/** Payload for `review:summary:progress`. */
export interface ReviewSummaryProgressEvent {
  repoPath: string;
  file: ReviewFileSummary;
  index: number;
  total: number;
  error?: string;
}

/** Payload for `review:summary:done`. */
export interface ReviewSummaryDoneEvent {
  repoPath: string;
  summary?: ReviewSummary | null;
  error?: string;
}

/**
 * Payload for `review:advice:progress` (streamed methodology advice).
 * `done: true` with `error: ""` signals the end of the stream.
 */
export interface ReviewAdviceProgressEvent {
  repoPath: string;
  text: string;
  done: boolean;
  error?: string;
}
