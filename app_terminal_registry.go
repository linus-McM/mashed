package main

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"mashed/internal/domain"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// recoverSessionsFromOutput parses tmux list-sessions output and registers sessions.
// Returns the number of sessions recovered.
func (a *App) recoverSessionsFromOutput(sessionOutput string) int {
	count := 0
	for _, line := range strings.Split(sessionOutput, "\n") {
		name := strings.TrimSpace(line)
		if name == "" {
			continue
		}

		var sessionType domain.SessionType
		var repoName string

		switch {
		case strings.HasPrefix(name, prefixTerminal):
			sessionType = domain.SessionTerminal
			repoName = parseRepoName(name, prefixTerminal)
		case strings.HasPrefix(name, prefixAgent):
			sessionType = domain.SessionAgent
			repoName = parseRepoName(name, prefixAgent)
		default:
			continue
		}

		session := domain.TerminalSession{
			SessionName: name,
			PaneTarget:  name,
			RepoName:    repoName,
			SessionType: sessionType,
			SpawnedAt:   time.Now(),
		}
		a.registerSession(session)
		count++
	}
	return count
}

// parseRepoName extracts the repo name from a session name like "prefix-repoName-timestamp".
// The repo name may itself contain hyphens, so we strip the prefix and the last "-timestamp" segment.
func parseRepoName(sessionName, prefix string) string {
	rest := strings.TrimPrefix(sessionName, prefix) // "repoName-timestamp" or "my-repo-timestamp"
	lastDash := strings.LastIndex(rest, "-")
	if lastDash <= 0 {
		return rest
	}
	return rest[:lastDash]
}

// Session name prefixes used by spawnSession and recovery.
const (
	prefixTerminal = "term-"
	prefixAgent    = "mashed-"
)

// Wails event names for terminal session lifecycle.
const (
	eventSessionAdded   = "terminal:session:added"
	eventSessionRemoved = "terminal:session:removed"
)

func (a *App) registerSession(session domain.TerminalSession) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.terminalSessions[session.SessionName] = session
}

// deregisterSession removes a session from the registry by name.
func (a *App) deregisterSession(sessionName string) {
	a.mu.Lock()
	delete(a.terminalSessions, sessionName)
	a.mu.Unlock()
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, eventSessionRemoved, sessionName)
	}
}

// ListRepoSessions returns all live terminal sessions for the given repo path.
// Dead sessions (whose managed session is no longer alive) are pruned from the registry.
func (a *App) ListRepoSessions(repoPath string) []domain.TerminalSession {
	if repoPath == "" {
		return []domain.TerminalSession{}
	}

	// Snapshot matching sessions under lock, then release before calling IsAlive
	// (which acquires its own mutex) to avoid nested lock acquisition.
	a.mu.Lock()
	candidates := make([]domain.TerminalSession, 0, len(a.terminalSessions))
	for _, sess := range a.terminalSessions {
		if sess.RepoPath == repoPath {
			candidates = append(candidates, sess)
		}
	}
	a.mu.Unlock()

	var alive []domain.TerminalSession
	var dead []string
	for _, sess := range candidates {
		if a.manager.IsAlive(sess.SessionName) {
			sess.IsAlive = true
			alive = append(alive, sess)
		} else {
			dead = append(dead, sess.SessionName)
		}
	}

	// Prune dead sessions under lock.
	if len(dead) > 0 {
		a.mu.Lock()
		for _, name := range dead {
			delete(a.terminalSessions, name)
		}
		a.mu.Unlock()
	}

	sort.Slice(alive, func(i, j int) bool {
		return alive[i].SpawnedAt.Before(alive[j].SpawnedAt)
	})

	if alive == nil {
		return []domain.TerminalSession{}
	}
	return alive
}

// KillTerminalSession removes a session from the registry and kills its managed PTY.
// Returns an error only if the session is not found in the registry; a kill
// failure (e.g., session already dead) is logged but not treated as an error.
func (a *App) KillTerminalSession(sessionName string) error {
	a.mu.Lock()
	_, exists := a.terminalSessions[sessionName]
	a.mu.Unlock()
	if !exists {
		return fmt.Errorf("session %q not found in registry", sessionName)
	}

	if err := a.manager.Kill(sessionName); err != nil {
		log.Printf("kill session %s (may already be dead): %v", sessionName, err)
	}

	a.deregisterSession(sessionName)
	return nil
}
