import { useMutation, useQuery, useQueryClient } from "@tanstack/vue-query";
import {
  createSession,
  deleteSession,
  getCurrentSession,
  listSessions,
  updateSession,
  type SessionWritePayload,
} from "../../../lib/api/sessions";
import { queryKeys } from "../../../lib/api/query-keys";

export function useSessionsQuery() {
  return useQuery({
    queryFn: () => listSessions(),
    queryKey: queryKeys.sessions(),
  });
}

export function useCurrentSessionQuery() {
  return useQuery({
    queryFn: () => getCurrentSession(),
    queryKey: queryKeys.currentSession(),
  });
}

export function useCreateSessionMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: SessionWritePayload) => createSession(payload),
    onSuccess() {
      queryClient.invalidateQueries({ queryKey: queryKeys.sessions() });
      queryClient.invalidateQueries({ queryKey: queryKeys.currentSession() });
    },
  });
}

export function useUpdateSessionMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      id,
      payload,
    }: {
      id: string;
      payload: SessionWritePayload;
    }) => updateSession(id, payload),
    onSuccess() {
      queryClient.invalidateQueries({ queryKey: queryKeys.sessions() });
      queryClient.invalidateQueries({ queryKey: queryKeys.currentSession() });
    },
  });
}

export function useDeleteSessionMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => deleteSession(id),
    onSuccess() {
      queryClient.invalidateQueries({ queryKey: queryKeys.sessions() });
      queryClient.invalidateQueries({ queryKey: queryKeys.currentSession() });
    },
  });
}
