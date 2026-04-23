<script lang="ts">
  import { onMount } from 'svelte';
  import { EventsOn } from '../wailsjs/runtime/runtime.js';
  import { GetNotifications, GetDevDir, GetConfig, ListLocalFonts, SetActiveContext, WriteConsoleLog, GetEditorSettings } from '../wailsjs/go/main/App.js';
  import Setup from './views/Setup.svelte';
  import NotificationFeed from './views/NotificationFeed.svelte';
  import AgentDetail from './views/AgentDetail.svelte';
  import SpawnAgent from './views/SpawnAgent.svelte';
  import Settings from './views/Settings.svelte';
  import WorkflowBuilder from './views/WorkflowBuilder.svelte';
  import NewRepoModal from './components/NewRepoModal.svelte';
  import AboutModal from './components/AboutModal.svelte';
  import QuestionSnackbarStack from './components/bmad/QuestionSnackbarStack.svelte';
  import {
    upsertQuestion,
    upsertIdle,
    dismissQuestion,
    dismissIdle,
  } from './components/bmad/questionSnackbarUtils';
  import type {
    SnackbarEntry,
    QuestionEventLike,
    IdleEventLike,
  } from './components/bmad/questionSnackbarUtils';
  import { Hexagon } from 'lucide-svelte';
  import TitleBar from './components/TitleBar.svelte';
  import { applyTheme } from './lib/stores/theme.js';
  import { loadSavedThemes, restoreImportedThemeFromConfig, loadBundledThemes } from './lib/themeInit.js';
  import { applyFont, registerLocalFonts } from './lib/stores/font.js';
  import { initEditorSettings } from './lib/stores/editorSettings.js';
  import { initMarkdownMenuSettings } from './lib/stores/markdownMenuSettings';
  import { addSession, removeSessionByName } from './lib/stores/sessions';
  import type { domain, main } from '../wailsjs/go/models';
  type WailsNotificationEvent = domain.NotificationEvent;

  /**
   * Boundary adapter for passing the in-memory `AgentNotification` view-model
   * to `AgentDetail`, whose prop is typed as the full Wails-generated
   * `NotificationEvent` class. At runtime they are the same JSON shape
   * (`id`, `read`, `pid`, `isSubAgent`, `convertValues` are never read by
   * AgentDetail in practice); the cast is isolated to this one adapter so
   * no template expression carries a type assertion.
   */
  function asWailsNotification(a: AgentNotification): WailsNotificationEvent {
    return a as unknown as WailsNotificationEvent;
  }

  // ---------------------------------------------------------------------------
  // Local types — view-model shapes that flow through the root component.
  // The canonical definitions live in the backend (internal/domain) and in
  // components that own the data (questionSnackbarUtils.ts, etc). Here we
  // narrow to the fields App actually reads/writes so the top-level view is
  // `strict`-friendly without importing the full Wails model surface.
  // ---------------------------------------------------------------------------

  type ViewRoute = 'loading' | 'setup' | 'feed' | 'detail' | 'settings' | 'workflows';

  /** Minimal shape of a notification/agent as it flows through App. */
  type AgentNotification = {
    agentId: string;
    agentName?: string;
    model?: string;
    repoName?: string;
    repoPath?: string;
    repoBranch?: string;
    eventType: string;
    summary?: string;
    timestamp?: string;
    tokensUsed?: number;
    tokensMax?: number;
    tmuxTarget?: string;
    priority?: number;
    parentAgentId?: string;
  };

  /** Event detail emitted by SpawnAgent on successful spawn. */
  type SpawnedDetail = {
    target: string;
    repo: { name: string; path: string; branch: string };
    model: string;
  };

  /** Payload dispatched from the in-workflow navigation snackbar. */
  type QuestionNavigateDetail = {
    repoPath?: string;
    question?: QuestionEventLike | null;
    entry?: { tmuxTarget?: string; [k: string]: unknown };
  };

  /** Payload of `terminal:session:added`. Mirrors `domain.TerminalSession`. */
  type TerminalSessionAdded = {
    repoPath: string;
    [k: string]: unknown;
  };

  /** Shape of `bmad:node:*:dismissed` — either a string nodeId or object. */
  type DismissedEvent = string | { nodeId?: string };

  /** Subset of `mashedConfig` this component reads. */
  type ConfigSlice = {
    theme?: string;
    importedTheme?: string;
    monoFont?: string;
    fontSize?: number;
    markdownMenu?: main.MarkdownMenuSettings;
  };

  // Dev-only `window.__mashed_gotoWorkflows` is declared in app.d.ts.

  let currentView: ViewRoute = 'loading';
  let selectedAgent: AgentNotification | null = null;
  let notifications: AgentNotification[] = [];
  let showSpawnModal = false;
  let showNewRepoModal = false;
  let showAboutModal = false;
  let builderRepoPath = '';
  let builderRepoBranch = '';
  let toastMessage = '';
  let toastTimeout: ReturnType<typeof setTimeout> | undefined;
  let questionQueue: SnackbarEntry[] = [];
  let pendingQuestion: QuestionEventLike | null = null;
  let pendingTmuxTarget = '';

  onMount(async () => {
    // Intercept console.error/warn/log and forward to Go session log file
    const _error = console.error;
    const _warn = console.warn;
    const _log = console.log;
    console.error = (...args: unknown[]) => { _error(...args); WriteConsoleLog('ERROR', args.map(String).join(' ')).catch(() => {}); };
    console.warn = (...args: unknown[]) => { _warn(...args); WriteConsoleLog('WARN', args.map(String).join(' ')).catch(() => {}); };
    console.log = (...args: unknown[]) => { _log(...args); WriteConsoleLog('LOG', args.map(String).join(' ')).catch(() => {}); };
    // Also catch unhandled errors
    window.addEventListener('error', (e) => WriteConsoleLog('UNCAUGHT', `${e.message} @ ${e.filename}:${e.lineno}`).catch(() => {}));
    window.addEventListener('unhandledrejection', (e) => WriteConsoleLog('UNHANDLED_REJECTION', String(e.reason)).catch(() => {}));

    const cfg = (await GetConfig()) as ConfigSlice;

    try { initMarkdownMenuSettings(cfg.markdownMenu); } catch {}

    // Load themes + fonts before rendering
    try {
      await loadSavedThemes();
      if (cfg.theme) {
        applyTheme(cfg.theme);
      } else if (cfg.importedTheme) {
        await restoreImportedThemeFromConfig(cfg);
      }
      const localFonts = await ListLocalFonts();
      registerLocalFonts(localFonts);
      applyFont(cfg.monoFont || '', cfg.fontSize || 0);
    } catch {}

    // Auto-import bundled themes (after loadSavedThemes so duplicates are skipped)
    try { await loadBundledThemes(); } catch (e) { console.warn('Bundled theme import failed:', e); }

    try { const es = await GetEditorSettings(); initEditorSettings(es); } catch {}

    const dir = await GetDevDir();
    if (dir && dir.length > 0) {
      currentView = 'feed';
      try { notifications = (await GetNotifications()) as AgentNotification[]; } catch {}
    } else {
      currentView = 'setup';
    }

    // Dev-only navigation seam for the Playwright AC spec — lets it
    // jump straight to WorkflowBuilder without driving the feed flow.
    // `import.meta.env.DEV` is a Vite static literal, so the entire
    // block dead-code-eliminates in production.
    if (import.meta.env.DEV && typeof window !== 'undefined') {
      window.__mashed_gotoWorkflows = (repoPath = '') => {
        builderRepoPath = repoPath || '';
        builderRepoBranch = '';
        currentView = 'workflows';
        return true;
      };
    }
  });

  EventsOn('agent:notification', (event: AgentNotification) => {
    // Dedup by agentId, or by tmuxTarget for non-terminal sessions (the frontend
    // spawn uses "spawned-{ts}" as agentId while the backend scanner uses a
    // PID-based ID, but both share the same tmuxTarget for the same session).
    // Don't match terminal sessions — they're only managed by the frontend.
    const idx = notifications.findIndex(n =>
      n.agentId === event.agentId ||
      (!!event.tmuxTarget && n.tmuxTarget === event.tmuxTarget && n.eventType !== 'terminal')
    );
    if (idx >= 0) {
      notifications[idx] = event;
    } else {
      notifications = [event, ...notifications];
    }
    notifications = notifications.sort((a, b) => (a.priority ?? 0) - (b.priority ?? 0));
  });

  EventsOn('agent:removed', (agentId: string) => {
    notifications = notifications.filter(n =>
      n.agentId !== agentId && n.parentAgentId !== agentId
    );
  });

  EventsOn('menu:navigate', (route: string) => {
    showSpawnModal = false;
    showNewRepoModal = false;
    showAboutModal = false;

    switch (route) {
      case 'settings': currentView = 'settings'; break;
      case 'spawn': showSpawnModal = true; break;
      case 'new-repo': showNewRepoModal = true; break;
      case 'feed': currentView = 'feed'; selectedAgent = null; break;
      case 'workflows':
        if (builderRepoPath) currentView = 'workflows';
        break;
    }
  });

  EventsOn('menu:about', () => {
    showAboutModal = true;
  });

  EventsOn('screenshot:taken', (path: string) => {
    const filename = path.split('/').pop();
    toastMessage = `Screenshot saved: ${filename}`;
    clearTimeout(toastTimeout);
    toastTimeout = setTimeout(() => { toastMessage = ''; }, 3000);
  });

  EventsOn('terminal:session:added', (session: TerminalSessionAdded) => {
    // `session` arrives shaped like `domain.TerminalSession`; `addSession`
    // accepts the same shape so the cast is boundary-only.
    addSession(session.repoPath, session as unknown as Parameters<typeof addSession>[1]);
  });

  EventsOn('terminal:session:removed', (sessionName: string) => {
    removeSessionByName(sessionName);
  });

  EventsOn('bmad:node:question', (event: QuestionEventLike) => {
    questionQueue = upsertQuestion(questionQueue, event);
  });

  EventsOn('bmad:node:question:dismissed', (event: DismissedEvent) => {
    const nodeId = typeof event === 'object' && event !== null ? event.nodeId : event;
    if (!nodeId) return;
    questionQueue = dismissQuestion(questionQueue, nodeId);
  });

  EventsOn('bmad:node:idle', (event: IdleEventLike) => {
    questionQueue = upsertIdle(questionQueue, event);
  });

  EventsOn('bmad:node:idle:dismissed', (event: DismissedEvent) => {
    const nodeId = typeof event === 'object' && event !== null ? event.nodeId : event;
    if (!nodeId) return;
    questionQueue = dismissIdle(questionQueue, nodeId);
  });

  // When an interactive node emits a PendingPrompt, the legacy idle snackbar
  // for that node becomes redundant. Clear it so only the amber Respond card
  // remains visible.
  EventsOn('bmad:node:awaiting_input', (event: { nodeId?: string } | string) => {
    const nodeId = typeof event === 'string' ? event : event?.nodeId;
    if (!nodeId) return;
    questionQueue = dismissIdle(questionQueue, nodeId);
    questionQueue = dismissQuestion(questionQueue, nodeId);
  });

  function handleQuestionNavigate(e: CustomEvent<QuestionNavigateDetail>): void {
    const { repoPath, question, entry } = e.detail;
    if (repoPath) builderRepoPath = repoPath;
    pendingQuestion = question ?? null;
    const entryTmux =
      entry && typeof (entry as { tmuxTarget?: unknown }).tmuxTarget === 'string'
        ? ((entry as { tmuxTarget?: string }).tmuxTarget ?? '')
        : '';
    const questionTmux = question?.tmuxTarget ?? '';
    // Idle entries have no question text but carry a tmuxTarget — click
    // should jump to the workflow view and open that node's terminal.
    pendingTmuxTarget = questionTmux || entryTmux || '';
    showSpawnModal = false;
    showNewRepoModal = false;
    showAboutModal = false;
    currentView = 'workflows';
  }

  function onSetupReady(): void {
    currentView = 'feed';
  }

  function addNotification(event: CustomEvent<AgentNotification>): void {
    const n = event.detail;
    const idx = notifications.findIndex(x =>
      x.agentId === n.agentId ||
      (!!n.tmuxTarget && x.tmuxTarget === n.tmuxTarget)
    );
    if (idx >= 0) {
      notifications[idx] = n;
    } else {
      notifications = [...notifications, n];
    }
  }

  function drillDown(agent: AgentNotification | null | undefined): void {
    console.log('drillDown called with:', agent?.agentId, agent?.tmuxTarget);
    selectedAgent = agent ?? null;
    currentView = 'detail';
  }

  function goBack(): void {
    SetActiveContext('', '');
    currentView = 'feed';
    selectedAgent = null;
    pendingQuestion = null;
  }

  function onSpawned(e: CustomEvent<SpawnedDetail>): void {
    showSpawnModal = false;
    // Navigate to the new agent's detail view
    const { target, repo, model } = e.detail;
    selectedAgent = {
      agentId: `spawned-${Date.now()}`,
      agentName: model,
      model: model,
      repoName: repo.name,
      repoPath: repo.path,
      repoBranch: repo.branch,
      eventType: 'running',
      tmuxTarget: target,
      tokensUsed: 0,
      tokensMax: model.includes('opus') ? 1000000 : 200000,
      summary: `Spawned in ${repo.name}`,
    };
    currentView = 'detail';
  }

  function openSettings(): void {
    currentView = 'settings';
  }

  function handleKeydown(e: KeyboardEvent): void {
    if (e.key === 'Escape') {
      if (showAboutModal) {
        showAboutModal = false;
      } else if (showNewRepoModal) {
        showNewRepoModal = false;
      } else if (showSpawnModal) {
        showSpawnModal = false;
      } else if (currentView === 'detail' || currentView === 'settings' || currentView === 'workflows') {
        goBack();
      }
    }
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<main>
  <TitleBar on:open-settings={openSettings} on:open-new-repo={() => showNewRepoModal = true} />
  {#if currentView === 'loading'}
    <div class="loading">
      <div class="loading-icon"><Hexagon size={48} /></div>
    </div>
  {:else if currentView === 'setup'}
    <Setup on:ready={onSetupReady} />
  {:else if currentView === 'feed'}
    <NotificationFeed
      {notifications}
      on:select={(e) => drillDown(e.detail)}
      on:notify={addNotification}
      on:spawn={() => showSpawnModal = true}
      on:open-workspace={(e) => { builderRepoPath = e.detail.path; builderRepoBranch = e.detail.branch || ''; currentView = 'workflows'; }}
    />
  {:else if currentView === 'workflows'}
    <WorkflowBuilder
      repoPath={builderRepoPath}
      repoBranch={builderRepoBranch}
      {pendingQuestion}
      {pendingTmuxTarget}
      on:back={goBack}
      on:question-responded={(e) => {
        const nodeId = e.detail?.nodeId;
        pendingQuestion = null;
        if (nodeId) questionQueue = dismissQuestion(questionQueue, nodeId);
      }}
      on:tmux-opened={() => { pendingTmuxTarget = ''; }}
    />
  {:else if currentView === 'settings'}
    <Settings on:back={goBack} />
  {:else if selectedAgent}
    <!-- Single cast at the boundary: AgentDetail types `agent` as the full
         Wails `NotificationEvent` class, while App tracks the narrower
         `AgentNotification` view-model. At runtime they're the same JSON. -->
    <AgentDetail agent={asWailsNotification(selectedAgent)} on:back={goBack} />
  {:else}
    <!-- Reached when `currentView === 'detail'` but no agent is selected.
         Fall back to the feed so the user is never stranded on a blank view. -->
    <NotificationFeed {notifications} on:select={(e) => drillDown(e.detail)} on:notify={addNotification} on:spawn={() => showSpawnModal = true} on:open-workspace={(e) => { builderRepoPath = e.detail.path; builderRepoBranch = e.detail.branch || ''; currentView = 'workflows'; }} />
  {/if}

  {#if showNewRepoModal}
    <NewRepoModal
      on:created={() => showNewRepoModal = false}
      on:cancel={() => showNewRepoModal = false}
    />
  {/if}

  {#if showSpawnModal}
    <SpawnAgent
      on:spawned={onSpawned}
      on:cancel={() => showSpawnModal = false}
    />
  {/if}

  {#if showAboutModal}
    <AboutModal on:close={() => showAboutModal = false} />
  {/if}

  {#if toastMessage}
    <div class="toast">{toastMessage}</div>
  {/if}

  <QuestionSnackbarStack
    questions={questionQueue}
    on:navigate={handleQuestionNavigate}
    on:dismiss={(e) => {
      const nodeId = e.detail?.entry?.nodeId;
      if (!nodeId) return;
      questionQueue = dismissIdle(questionQueue, nodeId);
      questionQueue = dismissQuestion(questionQueue, nodeId);
    }}
  />
</main>

<style>
  main {
    width: 100vw;
    height: 100vh;
    display: flex;
    flex-direction: column;
    background: var(--bg-deepest);
  }

  .loading {
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    --wails-draggable: drag;
  }

  .loading-icon {
    font-size: 48px;
    color: var(--accent-green);
    animation: pulse 1.5s ease-in-out infinite;
  }

  @keyframes pulse {
    0%, 100% { opacity: 0.3; }
    50% { opacity: 1; }
  }

  .toast {
    position: fixed;
    bottom: 24px;
    right: 24px;
    background: var(--bg-raised);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    padding: 10px 16px;
    font-family: var(--font-mono);
    font-size: 12px;
    color: var(--text-primary);
    z-index: 200;
    animation: toast-in 200ms ease-out;
  }

  @keyframes toast-in {
    from { opacity: 0; transform: translateY(8px); }
    to { opacity: 1; transform: translateY(0); }
  }
</style>
