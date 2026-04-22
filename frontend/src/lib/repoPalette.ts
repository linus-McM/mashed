// lib/repoPalette.ts — per-repo border-colour palette shared by NotificationFeed.
//
// The values below are *user-facing palette choices*, not theme tokens. They
// persist to localStorage under `mashed:repoBorderColors` keyed by repo name,
// so they need stable hex literals — a theme switch must NOT reshuffle a
// user's pinned colours. This is the one place in the UI where hex literals
// are intentional. See cerebrum 2026-04-10 for the general ban on hex —
// this module is the documented exception.

/**
 * Sentinel value meaning "no custom colour". Chosen to match
 * `--border-subtle` (the default repo-group border) so a user who hasn't
 * picked a colour sees the unmodified system border.
 *
 * localStorage values written before this constant existed still compare
 * equal, so no migration is needed.
 */
export const REPO_BORDER_NONE = '#1e2530';

/**
 * The fixed colour palette the user picks from in the per-repo border-colour
 * popover. Order is significant (appears left-to-right in the UI).
 *
 * Values mirror the macOS window-control palette + common accent hues:
 *   red, amber, green, blue, purple, teal, orange, red-2, accent-green,
 *   pink, and finally REPO_BORDER_NONE as the clear/default choice.
 */
export const REPO_BORDER_PALETTE: readonly string[] = [
  '#ff5f57',
  '#febc2e',
  '#28c840',
  '#3d9eff',
  '#9d6fff',
  '#00c4b3',
  '#f0a500',
  '#e84545',
  '#00e57a',
  '#ff6ec7',
  REPO_BORDER_NONE,
];
