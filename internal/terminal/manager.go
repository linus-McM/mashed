package terminal

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"

	"github.com/creack/pty"
)

// Sentinel errors for SessionManager.
var (
	ErrSessionExists   = errors.New("terminal: session already exists")
	ErrSessionNotFound = errors.New("terminal: session not found")
)

// SessionManager owns a map of named ManagedSession instances and provides
// Spawn/Get/Kill/List/FindByPID/Shutdown operations.
type SessionManager struct {
	mu       sync.Mutex
	sessions map[string]*ManagedSession
}

// NewSessionManager creates a new SessionManager.
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*ManagedSession),
	}
}

// Spawn creates a new PTY session with the given name, working directory, and command.
// If command is empty, the user's default shell is used.
func (sm *SessionManager) Spawn(ctx context.Context, name string, repoPath string, command string) (*ManagedSession, error) {
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

	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	cmd.Dir = repoPath
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{Cols: 80, Rows: 24})
	if err != nil {
		sm.mu.Lock()
		delete(sm.sessions, name)
		sm.mu.Unlock()
		return nil, &TerminalError{Op: "spawn", Err: fmt.Errorf("pty start %q: %w", name, err)}
	}

	if cmd.Process == nil || cmd.Process.Pid <= 0 {
		ptmx.Close()
		sm.mu.Lock()
		delete(sm.sessions, name)
		sm.mu.Unlock()
		return nil, &TerminalError{Op: "spawn", Err: fmt.Errorf("invalid PID for session %q", name)}
	}

	ms := newManagedSession(name, cmd, ptmx)
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
	delete(sm.sessions, name)
	sm.mu.Unlock()

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
		if ms.cmd.Process != nil && ms.cmd.Process.Pid == pid {
			return ms, true
		}
	}
	return nil, false
}

// Shutdown kills all sessions and clears the session map.
func (sm *SessionManager) Shutdown() {
	sm.mu.Lock()
	sessions := make([]*ManagedSession, 0, len(sm.sessions))
	for _, ms := range sm.sessions {
		if ms != nil {
			sessions = append(sessions, ms)
		}
	}
	sm.sessions = make(map[string]*ManagedSession)
	sm.mu.Unlock()

	for _, ms := range sessions {
		ms.Kill()
	}
}
