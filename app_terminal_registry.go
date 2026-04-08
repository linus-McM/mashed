package main

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
	"sort"

	"mashed/internal/domain"
	"mashed/internal/terminal"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) registerSession(session domain.TerminalSession) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.terminalSessions[session.SessionName] = session
}

// ListRepoSessions returns all live terminal sessions for the given repo path.
// Dead sessions (whose tmux session no longer exists) are pruned from the registry.
func (a *App) ListRepoSessions(repoPath string) []domain.TerminalSession {
	if repoPath == "" {
		return []domain.TerminalSession{}
	}

	// Call ListPanes outside the lock to avoid holding a.mu while shelling out.
	panes, err := a.panes.ListPanes()
	aliveSet := make(map[string]bool)
	if err != nil {
		var termErr *terminal.TerminalError
		if errors.As(err, &termErr) && errors.Is(termErr.Err, terminal.ErrTmuxNotRunning) {
			// tmux not running — all sessions are dead
		} else {
			log.Printf("ListPanes failed: %v", err)
		}
	} else {
		for _, p := range panes {
			aliveSet[p.SessionName] = true
		}
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	var result []domain.TerminalSession
	for name, sess := range a.terminalSessions {
		if sess.RepoPath != repoPath {
			continue
		}
		if aliveSet[name] {
			sess.IsAlive = true
			result = append(result, sess)
		} else {
			delete(a.terminalSessions, name)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].SpawnedAt.Before(result[j].SpawnedAt)
	})

	if result == nil {
		return []domain.TerminalSession{}
	}
	return result
}

// KillTerminalSession removes a session from the registry and kills its tmux session.
// Returns an error only if the session is not found in the registry; a tmux kill
// failure (e.g., session already dead) is logged but not treated as an error.
func (a *App) KillTerminalSession(sessionName string) error {
	a.mu.Lock()
	_, exists := a.terminalSessions[sessionName]
	if !exists {
		a.mu.Unlock()
		return fmt.Errorf("session %q not found in registry", sessionName)
	}
	delete(a.terminalSessions, sessionName)
	a.mu.Unlock()

	if err := exec.Command("tmux", "kill-session", "-t", sessionName).Run(); err != nil {
		log.Printf("tmux kill-session %s (may already be dead): %v", sessionName, err)
	}

	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "terminal:session:removed", sessionName)
	}
	return nil
}
