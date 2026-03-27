<script setup lang="ts">
defineProps<{
  description: string;
  eyebrow?: string;
  title: string;
}>();
</script>

<template>
  <header class="space-y-3">
    <p
      v-if="eyebrow"
      class="text-xs font-semibold uppercase tracking-[0.24em] text-primary"
    >
      {{ eyebrow }}
    </p>
    <div class="space-y-2">
      <h1
        class="font-display text-3xl font-semibold tracking-tight text-base-content md:text-4xl"
      >
        {{ title }}
      </h1>
      <p class="max-w-3xl text-sm leading-7 text-base-content/70 md:text-base">
        {{ description }}
      </p>
    </div>
  </header>
</template>
