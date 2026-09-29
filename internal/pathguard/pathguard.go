// Package pathguard confines file paths supplied by the webview to a set of
// allowed root directories. Paths and roots are resolved with Abs, Clean and
// EvalSymlinks before comparison, so ".." traversal and symlink escapes are
// rejected (spec R7).
package pathguard

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ErrOutsideRoot is returned when a path resolves outside every allowed root.
var ErrOutsideRoot = errors.New("path outside allowed roots")

// ErrDeniedPath is returned when a write targets a persistence-sensitive file
// under $HOME (shell rc files, ~/.ssh, LaunchAgents, ~/.gitconfig).
var ErrDeniedPath = errors.New("path is write-protected")

// HomeRoot returns the user's home directory.
func HomeRoot() (string, error) {
	return os.UserHomeDir()
}

// AllowedRoots returns $HOME plus devDir (when set and different), cleaned.
func AllowedRoots(devDir string) []string {
	var roots []string
	if home, err := HomeRoot(); err == nil && home != "" {
		roots = append(roots, filepath.Clean(home))
	}
	if devDir != "" {
		d := filepath.Clean(devDir)
		if len(roots) == 0 || roots[0] != d {
			roots = append(roots, d)
		}
	}
	return roots
}

// ResolveExisting resolves an existing path and returns it if it lies inside
// one of roots after symlink resolution.
func ResolveExisting(roots []string, p string) (string, error) {
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", fmt.Errorf("resolve %q: %w", p, err)
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", fmt.Errorf("resolve %q: %w", p, err)
	}
	if !within(roots, resolved) {
		return "", fmt.Errorf("%q: %w", p, ErrOutsideRoot)
	}
	return resolved, nil
}

// ResolveForWrite resolves a path that may not exist yet. The parent
// directory must exist and resolve inside a root; an existing target is
// resolved through symlinks (so writing a symlinked file writes its target)
// and must also lie inside a root.
func ResolveForWrite(roots []string, p string) (string, error) {
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", fmt.Errorf("resolve %q: %w", p, err)
	}
	if _, err := os.Lstat(abs); err == nil {
		return ResolveExisting(roots, abs)
	}
	parent, err := filepath.EvalSymlinks(filepath.Dir(abs))
	if err != nil {
		return "", fmt.Errorf("resolve parent of %q: %w", p, err)
	}
	target := filepath.Join(parent, filepath.Base(abs))
	if !within(roots, target) {
		return "", fmt.Errorf("%q: %w", p, ErrOutsideRoot)
	}
	return target, nil
}

// deniedFiles and deniedDirs are relative to $HOME: shell startup files, git
// and tmux config, SSH keys and LaunchAgents all give persistent code
// execution if written.
var (
	deniedFiles = []string{
		".zshrc", ".zprofile", ".zshenv", ".zlogin", ".zlogout",
		".bashrc", ".bash_profile", ".bash_login", ".profile",
		".gitconfig", filepath.Join(".config", "git", "config"), ".tmux.conf",
	}
	deniedDirs = []string{".ssh", filepath.Join("Library", "LaunchAgents")}
)

// CheckWriteDenylist rejects writes to persistence-sensitive locations under
// home. resolved should already be symlink-resolved. Names are compared
// case-insensitively (macOS APFS maps ~/.ZSHRC onto ~/.zshrc) and, where the
// target or a parent exists, by file identity as well.
func CheckWriteDenylist(home, resolved string) error {
	denied := fmt.Errorf("%q: %w", resolved, ErrDeniedPath)
	r := filepath.Clean(resolved)
	for _, base := range []string{resolveRoot(home), filepath.Clean(home)} {
		rel, err := filepath.Rel(base, r)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
			continue
		}
		for _, f := range deniedFiles {
			if strings.EqualFold(rel, f) {
				return denied
			}
		}
		for _, d := range deniedDirs {
			if strings.EqualFold(rel, d) || hasFoldPrefix(rel, d+string(os.PathSeparator)) {
				return denied
			}
		}
	}

	// Identity checks catch any spelling the name checks miss (other
	// normalisations, hard-to-predict aliases).
	if fi, err := os.Stat(r); err == nil {
		for _, f := range deniedFiles {
			if di, err := os.Stat(filepath.Join(home, f)); err == nil && os.SameFile(fi, di) {
				return denied
			}
		}
	}
	for _, d := range deniedDirs {
		di, err := os.Stat(filepath.Join(home, d))
		if err != nil {
			continue
		}
		for p := filepath.Dir(r); ; p = filepath.Dir(p) {
			if pi, err := os.Stat(p); err == nil && os.SameFile(pi, di) {
				return denied
			}
			if parent := filepath.Dir(p); parent == p {
				break
			}
		}
	}
	return nil
}

// hasFoldPrefix reports whether s starts with prefix, ignoring case.
func hasFoldPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && strings.EqualFold(s[:len(prefix)], prefix)
}

// within reports whether p equals or lies under any resolved root.
func within(roots []string, p string) bool {
	for _, root := range roots {
		if root == "" {
			continue
		}
		r := resolveRoot(root)
		if p == r || strings.HasPrefix(p, r+string(os.PathSeparator)) {
			return true
		}
	}
	return false
}

// resolveRoot returns the symlink-resolved form of root, or its cleaned
// absolute form when it cannot be resolved (for example, it does not exist).
func resolveRoot(root string) string {
	abs, err := filepath.Abs(root)
	if err != nil {
		return filepath.Clean(root)
	}
	if r, err := filepath.EvalSymlinks(abs); err == nil {
		return r
	}
	return abs
}
