import { defineStore } from "pinia";
import { useStorage } from "@vueuse/core";

export const useAppShellStore = defineStore("app-shell", () => {
  const sidebarPinned = useStorage("bell-admin-sidebar-pinned", true);
  const mobileMenuOpen = useStorage("bell-admin-mobile-menu-open", false);

  function closeMobileMenu() {
    mobileMenuOpen.value = false;
  }

  function openMobileMenu() {
    mobileMenuOpen.value = true;
  }

  function setSidebarPinned(value: boolean) {
    sidebarPinned.value = value;
  }

  function toggleSidebarPinned() {
    sidebarPinned.value = !sidebarPinned.value;
  }

  return {
    closeMobileMenu,
    mobileMenuOpen,
    openMobileMenu,
    setSidebarPinned,
    sidebarPinned,
    toggleSidebarPinned,
  };
});
