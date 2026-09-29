// Global ambient declarations for the frontend.
//
// TS does not allow `declare global` inside a Svelte/component module
// (it errors: "An ambient module declaration is only allowed at the top
// level in a file"), so any window augmentation MUST live here.

import type { WorkflowDef, GroupedMashedAssets } from './lib/types/wails';
import type { BmadInteractiveEventMap } from './types/bmadEvents';

export {};

declare global {
  interface Window {
    /**
     * Dev-only Playwright navigation seam defined in `App.svelte`.
     * Installed when `import.meta.env.DEV` is true; absent in production.
     * Returns `true` when the navigation happened so the test can await it.
     */
    __mashed_gotoWorkflows?: (repoPath?: string) => boolean;

    // ── WorkflowBuilder dev seams (installed only when import.meta.env.DEV) ──
    //
    // All four are typed as optional so the deletion in onDestroy does not
    // require a cast, and production-bundle access (dead-code-eliminated)
    // does not compile-fail if a downstream test happens to reference it.

    /** Inject a serialised workflow into the canvas without round-tripping through disk. */
    __mashed_loadWorkflowFixture?: (def: WorkflowDef) => boolean;
    /** Seed the mashed-assets sidebar section from fixture data. */
    __mashed_seedMashedAssets?: (grouped: GroupedMashedAssets | null | undefined) => boolean;
    /** Trigger the AssetsChanged refetch path without a real backend event. */
    __mashed_simulateAssetsChanged?: () => boolean;
    /** Flip a specific canvas node to `failed` and raise the sentinel toast. */
    __mashed_simulateNodeFailure?: (nodeId: string, message?: string) => boolean;
    /** Invoke a bmad:node:* handler directly (Playwright event-emit seam). */
    __mashedEmitBmadEvent?: <K extends keyof BmadInteractiveEventMap>(
      name: K,
      payload: BmadInteractiveEventMap[K],
    ) => boolean;
  }
}
