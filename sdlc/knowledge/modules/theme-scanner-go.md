---
type: Module
title: theme_scanner.go
description: "Graphify community 99: theme_scanner.go, theme_scanner_test.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T07:07:25Z" }
stale_after: "2026-10-13T07:07:25Z"
source_commit: ""
sources:
  - { id: theme_scanner, resource: theme_scanner.go, last_modified: "2026-09-29T07:07:25Z", digest: d02c5e5c45d34f49 }
  - { id: theme_scanner_test, resource: theme_scanner_test.go, last_modified: "2026-09-29T07:07:25Z", digest: b5300960a22f33a8 }
---

# Files
- `theme_scanner.go`
- `theme_scanner_test.go`

# Symbols
- theme_scanner.go (theme_scanner.go:L1)
- stripJSONC() (theme_scanner.go:L106)
- stripTrailingCommas() (theme_scanner.go:L166)
- isVSIXThemePath() (theme_scanner.go:L19)
- packageJSON (theme_scanner.go:L192)
- parseVSIXThemePath() (theme_scanner.go:L25)
- .ListVSCodiumThemes() (theme_scanner.go:L270)
- App (theme_scanner.go:L270)
- .ReadBundledThemeFile() (theme_scanner.go:L329)
- readAndResolveVSIXTheme() (theme_scanner.go:L346)
- readFileFromZip() (theme_scanner.go:L39)
- rawTheme (theme_scanner.go:L396)
- .ReadThemeFile() (theme_scanner.go:L407)
- .readThemeFileWithDepth() (theme_scanner.go:L426)
- .readThemeFromVSIX() (theme_scanner.go:L501)
- mergeThemes() (theme_scanner.go:L59)
- expandTilde() (theme_scanner.go:L86)
- TestStripJSONC() (theme_scanner_test.go:L104)
- TestTildeExpansion() (theme_scanner_test.go:L1164)
- TestStripJSONC_ResultIsValidJSON() (theme_scanner_test.go:L193)

# Depends on
- [bundled_themes_test.go](/modules/bundled-themes-test-go.md)
- [loadConfig](/modules/loadconfig.md)

# Inferred
- [loadConfig](/modules/loadconfig.md)

# Features
- no feature plan names these files
