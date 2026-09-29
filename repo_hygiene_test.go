package main

// Repository-configuration checks. Each TestRepo_* guards one requirement of
// sdlc/repo-health-remediation/spec.md so config changes get a red/green cycle.

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// repoRoot returns the absolute path of the repository root (this file's dir).
func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Dir(file)
}

// trackedFiles returns `git ls-files` output for the given pathspecs.
func trackedFiles(t *testing.T, pathspecs ...string) []string {
	t.Helper()
	args := append([]string{"ls-files", "--"}, pathspecs...)
	cmd := exec.Command("git", args...)
	cmd.Dir = repoRoot(t)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git ls-files: %v", err)
	}
	return strings.Fields(string(out))
}

// readRepoFile returns a repo file's contents.
func readRepoFile(t *testing.T, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(repoRoot(t), rel))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return string(b)
}

// hasLine reports whether text contains line as a whole line.
func hasLine(text, line string) bool {
	for _, l := range strings.Split(text, "\n") {
		if strings.TrimSpace(l) == line {
			return true
		}
	}
	return false
}

// R21: a fresh clone can `go build` because frontend/dist/.gitkeep is
// tracked, so `//go:embed all:frontend/dist` always matches a file.
func TestRepo_DistPlaceholderTracked(t *testing.T) {
	if got := trackedFiles(t, "frontend/dist/.gitkeep"); len(got) != 1 {
		t.Fatalf("frontend/dist/.gitkeep is not tracked (got %v)", got)
	}
	gi := readRepoFile(t, ".gitignore")
	for _, want := range []string{"frontend/dist/*", "!frontend/dist/.gitkeep"} {
		if !hasLine(gi, want) {
			t.Errorf(".gitignore lacks %q", want)
		}
	}
	if hasLine(gi, "frontend/dist") || hasLine(gi, "frontend/dist/") {
		t.Error(".gitignore still ignores the whole frontend/dist directory")
	}
}

// R22: npm is the only frontend package manager, pinned to Node 22, and
// every install path uses the lockfile (`npm ci`).
func TestRepo_SinglePackageManager(t *testing.T) {
	root := repoRoot(t)
	if _, err := os.Stat(filepath.Join(root, "frontend", "bun.lock")); !os.IsNotExist(err) {
		t.Error("frontend/bun.lock must be removed (npm is the package manager)")
	}
	if got := strings.TrimSpace(readRepoFile(t, ".nvmrc")); got != "22" {
		t.Errorf(".nvmrc = %q, want 22", got)
	}

	var pkg struct {
		Engines map[string]string `json:"engines"`
	}
	pkgJSON := readRepoFile(t, "frontend/package.json")
	if err := json.Unmarshal([]byte(pkgJSON), &pkg); err != nil {
		t.Fatalf("package.json: %v", err)
	}
	if got := pkg.Engines["node"]; got != ">=22 <23" {
		t.Errorf("package.json engines.node = %q, want \">=22 <23\"", got)
	}

	var wails map[string]any
	if err := json.Unmarshal([]byte(readRepoFile(t, "wails.json")), &wails); err != nil {
		t.Fatalf("wails.json: %v", err)
	}
	if got := wails["frontend:install"]; got != "npm ci" {
		t.Errorf("wails.json frontend:install = %v, want \"npm ci\"", got)
	}
	just := readRepoFile(t, "justfile")
	if strings.Contains(just, "npm install") || !strings.Contains(just, "npm ci") {
		t.Error("justfile must install with `npm ci`, never `npm install`")
	}

	// Wails tracks package.json by hash; keep it in sync so builds don't
	// dirty the tree.
	sum := md5.Sum([]byte(pkgJSON))
	if got := strings.TrimSpace(readRepoFile(t, "frontend/package.json.md5")); got != hex.EncodeToString(sum[:]) {
		t.Errorf("frontend/package.json.md5 = %s, want %x", got, sum)
	}
}

// ciWorkflow is the subset of a GitHub Actions workflow the tests inspect.
type ciWorkflow struct {
	On struct {
		Push        struct{ Branches []string } `yaml:"push"`
		PullRequest struct{ Branches []string } `yaml:"pull_request"`
	} `yaml:"on"`
	Permissions map[string]string `yaml:"permissions"`
	Jobs        map[string]struct {
		Steps []struct {
			Uses string            `yaml:"uses"`
			Run  string            `yaml:"run"`
			With map[string]any    `yaml:"with"`
			Env  map[string]string `yaml:"env"`
		} `yaml:"steps"`
	} `yaml:"jobs"`
}

// svelteCheckSHA256 pins .github/workflows/svelte-check.yml at the PR 3 base
// so its required-check name and behaviour survive (R24).
const svelteCheckSHA256 = "41f55a9f1eb0cddf4383e53cd5594a6c0c50f350dc2536723dcda2b43c0a9a19"

// R24: ci.yml runs the Go and frontend gates on push and PR to main and dev,
// read-only, with actions pinned by SHA and LFS off.
func TestRepo_CIWorkflow(t *testing.T) {
	var wf ciWorkflow
	if err := yaml.Unmarshal([]byte(readRepoFile(t, ".github/workflows/ci.yml")), &wf); err != nil {
		t.Fatalf("ci.yml: %v", err)
	}
	for _, br := range []string{"main", "dev"} {
		if !slices.Contains(wf.On.Push.Branches, br) || !slices.Contains(wf.On.PullRequest.Branches, br) {
			t.Errorf("ci.yml must run on push and pull_request to %s", br)
		}
	}
	if len(wf.Permissions) != 1 || wf.Permissions["contents"] != "read" {
		t.Errorf("permissions = %v, want only contents: read", wf.Permissions)
	}

	pinned := regexp.MustCompile(`@[0-9a-f]{40}$`)
	runs := map[string]string{}
	for _, job := range []string{"go", "frontend"} {
		j, ok := wf.Jobs[job]
		if !ok {
			t.Fatalf("ci.yml has no %q job", job)
		}
		var all strings.Builder
		for _, s := range j.Steps {
			if s.Uses != "" {
				if !pinned.MatchString(s.Uses) {
					t.Errorf("%s: %q is not pinned to a 40-hex SHA", job, s.Uses)
				}
				if strings.HasPrefix(s.Uses, "actions/checkout@") && s.With["lfs"] != false {
					t.Errorf("%s: checkout must set lfs: false", job)
				}
			}
			all.WriteString(s.Run + "\n")
		}
		runs[job] = all.String()
	}

	for _, want := range []string{
		"test -f frontend/dist/.gitkeep",
		"go build ./...",
		"go vet ./...",
		"go test -race ./...",
		"go test -count=50 -run 'AC2|LateClient' ./internal/terminal/", // R20 on ubuntu
		"sudo -E env \"PATH=$PATH\" go test ./internal/bmad/",           // R23 as root
		"! git ls-files | grep -E",                                      // R26 junk
		"! git grep -l -F \"/Users/",                                    // R26 personal paths
		"! git grep -l -E '5X8A9U",                                      // R27 signing id / email
	} {
		if !strings.Contains(runs["go"], want) {
			t.Errorf("go job lacks %q", want)
		}
	}
	for _, want := range []string{"npm ci", "npm run lint:tokens", "npx vitest run", "npm run build", "git status --porcelain"} {
		if !strings.Contains(runs["frontend"], want) {
			t.Errorf("frontend job lacks %q", want)
		}
	}

	sum := sha256.Sum256([]byte(readRepoFile(t, ".github/workflows/svelte-check.yml")))
	if got := hex.EncodeToString(sum[:]); got != svelteCheckSHA256 {
		t.Errorf("svelte-check.yml changed (sha256 %s); keep it as is so the required check survives", got)
	}
}

// R25: the local pre-push hook tests every package, including the root
// `mashed` package, like CI does.
func TestRepo_PrePushCoversRoot(t *testing.T) {
	var lh struct {
		PrePush struct {
			Commands map[string]struct {
				Run string `yaml:"run"`
			} `yaml:"commands"`
		} `yaml:"pre-push"`
	}
	if err := yaml.Unmarshal([]byte(readRepoFile(t, "lefthook.yml")), &lh); err != nil {
		t.Fatalf("lefthook.yml: %v", err)
	}
	cmd, ok := lh.PrePush.Commands["go-test-all"]
	if !ok {
		t.Fatal("lefthook pre-push has no go-test-all command")
	}
	if cmd.Run != "go test -short -count=1 ./..." {
		t.Errorf("pre-push go-test-all runs %q, want %q", cmd.Run, "go test -short -count=1 ./...")
	}
}

// junkPaths are generated or personal artefacts that must never be tracked.
var junkPaths = []string{
	"docs/repomixer", ".playwright-mcp", ".playwright-cli", "desloppify-workspace",
	"frontend/coverage", ".vite", "frontend/.claude", "todo.md",
}

// personalPath is built from pieces so this file never matches itself.
var personalPath = "/Users/" + "linus"

// gitGrepFiles lists tracked files containing needle (fixed string), outside
// docs/plans/, sdlc/ and the files that must spell the needle out.
func gitGrepFiles(t *testing.T, extended bool, needle string) []string {
	t.Helper()
	args := []string{"grep", "-l", "-F"}
	if extended {
		args = []string{"grep", "-l", "-E"}
	}
	args = append(args, needle, "--", ".",
		":!docs/plans", ":!sdlc", ":!repo_hygiene_test.go", ":!.github/workflows/ci.yml")
	cmd := exec.Command("git", args...)
	cmd.Dir = repoRoot(t)
	cmd.Env = cleanGitEnv()
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok && ee.ExitCode() == 1 {
			return nil // no matches
		}
		t.Fatalf("git grep: %v", err)
	}
	return strings.Fields(string(out))
}

// R26: generated artefacts and personal paths are not tracked, and stay out.
func TestRepo_NoTrackedJunk(t *testing.T) {
	if got := trackedFiles(t, junkPaths...); len(got) > 0 {
		t.Errorf("%d junk files still tracked, e.g. %v", len(got), got[:min(3, len(got))])
	}
	for _, p := range junkPaths {
		probe := p + "/probe"
		if p == "todo.md" {
			probe = p
		}
		cmd := exec.Command("git", "check-ignore", "-q", "--no-index", probe)
		cmd.Dir = repoRoot(t)
		cmd.Env = cleanGitEnv()
		if err := cmd.Run(); err != nil {
			t.Errorf("%s is not git-ignored", probe)
		}
	}
	if got := gitGrepFiles(t, false, personalPath); len(got) > 0 {
		t.Errorf("tracked files contain a personal home path: %v", got)
	}
}

// R27: the Apple team id and the personal email are gone from HEAD, the
// justfile signs with an overridable identity, and wails.json uses the
// project alias.
func TestRepo_NoPersonalData(t *testing.T) {
	needle := "5X8A9U" + "965U|linus\\.a\\." + "mcm"
	if got := gitGrepFiles(t, true, needle); len(got) > 0 {
		t.Errorf("tracked files contain the signing identity or personal email: %v", got)
	}
	just := readRepoFile(t, "justfile")
	if !strings.Contains(just, `env_var_or_default("MASHED_SIGN_IDENTITY", "-")`) {
		t.Error(`justfile must sign with env_var_or_default("MASHED_SIGN_IDENTITY", "-")`)
	}
	if strings.Contains(just, "Apple Development:") {
		t.Error("justfile still hard-codes an Apple Development identity")
	}
	var w struct {
		Author struct {
			Email string `json:"email"`
		} `json:"author"`
	}
	if err := json.Unmarshal([]byte(readRepoFile(t, "wails.json")), &w); err != nil {
		t.Fatalf("wails.json: %v", err)
	}
	if w.Author.Email != "support@surfseer.com" {
		t.Errorf("wails.json author.email = %q, want support@surfseer.com", w.Author.Email)
	}
}

// R28: bundled themes and fonts are stored with Git LFS going forward.
func TestRepo_LFSAttributes(t *testing.T) {
	ga := readRepoFile(t, ".gitattributes")
	for _, pat := range []string{"themes/*.vsix", "fonts/*.ttf"} {
		if !hasLine(ga, pat+" filter=lfs diff=lfs merge=lfs -text") {
			t.Errorf(".gitattributes does not route %s through LFS", pat)
		}
	}
}

// R28: `just build` refuses to bundle an asset that is still an LFS pointer.
func TestRepo_JustfileLFSGuard(t *testing.T) {
	justBin, err := exec.LookPath("just")
	if err != nil {
		t.Skip("just not installed")
	}
	run := func(dir string) error {
		cmd := exec.Command(justBin, "--justfile", filepath.Join(repoRoot(t), "justfile"),
			"--working-directory", repoRoot(t), "check-lfs")
		cmd.Env = append(cleanGitEnv(), "MASHED_ASSET_DIRS="+dir)
		return cmd.Run()
	}

	pointerDir := t.TempDir()
	pointer := "version https://git-lfs.github.com/spec/v1\noid sha256:00\nsize 1\n"
	if err := os.WriteFile(filepath.Join(pointerDir, "x.ttf"), []byte(pointer), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := run(pointerDir); err == nil {
		t.Error("check-lfs passed on an LFS pointer file")
	}

	realDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(realDir, "x.ttf"), []byte{0, 1, 0, 0}, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := run(realDir); err != nil {
		t.Errorf("check-lfs failed on a real binary: %v", err)
	}
	if !strings.Contains(readRepoFile(t, "justfile"), "build: build-helper check-lfs") {
		t.Error("just build must depend on check-lfs")
	}
}

// ignored reports whether path is git-ignored (path need not exist).
func ignored(t *testing.T, path string) bool {
	t.Helper()
	cmd := exec.Command("git", "check-ignore", "-q", "--no-index", path)
	cmd.Dir = repoRoot(t)
	cmd.Env = cleanGitEnv()
	return cmd.Run() == nil
}

// R29: LICENSE, one hook system, lint config, the tracked skill and the
// registry CSVs.
func TestRepo_Basics(t *testing.T) {
	lic := readRepoFile(t, "LICENSE")
	if !strings.HasPrefix(lic, "MIT License") {
		t.Error("LICENSE is not the MIT licence")
	}
	if !strings.Contains(lic, "Copyright (c) 2026 Linus McManamey\n") || strings.Contains(lic, "@") {
		t.Error("LICENSE copyright line must name the holder with no email")
	}

	if got := trackedFiles(t, ".githooks"); len(got) > 0 {
		t.Errorf(".githooks/ still tracked (%v); lefthook is the only hook system", got)
	}
	if strings.Contains(readRepoFile(t, "lefthook.yml"), "desloppify") {
		t.Error("lefthook.yml still carries the commented desloppify block")
	}

	var lint struct {
		Version string `yaml:"version"`
		Linters struct {
			Enable []string `yaml:"enable"`
		} `yaml:"linters"`
		Issues struct {
			NewFromRev string `yaml:"new-from-rev"`
		} `yaml:"issues"`
	}
	if err := yaml.Unmarshal([]byte(readRepoFile(t, ".golangci.yml")), &lint); err != nil {
		t.Fatalf(".golangci.yml: %v", err)
	}
	for _, l := range []string{"govet", "errcheck", "staticcheck", "gosec"} {
		if !slices.Contains(lint.Linters.Enable, l) {
			t.Errorf(".golangci.yml does not enable %s", l)
		}
	}
	if len(lint.Issues.NewFromRev) != 40 {
		t.Errorf(".golangci.yml issues.new-from-rev = %q, want the PR 4 base SHA", lint.Issues.NewFromRev)
	}

	// Owner decision: restore the mashed-refactor-asset skill; every other
	// .claude/ path stays personal.
	skill := ".claude/skills/mashed-refactor-asset/SKILL.md"
	if got := trackedFiles(t, skill); len(got) != 1 {
		t.Errorf("%s is not tracked", skill)
	}
	if ignored(t, skill) {
		t.Errorf("%s is git-ignored", skill)
	}
	for _, p := range []string{".claude/skills/other/SKILL.md", ".claude/settings.local.json", ".claude/worktrees/x/y"} {
		if !ignored(t, p) {
			t.Errorf("%s should be git-ignored", p)
		}
	}

	if !hasLine(readRepoFile(t, ".gitignore"), "!internal/bmad/testdata/*.csv") {
		t.Error(".gitignore must re-include internal/bmad/testdata/*.csv (registry data)")
	}
	if ignored(t, "internal/bmad/testdata/new.csv") {
		t.Error("internal/bmad/testdata/*.csv is git-ignored")
	}
}

// lineCount returns the number of lines in a repo file.
func lineCount(t *testing.T, rel string) int {
	t.Helper()
	return strings.Count(readRepoFile(t, rel), "\n")
}

// R30: no Go file in internal/bmad (tests included) exceeds 1000 lines.
func TestRepo_BmadFilesUnder1000Lines(t *testing.T) {
	files, err := filepath.Glob(filepath.Join(repoRoot(t), "internal", "bmad", "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		rel, _ := filepath.Rel(repoRoot(t), f)
		if n := lineCount(t, rel); n > 1000 {
			t.Errorf("%s has %d lines (max 1000)", rel, n)
		}
	}
}

// viewUnder1000 fails when a Svelte view exceeds 1000 lines (R31).
func viewUnder1000(t *testing.T, name string) {
	t.Helper()
	if n := lineCount(t, "frontend/src/views/"+name); n > 1000 {
		t.Errorf("frontend/src/views/%s has %d lines (max 1000)", name, n)
	}
}

func TestRepo_NotificationFeedUnder1000Lines(t *testing.T) { viewUnder1000(t, "NotificationFeed.svelte") }

func TestRepo_WorkflowBuilderUnder1000Lines(t *testing.T) { viewUnder1000(t, "WorkflowBuilder.svelte") }

func TestRepo_SettingsUnder1000Lines(t *testing.T) { viewUnder1000(t, "Settings.svelte") }

func TestRepo_AgentDetailUnder1000Lines(t *testing.T) { viewUnder1000(t, "AgentDetail.svelte") }

// R31: app_git.go delegates every git invocation to internal/git and has a
// single auto-commit helper instead of three copies.
func TestRepo_AppGitNoDirectGitExec(t *testing.T) {
	src := readRepoFile(t, "app_git.go")
	direct := regexp.MustCompile(`exec\.Command(Context)?\([^)]*"git"`)
	if m := direct.FindAllString(src, -1); len(m) > 0 {
		t.Errorf("app_git.go still runs git directly (%d sites), e.g. %q", len(m), m[0])
	}
	if n := strings.Count(src, "if autoCommit {"); n > 1 {
		t.Errorf("auto-commit block appears %d times; use autoCommitIfRequested", n)
	}
	if !strings.Contains(src, "func (a *App) autoCommitIfRequested(") {
		t.Error("app_git.go lacks autoCommitIfRequested")
	}
}

func TestRepo_GraphifyOutIgnored(t *testing.T) {
	cmd := exec.Command("git", "check-ignore", "-q", "graphify-out/x")
	cmd.Dir = repoRoot(t)
	if err := cmd.Run(); err != nil {
		t.Fatalf("graphify-out/ is not git-ignored: %v", err)
	}
}
