<script>
  import { createEventDispatcher, onMount } from 'svelte';
  import { X, Plus, Minus } from 'lucide-svelte';

  /** @type {string[]} */
  export let items = [];

  /** @type {import('svelte').EventDispatcher<{ save: string[]; close: void }>} */
  const dispatch = createEventDispatcher();

  /** @type {string[]} */
  let localItems = [];

  onMount(() => {
    localItems = items.length > 0 ? [...items] : [''];
  });

  function addItem() {
    localItems = [...localItems, ''];
    // Focus the new input after render
    setTimeout(() => {
      const inputs = /** @type {NodeListOf<HTMLInputElement>} */ (document.querySelectorAll('.item-input'));
      if (inputs.length > 0) inputs[inputs.length - 1].focus();
    }, 0);
  }

  /** @param {number} index */
  function removeItem(index) {
    localItems = localItems.filter((_, i) => i !== index);
    if (localItems.length === 0) localItems = [''];
  }

  /**
   * @param {number} index
   * @param {string} value
   */
  function updateItem(index, value) {
    localItems[index] = value;
    localItems = localItems;
  }

  function save() {
    // Filter out empty strings
    const filtered = localItems.filter(s => s.trim() !== '');
    dispatch('save', filtered);
  }

  function close() {
    dispatch('close');
  }

  /** @param {KeyboardEvent} e */
  function handleKeydown(e) {
    if (e.key === 'Escape') close();
  }

  /**
   * @param {KeyboardEvent} e
   * @param {number} _index
   */
  function handleItemKeydown(e, _index) {
    if (e.key === 'Enter') {
      e.preventDefault();
      addItem();
    }
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<!-- svelte-ignore a11y-click-events-have-key-events -->
<!-- svelte-ignore a11y-no-static-element-interactions -->
<div class="overlay" on:click={close}>
  <!-- svelte-ignore a11y-click-events-have-key-events -->
  <!-- svelte-ignore a11y-no-static-element-interactions -->
  <div class="modal" on:click|stopPropagation>
    <div class="modal-header">
      <span class="modal-title">Edit Array Items</span>
      <button class="close-btn" on:click={close}>
        <X size={14} />
      </button>
    </div>

    <div class="modal-body">
      <div class="items-list">
        {#each localItems as item, i}
          <div class="item-row">
            <span class="item-index">{i + 1}</span>
            <input
              class="item-input"
              type="text"
              value={item}
              on:input={(e) => updateItem(i, e.target.value)}
              on:keydown={(e) => handleItemKeydown(e, i)}
              placeholder="Enter item..."
            />
            <button class="remove-btn" on:click={() => removeItem(i)} title="Remove item">
              <Minus size={12} />
            </button>
          </div>
        {/each}
      </div>

      <button class="add-btn" on:click={addItem}>
        <Plus size={13} />
        Add Item
      </button>
    </div>

    <div class="modal-footer">
      <span class="item-count">{localItems.filter(s => s.trim() !== '').length} items</span>
      <div class="footer-buttons">
        <button class="btn cancel-btn" on:click={close}>Cancel</button>
        <button class="btn save-btn" on:click={save}>Save</button>
      </div>
    </div>
  </div>
</div>

<style>
  .overlay {
    position: fixed;
    inset: 0;
    background: var(--overlay-backdrop);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 100;
  }

  .modal {
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md, 6px);
    width: 400px;
    max-height: 80vh;
    display: flex;
    flex-direction: column;
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4);
  }

  .modal-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 14px;
    border-bottom: 1px solid var(--border-subtle);
    flex-shrink: 0;
  }

  .modal-title {
    font-family: var(--font-mono);
    font-size: 13px;
    font-weight: 600;
    color: var(--text-primary);
  }

  .close-btn {
    background: none;
    border: none;
    color: var(--text-muted);
    cursor: pointer;
    padding: 2px;
    border-radius: var(--radius-sm);
    display: flex;
    align-items: center;
    transition: color 100ms ease;
  }

  .close-btn:hover { color: var(--text-primary); }

  .modal-body {
    flex: 1;
    overflow-y: auto;
    padding: 12px 14px;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .items-list {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .item-row {
    display: flex;
    align-items: center;
    gap: var(--sp-sm);
  }

  .item-index {
    font-family: var(--font-mono);
    font-size: 9px;
    font-weight: 600;
    color: var(--text-muted);
    min-width: 18px;
    text-align: right;
    flex-shrink: 0;
  }

  .item-input {
    flex: 1;
    padding: var(--sp-xs) var(--sp-sm);
    background: var(--bg-deepest);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-sm);
    color: var(--text-primary);
    font-family: var(--font-mono);
    font-size: 11px;
    outline: none;
    transition: border-color 100ms ease;
  }

  .item-input:focus { border-color: var(--accent-green); }
  .item-input::placeholder { color: var(--text-muted); }

  .remove-btn {
    background: none;
    border: 1px solid transparent;
    color: var(--text-muted);
    cursor: pointer;
    padding: 4px;
    border-radius: var(--radius-sm);
    display: flex;
    align-items: center;
    flex-shrink: 0;
    transition: color 100ms ease, border-color 100ms ease;
  }

  .remove-btn:hover {
    color: var(--accent-red, #f85149);
    border-color: var(--accent-red, #f85149);
  }

  .add-btn {
    display: flex;
    align-items: center;
    gap: var(--sp-sm);
    width: 100%;
    padding: var(--sp-xs) 10px;
    background: var(--bg-elevated);
    border: 1px dashed var(--border-emphasis);
    border-radius: var(--radius-sm);
    color: var(--accent-green);
    font-family: var(--font-mono);
    font-size: 11px;
    font-weight: 500;
    cursor: pointer;
    transition: background 100ms ease, border-color 100ms ease;
  }

  .add-btn:hover {
    background: var(--bg-active);
    border-color: var(--accent-green);
  }

  .modal-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 10px 14px;
    border-top: 1px solid var(--border-subtle);
    flex-shrink: 0;
  }

  .item-count {
    font-family: var(--font-mono);
    font-size: var(--text-label);
    color: var(--text-muted);
  }

  .footer-buttons {
    display: flex;
    gap: 8px;
  }

  .btn {
    padding: var(--sp-xs) 14px;
    border-radius: var(--radius-sm);
    font-family: var(--font-mono);
    font-size: 11px;
    font-weight: 500;
    cursor: pointer;
    border: 1px solid var(--border-subtle);
    transition: background 100ms ease, border-color 100ms ease;
  }

  .cancel-btn {
    background: var(--bg-elevated);
    color: var(--text-primary);
  }

  .cancel-btn:hover { background: var(--bg-active); }

  .save-btn {
    background: var(--accent-green);
    color: var(--bg-deepest);
    border-color: var(--accent-green);
    font-weight: 600;
  }

  .save-btn:hover {
    opacity: 0.9;
  }
</style>
