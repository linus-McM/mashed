---
type: Module
title: bundled_themes_test.go
description: "Graphify community 45: bundled_themes_test.go, theme_scanner.go, theme_scanner_test.go"
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
- bundled_themes_test.go (bundled_themes_test.go:L1)
- TestScanVSIXDirectory_AC6_FiltersTmTheme() (bundled_themes_test.go:L116)
- TestScanVSIXDirectory_AC6_SkipsNonVSIXFiles() (bundled_themes_test.go:L139)
- chdirTemp() (bundled_themes_test.go:L16)
- TestScanVSIXDirectory_AC6_MatchesListVSCodiumThemes() (bundled_themes_test.go:L164)
- TestListBundledThemes_AC1_WithThemes() (bundled_themes_test.go:L206)
- TestListBundledThemes_AC1_SortedByLabel() (bundled_themes_test.go:L247)
- TestListBundledThemes_AC2_MissingDir() (bundled_themes_test.go:L275)
- TestScanVSIXDirectory_AC6_WithThemes() (bundled_themes_test.go:L28)
- TestListBundledThemes_AC2_EmptyDir() (bundled_themes_test.go:L287)
- TestReadBundledThemeFile_AC3_ReadsThemeJSON() (bundled_themes_test.go:L305)
- TestReadBundledThemeFile_AC3_StripsJSONC() (bundled_themes_test.go:L339)
- TestReadBundledThemeFile_AC3_ResolvesIncludes() (bundled_themes_test.go:L365)
- TestReadBundledThemeFile_AC3_NoConfigRequired() (bundled_themes_test.go:L416)
- TestReadBundledThemeFile_AC3_InvalidPath() (bundled_themes_test.go:L445)
- TestReadBundledThemeFile_AC3_MissingInternalFile() (bundled_themes_test.go:L473)
- TestScanVSIXDirectory_AC6_EmptyDir() (bundled_themes_test.go:L74)
- TestScanVSIXDirectory_AC6_MissingDir() (bundled_themes_test.go:L82)
- TestScanVSIXDirectory_AC6_CorruptVSIX() (bundled_themes_test.go:L89)
- scanVSIXDirectory() (theme_scanner.go:L205)
- makeVSIXThemePath() (theme_scanner.go:L34)
- createMockVSIX() (theme_scanner_test.go:L78)

# Depends on
- [theme_scanner.go](/modules/theme-scanner-go.md)

# Inferred
- [loadConfig](/modules/loadconfig.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
