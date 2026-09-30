import { describe, it, expect } from 'vitest';
import { isStatusToken } from '../types/status';
import type { StatusToken } from '../types/status';

// Story svelte-check-03 — Acceptance: `isStatusToken` narrows an unknown
// string to `StatusToken` at the frontend boundary. The union is sourced
// from internal/domain/types.go (AgentStatus + EventType) plus the UI-only
// `terminal` token. These tests lock that vocabulary down so accidental
// drift is caught in CI.

describe('isStatusToken', () => {
  it.each([
    // AgentStatusToken — from domain.AgentStatus
    'running',
    'open',
    'finished',
    'waiting',
    'blocked',
    'error',
    'queued',
    'done',
    // EventTypeToken — from domain.EventType (running/error already covered)
    'needs_response',
    'completed',
    'started',
    // UITerminalToken — frontend-only
    'terminal',
  ])('accepts backend status token %s', (token) => {
    expect(isStatusToken(token)).toBe(true);
    if (isStatusToken(token)) {
      // Compile-time: post-narrow, `token` is assignable to StatusToken.
      const narrowed: StatusToken = token;
      expect(typeof narrowed).toBe('string');
    }
  });

  it.each([
    '',
    'idle', // deliberately *not* in the union — used as a worstStatus seed only
    'unknown',
    'RUNNING', // case-sensitive
    ' running', // trimming is the caller's responsibility
    'running ',
    'pending', // must not slip in as a default
  ])('rejects unknown status string %s', (s) => {
    expect(isStatusToken(s)).toBe(false);
  });

  it.each([
    null,
    undefined,
    0,
    1,
    true,
    false,
    NaN,
    {},
    [],
    ['running'],
    { kind: 'running' },
    Symbol('running'),
  ])('rejects non-string input %s', (value) => {
    expect(isStatusToken(value)).toBe(false);
  });

  it('is usable as a TypeScript type predicate', () => {
    // A black-box use of the guard — exercises the `value is StatusToken`
    // branch so coverage reports the narrowing path, not just the boolean.
    const raw: unknown = 'running';
    let observed: StatusToken | null = null;
    if (isStatusToken(raw)) {
      observed = raw;
    }
    expect(observed).toBe('running');
  });
});
