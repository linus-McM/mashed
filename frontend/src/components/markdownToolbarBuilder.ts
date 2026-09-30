// Strategy: fallback (CSS masking)
//
// The primary-strategy imports from Story 03's plan (e.g. `@milkdown/crepe/icons`,
// `@milkdown/crepe/utils/group-builder`, `@milkdown/crepe/feature/toolbar/config`,
// `@milkdown/crepe/feature/latex/command`, `@milkdown/crepe/feature/latex/inline-latex`)
// are NOT listed in `@milkdown/crepe`'s `exports` map (verified against
// `node_modules/@milkdown/crepe/package.json`, v7.20.0). Importing them would fail
// module resolution at build time. The top-level bundle re-exports only
// `CrepeFeature` and the core — none of the per-item icons or commands.
//
// We therefore ship the fallback: the builder still creates the `'toolbar'` group
// and conditionally `addItem`s for each enabled key in the canonical order, using
// lightweight placeholder configs (icon string, no-op active/onRun). Story 06
// supplements this with container-level `data-toolbar-<key>="off"` attributes on
// `.milkdown` and CSS rules in `frontend/src/styles/crepe-mashed.css` that hide
// unwanted built-in items when a custom `buildToolbar` is not plumbed through.
//
// Contract preserved:
//   - Pure function (no external closures, no mutation of `settings`)
//   - Fixed order: bold -> italic -> strikethrough -> code -> link -> latex
//   - Only enabled keys get `addItem`; all-disabled still yields one empty group

import type { MarkdownMenuSettings } from '../lib/stores/markdownMenuSettings';

/**
 * Minimal structural type for the Crepe toolbar group builder. We type against
 * only the surface we use so we do not depend on Milkdown subpath exports that
 * are not in the package's `exports` map.
 */
export interface MinimalGroupBuilder<TItem> {
  addGroup(key: string): MinimalItemGroup<TItem>;
}

export interface MinimalItemGroup<TItem> {
  addItem(key: string, item: TItem): MinimalItemGroup<TItem>;
}

/**
 * Placeholder item config. Real Crepe item configs consume `ctx` via closures
 * that read the editor's command registry; in the fallback strategy those
 * commands are not importable, so we ship safe no-ops. When Story 06 wires the
 * builder into the editor, the built-in items stay active and CSS masking
 * hides the disabled ones — these placeholders do not fight that behavior
 * because we never pass this builder through `featureConfigs` until the
 * primary-strategy imports become available.
 */
export interface ToolbarItemConfig {
  icon: string;
  active: (ctx: unknown) => boolean;
  onRun: (ctx: unknown) => void;
}

type ToolbarKey = 'bold' | 'italic' | 'strikethrough' | 'code' | 'link' | 'latex';

interface ToolbarEntry {
  key: ToolbarKey;
  flag: keyof MarkdownMenuSettings;
  icon: string;
}

// Fixed order — matches Crepe's default `getGroups` sequence so a fully enabled
// toolbar renders identically to the framework default.
const ITEMS: readonly ToolbarEntry[] = [
  { key: 'bold', flag: 'bold', icon: 'bold' },
  { key: 'italic', flag: 'italic', icon: 'italic' },
  { key: 'strikethrough', flag: 'strikethrough', icon: 'strikethrough' },
  { key: 'code', flag: 'code', icon: 'code' },
  { key: 'link', flag: 'link', icon: 'link' },
  { key: 'latex', flag: 'latex', icon: 'functions' },
] as const;

/**
 * Build a Crepe `buildToolbar` callback scoped to the current settings.
 *
 * Pure: does not capture module-level state and never mutates `settings`. The
 * returned callback, when invoked, adds one `'toolbar'` group and conditionally
 * adds items for each truthy setting in the canonical order.
 */
export function buildToolbarFromSettings(
  settings: MarkdownMenuSettings,
): (builder: MinimalGroupBuilder<ToolbarItemConfig>) => void {
  return (builder: MinimalGroupBuilder<ToolbarItemConfig>): void => {
    const group = builder.addGroup('toolbar');
    for (const entry of ITEMS) {
      if (settings[entry.flag]) {
        group.addItem(entry.key, {
          icon: entry.icon,
          active: () => false,
          onRun: () => {
            /* placeholder — real command wired in when primary strategy lands */
          },
        });
      }
    }
  };
}
