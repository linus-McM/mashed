<script>
  import { Quit, WindowMinimise, WindowToggleMaximise } from '../../wailsjs/runtime/runtime.js';
</script>

<div class="titlebar">
  <div class="traffic-lights">
    <button class="tl-btn close" on:click={Quit} title="Close">
      <svg viewBox="0 0 12 12"><path d="M3.5 3.5l5 5M8.5 3.5l-5 5" stroke="currentColor" stroke-width="1.2" fill="none"/></svg>
    </button>
    <button class="tl-btn minimize" on:click={WindowMinimise} title="Minimize">
      <svg viewBox="0 0 12 12"><path d="M2.5 6h7" stroke="currentColor" stroke-width="1.2" fill="none"/></svg>
    </button>
    <button class="tl-btn fullscreen" on:click={WindowToggleMaximise} title="Fullscreen">
      <svg viewBox="0 0 12 12"><path d="M3 8.5L6 4l3 4.5H3z" fill="currentColor"/></svg>
    </button>
  </div>
  <div class="drag-region" on:dblclick={WindowToggleMaximise}></div>
</div>

<style>
  .titlebar {
    display: flex;
    align-items: center;
    height: 38px;
    padding: 0 12px;
    background: var(--bg-surface);
    border-bottom: 1px solid var(--border-subtle);
    flex-shrink: 0;
    --wails-draggable: drag;
    user-select: none;
    -webkit-user-select: none;
  }

  .drag-region {
    flex: 1;
    height: 100%;
    --wails-draggable: drag;
  }

  .traffic-lights {
    display: flex;
    gap: 8px;
    align-items: center;
    --wails-draggable: none;
  }

  .tl-btn {
    width: 12px;
    height: 12px;
    border-radius: 50%;
    border: none;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 0;
    transition: filter 120ms ease;
  }

  .tl-btn svg {
    width: 8px;
    height: 8px;
    color: transparent;
    transition: color 120ms ease;
  }

  .tl-btn.close { background: #ff5f57; }
  .tl-btn.minimize { background: #febc2e; }
  .tl-btn.fullscreen { background: #28c840; }

  .tl-btn:hover svg { color: rgba(0, 0, 0, 0.5); }
  .tl-btn:active { filter: brightness(0.8); }

  .traffic-lights:hover .tl-btn svg { color: rgba(0, 0, 0, 0.5); }
</style>
