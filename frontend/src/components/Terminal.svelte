<script>
  import { onMount, onDestroy } from 'svelte';
  import { GetTerminalPort } from '../../wailsjs/go/main/App.js';

  export let paneTarget = '';

  let terminalEl;
  let term;
  let ws;

  onMount(async () => {
    // Dynamic import xterm.js (it's an npm dep)
    const { Terminal } = await import('@xterm/xterm');
    const { FitAddon } = await import('@xterm/addon-fit');
    const { CanvasAddon } = await import('@xterm/addon-canvas');

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
    });

    const fitAddon = new FitAddon();
    term.loadAddon(fitAddon);

    // Use canvas renderer (NOT WebGL — WebKit compat)
    try {
      const canvasAddon = new CanvasAddon();
      term.loadAddon(canvasAddon);
    } catch (e) {
      console.warn('Canvas addon failed, using default renderer:', e);
    }

    term.open(terminalEl);
    fitAddon.fit();

    // Connect to Go WebSocket bridge
    const port = await GetTerminalPort();
    if (port && paneTarget) {
      const url = `ws://127.0.0.1:${port}/ws/${encodeURIComponent(paneTarget)}`;
      ws = new WebSocket(url);
      ws.binaryType = 'arraybuffer';

      ws.onopen = () => {
        term.write('\r\n\x1b[32mConnected to agent terminal\x1b[0m\r\n');
      };

      ws.onmessage = (evt) => {
        const data = evt.data instanceof ArrayBuffer
          ? new TextDecoder().decode(evt.data)
          : evt.data;
        term.write(data);
      };

      ws.onclose = () => {
        term.write('\r\n\x1b[33mDisconnected\x1b[0m\r\n');
      };

      ws.onerror = () => {
        term.write('\r\n\x1b[31mConnection error\x1b[0m\r\n');
      };

      // Send keystrokes to backend
      term.onData((data) => {
        if (ws && ws.readyState === WebSocket.OPEN) {
          ws.send(data);
        }
      });
    } else {
      term.write('No terminal pane available for this agent.\r\n');
      term.write('The agent may not be running in tmux.\r\n');
    }

    // Handle resize
    const resizeObserver = new ResizeObserver(() => fitAddon.fit());
    resizeObserver.observe(terminalEl);
  });

  onDestroy(() => {
    if (ws) ws.close();
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
