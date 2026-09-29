package pathguard

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// layout builds: home/ (root 1), dev/ (root 2, outside home), outside/ (no root).
func layout(t *testing.T) (home, dev, outside string) {
	t.Helper()
	base := t.TempDir()
	home = filepath.Join(base, "home")
	dev = filepath.Join(base, "dev")
	outside = filepath.Join(base, "outside")
	for _, d := range []string{home, dev, outside, filepath.Join(home, "proj"), filepath.Join(home, ".ssh"), filepath.Join(home, "Library", "LaunchAgents")} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	write := func(p string) {
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(filepath.Join(home, "proj", "a.txt"))
	write(filepath.Join(dev, "b.txt"))
	write(filepath.Join(outside, "secret.txt"))
	// sibling whose name has home as a string prefix
	write(home + "x")
	// symlink inside home pointing outside every root
	if err := os.Symlink(filepath.Join(outside, "secret.txt"), filepath.Join(home, "proj", "escape.txt")); err != nil {
		t.Fatal(err)
	}
	// symlinked directory inside home pointing outside every root
	if err := os.Symlink(outside, filepath.Join(home, "proj", "outdir")); err != nil {
		t.Fatal(err)
	}
	// symlink inside home pointing at another file inside home
	if err := os.Symlink(filepath.Join(home, "proj", "a.txt"), filepath.Join(home, "proj", "link.txt")); err != nil {
		t.Fatal(err)
	}
	return home, dev, outside
}

func TestAllowedRoots(t *testing.T) {
	home, dev, _ := layout(t)
	t.Setenv("HOME", home)
	got := AllowedRoots(dev)
	if len(got) != 2 || got[0] != home || got[1] != dev {
		t.Fatalf("AllowedRoots(%q) = %v, want [%s %s]", dev, got, home, dev)
	}
	if got := AllowedRoots(""); len(got) != 1 || got[0] != home {
		t.Fatalf("AllowedRoots(\"\") = %v, want [%s]", got, home)
	}
	if got := AllowedRoots(home); len(got) != 1 {
		t.Fatalf("AllowedRoots(home) should dedupe, got %v", got)
	}
}

func TestResolveExisting(t *testing.T) {
	home, dev, outside := layout(t)
	roots := []string{home, dev}
	cases := []struct {
		name    string
		path    string
		wantErr error
	}{
		{"inside home", filepath.Join(home, "proj", "a.txt"), nil},
		{"inside dev root outside home", filepath.Join(dev, "b.txt"), nil},
		{"symlink to file inside root", filepath.Join(home, "proj", "link.txt"), nil},
		{"absolute outside", filepath.Join(outside, "secret.txt"), ErrOutsideRoot},
		{"dotdot traversal", filepath.Join(home, "proj", "..", "..", "outside", "secret.txt"), ErrOutsideRoot},
		{"symlink escape", filepath.Join(home, "proj", "escape.txt"), ErrOutsideRoot},
		{"through symlinked dir", filepath.Join(home, "proj", "outdir", "secret.txt"), ErrOutsideRoot},
		{"prefix confusion", home + "x", ErrOutsideRoot},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ResolveExisting(roots, tc.path)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("ResolveExisting(%q) err = %v, want %v", tc.path, err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("ResolveExisting(%q) unexpected err: %v", tc.path, err)
			}
			if !filepath.IsAbs(got) {
				t.Fatalf("resolved path %q is not absolute", got)
			}
		})
	}
}

func TestResolveExisting_MissingFile(t *testing.T) {
	home, _, _ := layout(t)
	if _, err := ResolveExisting([]string{home}, filepath.Join(home, "nope.txt")); err == nil {
		t.Fatal("missing file should error")
	}
}

func TestResolveForWrite(t *testing.T) {
	home, dev, outside := layout(t)
	roots := []string{home, dev}
	cases := []struct {
		name    string
		path    string
		wantErr error
	}{
		{"new file inside home", filepath.Join(home, "proj", "new.txt"), nil},
		{"existing file inside dev", filepath.Join(dev, "b.txt"), nil},
		{"new file outside", filepath.Join(outside, "new.txt"), ErrOutsideRoot},
		{"dotdot traversal", filepath.Join(home, "..", "outside", "new.txt"), ErrOutsideRoot},
		{"existing symlink escape", filepath.Join(home, "proj", "escape.txt"), ErrOutsideRoot},
		{"new leaf in symlinked dir", filepath.Join(home, "proj", "outdir", "new.txt"), ErrOutsideRoot},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ResolveForWrite(roots, tc.path)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("ResolveForWrite(%q) err = %v, want %v", tc.path, err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("ResolveForWrite(%q) unexpected err: %v", tc.path, err)
			}
		})
	}
}

func TestResolveForWrite_SymlinkResolvesToTarget(t *testing.T) {
	home, _, _ := layout(t)
	got, err := ResolveForWrite([]string{home}, filepath.Join(home, "proj", "link.txt"))
	if err != nil {
		t.Fatal(err)
	}
	want, _ := filepath.EvalSymlinks(filepath.Join(home, "proj", "a.txt"))
	if got != want {
		t.Fatalf("ResolveForWrite(link) = %q, want symlink target %q", got, want)
	}
}

func TestCheckWriteDenylist(t *testing.T) {
	home, _, _ := layout(t)
	denied := []string{
		".zshrc", ".zprofile", ".zshenv", ".bashrc", ".bash_profile", ".profile", ".gitconfig",
		filepath.Join(".ssh", "authorized_keys"), filepath.Join(".ssh", "config"),
		filepath.Join("Library", "LaunchAgents", "x.plist"),
	}
	for _, rel := range denied {
		p := filepath.Join(home, rel)
		if err := CheckWriteDenylist(home, p); !errors.Is(err, ErrDeniedPath) {
			t.Errorf("CheckWriteDenylist(%q) = %v, want ErrDeniedPath", rel, err)
		}
	}
	allowed := []string{filepath.Join("proj", "a.txt"), ".zshrc.bak", filepath.Join("proj", ".zshrc"), ".sshx", ".sshrc.d"}
	for _, rel := range allowed {
		if err := CheckWriteDenylist(home, filepath.Join(home, rel)); err != nil {
			t.Errorf("CheckWriteDenylist(%q) = %v, want nil", rel, err)
		}
	}
}

// Security review finding: on case-insensitive APFS a mixed-case spelling
// reaches the same protected file, so the denylist must not be bypassable by
// changing case. Also covers the extra login and tool config files.
func TestCheckWriteDenylist_CaseAndAliases(t *testing.T) {
	home, _, _ := layout(t)
	mustWrite := func(p string) {
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mustWrite(filepath.Join(home, ".zshrc"))
	mustWrite(filepath.Join(home, ".ssh", "authorized_keys"))

	denied := []string{
		".ZSHRC", ".ZshRc", ".GITCONFIG",
		filepath.Join(".SSH", "authorized_keys"), filepath.Join(".Ssh", "new_key"),
		filepath.Join("library", "launchagents", "evil.plist"),
		filepath.Join("LIBRARY", "LaunchAgents", "evil.plist"),
		".zlogin", ".zlogout", ".bash_login", ".tmux.conf",
		filepath.Join(".config", "git", "config"), filepath.Join(".CONFIG", "Git", "Config"),
	}
	for _, rel := range denied {
		if err := CheckWriteDenylist(home, filepath.Join(home, rel)); !errors.Is(err, ErrDeniedPath) {
			t.Errorf("CheckWriteDenylist(%q) = %v, want ErrDeniedPath", rel, err)
		}
	}
}
