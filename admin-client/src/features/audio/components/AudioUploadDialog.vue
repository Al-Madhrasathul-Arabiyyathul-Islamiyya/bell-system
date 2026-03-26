<script setup lang="ts">
import { computed, reactive, ref } from "vue";
import { useRegleSchema } from "@regle/schemas";
import { z } from "zod";
import type { AudioFileType } from "../../../lib/api/audio";

type AudioUploadPayload = {
  file: File;
  fileType: AudioFileType;
  name: string;
};

withDefaults(
  defineProps<{
    error?: null | string;
    open: boolean;
    pending?: boolean;
  }>(),
  {
    error: null,
    pending: false,
  },
);

const emit = defineEmits<{
  close: [];
  save: [payload: AudioUploadPayload];
}>();

const file = ref<File | null>(null);
const form = reactive({
  fileType: "bell" as AudioFileType,
  name: "",
});

const schema = z.object({
  fileType: z.enum(["anthem", "bell", "other", "school_song"]),
  name: z.string().trim().min(1, "Audio name is required."),
});

const { r$ } = useRegleSchema(form, schema);

const selectedFileName = computed(() => {
  return file.value?.name ?? "No file selected";
});

async function handleSubmit() {
  const result = await r$.$validate();

  if (!result.valid) {
    return;
  }

  if (!file.value) {
    return;
  }

  emit("save", {
    file: file.value,
    fileType: form.fileType,
    name: form.name.trim(),
  });
}

function handleFileChange(event: Event) {
  const target = event.target as HTMLInputElement;
  file.value = target.files?.[0] ?? null;

  if (!form.name && file.value) {
    form.name = file.value.name;
  }
}

function reset() {
  file.value = null;
  form.name = "";
  form.fileType = "bell";
  r$.$reset();
}

function handleClose() {
  reset();
  emit("close");
}
</script>

<template>
  <div v-if="open" class="modal modal-open">
    <div class="modal-box max-w-lg">
      <button
        class="btn btn-circle btn-ghost btn-sm absolute top-4 right-4"
        type="button"
        :disabled="pending"
        aria-label="Close audio upload dialog"
        @click="handleClose"
      >
        ×
      </button>

      <h2 class="font-display pr-8 text-2xl font-semibold text-base-content">
        Upload Audio
      </h2>
      <p class="mt-2 text-sm leading-7 text-base-content/70">
        Upload an audio file using the backend multipart route and categorize it
        for schedule selection.
      </p>

      <div v-if="error" class="alert alert-error mt-5 text-sm" role="alert">
        {{ error }}
      </div>

      <form class="mt-6 space-y-5" @submit.prevent="handleSubmit">
        <fieldset class="fieldset">
          <legend class="fieldset-legend">Audio file</legend>
          <input
            class="file-input w-full"
            :disabled="pending"
            type="file"
            accept="audio/*"
            @change="handleFileChange"
          />
          <p class="fieldset-label">{{ selectedFileName }}</p>
          <p v-if="!file" class="fieldset-label text-error">
            Please select an audio file to upload.
          </p>
        </fieldset>

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
            @click="handleClose"
          >
            Cancel
          </button>
          <button class="btn btn-primary" type="submit" :disabled="pending">
            {{ pending ? "Uploading..." : "Upload Audio" }}
          </button>
        </div>
      </form>
    </div>
    <div class="modal-backdrop" @click="handleClose" />
  </div>
</template>
