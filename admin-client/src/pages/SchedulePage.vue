<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { Icon } from "@iconify/vue";
import AppConfirmDialog from "../components/app/AppConfirmDialog.vue";
import { AppPageHeader } from "./page-exports";
import CurrentScheduleCard from "../features/schedule/components/CurrentScheduleCard.vue";
import ScheduleItemFormDialog from "../features/schedule/components/ScheduleItemFormDialog.vue";
import ScheduleSessionPicker from "../features/schedule/components/ScheduleSessionPicker.vue";
import ScheduleTable from "../features/schedule/components/ScheduleTable.vue";
import {
  useCreateScheduleItemMutation,
  useCurrentScheduleQuery,
  useDeleteScheduleItemMutation,
  useScheduleItemsQuery,
  useUpdateScheduleItemMutation,
} from "../features/schedule/composables/use-schedule";
import { useScheduleScope } from "../features/schedule/composables/use-schedule-scope";
import { useAudioFilesQuery } from "../features/audio/composables/use-audio";
import { useSessionsQuery } from "../features/sessions/composables/use-sessions";
import { getPrimaryError, getUserFacingError } from "../lib/api/errors";
import type { ScheduleItemVm, ScheduleWritePayload } from "../lib/api/schedule";
import { useAuthStore } from "../stores/auth";
import { useToastStore } from "../stores/toast";

type ScheduleFormPayload = {
  days: number[];
  name: string;
  sessionId: string;
  soundId: string;
  time: string;
};

const authStore = useAuthStore();
const toastStore = useToastStore();

const selectedSessionId = ref<null | string>(null);
const dialogMode = ref<"create" | "edit">("create");
const dialogOpen = ref(false);
const dialogError = ref<null | string>(null);
const dialogItem = ref<null | ScheduleItemVm>(null);
const deleteTarget = ref<null | ScheduleItemVm>(null);

const sessionsQuery = useSessionsQuery();
const currentScheduleQuery = useCurrentScheduleQuery();
const audioFilesQuery = useAudioFilesQuery({
  page: 1,
  size: 200,
  sort: "name",
});
const createScheduleItemMutation = useCreateScheduleItemMutation();
const updateScheduleItemMutation = useUpdateScheduleItemMutation();
const deleteScheduleItemMutation = useDeleteScheduleItemMutation();

const sessions = computed(() => sessionsQuery.data.value?.items ?? []);
const userRole = computed(() => authStore.user.role);
const {
  allowedSessions,
  canSelectAnySession,
  defaultSessionId,
  isSessionAllowed,
} = useScheduleScope(sessions, userRole);

const scheduleQueryParams = computed(() => ({
  include: "session,sound",
  ...(selectedSessionId.value ? { sessionId: selectedSessionId.value } : {}),
}));

const scheduleItemsQuery = useScheduleItemsQuery(scheduleQueryParams);

const scheduleItems = computed(
  () => scheduleItemsQuery.data.value?.items ?? [],
);
const filteredScheduleItems = computed(() => {
  if (authStore.isAdmin) {
    return scheduleItems.value;
  }

  return scheduleItems.value.filter((item) => isSessionAllowed(item.sessionId));
});
const currentSchedule = computed(() => currentScheduleQuery.data.value ?? null);
const availableAudioFiles = computed(
  () => audioFilesQuery.data.value?.items ?? [],
);
const pageError = computed(() => {
  if (!sessionsQuery.error.value && !scheduleItemsQuery.error.value) {
    return null;
  }

  return getUserFacingError(
    sessionsQuery.error.value ?? scheduleItemsQuery.error.value,
  ).detail;
});
const currentScheduleError = computed(() => {
  if (!currentScheduleQuery.error.value) {
    return null;
  }

  return getUserFacingError(currentScheduleQuery.error.value).detail;
});
const formPending = computed(() => {
  return (
    createScheduleItemMutation.isPending.value ||
    updateScheduleItemMutation.isPending.value
  );
});

watch(
  [allowedSessions, canSelectAnySession],
  () => {
    if (!allowedSessions.value.length) {
      selectedSessionId.value = null;
      return;
    }

    if (!canSelectAnySession.value) {
      selectedSessionId.value = defaultSessionId.value;
      return;
    }

    if (
      selectedSessionId.value &&
      !allowedSessions.value.some(
        (session) => session.id === selectedSessionId.value,
      )
    ) {
      selectedSessionId.value = defaultSessionId.value;
    }
  },
  { immediate: true },
);

watch(selectedSessionId, (sessionId, previousSessionId) => {
  if (!sessionId || isSessionAllowed(sessionId)) {
    return;
  }

  selectedSessionId.value = defaultSessionId.value;

  if (sessionId !== previousSessionId) {
    toastStore.enqueue({
      detail: "Your role is limited to its assigned session schedule.",
      title: "Session Restricted",
      tone: "info",
    });
  }
});

function openCreateDialog() {
  if (!allowedSessions.value.length) {
    toastStore.enqueue({
      detail: "No permitted session is available for this account.",
      title: "Session Required",
      tone: "error",
    });
    return;
  }

  dialogMode.value = "create";
  dialogItem.value = null;
  dialogError.value = null;
  dialogOpen.value = true;
}

function openEditDialog(item: ScheduleItemVm) {
  dialogMode.value = "edit";
  dialogItem.value = item;
  dialogError.value = null;
  dialogOpen.value = true;
}

function closeDialog() {
  dialogOpen.value = false;
  dialogError.value = null;
}

function requestDelete(item: ScheduleItemVm) {
  deleteTarget.value = item;
}

function cancelDelete() {
  deleteTarget.value = null;
}

async function handleSaveScheduleItem(payload: ScheduleFormPayload) {
  dialogError.value = null;

  try {
    if (dialogMode.value === "create") {
      await createScheduleItemMutation.mutateAsync(
        toScheduleWritePayload(payload),
      );
      toastStore.enqueue({
        detail: `${payload.name} was created successfully.`,
        title: "Schedule Item Created",
        tone: "success",
      });
    } else if (dialogItem.value) {
      await updateScheduleItemMutation.mutateAsync({
        id: dialogItem.value.id,
        payload: toScheduleWritePayload(payload),
      });
      toastStore.enqueue({
        detail: `${payload.name} was updated successfully.`,
        title: "Schedule Item Updated",
        tone: "success",
      });
    }

    dialogOpen.value = false;
    dialogItem.value = null;
  } catch (error) {
    dialogError.value = toScheduleFormError(error);
  }
}

async function confirmDelete() {
  if (!deleteTarget.value) {
    return;
  }

  try {
    await deleteScheduleItemMutation.mutateAsync(deleteTarget.value.id);
    toastStore.enqueue({
      detail: `${deleteTarget.value.name} was deleted successfully.`,
      title: "Schedule Item Deleted",
      tone: "success",
    });
    deleteTarget.value = null;
  } catch (error) {
    toastStore.enqueue({
      detail: getUserFacingError(error).detail,
      title: "Delete Failed",
      tone: "error",
    });
  }
}

function onSessionChange(value: string) {
  selectedSessionId.value = value || null;
}

function toScheduleFormError(error: unknown) {
  const primaryError = getPrimaryError(error);
  const detail = primaryError.detail.toLowerCase();

  if (detail.includes("invalid time format")) {
    return "Time must use the HH:MM format.";
  }

  if (detail.includes("days must be between")) {
    return "Select valid schedule days before saving.";
  }

  if (detail.includes("name, time, soundid, and days are required")) {
    return "Name, time, audio file, and at least one day are required.";
  }

  if (detail.includes("failed to create schedule item")) {
    return "The schedule item could not be created. Check the selected session and audio, then try again.";
  }

  if (detail.includes("failed to update schedule item")) {
    return "The schedule item could not be updated. Refresh the page and try again.";
  }

  return getUserFacingError(error).detail;
}

function toScheduleWritePayload(
  payload: ScheduleFormPayload,
): ScheduleWritePayload {
  return {
    data: {
      attributes: {
        days: payload.days,
        name: payload.name,
        time: payload.time,
      },
      relationships: {
        session: {
          data: {
            id: payload.sessionId,
            type: "sessions",
          },
        },
        sound: {
          data: {
            id: payload.soundId,
            type: "audio-files",
          },
        },
      },
      type: "schedule-items",
    },
  };
}
</script>

<template>
  <section class="page-stack">
    <AppPageHeader
      eyebrow="Operations"
      title="Schedule"
      description="Manage the live bell timetable by session, confirm what is active right now, and keep each bell linked to the correct audio file."
    />

    <div v-if="pageError" class="alert alert-error text-sm" role="alert">
      {{ pageError }}
    </div>

    <div class="grid gap-6 xl:grid-cols-[minmax(0,1.25fr)_minmax(0,0.75fr)]">
      <article class="card border border-base-300 bg-base-100 shadow-sm">
        <div class="card-body gap-5">
          <div
            class="flex flex-col gap-4 xl:flex-row xl:items-end xl:justify-between"
          >
            <ScheduleSessionPicker
              :sessions="allowedSessions"
              :value="selectedSessionId"
              :can-select-any-session="canSelectAnySession"
              :loading="
                sessionsQuery.isLoading.value || sessionsQuery.isFetching.value
              "
              @update:value="onSessionChange"
            />

            <button
              class="btn btn-primary"
              type="button"
              @click="openCreateDialog"
            >
              <Icon icon="solar:calendar-add-bold-duotone" class="text-lg" />
              Create Schedule Item
            </button>
          </div>

          <ScheduleTable
            :items="filteredScheduleItems"
            :loading="
              scheduleItemsQuery.isLoading.value ||
              scheduleItemsQuery.isFetching.value
            "
            :deleting-id="deleteScheduleItemMutation.variables.value ?? null"
            @edit="openEditDialog"
            @delete="requestDelete"
          />
        </div>
      </article>

      <div class="space-y-6">
        <CurrentScheduleCard
          :schedule="currentSchedule"
          :loading="
            currentScheduleQuery.isLoading.value ||
            currentScheduleQuery.isFetching.value
          "
          :error="currentScheduleError"
        />
      </div>
    </div>

    <ScheduleItemFormDialog
      :open="dialogOpen"
      :mode="dialogMode"
      :item="dialogItem"
      :sessions="allowedSessions"
      :audio-files="availableAudioFiles"
      :pending="formPending"
      :error="dialogError"
      @close="closeDialog"
      @save="handleSaveScheduleItem"
    />

    <AppConfirmDialog
      :open="Boolean(deleteTarget)"
      title="Delete Schedule Item"
      :description="
        deleteTarget
          ? `Delete ${deleteTarget.name}. This bell time will be removed from the schedule permanently.`
          : ''
      "
      confirm-label="Delete Item"
      tone="error"
      :pending="deleteScheduleItemMutation.isPending.value"
      @close="cancelDelete"
      @confirm="confirmDelete"
    />
  </section>
</template>
