#!/usr/bin/env bash
# Test harness for `just sveltecheck-ratchet`.
#
# Story: docs/stories/svelte-check-00-scaffolding.md (Task 1, RED phase)
# Covers AC-2 scenarios:
#   - Scenario 1: count == baseline -> exit 0
#   - Scenario 2: count >  baseline -> exit non-zero, stderr "count rose above baseline"
#   - Scenario 3: count <  baseline -> exit 0 (decrease always allowed)
#   - Scenario 4: baseline file missing -> exit 2, stderr "baseline file missing"
#
# CONTRACT WITH IMPLEMENTER (task 3, fullstack-eng):
# The `just sveltecheck-ratchet` recipe MUST honour two environment overrides
# so this harness can exercise it deterministically without re-running the
# (~30s) live svelte-check pass:
#
#   TEST_COUNT     If set, the recipe SKIPS calling `just sveltecheck-count`
#                  and treats TEST_COUNT as the live count.
#   BASELINE_FILE  If set, the recipe reads the baseline integer from this
#                  path INSTEAD of `docs/plans/svelte-check-baseline.md`.
#
# When both env vars are unset, the recipe MUST fall back to the live count
# and the canonical baseline file (production behaviour).
#
# Baseline file format (canonical and override): a line matching
#   ^baseline:\s*([0-9]+)\s*$
# is parsed for the integer. Surrounding markdown/prose is allowed.
#
# Failure contract:
#   - count > baseline       -> exit code 1 (any non-zero acceptable),
#                               stderr matches /count rose above baseline/
#   - baseline file missing  -> exit code 2, stderr matches /baseline file missing/

set -euo pipefail

# ---------------------------------------------------------------------------
# Locate repo root so the harness runs from any cwd.
# ---------------------------------------------------------------------------
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

# ---------------------------------------------------------------------------
# Workspace for temporary baseline files. Cleaned up on exit.
# ---------------------------------------------------------------------------
TMP_DIR="$(mktemp -d -t sveltecheck-ratchet-XXXXXX)"
cleanup() { rm -rf "${TMP_DIR}"; }
trap cleanup EXIT

PASS=0
FAIL=0

# ---------------------------------------------------------------------------
# run_case <name> <expected_exit> <stderr_regex_or_empty> <env-assignments...>
#
# Invokes `just sveltecheck-ratchet` from the repo root with the supplied
# `KEY=VALUE` env assignments. Records PASS/FAIL based on exit code and
# (optional) stderr regex match.
# ---------------------------------------------------------------------------
run_case() {
  local name="$1"; shift
  local want_exit="$1"; shift
  local stderr_regex="$1"; shift
  # Remaining args are KEY=VALUE pairs forwarded to the recipe via `env`.

  local stdout_log="${TMP_DIR}/stdout.log"
  local stderr_log="${TMP_DIR}/stderr.log"
  local got_exit=0

  # `env -i` would also strip PATH, so use plain `env` to keep PATH/HOME but
  # add the test-only overrides on top.
  ( cd "${REPO_ROOT}" && env "$@" just sveltecheck-ratchet ) \
    >"${stdout_log}" 2>"${stderr_log}" || got_exit=$?

  local ok=1
  local diag=""

  if [[ "${want_exit}" == "0" ]]; then
    if [[ "${got_exit}" -ne 0 ]]; then
      ok=0
      diag+="expected exit 0, got ${got_exit}; "
    fi
  elif [[ "${want_exit}" == "nonzero" ]]; then
    if [[ "${got_exit}" -eq 0 ]]; then
      ok=0
      diag+="expected non-zero exit, got 0; "
    fi
  else
    if [[ "${got_exit}" -ne "${want_exit}" ]]; then
      ok=0
      diag+="expected exit ${want_exit}, got ${got_exit}; "
    fi
  fi

  if [[ -n "${stderr_regex}" ]]; then
    if ! grep -Eq "${stderr_regex}" "${stderr_log}"; then
      ok=0
      diag+="stderr did not match /${stderr_regex}/; "
    fi
  fi

  if (( ok == 1 )); then
    PASS=$((PASS + 1))
    printf '✓ PASS  %s\n' "${name}"
  else
    FAIL=$((FAIL + 1))
    printf '✗ FAIL  %s\n' "${name}"
    printf '        %s\n' "${diag}"
    printf '        --- stdout ---\n'
    sed 's/^/        /' "${stdout_log}" || true
    printf '        --- stderr ---\n'
    sed 's/^/        /' "${stderr_log}" || true
  fi
}

# ---------------------------------------------------------------------------
# Helpers for building baseline files in TMP_DIR.
# ---------------------------------------------------------------------------
write_baseline() {
  local path="$1"
  local value="$2"
  cat >"${path}" <<EOF
# svelte-check baseline (test fixture)

baseline: ${value}
EOF
}

BASELINE_AT_1520="${TMP_DIR}/baseline-1520.md"
write_baseline "${BASELINE_AT_1520}" 1520

MISSING_BASELINE="${TMP_DIR}/does-not-exist.md"
# Intentionally NOT created.

# ---------------------------------------------------------------------------
# Scenarios
# ---------------------------------------------------------------------------

# Scenario 1: count == baseline -> exit 0
run_case "scenario 1: count == baseline (1520 == 1520) exits 0" \
  0 "" \
  "TEST_COUNT=1520" "BASELINE_FILE=${BASELINE_AT_1520}"

# Scenario 2: count > baseline -> non-zero + stderr regex
run_case "scenario 2: count > baseline (1521 > 1520) exits non-zero with diagnostic" \
  nonzero "count rose above baseline" \
  "TEST_COUNT=1521" "BASELINE_FILE=${BASELINE_AT_1520}"

# Scenario 3: count < baseline -> exit 0 (decrease always OK)
run_case "scenario 3: count < baseline (1320 < 1520) exits 0" \
  0 "" \
  "TEST_COUNT=1320" "BASELINE_FILE=${BASELINE_AT_1520}"

# Scenario 4: baseline file missing -> exit 2 + stderr regex
run_case "scenario 4: baseline file missing exits 2 with diagnostic" \
  2 "baseline file missing" \
  "TEST_COUNT=1520" "BASELINE_FILE=${MISSING_BASELINE}"

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
TOTAL=$((PASS + FAIL))
printf '\n'
printf '%d/%d ratchet scenarios passed\n' "${PASS}" "${TOTAL}"

if (( FAIL > 0 )); then
  exit 1
fi

echo "all ${PASS} ratchet scenarios passed"
