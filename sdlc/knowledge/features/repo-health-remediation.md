---
type: Feature
title: Repo health remediation
description: "The Mashed desktop app (Wails, Go + Svelte, macOS) has security holes, broken features and a build that does not work from a fresh clone. Source review: `docs/plans/repo-health-remediation.md` (measur"
resource: sdlc/repo-health-remediation
tags: [feature, accepted]
status: stable
generated: { by: sdlc/0.8.1, at: "2026-09-29T12:23:14Z" }
verified:
  - { by: "human:linus-mcmanamey", at: "2026-09-29T11:16:08Z" }
  - { by: "human:linus-mcmanamey", at: "2026-09-29T11:35:20Z" }
  - { by: "human:linus-mcmanamey", at: "2026-09-29T12:23:00Z" }
stale_after: "2026-10-13T12:23:14Z"
source_commit: ab1f2eb4ec65fc2b1fce15503c11f0e310758404
sources:
  - { id: intent, resource: sdlc/repo-health-remediation/intent.md, last_modified: "2026-09-29T21:16:08+10:00", digest: 476d74d6341cb18a }
  - { id: spec, resource: sdlc/repo-health-remediation/spec.md, last_modified: "2026-09-29T21:35:20+10:00", digest: 39c3836c2ae03762 }
  - { id: plan, resource: sdlc/repo-health-remediation/plan.md, last_modified: "2026-09-29T22:23:00+10:00", digest: 630a4bebc7512078 }
---

# Problem
The Mashed desktop app (Wails, Go + Svelte, macOS) has security holes, broken features and a build that does not work from a fresh clone. Source review: `docs/plans/repo-health-remediation.md` (measured on `main` @ `5ca92ae`, before the `dev` merge; line references there must be re-pinned).

Today:
- Any web page open in the user's browser can connect to the terminal WebSocket and type into a live shell or Claude PTY. `wsUpgrader.CheckOrigin` returns `true` and there is no token (`internal/terminal/bridge.go:44-45`); session names are guessable (`app_spawn.go:21`).
- The frontend can read or write any file on disk. `WriteFile`, `ReadFile`, `ReadFileBase64` (`app_git.go:757-826`) have no path confinement; `ReadBundledThemeFile` (`theme_scanner.go:329`) skips the guard its siblings use.
- Branch names and paths from the UI reach `git` without `--` or leading-dash checks (option injection).
- "Review PR" and "Refactor plan" spawn Claude with a broken prompt: `-p %q` is built as one string (`app_git.go:896`, `app_review.go:359`) and split by `strings.Fields` (`internal/terminal/manager.go:55-58`), so Claude gets argv fragments with literal quotes.
- `SetDevDir` restarts scanning goroutines without cancelling the old ones and writes `devDir`/`provider`/`repoScanner` without a lock (`app_scan.go:20-37`); the asset watcher never follows the active repo; `saveConfig` is not atomic; the pty-helper socket is not `0600`.
- A fresh clone cannot `go build`/`go test` (`main.go:24` embeds `frontend/dist`); `npm ci` fails on an out-of-sync lockfile; some Go tests fail as root or flake on Linux; CI runs only svelte-check on `dev`; the pre-push hook tests only `./internal/...`.
- The repo tracks ~40 MB of generated junk (repomix XML, coverage with absolute paths, playwright output), ~66 MB of binary themes/fonts in plain git, and personal data (email in `wails.json`, Apple signing identity in `justfile`).
- Several files are too large to change safely (`internal/bmad/executor.go` ~3000 lines, four Svelte views over 1000 lines each, `app_git.go` duplicating `internal/git`).

Affected: the single local user (security and broken Claude spawns) and every contributor (cannot build, no CI gate).

# Outcome
All five phases of `docs/plans/repo-health-remediation.md` are done:
1. **Security** — terminal WebSocket requires a per-launch token and a Wails-origin allowlist; git args are guarded with `--` and leading-dash rejection; file bindings are confined; pty-helper socket is `0600`; config writes are atomic and locked.
2. **Broken features and races** — Claude spawns take a `[]string` argv; `SetDevDir` is re-entrant (cancel + wait, fields guarded, goleak test); the asset watcher follows `SetActiveContext`.
3. **Build and test health** — `frontend/dist` placeholder; npm is the single package manager with a synced lockfile; root-only and flaky tests fixed; new `ci.yml`; pre-push runs `go test ./...`.
4. **Repo hygiene** — junk untracked and ignored; themes/fonts tracked with Git LFS going forward; personal data removed from HEAD; LICENSE, one hook system, `.golangci.yml`, `todo.md` resolved.
5. **Structure** — split `executor.go` and the large Svelte views; move git shell-outs into `internal/git`; remove the `recoverSessions` no-op stub; code-split Monaco/Milkdown.

Success measure (done means all of):
- On a fresh clone in CI: `go build ./...`, `go vet ./...`, `go test -race ./...`, `npm ci`, `vitest run`, `vite build` all pass.
- Each P0 security item has a test proving the attack is blocked: WebSocket upgrade without token or with a foreign `Origin` returns `403`; a branch named `--help` / `-b` is rejected; `../` traversal and symlink escape return an error.
- `/security-review` reports no high-severity findings.

# Requirements
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

# Files
- `.claude/skills/mashed-refactor-asset/SKILL.md`
- `.gitattributes`
- `.githooks/pre-commit`
- `.github/workflows/ci.yml`
- `.gitignore`
- `.golangci.yml`
- `.nvmrc`
- `.playwright-cli/page-2026-04-08T02-07-40-459Z.yml`
- `.playwright-cli/page-2026-04-08T02-08-01-351Z.yml`
- `.playwright-cli/page-2026-04-08T02-08-43-225Z.yml`
- `.playwright-cli/page-2026-04-08T02-10-43-332Z.yml`
- `.playwright-cli/page-2026-04-08T02-11-05-458Z.yml`
- `.playwright-cli/page-2026-04-08T02-12-50-811Z.yml`
- `.playwright-cli/page-2026-04-08T02-13-10-615Z.yml`
- `.playwright-cli/page-2026-04-08T02-13-29-179Z.yml`
- `.playwright-cli/page-2026-04-08T02-17-42-304Z.yml`
- `.playwright-cli/page-2026-04-08T02-17-57-674Z.yml`
- `.playwright-cli/page-2026-04-08T02-20-15-065Z.yml`
- `.playwright-cli/page-2026-04-08T02-20-28-908Z.yml`
- `.playwright-cli/page-2026-04-08T03-32-30-687Z.yml`
- `.playwright-cli/page-2026-04-08T03-32-47-520Z.yml`
- `.playwright-cli/page-2026-04-08T03-33-27-801Z.yml`
- `.playwright-cli/page-2026-04-08T04-45-22-257Z.yml`
- `.playwright-cli/page-2026-04-08T04-47-03-333Z.yml`
- `.playwright-cli/page-2026-04-08T04-47-24-899Z.yml`
- `.playwright-cli/page-2026-04-08T04-47-46-615Z.yml`
- `.playwright-mcp/console-2026-04-01T21-15-54-193Z.log`
- `.playwright-mcp/console-2026-04-01T21-18-35-918Z.log`
- `.playwright-mcp/console-2026-04-01T21-22-07-700Z.log`
- `.playwright-mcp/console-2026-04-01T21-33-29-404Z.log`
- `.playwright-mcp/console-2026-04-01T21-38-05-974Z.log`
- `.playwright-mcp/console-2026-04-01T21-39-03-053Z.log`
- `.playwright-mcp/console-2026-04-01T21-42-53-767Z.log`
- `.playwright-mcp/console-2026-04-01T21-46-37-596Z.log`
- `.playwright-mcp/console-2026-04-01T21-47-35-278Z.log`
- `.playwright-mcp/console-2026-04-01T21-50-29-051Z.log`
- `.playwright-mcp/console-2026-04-01T21-51-53-146Z.log`
- `.playwright-mcp/console-2026-04-01T21-54-52-127Z.log`
- `.playwright-mcp/console-2026-04-01T21-59-06-260Z.log`
- `.playwright-mcp/console-2026-04-01T22-01-31-542Z.log`
- `.playwright-mcp/console-2026-04-01T22-09-33-251Z.log`
- `.playwright-mcp/console-2026-04-01T22-12-31-018Z.log`
- `.playwright-mcp/console-2026-04-01T22-15-15-831Z.log`
- `.playwright-mcp/console-2026-04-01T22-17-00-582Z.log`
- `.playwright-mcp/console-2026-04-01T23-12-01-648Z.log`
- `.playwright-mcp/console-2026-04-02T01-17-14-795Z.log`
- `.playwright-mcp/console-2026-04-02T01-17-22-689Z.log`
- `.playwright-mcp/page-2026-04-01T21-15-54-581Z.yml`
- `.playwright-mcp/page-2026-04-01T21-18-35-994Z.yml`
- `.playwright-mcp/page-2026-04-01T21-19-06-087Z.yml`
- `.playwright-mcp/page-2026-04-01T21-22-07-779Z.yml`
- `.playwright-mcp/page-2026-04-01T21-22-23-486Z.yml`
- `.playwright-mcp/page-2026-04-01T21-33-29-564Z.yml`
- `.playwright-mcp/page-2026-04-01T21-33-53-183Z.yml`
- `.playwright-mcp/page-2026-04-01T21-38-06-223Z.yml`
- `.playwright-mcp/page-2026-04-01T21-39-03-176Z.yml`
- `.playwright-mcp/page-2026-04-01T21-40-24-868Z.yml`
- `.playwright-mcp/page-2026-04-01T21-42-53-855Z.yml`
- `.playwright-mcp/page-2026-04-01T21-43-11-621Z.yml`
- `.playwright-mcp/page-2026-04-01T21-46-37-841Z.yml`
- `.playwright-mcp/page-2026-04-01T21-47-35-361Z.yml`
- `.playwright-mcp/page-2026-04-01T21-50-29-188Z.yml`
- `.playwright-mcp/page-2026-04-01T21-51-53-292Z.yml`
- `.playwright-mcp/page-2026-04-01T21-53-19-907Z.yml`
- `.playwright-mcp/page-2026-04-01T21-54-52-281Z.yml`
- `.playwright-mcp/page-2026-04-01T21-59-06-437Z.yml`
- `.playwright-mcp/page-2026-04-01T22-01-31-701Z.yml`
- `.playwright-mcp/page-2026-04-01T22-09-33-442Z.yml`
- `.playwright-mcp/page-2026-04-01T22-11-20-170Z.yml`
- `.playwright-mcp/page-2026-04-01T22-12-31-209Z.yml`
- `.playwright-mcp/page-2026-04-01T22-13-21-095Z.yml`
- `.playwright-mcp/page-2026-04-01T22-15-16-033Z.yml`
- `.playwright-mcp/page-2026-04-01T22-17-00-770Z.yml`
- `.playwright-mcp/page-2026-04-01T23-12-01-876Z.yml`
- `.playwright-mcp/page-2026-04-02T01-17-15-032Z.yml`
- `.playwright-mcp/page-2026-04-02T05-58-07-871Z.yml`
- `.playwright-mcp/page-2026-04-02T05-58-11-675Z.png`
- `.playwright-mcp/page-2026-04-04T08-14-29-283Z.yml`
- `.playwright-mcp/page-2026-04-04T08-14-48-697Z.png`
- `.playwright-mcp/page-2026-04-07T12-58-23-669Z.yml`
- `.playwright-mcp/page-2026-04-07T12-58-27-121Z.png`
- `.playwright-mcp/page-2026-04-07T12-59-05-350Z.yml`
- `.playwright-mcp/page-2026-04-07T12-59-09-337Z.png`
- `.playwright-mcp/page-2026-04-07T12-59-42-717Z.yml`
- `.playwright-mcp/page-2026-04-07T12-59-46-642Z.png`
- `.playwright-mcp/page-2026-04-07T13-00-20-995Z.png`
- `.playwright-mcp/page-2026-04-07T13-01-01-228Z.png`
- `.playwright-mcp/page-2026-04-07T13-02-25-079Z.yml`
- `.playwright-mcp/page-2026-04-07T13-02-31-992Z.yml`
- `.playwright-mcp/page-2026-04-07T13-02-43-583Z.yml`
- `.playwright-mcp/page-2026-04-07T13-02-56-519Z.yml`
- `.playwright-mcp/page-2026-04-07T13-03-02-262Z.png`
- `.playwright-mcp/page-2026-04-07T13-03-25-579Z.yml`
- `.playwright-mcp/page-2026-04-07T13-03-39-762Z.yml`
- `.playwright-mcp/page-2026-04-07T13-03-43-742Z.png`
- `.playwright-mcp/page-2026-04-07T13-04-04-100Z.png`
- `.playwright-mcp/page-2026-04-07T13-04-22-662Z.png`
- `.playwright-mcp/page-2026-04-07T13-05-10-829Z.yml`
- `.playwright-mcp/page-2026-04-07T13-05-19-074Z.png`
- `.playwright-mcp/page-2026-04-07T13-05-28-748Z.yml`
- `.playwright-mcp/page-2026-04-07T13-05-47-315Z.yml`
- `.playwright-mcp/page-2026-04-07T13-05-56-106Z.png`
- `.playwright-mcp/page-2026-04-07T13-13-36-476Z.yml`
- `.playwright-mcp/page-2026-04-07T13-14-18-389Z.png`
- `.playwright-mcp/page-2026-04-07T13-17-36-610Z.yml`
- `.playwright-mcp/page-2026-04-07T13-17-43-368Z.yml`
- `.playwright-mcp/page-2026-04-07T13-18-01-381Z.yml`
- `.playwright-mcp/page-2026-04-07T13-18-07-995Z.png`
- `.playwright-mcp/page-2026-04-07T13-18-27-334Z.yml`
- `.playwright-mcp/page-2026-04-07T13-18-45-049Z.yml`
- `.playwright-mcp/page-2026-04-07T13-18-51-156Z.png`
- `.playwright-mcp/page-2026-04-08T02-50-00-675Z.yml`
- `.playwright-mcp/page-2026-04-08T02-50-34-997Z.yml`
- `.playwright-mcp/page-2026-04-08T02-50-49-634Z.yml`
- `.playwright-mcp/page-2026-04-08T02-51-09-361Z.yml`
- `.playwright-mcp/page-2026-04-08T02-51-25-176Z.yml`
- `.playwright-mcp/page-2026-04-08T02-51-36-182Z.yml`
- `.playwright-mcp/page-2026-04-08T02-53-48-314Z.yml`
- `.playwright-mcp/page-2026-04-08T03-06-53-875Z.yml`
- `.playwright-mcp/page-2026-04-08T03-07-06-834Z.yml`
- `.playwright-mcp/page-2026-04-08T03-07-27-099Z.yml`
- `.playwright-mcp/page-2026-04-08T03-08-27-539Z.yml`
- `.playwright-mcp/page-2026-04-08T03-31-23-603Z.yml`
- `.playwright-mcp/page-2026-04-08T03-31-47-553Z.yml`
- `.repomixignore`
- `.sdlc.toml`
- `.vite/deps_temp_c90be2f4/package.json`
- `LICENSE`
- `README.md`
- `app.go` in [ensureGitignoreEntry](/modules/ensuregitignoreentry.md)
- `app_asset_follow_test.go`
- `app_config_test.go` in [setupTestConfig](/modules/setuptestconfig.md)
- `app_files_guard_test.go`
- `app_git.go` in [mimeForExt](/modules/mimeforext.md)
- `app_git_file_test.go`
- `app_git_guard_test.go`
- `app_git_spawn_test.go`
- `app_review.go` in [App](/modules/app-348.md)
- `app_review_scoped.go` in [loader_test.go](/modules/loader-test-go.md)
- `app_review_scoped_test.go` in [loader_test.go](/modules/loader-test-go.md)
- `app_review_test.go` in [App](/modules/app-348.md)
- `app_scan.go` in [ClaudeCodeProvider](/modules/claudecodeprovider.md)
- `app_scan_lifecycle_test.go`
- `app_sessions.go` in [SessionData](/modules/sessiondata.md)
- `app_spawn.go` in [App](/modules/app-355.md)
- `app_terminal_registry.go` in [App](/modules/app-349.md)
- `app_terminal_registry_test.go` in [app_terminal_registry_test.go](/modules/app-terminal-registry-test-go.md)
- `app_uiadapter.go` in [App](/modules/app-296.md)
- `app_uiadapter_bindings_test.go` in [setupTestConfig](/modules/setuptestconfig.md)
- `app_uiadapter_v3.go` in [App](/modules/app-361.md)
- `app_uiadapter_v3_test.go` in [app_uiadapter_v3_test.go](/modules/app-uiadapter-v3-test-go.md)
- `bundled_themes_test.go` in [bundled_themes_test.go](/modules/bundled-themes-test-go.md)
- `cmd/pty-helper/main.go` in [go_pkg_testing](/modules/go-pkg-testing.md)
- `cmd/pty-helper/main_test.go`
- `desloppify-workspace/iteration-1/benchmark.json`
- `desloppify-workspace/iteration-1/hook-setup/eval_metadata.json`
- `desloppify-workspace/iteration-1/hook-setup/with_skill/grading.json`
- `desloppify-workspace/iteration-1/hook-setup/with_skill/outputs/transcript.md`
- `desloppify-workspace/iteration-1/hook-setup/with_skill/timing.json`
- `desloppify-workspace/iteration-1/hook-setup/without_skill/grading.json`
- `desloppify-workspace/iteration-1/hook-setup/without_skill/outputs/step1-venv.log`
- `desloppify-workspace/iteration-1/hook-setup/without_skill/outputs/step2-pip-install.log`
- `desloppify-workspace/iteration-1/hook-setup/without_skill/outputs/step5-initial-scan.log`
- `desloppify-workspace/iteration-1/hook-setup/without_skill/outputs/step8-hook-setup.log`
- `desloppify-workspace/iteration-1/hook-setup/without_skill/outputs/transcript.md`
- `desloppify-workspace/iteration-1/hook-setup/without_skill/timing.json`
- `desloppify-workspace/iteration-1/setup-and-scan/eval_metadata.json`
- `desloppify-workspace/iteration-1/setup-and-scan/with_skill/grading.json`
- `desloppify-workspace/iteration-1/setup-and-scan/with_skill/outputs/scan-results.md`
- `desloppify-workspace/iteration-1/setup-and-scan/with_skill/outputs/setup-status.md`
- `desloppify-workspace/iteration-1/setup-and-scan/with_skill/outputs/transcript.md`
- `desloppify-workspace/iteration-1/setup-and-scan/with_skill/timing.json`
- `desloppify-workspace/iteration-1/setup-and-scan/without_skill/grading.json`
- `desloppify-workspace/iteration-1/setup-and-scan/without_skill/timing.json`
- `desloppify-workspace/iteration-1/sloppy-code-check/eval_metadata.json`
- `desloppify-workspace/iteration-1/sloppy-code-check/with_skill/grading.json`
- `desloppify-workspace/iteration-1/sloppy-code-check/with_skill/outputs/next-actions.txt`
- `desloppify-workspace/iteration-1/sloppy-code-check/with_skill/outputs/scan-results.txt`
- `desloppify-workspace/iteration-1/sloppy-code-check/with_skill/outputs/setup-log.txt`
- `desloppify-workspace/iteration-1/sloppy-code-check/with_skill/outputs/transcript.md`
- `desloppify-workspace/iteration-1/sloppy-code-check/with_skill/timing.json`
- `desloppify-workspace/iteration-1/sloppy-code-check/without_skill/grading.json`
- `desloppify-workspace/iteration-1/sloppy-code-check/without_skill/outputs/health-check-report.md`
- `desloppify-workspace/iteration-1/sloppy-code-check/without_skill/outputs/raw-command-outputs.txt`
- `desloppify-workspace/iteration-1/sloppy-code-check/without_skill/outputs/transcript.md`
- `desloppify-workspace/iteration-1/sloppy-code-check/without_skill/timing.json`
- `desloppify-workspace/iteration-2/benchmark.json`
- `desloppify-workspace/iteration-2/feedback.json`
- `desloppify-workspace/iteration-2/hook-setup/with_skill/grading.json`
- `desloppify-workspace/iteration-2/hook-setup/with_skill/outputs/transcript.md`
- `desloppify-workspace/iteration-2/hook-setup/with_skill/timing.json`
- `desloppify-workspace/iteration-2/hook-setup/without_skill/grading.json`
- `desloppify-workspace/iteration-2/hook-setup/without_skill/outputs/transcript.md`
- `desloppify-workspace/iteration-2/hook-setup/without_skill/timing.json`
- `desloppify-workspace/iteration-2/setup-and-scan/with_skill/grading.json`
- `desloppify-workspace/iteration-2/setup-and-scan/with_skill/outputs/query.json`
- `desloppify-workspace/iteration-2/setup-and-scan/with_skill/outputs/scorecard.png`
- `desloppify-workspace/iteration-2/setup-and-scan/with_skill/outputs/transcript.md`
- `desloppify-workspace/iteration-2/setup-and-scan/with_skill/timing.json`
- `desloppify-workspace/iteration-2/setup-and-scan/without_skill/grading.json`
- `desloppify-workspace/iteration-2/setup-and-scan/without_skill/outputs/transcript.md`
- `desloppify-workspace/iteration-2/setup-and-scan/without_skill/timing.json`
- `desloppify-workspace/iteration-2/sloppy-code-check/with_skill/grading.json`
- `desloppify-workspace/iteration-2/sloppy-code-check/with_skill/outputs/transcript.md`
- `desloppify-workspace/iteration-2/sloppy-code-check/with_skill/timing.json`
- `desloppify-workspace/iteration-2/sloppy-code-check/without_skill/grading.json`
- `desloppify-workspace/iteration-2/sloppy-code-check/without_skill/outputs/transcript.md`
- `desloppify-workspace/iteration-2/sloppy-code-check/without_skill/timing.json`
- `docs/DOCUMENTATION_SUMMARY.txt`
- `docs/SPECIFICATION.md` in [10. Wails Bindings (Go → Svelte API)](/modules/10-wails-bindings-go-svelte-api.md)
- `docs/agent_reports/skill-rectification-use-repo-code-2026-04-12.md`
- `docs/agent_reports/skill-validation-team-sprint-2026-04-12.md`
- `docs/agent_reports/skill-validation-use-repo-code-2026-04-12.md`
- `docs/plans/mashed-pty-helper-implementation-plan.md`
- `docs/plans/repo-health-remediation.md` in [WriteFile](/modules/writefile.md)
- `docs/repomixer/bmad-method/bmad-method.xml`
- `docs/repomixer/desloppify/desloppify.xml`
- `docs/repomixer/wails/wails.xml`
- `docs/repomixer/xyflow/xyflow.xml`
- `docs/reports/pty-fork-exec-investigation.md`
- `docs/stories/backlog.md`
- `docs/stories/markdown-toolbar-01-backend-config.md` in [SetEditorSettings](/modules/seteditorsettings.md)
- `docs/stories/markdown-toolbar-02-frontend-store.md` in [clearMarkdownMenuDirty](/modules/clearmarkdownmenudirty.md)
- `docs/stories/markdown-toolbar-03-toolbar-builder.md` in [markdownToolbarBuilder.test.ts](/modules/markdowntoolbarbuilder-test-ts.md)
- `docs/stories/markdown-toolbar-04-settings-layout-refactor.md`
- `docs/stories/markdown-toolbar-05-markdown-editor-panel.md` in [clearMarkdownMenuDirty](/modules/clearmarkdownmenudirty.md)
- `docs/stories/markdown-toolbar-06-editor-wiring.md` in [MarkdownEditor.test.ts](/modules/markdowneditor-test-ts.md)
- `docs/stories/markdown-toolbar-07-app-hydration.md`
- `docs/stories/markdown-toolbar-08-e2e-verification.md` in [clearMarkdownMenuDirty](/modules/clearmarkdownmenudirty.md)
- `docs/stories/old_stories/bmad-sprint-backlog.md`
- `docs/stories/old_stories/bridge-01-descriptive-session-names.md`
- `docs/stories/old_stories/bridge-02-tmux-adapter.md`
- `docs/stories/old_stories/lefthook-backlog.md`
- `docs/stories/old_stories/review-scoped-01-backend-scoped-advice.md` in [StreamScopedAdvice](/modules/streamscopedadvice.md)
- `docs/stories/old_stories/review-scoped-02-file-selection-ui.md`
- `docs/stories/old_stories/review-scoped-03-wired-scoped-flow.md` in [StreamScopedAdvice](/modules/streamscopedadvice.md)
- `docs/stories/old_stories/sprint2-03-repo-scoped-workflows.md` in [CreateFromTemplate](/modules/createfromtemplate.md)
- `docs/stories/old_stories/sprint2-04-repo-context-flow.md` in [CreateFromTemplate](/modules/createfromtemplate.md)
- `docs/stories/old_stories/sprint2-06-execution-bar-simplification.md`
- `docs/stories/old_stories/sprint2-08-repo-context-header.md`
- `docs/stories/old_stories/theme-01-backend-scanner.md` in [SetTheme](/modules/settheme.md)
- `docs/stories/old_stories/theme-02-converter.md` in [themeConverter.ts](/modules/themeconverter-ts.md)
- `docs/stories/old_stories/theme-03-store-refactor.md` in [applyTheme](/modules/applytheme.md)
- `docs/stories/old_stories/theme-04-settings-activation.md` in [activateImportedTheme](/modules/activateimportedtheme.md)
- `docs/stories/old_stories/theme-05-monaco-registration.md` in [activateImportedTheme](/modules/activateimportedtheme.md)
- `docs/stories/old_stories/vsix-01-backend-zip-reading.md`
- `docs/stories/old_stories/vsix-02-frontend-vsix-path-handling.md` in [activateImportedTheme](/modules/activateimportedtheme.md)
- `docs/stories/old_stories/vsix-sprint-backlog.md` in [activateImportedTheme](/modules/activateimportedtheme.md)
- `docs/stories/uiadapter-logging-1-infrastructure-and-boot.md`
- `docs/stories/uiadapter-logging-2-plumb-subcomponents.md`
- `docs/stories/uiadapter-logging-3-instrument-cache-and-network.md`
- `docs/stories/uiadapter-logging-4-instrument-pipeline.md`
- `docs/stories/uiadapter-logging-5-instrument-payload-shaping.md`
- `docs/stories/uiadapter-logging-6-tests-and-docs.md`
- `editor_settings_test.go` in [setupTestConfig](/modules/setuptestconfig.md)
- `font_scanner.go` in [font_scanner.go](/modules/font-scanner-go.md)
- `font_scanner_test.go`
- `fonts/CaskaydiaMonoNerdFontMono-Regular.ttf`
- `fonts/FiraCodeNerdFontMono-Regular.ttf`
- `fonts/HackNerdFontMono-Regular.ttf`
- `fonts/InconsolataNerdFontMono-Regular.ttf`
- `fonts/JetBrainsMonoNerdFontMono-Regular.ttf`
- `fonts/MononokiNerdFontMono-Regular.ttf`
- `fonts/SauceCodeProNerdFontMono-Regular.ttf`
- `fonts/SpaceMonoNerdFontMono-Regular.ttf`
- `fonts/UbuntuMonoNerdFontMono-Regular.ttf`
- `fonts/ZedMonoNerdFontMono-Regular.ttf`
- `frontend/.claude/scheduled_tasks.lock`
- `frontend/bun.lock`
- `frontend/coverage/base.css`
- `frontend/coverage/block-navigation.js`
- `frontend/coverage/clover.xml`
- `frontend/coverage/coverage-final.json`
- `frontend/coverage/favicon.png`
- `frontend/coverage/index.html`
- `frontend/coverage/prettify.css`
- `frontend/coverage/prettify.js`
- `frontend/coverage/sort-arrow-sprite.png`
- `frontend/coverage/sorter.js`
- `frontend/dist/.gitkeep`
- `frontend/package-lock.json`
- `frontend/package.json` in [frontend/package.json](/modules/frontend-package-json.md)
- `frontend/package.json.md5`
- `frontend/src/__tests__/entry-animations.test.ts` in [entry-animations.test.ts](/modules/entry-animations-test-ts.md)
- `frontend/src/__tests__/rgba-migration.test.ts` in [rgba-migration.test.ts](/modules/rgba-migration-test-ts.md)
- `frontend/src/__tests__/signature-moments.test.ts` in [signature-moments.test.ts](/modules/signature-moments-test-ts.md)
- `frontend/src/__tests__/sparkline-render.test.ts` in [ref_node_fs](/modules/ref-node-fs.md)
- `frontend/src/__tests__/token-normalization.test.ts` in [token-normalization.test.ts](/modules/token-normalization-test-ts.md)
- `frontend/src/__tests__/vite-keep-dist.test.ts`
- `frontend/src/components/EditorRouter.svelte` in [markdownMenuSettings.ts](/modules/markdownmenusettings-ts.md)
- `frontend/src/components/MarkdownEditor.svelte` in [markdownMenuSettings.ts](/modules/markdownmenusettings-ts.md)
- `frontend/src/components/MonacoEditor.svelte` in [MonacoEditor.svelte](/modules/monacoeditor-svelte.md)
- `frontend/src/components/Terminal.svelte` in [Story: pty-06 — Frontend Terminal and Session Cleanup](/modules/story-pty-06-frontend-terminal-and-session-cleanup.md)
- `frontend/src/components/__tests__/EditorRouter.lazy.test.ts`
- `frontend/src/components/__tests__/MarkdownEditor.guard.test.ts`
- `frontend/src/components/__tests__/MonacoEditor.guard.test.ts`
- `frontend/src/components/__tests__/Terminal.auth.test.ts`
- `frontend/src/components/__tests__/glow-btn.test.ts` in [glow-btn.test.ts](/modules/glow-btn-test-ts.md)
- `frontend/src/components/__tests__/neon-green.test.ts` in [neon-green.test.ts](/modules/neon-green-test-ts.md)
- `frontend/src/components/__tests__/remaining-colors.test.ts` in [remaining-colors.test.ts](/modules/remaining-colors-test-ts.md)
- `frontend/src/components/agent/CommitOutputPanel.svelte`
- `frontend/src/components/agent/FileStrip.svelte`
- `frontend/src/components/agent/SessionTabs.svelte`
- `frontend/src/components/agent/SubAgentPanel.svelte`
- `frontend/src/components/bmad/BuilderToolbar.svelte`
- `frontend/src/components/bmad/ProcessSidebar.svelte` in [svelte](/modules/svelte.md)
- `frontend/src/components/bmad/TerminalModal.svelte`
- `frontend/src/components/feed/AgentList.svelte`
- `frontend/src/components/feed/CommitOutputPanel.svelte`
- `frontend/src/components/feed/RepoActions.svelte`
- `frontend/src/components/feed/RepoHeader.svelte`
- `frontend/src/components/settings/EditorSettings.svelte`
- `frontend/src/components/settings/UIAdapterSettings.svelte`
- `frontend/src/components/settings/VSCodiumThemes.svelte`
- `frontend/src/lib/__tests__/workflowSerialisation.test.ts` in [workflowSerialisation.ts](/modules/workflowserialisation-ts.md)
- `frontend/src/lib/feed/repoTree.ts`
- `frontend/src/lib/workflowBuilder/canvasHandlers.js`
- `frontend/src/lib/workflowBuilder/execEvents.js`
- `frontend/src/lib/workflowBuilder/leaveIntercept.js`
- `frontend/src/views/AgentDetail.svelte` in [App.js](/modules/app-js.md)
- `frontend/src/views/NotificationFeed.svelte` in [svelte](/modules/svelte.md)
- `frontend/src/views/Settings.svelte` in [svelte](/modules/svelte.md)
- `frontend/src/views/WorkflowBuilder.svelte` in [svelte](/modules/svelte.md)
- `frontend/vite.config.js` in [frontend/package.json](/modules/frontend-package-json.md)
- `frontend/wailsjs/go/main/App.d.ts` in [models.ts](/modules/models-ts.md)
- `frontend/wailsjs/go/main/App.js` in [hydrate](/modules/hydrate.md)
- `frontend/wailsjs/go/models.ts` in [WorktreeInfo](/modules/worktreeinfo.md)
- `internal/bmad/artifacts_test.go` in [ResolveArtifactPath](/modules/resolveartifactpath.md)
- `internal/bmad/assets_write_test.go` in [assets_test.go](/modules/assets-test-go.md)
- `internal/bmad/executor.go` in [appendUpstreamContext](/modules/appendupstreamcontext.md)
- `internal/bmad/executor_command.go`
- `internal/bmad/executor_inputs.go`
- `internal/bmad/executor_interactive.go`
- `internal/bmad/executor_lifecycle.go`
- `internal/bmad/executor_lifecycle_test.go`
- `internal/bmad/executor_loaders.go`
- `internal/bmad/executor_loaders_test.go`
- `internal/bmad/executor_node_state.go`
- `internal/bmad/executor_schedule.go`
- `internal/bmad/executor_schedule_test.go`
- `internal/bmad/executor_test.go` in [session_naming_test.go](/modules/session-naming-test-go.md)
- `internal/bmad/question_fixtures_test.go`
- `internal/bmad/question_test.go` in [question_test.go](/modules/question-test-go.md)
- `internal/bmad/session_naming_test.go` in [session_naming_test.go](/modules/session-naming-test-go.md)
- `internal/bmad/skillgen_test.go` in [bmad/registry_test.go](/modules/bmad-registry-test-go.md)
- `internal/bmad/storage.go` in [BmadAgentConfig](/modules/bmadagentconfig.md)
- `internal/bmad/testdata/exec_verify/verify1_slash_injection.txt`
- `internal/bmad/testdata/exec_verify/verify2_idle_stable_pane.txt`
- `internal/bmad/testutil_root_test.go`
- `internal/fsutil/atomic.go`
- `internal/fsutil/atomic_test.go`
- `internal/git/branch.go`
- `internal/git/branch_test.go`
- `internal/git/commit.go`
- `internal/git/commit_test.go`
- `internal/git/files.go`
- `internal/git/files_test.go`
- `internal/git/refs.go`
- `internal/git/refs_test.go`
- `internal/git/remote.go`
- `internal/git/remote_test.go`
- `internal/pathguard/pathguard.go`
- `internal/pathguard/pathguard_test.go`
- `internal/scanner/claude.go` in [ClaudeCodeProvider](/modules/claudecodeprovider.md)
- `internal/scanner/watcher.go` in [AssetWatcher](/modules/assetwatcher.md)
- `internal/terminal/bridge.go` in [mockTmuxSession](/modules/mocktmuxsession.md)
- `internal/terminal/bridge_auth_dev_test.go`
- `internal/terminal/bridge_auth_test.go`
- `internal/terminal/bridge_test.go` in [mockTmuxSession](/modules/mocktmuxsession.md)
- `internal/terminal/manager.go` in [manager_test.go](/modules/manager-test-go.md)
- `internal/terminal/manager_argv_test.go`
- `internal/terminal/origins_dev.go`
- `internal/terminal/origins_prod.go`
- `internal/terminal/session.go` in [session_test.go](/modules/session-test-go.md)
- `internal/terminal/session_test.go` in [session_test.go](/modules/session-test-go.md)
- `justfile`
- `lefthook.yml`
- `main.go` in [main_test.go](/modules/main-test-go.md)
- `main_embed_test.go`
- `main_socket_test.go`
- `markdown_menu_test.go` in [setupTestConfig](/modules/setuptestconfig.md)
- `repo_hygiene_test.go`
- `scripts/api-diff.sh`
- `scripts/check-entry-chunk.mjs`
- `scripts/test-all.sh`
- `theme_scanner.go` in [bundled_themes_test.go](/modules/bundled-themes-test-go.md)
- `theme_scanner_test.go` in [bundled_themes_test.go](/modules/bundled-themes-test-go.md)
- `themes/JuanLias.ultra-instinct-theme-0.1.4.vsix`
- `themes/LhacenMed.cursor-noir-1.0.1.vsix`
- `themes/RINDAMAN2426.ubuntu-aubergine-theme-1.0.1.vsix`
- `themes/SeptWong.vscode-webstorm-theme-1.0.5.vsix`
- `themes/Zhangcy.claude-themes-by-zhangcy-0.0.6.vsix`
- `themes/bastndev.lynx-theme-4.1.0.vsix`
- `themes/jdinhlife.gruvbox-1.29.0.vsix`
- `themes/mhdstk.vibe-black-0.0.10.vsix`
- `themes/monokai.theme-monokai-pro-vscode-2.0.13.vsix`
- `themes/moondevaa.DarkPlusChocolate-0.1.1.vsix`
- `themes/ni3rav.andromeda-night-0.0.7.vsix`
- `themes/oderwat.indent-rainbow-8.3.1.vsix`
- `themes/pittaya-org.pittaya-theme-1.0.1.vsix`
- `themes/prettier.prettier-vscode-12.2.0.vsix`
- `themes/zhuangtongfa.material-theme-3.19.0.vsix`
- `todo.md`
- `wails.json` in [wails.json](/modules/wails-json.md)

# Review
- no review yet

# Status
- intent.md: accepted
- spec.md: accepted
- plan.md: accepted
- test-report: missing or failed
- deployed: nowhere

# Documents
- build: sdlc/repo-health-remediation/docs/build.html (9/9 showcase, 0 errors, 0 warnings)
- design: sdlc/repo-health-remediation/docs/design.html (9/9 showcase, 0 errors, 0 warnings)
- plan: sdlc/repo-health-remediation/docs/plan.html (9/9 showcase, 0 errors, 0 warnings)
