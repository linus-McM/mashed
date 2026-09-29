package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"mashed/internal/domain"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// spawnSession creates a new managed PTY session with the given prefix, working
// directory, and optional shell command. It returns the session name.
// If command is empty, the session starts the user's default shell.
// cols/rows set the initial PTY winsize; 0 falls back to defaults (80x24).
func (a *App) spawnSession(prefix, repoPath, command string, sessionType domain.SessionType, model string, cols, rows uint16) (string, error) {
	now := time.Now()
	repoName := repoNameFromDir(repoPath)
	sessionName := fmt.Sprintf("%s-%s-%d", prefix, repoName, now.Unix())

	if _, err := a.manager.Spawn(a.ctx, sessionName, repoPath, command, cols, rows); err != nil {
		return "", fmt.Errorf("spawn session failed: %w", err)
	}

	session := domain.TerminalSession{
		SessionName: sessionName,
		PaneTarget:  sessionName,
		RepoPath:    repoPath,
		RepoName:    repoName,
		SessionType: sessionType,
		Model:       model,
		SpawnedAt:   now,
		IsAlive:     true,
	}
	a.registerSession(session)
	runtime.EventsEmit(a.ctx, eventSessionAdded, session)

	log.Printf("spawned session %s at %s", sessionName, repoPath)
	return sessionName, nil
}

// spawnSessionArgv is spawnSession for a pre-split argv: every element
// reaches the process unchanged (R15/R16). Use it for commands built in code,
// such as claude with a multi-line prompt.
func (a *App) spawnSessionArgv(prefix, repoPath string, argv []string, sessionType domain.SessionType, model string) (string, error) {
	now := time.Now()
	repoName := repoNameFromDir(repoPath)
	sessionName := fmt.Sprintf("%s-%s-%d", prefix, repoName, now.Unix())

	if _, err := a.manager.SpawnArgv(a.ctx, sessionName, repoPath, argv, 0, 0); err != nil {
		return "", fmt.Errorf("spawn session failed: %w", err)
	}

	session := domain.TerminalSession{
		SessionName: sessionName,
		PaneTarget:  sessionName,
		RepoPath:    repoPath,
		RepoName:    repoName,
		SessionType: sessionType,
		Model:       model,
		SpawnedAt:   now,
		IsAlive:     true,
	}
	a.registerSession(session)
	a.emitEvent(eventSessionAdded, session)

	log.Printf("spawned session %s at %s", sessionName, repoPath)
	return sessionName, nil
}

// SpawnAgent starts a new Claude session in a managed PTY for the given repo.
// Returns the session name for the terminal bridge.
// cols/rows set the initial PTY winsize so claude's first paint matches the
// frontend viewport; 0 means use the helper defaults.
func (a *App) SpawnAgent(repoPath string, model string, cols, rows uint16) (string, error) {
	if repoPath == "" {
		return "", fmt.Errorf("empty repo path")
	}
	if model == "" {
		model = domain.DefaultAlias(a.ListModels())
	}
	cmd := fmt.Sprintf("claude --dangerously-skip-permissions --model %s", model)
	return a.spawnSession("mashed", repoPath, cmd, domain.SessionAgent, model, cols, rows)
}

// SpawnAgentWithCommand starts a Claude session using a fully built CLI command.
// Returns the session name.
func (a *App) SpawnAgentWithCommand(repoPath, command string, cols, rows uint16) (string, error) {
	if repoPath == "" || command == "" {
		return "", fmt.Errorf("repo path and command are required")
	}
	return a.spawnSession("mashed", repoPath, command, domain.SessionAgent, "", cols, rows)
}

// SpawnTerminal starts a plain shell PTY session in the given repo directory.
// Returns the session name for the terminal bridge.
func (a *App) SpawnTerminal(repoPath string, cols, rows uint16) (string, error) {
	if repoPath == "" {
		return "", fmt.Errorf("empty repo path")
	}
	return a.spawnSession("term", repoPath, "", domain.SessionTerminal, "", cols, rows)
}

// GetAgentLog returns the parsed log lines for an agent's latest session.
func (a *App) GetAgentLog(repoPath string) []domain.LogLine {
	if a.provider == nil || repoPath == "" {
		return nil
	}
	sessionDir := a.provider.SessionDir(repoPath)
	data := a.findLatestSession(sessionDir)
	if data == nil {
		return nil
	}
	// Return last 200 lines
	lines := data.LogLines
	if len(lines) > 200 {
		lines = lines[len(lines)-200:]
	}
	return lines
}

// KillAgent terminates an agent process, kills its managed session, and removes it from tracking.
// The tmuxTarget parameter is a fallback used when the agent was spawned from the UI
// and doesn't yet have a matching entry in the backend notification list.
func (a *App) KillAgent(agentID string, pid int, tmuxTarget string) error {
	a.mu.Lock()
	for _, n := range a.notifications {
		if n.AgentID == agentID && n.TmuxTarget != "" {
			tmuxTarget = n.TmuxTarget
			break
		}
	}
	a.mu.Unlock()

	if tmuxTarget != "" {
		sessionName := tmuxTarget
		if idx := strings.Index(sessionName, ":"); idx > 0 {
			sessionName = sessionName[:idx]
		}

		// Try manager kill first (managed PTY sessions)
		if err := a.manager.Kill(sessionName); err != nil {
			log.Printf("manager kill %s: %v (may be external)", sessionName, err)
		}

		a.deregisterSession(sessionName)
	}

	// If the process is still alive (e.g. externally spawned), signal directly
	if pid > 0 {
		proc, err := os.FindProcess(pid)
		if err == nil {
			if err := proc.Signal(os.Interrupt); err != nil {
				_ = proc.Kill()
			}
		}
	}

	// Also remove any sub-agents belonging to this parent
	a.engine.RemoveAgent(agentID)
	a.mu.Lock()
	pruned := a.notifications[:0]
	for _, n := range a.notifications {
		if n.AgentID == agentID {
			continue
		}
		if n.IsSubAgent && n.ParentAgentID == agentID {
			a.engine.RemoveAgent(n.AgentID)
			continue
		}
		pruned = append(pruned, n)
	}
	a.notifications = pruned
	a.mu.Unlock()

	// Emit updated list to frontend
	runtime.EventsEmit(a.ctx, "agent:removed", agentID)
	return nil
}
