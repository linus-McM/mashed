package scanner

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"conductor/internal/domain"
)

var (
	psLineRegex  = regexp.MustCompile(`^\s*(\d+)\s+(\d+)\s+([\d:.,-]+)\s+(.+)$`)
	modelRegex   = regexp.MustCompile(`--model\s+(\S+)`)
	sessionRegex = regexp.MustCompile(`--session-id\s+(\S+)`)

	excludedPatterns = []string{
		"grep",
		"Claude Helper",
		"Claude.app/Contents/MacOS",
		"context-mode",
		"/bin/zsh",
	}
)

// rawProcess holds intermediate parsed data from a ps line.
type rawProcess struct {
	pid       int
	ppid      int
	model     string
	sessionID string
	startedAt time.Time
}

// ScanProcesses discovers running Claude CLI processes and resolves their working directories.
func (p *ClaudeCodeProvider) ScanProcesses() ([]domain.AgentSession, error) {
	out, err := exec.Command("ps", "-eo", "pid,ppid,etime,args").Output()
	if err != nil {
		return nil, &ScanError{Op: "ps", Err: fmt.Errorf("executing ps: %w", err)}
	}

	var candidates []rawProcess
	sc := bufio.NewScanner(bytes.NewReader(out))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if !strings.Contains(line, "claude") {
			continue
		}
		if isExcludedProcess(line) {
			continue
		}
		rp, err := parsePSLine(line)
		if err != nil {
			continue
		}
		candidates = append(candidates, rp)
	}

	if len(candidates) == 0 {
		return nil, nil
	}

	// Parallel working directory lookups via goroutines.
	type result struct {
		idx     int
		session domain.AgentSession
		dir     string
		err     error
	}

	results := make([]result, len(candidates))
	var wg sync.WaitGroup

	for i, c := range candidates {
		wg.Add(1)
		go func(idx int, proc rawProcess) {
			defer wg.Done()
			dir, err := p.GetWorkingDir(proc.pid)
			results[idx] = result{
				idx: idx,
				session: domain.AgentSession{
					PID:       proc.pid,
					PPID:      proc.ppid,
					Model:     proc.model,
					StartedAt: proc.startedAt,
					SessionID: proc.sessionID,
				},
				dir: dir,
				err: err,
			}
		}(i, c)
	}
	wg.Wait()

	sessions := make([]domain.AgentSession, 0, len(results))
	for _, r := range results {
		if r.err != nil {
			continue // skip processes we can't resolve
		}
		p.cachePidDir(r.session.PID, r.dir)
		sessions = append(sessions, r.session)
	}

	return sessions, nil
}

// GetWorkingDir returns the working directory for a process.
// On macOS it uses lsof -Fn; on Linux it reads /proc/{pid}/cwd.
func (p *ClaudeCodeProvider) GetWorkingDir(pid int) (string, error) {
	if dir, ok := p.getCachedPidDir(pid); ok {
		return dir, nil
	}

	var dir string
	var err error

	switch runtime.GOOS {
	case "darwin":
		dir, err = getWorkingDirDarwin(pid)
	case "linux":
		dir, err = getWorkingDirLinux(pid)
	default:
		return "", &ScanError{
			PID: pid, Op: "cwd",
			Err: fmt.Errorf("unsupported platform %s: %w", runtime.GOOS, ErrNoWorkingDir),
		}
	}

	if err != nil {
		return "", &ScanError{PID: pid, Op: "cwd", Err: err}
	}

	p.cachePidDir(pid, dir)
	return dir, nil
}

// getWorkingDirDarwin uses lsof -Fn to find the cwd for a process on macOS.
// Output format: lines starting with 'f' (fd type) and 'n' (name).
// We look for "fcwd" followed by "n/path/...".
func getWorkingDirDarwin(pid int) (string, error) {
	out, err := exec.Command("lsof", "-Fn", "-p", strconv.Itoa(pid)).Output()
	if err != nil {
		return "", fmt.Errorf("lsof pid %d: %w", pid, ErrNoWorkingDir)
	}

	lines := strings.Split(string(out), "\n")
	foundCwd := false
	for _, line := range lines {
		if line == "fcwd" {
			foundCwd = true
			continue
		}
		if foundCwd && strings.HasPrefix(line, "n") {
			return line[1:], nil
		}
		if foundCwd {
			foundCwd = false
		}
	}

	return "", fmt.Errorf("no cwd in lsof output for pid %d: %w", pid, ErrNoWorkingDir)
}

// getWorkingDirLinux reads the /proc/{pid}/cwd symlink on Linux.
func getWorkingDirLinux(pid int) (string, error) {
	link := fmt.Sprintf("/proc/%d/cwd", pid)
	dir, err := os.Readlink(link)
	if err != nil {
		return "", fmt.Errorf("readlink %s: %w", link, ErrNoWorkingDir)
	}
	return dir, nil
}

// parsePSLine extracts process info from a single ps output line.
func parsePSLine(line string) (rawProcess, error) {
	matches := psLineRegex.FindStringSubmatch(line)
	if matches == nil {
		return rawProcess{}, fmt.Errorf("no regex match: %w", ErrProcessNotFound)
	}

	pid, err := strconv.Atoi(matches[1])
	if err != nil {
		return rawProcess{}, fmt.Errorf("parsing pid %q: %w", matches[1], err)
	}

	ppid, err := strconv.Atoi(matches[2])
	if err != nil {
		return rawProcess{}, fmt.Errorf("parsing ppid %q: %w", matches[2], err)
	}

	elapsed, err := parseEtime(matches[3])
	if err != nil {
		return rawProcess{}, fmt.Errorf("parsing etime %q: %w", matches[3], err)
	}

	args := matches[4]

	return rawProcess{
		pid:       pid,
		ppid:      ppid,
		model:     extractModel(args),
		sessionID: extractSessionID(args),
		startedAt: time.Now().Add(-elapsed),
	}, nil
}

// parseEtime converts ps etime format [[dd-]hh:]mm:ss to time.Duration.
func parseEtime(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	var days, hours, mins, secs int

	// Check for days component: dd-hh:mm:ss
	if idx := strings.Index(s, "-"); idx >= 0 {
		d, err := strconv.Atoi(s[:idx])
		if err != nil {
			return 0, fmt.Errorf("parsing days %q: %w", s[:idx], err)
		}
		days = d
		s = s[idx+1:]
	}

	parts := strings.Split(s, ":")
	switch len(parts) {
	case 3: // hh:mm:ss
		h, err := strconv.Atoi(parts[0])
		if err != nil {
			return 0, err
		}
		hours = h
		m, err := strconv.Atoi(parts[1])
		if err != nil {
			return 0, err
		}
		mins = m
		sec, err := strconv.Atoi(parts[2])
		if err != nil {
			return 0, err
		}
		secs = sec
	case 2: // mm:ss
		m, err := strconv.Atoi(parts[0])
		if err != nil {
			return 0, err
		}
		mins = m
		sec, err := strconv.Atoi(parts[1])
		if err != nil {
			return 0, err
		}
		secs = sec
	case 1: // ss
		sec, err := strconv.Atoi(parts[0])
		if err != nil {
			return 0, err
		}
		secs = sec
	default:
		return 0, fmt.Errorf("unexpected etime format: %s", s)
	}

	return time.Duration(days)*24*time.Hour +
		time.Duration(hours)*time.Hour +
		time.Duration(mins)*time.Minute +
		time.Duration(secs)*time.Second, nil
}

func extractModel(args string) string {
	m := modelRegex.FindStringSubmatch(args)
	if m != nil {
		return m[1]
	}
	return "claude"
}

func extractSessionID(args string) string {
	m := sessionRegex.FindStringSubmatch(args)
	if m != nil {
		return m[1]
	}
	return ""
}

func isExcludedProcess(line string) bool {
	for _, pattern := range excludedPatterns {
		if strings.Contains(line, pattern) {
			return true
		}
	}
	return false
}
