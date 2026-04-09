<script>
  import { onMount, onDestroy } from 'svelte';
  import { GetTerminalPort, GetAgentLog } from '../../wailsjs/go/main/App.js';
  import { EventsOn, EventsOff, ClipboardGetText, ClipboardSetText } from '../../wailsjs/runtime/runtime.js';
  import { currentTheme } from '../lib/stores/theme.js';
  import { currentMonoFont, currentFontSize } from '../lib/stores/font.js';

  export let paneTarget = '';
  export let repoPath = '';

  let terminalEl;
  let term;
  let ws;
  let logPollInterval;
  let resizeObserver;

  const KIND_COLORS = {
    ok:     '\x1b[32m',  // green
    info:   '\x1b[34m',  // blue
    warn:   '\x1b[33m',  // amber
    err:    '\x1b[31m',  // red
    dim:    '\x1b[90m',  // gray
    system: '\x1b[35m',  // purple
  };

  const KIND_ICONS = {
    ok:     '✓',
    info:   'ℹ',
    warn:   '⚠',
    err:    '✗',
    dim:    '·',
    system: '⬡',
  };

  function formatLogLine(line) {
    const color = KIND_COLORS[line.kind] || '\x1b[37m';
    const icon = KIND_ICONS[line.kind] || ' ';
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
  let onWindowFocus;

  async function pollLog() {
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
    } catch (e) {
      // Silently retry next poll
    }
  }

  onMount(async () => {
    const { Terminal } = await import('@xterm/xterm');
    const { FitAddon } = await import('@xterm/addon-fit');

    term = new Terminal({
      fontFamily: $currentMonoFont,
      fontSize: $currentFontSize,
      theme: $currentTheme.xterm,
      cursorBlink: true,
      cursorStyle: 'block',
      cursorInactiveStyle: 'outline',
      scrollback: 5000,
      disableStdin: !paneTarget, // Read-only when showing log view
      allowProposedApi: true,
    });

    const fitAddon = new FitAddon();
    term.loadAddon(fitAddon);

    // NOTE: @xterm/addon-canvas removed — it breaks text selection and scrolling
    // in macOS WKWebView. xterm.js 5.x default renderer handles both correctly.

    term.open(terminalEl);
    fitAddon.fit();
    // Focus the terminal so it receives keyboard input and shows the cursor
    term.focus();
    // Re-focus after a short delay to ensure WebView has settled
    setTimeout(() => term.focus(), 100);
    // Re-focus whenever the window regains focus so the cursor keeps blinking
    onWindowFocus = () => { if (term) term.focus(); };
    window.addEventListener('focus', onWindowFocus);

    // Clipboard helpers using Wails native API (bypasses webview restrictions)
    function copyText(text) {
      ClipboardSetText(text).catch(() => {});
    }

    function pasteToTerminal(text) {
      if (!text || !paneTarget) return;
      const encoder = new TextEncoder();
      if (ws && ws.readyState === WebSocket.OPEN) {
        ws.send(encoder.encode('\x1b[200~' + text + '\x1b[201~'));
      }
    }

    // Strip escape sequences that would break xterm.js behavior in the webview:
    // 1. Mouse tracking: prevents xterm.js from entering mouse-reporting mode,
    //    letting click+drag select text instead of forwarding events to tmux
    // 2. Alternate screen (smcup/rmcup): keeps xterm.js in normal buffer mode
    //    so mouse wheel scrolls the scrollback buffer instead of sending arrows
    // 3. Bracketed paste mode from tmux: we handle paste ourselves via Wails clipboard
    const stripRe = new RegExp(
      '\\x1b\\[\\?10(?:0[0-6]|15)[hl]' +  // mouse tracking on/off
      '|\\x1b\\[\\?1049[hl]' +              // alternate screen enter/exit
      '|\\x1b\\[\\?2004[hl]',               // bracketed paste on/off
      'g'
    );
    function stripControlSequences(data) {
      return data.replace(stripRe, '');
    }

    // Cmd+C copies selection (or sends ^C if nothing selected),
    // Cmd+V pastes from clipboard into the terminal.
    term.attachCustomKeyEventHandler((ev) => {
      const isMeta = ev.metaKey || ev.ctrlKey;
      if (ev.type !== 'keydown') return true;

      if (isMeta && ev.key === 'c') {
        const sel = term.getSelection();
        if (sel) {
          copyText(sel);
          term.clearSelection();
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

      return true;
    });

    resizeObserver = new ResizeObserver(() => {
      if (term && terminalEl) fitAddon.fit();
    });
    resizeObserver.observe(terminalEl);

    if (paneTarget) {
      // Live tmux terminal via WebSocket
      term.write('\x1b[90mConnecting to tmux session...\x1b[0m');
      const port = await GetTerminalPort();
      if (port) {
        const url = `ws://127.0.0.1:${port}/ws/${encodeURIComponent(paneTarget)}`;
        ws = new WebSocket(url);
        ws.binaryType = 'arraybuffer';

        ws.onopen = () => {
          // Send initial resize so tmux knows the real terminal size
          sendResize();
          // Clear the "Connecting..." message and any stale 1x1 rendering
          term.clear();
        };

        ws.onmessage = (evt) => {
          const raw = evt.data instanceof ArrayBuffer
            ? new TextDecoder().decode(evt.data)
            : evt.data;
          term.write(stripControlSequences(raw));
        };

        ws.onclose = () => {
          if (term) term.write('\r\n\x1b[33m[disconnected]\x1b[0m\r\n');
        };

        ws.onerror = () => {
          if (term) term.write('\r\n\x1b[31m[connection error]\x1b[0m\r\n');
        };

        // Send resize event to bridge so tmux/pty knows the real terminal dimensions
        function sendResize() {
          if (ws && ws.readyState === WebSocket.OPEN && term.cols && term.rows) {
            ws.send(JSON.stringify({ type: 'resize', cols: term.cols, rows: term.rows }));
          }
        }

        // Re-send resize whenever the terminal is re-fitted
        term.onResize(({ cols, rows }) => {
          sendResize();
        });

        // Send keystrokes as binary (bridge expects BinaryMessage for pty input)
        const encoder = new TextEncoder();
        term.onData((data) => {
          if (ws && ws.readyState === WebSocket.OPEN) {
            ws.send(encoder.encode(data));
          }
        });
      } else {
        term.write('\r\n\x1b[31m[terminal bridge not available]\x1b[0m\r\n');
      }
    } else if (repoPath) {
      // Live log view — poll JSONL session data
      term.write('\x1b[90mMonitoring session log...\x1b[0m\r\n');

      // Initial load
      await pollLog();

      // Poll every 2 seconds for new log lines
      logPollInterval = setInterval(pollLog, 2000);
    } else {
      term.write('\x1b[90m[no session data]\x1b[0m\r\n');
    }
  });

  onDestroy(() => {
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
</script>

<div
  class="terminal-wrapper"
  bind:this={terminalEl}
  on:click={() => term && term.focus()}
></div>

<style>
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
</style>
