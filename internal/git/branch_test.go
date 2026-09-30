package git

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestMain drops inherited GIT_* variables before any test runs. Git hooks
// export GIT_DIR / GIT_INDEX_FILE / GIT_WORK_TREE; when `go test` runs from a
// hook, every git subprocess would otherwise act on the real repository
// instead of the test's temp repo.
func TestMain(m *testing.M) {
	for _, kv := range os.Environ() {
		if name, _, ok := strings.Cut(kv, "="); ok && strings.HasPrefix(name, "GIT_") {
			os.Unsetenv(name)
		}
	}
	os.Exit(m.Run())
}

// runGit runs git in dir for test setup and fails the test on error.
func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return string(out)
}

// newRepo creates a temp repo on branch main with one commit (a.txt).
func newRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, "init", "-q", "-b", "main")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "Test")
	runGit(t, dir, "config", "commit.gpgsign", "false")
	writeFile(t, dir, "a.txt", "one\n")
	runGit(t, dir, "add", "-A")
	runGit(t, dir, "commit", "-q", "-m", "init")
	return dir
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestListBranches(t *testing.T) {
	ctx := context.Background()
	repo := newRepo(t)
	runGit(t, repo, "branch", "feature/x")

	got, err := ListBranches(ctx, repo)
	if err != nil {
		t.Fatalf("ListBranches: %v", err)
	}
	want := map[string]bool{"main": true, "feature/x": false}
	if len(got) != len(want) {
		t.Fatalf("got %+v, want %v", got, want)
	}
	for _, b := range got {
		if cur, ok := want[b.Name]; !ok || cur != b.Current {
			t.Errorf("unexpected branch %+v", b)
		}
	}

	if _, err := ListBranches(ctx, t.TempDir()); err == nil {
		t.Error("ListBranches on non-repo: want error")
	}
}

func TestCheckoutAndCurrentBranch(t *testing.T) {
	ctx := context.Background()
	repo := newRepo(t)
	runGit(t, repo, "branch", "dev")

	if out, err := Checkout(ctx, repo, "dev"); err != nil {
		t.Fatalf("Checkout: %v (%s)", err, out)
	}
	cur, err := CurrentBranch(ctx, repo)
	if err != nil || cur != "dev" {
		t.Fatalf("CurrentBranch = %q, %v; want dev", cur, err)
	}

	out, err := Checkout(ctx, repo, "missing")
	if err == nil {
		t.Fatal("Checkout of missing branch: want error")
	}
	if out == "" {
		t.Error("Checkout error should carry combined output")
	}

	if err := CheckoutBranch(ctx, repo, "main"); err != nil {
		t.Fatalf("CheckoutBranch: %v", err)
	}
	if cur, _ := CurrentBranch(ctx, repo); cur != "main" {
		t.Errorf("after CheckoutBranch on %q, want main", cur)
	}
	if err := CheckoutBranch(ctx, repo, "missing"); err == nil {
		t.Error("CheckoutBranch of missing branch: want error")
	}
}

func TestCurrentBranch_NotARepo(t *testing.T) {
	if _, err := CurrentBranch(context.Background(), t.TempDir()); err == nil {
		t.Error("CurrentBranch on non-repo: want error")
	}
}

func TestCreateBranch(t *testing.T) {
	ctx := context.Background()
	repo := newRepo(t)

	if out, err := CreateBranch(ctx, repo, "feature/new"); err != nil {
		t.Fatalf("CreateBranch: %v (%s)", err, out)
	}
	if cur, _ := CurrentBranch(ctx, repo); cur != "feature/new" {
		t.Errorf("current = %q, want feature/new", cur)
	}
	out, err := CreateBranch(ctx, repo, "feature/new")
	if err == nil || !strings.Contains(out, "already exists") {
		t.Errorf("duplicate CreateBranch = %v (%q); want already-exists error", err, out)
	}
}

func TestVerifyRef(t *testing.T) {
	ctx := context.Background()
	repo := newRepo(t)
	if err := VerifyRef(ctx, repo, "main"); err != nil {
		t.Errorf("VerifyRef(main): %v", err)
	}
	if err := VerifyRef(ctx, repo, "nope"); err == nil {
		t.Error("VerifyRef(nope): want error")
	}
}

func TestMerge(t *testing.T) {
	ctx := context.Background()
	repo := newRepo(t)
	runGit(t, repo, "checkout", "-q", "-b", "feat")
	writeFile(t, repo, "b.txt", "b\n")
	runGit(t, repo, "add", "-A")
	runGit(t, repo, "commit", "-q", "-m", "b")
	runGit(t, repo, "checkout", "-q", "main")

	if out, err := Merge(ctx, repo, "feat"); err != nil {
		t.Fatalf("Merge: %v (%s)", err, out)
	}
	if _, err := os.Stat(filepath.Join(repo, "b.txt")); err != nil {
		t.Errorf("merged file missing: %v", err)
	}
	if _, err := Merge(ctx, repo, "nope"); err == nil {
		t.Error("Merge of missing branch: want error")
	}
}

func TestMergeConflictAndAbort(t *testing.T) {
	ctx := context.Background()
	repo := newRepo(t)
	runGit(t, repo, "checkout", "-q", "-b", "feat")
	writeFile(t, repo, "a.txt", "feat\n")
	runGit(t, repo, "commit", "-q", "-am", "feat")
	runGit(t, repo, "checkout", "-q", "main")
	writeFile(t, repo, "a.txt", "main\n")
	runGit(t, repo, "commit", "-q", "-am", "main")

	out, err := Merge(ctx, repo, "feat")
	if err == nil || !strings.Contains(out, "CONFLICT") {
		t.Fatalf("Merge = %v (%q); want conflict", err, out)
	}
	if err := MergeAbort(ctx, repo); err != nil {
		t.Fatalf("MergeAbort: %v", err)
	}
	if st := runGit(t, repo, "status", "--porcelain"); st != "" {
		t.Errorf("after abort, status = %q; want clean", st)
	}
	if err := MergeAbort(ctx, repo); err == nil {
		t.Error("MergeAbort with no merge in progress: want error")
	}
}
