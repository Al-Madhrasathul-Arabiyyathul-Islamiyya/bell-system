import { computed, ref } from "vue";
import { defineStore } from "pinia";

export type SocketStatus = "CLOSED" | "CONNECTING" | "OPEN";

export type ConnectedClient = {
  clientName: string;
  clientType: string;
  connectedAt?: string;
  id?: string;
};

export const useRealtimeStore = defineStore("realtime", () => {
  const connectedClients = ref<ConnectedClient[]>([]);
  const lastEventType = ref<null | string>(null);
  const lastMessageAt = ref<null | string>(null);
  const logMessages = ref<string[]>([]);
  const socketStatus = ref<SocketStatus>("CLOSED");

  const isConnected = computed(() => socketStatus.value === "OPEN");

  function appendLog(message: string) {
    logMessages.value = [...logMessages.value.slice(-49), message];
  }

  function setConnectedClients(clients: ConnectedClient[]) {
    connectedClients.value = clients;
  }

  function setLastEvent(type: string) {
    lastEventType.value = type;
    lastMessageAt.value = new Date().toISOString();
  }

  function setSocketStatus(status: SocketStatus) {
    socketStatus.value = status;
  }

  return {
    appendLog,
    connectedClients,
    isConnected,
    lastEventType,
    lastMessageAt,
    logMessages,
    setConnectedClients,
    setLastEvent,
    setSocketStatus,
    socketStatus,
  };
});
