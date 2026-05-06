package bmad

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
)

// CleanupStaleSessions removes orphaned tmux sessions whose names start with
// SessionNamePrefix and are NOT referenced by any tracked execution.
//
// "no server running" from tmux list-sessions is treated as a no-op success
// (there is no daemon, therefore no orphans). Per-kill failures are joined
// into a single error so callers can use errors.Is on any failure cause
// without losing visibility into the remaining ones.
//
// Intended for app startup, where the executions map is empty and any
// bmad-prefixed session is unambiguously an orphan from a prior process run.
func (e *Executor) CleanupStaleSessions(ctx context.Context) error {
	out, listErr := e.runCmd(ctx, "tmux", "list-sessions", "-F", "#{session_name}")
	if listErr != nil {
		if isNoTmuxServer(listErr.Error()) || isNoTmuxServer(string(out)) {
			return nil
		}
		return fmt.Errorf("bmad cleanup: list-sessions: %w", listErr)
	}
	if isNoTmuxServer(string(out)) {
		return nil
	}

	live := e.liveSessionNames()

	var killErrs []error
	for _, name := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		name = strings.TrimSpace(name)
		if name == "" || !strings.HasPrefix(name, SessionNamePrefix) {
			continue
		}
		if _, tracked := live[name]; tracked {
			log.Printf("bmad cleanup: preserving tracked session %q", name)
			continue
		}
		if _, err := e.runCmd(ctx, "tmux", "kill-session", "-t", name); err != nil {
			killErrs = append(killErrs, fmt.Errorf("kill %q: %w", name, err))
			continue
		}
		log.Printf("bmad cleanup: killed orphan session %q", name)
	}
	if len(killErrs) > 0 {
		return fmt.Errorf("bmad cleanup: %w", errors.Join(killErrs...))
	}
	return nil
}

// isNoTmuxServer reports whether the supplied tmux output/error indicates that
// no tmux daemon is running. tmux's wording differs by platform: Linux prints
// "no server running on <socket>" while macOS prints
// "error connecting to <socket> (No such file or directory)".
func isNoTmuxServer(s string) bool {
	return strings.Contains(s, "no server running") ||
		strings.Contains(s, "error connecting to")
}

// liveSessionNames returns the bare tmux session names referenced by tracked
// executions. The ":window.pane" suffix is stripped so callers can compare
// directly against `tmux list-sessions` output.
func (e *Executor) liveSessionNames() map[string]struct{} {
	e.mu.RLock()
	states := make([]*execState, 0, len(e.executions))
	for _, s := range e.executions {
		states = append(states, s)
	}
	e.mu.RUnlock()

	live := make(map[string]struct{})
	for _, s := range states {
		s.mu.Lock()
		if s.exec != nil {
			for i := range s.exec.Nodes {
				if name := bareSessionName(s.exec.Nodes[i].TmuxTarget); name != "" {
					live[name] = struct{}{}
				}
			}
		}
		s.mu.Unlock()
	}
	return live
}

// bareSessionName strips the optional ":window.pane" suffix from a tmux
// target, returning just the session name.
func bareSessionName(target string) string {
	if i := strings.Index(target, ":"); i >= 0 {
		return target[:i]
	}
	return target
}
