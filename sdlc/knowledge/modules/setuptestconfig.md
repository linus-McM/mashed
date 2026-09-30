---
type: Module
title: setupTestConfig
description: "Graphify community 13: app_uiadapter.go, app_uiadapter_bindings_test.go, editor_settings_test.go, markdown_menu_test.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T20:55:29Z" }
stale_after: "2026-10-13T20:55:29Z"
source_commit: c5568e08ac3c9ebbe5442e8f9b0ce4b6fe270ffd
sources:
  - { id: app_uiadapter, resource: app_uiadapter.go, last_modified: "2026-04-21T21:06:39+10:00", digest: eebf36a959a4ef0e }
  - { id: app_uiadapter_bindings_test, resource: app_uiadapter_bindings_test.go, last_modified: "2026-04-21T21:06:39+10:00", digest: 189277263c74d9ce }
  - { id: editor_settings_test, resource: editor_settings_test.go, last_modified: "2026-04-10T09:11:59+10:00", digest: 874a181ff7c1785e }
  - { id: markdown_menu_test, resource: markdown_menu_test.go, last_modified: "2026-04-23T10:53:25+10:00", digest: 625b5fa30cfeec7d }
---

# Files
- `app_uiadapter.go`
- `app_uiadapter_bindings_test.go`
- `editor_settings_test.go`
- `markdown_menu_test.go`

# Symbols
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
- TestSetGetMarkdownMenuSettings_RoundTrip() (markdown_menu_test.go:L92)

# Depends on
- [loadConfig](/modules/loadconfig.md)

# Inferred
- [app_config_test.go](/modules/app-config-test-go.md)
- [loadConfig](/modules/loadconfig.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
