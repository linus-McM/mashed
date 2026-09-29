---
type: Module
title: mountSvelte.ts
description: "Graphify community 3: docs/stories/bmad-interactive-07-registry-entries.md, frontend/src/components/bmad/MarkdownBlock.svelte, frontend/src/components/bmad/RawViewToggle.svelte, frontend/src/component"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:16:08Z" }
stale_after: "2026-10-13T11:16:08Z"
source_commit: 85e7124b5a77a2f29bd9db818781712ea7a1049b
sources:
  - { id: bmad-interactive-07-registry-entries, resource: docs/stories/bmad-interactive-07-registry-entries.md, last_modified: "2026-04-20T15:36:58+10:00", digest: 47d47db81112ff4e }
  - { id: MarkdownBlock, resource: frontend/src/components/bmad/MarkdownBlock.svelte, last_modified: "2026-04-22T20:18:54+10:00", digest: 3f0e52590dddfec2 }
  - { id: RawViewToggle, resource: frontend/src/components/bmad/RawViewToggle.svelte, last_modified: "2026-04-22T12:56:23+10:00", digest: 21457097a1637e7e }
  - { id: SummaryCard, resource: frontend/src/components/bmad/SummaryCard.svelte, last_modified: "2026-04-22T08:19:28+10:00", digest: cb815de294ae47a0 }
  - { id: AstNode.test, resource: frontend/src/components/bmad/__tests__/AstNode.test.ts, last_modified: "2026-04-22T11:51:34+10:00", digest: 33789a8a8131bd3a }
  - { id: CommandNode.breadcrumb.test, resource: frontend/src/components/bmad/__tests__/CommandNode.breadcrumb.test.ts, last_modified: "2026-04-22T20:01:25+10:00", digest: d5f2905d83e01211 }
  - { id: CommandNode.status.test, resource: frontend/src/components/bmad/__tests__/CommandNode.status.test.ts, last_modified: "2026-04-22T20:01:25+10:00", digest: 3fa251016ae15b6d }
  - { id: ComparisonTable.test, resource: frontend/src/components/bmad/__tests__/ComparisonTable.test.ts, last_modified: "2026-04-22T08:19:28+10:00", digest: a201a2122ed92e1c }
  - { id: DecisionGroup.test, resource: frontend/src/components/bmad/__tests__/DecisionGroup.test.ts, last_modified: "2026-04-22T11:51:34+10:00", digest: 3bdedd5f8c85ea3d }
  - { id: DiagnosticsChip.test, resource: frontend/src/components/bmad/__tests__/DiagnosticsChip.test.ts, last_modified: "2026-04-22T20:01:25+10:00", digest: 5a1b2aad12debba0 }
  - { id: HintBanner.test, resource: frontend/src/components/bmad/__tests__/HintBanner.test.ts, last_modified: "2026-04-22T08:19:28+10:00", digest: 436fcbc3d6cca993 }
  - { id: MarkdownBlock.test, resource: frontend/src/components/bmad/__tests__/MarkdownBlock.test.ts, last_modified: "2026-04-22T08:19:28+10:00", digest: fffa845df450229f }
  - { id: MultiFileLoaderNode.test, resource: frontend/src/components/bmad/__tests__/MultiFileLoaderNode.test.ts, last_modified: "2026-04-22T20:01:25+10:00", digest: 05eb5149cdc50787 }
  - { id: ProcessNode.multiInput.test, resource: frontend/src/components/bmad/__tests__/ProcessNode.multiInput.test.ts, last_modified: "2026-04-22T20:01:25+10:00", digest: b50d732d5ea73c1a }
  - { id: ProcessNode.status.test, resource: frontend/src/components/bmad/__tests__/ProcessNode.status.test.ts, last_modified: "2026-04-22T20:01:25+10:00", digest: cb84331adc8f4a2d }
  - { id: RawViewToggle.test, resource: frontend/src/components/bmad/__tests__/RawViewToggle.test.ts, last_modified: "2026-04-22T12:56:23+10:00", digest: fcf457a4a4ad98ac }
  - { id: SummaryCard.test, resource: frontend/src/components/bmad/__tests__/SummaryCard.test.ts, last_modified: "2026-04-22T08:19:28+10:00", digest: c5e05e48c7be686b }
  - { id: linkSanitiser.test, resource: frontend/src/components/bmad/__tests__/linkSanitiser.test.ts, last_modified: "2026-04-22T08:19:28+10:00", digest: 4e352b3c3d0f7983 }
  - { id: mountSvelte, resource: frontend/src/components/bmad/__tests__/mountSvelte.ts, last_modified: "2026-04-22T20:01:25+10:00", digest: becfbd6c142f9b1b }
  - { id: linkSanitiser, resource: frontend/src/components/bmad/linkSanitiser.ts, last_modified: "2026-04-22T08:19:28+10:00", digest: c487727219545c6a }
  - { id: workflow, resource: frontend/src/types/workflow.ts, last_modified: "2026-04-23T13:05:14+10:00", digest: dde86e2d10f68917 }
  - { id: markdown-it, resource: markdown-it, last_modified: "2026-09-29T11:16:08Z", digest: missing }
---

# Files
- `docs/stories/bmad-interactive-07-registry-entries.md`
- `frontend/src/components/bmad/MarkdownBlock.svelte`
- `frontend/src/components/bmad/RawViewToggle.svelte`
- `frontend/src/components/bmad/SummaryCard.svelte`
- `frontend/src/components/bmad/__tests__/AstNode.test.ts`
- `frontend/src/components/bmad/__tests__/CommandNode.breadcrumb.test.ts`
- `frontend/src/components/bmad/__tests__/CommandNode.status.test.ts`
- `frontend/src/components/bmad/__tests__/ComparisonTable.test.ts`
- `frontend/src/components/bmad/__tests__/DecisionGroup.test.ts`
- `frontend/src/components/bmad/__tests__/DiagnosticsChip.test.ts`
- `frontend/src/components/bmad/__tests__/HintBanner.test.ts`
- `frontend/src/components/bmad/__tests__/MarkdownBlock.test.ts`
- `frontend/src/components/bmad/__tests__/MultiFileLoaderNode.test.ts`
- `frontend/src/components/bmad/__tests__/ProcessNode.multiInput.test.ts`
- `frontend/src/components/bmad/__tests__/ProcessNode.status.test.ts`
- `frontend/src/components/bmad/__tests__/RawViewToggle.test.ts`
- `frontend/src/components/bmad/__tests__/SummaryCard.test.ts`
- `frontend/src/components/bmad/__tests__/linkSanitiser.test.ts`
- `frontend/src/components/bmad/__tests__/mountSvelte.ts`
- `frontend/src/components/bmad/linkSanitiser.ts`
- `frontend/src/types/workflow.ts`
- `markdown-it`

# Symbols
- Exact registry shapes (docs/stories/bmad-interactive-07-registry-entries.md:L40)
- MarkdownBlock.svelte (frontend/src/components/bmad/MarkdownBlock.svelte:L1)
- RawViewToggle.svelte (frontend/src/components/bmad/RawViewToggle.svelte:L1)
- SummaryCard.svelte (frontend/src/components/bmad/SummaryCard.svelte:L1)
- AstNode.test.ts (frontend/src/components/bmad/__tests__/AstNode.test.ts:L1)
- UINode (frontend/src/components/bmad/__tests__/AstNode.test.ts:L14)
- FIXTURES (frontend/src/components/bmad/__tests__/AstNode.test.ts:L26)
- CommandNode.breadcrumb.test.ts (frontend/src/components/bmad/__tests__/CommandNode.breadcrumb.test.ts:L1)
- CommandNodeProps (frontend/src/components/bmad/__tests__/CommandNode.breadcrumb.test.ts:L24)
- BreadcrumbOverrides (frontend/src/components/bmad/__tests__/CommandNode.breadcrumb.test.ts:L30)
- makeNodeProps() (frontend/src/components/bmad/__tests__/CommandNode.breadcrumb.test.ts:L38)
- CommandNode.status.test.ts (frontend/src/components/bmad/__tests__/CommandNode.status.test.ts:L1)
- StatusToken (frontend/src/components/bmad/__tests__/CommandNode.status.test.ts:L18)
- CommandNodeProps (frontend/src/components/bmad/__tests__/CommandNode.status.test.ts:L20)
- StatusOverrides (frontend/src/components/bmad/__tests__/CommandNode.status.test.ts:L26)
- baseProps() (frontend/src/components/bmad/__tests__/CommandNode.status.test.ts:L32)
- ComparisonTable.test.ts (frontend/src/components/bmad/__tests__/ComparisonTable.test.ts:L1)
- ROWS (frontend/src/components/bmad/__tests__/ComparisonTable.test.ts:L10)
- COLUMNS (frontend/src/components/bmad/__tests__/ComparisonTable.test.ts:L9)
- DecisionGroup.test.ts (frontend/src/components/bmad/__tests__/DecisionGroup.test.ts:L1)
- $on() (frontend/src/components/bmad/__tests__/DecisionGroup.test.ts:L112)
- CASES (frontend/src/components/bmad/__tests__/DecisionGroup.test.ts:L20)
- $destroy() (frontend/src/components/bmad/__tests__/DecisionGroup.test.ts:L66)
- DiagnosticsChip.test.ts (frontend/src/components/bmad/__tests__/DiagnosticsChip.test.ts:L1)
- mount (frontend/src/components/bmad/__tests__/DiagnosticsChip.test.ts:L12)
- render() (frontend/src/components/bmad/__tests__/DiagnosticsChip.test.ts:L13)
- findUntrusted() (frontend/src/components/bmad/__tests__/DiagnosticsChip.test.ts:L18)
- findNotesChip() (frontend/src/components/bmad/__tests__/DiagnosticsChip.test.ts:L23)
- HintBanner.test.ts (frontend/src/components/bmad/__tests__/HintBanner.test.ts:L1)
- Tone (frontend/src/components/bmad/__tests__/HintBanner.test.ts:L10)
- TONES (frontend/src/components/bmad/__tests__/HintBanner.test.ts:L9)
- MarkdownBlock.test.ts (frontend/src/components/bmad/__tests__/MarkdownBlock.test.ts:L1)
- MultiFileLoaderNode.test.ts (frontend/src/components/bmad/__tests__/MultiFileLoaderNode.test.ts:L1)
- Entry (frontend/src/components/bmad/__tests__/MultiFileLoaderNode.test.ts:L19)
- MultiFileLoaderProps (frontend/src/components/bmad/__tests__/MultiFileLoaderNode.test.ts:L24)
- MultiFileLoaderOverrides (frontend/src/components/bmad/__tests__/MultiFileLoaderNode.test.ts:L30)
- makeProps() (frontend/src/components/bmad/__tests__/MultiFileLoaderNode.test.ts:L34)
- ProcessNode.multiInput.test.ts (frontend/src/components/bmad/__tests__/ProcessNode.multiInput.test.ts:L1)
- TestProcess (frontend/src/components/bmad/__tests__/ProcessNode.multiInput.test.ts:L22)
- TestNodeData (frontend/src/components/bmad/__tests__/ProcessNode.multiInput.test.ts:L30)
- ProcessNodeProps (frontend/src/components/bmad/__tests__/ProcessNode.multiInput.test.ts:L37)
- MultiInputOverrides (frontend/src/components/bmad/__tests__/ProcessNode.multiInput.test.ts:L43)
- makeProps() (frontend/src/components/bmad/__tests__/ProcessNode.multiInput.test.ts:L50)
- collectBreadcrumbRows() (frontend/src/components/bmad/__tests__/ProcessNode.multiInput.test.ts:L70)
- ProcessNode.status.test.ts (frontend/src/components/bmad/__tests__/ProcessNode.status.test.ts:L1)
- StatusToken (frontend/src/components/bmad/__tests__/ProcessNode.status.test.ts:L18)
- Phase (frontend/src/components/bmad/__tests__/ProcessNode.status.test.ts:L19)
- AgentRole (frontend/src/components/bmad/__tests__/ProcessNode.status.test.ts:L20)
- TestProcess (frontend/src/components/bmad/__tests__/ProcessNode.status.test.ts:L33)
- TestNodeData (frontend/src/components/bmad/__tests__/ProcessNode.status.test.ts:L41)
- ProcessNodeProps (frontend/src/components/bmad/__tests__/ProcessNode.status.test.ts:L51)
- StatusOverrides (frontend/src/components/bmad/__tests__/ProcessNode.status.test.ts:L57)
- baseProps() (frontend/src/components/bmad/__tests__/ProcessNode.status.test.ts:L71)
- RawViewToggle.test.ts (frontend/src/components/bmad/__tests__/RawViewToggle.test.ts:L1)
- Props (frontend/src/components/bmad/__tests__/RawViewToggle.test.ts:L11)
- SummaryCard.test.ts (frontend/src/components/bmad/__tests__/SummaryCard.test.ts:L1)
- linkSanitiser.test.ts (frontend/src/components/bmad/__tests__/linkSanitiser.test.ts:L1)
- mountSvelte.ts (frontend/src/components/bmad/__tests__/mountSvelte.ts:L1)
- WailsAppModule (frontend/src/components/bmad/__tests__/mountSvelte.ts:L124)
- WailsAppOverrides (frontend/src/components/bmad/__tests__/mountSvelte.ts:L130)
- MountedComponent (frontend/src/components/bmad/__tests__/mountSvelte.ts:L21)
- SvelteInit (frontend/src/components/bmad/__tests__/mountSvelte.ts:L22)
- makeMount() (frontend/src/components/bmad/__tests__/mountSvelte.ts:L24)
- mountComponent() (frontend/src/components/bmad/__tests__/mountSvelte.ts:L48)
- xyflowHandleStub() (frontend/src/components/bmad/__tests__/mountSvelte.ts:L73)
- linkSanitiser.ts (frontend/src/components/bmad/linkSanitiser.ts:L1)
- REJECTED_SCHEMES (frontend/src/components/bmad/linkSanitiser.ts:L12)
- REJECTED_PROTOCOL_RE (frontend/src/components/bmad/linkSanitiser.ts:L13)
- sanitizeUrl() (frontend/src/components/bmad/linkSanitiser.ts:L15)
- UNSAFE_URL_IN_TEXT (frontend/src/components/bmad/linkSanitiser.ts:L31)
- scrubUnsafeUrls() (frontend/src/components/bmad/linkSanitiser.ts:L35)
- CanvasNodeData (frontend/src/types/workflow.ts:L49)
- markdown-it (markdown-it:)

# Depends on
- [runtime.js](/modules/runtime-js.md)
- [svelte](/modules/svelte.md)
- [uiAst.ts](/modules/uiast-ts.md)
- [vitest](/modules/vitest.md)

# Inferred
- [Story 1: Native macOS Menu Bar Construction](/modules/story-1-native-macos-menu-bar-construction.md)

# Features
- no feature plan names these files
