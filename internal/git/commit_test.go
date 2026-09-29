package git

import (
	"context"
	"strings"
	"testing"
)

func TestStageAllAndStaged(t *testing.T) {
	ctx := context.Background()
	repo := newRepo(t)

	if stat, err := StagedStat(ctx, repo); err != nil || stat != "" {
		t.Fatalf("StagedStat on clean repo = %q, %v; want empty", stat, err)
	}

	writeFile(t, repo, "a.txt", "two\n")
	writeFile(t, repo, "new.txt", "new\n")
	if out, err := StageAll(ctx, repo); err != nil {
		t.Fatalf("StageAll: %v (%s)", err, out)
	}

	stat, err := StagedStat(ctx, repo)
	if err != nil || !strings.Contains(stat, "a.txt") || !strings.Contains(stat, "new.txt") {
		t.Errorf("StagedStat = %q, %v", stat, err)
	}
	diff, err := StagedDiff(ctx, repo)
	if err != nil || !strings.Contains(diff, "+two") || !strings.Contains(diff, "+new") {
		t.Errorf("StagedDiff = %q, %v", diff, err)
	}
}

func TestStageAll_NotARepo(t *testing.T) {
	out, err := StageAll(context.Background(), t.TempDir())
	if err == nil {
		t.Fatal("StageAll on non-repo: want error")
	}
	if !strings.Contains(strings.ToLower(out), "not a git repository") {
		t.Errorf("combined output = %q; want git's message", out)
	}
}

func TestCommit(t *testing.T) {
	ctx := context.Background()
	repo := newRepo(t)
	writeFile(t, repo, "a.txt", "changed\n")
	if _, err := StageAll(ctx, repo); err != nil {
		t.Fatal(err)
	}
	if out, err := Commit(ctx, repo, "change a"); err != nil {
		t.Fatalf("Commit: %v (%s)", err, out)
	}
	if msg := strings.TrimSpace(runGit(t, repo, "log", "-1", "--format=%s")); msg != "change a" {
		t.Errorf("last commit = %q", msg)
	}

	// Nothing staged: the combined output must carry "nothing to commit",
	// which gitCommitCore relies on.
	out, err := Commit(ctx, repo, "empty")
	if err == nil || !strings.Contains(out, "nothing to commit") {
		t.Errorf("empty Commit = %v (%q); want nothing-to-commit error", err, out)
	}
}
