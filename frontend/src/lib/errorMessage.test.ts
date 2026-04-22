import { describe, it, expect } from 'vitest';
import { errorMessage } from './errorMessage';

describe('errorMessage', () => {
  it('returns the message for an Error instance', () => {
    expect(errorMessage(new Error('boom'))).toBe('boom');
  });

  it('returns the message for a subclass of Error', () => {
    class CustomError extends Error {}
    expect(errorMessage(new CustomError('custom'))).toBe('custom');
  });

  it('returns a string input as-is', () => {
    expect(errorMessage('already-a-string')).toBe('already-a-string');
  });

  it('returns a non-empty string for null without throwing', () => {
    const out = errorMessage(null);
    expect(typeof out).toBe('string');
    expect(out.length).toBeGreaterThan(0);
  });

  it('returns a non-empty string for undefined without throwing', () => {
    const out = errorMessage(undefined);
    expect(typeof out).toBe('string');
    expect(out.length).toBeGreaterThan(0);
  });

  it('serialises a plain object to JSON containing its keys', () => {
    const out = errorMessage({ code: 'E42', detail: 'x' });
    expect(out).toContain('E42');
    expect(out).toContain('detail');
  });

  it('serialises an array', () => {
    expect(errorMessage([1, 2, 3])).toBe('[1,2,3]');
  });

  it('handles a circular reference by falling back to String()', () => {
    const o: { self?: unknown } = {};
    o.self = o;
    let out = '';
    expect(() => {
      out = errorMessage(o);
    }).not.toThrow();
    expect(typeof out).toBe('string');
    expect(out.length).toBeGreaterThan(0);
  });

  it('handles a number input', () => {
    expect(errorMessage(42)).toBe('42');
  });

  it('handles a boolean input', () => {
    expect(errorMessage(false)).toBe('false');
  });
});
