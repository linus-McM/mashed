// Package terminal tests for SessionManager lifecycle manager.
//
// Story pty-02 tests verified direct fork/exec spawning. With the Story 5
// rework, Spawn delegates to the helper client. Tests that need a live session
// use injectRemoteSession (pipe-backed) to avoid requiring a running helper.

package terminal

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// cmdOutputSettleTime is how long to wait for a short-lived command to
// produce PTY output before asserting on the scrollback buffer.
const cmdOutputSettleTime = 500 * time.Millisecond

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// injectRemoteSession creates a pipe-backed ManagedSession and registers it
// directly in the manager's session map. This avoids needing a helper binary.
// Returns the session and the write end of the pipe (caller can write PTY data).
func injectRemoteSession(t *testing.T, sm *SessionManager, name string, pid int) (*ManagedSession, *os.File) {
	t.Helper()
	r, w, err := os.Pipe()
	require.NoError(t, err)

	ms := newRemoteManagedSession(name, pid, r)

	sm.mu.Lock()
	sm.sessions[name] = ms
	sm.mu.Unlock()

	t.Cleanup(func() {
		w.Close()
		// closing w triggers readLoop EOF, which closes r and signals done
	})

	return ms, w
}

// ---------------------------------------------------------------------------
// Story 5 AC-3: Graceful degradation — nil helper client
// ---------------------------------------------------------------------------

func TestSessionManager_NilClient_SpawnReturnsErrHelperNotRunning(t *testing.T) {
	sm := NewSessionManager(nil)
	defer sm.Shutdown()

	ms, err := sm.Spawn(context.Background(), "test", t.TempDir(), "", 0, 0)
	assert.Nil(t, ms)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrHelperNotRunning),
		"expected ErrHelperNotRunning, got: %v", err)
}

func TestSessionManager_NilClient_SpawnWithCommand(t *testing.T) {
	sm := NewSessionManager(nil)
	defer sm.Shutdown()

	ms, err := sm.Spawn(context.Background(), "test", t.TempDir(), "echo hello", 0, 0)
	assert.Nil(t, ms)
	assert.ErrorIs(t, err, ErrHelperNotRunning)
}

func TestSessionManager_NilClient_ListReturnsEmpty(t *testing.T) {
	sm := NewSessionManager(nil)
	list := sm.List(nil)
	assert.Empty(t, list)
}

func TestSessionManager_NilClient_IsAliveReturnsFalse(t *testing.T) {
	sm := NewSessionManager(nil)
	assert.False(t, sm.IsAlive("anything"))
}

// ---------------------------------------------------------------------------
// Story 5 AC-2: Remote ManagedSession works with pipe-backed fd
// ---------------------------------------------------------------------------

func TestRemoteSession_IsAlive(t *testing.T) {
	sm := NewSessionManager(nil)
	defer sm.Shutdown()

	ms, _ := injectRemoteSession(t, sm, "alive-test", 12345)
	assert.True(t, ms.IsAlive(), "remote session should be alive")
	assert.Equal(t, "alive-test", ms.Name())
	assert.Equal(t, 12345, ms.pid)
}

func TestRemoteSession_NilCmd(t *testing.T) {
	sm := NewSessionManager(nil)
	defer sm.Shutdown()

	ms, _ := injectRemoteSession(t, sm, "nil-cmd-test", 99)
	assert.Nil(t, ms.cmd, "remote session should have nil cmd")
}

func TestRemoteSession_ReadLoop(t *testing.T) {
	sm := NewSessionManager(nil)
	defer sm.Shutdown()

	ms, w := injectRemoteSession(t, sm, "readloop-test", 100)

	_, err := w.Write([]byte("hello remote"))
	require.NoError(t, err)

	time.Sleep(cmdOutputSettleTime)

	snap := ms.scroll.Snapshot()
	assert.Contains(t, string(snap), "hello remote",
		"scrollback should contain data written to pipe")
}

// ---------------------------------------------------------------------------
// Get / Kill / IsAlive lifecycle (using injected remote sessions)
// ---------------------------------------------------------------------------

func TestSessionManager_GetKillLifecycle(t *testing.T) {
	sm := NewSessionManager(nil)
	defer sm.Shutdown()

	ms, _ := injectRemoteSession(t, sm, "sess1", 200)

	got, ok := sm.Get("sess1")
	require.True(t, ok)
	assert.Equal(t, ms, got)
	assert.True(t, sm.IsAlive("sess1"))

	err := sm.Kill("sess1")
	require.NoError(t, err)

	got, ok = sm.Get("sess1")
	assert.False(t, ok)
	assert.Nil(t, got)
	assert.False(t, sm.IsAlive("sess1"))
}

func TestSessionManager_KillNonExistent(t *testing.T) {
	sm := NewSessionManager(nil)
	err := sm.Kill("ghost")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrSessionNotFound)
}

func TestSessionManager_GetNonExistent(t *testing.T) {
	sm := NewSessionManager(nil)
	got, ok := sm.Get("nope")
	assert.False(t, ok)
	assert.Nil(t, got)
}

func TestSessionManager_IsAliveNonExistent(t *testing.T) {
	sm := NewSessionManager(nil)
	assert.False(t, sm.IsAlive("nope"))
}

// ---------------------------------------------------------------------------
// Duplicate session names (using injected sessions)
// ---------------------------------------------------------------------------

func TestSessionManager_DuplicateName(t *testing.T) {
	sm := NewSessionManager(nil)
	defer sm.Shutdown()

	ms, _ := injectRemoteSession(t, sm, "dup", 300)

	// With nil client, Spawn returns ErrHelperNotRunning before reaching
	// duplicate check. Verify the session is still unaffected.
	dup, err := sm.Spawn(context.Background(), "dup", t.TempDir(), "", 0, 0)
	assert.Nil(t, dup)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrHelperNotRunning)

	// Original should be unaffected
	assert.True(t, ms.IsAlive())
	assert.True(t, sm.IsAlive("dup"))
}

// ---------------------------------------------------------------------------
// FindByPID
// ---------------------------------------------------------------------------

func TestSessionManager_FindByPID(t *testing.T) {
	sm := NewSessionManager(nil)
	defer sm.Shutdown()

	ms, _ := injectRemoteSession(t, sm, "pid-lookup", 42)

	found, ok := sm.FindByPID(42)
	require.True(t, ok)
	assert.Equal(t, ms, found)
	assert.Equal(t, "pid-lookup", found.Name())
}

func TestSessionManager_FindByPID_Unknown(t *testing.T) {
	sm := NewSessionManager(nil)
	defer sm.Shutdown()

	injectRemoteSession(t, sm, "some-session", 500)

	found, ok := sm.FindByPID(99999)
	assert.False(t, ok)
	assert.Nil(t, found)
}

func TestSessionManager_FindByPID_EmptyManager(t *testing.T) {
	sm := NewSessionManager(nil)
	found, ok := sm.FindByPID(1)
	assert.False(t, ok)
	assert.Nil(t, found)
}

// ---------------------------------------------------------------------------
// Shutdown
// ---------------------------------------------------------------------------

func TestSessionManager_Shutdown(t *testing.T) {
	sm := NewSessionManager(nil)

	names := []string{"a", "b", "c"}
	sessions := make([]*ManagedSession, len(names))
	for i, name := range names {
		sessions[i], _ = injectRemoteSession(t, sm, name, 1000+i)
	}

	for _, name := range names {
		require.True(t, sm.IsAlive(name))
	}

	sm.Shutdown()

	for i, name := range names {
		assert.False(t, sessions[i].IsAlive(), "%s should be dead after shutdown", name)
		assert.False(t, sm.IsAlive(name))
	}

	list := sm.List(nil)
	assert.Empty(t, list)
}

func TestSessionManager_ShutdownIdempotent(t *testing.T) {
	sm := NewSessionManager(nil)
	injectRemoteSession(t, sm, "x", 2000)

	sm.Shutdown()
	assert.NotPanics(t, func() { sm.Shutdown() })
}

// ---------------------------------------------------------------------------
// List with filter
// ---------------------------------------------------------------------------

func TestSessionManager_List_All(t *testing.T) {
	sm := NewSessionManager(nil)
	defer sm.Shutdown()

	injectRemoteSession(t, sm, "l1", 3001)
	injectRemoteSession(t, sm, "l2", 3002)
	injectRemoteSession(t, sm, "l3", 3003)

	all := sm.List(nil)
	assert.Len(t, all, 3)
}

func TestSessionManager_List_WithFilter(t *testing.T) {
	sm := NewSessionManager(nil)
	defer sm.Shutdown()

	injectRemoteSession(t, sm, "agent-1", 4001)
	injectRemoteSession(t, sm, "agent-2", 4002)
	injectRemoteSession(t, sm, "shell-1", 4003)

	agents := sm.List(func(ms *ManagedSession) bool {
		return strings.HasPrefix(ms.Name(), "agent-")
	})
	assert.Len(t, agents, 2)
}

func TestSessionManager_List_Empty(t *testing.T) {
	sm := NewSessionManager(nil)
	list := sm.List(nil)
	assert.Empty(t, list)
}

// ---------------------------------------------------------------------------
// Concurrency safety
// ---------------------------------------------------------------------------

func TestSessionManager_ConcurrentInjectAndKill(t *testing.T) {
	t.Parallel()

	sm := NewSessionManager(nil)
	defer sm.Shutdown()

	const n = 5
	var wg sync.WaitGroup
	wg.Add(2)

	// Inject sessions concurrently
	go func() {
		defer wg.Done()
		for i := 0; i < n; i++ {
			name := fmt.Sprintf("conc-%d", i)
			r, w, _ := os.Pipe()
			ms := newRemoteManagedSession(name, 5000+i, r)
			sm.mu.Lock()
			sm.sessions[name] = ms
			sm.mu.Unlock()
			w.Close()
		}
	}()

	// Concurrently kill sessions
	go func() {
		defer wg.Done()
		for i := 0; i < n; i++ {
			name := fmt.Sprintf("conc-%d", i)
			_ = sm.Kill(name)
		}
	}()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// No panic or data race
	case <-time.After(10 * time.Second):
		t.Fatal("concurrent inject/kill timed out")
	}
}

// ---------------------------------------------------------------------------
// Kill remote session (no cmd.Wait panic)
// ---------------------------------------------------------------------------

func TestRemoteSession_KillNoCmdWait(t *testing.T) {
	sm := NewSessionManager(nil)
	defer sm.Shutdown()

	ms, _ := injectRemoteSession(t, sm, "kill-remote", 6000)
	require.Nil(t, ms.cmd, "remote session should have nil cmd")

	// Kill should not panic (no cmd.Wait call)
	assert.NotPanics(t, func() {
		ms.Kill()
	})
	assert.False(t, ms.IsAlive())
}
