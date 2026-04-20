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

import { writable, get } from 'svelte/store';

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
