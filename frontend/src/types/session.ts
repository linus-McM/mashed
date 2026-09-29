// types/session.ts — frontend-facing Session domain type.
//
// `Session` is a TerminalSession as modelled by Go (`domain.TerminalSession`
// in `internal/domain/types.go`). The canonical shape flows through
// `wailsjs/go/models.ts` and is re-exported under friendly names in
// `$lib/types/wails`. This module pulls those in and keeps the symbol
// co-located with the other TS type modules (`types/uiAst.ts`,
// `types/workflow.ts`, `types/theme.ts`).
//
// Svelte stores consuming sessions should import from here rather than
// reaching into the generated Wails models directly, so a future Go rename
// only needs the re-export in `$lib/types/wails` updated.

import type { Session as WailsSession } from '../lib/types/wails';

// `WailsSession` is the generated `TerminalSession` *class* — it carries a
// runtime `convertValues` method we never hand-construct. Pick just the data
// fields so literal-object constructions (`makeSession`) stay assignable.
type DataFields<T> = {
  [K in keyof T as T[K] extends (...args: never[]) => unknown ? never : K]: T[K];
};

/**
 * A live or recently-dead tmux session tied to a repo + model.
 *
 * Shape mirrors `domain.TerminalSession`:
 *   - sessionName  — unique tmux session name (also the display label)
 *   - paneTarget   — tmux target in `session:window.pane` form
 *   - repoPath     — absolute path to the repository the session is operating on
 *   - repoName     — basename of `repoPath` (display convenience)
 *   - sessionType  — free-form tag (e.g. "claude", "gemini", "bmad")
 *   - model        — model alias as returned by `ListModels()`
 *   - spawnedAt    — ISO-8601 timestamp from Go's `time.Time` marshal
 *   - isAlive      — whether the tmux pane is still running
 *
 * Wails marshals Go's `time.Time` as a string, but `models.ts` types it as
 * `any`. We narrow it to `string` here at the frontend boundary — `Date`
 * construction accepts ISO strings directly.
 */
export type Session = Omit<DataFields<WailsSession>, 'spawnedAt'> & {
  spawnedAt: string;
};

/**
 * Store shape for `repoSessions`: map from `repoPath` to its sessions list.
 * Empty-map is the valid "nothing loaded yet" state.
 */
export type SessionState = Record<string, Session[]>;
