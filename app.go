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
	"sync/atomic"
	"time"

	"mashed/internal/agent"
	"mashed/internal/bmad"
	"mashed/internal/domain"
	"mashed/internal/explain"
	"mashed/internal/fsutil"
	"mashed/internal/scanner"
	"mashed/internal/terminal"
	"mashed/internal/terminal/helper"
	"mashed/internal/uiadapter"

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
	Spawn(ctx context.Context, name, repoPath, command string, cols, rows uint16) (*terminal.ManagedSession, error)
	SpawnArgv(ctx context.Context, name, repoPath string, argv []string, cols, rows uint16) (*terminal.ManagedSession, error)
	Kill(name string) error
	IsAlive(name string) bool
	FindByPID(pid int) (*terminal.ManagedSession, bool)
	Shutdown()
}

// App is the main application struct bound to the Wails frontend.
type App struct {
	ctx     context.Context
	cancel  context.CancelFunc
	engine  *agent.NotificationEngine
	bridge  *terminal.Bridge
	manager sessionManager
	// prNumber finds the PR SpawnPRReview reviews; nil means latestOpenPR.
	// Tests replace it to avoid calling gh.
	prNumber         func(ctx context.Context, repoPath string) (string, error)
	panes            paneDiscoverer
	explainer        *explain.Explainer
	mu               sync.Mutex
	activeRepoPath   string
	activePaneTarget string

	// scan holds devDir, provider and repoScanner as one immutable snapshot;
	// readers call scanSnapshot() once per operation (R17). Restarts are
	// serialised by scanMu; scanCancel/scanWG stop the previous goroutines.
	scan       atomic.Pointer[scanState]
	scanMu     sync.Mutex
	scanCancel context.CancelFunc
	scanWG     sync.WaitGroup
	// newProvider builds the session provider; nil means
	// scanner.NewClaudeCodeProvider. Tests inject failures.
	newProvider func(dir string) (*scanner.ClaudeCodeProvider, error)
	// Running goroutine counters (tests assert at most one of each).
	scanLoops, sessionWatchers, engineConsumers atomic.Int32

	notifications []domain.NotificationEvent
	// tokenSamples is a per-agent rolling window of token counts powering
	// the notification-feed sparkline (uiqa-09). Keyed by agentID. Reads and
	// writes are protected by a.mu (shared with notifications to avoid
	// introducing a second mutex).
	tokenSamples map[string][]int

	bmadStorage  *bmad.Storage
	bmadExecutor *bmad.Executor
	assetWatcher *bmad.AssetWatcher

	terminalSessions map[string]domain.TerminalSession
	logFile          *os.File

	// shutdownHooks are drained in (*App).shutdown. Each hook is invoked
	// exactly once; errors are log.Printf'd and never block subsequent
	// hooks. Story uiadapter-logging-1 introduced this slice to register
	// the production log-file closer.
	shutdownHooks []func() error
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

// MarkdownMenuSettings holds per-action visibility toggles for the markdown
// formatting toolbar. Every field is a bool so any combination is valid.
type MarkdownMenuSettings struct {
	Bold          bool `json:"bold"`
	Italic        bool `json:"italic"`
	Strikethrough bool `json:"strikethrough"`
	Code          bool `json:"code"`
	Link          bool `json:"link"`
	Latex         bool `json:"latex"`
}

// mashedConfig persists user settings between launches.
//
// OllamaEnabled / UIAdapterEnabled / UIAdapterUntrustedExpanded omit
// omitempty so explicit false round-trips to disk; loadConfig
// distinguishes missing from false where the default is TRUE.
type mashedConfig struct {
	DevDir                     string                `json:"devDir"`
	Theme                      string                `json:"theme,omitempty"`
	VSCodiumExtPath            string                `json:"vscodiumExtPath,omitempty"`
	ImportedTheme              string                `json:"importedTheme,omitempty"`
	MonoFont                   string                `json:"monoFont,omitempty"`
	FontSize                   int                   `json:"fontSize,omitempty"`
	SidebarWidth               int                   `json:"sidebarWidth,omitempty"`
	EditorSettings             *EditorSettings       `json:"editorSettings,omitempty"`
	MarkdownMenu               *MarkdownMenuSettings `json:"markdownMenu,omitempty"`
	OllamaEnabled              bool                  `json:"ollamaEnabled"`
	OllamaModel                string                `json:"ollamaModel,omitempty"`
	UIAdapterEnabled           bool                  `json:"uiAdapterEnabled"`
	UIAdapterTimeoutMs         int                   `json:"uiAdapterTimeoutMs,omitempty"`
	UIAdapterUntrustedExpanded bool                  `json:"uiAdapterUntrustedExpanded"`

	// Plan v3 Story 18 — dynamic backend/router selection (runtime pick).
	Backend      string `json:"backend,omitempty"`      // "ollama" | "claude-api" | "claude-cli"
	ClaudeModel  string `json:"claudeModel,omitempty"`  // for backend=claude-api
	CLIModel     string `json:"cliModel,omitempty"`     // for backend=claude-cli
	RouterPolicy string `json:"routerPolicy,omitempty"` // enum — see Plan §3 Story 16
}

const (
	defaultOllamaModel = "gemma3:4b"
	// 30s budget covers Ollama gemma3:4b cold-load (~10s) + generation
	// (~5s) on a typical Mac. The previous 3s budget guaranteed a fallback
	// AST on the first call after a server restart, hiding the structured
	// menu the adapter would otherwise have produced.
	defaultUIAdapterTimeoutMs = 30000
)

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

// loadConfig reads the persisted config, applying defaults for missing keys.
// Two-pass decode distinguishes missing key (default true) from explicit false.
func loadConfig() mashedConfig {
	cfg := defaultConfig()
	data, err := os.ReadFile(configPath())
	if err != nil {
		return cfg
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		log.Printf("warning: malformed config.json, ignoring: %v", err)
		return defaultConfig()
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Printf("warning: malformed config.json, ignoring: %v", err)
		return defaultConfig()
	}

	if _, ok := raw["ollamaEnabled"]; !ok {
		cfg.OllamaEnabled = true
	}
	if _, ok := raw["uiAdapterEnabled"]; !ok {
		cfg.UIAdapterEnabled = true
	}
	if !validOllamaModelName(cfg.OllamaModel) {
		if cfg.OllamaModel != "" {
			log.Printf("warning: config.json ollamaModel %q fails validation, falling back to default", cfg.OllamaModel)
		}
		cfg.OllamaModel = defaultOllamaModel
	}
	if cfg.UIAdapterTimeoutMs == 0 {
		cfg.UIAdapterTimeoutMs = defaultUIAdapterTimeoutMs
	}
	return cfg
}

func defaultConfig() mashedConfig {
	return mashedConfig{
		OllamaEnabled:      true,
		OllamaModel:        defaultOllamaModel,
		UIAdapterEnabled:   true,
		UIAdapterTimeoutMs: defaultUIAdapterTimeoutMs,
	}
}

// ensurePrivateDir creates dir if needed and makes it owner-only (R13).
func ensurePrivateDir(dir string) error {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	return os.Chmod(dir, 0o700)
}

// writePrivateFile writes data atomically as an owner-only file (R13).
func writePrivateFile(path string, data []byte) error {
	if err := ensurePrivateDir(filepath.Dir(path)); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}
	if err := fsutil.WriteFileAtomic(path, data, 0o600); err != nil {
		return err
	}
	return os.Chmod(path, 0o600)
}

// configRecoveredHook is told where a malformed config.json was moved.
// (*App).registerConfigRecovery sets it at startup.
var configRecoveredHook func(quarantinedPath string)

// registerConfigRecovery emits `config:recovered` whenever saveConfig
// quarantines a malformed config.json (R14).
func (a *App) registerConfigRecovery() {
	configRecoveredHook = func(p string) {
		a.emitEvent("config:recovered", map[string]string{"quarantinedPath": p})
	}
}

// quarantineMalformedConfig moves an existing config.json that does not
// parse as JSON to config.json.corrupt-<unix>, so the next write never
// replaces the user's bytes with defaults (R14). It returns the new path, or
// "" when there was nothing to quarantine.
func quarantineMalformedConfig() (string, error) {
	path := configPath()
	data, err := os.ReadFile(path)
	if err != nil || json.Valid(data) {
		return "", nil
	}
	dst := fmt.Sprintf("%s.corrupt-%d", path, time.Now().Unix())
	if err := os.Rename(path, dst); err != nil {
		return "", fmt.Errorf("quarantine malformed config: %w", err)
	}
	log.Printf("warning: malformed config.json moved to %s", dst)
	return dst, nil
}

// saveConfig persists the config to disk atomically, owner-only. A malformed
// config.json on disk is quarantined first instead of being overwritten.
func saveConfig(cfg mashedConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	quarantined, err := quarantineMalformedConfig()
	if err != nil {
		return err
	}
	if err := writePrivateFile(configPath(), data); err != nil {
		return err
	}
	if quarantined != "" && configRecoveredHook != nil {
		configRecoveredHook(quarantined)
	}
	return nil
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
		tokenSamples:     make(map[string][]int),
	}
}

// startup is called by Wails when the app starts.
func (a *App) startup(ctx context.Context) {
	a.ctx, a.cancel = context.WithCancel(ctx)
	a.registerConfigRecovery()

	a.initSessionLog()

	// Engine needs Wails context for event emission
	a.engine = agent.NewNotificationEngine(a.ctx)
	go a.consumeEngineEvents() // once per app, not per SetDevDir (R17)

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
		a.scan.Store(&scanState{devDir: cfg.DevDir})
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
		var bmadOpts []bmad.Option
		if cfg.UIAdapterEnabled {
			level := uiadapter.ParseLogLevel(os.Getenv("UIADAPTER_LOG_LEVEL"))
			adapterLogger, closer, logErr := uiadapter.NewProductionLogger(level, "./logs")
			if logErr != nil {
				log.Printf("uiadapter: production logger fallback to stdout-only: %v", logErr)
			}
			if closer != nil {
				a.shutdownHooks = append(a.shutdownHooks, closer.Close)
			}
			// Pick the adapter implementation based on cfg.Backend. The
			// "claude-cli" path delegates to the user's local `claude`
			// binary (no API key required) and forces Haiku to keep
			// per-translation cost negligible — Opus would cost ~30x for
			// a task that only needs JSON shaping.
			var adapter uiadapter.Adapter
			switch cfg.Backend {
			case "claude-cli":
				cliModel := cfg.CLIModel
				if cliModel == "" {
					cliModel = "claude-haiku-4-5"
				}
				adapterLogger.Info("uiadapter.boot",
					"op", "uiadapter.boot",
					"boot_level", level.String(),
					"backend", "claude-cli",
					"model", cliModel,
					"timeout_ms", cfg.UIAdapterTimeoutMs,
				)
				adapter = newClaudeCLIAdapter(uiadapter.Config{
					Enabled:            true,
					ClaudeModelPrimary: cliModel,
					TimeoutMs:          cfg.UIAdapterTimeoutMs,
					ClaudeCLIBinary:    "claude",
					// --verbose is mandatory when --output-format is
					// stream-json (claude refuses with exit 1 otherwise);
					// --model pins Haiku for cost.
					ClaudeCLIExtraFlags: []string{"--verbose", "--model", cliModel},
				}, adapterLogger)
			default:
				adapterLogger.Info("uiadapter.boot",
					"op", "uiadapter.boot",
					"boot_level", level.String(),
					"backend", "ollama",
					"model", cfg.OllamaModel,
					"timeout_ms", cfg.UIAdapterTimeoutMs,
				)
				adapter = uiadapter.NewDefault(uiadapter.Config{
					Enabled:     true,
					Model:       cfg.OllamaModel,
					TimeoutMs:   cfg.UIAdapterTimeoutMs,
					MaxInflight: 1,
				}, adapterLogger)
			}
			bmadOpts = append(bmadOpts, bmad.WithAdapter(adapter))
		}
		a.bmadExecutor = bmad.NewExecutor(storage, func(event string, data interface{}) {
			runtime.EventsEmit(a.ctx, event, data)
		}, bmadOpts...)

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

	// Start asset watcher for skills/commands directories.
	assetRoots := bmad.AssetWatchRoots(a.activeRepoPath)
	aw := bmad.NewAssetWatcher(assetRoots, func(event string, data interface{}) {
		runtime.EventsEmit(a.ctx, event, data)
	})
	if err := aw.Start(a.ctx); err != nil {
		log.Printf("bmad: asset watcher start failed (non-fatal): %v", err)
	} else {
		a.assetWatcher = aw
	}
}

// shutdown is called by Wails when the app is closing.
func (a *App) shutdown(ctx context.Context) {
	if a.assetWatcher != nil {
		a.assetWatcher.Stop()
	}
	if a.cancel != nil {
		a.cancel()
	}
	a.stopScanning()
	if a.manager != nil {
		a.manager.Shutdown()
	}
	if a.bridge != nil {
		a.bridge.Stop()
	}
	// Drain the registered shutdown hooks (e.g. uiadapter log-file
	// closer). Errors are logged but never block subsequent hooks per
	// Story uiadapter-logging-1 AC-1.6.
	for _, h := range a.shutdownHooks {
		if err := h(); err != nil {
			log.Printf("uiadapter: shutdown hook error: %v", err)
		}
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

	// Start scanning first: if the provider fails, the previous scanners and
	// the persisted DevDir stay as they were (R18).
	if err := a.initScanning(dir); err != nil {
		return err
	}

	// Persist (load-modify-save to preserve Theme/VSCodiumExtPath)
	a.mu.Lock()
	cfg := loadConfig()
	cfg.DevDir = dir
	if err := saveConfig(cfg); err != nil {
		log.Printf("failed to save config: %v", err)
	}
	a.mu.Unlock()
	return nil
}

// GetDevDir returns the current development directory.
func (a *App) GetDevDir() string {
	return a.scanSnapshot().devDir
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

// DefaultMarkdownMenuSettings returns the default visibility for each markdown
// toolbar action. LaTeX defaults to off; all other actions default to on.
func (a *App) DefaultMarkdownMenuSettings() MarkdownMenuSettings {
	return MarkdownMenuSettings{
		Bold:          true,
		Italic:        true,
		Strikethrough: true,
		Code:          true,
		Link:          true,
		Latex:         false,
	}
}

// GetMarkdownMenuSettings returns persisted markdown menu settings, or defaults
// if none saved. Reads must never mutate the on-disk config.
func (a *App) GetMarkdownMenuSettings() MarkdownMenuSettings {
	cfg := loadConfig()
	if cfg.MarkdownMenu == nil {
		return a.DefaultMarkdownMenuSettings()
	}
	return *cfg.MarkdownMenu
}

// SetMarkdownMenuSettings persists markdown menu settings to config. All fields
// are bool, so no validation is required.
func (a *App) SetMarkdownMenuSettings(settings MarkdownMenuSettings) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	cfg := loadConfig()
	cfg.MarkdownMenu = &settings
	if err := saveConfig(cfg); err != nil {
		return fmt.Errorf("save markdown menu settings: %w", err)
	}
	return nil
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

	data, err := json.MarshalIndent(all, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling themes: %w", err)
	}
	return writePrivateFile(themesPath(), data)
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
	return writePrivateFile(themesPath(), data)
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

// TerminalAuth is what the webview needs to open an authenticated terminal
// WebSocket: the bridge port and the per-launch token (R5).
type TerminalAuth struct {
	Port  int    `json:"port"`
	Token string `json:"token"`
}

// GetTerminalAuth returns the bridge port and connection token. The token is
// sent as the `mashed.auth.<token>` WebSocket subprotocol, never in a URL.
func (a *App) GetTerminalAuth() TerminalAuth {
	return TerminalAuth{Port: a.bridge.GetTerminalPort(), Token: a.bridge.Token()}
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
