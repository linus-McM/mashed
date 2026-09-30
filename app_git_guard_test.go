package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"mashed/internal/git"
	"mashed/internal/pathguard"
)

// gitOut runs git in dir and returns trimmed stdout.
func gitOut(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git %v: %v", args, err)
	}
	return strings.TrimSpace(string(out))
}

// dirtyRepoInHome creates HOME=<tmp>/home and a repo inside it with an
// uncommitted change, so an auto-commit would visibly move HEAD.
func dirtyRepoInHome(t *testing.T) (app *App, repo string) {
	t.Helper()
	home := filepath.Join(t.TempDir(), "home")
	repo = filepath.Join(home, "repo")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	initTestGitRepoAt(t, repo)
	mustWrite(t, filepath.Join(repo, "dirty.txt"), "uncommitted\n")
	return &App{ctx: context.Background()}, repo
}

// R10: invalid refs are rejected before any auto-commit or git process runs.
func TestGitBindings_RejectInvalidRefs(t *testing.T) {
	app, repo := dirtyRepoInHome(t)
	head := gitOut(t, repo, "rev-parse", "HEAD")
	status := gitOut(t, repo, "status", "--porcelain")

	calls := map[string]func() error{
		"switch --help":        func() error { return app.GitSwitchBranch(repo, "--help", true) },
		"switch -b":            func() error { return app.GitSwitchBranch(repo, "-b", true) },
		"switch a..b":          func() error { return app.GitSwitchBranch(repo, "a..b", true) },
		"create bad prefix":    func() error { return app.GitCreateBranch(repo, "-x", "ok", true) },
		"create bad name":      func() error { return app.GitCreateBranch(repo, "feature", "--help", true) },
		"create bad composite": func() error { return app.GitCreateBranch(repo, "feature", "x.lock", true) },
		"merge into -b":        func() error { _, err := app.GitMergeInto(repo, "-b", true); return err },
		"merge into x@{1}":     func() error { _, err := app.GitMergeInto(repo, "x@{1}", true); return err },
	}
	for name, call := range calls {
		t.Run(name, func(t *testing.T) {
			if err := call(); !errors.Is(err, git.ErrInvalidRef) {
				t.Fatalf("err = %v, want ErrInvalidRef", err)
			}
			if got := gitOut(t, repo, "rev-parse", "HEAD"); got != head {
				t.Fatalf("HEAD moved from %s to %s (auto-commit ran)", head, got)
			}
			if got := gitOut(t, repo, "status", "--porcelain"); got != status {
				t.Fatalf("working tree changed: %q -> %q", status, got)
			}
		})
	}
}

// R10: valid names still work.
func TestGitBindings_ValidRefStillWorks(t *testing.T) {
	app, repo := dirtyRepoInHome(t)
	if err := app.GitCreateBranch(repo, "feature", "foo-1", true); err != nil {
		t.Fatalf("GitCreateBranch(feature/foo-1): %v", err)
	}
	if got := gitOut(t, repo, "rev-parse", "--abbrev-ref", "HEAD"); got != "feature/foo-1" {
		t.Fatalf("current branch = %q", got)
	}
}

// R11: every git/gh binding rejects a repo outside the allowed roots.
func TestGitBindings_RejectRepoOutsideRoots(t *testing.T) {
	base := t.TempDir()
	home := filepath.Join(base, "home")
	outside := filepath.Join(base, "outside")
	for _, d := range []string{home, outside} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("HOME", home)
	initTestGitRepoAt(t, outside)
	app := &App{ctx: context.Background()}

	calls := map[string]func() error{
		"GitListBranches":    func() error { _, err := app.GitListBranches(outside); return err },
		"GitSwitchBranch":    func() error { return app.GitSwitchBranch(outside, "main", false) },
		"GitCreateBranch":    func() error { return app.GitCreateBranch(outside, "feature", "x", false) },
		"GitCommit":          func() error { _, err := app.GitCommit(outside); return err },
		"GitCommitAndPush":   func() error { _, err := app.GitCommitAndPush(outside); return err },
		"GitPush":            func() error { _, err := app.GitPush(outside); return err },
		"GitForcePush":       func() error { _, err := app.GitForcePush(outside); return err },
		"GitPull":            func() error { _, err := app.GitPull(outside); return err },
		"GitMergeInto":       func() error { _, err := app.GitMergeInto(outside, "main", false); return err },
		"GitCommitPushAndPR": func() error { _, err := app.GitCommitPushAndPR(outside); return err },
		"GetWorktrees":       func() error { _, err := app.GetWorktrees(outside); return err },
		"ListRepoFiles":      func() error { _, err := app.ListRepoFiles(outside); return err },
		"ReadFileDiff":       func() error { _, err := app.ReadFileDiff(outside, "README.md"); return err },
		"ReadFileAtHead":     func() error { _, err := app.ReadFileAtHead(outside, "README.md"); return err },
		"SpawnPRReview":      func() error { _, err := app.SpawnPRReview(outside); return err },
	}
	for name, call := range calls {
		t.Run(name, func(t *testing.T) {
			if err := call(); !errors.Is(err, pathguard.ErrOutsideRoot) {
				t.Fatalf("err = %v, want ErrOutsideRoot", err)
			}
		})
	}

	// Streaming bindings report the rejection synchronously as an error event.
	var mu sync.Mutex
	var events []string
	orig := appEmitHook
	appEmitHook = func(name string, data ...any) {
		mu.Lock()
		defer mu.Unlock()
		events = append(events, name+" "+fmt.Sprint(data...))
	}
	t.Cleanup(func() { appEmitHook = orig })

	streams := map[string]func(){
		"StreamCodeReviewSummary": func() { app.StreamCodeReviewSummary(outside, "sonnet") },
		"StreamAdvice":            func() { app.StreamAdvice(outside, "default", "sonnet") },
		"StreamScopedAdvice":      func() { app.StreamScopedAdvice(outside, "default", "sonnet", []string{"README.md"}, "") },
	}
	for name, call := range streams {
		t.Run(name, func(t *testing.T) {
			mu.Lock()
			events = nil
			mu.Unlock()
			call()
			mu.Lock()
			defer mu.Unlock()
			if len(events) != 1 || !strings.Contains(events[0], "outside allowed roots") {
				t.Fatalf("events = %v, want one synchronous error event naming the rejection", events)
			}
		})
	}
}
