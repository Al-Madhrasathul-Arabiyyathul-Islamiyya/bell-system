<script setup lang="ts">
import { computed, ref, useTemplateRef } from "vue";
import { AppPageHeader, AppStatCard } from "./page-exports";
import AppConfirmDialog from "../components/app/AppConfirmDialog.vue";
import ChangePasswordCard from "../features/system/components/ChangePasswordCard.vue";
import SystemStateCard from "../features/system/components/SystemStateCard.vue";
import { useChangePasswordMutation } from "../features/auth/composables/use-auth";
import { getUserFacingError } from "../lib/api/errors";
import type { SystemStateVm } from "../lib/api/system";
import { useSystemControls } from "../features/system/composables/use-system-controls";
import { useAuthStore } from "../stores/auth";
import { useRealtimeStore } from "../stores/realtime";
import { useToastStore } from "../stores/toast";

const authStore = useAuthStore();
const toastStore = useToastStore();
const realtimeStore = useRealtimeStore();

const changePasswordCard = useTemplateRef<InstanceType<
  typeof ChangePasswordCard
> | null>("changePasswordCard");

const passwordError = ref<null | string>(null);
const showCancelDialog = ref(false);

const changePasswordMutation = useChangePasswordMutation();
const {
  cancelNextBell,
  cancelNextBellMutation,
  setSystemState,
  systemError,
  systemState,
  updateSystemStateMutation,
} = useSystemControls();
const showConnectedClients = computed(() => authStore.isAdmin);

async function handleSetState(state: SystemStateVm["state"]) {
  await setSystemState(state);
}

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
    // Toast feedback is already handled inside the shared system controls composable.
  } finally {
    showCancelDialog.value = false;
  }
}

async function handleChangePassword(payload: {
  newPassword: string;
  oldPassword: string;
}) {
  passwordError.value = null;

  try {
    await changePasswordMutation.mutateAsync(payload);
    changePasswordCard.value?.reset();
    toastStore.enqueue({
      detail: "Your password was updated successfully.",
      title: "Password Changed",
      tone: "success",
    });
  } catch (error) {
    passwordError.value = getUserFacingError(error).detail;
  }
}
</script>

<template>
  <section class="page-stack">
    <AppPageHeader
      eyebrow="Operations"
      title="System"
      description="Control bell playback, review current connection health, and manage the password for the current signed-in account."
    />

    <div
      class="grid gap-4"
      :class="showConnectedClients ? 'xl:grid-cols-3' : 'xl:grid-cols-2'"
    >
      <AppStatCard
        title="System State"
        icon="solar:settings-bold-duotone"
        :value="
          systemState?.state === 'paused'
            ? 'Paused'
            : systemState
              ? 'Active'
              : 'Loading'
        "
        description="Current scheduler state returned by the backend."
      />
      <AppStatCard
        title="Socket Status"
        icon="solar:wireless-charge-bold-duotone"
        :value="realtimeStore.socketStatus"
        description="Realtime admin websocket connection health."
      />
      <AppStatCard
        v-if="showConnectedClients"
        title="Connected Clients"
        icon="solar:devices-bold-duotone"
        :value="String(realtimeStore.connectedClients.length)"
        description="Clients currently connected to the websocket hub."
      />
    </div>

    <div v-if="systemError" class="alert alert-error text-sm" role="alert">
      {{ systemError }}
    </div>

    <div class="grid gap-6 xl:grid-cols-[minmax(0,1.2fr)_minmax(0,0.8fr)]">
      <SystemStateCard
        :state="systemState"
        :status="realtimeStore.socketStatus"
        :show-connected-clients="showConnectedClients"
        :connected-clients="realtimeStore.connectedClients.length"
        :state-pending="updateSystemStateMutation.isPending.value"
        :cancel-pending="cancelNextBellMutation.isPending.value"
        @set-state="handleSetState"
        @cancel-next-bell="requestCancelNextBell"
      />

      <ChangePasswordCard
        ref="changePasswordCard"
        :pending="changePasswordMutation.isPending.value"
        :error="passwordError"
        @save="handleChangePassword"
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
