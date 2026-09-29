import { describe, it, expect } from 'vitest';
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import {
  openTerminalSocket,
  watchEarlyClose,
  TERMINAL_AUTH_FAILED_MESSAGE,
} from '../../lib/terminalSocket';

// R5: the terminal pane authenticates with the per-launch token as a
// WebSocket subprotocol and announces an auth/bridge failure.

class FakeWebSocket {
  static last: FakeWebSocket | null = null;
  url: string;
  protocols: string[];
  binaryType = 'blob';
  onopen: (() => void) | null = null;
  onclose: (() => void) | null = null;
  constructor(url: string, protocols: string[]) {
    this.url = url;
    this.protocols = protocols;
    FakeWebSocket.last = this;
  }
}

describe('Terminal WebSocket auth (R5)', () => {
  it('offers mashed.v1 and the token subprotocol, token never in the URL', () => {
    const ws = openTerminalSocket(
      { port: 4242, token: 'abc123' },
      'term-repo:0.0',
      FakeWebSocket as unknown as typeof WebSocket,
    ) as unknown as FakeWebSocket;

    expect(ws.url).toBe('ws://127.0.0.1:4242/ws/term-repo%3A0.0');
    expect(ws.protocols).toEqual(['mashed.v1', 'mashed.auth.abc123']);
    expect(ws.url).not.toContain('abc123');
    expect(ws.binaryType).toBe('arraybuffer');
  });

  it('reports a close before open as an auth or bridge failure', () => {
    const ws = new FakeWebSocket('ws://x', []);
    const failures: string[] = [];
    watchEarlyClose(ws as unknown as WebSocket, (m) => failures.push(m));
    ws.onclose?.();
    expect(failures).toEqual([TERMINAL_AUTH_FAILED_MESSAGE]);
    expect(TERMINAL_AUTH_FAILED_MESSAGE).toBe('terminal authorisation failed or bridge unavailable');
  });

  it('does not report a close after a successful open', () => {
    const ws = new FakeWebSocket('ws://x', []);
    const failures: string[] = [];
    watchEarlyClose(ws as unknown as WebSocket, (m) => failures.push(m));
    ws.onopen?.();
    ws.onclose?.();
    expect(failures).toEqual([]);
  });

  it('Terminal.svelte uses GetTerminalAuth and an aria-live status with Retry', () => {
    const src = readFileSync(resolve(__dirname, '../Terminal.svelte'), 'utf8');
    expect(src).toMatch(/GetTerminalAuth\(\)/);
    expect(src).toMatch(/openTerminalSocket\(/);
    expect(src).toMatch(/watchEarlyClose\(/);
    expect(src).toMatch(/role="status"[^>]*aria-live="polite"[^>]*>[\s\S]*\{authError\}/);
    expect(src).toMatch(/<button[^>]*on:click=\{retryConnect\}[^>]*>\s*Retry\s*<\/button>/);
  });
});
