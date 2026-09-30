// Authenticated terminal WebSocket (spec R5). The per-launch token travels as
// the `mashed.auth.<token>` subprotocol: browsers cannot set headers on a
// WebSocket, and a subprotocol keeps the token out of URLs and logs.

export type TerminalAuth = { port: number; token: string };

export const TERMINAL_AUTH_FAILED_MESSAGE = 'terminal authorisation failed or bridge unavailable';

/** Open the bridge socket for paneTarget with the auth subprotocols. */
export function openTerminalSocket(
  auth: TerminalAuth,
  paneTarget: string,
  WebSocketImpl: typeof WebSocket = WebSocket,
): WebSocket {
  const url = `ws://127.0.0.1:${auth.port}/ws/${encodeURIComponent(paneTarget)}`;
  const ws = new WebSocketImpl(url, ['mashed.v1', `mashed.auth.${auth.token}`]);
  ws.binaryType = 'arraybuffer';
  return ws;
}

/**
 * Call onFail when the socket closes before it ever opened: the bridge
 * answered 403 (bad token or origin) or is not listening. Chains any onopen
 * / onclose handlers already set, and returns nothing.
 */
export function watchEarlyClose(ws: WebSocket, onFail: (message: string) => void): void {
  let opened = false;
  const prevOpen = ws.onopen;
  const prevClose = ws.onclose;
  ws.onopen = function (this: WebSocket, ev: Event) {
    opened = true;
    prevOpen?.call(this, ev);
  };
  ws.onclose = function (this: WebSocket, ev: CloseEvent) {
    if (!opened) onFail(TERMINAL_AUTH_FAILED_MESSAGE);
    prevClose?.call(this, ev);
  };
}
