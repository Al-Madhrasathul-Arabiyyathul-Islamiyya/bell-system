<script setup lang="ts">
import { computed, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useTitle } from "@vueuse/core";
import { getPrimaryError } from "../lib/api/errors";
import { appEnv } from "../lib/env";
import { useLoginMutation } from "../features/auth/composables/use-auth";
import { useAuthStore } from "../stores/auth";

const route = useRoute();
const router = useRouter();
const authStore = useAuthStore();
const loginMutation = useLoginMutation();

const username = ref("");
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
    username: username.value,
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
    <section class="login-panel">
      <div class="space-y-4">
        <p
          class="text-xs font-semibold uppercase tracking-[0.24em] text-primary"
        >
          Bell System
        </p>
        <h1
          class="font-display text-4xl font-semibold tracking-tight text-base-content"
        >
          Admin client foundation is wired.
        </h1>
        <p
          class="max-w-2xl text-sm leading-7 text-base-content/70 md:text-base"
        >
          Sign in with a live backend account. Auth state is persisted locally,
          so valid sessions survive hard refreshes until logout or a `401`
          response clears them.
        </p>
      </div>

      <div class="grid gap-6 lg:grid-cols-[minmax(0,1fr)_24rem]">
        <article class="card border border-base-300 bg-base-100 shadow-sm">
          <div class="card-body gap-4">
            <h2 class="card-title">What is already wired</h2>
            <ul class="space-y-3 text-sm leading-6 text-base-content/70">
              <li>
                Vue Router with guarded protected routes and public login
                screen.
              </li>
              <li>
                Pinia stores for auth state and persisted theme preference.
              </li>
              <li>
                TanStack Vue Query bootstrapped for server-state features.
              </li>
              <li>
                Tailwind CSS v4 and DaisyUI theme tokens integrated for the Bell
                System palette.
              </li>
            </ul>
          </div>
        </article>

        <article class="card border border-base-300 bg-base-100 shadow-sm">
          <div class="card-body gap-4">
            <h2 class="card-title">Admin Sign In</h2>
            <div
              v-if="routeError || loginError"
              class="alert alert-error text-sm"
              role="alert"
            >
              {{ routeError || loginError }}
            </div>
            <label class="form-control gap-2">
              <span class="label-text font-medium">Username</span>
              <input
                v-model="username"
                class="input input-bordered w-full"
                :disabled="loginMutation.isPending.value"
                placeholder="admin"
                type="text"
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
              {{ loginMutation.isPending.value ? "Signing In..." : "Sign In" }}
            </button>
            <p class="text-xs leading-6 text-base-content/60">
              Only administrator accounts should continue into this client.
            </p>
          </div>
        </article>
      </div>
    </section>
  </main>
</template>
