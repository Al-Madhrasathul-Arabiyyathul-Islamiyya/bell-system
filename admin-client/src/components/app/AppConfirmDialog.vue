<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    confirmLabel?: string;
    description: string;
    open: boolean;
    pending?: boolean;
    title: string;
    tone?: "error" | "info" | "primary" | "warning";
  }>(),
  {
    confirmLabel: "Confirm",
    pending: false,
    tone: "primary",
  },
);

const emit = defineEmits<{
  close: [];
  confirm: [];
}>();

function confirmClass() {
  if (props.tone === "error") {
    return "btn-error";
  }

  if (props.tone === "warning") {
    return "btn-warning";
  }

  if (props.tone === "info") {
    return "btn-info";
  }

  return "btn-primary";
}
</script>

<template>
  <div v-if="open" class="modal modal-open">
    <div class="modal-box max-w-md">
      <h2 class="font-display text-2xl font-semibold text-base-content">
        {{ title }}
      </h2>
      <p class="mt-3 text-sm leading-7 text-base-content/70">
        {{ description }}
      </p>

      <div class="modal-action">
        <button
          class="btn btn-ghost"
          type="button"
          :disabled="pending"
          @click="emit('close')"
        >
          Cancel
        </button>
        <button
          class="btn"
          :class="confirmClass()"
          type="button"
          :disabled="pending"
          @click="emit('confirm')"
        >
          {{ pending ? "Working..." : confirmLabel }}
        </button>
      </div>
    </div>
    <div class="modal-backdrop" @click="emit('close')" />
  </div>
</template>
