// Friendly-name re-exports for Wails-generated types.
//
// DO NOT hand-edit structural definitions here — this module only re-exports
// from ../../../wailsjs/go/models. Wails regenerates models.ts on `wails dev`,
// and because every export below is a type alias to a namespaced symbol, the
// regenerated shape flows through transparently.
//
// Add new re-exports when new top-level namespace members appear upstream.
// Remove re-exports only when the upstream symbol is removed.
//
// Usage:
//   import type { Workflow, ProcessDef } from '$lib/types/wails';
//
// (Requires the $lib alias; otherwise use the relative path
//  ../../lib/types/wails from src/components/**.)

import type { advice, bmad, domain, main } from '../../../wailsjs/go/models';

// ---------------------------------------------------------------------------
// advice/*
// ---------------------------------------------------------------------------
export type AdviceMode = advice.AdviceMode;

// ---------------------------------------------------------------------------
// bmad/*
// ---------------------------------------------------------------------------
export type AgentInfo = bmad.AgentInfo;
export type BmadAgentConfig = bmad.BmadAgentConfig;
export type ControlFlowNodeDef = bmad.ControlFlowNodeDef;
export type GroupedAgents = bmad.GroupedAgents;
export type ValidationIssue = bmad.ValidationIssue;
export type MashedAssetInfo = bmad.MashedAssetInfo;
export type GroupedMashedAssets = bmad.GroupedMashedAssets;
export type InputSpec = bmad.InputSpec;
export type InteractiveTurn = bmad.InteractiveTurn;
export type IterationGate = bmad.IterationGate;
export type ModuleDef = bmad.ModuleDef;
export type NodeInputEntry = bmad.NodeInputEntry;
export type OutputSpec = bmad.OutputSpec;
export type PendingPrompt = bmad.PendingPrompt;
export type Position = bmad.Position;
export type ProcessDef = bmad.ProcessDef;
export type SprintStory = bmad.SprintStory;
export type SprintEpic = bmad.SprintEpic;
export type SprintStatus = bmad.SprintStatus;
export type WorkflowEdge = bmad.WorkflowEdge;
export type WorkflowNode = bmad.WorkflowNode;
export type WorkflowDef = bmad.WorkflowDef;
export type WorkflowExecution = bmad.WorkflowExecution;

// ---------------------------------------------------------------------------
// domain/*
// ---------------------------------------------------------------------------
export type DiffFileStat = domain.DiffFileStat;
export type LogLine = domain.LogLine;
export type ModelInfo = domain.ModelInfo;
export type NotificationEvent = domain.NotificationEvent;
export type ScopedDiff = domain.ScopedDiff;
export type TerminalSession = domain.TerminalSession;
export type WorktreeInfo = domain.WorktreeInfo;

// ---------------------------------------------------------------------------
// main/*
// ---------------------------------------------------------------------------
export type BranchInfo = main.BranchInfo;
export type EditorSettings = main.EditorSettings;
export type LocalFontFile = main.LocalFontFile;
export type LocalFontFamily = main.LocalFontFamily;
export type NerdFontEntry = main.NerdFontEntry;
export type RepoChoice = main.RepoChoice;
export type RepoStatusInfo = main.RepoStatusInfo;
export type VSCodeThemeEntry = main.VSCodeThemeEntry;
// Upstream name is lowercased; re-export under the Pascal-cased friendly name.
export type MashedConfig = main.mashedConfig;

// ---------------------------------------------------------------------------
// Friendly short aliases for the most common frontend touchpoints.
// These are aliases, NOT new types: they resolve to the canonical class.
// ---------------------------------------------------------------------------
export type Agent = bmad.AgentInfo;
export type Workflow = bmad.WorkflowDef;
export type Session = domain.TerminalSession;
export type Notification = domain.NotificationEvent;
