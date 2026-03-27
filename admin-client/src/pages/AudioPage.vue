<script setup lang="ts">
import { computed, ref } from "vue";
import { Icon } from "@iconify/vue";
import { useStorage } from "@vueuse/core";
import AppConfirmDialog from "../components/app/AppConfirmDialog.vue";
import { AppPageHeader, AppStatCard } from "./page-exports";
import AudioFilesTable from "../features/audio/components/AudioFilesTable.vue";
import AudioMetadataDialog from "../features/audio/components/AudioMetadataDialog.vue";
import AudioPreviewPlayer from "../features/audio/components/AudioPreviewPlayer.vue";
import AudioUploadDialog from "../features/audio/components/AudioUploadDialog.vue";
import {
  useAudioChecksumsQuery,
  useAudioFilesQuery,
  useDeleteAudioMutation,
  useUpdateAudioMutation,
  useUploadAudioMutation,
} from "../features/audio/composables/use-audio";
import { getPrimaryError, getUserFacingError } from "../lib/api/errors";
import type {
  AudioFileType,
  AudioFileVm,
  AudioUpdatePayload,
} from "../lib/api/audio";
import { useToastStore } from "../stores/toast";

type AudioFilter = "all" | AudioFileType;
type AudioMetadataPayload = {
  fileType: AudioFileType;
  name: string;
};
type AudioUploadPayload = {
  file: File;
  fileType: AudioFileType;
  name: string;
};

const toastStore = useToastStore();

const currentPage = ref(1);
const fileTypeFilter = useStorage<AudioFilter>(
  "bell-admin-audio-file-type-filter",
  "all",
);
const pageSize = useStorage<number>("bell-admin-audio-page-size", 20);

const uploadDialogOpen = ref(false);
const uploadDialogError = ref<null | string>(null);

const metadataDialogOpen = ref(false);
const metadataDialogError = ref<null | string>(null);
const metadataTarget = ref<null | AudioFileVm>(null);

const deleteTarget = ref<null | AudioFileVm>(null);
const previewTarget = ref<null | AudioFileVm>(null);

const queryParams = computed(() => ({
  fileType: fileTypeFilter.value === "all" ? undefined : fileTypeFilter.value,
  page: currentPage.value,
  size: pageSize.value,
  sort: "name",
}));

const audioFilesQuery = useAudioFilesQuery(queryParams);
const audioChecksumsQuery = useAudioChecksumsQuery();
const uploadAudioMutation = useUploadAudioMutation();
const updateAudioMutation = useUpdateAudioMutation();
const deleteAudioMutation = useDeleteAudioMutation();

const audioFiles = computed(() => audioFilesQuery.data.value?.items ?? []);
const totalFiles = computed(() =>
  readMetaTotal(audioFilesQuery.data.value?.meta),
);
const totalPages = computed(() =>
  readMetaPages(audioFilesQuery.data.value?.meta),
);
const previewLabel = computed(() => {
  if (fileTypeFilter.value === "all") {
    return "All Types";
  }

  if (fileTypeFilter.value === "school_song") {
    return "School Song";
  }

  return (
    fileTypeFilter.value.charAt(0).toUpperCase() +
    fileTypeFilter.value.slice(1).replace("_", " ")
  );
});
const checksumTotal = computed(() => {
  const data = audioChecksumsQuery.data.value;

  if (!data) {
    return "Unavailable";
  }

  return String(data.items.length);
});
const pageError = computed(() => {
  if (!audioFilesQuery.error.value) {
    return null;
  }

  return getUserFacingError(audioFilesQuery.error.value).detail;
});

function openUploadDialog() {
  uploadDialogError.value = null;
  uploadDialogOpen.value = true;
}

function closeUploadDialog() {
  uploadDialogError.value = null;
  uploadDialogOpen.value = false;
}

function openMetadataDialog(file: AudioFileVm) {
  metadataTarget.value = file;
  metadataDialogError.value = null;
  metadataDialogOpen.value = true;
}

function closeMetadataDialog() {
  metadataDialogError.value = null;
  metadataDialogOpen.value = false;
  metadataTarget.value = null;
}

function selectPreview(file: AudioFileVm) {
  previewTarget.value = file;
}

function requestDelete(file: AudioFileVm) {
  deleteTarget.value = file;
}

function cancelDelete() {
  deleteTarget.value = null;
}

async function handleUpload(payload: AudioUploadPayload) {
  uploadDialogError.value = null;

  try {
    const formData = new FormData();
    formData.set("file", payload.file);
    formData.set("name", payload.name);
    formData.set("type", payload.fileType);

    await uploadAudioMutation.mutateAsync(formData);
    toastStore.enqueue({
      detail: `${payload.name} was uploaded successfully.`,
      title: "Audio Uploaded",
      tone: "success",
    });
    uploadDialogOpen.value = false;
  } catch (error) {
    uploadDialogError.value = toAudioFormError(error);
  }
}

async function handleMetadataSave(payload: AudioMetadataPayload) {
  metadataDialogError.value = null;

  if (!metadataTarget.value) {
    return;
  }

  try {
    await updateAudioMutation.mutateAsync({
      id: metadataTarget.value.id,
      payload: toAudioUpdatePayload(payload),
    });
    toastStore.enqueue({
      detail: `${payload.name} was updated successfully.`,
      title: "Audio Updated",
      tone: "success",
    });

    if (previewTarget.value?.id === metadataTarget.value.id) {
      previewTarget.value = {
        ...previewTarget.value,
        fileType: payload.fileType,
        name: payload.name,
      };
    }

    metadataDialogOpen.value = false;
    metadataTarget.value = null;
  } catch (error) {
    metadataDialogError.value = toAudioFormError(error);
  }
}

async function confirmDelete() {
  if (!deleteTarget.value) {
    return;
  }

  try {
    await deleteAudioMutation.mutateAsync(deleteTarget.value.id);
    toastStore.enqueue({
      detail: `${deleteTarget.value.name} was deleted successfully.`,
      title: "Audio Deleted",
      tone: "success",
    });

    if (previewTarget.value?.id === deleteTarget.value.id) {
      previewTarget.value = null;
    }

    deleteTarget.value = null;
  } catch (error) {
    toastStore.enqueue({
      detail: getUserFacingError(error).detail,
      title: "Delete Failed",
      tone: "error",
    });
  }
}

function onFileTypeFilterChange(value: AudioFilter) {
  fileTypeFilter.value = value;
  currentPage.value = 1;
}

function onPageSizeChange(value: number) {
  pageSize.value = value;
  currentPage.value = 1;
}

function nextPage() {
  if (currentPage.value < totalPages.value) {
    currentPage.value += 1;
  }
}

function previousPage() {
  if (currentPage.value > 1) {
    currentPage.value -= 1;
  }
}

function readMetaTotal(meta: Record<string, unknown> | undefined) {
  const value = meta?.total;

  return typeof value === "number" ? value : 0;
}

function readMetaPages(meta: Record<string, unknown> | undefined) {
  const value = meta?.page;

  if (
    value &&
    typeof value === "object" &&
    "pages" in value &&
    typeof value.pages === "number"
  ) {
    return Math.max(value.pages, 1);
  }

  return 1;
}

function toAudioFormError(error: unknown) {
  const primaryError = getPrimaryError(error);
  const detail = primaryError.detail.toLowerCase();

  if (detail.includes("file is required")) {
    return "Select an audio file before uploading.";
  }

  if (detail.includes("invalid file type")) {
    return "Choose one of the supported audio categories before continuing.";
  }

  return getUserFacingError(error).detail;
}

function toAudioUpdatePayload(
  payload: AudioMetadataPayload,
): AudioUpdatePayload {
  return {
    data: {
      attributes: {
        fileType: payload.fileType,
        name: payload.name,
      },
      type: "audio-files",
    },
  };
}
</script>

<template>
  <section class="page-stack">
    <AppPageHeader
      eyebrow="Media"
      title="Audio Library"
      description="Manage uploaded bell sounds and other system audio, preview files directly from the binary route, and keep metadata organized by file type."
    />

    <div class="grid gap-4 xl:grid-cols-3">
      <AppStatCard
        title="Total Audio Files"
        icon="solar:music-library-2-bold-duotone"
        :value="String(totalFiles)"
        description="Audio metadata records returned by the paginated backend list."
      />
      <AppStatCard
        title="Current Filter"
        icon="solar:filters-bold-duotone"
        :value="previewLabel"
        description="Active audio type filter persisted for this admin client."
      />
      <AppStatCard
        title="Checksums Indexed"
        icon="solar:shield-check-bold-duotone"
        :value="checksumTotal"
        description="Checksum entries available from the audio checksums sync endpoint."
      />
    </div>

    <div v-if="pageError" class="alert alert-error text-sm" role="alert">
      {{ pageError }}
    </div>

    <div class="grid gap-6 xl:grid-cols-[minmax(0,1.25fr)_minmax(0,0.75fr)]">
      <article class="card border border-base-300 bg-base-100 shadow-sm">
        <div class="card-body gap-5">
          <div
            class="flex flex-col gap-4 xl:flex-row xl:items-end xl:justify-between"
          >
            <div class="grid gap-4 md:grid-cols-2">
              <fieldset class="fieldset">
                <legend class="fieldset-legend">Audio type</legend>
                <select
                  class="select w-full"
                  :value="fileTypeFilter"
                  @change="
                    onFileTypeFilterChange(
                      ($event.target as HTMLSelectElement).value as AudioFilter,
                    )
                  "
                >
                  <option value="all">All Types</option>
                  <option value="bell">Bell</option>
                  <option value="anthem">Anthem</option>
                  <option value="school_song">School Song</option>
                  <option value="other">Other</option>
                </select>
              </fieldset>

              <fieldset class="fieldset">
                <legend class="fieldset-legend">Page Size</legend>
                <select
                  class="select w-full"
                  :value="String(pageSize)"
                  @change="
                    onPageSizeChange(
                      Number(($event.target as HTMLSelectElement).value),
                    )
                  "
                >
                  <option value="10">10 per page</option>
                  <option value="20">20 per page</option>
                  <option value="50">50 per page</option>
                </select>
              </fieldset>
            </div>

            <button
              class="btn btn-primary"
              type="button"
              @click="openUploadDialog"
            >
              <Icon icon="solar:upload-bold-duotone" class="text-lg" />
              Upload Audio
            </button>
          </div>

          <AudioFilesTable
            :files="audioFiles"
            :loading="
              audioFilesQuery.isLoading.value ||
              audioFilesQuery.isFetching.value
            "
            :deleting-id="deleteAudioMutation.variables.value ?? null"
            :preview-id="previewTarget?.id ?? null"
            @preview="selectPreview"
            @edit="openMetadataDialog"
            @delete="requestDelete"
          />

          <div
            class="flex flex-col gap-3 border-t border-base-300 pt-4 sm:flex-row sm:items-center sm:justify-between"
          >
            <p class="text-sm text-base-content/65">
              Showing {{ audioFiles.length }} of {{ totalFiles }} audio files.
            </p>
            <div class="join">
              <button
                class="btn join-item"
                type="button"
                :disabled="currentPage <= 1 || audioFilesQuery.isFetching.value"
                @click="previousPage"
              >
                Previous
              </button>
              <button class="btn join-item pointer-events-none" type="button">
                {{ currentPage }}
              </button>
              <button
                class="btn join-item"
                type="button"
                :disabled="
                  currentPage >= totalPages || audioFilesQuery.isFetching.value
                "
                @click="nextPage"
              >
                Next
              </button>
            </div>
          </div>
        </div>
      </article>

      <div class="space-y-6">
        <AudioPreviewPlayer :file="previewTarget" />
      </div>
    </div>

    <AudioUploadDialog
      :open="uploadDialogOpen"
      :pending="uploadAudioMutation.isPending.value"
      :error="uploadDialogError"
      @close="closeUploadDialog"
      @save="handleUpload"
    />

    <AudioMetadataDialog
      :open="metadataDialogOpen"
      :pending="updateAudioMutation.isPending.value"
      :error="metadataDialogError"
      :file="metadataTarget"
      @close="closeMetadataDialog"
      @save="handleMetadataSave"
    />

    <AppConfirmDialog
      :open="Boolean(deleteTarget)"
      title="Delete Audio File"
      :description="
        deleteTarget
          ? `Delete ${deleteTarget.name}. This removes both the metadata and the stored audio content.`
          : ''
      "
      confirm-label="Delete Audio"
      tone="error"
      :pending="deleteAudioMutation.isPending.value"
      @close="cancelDelete"
      @confirm="confirmDelete"
    />
  </section>
</template>
