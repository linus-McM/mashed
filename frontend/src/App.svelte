<script>
  import { onMount } from 'svelte';
  import { EventsOn } from '../wailsjs/runtime/runtime.js';
  import { GetNotifications, GetDevDir, GetConfig, ListLocalFonts } from '../wailsjs/go/main/App.js';
  import Setup from './views/Setup.svelte';
  import NotificationFeed from './views/NotificationFeed.svelte';
  import AgentDetail from './views/AgentDetail.svelte';
  import SpawnAgent from './views/SpawnAgent.svelte';
  import Settings from './views/Settings.svelte';
  import WorkflowBuilder from './views/WorkflowBuilder.svelte';
  import RepoPickerModal from './components/bmad/RepoPickerModal.svelte';
  import { Hexagon } from 'lucide-svelte';
  import TitleBar from './components/TitleBar.svelte';
  import { applyTheme } from './lib/stores/theme.js';
  import { loadSavedThemes, restoreImportedThemeFromConfig } from './lib/themeInit.js';
  import { applyFont, registerLocalFonts } from './lib/stores/font.js';

  let currentView = 'loading'; // 'loading' | 'setup' | 'feed' | 'detail' | 'settings' | 'workflows'
  let selectedAgent = null;
  let notifications = [];
  let showSpawnModal = false;
  let showRepoPicker = false;
  let builderRepoPath = '';

  onMount(async () => {
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

    const dir = await GetDevDir();
    if (dir && dir.length > 0) {
      currentView = 'feed';
      try { notifications = await GetNotifications(); } catch {}
    } else {
      currentView = 'setup';
    }
  });

  EventsOn('agent:notification', (event) => {
    const idx = notifications.findIndex(n => n.agentId === event.agentId);
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

  function onSetupReady() {
    currentView = 'feed';
  }

  function addNotification(event) {
    const n = event.detail;
    const idx = notifications.findIndex(x => x.agentId === n.agentId);
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
    currentView = 'feed';
    selectedAgent = null;
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

  function openWorkflows() {
    showRepoPicker = true;
  }

  function onRepoSelected(e) {
    builderRepoPath = e.detail.path;
    showRepoPicker = false;
    currentView = 'workflows';
  }

  function handleKeydown(e) {
    if (e.key === 'Escape') {
      if (showRepoPicker) {
        showRepoPicker = false;
      } else if (showSpawnModal) {
        showSpawnModal = false;
      } else if (currentView === 'detail' || currentView === 'settings' || currentView === 'workflows') {
        goBack();
      }
    }
    // Ctrl+N or Cmd+N to spawn
    if ((e.ctrlKey || e.metaKey) && e.key === 'n' && currentView === 'feed') {
      e.preventDefault();
      showSpawnModal = true;
    }
    // Ctrl+W or Cmd+W to open workflows
    if ((e.ctrlKey || e.metaKey) && e.key === 'w' && currentView === 'feed') {
      e.preventDefault();
      openWorkflows();
    }
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<main>
  <TitleBar on:open-settings={openSettings} on:open-workflows={openWorkflows} />
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
    />
  {:else if currentView === 'workflows'}
    <WorkflowBuilder repoPath={builderRepoPath} on:back={goBack} />
  {:else if currentView === 'settings'}
    <Settings on:back={goBack} />
  {:else}
    <AgentDetail agent={selectedAgent} on:back={goBack} />
  {/if}

  {#if showRepoPicker}
    <RepoPickerModal
      on:select={onRepoSelected}
      on:cancel={() => showRepoPicker = false}
    />
  {/if}

  {#if showSpawnModal}
    <SpawnAgent
      on:spawned={onSpawned}
      on:cancel={() => showSpawnModal = false}
    />
  {/if}
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
</style>
