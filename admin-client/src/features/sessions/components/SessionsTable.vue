<script setup lang="ts">
import type { SessionVm } from "../../../lib/api/sessions";

withDefaults(
  defineProps<{
    deletingId?: null | string;
    loading?: boolean;
    sessions: SessionVm[];
  }>(),
  {
    deletingId: null,
    loading: false,
  },
);

const emit = defineEmits<{
  delete: [session: SessionVm];
  edit: [session: SessionVm];
}>();

function formatTime(value: string) {
  if (value.length >= 5) {
    return value.slice(0, 5);
  }

  return value;
}
</script>

<template>
  <div class="overflow-x-auto rounded-box border border-base-300 bg-base-100">
    <table class="table table-zebra">
      <thead>
        <tr>
          <th>Session</th>
          <th>Start</th>
          <th>End</th>
          <th>Window</th>
          <th class="w-32 text-right">Actions</th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="loading">
          <td colspan="5">
            <div class="flex items-center justify-center py-10">
              <span class="loading loading-spinner loading-md text-primary" />
            </div>
          </td>
        </tr>

        <tr v-else-if="sessions.length === 0">
          <td colspan="5">
            <div class="flex flex-col items-center gap-3 py-10 text-center">
              <p class="text-base font-semibold text-base-content">
                No sessions configured
              </p>
              <p class="text-sm text-base-content/65">
                Create the first session to define a bell schedule window.
              </p>
            </div>
          </td>
        </tr>

        <tr v-for="session in sessions" :key="session.id">
          <td class="font-medium text-base-content">{{ session.name }}</td>
          <td class="text-sm text-base-content/75">
            {{ formatTime(session.startTime) }}
          </td>
          <td class="text-sm text-base-content/75">
            {{ formatTime(session.endTime) }}
          </td>
          <td>
            <span class="badge badge-outline font-medium">
              {{ formatTime(session.startTime) }} -
              {{ formatTime(session.endTime) }}
            </span>
          </td>
          <td>
            <div class="flex justify-end gap-2">
              <button
                class="btn btn-sm btn-outline"
                type="button"
                @click="emit('edit', session)"
              >
                Edit
              </button>
              <button
                class="btn btn-sm btn-error btn-soft"
                type="button"
                :disabled="deletingId === session.id"
                @click="emit('delete', session)"
              >
                {{ deletingId === session.id ? "Deleting..." : "Delete" }}
              </button>
            </div>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
