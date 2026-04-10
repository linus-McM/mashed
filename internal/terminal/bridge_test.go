// Package terminal tests for Bridge refactor — session routing via SessionManager.
// Story pty-03: Bridge Refactor — Route WebSockets to Managed Sessions
//
// RED Phase: These tests define expected behavior for all 5 ACs and 4 BDD scenarios.
// They should FAIL until the go-engineer refactors bridge.go.

package terminal

// All tests in this file require a real PTY (creack/pty fork/exec).
// Add testing.Short() skip guard to any new test functions.

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bmadpkg "mashed/internal/bmad"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// startBridgeWithManager creates a SessionManager and Bridge, starts the bridge,
// and registers cleanup. Returns the manager and bridge for further setup.
func startBridgeWithManager(t *testing.T, ctx context.Context) (*SessionManager, *Bridge) {
	t.Helper()

	sm := NewSessionManager(nil)
	b := NewBridge(sm, nil)
	require.NotNil(t, b, "NewBridge(manager) should return non-nil bridge")

	err := b.Start(ctx)
	require.NoError(t, err, "Bridge.Start should not error")

	t.Cleanup(func() {
		b.Stop()
		sm.Shutdown()
	})

	return sm, b
}

// dialBridgeWS dials the bridge's WebSocket endpoint for the given session name.
func dialBridgeWS(t *testing.T, b *Bridge, sessionName string) *websocket.Conn {
	t.Helper()

	port := b.GetTerminalPort()
	require.Greater(t, port, 0, "bridge port should be positive")

	url := fmt.Sprintf("ws://127.0.0.1:%d/ws/%s", port, sessionName)
	dialer := websocket.Dialer{}
	ws, _, err := dialer.Dial(url, nil)
	require.NoError(t, err, "WebSocket dial to %s should succeed", url)

	t.Cleanup(func() { ws.Close() })
	return ws
}

// readWSMessage reads one binary message from a WebSocket with a timeout.
func readWSMessage(t *testing.T, ws *websocket.Conn, timeout time.Duration) []byte {
	t.Helper()
	_ = ws.SetReadDeadline(time.Now().Add(timeout))
	_, data, err := ws.ReadMessage()
	require.NoError(t, err, "should read WS message within timeout")
	return data
}

// ---------------------------------------------------------------------------
// AC-1: Bridge constructor accepts a SessionManager (BDD Scenario 4)
// ---------------------------------------------------------------------------

func TestBridge_AC1_NewBridgeWithManager(t *testing.T) {
	if testing.Short() {
		t.Skip("requires PTY (fork/exec)")
	}
	sm := NewSessionManager(nil)
	defer sm.Shutdown()

	b := NewBridge(sm, nil)
	require.NotNil(t, b, "NewBridge should return a non-nil bridge")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := b.Start(ctx)
	require.NoError(t, err, "Bridge.Start should succeed")

	port := b.GetTerminalPort()
	assert.Greater(t, port, 0, "port should be positive after Start")

	b.Stop()
}

// ---------------------------------------------------------------------------
// AC-2: WebSocket connections route to managed sessions (BDD Scenario 1)
// ---------------------------------------------------------------------------

func TestBridge_AC2_WSRoutesToSession(t *testing.T) {
	t.Skip("requires running pty-helper binary (PTY spawn now delegated to helper process)")
	if testing.Short() {
		t.Skip("requires PTY (fork/exec)")
	}
	if os.Getenv("CI") != "" {
		t.Skip("requires real PTY, skipping in CI")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sm, b := startBridgeWithManager(t, ctx)

	_, err := sm.Spawn(ctx, "test-sess", t.TempDir(), "echo hello-bridge && cat")
	require.NoError(t, err, "Spawn should succeed")

	time.Sleep(outputSettleTime)

	ws := dialBridgeWS(t, b, "test-sess")

	data := readWSMessage(t, ws, wsReadTimeout)
	assert.Contains(t, string(data), "hello-bridge",
		"WS client should receive PTY output from the session")
}

// ---------------------------------------------------------------------------
// AC-3: Unknown/empty session names return proper HTTP errors (BDD Scenario 1b)
// ---------------------------------------------------------------------------

func TestBridge_AC3_ErrorResponses(t *testing.T) {
	if testing.Short() {
		t.Skip("requires PTY (fork/exec)")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, b := startBridgeWithManager(t, ctx)

	port := b.GetTerminalPort()
	require.Greater(t, port, 0, "bridge port should be positive")

	tests := []struct {
		name       string
		path       string
		wantStatus int
	}{
		{
			name:       "unknown session returns 404",
			path:       "/ws/ghost",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "empty session name returns 400",
			path:       "/ws/",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("http://127.0.0.1:%d%s", port, tt.path)
			resp, err := http.Get(url)
			require.NoError(t, err, "HTTP GET should not error")
			defer resp.Body.Close()

			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

// ---------------------------------------------------------------------------
// AC-4 (pty-03): All direct tmux exec is removed from bridge.go.
//
// Updated for bridge-03: the bridge now holds a TmuxAttacher interface and
// delegates to it, but it must NOT exec tmux directly. The literal "tmux"
// substring ban from pty-03 was relaxed because bridge.go legitimately
// references tmux types (TmuxSession, TmuxAttacher, tmuxAdapter). The
// intent — keep direct tmux command invocations out of bridge.go — is
// preserved by the per-subcommand checks below.
// ---------------------------------------------------------------------------

func TestBridge_AC4_NoTmuxReferences(t *testing.T) {
	if testing.Short() {
		t.Skip("requires PTY (fork/exec)")
	}
	bridgeSrc, err := os.ReadFile("bridge.go")
	require.NoError(t, err, "should be able to read bridge.go")

	src := string(bridgeSrc)

	forbidden := []struct {
		pattern string
		reason  string
	}{
		{"outputBuf", "outputBuf type should be removed (replaced by scrollBuffer)"},
		{"servePane", "servePane method should be removed (replaced by AddClient)"},
		{"attach-session", "tmux attach-session exec should be removed"},
		{"capture-pane", "tmux capture-pane exec should be removed"},
		{"send-keys", "tmux send-keys exec must stay out of bridge.go (delegated to TmuxAttacher)"},
		{"pipe-pane", "tmux pipe-pane exec must stay out of bridge.go (delegated to TmuxAttacher)"},
		{"exec.Command", "no exec.Command in bridge.go — tmux work goes through the adapter"},
		{"ErrPaneTargetEmpty", "pane target sentinel should be removed"},
		{"type outputBuf struct", "outputBuf type definition should be removed"},
		{"func newOutputBuf", "newOutputBuf constructor should be removed"},
		{"func (o *outputBuf) Append", "outputBuf.Append method should be removed"},
		{"func (o *outputBuf) Drain", "outputBuf.Drain method should be removed"},
	}

	for _, f := range forbidden {
		assert.False(t, strings.Contains(src, f.pattern),
			"bridge.go should not contain %q: %s", f.pattern, f.reason)
	}
}

// ---------------------------------------------------------------------------
// AC-5: Multiple clients can connect to the same session (BDD Scenario 3)
// ---------------------------------------------------------------------------

func TestBridge_AC5_MultipleClients(t *testing.T) {
	t.Skip("requires running pty-helper binary (PTY spawn now delegated to helper process)")
	if testing.Short() {
		t.Skip("requires PTY (fork/exec)")
	}
	if os.Getenv("CI") != "" {
		t.Skip("requires real PTY, skipping in CI")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sm, b := startBridgeWithManager(t, ctx)

	_, err := sm.Spawn(ctx, "shared", t.TempDir(), "cat")
	require.NoError(t, err, "Spawn should succeed")
	time.Sleep(outputSettleTime)

	wsA := dialBridgeWS(t, b, "shared")
	wsB := dialBridgeWS(t, b, "shared")

	err = wsA.WriteMessage(websocket.BinaryMessage, []byte("multi-test\n"))
	require.NoError(t, err, "client A write should succeed")

	// Read from both clients concurrently with blocking reads.
	var gotA, gotB bool
	var wg sync.WaitGroup
	wg.Add(2)

	readUntilMatch := func(ws *websocket.Conn, match string, found *bool) {
		defer wg.Done()
		_ = ws.SetReadDeadline(time.Now().Add(wsReadTimeout))
		for i := 0; i < 5; i++ {
			_, data, err := ws.ReadMessage()
			if err != nil {
				return
			}
			if strings.Contains(string(data), match) {
				*found = true
				return
			}
		}
	}

	go readUntilMatch(wsA, "multi-test", &gotA)
	go readUntilMatch(wsB, "multi-test", &gotB)

	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("timed out waiting for both clients to receive output")
	}

	assert.True(t, gotA, "client A should receive echoed output")
	assert.True(t, gotB, "client B should receive echoed output")
}

// ---------------------------------------------------------------------------
// BDD Scenario 4: Bridge start/stop lifecycle
// ---------------------------------------------------------------------------

func TestBridge_StartStop(t *testing.T) {
	if testing.Short() {
		t.Skip("requires PTY (fork/exec)")
	}
	sm := NewSessionManager(nil)
	defer sm.Shutdown()

	b := NewBridge(sm, nil)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := b.Start(ctx)
	require.NoError(t, err, "Bridge.Start should succeed")

	port := b.GetTerminalPort()
	assert.Greater(t, port, 0, "GetTerminalPort should return positive port after Start")

	// Verify the server accepts HTTP connections.
	url := fmt.Sprintf("http://127.0.0.1:%d/ws/", port)
	resp, err := http.Get(url)
	require.NoError(t, err, "HTTP request to running bridge should not error")
	resp.Body.Close()

	b.Stop()
	time.Sleep(100 * time.Millisecond)

	// Server should no longer accept connections.
	_, err = http.Get(url)
	assert.Error(t, err, "HTTP request to stopped bridge should error (connection refused)")
}

func TestBridge_StopIdempotent(t *testing.T) {
	if testing.Short() {
		t.Skip("requires PTY (fork/exec)")
	}
	sm := NewSessionManager(nil)
	defer sm.Shutdown()

	b := NewBridge(sm, nil)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := b.Start(ctx)
	require.NoError(t, err)

	assert.NotPanics(t, func() {
		b.Stop()
		b.Stop()
	}, "Stop should be idempotent and not panic on double-call")
}

// ---------------------------------------------------------------------------
// bridge-03: TmuxAdapter routing tests
//
// Interface design decision — Option A:
// These tests assume `NewBridge` takes a second parameter `TmuxAttacher`
// (interface defined in bridge.go) whose `Attach` returns a `TmuxSession`
// (also an interface defined in bridge.go). Rationale:
//
//  1. An interface lets tests inject a pure-Go mock with no tmux dependency.
//  2. Putting the interface in bridge.go — not tmux_adapter.go — avoids
//     coupling bridge-02 to the bridge consumer and keeps the bridge-side
//     contract local to the file that uses it.
//  3. The concrete *TmuxAttachment already has the required method set
//     (Read, SendInput, SendKey, Resize, Close), so it implicitly satisfies
//     TmuxSession — the go-engineer in Task #2 must change
//     *TmuxAdapter.Attach to return `TmuxSession` (or add a thin adapter
//     method). Either way, the bridge holds a `TmuxAttacher` field.
//
// If Option A creates friction, the go-engineer may adapt: the failing
// tests describe the contract, and the RED compile errors point at the
// missing symbols.
// ---------------------------------------------------------------------------

// mockResizeCall captures one Resize(cols, rows) invocation for assertions.
type mockResizeCall struct {
	cols uint16
	rows uint16
}

// mockTmuxSession is a pure-Go stand-in for the TmuxSession interface. It
// records every call and exposes channels so tests can synchronise without
// time.Sleep. The readCh drives canned bytes into the bridge's reader
// goroutine; closeCh signals that the bridge called Close.
type mockTmuxSession struct {
	inputCh  chan []byte
	keyCh    chan string
	resizeCh chan mockResizeCall
	closeCh  chan struct{}
	readCh   chan []byte
	closed   atomic.Bool
}

func newMockTmuxSession() *mockTmuxSession {
	return &mockTmuxSession{
		inputCh:  make(chan []byte, 16),
		keyCh:    make(chan string, 16),
		resizeCh: make(chan mockResizeCall, 16),
		closeCh:  make(chan struct{}),
		readCh:   make(chan []byte, 16),
	}
}

// Read returns the whole chunk in one call. That is fine for these tests
// because every canned payload is smaller than the bridge's proxy read buffer.
func (s *mockTmuxSession) Read(p []byte) (int, error) {
	if s.closed.Load() {
		return 0, io.EOF
	}
	chunk, ok := <-s.readCh
	if !ok {
		return 0, io.EOF
	}
	n := copy(p, chunk)
	return n, nil
}

func (s *mockTmuxSession) SendInput(data []byte) error {
	cp := make([]byte, len(data))
	copy(cp, data)
	select {
	case s.inputCh <- cp:
	default:
		// drop silently — tests use sufficient buffer (16)
	}
	return nil
}

func (s *mockTmuxSession) SendKey(key string) error {
	select {
	case s.keyCh <- key:
	default:
	}
	return nil
}

func (s *mockTmuxSession) Resize(cols, rows uint16) error {
	select {
	case s.resizeCh <- mockResizeCall{cols: cols, rows: rows}:
	default:
	}
	return nil
}

func (s *mockTmuxSession) Close() error {
	if s.closed.CompareAndSwap(false, true) {
		close(s.closeCh)
		close(s.readCh)
	}
	return nil
}

// mockTmuxAttacher implements TmuxAttacher and records every Attach call.
// attachErr, if non-nil, is returned from Attach instead of the session.
// attachFn, if non-nil, fully overrides the default Attach behavior — used
// by tests that need to return arbitrary (TmuxSession, error) tuples (e.g.
// the (nil, nil) regression test).
// attachCh fires once per Attach call so tests can wait for the routing
// decision event-driven instead of polling.
type mockTmuxAttacher struct {
	mu          sync.Mutex
	attachCount atomic.Int32
	attachErr   error
	attachFn    func(ctx context.Context, paneTarget string) (TmuxSession, error)
	lastTarget  string
	session     *mockTmuxSession
	attachCh    chan string
}

func newMockTmuxAttacher(session *mockTmuxSession) *mockTmuxAttacher {
	return &mockTmuxAttacher{
		session:  session,
		attachCh: make(chan string, 4),
	}
}

func (m *mockTmuxAttacher) Attach(ctx context.Context, paneTarget string) (TmuxSession, error) {
	m.attachCount.Add(1)
	m.mu.Lock()
	m.lastTarget = paneTarget
	m.mu.Unlock()
	select {
	case m.attachCh <- paneTarget:
	default:
	}
	if m.attachFn != nil {
		return m.attachFn(ctx, paneTarget)
	}
	if m.attachErr != nil {
		return nil, m.attachErr
	}
	return m.session, nil
}

func (m *mockTmuxAttacher) LastTarget() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.lastTarget
}

// startBridgeWithMockAdapter wires a SessionManager + Bridge to the given
// attacher, starts the bridge, and registers cleanup. It is intentionally
// separate from startBridgeWithManager so existing tests remain untouched.
func startBridgeWithMockAdapter(t *testing.T, ctx context.Context, adapter TmuxAttacher) (*SessionManager, *Bridge) {
	t.Helper()
	sm := NewSessionManager(nil)
	b := NewBridge(sm, adapter)
	require.NotNil(t, b, "NewBridge(sm, adapter) should return non-nil")
	require.NoError(t, b.Start(ctx), "Bridge.Start should not error")
	t.Cleanup(func() {
		b.Stop()
		sm.Shutdown()
	})
	return sm, b
}

// dialBridgeWSCtx dials the bridge's /ws/{name} endpoint using the given
// context, registering a Cleanup to close the returned conn.
func dialBridgeWSCtx(t *testing.T, ctx context.Context, b *Bridge, name string) *websocket.Conn {
	t.Helper()
	port := b.GetTerminalPort()
	require.Greater(t, port, 0, "bridge port should be positive")
	url := fmt.Sprintf("ws://127.0.0.1:%d/ws/%s", port, name)
	ws, _, err := websocket.DefaultDialer.DialContext(ctx, url, nil)
	require.NoError(t, err, "WebSocket dial to %s should succeed", url)
	t.Cleanup(func() { _ = ws.Close() })
	return ws
}

// ---------------------------------------------------------------------------
// AC-1: NewBridge accepts both SessionManager and TmuxAttacher
// ---------------------------------------------------------------------------

func TestBridge_AC1_NewBridgeAcceptsAdapter(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	adapter := newMockTmuxAttacher(newMockTmuxSession())
	_, b := startBridgeWithMockAdapter(t, ctx, adapter)

	assert.Greater(t, b.GetTerminalPort(), 0,
		"GetTerminalPort should return a positive port after Start")
}

// ---------------------------------------------------------------------------
// AC-2: PTY session takes precedence over the tmux path (BDD Scenario 1a)
// ---------------------------------------------------------------------------

func TestBridge_AC2_PTYSessionTakesPrecedence(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	adapter := newMockTmuxAttacher(newMockTmuxSession())
	sm, b := startBridgeWithMockAdapter(t, ctx, adapter)

	// Inject a minimal-viable ManagedSession directly into the manager map.
	// The goal is only to force the PTY branch; the session need not be
	// fully functional. os.Pipe() gives us a valid *os.File so that the
	// subsequent sm.Shutdown() → ms.Kill() → ptmx.Close() cascade does not
	// nil-deref.
	_, ptmxW, err := os.Pipe()
	require.NoError(t, err)
	fakeSession := &ManagedSession{
		name:    "my-shell",
		ptmx:    ptmxW,
		clients: make(map[*websocket.Conn]*sync.Mutex),
		scroll:  newScrollBuffer(scrollBufferDefaultCap),
		done:    make(chan struct{}),
	}
	sm.mu.Lock()
	sm.sessions["my-shell"] = fakeSession
	sm.mu.Unlock()

	// Dial and immediately close — the bridge should route through
	// AddClient (the PTY path) and never touch the adapter.
	ws := dialBridgeWSCtx(t, ctx, b, "my-shell")
	_ = ws.Close()

	// Race the (forbidden) Attach signal against a bounded window. Asserting
	// a negative needs a passive wait; the channel makes the wait
	// event-driven instead of a CPU-pinning busy-loop.
	select {
	case target := <-adapter.attachCh:
		t.Fatalf("adapter.Attach was called with %q — PTY path should have taken precedence", target)
	case <-time.After(200 * time.Millisecond):
		// PASS — adapter was not invoked within the window.
	}
	assert.Equal(t, int32(0), adapter.attachCount.Load(),
		"adapter.Attach must NOT be invoked when a PTY session matches the name")
}

// ---------------------------------------------------------------------------
// AC-3: BMAD-prefixed session routes through TmuxAdapter (BDD Scenario 1b)
// ---------------------------------------------------------------------------

func TestBridge_AC3_BMADPrefixRoutesToAdapter(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	mockSess := newMockTmuxSession()
	// Pre-load the canned payload; the bridge's reader goroutine will
	// pick it up as soon as it calls session.Read.
	mockSess.readCh <- []byte("line1\nline2\n")

	adapter := newMockTmuxAttacher(mockSess)
	_, b := startBridgeWithMockAdapter(t, ctx, adapter)

	sessionName := "bmad-foo-main-node-deadbeef"
	ws := dialBridgeWSCtx(t, ctx, b, sessionName)

	_ = ws.SetReadDeadline(time.Now().Add(1500 * time.Millisecond))
	mt, data, err := ws.ReadMessage()
	require.NoError(t, err, "client should receive the canned payload")
	assert.Equal(t, websocket.BinaryMessage, mt,
		"tmux output must be forwarded as binary frames")
	assert.Equal(t, "line1\nline2\n", string(data))

	assert.Equal(t, int32(1), adapter.attachCount.Load(),
		"adapter.Attach must be invoked exactly once")
	assert.Equal(t, sessionName+":0.0", adapter.LastTarget(),
		"pane target must be <sessionName>:0.0 (BMAD window 0, pane 0)")
}

// ---------------------------------------------------------------------------
// AC-4: Non-BMAD miss still returns 404 (BDD Scenario 1c)
// ---------------------------------------------------------------------------

func TestBridge_AC4_NonBMADMissReturns404(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	adapter := newMockTmuxAttacher(newMockTmuxSession())
	_, b := startBridgeWithMockAdapter(t, ctx, adapter)

	url := fmt.Sprintf("http://127.0.0.1:%d/ws/randomname", b.GetTerminalPort())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode,
		"unknown non-BMAD session must return 404")
	assert.Equal(t, int32(0), adapter.attachCount.Load(),
		"adapter.Attach must NOT be invoked for non-BMAD names")
}

// ---------------------------------------------------------------------------
// AC-5: Binary input frame forwards to SendInput (BDD Scenario 2b)
// ---------------------------------------------------------------------------

func TestBridge_AC5_BinaryInputFrameForwardsSendInput(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	mockSess := newMockTmuxSession()
	adapter := newMockTmuxAttacher(mockSess)
	_, b := startBridgeWithMockAdapter(t, ctx, adapter)

	ws := dialBridgeWSCtx(t, ctx, b, "bmad-input-test")

	require.NoError(t, ws.WriteMessage(websocket.BinaryMessage, []byte("hello")),
		"WS binary write should succeed")

	select {
	case got := <-mockSess.inputCh:
		assert.Equal(t, []byte("hello"), got,
			"SendInput must receive the exact binary frame payload")
	case <-time.After(1500 * time.Millisecond):
		t.Fatal("timed out waiting for SendInput to be called")
	}
}

// ---------------------------------------------------------------------------
// AC-6: Resize text frame forwards to Resize (BDD Scenario 2c)
// ---------------------------------------------------------------------------

func TestBridge_AC6_ResizeTextFrameForwardsResize(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	mockSess := newMockTmuxSession()
	adapter := newMockTmuxAttacher(mockSess)
	_, b := startBridgeWithMockAdapter(t, ctx, adapter)

	ws := dialBridgeWSCtx(t, ctx, b, "bmad-resize-test")

	payload, err := json.Marshal(resizeMsg{Type: "resize", Cols: 120, Rows: 40})
	require.NoError(t, err)
	require.NoError(t, ws.WriteMessage(websocket.TextMessage, payload),
		"WS text write should succeed")

	select {
	case got := <-mockSess.resizeCh:
		assert.Equal(t, mockResizeCall{cols: 120, rows: 40}, got,
			"Resize must receive the parsed cols/rows from the resize text frame")
	case <-time.After(1500 * time.Millisecond):
		t.Fatal("timed out waiting for Resize to be called")
	}
}

// ---------------------------------------------------------------------------
// AC-7: Attach error closes the WebSocket (BDD Scenario 3a)
// ---------------------------------------------------------------------------

func TestBridge_AC7_AttachErrorClosesWebSocket(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	adapter := newMockTmuxAttacher(newMockTmuxSession())
	adapter.attachErr = fmt.Errorf("%w: bmad-dead pane", ErrPaneDead)
	_, b := startBridgeWithMockAdapter(t, ctx, adapter)

	ws := dialBridgeWSCtx(t, ctx, b, "bmad-dead")

	// After an Attach failure the bridge should send a close frame and
	// tear down the connection. The client's next ReadMessage should
	// therefore return an error (close frame, EOF, or use-of-closed).
	_ = ws.SetReadDeadline(time.Now().Add(1500 * time.Millisecond))
	_, _, err := ws.ReadMessage()
	assert.Error(t, err, "client ReadMessage should error after server-side close")
	assert.Equal(t, int32(1), adapter.attachCount.Load(),
		"Attach must have been attempted exactly once")
}

// TestBridge_ProxyTmuxSessionHandlesNilSessionFromAttach defends against a
// buggy TmuxAttacher implementation that returns (nil, nil) from Attach in
// violation of the interface contract. The bridge must not panic on the
// deferred Close — instead it should send a close frame and tear down.
func TestBridge_ProxyTmuxSessionHandlesNilSessionFromAttach(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	adapter := &mockTmuxAttacher{
		attachCh: make(chan string, 1),
		attachFn: func(ctx context.Context, paneTarget string) (TmuxSession, error) {
			return nil, nil //nolint:nilnil // intentional contract violation
		},
	}
	_, b := startBridgeWithMockAdapter(t, ctx, adapter)

	ws := dialBridgeWSCtx(t, ctx, b, "bmad-nil-session")

	// The bridge must not panic on the deferred att.Close() despite att
	// being nil; the client should observe a normal close-frame teardown.
	_ = ws.SetReadDeadline(time.Now().Add(1500 * time.Millisecond))
	_, _, err := ws.ReadMessage()
	assert.Error(t, err, "client ReadMessage should error after server-side close")
	assert.Equal(t, int32(1), adapter.attachCount.Load(),
		"Attach must have been attempted exactly once")
}

// ---------------------------------------------------------------------------
// AC-8: WS close cancels the attachment (BDD Scenario 3b)
// ---------------------------------------------------------------------------

func TestBridge_AC8_WebSocketCloseCancelsAttachment(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	mockSess := newMockTmuxSession()
	adapter := newMockTmuxAttacher(mockSess)
	_, b := startBridgeWithMockAdapter(t, ctx, adapter)

	ws := dialBridgeWSCtx(t, ctx, b, "bmad-close-test")

	// Deliver one frame first so we are certain both proxy goroutines are
	// live before we tear the WebSocket down.
	mockSess.readCh <- []byte("bootstrap")
	_ = ws.SetReadDeadline(time.Now().Add(1500 * time.Millisecond))
	_, _, err := ws.ReadMessage()
	require.NoError(t, err, "client must read the bootstrap frame before closing")

	// Clean WS close from the client side.
	_ = ws.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, "client done"),
	)
	_ = ws.Close()

	select {
	case <-mockSess.closeCh:
		// Success — the bridge called attachment.Close().
	case <-time.After(1500 * time.Millisecond):
		t.Fatal("attachment.Close() was not called within 1.5s of WS close")
	}
}

// ---------------------------------------------------------------------------
// Nil adapter degrades gracefully (BDD Scenario 3c)
// ---------------------------------------------------------------------------

func TestBridge_NilAdapterDegradesGracefully(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Pass a typed nil through the helper. The bridge must accept nil and
	// then serve 404 for any bmad-prefixed name (no panic, no upgrade).
	_, b := startBridgeWithMockAdapter(t, ctx, nil)

	url := fmt.Sprintf("http://127.0.0.1:%d/ws/bmad-anything", b.GetTerminalPort())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode,
		"bridge with nil adapter must serve 404 (not 500, not an upgrade attempt) for bmad-prefixed names")
}

// ---------------------------------------------------------------------------
// Drift guard: local bmadSessionPrefix must match internal/bmad.SessionNamePrefix
// ---------------------------------------------------------------------------

func TestBridge_CompileTimeAssertBMADPrefixMatches(t *testing.T) {
	// bridge.go declares its own bmadSessionPrefix constant to avoid an
	// import cycle with internal/bmad. This test fails the build if those
	// two constants ever drift apart.
	assert.Equal(t, bmadpkg.SessionNamePrefix, bmadSessionPrefix,
		"local bmadSessionPrefix must match internal/bmad.SessionNamePrefix")
}

// ---------------------------------------------------------------------------
// AC-9 / AC-10 — App startup wiring (cross-referenced, not implemented here)
//
// AC-9: tmux on PATH  → app.go wires NewBridge with a non-nil TmuxAdapter.
// AC-10: tmux missing → app.go wires NewBridge with nil adapter + warning.
//
// These ACs live in app.go/OnStartup and are exercised either manually or
// from app_test.go (if/when one is added). They are intentionally NOT
// duplicated here — this file would otherwise have to import Wails
// runtime primitives that are unhelpful for the terminal package.
// ---------------------------------------------------------------------------
