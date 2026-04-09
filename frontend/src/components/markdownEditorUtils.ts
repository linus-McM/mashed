export type SaveStatus = 'idle' | 'modified' | 'saving' | 'saved' | 'error';

export interface DebouncedSave {
  schedule(): void;
  flush(): void;
  cancel(): void;
  destroy(): void;
}

export function createDebouncedSave(
  saveFn: () => void | Promise<void>,
  delayMs: number,
): DebouncedSave {
  let timer: ReturnType<typeof setTimeout> | null = null;
  let destroyed = false;

  function cancel() {
    if (timer !== null) {
      clearTimeout(timer);
      timer = null;
    }
  }

  return {
    schedule() {
      if (destroyed) return;
      cancel();
      timer = setTimeout(() => {
        timer = null;
        saveFn();
      }, delayMs);
    },
    flush() {
      if (destroyed) return;
      cancel();
      saveFn();
    },
    cancel,
    destroy() {
      destroyed = true;
      cancel();
    },
  };
}
