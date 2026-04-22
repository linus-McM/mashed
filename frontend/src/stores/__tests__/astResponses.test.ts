/**
 * @vitest-environment jsdom
 *
 * Story ui-ast-U7 — AC-2.
 * RED phase: `makeAstResponses` does not exist yet; these tests must fail
 * until `frontend/src/stores/astResponses.ts` exports the factory.
 */
import { describe, it, expect } from 'vitest';
import { get } from 'svelte/store';

import { makeAstResponses } from '../astResponses';

describe('astResponses store', () => {
  it('AC2_returns_writable — makeAstResponses yields a Writable<Record<string,string>>', () => {
    const r = makeAstResponses();
    expect(typeof r.subscribe).toBe('function');
    expect(typeof r.set).toBe('function');
    expect(typeof r.update).toBe('function');
    expect(get(r)).toEqual({});
  });

  it('AC2_responses_accumulate — three keys a/b/c accumulate into one map', () => {
    const r = makeAstResponses();
    r.update((m) => ({ ...m, a: 'alpha' }));
    r.update((m) => ({ ...m, b: 'beta' }));
    r.update((m) => ({ ...m, c: 'gamma' }));
    expect(get(r)).toEqual({ a: 'alpha', b: 'beta', c: 'gamma' });
  });

  it('AC2_per_instance_isolation — separate makeAstResponses() calls yield independent stores', () => {
    const a = makeAstResponses();
    const b = makeAstResponses();
    a.update((m) => ({ ...m, x: '1' }));
    b.update((m) => ({ ...m, y: '2' }));
    expect(get(a)).toEqual({ x: '1' });
    expect(get(b)).toEqual({ y: '2' });
  });
});
