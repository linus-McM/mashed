# svelte-check baseline

Snapshot of the svelte-check error count on `dev` at the time the ratchet
infrastructure landed. This file is the single source of truth parsed by
`just sveltecheck-ratchet` — the fenced line below is consumed by shell.

```
baseline: 1520
```

- **Date:** 2026-04-22
- **Commit:** `9fc63d7e03fb924f1e28b391d5600f5af47634be` (dev)
- **Recipe used:** `just sveltecheck-count`

## How to regenerate

1. Check out the branch that drops the count.
2. Run `just sveltecheck-count` and capture the integer output.
3. Update the fenced `baseline: <N>` line above to the new ceiling.
4. Update the date and commit SHA to the landing commit.
5. Commit the change alongside the migration work.

The ratchet allows the live count to drop at any time; only commits that
raise the count above `baseline` are rejected. The baseline must never
increase after this story — subsequent migration stories ratchet it
downward only.

## Ratchet contract

`just sveltecheck-ratchet` (defined in the repo `justfile`):

- Reads the `baseline:` line from this file (strict `^baseline: [0-9]+$`).
- Runs `just sveltecheck-count` unless the `TEST_COUNT` env var is set,
  in which case the integer in `TEST_COUNT` is used (test harness hook).
- Honours `BASELINE_FILE` env var to override the baseline path (test
  harness hook).
- Exit `0` when `current <= baseline`.
- Exit `1` with stderr `count rose above baseline` when `current > baseline`.
- Exit `2` with stderr `baseline file missing` when the file is absent.
