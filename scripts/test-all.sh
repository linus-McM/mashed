#!/usr/bin/env bash
# Full test suite used by the sdlc red/green gate.
# Strips Orca's tmux shim from PATH: it answers `has-session` with 0 for any
# name, which breaks internal/bmad/resume_ghost_test.go.
set -euo pipefail
cd "$(dirname "$0")/.."
PATH="$(printf '%s' "$PATH" | tr ':' '\n' | grep -v '/\.orca/' | paste -sd: -)"
export PATH
go test -race -count=1 ./...
(cd frontend && npx vitest run)
