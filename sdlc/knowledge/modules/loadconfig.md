---
type: Module
title: loadConfig
description: "Graphify community 16: app.go, app_config_test.go, app_uiadapter.go, app_uiadapter_bindings_test.go, app_uiadapter_v3.go, bundled_themes_test.go, theme_scanner.go, theme_scanner_test.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:16:15Z" }
stale_after: "2026-10-13T11:16:15Z"
source_commit: 412dea92db2fe0033e1ef1b18ae99e119ea06b8c
sources:
  - { id: app, resource: app.go, last_modified: "2026-05-07T18:18:02+10:00", digest: 295875db4f0bdc1e }
  - { id: app_config_test, resource: app_config_test.go, last_modified: "2026-04-28T12:36:05+10:00", digest: ca9b2bb89bd6ecc6 }
  - { id: app_uiadapter, resource: app_uiadapter.go, last_modified: "2026-04-21T21:06:39+10:00", digest: eebf36a959a4ef0e }
  - { id: app_uiadapter_bindings_test, resource: app_uiadapter_bindings_test.go, last_modified: "2026-04-21T21:06:39+10:00", digest: 189277263c74d9ce }
  - { id: app_uiadapter_v3, resource: app_uiadapter_v3.go, last_modified: "2026-04-23T11:43:31+10:00", digest: c6ede9f0c3335d2e }
  - { id: bundled_themes_test, resource: bundled_themes_test.go, last_modified: "2026-04-10T09:28:28+10:00", digest: 7521e18fece03e1e }
  - { id: theme_scanner, resource: theme_scanner.go, last_modified: "2026-04-10T09:28:28+10:00", digest: d02c5e5c45d34f49 }
  - { id: theme_scanner_test, resource: theme_scanner_test.go, last_modified: "2026-04-07T10:03:32+10:00", digest: b5300960a22f33a8 }
---

# Files
- `app.go`
- `app_config_test.go`
- `app_uiadapter.go`
- `app_uiadapter_bindings_test.go`
- `app_uiadapter_v3.go`
- `bundled_themes_test.go`
- `theme_scanner.go`
- `theme_scanner_test.go`

# Symbols
- mashedConfig (app.go:L122)
- configPath() (app.go:L155)
- loadConfig() (app.go:L168)
- defaultConfig() (app.go:L203)
- saveConfig() (app.go:L213)
- .SetDevDir() (app.go:L492)
- .GetConfig() (app.go:L525)
- .SetTheme() (app.go:L530)
- .SetVSCodiumExtPath() (app.go:L539)
- .SetMonoFont() (app.go:L548)
- .SetFontSize() (app.go:L557)
- .SetSidebarWidth() (app.go:L566)
- .SetMarkdownMenuSettings() (app.go:L678)
- TestU1_AC4_MashedConfig_UIAdapterFields_RoundTrip() (app_config_test.go:L12)
- .SetOllamaModel() (app_uiadapter.go:L111)
- .ProbeOllamaReachable() (app_uiadapter.go:L124)
- .ListOllamaModels() (app_uiadapter.go:L154)
- loadOllamaBaseURL() (app_uiadapter.go:L41)
- validOllamaModelName() (app_uiadapter.go:L62)
- App (app_uiadapter.go:L67)
- .SetUIAdapterEnabled() (app_uiadapter.go:L67)
- .SetOllamaEnabled() (app_uiadapter.go:L76)
- .SetUIAdapterUntrustedExpanded() (app_uiadapter.go:L86)
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
- TestBundledThemesDir_ReturnsPath() (bundled_themes_test.go:L490)
- TestScanVSIXDirectory_AC6_EmptyDir() (bundled_themes_test.go:L74)
- TestScanVSIXDirectory_AC6_MissingDir() (bundled_themes_test.go:L82)
- TestScanVSIXDirectory_AC6_CorruptVSIX() (bundled_themes_test.go:L89)
- stripJSONC() (theme_scanner.go:L106)
- stripTrailingCommas() (theme_scanner.go:L166)
- isVSIXThemePath() (theme_scanner.go:L19)
- scanVSIXDirectory() (theme_scanner.go:L205)
- parseVSIXThemePath() (theme_scanner.go:L25)
- .ListVSCodiumThemes() (theme_scanner.go:L270)
- App (theme_scanner.go:L270)
- bundledThemesDir() (theme_scanner.go:L295)
- .ListBundledThemes() (theme_scanner.go:L318)
- .ReadBundledThemeFile() (theme_scanner.go:L329)
- makeVSIXThemePath() (theme_scanner.go:L34)
- readAndResolveVSIXTheme() (theme_scanner.go:L346)
- readFileFromZip() (theme_scanner.go:L39)
- rawTheme (theme_scanner.go:L396)
- .ReadThemeFile() (theme_scanner.go:L407)
- .readThemeFileWithDepth() (theme_scanner.go:L426)
- .readThemeFromVSIX() (theme_scanner.go:L501)
- .SetImportedTheme() (theme_scanner.go:L519)
- mergeThemes() (theme_scanner.go:L59)
- expandTilde() (theme_scanner.go:L86)
- theme_scanner_test.go (theme_scanner_test.go:L1)
- TestStripJSONC() (theme_scanner_test.go:L104)
- TestReadThemeFile_VSIX_PathTraversal() (theme_scanner_test.go:L1040)
- TestReadThemeFile_VSIX_SizeLimit() (theme_scanner_test.go:L1070)
- TestReadThemeFile_VSIX_FileNotFound() (theme_scanner_test.go:L1100)
- TestReadThemeFile_VSIX_BackwardCompat() (theme_scanner_test.go:L1127)
- TestTildeExpansion() (theme_scanner_test.go:L1164)
- TestSetImportedTheme() (theme_scanner_test.go:L1212)
- TestConcurrentConfigWrites() (theme_scanner_test.go:L1232)
- TestReadThemeFile_ExtPathNotConfigured() (theme_scanner_test.go:L1295)
- TestListVSCodiumThemes_InaccessibleDirectory() (theme_scanner_test.go:L1318)
- TestListVSCodiumThemes_ExtensionID() (theme_scanner_test.go:L1337)
- mockTheme (theme_scanner_test.go:L18)
- TestStripJSONC_ResultIsValidJSON() (theme_scanner_test.go:L193)
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
- createMockVSIX() (theme_scanner_test.go:L78)
- TestReadThemeFile_VSIX_HappyPath() (theme_scanner_test.go:L797)
- TestReadThemeFile_VSIX_JSONCStripped() (theme_scanner_test.go:L830)
- TestReadThemeFile_VSIX_IncludeResolution() (theme_scanner_test.go:L870)
- TestReadThemeFile_VSIX_IncludeMultiLevel() (theme_scanner_test.go:L942)
- TestReadThemeFile_VSIX_IncludeDepthLimit() (theme_scanner_test.go:L988)

# Depends on
- [App](/modules/app.md)

# Inferred
- [setupTestConfig](/modules/setuptestconfig.md)

# Features
- no feature plan names these files
