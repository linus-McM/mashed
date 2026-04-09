package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"mashed/internal/agent"
	"mashed/internal/bmad"
	"mashed/internal/domain"
	"mashed/internal/explain"
	"mashed/internal/scanner"
	"mashed/internal/terminal"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// paneDiscoverer abstracts tmux pane discovery for testability.
type paneDiscoverer interface {
	ListPanes() ([]terminal.TmuxPane, error)
	InvalidateCache()
	FindPaneForPID(agentPID int) (*terminal.TmuxPane, error)
}

// sessionManager abstracts PTY session lifecycle for testability.
type sessionManager interface {
	Spawn(ctx context.Context, name, repoPath, command string) (*terminal.ManagedSession, error)
	Kill(name string) error
	IsAlive(name string) bool
	Shutdown()
}

// App is the main application struct bound to the Wails frontend.
type App struct {
	ctx         context.Context
	cancel      context.CancelFunc
	provider    *scanner.ClaudeCodeProvider
	repoScanner *scanner.RepoScanner
	engine      *agent.NotificationEngine
	bridge      *terminal.Bridge
	manager     sessionManager
	panes       paneDiscoverer
	explainer   *explain.Explainer
	mu          sync.Mutex

	devDir        string // root directory to scan for repos
	notifications []domain.NotificationEvent

	bmadStorage  *bmad.Storage
	bmadExecutor *bmad.Executor

	terminalSessions map[string]domain.TerminalSession
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
	sm := terminal.NewSessionManager()
	return &App{
		bridge:           terminal.NewBridge(sm),
		manager:          sm,
		panes:            terminal.NewPaneDiscovery(),
		terminalSessions: make(map[string]domain.TerminalSession),
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

	// PTY sessions don't survive restart — no recovery needed.

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

// shutdown is called by Wails when the app is closing.
func (a *App) shutdown(ctx context.Context) {
	if a.cancel != nil {
		a.cancel()
	}
	if a.manager != nil {
		a.manager.Shutdown()
	}
	if a.bridge != nil {
		a.bridge.Stop()
	}
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

// TakeScreenshot launches macOS screencapture and returns the saved file path.
func (a *App) TakeScreenshot() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("screencapture: %w", err)
	}

	filename := fmt.Sprintf("mashed-screenshot-%s.png", time.Now().Format("20060102-150405"))
	path := filepath.Join(home, "Desktop", filename)

	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	cmd := exec.CommandContext(ctx, "screencapture", "-i", "-x", path)
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return "", nil
		}
		return "", fmt.Errorf("screencapture: %w", err)
	}

	if a.cancel != nil {
		runtime.EventsEmit(a.ctx, "screenshot:taken", path)
	}
	return path, nil
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
