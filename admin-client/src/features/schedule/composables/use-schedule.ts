import { toValue, type MaybeRefOrGetter } from "vue";
import { useMutation, useQuery, useQueryClient } from "@tanstack/vue-query";
import {
  createScheduleItem,
  deleteScheduleItem,
  getCurrentSchedule,
  listScheduleItems,
  updateScheduleItem,
  type ScheduleListParams,
  type ScheduleWritePayload,
} from "../../../lib/api/schedule";
import { queryKeys } from "../../../lib/api/query-keys";

export function useScheduleItemsQuery(
  params: MaybeRefOrGetter<ScheduleListParams> = {},
) {
  return useQuery({
    queryFn: () => listScheduleItems(toValue(params)),
    queryKey: queryKeys.schedule(toValue(params)),
  });
}

export function useCurrentScheduleQuery() {
  return useQuery({
    queryFn: () => getCurrentSchedule(),
    queryKey: queryKeys.currentSchedule(),
  });
}

export function useCreateScheduleItemMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: ScheduleWritePayload) => createScheduleItem(payload),
    onSuccess() {
      queryClient.invalidateQueries({ queryKey: queryKeys.schedule() });
      queryClient.invalidateQueries({ queryKey: queryKeys.currentSchedule() });
    },
  });
}

export function useUpdateScheduleItemMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      id,
      payload,
    }: {
      id: string;
      payload: ScheduleWritePayload;
    }) => updateScheduleItem(id, payload),
    onSuccess() {
      queryClient.invalidateQueries({ queryKey: queryKeys.schedule() });
      queryClient.invalidateQueries({ queryKey: queryKeys.currentSchedule() });
    },
  });
}

export function useDeleteScheduleItemMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => deleteScheduleItem(id),
    onSuccess() {
      queryClient.invalidateQueries({ queryKey: queryKeys.schedule() });
      queryClient.invalidateQueries({ queryKey: queryKeys.currentSchedule() });
    },
  });
}
