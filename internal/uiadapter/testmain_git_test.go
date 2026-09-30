package uiadapter

import (
	"os"
	"strings"
	"testing"
)

// TestMain drops inherited GIT_* variables before any test runs. Git hooks
// export GIT_DIR (and in worktrees GIT_DIR points at .git/worktrees/<name>),
// which makes TestCodegen_NoDrift's `git rev-parse --show-toplevel` resolve
// to the package directory and fail when `go test` runs from pre-push.
func TestMain(m *testing.M) {
	for _, kv := range os.Environ() {
		if name, _, ok := strings.Cut(kv, "="); ok && strings.HasPrefix(name, "GIT_") {
			os.Unsetenv(name)
		}
	}
	os.Exit(m.Run())
}
