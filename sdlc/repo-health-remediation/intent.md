# Intent: Repo health remediation
Author: linus. Status: accepted. Risk: high.

## Problem
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

## Proposed outcome
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

## Affected users and systems
People:
- The local desktop user: terminal pane (`frontend/src/components/Terminal.svelte` must send the token), PR-review and refactor-plan actions (GitPanel), the code editor (`CodeEditor.svelte`), dev-dir and active-repo switching.
- Contributors and CI: new `ci.yml`, changed pre-push hook, npm-only frontend, Git LFS for binaries.

Systems:
- `internal/terminal` (`bridge.go` auth, `manager.go` argv spawn API), `app_spawn.go`.
- Root Wails bindings: `app_git.go`, `app_review.go`, `theme_scanner.go`, `app.go`, `app_scan.go`, `app_terminal_registry.go`.
- `main.go` (embed, pty-helper socket), `cmd/pty-helper`.
- `internal/git`, `internal/bmad/executor.go` (split only), large Svelte views in `frontend/src/views/`.
- Build and repo config: `justfile`, `wails.json`, `lefthook.yml` / `.githooks/`, `.gitignore`, `.gitattributes` (LFS), `frontend/package-lock.json`, `frontend/bun.lock`, `.github/workflows/`.
- External tools: git CLI, claude CLI, npm, Git LFS, GitHub Actions.

Stays the same: BMAD interactive execution behaviour and its invariants (`NodeAwaitingInput` as single pause state, §14.x guards); macOS-only integrations.

## Constraints
- Allowed roots for file bindings: anything under `$HOME`, resolved with `filepath.Abs` + `filepath.EvalSymlinks` and a separator-prefix check; paths outside `$HOME` and symlink escapes are rejected. Reuse the existing guard pattern (`theme_scanner.go:432-443`, `internal/bmad/validate.go`).
- Personal data is removed from HEAD only. No history rewrite, no force-push of `main`.
- Git LFS applies to themes/fonts going forward; existing blobs stay in history (no rewrite).
- Token and `Terminal.svelte` change ship together, or the terminal breaks.
- New `.github/workflows/ci.yml` (Go + frontend jobs, `ubuntu-latest`) on push and PR to `main` and `dev`; the existing `svelte-check.yml` is kept or folded in, not silently dropped.
- Delivery: one PR per phase, merged in order 1 → 5. Within phase 2, the argv spawn API lands before its callers.
- Phase 5 refactors must not change behaviour; existing tests stay green and BMAD invariants in `CLAUDE.md` hold.
- Reuse existing helpers: atomic temp-file + rename (`internal/bmad/storage.go`), `go.uber.org/goleak` (already in `go.mod`).
- The app stays macOS-only; Linux/Windows build tags are out of scope.

## Open questions
- Which origins the Wails webview actually sends (`wails://wails`, `http://wails.localhost`, dev `http://localhost:34115`) must be verified during design.
- Whether late-attaching terminal clients should receive scrollback after the process exits (possible real bug behind the flaky PTY test) — decide in design.
- Which vendored `.claude/skills` files ship with the repo (Task 4.1 `.gitignore` rule) — decide in design.
- Which LICENSE to add — owner to choose in design.
