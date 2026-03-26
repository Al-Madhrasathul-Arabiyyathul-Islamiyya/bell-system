<script setup lang="ts">
import { computed } from "vue";
import { Icon } from "@iconify/vue";
import type { CurrentScheduleResponse } from "../../../lib/api/schedule";

const props = withDefaults(
  defineProps<{
    error?: null | string;
    loading?: boolean;
    schedule: CurrentScheduleResponse | null;
  }>(),
  {
    error: null,
    loading: false,
  },
);

const currentItem = computed(() => {
  return (
    props.schedule?.items.find((item) => item.status === "current") ?? null
  );
});
</script>

<template>
  <article class="card border border-base-300 bg-base-100 shadow-sm">
    <div class="card-body gap-4">
      <div class="space-y-2">
        <p
          class="text-xs font-semibold uppercase tracking-[0.24em] text-primary"
        >
          Live Schedule
        </p>
        <h2 class="font-display text-3xl font-semibold text-base-content">
          {{ schedule?.session?.name ?? "No Active Session" }}
        </h2>
        <p class="text-sm leading-7 text-base-content/70">
          {{
            schedule?.session
              ? `${schedule.session.startTime?.slice(0, 5)} - ${schedule.session.endTime?.slice(0, 5)}`
              : "No session is active at the current time."
          }}
        </p>
      </div>

      <div v-if="error" class="alert alert-error text-sm" role="alert">
        {{ error }}
      </div>

      <div
        v-else-if="loading"
        class="flex items-center gap-3 rounded-box border border-base-300 bg-base-200/50 px-4 py-3 text-sm text-base-content/70"
      >
        <span class="loading loading-spinner loading-sm" />
        Loading the current schedule snapshot...
      </div>

      <div
        v-else-if="currentItem"
        class="rounded-box border border-success/30 bg-success/10 px-4 py-3"
      >
        <p
          class="text-xs font-semibold uppercase tracking-[0.2em] text-success"
        >
          Current Bell
        </p>
        <div class="mt-2 flex items-start justify-between gap-3">
          <div>
            <p class="text-lg font-semibold text-base-content">
              {{ currentItem.name }}
            </p>
            <p class="text-sm text-base-content/70">
              {{ currentItem.time }} •
              {{ currentItem.sound?.name ?? "No audio linked" }}
            </p>
          </div>
          <span class="badge badge-success badge-outline">Current</span>
        </div>
      </div>

      <ul class="space-y-3">
        <li
          v-for="item in schedule?.items ?? []"
          :key="item.id"
          class="flex items-start gap-3 rounded-box border border-base-300 px-4 py-3"
        >
          <div
            class="mt-0.5 rounded-full p-2"
            :class="
              item.status === 'current'
                ? 'bg-success/15 text-success'
                : item.status === 'completed'
                  ? 'bg-base-300 text-base-content/60'
                  : 'bg-primary/10 text-primary'
            "
          >
            <Icon
              :icon="
                item.status === 'current'
                  ? 'solar:alarm-play-bold-duotone'
                  : item.status === 'completed'
                    ? 'solar:check-circle-bold-duotone'
                    : 'solar:clock-circle-bold-duotone'
              "
              class="text-lg"
            />
          </div>
          <div class="min-w-0 flex-1">
            <div class="flex items-center justify-between gap-3">
              <p class="truncate font-medium text-base-content">
                {{ item.name }}
              </p>
              <span class="text-sm font-semibold text-base-content/70">
                {{ item.time }}
              </span>
            </div>
            <p class="truncate text-sm text-base-content/65">
              {{ item.sound?.name ?? "No audio linked" }}
            </p>
          </div>
        </li>
      </ul>

      <p
        v-if="!loading && !error && (schedule?.items.length ?? 0) === 0"
        class="text-sm leading-7 text-base-content/65"
      >
        No schedule items are active for the current session and day.
      </p>
    </div>
  </article>
</template>
