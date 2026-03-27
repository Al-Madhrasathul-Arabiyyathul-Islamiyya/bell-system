<script setup lang="ts">
import { computed } from "vue";
import { Icon } from "@iconify/vue";
import type { SystemStateVm } from "../../lib/api/system";

const props = withDefaults(
  defineProps<{
    iconOnly?: boolean;
    pending?: boolean;
    state: null | SystemStateVm["state"] | undefined;
  }>(),
  {
    iconOnly: false,
    pending: false,
  },
);

const emit = defineEmits<{
  toggle: [];
}>();

const isPaused = computed(() => props.state === "paused");
const buttonClass = computed(() => {
  return isPaused.value ? "btn-success" : "btn-warning";
});
const icon = computed(() => {
  return isPaused.value
    ? "solar:play-bold-duotone"
    : "solar:pause-bold-duotone";
});
const label = computed(() => {
  return isPaused.value ? "Resume System" : "Pause System";
});
const tooltip = computed(() => {
  return isPaused.value ? "Resume bell playback" : "Pause bell playback";
});
</script>

<template>
  <div class="tooltip tooltip-bottom" :data-tip="tooltip">
    <button
      class="btn"
      :class="[buttonClass, iconOnly ? 'btn-square btn-sm' : '']"
      type="button"
      :disabled="pending || !state"
      @click="emit('toggle')"
    >
      <span v-if="pending" class="loading loading-spinner loading-xs" />
      <Icon v-else :icon="icon" class="text-lg" />
      <span v-if="!iconOnly">
        {{ pending ? "Updating..." : label }}
      </span>
    </button>
  </div>
</template>
