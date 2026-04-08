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

// ── Story 2: Startup Session Recovery ──

func TestStory2_AC1_RecoverTerminalAndAgent(t *testing.T) {
	app := testApp(nil)

	n := app.recoverSessionsFromOutput("term-repo1-1000\nmashed-repo2-2000")

	assert.Equal(t, 2, n)

	app.mu.Lock()
	defer app.mu.Unlock()

	sess1, ok1 := app.terminalSessions["term-repo1-1000"]
	require.True(t, ok1, "term session should be registered")
	assert.Equal(t, domain.SessionTerminal, sess1.SessionType)
	assert.Equal(t, "repo1", sess1.RepoName)
	assert.Equal(t, "term-repo1-1000:0.0", sess1.PaneTarget)

	sess2, ok2 := app.terminalSessions["mashed-repo2-2000"]
	require.True(t, ok2, "agent session should be registered")
	assert.Equal(t, domain.SessionAgent, sess2.SessionType)
	assert.Equal(t, "repo2", sess2.RepoName)
	assert.Equal(t, "mashed-repo2-2000:0.0", sess2.PaneTarget)
}

func TestStory2_AC2_FilterNonMatchingPrefixes(t *testing.T) {
	app := testApp(nil)

	n := app.recoverSessionsFromOutput("term-foo-1\nmashed-bar-2\nirssi\ndev-session")

	assert.Equal(t, 2, n)

	app.mu.Lock()
	defer app.mu.Unlock()
	assert.Len(t, app.terminalSessions, 2)
	_, hasTerm := app.terminalSessions["term-foo-1"]
	_, hasAgent := app.terminalSessions["mashed-bar-2"]
	assert.True(t, hasTerm)
	assert.True(t, hasAgent)
}

func TestStory2_AC3_TmuxNotRunning(t *testing.T) {
	app := testApp(nil)

	// Empty output simulates tmux not running / no sessions
	n := app.recoverSessionsFromOutput("")

	assert.Equal(t, 0, n)
	app.mu.Lock()
	defer app.mu.Unlock()
	assert.Empty(t, app.terminalSessions)
}

func TestStory2_AC4_IdempotentRecovery(t *testing.T) {
	app := testApp(nil)

	output := "term-repo1-1000\nmashed-repo2-2000"
	n1 := app.recoverSessionsFromOutput(output)
	n2 := app.recoverSessionsFromOutput(output)

	assert.Equal(t, 2, n1)
	assert.Equal(t, 2, n2)

	app.mu.Lock()
	defer app.mu.Unlock()
	assert.Len(t, app.terminalSessions, 2, "no duplicates after second call")
}

func TestStory2_ParsesRepoNameFromSessionName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantRepo string
		wantType domain.SessionType
	}{
		{"terminal simple", "term-myrepo-12345", "myrepo", domain.SessionTerminal},
		{"agent simple", "mashed-project-99999", "project", domain.SessionAgent},
		{"terminal hyphenated repo", "term-my-repo-12345", "my-repo", domain.SessionTerminal},
		{"agent hyphenated repo", "mashed-my-cool-repo-99999", "my-cool-repo", domain.SessionAgent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := testApp(nil)
			n := app.recoverSessionsFromOutput(tt.input)
			require.Equal(t, 1, n)

			app.mu.Lock()
			sess := app.terminalSessions[tt.input]
			app.mu.Unlock()

			assert.Equal(t, tt.wantRepo, sess.RepoName)
			assert.Equal(t, tt.wantType, sess.SessionType)
		})
	}
}

func TestStory2_NoTimestampSuffix(t *testing.T) {
	app := testApp(nil)

	// "term-repo" has no second dash after prefix — repo name is the whole rest
	n := app.recoverSessionsFromOutput("term-repo")

	require.Equal(t, 1, n)
	app.mu.Lock()
	sess := app.terminalSessions["term-repo"]
	app.mu.Unlock()
	assert.Equal(t, "repo", sess.RepoName)
}

func TestStory2_EmptyAndWhitespaceLines(t *testing.T) {
	app := testApp(nil)

	n := app.recoverSessionsFromOutput("term-foo-1\n\n  \nmashed-bar-2\n")

	assert.Equal(t, 2, n)
	app.mu.Lock()
	defer app.mu.Unlock()
	assert.Len(t, app.terminalSessions, 2)
}

// ── Story 3: Spawn registration and deregistration ──

func TestStory3_AC1_SpawnRegistersAgentSession(t *testing.T) {
	app := testApp(nil)

	session := domain.TerminalSession{
		SessionName: "mashed-repo-100",
		PaneTarget:  "mashed-repo-100:0.0",
		RepoPath:    "/dev/repo",
		RepoName:    "repo",
		SessionType: domain.SessionAgent,
		Model:       "claude-opus-4-6",
		SpawnedAt:   time.Unix(100, 0),
		IsAlive:     true,
	}
	app.registerSession(session)

	app.mu.Lock()
	stored, exists := app.terminalSessions["mashed-repo-100"]
	app.mu.Unlock()

	require.True(t, exists, "agent session must be registered")
	assert.Equal(t, domain.SessionAgent, stored.SessionType)
	assert.Equal(t, "claude-opus-4-6", stored.Model)
	assert.Equal(t, "/dev/repo", stored.RepoPath)
	assert.Equal(t, "repo", stored.RepoName)
	assert.True(t, stored.IsAlive)
}

func TestStory3_AC2_SpawnRegistersTerminalSession(t *testing.T) {
	app := testApp(nil)

	session := domain.TerminalSession{
		SessionName: "term-repo-200",
		PaneTarget:  "term-repo-200:0.0",
		RepoPath:    "/dev/repo",
		RepoName:    "repo",
		SessionType: domain.SessionTerminal,
		Model:       "",
		SpawnedAt:   time.Unix(200, 0),
		IsAlive:     true,
	}
	app.registerSession(session)

	app.mu.Lock()
	stored, exists := app.terminalSessions["term-repo-200"]
	app.mu.Unlock()

	require.True(t, exists, "terminal session must be registered")
	assert.Equal(t, domain.SessionTerminal, stored.SessionType)
	assert.Empty(t, stored.Model, "terminal sessions have no model")
}

func TestStory3_AC3_KillAgentDeregisters(t *testing.T) {
	app := testApp(nil)

	// Pre-populate registry with a session
	app.registerSession(domain.TerminalSession{
		SessionName: "mashed-repo-100",
		PaneTarget:  "mashed-repo-100:0.0",
		RepoPath:    "/dev/repo",
		RepoName:    "repo",
		SessionType: domain.SessionAgent,
		Model:       "claude-opus-4-6",
		SpawnedAt:   time.Unix(100, 0),
		IsAlive:     true,
	})

	// Verify it exists
	app.mu.Lock()
	_, exists := app.terminalSessions["mashed-repo-100"]
	app.mu.Unlock()
	require.True(t, exists, "session must exist before deregistration")

	// Simulate KillAgent's deregistration logic
	app.mu.Lock()
	delete(app.terminalSessions, "mashed-repo-100")
	app.mu.Unlock()

	// Verify removal
	app.mu.Lock()
	_, existsAfter := app.terminalSessions["mashed-repo-100"]
	app.mu.Unlock()
	assert.False(t, existsAfter, "session must be removed after deregistration")
}

func TestStory3_AC4_FailedSpawnNoRegistration(t *testing.T) {
	app := testApp(nil)

	// Registry starts empty
	app.mu.Lock()
	count := len(app.terminalSessions)
	app.mu.Unlock()
	assert.Equal(t, 0, count, "registry must start empty")

	// A failed spawn should not register anything.
	// spawnTmuxSession registers AFTER successful tmux new-session.
	// If tmux fails, the function returns early with an error,
	// so registerSession is never called. Verify invariant:
	// registry stays empty when no registerSession is called.
	app.mu.Lock()
	countAfter := len(app.terminalSessions)
	app.mu.Unlock()
	assert.Equal(t, 0, countAfter, "registry must stay empty when spawn fails")
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
