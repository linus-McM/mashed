package main

import (
	"net"
	"os"
	"path/filepath"
	"testing"
)

// R12: the helper socket lives in a private 0700 directory.
func TestSetupHelperSocketDir_Mode0700(t *testing.T) {
	dir, cleanup, err := setupHelperSocketDir()
	if err != nil {
		t.Fatalf("setupHelperSocketDir: %v", err)
	}
	defer cleanup()
	fi, err := os.Stat(dir)
	if err != nil {
		t.Fatal(err)
	}
	if perm := fi.Mode().Perm(); perm != 0o700 {
		t.Fatalf("socket dir mode = %o, want 700", perm)
	}
}

// R12: shutdown removes the socket directory.
func TestHelperShutdown_RemovesSocketDir(t *testing.T) {
	dir, cleanup, err := setupHelperSocketDir()
	if err != nil {
		t.Fatal(err)
	}
	cleanup()
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("socket dir still exists after cleanup: %v", err)
	}
}

// R12: the parent refuses to dial a socket other users can access.
func TestDialRefusesWorldAccessibleSocket(t *testing.T) {
	// Short dir: t.TempDir() paths exceed macOS's 104-byte socket path limit.
	dir, err := os.MkdirTemp("", "ms")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	path := filepath.Join(dir, "s.sock")
	ln, err := net.Listen("unix", path)
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	if err := os.Chmod(path, 0o777); err != nil {
		t.Fatal(err)
	}
	if c, err := dialHelperSecure(path); err == nil {
		c.Close()
		t.Fatal("dialHelperSecure accepted a 0777 socket")
	}

	if err := os.Chmod(path, 0o600); err != nil {
		t.Fatal(err)
	}
	c, err := dialHelperSecure(path)
	if err != nil {
		t.Fatalf("dialHelperSecure(0600) = %v, want nil", err)
	}
	c.Close()
}
