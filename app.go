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
	"path/filepath"

	"conductor/internal/agent"
	"conductor/internal/domain"
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

		// Parse the latest session data for this agent
		sessionDir := a.provider.SessionDir(dir)
		sessionData := a.findLatestSession(sessionDir)

		var tokensUsed int64
		if sessionData != nil {
			tokensUsed = sessionData.TotalTokens
		}

		// Determine status from session data
		status := domain.StatusRunning
		if sessionData != nil {
			status = a.inferStatus(sessionData)
		}

		ag := domain.Agent{
			ID:         agentID,
			Name:       model,
			Model:      model,
			Status:     status,
			PID:        s.PID,
			TokensUsed: tokensUsed,
			Elapsed:    time.Since(s.StartedAt),
		}

		// Check for tmux pane
		if pane, err := a.panes.FindPaneForPID(s.PID); err == nil && pane != nil {
			ag.HasTmuxPane = true
		}

		// Feed to engine (only emits on state transitions)
		repoName := repoNameFromDir(dir)
		_ = a.engine.ProcessAgentUpdate(ag, repoName, "")
	}

	// Emit repos to frontend
	repos, _ := a.repoScanner.ScanRepos(nil)
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

// findLatestSession finds the most recently modified .jsonl in a session directory.
func (a *App) findLatestSession(dir string) *domain.SessionData {
	entries, err := a.provider.ParseSession(dir)
	if err != nil {
		return nil
	}
	return entries
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
