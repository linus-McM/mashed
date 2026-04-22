// types/bmadEvents.ts — typed event payload interfaces for `bmad:*` events.
//
// Story svelte-check-04a — shared between the Wails event subscriptions in
// WorkflowBuilder.svelte and (eventually) any test harness / shim that needs
// to fabricate payloads.  Kept permissive (every field optional where the
// backend permits it) so the narrowing happens explicitly at each handler,
// not silently at the type boundary.

/** Payload for `bmad:node:awaiting_input`. */
export interface BmadAwaitingInputEvent {
  execId?: string;
  nodeId: string;
  inputId: string;
  prompt?: string;
  shape?: string;
  options?: string[];
  round?: number;
  createdAt?: number;
  promptId?: string;
  required?: boolean;
  helpText?: string;
  maxLength?: number;
  maxRounds?: number;
  structured?: string;
  lastOutput?: string;
}

/** Payload for `bmad:node:input_resolved`. Carries `valueHash` — NEVER the raw value. */
export interface BmadInputResolvedEvent {
  execId?: string;
  nodeId: string;
  inputId: string;
  valueHash?: string;
  round?: number;
}

/** Payload for `bmad:node:input_invalid`. */
export interface BmadInputInvalidEvent {
  execId?: string;
  nodeId: string;
  inputId: string;
  reason?: string;
}

/** Payload for `bmad:node:round_complete`. */
export interface BmadRoundCompleteEvent {
  execId?: string;
  nodeId: string;
  round?: number;
  outputKey?: string;
}

/** Payload for `bmad:node:gate_satisfied`. */
export interface BmadGateSatisfiedEvent {
  execId?: string;
  nodeId: string;
  round?: number;
  reason?: string;
}

/** Payload for `bmad:node:round_limit`. */
export interface BmadRoundLimitEvent {
  execId?: string;
  nodeId: string;
  round?: number;
}

/** Payload for `bmad:node:aborted`. */
export interface BmadAbortedEvent {
  execId?: string;
  nodeId: string;
  reason?: string;
}

/** Payload for `bmad:node:status` (autonomous live status tick). */
export interface BmadNodeStatusEvent {
  execId?: string;
  nodeId: string;
  status?: string;
  tmuxTarget?: string;
  iteration?: number;
  message?: string;
}

/** Payload for `bmad:node:artifacts` (per-node artifact resolution). */
export interface BmadNodeArtifactsEvent {
  execId?: string;
  nodeId: string;
  found?: string[];
  missing?: string[];
  paths?: Record<string, string>;
}

/** Payload for `bmad:execution:status` (execution-level lifecycle). */
export interface BmadExecutionStatusEvent {
  execId?: string;
  status?: string;
}

/** Payload for `bmad:sprint:updated` (story status delta). */
export interface BmadSprintUpdatedEvent {
  storyId?: string;
  status?: string;
}

/**
 * Union of every `bmad:node:*` interactive-input event.  Indexed by the
 * wire event name so the handler map in WorkflowBuilder can be statically
 * typed — a missing name is a compile error, and the handler parameter
 * narrows to the correct interface.
 */
export interface BmadInteractiveEventMap {
  'bmad:node:awaiting_input': BmadAwaitingInputEvent;
  'bmad:node:input_resolved': BmadInputResolvedEvent;
  'bmad:node:input_invalid': BmadInputInvalidEvent;
  'bmad:node:round_complete': BmadRoundCompleteEvent;
  'bmad:node:gate_satisfied': BmadGateSatisfiedEvent;
  'bmad:node:round_limit': BmadRoundLimitEvent;
  'bmad:node:aborted': BmadAbortedEvent;
}

/**
 * Handler map type — parameterised on the event name so each handler
 * signature is pinned to its event's payload type. Used as
 * `BmadInteractiveHandlers` when building the `EventsOn` subscription
 * table so no handler can drift from its payload shape.
 */
export type BmadInteractiveHandlers = {
  [K in keyof BmadInteractiveEventMap]: (event: BmadInteractiveEventMap[K]) => void;
};
