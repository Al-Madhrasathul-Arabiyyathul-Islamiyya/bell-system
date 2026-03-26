<script setup lang="ts">
import { appEnv } from "../lib/env";
import { AppFeatureCard, AppPageHeader } from "./page-exports";
</script>

<template>
  <section class="page-stack">
    <AppPageHeader
      eyebrow="Planning"
      title="System"
      description="System controls, live connection status, and current-user account actions will live here once the API and websocket layers are wired."
    />

    <div class="grid gap-4 lg:grid-cols-[minmax(0,1fr)_22rem]">
      <AppFeatureCard
        title="Planned implementation"
        icon="solar:settings-bold-duotone"
        :items="[
          'System state query plus pause and resume mutation hooks.',
          'Cancel-next-bell action with a friendly 404 fallback.',
          'Change-password flow routed through the plain JSON exception endpoint.',
        ]"
      />

      <article class="card bg-base-100 shadow-sm">
        <div class="card-body gap-3">
          <h2 class="card-title">Configured endpoints</h2>
          <div class="space-y-2">
            <p class="text-sm text-base-content/70">REST base</p>
            <div class="mockup-code text-xs">
              <pre><code>{{ appEnv.apiBaseUrl }}</code></pre>
            </div>
          </div>
          <div class="space-y-2">
            <p class="text-sm text-base-content/70">WebSocket base</p>
            <div class="mockup-code text-xs">
              <pre><code>{{ appEnv.wsBaseUrl }}</code></pre>
            </div>
          </div>
        </div>
      </article>
    </div>
  </section>
</template>
