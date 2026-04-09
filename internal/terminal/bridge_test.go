// Package terminal tests for Bridge refactor — session routing via SessionManager.
// Story pty-03: Bridge Refactor — Route WebSockets to Managed Sessions
//
// RED Phase: These tests define expected behavior for all 5 ACs and 4 BDD scenarios.
// They should FAIL until the go-engineer refactors bridge.go.

package terminal

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// startBridgeWithManager creates a SessionManager and Bridge, starts the bridge,
// and registers cleanup. Returns the manager and bridge for further setup.
func startBridgeWithManager(t *testing.T, ctx context.Context) (*SessionManager, *Bridge) {
	t.Helper()

	sm := NewSessionManager()
	b := NewBridge(sm)
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
	sm := NewSessionManager()
	defer sm.Shutdown()

	b := NewBridge(sm)
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
// AC-4: All tmux code is removed from bridge.go (BDD Scenario 2)
// ---------------------------------------------------------------------------

func TestBridge_AC4_NoTmuxReferences(t *testing.T) {
	bridgeSrc, err := os.ReadFile("bridge.go")
	require.NoError(t, err, "should be able to read bridge.go")

	src := string(bridgeSrc)

	forbidden := []struct {
		pattern string
		reason  string
	}{
		{"tmux", "no tmux references should remain"},
		{"outputBuf", "outputBuf type should be removed (replaced by scrollBuffer)"},
		{"servePane", "servePane method should be removed (replaced by AddClient)"},
		{"attach-session", "tmux attach-session exec should be removed"},
		{"capture-pane", "tmux capture-pane exec should be removed"},
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
	sm := NewSessionManager()
	defer sm.Shutdown()

	b := NewBridge(sm)

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
	sm := NewSessionManager()
	defer sm.Shutdown()

	b := NewBridge(sm)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := b.Start(ctx)
	require.NoError(t, err)

	assert.NotPanics(t, func() {
		b.Stop()
		b.Stop()
	}, "Stop should be idempotent and not panic on double-call")
}
