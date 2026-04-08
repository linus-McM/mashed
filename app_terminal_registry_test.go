package main

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"mashed/internal/domain"
	"mashed/internal/terminal"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── fakePaneDiscovery satisfies paneDiscoverer for testing ──

type fakePaneDiscovery struct {
	panes []terminal.TmuxPane
	err   error
}

func (f *fakePaneDiscovery) ListPanes() ([]terminal.TmuxPane, error) {
	return f.panes, f.err
}

func (f *fakePaneDiscovery) InvalidateCache() {}

func (f *fakePaneDiscovery) FindPaneForPID(_ int) (*terminal.TmuxPane, error) {
	return nil, nil
}

// testApp builds an App with a fake pane discoverer for registry tests.
func testApp(panes []terminal.TmuxPane) *App {
	return &App{
		panes:            &fakePaneDiscovery{panes: panes},
		terminalSessions: make(map[string]domain.TerminalSession),
	}
}

// ── AC-1: TerminalSession JSON serialization ──

func TestStory1_AC1_TerminalSessionJSONTags(t *testing.T) {
	ts := domain.TerminalSession{
		SessionName: "term-myrepo-1712600000",
		PaneTarget:  "term-myrepo-1712600000:0.0",
		RepoPath:    "/dev/myrepo",
		RepoName:    "myrepo",
		SessionType: domain.SessionTerminal,
		Model:       "",
		SpawnedAt:   time.Unix(1712600000, 0),
		IsAlive:     true,
	}

	data, err := json.Marshal(ts)
	require.NoError(t, err)

	var m map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &m))

	for _, key := range []string{
		"sessionName", "paneTarget", "repoPath", "repoName",
		"sessionType", "model", "spawnedAt", "isAlive",
	} {
		assert.Contains(t, m, key, "JSON missing key %q", key)
	}
}

func TestStory1_AC1_TerminalSessionRoundTrip(t *testing.T) {
	original := domain.TerminalSession{
		SessionName: "mashed-foo-100",
		PaneTarget:  "mashed-foo-100:0.0",
		RepoPath:    "/dev/foo",
		RepoName:    "foo",
		SessionType: domain.SessionAgent,
		Model:       "claude-opus-4-6",
		SpawnedAt:   time.Unix(100, 0),
		IsAlive:     false,
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded domain.TerminalSession
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, original.SessionName, decoded.SessionName)
	assert.Equal(t, original.PaneTarget, decoded.PaneTarget)
	assert.Equal(t, original.RepoPath, decoded.RepoPath)
	assert.Equal(t, original.RepoName, decoded.RepoName)
	assert.Equal(t, original.SessionType, decoded.SessionType)
	assert.Equal(t, original.Model, decoded.Model)
	assert.True(t, original.SpawnedAt.Equal(decoded.SpawnedAt))
	assert.Equal(t, original.IsAlive, decoded.IsAlive)
}

// ── AC-2: ListRepoSessions filters by repo and sorts ──

func TestStory1_AC2_FilterByRepo(t *testing.T) {
	app := testApp([]terminal.TmuxPane{
		{SessionName: "mashed-foo-100"},
		{SessionName: "term-bar-200"},
	})
	app.registerSession(domain.TerminalSession{
		SessionName: "mashed-foo-100",
		RepoPath:    "/dev/foo",
		SpawnedAt:   time.Unix(100, 0),
	})
	app.registerSession(domain.TerminalSession{
		SessionName: "term-bar-200",
		RepoPath:    "/dev/bar",
		SpawnedAt:   time.Unix(200, 0),
	})

	result := app.ListRepoSessions("/dev/foo")

	require.Len(t, result, 1)
	assert.Equal(t, "mashed-foo-100", result[0].SessionName)
	assert.True(t, result[0].IsAlive)
}

func TestStory1_AC2_EmptyRepoPath(t *testing.T) {
	app := testApp(nil)
	app.registerSession(domain.TerminalSession{
		SessionName: "mashed-foo-100",
		RepoPath:    "/dev/foo",
	})

	result := app.ListRepoSessions("")

	assert.Empty(t, result)
}

func TestStory1_AC2_SortBySpawnedAt(t *testing.T) {
	app := testApp([]terminal.TmuxPane{
		{SessionName: "term-foo-300"},
		{SessionName: "term-foo-100"},
		{SessionName: "term-foo-200"},
	})
	app.registerSession(domain.TerminalSession{
		SessionName: "term-foo-300", RepoPath: "/dev/foo", SpawnedAt: time.Unix(300, 0),
	})
	app.registerSession(domain.TerminalSession{
		SessionName: "term-foo-100", RepoPath: "/dev/foo", SpawnedAt: time.Unix(100, 0),
	})
	app.registerSession(domain.TerminalSession{
		SessionName: "term-foo-200", RepoPath: "/dev/foo", SpawnedAt: time.Unix(200, 0),
	})

	result := app.ListRepoSessions("/dev/foo")

	require.Len(t, result, 3)
	assert.Equal(t, "term-foo-100", result[0].SessionName)
	assert.Equal(t, "term-foo-200", result[1].SessionName)
	assert.Equal(t, "term-foo-300", result[2].SessionName)
}

// ── AC-3: ListRepoSessions prunes dead sessions ──

func TestStory1_AC3_PruneDeadSessions(t *testing.T) {
	app := testApp([]terminal.TmuxPane{
		{SessionName: "term-foo-alive"},
	})
	app.registerSession(domain.TerminalSession{
		SessionName: "term-foo-alive", RepoPath: "/dev/foo", SpawnedAt: time.Unix(100, 0),
	})
	app.registerSession(domain.TerminalSession{
		SessionName: "term-foo-dead", RepoPath: "/dev/foo", SpawnedAt: time.Unix(200, 0),
	})

	result := app.ListRepoSessions("/dev/foo")

	require.Len(t, result, 1)
	assert.Equal(t, "term-foo-alive", result[0].SessionName)
	assert.True(t, result[0].IsAlive)

	app.mu.Lock()
	_, exists := app.terminalSessions["term-foo-dead"]
	app.mu.Unlock()
	assert.False(t, exists, "dead session should be pruned from registry")
}

func TestStory1_AC3_TmuxNotRunning(t *testing.T) {
	app := &App{
		panes: &fakePaneDiscovery{
			err: &terminal.TerminalError{Op: "list_panes", Err: terminal.ErrTmuxNotRunning},
		},
		terminalSessions: make(map[string]domain.TerminalSession),
	}
	app.registerSession(domain.TerminalSession{
		SessionName: "term-foo-100", RepoPath: "/dev/foo", SpawnedAt: time.Unix(100, 0),
	})

	result := app.ListRepoSessions("/dev/foo")

	assert.Empty(t, result, "all sessions pruned when tmux not running")
	app.mu.Lock()
	_, exists := app.terminalSessions["term-foo-100"]
	app.mu.Unlock()
	assert.False(t, exists)
}

// ── AC-4: KillTerminalSession removes from registry ──

func TestStory1_AC4_KillRemovesFromRegistry(t *testing.T) {
	app := testApp(nil)
	app.registerSession(domain.TerminalSession{
		SessionName: "term-foo-100", RepoPath: "/dev/foo",
	})

	app.mu.Lock()
	_, exists := app.terminalSessions["term-foo-100"]
	app.mu.Unlock()
	require.True(t, exists)

	err := app.KillTerminalSession("term-foo-100")
	require.NoError(t, err)

	app.mu.Lock()
	_, existsAfter := app.terminalSessions["term-foo-100"]
	app.mu.Unlock()
	assert.False(t, existsAfter, "session removed from registry after kill")
}

// ── AC-5: KillTerminalSession handles already-dead and not-found ──

func TestStory1_AC5_KillAlreadyDead(t *testing.T) {
	app := testApp(nil)
	app.registerSession(domain.TerminalSession{
		SessionName: "term-foo-999", RepoPath: "/dev/foo",
	})

	err := app.KillTerminalSession("term-foo-999")
	require.NoError(t, err)

	app.mu.Lock()
	_, exists := app.terminalSessions["term-foo-999"]
	app.mu.Unlock()
	assert.False(t, exists, "already-dead session removed from registry")
}

func TestStory1_AC5_KillNotFound(t *testing.T) {
	app := testApp(nil)

	err := app.KillTerminalSession("nonexistent")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// ── Extra: registerSession overwrite, init, concurrency ──

func TestStory1_RegisterOverwrite(t *testing.T) {
	app := testApp(nil)
	app.registerSession(domain.TerminalSession{
		SessionName: "term-foo-100", RepoPath: "/dev/foo",
		SessionType: domain.SessionTerminal,
	})
	app.registerSession(domain.TerminalSession{
		SessionName: "term-foo-100", RepoPath: "/dev/foo",
		SessionType: domain.SessionAgent, Model: "claude-opus-4-6",
	})

	app.mu.Lock()
	stored := app.terminalSessions["term-foo-100"]
	app.mu.Unlock()

	assert.Equal(t, domain.SessionAgent, stored.SessionType, "second register overwrites")
	assert.Equal(t, "claude-opus-4-6", stored.Model)
}

func TestStory1_NewAppInitializesMap(t *testing.T) {
	app := NewApp()
	require.NotNil(t, app.terminalSessions, "terminalSessions map must be initialized")
	assert.Empty(t, app.terminalSessions)
}

func TestStory1_RegisterSessionConcurrent(t *testing.T) {
	app := testApp(nil)
	done := make(chan struct{})

	for i := 0; i < 100; i++ {
		go func(n int) {
			app.registerSession(domain.TerminalSession{
				SessionName: fmt.Sprintf("term-foo-%d", n),
				RepoPath:    "/dev/foo",
			})
			done <- struct{}{}
		}(i)
	}

	for i := 0; i < 100; i++ {
		<-done
	}

	app.mu.Lock()
	count := len(app.terminalSessions)
	app.mu.Unlock()
	assert.Greater(t, count, 0)
}
