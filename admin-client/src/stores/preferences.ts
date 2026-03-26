import { computed, watch } from "vue";
import { defineStore } from "pinia";
import { usePreferredDark, useStorage } from "@vueuse/core";

type ThemePreference = "system" | "light" | "dark";
type ResolvedTheme = "light" | "dark";

export const usePreferencesStore = defineStore("preferences", () => {
  const preferredDark = usePreferredDark();
  const themePreference = useStorage<ThemePreference>(
    "bell-system-theme",
    "system",
  );

  const resolvedTheme = computed<ResolvedTheme>(() => {
    if (themePreference.value === "system") {
      return preferredDark.value ? "dark" : "light";
    }

    return themePreference.value;
  });

  watch(
    resolvedTheme,
    (theme) => {
      document.documentElement.setAttribute("data-theme", theme);
    },
    { immediate: true },
  );

  return {
    resolvedTheme,
    themePreference,
  };
});
