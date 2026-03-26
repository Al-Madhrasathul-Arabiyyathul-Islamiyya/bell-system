<script setup lang="ts">
import { useTitle } from "@vueuse/core";
import { RouterLink } from "vue-router";
import { appEnv } from "../lib/env";

useTitle(`Not Found • ${appEnv.appName}`);
</script>

<template>
  <main class="flex min-h-screen items-center justify-center bg-base-200 px-4">
    <article class="card w-full max-w-xl bg-base-100 shadow-2xl">
      <div class="card-body gap-4">
        <p
          class="text-xs font-semibold uppercase tracking-[0.24em] text-primary"
        >
          404
        </p>
        <h1
          class="font-display text-4xl font-semibold tracking-tight text-base-content"
        >
          Page not found
        </h1>
        <p class="text-sm leading-7 text-base-content/70">
          This route is outside the current admin-client shell. Use the
          navigation or return to the dashboard.
        </p>
        <div class="card-actions justify-start">
          <RouterLink class="btn btn-primary" :to="{ name: 'dashboard' }"
            >Go to dashboard</RouterLink
          >
        </div>
      </div>
    </article>
  </main>
</template>
