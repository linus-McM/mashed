# Plan: Repo health remediation

**Date:** 2026-09-29
**Status:** draft — awaiting approval
**Owner:** lead
**Source:** full-repo review (build, tests, code read-through) on `main` @ `5ca92ae`

---

## Baseline (measured on a clean clone, Linux, Go 1.25, Node 22)

| Check | Result |
|---|---|
| `go build ./...` | ❌ fails — `main.go:24` embeds `frontend/dist`, which does not exist until the frontend is built |
| `go test ./internal/...` | ⚠️ 4 failures (3 root-only permission tests in `internal/bmad`, 1 flaky PTY test in `internal/terminal`) |
| `go test .` (root package) | ✅ passes once `frontend/dist` exists |
| `npm ci` (frontend) | ❌ `frontend/package-lock.json` out of sync with `package.json` (missing `esbuild@0.28.2`) |
| `vite build` | ✅ builds; one chunk > 2.6 MB |
| `vitest run` | ✅ 515 / 515 |
| `npm run lint:tokens` | ✅ clean |

## Goal

Every check above green on a fresh clone, enforced by CI, with the security issues below closed.

Priorities: **P0** = security / broken feature, **P1** = correctness / build health, **P2** = hygiene, **P3** = structure.

---

## Phase 1 — Security *(P0)*

### Task 1.1 — Lock down the terminal WebSocket

- **Where:** `internal/terminal/bridge.go:43-46` (`wsUpgrader.CheckOrigin` returns `true`), listener at `bridge.go:110`.
- **Problem:** any web page open in the user's browser can connect to `ws://127.0.0.1:<port>/ws/<name>` and type into a live shell or Claude PTY. Session names are guessable (`term-<repo>-<unix>`, `app_spawn.go:21`).
- **Fix:** generate a random per-launch token at startup, pass it to the frontend via a Wails binding, require it as a query param / header on upgrade. Additionally reject origins other than the Wails webview origin (`wails://wails`, `http://wails.localhost`, and `http://localhost:34115` in dev).
- **Acceptance:** a test that upgrades without the token (or with a foreign `Origin`) gets `403`; existing terminal tests updated to pass the token.

### Task 1.2 — Stop git option injection

- **Where:** `app_git.go:212, 242, 619, 625, 666, 851` — branch names / file paths from the frontend passed straight to `git`.
- **Fix:** insert `--` before user-supplied path args; for branch args, reject values starting with `-` (or validate with `git check-ref-format --branch`).
- **Acceptance:** unit test that a branch named `--help` / `-b` is rejected, not executed as a flag.

### Task 1.3 — Restrict file-access bindings

- **Where:** `WriteFile`, `ReadFile`, `ReadFileBase64` (`app_git.go:757-826`); `ReadBundledThemeFile` (`theme_scanner.go:329`) lacks the path check its sibling has.
- **Fix:** resolve with `filepath.Abs` + `filepath.EvalSymlinks`, then require the result to be inside the dev dir, an open repo, or the app config dir. Add the missing check to `ReadBundledThemeFile`.
- **Acceptance:** tests for `../` traversal and symlink escape returning an error.

### Task 1.4 — Smaller hardening

- `main.go:136` / `cmd/pty-helper`: `chmod 0600` the helper Unix socket after creating it.
- `app.go:139` `saveConfig`: write to temp file + `os.Rename` (atomic).
- `GetConfig` / `ReadThemeFile`: take `a.mu` around `loadConfig`.

---

## Phase 2 — Broken features & races *(P0/P1)*

### Task 2.1 — Fix prompt splitting for spawned Claude sessions *(P0)*

- **Where:** `app_git.go:896` (SpawnPRReview), `app_review.go:359` (SpawnRefactorPlan) build `claude ... -p %q` as one string; `internal/terminal/manager.go:58` splits it with `strings.Fields`.
- **Effect:** Claude receives dozens of argv fragments with literal `"` characters instead of one prompt.
- **Fix:** add a spawn path that takes `[]string` args (e.g. `SpawnArgs(name, dir string, argv []string)`), and use it from both callers. Keep the string form for shell commands typed by the user.
- **Acceptance:** test asserting the spawned argv for a prompt containing spaces and quotes is exactly `["claude", ..., "-p", "<prompt>"]`.

### Task 2.2 — Make `SetDevDir` re-entrant *(P1)*

- **Where:** `app.go:353` → `initScanning` (`app_scan.go:20-37`) starts `scanLoop`, `watchSessions`, `consumeEngineEvents` every call with no cancellation. `a.devDir`, `a.provider`, `a.repoScanner` are written unlocked while those goroutines read them (`app_scan.go:108`, `app_git.go:98-122`, `app.go:382`).
- **Fix:** hold a `context.CancelFunc` for the scanning goroutines; cancel + wait before restarting. Guard the three fields with `a.mu` (or an `atomic.Pointer` snapshot struct).
- **Acceptance:** test calling `SetDevDir` N times and asserting goroutine count is stable (`go.uber.org/goleak` is already a dependency); `go test -race` clean.

### Task 2.3 — Asset watcher follows the active repo *(P1)*

- **Where:** `app.go:231` builds watch roots from `a.activeRepoPath` at startup (always empty); `SetActiveContext` (`app.go:284`) never updates the watcher.
- **Fix:** on `SetActiveContext`, add the repo's `.claude/skills` / `.claude/commands` to the watcher (and remove the previous one).

---

## Phase 3 — Build & test health *(P1)*

### Task 3.1 — Fresh-clone `go build` / `go test` works

- Commit `frontend/dist/.gitkeep` (and un-ignore just that file), **or** add a `just test` recipe that builds the frontend first. Prefer the placeholder: it lets `go vet ./...` and editors work without a Node toolchain.

### Task 3.2 — Sync the frontend lockfile

- Pick one package manager. Recommend **npm** (the `justfile`/Wails use it): delete `frontend/bun.lock`, run `npm install` in `frontend/`, commit the updated `package-lock.json`.
- **Acceptance:** `cd frontend && npm ci` succeeds.

### Task 3.3 — Fix failing Go tests

- `internal/bmad`: `TestAC3_VerifyArtifacts_StatErrorNotNotExist`, `TestWriteMashedAssetFrontmatter_WriteFailure`, `TestGenerateSkillFiles_ErrorOnReadOnlyDir` rely on read-only dirs, which root ignores. Add `if os.Geteuid() == 0 { t.Skip("permission test; running as root") }`.
- `internal/terminal`: `TestManagedSession_AC2_TwoClientsReceiveOutput` fails ~4/5 on Linux — `echo hello` exits before clients attach. Attach clients before starting the process, or use a command that waits for input (e.g. `cat`) and write `hello` to it. Also decide whether late-attaching clients should receive scrollback after exit (possible real bug).
- **Acceptance:** `go test -count=10 ./...` green on macOS and in a Linux container.

### Task 3.4 — Add CI

- `.github/workflows/ci.yml` on push + PR:
  - **go job:** `ubuntu-latest`, `actions/setup-go` (1.25), create `frontend/dist` placeholder, `go vet ./...`, `go test -race ./...`.
  - **frontend job:** `actions/setup-node` (22), `npm ci`, `npm run lint:tokens`, `npx vitest run`, `npm run build`.
  - Optional `macos-latest` job for PTY / `screencapture`-dependent tests.
- Make lefthook's pre-push run `go test ./...` (currently only `./internal/...`, `lefthook.yml:52`) so the root package is covered locally too.

---

## Phase 4 — Repo hygiene *(P2)*

### Task 4.1 — Untrack junk (~40 MB of text + generated files)

`git rm -r --cached` then fix `.gitignore` so they stay out:

- `docs/repomixer/**/*.xml` (~34.6 MB; `wails.xml` alone 27 MB)
- `.playwright-mcp/` (98 files), `.playwright-cli/`
- `desloppify-workspace/`
- `frontend/coverage/` (contains `/Users/linus/...` absolute paths)
- `.vite/`
- `frontend/.claude/scheduled_tasks.lock`
- Review the triple-negation `.claude/skills` rule in `.gitignore` — 43 vendored skill files are tracked despite it; decide intentionally which skills ship with the repo.

History rewrite (`git filter-repo`) is optional; untracking is enough to stop growth.

### Task 4.2 — Large binaries

- `themes/*.vsix` (~36 MB) and `fonts/*.ttf` (~30 MB): move to Git LFS, or download at build time via a `just fetch-assets` recipe.

### Task 4.3 — De-personalise config

- `justfile:43,51,52`: replace the hard-coded `Apple Development: linus McManamey (5X8A9U965U)` with `env_var_or_default("MASHED_SIGN_IDENTITY", "-")` (`-` = ad-hoc signing, works for local dev).
- `wails.json:11`: personal email → project/contact alias if the repo will be public.

### Task 4.4 — Repo basics

- Add a `LICENSE`.
- Pick one hook system: `lefthook.yml` **or** `.githooks/` — remove the other. Re-enable or delete the commented-out desloppify step (`lefthook.yml:34-47`).
- Add `.golangci.yml` (start with `govet`, `errcheck`, `staticcheck`, `gosec`) and a Prettier / `svelte-check` step for the frontend.
- `todo.md` is both tracked and gitignored — move items into `docs/stories/` or GitHub issues and delete the file.

---

## Phase 5 — Structure *(P3, after phases 1-4)*

- Split `internal/bmad/executor.go` (1913 lines) by concern (scheduling, node execution, artifact propagation, events).
- Split `frontend/src/views/NotificationFeed.svelte` (1955), `WorkflowBuilder.svelte` (1643), `AgentDetail.svelte` (1306), `Settings.svelte` (1014) into child components.
- `app_git.go` (898 lines) duplicates `internal/git`; move shell-outs into `internal/git` and dedupe the auto-commit-then-checkout block repeated at `:199`, `:221`, `:592`.
- Remove the no-op `recoverSessions` stub (`app_terminal_registry.go:16`) or implement it using `recoverSessionsFromOutput`.
- Code-split Monaco / Milkdown with dynamic `import()` to bring the main chunk under the Vite warning limit.

## Platform note

The app is macOS-only today (`screencapture` `app.go:314`, `open` `font_scanner.go:166`, `codesign` in `justfile`, `mac.Options` in `main.go`, `lsof` in `internal/scanner/processes.go:148`, Unix-socket PTY helper). Not in scope here, but guard these behind build tags (`//go:build darwin`) if Linux/Windows support is ever wanted — it also makes Linux CI cleaner.

---

## Suggested order

1. Task 2.1 — prompt splitting (small, user-visible)
2. Task 1.1 — WebSocket token/origin
3. Task 1.2 — git `--`
4. Tasks 3.1 + 3.2 — fresh-clone build + lockfile
5. Task 3.3 — fix failing tests
6. Task 3.4 — CI
7. Task 4.1 — untrack junk
8. Remaining Phase 1/2 tasks, then Phase 4, then Phase 5

## Local setup checklist (for picking this up)

```sh
git pull origin main
cd frontend && npm install && cd ..   # regenerates lockfile (Task 3.2)
just dev                              # builds pty-helper + runs wails dev
go test ./...                         # after `npm run build` or the dist placeholder
cd frontend && npx vitest run
```
