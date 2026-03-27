<script setup lang="ts">
import type { CurrentScheduleResponse } from "../../../lib/api/schedule";

type ScheduleItem = CurrentScheduleResponse["items"][number];

withDefaults(
  defineProps<{
    items: ScheduleItem[];
    loading?: boolean;
    title?: string;
  }>(),
  {
    loading: false,
    title: "Upcoming Bells",
  },
);
</script>

<template>
  <article class="card border border-base-300 bg-base-100 shadow-sm">
    <div class="card-body gap-4">
      <div class="space-y-2">
        <p
          class="text-xs font-semibold uppercase tracking-[0.24em] text-primary"
        >
          Queue
        </p>
        <h2 class="font-display text-3xl font-semibold text-base-content">
          {{ title }}
        </h2>
        <p class="text-sm leading-7 text-base-content/70">
          The next bells expected to ring from the live session schedule.
        </p>
      </div>

      <div
        v-if="loading"
        class="flex items-center gap-3 rounded-box border border-base-300 bg-base-200/50 px-4 py-3 text-sm text-base-content/70"
      >
        <span class="loading loading-spinner loading-sm" />
        Loading upcoming schedule items...
      </div>

      <ul v-else-if="items.length > 0" class="space-y-3">
        <li
          v-for="item in items"
          :key="item.id"
          class="flex items-start justify-between gap-4 rounded-box border border-base-300 px-4 py-3"
        >
          <div class="min-w-0">
            <p class="truncate font-medium text-base-content">
              {{ item.name }}
            </p>
            <p class="truncate text-sm text-base-content/65">
              {{ item.sound?.name ?? "No audio linked" }}
            </p>
          </div>
          <span class="shrink-0 text-sm font-semibold text-base-content/70">
            {{ item.time }}
          </span>
        </li>
      </ul>

      <p v-else class="text-sm leading-7 text-base-content/65">
        No upcoming bells are queued right now.
      </p>
    </div>
  </article>
</template>
