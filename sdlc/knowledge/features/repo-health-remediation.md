---
type: Feature
title: Repo health remediation
description: "The Mashed desktop app (Wails, Go + Svelte, macOS) has security holes, broken features and a build that does not work from a fresh clone. Source review: `docs/plans/repo-health-remediation.md` (measur"
resource: sdlc/repo-health-remediation
tags: [feature, accepted]
status: stable
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:16:08Z" }
stale_after: "2026-10-13T11:16:08Z"
source_commit: 85e7124b5a77a2f29bd9db818781712ea7a1049b
sources:
  - { id: intent, resource: sdlc/repo-health-remediation/intent.md, last_modified: "2026-09-29T20:00:59+10:00", digest: 476d74d6341cb18a }
verified:
  - { by: "human:linus-mcmanamey", at: "2026-09-29T11:16:08Z" }
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
- spec.md not written yet

# Files
- plan.md not written yet

# Review
- no review yet

# Status
- intent.md: accepted
- spec.md: missing
- plan.md: missing
- test-report: missing or failed
- deployed: nowhere

# Documents
- plan: sdlc/repo-health-remediation/docs/plan.html (9/9 showcase, 0 errors, 0 warnings)
