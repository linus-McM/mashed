package terminal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Sentinel errors for the terminal bridge.
var ErrBridgeClosed = errors.New("terminal: bridge closed")

// bmadSessionPrefix is the local copy of internal/bmad.SessionNamePrefix.
// We duplicate the constant rather than importing internal/bmad to avoid an
// import cycle (the bmad package and its executor depend on this terminal
// package). The compile-time test TestBridge_CompileTimeAssertBMADPrefixMatches
// in bridge_test.go imports both packages and asserts these stay in sync.
const bmadSessionPrefix = "bmad-"

// proxyReadBufSize is the read buffer used by proxyTmuxSession's reader
// goroutine. 4 KiB matches a typical PTY chunk size and minimises ws frame
// fragmentation while keeping per-attachment memory bounded.
const proxyReadBufSize = 4096

// closeReasonMaxBytes is the maximum byte length for a WebSocket close
// reason as enforced by RFC 6455 section 5.5.1.
const closeReasonMaxBytes = 123

// closeWriteTimeout bounds how long the proxy waits when writing a final
// close control frame. The value is short because the connection is about to
// be torn down anyway.
const closeWriteTimeout = time.Second

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // allow Wails webview origin
	},
}

// resizeMsgType is the discriminator value the xterm.js frontend sets on
// the resize JSON frame ({"type":"resize",...}).
const resizeMsgType = "resize"

// resizeMsg is sent by the xterm.js frontend to resize the pty.
type resizeMsg struct {
	Type string `json:"type"`
	Cols uint16 `json:"cols"`
	Rows uint16 `json:"rows"`
}

// TmuxSession is the minimal contract proxyTmuxSession needs from a live
// tmux pane attachment. *TmuxAttachment from tmux_adapter.go satisfies this
// interface implicitly. Defined in bridge.go (not tmux_adapter.go) so the
// bridge consumer owns its own contract and tests can supply pure-Go mocks
// without touching the adapter package.
type TmuxSession interface {
	io.Reader
	SendInput(data []byte) error
	SendKey(key string) error
	Resize(cols, rows uint16) error
	Close() error
}

// TmuxAttacher is the minimal contract the Bridge needs to attach to a tmux
// pane. The concrete *TmuxAdapter satisfies this after Attach was changed to
// return TmuxSession (instead of *TmuxAttachment) — see tmux_adapter.go.
// A nil TmuxAttacher means "tmux path disabled" — handleWS responds with
// 404 for any bmad-prefixed name in that mode.
type TmuxAttacher interface {
	Attach(ctx context.Context, paneTarget string) (TmuxSession, error)
}

// Bridge serves WebSocket connections that route to ManagedSession instances
// via a SessionManager, falling through to a TmuxAttacher when the requested
// session name is not in the manager but carries the BMAD prefix.
type Bridge struct {
	listener    net.Listener
	server      *http.Server
	port        int
	manager     *SessionManager
	tmuxAdapter TmuxAttacher
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewBridge creates a new terminal bridge backed by the given SessionManager
// and (optional) TmuxAttacher. A nil tmuxAdapter disables the BMAD tmux path:
// any /ws/bmad-* request that misses the SessionManager returns 404 instead
// of attempting an attach. Call Start() to begin serving.
func NewBridge(manager *SessionManager, tmuxAdapter TmuxAttacher) *Bridge {
	return &Bridge{
		manager:     manager,
		tmuxAdapter: tmuxAdapter,
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

	// PTY-managed session takes precedence — the existing AddClient path is
	// kept fully intact so shell/agent sessions are unaffected.
	if session, ok := b.manager.Get(name); ok {
		ws, err := wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("terminal bridge: websocket upgrade: %v", err)
			return
		}
		session.AddClient(ws)
		return
	}

	// On miss, fall through to the BMAD tmux path only if the adapter is
	// wired AND the name carries the BMAD prefix. Anything else is a 404 —
	// callers depending on 404 for unknown PTY sessions still work.
	if b.tmuxAdapter == nil || !strings.HasPrefix(name, bmadSessionPrefix) {
		http.Error(w, "session not found", http.StatusNotFound)
		return
	}

	ws, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("terminal bridge: websocket upgrade: %v", err)
		return
	}
	b.proxyTmuxSession(r.Context(), ws, name)
}

// proxyTmuxSession bridges a single WebSocket to a live tmux pane. It owns
// the WebSocket and the attachment for the call duration:
//
//   - On Attach error → sends a close control frame with the truncated reason
//     and returns. ws.Close runs unconditionally.
//   - On success → spawns two goroutines (att→ws and ws→att) and waits for
//     either to exit. When either side errors, both goroutines join via the
//     WaitGroup, then att.Close + ws.Close run as deferred cleanup.
//
// Concurrency model
// =================
//
// Two goroutines, one WaitGroup, one shared cancel:
//
//  1. Reader (att → ws): blocks on att.Read; writes binary frames. On exit
//     it cancels the shared ctx AND closes ws so the writer's blocked
//     ws.ReadMessage unblocks.
//  2. Writer (ws → att): blocks on ws.ReadMessage; dispatches binary frames
//     to SendInput and resize text frames to Resize. On exit it cancels
//     the shared ctx AND closes att so the reader's blocked att.Read
//     unblocks (att.Close cascades to release the underlying FIFO/poller).
//
// Both ws.Close and att.Close are idempotent, so the deferred cleanups at
// the end of the function are safe no matter which goroutine called them
// first. The cross-cancel pattern guarantees that closing either side
// always cascades to the other — there is no scenario where one goroutine
// stays blocked after the other has exited.
//
// The name parameter may be either a bare session name ("bmad-foo-…") or a
// full pane target ("bmad-foo-…:0.0"). The frontend passes TmuxTarget (the
// full form) directly via the WebSocket URL path, while tests dial with the
// bare name. normalizeBMADPaneTarget strips any existing ":window.pane"
// suffix before reappending the canonical ":0.0" so both forms resolve to
// the same pane.
func (b *Bridge) proxyTmuxSession(parentCtx context.Context, ws *websocket.Conn, name string) {
	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	paneTarget := normalizeBMADPaneTarget(name)

	att, err := b.tmuxAdapter.Attach(ctx, paneTarget)
	if err != nil {
		reason := truncateForClose(err.Error())
		closeMsg := websocket.FormatCloseMessage(websocket.CloseInternalServerErr, reason)
		_ = ws.WriteControl(websocket.CloseMessage, closeMsg, time.Now().Add(closeWriteTimeout))
		_ = ws.Close()
		return
	}
	// Defensive: TmuxAttacher's contract requires a non-nil session when err
	// is nil, but we cannot reason about third-party implementations. Without
	// this guard a buggy attacher returning (nil, nil) would panic on the
	// deferred att.Close() and the goroutines below.
	if att == nil {
		closeMsg := websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "attach returned nil session")
		_ = ws.WriteControl(websocket.CloseMessage, closeMsg, time.Now().Add(closeWriteTimeout))
		_ = ws.Close()
		return
	}
	// Idempotent cleanup — either goroutine may have already called these.
	defer func() { _ = att.Close() }()
	defer func() { _ = ws.Close() }()

	var wg sync.WaitGroup
	wg.Add(2)

	// Reader: att → ws (binary frames). On any error: cancel ctx + close ws
	// so the writer's blocked ReadMessage unblocks.
	go func() {
		defer wg.Done()
		defer func() {
			cancel()
			_ = ws.Close()
		}()
		buf := make([]byte, proxyReadBufSize)
		for {
			n, rerr := att.Read(buf)
			if n > 0 {
				if werr := ws.WriteMessage(websocket.BinaryMessage, buf[:n]); werr != nil {
					return
				}
			}
			if rerr != nil {
				return
			}
		}
	}()

	// Writer: ws → att (input + resize). On any error: cancel ctx + close
	// att so the reader's blocked Read unblocks.
	go func() {
		defer wg.Done()
		defer func() {
			cancel()
			_ = att.Close()
		}()
		for {
			mt, data, rerr := ws.ReadMessage()
			if rerr != nil {
				return
			}
			switch mt {
			case websocket.BinaryMessage:
				if serr := att.SendInput(data); serr != nil {
					return
				}
			case websocket.TextMessage:
				var msg resizeMsg
				if jerr := json.Unmarshal(data, &msg); jerr != nil {
					continue
				}
				if msg.Type == resizeMsgType {
					_ = att.Resize(msg.Cols, msg.Rows)
				}
			}
		}
	}()

	wg.Wait()
}

// truncateForClose enforces RFC 6455 section 5.5.1's 123-byte limit on the
// reason string of a WebSocket close control frame. Truncation is done at
// the byte level — UTF-8 boundaries are not respected because callers pass
// machine-generated error strings that are ASCII in practice.
func truncateForClose(reason string) string {
	if len(reason) <= closeReasonMaxBytes {
		return reason
	}
	return reason[:closeReasonMaxBytes]
}

// normalizeBMADPaneTarget accepts either a bare BMAD session name
// ("bmad-foo-main-node-deadbeef") or a full pane target carrying a
// ":window.pane" suffix ("bmad-foo-main-node-deadbeef:0.0") and returns the
// canonical pane target that BMAD guarantees exists.
//
// BMAD always launches one window with one pane via `tmux new-session`, so
// ":0.0" is the only addressable target. The frontend passes the stored
// TmuxTarget (which already includes ":0.0") directly via the WebSocket URL
// path, while unit tests dial with the bare session name. Stripping any
// existing suffix before reappending ":0.0" makes both call sites converge
// on the same target instead of producing ":0.0:0.0" for the frontend path.
func normalizeBMADPaneTarget(name string) string {
	if i := strings.Index(name, ":"); i >= 0 {
		name = name[:i]
	}
	return name + ":0.0"
}
