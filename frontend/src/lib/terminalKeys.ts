// Key handling for the embedded xterm.js terminal.
//
// xterm.js 6 sends CR for both Enter and Shift+Enter, so Claude Code cannot
// tell them apart and Shift+Enter submits. Until the kitty keyboard protocol
// ships in a stable xterm.js (6.1/7.0, `vtExtensions.kittyKeyboard`), map
// Shift+Enter to Ctrl+J (LF), which Claude Code treats as "insert newline"
// in every terminal.

/** Bytes sent for Shift+Enter: Ctrl+J / line feed. */
export const SHIFT_ENTER_SEQUENCE = '\n';

/**
 * Build an xterm `attachCustomKeyEventHandler` callback. It returns false
 * (xterm must not handle the key) for Shift+Enter and sends the newline
 * sequence itself on keydown; every other key is left to xterm.
 */
export function createKeyHandler(send: (data: string) => void): (e: KeyboardEvent) => boolean {
  return (e: KeyboardEvent): boolean => {
    const shiftEnterOnly = e.key === 'Enter' && e.shiftKey && !e.ctrlKey && !e.altKey && !e.metaKey;
    if (!shiftEnterOnly) return true;
    if (e.type === 'keydown') send(SHIFT_ENTER_SEQUENCE);
    return false;
  };
}
