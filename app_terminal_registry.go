package main

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
	"sort"
	"strings"
	"time"

	"mashed/internal/domain"
	"mashed/internal/terminal"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// recoverSessions scans tmux for existing sessions matching mashed's naming
// convention and registers them in the terminal session registry.
func (a *App) recoverSessions() {
	out, err := exec.Command("tmux", "list-sessions", "-F", "#{session_name}").Output()
	if err != nil {
		log.Printf("session recovery: tmux list-sessions failed (tmux may not be running): %v", err)
		return
	}
	n := a.recoverSessionsFromOutput(string(out))
	if n > 0 {
		log.Printf("session recovery: recovered %d session(s)", n)
	}
}

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
		case strings.HasPrefix(name, "term-"):
			sessionType = domain.SessionTerminal
			repoName = parseRepoName(name, "term-")
		case strings.HasPrefix(name, "mashed-"):
			sessionType = domain.SessionAgent
			repoName = parseRepoName(name, "mashed-")
		default:
			continue
		}

		session := domain.TerminalSession{
			SessionName: name,
			PaneTarget:  name + ":0.0",
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
