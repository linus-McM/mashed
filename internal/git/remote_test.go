package git

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
)

// newRepoWithRemote returns a repo whose origin is a local bare repo, with
// main pushed and tracking origin/main.
func newRepoWithRemote(t *testing.T) (repo, remote string) {
	t.Helper()
	remote = filepath.Join(t.TempDir(), "remote.git")
	runGit(t, filepath.Dir(remote), "init", "-q", "--bare", "-b", "main", remote)
	repo = newRepo(t)
	runGit(t, repo, "remote", "add", "origin", remote)
	runGit(t, repo, "push", "-q", "-u", "origin", "main")
	return repo, remote
}

// cloneOf clones remote into a fresh temp dir with an identity configured.
func cloneOf(t *testing.T, remote string) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "clone")
	runGit(t, filepath.Dir(dir), "clone", "-q", remote, dir)
	runGit(t, dir, "config", "user.email", "other@example.com")
	runGit(t, dir, "config", "user.name", "Other")
	runGit(t, dir, "config", "commit.gpgsign", "false")
	return dir
}

func commitFile(t *testing.T, dir, name, content, msg string) {
	t.Helper()
	writeFile(t, dir, name, content)
	runGit(t, dir, "add", "-A")
	runGit(t, dir, "commit", "-q", "-m", msg)
}

func TestPushAndAheadBehind(t *testing.T) {
	ctx := context.Background()
	repo, remote := newRepoWithRemote(t)

	commitFile(t, repo, "b.txt", "b\n", "b")
	ahead, behind, err := AheadBehind(ctx, repo)
	if err != nil || ahead != 1 || behind != 0 {
		t.Fatalf("AheadBehind = %d,%d,%v; want 1,0", ahead, behind, err)
	}

	if out, err := Push(ctx, repo); err != nil {
		t.Fatalf("Push: %v (%s)", err, out)
	}
	if got := strings.TrimSpace(runGit(t, remote, "log", "-1", "--format=%s", "main")); got != "b" {
		t.Errorf("remote head = %q, want b", got)
	}
	if a, b, _ := AheadBehind(ctx, repo); a != 0 || b != 0 {
		t.Errorf("after push AheadBehind = %d,%d; want 0,0", a, b)
	}
}

func TestPush_Rejected(t *testing.T) {
	ctx := context.Background()
	repo, remote := newRepoWithRemote(t)
	other := cloneOf(t, remote)
	commitFile(t, other, "o.txt", "o\n", "other")
	runGit(t, other, "push", "-q", "origin", "main")

	commitFile(t, repo, "b.txt", "b\n", "mine")
	out, err := Push(ctx, repo)
	if err == nil {
		t.Fatal("diverged Push: want error")
	}
	// GitPush classifies conflicts from this combined output.
	if !strings.Contains(out, "rejected") && !strings.Contains(out, "fetch first") &&
		!strings.Contains(out, "non-fast-forward") {
		t.Errorf("Push output lacks rejection marker: %q", out)
	}
}

func TestForcePush(t *testing.T) {
	ctx := context.Background()
	repo, remote := newRepoWithRemote(t)
	commitFile(t, repo, "b.txt", "b\n", "b")
	runGit(t, repo, "push", "-q")
	runGit(t, repo, "commit", "-q", "--amend", "-m", "b amended")

	if out, err := ForcePush(ctx, repo); err != nil {
		t.Fatalf("ForcePush: %v (%s)", err, out)
	}
	if got := strings.TrimSpace(runGit(t, remote, "log", "-1", "--format=%s", "main")); got != "b amended" {
		t.Errorf("remote head = %q, want b amended", got)
	}

	noRemote := newRepo(t)
	if _, err := ForcePush(ctx, noRemote); err == nil {
		t.Error("ForcePush without origin: want error")
	}
}

func TestPull(t *testing.T) {
	ctx := context.Background()
	repo, remote := newRepoWithRemote(t)
	other := cloneOf(t, remote)
	commitFile(t, other, "o.txt", "o\n", "other")
	runGit(t, other, "push", "-q", "origin", "main")

	if out, err := Pull(ctx, repo); err != nil {
		t.Fatalf("Pull: %v (%s)", err, out)
	}
	if got := strings.TrimSpace(runGit(t, repo, "log", "-1", "--format=%s")); got != "other" {
		t.Errorf("head after pull = %q, want other", got)
	}

	if _, err := Pull(ctx, newRepo(t)); err == nil {
		t.Error("Pull without remote: want error")
	}
}

func TestAheadBehind_NoUpstream(t *testing.T) {
	if _, _, err := AheadBehind(context.Background(), newRepo(t)); err == nil {
		t.Error("AheadBehind without upstream: want error")
	}
}

func TestPorcelainStatus(t *testing.T) {
	ctx := context.Background()
	repo := newRepo(t)
	if st, err := PorcelainStatus(ctx, repo); err != nil || st != "" {
		t.Fatalf("clean PorcelainStatus = %q, %v", st, err)
	}
	writeFile(t, repo, "u.txt", "u\n")
	if st, err := PorcelainStatus(ctx, repo); err != nil || !strings.Contains(st, "?? u.txt") {
		t.Errorf("dirty PorcelainStatus = %q, %v", st, err)
	}
	if _, err := PorcelainStatus(ctx, t.TempDir()); err == nil {
		t.Error("PorcelainStatus on non-repo: want error")
	}
}
