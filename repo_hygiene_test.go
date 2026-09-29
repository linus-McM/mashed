package main

// Repository-configuration checks. Each TestRepo_* guards one requirement of
// sdlc/repo-health-remediation/spec.md so config changes get a red/green cycle.

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
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

// R22: npm is the only frontend package manager, pinned to Node 22, and
// every install path uses the lockfile (`npm ci`).
func TestRepo_SinglePackageManager(t *testing.T) {
	root := repoRoot(t)
	if _, err := os.Stat(filepath.Join(root, "frontend", "bun.lock")); !os.IsNotExist(err) {
		t.Error("frontend/bun.lock must be removed (npm is the package manager)")
	}
	if got := strings.TrimSpace(readRepoFile(t, ".nvmrc")); got != "22" {
		t.Errorf(".nvmrc = %q, want 22", got)
	}

	var pkg struct {
		Engines map[string]string `json:"engines"`
	}
	pkgJSON := readRepoFile(t, "frontend/package.json")
	if err := json.Unmarshal([]byte(pkgJSON), &pkg); err != nil {
		t.Fatalf("package.json: %v", err)
	}
	if got := pkg.Engines["node"]; got != ">=22 <23" {
		t.Errorf("package.json engines.node = %q, want \">=22 <23\"", got)
	}

	var wails map[string]any
	if err := json.Unmarshal([]byte(readRepoFile(t, "wails.json")), &wails); err != nil {
		t.Fatalf("wails.json: %v", err)
	}
	if got := wails["frontend:install"]; got != "npm ci" {
		t.Errorf("wails.json frontend:install = %v, want \"npm ci\"", got)
	}
	just := readRepoFile(t, "justfile")
	if strings.Contains(just, "npm install") || !strings.Contains(just, "npm ci") {
		t.Error("justfile must install with `npm ci`, never `npm install`")
	}

	// Wails tracks package.json by hash; keep it in sync so builds don't
	// dirty the tree.
	sum := md5.Sum([]byte(pkgJSON))
	if got := strings.TrimSpace(readRepoFile(t, "frontend/package.json.md5")); got != hex.EncodeToString(sum[:]) {
		t.Errorf("frontend/package.json.md5 = %s, want %x", got, sum)
	}
}

func TestRepo_GraphifyOutIgnored(t *testing.T) {
	cmd := exec.Command("git", "check-ignore", "-q", "graphify-out/x")
	cmd.Dir = repoRoot(t)
	if err := cmd.Run(); err != nil {
		t.Fatalf("graphify-out/ is not git-ignored: %v", err)
	}
}
