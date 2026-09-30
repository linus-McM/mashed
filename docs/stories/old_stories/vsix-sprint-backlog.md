# Sprint Backlog: Read themes directly from .vsix files

## Sprint Backlog

| # | Story | Priority | Domain | Size | Depends On |
|---|-------|----------|--------|------|------------|
| vsix-01 | Read themes directly from .vsix zip archives | P0 | backend | M | none |
| vsix-02 | Update frontend path handling for VSIX-encoded theme paths | P0 | frontend | S | vsix-01 |

**Total Stories:** 2
**Ready for Sprint:** vsix-01 (vsix-02 is ready once vsix-01 is complete)
**Recommended Sprint Order:** vsix-01, vsix-02

## Summary

**vsix-01 (backend, M):** Modifies `theme_scanner.go` to scan `.vsix` zip archives instead of extracted directories. Adds VSIX path helpers (`isVSIXThemePath`, `parseVSIXThemePath`, `makeVSIXThemePath`, `readFileFromZip`), extracts `mergeThemes` from inline code, implements `readThemeFromVSIX` with full security checks and include resolution within zip archives, and rewrites `ListVSCodiumThemes` to iterate `.vsix` files. Backward compatibility preserved -- `ReadThemeFile` dispatches based on `::vsix::` separator presence. 5 tasks, 8 acceptance criteria.

**vsix-02 (frontend, S):** Updates `extractExtensionId` and `makeThemeId` in `themeInit.js` to handle `::vsix::` separator, exports `makeThemeId`, and updates `Settings.svelte` to use the shared function instead of inline duplication. 3 tasks, 5 acceptance criteria.

Story files written to: `/Users/dev/Development/mashed/docs/stories/`
