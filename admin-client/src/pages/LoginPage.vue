<script setup lang="ts">
import { computed, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useTitle } from "@vueuse/core";
import { Icon } from "@iconify/vue";
import { getPrimaryError } from "../lib/api/errors";
import { appEnv } from "../lib/env";
import { useLoginMutation } from "../features/auth/composables/use-auth";
import { useAuthStore } from "../stores/auth";
import { usePreferencesStore } from "../stores/preferences";

const route = useRoute();
const router = useRouter();
const authStore = useAuthStore();
const preferencesStore = usePreferencesStore();
const loginMutation = useLoginMutation();

const email = ref("");
const password = ref("");

useTitle(`Login • ${appEnv.appName}`);

const redirectTarget = computed(() => {
  return typeof route.query.redirect === "string"
    ? route.query.redirect
    : "/dashboard";
});

const routeError = computed(() => {
  if (route.query.error === "admin-only") {
    return "This admin client only allows administrator accounts.";
  }

  return null;
});

const loginError = computed(() => {
  if (!loginMutation.error.value) {
    return null;
  }

  return getPrimaryError(loginMutation.error.value).detail;
});

async function handleLogin() {
  await loginMutation.mutateAsync({
    password: password.value,
    username: email.value,
  });

  if (!authStore.isAdmin) {
    authStore.clearAuth();
    return;
  }

  await router.push(redirectTarget.value);
}
</script>

<template>
  <main class="login-shell">
    <div class="mx-auto flex w-full max-w-6xl justify-end px-2">
      <div class="dropdown dropdown-end">
        <button class="btn btn-ghost btn-circle" type="button">
          <Icon
            :icon="
              preferencesStore.themePreference === 'system'
                ? 'solar:monitor-bold-duotone'
                : preferencesStore.resolvedTheme === 'dark'
                  ? 'solar:moon-stars-bold-duotone'
                  : 'solar:sun-bold-duotone'
            "
            class="text-xl"
          />
        </button>
        <ul
          class="menu dropdown-content z-10 mt-3 w-40 rounded-box border border-base-300 bg-base-100 p-2 shadow-xl"
        >
          <li>
            <button
              type="button"
              @click="preferencesStore.themePreference = 'system'"
            >
              System
            </button>
          </li>
          <li>
            <button
              type="button"
              @click="preferencesStore.themePreference = 'light'"
            >
              Light
            </button>
          </li>
          <li>
            <button
              type="button"
              @click="preferencesStore.themePreference = 'dark'"
            >
              Dark
            </button>
          </li>
        </ul>
      </div>
    </div>

    <section class="login-panel pt-4">
      <article
        class="card mx-auto w-full max-w-md border border-base-300 bg-base-100 shadow-xl"
      >
        <div class="card-body gap-5 p-8">
          <div class="flex flex-col items-center gap-4 text-center">
            <img
              class="size-20 rounded-box bg-primary/10 p-3"
              src="/logo.svg"
              alt="Arabiyya Bell System logo"
            />
            <h1
              class="font-display text-4xl font-semibold tracking-tight text-base-content"
            >
              Arabiyya Bell System
            </h1>
          </div>

          <div class="space-y-4">
            <div
              v-if="routeError || loginError"
              class="alert alert-error text-sm"
              role="alert"
            >
              {{ routeError || loginError }}
            </div>

            <label class="form-control gap-2">
              <span class="label-text font-medium">Email</span>
              <input
                v-model="email"
                class="input input-bordered w-full"
                :disabled="loginMutation.isPending.value"
                placeholder="admin@example.com"
                type="email"
              />
            </label>
            <label class="form-control gap-2">
              <span class="label-text font-medium">Password</span>
              <input
                v-model="password"
                class="input input-bordered w-full"
                :disabled="loginMutation.isPending.value"
                placeholder="Password"
                type="password"
                @keydown.enter="handleLogin"
              />
            </label>
            <button
              class="btn btn-primary mt-2"
              type="button"
              :class="{ 'btn-disabled': loginMutation.isPending.value }"
              :disabled="loginMutation.isPending.value"
              @click="handleLogin"
            >
              {{ loginMutation.isPending.value ? "Logging In..." : "Login" }}
            </button>
          </div>
        </div>
      </article>
    </section>
  </main>
</template>
