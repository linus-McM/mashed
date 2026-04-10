import { describe, it, expect } from 'vitest';
import { parseFriendlyTarget, type FriendlyTarget } from './bmadSessionName';

describe('parseFriendlyTarget', () => {
  it('AC-6: parses valid target with :0.0 suffix', () => {
    const input = 'bmad-surfseer-main-create-story-a1b2c3d4:0.0';
    const result = parseFriendlyTarget(input);
    expect(result).not.toBeNull();
    expect(result).toEqual<FriendlyTarget>({
      repo: 'surfseer',
      branch: 'main',
      label: 'create-story',
      hash: 'a1b2c3d4',
      raw: 'bmad-surfseer-main-create-story-a1b2c3d4:0.0',
    });
  });

  it('AC-6: parses valid target without :0.0 suffix', () => {
    const input = 'bmad-surfseer-main-create-story-a1b2c3d4';
    const result = parseFriendlyTarget(input);
    expect(result).not.toBeNull();
    expect(result).toEqual<FriendlyTarget>({
      repo: 'surfseer',
      branch: 'main',
      label: 'create-story',
      hash: 'a1b2c3d4',
      raw: 'bmad-surfseer-main-create-story-a1b2c3d4',
    });
  });

  it('AC-6: label with internal dashes parses correctly', () => {
    const input = 'bmad-repo-main-draft-prd-deadbeef:0.0';
    const result = parseFriendlyTarget(input);
    expect(result).not.toBeNull();
    expect(result?.repo).toBe('repo');
    expect(result?.branch).toBe('main');
    expect(result?.label).toBe('draft-prd');
    expect(result?.hash).toBe('deadbeef');
  });

  it('AC-6: single-word label', () => {
    const input = 'bmad-foo-bar-baz-12345678';
    const result = parseFriendlyTarget(input);
    expect(result).not.toBeNull();
    expect(result?.repo).toBe('foo');
    expect(result?.branch).toBe('bar');
    expect(result?.label).toBe('baz');
    expect(result?.hash).toBe('12345678');
  });

  it('AC-7: returns null for empty string', () => {
    expect(parseFriendlyTarget('')).toBeNull();
  });

  it('AC-7: returns null for garbage string', () => {
    expect(parseFriendlyTarget('garbage')).toBeNull();
  });

  it('AC-7: returns null for bmad prefix with insufficient components', () => {
    expect(parseFriendlyTarget('bmad-')).toBeNull();
    expect(parseFriendlyTarget('bmad-short')).toBeNull();
    expect(parseFriendlyTarget('bmad-a-b')).toBeNull();
    expect(parseFriendlyTarget('bmad-a-b-c')).toBeNull();
  });

  it('AC-7: returns null for non-hex hash suffix', () => {
    expect(parseFriendlyTarget('bmad-repo-main-label-ZZZZZZZZ')).toBeNull();
    expect(parseFriendlyTarget('bmad-repo-main-label-abc')).toBeNull();
    expect(parseFriendlyTarget('bmad-repo-main-label-abcdefghi')).toBeNull();
  });

  it('AC-7: returns null for wrong prefix', () => {
    expect(parseFriendlyTarget('foo-repo-main-label-12345678')).toBeNull();
    expect(parseFriendlyTarget('BMAD-repo-main-label-12345678')).toBeNull();
  });
});
