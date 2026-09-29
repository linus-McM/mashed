// If these drift from Go DefaultMarkdownMenuSettings, Story 01 is authoritative.
//
// Frontend-side DEFAULTS exist as a safety net for hydration failures only. The
// Go backend (`internal/...` DefaultMarkdownMenuSettings) is the source of truth;
// keep this constant in lockstep with it.

import { writable, get, type Writable } from 'svelte/store';
import { SetMarkdownMenuSettings } from '../../../wailsjs/go/main/App.js';
import type { main } from '../../../wailsjs/go/models';

export type MarkdownMenuSettings = main.MarkdownMenuSettings;

const DEFAULTS: MarkdownMenuSettings = {
  bold: true,
  italic: true,
  strikethrough: true,
  code: true,
  link: true,
  latex: false,
};

export const markdownMenuSettings: Writable<MarkdownMenuSettings> = writable({ ...DEFAULTS });
export const markdownMenuDirty: Writable<boolean> = writable(false);

/**
 * Merge backend settings into the store, falling back to DEFAULTS per field.
 * Null/undefined input resets to DEFAULTS — safe for pre-migration configs
 * where `cfg.markdownMenu` may be absent.
 */
export function initMarkdownMenuSettings(
  settings: Partial<MarkdownMenuSettings> | null | undefined,
): void {
  const merged: MarkdownMenuSettings = { ...DEFAULTS };
  if (settings) {
    const source = settings as Record<string, unknown>;
    const target = merged as unknown as Record<string, unknown>;
    for (const key of Object.keys(DEFAULTS) as (keyof MarkdownMenuSettings)[]) {
      const value = source[key];
      if (value !== undefined && value !== null) {
        target[key] = value;
      }
    }
  }
  markdownMenuSettings.set(merged);
}

/**
 * Update a single menu item and persist the full object to the Go backend.
 *
 * Order is critical:
 *   1. Store is updated synchronously so subscribers (Settings panel) see the
 *      new value immediately.
 *   2. Dirty flag is flipped synchronously BEFORE awaiting Wails so the
 *      MarkdownEditor re-init guard trips before persistence returns (AC-6).
 *   3. Wails call is awaited last.
 *
 * Tradeoff on rejection: if `SetMarkdownMenuSettings` rejects, the store value
 * is already updated locally. We log and do NOT revert — the user sees the UI
 * reflect the toggle; persistence fails silently. Re-opening the app re-reads
 * from disk, reverting the failed change. Acceptable for MVP.
 */
export async function updateMarkdownMenuItem<K extends keyof MarkdownMenuSettings>(
  key: K,
  value: MarkdownMenuSettings[K],
): Promise<void> {
  markdownMenuSettings.update(current => ({ ...current, [key]: value }));
  markdownMenuDirty.set(true);
  try {
    await SetMarkdownMenuSettings(get(markdownMenuSettings));
  } catch (err) {
    // See tradeoff comment above — do not revert the store.
    console.error('SetMarkdownMenuSettings failed', err);
  }
}

/**
 * Reset the dirty flag. Does NOT persist — persistence already happened in
 * `updateMarkdownMenuItem`. Called by `MarkdownEditor.svelte` after it has
 * consumed the dirty signal and re-initialised its toolbar.
 */
export function clearMarkdownMenuDirty(): void {
  markdownMenuDirty.set(false);
}
