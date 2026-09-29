---
type: Module
title: setupTestConfig
description: "Graphify community 24: app_config_test.go, editor_settings_test.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:27Z" }
stale_after: "2026-10-13T11:35:27Z"
source_commit: a2484a9a3b200a7f406028196e6f86cc4cf9306f
sources:
  - { id: app_config_test, resource: app_config_test.go, last_modified: "2026-04-28T12:36:05+10:00", digest: ca9b2bb89bd6ecc6 }
  - { id: editor_settings_test, resource: editor_settings_test.go, last_modified: "2026-04-10T09:11:59+10:00", digest: 874a181ff7c1785e }
---

# Files
- `app_config_test.go`
- `editor_settings_test.go`

# Symbols
- app_config_test.go (app_config_test.go:L1)
- TestU1_AC4_MashedConfig_UIAdapterFields_RoundTrip() (app_config_test.go:L12)
- TestU1_AC4_MashedConfig_LegacyLoad_DefaultsApplied() (app_config_test.go:L42)
- TestU1_AC4_MashedConfig_ExplicitOptOut_Honored() (app_config_test.go:L54)
- TestU1_AC4_MashedConfig_MalformedJSON_ReturnsDefaults() (app_config_test.go:L64)
- TestU1_AC4_MashedConfig_MissingFile_ReturnsDefaults() (app_config_test.go:L75)
- TestU5_LoadConfig_InvalidModelFallsBackToDefault() (app_config_test.go:L87)
- editor_settings_test.go (editor_settings_test.go:L1)
- TestSetEditorSettings_AC3_ValidRoundTrip() (editor_settings_test.go:L140)
- TestSetEditorSettings_AC3_PreservesOtherConfig() (editor_settings_test.go:L162)
- setupTestConfig() (editor_settings_test.go:L18)
- TestSetEditorSettings_AC4_InvalidTabSize() (editor_settings_test.go:L192)
- TestSetEditorSettings_AC4_ValidTabSizeBounds() (editor_settings_test.go:L218)
- TestSetEditorSettings_AC4_InvalidCursorStyle() (editor_settings_test.go:L240)
- TestSetEditorSettings_AC4_ValidCursorStyles() (editor_settings_test.go:L264)
- TestSetEditorSettings_AC4_InvalidWordWrap() (editor_settings_test.go:L279)
- TestSetEditorSettings_AC4_ValidWordWrapValues() (editor_settings_test.go:L303)
- TestSetEditorSettings_AC4_InvalidLineNumbers() (editor_settings_test.go:L318)
- TestSetEditorSettings_AC4_ValidLineNumbers() (editor_settings_test.go:L342)
- TestSetEditorSettings_AC4_InvalidRenderWhitespace() (editor_settings_test.go:L357)
- TestSetEditorSettings_AC4_ValidRenderWhitespace() (editor_settings_test.go:L381)
- TestSetEditorSettings_AC4_InvalidCursorBlinking() (editor_settings_test.go:L396)
- TestSetEditorSettings_AC4_ValidCursorBlinking() (editor_settings_test.go:L420)
- TestSetEditorSettings_AC4_InvalidRenderLineHighlight() (editor_settings_test.go:L435)
- expectedDefaults() (editor_settings_test.go:L44)
- TestSetEditorSettings_AC4_ValidRenderLineHighlight() (editor_settings_test.go:L459)
- TestSetEditorSettings_AC4_ConfigUnchangedOnError() (editor_settings_test.go:L475)
- TestSetEditorSettings_ConcurrentAccess() (editor_settings_test.go:L500)
- TestBackwardCompat_AC5_LegacyConfigNoEditorSettings() (editor_settings_test.go:L527)
- TestBackwardCompat_AC5_ConfigWithEditorSettings() (editor_settings_test.go:L538)
- TestBackwardCompat_AC5_EmptyConfigFile() (editor_settings_test.go:L566)
- TestEditorSettings_AC1_JSONTags() (editor_settings_test.go:L578)
- TestEditorSettings_AC1_JSONRoundTrip() (editor_settings_test.go:L615)
- TestDefaultEditorSettings_AC1_AllFieldsPresent() (editor_settings_test.go:L67)
- TestDefaultEditorSettings_ConsistentValues() (editor_settings_test.go:L72)
- TestGetEditorSettings_AC2_DefaultsWhenNil() (editor_settings_test.go:L83)
- TestGetEditorSettings_AC2_NoConfigFile() (editor_settings_test.go:L91)
- TestGetEditorSettings_ExistingConfig() (editor_settings_test.go:L99)

# Depends on
- [markdown_menu_test.go](/modules/markdown-menu-test-go.md)

# Inferred
- [loadConfig](/modules/loadconfig.md)
- [markdown_menu_test.go](/modules/markdown-menu-test-go.md)

# Features
- no feature plan names these files
