<script setup lang="ts">
import { computed } from "vue";
import { Icon } from "@iconify/vue";
import { useTimeAgo } from "@vueuse/core";
import type { SocketStatus } from "../../../stores/realtime";

const props = defineProps<{
  connectedClients: number;
  lastEventType: null | string;
  lastMessageAt: null | string;
  showConnectedClients?: boolean;
  socketStatus: SocketStatus;
}>();

const lastMessageTimeAgo = useTimeAgo(
  computed(() => props.lastMessageAt ?? new Date().toISOString()),
);

const socketBadgeClass = computed(() => {
  if (props.socketStatus === "OPEN") {
    return "badge-success";
  }

  if (props.socketStatus === "CONNECTING") {
    return "badge-warning";
  }

  return "badge-error";
});
</script>

<template>
  <article class="card border border-base-300 bg-base-100 shadow-sm">
    <div class="card-body gap-5">
      <div class="flex items-start justify-between gap-4">
        <div class="space-y-2">
          <p
            class="text-xs font-semibold uppercase tracking-[0.24em] text-primary"
          >
            Realtime
          </p>
          <h2 class="font-display text-3xl font-semibold text-base-content">
            Live Connectivity
          </h2>
          <p class="text-sm leading-7 text-base-content/70">
            Websocket health and the most recent realtime activity observed by
            the admin client.
          </p>
        </div>
        <span class="badge badge-soft" :class="socketBadgeClass">
          Socket {{ socketStatus.toLowerCase() }}
        </span>
      </div>

      <div
        class="grid gap-4"
        :class="showConnectedClients ? 'md:grid-cols-3' : 'md:grid-cols-2'"
      >
        <div class="rounded-box bg-base-200 p-4">
          <p class="text-sm text-base-content/65">Last event</p>
          <p class="mt-2 text-lg font-semibold text-base-content">
            {{ lastEventType ?? "No events yet" }}
          </p>
        </div>
        <div class="rounded-box bg-base-200 p-4">
          <p class="text-sm text-base-content/65">Last message</p>
          <p class="mt-2 text-lg font-semibold text-base-content">
            {{ lastMessageAt ? lastMessageTimeAgo : "Waiting for events" }}
          </p>
        </div>
        <div v-if="showConnectedClients" class="rounded-box bg-base-200 p-4">
          <p class="text-sm text-base-content/65">Connected clients</p>
          <div class="mt-2 flex items-center gap-3">
            <Icon icon="solar:devices-bold-duotone" class="text-2xl" />
            <p class="text-lg font-semibold text-base-content">
              {{ connectedClients }}
            </p>
          </div>
        </div>
      </div>
    </div>
  </article>
</template>
