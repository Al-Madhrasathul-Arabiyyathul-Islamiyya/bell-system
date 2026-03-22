import { computed, watch } from "vue";
import { useWebSocket } from "@vueuse/core";
import { useQueryClient } from "@tanstack/vue-query";
import { appEnv } from "../../../lib/env";
import { queryKeys } from "../../../lib/api/query-keys";
import { useAuthStore } from "../../../stores/auth";
import {
  useRealtimeStore,
  type ConnectedClient,
  type SocketStatus,
} from "../../../stores/realtime";

type SocketMessage =
  | {
      clients?: ConnectedClient[];
      type: "connected_clients";
    }
  | {
      type: "schedules_updated";
    }
  | {
      type: "audio_files_updated";
    }
  | {
      type: "system_state_changed";
    }
  | {
      message?: string;
      type: "system_log";
    }
  | {
      type: string;
      [key: string]: unknown;
    };

function toSocketStatus(status: string): SocketStatus {
  if (status === "OPEN") {
    return "OPEN";
  }

  if (status === "CONNECTING") {
    return "CONNECTING";
  }

  return "CLOSED";
}

export function useBellSystemSocket() {
  const authStore = useAuthStore();
  const realtimeStore = useRealtimeStore();
  const queryClient = useQueryClient();

  const socketUrl = computed(() => {
    const url = new URL(appEnv.wsBaseUrl);
    url.searchParams.set("client_name", "admin-client");
    url.searchParams.set("client_type", "admin");

    if (authStore.token) {
      url.searchParams.set("token", authStore.token);
    }

    return url.toString();
  });

  const socket = useWebSocket(socketUrl, {
    autoReconnect: {
      delay: 2_000,
      retries: 10,
    },
    autoClose: true,
    immediate: Boolean(authStore.token),
    onConnected() {
      realtimeStore.setSocketStatus("OPEN");
    },
    onDisconnected() {
      realtimeStore.setSocketStatus("CLOSED");
    },
    onError() {
      realtimeStore.setSocketStatus("CLOSED");
    },
    onMessage(_, event) {
      const message = parseSocketMessage(event.data);

      if (!message) {
        return;
      }

      realtimeStore.setLastEvent(message.type);

      if (message.type === "connected_clients") {
        realtimeStore.setConnectedClients(
          Array.isArray(message.clients) ? message.clients : [],
        );
        return;
      }

      if (message.type === "schedules_updated") {
        queryClient.invalidateQueries({ queryKey: queryKeys.schedule() });
        queryClient.invalidateQueries({
          queryKey: queryKeys.currentSchedule(),
        });
        return;
      }

      if (message.type === "audio_files_updated") {
        queryClient.invalidateQueries({ queryKey: queryKeys.audio() });
        queryClient.invalidateQueries({ queryKey: queryKeys.audioChecksums() });
        return;
      }

      if (message.type === "system_state_changed") {
        queryClient.invalidateQueries({ queryKey: queryKeys.systemState() });
        return;
      }

      if (
        message.type === "system_log" &&
        typeof message.message === "string"
      ) {
        realtimeStore.appendLog(message.message);
      }
    },
  });

  watch(socket.status, (status) => {
    realtimeStore.setSocketStatus(toSocketStatus(status));
  });

  return socket;
}

function parseSocketMessage(data: string): null | SocketMessage {
  try {
    return JSON.parse(data) as SocketMessage;
  } catch {
    return null;
  }
}
