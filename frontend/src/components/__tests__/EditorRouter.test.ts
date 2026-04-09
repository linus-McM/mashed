import { describe, it, expect } from 'vitest';
import { getEditorType } from '../editorUtils';

describe('getEditorType', () => {
  describe('AC-1: Markdown routing', () => {
    it.each([
      ['document.md'],
      ['README.mdx'],
      ['notes.markdown'],
    ])('routes %s to markdown', (filePath) => {
      expect(getEditorType(filePath)).toBe('markdown');
    });
  });

  describe('AC-2: Image routing', () => {
    it.each([
      ['photo.png'],
      ['banner.jpg'],
      ['avatar.jpeg'],
      ['animation.gif'],
      ['logo.svg'],
      ['hero.webp'],
      ['icon.bmp'],
      ['favicon.ico'],
    ])('routes %s to image', (filePath) => {
      expect(getEditorType(filePath)).toBe('image');
    });
  });

  describe('AC-3: Code fallback', () => {
    it.each([
      ['index.ts'],
      ['main.go'],
      ['config.json'],
    ])('routes %s to code', (filePath) => {
      expect(getEditorType(filePath)).toBe('code');
    });

    it('routes dotfiles with no extension to code', () => {
      expect(getEditorType('.gitignore')).toBe('code');
    });

    it('routes empty string to code', () => {
      expect(getEditorType('')).toBe('code');
    });

    it('routes undefined to code', () => {
      expect(getEditorType(undefined as unknown as string)).toBe('code');
    });

    it('routes null to code', () => {
      expect(getEditorType(null as unknown as string)).toBe('code');
    });
  });

  describe('AC-5: Case-insensitive matching', () => {
    it.each([
      ['README.MD', 'markdown'],
      ['photo.PNG', 'image'],
      ['STYLE.CSS', 'code'],
      ['Notes.Markdown', 'markdown'],
      ['HERO.WEBP', 'image'],
    ])('routes %s to %s (case-insensitive)', (filePath, expected) => {
      expect(getEditorType(filePath)).toBe(expected);
    });
  });
});
