# Sprint Backlog: VSCodium Theme Loading

## Stories

| # | Story | Priority | Domain | Size | Depends On |
|---|-------|----------|--------|------|------------|
| 1 | [Go Backend: Extension Scanner & Theme Reader](theme-01-backend-scanner.md) | P0 | backend | M | none |
| 2 | [Frontend: Theme Converter (B2 Hand-Rolled)](theme-02-converter.md) | P0 | frontend | M | none |
| 3 | [Theme Store Refactor & Consumer Updates](theme-03-store-refactor.md) | P0 | frontend | M | none |
| 4 | [Settings UI: Theme Scanning, Activation & Startup Restore](theme-04-settings-activation.md) | P1 | fullstack | L | Story 1, Story 2, Story 3 |
| 5 | [Monaco Runtime Registration, Validation & Error Recovery](theme-05-monaco-registration.md) | P1 | frontend | S | Story 3, Story 4 |

## Review Issue Coverage

All 5 critical and 7 high-priority issues from the adversarial review are addressed:

| Issue | Severity | Story | Fix |
|-------|----------|-------|-----|
| C-1: Path traversal via symlinks | CRITICAL | Story 1 | `filepath.EvalSymlinks` + separator suffix |
| C-2: TitleBar breakage after store refactor | CRITICAL | Story 3 | Update to `$themeIds` and `$allThemes` |
| C-3: JSONC theme files fail JSON.parse | CRITICAL | Story 1 | Go-side JSONC stripping |
| C-4: Size check after full file read | CRITICAL | Story 1 | `os.Stat` before `os.ReadFile` |
| C-5: Imported theme lost on restart | CRITICAL | Story 4 | `restoreImportedThemeFromConfig` in App.svelte |
| H-1: ReadThemeFileResolved not wired | HIGH | Story 1 | Include resolution merged into ReadThemeFile |
| H-2: Config race condition | HIGH | Story 1 | Mutex around all config ops |
| H-3: Tilde expansion bug | HIGH | Story 1 | Match `~/` or `~` only, check errors |
| H-5: Token specificity reversed | HIGH | Story 2 | Reverse iteration in convertTokenColors |
| H-6: Alpha in CSS vars | HIGH | Story 2 | Strip alpha from background vars |
| H-7: tmTheme filter missing | HIGH | Story 1 | `.json` suffix filter in scanner |
| M-1: No activation guard | MEDIUM | Story 4 | `activatingPath` guard variable |
| M-5: Theme ID collisions | MEDIUM | Story 4 | extensionId in theme ID |
| L-1: dimColor hex normalization | LOW | Story 2 | `normalizeHex` utility |
| L-3: HC themes default to vs-dark | LOW | Story 2 | `getMonacoBase` with hc-black/hc-light |

## Approach

- **B2 only** (hand-rolled token converter, zero npm dependencies)
- B1 dropped (npm package may not exist)
- Approach C (Shiki/WASM) deferred as a future upgrade path

**Total Stories:** 5
**Ready for Sprint:** Stories 1, 2, 3 (no dependencies, can start in parallel)
**Blocked:** Story 4 (needs 1+2+3), Story 5 (needs 3+4)

**Recommended Sprint Order:**

1. Stories 1, 2, 3 -- start in parallel (independent domains: backend, frontend module, frontend store)
2. Story 4 -- after all three foundations land (integration story)
3. Story 5 -- after Story 4 (polish and robustness)

Story files written to: `/Users/linus/Development/mashed/docs/stories/`
