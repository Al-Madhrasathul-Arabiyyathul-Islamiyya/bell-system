<script setup lang="ts">
import type { ScheduleItemVm } from "../../../lib/api/schedule";

defineProps<{
  deletingId?: null | string;
  items: ScheduleItemVm[];
  loading?: boolean;
}>();

const emit = defineEmits<{
  delete: [item: ScheduleItemVm];
  edit: [item: ScheduleItemVm];
}>();

const dayLabels: Record<number, string> = {
  1: "Sun",
  2: "Mon",
  3: "Tue",
  4: "Wed",
  5: "Thu",
  6: "Fri",
  7: "Sat",
};
</script>

<template>
  <div class="overflow-x-auto rounded-box border border-base-300">
    <table class="table table-zebra">
      <thead>
        <tr>
          <th>Name</th>
          <th>Time</th>
          <th>Days</th>
          <th>Session</th>
          <th>Audio</th>
          <th class="w-36 text-right">Actions</th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="loading">
          <td
            colspan="6"
            class="py-10 text-center text-sm text-base-content/65"
          >
            Loading schedule items...
          </td>
        </tr>
        <tr v-else-if="items.length === 0">
          <td
            colspan="6"
            class="py-10 text-center text-sm text-base-content/65"
          >
            No schedule items matched the current session filter.
          </td>
        </tr>
        <tr v-for="item in items" :key="item.id">
          <td>
            <div class="font-medium text-base-content">{{ item.name }}</div>
          </td>
          <td>
            <span class="font-semibold text-base-content/75">
              {{ item.time.slice(0, 5) }}
            </span>
          </td>
          <td>
            <div class="flex flex-wrap gap-1">
              <span
                v-for="day in item.days"
                :key="day"
                class="badge badge-outline badge-sm"
              >
                {{ dayLabels[day] ?? day }}
              </span>
            </div>
          </td>
          <td>{{ item.sessionName ?? "No session" }}</td>
          <td>{{ item.soundName ?? "No audio linked" }}</td>
          <td>
            <div class="flex justify-end gap-2">
              <button
                class="btn btn-sm"
                type="button"
                @click="emit('edit', item)"
              >
                Edit
              </button>
              <button
                class="btn btn-sm btn-error btn-outline"
                type="button"
                :disabled="deletingId === item.id"
                @click="emit('delete', item)"
              >
                {{ deletingId === item.id ? "Deleting..." : "Delete" }}
              </button>
            </div>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
