<script>
  import { onMount } from 'svelte';
  import { EventsOn } from '../wailsjs/runtime/runtime.js';
  import { GetNotifications, GetDevDir } from '../wailsjs/go/main/App.js';
  import Setup from './views/Setup.svelte';
  import NotificationFeed from './views/NotificationFeed.svelte';
  import AgentDetail from './views/AgentDetail.svelte';

  let currentView = 'loading'; // 'loading' | 'setup' | 'feed' | 'detail'
  let selectedAgent = null;
  let notifications = [];

  onMount(async () => {
    try {
      const dir = await GetDevDir();
      console.log('GetDevDir returned:', JSON.stringify(dir));
      if (dir && dir.length > 0) {
        currentView = 'feed';
        notifications = await GetNotifications();
      } else {
        currentView = 'setup';
      }
    } catch (e) {
      console.error('GetDevDir failed:', e);
      currentView = 'setup';
    }
  });

  // Live notification updates
  EventsOn('agent:notification', (event) => {
    const idx = notifications.findIndex(n => n.agentId === event.agentId);
    if (idx >= 0) {
      notifications[idx] = event;
    } else {
      notifications = [event, ...notifications];
    }
    notifications = notifications.sort((a, b) => a.priority - b.priority);
  });

  function onSetupReady() {
    currentView = 'feed';
  }

  function drillDown(agent) {
    selectedAgent = agent;
    currentView = 'detail';
  }

  function goBack() {
    currentView = 'feed';
    selectedAgent = null;
  }

  function handleKeydown(e) {
    if (e.key === 'Escape' && currentView === 'detail') goBack();
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<main>
  {#if currentView === 'loading'}
    <div class="loading">
      <div class="loading-icon">⬡</div>
    </div>
  {:else if currentView === 'setup'}
    <Setup on:ready={onSetupReady} />
  {:else if currentView === 'feed'}
    <NotificationFeed {notifications} on:select={(e) => drillDown(e.detail)} />
  {:else}
    <AgentDetail agent={selectedAgent} on:back={goBack} />
  {/if}
</main>

<style>
  main {
    width: 100vw;
    height: 100vh;
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
