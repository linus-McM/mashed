package main

import (
	"os"
	"strings"
	"testing"
)

// TestMain drops inherited GIT_* variables before any test runs. Git hooks
// export GIT_DIR / GIT_INDEX_FILE / GIT_WORK_TREE; when `go test` runs from a
// hook, every git subprocess (test helpers and production code alike) would
// otherwise operate on the real repository instead of the test's temp repo.
func TestMain(m *testing.M) {
	for _, kv := range os.Environ() {
		if name, _, ok := strings.Cut(kv, "="); ok && strings.HasPrefix(name, "GIT_") {
			os.Unsetenv(name)
		}
	}
	os.Exit(m.Run())
}

// cleanGitEnv returns the process environment without GIT_* variables, for
// git subprocesses started by tests.
func cleanGitEnv() []string {
	env := os.Environ()
	out := make([]string, 0, len(env))
	for _, kv := range env {
		if strings.HasPrefix(kv, "GIT_") {
			continue
		}
		out = append(out, kv)
	}
	return out
}
