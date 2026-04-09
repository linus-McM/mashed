export type ImageViewerState =
  | { status: 'loading' }
  | { status: 'loaded'; dataUri: string }
  | { status: 'error'; message: string };

export function buildImagePath(repoPath: string, filePath: string): string {
  const cleanRepo = repoPath.replace(/\/+$/, '');
  const cleanFile = filePath.replace(/^\/+/, '');
  return cleanRepo + '/' + cleanFile;
}

export function isDataUri(str: string): boolean {
  return str.startsWith('data:image/') && str.includes(';base64,');
}
