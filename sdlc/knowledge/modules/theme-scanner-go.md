---
type: Module
title: theme_scanner.go
description: "Graphify community 46: bundled_themes_test.go, theme_scanner.go, theme_scanner_test.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T12:23:00Z" }
stale_after: "2026-10-13T12:23:00Z"
source_commit: c111e9108518da0d6e68dc3fa508835faaebfd96
sources:
  - { id: bundled_themes_test, resource: bundled_themes_test.go, last_modified: "2026-04-10T09:28:28+10:00", digest: 7521e18fece03e1e }
  - { id: theme_scanner, resource: theme_scanner.go, last_modified: "2026-04-10T09:28:28+10:00", digest: d02c5e5c45d34f49 }
  - { id: theme_scanner_test, resource: theme_scanner_test.go, last_modified: "2026-04-07T10:03:32+10:00", digest: b5300960a22f33a8 }
---

# Files
- `bundled_themes_test.go`
- `theme_scanner.go`
- `theme_scanner_test.go`

# Symbols
- TestBundledThemesDir_ReturnsPath() (bundled_themes_test.go:L490)
- theme_scanner.go (theme_scanner.go:L1)
- stripJSONC() (theme_scanner.go:L106)
- stripTrailingCommas() (theme_scanner.go:L166)
- isVSIXThemePath() (theme_scanner.go:L19)
- packageJSON (theme_scanner.go:L192)
- parseVSIXThemePath() (theme_scanner.go:L25)
- .ListVSCodiumThemes() (theme_scanner.go:L270)
- App (theme_scanner.go:L270)
- bundledThemesDir() (theme_scanner.go:L295)
- .ListBundledThemes() (theme_scanner.go:L318)
- .ReadBundledThemeFile() (theme_scanner.go:L329)
- readAndResolveVSIXTheme() (theme_scanner.go:L346)
- readFileFromZip() (theme_scanner.go:L39)
- rawTheme (theme_scanner.go:L396)
- .ReadThemeFile() (theme_scanner.go:L407)
- .readThemeFileWithDepth() (theme_scanner.go:L426)
- .readThemeFromVSIX() (theme_scanner.go:L501)
- .SetImportedTheme() (theme_scanner.go:L519)
- mergeThemes() (theme_scanner.go:L59)
- expandTilde() (theme_scanner.go:L86)
- TestStripJSONC() (theme_scanner_test.go:L104)
- TestTildeExpansion() (theme_scanner_test.go:L1164)
- TestStripJSONC_ResultIsValidJSON() (theme_scanner_test.go:L193)

# Depends on
- [bundled_themes_test.go](/modules/bundled-themes-test-go.md)

# Inferred
- [loadConfig](/modules/loadconfig.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
