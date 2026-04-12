# Sprint Backlog: Code Review Summarisation Panel

## Sprint Backlog

| # | Story | Priority | Domain | Size | Depends On | Key Files |
|---|-------|----------|--------|------|------------|-----------|
| 1 | [review-01](review-01-advice-backend.md): Advice Mode Discovery & Parsing Backend | P0 | backend | M | none | `internal/advice/types.go`, `internal/advice/loader.go` |
| 2 | [review-02](review-02-review-backend.md): Code Review Summary & Advice Streaming Backend | P0 | backend | L | review-01 | `app_review.go` |
| 3 | [review-03](review-03-summarisation-modal.md): Summarisation Modal Frontend | P0 | frontend | L | review-02 | `frontend/src/views/SummarisationModal.svelte` |
| 4 | [review-04](review-04-refactor-plan-agent.md): Refactor Plan Agent | P1 | fullstack | M | review-02, review-03 | `app_review.go`, `SummarisationModal.svelte` |
| 5 | [review-05](review-05-bundled-advice-files.md): Bundled Default Advice Files | P1 | backend | S | review-01 | `internal/advice/defaults/*.md` |
| 6 | [review-06](review-06-gitpanel-integration.md): GitPanel Integration & Wiring | P1 | fullstack | S | review-03 | `frontend/src/components/bmad/GitPanel.svelte` |

**Total Stories:** 6
**Ready for Sprint:** Stories 1, 2, 3, 4, 5, 6 (all status: ready)

## Dependency Graph

```
review-01 (Advice Backend)
  |
  +---> review-02 (Review Backend)
  |       |
  |       +---> review-03 (Summarisation Modal)
  |       |       |
  |       |       +---> review-04 (Refactor Plan Agent)
  |       |       |
  |       |       +---> review-06 (GitPanel Integration)
  |       |
  |       +---> review-04 (Refactor Plan Agent)
  |
  +---> review-05 (Bundled Advice Files)
```

## Recommended Sprint Order

1. **review-01** -- Foundation types and loading logic (no dependencies)
2. **review-05** -- Bundled advice files (depends on review-01, can run in parallel with review-02)
3. **review-02** -- Streaming backend (depends on review-01)
4. **review-03** -- Frontend modal (depends on review-02)
5. **review-04** -- Refactor plan agent (depends on review-02 + review-03)
6. **review-06** -- GitPanel wiring (depends on review-03, can run in parallel with review-04)

## Parallelization Opportunities

- **review-01** then **review-02 + review-05** in parallel (different packages, no file conflicts)
- **review-04 + review-06** in parallel after review-03 (review-04 modifies `app_review.go`, review-06 modifies `GitPanel.svelte` -- no conflicts)

## Risk Summary

| Risk | Mitigation |
|------|------------|
| Claude CLI not on PATH | All streaming methods check `exec.LookPath` first, emit error events |
| Large diffs overload Claude | Per-file truncation (500 lines), file cap (50 files) |
| go:embed empty directory | Use `all:defaults` pattern; review-05 populates files |
| Wails bindings stale | review-06 explicitly includes binding regeneration task |
| Concurrent review race | `sync.Map` guard in review-02 prevents double-run |
