package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mashed/internal/pathguard"
)

// R9: ReadFileDiff and ReadFileAtHead confine filePath to the repo.

func TestReadFileDiff_RejectsTraversal(t *testing.T) {
	base := t.TempDir()
	repo := filepath.Join(base, "repo")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatal(err)
	}
	repo = initTestGitRepoAt(t, repo)
	mustWrite(t, filepath.Join(base, "outside.txt"), "TOP-SECRET-CONTENT")

	app := appWithDevDirCtx(repo) // repo is an allowed root
	out, err := app.ReadFileDiff(repo, "../outside.txt")
	if !errors.Is(err, pathguard.ErrOutsideRoot) {
		t.Fatalf("ReadFileDiff err = %v, want ErrOutsideRoot", err)
	}
	if strings.Contains(out, "TOP-SECRET-CONTENT") {
		t.Fatal("ReadFileDiff leaked a file outside the repo")
	}
}

func TestReadFileAtHead_RejectsAbsolute(t *testing.T) {
	repo := initTestGitRepo(t)
	app := appWithDevDirCtx(repo) // repo is an allowed root
	for _, p := range []string{"/etc/passwd", "../x", "-p"} {
		if _, err := app.ReadFileAtHead(repo, p); !errors.Is(err, pathguard.ErrOutsideRoot) {
			t.Errorf("ReadFileAtHead(%q) err = %v, want ErrOutsideRoot", p, err)
		}
	}
}

func TestReadFileAtHead_DeletedTrackedFileStillReadable(t *testing.T) {
	repo := initTestGitRepo(t)
	commitFile(t, repo, "gone.txt", "was here\n")
	if err := os.Remove(filepath.Join(repo, "gone.txt")); err != nil {
		t.Fatal(err)
	}
	app := appWithDevDirCtx(repo) // repo is an allowed root
	out, err := app.ReadFileAtHead(repo, "gone.txt")
	if err != nil || out != "was here\n" {
		t.Fatalf("ReadFileAtHead(deleted) = %q, %v", out, err)
	}
}

// Regression guard (already passes before R9): a dash-prefixed path is never
// parsed as a git option.
func TestReadFileDiff_DashPathCreatesNoFile(t *testing.T) {
	repo := initTestGitRepo(t)
	target := filepath.Join(t.TempDir(), "x")
	app := appWithDevDirCtx(repo) // repo is an allowed root
	_, _ = app.ReadFileDiff(repo, "--output="+target)
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatal("dash-prefixed filePath was treated as a git option")
	}
}

// initTestGitRepoAt is initTestGitRepo for a caller-chosen directory.
func initTestGitRepoAt(t *testing.T, dir string) string {
	t.Helper()
	gitRun(t, dir, "init")
	gitRun(t, dir, "config", "user.email", "test@test.com")
	gitRun(t, dir, "config", "user.name", "Test")
	mustWrite(t, filepath.Join(dir, "README.md"), "# Test\n")
	gitRun(t, dir, "add", ".")
	gitRun(t, dir, "commit", "-m", "initial")
	return dir
}
