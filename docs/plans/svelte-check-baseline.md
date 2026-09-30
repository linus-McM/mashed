# svelte-check baseline

Snapshot of the svelte-check error count. **MIGRATION COMPLETE** as of story
`svelte-check-07` (2026-04-22) — the baseline is now `0` and CI/lefthook run
strict `just sveltecheck` instead of `just sveltecheck-ratchet`.

```
baseline: 0
```

- **Date:** 2026-04-22 (migration complete)
- **Recipe used:** `just sveltecheck` (strict zero-error gate)
- **Starting count:** 1520 (Phase 0 baseline, 2026-04-22)
- **Final count:** 0 (Phase 7 landing)

## Migration complete

Phases 0→7 landed the full TypeScript retype of the BMAD frontend. From this
story onward the ratchet infrastructure is retired — any commit that
introduces a single svelte-check error is blocked at pre-commit by lefthook
and in CI by `.github/workflows/svelte-check.yml`.

- CI runs `just sveltecheck` (strict) with `continue-on-error: false`.
- Pre-commit runs `just sveltecheck` via lefthook.
- The `sveltecheck-ratchet` recipe has been removed from `justfile` —
  `just sveltecheck` is the canonical gate.

If a future migration wave ever needs to reintroduce a ratchet (e.g. after
upgrading to a new svelte-check version that surfaces fresh issues), revive
the recipe from the git history of story `svelte-check-00` and set
`baseline:` above to the new starting ceiling.
