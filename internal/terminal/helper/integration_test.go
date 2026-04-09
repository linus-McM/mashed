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

const testShellCat = "/bin/cat"

func catSpawnReq(id string) SpawnRequest {
	return SpawnRequest{ID: id, Shell: testShellCat, Cols: 80, Rows: 24}
}

func countOpenFds(t *testing.T) int {
	t.Helper()
	// /dev/fd enumeration can transiently fail on macOS during fd churn.
	for range 3 {
		entries, err := os.ReadDir("/dev/fd")
		if err == nil {
			return len(entries)
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Skip("cannot read /dev/fd reliably — skipping fd leak test")
	return 0
}

func waitForPIDDeath(t *testing.T, pid int, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if err := syscall.Kill(pid, 0); err != nil {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Errorf("process %d still alive after %v", pid, timeout)
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

	fi, err := ptmx.Stat()
	require.NoError(t, err, "received fd should be valid")
	assert.NotNil(t, fi)

	buf := make([]byte, 512)
	ptmx.SetReadDeadline(time.Now().Add(3 * time.Second))
	n, readErr := ptmx.Read(buf)
	if readErr != nil && readErr != io.EOF {
		require.NoError(t, readErr, "read from PTY fd")
	}
	assert.Contains(t, string(buf[:n]), "hello", "PTY output should contain echo's argument")
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

	ptmx, pid, err := client.Spawn(context.Background(), catSpawnReq("ac2-cat"))
	require.NoError(t, err, "spawn cat should succeed")
	require.NotNil(t, ptmx)
	defer ptmx.Close()
	require.Greater(t, pid, 0)

	err = syscall.Kill(pid, 0)
	require.NoError(t, err, "process should be alive before kill")

	_, loaded := srv.sessions.Load("ac2-cat")
	require.True(t, loaded, "session should exist in server map before kill")

	err = client.Kill("ac2-cat", syscall.SIGTERM)
	require.NoError(t, err, "kill should not return error")

	waitForPIDDeath(t, pid, 3*time.Second)

	err = syscall.Kill(pid, 0)
	assert.Error(t, err, "process should be dead after kill")

	// Brief pause for server-side session cleanup.
	time.Sleep(200 * time.Millisecond)

	_, loaded = srv.sessions.Load("ac2-cat")
	assert.False(t, loaded, "session should be removed from server map after kill")
}

// TestIntegration_BidirectionalIO proves bidirectional I/O through the received
// PTY fd — write to cat's stdin, read echo back via stdout.
func TestIntegration_BidirectionalIO(t *testing.T) {
	if !canPTYSpawn(t) {
		t.Skip("requires PTY spawn (sandbox)")
	}

	_, sockPath := startTestServer(t)

	client, err := Dial(sockPath)
	require.NoError(t, err)
	defer client.Close()

	ptmx, pid, err := client.Spawn(context.Background(), catSpawnReq("bidir-cat"))
	require.NoError(t, err)
	require.NotNil(t, ptmx)
	defer ptmx.Close()
	require.Greater(t, pid, 0)

	_, err = ptmx.Write([]byte("HELLO_FROM_MAIN\n"))
	require.NoError(t, err)

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
			ptmx, pid, err := client.Spawn(context.Background(), catSpawnReq(id))
			res := spawnResult{id: id, ptmx: ptmx, pid: pid, err: err}
			if ptmx != nil {
				res.fdNum = ptmx.Fd()
			}
			results[idx] = res
		}(i)
	}
	wg.Wait()

	fds := make(map[uintptr]bool)
	pids := make(map[int]bool)
	for i, r := range results {
		require.NoError(t, r.err, "spawn %d should succeed", i)
		require.NotNil(t, r.ptmx, "ptmx %d should not be nil", i)
		defer results[i].ptmx.Close()
		assert.Greater(t, r.pid, 0, "PID %d should be positive", i)

		assert.False(t, fds[r.fdNum], "fd %d should be unique (got %d)", i, r.fdNum)
		fds[r.fdNum] = true

		assert.False(t, pids[r.pid], "PID %d should be unique (got %d)", i, r.pid)
		pids[r.pid] = true
	}

	assert.Len(t, fds, spawnCount, "should have %d unique fds", spawnCount)
	assert.Len(t, pids, spawnCount, "should have %d unique PIDs", spawnCount)

	// Verify each fd maps to the correct session by writing distinct markers.
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
		assert.Contains(t, string(buf[:n]), fmt.Sprintf("marker-%d", i),
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
	warmPtmx, warmPID, err := client.Spawn(context.Background(), catSpawnReq("warmup"))
	require.NoError(t, err)
	warmPtmx.Close()
	err = client.Kill("warmup", syscall.SIGTERM)
	require.NoError(t, err)
	waitForPIDDeath(t, warmPID, 2*time.Second)

	fdsBefore := countOpenFds(t)

	start := time.Now()
	for i := 0; i < cycles; i++ {
		id := fmt.Sprintf("latency-%d", i)
		ptmx, pid, err := client.Spawn(context.Background(), catSpawnReq(id))
		require.NoError(t, err, "spawn cycle %d should succeed", i)
		require.Greater(t, pid, 0, "PID should be positive in cycle %d", i)

		ptmx.Close()

		err = client.Kill(id, syscall.SIGTERM)
		require.NoError(t, err, "kill cycle %d should succeed", i)

		waitForPIDDeath(t, pid, 2*time.Second)
	}
	totalDuration := time.Since(start)

	avgLatency := totalDuration / cycles
	t.Logf("Total: %v, Average: %v per cycle (%d cycles)", totalDuration, avgLatency, cycles)
	assert.Less(t, avgLatency, 200*time.Millisecond,
		"average latency should be < 200ms per cycle (got %v)", avgLatency)

	sessionCount := 0
	srv.sessions.Range(func(_, _ any) bool { sessionCount++; return true })
	assert.Equal(t, 0, sessionCount, "all sessions should be cleaned up")

	fdsAfter := countOpenFds(t)
	fdDelta := fdsAfter - fdsBefore
	t.Logf("FDs before: %d, after: %d, delta: %d", fdsBefore, fdsAfter, fdDelta)
	assert.LessOrEqual(t, fdDelta, 2,
		"fd leak detected: before=%d after=%d delta=%d", fdsBefore, fdsAfter, fdDelta)
}
