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

// R16 + PR 2 security review: the review agent gets no shell at all. Claude
// Code auto-approves read-only-looking commands such as `git log`, which
// accept --output=<file>, so any Bash access lets it write arbitrary files.
// The diff is fetched by the app and handed over as a file to Read.
func TestSpawnPRReview_ArgvExact(t *testing.T) {
	app, fake, repo := spawnFixture(t)
	app.prNumber = func(context.Context, string) (string, error) { return "12", nil }
	app.prDiff = func(_ context.Context, _ string, n string) (string, error) {
		return "diff --git a/x b/x\n+added line for PR " + n + "\n", nil
	}

	_, err := app.SpawnPRReview(repo)
	require.NoError(t, err)

	argvs := fake.spawnedArgvs()
	require.Len(t, argvs, 1)
	argv := argvs[0]
	model := domain.DefaultAlias(app.ListModels())
	require.Len(t, argv, 9)
	assert.Equal(t, []string{"claude", "--model", model, "--allowedTools", "Read,Grep,Glob", "--disallowedTools", "Bash", "-p"}, argv[:8])
	prompt := argv[8]
	assert.Contains(t, prompt, "Review PR #12")
	assert.Contains(t, prompt, "\n", "multi-line prompt stays one element")
	assert.NotContains(t, strings.Join(argv, " "), "--dangerously-skip-permissions")

	// The prompt names a private temp file holding the diff.
	var diffPath string
	for _, f := range strings.Fields(prompt) {
		if strings.HasSuffix(f, ".diff") {
			diffPath = f
		}
	}
	require.NotEmpty(t, diffPath, "prompt must name the diff file")
	b, err := os.ReadFile(diffPath)
	require.NoError(t, err)
	assert.Contains(t, string(b), "+added line for PR 12")
	fi, err := os.Stat(diffPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), fi.Mode().Perm())
	t.Cleanup(func() { os.Remove(diffPath) })
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
