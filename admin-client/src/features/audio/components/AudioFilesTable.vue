<script setup lang="ts">
import { computed } from "vue";
import type { AudioFileVm } from "../../../lib/api/audio";
import AudioFileTypeBadge from "./AudioFileTypeBadge.vue";

const props = withDefaults(
  defineProps<{
    deletingId?: null | string;
    loading?: boolean;
    previewId?: null | string;
    files: AudioFileVm[];
  }>(),
  {
    deletingId: null,
    loading: false,
    previewId: null,
  },
);

const emit = defineEmits<{
  delete: [file: AudioFileVm];
  edit: [file: AudioFileVm];
  preview: [file: AudioFileVm];
}>();

const empty = computed(() => !props.loading && props.files.length === 0);

function formatDate(value: string) {
  return new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
}
</script>

<template>
  <div class="overflow-x-auto rounded-box border border-base-300 bg-base-100">
    <table class="table table-zebra">
      <thead>
        <tr>
          <th>Name</th>
          <th>Type</th>
          <th>Checksum</th>
          <th>Updated</th>
          <th class="w-48 text-right">Actions</th>
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

        <tr v-else-if="empty">
          <td colspan="5">
            <div class="flex flex-col items-center gap-3 py-10 text-center">
              <p class="text-base font-semibold text-base-content">
                No audio files found
              </p>
              <p class="text-sm text-base-content/65">
                Adjust the file type filter or upload a new audio file.
              </p>
            </div>
          </td>
        </tr>

        <tr v-for="file in files" :key="file.id">
          <td class="font-medium text-base-content">{{ file.name }}</td>
          <td>
            <AudioFileTypeBadge :file-type="file.fileType" />
          </td>
          <td class="max-w-52 truncate text-sm text-base-content/65">
            {{ file.checksum }}
          </td>
          <td class="text-sm text-base-content/70">
            {{ formatDate(file.updatedAt) }}
          </td>
          <td>
            <div class="flex justify-end gap-2">
              <button
                class="btn btn-sm"
                :class="previewId === file.id ? 'btn-primary' : 'btn-outline'"
                type="button"
                @click="emit('preview', file)"
              >
                Preview
              </button>
              <button
                class="btn btn-sm btn-outline"
                type="button"
                @click="emit('edit', file)"
              >
                Edit
              </button>
              <button
                class="btn btn-sm btn-error btn-soft"
                type="button"
                :disabled="deletingId === file.id"
                @click="emit('delete', file)"
              >
                {{ deletingId === file.id ? "Deleting..." : "Delete" }}
              </button>
            </div>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
