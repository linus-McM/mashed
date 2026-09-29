package main

import (
	"os"
	"path/filepath"
	"testing"
)

// lfsPointer is what a clone without git-lfs has in place of a binary.
const lfsPointer = "version https://git-lfs.github.com/spec/v1\noid sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef\nsize 1234\n"

// R28: ListLocalFonts skips LFS pointer files instead of registering them as
// fonts (which would inject a broken @font-face).
func TestListLocalFonts_SkipsLFSPointer(t *testing.T) {
	dir := t.TempDir()
	fonts := filepath.Join(dir, "fonts")
	if err := os.MkdirAll(fonts, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fonts, "PointerMono-Regular.ttf"), []byte(lfsPointer), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fonts, "RealMono-Regular.ttf"), []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x10}, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)

	fams := (&App{}).ListLocalFonts()
	var names []string
	for _, f := range fams {
		for _, file := range f.Files {
			names = append(names, file.FileName)
		}
	}
	for _, n := range names {
		if n == "PointerMono-Regular.ttf" {
			t.Fatalf("LFS pointer registered as a font: %v", names)
		}
	}
	if len(names) != 1 || names[0] != "RealMono-Regular.ttf" {
		t.Fatalf("fonts = %v, want only RealMono-Regular.ttf", names)
	}
}
