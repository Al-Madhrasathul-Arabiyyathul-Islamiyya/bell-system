import { toValue, type MaybeRefOrGetter } from "vue";
import { useMutation, useQuery, useQueryClient } from "@tanstack/vue-query";
import {
  createUser,
  deleteUser,
  getUser,
  listUsers,
  updateUser,
  type UserWritePayload,
  type UsersQueryParams,
} from "../../../lib/api/users";
import { queryKeys } from "../../../lib/api/query-keys";

export function useUsersQuery(params: MaybeRefOrGetter<UsersQueryParams> = {}) {
  return useQuery({
    queryFn: () => listUsers(toValue(params)),
    queryKey: queryKeys.users(toValue(params)),
  });
}

export function useUserQuery(id: string) {
  return useQuery({
    enabled: Boolean(id),
    queryFn: () => getUser(id),
    queryKey: [...queryKeys.users(), "detail", id] as const,
  });
}

export function useCreateUserMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: UserWritePayload) => createUser(payload),
    onSuccess() {
      queryClient.invalidateQueries({ queryKey: ["users"] });
    },
  });
}

export function useUpdateUserMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: UserWritePayload }) =>
      updateUser(id, payload),
    onSuccess(_, variables) {
      queryClient.invalidateQueries({ queryKey: ["users"] });
      queryClient.invalidateQueries({
        queryKey: [...queryKeys.users(), "detail", variables.id] as const,
      });
    },
  });
}

export function useDeleteUserMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => deleteUser(id),
    onSuccess() {
      queryClient.invalidateQueries({ queryKey: ["users"] });
    },
  });
}
