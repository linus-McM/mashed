---
type: Module
title: markdown_menu_test.go
description: "Graphify community 181: app.go, app_config_test.go, markdown_menu_test.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T07:07:25Z" }
stale_after: "2026-10-13T07:07:25Z"
source_commit: ""
sources:
  - { id: app, resource: app.go, last_modified: "2026-09-29T07:07:25Z", digest: 295875db4f0bdc1e }
  - { id: app_config_test, resource: app_config_test.go, last_modified: "2026-09-29T07:07:25Z", digest: ca9b2bb89bd6ecc6 }
  - { id: markdown_menu_test, resource: markdown_menu_test.go, last_modified: "2026-09-29T07:07:25Z", digest: 625b5fa30cfeec7d }
---

# Files
- `app.go`
- `app_config_test.go`
- `markdown_menu_test.go`

# Symbols
- configPath() (app.go:L155)
- TestU1_AC4_MashedConfig_UIAdapterFields_RoundTrip() (app_config_test.go:L12)
- markdown_menu_test.go (markdown_menu_test.go:L1)
- TestConfigMarkdownMenuOmitempty() (markdown_menu_test.go:L121)
- TestConfigMarkdownMenu_PresentWhenSet() (markdown_menu_test.go:L138)
- TestSetMarkdownMenuSettings_PreservesOtherConfig() (markdown_menu_test.go:L160)
- expectedMarkdownMenuDefaults() (markdown_menu_test.go:L17)
- TestBackwardCompat_LegacyConfigNoMarkdownMenu() (markdown_menu_test.go:L195)
- TestSetMarkdownMenuSettings_ConcurrentRace() (markdown_menu_test.go:L206)
- TestMarkdownMenuSettings_JSONTags() (markdown_menu_test.go:L251)
- TestMarkdownMenuSettings_JSONRoundTrip() (markdown_menu_test.go:L274)
- TestMarkdownMenu_ConfigPathLocation() (markdown_menu_test.go:L292)
- TestDefaultMarkdownMenuSettings() (markdown_menu_test.go:L32)
- TestDefaultMarkdownMenuSettings_Consistent() (markdown_menu_test.go:L47)
- TestGetMarkdownMenuSettings_DefaultsWhenNil() (markdown_menu_test.go:L57)
- TestGetMarkdownMenuSettings_NoConfigFile() (markdown_menu_test.go:L75)

# Depends on
- [setupTestConfig](/modules/setuptestconfig.md)

# Inferred
- [loadConfig](/modules/loadconfig.md)
- [setupTestConfig](/modules/setuptestconfig.md)

# Features
- no feature plan names these files
