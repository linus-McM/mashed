package main

import (
	"os"
	"path/filepath"
	"testing"
)

// R12: the helper socket is created owner-only (0600), regardless of umask.
func TestListenSocket_Mode0600(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pty.sock")
	ln, err := listenSocket(path)
	if err != nil {
		t.Fatalf("listenSocket: %v", err)
	}
	defer ln.Close()

	fi, err := os.Lstat(path)
	if err != nil {
		t.Fatal(err)
	}
	if perm := fi.Mode().Perm(); perm != 0o600 {
		t.Fatalf("socket mode = %o, want 600", perm)
	}
}
