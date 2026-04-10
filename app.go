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
	"strings"
	"sync"
	"time"

	"mashed/internal/agent"
	"mashed/internal/bmad"
	"mashed/internal/domain"
	"mashed/internal/explain"
	"mashed/internal/scanner"
	"mashed/internal/terminal"
	"mashed/internal/terminal/helper"

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
	FindByPID(pid int) (*terminal.ManagedSession, bool)
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
	mu               sync.Mutex
	activeRepoPath   string
	activePaneTarget string

	devDir        string // root directory to scan for repos
	notifications []domain.NotificationEvent

	bmadStorage  *bmad.Storage
	bmadExecutor *bmad.Executor

	terminalSessions map[string]domain.TerminalSession
	logFile          *os.File
}

// VSCodeThemeEntry represents a single color theme found in a VSCodium extension.
type VSCodeThemeEntry struct {
	Label       string `json:"label"`
	ExtensionID string `json:"extensionId"`
	ThemePath   string `json:"themePath"`
	UITheme     string `json:"uiTheme"`
}

// EditorSettings holds Monaco editor configuration options.
type EditorSettings struct {
	MinimapEnabled          bool   `json:"minimapEnabled"`
	WordWrap                string `json:"wordWrap"`
	LineNumbers             string `json:"lineNumbers"`
	RenderWhitespace        string `json:"renderWhitespace"`
	TabSize                 int    `json:"tabSize"`
	InsertSpaces            bool   `json:"insertSpaces"`
	CursorStyle             string `json:"cursorStyle"`
	CursorBlinking          string `json:"cursorBlinking"`
	BracketPairColorization bool   `json:"bracketPairColorization"`
	RenderLineHighlight     string `json:"renderLineHighlight"`
	FontLigatures           bool   `json:"fontLigatures"`
	ScrollBeyondLastLine    bool   `json:"scrollBeyondLastLine"`
	SmoothScrolling         bool   `json:"smoothScrolling"`
}

// mashedConfig persists user settings between launches.
type mashedConfig struct {
	DevDir          string          `json:"devDir"`
	Theme           string          `json:"theme,omitempty"`
	VSCodiumExtPath string          `json:"vscodiumExtPath,omitempty"`
	ImportedTheme   string          `json:"importedTheme,omitempty"`
	MonoFont        string          `json:"monoFont,omitempty"`
	FontSize        int             `json:"fontSize,omitempty"`
	SidebarWidth    int             `json:"sidebarWidth,omitempty"`
	EditorSettings  *EditorSettings `json:"editorSettings,omitempty"`
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

// NewApp creates a new App instance. The helperClient may be nil; Spawn will
// return terminal.ErrHelperNotRunning until a client is provided.
//
// The tmux adapter is wired only when the tmux binary is on PATH; otherwise
// the bridge gets a nil adapter and BMAD-prefixed WebSocket requests fall
// through to a 404 instead of attempting an upgrade — PTY shells keep
// working unchanged.
func NewApp(helperClient *helper.Client) *App {
	sm := terminal.NewSessionManager(helperClient)

	var tmuxAdapter terminal.TmuxAttacher
	if terminal.IsTmuxAvailable() {
		tmuxAdapter = terminal.NewTmuxAdapter(nil)
	} else {
		log.Printf("app: tmux not on PATH; BMAD terminal streaming disabled")
	}

	return &App{
		bridge:           terminal.NewBridge(sm, tmuxAdapter),
		manager:          sm,
		panes:            terminal.NewPaneDiscovery(),
		terminalSessions: make(map[string]domain.TerminalSession),
	}
}

// startup is called by Wails when the app starts.
func (a *App) startup(ctx context.Context) {
	a.ctx, a.cancel = context.WithCancel(ctx)

	a.initSessionLog()

	// Engine needs Wails context for event emission
	a.engine = agent.NewNotificationEngine(a.ctx)

	// Start the terminal WebSocket bridge
	if err := a.bridge.Start(a.ctx); err != nil {
		log.Printf("terminal bridge start failed: %v", err)
	}

	// PTY sessions don't survive restart — no recovery needed.

	// Initialize the diff explainer (uses ANTHROPIC_API_KEY from env)
	a.explainer = explain.New()

	// Discover available models from Claude CLI (async, non-blocking).
	go a.initModelCache()

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

		// Clean up any BMAD tmux sessions left over from prior runs.
		// Executions map is empty here (no workflows can have started yet),
		// so any bmad-prefixed session is unambiguously an orphan.
		// Bounded by a 5s timeout so a hung tmux daemon can't block startup.
		cleanupCtx, cleanupCancel := context.WithTimeout(a.ctx, 5*time.Second)
		if err := a.bmadExecutor.CleanupStaleSessions(cleanupCtx); err != nil {
			log.Printf("app: bmad session cleanup error (non-fatal): %v", err)
		}
		cleanupCancel()
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

// PickFile opens the native OS file picker dialog and returns the selected file path.
func (a *App) PickFile(title string) (string, error) {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: title,
	})
	if err != nil {
		return "", fmt.Errorf("file dialog: %w", err)
	}
	return path, nil
}

// SetActiveContext stores the current repo path and pane target for screenshot routing.
func (a *App) SetActiveContext(repoPath, paneTarget string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.activeRepoPath = repoPath
	a.activePaneTarget = paneTarget
}

// TakeScreenshot launches macOS screencapture and saves under {repoPath}/.screenshots/.
func (a *App) TakeScreenshot(repoPath, paneTarget string) (string, error) {
	if repoPath == "" {
		return "", fmt.Errorf("empty repoPath")
	}

	screenshotDir := filepath.Join(repoPath, ".screenshots")
	if err := os.MkdirAll(screenshotDir, 0755); err != nil {
		return "", fmt.Errorf("screencapture: %w", err)
	}

	if err := ensureGitignoreEntry(repoPath, ".screenshots/"); err != nil {
		return "", fmt.Errorf("screencapture: %w", err)
	}

	filename := fmt.Sprintf("screenshot-%s.png", time.Now().Format("20060102-150405"))
	path := filepath.Join(screenshotDir, filename)

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
		runtime.EventsEmit(a.ctx, "screenshot:inject", map[string]string{"path": path, "paneTarget": paneTarget})
		runtime.EventsEmit(a.ctx, "screenshot:taken", path)
	}
	return path, nil
}

// ensureGitignoreEntry appends entry to .gitignore if not already present.
func ensureGitignoreEntry(repoPath, entry string) error {
	gitignorePath := filepath.Join(repoPath, ".gitignore")
	content, err := os.ReadFile(gitignorePath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("read .gitignore: %w", err)
	}
	if strings.Contains(string(content), entry) {
		return nil
	}
	f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open .gitignore: %w", err)
	}
	defer f.Close()
	if len(content) > 0 && !strings.HasSuffix(string(content), "\n") {
		fmt.Fprintln(f)
	}
	fmt.Fprintf(f, "\n# mashed screenshots\n%s\n", entry)
	return nil
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

// SetSidebarWidth persists the sidebar width to config.
func (a *App) SetSidebarWidth(width int) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	cfg := loadConfig()
	cfg.SidebarWidth = width
	return saveConfig(cfg)
}

// DefaultEditorSettings returns sensible defaults for all editor options.
func (a *App) DefaultEditorSettings() EditorSettings {
	return EditorSettings{
		MinimapEnabled:          false,
		WordWrap:                "off",
		LineNumbers:             "on",
		RenderWhitespace:        "none",
		TabSize:                 2,
		InsertSpaces:            true,
		CursorStyle:             "line",
		CursorBlinking:          "blink",
		BracketPairColorization: true,
		RenderLineHighlight:     "line",
		FontLigatures:           false,
		ScrollBeyondLastLine:    false,
		SmoothScrolling:         false,
	}
}

// GetEditorSettings returns persisted editor settings, or defaults if none saved.
func (a *App) GetEditorSettings() EditorSettings {
	cfg := loadConfig()
	if cfg.EditorSettings == nil {
		return a.DefaultEditorSettings()
	}
	return *cfg.EditorSettings
}

// validateEditorSettings checks that all enum and range fields are valid.
func validateEditorSettings(es EditorSettings) error {
	if es.TabSize < 2 || es.TabSize > 8 {
		return fmt.Errorf("tabSize must be between 2 and 8, got %d", es.TabSize)
	}

	validWordWrap := map[string]bool{"off": true, "on": true, "wordWrapColumn": true, "bounded": true}
	if !validWordWrap[es.WordWrap] {
		return fmt.Errorf("wordWrap must be one of off, on, wordWrapColumn, bounded; got %q", es.WordWrap)
	}

	validLineNumbers := map[string]bool{"on": true, "off": true, "relative": true, "interval": true}
	if !validLineNumbers[es.LineNumbers] {
		return fmt.Errorf("lineNumbers must be one of on, off, relative, interval; got %q", es.LineNumbers)
	}

	validCursorStyle := map[string]bool{"line": true, "block": true, "underline": true, "line-thin": true, "block-outline": true, "underline-thin": true}
	if !validCursorStyle[es.CursorStyle] {
		return fmt.Errorf("cursorStyle must be one of line, block, underline, line-thin, block-outline, underline-thin; got %q", es.CursorStyle)
	}

	validCursorBlinking := map[string]bool{"blink": true, "smooth": true, "phase": true, "expand": true, "solid": true}
	if !validCursorBlinking[es.CursorBlinking] {
		return fmt.Errorf("cursorBlinking must be one of blink, smooth, phase, expand, solid; got %q", es.CursorBlinking)
	}

	validRenderWhitespace := map[string]bool{"none": true, "boundary": true, "selection": true, "trailing": true, "all": true}
	if !validRenderWhitespace[es.RenderWhitespace] {
		return fmt.Errorf("renderWhitespace must be one of none, boundary, selection, trailing, all; got %q", es.RenderWhitespace)
	}

	validRenderLineHighlight := map[string]bool{"none": true, "gutter": true, "line": true, "all": true}
	if !validRenderLineHighlight[es.RenderLineHighlight] {
		return fmt.Errorf("renderLineHighlight must be one of none, gutter, line, all; got %q", es.RenderLineHighlight)
	}

	return nil
}

// SetEditorSettings validates and persists editor settings to config.
func (a *App) SetEditorSettings(settings EditorSettings) error {
	if err := validateEditorSettings(settings); err != nil {
		return err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	cfg := loadConfig()
	cfg.EditorSettings = &settings
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

// initSessionLog creates a timestamped log file in .logs/ for this session.
func (a *App) initSessionLog() {
	dir := filepath.Join(".", ".logs")
	os.MkdirAll(dir, 0755)
	name := fmt.Sprintf("session-%s.log", time.Now().Format("2006-01-02T15-04-05"))
	f, err := os.Create(filepath.Join(dir, name))
	if err != nil {
		log.Printf("WARNING: could not create session log: %v", err)
		return
	}
	a.logFile = f
	fmt.Fprintf(f, "=== mashed session started %s ===\n", time.Now().Format(time.RFC3339))
}

// WriteConsoleLog receives a frontend console message and appends it to the session log.
func (a *App) WriteConsoleLog(level, message string) {
	if a.logFile == nil {
		return
	}
	fmt.Fprintf(a.logFile, "[%s] %s %s\n", time.Now().Format("15:04:05.000"), level, message)
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
