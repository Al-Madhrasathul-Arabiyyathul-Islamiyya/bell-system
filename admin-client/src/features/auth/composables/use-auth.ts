import { useMutation, useQueryClient } from "@tanstack/vue-query";
import {
  changePassword,
  login,
  logout,
  type ChangePasswordPayload,
  type LoginPayload,
} from "../../../lib/api/auth";
import { queryKeys } from "../../../lib/api/query-keys";
import { useAuthStore } from "../../../stores/auth";

export function useLoginMutation() {
  const authStore = useAuthStore();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: LoginPayload) => login(payload),
    onSuccess(response) {
      authStore.setAuth(response.token, {
        id: response.user.id ?? null,
        role: response.user.attributes.role ?? null,
        username: response.user.attributes.username ?? null,
      });
      queryClient.invalidateQueries({ queryKey: queryKeys.authUser() });
    },
  });
}

export function useLogoutMutation() {
  const authStore = useAuthStore();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => logout(),
    onSettled() {
      authStore.clearAuth();
      queryClient.removeQueries({ queryKey: queryKeys.authUser() });
    },
  });
}

export function useChangePasswordMutation() {
  return useMutation({
    mutationFn: (payload: ChangePasswordPayload) => changePassword(payload),
  });
}
