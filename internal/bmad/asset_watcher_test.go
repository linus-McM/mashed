package bmad

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

// TestAssetWatcher_AC1_SingleWriteDebouncedEvent verifies that a single
// file write produces exactly one debounced event after ~500ms.
func TestAssetWatcher_AC1_SingleWriteDebouncedEvent(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "skills", "my-skill")
	require.NoError(t, os.MkdirAll(skillDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# test"), 0o644))

	var count atomic.Int32
	emit := func(event string, _ interface{}) {
		assert.Equal(t, assetChangeEvent, event)
		count.Add(1)
	}

	w := NewAssetWatcher([]string{filepath.Join(dir, "skills")}, emit)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	require.NoError(t, w.Start(ctx))
	defer w.Stop()

	// Let watcher initialise.
	time.Sleep(100 * time.Millisecond)

	// Single write.
	require.NoError(t, os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# modified"), 0o644))

	// Wait for debounce (500ms) + margin.
	time.Sleep(800 * time.Millisecond)
	assert.Equal(t, int32(1), count.Load(), "expected exactly one debounced event")

	// Confirm no spurious second event.
	time.Sleep(300 * time.Millisecond)
	assert.Equal(t, int32(1), count.Load(), "no second event should fire")
}

// TestAssetWatcher_AC2_BurstCoalescesToOneEvent verifies that five rapid
// writes coalesce into a single debounced event.
func TestAssetWatcher_AC2_BurstCoalescesToOneEvent(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "skills", "my-skill")
	require.NoError(t, os.MkdirAll(skillDir, 0o755))
	skillFile := filepath.Join(skillDir, "SKILL.md")
	require.NoError(t, os.WriteFile(skillFile, []byte("# test"), 0o644))

	var count atomic.Int32
	emit := func(_ string, _ interface{}) { count.Add(1) }

	w := NewAssetWatcher([]string{filepath.Join(dir, "skills")}, emit)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	require.NoError(t, w.Start(ctx))
	defer w.Stop()

	time.Sleep(100 * time.Millisecond)

	// Five rapid writes within ~200ms.
	for i := 0; i < 5; i++ {
		require.NoError(t, os.WriteFile(skillFile, []byte(fmt.Sprintf("# write %d", i)), 0o644))
		time.Sleep(40 * time.Millisecond)
	}

	// Wait for debounce after last write.
	time.Sleep(800 * time.Millisecond)
	assert.Equal(t, int32(1), count.Load(), "burst should coalesce to one event")
}

// TestAssetWatcher_AC3_NewSubdirectoryWatchedDynamically verifies that
// files written inside a newly created subdirectory trigger events.
func TestAssetWatcher_AC3_NewSubdirectoryWatchedDynamically(t *testing.T) {
	dir := t.TempDir()
	skillsRoot := filepath.Join(dir, "skills")
	require.NoError(t, os.MkdirAll(filepath.Join(skillsRoot, "existing-skill"), 0o755))

	var count atomic.Int32
	emit := func(_ string, _ interface{}) { count.Add(1) }

	w := NewAssetWatcher([]string{skillsRoot}, emit)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	require.NoError(t, w.Start(ctx))
	defer w.Stop()

	time.Sleep(100 * time.Millisecond)

	// Create new subdirectory.
	newSkillDir := filepath.Join(skillsRoot, "new-skill")
	require.NoError(t, os.MkdirAll(newSkillDir, 0o755))

	// Wait for the watcher to pick up the new directory (story: 1 second later).
	time.Sleep(200 * time.Millisecond)

	// Write a .md file inside the new directory.
	require.NoError(t, os.WriteFile(filepath.Join(newSkillDir, "SKILL.md"), []byte("# new"), 0o644))

	// Wait for debounce.
	time.Sleep(800 * time.Millisecond)
	assert.GreaterOrEqual(t, count.Load(), int32(1), "event should fire for file in new subdirectory")
}

// TestAssetWatcher_AC4_NonexistentRootSilentlySkipped verifies that
// Start returns nil when some configured roots do not exist, and
// existing roots are still watched normally.
func TestAssetWatcher_AC4_NonexistentRootSilentlySkipped(t *testing.T) {
	dir := t.TempDir()
	existingRoot := filepath.Join(dir, "skills")
	require.NoError(t, os.MkdirAll(existingRoot, 0o755))

	var count atomic.Int32
	w := NewAssetWatcher([]string{
		existingRoot,
		filepath.Join(dir, "nonexistent1"),
		filepath.Join(dir, "commands"), // does not exist
		filepath.Join(dir, "nonexistent2"),
	}, func(_ string, _ interface{}) { count.Add(1) })

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := w.Start(ctx)
	require.NoError(t, err, "Start must not fail for missing roots")
	defer w.Stop()

	time.Sleep(100 * time.Millisecond)
	require.NoError(t, os.WriteFile(filepath.Join(existingRoot, "test.md"), []byte("# hi"), 0o644))
	time.Sleep(800 * time.Millisecond)
	assert.Equal(t, int32(1), count.Load(), "existing root should still be watched")
}

// TestAssetWatcher_AC5_CleansUpOnContextCancel verifies that cancelling
// the context causes Stop to complete within 1 second and leaks no goroutines.
func TestAssetWatcher_AC5_CleansUpOnContextCancel(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	dir := t.TempDir()
	skillsRoot := filepath.Join(dir, "skills")
	require.NoError(t, os.MkdirAll(skillsRoot, 0o755))

	w := NewAssetWatcher([]string{skillsRoot}, func(string, interface{}) {})
	ctx, cancel := context.WithCancel(context.Background())

	require.NoError(t, w.Start(ctx))

	time.Sleep(100 * time.Millisecond)

	cancel()

	done := make(chan struct{})
	go func() {
		w.Stop()
		close(done)
	}()

	select {
	case <-done:
		// Success — Stop completed promptly.
	case <-time.After(1 * time.Second):
		t.Fatal("Stop did not complete within 1 second")
	}
}

// TestAssetWatcher_AC6_NonMarkdownFilesIgnored verifies that .DS_Store,
// .swp, and other non-markdown files do not trigger events.
func TestAssetWatcher_AC6_NonMarkdownFilesIgnored(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "skills", "my-skill")
	require.NoError(t, os.MkdirAll(skillDir, 0o755))

	var count atomic.Int32
	emit := func(_ string, _ interface{}) { count.Add(1) }

	w := NewAssetWatcher([]string{filepath.Join(dir, "skills")}, emit)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	require.NoError(t, w.Start(ctx))
	defer w.Stop()

	time.Sleep(100 * time.Millisecond)

	// Create non-markdown files.
	require.NoError(t, os.WriteFile(filepath.Join(skillDir, ".DS_Store"), []byte(""), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(skillDir, "foo.swp"), []byte(""), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(skillDir, "data.json"), []byte("{}"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(skillDir, "notes.txt"), []byte("hi"), 0o644))

	// Wait past the debounce period.
	time.Sleep(800 * time.Millisecond)
	assert.Equal(t, int32(0), count.Load(), "non-markdown files must not trigger events")
}

// TestAssetWatcher_DoubleStartReturnsError verifies that calling Start
// twice on the same watcher returns an error.
func TestAssetWatcher_DoubleStartReturnsError(t *testing.T) {
	dir := t.TempDir()
	w := NewAssetWatcher([]string{dir}, func(string, interface{}) {})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	require.NoError(t, w.Start(ctx))
	defer w.Stop()

	err := w.Start(ctx)
	assert.Error(t, err, "second Start should return error")
	assert.Contains(t, err.Error(), "already started")
}

// TestAssetWatchRoots verifies the root computation helper.
func TestAssetWatchRoots(t *testing.T) {
	t.Run("with repo path", func(t *testing.T) {
		roots := AssetWatchRoots("/tmp/myrepo")
		assert.GreaterOrEqual(t, len(roots), 4, "should include repo-local + global roots")
		assert.Equal(t, "/tmp/myrepo/.claude/skills", roots[0])
		assert.Equal(t, "/tmp/myrepo/.claude/commands", roots[1])
	})

	t.Run("without repo path", func(t *testing.T) {
		roots := AssetWatchRoots("")
		// Only global roots.
		for _, r := range roots {
			assert.NotContains(t, r, "/.claude/skills"+string(filepath.Separator), "should not contain repo-local path with empty repo")
		}
		assert.GreaterOrEqual(t, len(roots), 2, "should include at least global roots")
	})
}

// TestIsMarkdown verifies the extension filter.
func TestIsMarkdown(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"SKILL.md", true},
		{"README.MD", true},
		{"notes.md", true},
		{".DS_Store", false},
		{"foo.swp", false},
		{"data.json", false},
		{"file.txt", false},
		{"noext", false},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, isMarkdown(tt.path), "isMarkdown(%q)", tt.path)
	}
}
