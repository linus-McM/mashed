<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  // xterm.js ships its layout stylesheet separately from the JS module.
  // Without this import .xterm-viewport stays `position: static` so the
  // .xterm-screen rows end up rendered ~46k px below the wrapper instead
  // of being overlaid by the viewport — the terminal modal appears blank
  // even though content is streaming into the DOM. This is the required
  // import per xterm.js docs and must stay statically imported so Vite
  // bundles it with the component chunk.
  import '@xterm/xterm/css/xterm.css';
  import type { Terminal as Xterm, ITerminalOptions } from '@xterm/xterm';
  import type { FitAddon } from '@xterm/addon-fit';
  import { GetTerminalAuth, GetAgentLog } from '../../wailsjs/go/main/App.js';
  import { openTerminalSocket, watchEarlyClose, TERMINAL_AUTH_FAILED_MESSAGE } from '../lib/terminalSocket';
  import { EventsOn, ClipboardGetText, ClipboardSetText } from '../../wailsjs/runtime/runtime.js';
  import { currentTheme } from '../lib/stores/theme.js';
  import { currentMonoFont, currentFontSize } from '../lib/stores/font.js';
  import { editorSettings } from '../lib/stores/editorSettings.js';
  import type { AgentLogLine, LogLineKind, PtyResizeFrame, ScreenshotInjectEvent } from '../types/pty';

  type XtermCursorStyle = NonNullable<ITerminalOptions['cursorStyle']>;

  /** Map Monaco cursorStyle values to xterm.js cursorStyle. */
  function mapMonacoCursorToXterm(monacoStyle: string): XtermCursorStyle {
    switch (monacoStyle) {
      case 'line':
      case 'line-thin':
        return 'bar';
      case 'underline':
      case 'underline-thin':
        return 'underline';
      case 'block':
      case 'block-outline':
        return 'block';
      default:
        return 'bar';
    }
  }

  /** Map Monaco cursorBlinking values to xterm.js cursorBlink boolean. */
  function mapCursorBlinkToXterm(monacoBlinking: string): boolean {
    return monacoBlinking !== 'solid';
  }

  export let paneTarget = '';
  export let repoPath = '';

  let terminalEl: HTMLDivElement | undefined;
  let term: Xterm | null = null;
  let ws: WebSocket | null = null;
  let logPollInterval: ReturnType<typeof setInterval> | undefined;
  let resizeObserver: ResizeObserver | undefined;
  let unsubScreenshot: (() => void) | undefined;

  const KIND_COLORS: Record<LogLineKind, string> = {
    ok:     '\x1b[32m',  // green
    info:   '\x1b[34m',  // blue
    warn:   '\x1b[33m',  // amber
    err:    '\x1b[31m',  // red
    dim:    '\x1b[90m',  // gray
    system: '\x1b[35m',  // purple
  };

  const KIND_ICONS: Record<LogLineKind, string> = {
    ok:     '✓',
    info:   'ℹ',
    warn:   '⚠',
    err:    '✗',
    dim:    '·',
    system: '⬡',
  };

  function colorFor(kind: string): string {
    return (KIND_COLORS as Record<string, string>)[kind] ?? '\x1b[37m';
  }

  function iconFor(kind: string): string {
    return (KIND_ICONS as Record<string, string>)[kind] ?? ' ';
  }

  function formatLogLine(line: AgentLogLine): string {
    const color = colorFor(line.kind);
    const icon = iconFor(line.kind);
    const reset = '\x1b[0m';
    const dimColor = '\x1b[90m';

    let ts = '';
    if (line.ts) {
      const d = new Date(line.ts);
      ts = `${dimColor}${d.toLocaleTimeString()}${reset} `;
    }

    return `${ts}${color}${icon}${reset}  ${color}${line.text}${reset}`;
  }

  let lastLineCount = 0;
  let onWindowFocus: (() => void) | undefined;

  async function pollLog(): Promise<void> {
    if (!repoPath || !term) return;
    try {
      const lines = await GetAgentLog(repoPath);
      if (!lines || lines.length === 0) return;

      // Only write new lines since last poll
      if (lines.length > lastLineCount) {
        const newLines = lines.slice(lastLineCount);
        for (const line of newLines) {
          term.write(formatLogLine(line) + '\r\n');
        }
        lastLineCount = lines.length;
      }
    } catch (_e) {
      // Silently retry next poll
    }
  }

  onMount(async () => {
    const { Terminal } = await import('@xterm/xterm');
    const { FitAddon } = await import('@xterm/addon-fit');

    const localTerm: Xterm = new Terminal({
      fontFamily: $currentMonoFont,
      fontSize: $currentFontSize,
      theme: $currentTheme.xterm,
      cursorBlink: mapCursorBlinkToXterm($editorSettings.cursorBlinking),
      cursorStyle: mapMonacoCursorToXterm($editorSettings.cursorStyle),
      cursorInactiveStyle: 'outline',
      scrollback: 5000,
      disableStdin: !paneTarget, // Read-only when showing log view
      allowProposedApi: true,
    });
    term = localTerm;

    const fitAddon: FitAddon = new FitAddon();
    localTerm.loadAddon(fitAddon);

    // NOTE: @xterm/addon-canvas removed — it breaks text selection and scrolling
    // in macOS WKWebView. xterm.js 5.x default renderer handles both correctly.

    if (!terminalEl) return;
    localTerm.open(terminalEl);
    fitAddon.fit();
    // Focus the terminal so it receives keyboard input and shows the cursor
    localTerm.focus();
    // Re-focus after a short delay to ensure WebView has settled
    setTimeout(() => localTerm.focus(), 100);
    // Re-focus whenever the window regains focus so the cursor keeps blinking
    onWindowFocus = () => { if (term) term.focus(); };
    window.addEventListener('focus', onWindowFocus);

    // Clipboard helpers using Wails native API (bypasses webview restrictions)
    function copyText(text: string): void {
      ClipboardSetText(text).catch(() => {});
    }

    function pasteToTerminal(text: string): void {
      if (!text || !paneTarget) return;
      const encoder = new TextEncoder();
      if (ws && ws.readyState === WebSocket.OPEN) {
        ws.send(encoder.encode('\x1b[200~' + text + '\x1b[201~'));
      }
    }
    pasteFn = pasteToTerminal;

    // Cmd+C copies selection (or sends ^C if nothing selected),
    // Cmd+V pastes from clipboard into the terminal.
    localTerm.attachCustomKeyEventHandler((ev: KeyboardEvent): boolean => {
      const isMeta = ev.metaKey || ev.ctrlKey;
      if (ev.type !== 'keydown') return true;

      if (isMeta && ev.key === 'c') {
        const sel = localTerm.getSelection();
        if (sel) {
          copyText(sel);
          localTerm.clearSelection();
          return false;
        }
        return true; // no selection — send ^C
      }

      if (isMeta && ev.key === 'v') {
        ClipboardGetText()
          .then(pasteToTerminal)
          .catch(() => {});
        return false;
      }

      // Cmd+K — wipe xterm scrollback + ask the running app to redraw.
      // Clears the visible artifacts that scrollback accumulates when the
      // PTY has been resized (claude/zsh repaint at the new size but old
      // frames remain in history).
      if (isMeta && ev.key === 'k') {
        localTerm.clear();
        if (ws && ws.readyState === WebSocket.OPEN && paneTarget) {
          ws.send(new TextEncoder().encode('\x0c'));
        }
        return false;
      }

      return true;
    });

    resizeObserver = new ResizeObserver(() => {
      if (term && terminalEl) fitAddon.fit();
    });
    resizeObserver.observe(terminalEl);

    if (paneTarget) {
      // Live terminal via WebSocket — don't write anything before connect
      // to avoid scroll offset that misaligns the selection overlay.
      // Re-send resize whenever the terminal is re-fitted, and send keystrokes
      // as binary (bridge expects BinaryMessage for pty input). Registered
      // once; they always target the current socket, including after Retry.
      localTerm.onResize(() => sendResize());
      const encoder = new TextEncoder();
      localTerm.onData((data: string) => {
        if (ws && ws.readyState === WebSocket.OPEN) {
          ws.send(encoder.encode(data));
        }
      });
      await connectLive();
    } else if (repoPath) {
      // Live log view — poll JSONL session data
      localTerm.write('\x1b[90mMonitoring session log...\x1b[0m\r\n');

      // Initial load
      await pollLog();

      // Poll every 2 seconds for new log lines
      logPollInterval = setInterval(pollLog, 2000);
    } else {
      localTerm.write('\x1b[90m[no session data]\x1b[0m\r\n');
    }
  });

  /** Announced (aria-live) reason when the live socket could not connect. */
  let authError = '';
  /** pasteToTerminal from onMount, for the screenshot injector in connectLive. */
  let pasteFn: ((text: string) => void) | null = null;

  // Send resize event to bridge so the PTY knows the real terminal dimensions
  function sendResize(): void {
    if (ws && term && ws.readyState === WebSocket.OPEN && term.cols && term.rows) {
      const frame: PtyResizeFrame = { type: 'resize', cols: term.cols, rows: term.rows };
      ws.send(JSON.stringify(frame));
    }
  }

  /** Open the authenticated bridge socket for paneTarget (R5). */
  async function connectLive(): Promise<void> {
    const localTerm = term;
    if (!localTerm) return;
    authError = '';
    const auth = await GetTerminalAuth();
    if (!auth || !auth.port) {
      authError = TERMINAL_AUTH_FAILED_MESSAGE;
      return;
    }
    const localWs = openTerminalSocket(auth, paneTarget);
    ws = localWs;

    localWs.onopen = () => {
      // Send resize so PTY learns the real dimensions.
      sendResize();

      // The bridge replays the scroll buffer on connect. Write directly.
      const decoder = new TextDecoder();
      localWs.onmessage = (evt: MessageEvent<ArrayBuffer | string>) => {
        const raw = evt.data instanceof ArrayBuffer
          ? decoder.decode(evt.data)
          : evt.data;
        localTerm.write(raw);
      };

      // NOTE: previously sent Ctrl+L (\x0c) here as a "redraw nudge", but
      // for fresh sessions claude/zsh hadn't finished init by the time it
      // arrived — it landed in the input buffer and got echoed as a literal
      // `^L` glyph. Resize SIGWINCH already triggers a clean repaint, so
      // the nudge isn't needed.

      // Listen for screenshot path injection scoped to this terminal's pane
      if (unsubScreenshot) unsubScreenshot();
      unsubScreenshot = EventsOn('screenshot:inject', (data: ScreenshotInjectEvent) => {
        if (data.paneTarget !== paneTarget) return;
        pasteFn?.(data.path);
        if (localWs.readyState === WebSocket.OPEN) {
          localWs.send(new TextEncoder().encode('\r'));
        }
      });
    };

    localWs.onclose = () => {
      if (term) term.write('\r\n\x1b[33m[disconnected]\x1b[0m\r\n');
    };

    // A close before open means the bridge refused us (403) or is down.
    watchEarlyClose(localWs, (message) => { authError = message; });
  }

  function retryConnect(): void {
    if (ws) ws.close();
    ws = null;
    connectLive();
  }

  onDestroy(() => {
    if (unsubScreenshot) unsubScreenshot();
    if (resizeObserver) resizeObserver.disconnect();
    if (onWindowFocus) window.removeEventListener('focus', onWindowFocus);
    if (ws) ws.close();
    if (logPollInterval) clearInterval(logPollInterval);
    if (term) { term.dispose(); term = null; }
  });

  // Live theme switching — xterm supports setting theme via options
  $: if (term && $currentTheme) {
    term.options.theme = $currentTheme.xterm;
  }

  // Live font switching
  $: if (term && $currentMonoFont) {
    term.options.fontFamily = $currentMonoFont;
  }
  $: if (term && $currentFontSize) {
    term.options.fontSize = $currentFontSize;
  }

  // Live cursor style/blink switching from editor settings
  $: if (term && $editorSettings) {
    const newStyle = mapMonacoCursorToXterm($editorSettings.cursorStyle);
    const newBlink = mapCursorBlinkToXterm($editorSettings.cursorBlinking);
    if (term.options.cursorStyle !== newStyle) term.options.cursorStyle = newStyle;
    if (term.options.cursorBlink !== newBlink) term.options.cursorBlink = newBlink;
  }
</script>

<!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
<div
  class="terminal-wrapper"
  role="application"
  bind:this={terminalEl}
  on:click={() => term && term.focus()}
  on:keydown={() => {}}
></div>
{#if authError}
  <div class="terminal-status" role="status" aria-live="polite">
    <span>{authError}</span>
    <button type="button" on:click={retryConnect}>Retry</button>
  </div>
{/if}

<style>
  .terminal-status {
    display: flex;
    align-items: center;
    gap: var(--sp-sm);
    padding: var(--sp-xs) var(--sp-sm);
    font-size: var(--text-label);
    color: var(--text-secondary);
    background: var(--bg-deepest);
    border-top: 1px solid var(--border-subtle);
  }
  .terminal-status button {
    font: inherit;
    color: var(--text-primary);
    background: transparent;
    border: 1px solid var(--border-subtle);
    border-radius: var(--sp-xs);
    padding: var(--sp-2xs) var(--sp-sm);
    cursor: pointer;
  }
  .terminal-wrapper {
    width: 100%;
    height: 100%;
    background: var(--bg-deepest);
    border-radius: 0;
    overflow: hidden;
    user-select: none;
    -webkit-user-select: none;
    /* Prevent Wails frameless window from intercepting mouse events */
    --wails-draggable: no-drag;
  }
  .terminal-wrapper :global(.xterm) {
    padding: 0;
  }
  .terminal-wrapper :global(.xterm .xterm-screen) {
    cursor: text;
  }
  /* Hide xterm.js selection overlay — it renders one row above the actual text
     in WKWebView due to layout offset from internal measurement elements.
     The native browser selection is used instead, styled to match the theme. */
  .terminal-wrapper :global(.xterm-selection) {
    display: none !important;
  }
  .terminal-wrapper :global(.xterm *::selection) {
    background: var(--border-emphasis, #2a3340) !important;
  }
  /* Hide internal xterm.js measurement/input helpers that render visibly in WKWebView */
  .terminal-wrapper :global(.xterm-helper-textarea) {
    position: absolute !important;
    opacity: 0 !important;
    height: 0 !important;
    width: 0 !important;
    overflow: hidden !important;
  }
  .terminal-wrapper :global(.xterm-width-cache-measure-container) {
    position: absolute !important;
    top: -9999px !important;
    visibility: hidden !important;
  }
</style>
