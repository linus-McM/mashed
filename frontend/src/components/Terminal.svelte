<script>
  import { onMount, onDestroy } from 'svelte';
  import { GetTerminalPort, GetAgentLog } from '../../wailsjs/go/main/App.js';
  import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime.js';

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
      fontFamily: "'JetBrains Mono', monospace",
      fontSize: 13,
      theme: {
        background: '#07080a',
        foreground: '#c8d4e0',
        cursor: '#00e57a',
        selectionBackground: '#2a3340',
        black: '#07080a',
        red: '#e84545',
        green: '#00e57a',
        yellow: '#f0a500',
        blue: '#3d9eff',
        magenta: '#9d6fff',
        cyan: '#00c4b3',
        white: '#c8d4e0',
      },
      cursorBlink: true,
      cursorStyle: 'block',
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
          term.write('\x1b[32m● Connected to agent terminal\x1b[0m\r\n\r\n');
        };

        ws.onmessage = (evt) => {
          const data = evt.data instanceof ArrayBuffer
            ? new TextDecoder().decode(evt.data)
            : evt.data;
          term.write(data);
        };

        ws.onclose = () => {
          term.write('\r\n\x1b[33m● Disconnected\x1b[0m\r\n');
        };

        ws.onerror = () => {
          term.write('\r\n\x1b[31m● Connection error\x1b[0m\r\n');
        };

        term.onData((data) => {
          if (ws && ws.readyState === WebSocket.OPEN) {
            ws.send(data);
          }
        });
      }
    } else if (repoPath) {
      // Live log view — poll JSONL session data
      term.write('\x1b[90m● Live agent activity log\x1b[0m\r\n');
      term.write('\x1b[90m  Streaming from session JSONL...\x1b[0m\r\n\r\n');

      // Initial load
      await pollLog();

      // Poll every 2 seconds for new log lines
      logPollInterval = setInterval(pollLog, 2000);
    } else {
      term.write('\x1b[90mNo agent session data available.\x1b[0m\r\n');
    }
  });

  onDestroy(() => {
    if (ws) ws.close();
    if (logPollInterval) clearInterval(logPollInterval);
    if (term) term.dispose();
  });
</script>

<div class="terminal-wrapper" bind:this={terminalEl}></div>

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
