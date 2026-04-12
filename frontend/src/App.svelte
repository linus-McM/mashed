<script>
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
  import { Hexagon } from 'lucide-svelte';
  import TitleBar from './components/TitleBar.svelte';
  import { applyTheme } from './lib/stores/theme.js';
  import { loadSavedThemes, restoreImportedThemeFromConfig, loadBundledThemes } from './lib/themeInit.js';
  import { applyFont, registerLocalFonts } from './lib/stores/font.js';
  import { initEditorSettings } from './lib/stores/editorSettings.js';
  import { addSession, removeSessionByName } from './lib/stores/sessions.js';

  let currentView = 'loading'; // 'loading' | 'setup' | 'feed' | 'detail' | 'settings' | 'workflows'
  let selectedAgent = null;
  let notifications = [];
  let showSpawnModal = false;
  let showNewRepoModal = false;
  let showAboutModal = false;
  let builderRepoPath = '';
  let builderRepoBranch = '';
  let toastMessage = '';
  let toastTimeout;
  let questionQueue = [];
  let pendingQuestion = null;

  onMount(async () => {
    // Intercept console.error/warn/log and forward to Go session log file
    const _error = console.error;
    const _warn = console.warn;
    const _log = console.log;
    console.error = (...args) => { _error(...args); WriteConsoleLog('ERROR', args.map(String).join(' ')).catch(() => {}); };
    console.warn = (...args) => { _warn(...args); WriteConsoleLog('WARN', args.map(String).join(' ')).catch(() => {}); };
    console.log = (...args) => { _log(...args); WriteConsoleLog('LOG', args.map(String).join(' ')).catch(() => {}); };
    // Also catch unhandled errors
    window.addEventListener('error', (e) => WriteConsoleLog('UNCAUGHT', `${e.message} @ ${e.filename}:${e.lineno}`).catch(() => {}));
    window.addEventListener('unhandledrejection', (e) => WriteConsoleLog('UNHANDLED_REJECTION', String(e.reason)).catch(() => {}));

    const cfg = await GetConfig();

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
      try { notifications = await GetNotifications(); } catch {}
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

  EventsOn('agent:notification', (event) => {
    // Dedup by agentId, or by tmuxTarget for non-terminal sessions (the frontend
    // spawn uses "spawned-{ts}" as agentId while the backend scanner uses a
    // PID-based ID, but both share the same tmuxTarget for the same session).
    // Don't match terminal sessions — they're only managed by the frontend.
    const idx = notifications.findIndex(n =>
      n.agentId === event.agentId ||
      (event.tmuxTarget && n.tmuxTarget === event.tmuxTarget && n.eventType !== 'terminal')
    );
    if (idx >= 0) {
      notifications[idx] = event;
    } else {
      notifications = [event, ...notifications];
    }
    notifications = notifications.sort((a, b) => a.priority - b.priority);
  });

  EventsOn('agent:removed', (agentId) => {
    notifications = notifications.filter(n =>
      n.agentId !== agentId && n.parentAgentId !== agentId
    );
  });

  EventsOn('menu:navigate', (route) => {
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

  EventsOn('screenshot:taken', (path) => {
    const filename = path.split('/').pop();
    toastMessage = `Screenshot saved: ${filename}`;
    clearTimeout(toastTimeout);
    toastTimeout = setTimeout(() => { toastMessage = ''; }, 3000);
  });

  EventsOn('terminal:session:added', (session) => {
    addSession(session.repoPath, session);
  });

  EventsOn('terminal:session:removed', (sessionName) => {
    removeSessionByName(sessionName);
  });

  EventsOn('bmad:node:question', (event) => {
    questionQueue = upsertQuestion(questionQueue, event);
  });

  EventsOn('bmad:node:question:dismissed', (event) => {
    const nodeId = event && typeof event === 'object' ? event.nodeId : event;
    if (!nodeId) return;
    questionQueue = dismissQuestion(questionQueue, nodeId);
  });

  EventsOn('bmad:node:idle', (event) => {
    questionQueue = upsertIdle(questionQueue, event);
  });

  EventsOn('bmad:node:idle:dismissed', (event) => {
    const nodeId = event && typeof event === 'object' ? event.nodeId : event;
    if (!nodeId) return;
    questionQueue = dismissIdle(questionQueue, nodeId);
  });

  function handleQuestionNavigate(e) {
    const { repoPath, question } = e.detail;
    if (repoPath) builderRepoPath = repoPath;
    pendingQuestion = question;
    showSpawnModal = false;
    showNewRepoModal = false;
    showAboutModal = false;
    currentView = 'workflows';
  }

  function onSetupReady() {
    currentView = 'feed';
  }

  function addNotification(event) {
    const n = event.detail;
    const idx = notifications.findIndex(x =>
      x.agentId === n.agentId ||
      (n.tmuxTarget && x.tmuxTarget === n.tmuxTarget)
    );
    if (idx >= 0) {
      notifications[idx] = n;
    } else {
      notifications = [...notifications, n];
    }
  }

  function drillDown(agent) {
    console.log('drillDown called with:', agent?.agentId, agent?.tmuxTarget);
    selectedAgent = agent;
    currentView = 'detail';
  }

  function goBack() {
    SetActiveContext('', '');
    currentView = 'feed';
    selectedAgent = null;
    pendingQuestion = null;
  }

  function onSpawned(e) {
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

  function openSettings() {
    currentView = 'settings';
  }

  function handleKeydown(e) {
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
      on:back={goBack}
      on:question-responded={(e) => {
        const nodeId = e.detail?.nodeId;
        pendingQuestion = null;
        if (nodeId) questionQueue = dismissQuestion(questionQueue, nodeId);
      }}
    />
  {:else if currentView === 'settings'}
    <Settings on:back={goBack} />
  {:else}
    <AgentDetail agent={selectedAgent} on:back={goBack} />
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

  <QuestionSnackbarStack questions={questionQueue} on:navigate={handleQuestionNavigate} />
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
