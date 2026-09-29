import type { MarkdownMenuSettings } from '../lib/stores/markdownMenuSettings';

export type SaveStatus = 'idle' | 'modified' | 'saving' | 'saved' | 'error';

/**
 * Canonical ordered keys for the markdown selection-toolbar. Mirrors
 * `ITEMS` in `markdownToolbarBuilder.ts` and the `:nth-of-type` order in
 * `crepe-mashed.css`. Exported for tests and downstream consumers.
 */
export const TOOLBAR_KEYS = [
  'bold',
  'italic',
  'strikethrough',
  'code',
  'link',
  'latex',
] as const;

export type ToolbarKey = (typeof TOOLBAR_KEYS)[number];

/**
 * Apply `data-toolbar-<key>="on"|"off"` attributes to the `.milkdown` root
 * within the given container. Pure DOM side-effect; safe to call with a
 * null container (no-op).
 *
 * CSS rules in `crepe-mashed.css` pick up the `off` attributes and hide the
 * corresponding default Crepe toolbar button. No editor teardown required.
 *
 * Story 06 fallback strategy: the primary featureConfigs.Toolbar.buildToolbar
 * path cannot be used because Crepe v7.20.0 does not export the needed
 * internal subpaths — see `markdownToolbarBuilder.ts` header.
 */
export function applyToolbarAttributes(
  container: Element | null,
  settings: MarkdownMenuSettings,
): void {
  if (!container) return;
  const root = container.querySelector('.milkdown') ?? container;
  for (const key of TOOLBAR_KEYS) {
    root.setAttribute(`data-toolbar-${key}`, settings[key] ? 'on' : 'off');
  }
}

/**
 * Pure decision helper used by MarkdownEditor.svelte's deferred-apply
 * reactive block. Given the previous applied snapshot (JSON string), the
 * current dirty flag, editor-mounted flag, and loading flag, returns the
 * new JSON snapshot to persist and apply if an apply should happen now —
 * or `null` if the guards block, or settings are unchanged.
 *
 * AC-2: dirty=true → null (blocked while user toggling in Settings)
 * AC-3: dirty=false + mounted + !loading + changed → new JSON
 * AC-4: dirty=false + unchanged → null (no redundant apply)
 * AC-6: loading=true → null (init in flight)
 */
export function computeToolbarApplyTarget(args: {
  dirty: boolean;
  mounted: boolean;
  loading: boolean;
  settings: MarkdownMenuSettings;
  lastApplied: string;
}): string | null {
  if (args.dirty) return null;
  if (!args.mounted) return null;
  if (args.loading) return null;
  const current = JSON.stringify(args.settings);
  if (current === args.lastApplied) return null;
  return current;
}


export interface DebouncedSave {
  schedule(): void;
  flush(): void;
  cancel(): void;
  destroy(): void;
}

export function createDebouncedSave(
  saveFn: () => void | Promise<void>,
  delayMs: number,
): DebouncedSave {
  let timer: ReturnType<typeof setTimeout> | null = null;
  let destroyed = false;

  function cancel() {
    if (timer !== null) {
      clearTimeout(timer);
      timer = null;
    }
  }

  return {
    schedule() {
      if (destroyed) return;
      cancel();
      timer = setTimeout(() => {
        timer = null;
        saveFn();
      }, delayMs);
    },
    flush() {
      if (destroyed) return;
      cancel();
      saveFn();
    },
    cancel,
    destroy() {
      destroyed = true;
      cancel();
    },
  };
}
