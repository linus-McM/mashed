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

	"mashed/internal/agent"
	"mashed/internal/bmad"
	"mashed/internal/domain"
	"mashed/internal/explain"
	"mashed/internal/git"
	"mashed/internal/scanner"
	"mashed/internal/terminal"

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

	bmadStorage  *bmad.Storage
	bmadExecutor *bmad.Executor
}

// VSCodeThemeEntry represents a single color theme found in a VSCodium extension.
type VSCodeThemeEntry struct {
	Label       string `json:"label"`
	ExtensionID string `json:"extensionId"`
	ThemePath   string `json:"themePath"`
	UITheme     string `json:"uiTheme"`
}

// mashedConfig persists user settings between launches.
type mashedConfig struct {
	DevDir          string `json:"devDir"`
	Theme           string `json:"theme,omitempty"`
	VSCodiumExtPath string `json:"vscodiumExtPath,omitempty"`
	ImportedTheme   string `json:"importedTheme,omitempty"`
	MonoFont        string `json:"monoFont,omitempty"`
	FontSize        int    `json:"fontSize,omitempty"`
}

// configPath returns the path to the mashed config file.
func configPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".mashed", "config.json")
}

// themesPath returns the path to the saved themes file.
func themesPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".mashed", "themes.json")
}

// loadConfig reads the persisted config, or returns empty config.
func loadConfig() mashedConfig {
	data, err := os.ReadFile(configPath())
	if err != nil {
		return mashedConfig{}
	}
	var cfg mashedConfig
	json.Unmarshal(data, &cfg)
	return cfg
}

// saveConfig persists the config to disk.
func saveConfig(cfg mashedConfig) error {
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

	// Restore devDir from config so GetDevDir() works even if scanning fails.
	cfg := loadConfig()
	if cfg.DevDir != "" {
		a.devDir = cfg.DevDir
		if err := a.initScanning(cfg.DevDir); err != nil {
			log.Printf("scanning failed for %s: %v — app will show feed but may be empty", cfg.DevDir, err)
		}
	}

	// Initialize BMAD subsystem.
	home, _ := os.UserHomeDir()
	bmadDir := filepath.Join(home, ".mashed")
	storage, err := bmad.NewStorage(bmadDir)
	if err != nil {
		log.Printf("bmad storage init failed: %v", err)
	} else {
		a.bmadStorage = storage
		a.bmadExecutor = bmad.NewExecutor(storage, func(event string, data interface{}) {
			runtime.EventsEmit(a.ctx, event, data)
		})
	}
}

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

// sanitizeID removes characters that would break agent ID parsing.
func sanitizeID(s string) string {
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' {
			out = append(out, c)
		}
	}
	if len(out) == 0 {
		return "agent"
	}
	return string(out)
}

// repoNameFromDir extracts the repo name from a directory path.
func repoNameFromDir(dir string) string {
	if dir == "" {
		return "unknown"
	}
	return filepath.Base(dir)
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

	// Persist (load-modify-save to preserve Theme/VSCodiumExtPath)
	a.mu.Lock()
	cfg := loadConfig()
	cfg.DevDir = dir
	if err := saveConfig(cfg); err != nil {
		a.mu.Unlock()
		log.Printf("failed to save config: %v", err)
	} else {
		a.mu.Unlock()
	}

	return a.initScanning(dir)
}

// GetDevDir returns the current development directory.
func (a *App) GetDevDir() string {
	return a.devDir
}

// GetConfig returns the full persisted config for the frontend.
func (a *App) GetConfig() mashedConfig {
	return loadConfig()
}

// SetTheme persists the selected theme ID to config.
func (a *App) SetTheme(id string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	cfg := loadConfig()
	cfg.Theme = id
	return saveConfig(cfg)
}

// SetVSCodiumExtPath persists the VSCodium extension path to config.
func (a *App) SetVSCodiumExtPath(path string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	cfg := loadConfig()
	cfg.VSCodiumExtPath = path
	return saveConfig(cfg)
}

// SetMonoFont persists the selected mono font family to config.
func (a *App) SetMonoFont(fontFamily string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	cfg := loadConfig()
	cfg.MonoFont = fontFamily
	return saveConfig(cfg)
}

// SetFontSize persists the selected font size to config.
func (a *App) SetFontSize(size int) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	cfg := loadConfig()
	cfg.FontSize = size
	return saveConfig(cfg)
}

// GetSavedThemes returns all saved imported themes as a JSON string.
// The format is {"themeId": { label, css, monaco, xterm }, ...}.
func (a *App) GetSavedThemes() string {
	data, err := os.ReadFile(themesPath())
	if err != nil {
		return "{}"
	}
	return string(data)
}

// SaveTheme persists a converted theme to ~/.mashed/themes.json.
func (a *App) SaveTheme(id string, themeJSON string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Load existing themes
	all := make(map[string]json.RawMessage)
	if data, err := os.ReadFile(themesPath()); err == nil {
		json.Unmarshal(data, &all)
	}

	all[id] = json.RawMessage(themeJSON)

	dir := filepath.Dir(themesPath())
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating themes dir: %w", err)
	}
	data, err := json.MarshalIndent(all, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling themes: %w", err)
	}
	return os.WriteFile(themesPath(), data, 0644)
}

// RemoveTheme removes a saved theme from ~/.mashed/themes.json.
func (a *App) RemoveTheme(id string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	all := make(map[string]json.RawMessage)
	if data, err := os.ReadFile(themesPath()); err == nil {
		json.Unmarshal(data, &all)
	}

	delete(all, id)

	data, err := json.MarshalIndent(all, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling themes: %w", err)
	}
	return os.WriteFile(themesPath(), data, 0644)
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

// RepoMtimes returns the modification times of .git/index and the repo root directory.
// The frontend polls this cheaply to detect when a full refresh is needed.
func (a *App) RepoMtimes(repoPath string) (map[string]int64, error) {
	if repoPath == "" {
		return nil, fmt.Errorf("empty repo path")
	}
	result := map[string]int64{"index": 0, "root": 0}

	indexPath := filepath.Join(repoPath, ".git", "index")
	if fi, err := os.Stat(indexPath); err == nil {
		result["index"] = fi.ModTime().UnixMilli()
	}

	if fi, err := os.Stat(repoPath); err == nil {
		result["root"] = fi.ModTime().UnixMilli()
	}

	return result, nil
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
	sessionName := fmt.Sprintf("mashed-%s-%d", repoName, time.Now().Unix())

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

// RepoStatus returns git dirty state, open PR count, and ahead/behind counts for a repo.
func (a *App) RepoStatus(repoPath string) map[string]interface{} {
	result := map[string]interface{}{
		"dirty":     false,
		"openPRs":   0,
		"ahead":     0,
		"behind":    0,
		"protected": false,
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

		// Check ahead/behind remote tracking branch
		revCmd := exec.CommandContext(a.ctx, "git", "-C", repoPath,
			"rev-list", "--left-right", "--count", "HEAD...@{upstream}")
		if revOut, err := revCmd.Output(); err == nil {
			parts := strings.Fields(strings.TrimSpace(string(revOut)))
			if len(parts) == 2 {
				var ahead, behind int
				fmt.Sscanf(parts[0], "%d", &ahead)
				fmt.Sscanf(parts[1], "%d", &behind)
				result["ahead"] = ahead
				result["behind"] = behind
			}
		}

		// Check branch protection rules via gh API
		ghProtCmd := exec.CommandContext(a.ctx, "gh", "api",
			fmt.Sprintf("repos/{owner}/{repo}/branches/%s/protection", branch),
			"--jq", ".required_status_checks // empty",
		)
		ghProtCmd.Dir = repoPath
		if protOut, err := ghProtCmd.Output(); err == nil && len(strings.TrimSpace(string(protOut))) > 0 {
			result["protected"] = true
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

// GitCommitStreaming stages, generates an AI commit message, and commits,
// emitting progress events to the frontend at each step. On failure, calls
// Claude to explain what went wrong.
func (a *App) GitCommitStreaming(repoPath string) {
	emit := func(step, output, errMsg, explanation string, done bool) {
		runtime.EventsEmit(a.ctx, "git:commit:progress", map[string]interface{}{
			"repoPath":    repoPath,
			"step":        step,
			"output":      output,
			"error":       errMsg,
			"explanation":  explanation,
			"done":        done,
		})
	}

	handleErr := func(step, output string, err error) {
		fullErr := fmt.Sprintf("%s: %v", step, err)
		if output != "" {
			fullErr += "\n" + output
		}

		// Ask Claude to explain the failure
		explanation := ""
		prompt := fmt.Sprintf(
			"A git commit operation failed during the %q step. Explain this error concisely (2-3 sentences) and suggest a fix.\n\nError:\n%s\n\nOutput:\n%s",
			step, err.Error(), output,
		)
		claudeCmd := exec.CommandContext(a.ctx, "claude", "-p", prompt)
		claudeCmd.Dir = repoPath
		if expOut, expErr := claudeCmd.Output(); expErr == nil {
			explanation = strings.TrimSpace(string(expOut))
		}

		emit(step, output, fullErr, explanation, true)
	}

	go func() {
		// Step 1: Stage
		emit("Staging changes...", "", "", "", false)
		addCmd := exec.CommandContext(a.ctx, "git", "-C", repoPath, "add", "-A")
		if out, err := addCmd.CombinedOutput(); err != nil {
			handleErr("git add", string(out), err)
			return
		}
		emit("Staged all changes", "", "", "", false)

		// Step 2: Check for changes
		statusCmd := exec.CommandContext(a.ctx, "git", "-C", repoPath, "diff", "--cached", "--stat")
		statusOut, err := statusCmd.Output()
		if err != nil || len(statusOut) == 0 {
			emit("Nothing to commit", string(statusOut), "Nothing to commit — working tree clean", "", true)
			return
		}
		statText := strings.TrimSpace(string(statusOut))
		emit("Changes found", statText, "", "", false)

		// Step 3: Get diff for AI
		emit("Generating commit message...", "", "", "", false)
		diffCmd := exec.CommandContext(a.ctx, "git", "-C", repoPath, "diff", "--cached")
		diffOut, _ := diffCmd.Output()
		diffText := string(diffOut)
		if len(diffText) > 8000 {
			diffText = diffText[:8000] + "\n... (truncated)"
		}

		// Step 4: Generate commit message
		prompt := fmt.Sprintf("Write a concise git commit message (1-2 lines max, no quotes, no markdown) for this diff:\n\n%s", diffText)
		claudeCmd := exec.CommandContext(a.ctx, "claude", "-p", prompt)
		claudeCmd.Dir = repoPath
		msgOut, err := claudeCmd.Output()
		commitMsg := strings.TrimSpace(string(msgOut))
		if err != nil || commitMsg == "" {
			commitMsg = "update: " + strings.TrimSpace(string(statusOut))
			if idx := strings.IndexByte(commitMsg, '\n'); idx > 0 {
				commitMsg = commitMsg[:idx]
			}
			emit("Using fallback commit message", commitMsg, "", "", false)
		} else {
			emit("Commit message ready", commitMsg, "", "", false)
		}

		// Step 5: Commit
		emit("Committing...", "", "", "", false)
		commitCmd := exec.CommandContext(a.ctx, "git", "-C", repoPath, "commit", "-m", commitMsg)
		if out, err := commitCmd.CombinedOutput(); err != nil {
			handleErr("git commit", string(out), err)
			return
		}

		emit("Committed", commitMsg, "", "", true)
	}()
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

// GitPush pushes the current branch to origin without committing first.
// Returns a structured result: "ok" on success, or "conflict:<message>" when
// the push is rejected due to diverged history (non-fast-forward).
func (a *App) GitPush(repoPath string) (string, error) {
	if repoPath == "" {
		return "", fmt.Errorf("repo path is required")
	}

	pushCmd := exec.CommandContext(a.ctx, "git", "-C", repoPath, "push", "-u", "origin", "HEAD")
	out, err := pushCmd.CombinedOutput()
	if err != nil {
		outStr := string(out)
		// Detect non-fast-forward (diverged history) vs other errors
		if strings.Contains(outStr, "non-fast-forward") ||
			strings.Contains(outStr, "rejected") ||
			strings.Contains(outStr, "fetch first") {
			return "conflict:" + strings.TrimSpace(outStr), nil
		}
		return "", fmt.Errorf("git push: %w (%s)", err, outStr)
	}
	return "ok", nil
}

// GitForcePush force-pushes the current branch to origin with --force-with-lease
// for safety (fails if someone else pushed since your last fetch).
func (a *App) GitForcePush(repoPath string) (string, error) {
	if repoPath == "" {
		return "", fmt.Errorf("repo path is required")
	}

	pushCmd := exec.CommandContext(a.ctx, "git", "-C", repoPath, "push", "--force-with-lease", "-u", "origin", "HEAD")
	out, err := pushCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git force push: %w (%s)", err, string(out))
	}
	return strings.TrimSpace(string(out)), nil
}

// GitPull pulls remote changes into the current branch.
func (a *App) GitPull(repoPath string) (string, error) {
	if repoPath == "" {
		return "", fmt.Errorf("repo path is required")
	}
	cmd := exec.CommandContext(a.ctx, "git", "-C", repoPath, "pull")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git pull: %w (%s)", err, string(out))
	}
	return strings.TrimSpace(string(out)), nil
}

// GitMergeInto merges the current branch into targetBranch.
// If autoCommit is true, commits current changes before merging.
// On merge failure, aborts the merge and checks out the original branch.
func (a *App) GitMergeInto(repoPath, targetBranch string, autoCommit bool) (string, error) {
	if repoPath == "" || targetBranch == "" {
		return "", fmt.Errorf("repo path and target branch are required")
	}

	// Get current branch name
	branchCmd := exec.CommandContext(a.ctx, "git", "-C", repoPath, "rev-parse", "--abbrev-ref", "HEAD")
	branchOut, err := branchCmd.Output()
	if err != nil {
		return "", fmt.Errorf("get current branch: %w", err)
	}
	sourceBranch := strings.TrimSpace(string(branchOut))

	if sourceBranch == targetBranch {
		return "", fmt.Errorf("already on %s — nothing to merge", targetBranch)
	}

	// Auto-commit current changes if requested
	if autoCommit {
		if _, err := a.GitCommit(repoPath); err != nil {
			if !strings.Contains(err.Error(), "nothing to commit") {
				return "", fmt.Errorf("auto-commit failed: %w", err)
			}
		}
	}

	// Switch to target branch
	checkoutCmd := exec.CommandContext(a.ctx, "git", "-C", repoPath, "checkout", targetBranch)
	if out, err := checkoutCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("checkout %s: %w (%s)", targetBranch, err, string(out))
	}

	// Merge source into target
	mergeCmd := exec.CommandContext(a.ctx, "git", "-C", repoPath, "merge", sourceBranch)
	mergeOut, mergeErr := mergeCmd.CombinedOutput()
	if mergeErr != nil {
		// Abort the failed merge and return to the original branch
		_ = exec.CommandContext(a.ctx, "git", "-C", repoPath, "merge", "--abort").Run()
		_ = exec.CommandContext(a.ctx, "git", "-C", repoPath, "checkout", sourceBranch).Run()
		return "", fmt.Errorf("merge %s into %s failed: %w (%s)", sourceBranch, targetBranch, mergeErr, string(mergeOut))
	}

	return fmt.Sprintf("Merged %s into %s", sourceBranch, targetBranch), nil
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

// ── BMAD Workflow CRUD ──

// ListBmadWorkflows returns all user-saved workflows.
func (a *App) ListBmadWorkflows() ([]bmad.WorkflowDef, error) {
	if a.bmadStorage == nil {
		return nil, fmt.Errorf("bmad storage not initialized")
	}
	return a.bmadStorage.ListWorkflows()
}

// GetBmadWorkflow loads a single workflow by ID.
func (a *App) GetBmadWorkflow(id string) (bmad.WorkflowDef, error) {
	if a.bmadStorage == nil {
		return bmad.WorkflowDef{}, fmt.Errorf("bmad storage not initialized")
	}
	return a.bmadStorage.LoadWorkflow(id)
}

// SaveBmadWorkflow persists a workflow definition.
func (a *App) SaveBmadWorkflow(wf bmad.WorkflowDef) error {
	if a.bmadStorage == nil {
		return fmt.Errorf("bmad storage not initialized")
	}
	return a.bmadStorage.SaveWorkflow(wf)
}

// DeleteBmadWorkflow removes a workflow by ID.
func (a *App) DeleteBmadWorkflow(id string) error {
	if a.bmadStorage == nil {
		return fmt.Errorf("bmad storage not initialized")
	}
	return a.bmadStorage.DeleteWorkflow(id)
}

// ── BMAD Process Registry ──

// GetBmadProcesses returns the full process catalog.
func (a *App) GetBmadProcesses() []bmad.ProcessDef {
	return bmad.AllProcesses()
}

// GetBmadProcessesByPhase returns processes filtered by lifecycle phase.
func (a *App) GetBmadProcessesByPhase(phase string) []bmad.ProcessDef {
	return bmad.ProcessesByPhase(bmad.BmadPhase(phase))
}

// ListBmadWorkflowsByRepo returns workflows scoped to a specific repository path.
func (a *App) ListBmadWorkflowsByRepo(repoPath string) ([]bmad.WorkflowDef, error) {
	if a.bmadStorage == nil {
		return nil, fmt.Errorf("bmad storage not initialized")
	}
	return a.bmadStorage.ListWorkflowsByRepo(repoPath)
}

// ── BMAD Templates ──

// ListBmadTemplates returns the 6 built-in workflow templates.
func (a *App) ListBmadTemplates() []bmad.WorkflowDef {
	return bmad.BuiltinTemplates()
}

// CreateFromTemplate deep-copies a built-in template into a user workflow.
func (a *App) CreateFromTemplate(templateID, repoPath string) (bmad.WorkflowDef, error) {
	if a.bmadStorage == nil {
		return bmad.WorkflowDef{}, fmt.Errorf("bmad storage not initialized")
	}

	var tpl bmad.WorkflowDef
	var found bool
	for _, t := range bmad.BuiltinTemplates() {
		if t.ID == templateID {
			tpl = t
			found = true
			break
		}
	}
	if !found {
		return bmad.WorkflowDef{}, fmt.Errorf("template %q not found", templateID)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	wf := bmad.WorkflowDef{
		ID:          fmt.Sprintf("wf-%s-%d", templateID, time.Now().UnixMilli()),
		Name:        "Copy of " + tpl.Name,
		Description: tpl.Description,
		Nodes:       make([]bmad.WorkflowNode, len(tpl.Nodes)),
		Edges:       make([]bmad.WorkflowEdge, len(tpl.Edges)),
		IsTemplate:  false,
		TemplateID:  templateID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	copy(wf.Nodes, tpl.Nodes)
	copy(wf.Edges, tpl.Edges)
	wf.RepoPath = repoPath

	if err := a.bmadStorage.SaveWorkflow(wf); err != nil {
		return bmad.WorkflowDef{}, fmt.Errorf("saving workflow from template: %w", err)
	}
	return wf, nil
}

// ── BMAD Sprint Status ──

// GetSprintStatus reads and parses the sprint-status.yaml for the given repo.
func (a *App) GetSprintStatus(repoPath string) (bmad.SprintStatus, error) {
	return bmad.ParseSprintStatus(repoPath)
}

// UpdateStoryStatus modifies a story's status in sprint-status.yaml.
func (a *App) UpdateStoryStatus(repoPath, storyID, newStatus string) error {
	return bmad.UpdateStoryStatus(repoPath, storyID, newStatus)
}

// ── BMAD Execution ──

// StartBmadWorkflow begins executing a workflow and returns the execution ID.
func (a *App) StartBmadWorkflow(workflowID, repoPath, model string) (string, error) {
	if a.bmadExecutor == nil {
		return "", fmt.Errorf("bmad executor not initialized")
	}
	exec, err := a.bmadExecutor.StartWorkflow(workflowID, repoPath, model)
	if err != nil {
		return "", err
	}
	return exec.ID, nil
}

// PauseBmadWorkflow pauses a running execution.
func (a *App) PauseBmadWorkflow(execID string) error {
	if a.bmadExecutor == nil {
		return fmt.Errorf("bmad executor not initialized")
	}
	return a.bmadExecutor.PauseWorkflow(execID)
}

// ResumeBmadWorkflow resumes a paused execution.
func (a *App) ResumeBmadWorkflow(execID string) error {
	if a.bmadExecutor == nil {
		return fmt.Errorf("bmad executor not initialized")
	}
	return a.bmadExecutor.ResumeWorkflow(execID)
}

// StopBmadWorkflow cancels a running execution.
func (a *App) StopBmadWorkflow(execID string) error {
	if a.bmadExecutor == nil {
		return fmt.Errorf("bmad executor not initialized")
	}
	return a.bmadExecutor.StopWorkflow(execID)
}

// GetBmadExecution returns the current state of an execution.
func (a *App) GetBmadExecution(execID string) (*bmad.WorkflowExecution, error) {
	if a.bmadExecutor == nil {
		return nil, fmt.Errorf("bmad executor not initialized")
	}
	return a.bmadExecutor.GetExecution(execID)
}

// ── BMAD Agent Management ──

// ListBmadAgents returns all custom agent configurations.
func (a *App) ListBmadAgents() ([]bmad.BmadAgentConfig, error) {
	if a.bmadStorage == nil {
		return nil, fmt.Errorf("bmad storage not initialized")
	}
	return a.bmadStorage.ListAgents()
}

// SaveBmadAgent persists a custom agent configuration.
func (a *App) SaveBmadAgent(agent bmad.BmadAgentConfig) error {
	if a.bmadStorage == nil {
		return fmt.Errorf("bmad storage not initialized")
	}
	return a.bmadStorage.SaveAgent(agent)
}

// DeleteBmadAgent removes a custom agent by ID.
func (a *App) DeleteBmadAgent(id string) error {
	if a.bmadStorage == nil {
		return fmt.Errorf("bmad storage not initialized")
	}
	return a.bmadStorage.DeleteAgent(id)
}

// ── BMAD Modules ──

// GetBmadModules returns all available BMAD modules.
func (a *App) GetBmadModules() []bmad.ModuleDef {
	return bmad.GetModules()
}
