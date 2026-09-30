package terminal

import (
	"context"
	"errors"
	"os"
	"sync"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mashed/internal/terminal/helper"
)

// fakeSpawner records SpawnRequests and hands back a pipe as the "pty".
type fakeSpawner struct {
	mu   sync.Mutex
	reqs []helper.SpawnRequest
}

func (f *fakeSpawner) Spawn(_ context.Context, req helper.SpawnRequest) (*os.File, int, error) {
	f.mu.Lock()
	f.reqs = append(f.reqs, req)
	f.mu.Unlock()
	r, w, err := os.Pipe()
	if err != nil {
		return nil, 0, err
	}
	w.Close() // readLoop sees EOF immediately; the session just ends
	return r, 424242, nil
}
func (f *fakeSpawner) Kill(string, syscall.Signal) error { return nil }
func (f *fakeSpawner) Close() error                      { return nil }

func (f *fakeSpawner) last(t *testing.T) helper.SpawnRequest {
	t.Helper()
	f.mu.Lock()
	defer f.mu.Unlock()
	require.NotEmpty(t, f.reqs)
	return f.reqs[len(f.reqs)-1]
}

// R15: argv elements reach the helper unchanged, one element each.
func TestSpawnArgv_PassesArgsVerbatim(t *testing.T) {
	fs := &fakeSpawner{}
	sm := newSessionManagerWith(fs)
	prompt := "Review PR #12:\n  - don't \"quote\" me\n  - $(rm -rf /) `x`"
	argv := []string{"sh", "--model", "sonnet", "-p", prompt}

	_, err := sm.SpawnArgv(context.Background(), "s1", t.TempDir(), argv, 0, 0)
	require.NoError(t, err)

	req := fs.last(t)
	assert.Equal(t, resolveExecutable("sh"), req.Shell)
	assert.Equal(t, []string{"--model", "sonnet", "-p", prompt}, req.Args)
}

// xterm.js 6 supports synchronized output (DEC 2026); tell Claude Code so,
// since it only auto-detects a fixed list of terminals.
func TestSpawnArgv_EnvForcesClaudeSyncOutput(t *testing.T) {
	fs := &fakeSpawner{}
	sm := newSessionManagerWith(fs)
	_, err := sm.SpawnArgv(context.Background(), "s-env", t.TempDir(), []string{"sh"}, 0, 0)
	require.NoError(t, err)
	assert.Contains(t, fs.last(t).Env, "CLAUDE_CODE_FORCE_SYNC_OUTPUT=1")
}

func TestSpawnArgv_RejectsEmpty(t *testing.T) {
	sm := newSessionManagerWith(&fakeSpawner{})
	for _, argv := range [][]string{nil, {}, {""}} {
		_, err := sm.SpawnArgv(context.Background(), "s", "", argv, 0, 0)
		var te *TerminalError
		assert.True(t, errors.As(err, &te), "argv %q: err = %v, want TerminalError", argv, err)
	}
}

// Spawn keeps its string form: "" means $SHELL, otherwise whitespace split.
func TestSpawn_StringFormUnchanged(t *testing.T) {
	fs := &fakeSpawner{}
	sm := newSessionManagerWith(fs)
	t.Setenv("SHELL", "/bin/sh")

	_, err := sm.Spawn(context.Background(), "a", "", "", 0, 0)
	require.NoError(t, err)
	assert.Equal(t, resolveExecutable("/bin/sh"), fs.last(t).Shell)
	assert.Empty(t, fs.last(t).Args)

	_, err = sm.Spawn(context.Background(), "b", "", "echo  a b", 0, 0)
	require.NoError(t, err)
	assert.Equal(t, resolveExecutable("echo"), fs.last(t).Shell)
	assert.Equal(t, []string{"a", "b"}, fs.last(t).Args)
}

// A nil *helper.Client must not become a typed-nil interface.
func TestSessionManager_NilClient_SpawnArgvReturnsErrHelperNotRunning(t *testing.T) {
	sm := NewSessionManager(nil)
	_, err := sm.SpawnArgv(context.Background(), "x", "", []string{"sh"}, 0, 0)
	assert.ErrorIs(t, err, ErrHelperNotRunning)
	sm.Shutdown() // must not panic on a nil client
}
