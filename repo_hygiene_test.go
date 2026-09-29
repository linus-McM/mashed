package main

// Repository-configuration checks. Each TestRepo_* guards one requirement of
// sdlc/repo-health-remediation/spec.md so config changes get a red/green cycle.

import (
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

func TestRepo_GraphifyOutIgnored(t *testing.T) {
	cmd := exec.Command("git", "check-ignore", "-q", "graphify-out/x")
	cmd.Dir = repoRoot(t)
	if err := cmd.Run(); err != nil {
		t.Fatalf("graphify-out/ is not git-ignored: %v", err)
	}
}
