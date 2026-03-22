<script setup lang="ts">
import { AppFeatureCard, AppPageHeader } from "./page-exports";
</script>

<template>
  <section class="page-stack">
    <AppPageHeader
      eyebrow="Planning"
      title="Audio Library"
      description="This page will handle paginated audio metadata, uploads, preview links, and CRUD operations with multipart exceptions to the JSON:API path."
    />

    <AppFeatureCard
      title="Planned implementation"
      icon="solar:music-library-2-bold-duotone"
      :items="[
        'Paginated query layer with file-type filters and optimistic refresh.',
        'Multipart upload form with validation and progress feedback.',
        'Binary preview routes using the public content endpoint.',
      ]"
    />
  </section>
</template>
