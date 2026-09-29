package main

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// R19: the asset watcher follows SetActiveContext.
func TestAssetWatcher_FollowsActiveRepo(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	repoA, repoB := t.TempDir(), t.TempDir()
	for _, r := range []string{repoA, repoB} {
		require.NoError(t, os.MkdirAll(filepath.Join(r, ".claude", "skills"), 0o755))
	}

	var changes atomic.Int32
	orig := appEmitHook
	appEmitHook = func(name string, _ ...any) {
		if name == "bmad:assets:changed" {
			changes.Add(1)
		}
	}
	t.Cleanup(func() { appEmitHook = orig })

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	app := &App{ctx: ctx}
	defer app.stopAssetWatcher()

	app.SetActiveContext(repoA, "")
	app.SetActiveContext(repoB, "")
	w := app.currentAssetWatcher()
	require.NotNil(t, w)

	// A change under repoB is reported within the debounce window + 1s.
	require.NoError(t, os.WriteFile(filepath.Join(repoB, ".claude", "skills", "b.md"), []byte("b"), 0o644))
	require.Eventually(t, func() bool { return changes.Load() > 0 }, 1500*time.Millisecond, 20*time.Millisecond,
		"change under the active repo must be reported")

	// A change under the previous repo is not.
	changes.Store(0)
	require.NoError(t, os.WriteFile(filepath.Join(repoA, ".claude", "skills", "a.md"), []byte("a"), 0o644))
	time.Sleep(1200 * time.Millisecond)
	assert.Equal(t, int32(0), changes.Load(), "previous repo is no longer watched")

	// Re-selecting the same repo does not restart the watcher.
	app.SetActiveContext(repoB, "other-pane")
	assert.Same(t, w, app.currentAssetWatcher())
}

// R19: shutdown and SetActiveContext may race without a data race.
func TestAssetWatcher_ShutdownRacesSetActiveContext(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	orig := appEmitHook
	appEmitHook = func(string, ...any) {}
	t.Cleanup(func() { appEmitHook = orig })

	ctx, cancel := context.WithCancel(context.Background())
	app := &App{ctx: ctx, cancel: cancel}
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			app.SetActiveContext(t.TempDir(), "")
		}(i)
	}
	app.shutdown(ctx)
	wg.Wait()
	app.stopAssetWatcher()
}
