import { computed } from "vue";
import { defineStore } from "pinia";
import { useColorMode } from "@vueuse/core";

type AppTheme = "light" | "dark";

export const usePreferencesStore = defineStore("preferences", () => {
  const colorMode = useColorMode<AppTheme>({
    attribute: "data-theme",
    emitAuto: false,
    initialValue: "light",
    modes: {
      dark: "dark",
      light: "light",
    },
  });

  const theme = computed<AppTheme>({
    get: () => (colorMode.value === "dark" ? "dark" : "light"),
    set: (value) => {
      colorMode.value = value;
    },
  });

  function toggleTheme() {
    theme.value = theme.value === "light" ? "dark" : "light";
  }

  return {
    theme,
    toggleTheme,
  };
});
