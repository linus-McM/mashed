package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mashed/internal/domain"
)

// spawnFixture returns an App with a recording session manager, a repo
// inside a temp HOME, a primed model cache and silenced event emission.
func spawnFixture(t *testing.T) (*App, *fakeSessionManager, string) {
	t.Helper()
	home := filepath.Join(t.TempDir(), "home")
	repo := filepath.Join(home, "repo")
	require.NoError(t, os.MkdirAll(repo, 0o755))
	t.Setenv("HOME", home)

	modelCacheOnce.Do(func() { modelCache = nil }) // never shell out to claude
	orig := appEmitHook
	appEmitHook = func(string, ...any) {}
	t.Cleanup(func() { appEmitHook = orig })

	fake := newFakeManager()
	app := &App{
		ctx:              context.Background(),
		manager:          fake,
		terminalSessions: map[string]domain.TerminalSession{},
	}
	return app, fake, repo
}

const wantReviewAllowedTools = "Read,Grep,Glob,Bash(gh pr diff *),Bash(gh pr view *),Bash(git diff *),Bash(git log *),Bash(git show *)"

// R16: the review prompt reaches claude as exactly one argv element, and the
// spawn no longer skips permissions.
func TestSpawnPRReview_ArgvExact(t *testing.T) {
	app, fake, repo := spawnFixture(t)
	app.prNumber = func(context.Context, string) (string, error) { return "12", nil }

	_, err := app.SpawnPRReview(repo)
	require.NoError(t, err)

	argvs := fake.spawnedArgvs()
	require.Len(t, argvs, 1)
	argv := argvs[0]
	model := domain.DefaultAlias(app.ListModels())
	require.Len(t, argv, 7)
	assert.Equal(t, []string{"claude", "--model", model, "--allowedTools", wantReviewAllowedTools, "-p"}, argv[:6])
	prompt := argv[6]
	assert.Contains(t, prompt, "Review PR #12")
	assert.Contains(t, prompt, "gh pr diff 12")
	assert.Contains(t, prompt, "\n", "multi-line prompt stays one element")
	assert.NotContains(t, strings.Join(argv, " "), "--dangerously-skip-permissions")
}

// R16: a PR number that is not all digits is rejected before any spawn.
func TestSpawnPRReview_RejectsNonDigitPRNumber(t *testing.T) {
	for _, n := range []string{"12; rm -rf /", "--help", "", "1 2", "0x1"} {
		app, fake, repo := spawnFixture(t)
		pr := n
		app.prNumber = func(context.Context, string) (string, error) { return pr, nil }
		_, err := app.SpawnPRReview(repo)
		assert.Error(t, err, "PR number %q", n)
		assert.Empty(t, fake.spawnedArgvs(), "PR number %q must not spawn", n)
	}
}
