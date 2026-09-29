---
type: Module
title: createMockVSIX
description: "Graphify community 68: bundled_themes_test.go, theme_scanner.go, theme_scanner_test.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T20:55:29Z" }
stale_after: "2026-10-13T20:55:29Z"
source_commit: c5568e08ac3c9ebbe5442e8f9b0ce4b6fe270ffd
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
- TestReadBundledThemeFile_AC3_StripsJSONC() (bundled_themes_test.go:L340)
- TestReadBundledThemeFile_AC3_ResolvesIncludes() (bundled_themes_test.go:L367)
- TestReadBundledThemeFile_AC3_NoConfigRequired() (bundled_themes_test.go:L419)
- TestReadBundledThemeFile_AC3_InvalidPath() (bundled_themes_test.go:L449)
- TestReadBundledThemeFile_AC3_MissingInternalFile() (bundled_themes_test.go:L477)
- TestScanVSIXDirectory_AC6_EmptyDir() (bundled_themes_test.go:L74)
- TestScanVSIXDirectory_AC6_MissingDir() (bundled_themes_test.go:L82)
- TestScanVSIXDirectory_AC6_CorruptVSIX() (bundled_themes_test.go:L89)
- scanVSIXDirectory() (theme_scanner.go:L207)
- makeVSIXThemePath() (theme_scanner.go:L36)
- TestReadThemeFile_VSIX_PathTraversal() (theme_scanner_test.go:L1043)
- useBundledThemesDir() (theme_scanner_test.go:L1379)
- TestReadBundledThemeFile_RejectsOutsideBundledDir() (theme_scanner_test.go:L1387)
- TestReadBundledThemeFile_RejectsSymlinkEscape() (theme_scanner_test.go:L1412)
- createMockVSIX() (theme_scanner_test.go:L81)

# Depends on
- [theme_scanner.go](/modules/theme-scanner-go.md)

# Inferred
- [loadConfig](/modules/loadconfig.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
