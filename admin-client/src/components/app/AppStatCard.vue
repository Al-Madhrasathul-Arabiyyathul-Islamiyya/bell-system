<script setup lang="ts">
defineProps<{
  description: string;
  title: string;
  value: string;
}>();
</script>

<template>
  <article class="card bg-base-100 shadow-sm">
    <div class="card-body gap-2">
      <p
        class="text-xs font-semibold uppercase tracking-[0.2em] text-base-content/50"
      >
        {{ title }}
      </p>
      <p class="text-3xl font-semibold text-base-content">
        {{ value }}
      </p>
      <p class="text-sm leading-6 text-base-content/70">
        {{ description }}
      </p>
    </div>
  </article>
</template>
