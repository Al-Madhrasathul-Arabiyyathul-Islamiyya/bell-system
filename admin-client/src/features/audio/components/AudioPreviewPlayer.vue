<script setup lang="ts">
import { computed } from "vue";
import { Icon } from "@iconify/vue";
import type { AudioFileVm } from "../../../lib/api/audio";
import AudioFileTypeBadge from "./AudioFileTypeBadge.vue";

const props = defineProps<{
  file: null | AudioFileVm;
}>();

const previewLabel = computed(() => {
  if (!props.file) {
    return "Select an audio file to preview it here.";
  }

  return props.file.name;
});
</script>

<template>
  <article class="card border border-base-300 bg-base-100 shadow-sm">
    <div class="card-body gap-5">
      <div class="space-y-2">
        <p
          class="text-xs font-semibold uppercase tracking-[0.24em] text-primary"
        >
          Preview
        </p>
        <h2 class="font-display text-3xl font-semibold text-base-content">
          Audio Player
        </h2>
        <p class="text-sm leading-7 text-base-content/70">
          Preview the selected file directly from the public binary content
          route before editing or deleting it.
        </p>
      </div>

      <div
        v-if="!file"
        class="flex min-h-48 flex-col items-center justify-center gap-3 rounded-box border border-dashed border-base-300 bg-base-200 px-6 text-center"
      >
        <Icon
          icon="solar:music-note-3-bold-duotone"
          class="text-5xl text-base-content/35"
        />
        <div class="space-y-1">
          <p class="text-base font-semibold text-base-content">
            No file selected
          </p>
          <p class="text-sm text-base-content/65">
            {{ previewLabel }}
          </p>
        </div>
      </div>

      <div v-else class="space-y-4">
        <div class="space-y-2">
          <div class="flex items-center gap-2">
            <AudioFileTypeBadge :file-type="file.fileType" />
            <span class="text-sm text-base-content/60">Previewing now</span>
          </div>
          <h3 class="text-xl font-semibold text-base-content">
            {{ file.name }}
          </h3>
          <p class="text-sm break-all text-base-content/65">
            Checksum: {{ file.checksum }}
          </p>
        </div>

        <audio
          class="w-full"
          controls
          preload="metadata"
          :src="file.contentUrl"
        >
          Your browser does not support HTML audio playback.
        </audio>

        <div class="flex justify-end">
          <a
            class="btn btn-outline"
            :href="file.contentUrl"
            target="_blank"
            rel="noreferrer"
          >
            Open Binary
          </a>
        </div>
      </div>
    </div>
  </article>
</template>
