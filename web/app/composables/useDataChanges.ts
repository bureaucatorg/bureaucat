type Listener = () => void;

const listeners = new Set<Listener>();
let timer: ReturnType<typeof setTimeout> | undefined;

// Coalesces bursts (e.g. the assistant making several writes in a row) into one refresh.
export function notifyDataChanged() {
  clearTimeout(timer);
  timer = setTimeout(() => listeners.forEach((fn) => fn()), 300);
}

export function onDataChanged(fn: Listener) {
  onMounted(() => listeners.add(fn));
  onUnmounted(() => listeners.delete(fn));
}
