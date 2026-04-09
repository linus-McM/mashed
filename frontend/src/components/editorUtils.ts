export type EditorType = 'markdown' | 'image' | 'code';

const MARKDOWN_EXTENSIONS = new Set(['md', 'mdx', 'markdown']);
const IMAGE_EXTENSIONS = new Set(['png', 'jpg', 'jpeg', 'gif', 'svg', 'webp', 'bmp', 'ico']);

export function getEditorType(filePath: string): EditorType {
  if (!filePath) return 'code';

  const ext = filePath.split('.').pop()?.toLowerCase() ?? '';

  if (MARKDOWN_EXTENSIONS.has(ext)) return 'markdown';
  if (IMAGE_EXTENSIONS.has(ext)) return 'image';
  return 'code';
}
