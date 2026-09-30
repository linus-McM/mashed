// types/pty.ts — typed event payload interfaces for Wails-emitted PTY and
// screenshot events consumed by the Terminal component.
//
// Story svelte-check-04c (Phase 4c) — per cerebrum 2026-04-09 the PTY
// architecture uses a signed helper + `pty:<sessionId>:data|exit|resize`
// event names.  These interfaces pin the handler signatures in
// Terminal.svelte and anywhere else a test shim needs to fabricate payloads.

/** Payload for `pty:<sessionId>:data`. */
export interface PtyDataEvent {
  sessionId: string;
  chunk: string;
}

/** Payload for `pty:<sessionId>:exit`. */
export interface PtyExitEvent {
  sessionId: string;
  exitCode?: number;
}

/** Payload for `pty:<sessionId>:resize`. */
export interface PtyResizeEvent {
  sessionId: string;
  cols: number;
  rows: number;
}

/**
 * Payload for `screenshot:inject` (emitted from app.go when a screenshot
 * file is dropped / captured and should be pasted into the active terminal).
 */
export interface ScreenshotInjectEvent {
  path: string;
  paneTarget: string;
}

/** Discriminated kind tag for agent-log lines rendered into the terminal. */
export type LogLineKind = 'ok' | 'info' | 'warn' | 'err' | 'dim' | 'system';

/** Shape of a single agent log line as returned by `GetAgentLog`. */
export interface AgentLogLine {
  kind: LogLineKind | string;
  text: string;
  ts?: string;
}

/**
 * Resize message sent to the PTY bridge over the WebSocket as JSON.
 * Matches the `{ type: "resize", cols, rows }` frame the Go bridge expects.
 */
export interface PtyResizeFrame {
  type: 'resize';
  cols: number;
  rows: number;
}
