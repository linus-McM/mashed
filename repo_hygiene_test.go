package main

// Repository-configuration checks. Each TestRepo_* guards one requirement of
// sdlc/repo-health-remediation/spec.md so config changes get a red/green cycle.

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// repoRoot returns the absolute path of the repository root (this file's dir).
func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Dir(file)
}

// trackedFiles returns `git ls-files` output for the given pathspecs.
func trackedFiles(t *testing.T, pathspecs ...string) []string {
	t.Helper()
	args := append([]string{"ls-files", "--"}, pathspecs...)
	cmd := exec.Command("git", args...)
	cmd.Dir = repoRoot(t)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git ls-files: %v", err)
	}
	return strings.Fields(string(out))
}

// readRepoFile returns a repo file's contents.
func readRepoFile(t *testing.T, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(repoRoot(t), rel))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return string(b)
}

// hasLine reports whether text contains line as a whole line.
func hasLine(text, line string) bool {
	for _, l := range strings.Split(text, "\n") {
		if strings.TrimSpace(l) == line {
			return true
		}
	}
	return false
}

// R21: a fresh clone can `go build` because frontend/dist/.gitkeep is
// tracked, so `//go:embed all:frontend/dist` always matches a file.
func TestRepo_DistPlaceholderTracked(t *testing.T) {
	if got := trackedFiles(t, "frontend/dist/.gitkeep"); len(got) != 1 {
		t.Fatalf("frontend/dist/.gitkeep is not tracked (got %v)", got)
	}
	gi := readRepoFile(t, ".gitignore")
	for _, want := range []string{"frontend/dist/*", "!frontend/dist/.gitkeep"} {
		if !hasLine(gi, want) {
			t.Errorf(".gitignore lacks %q", want)
		}
	}
	if hasLine(gi, "frontend/dist") || hasLine(gi, "frontend/dist/") {
		t.Error(".gitignore still ignores the whole frontend/dist directory")
	}
}

func TestRepo_GraphifyOutIgnored(t *testing.T) {
	cmd := exec.Command("git", "check-ignore", "-q", "graphify-out/x")
	cmd.Dir = repoRoot(t)
	if err := cmd.Run(); err != nil {
		t.Fatalf("graphify-out/ is not git-ignored: %v", err)
	}
}
