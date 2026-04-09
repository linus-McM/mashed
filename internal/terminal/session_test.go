// Package terminal tests for ManagedSession and scrollBuffer.
// Story pty-01: ManagedSession with Scrollback Ring Buffer
//
// RED Phase: These tests define expected behavior and MUST FAIL
// until the go-engineer implements session.go.

package terminal

// All tests in this file require a real PTY (creack/pty fork/exec).
// Add testing.Short() skip guard to any new test functions.

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	// wsReadTimeout is the default deadline for reading a WebSocket message in tests.
	wsReadTimeout = 5 * time.Second
	// outputSettleTime allows PTY output to be buffered before a late client connects.
	outputSettleTime = 500 * time.Millisecond
	// processExitWait allows a short-lived process to exit and EOF to propagate.
	processExitWait = 1 * time.Second
	// clientEvictWait allows the session to detect and evict a failed client.
	clientEvictWait = 300 * time.Millisecond
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// wsTestPair creates a connected WebSocket pair for testing.
// Returns the server-side conn (passed to AddClient) and client-side conn
// (used by the test to read data the session sends).
func wsTestPair(t *testing.T) (server *websocket.Conn, client *websocket.Conn) {
	t.Helper()

	serverCh := make(chan *websocket.Conn, 1)
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("ws upgrade failed: %v", err)
			return
		}
		serverCh <- c
	}))
	t.Cleanup(s.Close)

	wsURL := "ws" + strings.TrimPrefix(s.URL, "http")
	c, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err, "ws dial")
	t.Cleanup(func() { c.Close() })

	sc := <-serverCh
	t.Cleanup(func() { sc.Close() })

	return sc, c
}

// startTestSession starts a real PTY process and wraps it in a ManagedSession.
// Cleanup kills the process group and closes the PTY even if the stub Kill is a no-op.
func startTestSession(t *testing.T, shellCmd string) *ManagedSession {
	t.Helper()

	cmd := exec.Command("/bin/sh", "-c", shellCmd)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	ptmx, err := pty.Start(cmd)
	require.NoError(t, err, "pty start")

	sess := newManagedSession("test", cmd, ptmx)
	t.Cleanup(func() {
		sess.Kill() // may be a stub no-op
		// Fallback: ensure the real process is killed even with stubs.
		if cmd.Process != nil {
			_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
		}
		ptmx.Close()
		_ = cmd.Wait()
	})
	return sess
}

// readWSTimeout reads one WebSocket message with a deadline.
func readWSTimeout(t *testing.T, conn *websocket.Conn, timeout time.Duration) ([]byte, error) {
	t.Helper()
	_ = conn.SetReadDeadline(time.Now().Add(timeout))
	_, data, err := conn.ReadMessage()
	return data, err
}

// ---------------------------------------------------------------------------
// AC-1: Scrollback Ring Buffer (BDD Scenario 1)
// ---------------------------------------------------------------------------

func TestScrollBuffer_AC1_WriteAndSnapshot(t *testing.T) {
	if testing.Short() {
		t.Skip("requires PTY (fork/exec)")
	}
	t.Parallel()

	tests := []struct {
		name      string
		capacity  int
		writes    [][]byte
		wantLen   int
		wantBytes []byte
	}{
		{
			name:      "empty buffer returns empty snapshot",
			capacity:  1024,
			writes:    nil,
			wantLen:   0,
			wantBytes: []byte{},
		},
		{
			name:      "under capacity returns all data",
			capacity:  1024,
			writes:    [][]byte{[]byte("hello")},
			wantLen:   5,
			wantBytes: []byte("hello"),
		},
		{
			name:      "exact capacity returns all data",
			capacity:  10,
			writes:    [][]byte{[]byte("0123456789")},
			wantLen:   10,
			wantBytes: []byte("0123456789"),
		},
		{
			name:      "over capacity keeps most recent data",
			capacity:  10,
			writes:    [][]byte{[]byte("abcdefghij"), []byte("KLMNO")},
			wantLen:   10,
			wantBytes: []byte("fghijKLMNO"),
		},
		{
			name:      "BDD: wraps at 1024 when 1500 written",
			capacity:  1024,
			writes:    [][]byte{bytes.Repeat([]byte("A"), 1024), bytes.Repeat([]byte("B"), 476)},
			wantLen:   1024,
			wantBytes: append(bytes.Repeat([]byte("A"), 548), bytes.Repeat([]byte("B"), 476)...),
		},
		{
			name:      "multiple small writes accumulate",
			capacity:  1024,
			writes:    [][]byte{[]byte("aaa"), []byte("bbb"), []byte("ccc")},
			wantLen:   9,
			wantBytes: []byte("aaabbbccc"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sb := newScrollBuffer(tt.capacity)
			for _, w := range tt.writes {
				sb.Write(w)
			}

			snap := sb.Snapshot()
			require.NotNil(t, snap, "Snapshot must never return nil")
			assert.Equal(t, tt.wantLen, len(snap), "snapshot length")
			assert.Equal(t, tt.wantBytes, snap, "snapshot content")
		})
	}
}

func TestScrollBuffer_AC1_LargeOverCapacity(t *testing.T) {
	if testing.Short() {
		t.Skip("requires PTY (fork/exec)")
	}
	t.Parallel()

	// Write 2MB to a 1MB buffer — only the second 1MB should survive.
	capacity := 1 << 20
	sb := newScrollBuffer(capacity)

	first := bytes.Repeat([]byte("X"), capacity)
	sb.Write(first)
	second := bytes.Repeat([]byte("Y"), capacity)
	sb.Write(second)

	snap := sb.Snapshot()
	require.NotNil(t, snap, "Snapshot must never return nil")
	assert.Equal(t, capacity, len(snap), "snapshot should be exactly 1MB")
	assert.Equal(t, second, snap, "snapshot should contain only the most recent 1MB")
}

func TestScrollBuffer_AC1_SnapshotIndependentCopy(t *testing.T) {
	if testing.Short() {
		t.Skip("requires PTY (fork/exec)")
	}
	t.Parallel()

	sb := newScrollBuffer(1024)
	sb.Write([]byte("hello"))

	snap1 := sb.Snapshot()
	require.NotNil(t, snap1, "first Snapshot must not be nil")
	require.Equal(t, []byte("hello"), snap1)

	// Mutate the returned slice.
	for i := range snap1 {
		snap1[i] = 'Z'
	}

	snap2 := sb.Snapshot()
	require.NotNil(t, snap2, "second Snapshot must not be nil")
	assert.Equal(t, []byte("hello"), snap2, "mutating first snapshot must not affect the buffer")
}

// ---------------------------------------------------------------------------
// AC-2: Multi-Client Fan-Out (BDD Scenario 2)
// ---------------------------------------------------------------------------

func TestManagedSession_AC2_TwoClientsReceiveOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("requires PTY (fork/exec)")
	}
	// BDD Scenario 2: Two clients receive same output
	sess := startTestSession(t, "echo hello")

	srv1, cli1 := wsTestPair(t)
	srv2, cli2 := wsTestPair(t)

	sess.AddClient(srv1)
	sess.AddClient(srv2)

	data1, err1 := readWSTimeout(t, cli1, wsReadTimeout)
	require.NoError(t, err1, "client 1 should receive data")
	assert.Contains(t, string(data1), "hello", "client 1 output should contain 'hello'")

	data2, err2 := readWSTimeout(t, cli2, wsReadTimeout)
	require.NoError(t, err2, "client 2 should receive data")
	assert.Contains(t, string(data2), "hello", "client 2 output should contain 'hello'")
}

// ---------------------------------------------------------------------------
// AC-3: Atomic Scrollback Replay (BDD Scenario 3)
// ---------------------------------------------------------------------------

func TestManagedSession_AC3_NewClientGetsSnapshot(t *testing.T) {
	if testing.Short() {
		t.Skip("requires PTY (fork/exec)")
	}
	// BDD Scenario 3: New client gets scrollback then live data
	sess := startTestSession(t, "printf 'line1\\nline2\\n'; sleep 10")

	time.Sleep(outputSettleTime)

	srv, cli := wsTestPair(t)
	sess.AddClient(srv)

	data, err := readWSTimeout(t, cli, wsReadTimeout)
	require.NoError(t, err, "late client should receive snapshot")
	assert.Contains(t, string(data), "line1", "snapshot should contain line1")
	assert.Contains(t, string(data), "line2", "snapshot should contain line2")
}

// ---------------------------------------------------------------------------
// AC-4: Kill and Cleanup (BDD Scenario 4)
// ---------------------------------------------------------------------------

func TestManagedSession_AC4_KillTerminatesProcess(t *testing.T) {
	if testing.Short() {
		t.Skip("requires PTY (fork/exec)")
	}
	// BDD Scenario 4: Kill terminates process group
	sess := startTestSession(t, "sleep 300")

	// Session must report alive before Kill.
	require.True(t, sess.IsAlive(), "session should be alive before Kill")

	sess.Kill()

	assert.False(t, sess.IsAlive(), "session should be dead after Kill")
}

func TestManagedSession_AC4_DoubleKillNoPanic(t *testing.T) {
	if testing.Short() {
		t.Skip("requires PTY (fork/exec)")
	}
	// BDD Scenario 4: Calling Kill again does not panic
	sess := startTestSession(t, "sleep 300")

	assert.NotPanics(t, func() {
		sess.Kill()
		sess.Kill()
	}, "double Kill must not panic")
}

func TestManagedSession_AC4_NaturalExitDetected(t *testing.T) {
	if testing.Short() {
		t.Skip("requires PTY (fork/exec)")
	}
	// BDD Scenario 4: Session cleanup after process exits naturally
	sess := startTestSession(t, "echo done")

	time.Sleep(processExitWait)

	assert.False(t, sess.IsAlive(), "session should detect natural process exit via EOF")
}

// ---------------------------------------------------------------------------
// AC-5: Client Failure Isolation (BDD Scenario 2, failed-client variant)
// ---------------------------------------------------------------------------

func TestManagedSession_AC5_FailedClientIsolation(t *testing.T) {
	if testing.Short() {
		t.Skip("requires PTY (fork/exec)")
	}
	// BDD Scenario 2: Failed client is removed without killing session
	sess := startTestSession(t, "while true; do echo tick; sleep 0.1; done")

	srv1, cli1 := wsTestPair(t)
	srv2, cli2 := wsTestPair(t)

	sess.AddClient(srv1)
	sess.AddClient(srv2)

	time.Sleep(clientEvictWait)

	// Forcibly close client 1 to simulate failure.
	cli1.Close()
	srv1.Close()

	time.Sleep(clientEvictWait)

	assert.True(t, sess.IsAlive(), "session must stay alive after client failure")

	data, err := readWSTimeout(t, cli2, wsReadTimeout)
	require.NoError(t, err, "client 2 must still receive data after client 1 failure")
	assert.NotEmpty(t, data, "client 2 should receive non-empty output")
}
