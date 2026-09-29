package fsutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteFileAtomic_WritesAndSetsMode(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := WriteFileAtomic(path, []byte(`{"a":1}`), 0o600); err != nil {
		t.Fatalf("WriteFileAtomic: %v", err)
	}
	b, _ := os.ReadFile(path)
	if string(b) != `{"a":1}` {
		t.Fatalf("content = %q", b)
	}
	fi, _ := os.Stat(path)
	if perm := fi.Mode().Perm(); perm != 0o600 {
		t.Fatalf("mode = %o, want 600", perm)
	}
}

func TestWriteFileAtomic_ExistingTargetKeepsMode(t *testing.T) {
	path := filepath.Join(t.TempDir(), "notes.md")
	if err := os.WriteFile(path, []byte("old"), 0o640); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o640); err != nil {
		t.Fatal(err)
	}
	if err := WriteFileAtomic(path, []byte("new"), 0o600); err != nil {
		t.Fatal(err)
	}
	fi, _ := os.Stat(path)
	if perm := fi.Mode().Perm(); perm != 0o640 {
		t.Fatalf("mode = %o, want existing 640 kept", perm)
	}
}

func TestWriteFileAtomic_FailureLeavesOriginalAndNoTemp(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("read-only dir is ineffective as root")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte("ORIGINAL"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dir, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(dir, 0o700) })

	if err := WriteFileAtomic(path, []byte("NEW"), 0o600); err == nil {
		t.Fatal("expected error writing into a read-only dir")
	}
	if b, _ := os.ReadFile(path); string(b) != "ORIGINAL" {
		t.Fatalf("original changed: %q", b)
	}
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if e.Name() != "config.json" {
			t.Fatalf("leftover file %q", e.Name())
		}
	}
}

func TestWriteFileAtomic_NoTempLeftOnSuccess(t *testing.T) {
	dir := t.TempDir()
	for i := 0; i < 3; i++ {
		if err := WriteFileAtomic(filepath.Join(dir, "x.json"), []byte("v"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.Contains(e.Name(), "tmp") || e.Name() != "x.json" {
			t.Fatalf("leftover file %q", e.Name())
		}
	}
}
