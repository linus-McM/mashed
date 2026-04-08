package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"mashed/internal/domain"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// spawnTmuxSession creates a new tmux session with the given prefix, working
// directory, and optional shell command. It returns the pane target string.
// If command is empty, the session starts a default shell.
func (a *App) spawnTmuxSession(prefix, repoPath, command string, sessionType domain.SessionType, model string) (string, error) {
	repoName := repoNameFromDir(repoPath)
	sessionName := fmt.Sprintf("%s-%s-%d", prefix, repoName, time.Now().Unix())

	args := []string{"new-session", "-d", "-s", sessionName, "-c", repoPath}
	if command != "" {
		args = append(args, command)
	}

	tmuxCmd := exec.CommandContext(a.ctx, "tmux", args...)
	if out, err := tmuxCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("tmux new-session failed: %w (%s)", err, string(out))
	}

	target := fmt.Sprintf("%s:0.0", sessionName)
	a.panes.InvalidateCache()

	session := domain.TerminalSession{
		SessionName: sessionName,
		PaneTarget:  target,
		RepoPath:    repoPath,
		RepoName:    repoName,
		SessionType: sessionType,
		Model:       model,
		SpawnedAt:   time.Now(),
		IsAlive:     true,
	}
	a.registerSession(session)
	runtime.EventsEmit(a.ctx, "terminal:session:added", session)

	log.Printf("spawned tmux session %s at %s", sessionName, repoPath)
	return target, nil
}

// SpawnAgent starts a new Claude session in a tmux pane for the given repo.
// Returns the tmux pane target string for the terminal bridge.
func (a *App) SpawnAgent(repoPath string, model string) (string, error) {
	if repoPath == "" {
		return "", fmt.Errorf("empty repo path")
	}
	if model == "" {
		model = "claude-opus-4-6"
	}
	cmd := fmt.Sprintf("claude --dangerously-skip-permissions --model %s", model)
	return a.spawnTmuxSession("mashed", repoPath, cmd, domain.SessionAgent, model)
}

// SpawnAgentWithCommand starts a Claude session using a fully built CLI command.
// Returns the tmux pane target string.
func (a *App) SpawnAgentWithCommand(repoPath, command string) (string, error) {
	if repoPath == "" || command == "" {
		return "", fmt.Errorf("repo path and command are required")
	}
	return a.spawnTmuxSession("mashed", repoPath, command, domain.SessionAgent, "")
}

// SpawnTerminal starts a plain shell tmux session in the given repo directory.
// Returns the tmux pane target string for the terminal bridge.
func (a *App) SpawnTerminal(repoPath string) (string, error) {
	if repoPath == "" {
		return "", fmt.Errorf("empty repo path")
	}
	return a.spawnTmuxSession("term", repoPath, "", domain.SessionTerminal, "")
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

// KillAgent terminates an agent process, kills its tmux session, and removes it from tracking.
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
		if err := exec.Command("tmux", "kill-session", "-t", sessionName).Run(); err != nil {
			log.Printf("tmux kill-session %s failed: %v", sessionName, err)
		}

		// Deregister from terminal session registry
		a.mu.Lock()
		delete(a.terminalSessions, sessionName)
		a.mu.Unlock()
		runtime.EventsEmit(a.ctx, "terminal:session:removed", sessionName)
	}

	// If the process is still alive (e.g. tmux kill didn't reach it), signal directly
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
