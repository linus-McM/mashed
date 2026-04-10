package main

import (
	"fmt"
	"log"
	"os/exec"
	"sort"
	"strings"
	"time"

	"mashed/internal/domain"
	"mashed/internal/scanner"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// initScanning starts all background goroutines for a given dev directory.
// Returns an error if the provider fails to initialize so callers can react.
func (a *App) initScanning(devDir string) error {
	a.devDir = devDir

	provider, err := scanner.NewClaudeCodeProvider(devDir)
	if err != nil {
		log.Printf("failed to init claude provider: %v", err)
		return fmt.Errorf("init claude provider: %w", err)
	}
	a.provider = provider
	a.repoScanner = scanner.NewRepoScanner(devDir)

	go a.scanLoop()
	go a.watchSessions()
	go a.consumeEngineEvents()

	// Tell frontend setup is done
	runtime.EventsEmit(a.ctx, "needs-setup", false)
	return nil
}

// scanLoop polls for running processes every 5 seconds.
func (a *App) scanLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	a.doScan()

	for {
		select {
		case <-a.ctx.Done():
			return
		case <-ticker.C:
			a.doScan()
		}
	}
}

// resolveTmuxTarget returns the terminal target for a process PID.
// Managed PTY sessions take priority; tmux panes are the fallback for
// externally spawned sessions (BMAD, manual tmux).
func (a *App) resolveTmuxTarget(pid int) string {
	if sess, ok := a.manager.FindByPID(pid); ok {
		return sess.Name()
	}
	if pane, err := a.panes.FindPaneForPID(pid); err == nil && pane != nil {
		return pane.Target()
	}
	return ""
}

// doScan performs one round of process scanning and feeds updates to the engine.
func (a *App) doScan() {
	sessions, err := a.provider.ScanProcesses()
	if err != nil {
		log.Printf("process scan error: %v", err)
		return
	}

	// Build repo info for branch lookups
	repos, _ := a.repoScanner.ScanRepos(nil)
	repoBranch := make(map[string]string) // path -> branch
	for _, r := range repos {
		repoBranch[r.Path] = r.Branch
	}

	// Track which agent IDs are alive this scan so we can prune dead ones
	seenAgentIDs := make(map[string]bool)

	// Sort sessions newest-first so the most recently started process claims
	// the most recently modified session file (prevents stale file mismatches
	// when multiple processes lack --session-id).
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].StartedAt.After(sessions[j].StartedAt)
	})

	// Track claimed session files so two agents in the same repo don't share one
	claimedSessions := make(map[string]bool) // sessionDir/sessionID -> true

	for _, s := range sessions {
		dir, err := a.provider.GetWorkingDir(s.PID)
		if err != nil {
			continue
		}
		// Resolve to git repo root so agents in subdirs group under the repo
		if out, err := exec.Command("git", "-C", dir, "rev-parse", "--show-toplevel").Output(); err == nil {
			dir = strings.TrimSpace(string(out))
		}
		// Skip agents whose repo root is outside the configured dev directory
		if !strings.HasPrefix(dir, a.devDir+"/") && dir != a.devDir {
			continue
		}

		agentID := fmt.Sprintf("pid-%d", s.PID)
		seenAgentIDs[agentID] = true
		model := s.Model
		if model == "" {
			model = "claude"
		}

		// Parse session data — use the agent's specific session file if possible
		sessionDir := a.provider.SessionDir(dir)
		var sessionData *domain.SessionData
		if s.SessionID != "" {
			sessionData = a.findSessionByID(sessionDir, s.SessionID)
			if sessionData != nil {
				claimedSessions[sessionDir+"/"+s.SessionID] = true
			}
		}
		if sessionData == nil {
			sessionData = a.findUnclaimed(sessionDir, claimedSessions)
		}

		var tokensUsed int64
		if sessionData != nil {
			tokensUsed = sessionData.TotalTokens
		}
		// Look up context window from model registry.
		modelInfo := domain.ModelByAlias(model)
		tokensMax := int64(modelInfo.ContextWindow)

		// Determine status from session data
		status := domain.StatusRunning
		if sessionData != nil {
			status = a.inferStatus(sessionData)
		}

		tmuxTarget := a.resolveTmuxTarget(s.PID)
		if tmuxTarget == "" {
			continue
		}

		var logLines []domain.LogLine
		if sessionData != nil && len(sessionData.LogLines) > 0 {
			logLines = sessionData.LogLines
		}

		ag := domain.Agent{
			ID:          agentID,
			Name:        model,
			Model:       model,
			Status:      status,
			PID:         s.PID,
			TokensUsed:  tokensUsed,
			TokensMax:   tokensMax,
			Elapsed:     time.Since(s.StartedAt),
			HasTmuxPane: tmuxTarget != "",
			TmuxTarget:  tmuxTarget,
			RepoPath:    dir,
			LogLines:    logLines,
		}

		repoName := repoNameFromDir(dir)
		branch := repoBranch[dir]

		_ = a.engine.ProcessAgentUpdate(ag, repoName, branch)

		// Emit sub-agent events from the session's parsed sub-agents
		if sessionData != nil {
			for _, sub := range sessionData.SubAgents {
				subID := fmt.Sprintf("%s-sub-%s-%s", agentID, sanitizeID(sub.Name), sub.ToolUseID)
				subStatus := domain.StatusRunning
				if sub.Status == "done" {
					subStatus = domain.StatusDone
				}

				subAg := domain.Agent{
					ID:          subID,
					Name:        sub.Name,
					Model:       model,
					Status:      subStatus,
					PID:         0, // sub-agents don't have their own PID
					TokensUsed:  0,
					TokensMax:   0,
					HasTmuxPane: true,
					TmuxTarget:  tmuxTarget, // use parent's pane
					RepoPath:    dir,
					LogLines:    sub.LogLines,
					// Sub-agent metadata carried via the engine event
					SubAgentInfo: &sub,
				}

				_ = a.engine.ProcessAgentUpdate(subAg, repoName, branch)
			}
		}
	}

	// Prune agents whose process is gone (session killed / exited).
	// Sub-agents are pruned if their parent is gone.
	a.mu.Lock()
	pruned := a.notifications[:0]
	for _, n := range a.notifications {
		keep := seenAgentIDs[n.AgentID]
		if n.IsSubAgent {
			keep = seenAgentIDs[n.ParentAgentID]
		}
		if keep {
			pruned = append(pruned, n)
		} else {
			a.engine.RemoveAgent(n.AgentID)
		}
	}
	a.notifications = pruned
	a.mu.Unlock()

	// Emit repos to frontend
	runtime.EventsEmit(a.ctx, "repos", repos)
}
