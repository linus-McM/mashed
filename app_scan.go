package main

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"mashed/internal/agent"
	"mashed/internal/domain"
	"mashed/internal/scanner"
)

// scanState is one immutable snapshot of what scanning works against.
// provider and repoScanner are nil when scanning could not start.
type scanState struct {
	devDir      string // root directory to scan for repos (symlink-resolved)
	provider    *scanner.ClaudeCodeProvider
	repoScanner *scanner.RepoScanner
}

// scanSnapshot returns the current scan state; never nil.
func (a *App) scanSnapshot() *scanState {
	if st := a.scan.Load(); st != nil {
		return st
	}
	return &scanState{}
}

// initScanning (re)starts the background scanners for devDir. It is safe to
// call repeatedly (R17): the new provider is built first — on failure the
// previous scanners keep running (R18) — then the old goroutines are
// cancelled and waited for before the new ones start.
func (a *App) initScanning(devDir string) error {
	// Canonicalize devDir so the symlink-vs-real-path comparison in doScan
	// matches what `git rev-parse --show-toplevel` returns.
	if resolved, err := filepath.EvalSymlinks(devDir); err == nil {
		devDir = resolved
	}

	newProvider := a.newProvider
	if newProvider == nil {
		newProvider = scanner.NewClaudeCodeProvider
	}
	provider, err := newProvider(devDir)
	if err != nil {
		log.Printf("failed to init claude provider: %v", err)
		return fmt.Errorf("init claude provider: %w", err)
	}
	st := &scanState{devDir: devDir, provider: provider, repoScanner: scanner.NewRepoScanner(devDir)}

	a.scanMu.Lock()
	defer a.scanMu.Unlock()
	a.stopScanningLocked()
	a.scan.Store(st)

	ctx, cancel := context.WithCancel(a.ctx)
	a.scanCancel = cancel
	a.scanWG.Add(2)
	go func() { defer a.scanWG.Done(); a.scanLoop(ctx, st) }()
	go func() { defer a.scanWG.Done(); a.watchSessions(ctx, st) }()

	// Tell frontend setup is done
	a.emitEvent("needs-setup", false)
	return nil
}

// stopScanning cancels the scanners and waits for them to exit.
func (a *App) stopScanning() {
	a.scanMu.Lock()
	defer a.scanMu.Unlock()
	a.stopScanningLocked()
}

func (a *App) stopScanningLocked() {
	if a.scanCancel != nil {
		a.scanCancel()
		a.scanCancel = nil
	}
	a.scanWG.Wait()
}

// scanLoop polls for running processes every 5 seconds until ctx ends.
func (a *App) scanLoop(ctx context.Context, st *scanState) {
	a.scanLoops.Add(1)
	defer a.scanLoops.Add(-1)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	a.doScan(st)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			a.doScan(st)
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
func (a *App) doScan(st *scanState) {
	if st.provider == nil || st.repoScanner == nil {
		return
	}
	sessions, err := st.provider.ScanProcesses()
	if err != nil {
		log.Printf("process scan error: %v", err)
		return
	}

	// Build repo info for branch lookups
	repos, _ := st.repoScanner.ScanRepos(nil)
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
		dir, err := st.provider.GetWorkingDir(s.PID)
		if err != nil {
			continue
		}
		// Resolve to git repo root so agents in subdirs group under the repo
		if out, err := exec.Command("git", "-C", dir, "rev-parse", "--show-toplevel").Output(); err == nil {
			dir = strings.TrimSpace(string(out))
		}
		// Skip agents whose repo root is outside the configured dev directory
		if !strings.HasPrefix(dir, st.devDir+"/") && dir != st.devDir {
			continue
		}

		agentID := fmt.Sprintf("pid-%d", s.PID)
		seenAgentIDs[agentID] = true
		model := s.Model
		if model == "" {
			model = "claude"
		}

		// Parse session data — use the agent's specific session file if possible
		sessionDir := st.provider.SessionDir(dir)
		var sessionData *domain.SessionData
		if s.SessionID != "" {
			sessionData = a.findSessionByID(st.provider, sessionDir, s.SessionID)
			if sessionData != nil {
				claimedSessions[sessionDir+"/"+s.SessionID] = true
			}
		}
		if sessionData == nil {
			sessionData = a.findUnclaimed(st.provider, sessionDir, claimedSessions)
		}

		var tokensUsed int64
		if sessionData != nil {
			tokensUsed = sessionData.TotalTokens
		}
		// Look up context window from model registry.
		modelInfo := domain.ModelByAlias(a.ListModels(), model)
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

		// Record the rolling token-sample window for sparkline rendering (uiqa-09).
		// The helper throttles by absolute-delta so sparse updates don't collapse
		// the 20-sample window into a few seconds of noise.
		a.mu.Lock()
		samples := agent.MaybeAppendTokenSample(a.tokenSamples[agentID], tokensUsed)
		a.tokenSamples[agentID] = samples
		// Defensive copy so downstream readers can't mutate the cached slice.
		samplesCopy := append([]int(nil), samples...)
		a.mu.Unlock()

		ag := domain.Agent{
			ID:           agentID,
			Name:         model,
			Model:        model,
			Status:       status,
			PID:          s.PID,
			TokensUsed:   tokensUsed,
			TokenSamples: samplesCopy,
			TokensMax:    tokensMax,
			Elapsed:      time.Since(s.StartedAt),
			HasTmuxPane:  tmuxTarget != "",
			TmuxTarget:   tmuxTarget,
			RepoPath:     dir,
			LogLines:     logLines,
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
			delete(a.tokenSamples, n.AgentID)
		}
	}
	a.notifications = pruned
	a.mu.Unlock()

	// Emit repos to frontend
	a.emitEvent("repos", repos)
}
