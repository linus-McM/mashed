package git

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestListFiles(t *testing.T) {
	ctx := context.Background()
	repo := newRepo(t)
	writeFile(t, repo, ".gitignore", "ignored.txt\n")
	writeFile(t, repo, "ignored.txt", "x\n")
	writeFile(t, repo, "c.txt", "c\n")

	got, err := ListFiles(ctx, repo)
	if err != nil {
		t.Fatalf("ListFiles: %v", err)
	}
	want := []string{".gitignore", "a.txt", "c.txt"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ListFiles = %v, want %v", got, want)
	}

	if _, err := ListFiles(ctx, t.TempDir()); err == nil {
		t.Error("ListFiles on non-repo: want error")
	}
}

func TestFileDiff(t *testing.T) {
	ctx := context.Background()
	repo := newRepo(t)
	writeFile(t, repo, "a.txt", "changed\n")

	out, err := FileDiff(ctx, repo, "a.txt")
	if err != nil || !strings.Contains(out, "+changed") || !strings.Contains(out, "-one") {
		t.Errorf("FileDiff = %q, %v", out, err)
	}

	if _, err := FileDiff(ctx, t.TempDir(), "a.txt"); err == nil {
		t.Error("FileDiff on non-repo: want error")
	}
}

func TestNoIndexDiff(t *testing.T) {
	ctx := context.Background()
	repo := newRepo(t)
	writeFile(t, repo, "untracked.txt", "fresh\n")

	// git exits 1 when the files differ; the output is what matters.
	out, _ := NoIndexDiff(ctx, repo, filepath.Join(repo, "untracked.txt"))
	if !strings.Contains(out, "+fresh") {
		t.Errorf("NoIndexDiff output = %q", out)
	}

	out, err := NoIndexDiff(ctx, repo, filepath.Join(repo, "missing.txt"))
	if err == nil || out != "" {
		t.Errorf("NoIndexDiff of missing file = %q, %v; want error and no output", out, err)
	}
}

func TestShowAtHead(t *testing.T) {
	ctx := context.Background()
	repo := newRepo(t)
	if err := os.MkdirAll(filepath.Join(repo, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	commitFile(t, repo, filepath.Join("sub", "n.txt"), "nested\n", "nested")
	writeFile(t, repo, "a.txt", "worktree edit\n")

	if out, err := ShowAtHead(ctx, repo, "a.txt"); err != nil || out != "one\n" {
		t.Errorf("ShowAtHead(a.txt) = %q, %v; want committed content", out, err)
	}
	if out, err := ShowAtHead(ctx, repo, filepath.Join("sub", "n.txt")); err != nil || out != "nested\n" {
		t.Errorf("ShowAtHead(sub/n.txt) = %q, %v", out, err)
	}
	if _, err := ShowAtHead(ctx, repo, "missing.txt"); err == nil {
		t.Error("ShowAtHead(missing): want error")
	}
}

func TestBranchDiff(t *testing.T) {
	ctx := context.Background()
	repo := newRepo(t)
	runGit(t, repo, "checkout", "-q", "-b", "feat")
	commitFile(t, repo, "f.txt", "feature\n", "feat")

	out, err := BranchDiff(ctx, repo, "main")
	if err != nil || !strings.Contains(out, "+feature") {
		t.Errorf("BranchDiff = %q, %v", out, err)
	}
	if _, err := BranchDiff(ctx, repo, "nope"); err == nil {
		t.Error("BranchDiff against missing base: want error")
	}
}
