package bmad

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

const (
	// assetChangeEvent is the Wails event name emitted on asset file changes.
	assetChangeEvent = "bmad:assets:changed"

	// assetDebounceDelay is the quiet period before emitting a coalesced change event.
	assetDebounceDelay = 500 * time.Millisecond
)

// AssetWatcher watches skill and command directories for .md file changes
// and emits a debounced event so the frontend can refresh.
type AssetWatcher struct {
	roots   []string
	emit    func(string, interface{})
	watcher *fsnotify.Watcher
	cancel  context.CancelFunc
	done    chan struct{}
	once    sync.Once
}

// NewAssetWatcher creates a watcher for the given root directories.
// The emit function is called with the event name and nil data when
// changes are detected after the debounce window.
func NewAssetWatcher(roots []string, emit func(string, interface{})) *AssetWatcher {
	return &AssetWatcher{
		roots: roots,
		emit:  emit,
		done:  make(chan struct{}),
	}
}

// Start begins watching the configured roots. Nonexistent roots are silently
// skipped. Returns nil even if no roots exist on disk. Must not be called twice.
func (w *AssetWatcher) Start(ctx context.Context) error {
	if w.watcher != nil {
		return fmt.Errorf("bmad: asset watcher already started")
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("bmad: creating asset watcher: %w", err)
	}
	w.watcher = watcher

	ctx, w.cancel = context.WithCancel(ctx)

	for _, root := range w.roots {
		w.addRecursive(root)
	}

	go w.loop(ctx)

	return nil
}

// Stop cancels the watcher and waits for the goroutine to exit.
// Safe to call multiple times.
func (w *AssetWatcher) Stop() {
	w.once.Do(func() {
		if w.cancel != nil {
			w.cancel()
			<-w.done
		}
	})
}

// addRecursive adds a directory and all its subdirectories to the watcher.
// Nonexistent or inaccessible directories are silently skipped.
func (w *AssetWatcher) addRecursive(root string) {
	fi, err := os.Stat(root)
	if err != nil || !fi.IsDir() {
		return
	}
	if err := w.watcher.Add(root); err != nil {
		log.Printf("bmad: asset watcher: cannot watch %s: %v", root, err)
		return
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() {
			w.addRecursive(filepath.Join(root, e.Name()))
		}
	}
}

func isMarkdown(path string) bool {
	return strings.EqualFold(filepath.Ext(path), ".md")
}

// loop is the main event loop. It debounces filesystem events and emits
// the coalesced change event after the quiet period.
func (w *AssetWatcher) loop(ctx context.Context) {
	defer close(w.done)
	defer w.watcher.Close()

	var timer *time.Timer
	var timerC <-chan time.Time

	for {
		select {
		case <-ctx.Done():
			if timer != nil {
				timer.Stop()
			}
			return

		case <-timerC:
			w.emit(assetChangeEvent, nil)
			timer = nil
			timerC = nil

		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}

			// New directory: add it to the watch set for dynamic subdirectory handling.
			if event.Has(fsnotify.Create) {
				fi, err := os.Stat(event.Name)
				if err == nil && fi.IsDir() {
					w.addRecursive(event.Name)
					continue
				}
			}

			// Only debounce .md file events.
			if !isMarkdown(event.Name) {
				continue
			}

			// Reset the debounce timer.
			if timer != nil {
				timer.Stop()
			}
			timer = time.NewTimer(assetDebounceDelay)
			timerC = timer.C

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("bmad: asset watcher error: %v", err)
		}
	}
}

// AssetWatchRoots computes the skill/command root directories to watch.
// If repoPath is non-empty, repo-local roots are included. Global roots
// under ~/.claude/ are always included.
func AssetWatchRoots(repoPath string) []string {
	var roots []string
	if repoPath != "" {
		roots = append(roots,
			filepath.Join(repoPath, ".claude", "skills"),
			filepath.Join(repoPath, ".claude", "commands"),
		)
	}
	home, _ := os.UserHomeDir()
	if home != "" {
		roots = append(roots,
			filepath.Join(home, ".claude", "skills"),
			filepath.Join(home, ".claude", "commands"),
		)
	}
	return roots
}
