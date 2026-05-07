package terminal

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// loginPathOnce captures the login-shell PATH on first call, then memoises it.
// Needed because mashed launched from Finder inherits the stripped GUI PATH
// (basically /usr/bin:/bin:/usr/sbin:/sbin), which doesn't include user bins
// like ~/.local/bin or /opt/homebrew/bin where `claude` typically lives.
var (
	loginPathOnce sync.Once
	loginPath     string
)

// loginShellPATH returns the PATH the user's login shell would expose, or ""
// if it can't be determined. Cached after first invocation.
func loginShellPATH() string {
	loginPathOnce.Do(func() {
		shell := os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/zsh"
		}
		// `-l` runs login shell (sources .zprofile / .bash_profile),
		// `-i` makes it interactive (sources .zshrc / .bashrc).
		// Together they reproduce a normal terminal session's PATH.
		out, err := exec.Command(shell, "-lic", "echo -n $PATH").Output()
		if err != nil {
			return
		}
		loginPath = strings.TrimSpace(string(out))
	})
	return loginPath
}

// resolveExecutable returns the absolute path of name. If name already contains
// a path separator it is returned as-is. Lookup falls back from login-shell
// PATH → exec.LookPath → original name. Needed because the PTY helper does
// exec.Command(name, ...) which resolves against the helper process's own
// PATH (which on Finder-launched apps is the stripped GUI PATH and won't
// include ~/.local/bin or /opt/homebrew/bin where `claude` typically lives).
func resolveExecutable(name string) string {
	if name == "" || strings.ContainsRune(name, os.PathSeparator) {
		return name
	}
	if lp := loginShellPATH(); lp != "" {
		for _, dir := range strings.Split(lp, ":") {
			if dir == "" {
				continue
			}
			candidate := filepath.Join(dir, name)
			if fi, err := os.Stat(candidate); err == nil && !fi.IsDir() && fi.Mode()&0o111 != 0 {
				return candidate
			}
		}
	}
	if abs, err := exec.LookPath(name); err == nil {
		return abs
	}
	return name
}

// applyLoginPATH returns env with PATH replaced by the login-shell PATH if one
// could be determined. Otherwise env is returned unchanged.
func applyLoginPATH(env []string) []string {
	lp := loginShellPATH()
	if lp == "" {
		return env
	}
	for i, e := range env {
		if strings.HasPrefix(e, "PATH=") {
			env[i] = "PATH=" + lp
			return env
		}
	}
	return append(env, "PATH="+lp)
}
