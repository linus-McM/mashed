/**
 * Convert a flat array of file paths into a tree structure.
 * @param {string[]} paths - flat array of relative paths, e.g. ["src/main.go", "src/lib/util.go"]
 * @returns {TreeNode[]} sorted tree -- directories first, then alphabetical
 *
 * TreeNode = { name: string, path: string, type: 'file' | 'dir', children?: TreeNode[] }
 */
export function flatPathsToTree(paths) {
  const unique = [...new Set(paths)];
  const root = { name: '', path: '', type: 'dir', children: [] };

  for (const p of unique) {
    const parts = p.split('/');
    let current = root;

    for (let i = 0; i < parts.length; i++) {
      const name = parts[i];
      const isFile = i === parts.length - 1;
      const partPath = parts.slice(0, i + 1).join('/');

      if (isFile) {
        current.children.push({ name, path: partPath, type: 'file' });
      } else {
        let dir = current.children.find(c => c.type === 'dir' && c.name === name);
        if (!dir) {
          dir = { name, path: partPath, type: 'dir', children: [] };
          current.children.push(dir);
        }
        current = dir;
      }
    }
  }

  // Sort recursively: dirs first, then alphabetical
  function sortTree(nodes) {
    nodes.sort((a, b) => {
      if (a.type !== b.type) return a.type === 'dir' ? -1 : 1;
      return a.name.localeCompare(b.name);
    });
    for (const n of nodes) {
      if (n.children) sortTree(n.children);
    }
  }

  sortTree(root.children);
  return root.children;
}
