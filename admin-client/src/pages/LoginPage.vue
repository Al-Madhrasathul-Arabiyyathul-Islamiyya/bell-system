<script setup lang="ts">
import { computed, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useTitle } from "@vueuse/core";
import { appEnv } from "../lib/env";
import { useAuthStore } from "../stores/auth";

const route = useRoute();
const router = useRouter();
const authStore = useAuthStore();

const username = ref("");
const password = ref("");

useTitle(`Login • ${appEnv.appName}`);

const redirectTarget = computed(() => {
  return typeof route.query.redirect === "string"
    ? route.query.redirect
    : "/dashboard";
});

function handlePlaceholderLogin() {
  authStore.loginPlaceholder(username.value);
  router.push(redirectTarget.value);
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
          This page is still on a placeholder session flow so the app shell can
          be exercised before the auth API and JSON:API transport layer are
          implemented.
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
            <h2 class="card-title">Temporary Access</h2>
            <label class="form-control gap-2">
              <span class="label-text font-medium">Username</span>
              <input
                v-model="username"
                class="input input-bordered w-full"
                placeholder="admin"
                type="text"
              />
            </label>
            <label class="form-control gap-2">
              <span class="label-text font-medium">Password</span>
              <input
                v-model="password"
                class="input input-bordered w-full"
                placeholder="Pending API integration"
                type="password"
              />
            </label>
            <button
              class="btn btn-primary mt-2"
              type="button"
              @click="handlePlaceholderLogin"
            >
              Enter Placeholder Session
            </button>
            <p class="text-xs leading-6 text-base-content/60">
              This only persists a local placeholder token for shell
              development. Real login will replace it.
            </p>
          </div>
        </article>
      </div>
    </section>
  </main>
</template>
