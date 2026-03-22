<script setup lang="ts">
import { Icon } from "@iconify/vue";

defineProps<{
  icon: string;
  items: string[];
  title: string;
}>();
</script>

<template>
  <article class="card border border-base-300 bg-base-100 shadow-sm">
    <div class="card-body gap-4">
      <div class="flex items-center gap-3">
        <div
          class="flex size-12 items-center justify-center rounded-box bg-primary/12 text-primary"
        >
          <Icon :icon="icon" class="text-2xl" />
        </div>
        <h2 class="text-lg font-semibold text-base-content">
          {{ title }}
        </h2>
      </div>
      <ul class="space-y-2 text-sm leading-6 text-base-content/70">
        <li v-for="item in items" :key="item" class="flex items-start gap-2">
          <span class="mt-2 size-1.5 rounded-full bg-primary" />
          <span>{{ item }}</span>
        </li>
      </ul>
    </div>
  </article>
</template>
