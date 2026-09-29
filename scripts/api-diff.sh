#!/usr/bin/env bash
# Compare the exported API of a Go package between a base ref and the working
# tree (repo health R30: file splits must not change the API).
# Usage: scripts/api-diff.sh <base-ref> [package]   (default ./internal/bmad)
# Prints nothing and exits 0 when the documented API is identical.
set -euo pipefail
base="${1:?usage: api-diff.sh <base-ref> [package]}"
pkg="${2:-./internal/bmad}"
root="$(cd "$(dirname "$0")/.." && pwd)"
tmp="$(mktemp -d)"
trap 'git -C "$root" worktree remove --force "$tmp/base" >/dev/null 2>&1 || true; rm -rf "$tmp"' EXIT

git -C "$root" worktree add --detach -q "$tmp/base" "$base"
# go doc output order follows files; sort so pure moves compare equal.
(cd "$tmp/base" && go doc -all "$pkg" | sort) > "$tmp/base.txt"
(cd "$root" && go doc -all "$pkg" | sort) > "$tmp/head.txt"
diff -u "$tmp/base.txt" "$tmp/head.txt"
