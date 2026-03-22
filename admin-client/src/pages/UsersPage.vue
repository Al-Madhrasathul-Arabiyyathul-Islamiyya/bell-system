<script setup lang="ts">
import { AppFeatureCard, AppPageHeader } from "./page-exports";
</script>

<template>
  <section class="page-stack">
    <AppPageHeader
      eyebrow="Planning"
      title="Users"
      description="This route will become the paginated user-management surface for admin-only CRUD, role changes, and password reset flows."
    />

    <AppFeatureCard
      title="Planned implementation"
      icon="solar:users-group-rounded-bold-duotone"
      :items="[
        'Server-side pagination and role filtering using query keys that match the API spec.',
        'Regle-powered forms with action-specific mutation hooks.',
        'Inline mapping for JSON:API validation errors returned by the backend.',
      ]"
    />
  </section>
</template>
