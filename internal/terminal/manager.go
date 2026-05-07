package terminal

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"syscall"

	"mashed/internal/terminal/helper"
)

// Sentinel errors for SessionManager.
var (
	ErrSessionExists      = errors.New("terminal: session already exists")
	ErrSessionNotFound    = errors.New("terminal: session not found")
	ErrHelperNotRunning   = errors.New("terminal: PTY helper not running")
)

// SessionManager owns a map of named ManagedSession instances and provides
// Spawn/Get/Kill/List/FindByPID/Shutdown operations.
type SessionManager struct {
	mu           sync.Mutex
	sessions     map[string]*ManagedSession
	helperClient *helper.Client
}

// NewSessionManager creates a new SessionManager. The client may be nil;
// Spawn will return ErrHelperNotRunning in that case.
func NewSessionManager(client *helper.Client) *SessionManager {
	return &SessionManager{
		sessions:     make(map[string]*ManagedSession),
		helperClient: client,
	}
}

// Spawn creates a new PTY session with the given name, working directory, and command.
// If command is empty, the user's default shell is used.
// cols/rows set the initial PTY winsize; pass 0 to use the defaults (80x24).
// The helper client must be non-nil; otherwise ErrHelperNotRunning is returned.
func (sm *SessionManager) Spawn(ctx context.Context, name string, repoPath string, command string, cols, rows uint16) (*ManagedSession, error) {
	if sm.helperClient == nil {
		return nil, ErrHelperNotRunning
	}

	var parts []string
	if command == "" {
		shell := os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/zsh"
		}
		parts = []string{shell}
	} else {
		// NOTE: strings.Fields splits on whitespace only — quoted arguments are not
		// handled. This is a documented limitation per the story spec. Callers with
		// complex commands should pre-split arguments.
		parts = strings.Fields(command)
		if len(parts) == 0 {
			return nil, &TerminalError{Op: "spawn", Err: fmt.Errorf("empty command for session %q", name)}
		}
	}

	sm.mu.Lock()
	if _, exists := sm.sessions[name]; exists {
		sm.mu.Unlock()
		return nil, fmt.Errorf("session %q: %w", name, ErrSessionExists)
	}
	sm.sessions[name] = nil // reserve slot
	sm.mu.Unlock()

	if cols == 0 {
		cols = 80
	}
	if rows == 0 {
		rows = 24
	}

	env := applyLoginPATH(append(os.Environ(), "TERM=xterm-256color"))
	ptmx, pid, err := sm.helperClient.Spawn(ctx, helper.SpawnRequest{
		ID:    name,
		Shell: resolveExecutable(parts[0]),
		Args:  parts[1:],
		Env:   env,
		Cwd:   repoPath,
		Cols:  cols,
		Rows:  rows,
	})
	if err != nil {
		sm.mu.Lock()
		delete(sm.sessions, name)
		sm.mu.Unlock()
		return nil, &TerminalError{Op: "spawn", Err: fmt.Errorf("helper spawn %q: %w", name, err)}
	}

	ms := newRemoteManagedSession(name, pid, ptmx)
	sm.mu.Lock()
	sm.sessions[name] = ms
	sm.mu.Unlock()

	return ms, nil
}

// Get returns the session with the given name and true, or nil and false.
func (sm *SessionManager) Get(name string) (*ManagedSession, bool) {
	sm.mu.Lock()
	ms, ok := sm.sessions[name]
	sm.mu.Unlock()
	if !ok || ms == nil {
		return nil, false
	}
	return ms, true
}

// Kill terminates the session with the given name and removes it from the manager.
func (sm *SessionManager) Kill(name string) error {
	sm.mu.Lock()
	ms, ok := sm.sessions[name]
	if !ok || ms == nil {
		sm.mu.Unlock()
		return fmt.Errorf("session %q: %w", name, ErrSessionNotFound)
	}
	client := sm.helperClient
	delete(sm.sessions, name)
	sm.mu.Unlock()

	if client != nil {
		_ = client.Kill(name, syscall.SIGTERM)
	}
	ms.Kill()
	return nil
}

// IsAlive returns true if a session with the given name exists and is alive.
func (sm *SessionManager) IsAlive(name string) bool {
	sm.mu.Lock()
	ms, ok := sm.sessions[name]
	sm.mu.Unlock()
	if !ok || ms == nil {
		return false
	}
	return ms.IsAlive()
}

// List returns all sessions matching the filter. If filter is nil, all sessions are returned.
func (sm *SessionManager) List(filter func(*ManagedSession) bool) []*ManagedSession {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	var result []*ManagedSession
	for _, ms := range sm.sessions {
		if ms == nil {
			continue
		}
		if filter == nil || filter(ms) {
			result = append(result, ms)
		}
	}
	return result
}

// FindByPID returns the session whose process has the given PID, or nil and false.
func (sm *SessionManager) FindByPID(pid int) (*ManagedSession, bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for _, ms := range sm.sessions {
		if ms == nil {
			continue
		}
		if ms.pid == pid {
			return ms, true
		}
	}
	return nil, false
}

// Shutdown kills all sessions, clears the session map, and closes the helper client.
func (sm *SessionManager) Shutdown() {
	sm.mu.Lock()
	sessions := make(map[string]*ManagedSession, len(sm.sessions))
	for name, ms := range sm.sessions {
		if ms != nil {
			sessions[name] = ms
		}
	}
	client := sm.helperClient
	sm.sessions = make(map[string]*ManagedSession)
	sm.mu.Unlock()

	for name, ms := range sessions {
		if client != nil {
			_ = client.Kill(name, syscall.SIGTERM)
		}
		ms.Kill()
	}

	if client != nil {
		_ = client.Close()
	}
}
