# Sprint Backlog: Lefthook Pre-commit Harness + Testing Infrastructure

## Sprint Backlog

| # | Story | Priority | Domain | Size | Depends On |
|---|-------|----------|--------|------|------------|
| 1 | Coverage enforcement script | P0 | backend | S | none |
| 2 | PTY test isolation with testing.Short() guards | P0 | backend | S | none |
| 3 | Frontend testing infrastructure with Vitest | P0 | frontend | M | none |
| 4 | Lefthook pre-commit and pre-push configuration | P0 | fullstack | M | Stories 1, 2, 3 |
| 5 | Justfile targets for testing and hook management | P1 | fullstack | S | Stories 1, 3, 4 |

**Total Stories:** 5
**Ready for Sprint:** Stories 1, 2, 3 (no dependencies), then Story 4, then Story 5
**Recommended Sprint Order:** 1, 2, 3 (parallel), 4, 5

## Dependency Graph

```
Story 1 (coverage script) ──┐
Story 2 (PTY test isolation) ┼──> Story 4 (lefthook config) ──> Story 5 (justfile targets)
Story 3 (frontend vitest)  ──┘
```

Stories 1, 2, and 3 are independent and can be executed in parallel by separate agents. Story 4 depends on all three. Story 5 depends on Stories 1, 3, and 4.

## Files Created/Modified Summary

| Story | Action | File | Purpose |
|-------|--------|------|---------|
| 1 | Create | `scripts/check-coverage.sh` | Per-package Go coverage threshold enforcement |
| 2 | Edit | `internal/terminal/session_test.go` | Add testing.Short() skip guards |
| 2 | Edit | `internal/terminal/manager_test.go` | Add testing.Short() skip guards |
| 2 | Edit | `internal/terminal/bridge_test.go` | Assess and optionally add skip guards |
| 3 | Create | `frontend/vitest.config.js` | Vitest configuration for Svelte 4 |
| 3 | Edit | `frontend/package.json` | Add vitest devDep, test/check scripts |
| 3 | Edit | `frontend/src/lib/themeConverter.test.js` | Migrate node:test to vitest |
| 4 | Create | `lefthook.yml` | Pre-commit and pre-push hook definitions |
| 5 | Edit | `justfile` | Add test-all, test-cover, lint, hooks-install, pre-check targets |

## Sprint Notes

- **Parallelism opportunity:** Stories 1, 2, and 3 can be assigned to separate agents and executed simultaneously. This reduces wall-clock sprint time from 5 sequential stories to 3 sequential phases (1+2+3 parallel, then 4, then 5).
- **Skills required:** `/golang-testing` for Stories 1 and 2, frontend/Vitest knowledge for Story 3, shell scripting for Stories 1 and 4.
- **Risk:** Story 4 is the integration point -- if any of Stories 1-3 have issues, Story 4 will surface them during `lefthook run pre-commit` / `lefthook run pre-push` verification.
- **Prerequisites:** `lefthook` must be installed on the developer's machine (`brew install lefthook` or `go install github.com/evilmartians/lefthook@latest`).

## Story Files

- `/Users/linus/Development/mashed/docs/stories/lefthook-01-coverage-script.md`
- `/Users/linus/Development/mashed/docs/stories/lefthook-02-pty-test-isolation.md`
- `/Users/linus/Development/mashed/docs/stories/lefthook-03-frontend-testing.md`
- `/Users/linus/Development/mashed/docs/stories/lefthook-04-lefthook-config.md`
- `/Users/linus/Development/mashed/docs/stories/lefthook-05-justfile-targets.md`
