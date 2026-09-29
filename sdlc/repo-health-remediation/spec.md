# Spec: Repo health remediation
From: intent.md (2026-09-29). Status: accepted. Risk: high.

## Requirements
1. R1 (Outcome 1; Success measure 'WebSocket 403'; PR 1). If a WebSocket upgrade to /ws/<any> does not offer the subprotocol `mashed.auth.<token>` with this launch's token, the server returns HTTP 403 and does not upgrade. This applies to an existing PTY session name, an unknown name and a bmad-* name. The check runs before any session lookup, so an unknown session also gets 403, not the 404 returned today (internal/terminal/bridge.go:155-190). Test: internal/terminal/bridge_auth_test.go TestBridge_RejectsMissingToken, TestBridge_RejectsWrongToken, TestBridge_UnknownSessionWithoutToken403.
2. R2 (Outcome 1; Success measure; PR 1). With a valid token, an upgrade whose Origin is missing or not in the build's allowlist returns 403. A non-dev build accepts only `wails://wails` and rejects `http://localhost:34115`. A `-tags dev` build also accepts `http://localhost:34115` and `http://127.0.0.1:34115`. Test: TestBridge_RejectsForeignOrigin (https://evil.example), TestBridge_RejectsEmptyOrigin, TestBridge_ProdRejectsDevOrigin (default tags), TestBridge_DevAcceptsLocalhost34115 (`-tags dev`).
3. R3 (Outcome 1; PR 1). An upgrade whose Host header is not `127.0.0.1:<bridge port>` returns 403, which blocks DNS rebinding. Test: TestBridge_RejectsForeignHost.
4. R4 (Outcome 1; PR 1). Each Bridge has its own 64-hex-character token from crypto/rand, compared with crypto/subtle.ConstantTimeCompare. The token never appears in log output or in any URL. If crypto/rand fails, Start returns an error and the bridge does not listen (fail closed). Test: TestBridge_TokenUniquePerInstance, TestBridge_TokenNotLogged (captures log output during a rejected upgrade), TestNewBridge_RandFailureFailsClosed (injected reader).
5. R5 (Constraint 'token and Terminal.svelte ship together'; PR 1). Terminal.svelte gets {port, token} from the new GetTerminalAuth binding and opens `new WebSocket(url, ['mashed.v1', 'mashed.auth.'+token])`. The server echoes `mashed.v1`. If the socket closes before it opens, the pane shows 'terminal authorisation failed or bridge unavailable' in an aria-live region with a Retry control, rather than the current generic '[connection error]' written into the xterm buffer (Terminal.svelte:257-263). GetTerminalPort stays, because WorkflowBuilder.svelte:13 imports it. The bridge change and the frontend change land in the same commit. Test: vitest Terminal.auth.test.ts (a mocked WebSocket asserts the constructor args and the aria-live message on an early close); Go TestBridge_ValidTokenAndOriginUpgrades asserts the negotiated subprotocol.
6. R6 (Open question 'Wails origins'; PR 1 merge gate). Before PR 1 merges, the Origin header that WKWebView sends is recorded from a `wails build` binary and from `wails dev`, and it must match the allowlist in R2. The PR description records the observed values. If WKWebView sends `null` or no Origin, PR 1 stops and goes back to design. Test: manual checklist in the PR description.
7. R7 (Outcome 1; Constraint 'allowed roots under $HOME', widened by owner decision C1/C2; Success measure '../ traversal and symlink escape'; PR 1). ReadFile, ReadFileBase64 and WriteFile (app_git.go:757-826) accept a path only if it resolves inside an allowed root: $HOME or the configured DevDir (pathguard.AllowedRoots). Otherwise they return an error wrapping pathguard.ErrOutsideRoot and do not touch the file. Rejected cases: (a) a path outside every root; (b) `$HOME/x/../../etc/passwd`; (c) a symlink under a root whose target is outside every root; (d) for WriteFile, a new leaf inside a symlinked directory that resolves outside every root. WriteFile also rejects, with ErrDeniedPath, any target in the persistence denylist, even inside $HOME: ~/.zshrc, ~/.zprofile, ~/.zshenv, ~/.bashrc, ~/.bash_profile, ~/.profile, ~/.gitconfig, ~/.ssh/**, ~/Library/LaunchAgents/**. Paths inside a root behave as before, including creating a new file. When CodeEditor gets a rejection, it shows an announced (aria-live) error naming the reason and keeps the edit buffer. Test: internal/pathguard/pathguard_test.go (table-driven, t.TempDir as roots, including a DevDir outside HOME and each denylist entry); app_files_guard_test.go (HOME set to a TempDir with t.Setenv); vitest CodeEditor.guard.test.ts.
8. R8 (Outcome 1 'ReadBundledThemeFile skips the guard'; PR 1). ReadBundledThemeFile (theme_scanner.go:329-341) rejects a vsixPath that resolves outside bundledThemesDir() (theme_scanner.go:295-313), including through a symlink. The root is the bundled themes directory, not $HOME, because the bundle lives in /Applications. Test: theme_scanner_test.go TestReadBundledThemeFile_RejectsOutsideBundledDir, TestReadBundledThemeFile_RejectsSymlinkEscape.
9. R9 (Outcome 1; PR 1). ReadFileDiff and ReadFileAtHead (app_git.go:828-861) resolve filePath against the repo root before running any git command. This covers the untracked `git diff --no-index /dev/null <repo>/<file>` fallback (app_git.go:836), which today can read any file through `../`. Test: app_git_file_test.go TestReadFileDiff_RejectsTraversal (`../../secret` in a temp repo returns an error, and the secret's content never appears), TestReadFileAtHead_RejectsAbsolute.
10. R10 (Outcome 1 'git args guarded'; Success measure '--help / -b rejected'; PR 1). GitSwitchBranch, GitCreateBranch (the prefix, the name and the composed name) and GitMergeInto return an error wrapping git.ErrInvalidRef for `--help`, `-b`, `-`, `a..b`, `x@{1}`, `a b`, `foo.lock`, `/x` and `x/`. This happens before any auto-commit or git process runs, so HEAD and the working tree are unchanged. Valid names such as `feature/foo-1` still work. Test: internal/git/refs_test.go (table-driven); app_git_guard_test.go uses a temp repo with autoCommit=true and asserts that the HEAD sha and `git status --porcelain` are unchanged after a rejection.
11. R11 (Outcome 1; PR 1). Every git exec that takes a UI-supplied path puts `--` before the path. Checkout of a UI ref uses `git checkout <ref> --`. Every git or gh binding that takes repoPath rejects a path that is not an existing directory inside an allowed root, $HOME or the configured DevDir (see concern C2). Test: app_git_guard_test.go TestGitBindings_RejectRepoOutsideRoots (a table over the bindings) and TestReadFileDiff_DashPathCreatesNoFile (filePath `--output=/tmp/x`).
12. R12 (Outcome 1 'pty-helper socket 0600'; PR 1). The helper socket sits in a new 0700 directory from `os.MkdirTemp("", "mashed-pty-")` and has mode 0600 when the parent dials. The parent refuses to dial if the mode is broader. Shutdown removes the directory. Test: cmd/pty-helper/main_test.go (stat the socket and its directory); main_socket_test.go TestDialRefusesWorldAccessibleSocket.
13. R13 (Outcome 1 'config writes atomic'; Constraint 'reuse the storage.go atomic helper'; PR 1). The temp-plus-rename pattern in internal/bmad/storage.go:225-239 is extracted into internal/fsutil.WriteFileAtomic, which uses a unique temp file (os.CreateTemp) and fsyncs before the rename. bmad.atomicWriteJSON delegates to it in the same PR, which removes the fixed `.tmp` name collision. saveConfig, SaveTheme, RemoveTheme and WriteFile all write through it. If any step fails (create, write, fsync, rename), the original file is byte-for-byte unchanged and no temp file is left behind. Config and theme files are 0600 in a 0700 directory. An existing target keeps its mode. Test: internal/fsutil/atomic_test.go (fault injection through a read-only directory with skipIfRoot; mode preservation; no leftover temp files); internal/bmad existing storage tests unchanged and green.
14. R14 (Outcome 1 'config locked'; PR 1). Set* and Get* config calls run concurrently from 50 goroutines lose no field and pass `go test -race`. When config.json is malformed (today loadConfig returns defaults, app.go:176-183, and the next saveConfig overwrites the user's file), the first mutator renames it to `config.json.corrupt-<unix>` before writing and emits `config:recovered`. The original bytes stay in the quarantined file, and a malformed file is never overwritten in place. Test: app_config_test.go TestConfig_ConcurrentSetters_NoLostUpdate, TestConfig_MalformedIsQuarantinedNotOverwritten.
15. R15 (Outcome 2 'argv spawn API'; Constraint 'API lands before callers'; PR 2 commit 1). SessionManager.SpawnArgv sends argv[0] (resolved) as Shell and argv[1:] unchanged as Args, including elements with spaces, quotes, newlines and `$(x)`. An empty argv, or an empty argv[0], returns a TerminalError. Spawn("") still starts $SHELL, and Spawn("a b") still splits on whitespace. Test: internal/terminal/manager_argv_test.go with a fake helperSpawner.
16. R16 (Outcome 2; owner decision C3; PR 2 commit 2). SpawnPRReview and SpawnRefactorPlan no longer pass `--dangerously-skip-permissions`. SpawnPRReview spawns exactly `["claude","--model",<model>,"--allowedTools",<review allowlist>,"-p",<prompt>]`. The review allowlist is Read, Grep, Glob, `Bash(gh pr diff:*)`, `Bash(gh pr view:*)`, `Bash(git diff:*)`, `Bash(git log:*)` and `Bash(git show:*)`. SpawnRefactorPlan spawns the same shape with an allowlist of Read, Grep, Glob and Write limited to its plan path under .claude/plans/. <prompt> is the Go string before formatting, with nothing added. The exact rule syntax is checked against the installed claude CLI during build. SpawnPRReview rejects a PR number that is not all digits. SpawnAgentWithCommand keeps the string form for commands the user types. Test: app_review_test.go TestSpawnRefactorPlan_ArgvExact and app_git_spawn_test.go TestSpawnPRReview_ArgvExact (the prompt contains spaces, `"`, `'` and newlines; argv contains no `--dangerously-skip-permissions`), using the recording fake sessionManager; a manual smoke run of each spawn records that the review completes and the plan file is written.
17. R17 (Outcome 2 'SetDevDir re-entrant, goleak'; PR 2). Calling SetDevDir 20 times with alternating valid directories and then shutting down leaves no goroutines beyond the baseline. At most one scanLoop and one watchSessions run at any time. consumeEngineEvents runs exactly once, started in startup (today it is restarted on every initScanning, app_scan.go:37-39). Concurrent SetDevDir, GetDevDir and doScan pass -race. Test: app_scan_lifecycle_test.go with goleak.VerifyNone(t, goleak.IgnoreCurrent()), a running-goroutine counter, and `go test -race`.
18. R18 (Outcome 2; PR 2). If SetDevDir gets a directory where the provider fails to start, it returns an error, the previous devDir's scanners keep running, and config.json keeps the previous DevDir. Test: TestSetDevDir_ProviderFailureKeepsPreviousState.
19. R19 (Outcome 2 'asset watcher follows SetActiveContext'; PR 2). After SetActiveContext(repoA) and then SetActiveContext(repoB), creating a .md file under repoB/.claude/skills emits the asset-change event within the debounce window plus 1 s, and a change under repoA/.claude/skills emits nothing. Calling SetActiveContext again with the same repo does not restart the watcher. The AssetWatcher API does not change. Test: app_asset_follow_test.go with t.TempDir repos and a captured emit func.
20. R20 (Open question 'late-client scrollback'; Outcome 3 'flaky tests'; PR 2). A client that attaches after the process exits receives the full scrollback and then close code 1000 'process exited'. Its input is never written to the PTY. Scrollback is served only after R1-R3 auth passes. TestManagedSession_AC2 is rewritten to use `cat` with client-sent input, and passes at -count=50 on macOS and Linux. Test: internal/terminal/session_test.go TestManagedSession_LateClientAfterExitGetsScrollback and the rewritten AC2 test.
21. R21 (Outcome 3; Success measure 'fresh clone'; PR 3). On a fresh clone with no Node toolchain, `go build ./...`, `go vet ./...` and `go test -race ./...` pass. After `npm run build`, frontend/dist/.gitkeep still exists and `git status --porcelain` is empty. A Vite closeBundle plugin recreates the file after emptyOutDir removes it. Test: ci.yml go job (`test -f frontend/dist/.gitkeep`, no npm step first); ci.yml frontend job clean-tree step.
22. R22 (Outcome 3 'npm single manager'; PR 3). frontend/bun.lock is gone. On Node 22, `cd frontend && npm ci` succeeds. `.nvmrc` and the package.json engines field both pin Node 22. The justfile `build` recipe and the wails.json `frontend:install` both use `npm ci`. Test: ci.yml frontend job.
23. R23 (Outcome 3 'root-only tests'; PR 3). The three permission-dependent tests in internal/bmad skip when euid is 0 (skillgen_test.go:166, artifacts_test.go:149, assets_write_test.go:209) and pass otherwise. They use a shared skipIfRoot(t) helper. Test: `go test ./internal/bmad/...` run as root in a golang:1.25 container and as a normal user; the result is recorded in the PR.
24. R24 (Outcome 3 'ci.yml'; Constraint 'svelte-check kept'; PR 3). ci.yml runs on push and pull_request to main and dev, with `permissions: contents: read`, actions pinned by SHA, and checkout using `lfs: false`. It runs: go build, go vet, go test -race, npm ci, lint:tokens, vitest run and vite build. svelte-check.yml stays unchanged, so its required-check name survives. Test: a PR against dev shows the `go`, `frontend` and `svelte-check` checks, all green.
25. R25 (Outcome 3 'pre-push'; PR 3). The lefthook pre-push hook runs `go test -short -count=1 ./...`, which covers the root package. Test: `lefthook run pre-push` output lists package `mashed`.
26. R26 (Outcome 4 'junk untracked'; PR 4). `git ls-files` returns nothing under docs/repomixer, .playwright-mcp, .playwright-cli, desloppify-workspace, frontend/coverage, .vite or frontend/.claude, and each of these paths is matched by `git check-ignore`. No tracked file outside docs/plans and sdlc/ contains '/Users/linus'. Test: a CI step `! git ls-files | grep -E '<pattern>'` plus a git grep check.
27. R27 (Outcome 4 'personal data from HEAD'; Constraint 'no rewrite'; PR 4). At HEAD, `git grep -nE '5X8A9U965U|linus\.a\.mcm'` returns nothing outside sdlc/. The justfile signs with `env_var_or_default("MASHED_SIGN_IDENTITY", "-")`. The LICENSE copyright line names the holder with no email. There is no force-push to main. Test: a CI grep step, and a macOS `just build` smoke test with the variable unset.
28. R28 (Outcome 4 'LFS going forward'; Constraint 'no rewrite'; PR 4). .gitattributes routes themes/*.vsix and fonts/*.ttf through LFS, and `git add --renormalize` makes them LFS pointers at HEAD. Existing blobs stay in history. With pointer files present, the Go tests pass and the theme and font scanners skip non-zip and non-font files without error. `just build` fails loudly if a bundled asset is still a pointer. Test: `git lfs ls-files` lists them; CI is green with lfs:false; the justfile guard is checked by hand with a pointer file.
29. R29 (Outcome 4 'LICENSE, hooks, golangci, todo.md'; PR 4). A LICENSE file of the owner's choice exists. .githooks/ is removed and lefthook is the only hook system, with the commented desloppify block removed. .golangci.yml enables govet, errcheck, staticcheck and gosec, with issues.new-from-rev pinned to the PR 4 base. todo.md is untracked and its items are in docs/stories/backlog.md. The `.claude` triple-negation block (.gitignore:1-15) collapses to ignore `.claude/` except `.claude/skills/mashed-refactor-asset/` (see open question 3). `!internal/bmad/testdata/*.csv` is present. Test: file-existence checks in the PR 4 review; `golangci-lint run` green.
30. R30 (Outcome 5; Constraint 'no behaviour change, BMAD invariants hold'; PR 5). After executor.go (3038 lines) is split into same-package files, `go doc -all ./internal/bmad` output is identical apart from file positions, and no file in internal/bmad exceeds 1000 lines. The existing executor_suspend, executor_interactive and executor_respond tests pass unmodified, which proves the NodeAwaitingInput/activeOutEdges invariant and the §14.2-14.4 guards still hold. Test: CI plus an API-diff script in the PR.
31. R31 (Outcome 5; PR 5). NotificationFeed, WorkflowBuilder, Settings and AgentDetail are each under 1000 lines, and their keyboard focus order and ARIA labels are unchanged. app_git.go runs no `exec.Command*(..., "git"` directly, and the auto-commit block (app_git.go:204-210, 231-239, 609-616) exists once. The recoverSessions stub (app_terminal_registry.go:15-16) is gone. MarkdownEditor/Milkdown load through dynamic import() behind an aria-busy placeholder (Monaco is already dynamic at MonacoEditor.svelte:492). The entry chunk is under 500 kB with the default chunkSizeWarningLimit restored. Test: a `wc -l` check; a grep check; vite build output with no warning; the existing vitest and Playwright AC suites green.
32. R32 (Success measure '/security-review'; all PRs). `/security-review` on each phase PR, and on the final state, reports no high-severity findings. Test: review output attached to each PR's review.md.

## Design
Approach: risk-first. The design fails closed at every trust boundary between the webview and the Go backend. It ships as five PRs merged in order 1 to 5, with line references re-pinned to HEAD 4ff58d4. I followed three leads outside the context pack: the live intent.md, the loadConfig/saveConfig code (app.go:170-222), and atomicWriteJSON (internal/bmad/storage.go:225-239).

### PR 1: Security
- **Terminal gate (internal/terminal/bridge.go).**
  - A new `authorize(r)` runs first in handleWS, before `b.manager.Get` (bridge.go:163). The PTY path and the BMAD tmux path (bridge.go:174-189) both go through it.
  - It checks three things in order, and any failure returns a bare 403:
    1. The Host header is exactly `127.0.0.1:<port>`.
    2. The Origin is in the allowlist. A missing Origin is rejected.
    3. The offered subprotocol `mashed.auth.<64 hex>` matches the launch token (crypto/subtle compare).
  - The reason and the Origin are logged; the token never is.
  - The token is 32 bytes from crypto/rand, generated in NewBridge. If crypto/rand fails, the bridge does not listen.
  - The upgrader echoes `mashed.v1` and replaces the always-true CheckOrigin at bridge.go:43-47. The server gets `ReadHeaderTimeout: 5s`.
  - The allowlist lives in two files. origins_prod.go (`//go:build !dev`) allows only `wails://wails`. origins_dev.go (`//go:build dev`) also allows `http://localhost:34115` and `http://127.0.0.1:34115`.
- **Terminal frontend.** A new binding `GetTerminalAuth() TerminalAuth{Port, Token}` feeds Terminal.svelte:210-213 in the same commit. A close before open shows an announced (aria-live) auth or bridge message with a Retry control. GetTerminalPort stays, because WorkflowBuilder.svelte:13 imports it.
- **internal/pathguard (new, pure Go).**
  - `HomeRoot()` resolves the home directory.
  - `ResolveExisting(root,p)` does Abs, Clean and EvalSymlinks on both path and root, then requires the result to equal root or start with root plus a separator.
  - `ResolveForWrite(root,p)` resolves the parent directory and appends the base name. If the target exists, it must also resolve inside root.
  - The sentinel error is `ErrOutsideRoot`. This generalises theme_scanner.go:437-450 and internal/bmad/validate.go:95-104.
  - Callers:
    - ReadFile, ReadFileBase64 and WriteFile use AllowedRoots ($HOME plus the configured DevDir). WriteFile also checks the persistence denylist (C1).
    - ReadBundledThemeFile uses bundledThemesDir() as its root. It is not under $HOME, because the app bundle lives in /Applications.
    - ReadFileDiff and ReadFileAtHead resolve against the repo root before either git call, which covers the `--no-index` fallback at app_git.go:836.
    - A private `a.repoDir(repoPath)` (pathguard against AllowedRoots plus an IsDir check) guards every git and gh binding.
- **internal/git/refs.go.** `ValidateBranchName` is a pure-Go version of `check-ref-format --branch` that also rejects a leading '-'. Its sentinel is `ErrInvalidRef`. It is called in GitSwitchBranch (which becomes `checkout <ref> --`), GitCreateBranch (prefix, name and composed name) and GitMergeInto. Validation runs before autoCommit.
- **internal/fsutil.WriteFileAtomic.** It is extracted from the storage.go pattern, as the intent's reuse constraint requires, and uses a unique CreateTemp name, fsync, chmod and rename, cleaning up on failure. An existing target keeps its mode. bmad.atomicWriteJSON delegates to it in the same PR.
- **Config store.**
  - A dedicated `cfgMu` guards every config read-modify-write and every config read. a.mu is never held while taking cfgMu.
  - `loadConfig() (mashedConfig, error)` tells not-exist apart from malformed (`ErrConfigCorrupt`).
  - A malformed file is renamed to `config.json.corrupt-<unix>`, and `config:recovered` is emitted before the next write.
  - Writes are 0600 in a 0700 directory.
- **pty-helper socket.** The parent creates the socket directory with MkdirTemp (0700). The helper sets umask 0177, listens, then chmods the socket to 0600. The parent checks the mode before helper.Dial, and shutdown removes the directory.

### PR 2: Broken features and races
Order: commit 1 adds the API, commit 2 switches the callers.
- **SpawnArgv.** `SessionManager.SpawnArgv(ctx,name,repoPath,argv,cols,rows)` passes `Shell: resolveExecutable(argv[0]), Args: argv[1:]` straight through. `Spawn(command string)` becomes a thin wrapper: an empty command means $SHELL, otherwise strings.Fields then SpawnArgv. A small unexported `helperSpawner` interface is the test seam.
- **Spawn callers.** `spawnSessionArgv` is used by SpawnPRReview (app_git.go:895-897, with a digits-only PR number) and SpawnRefactorPlan (app_review.go:357-359). Both spawns drop `--dangerously-skip-permissions` and pass a per-spawn `--allowedTools` list (R16, C3).
- **Scan lifecycle.**
  - `scanState{devDir, provider, repoScanner}` sits behind an `atomic.Pointer`, and every reader takes one snapshot per operation. The readers are app_scan.go:78-126, app_spawn.go:80-83, app_git.go:62-65/98-122/160-162, app_sessions.go:16-77 and app.go:520-522.
  - `scanMu` serialises restarts.
  - A restart builds the new provider first and returns an error if that fails, leaving the old scanners running. It then persists the config, cancels the old context, waits on scanWG, stores the new pointer, and starts scanLoop and watchSessions under a child of a.ctx.
  - consumeEngineEvents starts once in startup.
- **Asset watcher.** SetActiveContext records the previous repo under a.mu. If the repo changed, it calls `swapAssetWatcher` under `watcherMu`, which builds and starts a new watcher from `bmad.AssetWatchRoots(newRepo)` and then stops the old one. If Start fails, the old watcher stays. The AssetWatcher API does not change.
- **Session scrollback.** session.go gets an `exited` flag. A client that attaches after exit gets the scrollback and then close 1000 'process exited', with no reader goroutine and no input path.

### PR 3: Build and test health
- **dist placeholder.** frontend/dist/.gitkeep is committed, with .gitignore rules `frontend/dist/*` and `!frontend/dist/.gitkeep`. A Vite closeBundle plugin recreates the file after each build, and startup warns when the embedded FS has no index.html.
- **npm only.** frontend/bun.lock is deleted and package-lock.json regenerated on Node 22. `.nvmrc` and an engines field pin Node 22, and the justfile and wails.json installs use `npm ci`.
- **Tests.** The root-only tests use a shared `skipIfRoot`, and the AC2 test is rewritten.
- **CI.** New .github/workflows/ci.yml with jobs `go` and `frontend`, `contents: read`, actions pinned by SHA and `lfs: false`. svelte-check.yml is unchanged.
- **Pre-push.** The lefthook pre-push hook runs `go test -short -count=1 ./...`.

### PR 4: Repo hygiene
- **Untrack junk.** `git rm -r --cached` removes about 187 tracked junk files, and .gitignore is extended.
- **.gitignore rules.** The `.claude` block collapses (open question 3), and `!internal/bmad/testdata/*.csv` is added so registry CSVs stay tracked (registry_fs.go:10).
- **LFS.** .gitattributes routes themes/*.vsix and fonts/*.ttf through LFS, followed by `git add --renormalize` with no history rewrite. `just build` has a guard against LFS pointer files.
- **Personal data.** The justfile uses `MASHED_SIGN_IDENTITY` (default `-`), and the wails.json author email becomes an alias.
- **LICENSE.** It is added with a name-only copyright line.
- **Hooks.** .githooks/ is removed, along with the commented desloppify block in lefthook.yml.
- **Lint.** .golangci.yml is added (govet, errcheck, staticcheck, gosec) with new-from-rev.
- **todo.md.** It is untracked, and its items move to docs/stories/backlog.md.

### PR 5: Structure (no behaviour change)
- **executor.go.** It is split by concern into same-package files: scheduling, node execution, interactive suspend/resume, outputs, loaders and events. The exported API does not change.
- **Svelte views.** The four large views are split into child components.
- **Git shell-outs.** app_git.go's shell-outs move to internal/git behind a `Runner` seam, and the auto-commit block is deduplicated. Every validator and every `--` stays in place.
- **Stub removed.** The recoverSessions stub is deleted.
- **Code split.** Milkdown is lazy-loaded through dynamic import() with an aria-busy placeholder, and the default chunkSizeWarningLimit is restored.

### Interfaces
- **Go APIs:**
  - `SpawnArgv(ctx context.Context, name, repoPath string, argv []string, cols, rows uint16) (*ManagedSession, error)`
  - `Bridge.Token() string`, `SubprotocolV1 = "mashed.v1"`
  - `GetTerminalAuth() TerminalAuth`
  - `pathguard.{HomeRoot, AllowedRoots, ResolveExisting, ResolveForWrite, CheckWriteDenylist, ErrOutsideRoot, ErrDeniedPath}`
  - `fsutil.WriteFileAtomic(path, data, perm)`
  - `git.{ValidateBranchName, ErrInvalidRef}`
- **Private helpers:** `repoDir`, `spawnSessionArgv`, `scanState`, `loadConfig() (mashedConfig, error)`, `ErrConfigCorrupt`, `swapAssetWatcher`.
- **Event:** `config:recovered {quarantinedPath}`.
- **Wire contract:** `ws://127.0.0.1:<port>/ws/<name>` with Sec-WebSocket-Protocol `mashed.v1, mashed.auth.<hex>`. The server answers 403 on rejection. An exited session sends its scrollback, then close 1000.
- **Unchanged:** every other Wails binding keeps its signature; only its error cases grow.
- **Environment:** `MASHED_SIGN_IDENTITY`.

### Rollback
Each PR can be reverted on its own. PR 1's bridge change and its Terminal.svelte change must be reverted together.

### Rejected alternatives (carry into plan.md Risks)
- **Token in a URL query parameter** (minimal and longevity designs). Rejected: URLs end up in logs and error strings (bridge.go logs failures). A browser cannot set WebSocket headers, so the token goes in the Sec-WebSocket-Protocol subprotocol instead, the same pattern Kubernetes uses.
- **Wrapping the whole mux in `Auth.Wrap` middleware and changing the NewBridge signature** (longevity). Rejected: it touches the 9 NewBridge call sites for little gain. A single `authorize()` at the top of handleWS covers both routes today. A future route must call it too, and code review checks that.
- **Replacing GetTerminalPort with GetTerminalEndpoint** (longevity). Rejected: WorkflowBuilder.svelte:13 imports GetTerminalPort. The new GetTerminalAuth binding is added and the old one is kept.
- **Confining ReadBundledThemeFile to $HOME** (longevity). Rejected as a correctness bug: bundledThemesDir() sits next to the executable inside /Applications, so every bundled theme read would fail in packaged builds.
- **Validating refs by shelling out to `git check-ref-format`** (minimal). Rejected: that passes untrusted input to a process. Validation is pure Go and runs before autoCommit.
- **Only adding `--` to the ReadFileDiff `--no-index` fallback** (minimal). Rejected: it still allows `../` traversal. filePath is confined to the repo root instead.
- **Relying on atomic rename alone and leaving config reads unlocked** (minimal). Rejected: it loses updates between concurrent read-modify-write setters, and it keeps the data-loss path where malformed config is replaced by defaults and then saved over the user's file (app.go:176-183).
- **Exporting bmad.atomicWriteJSON as-is** (minimal). Rejected: it is JSON-only and uses a fixed `.tmp` name. Its pattern is extracted into fsutil.WriteFileAtomic and bmad delegates to it, which still honours the intent's "reuse storage.go" constraint.
- **A configStore.Update rewrite of about 20 setters in PR 1** (longevity). Rejected as churn beyond need. A cfgMu around the existing load and save functions gives the same safety.
- **A new AssetWatcher.SetRoots API with per-root directory tracking** (longevity). Rejected: it is riskier inside internal/bmad. The watcher is instead rebuilt when the repo changes, starting the new one before stopping the old, with no API change.
- **Keeping consumeEngineEvents inside initScanning** (minimal). Rejected: it duplicates consumers on every SetDevDir call. It now starts once in startup.
- **Folding svelte-check into ci.yml and deleting svelte-check.yml** (longevity). Rejected: it renames a required branch-protection check. The file is kept unchanged.
- **Random suffixes on session names and `--end-of-options`** (longevity). Not adopted: once the WebSocket is authenticated, guessing a session name gives an attacker nothing, and `--end-of-options` requires git 2.24 or later. The intent's "guessable names" problem is covered by R1.
- **`git lfs migrate import --no-rewrite`** (minimal). Replaced by `.gitattributes` plus `git add --renormalize`, which has the same effect and is simpler to review.

### Out of scope
Rejecting all of $HOME in favour of a narrower root is out of scope because the intent fixed the root; it is flagged as C1 and C2 instead. Also out of scope: any history rewrite to purge personal data, and Linux or Windows WebView2 origins.

## Concerns
- **C1 Scope of the $HOME root.** Confining file bindings to $HOME still lets a compromised webview read ~/.ssh and ~/.aws, and write ~/.zshrc, ~/.bashrc, ~/.profile, ~/.zprofile, ~/.gitconfig, ~/.ssh/* or ~/Library/LaunchAgents/*, which would give it persistent code execution. The design adds a write-only denylist for those paths. The codebase would then have two confinement policies: $HOME here, and the repo root for ShapeFile under §14.2.
  - Owner: linus (product owner) with security (phase 1 implementer).
  - Policy conflict: Partly. The intent fixes the root at $HOME, while least privilege points to something narrower. The denylist narrows the intent's rule without contradicting it. The owner must sign it off, or pick a narrower root.
  - Resolution (linus, 2026-09-29): signed off. Roots are $HOME plus the configured DevDir, with the write denylist in R7.
- **C2 Repos outside $HOME.** Repos and dev directories outside $HOME (for example /Volumes/dev or /opt/src) will be rejected by the file and git bindings. This affects CodeEditor, which builds absolute paths itself (CodeEditor.svelte:65-70), and GitPanel. A rejected auto-save currently shows only saveStatus 'error' and a console log (CodeEditor.svelte:103-110), so it can look like silent data loss.
  - Owner: linus (product owner) with CodeEditor/UX.
  - Policy conflict: Yes. The intent's $HOME-only constraint breaks existing workflows outside $HOME. Two options: accept the regression, or add the configured DevDir as a second root. In either case CodeEditor must show an announced error that names the reason and keeps the edit buffer.
  - Resolution (linus, 2026-09-29): DevDir is added as a second root. This widens the intent's $HOME-only constraint; the product owner approved the change. CodeEditor shows an announced error (R7).
- **C3 Remaining webview trust.** The terminal gate does not remove the underlying trust in the webview. Any XSS in rendered markdown or theme JSON can still call SpawnAgentWithCommand (app_spawn.go:62) and WriteFile. Review PR also runs claude with --dangerously-skip-permissions over untrusted PR content (app_git.go:888-896), which is a prompt-injection path that fixing argv does not close.
  - Owner: linus / security.
  - Policy conflict: Out of the intent's scope. Recommended follow-up intent: sanitise the Milkdown and markdown renderers, add a CSP in the Wails asset server, and drop the skip-permissions flag for review and refactor spawns.
  - Resolution (linus, 2026-09-29): dropping `--dangerously-skip-permissions` for SpawnPRReview and SpawnRefactorPlan moves into PR 2 (R16). CSP and renderer sanitising stay out of scope, for a follow-up intent.
- **C4 Prompt visible to other local users.** Passing the prompt as argv makes it visible through ps to other local users on the machine. PR diffs and repo content go to the external claude CLI; that data flow is unchanged.
  - Owner: phase 2 implementer.
  - Policy conflict: none. Consider stdin for large or sensitive prompts later.
- **C5 Webview Origin unverified.** Rejecting an empty Origin fails closed, but if WKWebView omits Origin or sends `null` for the custom scheme, the terminal breaks for everyone. A missing dev origin would also break every contributor's terminal.
  - Owner: phase 1 implementer (security) / contributors.
  - Policy conflict: none. R6 makes checking the Origin on real builds a merge gate for PR 1.
- **C6 Auth failures in the UI.** A 403 from the new gate shows only the generic '[connection error]' written into the xterm buffer, which screen readers do not announce.
  - Owner: security (phase 1) with frontend/UX.
  - Policy conflict: Security wants rejections that disclose little, and UX wants errors that say what went wrong. These reconcile: the server returns a bare 403, and the client, which knows it sent a token, shows an announced 'terminal authorisation failed or bridge unavailable' message with a Retry control (R5).
- **C7 Late-client scrollback exposure.** Replaying scrollback lets any client that holds the token read past output, including secrets typed into the shell.
  - Owner: internal/terminal owner.
  - Policy conflict: none. Scrollback is bounded by the existing buffer and served only after the R1-R3 auth passes.
- **C8 Git ref validation UX.** Rejecting refs with a leading '-' or other invalid refs will make some GitPanel actions fail. Without an inline error, users see only a generic backend failure.
  - Owner: security (git guard) with GitPanel/UX.
  - Policy conflict: none. GitPanel should check names as the user types (aria-invalid plus a linked description).
- **C9 Personal data and junk remain in history.** The email in wails.json, the Apple signing identity with team ID 5X8A9U965U in justfile:48,56,57,75, and the tracked junk (coverage files with absolute paths, repomix XML) are removed from HEAD only, and stay in history on the GitHub remote.
  - Owner: linus (repo/privacy owner).
  - Policy conflict: Yes. Data minimisation conflicts with the intent's no-history-rewrite and no-force-push constraint. The intent says no-rewrite wins. Record this as an accepted residual risk, not as remediated.
  - Resolution (linus, 2026-09-29): accepted as a residual risk for this intent. A separate follow-up intent will plan a coordinated history rewrite.
- **C10 LFS bandwidth.** GitHub LFS gives 1 GB of free bandwidth a month, and a full clone pulls about 66 MB of LFS content, so roughly 15 full clones a month. CI and CI logs must not upload coverage with absolute paths or ~/.mashed data.
  - Owner: linus (repo/CI maintainer).
  - Policy conflict: none. CI stays on `lfs: false`, and releases may need an LFS data pack or a `just fetch-assets` recipe.
- **C11 Ad-hoc signing and the PTY helper.** Ad-hoc signing ('-') with `--options runtime` and the PTY entitlements must be checked on current macOS. If the helper cannot spawn PTYs when signed ad hoc, contributors must set MASHED_SIGN_IDENTITY.
  - Owner: linus (release owner).
  - Policy conflict: none
- **C12 Licence compatibility.** The LICENSE choice must be compatible with the licences of the vendored themes and fonts (~66 MB) and of any shipped .claude skill. The copyright line must name the holder with no email, so it does not reintroduce personal data.
  - Owner: linus (owner/legal).
  - Policy conflict: Possible. A permissive project licence may not cover third-party font and theme licences. List each asset's licence before choosing.
- **C13 Invariants through the phase 5 split.** Splitting executor.go and moving the git shell-outs must keep: the NodeAwaitingInput and activeOutEdges invariant; the §14.2 ShapeFile repo-root guard; §14.3 valueHash-only events (internal/bmad/events.go:86-93); §14.4 registry-only OptionsRef; and every ref validator and `--` from phase 1. Splitting the Svelte views must keep focus order and ARIA.
  - Owner: phase 5 implementer / BMAD maintainer.
  - Policy conflict: Partly. The 'no behaviour change' rule conflicts with adding the aria-busy placeholder that lazy-loading needs. The spec treats that placeholder as an allowed change.
- **C14 Local hooks and CI agree.** Any skill and hook files that ship run on contributor machines. The lefthook pre-push must match the CI gate.
  - Owner: linus (contributors).
  - Policy conflict: none
- **C15 Stale source plan.** The source plan's '43 vendored skill files' is stale: no .claude paths are tracked at HEAD 4ff58d4. executor.go is now 3038 lines, not ~1913. Every line reference in docs/plans/repo-health-remediation.md has been re-pinned in this spec.
  - Owner: phase implementers.
  - Policy conflict: none
- **C16 Stalls during SetDevDir restarts.** A SetDevDir restart waits for an in-flight doScan (lsof and git calls) to finish, so a restart can stall for a bounded time.
  - Owner: phase 2 implementer.
  - Policy conflict: none. A bounded stall is accepted over leaking goroutines. Passing the context through doScan's exec calls is optional.
- **Tech lead.** linus (recorded at design acceptance, 2026-09-29).

## Open questions
- Webview origins: answered provisionally. `wails://wails` is the macOS webview origin in both production and `wails dev` (Wails v2.12.0 darwin frontend.go:40). `http://localhost:34115` and `http://127.0.0.1:34115` are the browser dev server, allowed only under the `dev` build tag. `http://wails.localhost` is the Windows WebView2 origin and is out of scope, since the app is macOS-only. Still to confirm on real `wails build` and `wails dev` binaries before PR 1 merges (R6); the phase 1 implementer owns this.
- Late-client scrollback: answered yes. A client that attaches after the process exits gets the scrollback, read-only and only after auth, then close 1000 'process exited'. This is the real bug behind the flaky TestManagedSession_AC2 (internal/terminal/session.go:207-225). See R20.
- Vendored .claude/skills: answered in part, the rest reassigned to linus. No `.claude/` file is tracked at HEAD. The only skill the product depends on is `mashed-refactor-asset`, which the UI tells users to run (frontend/src/components/bmad/ProcessSidebar.svelte:30,440), but it is missing on disk. The .gitignore rule collapses to ignore `.claude/` except `.claude/skills/mashed-refactor-asset/`. The owner must choose, before PR 4 merges, between restoring and committing that skill after a licence check, or dropping the UI reference and ignoring all of `.claude/`.
- LICENSE: reassigned to linus (owner). PR 4 is blocked until it is chosen. MIT is the suggested default, provided it is compatible with the bundled theme and font licences (C12). The copyright line names the holder with no email.

## Proof
Each item names its test file or check.

### Go unit and integration tests
- **Terminal bridge (R1-R4):** internal/terminal/bridge_auth_test.go.
  - TestBridge_RejectsMissingToken, RejectsWrongToken and UnknownSessionWithoutToken403.
  - RejectsForeignOrigin, RejectsEmptyOrigin, ProdRejectsDevOrigin, and DevAcceptsLocalhost34115 under `-tags dev`.
  - RejectsForeignHost, TokenUniquePerInstance, TokenNotLogged, ValidTokenAndOriginUpgrades.
  - TestNewBridge_RandFailureFailsClosed.
- **Path confinement (R7):** internal/pathguard/pathguard_test.go (table test covering traversal, symlink escape, a DevDir root outside HOME, each denylist entry, and a new leaf under a symlinked directory), plus app_files_guard_test.go.
- **Bundled themes (R8):** theme_scanner_test.go TestReadBundledThemeFile_RejectsOutsideBundledDir and RejectsSymlinkEscape.
- **Repo-relative git reads (R9, R11):** app_git_file_test.go TestReadFileDiff_RejectsTraversal, TestReadFileAtHead_RejectsAbsolute and TestReadFileDiff_DashPathCreatesNoFile.
- **Refs and repo guard (R10, R11):** internal/git/refs_test.go (table test); app_git_guard_test.go runs against a temp repo with autoCommit=true, checks that HEAD and `git status --porcelain` are unchanged, and includes TestGitBindings_RejectRepoOutsideHome.
- **Socket permissions (R12):** cmd/pty-helper/main_test.go and main_socket_test.go TestDialRefusesWorldAccessibleSocket.
- **Atomic writes (R13):** internal/fsutil/atomic_test.go uses fault injection, skipIfRoot, a mode-preservation check and a no-leftover-temp check. The existing internal/bmad storage tests stay green.
- **Config store (R14):** app_config_test.go TestConfig_ConcurrentSetters_NoLostUpdate (run with -race) and TestConfig_MalformedIsQuarantinedNotOverwritten.
- **Argv spawn (R15-R16):** internal/terminal/manager_argv_test.go uses a fake helperSpawner; app_review_test.go TestSpawnRefactorPlan_ArgvExact and app_git_spawn_test.go TestSpawnPRReview_ArgvExact use the recording fake sessionManager.
- **Scan lifecycle (R17-R18):** app_scan_lifecycle_test.go uses goleak.VerifyNone with IgnoreCurrent, a goroutine counter, a single consumeEngineEvents, and -race. It includes TestSetDevDir_ProviderFailureKeepsPreviousState.
- **Asset watcher (R19):** app_asset_follow_test.go.
- **Scrollback (R20):** internal/terminal/session_test.go TestManagedSession_LateClientAfterExitGetsScrollback, plus the AC2 test rewritten to use `cat`, run at -count=50.

### Frontend
- **Terminal auth (R5):** frontend vitest Terminal.auth.test.ts checks the WebSocket constructor protocols and the aria-live message shown on an early close.
- **Editor rejection (R7):** frontend vitest CodeEditor.guard.test.ts checks that a rejected save shows an aria-live error naming the reason and keeps the edit buffer.

### CI and build checks
- **Build jobs (R21-R24):** .github/workflows/ci.yml.
  - The `go` job runs test -f frontend/dist/.gitkeep, go build, go vet and go test -race with no npm step.
  - The `frontend` job runs npm ci, lint:tokens, vitest run and vite build, then checks that `git status --porcelain` is empty.
  - svelte-check.yml keeps running unchanged.
- **Root tests (R23):** run in a golang:1.25 container as root; the log is recorded in the PR.
- **Pre-push (R25):** `lefthook run pre-push` output lists package `mashed`.
- **Hygiene greps (R26-R27):** CI grep steps for junk paths, '/Users/linus', and `5X8A9U965U|linus\.a\.mcm`.
- **LFS (R28):** `git lfs ls-files` lists the themes and fonts, CI is green with lfs:false, and the justfile pointer guard is checked by hand.
- **Repo files (R29):** file-existence checks plus `golangci-lint run` green.
- **Phase 5 (R30-R31):** a `go doc -all ./internal/bmad` API-diff script; the executor_suspend, executor_interactive and executor_respond tests pass unmodified; a `wc -l` check; a grep that no git exec remains in app_git.go; vite build output with no chunk warning; existing vitest and Playwright AC suites green.

### Manual gates
- **R6:** observed WKWebView Origin values recorded in the PR 1 description.
- **R27:** macOS `just build` smoke test with MASHED_SIGN_IDENTITY unset.
- **R32:** /security-review output attached to each PR's review.md.
