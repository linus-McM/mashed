package agent

import (
	"testing"
	"time"

	"conductor/internal/domain"
)

// drainEvents reads all available events from the engine channel within a timeout.
func drainEvents(t *testing.T, engine *NotificationEngine, timeout time.Duration) []domain.NotificationEvent {
	t.Helper()
	var events []domain.NotificationEvent
	deadline := time.After(timeout)
	for {
		select {
		case evt, ok := <-engine.Events():
			if !ok {
				return events
			}
			events = append(events, evt)
		case <-deadline:
			return events
		}
	}
}

func TestStateTransition_RunningToError(t *testing.T) {
	engine := NewNotificationEngine(nil) // nil ctx skips Wails emit

	agent := domain.Agent{
		ID:     "pid-1234",
		Name:   "test-agent",
		Status: domain.StatusRunning,
		Model:  "opus",
	}

	// First transition: unknown -> running (emits EventRunning)
	if err := engine.ProcessAgentUpdate(agent, "myrepo", "main"); err != nil {
		t.Fatalf("ProcessAgentUpdate (running) error: %v", err)
	}

	// Second transition: running -> error (emits EventError)
	agent.Status = domain.StatusError
	if err := engine.ProcessAgentUpdate(agent, "myrepo", "main"); err != nil {
		t.Fatalf("ProcessAgentUpdate (error) error: %v", err)
	}

	events := drainEvents(t, engine, 100*time.Millisecond)
	if len(events) != 2 {
		t.Fatalf("got %d events, want 2", len(events))
	}

	if events[0].EventType != domain.EventRunning {
		t.Errorf("events[0].EventType = %q, want %q", events[0].EventType, domain.EventRunning)
	}
	if events[1].EventType != domain.EventError {
		t.Errorf("events[1].EventType = %q, want %q", events[1].EventType, domain.EventError)
	}
}

func TestDeduplication_SameStateTwice(t *testing.T) {
	engine := NewNotificationEngine(nil)

	agent := domain.Agent{
		ID:     "pid-5678",
		Name:   "dedup-agent",
		Status: domain.StatusRunning,
		Model:  "sonnet",
	}

	// First: unknown -> running (emits)
	if err := engine.ProcessAgentUpdate(agent, "repo", "main"); err != nil {
		t.Fatal(err)
	}

	// Second: running -> running (should NOT emit)
	if err := engine.ProcessAgentUpdate(agent, "repo", "main"); err != nil {
		t.Fatal(err)
	}

	// Third: running -> running again (should NOT emit)
	if err := engine.ProcessAgentUpdate(agent, "repo", "main"); err != nil {
		t.Fatal(err)
	}

	events := drainEvents(t, engine, 100*time.Millisecond)
	if len(events) != 1 {
		t.Errorf("got %d events, want 1 (deduplication should suppress repeats)", len(events))
	}
}

func TestAgentRestart_CompletedStartedRunning(t *testing.T) {
	engine := NewNotificationEngine(nil)

	agent := domain.Agent{
		ID:     "pid-9012",
		Name:   "restart-agent",
		Status: domain.StatusRunning,
		Model:  "opus",
	}

	// Phase 1: initial running
	if err := engine.ProcessAgentUpdate(agent, "repo", "feat"); err != nil {
		t.Fatal(err)
	}

	// Phase 2: completed
	agent.Status = domain.StatusDone
	if err := engine.ProcessAgentUpdate(agent, "repo", "feat"); err != nil {
		t.Fatal(err)
	}

	// Phase 3: restarted (queued = started)
	agent.Status = domain.StatusQueued
	if err := engine.ProcessAgentUpdate(agent, "repo", "feat"); err != nil {
		t.Fatal(err)
	}

	// Phase 4: running again
	agent.Status = domain.StatusRunning
	if err := engine.ProcessAgentUpdate(agent, "repo", "feat"); err != nil {
		t.Fatal(err)
	}

	events := drainEvents(t, engine, 100*time.Millisecond)
	if len(events) != 4 {
		t.Fatalf("got %d events, want 4 (running, completed, started, running)", len(events))
	}

	wantTypes := []domain.EventType{
		domain.EventRunning,
		domain.EventCompleted,
		domain.EventStarted,
		domain.EventRunning,
	}
	for i, want := range wantTypes {
		if events[i].EventType != want {
			t.Errorf("events[%d].EventType = %q, want %q", i, events[i].EventType, want)
		}
	}
}

func TestPriorityOrdering(t *testing.T) {
	tests := []struct {
		name     string
		evtType  domain.EventType
		wantPrio int
	}{
		{"needs_response is 0", domain.EventNeedsResponse, 0},
		{"error is 1", domain.EventError, 1},
		{"completed is 2", domain.EventCompleted, 2},
		{"running is 3", domain.EventRunning, 3},
		{"started is 4", domain.EventStarted, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := priorityFor(tt.evtType)
			if got != tt.wantPrio {
				t.Errorf("priorityFor(%q) = %d, want %d", tt.evtType, got, tt.wantPrio)
			}
		})
	}

	// Verify relative ordering: needs_response < error < completed < running < started
	for i := 1; i < len(tests); i++ {
		prev := priorityFor(tests[i-1].evtType)
		curr := priorityFor(tests[i].evtType)
		if prev >= curr {
			t.Errorf("priority(%q)=%d should be less than priority(%q)=%d",
				tests[i-1].evtType, prev, tests[i].evtType, curr)
		}
	}
}

func TestSortByPriority(t *testing.T) {
	now := time.Now()
	events := []domain.NotificationEvent{
		{EventType: domain.EventRunning, Priority: 3, Timestamp: now},
		{EventType: domain.EventError, Priority: 1, Timestamp: now.Add(-1 * time.Second)},
		{EventType: domain.EventCompleted, Priority: 2, Timestamp: now},
		{EventType: domain.EventNeedsResponse, Priority: 0, Timestamp: now},
		{EventType: domain.EventStarted, Priority: 4, Timestamp: now},
	}

	sorted := SortByPriority(events)

	wantOrder := []domain.EventType{
		domain.EventNeedsResponse,
		domain.EventError,
		domain.EventCompleted,
		domain.EventRunning,
		domain.EventStarted,
	}

	for i, want := range wantOrder {
		if sorted[i].EventType != want {
			t.Errorf("sorted[%d].EventType = %q, want %q", i, sorted[i].EventType, want)
		}
	}
}

func TestProcessAgentUpdate_EmptyID(t *testing.T) {
	engine := NewNotificationEngine(nil)

	agent := domain.Agent{
		ID:     "",
		Name:   "no-id",
		Status: domain.StatusRunning,
	}

	err := engine.ProcessAgentUpdate(agent, "repo", "main")
	if err == nil {
		t.Fatal("expected error for empty agent ID, got nil")
	}
}

func TestActiveAgentCount(t *testing.T) {
	engine := NewNotificationEngine(nil)

	if got := engine.ActiveAgentCount(); got != 0 {
		t.Errorf("initial ActiveAgentCount = %d, want 0", got)
	}

	for i := 0; i < 3; i++ {
		a := domain.Agent{
			ID:     "pid-" + string(rune('A'+i)),
			Name:   "agent",
			Status: domain.StatusRunning,
		}
		_ = engine.ProcessAgentUpdate(a, "repo", "main")
	}

	if got := engine.ActiveAgentCount(); got != 3 {
		t.Errorf("ActiveAgentCount after 3 agents = %d, want 3", got)
	}

	engine.RemoveAgent("pid-A")
	if got := engine.ActiveAgentCount(); got != 2 {
		t.Errorf("ActiveAgentCount after removal = %d, want 2", got)
	}
}
