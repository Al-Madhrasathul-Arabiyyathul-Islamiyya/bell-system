import { computed } from "vue";
import { getUserFacingError } from "../../../lib/api/errors";
import type { SystemStateVm } from "../../../lib/api/system";
import { useToastStore } from "../../../stores/toast";
import {
  useCancelNextBellMutation,
  useSystemStateQuery,
  useUpdateSystemStateMutation,
} from "./use-system";

export function useSystemControls() {
  const toastStore = useToastStore();
  const systemStateQuery = useSystemStateQuery();
  const updateSystemStateMutation = useUpdateSystemStateMutation();
  const cancelNextBellMutation = useCancelNextBellMutation();

  const systemState = computed(() => systemStateQuery.data.value ?? null);
  const systemError = computed(() => {
    if (!systemStateQuery.error.value) {
      return null;
    }

    return getUserFacingError(systemStateQuery.error.value).detail;
  });

  async function setSystemState(state: SystemStateVm["state"]) {
    try {
      await updateSystemStateMutation.mutateAsync({
        data: {
          attributes: {
            state,
          },
          type: "system-state",
        },
      });

      toastStore.enqueue({
        detail:
          state === "paused"
            ? "Bell playback has been paused."
            : "Bell playback has been resumed.",
        title: state === "paused" ? "System Paused" : "System Resumed",
        tone: "success",
      });
    } catch (error) {
      toastStore.enqueue({
        detail: getUserFacingError(error).detail,
        title: "State Update Failed",
        tone: "error",
      });
    }
  }

  async function toggleSystemState() {
    const nextState =
      systemState.value?.state === "paused" ? "active" : "paused";

    await setSystemState(nextState);
  }

  async function cancelNextBell() {
    try {
      await cancelNextBellMutation.mutateAsync();
      toastStore.enqueue({
        detail: "The next pending bell was cancelled successfully.",
        title: "Bell Cancelled",
        tone: "success",
      });
    } catch (error) {
      const userError = getUserFacingError(error);

      toastStore.enqueue({
        detail:
          userError.status === 404
            ? "No pending bell is available to cancel right now."
            : userError.detail,
        title: userError.status === 404 ? "Nothing to Cancel" : "Action Failed",
        tone: userError.status === 404 ? "info" : "error",
      });
      throw error;
    }
  }

  return {
    cancelNextBell,
    cancelNextBellMutation,
    setSystemState,
    systemError,
    systemState,
    systemStateQuery,
    toggleSystemState,
    updateSystemStateMutation,
  };
}
