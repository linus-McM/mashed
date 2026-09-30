package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"mashed/internal/domain"
	"mashed/internal/scanner"
)

// watchSessions starts fsnotify-based session file watching.
// It runs until ctx ends and scans st on every session file change.
func (a *App) watchSessions(ctx context.Context, st *scanState) {
	a.sessionWatchers.Add(1)
	defer a.sessionWatchers.Add(-1)

	ch, err := st.provider.WatchSessions(ctx)
	if err != nil {
		log.Printf("session watcher error: %v", err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-ch:
			if !ok {
				return
			}
			// Trigger a scan on any session file change
			a.doScan(st)
		}
	}
}

// findSessionByID parses a specific session file by its ID.
func (a *App) findSessionByID(p *scanner.ClaudeCodeProvider, sessionDir, sessionID string) *domain.SessionData {
	path := filepath.Join(sessionDir, sessionID+".jsonl")
	if _, err := os.Stat(path); err != nil {
		return nil
	}
	data, err := p.ParseSession(path)
	if err != nil {
		return nil
	}
	return data
}

// findLatestSession finds and parses the most recently modified .jsonl file in a session directory.
func (a *App) findLatestSession(p *scanner.ClaudeCodeProvider, sessionDir string) *domain.SessionData {
	entries, err := os.ReadDir(sessionDir)
	if err != nil {
		return nil
	}

	var latestPath string
	var latestMod time.Time

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".jsonl" {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.ModTime().After(latestMod) {
			latestMod = info.ModTime()
			latestPath = filepath.Join(sessionDir, entry.Name())
		}
	}

	if latestPath == "" {
		return nil
	}

	data, err := p.ParseSession(latestPath)
	if err != nil {
		log.Printf("parse session %s: %v", latestPath, err)
		return nil
	}
	return data
}

// consumeEngineEvents reads from the engine's event channel and maintains the notification list.
func (a *App) consumeEngineEvents() {
	a.engineConsumers.Add(1)
	defer a.engineConsumers.Add(-1)
	for {
		select {
		case <-a.ctx.Done():
			return
		case evt, ok := <-a.engine.Events():
			if !ok {
				return
			}
			a.mu.Lock()
			found := false
			for i, n := range a.notifications {
				if n.AgentID == evt.AgentID {
					a.notifications[i] = evt
					found = true
					break
				}
			}
			if !found {
				a.notifications = append(a.notifications, evt)
			}
			sort.Slice(a.notifications, func(i, j int) bool {
				return a.notifications[i].Priority < a.notifications[j].Priority
			})
			a.mu.Unlock()
		}
	}
}

// findUnclaimed finds the most recent session file that hasn't been claimed by another agent.
func (a *App) findUnclaimed(p *scanner.ClaudeCodeProvider, sessionDir string, claimed map[string]bool) *domain.SessionData {
	entries, err := os.ReadDir(sessionDir)
	if err != nil {
		return nil
	}

	// Collect all .jsonl files sorted by mod time descending
	type fileEntry struct {
		path string
		key  string
		mod  time.Time
	}
	var files []fileEntry
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".jsonl" {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		sid := strings.TrimSuffix(entry.Name(), ".jsonl")
		files = append(files, fileEntry{
			path: filepath.Join(sessionDir, entry.Name()),
			key:  sessionDir + "/" + sid,
			mod:  info.ModTime(),
		})
	}

	// Sort newest first
	sort.Slice(files, func(i, j int) bool {
		return files[i].mod.After(files[j].mod)
	})

	for _, f := range files {
		if claimed[f.key] {
			continue
		}
		data, err := p.ParseSession(f.path)
		if err != nil {
			continue
		}
		claimed[f.key] = true
		return data
	}
	return nil
}

// inferStatus determines agent status from session data.
func (a *App) inferStatus(data *domain.SessionData) domain.AgentStatus {
	if len(data.LogLines) == 0 {
		return domain.StatusRunning
	}

	last := data.LogLines[len(data.LogLines)-1]
	if last.Kind == domain.LogErr {
		return domain.StatusError
	}

	// If the assistant's last action was AskUserQuestion — waiting for user
	if data.LastToolName == "AskUserQuestion" && data.HasPendingToolUse {
		return domain.StatusWaiting
	}

	// If assistant sent a tool_use and we haven't seen the user's tool_result yet — running
	if data.HasPendingToolUse {
		return domain.StatusRunning
	}

	// Last message was from assistant with no pending tool use — Claude is done talking
	if data.LastMessageType == "assistant" {
		// Check if this looks like a task completion (no tool calls, just text)
		if last.Kind == domain.LogInfo {
			return domain.StatusFinished
		}
		return domain.StatusOpen
	}

	return domain.StatusRunning
}
