import { describe, it, expect, vi } from 'vitest';
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { createKeyHandler, SHIFT_ENTER_SEQUENCE } from '../terminalKeys';

type KeyInit = { type?: string; key?: string; shiftKey?: boolean; ctrlKey?: boolean; altKey?: boolean; metaKey?: boolean };
const ev = (init: KeyInit) =>
  ({ type: 'keydown', key: 'Enter', shiftKey: false, ctrlKey: false, altKey: false, metaKey: false, ...init }) as KeyboardEvent;

describe('terminal key handling', () => {
  it('Shift+Enter sends a newline (Ctrl+J) instead of submitting', () => {
    const send = vi.fn();
    const handle = createKeyHandler(send);
    expect(handle(ev({ shiftKey: true }))).toBe(false); // xterm must not also send CR
    expect(send).toHaveBeenCalledWith('\n');
    expect(SHIFT_ENTER_SEQUENCE).toBe('\n');
  });

  it('ignores the keyup of Shift+Enter without sending again', () => {
    const send = vi.fn();
    const handle = createKeyHandler(send);
    expect(handle(ev({ type: 'keyup', shiftKey: true }))).toBe(false);
    expect(send).not.toHaveBeenCalled();
  });

  it('leaves plain Enter and other chords to xterm', () => {
    const send = vi.fn();
    const handle = createKeyHandler(send);
    expect(handle(ev({}))).toBe(true);
    expect(handle(ev({ shiftKey: true, ctrlKey: true }))).toBe(true);
    expect(handle(ev({ shiftKey: true, metaKey: true }))).toBe(true);
    expect(handle(ev({ key: 'a', shiftKey: true }))).toBe(true);
    expect(send).not.toHaveBeenCalled();
  });
});

describe('xterm.js version', () => {
  const pkg = JSON.parse(readFileSync(resolve(__dirname, '../../../package.json'), 'utf8'));
  const deps = { ...pkg.dependencies, ...pkg.devDependencies } as Record<string, string>;

  it('uses @xterm/xterm 6+ (synchronized output, DEC 2026)', () => {
    expect(Number(String(deps['@xterm/xterm']).replace(/^[^\d]*/, '').split('.')[0])).toBeGreaterThanOrEqual(6);
  });

  it('drops the unused, v5-only canvas addon', () => {
    expect(deps['@xterm/addon-canvas']).toBeUndefined();
  });
});
