import { computed, ref } from "vue";
import { defineStore } from "pinia";

export type ToastTone = "error" | "info" | "success" | "warning";

export type ToastItem = {
  detail: string;
  id: number;
  timeoutMs: number;
  title: string;
  tone: ToastTone;
};

const DEFAULT_TIMEOUT_MS = 10_000;

export const useToastStore = defineStore("toast", () => {
  const items = ref<ToastItem[]>([]);
  let nextToastId = 1;
  const timers = new Map<number, ReturnType<typeof setTimeout>>();

  const hasToasts = computed(() => items.value.length > 0);

  function enqueue(
    toast: Omit<ToastItem, "id" | "timeoutMs"> & {
      timeoutMs?: number;
    },
  ) {
    const id = nextToastId++;
    const item: ToastItem = {
      ...toast,
      id,
      timeoutMs: toast.timeoutMs ?? DEFAULT_TIMEOUT_MS,
    };

    items.value.push(item);

    const timer = setTimeout(() => {
      dismiss(id);
    }, item.timeoutMs);

    timers.set(id, timer);

    return id;
  }

  function dismiss(id: number) {
    items.value = items.value.filter((item) => item.id !== id);

    const timer = timers.get(id);

    if (timer) {
      clearTimeout(timer);
      timers.delete(id);
    }
  }

  function clearAll() {
    items.value.forEach((item) => dismiss(item.id));
  }

  return {
    clearAll,
    dismiss,
    enqueue,
    hasToasts,
    items,
  };
});
