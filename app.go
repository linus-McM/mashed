package main

import (
	"context"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"conductor/internal/agent"
	"conductor/internal/domain"
	"conductor/internal/explain"
	"conductor/internal/git"
	"conductor/internal/scanner"
	"conductor/internal/terminal"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the main application struct bound to the Wails frontend.
type App struct {
	ctx      context.Context
	cancel   context.CancelFunc
	provider    *scanner.ClaudeCodeProvider
	repoScanner *scanner.RepoScanner
	engine      *agent.NotificationEngine
	bridge      *terminal.Bridge
	panes       *terminal.PaneDiscovery
	explainer   *explain.Explainer
	mu          sync.Mutex

	devDir        string // root directory to scan for repos
	notifications []domain.NotificationEvent
}

// conductorConfig persists user settings between launches.
type conductorConfig struct {
	DevDir string `json:"devDir"`
}

// configPath returns the path to the conductor config file.
func configPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".conductor", "config.json")
}

// loadConfig reads the persisted config, or returns empty config.
func loadConfig() conductorConfig {
	data, err := os.ReadFile(configPath())
	if err != nil {
		return conductorConfig{}
	}
	var cfg conductorConfig
	json.Unmarshal(data, &cfg)
	return cfg
}

// saveConfig persists the config to disk.
func saveConfig(cfg conductorConfig) error {
	dir := filepath.Dir(configPath())
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	return os.WriteFile(configPath(), data, 0644)
}

// NewApp creates a new App instance.
func NewApp() *App {
	return &App{
		bridge: terminal.NewBridge(),
		panes:  terminal.NewPaneDiscovery(),
	}
}

// startup is called by Wails when the app starts.
func (a *App) startup(ctx context.Context) {
	a.ctx, a.cancel = context.WithCancel(ctx)

	// Engine needs Wails context for event emission
	a.engine = agent.NewNotificationEngine(a.ctx)

	// Start the terminal WebSocket bridge
	if err := a.bridge.Start(a.ctx); err != nil {
		log.Printf("terminal bridge start failed: %v", err)
	}

	// Initialize the diff explainer (uses ANTHROPIC_API_KEY from env)
	a.explainer = explain.New()

	// Check for saved config — if dir exists, start scanning immediately.
	// If not, frontend will detect empty GetDevDir() on mount and show setup.
	cfg := loadConfig()
	if cfg.DevDir != "" {
		a.initScanning(cfg.DevDir)
	}
}

// initScanning starts all background goroutines for a given dev directory.
func (a *App) initScanning(devDir string) {
	a.devDir = devDir

	provider, err := scanner.NewClaudeCodeProvider(devDir)
	if err != nil {
		log.Printf("failed to init claude provider: %v", err)
		return
	}
	a.provider = provider
	a.repoScanner = scanner.NewRepoScanner(devDir)

	go a.scanLoop()
	go a.watchSessions()
	go a.consumeEngineEvents()

	// Tell frontend setup is done
	runtime.EventsEmit(a.ctx, "needs-setup", false)
}

// shutdown is called by Wails when the app is closing.
func (a *App) shutdown(ctx context.Context) {
	if a.cancel != nil {
		a.cancel()
	}
	if a.bridge != nil {
		a.bridge.Stop()
	}
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

	// Track claimed session files so two agents in the same repo don't share one
	claimedSessions := make(map[string]bool) // sessionDir/sessionID -> true

	for _, s := range sessions {
		dir, err := a.provider.GetWorkingDir(s.PID)
		if err != nil {
			continue
		}

		agentID := fmt.Sprintf("pid-%d", s.PID)
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
		var tokensMax int64 = 200000 // default context window
		if sessionData != nil {
			tokensUsed = sessionData.TotalTokens
		}
		// Opus models have 1M context
		if model == "claude-opus-4-6" || model == "opus" {
			tokensMax = 1000000
		}

		// Determine status from session data
		status := domain.StatusRunning
		if sessionData != nil {
			status = a.inferStatus(sessionData)
		}

		// Look up tmux pane target — skip agents without a tmux session
		var tmuxTarget string
		if pane, err := a.panes.FindPaneForPID(s.PID); err == nil && pane != nil {
			tmuxTarget = pane.Target()
		}
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
	}

	// Emit repos to frontend
	runtime.EventsEmit(a.ctx, "repos", repos)
}

// watchSessions starts fsnotify-based session file watching.
func (a *App) watchSessions() {
	ch, err := a.provider.WatchSessions(a.ctx)
	if err != nil {
		log.Printf("session watcher error: %v", err)
		return
	}

	for {
		select {
		case <-a.ctx.Done():
			return
		case _, ok := <-ch:
			if !ok {
				return
			}
			// Trigger a scan on any session file change
			a.doScan()
		}
	}
}

// consumeEngineEvents reads from the engine's event channel and maintains the notification list.
func (a *App) consumeEngineEvents() {
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
func (a *App) findUnclaimed(sessionDir string, claimed map[string]bool) *domain.SessionData {
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

	// Return the first unclaimed file
	for _, f := range files {
		if claimed[f.key] {
			continue
		}
		data, err := a.provider.ParseSession(f.path)
		if err != nil {
			continue
		}
		claimed[f.key] = true
		return data
	}
	return nil
}

// findSessionByID parses a specific session file by its ID.
func (a *App) findSessionByID(sessionDir, sessionID string) *domain.SessionData {
	path := filepath.Join(sessionDir, sessionID+".jsonl")
	if _, err := os.Stat(path); err != nil {
		return nil
	}
	data, err := a.provider.ParseSession(path)
	if err != nil {
		return nil
	}
	return data
}

// findLatestSession finds and parses the most recently modified .jsonl file in a session directory.
func (a *App) findLatestSession(sessionDir string) *domain.SessionData {
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

	data, err := a.provider.ParseSession(latestPath)
	if err != nil {
		log.Printf("parse session %s: %v", latestPath, err)
		return nil
	}
	return data
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

// repoNameFromDir extracts the repo name from a directory path.
func repoNameFromDir(dir string) string {
	if dir == "" {
		return "unknown"
	}
	// Use last path component
	for i := len(dir) - 1; i >= 0; i-- {
		if dir[i] == '/' {
			return dir[i+1:]
		}
	}
	return dir
}

// --- Wails-bound methods (called from Svelte frontend) ---

// PickDirectory opens the native OS directory picker dialog and returns the selected path.
func (a *App) PickDirectory() (string, error) {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title:                "Choose Development Directory",
		CanCreateDirectories: true,
	})
	if err != nil {
		return "", fmt.Errorf("directory dialog: %w", err)
	}
	return dir, nil
}

// SetDevDir saves the chosen directory and starts scanning.
func (a *App) SetDevDir(dir string) error {
	if dir == "" {
		return fmt.Errorf("empty directory path")
	}
	// Verify it exists
	info, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("directory not accessible: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", dir)
	}

	// Persist
	if err := saveConfig(conductorConfig{DevDir: dir}); err != nil {
		log.Printf("failed to save config: %v", err)
	}

	a.initScanning(dir)
	return nil
}

// GetDevDir returns the current development directory.
func (a *App) GetDevDir() string {
	return a.devDir
}

// GetNotifications returns the current notification list sorted by priority.
func (a *App) GetNotifications() []domain.NotificationEvent {
	a.mu.Lock()
	defer a.mu.Unlock()
	result := make([]domain.NotificationEvent, len(a.notifications))
	copy(result, a.notifications)
	return result
}

// GetTerminalPort returns the WebSocket terminal bridge port.
func (a *App) GetTerminalPort() int {
	return a.bridge.GetTerminalPort()
}

// GetScopedDiff returns the changed files for a directory.
func (a *App) GetScopedDiff(dir string) (*domain.ScopedDiff, error) {
	return git.ScopedDiff(dir)
}

// GetWorktrees returns worktrees for a repo.
func (a *App) GetWorktrees(repoPath string) ([]domain.WorktreeInfo, error) {
	return git.DetectWorktrees(repoPath)
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

// SpawnAgent starts a new Claude session in a tmux pane for the given repo.
// Returns the tmux pane target string for the terminal bridge.
func (a *App) SpawnAgent(repoPath string, model string) (string, error) {
	if repoPath == "" {
		return "", fmt.Errorf("empty repo path")
	}
	if model == "" {
		model = "claude-opus-4-6"
	}

	// Derive a session name from the repo
	repoName := repoNameFromDir(repoPath)
	sessionName := fmt.Sprintf("conductor-%s-%d", repoName, time.Now().Unix())

	// Build the claude command
	cmd := fmt.Sprintf("claude --dangerously-skip-permissions --model %s", model)

	// Ensure tmux server is running, create session with claude inside it
	tmuxCmd := exec.CommandContext(a.ctx, "tmux", "new-session", "-d",
		"-s", sessionName,
		"-c", repoPath,
		cmd,
	)
	if out, err := tmuxCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("tmux new-session failed: %w (%s)", err, string(out))
	}

	// The pane target is sessionName:0.0 (first window, first pane)
	target := fmt.Sprintf("%s:0.0", sessionName)

	// Invalidate pane cache so discovery picks it up immediately
	a.panes.InvalidateCache()

	log.Printf("spawned agent in tmux session %s at %s", sessionName, repoPath)
	return target, nil
}

// SpawnAgentWithCommand starts a Claude session using a fully built CLI command.
// Returns the tmux pane target string.
func (a *App) SpawnAgentWithCommand(repoPath, command string) (string, error) {
	if repoPath == "" || command == "" {
		return "", fmt.Errorf("repo path and command are required")
	}

	repoName := repoNameFromDir(repoPath)
	sessionName := fmt.Sprintf("mashed-%s-%d", repoName, time.Now().Unix())

	tmuxCmd := exec.CommandContext(a.ctx, "tmux", "new-session", "-d",
		"-s", sessionName,
		"-c", repoPath,
		command,
	)
	if out, err := tmuxCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("tmux new-session failed: %w (%s)", err, string(out))
	}

	target := fmt.Sprintf("%s:0.0", sessionName)
	a.panes.InvalidateCache()

	log.Printf("spawned agent with command in tmux session %s at %s", sessionName, repoPath)
	return target, nil
}

// SpawnTerminal starts a plain shell tmux session in the given repo directory.
// Returns the tmux pane target string for the terminal bridge.
func (a *App) SpawnTerminal(repoPath string) (string, error) {
	if repoPath == "" {
		return "", fmt.Errorf("empty repo path")
	}

	repoName := repoNameFromDir(repoPath)
	sessionName := fmt.Sprintf("term-%s-%d", repoName, time.Now().Unix())

	tmuxCmd := exec.CommandContext(a.ctx, "tmux", "new-session", "-d",
		"-s", sessionName,
		"-c", repoPath,
	)
	if out, err := tmuxCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("tmux new-session failed: %w (%s)", err, string(out))
	}

	target := fmt.Sprintf("%s:0.0", sessionName)
	a.panes.InvalidateCache()

	log.Printf("spawned terminal in tmux session %s at %s", sessionName, repoPath)
	return target, nil
}

// ListRepoChoices returns the repos available for spawning agents.
func (a *App) ListRepoChoices() []map[string]string {
	if a.repoScanner == nil {
		return nil
	}
	repos, err := a.repoScanner.ScanRepos(nil)
	if err != nil {
		return nil
	}
	choices := make([]map[string]string, 0, len(repos))
	for _, r := range repos {
		choices = append(choices, map[string]string{
			"name":   r.Name,
			"path":   r.Path,
			"branch": r.Branch,
		})
	}
	return choices
}

// ReadFile returns the contents of a file as a string.
func (a *App) ReadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read file %s: %w", path, err)
	}
	// Cap at 1MB to avoid sending huge files to frontend
	if len(data) > 1024*1024 {
		return string(data[:1024*1024]) + "\n... (truncated at 1MB)", nil
	}
	return string(data), nil
}

// ReadFileDiff returns the git diff for a specific file.
func (a *App) ReadFileDiff(repoPath, filePath string) (string, error) {
	cmd := exec.CommandContext(a.ctx, "git", "-C", repoPath, "diff", "HEAD", "--", filePath)
	out, err := cmd.Output()
	if err != nil {
		// Try without HEAD for untracked files
		cmd2 := exec.CommandContext(a.ctx, "git", "-C", repoPath, "diff", "--no-index", "/dev/null", filepath.Join(repoPath, filePath))
		out2, _ := cmd2.Output()
		if len(out2) > 0 {
			return string(out2), nil
		}
		return "", fmt.Errorf("git diff %s: %w", filePath, err)
	}
	return string(out), nil
}

// ReadFileAtHead returns the content of a file at the HEAD commit.
func (a *App) ReadFileAtHead(repoPath, filePath string) (string, error) {
	cmd := exec.CommandContext(a.ctx, "git", "-C", repoPath, "show", "HEAD:"+filePath)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git show HEAD:%s: %w", filePath, err)
	}
	if len(out) > 1024*1024 {
		return string(out[:1024*1024]) + "\n... (truncated at 1MB)", nil
	}
	return string(out), nil
}

// KillAgent terminates an agent process and removes it from tracking.
func (a *App) KillAgent(agentID string, pid int) error {
	if pid > 0 {
		proc, err := os.FindProcess(pid)
		if err != nil {
			return fmt.Errorf("find process %d: %w", pid, err)
		}
		// Send SIGTERM for graceful shutdown
		if err := proc.Signal(os.Interrupt); err != nil {
			// Process may already be dead — try SIGKILL
			_ = proc.Kill()
		}
	}

	// Remove from engine tracking
	a.engine.RemoveAgent(agentID)

	// Remove from notification list
	a.mu.Lock()
	for i, n := range a.notifications {
		if n.AgentID == agentID {
			a.notifications = append(a.notifications[:i], a.notifications[i+1:]...)
			break
		}
	}
	a.mu.Unlock()

	// Emit updated list to frontend
	runtime.EventsEmit(a.ctx, "agent:removed", agentID)
	return nil
}

// MarkRead marks a notification as read.
func (a *App) MarkRead(agentID string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for i, n := range a.notifications {
		if n.AgentID == agentID {
			a.notifications[i].Read = true
			break
		}
	}
}

// GitListBranches returns all local branches for a repo, with the current branch marked.
func (a *App) GitListBranches(repoPath string) ([]map[string]interface{}, error) {
	cmd := exec.CommandContext(a.ctx, "git", "-C", repoPath, "branch", "--format=%(refname:short)\t%(HEAD)")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git branch: %w", err)
	}

	var branches []map[string]interface{}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		name := parts[0]
		current := len(parts) > 1 && strings.TrimSpace(parts[1]) == "*"
		branches = append(branches, map[string]interface{}{
			"name":    name,
			"current": current,
		})
	}
	return branches, nil
}

// GitSwitchBranch switches to an existing branch with optional auto-commit.
func (a *App) GitSwitchBranch(repoPath, branch string, autoCommit bool) error {
	if repoPath == "" || branch == "" {
		return fmt.Errorf("repo path and branch name are required")
	}

	if autoCommit {
		if _, err := a.GitCommit(repoPath); err != nil {
			if !strings.Contains(err.Error(), "nothing to commit") {
				return fmt.Errorf("auto-commit failed: %w", err)
			}
		}
	}

	cmd := exec.CommandContext(a.ctx, "git", "-C", repoPath, "checkout", branch)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git checkout: %w (%s)", err, string(out))
	}
	return nil
}

// GitCreateBranch creates a new branch with optional auto-commit of current changes.
// prefix is e.g. "feature", "hotfix"; name is the branch slug.
func (a *App) GitCreateBranch(repoPath, prefix, name string, autoCommit bool) error {
	if repoPath == "" || name == "" {
		return fmt.Errorf("repo path and branch name are required")
	}

	branchName := name
	if prefix != "" {
		branchName = prefix + "/" + name
	}

	// Auto-commit current changes if requested
	if autoCommit {
		if _, err := a.GitCommit(repoPath); err != nil {
			// Ignore "nothing to commit" — that's fine
			if !strings.Contains(err.Error(), "nothing to commit") {
				return fmt.Errorf("auto-commit failed: %w", err)
			}
		}
	}

	// Create and checkout the new branch
	cmd := exec.CommandContext(a.ctx, "git", "-C", repoPath, "checkout", "-b", branchName)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git checkout -b: %w (%s)", err, string(out))
	}

	return nil
}

// RepoStatus returns git dirty state and open PR count for a repo.
func (a *App) RepoStatus(repoPath string) map[string]interface{} {
	result := map[string]interface{}{
		"dirty":   false,
		"openPRs": 0,
	}

	// Check dirty (uncommitted changes including untracked files)
	statusCmd := exec.CommandContext(a.ctx, "git", "-C", repoPath, "status", "--porcelain")
	if out, err := statusCmd.Output(); err == nil && len(out) > 0 {
		result["dirty"] = true
	}

	// Check open PRs for the current branch
	branchCmd := exec.CommandContext(a.ctx, "git", "-C", repoPath, "rev-parse", "--abbrev-ref", "HEAD")
	branchOut, err := branchCmd.Output()
	if err == nil {
		branch := strings.TrimSpace(string(branchOut))
		ghCmd := exec.CommandContext(a.ctx, "gh", "pr", "list",
			"--state", "open",
			"--head", branch,
			"--json", "number",
			"--jq", "length",
		)
		ghCmd.Dir = repoPath
		if prOut, err := ghCmd.Output(); err == nil {
			count := strings.TrimSpace(string(prOut))
			if count != "" && count != "0" {
				n := 0
				fmt.Sscanf(count, "%d", &n)
				result["openPRs"] = n
			}
		}
	}

	return result
}

// GitCommit stages all changes, generates an AI commit message, and commits.
// Returns the commit message used.
func (a *App) GitCommit(repoPath string) (string, error) {
	// Stage all changes
	addCmd := exec.CommandContext(a.ctx, "git", "-C", repoPath, "add", "-A")
	if out, err := addCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("git add: %w (%s)", err, string(out))
	}

	// Check there's something to commit
	statusCmd := exec.CommandContext(a.ctx, "git", "-C", repoPath, "diff", "--cached", "--stat")
	statusOut, err := statusCmd.Output()
	if err != nil || len(statusOut) == 0 {
		return "", fmt.Errorf("nothing to commit")
	}

	// Get the diff for the AI to summarize
	diffCmd := exec.CommandContext(a.ctx, "git", "-C", repoPath, "diff", "--cached")
	diffOut, _ := diffCmd.Output()
	diffText := string(diffOut)
	if len(diffText) > 8000 {
		diffText = diffText[:8000] + "\n... (truncated)"
	}

	// Generate commit message using Claude CLI
	prompt := fmt.Sprintf("Write a concise git commit message (1-2 lines max, no quotes, no markdown) for this diff:\n\n%s", diffText)
	claudeCmd := exec.CommandContext(a.ctx, "claude", "-p", prompt)
	claudeCmd.Dir = repoPath
	msgOut, err := claudeCmd.Output()
	commitMsg := strings.TrimSpace(string(msgOut))
	if err != nil || commitMsg == "" {
		// Fallback: use the stat summary
		commitMsg = "update: " + strings.TrimSpace(string(statusOut))
		// Keep first line only
		if idx := strings.IndexByte(commitMsg, '\n'); idx > 0 {
			commitMsg = commitMsg[:idx]
		}
	}

	// Commit
	commitCmd := exec.CommandContext(a.ctx, "git", "-C", repoPath, "commit", "-m", commitMsg)
	if out, err := commitCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("git commit: %w (%s)", err, string(out))
	}

	return commitMsg, nil
}

// GitCommitAndPush commits (via GitCommit) then pushes to origin.
// Creates the remote branch if it doesn't exist.
func (a *App) GitCommitAndPush(repoPath string) (string, error) {
	msg, err := a.GitCommit(repoPath)
	if err != nil {
		return "", err
	}

	// Push with -u to set upstream, creating branch if needed
	pushCmd := exec.CommandContext(a.ctx, "git", "-C", repoPath, "push", "-u", "origin", "HEAD")
	if out, err := pushCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("git push: %w (%s)", err, string(out))
	}

	return msg, nil
}

// GitCommitPushAndPR commits, pushes, and creates a PR with an extensive description.
// Returns the PR URL.
func (a *App) GitCommitPushAndPR(repoPath string) (string, error) {
	_, err := a.GitCommitAndPush(repoPath)
	if err != nil {
		return "", err
	}

	// Get the current branch
	branchCmd := exec.CommandContext(a.ctx, "git", "-C", repoPath, "rev-parse", "--abbrev-ref", "HEAD")
	branchOut, err := branchCmd.Output()
	if err != nil {
		return "", fmt.Errorf("get branch: %w", err)
	}
	branch := strings.TrimSpace(string(branchOut))

	// Get the full diff against main/master for the PR body
	var baseBranch string
	for _, candidate := range []string{"main", "master"} {
		checkCmd := exec.CommandContext(a.ctx, "git", "-C", repoPath, "rev-parse", "--verify", candidate)
		if checkCmd.Run() == nil {
			baseBranch = candidate
			break
		}
	}
	if baseBranch == "" {
		baseBranch = "main"
	}

	diffCmd := exec.CommandContext(a.ctx, "git", "-C", repoPath, "diff", baseBranch+"...HEAD")
	diffOut, _ := diffCmd.Output()
	diffText := string(diffOut)
	if len(diffText) > 12000 {
		diffText = diffText[:12000] + "\n... (truncated)"
	}

	// Generate extensive PR description using Claude CLI
	prompt := fmt.Sprintf(`Write an extensive pull request description for this diff from branch "%s". Include:
- A clear title (first line, under 70 chars)
- ## Summary section with bullet points
- ## Changes section describing each file changed
- ## Test Plan section
Keep it factual based on the diff.

%s`, branch, diffText)
	claudeCmd := exec.CommandContext(a.ctx, "claude", "-p", prompt)
	claudeCmd.Dir = repoPath
	prOut, err := claudeCmd.Output()
	prBody := strings.TrimSpace(string(prOut))
	if err != nil || prBody == "" {
		prBody = fmt.Sprintf("Changes from branch %s", branch)
	}

	// Extract title (first line) and body (rest)
	prTitle := branch
	if idx := strings.IndexByte(prBody, '\n'); idx > 0 {
		prTitle = strings.TrimSpace(prBody[:idx])
		prBody = strings.TrimSpace(prBody[idx+1:])
		// Strip markdown heading prefix from title
		prTitle = strings.TrimLeft(prTitle, "# ")
	}
	if len(prTitle) > 70 {
		prTitle = prTitle[:67] + "..."
	}

	// Create PR using gh CLI
	ghCmd := exec.CommandContext(a.ctx, "gh", "pr", "create",
		"--title", prTitle,
		"--body", prBody,
		"--base", baseBranch,
	)
	ghCmd.Dir = repoPath
	ghOut, err := ghCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("gh pr create: %w (%s)", err, string(ghOut))
	}

	return strings.TrimSpace(string(ghOut)), nil
}

// ListRepoFiles returns all tracked (and untracked non-ignored) files in a repo.
func (a *App) ListRepoFiles(repoPath string) ([]string, error) {
	if repoPath == "" {
		return nil, fmt.Errorf("empty repo path")
	}
	// git ls-files returns tracked files; --others --exclude-standard adds untracked non-ignored
	cmd := exec.CommandContext(a.ctx, "git", "-C", repoPath, "ls-files", "--cached", "--others", "--exclude-standard")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git ls-files: %w", err)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var files []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			files = append(files, line)
		}
	}
	sort.Strings(files)
	return files, nil
}

// WriteFile writes content to a file on disk.
func (a *App) WriteFile(path, content string) error {
	if path == "" {
		return fmt.Errorf("empty file path")
	}
	return os.WriteFile(path, []byte(content), 0644)
}

// SpawnPRReview spawns a Claude agent to do an adversarial review of the latest PR.
// Returns the tmux pane target.
func (a *App) SpawnPRReview(repoPath string) (string, error) {
	// Find the latest PR number for this repo
	ghCmd := exec.CommandContext(a.ctx, "gh", "pr", "list", "--state", "open", "--limit", "1", "--json", "number", "--jq", ".[0].number")
	ghCmd.Dir = repoPath
	prOut, err := ghCmd.Output()
	if err != nil {
		return "", fmt.Errorf("no open PRs found: %w", err)
	}
	prNumber := strings.TrimSpace(string(prOut))
	if prNumber == "" {
		return "", fmt.Errorf("no open PRs found")
	}

	repoName := repoNameFromDir(repoPath)
	sessionName := fmt.Sprintf("review-%s-%d", repoName, time.Now().Unix())

	prompt := fmt.Sprintf(`You are an adversarial code reviewer. Review PR #%s in this repo thoroughly.
Look for: bugs, security vulnerabilities, race conditions, edge cases, performance issues,
missing error handling, breaking changes, and any code that could fail in production.
Be specific — cite file names and line numbers. Don't be nice, be thorough.
Start by running: gh pr diff %s`, prNumber, prNumber)

	cmd := fmt.Sprintf("claude --dangerously-skip-permissions --model claude-opus-4-6 -p %q", prompt)

	tmuxCmd := exec.CommandContext(a.ctx, "tmux", "new-session", "-d",
		"-s", sessionName,
		"-c", repoPath,
		cmd,
	)
	if out, err := tmuxCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("tmux new-session failed: %w (%s)", err, string(out))
	}

	target := fmt.Sprintf("%s:0.0", sessionName)
	a.panes.InvalidateCache()
	return target, nil
}

// ExplainDiffHunk returns an AI-generated explanation of why a diff hunk was changed.
func (a *App) ExplainDiffHunk(repoPath, filePath, hunkText string) (string, error) {
	return a.explainer.Explain(a.ctx, repoPath, filePath, hunkText)
}

// IsExplainAvailable returns true if the claude CLI is on PATH.
func (a *App) IsExplainAvailable() bool {
	_, err := exec.LookPath("claude")
	return err == nil
}
