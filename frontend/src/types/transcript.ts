// Transcript domain types for BMAD interactive input.
//
// The `Turn` type was previously re-exported from `TranscriptPane.svelte`
// via `/** @typedef */`, but Svelte's namespace export model does not expose
// types declared inside `<script>` blocks (Svelte namespace tracks component
// props + events, not arbitrary type aliases). Hoisting to a dedicated `.ts`
// module makes the type importable from anywhere without the namespace
// ambiguity — see story svelte-check-07 AC-1.

import type { InteractiveTurn } from '../lib/types/wails';

/** Role of a participant in a BMAD interactive round. */
export type TurnRole = 'claude' | 'user';

/**
 * A single turn in a node's interactive transcript. `round` increases by one
 * each time Claude completes a reply; a `user` turn is typically followed by
 * a `claude` turn at the same round number.
 */
export interface Turn {
  round: number;
  role: TurnRole;
  inputId: string;
  content: string;
  timestamp: number;
}

/**
 * Narrow the Wails-generated `InteractiveTurn` (whose `role` is typed as the
 * broad `string` Go emits) into the discriminated `Turn` shape the UI
 * consumes. Unknown roles fall back to `'claude'` so the transcript still
 * renders; empty inputId/timestamp default to safe zero values.
 */
export function normaliseInteractiveTurn(raw: InteractiveTurn): Turn {
  const role: TurnRole = raw.role === 'user' ? 'user' : 'claude';
  return {
    round: raw.round,
    role,
    inputId: raw.inputId ?? '',
    content: raw.content,
    timestamp: raw.timestamp ?? 0,
  };
}
