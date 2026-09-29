---
type: Module
title: app_config_test.go
description: "Graphify community 99: app.go, app_config_test.go, markdown_menu_test.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T20:55:29Z" }
stale_after: "2026-10-13T20:55:29Z"
source_commit: c5568e08ac3c9ebbe5442e8f9b0ce4b6fe270ffd
sources:
  - { id: app, resource: app.go, last_modified: "2026-09-30T06:52:18+10:00", digest: 19038d72636ae53b }
  - { id: app_config_test, resource: app_config_test.go, last_modified: "2026-09-30T01:14:01+10:00", digest: d2754cf35a1ecb42 }
  - { id: markdown_menu_test, resource: markdown_menu_test.go, last_modified: "2026-04-23T10:53:25+10:00", digest: 625b5fa30cfeec7d }
---

# Files
- `app.go`
- `app_config_test.go`
- `markdown_menu_test.go`

# Symbols
- configPath() (app.go:L172)
- themesPath() (app.go:L178)
- ensurePrivateDir() (app.go:L230)
- writePrivateFile() (app.go:L238)
- quarantineMalformedConfig() (app.go:L264)
- .GetSavedThemes() (app.go:L807)
- .SaveTheme() (app.go:L816)
- .RemoveTheme() (app.go:L836)
- app_config_test.go (app_config_test.go:L1)
- TestConfig_FilesAre0600In0700Dir() (app_config_test.go:L113)
- TestConfig_ConcurrentSetters_NoLostUpdate() (app_config_test.go:L131)
- TestConfig_MalformedIsQuarantinedNotOverwritten() (app_config_test.go:L149)
- TestU1_AC4_MashedConfig_UIAdapterFields_RoundTrip() (app_config_test.go:L15)
- TestConfig_ValidConfigNotQuarantined() (app_config_test.go:L187)
- TestConfig_ReadsNeverSeeTornWrite() (app_config_test.go:L195)
- TestU1_AC4_MashedConfig_LegacyLoad_DefaultsApplied() (app_config_test.go:L45)
- TestU1_AC4_MashedConfig_ExplicitOptOut_Honored() (app_config_test.go:L57)
- TestU1_AC4_MashedConfig_MalformedJSON_ReturnsDefaults() (app_config_test.go:L67)
- TestU1_AC4_MashedConfig_MissingFile_ReturnsDefaults() (app_config_test.go:L78)
- TestU5_LoadConfig_InvalidModelFallsBackToDefault() (app_config_test.go:L90)
- TestMarkdownMenu_ConfigPathLocation() (markdown_menu_test.go:L292)

# Depends on
- [WriteFileAtomic](/modules/writefileatomic.md)

# Inferred
- [loadConfig](/modules/loadconfig.md)
- [setupTestConfig](/modules/setuptestconfig.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
