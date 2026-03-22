<script setup lang="ts">
import { AppFeatureCard, AppPageHeader, AppStatCard } from "./page-exports";
</script>

<template>
  <section class="page-stack">
    <AppPageHeader
      eyebrow="Overview"
      title="Dashboard"
      description="The shell is now ready for the system-state summary, current session card, today schedule summary, connected clients, and quick actions."
    />

    <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
      <AppStatCard
        title="Router"
        value="Ready"
        description="Protected routing, default redirects, and page titles are active."
      />
      <AppStatCard
        title="State"
        value="Ready"
        description="Pinia manages auth persistence and app preferences without leaking server state."
      />
      <AppStatCard
        title="Query Cache"
        value="Ready"
        description="Vue Query is registered for route-level resource loading and invalidation."
      />
      <AppStatCard
        title="Theme"
        value="Ready"
        description="DaisyUI theme persistence is driven through VueUse color-mode storage."
      />
    </div>

    <div class="grid gap-4 xl:grid-cols-3">
      <AppFeatureCard
        title="Immediate next work"
        icon="solar:clipboard-list-bold-duotone"
        :items="[
          'Implement auth API transport and replace the placeholder login flow.',
          'Add the shared ofetch client with JSON:API-aware request and response helpers.',
          'Create the WebSocket composable with query invalidation hooks.',
        ]"
      />
      <AppFeatureCard
        title="Module split"
        icon="solar:widget-add-bold-duotone"
        :items="[
          'Dashboard as the integration point for system state, schedule current, and connected clients.',
          'Users, sessions, schedule, audio, and system pages are already stubbed for feature delivery.',
          'The shell and nav are stable enough to start module work immediately.',
        ]"
      />
      <AppFeatureCard
        title="VueUse baseline"
        icon="solar:bolt-circle-bold-duotone"
        :items="[
          'useStorage persists auth and theme state.',
          'useColorMode binds the DaisyUI light and dark themes.',
          'useBreakpoints, useOnline, useNow, and useTitle are already active in the shell.',
        ]"
      />
    </div>
  </section>
</template>
