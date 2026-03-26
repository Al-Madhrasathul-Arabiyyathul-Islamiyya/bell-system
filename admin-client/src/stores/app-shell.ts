import { defineStore } from "pinia";
import { useStorage } from "@vueuse/core";

export const useAppShellStore = defineStore("app-shell", () => {
  const sidebarCollapsed = useStorage("bell-admin-sidebar-collapsed", false);
  const mobileMenuOpen = useStorage("bell-admin-mobile-menu-open", false);

  function closeMobileMenu() {
    mobileMenuOpen.value = false;
  }

  function openMobileMenu() {
    mobileMenuOpen.value = true;
  }

  function setSidebarCollapsed(value: boolean) {
    sidebarCollapsed.value = value;
  }

  function toggleSidebarCollapsed() {
    sidebarCollapsed.value = !sidebarCollapsed.value;
  }

  return {
    closeMobileMenu,
    mobileMenuOpen,
    openMobileMenu,
    setSidebarCollapsed,
    sidebarCollapsed,
    toggleSidebarCollapsed,
  };
});
