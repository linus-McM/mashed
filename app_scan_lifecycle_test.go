package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"mashed/internal/scanner"
)

// appWithDevDir returns an App whose scan state reports devDir, without
// starting any scanning goroutines (for tests of devDir-dependent bindings).
func appWithDevDir(dir string) *App {
	a := &App{}
	a.scan.Store(&scanState{devDir: dir})
	return a
}

// appWithDevDirCtx is appWithDevDir with a background context set.
func appWithDevDirCtx(dir string) *App {
	a := appWithDevDir(dir)
	a.ctx = context.Background()
	return a
}

// lifecycleApp returns an App ready for SetDevDir: temp HOME (config
// writes), a cancellable ctx, a fake session manager and silenced events.
func lifecycleApp(t *testing.T) (*App, context.CancelFunc) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	orig := appEmitHook
	appEmitHook = func(string, ...any) {}
	t.Cleanup(func() { appEmitHook = orig })
	ctx, cancel := context.WithCancel(context.Background())
	return &App{ctx: ctx, cancel: cancel, manager: newFakeManager()}, cancel
}

// R17: re-entrant SetDevDir: no goroutine leak, at most one scanLoop and one
// watchSessions at a time, and no engine consumer started per call.
func TestSetDevDir_RestartsWithoutLeaking(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	app, cancel := lifecycleApp(t)
	dirs := []string{t.TempDir(), t.TempDir()}

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		require.NoError(t, app.SetDevDir(dirs[i%2]))
		// Concurrent readers must be race-free (run with -race).
		wg.Add(1)
		go func() { defer wg.Done(); _ = app.GetDevDir() }()
		assert.LessOrEqual(t, app.scanLoops.Load(), int32(1), "at most one scanLoop")
		assert.LessOrEqual(t, app.sessionWatchers.Load(), int32(1), "at most one watchSessions")
	}
	wg.Wait()
	assert.Equal(t, int32(0), app.engineConsumers.Load(), "SetDevDir must not start engine consumers")

	want, _ := filepath.EvalSymlinks(dirs[1])
	assert.Equal(t, want, app.GetDevDir())

	app.stopScanning()
	cancel()
	require.Eventually(t, func() bool {
		return app.scanLoops.Load() == 0 && app.sessionWatchers.Load() == 0
	}, 2*time.Second, 10*time.Millisecond)
}

// R18: a directory whose provider fails leaves the previous scanners, devDir
// and persisted config untouched.
func TestSetDevDir_ProviderFailureKeepsPreviousState(t *testing.T) {
	app, cancel := lifecycleApp(t)
	defer func() { app.stopScanning(); cancel() }()

	good, bad := t.TempDir(), t.TempDir()
	goodResolved, _ := filepath.EvalSymlinks(good)
	badResolved, _ := filepath.EvalSymlinks(bad)
	app.newProvider = func(dir string) (*scanner.ClaudeCodeProvider, error) {
		if dir == badResolved {
			return nil, errors.New("provider boom")
		}
		return scanner.NewClaudeCodeProvider(dir)
	}

	require.NoError(t, app.SetDevDir(good))
	require.Error(t, app.SetDevDir(bad))

	assert.Equal(t, goodResolved, app.GetDevDir(), "devDir unchanged")
	assert.Equal(t, int32(1), app.scanLoops.Load(), "previous scanLoop still running")
	cfg := loadConfig()
	assert.Equal(t, good, cfg.DevDir, "config keeps the previous DevDir")
	_, err := os.Stat(configPath())
	require.NoError(t, err)
}
