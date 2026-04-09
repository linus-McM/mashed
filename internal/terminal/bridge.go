package terminal

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

// Sentinel errors for the terminal bridge.
var ErrBridgeClosed = errors.New("terminal: bridge closed")

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

// Bridge serves WebSocket connections that route to ManagedSession instances
// via a SessionManager.
type Bridge struct {
	listener net.Listener
	server   *http.Server
	port     int
	manager  *SessionManager
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewBridge creates a new terminal bridge backed by the given SessionManager.
// Call Start() to begin serving.
func NewBridge(manager *SessionManager) *Bridge {
	return &Bridge{
		manager: manager,
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

// Stop gracefully shuts down the bridge.
func (b *Bridge) Stop() {
	if b.cancel != nil {
		b.cancel()
	}
}

func (b *Bridge) shutdown() {
	if b.server != nil {
		b.server.Close()
	}
}

func (b *Bridge) handleWS(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/ws/")
	if name == "" {
		http.Error(w, "missing session name", http.StatusBadRequest)
		return
	}

	session, ok := b.manager.Get(name)
	if !ok {
		http.Error(w, "session not found", http.StatusNotFound)
		return
	}

	ws, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("terminal bridge: websocket upgrade: %v", err)
		return
	}

	session.AddClient(ws)
}

