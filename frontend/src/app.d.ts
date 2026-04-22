// Global ambient declarations for the frontend.
//
// TS does not allow `declare global` inside a Svelte/component module
// (it errors: "An ambient module declaration is only allowed at the top
// level in a file"), so any window augmentation MUST live here.

export {};

declare global {
  interface Window {
    /**
     * Dev-only Playwright navigation seam defined in `App.svelte`.
     * Installed when `import.meta.env.DEV` is true; absent in production.
     * Returns `true` when the navigation happened so the test can await it.
     */
    __mashed_gotoWorkflows?: (repoPath?: string) => boolean;
  }
}
