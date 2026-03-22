<script setup lang="ts">
import { AppFeatureCard, AppPageHeader } from "./page-exports";
</script>

<template>
  <section class="page-stack">
    <AppPageHeader
      eyebrow="Planning"
      title="Schedule"
      description="The schedule workspace will host filtered listings, session and day controls, and JSON:API relationship-aware forms for sounds and sessions."
    />

    <AppFeatureCard
      title="Planned implementation"
      icon="solar:calendar-mark-bold-duotone"
      :items="[
        'Route-driven filters for day and session selection.',
        'Query invalidation from mutations and websocket schedule updates.',
        'Reusable table and form sections rather than page-local fetch code.',
      ]"
    />
  </section>
</template>
