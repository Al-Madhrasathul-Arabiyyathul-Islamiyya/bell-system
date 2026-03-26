<script setup lang="ts">
import { reactive, watch } from "vue";
import { useRegleSchema } from "@regle/schemas";
import { z } from "zod";
import type { AudioFileType, AudioFileVm } from "../../../lib/api/audio";

type AudioMetadataPayload = {
  fileType: AudioFileType;
  name: string;
};

const props = withDefaults(
  defineProps<{
    error?: null | string;
    open: boolean;
    pending?: boolean;
    file?: null | AudioFileVm;
  }>(),
  {
    error: null,
    pending: false,
    file: null,
  },
);

const emit = defineEmits<{
  close: [];
  save: [payload: AudioMetadataPayload];
}>();

const form = reactive({
  fileType: "bell" as AudioFileType,
  name: "",
});

const schema = z.object({
  fileType: z.enum(["anthem", "bell", "other", "school_song"]),
  name: z.string().trim().min(1, "Audio name is required."),
});

const { r$ } = useRegleSchema(form, schema);

watch(
  () => [props.open, props.file?.id] as const,
  () => {
    if (!props.open) {
      return;
    }

    form.name = props.file?.name ?? "";
    form.fileType = props.file?.fileType ?? "bell";
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
    fileType: form.fileType,
    name: form.name.trim(),
  });
}
</script>

<template>
  <div v-if="open" class="modal modal-open">
    <div class="modal-box max-w-lg">
      <button
        class="btn btn-circle btn-ghost btn-sm absolute top-4 right-4"
        type="button"
        :disabled="pending"
        aria-label="Close audio metadata dialog"
        @click="emit('close')"
      >
        ×
      </button>

      <h2 class="font-display pr-8 text-2xl font-semibold text-base-content">
        Edit Audio Metadata
      </h2>
      <p class="mt-2 text-sm leading-7 text-base-content/70">
        Update the display name or type without re-uploading the underlying
        audio file.
      </p>

      <div v-if="error" class="alert alert-error mt-5 text-sm" role="alert">
        {{ error }}
      </div>

      <form class="mt-6 space-y-5" @submit.prevent="handleSubmit">
        <fieldset class="fieldset">
          <legend class="fieldset-legend">Audio name</legend>
          <input
            v-model="r$.$value.name"
            class="input w-full"
            :disabled="pending"
            type="text"
            placeholder="bell.mp3"
          />
          <p v-if="r$.name.$error" class="fieldset-label text-error">
            {{ r$.name.$errors[0] }}
          </p>
        </fieldset>

        <fieldset class="fieldset">
          <legend class="fieldset-legend">Audio type</legend>
          <select
            v-model="r$.$value.fileType"
            class="select w-full"
            :disabled="pending"
          >
            <option value="bell">Bell</option>
            <option value="anthem">Anthem</option>
            <option value="school_song">School Song</option>
            <option value="other">Other</option>
          </select>
          <p v-if="r$.fileType.$error" class="fieldset-label text-error">
            {{ r$.fileType.$errors[0] }}
          </p>
        </fieldset>

        <div class="modal-action mt-8">
          <button
            class="btn btn-ghost"
            type="button"
            :disabled="pending"
            @click="emit('close')"
          >
            Cancel
          </button>
          <button class="btn btn-primary" type="submit" :disabled="pending">
            {{ pending ? "Saving..." : "Save Changes" }}
          </button>
        </div>
      </form>
    </div>
    <div class="modal-backdrop" @click="emit('close')" />
  </div>
</template>
