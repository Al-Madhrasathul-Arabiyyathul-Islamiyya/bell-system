<script setup lang="ts">
import { computed, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useTitle } from "@vueuse/core";
import AppThemeSwitcher from "../components/app/AppThemeSwitcher.vue";
import { getUserFacingError } from "../lib/api/errors";
import { appEnv } from "../lib/env";
import { useLoginMutation } from "../features/auth/composables/use-auth";

const route = useRoute();
const router = useRouter();
const loginMutation = useLoginMutation();

const username = ref("");
const password = ref("");

useTitle(`Login • ${appEnv.appName}`);

const redirectTarget = computed(() => {
  return typeof route.query.redirect === "string"
    ? route.query.redirect
    : "/dashboard";
});
const sessionExpired = computed(() => route.query.reason === "session-expired");

const loginError = computed(() => {
  if (!loginMutation.error.value) {
    return null;
  }

  return getUserFacingError(loginMutation.error.value).detail;
});

async function handleLogin() {
  await loginMutation.mutateAsync({
    password: password.value,
    username: username.value,
  });

  await router.push(redirectTarget.value);
}
</script>

<template>
  <main class="login-shell">
    <div class="mx-auto flex w-full max-w-6xl justify-end px-2">
      <AppThemeSwitcher icon-only />
    </div>

    <section class="login-panel pt-4">
      <article class="card mx-auto w-full max-w-md bg-base-100 shadow-2xl">
        <div class="card-body gap-6">
          <div class="flex flex-col items-center gap-4 text-center">
            <figure class="avatar">
              <div class="w-20 rounded-box bg-primary/10 p-3">
                <img src="/logo.svg" alt="Arabiyya Bell System logo" />
              </div>
            </figure>
            <h1
              class="font-display text-4xl font-semibold tracking-tight text-base-content"
            >
              Arabiyya Bell System
            </h1>
          </div>

          <div class="space-y-5">
            <div
              v-if="sessionExpired && !loginError"
              class="alert alert-warning text-sm"
              role="alert"
            >
              Your session expired. Please sign in again.
            </div>

            <div
              v-if="loginError"
              class="alert alert-error text-sm"
              role="alert"
            >
              {{ loginError }}
            </div>

            <fieldset class="fieldset">
              <legend class="fieldset-legend">Username</legend>
              <input
                v-model="username"
                class="input w-full"
                :disabled="loginMutation.isPending.value"
                placeholder="admin"
                type="text"
              />
              <legend class="fieldset-legend">Password</legend>
              <input
                v-model="password"
                class="input w-full"
                :disabled="loginMutation.isPending.value"
                placeholder="Password"
                type="password"
                @keydown.enter="handleLogin"
              />
            </fieldset>
            <button
              class="btn btn-primary w-full"
              type="button"
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
