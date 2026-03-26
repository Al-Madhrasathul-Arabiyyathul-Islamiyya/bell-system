<script setup lang="ts">
import { computed } from "vue";
import type { AudioFileType } from "../../../lib/api/audio";

const props = defineProps<{
  fileType: AudioFileType;
}>();

const label = computed(() => {
  if (props.fileType === "school_song") {
    return "School Song";
  }

  if (props.fileType === "anthem") {
    return "Anthem";
  }

  if (props.fileType === "bell") {
    return "Bell";
  }

  return "Other";
});

const badgeClass = computed(() => {
  if (props.fileType === "bell") {
    return "badge-primary";
  }

  if (props.fileType === "anthem") {
    return "badge-secondary";
  }

  if (props.fileType === "school_song") {
    return "badge-accent";
  }

  return "badge-outline";
});
</script>

<template>
  <span class="badge badge-soft font-medium" :class="badgeClass">
    {{ label }}
  </span>
</template>
