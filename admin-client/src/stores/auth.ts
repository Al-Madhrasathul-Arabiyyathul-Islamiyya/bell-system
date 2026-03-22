import { computed } from "vue";
import { defineStore } from "pinia";
import { useStorage } from "@vueuse/core";

export const useAuthStore = defineStore("auth", () => {
  const token = useStorage<string | null>("bell-admin-token", null);
  const username = useStorage<string | null>("bell-admin-username", null);

  const isAuthenticated = computed(() => Boolean(token.value));

  function loginPlaceholder(nextUsername: string) {
    token.value = "placeholder-session-token";
    username.value = nextUsername.trim() || "admin";
  }

  function logout() {
    token.value = null;
    username.value = null;
  }

  return {
    isAuthenticated,
    loginPlaceholder,
    logout,
    token,
    username,
  };
});
