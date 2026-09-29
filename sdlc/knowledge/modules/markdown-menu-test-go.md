---
type: Module
title: markdown_menu_test.go
description: "Graphify community 112: app.go, app_uiadapter_bindings_test.go, editor_settings_test.go, markdown_menu_test.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:27Z" }
stale_after: "2026-10-13T11:35:27Z"
source_commit: a2484a9a3b200a7f406028196e6f86cc4cf9306f
sources:
  - { id: app, resource: app.go, last_modified: "2026-05-07T18:18:02+10:00", digest: 295875db4f0bdc1e }
  - { id: app_uiadapter_bindings_test, resource: app_uiadapter_bindings_test.go, last_modified: "2026-04-21T21:06:39+10:00", digest: 189277263c74d9ce }
  - { id: editor_settings_test, resource: editor_settings_test.go, last_modified: "2026-04-10T09:11:59+10:00", digest: 874a181ff7c1785e }
  - { id: markdown_menu_test, resource: markdown_menu_test.go, last_modified: "2026-04-23T10:53:25+10:00", digest: 625b5fa30cfeec7d }
---

# Files
- `app.go`
- `app_uiadapter_bindings_test.go`
- `editor_settings_test.go`
- `markdown_menu_test.go`

# Symbols
- configPath() (app.go:L155)
- TestU5_AC8_App_SetUIAdapterUntrustedExpanded_Persists() (app_uiadapter_bindings_test.go:L210)
- TestU5_App_SetOllamaEnabled_Persists() (app_uiadapter_bindings_test.go:L226)
- TestU5_AC1_App_SetUIAdapterEnabled_Persists() (app_uiadapter_bindings_test.go:L41)
- readRawConfig() (editor_settings_test.go:L35)
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
- TestSetGetMarkdownMenuSettings_RoundTrip() (markdown_menu_test.go:L92)

# Depends on
- no EXTRACTED edges to other modules

# Inferred
- [loadConfig](/modules/loadconfig.md)
- [setupTestConfig](/modules/setuptestconfig.md)

# Features
- no feature plan names these files
