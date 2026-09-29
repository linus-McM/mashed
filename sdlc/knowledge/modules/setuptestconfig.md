---
type: Module
title: setupTestConfig
description: "Graphify community 24: app_config_test.go, app_uiadapter.go, app_uiadapter_bindings_test.go, editor_settings_test.go, markdown_menu_test.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T10:00:58Z" }
stale_after: "2026-10-13T10:00:58Z"
source_commit: 54a45892a8f0a2af4b1dccb63269c628ab15b3cd
sources:
  - { id: app_config_test, resource: app_config_test.go, last_modified: "2026-04-28T12:36:05+10:00", digest: ca9b2bb89bd6ecc6 }
  - { id: app_uiadapter, resource: app_uiadapter.go, last_modified: "2026-04-21T21:06:39+10:00", digest: eebf36a959a4ef0e }
  - { id: app_uiadapter_bindings_test, resource: app_uiadapter_bindings_test.go, last_modified: "2026-04-21T21:06:39+10:00", digest: 189277263c74d9ce }
  - { id: editor_settings_test, resource: editor_settings_test.go, last_modified: "2026-04-10T09:11:59+10:00", digest: 874a181ff7c1785e }
  - { id: markdown_menu_test, resource: markdown_menu_test.go, last_modified: "2026-04-23T10:53:25+10:00", digest: 625b5fa30cfeec7d }
---

# Files
- `app_config_test.go`
- `app_uiadapter.go`
- `app_uiadapter_bindings_test.go`
- `editor_settings_test.go`
- `markdown_menu_test.go`

# Symbols
- app_config_test.go (app_config_test.go:L1)
- TestU1_AC4_MashedConfig_LegacyLoad_DefaultsApplied() (app_config_test.go:L42)
- TestU1_AC4_MashedConfig_ExplicitOptOut_Honored() (app_config_test.go:L54)
- TestU1_AC4_MashedConfig_MalformedJSON_ReturnsDefaults() (app_config_test.go:L64)
- TestU1_AC4_MashedConfig_MissingFile_ReturnsDefaults() (app_config_test.go:L75)
- TestU5_LoadConfig_InvalidModelFallsBackToDefault() (app_config_test.go:L87)
- swapOllamaBaseURLForTest() (app_uiadapter.go:L50)
- app_uiadapter_bindings_test.go (app_uiadapter_bindings_test.go:L1)
- TestU5_AC4_App_ProbeOllamaReachable_DeadSocket() (app_uiadapter_bindings_test.go:L153)
- TestU5_AC4_App_ProbeOllamaReachable_LiveServer() (app_uiadapter_bindings_test.go:L166)
- TestU5_AC4_App_ProbeOllamaReachable_Non2xx() (app_uiadapter_bindings_test.go:L180)
- TestU5_AC8_MashedConfig_UntrustedExpanded_DefaultFalse() (app_uiadapter_bindings_test.go:L194)
- TestU5_AC8_MashedConfig_UntrustedExpanded_ExplicitTrue_Honored() (app_uiadapter_bindings_test.go:L202)
- TestU5_AC8_App_SetUIAdapterUntrustedExpanded_Persists() (app_uiadapter_bindings_test.go:L210)
- TestU5_App_SetOllamaEnabled_Persists() (app_uiadapter_bindings_test.go:L226)
- withBindingsBaseURL() (app_uiadapter_bindings_test.go:L23)
- TestU5_App_ListOllamaModels_ReturnsSortedList() (app_uiadapter_bindings_test.go:L241)
- TestU5_App_ListOllamaModels_EmptyOK() (app_uiadapter_bindings_test.go:L258)
- TestU5_App_ListOllamaModels_WrapsErrOllamaUnreachable() (app_uiadapter_bindings_test.go:L272)
- TestU5_ErrInvalidConfig_IsDistinctSentinel() (app_uiadapter_bindings_test.go:L284)
- closedSocketURL() (app_uiadapter_bindings_test.go:L30)
- TestU5_AC1_App_SetUIAdapterEnabled_Persists() (app_uiadapter_bindings_test.go:L41)
- TestU5_AC1_App_SetUIAdapterEnabled_RoundTripBothDirections() (app_uiadapter_bindings_test.go:L56)
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
- readRawConfig() (editor_settings_test.go:L35)
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
- [loadConfig](/modules/loadconfig.md)

# Inferred
- [loadConfig](/modules/loadconfig.md)

# Features
- no feature plan names these files
