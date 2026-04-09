import { describe, it, expect } from 'vitest';
import { buildImagePath, isDataUri, type ImageViewerState } from '../imageViewerUtils';

describe('ImageViewer utilities', () => {
  describe('AC-1: buildImagePath constructs correct full path', () => {
    it.each([
      ['/home/user/repo', 'assets/logo.png', '/home/user/repo/assets/logo.png'],
      ['/home/user/repo/', 'assets/logo.png', '/home/user/repo/assets/logo.png'],
      ['/home/user/repo', '/assets/logo.png', '/home/user/repo/assets/logo.png'],
      ['/home/user/repo/', '/assets/logo.png', '/home/user/repo/assets/logo.png'],
      ['/repo', 'image.png', '/repo/image.png'],
      ['/repo', 'a/b/c/d/image.png', '/repo/a/b/c/d/image.png'],
      ['', 'file.png', '/file.png'],
      ['/repo', '', '/repo/'],
      ['/repo///', '///file.png', '/repo/file.png'],
    ])('buildImagePath(%s, %s) => %s', (repo, file, expected) => {
      expect(buildImagePath(repo, file)).toBe(expected);
    });
  });

  describe('AC-1: isDataUri validates data URI format', () => {
    it.each([
      ['data:image/png;base64,iVBOR...', true],
      ['data:image/jpeg;base64,/9j/4AAQ...', true],
      ['data:image/svg+xml;base64,PHN2Zw==', true],
      ['', false],
      ['not-a-data-uri', false],
      ['data:', false],
      ['data:text/plain;base64,abc', false],
      [' data:image/png;base64,abc', false],
    ])('isDataUri(%s) => %s', (input, expected) => {
      expect(isDataUri(input)).toBe(expected);
    });
  });

  describe('AC-4/AC-5: ImageViewerState discriminated union', () => {
    it('loading state has status only', () => {
      const state: ImageViewerState = { status: 'loading' };
      expect(state.status).toBe('loading');
    });

    it('loaded state carries dataUri', () => {
      const state: ImageViewerState = {
        status: 'loaded',
        dataUri: 'data:image/png;base64,abc',
      };
      expect(state.status).toBe('loaded');
      expect(state.dataUri).toContain('data:image/');
    });

    it('error state carries message', () => {
      const state: ImageViewerState = {
        status: 'error',
        message: 'Failed to read file',
      };
      expect(state.status).toBe('error');
      expect(state.message).toBe('Failed to read file');
    });
  });
});
