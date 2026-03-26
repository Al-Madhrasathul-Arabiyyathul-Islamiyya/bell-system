<script setup lang="ts">
import { computed, ref } from "vue";
import { Icon } from "@iconify/vue";
import AppConfirmDialog from "../components/app/AppConfirmDialog.vue";
import { AppPageHeader, AppStatCard } from "./page-exports";
import SessionFormDialog from "../features/sessions/components/SessionFormDialog.vue";
import SessionsTable from "../features/sessions/components/SessionsTable.vue";
import {
  useCreateSessionMutation,
  useCurrentSessionQuery,
  useDeleteSessionMutation,
  useSessionsQuery,
  useUpdateSessionMutation,
} from "../features/sessions/composables/use-sessions";
import { getPrimaryError, getUserFacingError } from "../lib/api/errors";
import type { SessionVm, SessionWritePayload } from "../lib/api/sessions";
import { useToastStore } from "../stores/toast";

type SessionFormPayload = {
  endTime: string;
  name: string;
  startTime: string;
};

const toastStore = useToastStore();

const dialogMode = ref<"create" | "edit">("create");
const dialogOpen = ref(false);
const dialogError = ref<null | string>(null);
const dialogSession = ref<null | SessionVm>(null);

const deleteTarget = ref<null | SessionVm>(null);

const sessionsQuery = useSessionsQuery();
const currentSessionQuery = useCurrentSessionQuery();
const createSessionMutation = useCreateSessionMutation();
const updateSessionMutation = useUpdateSessionMutation();
const deleteSessionMutation = useDeleteSessionMutation();

const sessions = computed(() => sessionsQuery.data.value?.items ?? []);
const totalSessions = computed(() => sessions.value.length);
const sessionsLoading = computed(() => {
  return sessionsQuery.isLoading.value || sessionsQuery.isFetching.value;
});
const currentSession = computed(() => {
  if (!currentSessionQuery.error.value) {
    return currentSessionQuery.data.value ?? null;
  }

  const error = getPrimaryError(currentSessionQuery.error.value);

  if (Number(error.status) === 404) {
    return null;
  }

  return null;
});
const currentSessionLabel = computed(() => {
  if (!currentSession.value) {
    return "No Active Session";
  }

  return currentSession.value.name;
});
const currentSessionWindow = computed(() => {
  if (!currentSession.value) {
    return "No session currently covers the active time window.";
  }

  return `${formatTime(currentSession.value.startTime)} - ${formatTime(currentSession.value.endTime)}`;
});
const pageError = computed(() => {
  if (!sessionsQuery.error.value) {
    return null;
  }

  return getUserFacingError(sessionsQuery.error.value).detail;
});
const currentSessionError = computed(() => {
  if (!currentSessionQuery.error.value) {
    return null;
  }

  const error = getPrimaryError(currentSessionQuery.error.value);

  if (Number(error.status) === 404) {
    return null;
  }

  return getUserFacingError(currentSessionQuery.error.value).detail;
});
const currentMutationPending = computed(() => {
  return (
    createSessionMutation.isPending.value ||
    updateSessionMutation.isPending.value
  );
});

function openCreateDialog() {
  dialogMode.value = "create";
  dialogSession.value = null;
  dialogError.value = null;
  dialogOpen.value = true;
}

function openEditDialog(session: SessionVm) {
  dialogMode.value = "edit";
  dialogSession.value = session;
  dialogError.value = null;
  dialogOpen.value = true;
}

function closeDialog() {
  dialogOpen.value = false;
  dialogError.value = null;
}

async function handleSaveSession(payload: SessionFormPayload) {
  dialogError.value = null;

  try {
    if (dialogMode.value === "create") {
      await createSessionMutation.mutateAsync(toSessionWritePayload(payload));
      toastStore.enqueue({
        detail: `${payload.name} was created successfully.`,
        title: "Session Created",
        tone: "success",
      });
    } else if (dialogSession.value) {
      await updateSessionMutation.mutateAsync({
        id: dialogSession.value.id,
        payload: toSessionWritePayload(payload),
      });
      toastStore.enqueue({
        detail: `${payload.name} was updated successfully.`,
        title: "Session Updated",
        tone: "success",
      });
    }

    dialogOpen.value = false;
    dialogSession.value = null;
  } catch (error) {
    dialogError.value = toSessionFormError(error);
  }
}

function requestDelete(session: SessionVm) {
  deleteTarget.value = session;
}

function cancelDelete() {
  deleteTarget.value = null;
}

async function confirmDelete() {
  if (!deleteTarget.value) {
    return;
  }

  try {
    await deleteSessionMutation.mutateAsync(deleteTarget.value.id);
    toastStore.enqueue({
      detail: `${deleteTarget.value.name} was deleted successfully.`,
      title: "Session Deleted",
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

function formatTime(value: string) {
  return value.slice(0, 5);
}

function toSessionFormError(error: unknown) {
  const primaryError = getPrimaryError(error);

  if (
    Number(primaryError.status) === 409 ||
    primaryError.detail.toLowerCase().includes("conflict")
  ) {
    return "This session conflicts with an existing session time range.";
  }

  if (primaryError.detail.toLowerCase().includes("starttime")) {
    return "Start time must use the HH:MM format.";
  }

  if (primaryError.detail.toLowerCase().includes("endtime")) {
    return "End time must use the HH:MM format.";
  }

  return getUserFacingError(error).detail;
}

function toSessionWritePayload(
  payload: SessionFormPayload,
): SessionWritePayload {
  return {
    data: {
      attributes: {
        endTime: payload.endTime,
        name: payload.name,
        startTime: payload.startTime,
      },
      type: "sessions",
    },
  };
}
</script>

<template>
  <section class="page-stack">
    <AppPageHeader
      eyebrow="Scheduling"
      title="Sessions"
      description="Manage bell schedule windows, review the currently active session, and keep session time ranges aligned with backend schedule rules."
    />

    <div class="grid gap-4 xl:grid-cols-3">
      <AppStatCard
        title="Total Sessions"
        icon="solar:calendar-date-bold-duotone"
        :value="String(totalSessions)"
        description="All configured sessions returned by the backend."
      />
      <AppStatCard
        title="Current Session"
        icon="solar:clock-circle-bold-duotone"
        :value="currentSessionLabel"
        :description="currentSessionWindow"
      />
      <article class="card border border-base-300 bg-base-100 shadow-sm">
        <div class="card-body gap-3">
          <p
            class="text-xs font-semibold uppercase tracking-[0.24em] text-primary"
          >
            Session Rules
          </p>
          <div class="space-y-2">
            <p class="text-base font-semibold text-base-content">
              Time windows should not overlap.
            </p>
            <p class="text-sm leading-7 text-base-content/70">
              The backend rejects conflicting session windows, so create and
              edit flows surface those conflicts inline before you continue.
            </p>
          </div>
        </div>
      </article>
    </div>

    <div
      v-if="pageError || currentSessionError"
      class="alert alert-error text-sm"
      role="alert"
    >
      {{ pageError || currentSessionError }}
    </div>

    <article class="card border border-base-300 bg-base-100 shadow-sm">
      <div class="card-body gap-5">
        <div
          class="flex flex-col gap-4 xl:flex-row xl:items-end xl:justify-between"
        >
          <div class="space-y-1">
            <h2 class="font-display text-2xl font-semibold text-base-content">
              Session Windows
            </h2>
            <p class="text-sm leading-7 text-base-content/70">
              Sessions define the active time windows used by current-session
              lookups and schedule grouping.
            </p>
          </div>

          <button
            class="btn btn-primary"
            type="button"
            @click="openCreateDialog"
          >
            <Icon icon="solar:calendar-add-bold-duotone" class="text-lg" />
            Create Session
          </button>
        </div>

        <SessionsTable
          :sessions="sessions"
          :loading="sessionsLoading"
          :deleting-id="deleteSessionMutation.variables.value ?? null"
          @edit="openEditDialog"
          @delete="requestDelete"
        />
      </div>
    </article>

    <SessionFormDialog
      :open="dialogOpen"
      :mode="dialogMode"
      :session="dialogSession"
      :pending="currentMutationPending"
      :error="dialogError"
      @close="closeDialog"
      @save="handleSaveSession"
    />

    <AppConfirmDialog
      :open="Boolean(deleteTarget)"
      title="Delete Session"
      :description="
        deleteTarget
          ? `Delete ${deleteTarget.name}. This removes the session window permanently.`
          : ''
      "
      confirm-label="Delete Session"
      tone="error"
      :pending="deleteSessionMutation.isPending.value"
      @close="cancelDelete"
      @confirm="confirmDelete"
    />
  </section>
</template>
