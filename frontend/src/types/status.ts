// types/status.ts — frontend-facing agent/event status vocabulary.
//
// StatusToken is the union of *every* status string the frontend may see in
// practice. It is sourced from three backend enums, all declared in
// `internal/domain/types.go`:
//
//   AgentStatus (line ~10):
//     running, open, finished, waiting, blocked, error, queued, done
//
//   EventType (line ~157):
//     needs_response, error, completed, running, started
//
//   SubAgentStatus (string, line ~192 comment):
//     running, done
//
// Plus one frontend-only UI token:
//
//   terminal — shell session surfaced in NotificationFeed (no agent state)
//
// Never invent tokens the backend does not emit. If you need a new one, add
// the constant on the Go side first (domain/types.go) and then mirror it
// here. Keep this list alphabetical inside each group for legibility.
//
// This union is intentionally closed. At component boundaries that receive a
// string from Wails events or DOM attributes, narrow once with `isStatusToken`
// or cast once with `as StatusToken`. Do NOT pepper casts across call sites.

/**
 * AgentStatus tokens from `domain.AgentStatus` (internal/domain/types.go).
 */
export type AgentStatusToken =
  | 'running'
  | 'open'
  | 'finished'
  | 'waiting'
  | 'blocked'
  | 'error'
  | 'queued'
  | 'done';

/**
 * EventType tokens from `domain.EventType` (internal/domain/types.go).
 * Overlaps with AgentStatus for `running` / `error`; `needs_response`,
 * `completed`, `started` are event-only.
 */
export type EventTypeToken =
  | 'needs_response'
  | 'error'
  | 'completed'
  | 'running'
  | 'started';

/** UI-only status for a shell/terminal session (not backed by an agent). */
export type UITerminalToken = 'terminal';

/**
 * The complete closed set of status-like strings the UI may render.
 * Maps to `Record<StatusToken, string>` colour/label tables.
 */
export type StatusToken =
  | AgentStatusToken
  | EventTypeToken
  | UITerminalToken;

/**
 * Known StatusToken values as a runtime frozen set. Only use for `has` checks;
 * do not export as an array — the order is not meaningful.
 */
const STATUS_TOKENS: ReadonlySet<StatusToken> = new Set<StatusToken>([
  // AgentStatusToken
  'running',
  'open',
  'finished',
  'waiting',
  'blocked',
  'error',
  'queued',
  'done',
  // EventTypeToken (running/error already present, Set dedupes)
  'needs_response',
  'completed',
  'started',
  // UITerminalToken
  'terminal',
]);

/**
 * Type guard: narrows `unknown` to `StatusToken`.
 * Use this (exactly once) at the external boundary — a Wails event payload,
 * a DOM data attribute, or a localStorage read. Downstream code should
 * operate on `StatusToken` values directly without re-checking.
 */
export function isStatusToken(value: unknown): value is StatusToken {
  return typeof value === 'string' && STATUS_TOKENS.has(value as StatusToken);
}
