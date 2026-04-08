package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"mashed/internal/domain"
)

// Compile-time interface check.
var _ domain.AgentProvider = (*ClaudeCodeProvider)(nil)

// ClaudeCodeProvider implements AgentProvider for Claude Code CLI sessions.
type ClaudeCodeProvider struct {
	devDir    string
	homeDir   string
	claudeDir string // ~/.claude/projects/

	// pid -> working dir cache
	pidDirMu    sync.RWMutex
	pidDirCache map[int]pidDirEntry

	// incremental parser state per session file
	parserMu sync.Mutex
	parsers  map[string]*sessionParserState
}

type pidDirEntry struct {
	dir      string
	cachedAt time.Time
}

const pidDirCacheTTL = 30 * time.Second

// NewClaudeCodeProvider creates a new ClaudeCodeProvider.
func NewClaudeCodeProvider(devDir string) (*ClaudeCodeProvider, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, &ScanError{Op: "init", Err: fmt.Errorf("getting home dir: %w", err)}
	}

	if devDir == "" {
		devDir = os.Getenv("MASHED_DEV_DIR")
	}
	if devDir == "" {
		devDir = filepath.Join(home, "Development")
	}

	claudeDir := filepath.Join(home, ".claude", "projects")

	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		return nil, &ScanError{Op: "init", Err: fmt.Errorf("creating claude dir %s: %w", claudeDir, err)}
	}

	return &ClaudeCodeProvider{
		devDir:      devDir,
		homeDir:     home,
		claudeDir:   claudeDir,
		pidDirCache: make(map[int]pidDirEntry),
		parsers:     make(map[string]*sessionParserState),
	}, nil
}

// SessionDir returns the session storage directory for a given repo path.
// /Users/linus/Development/mashed -> ~/.claude/projects/-Users-linus-Development-mashed
func (p *ClaudeCodeProvider) SessionDir(repoPath string) string {
	key := strings.ReplaceAll(repoPath, "/", "-")
	return filepath.Join(p.claudeDir, key)
}

// DevDir returns the configured development directory.
func (p *ClaudeCodeProvider) DevDir() string {
	return p.devDir
}

func (p *ClaudeCodeProvider) cachePidDir(pid int, dir string) {
	p.pidDirMu.Lock()
	p.pidDirCache[pid] = pidDirEntry{dir: dir, cachedAt: time.Now()}
	p.pidDirMu.Unlock()
}

func (p *ClaudeCodeProvider) getCachedPidDir(pid int) (string, bool) {
	p.pidDirMu.RLock()
	entry, ok := p.pidDirCache[pid]
	p.pidDirMu.RUnlock()

	if !ok || time.Since(entry.cachedAt) > pidDirCacheTTL {
		return "", false
	}
	return entry.dir, true
}
