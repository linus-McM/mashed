// Package agent implements the notification state machine for agent lifecycle events.
package agent

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"conductor/internal/domain"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// Sentinel errors for the notification engine.
var (
	ErrEngineNotStarted = errors.New("agent: engine not started")
	ErrAgentNotFound    = errors.New("agent: agent not found")
)

// EngineError represents an error in the notification engine.
type EngineError struct {
	AgentID string
	Op      string
	Err     error
}

func (e *EngineError) Error() string {
	if e.AgentID != "" {
		return fmt.Sprintf("engine %s (agent %s): %v", e.Op, e.AgentID, e.Err)
	}
	return fmt.Sprintf("engine %s: %v", e.Op, e.Err)
}

func (e *EngineError) Unwrap() error { return e.Err }

// agentState tracks the last known state of an agent for deduplication.
type agentState struct {
	status    domain.AgentStatus
	lastEvent time.Time
}

// NotificationEngine manages agent state transitions and emits notification events.
// It deduplicates by only emitting when an agent's status actually changes
// (e.g. running->error, not running->running).
type NotificationEngine struct {
	ctx context.Context // Wails app context for EventsEmit

	mu     sync.RWMutex
	states map[string]*agentState // AgentID -> last known state

	events chan domain.NotificationEvent
}

// NewNotificationEngine creates a new engine with the given Wails app context.
func NewNotificationEngine(ctx context.Context) *NotificationEngine {
	return &NotificationEngine{
		ctx:    ctx,
		states: make(map[string]*agentState),
		events: make(chan domain.NotificationEvent, 128),
	}
}

// Events returns the read-only channel of notification events.
func (e *NotificationEngine) Events() <-chan domain.NotificationEvent {
	return e.events
}

// ProcessAgentUpdate checks for state transitions and emits events only on change.
// Priority: 0=needs-response, 1=error, 2=completed, 3=running, 4=started.
func (e *NotificationEngine) ProcessAgentUpdate(agent domain.Agent, repoName, repoBranch string) error {
	if agent.ID == "" {
		return &EngineError{Op: "process-update", Err: fmt.Errorf("empty agent ID: %w", ErrAgentNotFound)}
	}

	newStatus := agent.Status
	eventType := statusToEventType(newStatus)

	e.mu.Lock()
	prev, existed := e.states[agent.ID]

	// Only emit on state TRANSITIONS — skip if same status
	if existed && prev.status == newStatus {
		e.mu.Unlock()
		return nil
	}

	now := time.Now()
	e.states[agent.ID] = &agentState{
		status:    newStatus,
		lastEvent: now,
	}
	e.mu.Unlock()

	event := domain.NotificationEvent{
		ID:         fmt.Sprintf("%s-%d", agent.ID, now.UnixNano()),
		AgentID:    agent.ID,
		AgentName:  agent.Name,
		Model:      agent.Model,
		RepoName:   repoName,
		RepoBranch: repoBranch,
		EventType:  eventType,
		Summary:    buildSummary(agent, existed),
		Timestamp:  now,
		Priority:   priorityFor(eventType),
		TokensUsed: agent.TokensUsed,
		TokensMax:  agent.TokensMax,
		TmuxTarget: agent.TmuxTarget,
		RepoPath:   agent.RepoPath,
		PID:        agent.PID,
	}

	// Emit to Wails frontend
	if e.ctx != nil {
		wailsRuntime.EventsEmit(e.ctx, "agent:notification", event)
	}

	// Non-blocking send to internal channel
	select {
	case e.events <- event:
	default:
		// Channel full — drop to avoid blocking the scanner goroutine
	}

	return nil
}

// SortByPriority returns events sorted by priority (lower number = higher urgency),
// then by timestamp (newest first) within the same priority level.
func SortByPriority(events []domain.NotificationEvent) []domain.NotificationEvent {
	sorted := make([]domain.NotificationEvent, len(events))
	copy(sorted, events)

	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Priority != sorted[j].Priority {
			return sorted[i].Priority < sorted[j].Priority
		}
		return sorted[i].Timestamp.After(sorted[j].Timestamp)
	})

	return sorted
}

// RemoveAgent cleans up tracking state for a terminated agent.
func (e *NotificationEngine) RemoveAgent(agentID string) {
	e.mu.Lock()
	delete(e.states, agentID)
	e.mu.Unlock()
}

// ActiveAgentCount returns the number of agents currently tracked.
func (e *NotificationEngine) ActiveAgentCount() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.states)
}

// GetAgentStatus returns the last known status for an agent, or false if not tracked.
func (e *NotificationEngine) GetAgentStatus(agentID string) (domain.AgentStatus, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	state, ok := e.states[agentID]
	if !ok {
		return "", false
	}
	return state.status, true
}

// priorityFor maps event types to priority numbers.
// 0=needs-response, 1=error, 2=completed, 3=running, 4=started.
func priorityFor(et domain.EventType) int {
	switch et {
	case domain.EventNeedsResponse:
		return 0
	case domain.EventError:
		return 1
	case domain.EventCompleted:
		return 2
	case domain.EventRunning:
		return 3
	case domain.EventStarted:
		return 4
	default:
		return 5
	}
}

func statusToEventType(status domain.AgentStatus) domain.EventType {
	switch status {
	case domain.StatusRunning:
		return domain.EventRunning
	case domain.StatusBlocked:
		return domain.EventNeedsResponse
	case domain.StatusError:
		return domain.EventError
	case domain.StatusDone:
		return domain.EventCompleted
	case domain.StatusQueued:
		return domain.EventStarted
	default:
		return domain.EventRunning
	}
}

func buildSummary(agent domain.Agent, wasTracked bool) string {
	if !wasTracked {
		return fmt.Sprintf("Agent %s started (%s)", agent.Name, agent.Model)
	}

	switch agent.Status {
	case domain.StatusDone:
		tokens := agent.TokensUsed / (1000 * 1000)
		return fmt.Sprintf("Agent %s completed (%dM tokens)", agent.Name, tokens)
	case domain.StatusError:
		return fmt.Sprintf("Agent %s encountered an error", agent.Name)
	case domain.StatusBlocked:
		return fmt.Sprintf("Agent %s needs your response", agent.Name)
	case domain.StatusRunning:
		return fmt.Sprintf("Agent %s resumed", agent.Name)
	default:
		return fmt.Sprintf("Agent %s: %s", agent.Name, agent.Status)
	}
}
