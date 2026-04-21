#!/usr/bin/env bash
set -euo pipefail

# Coverage threshold enforcement for Go packages
# Used by: lefthook pre-push, just test-cover

# Parallel arrays (bash 3.2 compatible — no associative arrays on macOS)
packages=(
  "internal/bmad"
  "internal/agent"
  "internal/scanner"
  "internal/git"
)
thresholds=(
  85
  55
  30
  30
)

tmpdir=$(mktemp -d)
trap 'rm -rf "$tmpdir"' EXIT

failed=0

echo "=== Go Coverage Threshold Check ==="

for i in "${!packages[@]}"; do
  pkg="${packages[$i]}"
  threshold="${thresholds[$i]}"
  coverfile="$tmpdir/${pkg//\//_}.out"

  # set +e: go test may exit non-zero for low coverage; we parse the profile regardless.
  # -tags testing compiles files behind //go:build testing (ui-ast-U4 MockAdapter +
  # adapter/flatten/gate tests). Without the tag, those tests are silently excluded.
  set +e
  go test -tags testing -short -coverprofile="$coverfile" -count=1 "./$pkg/..." > /dev/null 2>&1
  set -e

  # Extract coverage percentage
  if [[ -s "$coverfile" ]]; then
    actual=$(go tool cover -func="$coverfile" | awk '/^total:/ { gsub(/%/, "", $3); print $3 }')
  else
    actual=0
  fi

  # Compare using awk (bash can't do float math)
  pass=$(awk "BEGIN { print ($actual >= $threshold) ? 1 : 0 }")

  if [[ "$pass" -eq 1 ]]; then
    printf "%-20s %5s%% >= %d%%\tPASS\n" "$pkg:" "$actual" "$threshold"
  else
    printf "%-20s %5s%% <  %d%%\tFAIL\n" "$pkg:" "$actual" "$threshold"
    failed=$((failed + 1))
  fi
done

if [[ "$failed" -gt 0 ]]; then
  echo "=== $failed package(s) below threshold ==="
  exit 1
else
  echo "=== All packages passed ==="
  exit 0
fi
