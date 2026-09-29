---
type: Module
title: loadConfig
description: "Graphify community 42: app.go, app_config_test.go, app_uiadapter.go, app_uiadapter_bindings_test.go, app_uiadapter_v3.go, theme_scanner_test.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T15:22:13Z" }
stale_after: "2026-10-13T15:22:13Z"
source_commit: 918616f42268500be0e26faa9f92720109b96761
sources:
  - { id: app, resource: app.go, last_modified: "2026-09-30T01:21:31+10:00", digest: 774e87e5f070ece0 }
  - { id: app_config_test, resource: app_config_test.go, last_modified: "2026-09-30T01:14:01+10:00", digest: d2754cf35a1ecb42 }
  - { id: app_uiadapter, resource: app_uiadapter.go, last_modified: "2026-04-21T21:06:39+10:00", digest: eebf36a959a4ef0e }
  - { id: app_uiadapter_bindings_test, resource: app_uiadapter_bindings_test.go, last_modified: "2026-04-21T21:06:39+10:00", digest: 189277263c74d9ce }
  - { id: app_uiadapter_v3, resource: app_uiadapter_v3.go, last_modified: "2026-04-23T11:43:31+10:00", digest: c6ede9f0c3335d2e }
  - { id: theme_scanner_test, resource: theme_scanner_test.go, last_modified: "2026-09-30T00:54:22+10:00", digest: f4deb15bb2bfe33c }
---

# Files
- `app.go`
- `app_config_test.go`
- `app_uiadapter.go`
- `app_uiadapter_bindings_test.go`
- `app_uiadapter_v3.go`
- `theme_scanner_test.go`

# Symbols
- loadConfig() (app.go:L169)
- defaultConfig() (app.go:L204)
- saveConfig() (app.go:L264)
- .SetDevDir() (app.go:L550)
- .SetTheme() (app.go:L588)
- .SetVSCodiumExtPath() (app.go:L597)
- .SetMonoFont() (app.go:L606)
- .SetFontSize() (app.go:L615)
- .SetSidebarWidth() (app.go:L624)
- TestU1_AC4_MashedConfig_UIAdapterFields_RoundTrip() (app_config_test.go:L15)
- .SetOllamaEnabled() (app_uiadapter.go:L76)
- .SetUIAdapterTimeoutMs() (app_uiadapter.go:L97)
- TestU5_AC3_App_SetOllamaModel_ValidatesName() (app_uiadapter_bindings_test.go:L111)
- TestU5_AC2_App_SetUIAdapterTimeoutMs_BoundsCheck() (app_uiadapter_bindings_test.go:L71)
- .ListClaudeModels() (app_uiadapter_v3.go:L100)
- .ListRouterPolicies() (app_uiadapter_v3.go:L107)
- App (app_uiadapter_v3.go:L39)
- .SetBackend() (app_uiadapter_v3.go:L39)
- .SetClaudeModel() (app_uiadapter_v3.go:L52)
- .SetCLIModel() (app_uiadapter_v3.go:L66)
- .SetRouterPolicy() (app_uiadapter_v3.go:L79)
- .ListBackendsAvailable() (app_uiadapter_v3.go:L93)
- theme_scanner_test.go (theme_scanner_test.go:L1)
- TestReadThemeFile_VSIX_FileNotFound() (theme_scanner_test.go:L1103)
- TestReadThemeFile_VSIX_BackwardCompat() (theme_scanner_test.go:L1130)
- TestSetImportedTheme() (theme_scanner_test.go:L1215)
- TestConcurrentConfigWrites() (theme_scanner_test.go:L1235)
- TestReadThemeFile_ExtPathNotConfigured() (theme_scanner_test.go:L1298)
- TestListVSCodiumThemes_InaccessibleDirectory() (theme_scanner_test.go:L1321)
- TestListVSCodiumThemes_ExtensionID() (theme_scanner_test.go:L1340)
- mockTheme (theme_scanner_test.go:L21)
- TestListVSCodiumThemes_HappyPath() (theme_scanner_test.go:L224)
- createMockExtension() (theme_scanner_test.go:L27)
- TestListVSCodiumThemes_FiltersTmTheme() (theme_scanner_test.go:L294)
- TestListVSCodiumThemes_NotConfigured() (theme_scanner_test.go:L339)
- TestListVSCodiumThemes_CorruptPackageJSON() (theme_scanner_test.go:L362)
- TestReadThemeFile_HappyPath() (theme_scanner_test.go:L405)
- TestReadThemeFile_JSONCStripped() (theme_scanner_test.go:L438)
- TestReadThemeFile_PathTraversal() (theme_scanner_test.go:L477)
- TestReadThemeFile_SymlinkOutside() (theme_scanner_test.go:L530)
- TestReadThemeFile_PrefixConfusion() (theme_scanner_test.go:L565)
- TestReadThemeFile_SizeLimit() (theme_scanner_test.go:L595)
- TestReadThemeFile_IncludeResolution() (theme_scanner_test.go:L629)
- TestReadThemeFile_IncludeMultiLevel() (theme_scanner_test.go:L702)
- TestReadThemeFile_IncludeDepthLimit() (theme_scanner_test.go:L747)
- TestReadThemeFile_VSIX_IncludeDepthLimit() (theme_scanner_test.go:L991)

# Depends on
- [App](/modules/app-109.md)
- [createMockVSIX](/modules/createmockvsix.md)
- [.ReadThemeFile](/modules/readthemefile.md)
- [setupTestConfig](/modules/setuptestconfig.md)

# Inferred
- [App](/modules/app-356.md)
- [createMockVSIX](/modules/createmockvsix.md)
- [setupTestConfig](/modules/setuptestconfig.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
