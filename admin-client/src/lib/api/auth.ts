import { apiJson } from "./client";
import type { AuthUser, JsonApiResource } from "./types";

export type LoginPayload = {
  password: string;
  username: string;
};

export type LoginResponse = {
  token: string;
  user: JsonApiResource<{
    createdAt?: string;
    role?: AuthUser["role"];
    username?: string;
  }>;
};

export type ChangePasswordPayload = {
  newPassword: string;
  oldPassword: string;
};

export async function login(payload: LoginPayload) {
  return apiJson<LoginResponse>("/auth/login", {
    body: payload,
    method: "POST" as const,
    requestType: "json",
  });
}

export async function logout() {
  return apiJson<void>("/auth/logout", {
    method: "POST" as const,
    requestType: "json",
  });
}

export async function changePassword(payload: ChangePasswordPayload) {
  return apiJson<void>("/auth/change-password", {
    body: payload,
    method: "POST" as const,
    requestType: "json",
  });
}
