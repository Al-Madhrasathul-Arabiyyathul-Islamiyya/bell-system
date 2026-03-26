<script setup lang="ts">
import { computed, reactive, watch } from "vue";
import { useRegleSchema } from "@regle/schemas";
import { z } from "zod";
import type { AudioFileVm } from "../../../lib/api/audio";
import type { ScheduleItemVm } from "../../../lib/api/schedule";
import type { SessionVm } from "../../../lib/api/sessions";

type ScheduleFormPayload = {
  days: number[];
  name: string;
  sessionId: string;
  soundId: string;
  time: string;
};

const props = withDefaults(
  defineProps<{
    audioFiles: AudioFileVm[];
    error?: null | string;
    item?: null | ScheduleItemVm;
    mode: "create" | "edit";
    open: boolean;
    pending?: boolean;
    sessions: SessionVm[];
  }>(),
  {
    error: null,
    item: null,
    pending: false,
  },
);

const emit = defineEmits<{
  close: [];
  save: [payload: ScheduleFormPayload];
}>();

const dayOptions = [
  { label: "Sun", value: 1 },
  { label: "Mon", value: 2 },
  { label: "Tue", value: 3 },
  { label: "Wed", value: 4 },
  { label: "Thu", value: 5 },
  { label: "Fri", value: 6 },
  { label: "Sat", value: 7 },
] as const;

const form = reactive({
  days: [] as number[],
  name: "",
  sessionId: "",
  soundId: "",
  time: "",
});

const schema = z.object({
  days: z
    .array(z.number().int().min(1).max(7))
    .min(1, "Select at least one day."),
  name: z.string().trim().min(1, "Schedule item name is required."),
  sessionId: z.string().trim().min(1, "Choose a session."),
  soundId: z.string().trim().min(1, "Choose an audio file."),
  time: z
    .string()
    .trim()
    .regex(/^([01]\d|2[0-3]):[0-5]\d$/, "Time must use the HH:MM format."),
});

const { r$ } = useRegleSchema(form, schema);

const dialogTitle = computed(() => {
  return props.mode === "create"
    ? "Create Schedule Item"
    : "Edit Schedule Item";
});

const submitLabel = computed(() => {
  if (props.pending) {
    return props.mode === "create" ? "Creating..." : "Saving...";
  }

  return props.mode === "create" ? "Create Item" : "Save Changes";
});

watch(
  () =>
    [
      props.open,
      props.item?.id,
      props.sessions.map((session) => session.id).join(","),
    ] as const,
  () => {
    if (!props.open) {
      return;
    }

    form.name = props.item?.name ?? "";
    form.time = props.item?.time ?? "";
    form.sessionId = props.item?.sessionId ?? props.sessions[0]?.id ?? "";
    form.soundId = props.item?.soundId ?? "";
    form.days = props.item?.days ? [...props.item.days] : [];
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
    days: [...form.days].sort((left, right) => left - right),
    name: form.name.trim(),
    sessionId: form.sessionId,
    soundId: form.soundId,
    time: form.time.trim(),
  });
}

function updateDay(day: number, checked: boolean) {
  if (checked && !form.days.includes(day)) {
    form.days = [...form.days, day];
    return;
  }

  if (!checked) {
    form.days = form.days.filter((value) => value !== day);
  }
}
</script>

<template>
  <div v-if="open" class="modal modal-open">
    <div class="modal-box max-w-2xl">
      <button
        class="btn btn-circle btn-ghost btn-sm absolute top-4 right-4"
        type="button"
        :disabled="pending"
        aria-label="Close schedule form"
        @click="emit('close')"
      >
        ×
      </button>

      <h2 class="font-display pr-8 text-2xl font-semibold text-base-content">
        {{ dialogTitle }}
      </h2>
      <p class="mt-2 text-sm leading-7 text-base-content/70">
        Link a bell time to a session and audio file, then choose the days on
        which it should run.
      </p>

      <div v-if="error" class="alert alert-error mt-5 text-sm" role="alert">
        {{ error }}
      </div>

      <form
        class="mt-6 grid gap-4 md:grid-cols-2"
        @submit.prevent="handleSubmit"
      >
        <fieldset class="fieldset md:col-span-2">
          <legend class="fieldset-legend">Item name</legend>
          <input
            v-model="r$.$value.name"
            class="input w-full"
            :disabled="pending"
            type="text"
            placeholder="Morning assembly bell"
          />
          <p v-if="r$.name.$error" class="fieldset-label text-error">
            {{ r$.name.$errors[0] }}
          </p>
        </fieldset>

        <fieldset class="fieldset">
          <legend class="fieldset-legend">Time</legend>
          <input
            v-model="r$.$value.time"
            class="input w-full"
            :disabled="pending"
            type="time"
          />
          <p v-if="r$.time.$error" class="fieldset-label text-error">
            {{ r$.time.$errors[0] }}
          </p>
        </fieldset>

        <fieldset class="fieldset">
          <legend class="fieldset-legend">Session</legend>
          <select
            v-model="r$.$value.sessionId"
            class="select w-full"
            :disabled="pending || sessions.length <= 1"
          >
            <option
              v-for="session in sessions"
              :key="session.id"
              :value="session.id"
            >
              {{ session.name }}
            </option>
          </select>
          <p v-if="r$.sessionId.$error" class="fieldset-label text-error">
            {{ r$.sessionId.$errors[0] }}
          </p>
        </fieldset>

        <fieldset class="fieldset md:col-span-2">
          <legend class="fieldset-legend">Audio file</legend>
          <select
            v-model="r$.$value.soundId"
            class="select w-full"
            :disabled="pending"
          >
            <option value="" disabled>Select audio</option>
            <option v-for="file in audioFiles" :key="file.id" :value="file.id">
              {{ file.name }} ({{ file.fileType.replace("_", " ") }})
            </option>
          </select>
          <p v-if="r$.soundId.$error" class="fieldset-label text-error">
            {{ r$.soundId.$errors[0] }}
          </p>
        </fieldset>

        <fieldset class="fieldset md:col-span-2">
          <legend class="fieldset-legend">Days</legend>
          <div class="grid grid-cols-2 gap-3 sm:grid-cols-4 md:grid-cols-7">
            <label
              v-for="day in dayOptions"
              :key="day.value"
              class="label cursor-pointer justify-start gap-2 rounded-box border border-base-300 px-3 py-2"
            >
              <input
                class="checkbox checkbox-sm"
                type="checkbox"
                :checked="form.days.includes(day.value)"
                :disabled="pending"
                @change="
                  updateDay(
                    day.value,
                    ($event.target as HTMLInputElement).checked,
                  )
                "
              />
              <span class="label-text">{{ day.label }}</span>
            </label>
          </div>
          <p v-if="r$.days.$error" class="fieldset-label text-error">
            Select at least one day.
          </p>
        </fieldset>

        <div class="modal-action md:col-span-2 mt-2">
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
