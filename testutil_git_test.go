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
// Test git subprocesses therefore inherit a clean os.Environ().
func TestMain(m *testing.M) {
	for _, kv := range os.Environ() {
		if name, _, ok := strings.Cut(kv, "="); ok && strings.HasPrefix(name, "GIT_") {
			os.Unsetenv(name)
		}
	}
	os.Exit(m.Run())
}
