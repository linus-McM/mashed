<script>
  import { EventsOn } from '../wailsjs/runtime/runtime.js';
  import NotificationFeed from './views/NotificationFeed.svelte';
  import AgentDetail from './views/AgentDetail.svelte';

  let currentView = 'feed'; // 'feed' or 'detail'
  let selectedAgent = null;
  let notifications = [];

  EventsOn('notification', (event) => {
    // Update or add notification
    const idx = notifications.findIndex(n => n.agentId === event.agentId);
    if (idx >= 0) {
      notifications[idx] = event;
    } else {
      notifications = [event, ...notifications];
    }
    notifications = notifications.sort((a, b) => a.priority - b.priority);
  });

  function drillDown(agent) {
    selectedAgent = agent;
    currentView = 'detail';
  }

  function goBack() {
    currentView = 'feed';
    selectedAgent = null;
  }

  // Keyboard navigation
  function handleKeydown(e) {
    if (e.key === 'Escape' && currentView === 'detail') goBack();
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<main>
  {#if currentView === 'feed'}
    <NotificationFeed {notifications} on:select={(e) => drillDown(e.detail)} />
  {:else}
    <AgentDetail agent={selectedAgent} on:back={goBack} />
  {/if}
</main>

<style>
  main { width: 100vw; height: 100vh; background: var(--bg-deepest); }
</style>
