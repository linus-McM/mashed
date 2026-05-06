package scanner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"mashed/internal/domain"
)

// RepoScanner discovers git repositories and merges them with agent session data.
type RepoScanner struct {
	devDir string

	mu        sync.RWMutex
	repoCache map[string]*repoCacheEntry
	cacheTTL  time.Duration
}

type repoCacheEntry struct {
	info      domain.RepoInfo
	fetchedAt time.Time
}

// NewRepoScanner creates a RepoScanner for the given development directory.
func NewRepoScanner(devDir string) *RepoScanner {
	if devDir == "" {
		devDir = os.Getenv("MASHED_DEV_DIR")
	}
	if devDir == "" {
		home, _ := os.UserHomeDir()
		devDir = filepath.Join(home, "Development")
	}
	return &RepoScanner{
		devDir:    devDir,
		repoCache: make(map[string]*repoCacheEntry),
		cacheTTL:  30 * time.Second,
	}
}

// ScanRepos discovers git repos under devDir and merges with agent session data.
// agents is keyed by repo absolute path.
func (rs *RepoScanner) ScanRepos(agents map[string][]domain.AgentSession) ([]domain.RepoInfo, error) {
	entries, err := os.ReadDir(rs.devDir)
	if err != nil {
		return nil, &ScanError{Op: "readdir", Err: fmt.Errorf("reading %s: %w", rs.devDir, err)}
	}

	var repos []domain.RepoInfo
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		repoPath := filepath.Join(rs.devDir, entry.Name())

		isDir := entry.IsDir()
		if !isDir && entry.Type()&os.ModeSymlink != 0 {
			if fi, err := os.Stat(repoPath); err == nil && fi.IsDir() {
				isDir = true
			}
		}
		if !isDir {
			continue
		}

		info, err := rs.getRepoInfo(repoPath)
		if err != nil {
			continue // skip non-git dirs silently
		}

		if sessions, ok := agents[repoPath]; ok {
			info.Agents = sessions
		}

		repos = append(repos, info)
	}

	return repos, nil
}

// getRepoInfo returns cached git metadata, refreshing if older than cacheTTL (30s).
func (rs *RepoScanner) getRepoInfo(repoPath string) (domain.RepoInfo, error) {
	rs.mu.RLock()
	cached, ok := rs.repoCache[repoPath]
	rs.mu.RUnlock()

	if ok && time.Since(cached.fetchedAt) < rs.cacheTTL {
		return cached.info, nil
	}

	info, err := rs.fetchGitInfo(repoPath)
	if err != nil {
		return domain.RepoInfo{}, err
	}

	rs.mu.Lock()
	rs.repoCache[repoPath] = &repoCacheEntry{
		info:      info,
		fetchedAt: time.Now(),
	}
	rs.mu.Unlock()

	return info, nil
}

// fetchGitInfo runs batched git commands in a single shell invocation.
// Combines: is-git-repo, branch, last commit message, commit epoch, dirty status.
func (rs *RepoScanner) fetchGitInfo(repoPath string) (domain.RepoInfo, error) {
	script := strings.Join([]string{
		`git rev-parse --is-inside-work-tree 2>/dev/null`,
		`echo "---SPLIT---"`,
		`git branch --show-current 2>/dev/null || git rev-parse --short HEAD 2>/dev/null`,
		`echo "---SPLIT---"`,
		`git log --oneline -1 --format='%s' 2>/dev/null`,
		`echo "---SPLIT---"`,
		`git log -1 --format='%ct' 2>/dev/null`,
		`echo "---SPLIT---"`,
		`git status --porcelain 2>/dev/null`,
	}, " && ")

	cmd := exec.Command("sh", "-c", script)
	cmd.Dir = repoPath
	out, err := cmd.Output()
	if err != nil {
		return domain.RepoInfo{}, &ScanError{Op: "git-batch", Err: fmt.Errorf("%s: %w", repoPath, ErrNotGitRepo)}
	}

	sections := strings.Split(string(out), "---SPLIT---\n")
	if len(sections) < 5 {
		return domain.RepoInfo{}, &ScanError{Op: "git-batch", Err: fmt.Errorf("%s: incomplete output: %w", repoPath, ErrNotGitRepo)}
	}

	isGit := strings.TrimSpace(sections[0])
	if isGit != "true" {
		return domain.RepoInfo{}, &ScanError{Op: "git-batch", Err: fmt.Errorf("%s: %w", repoPath, ErrNotGitRepo)}
	}

	branch := strings.TrimSpace(sections[1])
	lastCommit := strings.TrimSpace(sections[2])

	var commitDate time.Time
	if epoch, err := strconv.ParseInt(strings.TrimSpace(sections[3]), 10, 64); err == nil {
		commitDate = time.Unix(epoch, 0)
	}

	dirty := strings.TrimSpace(sections[4]) != ""

	return domain.RepoInfo{
		Name:           filepath.Base(repoPath),
		Path:           repoPath,
		Branch:         branch,
		LastCommit:     lastCommit,
		LastCommitDate: commitDate,
		Dirty:          dirty,
	}, nil
}

// InvalidateCache removes a repo from the cache, forcing a refresh on next scan.
func (rs *RepoScanner) InvalidateCache(repoPath string) {
	rs.mu.Lock()
	delete(rs.repoCache, repoPath)
	rs.mu.Unlock()
}
