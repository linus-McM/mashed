---
type: Module
title: .ReadThemeFile
description: "Graphify community 101: bundled_themes_test.go, theme_scanner.go, theme_scanner_test.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T15:22:13Z" }
stale_after: "2026-10-13T15:22:13Z"
source_commit: 918616f42268500be0e26faa9f92720109b96761
sources:
  - { id: bundled_themes_test, resource: bundled_themes_test.go, last_modified: "2026-09-30T00:54:22+10:00", digest: efea78a949fc9cff }
  - { id: theme_scanner, resource: theme_scanner.go, last_modified: "2026-09-30T00:54:22+10:00", digest: ade5ed9563586e37 }
  - { id: theme_scanner_test, resource: theme_scanner_test.go, last_modified: "2026-09-30T00:54:22+10:00", digest: f4deb15bb2bfe33c }
---

# Files
- `bundled_themes_test.go`
- `theme_scanner.go`
- `theme_scanner_test.go`

# Symbols
- TestBundledThemesDir_ReturnsPath() (bundled_themes_test.go:L494)
- stripJSONC() (theme_scanner.go:L108)
- stripTrailingCommas() (theme_scanner.go:L168)
- isVSIXThemePath() (theme_scanner.go:L21)
- parseVSIXThemePath() (theme_scanner.go:L27)
- .ListVSCodiumThemes() (theme_scanner.go:L272)
- App (theme_scanner.go:L272)
- bundledThemesDir() (theme_scanner.go:L297)
- .ListBundledThemes() (theme_scanner.go:L320)
- .ReadBundledThemeFile() (theme_scanner.go:L331)
- readAndResolveVSIXTheme() (theme_scanner.go:L362)
- readFileFromZip() (theme_scanner.go:L41)
- rawTheme (theme_scanner.go:L412)
- .ReadThemeFile() (theme_scanner.go:L423)
- .readThemeFileWithDepth() (theme_scanner.go:L442)
- .readThemeFromVSIX() (theme_scanner.go:L517)
- .SetImportedTheme() (theme_scanner.go:L535)
- mergeThemes() (theme_scanner.go:L61)
- expandTilde() (theme_scanner.go:L88)
- TestStripJSONC() (theme_scanner_test.go:L107)
- TestTildeExpansion() (theme_scanner_test.go:L1167)
- TestStripJSONC_ResultIsValidJSON() (theme_scanner_test.go:L196)

# Depends on
- [App](/modules/app-73.md)
- [createMockVSIX](/modules/createmockvsix.md)

# Inferred
- [loadConfig](/modules/loadconfig.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
