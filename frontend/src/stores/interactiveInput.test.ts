// Store tests for src/stores/interactiveInput.ts (S6).
//
// Covers:
//  - upsertPrompt replaces existing (nodeId,inputId) entry, appends new
//  - resolveInput removes matching entry + closes activeModal if matching
//  - setValidationError stores reason, auto-clears after 4s
//  - updateRound writes round per node
//  - openModal / closeModal / openModalForNode behaviour
//  - dismissNode removes all prompts for a node
//  - pushToast stores last toast

import { describe, it, expect, beforeEach, vi } from 'vitest';
import { get } from 'svelte/store';
import {
  interactiveInput,
  upsertPrompt,
  resolveInput,
  setValidationError,
  updateRound,
  openModal,
  closeModal,
  openModalForNode,
  pushToast,
  dismissNode,
  resetInteractiveInput,
  validationKey,
  pendingPrompt,
  pendingAst,
  type PendingPrompt,
} from './interactiveInput';

function makePrompt(overrides: Partial<PendingPrompt> = {}): PendingPrompt {
  return {
    execId: 'exec-1',
    nodeId: 'n1',
    inputId: 'topic',
    prompt: 'Pick one',
    shape: 'choice',
    options: ['a', 'b'],
    round: 1,
    createdAt: 1_700_000_000_000,
    promptId: 'p-1',
    required: true,
    ...overrides,
  };
}

describe('interactiveInput store', () => {
  beforeEach(() => {
    resetInteractiveInput();
    vi.useFakeTimers();
  });

  it('upsertPrompt appends a new prompt', () => {
    upsertPrompt(makePrompt());
    const state = get(interactiveInput);
    expect(state.pendingPrompts).toHaveLength(1);
    expect(state.pendingPrompts[0].nodeId).toBe('n1');
  });

  it('upsertPrompt replaces an existing (nodeId, inputId) entry', () => {
    upsertPrompt(makePrompt({ prompt: 'first' }));
    upsertPrompt(makePrompt({ prompt: 'second' }));
    const state = get(interactiveInput);
    expect(state.pendingPrompts).toHaveLength(1);
    expect(state.pendingPrompts[0].prompt).toBe('second');
  });

  it('upsertPrompt keeps separate entries for different inputIds', () => {
    upsertPrompt(makePrompt({ inputId: 'topic' }));
    upsertPrompt(makePrompt({ inputId: 'approval' }));
    expect(get(interactiveInput).pendingPrompts).toHaveLength(2);
  });

  it('resolveInput removes the matching entry', () => {
    upsertPrompt(makePrompt({ inputId: 'topic' }));
    upsertPrompt(makePrompt({ inputId: 'approval' }));
    resolveInput('n1', 'topic');
    const state = get(interactiveInput);
    expect(state.pendingPrompts).toHaveLength(1);
    expect(state.pendingPrompts[0].inputId).toBe('approval');
  });

  it('resolveInput closes activeModal when matching', () => {
    const p = makePrompt();
    upsertPrompt(p);
    openModal(p);
    expect(get(interactiveInput).activeModal).not.toBeNull();
    resolveInput('n1', 'topic');
    expect(get(interactiveInput).activeModal).toBeNull();
  });

  it('setValidationError stores reason and auto-clears after 4s', () => {
    setValidationError('n1', 'topic', 'value must be one of [a b]');
    const key = validationKey('n1', 'topic');
    expect(get(interactiveInput).validationError[key]).toBe('value must be one of [a b]');
    vi.advanceTimersByTime(4000);
    expect(get(interactiveInput).validationError[key]).toBeUndefined();
  });

  it('updateRound writes a per-node round', () => {
    updateRound('n1', 3);
    expect(get(interactiveInput).rounds['n1']).toBe(3);
    updateRound('n1', 4);
    expect(get(interactiveInput).rounds['n1']).toBe(4);
  });

  it('openModalForNode opens the first matching prompt', () => {
    upsertPrompt(makePrompt({ inputId: 'topic' }));
    const opened = openModalForNode('n1');
    expect(opened?.inputId).toBe('topic');
    expect(get(interactiveInput).activeModal?.inputId).toBe('topic');
  });

  it('closeModal nulls out activeModal but leaves prompts intact', () => {
    const p = makePrompt();
    upsertPrompt(p);
    openModal(p);
    closeModal();
    expect(get(interactiveInput).activeModal).toBeNull();
    expect(get(interactiveInput).pendingPrompts).toHaveLength(1);
  });

  it('dismissNode removes all prompts for a node + clears modal', () => {
    upsertPrompt(makePrompt({ inputId: 'a' }));
    upsertPrompt(makePrompt({ inputId: 'b' }));
    openModal(makePrompt({ inputId: 'a' }));
    dismissNode('n1');
    const state = get(interactiveInput);
    expect(state.pendingPrompts).toHaveLength(0);
    expect(state.activeModal).toBeNull();
  });

  it('pushToast stores the last toast with unique id', () => {
    pushToast('success', 'gate satisfied');
    const first = get(interactiveInput).lastToast;
    expect(first?.kind).toBe('success');
    pushToast('warn', 'round limit');
    const second = get(interactiveInput).lastToast;
    expect(second?.kind).toBe('warn');
    expect(second?.id).not.toBe(first?.id);
  });
});

// ui-ast-U6 — PendingPrompt.structured + pendingAst derived store (spec §6.1).
// Fails until T2 (GREEN) adds `structured`, `pendingPrompt`, and `pendingAst`.

describe('ui-ast-U6 pendingAst derived store', () => {
  beforeEach(() => {
    resetInteractiveInput();
    pendingPrompt.set(null);
  });

  it('AC1_structured_field_optional — PendingPrompt.structured typechecks as string | undefined', () => {
    // Both variants must compile — the contract is enforced by tsc, not at runtime.
    const withStructured: PendingPrompt = { ...makePrompt(), structured: 'x' };
    const withoutStructured: PendingPrompt = makePrompt();
    expect(withStructured.structured).toBe('x');
    expect(withoutStructured.structured).toBeUndefined();
  });

  it('AC2_pending_ast_parses_valid_v1 — emits parsed UIAST with version "1"', () => {
    pendingPrompt.set(makePrompt({ structured: JSON.stringify({ version: '1', nodes: [] }) }));
    const ast = get(pendingAst);
    expect(ast).not.toBeNull();
    expect(ast?.version).toBe('1');
  });

  it('AC3_pending_ast_null_on_malformed — emits null when structured is not JSON', () => {
    pendingPrompt.set(makePrompt({ structured: 'not-json' }));
    expect(get(pendingAst)).toBeNull();
  });

  it('AC4_pending_ast_null_on_unknown_version — emits null when version !== "1"', () => {
    pendingPrompt.set(makePrompt({ structured: '{"version":"2"}' }));
    expect(get(pendingAst)).toBeNull();
  });
});
