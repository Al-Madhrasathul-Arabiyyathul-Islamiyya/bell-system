<script setup lang="ts">
import { computed, ref } from "vue";
import { AppPageHeader, AppStatCard } from "./page-exports";
import AppConfirmDialog from "../components/app/AppConfirmDialog.vue";
import AppSystemStateToggle from "../components/app/AppSystemStateToggle.vue";
import CurrentScheduleCard from "../features/schedule/components/CurrentScheduleCard.vue";
import DashboardRealtimeCard from "../features/dashboard/components/DashboardRealtimeCard.vue";
import DashboardUpcomingList from "../features/dashboard/components/DashboardUpcomingList.vue";
import { useCurrentScheduleSummary } from "../features/dashboard/composables/use-current-schedule-summary";
import { useCurrentScheduleQuery } from "../features/schedule/composables/use-schedule";
import { useSystemControls } from "../features/system/composables/use-system-controls";
import { useAuthStore } from "../stores/auth";
import { useRealtimeStore } from "../stores/realtime";

const authStore = useAuthStore();
const realtimeStore = useRealtimeStore();
const showCancelDialog = ref(false);

const currentScheduleQuery = useCurrentScheduleQuery();
const { nextItem, upcomingItems } = useCurrentScheduleSummary(
  computed(() => currentScheduleQuery.data.value ?? null),
);
const {
  cancelNextBell,
  cancelNextBellMutation,
  systemError,
  systemState,
  toggleSystemState,
  updateSystemStateMutation,
} = useSystemControls();

const limitedUpcomingItems = computed(() => upcomingItems.value.slice(0, 4));
const showConnectedClients = computed(() => authStore.isAdmin);

function requestCancelNextBell() {
  showCancelDialog.value = true;
}

function closeCancelDialog() {
  showCancelDialog.value = false;
}

async function handleCancelNextBell() {
  try {
    await cancelNextBell();
  } catch {
    // Toast feedback is already handled in the shared system controls composable.
  } finally {
    showCancelDialog.value = false;
  }
}
</script>

<template>
  <section class="page-stack">
    <AppPageHeader
      eyebrow="Overview"
      title="Dashboard"
      description="Monitor the live bell state, the current schedule snapshot, upcoming bells, and realtime connectivity from one operational surface."
    />

    <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
      <AppStatCard
        title="System State"
        :value="
          systemState?.state === 'paused'
            ? 'Paused'
            : systemState
              ? 'Active'
              : 'Loading'
        "
        description="Current scheduler state from the backend."
      />
      <AppStatCard
        title="Current Session"
        :value="
          currentScheduleQuery.data.value?.session?.name ?? 'No Active Session'
        "
        :description="
          currentScheduleQuery.data.value?.session
            ? `${currentScheduleQuery.data.value.session.startTime?.slice(0, 5)} - ${currentScheduleQuery.data.value.session.endTime?.slice(0, 5)}`
            : 'No session is active at the current time.'
        "
      />
      <AppStatCard
        title="Next Bell"
        :value="nextItem?.name ?? 'None Queued'"
        :description="
          nextItem
            ? `${nextItem.time} • ${nextItem.sound?.name ?? 'No audio linked'}`
            : 'No pending bell is queued right now.'
        "
      />
    </div>

    <div v-if="systemError" class="alert alert-error text-sm" role="alert">
      {{ systemError }}
    </div>

    <div class="grid gap-6 xl:grid-cols-[minmax(0,1.15fr)_minmax(0,0.85fr)]">
      <CurrentScheduleCard
        :schedule="currentScheduleQuery.data.value ?? null"
        :loading="currentScheduleQuery.isLoading.value"
        :error="
          currentScheduleQuery.error.value
            ? 'The live schedule could not be loaded right now.'
            : null
        "
      />

      <article class="card border border-base-300 bg-base-100 shadow-sm">
        <div class="card-body gap-5">
          <div class="space-y-2">
            <p
              class="text-xs font-semibold uppercase tracking-[0.24em] text-primary"
            >
              Controls
            </p>
            <h2 class="font-display text-3xl font-semibold text-base-content">
              System Actions
            </h2>
            <p class="text-sm leading-7 text-base-content/70">
              Pause or resume bell playback quickly, and cancel the next pending
              bell when the live plan needs intervention.
            </p>
          </div>

          <div class="grid gap-4 md:grid-cols-2">
            <div class="rounded-box bg-base-200 p-4">
              <p class="text-sm text-base-content/65">Current state</p>
              <p
                class="mt-2 text-2xl font-semibold capitalize text-base-content"
              >
                {{ systemState?.state ?? "Loading" }}
              </p>
            </div>
            <div class="rounded-box bg-base-200 p-4">
              <p class="text-sm text-base-content/65">Upcoming bells</p>
              <p class="mt-2 text-2xl font-semibold text-base-content">
                {{ upcomingItems.length }}
              </p>
            </div>
          </div>

          <div class="flex flex-wrap gap-3">
            <AppSystemStateToggle
              :state="systemState?.state"
              :pending="updateSystemStateMutation.isPending.value"
              @toggle="toggleSystemState"
            />
            <button
              class="btn btn-error btn-soft"
              type="button"
              :disabled="cancelNextBellMutation.isPending.value"
              @click="requestCancelNextBell"
            >
              {{
                cancelNextBellMutation.isPending.value
                  ? "Cancelling..."
                  : "Cancel Next Bell"
              }}
            </button>
          </div>
        </div>
      </article>
    </div>

    <div class="grid gap-6 xl:grid-cols-[minmax(0,1fr)_minmax(0,1fr)]">
      <DashboardUpcomingList
        :items="limitedUpcomingItems"
        :loading="currentScheduleQuery.isLoading.value"
      />

      <DashboardRealtimeCard
        :socket-status="realtimeStore.socketStatus"
        :last-event-type="realtimeStore.lastEventType"
        :last-message-at="realtimeStore.lastMessageAt"
        :connected-clients="realtimeStore.connectedClients.length"
        :show-connected-clients="showConnectedClients"
      />
    </div>

    <AppConfirmDialog
      :open="showCancelDialog"
      title="Cancel Next Bell"
      description="Cancel the next pending bell event. Use this only when the next scheduled bell should not ring."
      confirm-label="Cancel Bell"
      tone="warning"
      :pending="cancelNextBellMutation.isPending.value"
      @close="closeCancelDialog"
      @confirm="handleCancelNextBell"
    />
  </section>
</template>
