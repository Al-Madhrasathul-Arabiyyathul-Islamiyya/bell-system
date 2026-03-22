import { computed } from "vue";
import { defineStore } from "pinia";
import { useStorage } from "@vueuse/core";
import type { AuthUser } from "../lib/api/types";

export const useAuthStore = defineStore("auth", () => {
  const token = useStorage<string | null>("bell-admin-token", null);
  const user = useStorage<AuthUser>("bell-admin-user", {
    id: null,
    role: null,
    username: null,
  });

  const isAuthenticated = computed(() => Boolean(token.value));
  const isAdmin = computed(() => user.value.role === "admin");
  const username = computed(() => user.value.username);

  function loginPlaceholder(nextUsername: string) {
    token.value = "placeholder-session-token";
    user.value = {
      id: null,
      role: "admin",
      username: nextUsername.trim() || "admin",
    };
  }

  function setAuth(nextToken: string, nextUser: Partial<AuthUser>) {
    token.value = nextToken;
    user.value = {
      id: nextUser.id ?? null,
      role: nextUser.role ?? null,
      username: nextUser.username ?? null,
    };
  }

  function clearAuth() {
    token.value = null;
    user.value = {
      id: null,
      role: null,
      username: null,
    };
  }

  return {
    clearAuth,
    isAuthenticated,
    isAdmin,
    loginPlaceholder,
    logout: clearAuth,
    setAuth,
    token,
    user,
    username,
  };
});
