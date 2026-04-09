package helper

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/creack/pty"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testSockPath returns a short socket path under /tmp to stay within macOS's
// 104-char limit for Unix domain socket addresses.
func testSockPath(t *testing.T) string {
	t.Helper()
	f, err := os.CreateTemp("/tmp", "pty-test-*.sock")
	require.NoError(t, err)
	path := f.Name()
	f.Close()
	os.Remove(path)
	t.Cleanup(func() { os.Remove(path) })
	return path
}

// canPTYSpawn returns true if pty.Start will work in this environment.
// Detects sandboxes that block fork/exec or PTY allocation.
func canPTYSpawn(t *testing.T) bool {
	t.Helper()
	cmd := exec.Command("/bin/echo", "probe")
	ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{Cols: 80, Rows: 24})
	if err != nil {
		return false
	}
	ptmx.Close()
	cmd.Wait()
	return true
}

// dialHelper connects to the server's Unix socket and returns a *net.UnixConn.
func dialHelper(t *testing.T, sockPath string) *net.UnixConn {
	t.Helper()
	conn, err := net.DialUnix("unix", nil, &net.UnixAddr{Name: sockPath, Net: "unix"})
	require.NoError(t, err, "dial helper socket")
	t.Cleanup(func() { conn.Close() })
	return conn
}

// startTestServer creates a Server, listens on a temp socket, and runs Serve in
// a goroutine. Returns the server, socket path.
func startTestServer(t *testing.T) (*Server, string) {
	t.Helper()

	sockPath := testSockPath(t)
	ln, err := net.ListenUnix("unix", &net.UnixAddr{Name: sockPath, Net: "unix"})
	require.NoError(t, err)

	srv := NewServer(os.Getpid())

	go srv.Serve(ln)

	t.Cleanup(func() {
		ln.Close()    // stop accepting; unblocks Serve and handleConn reads
		srv.Shutdown() // wait for in-flight handlers, then kill sessions
	})

	return srv, sockPath
}

// sendSpawn sends a SpawnRequest and reads back the SpawnResponse.
func sendSpawn(t *testing.T, conn *net.UnixConn, req SpawnRequest) SpawnResponse {
	t.Helper()

	err := WriteMessage(conn, MsgSpawn, req)
	require.NoError(t, err, "write spawn request")

	env, err := ReadMessage(conn)
	require.NoError(t, err, "read spawn response")
	assert.Equal(t, MsgSpawn, env.Type)

	var resp SpawnResponse
	require.NoError(t, json.Unmarshal(env.Data, &resp))
	return resp
}

// sendKill sends a KillRequest and reads back a response envelope.
func sendKill(t *testing.T, conn *net.UnixConn, req KillRequest) Envelope {
	t.Helper()

	err := WriteMessage(conn, MsgKill, req)
	require.NoError(t, err, "write kill request")

	env, err := ReadMessage(conn)
	require.NoError(t, err, "read kill response")
	return env
}

func TestServerSpawnSuccess(t *testing.T) {
	if !canPTYSpawn(t) {
		t.Skip("requires PTY spawn (sandbox)")
	}

	_, sockPath := startTestServer(t)
	conn := dialHelper(t, sockPath)

	resp := sendSpawn(t, conn, SpawnRequest{
		ID:    "s1",
		Shell: "/bin/echo",
		Args:  []string{"hello"},
		Cols:  80,
		Rows:  24,
	})

	require.Empty(t, resp.Error, "spawn should succeed")
	assert.Greater(t, resp.PID, 0, "PID should be positive")
	assert.Equal(t, "s1", resp.ID)

	// Receive the PTY fd via SCM_RIGHTS.
	fd, err := RecvFd(conn)
	require.NoError(t, err, "recv fd")
	assert.Greater(t, fd, 0, "fd should be positive")

	// Read output from the PTY fd.
	f := os.NewFile(uintptr(fd), "pty")
	defer f.Close()

	buf := make([]byte, 256)
	f.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, _ := f.Read(buf)
	output := string(buf[:n])
	assert.Contains(t, output, "hello", "should read echo output from pty fd")
}

func TestServerSpawnBadShell(t *testing.T) {
	_, sockPath := startTestServer(t)
	conn := dialHelper(t, sockPath)

	resp := sendSpawn(t, conn, SpawnRequest{
		ID:    "bad",
		Shell: "/nonexistent/shell",
		Cols:  80,
		Rows:  24,
	})

	assert.NotEmpty(t, resp.Error, "spawn with bad shell should return error")
	assert.Equal(t, "bad", resp.ID)
	assert.Equal(t, 0, resp.PID, "PID should be 0 on failure")
}

func TestServerKillSuccess(t *testing.T) {
	if !canPTYSpawn(t) {
		t.Skip("requires PTY spawn (sandbox)")
	}

	_, sockPath := startTestServer(t)
	conn := dialHelper(t, sockPath)

	resp := sendSpawn(t, conn, SpawnRequest{
		ID:    "killme",
		Shell: "/bin/cat",
		Cols:  80,
		Rows:  24,
	})
	require.Empty(t, resp.Error)
	require.Greater(t, resp.PID, 0)

	fd, err := RecvFd(conn)
	require.NoError(t, err)
	syscall.Close(fd)

	err = syscall.Kill(resp.PID, 0)
	require.NoError(t, err, "process should be alive before kill")

	killEnv := sendKill(t, conn, KillRequest{
		ID:     "killme",
		Signal: int(syscall.SIGTERM),
	})
	assert.Equal(t, MsgKill, killEnv.Type)

	time.Sleep(200 * time.Millisecond)

	err = syscall.Kill(resp.PID, 0)
	assert.Error(t, err, "process should be dead after kill")
}

func TestServerKillNotFound(t *testing.T) {
	_, sockPath := startTestServer(t)
	conn := dialHelper(t, sockPath)

	killEnv := sendKill(t, conn, KillRequest{
		ID:     "nonexistent",
		Signal: int(syscall.SIGTERM),
	})
	assert.Equal(t, MsgKill, killEnv.Type)

	// Parse the response to verify error message.
	var resp struct {
		ID    string `json:"id"`
		Error string `json:"error,omitempty"`
	}
	require.NoError(t, json.Unmarshal(killEnv.Data, &resp))
	assert.Equal(t, "nonexistent", resp.ID)
	assert.Contains(t, resp.Error, "not found")
}

func TestServerMultipleConcurrentSessions(t *testing.T) {
	if !canPTYSpawn(t) {
		t.Skip("requires PTY spawn (sandbox)")
	}

	_, sockPath := startTestServer(t)
	conn := dialHelper(t, sockPath)

	pids := make(map[int]bool)
	for i := 0; i < 3; i++ {
		id := fmt.Sprintf("session-%d", i)
		resp := sendSpawn(t, conn, SpawnRequest{
			ID:    id,
			Shell: "/bin/cat",
			Cols:  80,
			Rows:  24,
		})
		require.Empty(t, resp.Error, "spawn %s should succeed", id)
		require.Greater(t, resp.PID, 0)
		assert.False(t, pids[resp.PID], "PID %d should be unique", resp.PID)
		pids[resp.PID] = true

		fd, err := RecvFd(conn)
		require.NoError(t, err)
		syscall.Close(fd)
	}

	assert.Len(t, pids, 3, "should have 3 unique PIDs")
}

func TestServerShutdownCleansSessions(t *testing.T) {
	if !canPTYSpawn(t) {
		t.Skip("requires PTY spawn (sandbox)")
	}

	srv, sockPath := startTestServer(t)
	conn := dialHelper(t, sockPath)

	var sessionPIDs []int
	for i := 0; i < 2; i++ {
		id := fmt.Sprintf("shutdown-%d", i)
		resp := sendSpawn(t, conn, SpawnRequest{
			ID:    id,
			Shell: "/bin/cat",
			Cols:  80,
			Rows:  24,
		})
		require.Empty(t, resp.Error)
		require.Greater(t, resp.PID, 0)
		sessionPIDs = append(sessionPIDs, resp.PID)

		fd, err := RecvFd(conn)
		require.NoError(t, err)
		syscall.Close(fd)
	}

	for _, pid := range sessionPIDs {
		err := syscall.Kill(pid, 0)
		require.NoError(t, err, "pid %d should be alive before shutdown", pid)
	}

	srv.Shutdown()

	time.Sleep(200 * time.Millisecond)

	for _, pid := range sessionPIDs {
		err := syscall.Kill(pid, 0)
		assert.Error(t, err, "pid %d should be dead after shutdown", pid)
	}
}

func TestServerMonitorParentExitsOnDeadParent(t *testing.T) {
	srv := NewServer(999999999)

	exited := make(chan struct{})
	origExit := osExit
	osExit = func(code int) {
		close(exited)
	}
	t.Cleanup(func() { osExit = origExit })

	go srv.MonitorParent(999999999)

	select {
	case <-exited:
		// Success — monitor detected dead parent.
	case <-time.After(5 * time.Second):
		t.Fatal("MonitorParent did not exit within 5s for dead parent PID")
	}
}

func TestServerSpawnDefaultShell(t *testing.T) {
	if !canPTYSpawn(t) {
		t.Skip("requires PTY spawn (sandbox)")
	}

	_, sockPath := startTestServer(t)
	conn := dialHelper(t, sockPath)

	resp := sendSpawn(t, conn, SpawnRequest{
		ID:   "default-shell",
		Cols: 80,
		Rows: 24,
	})

	require.Empty(t, resp.Error, "spawn with default shell should succeed")
	assert.Greater(t, resp.PID, 0)

	fd, err := RecvFd(conn)
	require.NoError(t, err)
	syscall.Close(fd)

	sendKill(t, conn, KillRequest{
		ID:     "default-shell",
		Signal: int(syscall.SIGTERM),
	})
}

func TestServerShutdownEmpty(t *testing.T) {
	// Shutdown on an empty server should not panic.
	srv := NewServer(os.Getpid())
	srv.Shutdown()
}

func TestServerNewServer(t *testing.T) {
	srv := NewServer(42)
	assert.NotNil(t, srv)
	assert.Equal(t, 42, srv.parentPID)
}

// TestServerShutdownWithSessions tests Shutdown with pre-populated sessions
// (no fork/exec required — uses a sleep command or /dev/null as ptmx stand-in).
func TestServerShutdownWithMockSessions(t *testing.T) {
	srv := NewServer(os.Getpid())

	// Create mock sessions with opened /dev/null as ptmx stand-in.
	for i := 0; i < 2; i++ {
		f, err := os.Open("/dev/null")
		require.NoError(t, err)
		id := fmt.Sprintf("mock-%d", i)
		cmd := exec.Command("/bin/echo") // won't be started
		cmd.Process = &os.Process{Pid: 99999 + i}
		srv.sessions.Store(id, &helperSession{
			cmd:  cmd,
			ptmx: f,
		})
	}

	// Verify sessions exist.
	count := 0
	srv.sessions.Range(func(_, _ any) bool { count++; return true })
	assert.Equal(t, 2, count, "should have 2 sessions before shutdown")

	srv.Shutdown()

	// Verify all sessions removed.
	count = 0
	srv.sessions.Range(func(_, _ any) bool { count++; return true })
	assert.Equal(t, 0, count, "should have 0 sessions after shutdown")
}

// TestServerKillWithMockSession exercises the kill path via IPC with a
// pre-populated session (no fork/exec required).
func TestServerKillWithMockSession(t *testing.T) {
	srv, sockPath := startTestServer(t)
	conn := dialHelper(t, sockPath)

	// Pre-populate a session with a mock process and /dev/null as ptmx.
	f, err := os.Open("/dev/null")
	require.NoError(t, err)
	cmd := exec.Command("/bin/echo")
	cmd.Process = &os.Process{Pid: 99999}
	srv.sessions.Store("mock-kill", &helperSession{
		cmd:  cmd,
		ptmx: f,
	})

	// Send kill request via IPC.
	killEnv := sendKill(t, conn, KillRequest{
		ID:     "mock-kill",
		Signal: int(syscall.SIGTERM),
	})
	assert.Equal(t, MsgKill, killEnv.Type)

	// Session should be removed from the map.
	_, loaded := srv.sessions.Load("mock-kill")
	assert.False(t, loaded, "session should be removed after kill")
}

// TestServerHandleConnUnknownType tests that unknown message types are handled gracefully.
func TestServerHandleConnUnknownType(t *testing.T) {
	_, sockPath := startTestServer(t)
	conn := dialHelper(t, sockPath)

	// Send an unknown message type — server should log it and continue.
	err := WriteMessage(conn, MessageType("unknown"), struct{}{})
	require.NoError(t, err)

	// Server should still be alive — send a valid message.
	resp := sendSpawn(t, conn, SpawnRequest{
		ID:    "after-unknown",
		Shell: "/nonexistent/shell",
		Cols:  80,
		Rows:  24,
	})
	assert.NotEmpty(t, resp.Error, "should get error response for bad shell")
	assert.Equal(t, "after-unknown", resp.ID)
}

// TestServerHandleConnBadPayload sends messages with invalid JSON data fields to
// exercise the unmarshal error branches in handleConn.
func TestServerHandleConnBadPayload(t *testing.T) {
	_, sockPath := startTestServer(t)
	conn := dialHelper(t, sockPath)

	// Construct raw wire bytes with valid envelope but invalid data field.
	// The envelope JSON is valid, but the "data" field contains invalid JSON for
	// SpawnRequest/KillRequest unmarshaling.
	sendRaw := func(msgType string, badData string) {
		t.Helper()
		payload := []byte(fmt.Sprintf(`{"type":"%s","data":%s}`, msgType, badData))
		var lenBuf [4]byte
		binary.BigEndian.PutUint32(lenBuf[:], uint32(len(payload)))
		_, err := conn.Write(lenBuf[:])
		require.NoError(t, err)
		_, err = conn.Write(payload)
		require.NoError(t, err)
	}

	// Send spawn with data that's valid JSON but wrong shape (will fail unmarshal to SpawnRequest).
	sendRaw("spawn", `"not-an-object"`)
	// Send kill with bad data.
	sendRaw("kill", `"not-an-object"`)

	// Brief pause for server to process the bad messages.
	time.Sleep(100 * time.Millisecond)

	// Server should still be alive — send a valid message after bad ones.
	resp := sendSpawn(t, conn, SpawnRequest{
		ID:    "after-bad",
		Shell: "/nonexistent/shell",
		Cols:  80,
		Rows:  24,
	})
	assert.NotEmpty(t, resp.Error)
	assert.Equal(t, "after-bad", resp.ID)
}

// TestServerServeAcceptsMultipleConnections tests that Serve handles multiple connections.
func TestServerServeAcceptsMultipleConnections(t *testing.T) {
	_, sockPath := startTestServer(t)

	// Connect twice, send spawn with bad shell on each.
	for i := 0; i < 2; i++ {
		conn := dialHelper(t, sockPath)
		resp := sendSpawn(t, conn, SpawnRequest{
			ID:    fmt.Sprintf("conn-%d", i),
			Shell: "/nonexistent/shell",
			Cols:  80,
			Rows:  24,
		})
		assert.NotEmpty(t, resp.Error)
		assert.Equal(t, fmt.Sprintf("conn-%d", i), resp.ID)
	}
}

// TestServerMonitorParentAliveDoesNotExit tests that MonitorParent keeps running
// when the parent is alive.
func TestServerMonitorParentAliveDoesNotExit(t *testing.T) {
	srv := NewServer(os.Getpid())

	exited := make(chan struct{})
	origExit := osExit
	osExit = func(code int) {
		close(exited)
	}
	t.Cleanup(func() { osExit = origExit })

	go srv.MonitorParent(os.Getpid())

	select {
	case <-exited:
		t.Fatal("MonitorParent should NOT exit when parent is alive")
	case <-time.After(3 * time.Second):
		// Success — monitor stayed alive.
	}
}

func TestServerConcurrentSpawnRace(t *testing.T) {
	if !canPTYSpawn(t) {
		t.Skip("requires PTY spawn (sandbox)")
	}

	_, sockPath := startTestServer(t)

	var wg sync.WaitGroup
	errs := make(chan error, 5)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			conn := dialHelper(t, sockPath)
			resp := sendSpawn(t, conn, SpawnRequest{
				ID:    fmt.Sprintf("race-%d", idx),
				Shell: "/bin/cat",
				Cols:  80,
				Rows:  24,
			})
			if resp.Error != "" {
				errs <- fmt.Errorf("spawn race-%d: %s", idx, resp.Error)
				return
			}
			fd, err := RecvFd(conn)
			if err != nil {
				errs <- fmt.Errorf("recvfd race-%d: %v", idx, err)
				return
			}
			syscall.Close(fd)
		}(i)
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Error(err)
	}
}
