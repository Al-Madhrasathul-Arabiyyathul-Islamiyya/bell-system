<script setup lang="ts">
import { computed, reactive, watch } from "vue";
import { useRegleSchema } from "@regle/schemas";
import { z } from "zod";
import type { SessionVm } from "../../../lib/api/sessions";

type SessionFormPayload = {
  endTime: string;
  name: string;
  startTime: string;
};

const timePattern = /^([01]\d|2[0-3]):([0-5]\d)$/;

const props = withDefaults(
  defineProps<{
    error?: null | string;
    mode: "create" | "edit";
    open: boolean;
    pending?: boolean;
    session?: null | SessionVm;
  }>(),
  {
    error: null,
    pending: false,
    session: null,
  },
);

const emit = defineEmits<{
  close: [];
  save: [payload: SessionFormPayload];
}>();

const form = reactive({
  endTime: "",
  name: "",
  startTime: "",
});

const schema = z
  .object({
    endTime: z
      .string()
      .trim()
      .regex(timePattern, "End time must use HH:MM format."),
    name: z.string().trim().min(1, "Session name is required."),
    startTime: z
      .string()
      .trim()
      .regex(timePattern, "Start time must use HH:MM format."),
  })
  .refine((value) => value.endTime > value.startTime, {
    message: "End time must be later than the start time.",
    path: ["endTime"],
  });

const { r$ } = useRegleSchema(form, schema);

const dialogTitle = computed(() => {
  return props.mode === "create" ? "Create Session" : "Edit Session";
});

const submitLabel = computed(() => {
  if (props.pending) {
    return props.mode === "create" ? "Creating..." : "Saving...";
  }

  return props.mode === "create" ? "Create Session" : "Save Changes";
});

watch(
  () => [props.mode, props.open, props.session?.id] as const,
  () => {
    if (!props.open) {
      return;
    }

    form.name = props.session?.name ?? "";
    form.startTime = normalizeTimeValue(props.session?.startTime);
    form.endTime = normalizeTimeValue(props.session?.endTime);
    r$.$reset();
  },
  { immediate: true },
);

async function handleSubmit() {
  const result = await r$.$validate();

  if (!result.valid) {
    return;
  }

  emit("save", {
    endTime: form.endTime.trim(),
    name: form.name.trim(),
    startTime: form.startTime.trim(),
  });
}

function normalizeTimeValue(value: string | undefined) {
  if (!value) {
    return "";
  }

  return value.slice(0, 5);
}
</script>

<template>
  <div v-if="open" class="modal modal-open">
    <div class="modal-box max-w-lg">
      <button
        class="btn btn-circle btn-ghost btn-sm absolute top-4 right-4"
        type="button"
        :disabled="pending"
        aria-label="Close session form"
        @click="emit('close')"
      >
        ×
      </button>

      <h2 class="font-display pr-8 text-2xl font-semibold text-base-content">
        {{ dialogTitle }}
      </h2>
      <p class="mt-2 text-sm leading-7 text-base-content/70">
        {{
          mode === "create"
            ? "Create a session window for schedule items and bell playback."
            : "Update the session name or time range."
        }}
      </p>

      <div v-if="error" class="alert alert-error mt-5 text-sm" role="alert">
        {{ error }}
      </div>

      <form
        class="mt-6 grid gap-4 md:grid-cols-2"
        @submit.prevent="handleSubmit"
      >
        <fieldset class="fieldset md:col-span-2">
          <legend class="fieldset-legend">Session name</legend>
          <input
            v-model="r$.$value.name"
            class="input w-full"
            :disabled="pending"
            type="text"
            placeholder="Morning Session"
          />
          <p v-if="r$.name.$error" class="fieldset-label text-error">
            {{ r$.name.$errors[0] }}
          </p>
        </fieldset>

        <fieldset class="fieldset">
          <legend class="fieldset-legend">Start time</legend>
          <input
            v-model="r$.$value.startTime"
            class="input w-full"
            :disabled="pending"
            type="time"
          />
          <p class="fieldset-label">24-hour format.</p>
          <p v-if="r$.startTime.$error" class="fieldset-label text-error">
            {{ r$.startTime.$errors[0] }}
          </p>
        </fieldset>

        <fieldset class="fieldset">
          <legend class="fieldset-legend">End time</legend>
          <input
            v-model="r$.$value.endTime"
            class="input w-full"
            :disabled="pending"
            type="time"
          />
          <p class="fieldset-label">Must be later than the start time.</p>
          <p v-if="r$.endTime.$error" class="fieldset-label text-error">
            {{ r$.endTime.$errors[0] }}
          </p>
        </fieldset>

        <div class="modal-action mt-6 md:col-span-2">
          <button
            class="btn btn-ghost"
            type="button"
            :disabled="pending"
            @click="emit('close')"
          >
            Cancel
          </button>
          <button class="btn btn-primary" type="submit" :disabled="pending">
            {{ submitLabel }}
          </button>
        </div>
      </form>
    </div>
    <div class="modal-backdrop" @click="emit('close')" />
  </div>
</template>
