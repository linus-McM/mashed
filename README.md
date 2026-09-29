# Mashed

A notification-first IDE for multi-agent development. Built for solo AI-heavy developers running multiple Claude Code (and other agent) sessions in parallel.

Mashed is a [Wails v2](https://wails.io) desktop application: a single Go binary that embeds a Svelte + TypeScript frontend and ships as a signed macOS app.

## What it does

- **Notification feed** — single-pane command center for every agent session, diff, and alert.
- **Multi-agent orchestration** — spawn, watch, and switch between Claude Code sessions (Opus / Sonnet / Haiku), Gemini, and tmux-based swarms.
- **BMAD workflow canvas** — graph-based artifact pipelines (`_bmad-output/`) wired through typed nodes in `internal/bmad/`.
- **Git panel** — scoped diffs, branch info, review flows, worktree awareness (`app_git.go`, `app_review.go`).
- **Embedded terminal** — PTY-backed panes via the sandboxed `mashed-pty-helper` (required for macOS Sequoia PTY entitlements).
- **Process scanner** — discovers running agents across tmux sessions and the process tree.
- **Editor + themes** — Monaco editor, VSCodium theme scanning, Nerd Font detection.
- **Screenshots** — one-shot capture of the active pane with context tagging.

## Tech stack

| Layer     | Stack                                                                 |
|-----------|-----------------------------------------------------------------------|
| Desktop   | Wails v2 (`github.com/wailsapp/wails/v2`)                             |
| Backend   | Go 1.25, testify, fsnotify, gorilla/websocket, creack/pty, yaml.v3    |
| Frontend  | Svelte + Vite + TypeScript, Monaco editor                             |
| Design    | Geist / Geist Mono / JetBrains Mono, dark-only, neon-green accent     |
| Platform  | macOS (Apple-signed `.app` with hardened runtime + entitlements)      |

Full inventory: see `go.mod`, `frontend/package.json`, and `DESIGN.md`.

## Repository layout

```
.
├── cmd/                  # Go entry points (pty-helper, etc.)
├── internal/             # Go packages — bmad, git, terminal, scanner, etc.
├── frontend/             # Svelte + Vite app
│   └── src/
│       ├── components/   # BMAD canvas, GitPanel, MonacoEditor, ProcessSidebar
│       ├── views/        # NotificationFeed, WorkflowBuilder, AgentDetail, Settings
│       └── lib/          # stores, utilities
├── build/                # Wails build output + darwin entitlements
├── fonts/                # bundled Geist / JetBrains Mono
├── scripts/              # coverage, tmux viewer helpers
├── tests/ac/             # Playwright acceptance tests
├── main.go               # Wails bootstrap + native menu
├── app*.go               # Wails bindings (git, bmad, review, terminal, …)
├── wails.json            # Wails config
├── justfile              # dev / build / test recipes
└── .wolf/                # OpenWolf context (anatomy, cerebrum, memory)
```

Canonical per-file navigation lives in `.wolf/anatomy.md` (auto-maintained by OpenWolf).

## Getting started

### Prerequisites

- macOS (Darwin 25+ recommended — Sequoia PTY entitlements assumed)
- Go 1.25+
- Node.js 20+ and npm
- [Wails CLI v2.12+](https://wails.io/docs/gettingstarted/installation)
- [just](https://github.com/casey/just) (task runner)
- Codesigning identity — `just build` signs ad hoc (`-`) by default, which works for local builds; set `MASHED_SIGN_IDENTITY` to your Apple Development identity to sign for release.

### Develop

```sh
just dev          # builds pty-helper, starts `wails dev` with HMR
```

The Vite dev server binds automatically; Wails exposes Go methods on `http://localhost:34115` for browser devtools.

### Build + run

```sh
just build-helper # build + sign mashed-pty-helper only
just build        # full build: helper + frontend + wails bundle + codesign
just run          # build and launch the .app
```

Output: `build/bin/mashed.app`.

### Test

```sh
just test         # Go tests (internal/...)
just test-all     # Go + frontend vitest
just test-cover   # per-package Go coverage gate (scripts/check-coverage.sh)
just lint         # go vet ./...
just pre-check    # run lefthook pre-commit hooks
```

Playwright AC specs live in `tests/ac/` and use `playwright.config.ts`.

## Keyboard shortcuts (native menu)

| Shortcut         | Action                   |
|------------------|--------------------------|
| `⌘ N`            | New Agent                |
| `⌘ ⇧ N`          | New Repository…          |
| `⌘ O`            | Open Workspace…          |
| `⌘ ⇧ S`          | Take Screenshot          |
| `⌘ 1` / `⌘ 2`    | Feed / Workflows         |
| `⌘ ,`            | Settings                 |
| `⌘ M`            | Minimize                 |
| `⌃ ⌘ F`          | Toggle Fullscreen        |
| `⌘ Q`            | Quit                     |

## BMAD artifact system

- Canonical path map and resolver in `internal/bmad/artifacts.go`.
- Artifacts resolve to `{repoPath}/_bmad-output/{category}/{artifact}`.
- `ResolveArtifactPath(name, repoPath)` returns `""` for unmapped artifacts (`"code"`, `"tests"`, `"any-doc"`) — callers handle.
- Shared constant `bmadOutputDir = "_bmad-output"` in `internal/bmad/sprint.go`.

## Design system

Dark-only, industrial/utilitarian. Full spec in `DESIGN.md`:

- Backgrounds `#07080a → #0d0f12 → #12151a → #181c23`
- Accent `#00e57a` (neon green), alert `#f0a500`, error `#e84545`, info `#3d9eff`, agent `#9d6fff`
- 4px spacing base, compact density
- Motion capped at 150ms

## Project conventions

- OpenWolf-managed — read `.wolf/OPENWOLF.md` each session; consult `.wolf/anatomy.md` before opening files and `.wolf/cerebrum.md` before writing code.
- Bugs and recurring fixes live in `.wolf/buglog.json`.
- Sprint/story artifacts live in `_bmad-output/`.

## Debug logging

The UI adapter ships structured `slog` debug logging. Set `UIADAPTER_LOG_LEVEL=debug`
to see per-request traces; default `info` keeps production quiet. Records fan out to
stdout (human-readable) and `./logs/uiadapter-YYYYMMDD.log` (JSON, machine-grep).
Sanitize discipline is strict: lengths, hashes, durations, IDs only — never raw
payloads. Tail with:

    tail -f logs/uiadapter-*.log | jq 'select(.op=="client.chat")'

Shortcut: `just trace` runs `wails dev` with debug enabled and tees both streams.

## License

See `LICENSE` (when present). Repo URL in-app: set by `repoURL` in `main.go`.
