package agent

import (
	"encoding/json"
	"strings"
	"sync"
	"testing"

	"mashed/internal/domain"
)

// Story uiqa-09: Render SparkLine in NotificationFeed.
// Covers backend helper + race test for rolling token-sample window.

func TestMaybeAppendTokenSample_CapEnforced(t *testing.T) {
	var samples []int
	// Append 25 distinct values with large deltas so throttling never suppresses them.
	for i := 1; i <= 25; i++ {
		samples = MaybeAppendTokenSample(samples, int64(i*1000))
	}
	if got, want := len(samples), MaxTokenSamples; got != want {
		t.Fatalf("len(samples) = %d, want %d", got, want)
	}
	// Newest sample (25000) must be last.
	if samples[len(samples)-1] != 25000 {
		t.Errorf("tail sample = %d, want 25000", samples[len(samples)-1])
	}
	// Oldest remaining should be the 6th appended (first 5 dropped).
	if samples[0] != 6000 {
		t.Errorf("head sample = %d, want 6000 (first 5 dropped)", samples[0])
	}
}

func TestMaybeAppendTokenSample_FirstSampleAlwaysAppends(t *testing.T) {
	samples := MaybeAppendTokenSample(nil, 42)
	if len(samples) != 1 || samples[0] != 42 {
		t.Fatalf("first append = %v, want [42]", samples)
	}
}

func TestMaybeAppendTokenSample_DeltaThrottle(t *testing.T) {
	// Story Design Brief: minimum delta 50 tokens.
	samples := MaybeAppendTokenSample(nil, 1000)
	// Small delta (+10) should NOT append.
	samples = MaybeAppendTokenSample(samples, 1010)
	if len(samples) != 1 {
		t.Fatalf("small delta appended; got len=%d, want 1", len(samples))
	}
	// Large delta (+60) SHOULD append.
	samples = MaybeAppendTokenSample(samples, 1060)
	if len(samples) != 2 || samples[1] != 1060 {
		t.Fatalf("large delta not appended; samples=%v", samples)
	}
	// Regression: same value with zero delta must be dropped.
	samples = MaybeAppendTokenSample(samples, 1060)
	if len(samples) != 2 {
		t.Errorf("zero delta appended; samples=%v", samples)
	}
}

func TestMaybeAppendTokenSample_DeltaBypassedForDecreases(t *testing.T) {
	// A decrease of >= minDelta counts as significant (abs delta).
	samples := MaybeAppendTokenSample(nil, 5000)
	samples = MaybeAppendTokenSample(samples, 4900) // -100, abs >= 50 → append
	if len(samples) != 2 {
		t.Errorf("negative-delta sample dropped; samples=%v", samples)
	}
}

func TestMaybeAppendTokenSample_ThreadSafeRace(t *testing.T) {
	// Simulate 100 concurrent writers protected by a single mutex.
	// Run with `go test -race` — any data race fails the test.
	var (
		mu      sync.Mutex
		samples []int
	)
	const writers = 100
	var wg sync.WaitGroup
	wg.Add(writers)
	for i := 0; i < writers; i++ {
		go func(n int) {
			defer wg.Done()
			mu.Lock()
			samples = MaybeAppendTokenSample(samples, int64((n+1)*100))
			mu.Unlock()
		}(i)
	}
	wg.Wait()

	mu.Lock()
	defer mu.Unlock()
	if len(samples) > MaxTokenSamples {
		t.Fatalf("len(samples) = %d, must not exceed cap %d", len(samples), MaxTokenSamples)
	}
}

func TestAgentJSON_IncludesTokenSamples(t *testing.T) {
	ag := domain.Agent{
		ID:           "pid-1",
		TokensUsed:   300,
		TokenSamples: []int{100, 200, 300},
	}
	raw, err := json.Marshal(ag)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	if !strings.Contains(string(raw), `"tokenSamples":[100,200,300]`) {
		t.Errorf("json missing tokenSamples field: %s", string(raw))
	}
}

func TestNotificationEventJSON_IncludesTokenSamples(t *testing.T) {
	// AC-6: Wails binding propagation — NotificationEvent is the DTO emitted to frontend.
	evt := domain.NotificationEvent{
		ID:           "evt-1",
		AgentID:      "pid-1",
		TokensUsed:   500,
		TokenSamples: []int{100, 200, 300, 400, 500},
	}
	raw, err := json.Marshal(evt)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	if !strings.Contains(string(raw), `"tokenSamples":[100,200,300,400,500]`) {
		t.Errorf("notification event json missing tokenSamples: %s", string(raw))
	}
}

func TestNotificationEventJSON_OmitEmptyTokenSamples(t *testing.T) {
	// AC-5 support: when no samples exist, the field is omitted to avoid sending `null`
	// which the frontend guard `agent.tokenSamples?.length > 1` tolerates but we want clean JSON.
	evt := domain.NotificationEvent{ID: "evt-1"}
	raw, err := json.Marshal(evt)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	if strings.Contains(string(raw), `tokenSamples`) {
		t.Errorf("empty tokenSamples should be omitted; json=%s", string(raw))
	}
}

func TestProcessAgentUpdate_PropagatesTokenSamples(t *testing.T) {
	// AC-6: NotificationEngine must copy Agent.TokenSamples into NotificationEvent.TokenSamples.
	engine := NewNotificationEngine(nil)
	ag := domain.Agent{
		ID:           "pid-9001",
		Name:         "opus",
		Model:        "opus",
		Status:       domain.StatusRunning,
		TokensUsed:   1234,
		TokenSamples: []int{10, 20, 30},
	}
	if err := engine.ProcessAgentUpdate(ag, "repo", "main"); err != nil {
		t.Fatalf("ProcessAgentUpdate error: %v", err)
	}

	events := drainEvents(t, engine, 50e6) // 50ms
	if len(events) != 1 {
		t.Fatalf("got %d events, want 1", len(events))
	}
	got := events[0].TokenSamples
	if len(got) != 3 || got[0] != 10 || got[1] != 20 || got[2] != 30 {
		t.Errorf("event.TokenSamples = %v, want [10 20 30]", got)
	}
}
