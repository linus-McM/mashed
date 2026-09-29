---
type: Module
title: loadConfig
description: "Graphify community 16: app.go, app_uiadapter.go, app_uiadapter_v3.go, theme_scanner.go, theme_scanner_test.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:27Z" }
stale_after: "2026-10-13T11:35:27Z"
source_commit: a2484a9a3b200a7f406028196e6f86cc4cf9306f
sources:
  - { id: app, resource: app.go, last_modified: "2026-05-07T18:18:02+10:00", digest: 295875db4f0bdc1e }
  - { id: app_uiadapter, resource: app_uiadapter.go, last_modified: "2026-04-21T21:06:39+10:00", digest: eebf36a959a4ef0e }
  - { id: app_uiadapter_v3, resource: app_uiadapter_v3.go, last_modified: "2026-04-23T11:43:31+10:00", digest: c6ede9f0c3335d2e }
  - { id: theme_scanner, resource: theme_scanner.go, last_modified: "2026-04-10T09:28:28+10:00", digest: d02c5e5c45d34f49 }
  - { id: theme_scanner_test, resource: theme_scanner_test.go, last_modified: "2026-04-07T10:03:32+10:00", digest: b5300960a22f33a8 }
---

# Files
- `app.go`
- `app_uiadapter.go`
- `app_uiadapter_v3.go`
- `theme_scanner.go`
- `theme_scanner_test.go`

# Symbols
- loadConfig() (app.go:L168)
- saveConfig() (app.go:L213)
- .SetDevDir() (app.go:L492)
- .SetTheme() (app.go:L530)
- .SetVSCodiumExtPath() (app.go:L539)
- .SetMonoFont() (app.go:L548)
- .SetFontSize() (app.go:L557)
- .SetSidebarWidth() (app.go:L566)
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
- .ListClaudeModels() (app_uiadapter_v3.go:L100)
- .ListRouterPolicies() (app_uiadapter_v3.go:L107)
- App (app_uiadapter_v3.go:L39)
- .SetBackend() (app_uiadapter_v3.go:L39)
- .SetClaudeModel() (app_uiadapter_v3.go:L52)
- .SetCLIModel() (app_uiadapter_v3.go:L66)
- .SetRouterPolicy() (app_uiadapter_v3.go:L79)
- .ListBackendsAvailable() (app_uiadapter_v3.go:L93)
- .SetImportedTheme() (theme_scanner.go:L519)
- theme_scanner_test.go (theme_scanner_test.go:L1)
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

# Depends on
- [App](/modules/app.md)
- [app_uiadapter_bindings_test.go](/modules/app-uiadapter-bindings-test-go.md)
- [bundled_themes_test.go](/modules/bundled-themes-test-go.md)
- [markdown_menu_test.go](/modules/markdown-menu-test-go.md)
- [.ReadThemeFile](/modules/readthemefile.md)

# Inferred
- [bundled_themes_test.go](/modules/bundled-themes-test-go.md)
- [markdown_menu_test.go](/modules/markdown-menu-test-go.md)

# Features
- no feature plan names these files
