<script setup lang="ts">
import { Icon } from "@iconify/vue";

defineProps<{
  icon: string;
  items: string[];
  title: string;
}>();
</script>

<template>
  <article class="card bg-base-100 shadow-sm">
    <div class="card-body gap-4">
      <div class="flex items-center gap-3">
        <div class="avatar placeholder">
          <div class="size-12 rounded-box bg-primary text-primary-content">
            <Icon :icon="icon" class="text-2xl" />
          </div>
        </div>
        <h2 class="text-lg font-semibold text-base-content">
          {{ title }}
        </h2>
      </div>
      <ul class="menu rounded-box bg-base-200/60">
        <li v-for="item in items" :key="item">
          <span class="leading-6 text-base-content/70">{{ item }}</span>
        </li>
      </ul>
    </div>
  </article>
</template>
