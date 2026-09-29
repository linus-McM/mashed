---
type: Module
title: loadConfig
description: "Graphify community 16: app.go, app_config_test.go, app_uiadapter_bindings_test.go, theme_scanner_test.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T12:23:14Z" }
stale_after: "2026-10-13T12:23:14Z"
source_commit: ab1f2eb4ec65fc2b1fce15503c11f0e310758404
sources:
  - { id: app, resource: app.go, last_modified: "2026-05-07T18:18:02+10:00", digest: 295875db4f0bdc1e }
  - { id: app_config_test, resource: app_config_test.go, last_modified: "2026-04-28T12:36:05+10:00", digest: ca9b2bb89bd6ecc6 }
  - { id: app_uiadapter_bindings_test, resource: app_uiadapter_bindings_test.go, last_modified: "2026-04-21T21:06:39+10:00", digest: 189277263c74d9ce }
  - { id: theme_scanner_test, resource: theme_scanner_test.go, last_modified: "2026-04-07T10:03:32+10:00", digest: b5300960a22f33a8 }
---

# Files
- `app.go`
- `app_config_test.go`
- `app_uiadapter_bindings_test.go`
- `theme_scanner_test.go`

# Symbols
- configPath() (app.go:L155)
- loadConfig() (app.go:L168)
- defaultConfig() (app.go:L203)
- saveConfig() (app.go:L213)
- .SetDevDir() (app.go:L492)
- .SetTheme() (app.go:L530)
- .SetMonoFont() (app.go:L548)
- .SetFontSize() (app.go:L557)
- .SetSidebarWidth() (app.go:L566)
- TestU1_AC4_MashedConfig_UIAdapterFields_RoundTrip() (app_config_test.go:L12)
- TestU5_AC3_App_SetOllamaModel_ValidatesName() (app_uiadapter_bindings_test.go:L111)
- TestU5_AC2_App_SetUIAdapterTimeoutMs_BoundsCheck() (app_uiadapter_bindings_test.go:L71)
- theme_scanner_test.go (theme_scanner_test.go:L1)
- TestReadThemeFile_VSIX_PathTraversal() (theme_scanner_test.go:L1040)
- TestReadThemeFile_VSIX_SizeLimit() (theme_scanner_test.go:L1070)
- TestReadThemeFile_VSIX_FileNotFound() (theme_scanner_test.go:L1100)
- TestReadThemeFile_VSIX_BackwardCompat() (theme_scanner_test.go:L1127)
- TestSetImportedTheme() (theme_scanner_test.go:L1212)
- TestConcurrentConfigWrites() (theme_scanner_test.go:L1232)
- TestReadThemeFile_ExtPathNotConfigured() (theme_scanner_test.go:L1295)
- TestListVSCodiumThemes_InaccessibleDirectory() (theme_scanner_test.go:L1318)
- TestListVSCodiumThemes_ExtensionID() (theme_scanner_test.go:L1337)
- mockTheme (theme_scanner_test.go:L18)
- TestListVSCodiumThemes_HappyPath() (theme_scanner_test.go:L221)
- createMockExtension() (theme_scanner_test.go:L24)
- TestListVSCodiumThemes_FiltersTmTheme() (theme_scanner_test.go:L291)
- TestListVSCodiumThemes_NotConfigured() (theme_scanner_test.go:L336)
- TestListVSCodiumThemes_CorruptPackageJSON() (theme_scanner_test.go:L359)
- TestReadThemeFile_HappyPath() (theme_scanner_test.go:L402)
- TestReadThemeFile_JSONCStripped() (theme_scanner_test.go:L435)
- TestReadThemeFile_PathTraversal() (theme_scanner_test.go:L474)
- TestReadThemeFile_SymlinkOutside() (theme_scanner_test.go:L527)
- TestReadThemeFile_PrefixConfusion() (theme_scanner_test.go:L562)
- TestReadThemeFile_SizeLimit() (theme_scanner_test.go:L592)
- TestReadThemeFile_IncludeResolution() (theme_scanner_test.go:L626)
- TestReadThemeFile_IncludeMultiLevel() (theme_scanner_test.go:L699)
- TestReadThemeFile_IncludeDepthLimit() (theme_scanner_test.go:L744)
- TestReadThemeFile_VSIX_HappyPath() (theme_scanner_test.go:L797)
- TestReadThemeFile_VSIX_JSONCStripped() (theme_scanner_test.go:L830)
- TestReadThemeFile_VSIX_IncludeResolution() (theme_scanner_test.go:L870)
- TestReadThemeFile_VSIX_IncludeMultiLevel() (theme_scanner_test.go:L942)
- TestReadThemeFile_VSIX_IncludeDepthLimit() (theme_scanner_test.go:L988)

# Depends on
- [App](/modules/app.md)
- [bundled_themes_test.go](/modules/bundled-themes-test-go.md)
- [theme_scanner.go](/modules/theme-scanner-go.md)

# Inferred
- [App](/modules/app-296.md)
- [bundled_themes_test.go](/modules/bundled-themes-test-go.md)
- [setupTestConfig](/modules/setuptestconfig.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
