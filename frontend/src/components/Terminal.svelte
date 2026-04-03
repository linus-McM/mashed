<script>
  import { onMount, onDestroy } from 'svelte';
  import { GetTerminalPort, GetAgentLog } from '../../wailsjs/go/main/App.js';
  import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime.js';
  import { currentTheme } from '../lib/stores/theme.js';
  import { currentMonoFont, currentFontSize } from '../lib/stores/font.js';

  export let paneTarget = '';
  export let repoPath = '';

  let terminalEl;
  let term;
  let ws;
  let logPollInterval;

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
    });

    const fitAddon = new FitAddon();
    term.loadAddon(fitAddon);

    // Try canvas addon but don't fail if unavailable
    try {
      const { CanvasAddon } = await import('@xterm/addon-canvas');
      const canvasAddon = new CanvasAddon();
      term.loadAddon(canvasAddon);
    } catch (e) {
      // Fall back to default renderer
    }

    term.open(terminalEl);
    fitAddon.fit();
    // Focus the terminal so it receives keyboard input and shows the cursor
    term.focus();
    // Re-focus after a short delay to ensure WebView has settled
    setTimeout(() => term.focus(), 100);
    // Re-focus whenever the window regains focus so the cursor keeps blinking
    onWindowFocus = () => { if (term) term.focus(); };
    window.addEventListener('focus', onWindowFocus);

    const resizeObserver = new ResizeObserver(() => fitAddon.fit());
    resizeObserver.observe(terminalEl);

    if (paneTarget) {
      // Live tmux terminal via WebSocket
      const port = await GetTerminalPort();
      if (port) {
        const url = `ws://127.0.0.1:${port}/ws/${encodeURIComponent(paneTarget)}`;
        ws = new WebSocket(url);
        ws.binaryType = 'arraybuffer';

        ws.onopen = () => {
          // Send initial resize so tmux knows the real terminal size
          sendResize();
          // Clear any stale rendering from the initial 1x1 pty
          term.clear();
        };

        ws.onmessage = (evt) => {
          const data = evt.data instanceof ArrayBuffer
            ? new TextDecoder().decode(evt.data)
            : evt.data;
          term.write(data);
        };

        ws.onclose = () => {
          term.write('\r\n\x1b[33m[disconnected]\x1b[0m\r\n');
        };

        ws.onerror = () => {
          term.write('\r\n\x1b[31m[connection error]\x1b[0m\r\n');
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
      }
    } else if (repoPath) {
      // Live log view — poll JSONL session data

      // Initial load
      await pollLog();

      // Poll every 2 seconds for new log lines
      logPollInterval = setInterval(pollLog, 2000);
    } else {
      term.write('\x1b[90m[no session data]\x1b[0m\r\n');
    }
  });

  onDestroy(() => {
    if (onWindowFocus) window.removeEventListener('focus', onWindowFocus);
    if (ws) ws.close();
    if (logPollInterval) clearInterval(logPollInterval);
    if (term) term.dispose();
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
  on:keydown={() => {}}
  role="textbox"
  tabindex="0"
  aria-label="Terminal"
></div>

<style>
  .terminal-wrapper {
    width: 100%;
    height: 100%;
    background: var(--bg-deepest);
    border-radius: 0;
    overflow: hidden;
  }
  .terminal-wrapper :global(.xterm) {
    padding: 8px;
  }
</style>
