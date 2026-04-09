//go:build darwin

package helper

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// countOpenFds returns the number of open file descriptors for the current process.
func countOpenFds(t *testing.T) int {
	t.Helper()
	entries, err := os.ReadDir("/dev/fd")
	if err != nil {
		t.Fatalf("cannot read /dev/fd: %v", err)
	}
	return len(entries)
}

// TestIntegration_AC1_SpawnAndReadFd proves the full helper server -> client -> fd
// passing pipeline works end-to-end using SCM_RIGHTS.
func TestIntegration_AC1_SpawnAndReadFd(t *testing.T) {
	if !canPTYSpawn(t) {
		t.Skip("requires PTY spawn (sandbox)")
	}

	_, sockPath := startTestServer(t)

	client, err := Dial(sockPath)
	require.NoError(t, err, "client dial should succeed")
	defer client.Close()

	ptmx, pid, err := client.Spawn(context.Background(), SpawnRequest{
		ID:    "ac1-echo",
		Shell: "/bin/echo",
		Args:  []string{"hello"},
		Cols:  80,
		Rows:  24,
	})
	require.NoError(t, err, "spawn should succeed")
	require.NotNil(t, ptmx, "returned file must not be nil")
	defer ptmx.Close()

	assert.Greater(t, pid, 0, "PID should be positive")

	// Verify the fd is valid by checking Stat.
	fi, err := ptmx.Stat()
	require.NoError(t, err, "os.NewFile on received fd should produce a valid file")
	assert.NotNil(t, fi, "file info should not be nil")

	// Read output from the PTY — echo should produce "hello".
	buf := make([]byte, 512)
	ptmx.SetReadDeadline(time.Now().Add(3 * time.Second))
	n, readErr := ptmx.Read(buf)
	// PTY reads may return io.EOF after the process exits; that's fine as long as we got data.
	if readErr != nil && readErr != io.EOF {
		require.NoError(t, readErr, "read from PTY fd")
	}
	output := string(buf[:n])
	assert.Contains(t, output, "hello", "PTY output should contain echo's argument")
}

// TestIntegration_AC2_KillTerminatesProcess proves that sending a KillRequest
// terminates a spawned process and removes the session from the server.
func TestIntegration_AC2_KillTerminatesProcess(t *testing.T) {
	if !canPTYSpawn(t) {
		t.Skip("requires PTY spawn (sandbox)")
	}

	srv, sockPath := startTestServer(t)

	client, err := Dial(sockPath)
	require.NoError(t, err)
	defer client.Close()

	// Spawn a long-running process (/bin/cat blocks forever waiting for input).
	ptmx, pid, err := client.Spawn(context.Background(), SpawnRequest{
		ID:    "ac2-cat",
		Shell: "/bin/cat",
		Cols:  80,
		Rows:  24,
	})
	require.NoError(t, err, "spawn cat should succeed")
	require.NotNil(t, ptmx)
	defer ptmx.Close()
	require.Greater(t, pid, 0)

	// Verify process is alive.
	err = syscall.Kill(pid, 0)
	require.NoError(t, err, "process should be alive before kill")

	// Verify session exists in the server's map.
	_, loaded := srv.sessions.Load("ac2-cat")
	require.True(t, loaded, "session should exist in server map before kill")

	// Send KillRequest via Client.
	err = client.Kill("ac2-cat", syscall.SIGTERM)
	require.NoError(t, err, "kill should not return error")

	// Wait for the process to actually exit.
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if err := syscall.Kill(pid, 0); err != nil {
			break // process is dead
		}
		time.Sleep(50 * time.Millisecond)
	}

	err = syscall.Kill(pid, 0)
	assert.Error(t, err, "process should be dead after kill")

	// Wait briefly for the server to process the kill and remove the session.
	time.Sleep(200 * time.Millisecond)

	// Verify session is removed from the server's map.
	_, loaded = srv.sessions.Load("ac2-cat")
	assert.False(t, loaded, "session should be removed from server map after kill")
}

// TestIntegration_BidirectionalIO proves that bytes flow both directions through
// the received PTY fd — the story's AC-3 requirement.
func TestIntegration_BidirectionalIO(t *testing.T) {
	if !canPTYSpawn(t) {
		t.Skip("requires PTY spawn (sandbox)")
	}

	_, sockPath := startTestServer(t)

	client, err := Dial(sockPath)
	require.NoError(t, err)
	defer client.Close()

	// Spawn /bin/cat which echoes stdin to stdout.
	ptmx, pid, err := client.Spawn(context.Background(), SpawnRequest{
		ID:    "bidir-cat",
		Shell: "/bin/cat",
		Cols:  80,
		Rows:  24,
	})
	require.NoError(t, err)
	require.NotNil(t, ptmx)
	defer ptmx.Close()
	require.Greater(t, pid, 0)

	// Write to the PTY fd — this goes to cat's stdin.
	_, err = ptmx.Write([]byte("HELLO_FROM_MAIN\n"))
	require.NoError(t, err)

	// Read from the PTY fd — cat echoes back via stdout.
	buf := make([]byte, 4096)
	ptmx.SetReadDeadline(time.Now().Add(5 * time.Second))
	var collected []byte
	for {
		n, readErr := ptmx.Read(buf)
		if n > 0 {
			collected = append(collected, buf[:n]...)
		}
		if len(collected) > 0 && strings.Contains(string(collected), "HELLO_FROM_MAIN") {
			break
		}
		if readErr != nil {
			t.Fatalf("read failed before seeing echo: %v (got %q)", readErr, collected)
		}
	}
	assert.Contains(t, string(collected), "HELLO_FROM_MAIN",
		"bidirectional I/O: written data should echo back through PTY fd")

	// Kill the session to clean up.
	_ = client.Kill("bidir-cat", syscall.SIGTERM)
}

// TestIntegration_ConcurrentSpawnsFdIsolation proves that multiple concurrent
// SpawnRequests each receive a distinct, valid fd mapped to the correct session.
func TestIntegration_ConcurrentSpawnsFdIsolation(t *testing.T) {
	if !canPTYSpawn(t) {
		t.Skip("requires PTY spawn (sandbox)")
	}

	_, sockPath := startTestServer(t)

	const spawnCount = 4

	type spawnResult struct {
		id    string
		ptmx  *os.File
		pid   int
		fdNum uintptr
		err   error
	}

	results := make([]spawnResult, spawnCount)
	var wg sync.WaitGroup

	// Each goroutine needs its own client because Spawn holds a mutex.
	for i := 0; i < spawnCount; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			client, err := Dial(sockPath)
			if err != nil {
				results[idx] = spawnResult{err: fmt.Errorf("dial: %w", err)}
				return
			}
			defer client.Close()

			id := fmt.Sprintf("ac3-cat-%d", idx)
			ptmx, pid, err := client.Spawn(context.Background(), SpawnRequest{
				ID:    id,
				Shell: "/bin/cat",
				Cols:  80,
				Rows:  24,
			})
			res := spawnResult{
				id:   id,
				ptmx: ptmx,
				pid:  pid,
				err:  err,
			}
			if ptmx != nil {
				res.fdNum = ptmx.Fd()
			}
			results[idx] = res
		}(i)
	}
	wg.Wait()

	// Verify all spawns succeeded and collect fds/pids.
	fds := make(map[uintptr]bool)
	pids := make(map[int]bool)
	for i, r := range results {
		require.NoError(t, r.err, "spawn %d should succeed", i)
		require.NotNil(t, r.ptmx, "ptmx %d should not be nil", i)
		defer results[i].ptmx.Close()
		assert.Greater(t, r.pid, 0, "PID %d should be positive", i)

		// Fds should be distinct.
		assert.False(t, fds[r.fdNum], "fd %d should be unique (got %d)", i, r.fdNum)
		fds[r.fdNum] = true

		// PIDs should be distinct.
		assert.False(t, pids[r.pid], "PID %d should be unique (got %d)", i, r.pid)
		pids[r.pid] = true
	}

	assert.Len(t, fds, spawnCount, "should have %d unique fds", spawnCount)
	assert.Len(t, pids, spawnCount, "should have %d unique PIDs", spawnCount)

	// Verify each fd maps to the correct session by writing distinct data
	// and reading it back. /bin/cat echoes stdin to stdout via PTY.
	for i, r := range results {
		marker := fmt.Sprintf("marker-%d\n", i)
		_, err := r.ptmx.Write([]byte(marker))
		require.NoError(t, err, "write to ptmx %d should succeed", i)

		buf := make([]byte, 256)
		r.ptmx.SetReadDeadline(time.Now().Add(2 * time.Second))
		n, readErr := r.ptmx.Read(buf)
		if readErr != nil && readErr != io.EOF {
			require.NoError(t, readErr, "read from ptmx %d", i)
		}
		output := string(buf[:n])
		assert.Contains(t, output, fmt.Sprintf("marker-%d", i),
			"ptmx %d should echo its own distinct marker", i)
	}
}

// TestIntegration_AC4_LatencyAndFdLeaks measures round-trip spawn+kill latency
// and verifies no file descriptor leaks occur.
func TestIntegration_AC4_LatencyAndFdLeaks(t *testing.T) {
	if !canPTYSpawn(t) {
		t.Skip("requires PTY spawn (sandbox)")
	}

	srv, sockPath := startTestServer(t)

	client, err := Dial(sockPath)
	require.NoError(t, err)
	defer client.Close()

	const cycles = 10

	// Warm up — let fd counting settle.
	warmPtmx, _, err := client.Spawn(context.Background(), SpawnRequest{
		ID:    "warmup",
		Shell: "/bin/cat",
		Cols:  80,
		Rows:  24,
	})
	require.NoError(t, err)
	warmPtmx.Close()
	err = client.Kill("warmup", syscall.SIGTERM)
	require.NoError(t, err)
	time.Sleep(100 * time.Millisecond)

	fdsBefore := countOpenFds(t)

	start := time.Now()
	for i := 0; i < cycles; i++ {
		id := fmt.Sprintf("latency-%d", i)
		ptmx, pid, err := client.Spawn(context.Background(), SpawnRequest{
			ID:    id,
			Shell: "/bin/cat",
			Cols:  80,
			Rows:  24,
		})
		require.NoError(t, err, "spawn cycle %d should succeed", i)
		require.Greater(t, pid, 0, "PID should be positive in cycle %d", i)

		// Close our copy of the PTY fd immediately.
		ptmx.Close()

		// Kill the session.
		err = client.Kill(id, syscall.SIGTERM)
		require.NoError(t, err, "kill cycle %d should succeed", i)

		// Wait for process to die and server to clean up.
		deadline := time.Now().Add(2 * time.Second)
		for time.Now().Before(deadline) {
			if err := syscall.Kill(pid, 0); err != nil {
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
	totalDuration := time.Since(start)

	avgLatency := totalDuration / cycles
	t.Logf("Total: %v, Average: %v per cycle (%d cycles)", totalDuration, avgLatency, cycles)
	assert.Less(t, avgLatency, 200*time.Millisecond,
		"average latency should be < 200ms per cycle (got %v)", avgLatency)

	// Verify no session leaks in the server.
	sessionCount := 0
	srv.sessions.Range(func(_, _ any) bool { sessionCount++; return true })
	assert.Equal(t, 0, sessionCount, "all sessions should be cleaned up")

	// Give OS a moment to reclaim fds.
	time.Sleep(200 * time.Millisecond)

	fdsAfter := countOpenFds(t)
	// Allow a small delta (1-2 fds) for runtime/GC fluctuation.
	fdDelta := fdsAfter - fdsBefore
	t.Logf("FDs before: %d, after: %d, delta: %d", fdsBefore, fdsAfter, fdDelta)
	assert.LessOrEqual(t, fdDelta, 2,
		"fd leak detected: before=%d after=%d delta=%d", fdsBefore, fdsAfter, fdDelta)
}
