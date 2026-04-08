package terminal

import (
	"context"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"
)

// Sentinel errors for the terminal bridge.
var (
	ErrBridgeClosed    = errors.New("terminal: bridge closed")
	ErrPaneTargetEmpty = errors.New("terminal: pane target is empty")
)

const (
	outputBufMax = 64 * 1024 // 64KB ring buffer for slow consumers
	ptyReadSize  = 4096
)

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // allow Wails webview origin
	},
}

// resizeMsg is sent by the xterm.js frontend to resize the pty.
type resizeMsg struct {
	Type string `json:"type"`
	Cols uint16 `json:"cols"`
	Rows uint16 `json:"rows"`
}

// Bridge serves WebSocket connections that attach to tmux panes via pty.
// Each WebSocket connection spawns a tmux attach-session under a pty,
// with one goroutine pair for bidirectional I/O.
type Bridge struct {
	mu       sync.Mutex
	listener net.Listener
	server   *http.Server
	port     int
	conns    map[*websocket.Conn]context.CancelFunc
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewBridge creates a new terminal bridge. Call Start() to begin serving.
func NewBridge() *Bridge {
	return &Bridge{
		conns: make(map[*websocket.Conn]context.CancelFunc),
	}
}

// Start begins listening on a random localhost port for WebSocket connections.
func (b *Bridge) Start(ctx context.Context) error {
	b.ctx, b.cancel = context.WithCancel(ctx)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return &TerminalError{Op: "bridge_start", Err: fmt.Errorf("listen: %w", err)}
	}
	b.listener = ln
	b.port = ln.Addr().(*net.TCPAddr).Port

	mux := http.NewServeMux()
	mux.HandleFunc("/ws/", b.handleWS)

	b.server = &http.Server{Handler: mux}

	go func() {
		if err := b.server.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("terminal bridge: serve error: %v", err)
		}
	}()

	go func() {
		<-b.ctx.Done()
		b.shutdown()
	}()

	return nil
}

// GetTerminalPort returns the port the WebSocket server is listening on.
// Exposed as a Wails binding for the Svelte frontend to connect xterm.js.
func (b *Bridge) GetTerminalPort() int {
	return b.port
}

// Stop gracefully shuts down the bridge and all active connections.
func (b *Bridge) Stop() {
	if b.cancel != nil {
		b.cancel()
	}
}

func (b *Bridge) shutdown() {
	b.mu.Lock()
	for ws, cancel := range b.conns {
		cancel()
		ws.Close()
	}
	b.conns = make(map[*websocket.Conn]context.CancelFunc)
	b.mu.Unlock()

	if b.server != nil {
		b.server.Close()
	}
}

func (b *Bridge) trackConn(ws *websocket.Conn, cancel context.CancelFunc) {
	b.mu.Lock()
	b.conns[ws] = cancel
	b.mu.Unlock()
}

func (b *Bridge) untrackConn(ws *websocket.Conn) {
	b.mu.Lock()
	if cancel, ok := b.conns[ws]; ok {
		cancel()
		delete(b.conns, ws)
	}
	b.mu.Unlock()
}

func (b *Bridge) handleWS(w http.ResponseWriter, r *http.Request) {
	// Extract pane target from URL path: /ws/{target}
	paneTarget := strings.TrimPrefix(r.URL.Path, "/ws/")
	if paneTarget == "" {
		// Fall back to query parameter
		paneTarget = r.URL.Query().Get("pane")
	}
	if paneTarget == "" {
		http.Error(w, "missing pane target in URL path or query", http.StatusBadRequest)
		return
	}

	ws, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("terminal bridge: websocket upgrade: %v", err)
		return
	}

	connCtx, connCancel := context.WithCancel(b.ctx)
	b.trackConn(ws, connCancel)

	go b.servePane(connCtx, ws, paneTarget)
}

// servePane runs tmux attach-session under a pty and bridges I/O to the WebSocket.
// Two goroutines handle the bidirectional flow:
//   - pty reader: reads pty output into a 64KB ring buffer, flushes to WebSocket
//   - ws reader: reads WebSocket messages, writes input to pty (or handles resize)
func (b *Bridge) servePane(ctx context.Context, ws *websocket.Conn, target string) {
	defer func() {
		b.untrackConn(ws)
		ws.Close()
	}()

	// Disable tmux mouse mode so xterm.js handles text selection natively.
	// This is session-scoped and doesn't affect other terminal emulators.
	_ = exec.CommandContext(ctx, "tmux", "set-option", "-t", target, "mouse", "off").Run()

	// Send scroll history above the visible pane so the frontend has scrollback.
	// -p prints to stdout, -e preserves ANSI escape sequences for colors,
	// -S -5000 captures up to 5000 lines of history (matches xterm.js scrollback),
	// -E -1 = last line before the visible area.
	historyCmd := exec.CommandContext(ctx, "tmux", "capture-pane", "-t", target,
		"-p", "-e", "-S", "-5000", "-E", "-1")
	if histOut, err := historyCmd.Output(); err == nil {
		histOut = bytes.TrimRight(histOut, "\n")
		if len(histOut) > 0 {
			// xterm.js expects \r\n line endings for correct rendering
			histOut = bytes.ReplaceAll(histOut, []byte("\n"), []byte("\r\n"))
			histOut = append(histOut, '\r', '\n')
			if wsErr := ws.WriteMessage(websocket.BinaryMessage, histOut); wsErr != nil {
				return
			}
		}
	}

	cmd := exec.CommandContext(ctx, "tmux", "attach-session", "-t", target)
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")

	// Start with 1x1 — the frontend will send the real size immediately on connect.
	// This prevents tmux from rendering a full frame at the wrong dimensions.
	ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{Cols: 1, Rows: 1})
	if err != nil {
		sendWSClose(ws, fmt.Errorf("pty start for %s: %w", target, err))
		return
	}
	defer func() {
		ptmx.Close()
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		_ = cmd.Wait()
	}()

	obuf := newOutputBuf(outputBufMax)
	var wg sync.WaitGroup
	wg.Add(2)

	// Goroutine 1: pty → outputBuf → WebSocket.
	// Reads pty output into the ring buffer (never blocks on slow consumer),
	// then flushes accumulated bytes to the WebSocket.
	go func() {
		defer wg.Done()
		defer ws.Close() // signal ws reader to exit
		buf := make([]byte, ptyReadSize)
		for {
			n, readErr := ptmx.Read(buf)
			if n > 0 {
				obuf.Append(buf[:n])
				data := obuf.Drain()
				if err := ws.WriteMessage(websocket.BinaryMessage, data); err != nil {
					return
				}
			}
			if readErr != nil {
				if !errors.Is(readErr, io.EOF) && !errors.Is(readErr, os.ErrClosed) {
					log.Printf("terminal bridge: pty read (%s): %v", target, readErr)
				}
				return
			}
		}
	}()

	// Goroutine 2: WebSocket → pty.
	// Binary messages are terminal input written to the pty.
	// Text messages are JSON control commands (resize).
	go func() {
		defer wg.Done()
		defer ptmx.Close() // signal pty reader to exit
		for {
			msgType, msg, readErr := ws.ReadMessage()
			if readErr != nil {
				if !websocket.IsCloseError(readErr, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					log.Printf("terminal bridge: ws read (%s): %v", target, readErr)
				}
				return
			}

			switch msgType {
			case websocket.BinaryMessage:
				if _, err := ptmx.Write(msg); err != nil {
					if !errors.Is(err, os.ErrClosed) {
						log.Printf("terminal bridge: pty write (%s): %v", target, err)
					}
					return
				}
			case websocket.TextMessage:
				var rm resizeMsg
				if err := json.Unmarshal(msg, &rm); err != nil {
					continue
				}
				if rm.Type == "resize" && rm.Cols > 0 && rm.Rows > 0 {
					_ = pty.Setsize(ptmx, &pty.Winsize{
						Cols: rm.Cols,
						Rows: rm.Rows,
					})
				}
			}
		}
	}()

	wg.Wait()
}

func sendWSClose(ws *websocket.Conn, err error) {
	msg := websocket.FormatCloseMessage(websocket.CloseInternalServerErr, err.Error())
	_ = ws.WriteMessage(websocket.CloseMessage, msg)
}

// outputBuf is a bounded buffer for pty output destined for a WebSocket consumer.
// Append never blocks; if the buffer exceeds max bytes, oldest data is discarded.
// Drain returns all accumulated data and resets the buffer.
type outputBuf struct {
	mu  sync.Mutex
	buf []byte
	max int
}

func newOutputBuf(max int) *outputBuf {
	return &outputBuf{
		buf: make([]byte, 0, ptyReadSize),
		max: max,
	}
}

// Append adds data to the buffer, discarding oldest bytes if over capacity.
func (o *outputBuf) Append(p []byte) {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.buf = append(o.buf, p...)
	if len(o.buf) > o.max {
		excess := len(o.buf) - o.max
		trimmed := make([]byte, o.max)
		copy(trimmed, o.buf[excess:])
		o.buf = trimmed
	}
}

// Drain returns all buffered data and resets the buffer.
// The returned slice is owned by the caller.
func (o *outputBuf) Drain() []byte {
	o.mu.Lock()
	defer o.mu.Unlock()

	out := o.buf
	o.buf = make([]byte, 0, ptyReadSize)
	return out
}
