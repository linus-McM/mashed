// Store for BMAD interactive input state (S6).
//
// Consumes events from the S3/S4 backend:
//   bmad:node:awaiting_input  — new PendingPrompt queued
//   bmad:node:input_resolved  — {execId, nodeId, inputId, round, valueHash}
//   bmad:node:input_invalid   — {execId, nodeId, inputId, reason}
//   bmad:node:round_complete  — {execId, nodeId, round, outputKey}
//   bmad:node:gate_satisfied  — {execId, nodeId, round, reason}
//   bmad:node:round_limit     — {execId, nodeId, round}
//   bmad:node:aborted         — {execId, nodeId, reason}
//
// valueHash is NEVER surfaced to consumers — it's ignored on resolve (AC-4).

import { writable, derived, get } from 'svelte/store';
import type { UIAST } from '../types/uiAst';

export interface PendingPrompt {
  execId?: string;
  nodeId: string;
  inputId: string;
  prompt: string;
  shape: string;
  options?: string[];
  round: number;
  createdAt: number;
  promptId: string;
  required?: boolean;
  helpText?: string;
  repoName?: string;
  repoPath?: string;
  maxLength?: number;
  maxRounds?: number;
  lastOutput?: string;
  structured?: string;
}

export type ToastKind = 'info' | 'success' | 'warn' | 'error';

export interface Toast {
  kind: ToastKind;
  text: string;
  id: number;
}

export interface InteractiveInputState {
  pendingPrompts: PendingPrompt[];
  rounds: Record<string, number>;
  activeModal: PendingPrompt | null;
  lastToast: Toast | null;
  validationError: Record<string, string>;
}

const initial: InteractiveInputState = {
  pendingPrompts: [],
  rounds: {},
  activeModal: null,
  lastToast: null,
  validationError: {},
};

export const interactiveInput = writable<InteractiveInputState>(initial);

export function validationKey(nodeId: string, inputId: string): string {
  return `${nodeId}::${inputId}`;
}

export function upsertPrompt(prompt: PendingPrompt): void {
  interactiveInput.update((s) => {
    const filtered = s.pendingPrompts.filter(
      (p) => !(p.nodeId === prompt.nodeId && p.inputId === prompt.inputId),
    );
    return { ...s, pendingPrompts: [...filtered, prompt] };
  });
}

export function resolveInput(nodeId: string, inputId: string): void {
  interactiveInput.update((s) => {
    const pendingPrompts = s.pendingPrompts.filter(
      (p) => !(p.nodeId === nodeId && p.inputId === inputId),
    );
    const activeModal =
      s.activeModal && s.activeModal.nodeId === nodeId && s.activeModal.inputId === inputId
        ? null
        : s.activeModal;
    const key = validationKey(nodeId, inputId);
    const { [key]: _removed, ...rest } = s.validationError;
    return { ...s, pendingPrompts, activeModal, validationError: rest };
  });
}

export function setValidationError(nodeId: string, inputId: string, reason: string): void {
  const key = validationKey(nodeId, inputId);
  interactiveInput.update((s) => ({
    ...s,
    validationError: { ...s.validationError, [key]: reason },
  }));
  // Auto-clear after 4s so stale errors don't linger when the modal is reopened.
  setTimeout(() => {
    interactiveInput.update((s) => {
      if (s.validationError[key] !== reason) return s;
      const { [key]: _removed, ...rest } = s.validationError;
      return { ...s, validationError: rest };
    });
  }, 4000);
}

export function updateRound(nodeId: string, round: number): void {
  interactiveInput.update((s) => ({ ...s, rounds: { ...s.rounds, [nodeId]: round } }));
}

export function openModal(prompt: PendingPrompt | null): void {
  interactiveInput.update((s) => ({ ...s, activeModal: prompt }));
}

export function closeModal(): void {
  interactiveInput.update((s) => ({ ...s, activeModal: null }));
}

export function openModalForNode(nodeId: string): PendingPrompt | null {
  const state = get(interactiveInput);
  const prompt = state.pendingPrompts.find((p) => p.nodeId === nodeId) || null;
  if (prompt) openModal(prompt);
  return prompt;
}

let toastSeq = 0;
export function pushToast(kind: ToastKind, text: string): void {
  toastSeq += 1;
  const toast: Toast = { kind, text, id: toastSeq };
  interactiveInput.update((s) => ({ ...s, lastToast: toast }));
}

export function dismissNode(nodeId: string): void {
  interactiveInput.update((s) => {
    const pendingPrompts = s.pendingPrompts.filter((p) => p.nodeId !== nodeId);
    const activeModal =
      s.activeModal && s.activeModal.nodeId === nodeId ? null : s.activeModal;
    return { ...s, pendingPrompts, activeModal };
  });
}

export function resetInteractiveInput(): void {
  interactiveInput.set({ ...initial, validationError: {}, rounds: {} });
}

// ui-ast-U6 — §6.1: singular `pendingPrompt` writable + `pendingAst` derived.
//
// The list store (`interactiveInput.pendingPrompts`) is the transport layer;
// `pendingPrompt` is the UI-facing "currently-in-focus" slot that modal code
// subscribes to. Kept writable (not derived) because the AST renderer needs
// explicit control over which prompt is projected — e.g. when the modal swaps
// between queued prompts without a backing list mutation (§6.3).
export const pendingPrompt = writable<PendingPrompt | null>(null);

export const pendingAst = derived<typeof pendingPrompt, UIAST | null>(
  pendingPrompt,
  ($p): UIAST | null => {
    if (!$p?.structured) return null;
    try {
      const ast = JSON.parse($p.structured) as UIAST;
      if (ast?.version !== '1' || !Array.isArray(ast.nodes)) return null;
      return ast;
    } catch {
      return null;
    }
  },
);

// Dev-only test seam — Playwright / Claude-in-Chrome seed pendingPrompt to
// exercise the modal AST path without going through the backend emit cycle.
// `import.meta.env.DEV` is a Vite static literal so this block is dead code
// in production bundles.
if (import.meta.env.DEV && typeof window !== 'undefined') {
  (window as unknown as { __mashed_setPendingPromptForTests?: (p: PendingPrompt | null) => void })
    .__mashed_setPendingPromptForTests = (p) => pendingPrompt.set(p);
}
