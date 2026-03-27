import { useMutation, useQuery, useQueryClient } from "@tanstack/vue-query";
import {
  cancelNextBell,
  getSystemState,
  updateSystemState,
  type UpdateSystemStatePayload,
} from "../../../lib/api/system";
import { queryKeys } from "../../../lib/api/query-keys";

export function useSystemStateQuery() {
  return useQuery({
    queryFn: () => getSystemState(),
    queryKey: queryKeys.systemState(),
  });
}

export function useUpdateSystemStateMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: UpdateSystemStatePayload) =>
      updateSystemState(payload),
    onSuccess() {
      queryClient.invalidateQueries({ queryKey: queryKeys.systemState() });
    },
  });
}

export function useCancelNextBellMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => cancelNextBell(),
    onSuccess() {
      queryClient.invalidateQueries({ queryKey: queryKeys.systemState() });
    },
  });
}
