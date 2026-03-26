<script setup lang="ts">
import { computed } from "vue";
import { Icon } from "@iconify/vue";
import type { SystemStateVm } from "../../../lib/api/system";
import type { SocketStatus } from "../../../stores/realtime";

const props = withDefaults(
  defineProps<{
    cancelPending?: boolean;
    connectedClients: number;
    state: null | SystemStateVm;
    statePending?: boolean;
    status: SocketStatus;
  }>(),
  {
    cancelPending: false,
    statePending: false,
  },
);

const emit = defineEmits<{
  cancelNextBell: [];
  setState: [state: SystemStateVm["state"]];
}>();

const formattedUpdatedAt = computed(() => {
  if (!props.state?.lastUpdated) {
    return "Waiting for system state...";
  }

  return new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "medium",
  }).format(new Date(props.state.lastUpdated));
});

const stateBadgeClass = computed(() => {
  return props.state?.state === "paused" ? "badge-warning" : "badge-success";
});

const socketBadgeClass = computed(() => {
  if (props.status === "OPEN") {
    return "badge-success";
  }

  if (props.status === "CONNECTING") {
    return "badge-warning";
  }

  return "badge-error";
});
</script>

<template>
  <article class="card border border-base-300 bg-base-100 shadow-sm">
    <div class="card-body gap-6">
      <div class="flex flex-wrap items-start justify-between gap-4">
        <div class="space-y-2">
          <p
            class="text-xs font-semibold uppercase tracking-[0.24em] text-primary"
          >
            Runtime State
          </p>
          <h2 class="font-display text-3xl font-semibold text-base-content">
            System Control
          </h2>
          <p class="max-w-2xl text-sm leading-7 text-base-content/70">
            Pause or resume bell playback, review connection health, and cancel
            the next pending bell when needed.
          </p>
        </div>

        <div class="flex flex-wrap items-center gap-2">
          <span class="badge badge-soft" :class="stateBadgeClass">
            {{ state?.state === "paused" ? "Paused" : "Active" }}
          </span>
          <span class="badge badge-soft" :class="socketBadgeClass">
            Socket {{ status.toLowerCase() }}
          </span>
        </div>
      </div>

      <div class="grid gap-4 md:grid-cols-3">
        <div class="rounded-box bg-base-200 p-4">
          <p class="text-sm text-base-content/65">Current state</p>
          <p class="mt-2 text-2xl font-semibold capitalize text-base-content">
            {{ state?.state ?? "Loading" }}
          </p>
        </div>
        <div class="rounded-box bg-base-200 p-4">
          <p class="text-sm text-base-content/65">Connected clients</p>
          <p class="mt-2 text-2xl font-semibold text-base-content">
            {{ connectedClients }}
          </p>
        </div>
        <div class="rounded-box bg-base-200 p-4">
          <p class="text-sm text-base-content/65">Last updated</p>
          <p class="mt-2 text-sm font-semibold leading-6 text-base-content">
            {{ formattedUpdatedAt }}
          </p>
        </div>
      </div>

      <div class="flex flex-wrap gap-3">
        <button
          class="btn btn-primary"
          type="button"
          :disabled="statePending || state?.state === 'active'"
          @click="emit('setState', 'active')"
        >
          <Icon icon="solar:play-bold-duotone" class="text-lg" />
          {{
            statePending && state?.state !== "active"
              ? "Updating..."
              : "Resume System"
          }}
        </button>
        <button
          class="btn btn-warning"
          type="button"
          :disabled="statePending || state?.state === 'paused'"
          @click="emit('setState', 'paused')"
        >
          <Icon icon="solar:pause-bold-duotone" class="text-lg" />
          {{
            statePending && state?.state !== "paused"
              ? "Updating..."
              : "Pause System"
          }}
        </button>
        <button
          class="btn btn-error btn-soft"
          type="button"
          :disabled="cancelPending"
          @click="emit('cancelNextBell')"
        >
          <Icon icon="solar:alarm-pause-bold-duotone" class="text-lg" />
          {{ cancelPending ? "Cancelling..." : "Cancel Next Bell" }}
        </button>
      </div>
    </div>
  </article>
</template>
