// Package terminal tests for SessionManager lifecycle manager.
// Story pty-02: SessionManager Lifecycle Manager
//
// RED Phase: These tests define expected behavior for all 6 ACs and 5 BDD scenarios.
// They should FAIL until the go-engineer implements the feature.

package terminal

// All tests in this file require a real PTY (creack/pty fork/exec).
// Add testing.Short() skip guard to any new test functions.

import (
	"context"
	"errors"
	"fmt"
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

// spawnShell is a convenience wrapper that spawns a default shell session.
func spawnShell(t *testing.T, sm *SessionManager, name string) *ManagedSession {
	t.Helper()
	ms, err := sm.Spawn(context.Background(), name, t.TempDir(), "")
	require.NoError(t, err, "Spawn(%q) should not error", name)
	require.NotNil(t, ms, "Spawn(%q) should return non-nil session", name)
	return ms
}

// ---------------------------------------------------------------------------
// AC-1: Spawn creates a PTY session (BDD Scenario 1, 3)
// ---------------------------------------------------------------------------

func TestSessionManager_AC1_SpawnDefaultShell(t *testing.T) {
	if testing.Short() {
		t.Skip("requires PTY (fork/exec)")
	}
	// BDD Scenario 1: Spawn and retrieve a session
	// BDD Scenario 3: Empty command uses SHELL env
	sm := NewSessionManager()
	defer sm.Shutdown()

	ctx := context.Background()
	dir := t.TempDir()

	ms, err := sm.Spawn(ctx, "default-shell", dir, "")
	require.NoError(t, err)
	require.NotNil(t, ms)

	// Session should be alive
	assert.True(t, ms.IsAlive(), "spawned session should be alive")

	// Session should have a valid PID
	assert.Greater(t, ms.pid, 0, "session PID should be positive")

	// Session name should match
	assert.Equal(t, "default-shell", ms.Name())
}

func TestSessionManager_AC1_SpawnWithRepoPath(t *testing.T) {
	if testing.Short() {
		t.Skip("requires PTY (fork/exec)")
	}
	sm := NewSessionManager()
	defer sm.Shutdown()

	dir := t.TempDir()
	ms, err := sm.Spawn(context.Background(), "repo-session", dir, "")
	require.NoError(t, err)
	require.NotNil(t, ms)

	// Verify the process working directory was set
	assert.Equal(t, dir, ms.cmd.Dir, "session cmd.Dir should be the provided repoPath")
}

func TestSessionManager_AC1_SpawnSetsTermEnv(t *testing.T) {
	if testing.Short() {
		t.Skip("requires PTY (fork/exec)")
	}
	sm := NewSessionManager()
	defer sm.Shutdown()

	ms, err := sm.Spawn(context.Background(), "env-check", t.TempDir(), "")
	require.NoError(t, err)
	require.NotNil(t, ms)

	// TERM=xterm-256color should be in the process environment
	found := false
	for _, env := range ms.cmd.Env {
		if env == "TERM=xterm-256color" {
			found = true
			break
		}
	}
	assert.True(t, found, "TERM=xterm-256color should be in the process environment")
}

func TestSessionManager_AC1_SpawnSetsProcessGroup(t *testing.T) {
	if testing.Short() {
		t.Skip("requires PTY (fork/exec)")
	}
	sm := NewSessionManager()
	defer sm.Shutdown()

	ms, err := sm.Spawn(context.Background(), "pgid-check", t.TempDir(), "")
	require.NoError(t, err)
	require.NotNil(t, ms)

	require.NotNil(t, ms.cmd.SysProcAttr, "SysProcAttr should be set")
	assert.True(t, ms.cmd.SysProcAttr.Setpgid, "Setpgid should be true")
}

// ---------------------------------------------------------------------------
// AC-2: Reject duplicate session names (BDD Scenario 2)
// ---------------------------------------------------------------------------

func TestSessionManager_AC2_DuplicateName(t *testing.T) {
	if testing.Short() {
		t.Skip("requires PTY (fork/exec)")
	}
	// BDD Scenario 2: Reject duplicate name
	sm := NewSessionManager()
	defer sm.Shutdown()

	ctx := context.Background()
	dir := t.TempDir()

	// Spawn the first session
	original, err := sm.Spawn(ctx, "dup", dir, "")
	require.NoError(t, err)
	require.NotNil(t, original)

	// Spawning again with the same name should fail
	duplicate, err := sm.Spawn(ctx, "dup", dir, "")
	assert.Nil(t, duplicate, "duplicate spawn should return nil session")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrSessionExists), "error should be ErrSessionExists, got: %v", err)

	// Original session should still be alive
	assert.True(t, original.IsAlive(), "original session should be unaffected")
	assert.True(t, sm.IsAlive("dup"), "manager should report original alive")
}

// ---------------------------------------------------------------------------
// AC-3: Get, Kill, and IsAlive lifecycle (BDD Scenario 1)
// ---------------------------------------------------------------------------

func TestSessionManager_AC3_GetKillLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("requires PTY (fork/exec)")
	}
	// BDD Scenario 1: Kill removes session from manager
	sm := NewSessionManager()
	defer sm.Shutdown()

	ms := spawnShell(t, sm, "sess1")

	// Get should return the session
	got, ok := sm.Get("sess1")
	require.True(t, ok, "Get should find the session")
	assert.Equal(t, ms, got, "Get should return the same session")
	assert.True(t, sm.IsAlive("sess1"), "IsAlive should be true before kill")

	// Kill the session
	err := sm.Kill("sess1")
	require.NoError(t, err, "Kill should not error")

	// After kill: Get returns false, IsAlive returns false
	got, ok = sm.Get("sess1")
	assert.False(t, ok, "Get should return false after Kill")
	assert.Nil(t, got, "Get should return nil after Kill")
	assert.False(t, sm.IsAlive("sess1"), "IsAlive should be false after Kill")
}

func TestSessionManager_AC3_KillNonExistent(t *testing.T) {
	if testing.Short() {
		t.Skip("requires PTY (fork/exec)")
	}
	// BDD Scenario 1: Kill non-existent session returns error
	sm := NewSessionManager()
	defer sm.Shutdown()

	err := sm.Kill("ghost")
	require.Error(t, err, "Kill non-existent should error")
	assert.True(t, errors.Is(err, ErrSessionNotFound), "error should be ErrSessionNotFound, got: %v", err)
}

func TestSessionManager_AC3_GetNonExistent(t *testing.T) {
	if testing.Short() {
		t.Skip("requires PTY (fork/exec)")
	}
	sm := NewSessionManager()

	got, ok := sm.Get("nope")
	assert.False(t, ok)
	assert.Nil(t, got)
}

func TestSessionManager_AC3_IsAliveNonExistent(t *testing.T) {
	if testing.Short() {
		t.Skip("requires PTY (fork/exec)")
	}
	sm := NewSessionManager()

	assert.False(t, sm.IsAlive("nope"), "IsAlive should be false for non-existent session")
}

// ---------------------------------------------------------------------------
// AC-4: FindByPID (BDD Scenario 5)
// ---------------------------------------------------------------------------

func TestSessionManager_AC4_FindByPID(t *testing.T) {
	if testing.Short() {
		t.Skip("requires PTY (fork/exec)")
	}
	// BDD Scenario 5: Find existing session by PID
	sm := NewSessionManager()
	defer sm.Shutdown()

	ms := spawnShell(t, sm, "pid-lookup")
	pid := ms.pid
	require.Greater(t, pid, 0, "PID must be positive")

	found, ok := sm.FindByPID(pid)
	require.True(t, ok, "FindByPID should find the session")
	assert.Equal(t, ms, found, "FindByPID should return the matching session")
	assert.Equal(t, "pid-lookup", found.Name())
}

func TestSessionManager_AC4_FindByPID_Unknown(t *testing.T) {
	if testing.Short() {
		t.Skip("requires PTY (fork/exec)")
	}
	// BDD Scenario 5: FindByPID returns nil for unknown PID
	sm := NewSessionManager()
	defer sm.Shutdown()

	_ = spawnShell(t, sm, "some-session")

	found, ok := sm.FindByPID(99999)
	assert.False(t, ok, "FindByPID should return false for unknown PID")
	assert.Nil(t, found, "FindByPID should return nil for unknown PID")
}

func TestSessionManager_AC4_FindByPID_EmptyManager(t *testing.T) {
	if testing.Short() {
		t.Skip("requires PTY (fork/exec)")
	}
	sm := NewSessionManager()

	found, ok := sm.FindByPID(1)
	assert.False(t, ok)
	assert.Nil(t, found)
}

// ---------------------------------------------------------------------------
// AC-5: Shutdown kills all sessions (BDD Scenario 4)
// ---------------------------------------------------------------------------

func TestSessionManager_AC5_Shutdown(t *testing.T) {
	if testing.Short() {
		t.Skip("requires PTY (fork/exec)")
	}
	// BDD Scenario 4: Shutdown kills all sessions
	sm := NewSessionManager()

	names := []string{"a", "b", "c"}
	sessions := make([]*ManagedSession, len(names))
	for i, name := range names {
		sessions[i] = spawnShell(t, sm, name)
	}

	// All alive before shutdown
	for _, name := range names {
		require.True(t, sm.IsAlive(name), "%s should be alive before shutdown", name)
	}

	sm.Shutdown()

	// All dead after shutdown
	for i, name := range names {
		assert.False(t, sessions[i].IsAlive(), "%s ManagedSession should be dead after shutdown", name)
		assert.False(t, sm.IsAlive(name), "%s should not be alive in manager after shutdown", name)
	}

	// List should return empty
	list := sm.List(nil)
	assert.Empty(t, list, "List should be empty after Shutdown")
}

func TestSessionManager_AC5_ShutdownIdempotent(t *testing.T) {
	if testing.Short() {
		t.Skip("requires PTY (fork/exec)")
	}
	sm := NewSessionManager()
	_ = spawnShell(t, sm, "x")

	sm.Shutdown()
	// Second Shutdown should not panic
	assert.NotPanics(t, func() { sm.Shutdown() })
}

// ---------------------------------------------------------------------------
// AC-6: Spawn with command string (BDD Scenario 1)
// ---------------------------------------------------------------------------

func TestSessionManager_AC6_SpawnWithCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("requires PTY (fork/exec)")
	}
	sm := NewSessionManager()
	defer sm.Shutdown()

	ms, err := sm.Spawn(context.Background(), "echo-test", t.TempDir(), "echo hello world")
	require.NoError(t, err)
	require.NotNil(t, ms)

	// The command should have been split into executable + args
	assert.True(t,
		strings.HasSuffix(ms.cmd.Path, "/echo") || ms.cmd.Path == "echo",
		"cmd.Path should be echo, got: %s", ms.cmd.Path)

	time.Sleep(cmdOutputSettleTime)

	snap := ms.scroll.Snapshot()
	assert.Contains(t, string(snap), "hello world",
		"PTY output should contain 'hello world'")
}

func TestSessionManager_AC6_SpawnWithCommandArgs(t *testing.T) {
	if testing.Short() {
		t.Skip("requires PTY (fork/exec)")
	}
	sm := NewSessionManager()
	defer sm.Shutdown()

	ms, err := sm.Spawn(context.Background(), "printf-test", t.TempDir(), "printf abc")
	require.NoError(t, err)
	require.NotNil(t, ms)

	time.Sleep(cmdOutputSettleTime)

	snap := ms.scroll.Snapshot()
	assert.Contains(t, string(snap), "abc",
		"PTY output should contain 'abc'")
}

// ---------------------------------------------------------------------------
// List with filter
// ---------------------------------------------------------------------------

func TestSessionManager_List_All(t *testing.T) {
	if testing.Short() {
		t.Skip("requires PTY (fork/exec)")
	}
	sm := NewSessionManager()
	defer sm.Shutdown()

	spawnShell(t, sm, "l1")
	spawnShell(t, sm, "l2")
	spawnShell(t, sm, "l3")

	all := sm.List(nil)
	assert.Len(t, all, 3, "List(nil) should return all sessions")
}

func TestSessionManager_List_WithFilter(t *testing.T) {
	if testing.Short() {
		t.Skip("requires PTY (fork/exec)")
	}
	sm := NewSessionManager()
	defer sm.Shutdown()

	spawnShell(t, sm, "agent-1")
	spawnShell(t, sm, "agent-2")
	spawnShell(t, sm, "shell-1")

	agents := sm.List(func(ms *ManagedSession) bool {
		return strings.HasPrefix(ms.Name(), "agent-")
	})
	assert.Len(t, agents, 2, "filter should return only agent- sessions")
}

func TestSessionManager_List_Empty(t *testing.T) {
	if testing.Short() {
		t.Skip("requires PTY (fork/exec)")
	}
	sm := NewSessionManager()
	list := sm.List(nil)
	assert.Empty(t, list)
}

// ---------------------------------------------------------------------------
// Concurrency safety
// ---------------------------------------------------------------------------

func TestSessionManager_ConcurrentSpawnAndKill(t *testing.T) {
	if testing.Short() {
		t.Skip("requires PTY (fork/exec)")
	}
	t.Parallel()

	sm := NewSessionManager()
	defer sm.Shutdown()

	ctx := context.Background()
	dir := t.TempDir()

	const n = 5
	var wg sync.WaitGroup
	wg.Add(2)

	// Spawn sessions concurrently
	go func() {
		defer wg.Done()
		for i := 0; i < n; i++ {
			name := fmt.Sprintf("conc-%d", i)
			_, _ = sm.Spawn(ctx, name, dir, "")
		}
	}()

	// Concurrently kill sessions (some may not exist yet — that's fine)
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
		// No panic or data race (run with -race to verify)
	case <-time.After(10 * time.Second):
		t.Fatal("concurrent spawn/kill timed out")
	}
}
