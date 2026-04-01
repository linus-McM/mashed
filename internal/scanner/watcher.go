package scanner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"

	"conductor/internal/domain"
)

// WatchSessions watches ~/.claude/projects/ for session file changes and emits events.
// It watches the top-level projects dir for new repo directories, and each repo dir
// for .jsonl file creates/modifications.
func (p *ClaudeCodeProvider) WatchSessions(ctx context.Context) (<-chan domain.SessionEvent, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, &WatchError{Op: "create", Path: p.claudeDir, Err: fmt.Errorf("creating watcher: %w", err)}
	}

	// Watch the top-level projects directory
	if err := watcher.Add(p.claudeDir); err != nil {
		watcher.Close()
		return nil, &WatchError{Op: "add", Path: p.claudeDir, Err: fmt.Errorf("watching projects dir: %w", err)}
	}

	// Watch existing repo subdirectories
	entries, err := os.ReadDir(p.claudeDir)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				_ = watcher.Add(filepath.Join(p.claudeDir, entry.Name()))
			}
		}
	}

	ch := make(chan domain.SessionEvent, 64)

	go p.watchLoop(ctx, watcher, ch)

	return ch, nil
}

func (p *ClaudeCodeProvider) watchLoop(ctx context.Context, watcher *fsnotify.Watcher, ch chan<- domain.SessionEvent) {
	defer close(ch)
	defer watcher.Close()

	for {
		select {
		case <-ctx.Done():
			return

		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			se, shouldEmit := p.handleFSEvent(event, watcher)
			if !shouldEmit {
				continue
			}

			select {
			case ch <- se:
			case <-ctx.Done():
				return
			}

		case _, ok := <-watcher.Errors:
			if !ok {
				return
			}
			// Watcher errors are non-fatal; continue watching
		}
	}
}

func (p *ClaudeCodeProvider) handleFSEvent(event fsnotify.Event, watcher *fsnotify.Watcher) (domain.SessionEvent, bool) {
	path := event.Name

	// New directory under projects/ — start watching it for session files
	if event.Has(fsnotify.Create) {
		fi, err := os.Stat(path)
		if err == nil && fi.IsDir() {
			_ = watcher.Add(path)
			return domain.SessionEvent{}, false
		}
	}

	// Only emit events for .jsonl files
	if !strings.HasSuffix(path, ".jsonl") {
		return domain.SessionEvent{}, false
	}

	if !event.Has(fsnotify.Create) && !event.Has(fsnotify.Write) {
		return domain.SessionEvent{}, false
	}

	isNew := event.Has(fsnotify.Create)
	repoPath := p.repoPathFromSessionDir(filepath.Dir(path))

	return domain.SessionEvent{
		SessionPath: path,
		RepoPath:    repoPath,
		IsNew:       isNew,
		Timestamp:   time.Now(),
	}, true
}

// repoPathFromSessionDir converts a session dir key back to a repo path.
// ~/.claude/projects/-Users-linus-Development-mashed -> /Users/linus/Development/mashed
func (p *ClaudeCodeProvider) repoPathFromSessionDir(sessionDir string) string {
	key := filepath.Base(sessionDir)
	if strings.HasPrefix(key, "-") {
		key = key[1:]
	}
	return "/" + strings.ReplaceAll(key, "-", "/")
}
