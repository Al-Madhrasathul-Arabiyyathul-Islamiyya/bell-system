<script setup lang="ts">
import { AppFeatureCard, AppPageHeader } from "./page-exports";
</script>

<template>
  <section class="page-stack">
    <AppPageHeader
      eyebrow="Planning"
      title="Sessions"
      description="This route is reserved for the session list, CRUD form flows, and conflict handling against the backend schedule window rules."
    />

    <AppFeatureCard
      title="Planned implementation"
      icon="solar:calendar-bold-duotone"
      :items="[
        'Fetch the unpaginated sessions collection through Vue Query.',
        'Use Regle and Zod-backed schemas for create and edit forms.',
        'Surface backend 409 conflict responses inline in the form shell.',
      ]"
    />
  </section>
</template>
