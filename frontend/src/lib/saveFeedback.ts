// Shared auto-save routine for the editors. A rejected save never touches
// the editor buffer; it returns a message the editor announces (spec R7).

export type SaveResult = { status: 'saved' | 'error'; message: string };

/** Turn a backend save rejection into a user-facing reason. */
export function describeSaveError(e: unknown): string {
  const raw = e instanceof Error ? e.message : String(e ?? '');
  if (/outside allowed roots/i.test(raw)) {
    return 'Save blocked: this file is outside your home folder and dev directory.';
  }
  if (/write-protected/i.test(raw)) {
    return 'Save blocked: this file is write-protected.';
  }
  return raw ? `Save failed: ${raw}` : 'Save failed.';
}

/** Read the buffer once, write it, and report the outcome. */
export async function runSave(
  getContent: () => string,
  write: (content: string) => Promise<unknown>,
): Promise<SaveResult> {
  try {
    await write(getContent());
    return { status: 'saved', message: '' };
  } catch (e) {
    return { status: 'error', message: describeSaveError(e) };
  }
}
