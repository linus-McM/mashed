package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"mashed/internal/pathguard"
)

// guardFixture sets HOME to a temp dir and returns an App whose DevDir is a
// second temp dir outside HOME, plus an "outside" dir under neither root.
func guardFixture(t *testing.T) (app *App, home, dev, outside string) {
	t.Helper()
	base := t.TempDir()
	home = filepath.Join(base, "home")
	dev = filepath.Join(base, "dev")
	outside = filepath.Join(base, "outside")
	for _, d := range []string{home, dev, outside, filepath.Join(home, "proj"), filepath.Join(home, ".ssh")} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("HOME", home)
	mustWrite(t, filepath.Join(home, "proj", "a.txt"), "inside")
	mustWrite(t, filepath.Join(dev, "b.txt"), "dev")
	mustWrite(t, filepath.Join(outside, "secret.txt"), "SECRET")
	if err := os.Symlink(filepath.Join(outside, "secret.txt"), filepath.Join(home, "proj", "escape.txt")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(home, "proj", "outdir")); err != nil {
		t.Fatal(err)
	}
	app = appWithDevDir(dev)
	return app, home, dev, outside
}

func mustWrite(t *testing.T, p, s string) {
	t.Helper()
	if err := os.WriteFile(p, []byte(s), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestFileBindings_RejectOutsideRoots(t *testing.T) {
	app, home, _, outside := guardFixture(t)
	bad := map[string]string{
		"absolute outside": filepath.Join(outside, "secret.txt"),
		"dotdot traversal": filepath.Join(home, "proj", "..", "..", "outside", "secret.txt"),
		"symlink escape":   filepath.Join(home, "proj", "escape.txt"),
		"symlinked dir":    filepath.Join(home, "proj", "outdir", "secret.txt"),
	}
	for name, p := range bad {
		t.Run("ReadFile/"+name, func(t *testing.T) {
			out, err := app.ReadFile(p)
			if !errors.Is(err, pathguard.ErrOutsideRoot) {
				t.Fatalf("ReadFile err = %v, want ErrOutsideRoot", err)
			}
			if out != "" {
				t.Fatalf("ReadFile leaked content %q", out)
			}
		})
		t.Run("ReadFileBase64/"+name, func(t *testing.T) {
			if _, err := app.ReadFileBase64(p); !errors.Is(err, pathguard.ErrOutsideRoot) {
				t.Fatalf("ReadFileBase64 err = %v, want ErrOutsideRoot", err)
			}
		})
		t.Run("WriteFile/"+name, func(t *testing.T) {
			if err := app.WriteFile(p, "pwned"); !errors.Is(err, pathguard.ErrOutsideRoot) {
				t.Fatalf("WriteFile err = %v, want ErrOutsideRoot", err)
			}
		})
	}
	// The outside file is untouched by every WriteFile attempt.
	if b, _ := os.ReadFile(filepath.Join(outside, "secret.txt")); string(b) != "SECRET" {
		t.Fatalf("outside file modified: %q", b)
	}
	// A new leaf in a symlinked dir that points outside must not be created.
	if err := app.WriteFile(filepath.Join(home, "proj", "outdir", "new.txt"), "x"); !errors.Is(err, pathguard.ErrOutsideRoot) {
		t.Fatalf("WriteFile new leaf err = %v, want ErrOutsideRoot", err)
	}
	if _, err := os.Stat(filepath.Join(outside, "new.txt")); !os.IsNotExist(err) {
		t.Fatal("new leaf was created outside the roots")
	}
}

func TestWriteFile_RejectsDenylist(t *testing.T) {
	app, home, _, _ := guardFixture(t)
	for _, rel := range []string{".zshrc", ".gitconfig", filepath.Join(".ssh", "authorized_keys")} {
		if err := app.WriteFile(filepath.Join(home, rel), "x"); !errors.Is(err, pathguard.ErrDeniedPath) {
			t.Errorf("WriteFile(%s) err = %v, want ErrDeniedPath", rel, err)
		}
		if _, err := os.Stat(filepath.Join(home, rel)); !os.IsNotExist(err) {
			t.Errorf("%s was created", rel)
		}
	}
}

// Security review finding: WriteFile must reject mixed-case spellings of the
// protected files and directories (case-insensitive APFS maps them to the
// real targets) and must leave the real files untouched.
func TestWriteFile_RejectsDenylistCaseVariants(t *testing.T) {
	app, home, _, _ := guardFixture(t)
	mustWrite(t, filepath.Join(home, ".zshrc"), "ORIGINAL-RC")
	if err := os.MkdirAll(filepath.Join(home, "Library", "LaunchAgents"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, rel := range []string{
		".ZSHRC",
		filepath.Join(".SSH", "authorized_keys"),
		filepath.Join("library", "launchagents", "evil.plist"),
	} {
		if err := app.WriteFile(filepath.Join(home, rel), "pwned"); !errors.Is(err, pathguard.ErrDeniedPath) {
			t.Errorf("WriteFile(%s) err = %v, want ErrDeniedPath", rel, err)
		}
	}
	if b, _ := os.ReadFile(filepath.Join(home, ".zshrc")); string(b) != "ORIGINAL-RC" {
		t.Fatalf(".zshrc modified: %q", b)
	}
	if _, err := os.Stat(filepath.Join(home, "Library", "LaunchAgents", "evil.plist")); !os.IsNotExist(err) {
		t.Fatal("LaunchAgent plist was written")
	}
}

func TestFileBindings_AllowInsideRoots(t *testing.T) {
	app, home, dev, _ := guardFixture(t)
	if s, err := app.ReadFile(filepath.Join(home, "proj", "a.txt")); err != nil || s != "inside" {
		t.Fatalf("ReadFile inside home = %q, %v", s, err)
	}
	if s, err := app.ReadFile(filepath.Join(dev, "b.txt")); err != nil || s != "dev" {
		t.Fatalf("ReadFile inside DevDir = %q, %v", s, err)
	}
	newFile := filepath.Join(home, "proj", "new.txt")
	if err := app.WriteFile(newFile, "created"); err != nil {
		t.Fatalf("WriteFile new file inside home: %v", err)
	}
	if b, _ := os.ReadFile(newFile); string(b) != "created" {
		t.Fatalf("new file content = %q", b)
	}
	if _, err := app.ReadFileBase64(filepath.Join(dev, "b.txt")); err != nil {
		t.Fatalf("ReadFileBase64 inside DevDir: %v", err)
	}
}

func TestWriteFile_SymlinkInsideRootPreservesLink(t *testing.T) {
	app, home, _, _ := guardFixture(t)
	target := filepath.Join(home, "proj", "a.txt")
	link := filepath.Join(home, "proj", "link.txt")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	if err := app.WriteFile(link, "via link"); err != nil {
		t.Fatalf("WriteFile(link): %v", err)
	}
	fi, err := os.Lstat(link)
	if err != nil || fi.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("link is no longer a symlink (mode %v, err %v)", fi.Mode(), err)
	}
	if b, _ := os.ReadFile(target); string(b) != "via link" {
		t.Fatalf("target content = %q, want %q", b, "via link")
	}
}
