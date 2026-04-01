// Package terminal provides tmux pane discovery and WebSocket terminal bridging.
package terminal

import (
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Sentinel errors for pane discovery.
var (
	ErrTmuxNotRunning = errors.New("tmux server not running")
	ErrPaneNotFound   = errors.New("pane not found for agent PID")
	ErrPPIDWalkLimit  = errors.New("PPID chain exceeded max depth")
)

// TerminalError provides structured error context for terminal operations.
type TerminalError struct {
	Op  string // operation that failed
	PID int    // process ID involved, 0 if not applicable
	Err error  // underlying error
}

func (e *TerminalError) Error() string {
	if e.PID != 0 {
		return fmt.Sprintf("terminal: %s (pid=%d): %v", e.Op, e.PID, e.Err)
	}
	return fmt.Sprintf("terminal: %s: %v", e.Op, e.Err)
}

func (e *TerminalError) Unwrap() error {
	return e.Err
}

// TmuxPane represents a discovered tmux pane with its process and target info.
type TmuxPane struct {
	PanePID     int    // PID of the process running in the pane
	PaneID      string // tmux pane ID like %0, %1
	SessionName string
	WindowIndex int
	PaneIndex   int
	PaneTTY     string // tty device path (e.g., /dev/ttys005)
}

// Target returns the tmux target string for this pane (session:window.pane).
func (p TmuxPane) Target() string {
	return fmt.Sprintf("%s:%d.%d", p.SessionName, p.WindowIndex, p.PaneIndex)
}

const (
	maxPPIDDepth = 8
	paneCacheTTL = 5 * time.Second
)

// PaneDiscovery discovers and caches tmux pane information.
type PaneDiscovery struct {
	mu          sync.Mutex
	cachedPanes []TmuxPane
	cacheTime   time.Time
}

// NewPaneDiscovery creates a new PaneDiscovery instance.
func NewPaneDiscovery() *PaneDiscovery {
	return &PaneDiscovery{}
}

// ListPanes returns all tmux panes, using a 5-second cache.
func (pd *PaneDiscovery) ListPanes() ([]TmuxPane, error) {
	pd.mu.Lock()
	defer pd.mu.Unlock()

	if time.Since(pd.cacheTime) < paneCacheTTL && pd.cachedPanes != nil {
		return pd.cachedPanes, nil
	}

	panes, err := discoverPanes()
	if err != nil {
		return nil, err
	}

	pd.cachedPanes = panes
	pd.cacheTime = time.Now()
	return panes, nil
}

// FindPaneForPID finds the tmux pane containing the given agent PID by walking
// the PPID chain up to maxPPIDDepth levels. An agent process is a descendant
// of the shell running inside the tmux pane.
func (pd *PaneDiscovery) FindPaneForPID(agentPID int) (*TmuxPane, error) {
	panes, err := pd.ListPanes()
	if err != nil {
		return nil, &TerminalError{Op: "find_pane", PID: agentPID, Err: err}
	}

	panePIDMap := make(map[int]*TmuxPane, len(panes))
	for i := range panes {
		panePIDMap[panes[i].PanePID] = &panes[i]
	}

	pid := agentPID
	for depth := 0; depth < maxPPIDDepth; depth++ {
		if pane, ok := panePIDMap[pid]; ok {
			return pane, nil
		}
		ppid, err := getParentPID(pid)
		if err != nil {
			return nil, &TerminalError{
				Op:  "find_pane",
				PID: agentPID,
				Err: fmt.Errorf("ppid lookup at depth %d: %w", depth, err),
			}
		}
		if ppid <= 1 {
			break
		}
		pid = ppid
	}

	return nil, &TerminalError{Op: "find_pane", PID: agentPID, Err: ErrPaneNotFound}
}

// InvalidateCache forces the next ListPanes call to re-query tmux.
func (pd *PaneDiscovery) InvalidateCache() {
	pd.mu.Lock()
	pd.cacheTime = time.Time{}
	pd.mu.Unlock()
}

func discoverPanes() ([]TmuxPane, error) {
	cmd := exec.Command("tmux", "list-panes", "-a", "-F",
		"#{pane_pid}\t#{pane_id}\t#{session_name}\t#{window_index}\t#{pane_index}\t#{pane_tty}")
	out, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			stderr := string(exitErr.Stderr)
			if strings.Contains(stderr, "no server running") ||
				strings.Contains(stderr, "no current client") {
				return nil, &TerminalError{Op: "list_panes", Err: ErrTmuxNotRunning}
			}
		}
		return nil, &TerminalError{Op: "list_panes", Err: fmt.Errorf("tmux list-panes: %w", err)}
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	panes := make([]TmuxPane, 0, len(lines))
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 6)
		if len(parts) < 6 {
			continue
		}

		pid, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}
		winIdx, err := strconv.Atoi(parts[3])
		if err != nil {
			continue
		}
		paneIdx, err := strconv.Atoi(parts[4])
		if err != nil {
			continue
		}

		panes = append(panes, TmuxPane{
			PanePID:     pid,
			PaneID:      parts[1],
			SessionName: parts[2],
			WindowIndex: winIdx,
			PaneIndex:   paneIdx,
			PaneTTY:     parts[5],
		})
	}
	return panes, nil
}

func getParentPID(pid int) (int, error) {
	cmd := exec.Command("ps", "-o", "ppid=", "-p", strconv.Itoa(pid))
	out, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("ps ppid for pid %d: %w", pid, err)
	}
	ppid, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		return 0, fmt.Errorf("parse ppid for pid %d: %w", pid, err)
	}
	return ppid, nil
}
