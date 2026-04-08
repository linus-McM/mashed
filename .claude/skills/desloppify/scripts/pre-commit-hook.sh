#!/usr/bin/env bash
# Desloppify pre-commit hook
# Runs a scan and reports the strict score before each commit.
# Warns on score drops but does not block the commit.

set -euo pipefail

VENV_DIR=".venv-desloppify"

# Skip if desloppify isn't set up
if [ ! -d "$VENV_DIR" ]; then
  exit 0
fi

# Activate venv
# shellcheck disable=SC1091
source "$VENV_DIR/bin/activate"

echo "=== desloppify: scanning before commit ==="

# Run scan (suppress non-zero exit from findings)
desloppify scan --path . 2>&1 || true

# Report the strict score
SCORE=$(desloppify status --json 2>/dev/null \
  | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('strict_score', 'N/A'))" \
  2>/dev/null || echo "N/A")

echo "=== desloppify strict score: $SCORE ==="

deactivate
exit 0
