import { ref, onMounted, onBeforeUnmount } from "vue";

export function useDropdown() {
  const open = ref(false);
  const triggerRef = ref<HTMLElement | null>(null);
  const dropdownRef = ref<HTMLElement | null>(null);

  function toggle() {
    open.value = !open.value;
  }

  function close() {
    open.value = false;
  }

  function handleClickOutside(event: MouseEvent) {
    const target = event.target as Node;

    if (
      open.value &&
      triggerRef.value &&
      dropdownRef.value &&
      !triggerRef.value.contains(target) &&
      !dropdownRef.value.contains(target)
    ) {
      close();
    }
  }

  onMounted(() => {
    document.addEventListener("click", handleClickOutside);
  });

  onBeforeUnmount(() => {
    document.removeEventListener("click", handleClickOutside);
  });

  return {
    open,
    toggle,
    close,
    triggerRef,
    dropdownRef,
  };
}
