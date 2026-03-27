import { createFetch, type FetchContext, type FetchOptions } from "ofetch";
import { appEnv } from "../env";
import { pinia } from "../../app/pinia";
import { useAuthStore } from "../../stores/auth";
import type {
  JsonApiCollectionDocument,
  JsonApiDocument,
  JsonApiErrorDocument,
} from "./types";

type RequestMode = "binary" | "json" | "jsonapi" | "multipart";

const apiFetch = createFetch({
  defaults: {
    baseURL: appEnv.apiBaseUrl,
    onRequest({ options }: FetchContext) {
      applyAuthHeader(options);
    },
    onResponseError({ response }: FetchContext) {
      const authStore = useAuthStore(pinia);

      if (response?.status === 401 && authStore.isAuthenticated) {
        authStore.clearAuth();
        redirectToLoginIfNeeded();
      }
    },
    retry: 0,
  },
});

export type ApiRequestOptions = Omit<FetchOptions, "baseURL" | "headers"> & {
  headers?: HeadersInit;
  requestType?: RequestMode;
};

export async function apiJson<TResponse>(
  path: string,
  options: ApiRequestOptions = {},
) {
  return apiFetch<TResponse>(
    path,
    withMode(options, options.requestType ?? "json") as FetchOptions<"json">,
  );
}

export async function apiJsonApiDocument<TAttributes>(
  path: string,
  options: ApiRequestOptions = {},
) {
  return apiFetch<JsonApiDocument<TAttributes>>(
    path,
    withMode(options, "jsonapi") as FetchOptions<"json">,
  );
}

export async function apiJsonApiCollection<TAttributes>(
  path: string,
  options: ApiRequestOptions = {},
) {
  return apiFetch<JsonApiCollectionDocument<TAttributes>>(
    path,
    withMode(options, "jsonapi") as FetchOptions<"json">,
  );
}

export async function apiJsonApiWrite<
  TResponse = JsonApiDocument | JsonApiErrorDocument,
>(path: string, options: ApiRequestOptions = {}) {
  return apiFetch<TResponse>(
    path,
    withMode(options, "jsonapi") as FetchOptions<"json">,
  );
}

export async function apiMultipart<TResponse>(
  path: string,
  options: ApiRequestOptions = {},
) {
  return apiFetch<TResponse>(
    path,
    withMode(options, "multipart") as FetchOptions<"json">,
  );
}

export async function apiBinary(path: string, options: ApiRequestOptions = {}) {
  return apiFetch<Blob, "blob">(path, {
    ...(withMode(options, "binary") as FetchOptions<"blob">),
    responseType: "blob",
  });
}

function applyAuthHeader(options: FetchOptions) {
  const authStore = useAuthStore(pinia);

  if (!authStore.token) {
    return;
  }

  const headers = new Headers(options.headers);
  headers.set("Authorization", `Bearer ${authStore.token}`);
  options.headers = headers;
}

function withMode(options: ApiRequestOptions, mode: RequestMode) {
  const headers = new Headers(options.headers);

  if (mode === "jsonapi") {
    headers.set("Accept", "application/vnd.api+json");

    if (options.method && options.method !== "GET") {
      headers.set("Content-Type", "application/vnd.api+json");
    }
  } else if (mode === "json") {
    headers.set("Accept", "application/json");

    if (
      options.method &&
      options.method !== "GET" &&
      !(options.body instanceof FormData)
    ) {
      headers.set("Content-Type", "application/json");
    }
  } else if (mode === "multipart") {
    headers.set("Accept", "application/json");
    headers.delete("Content-Type");
  } else if (mode === "binary") {
    headers.set("Accept", "*/*");
  }

  return {
    ...options,
    headers,
  };
}

function redirectToLoginIfNeeded() {
  if (typeof window === "undefined") {
    return;
  }

  if (window.location.pathname === "/login") {
    return;
  }

  const redirectTarget = `${window.location.pathname}${window.location.search}${window.location.hash}`;
  const nextUrl = new URL("/login", window.location.origin);

  nextUrl.searchParams.set("reason", "session-expired");
  nextUrl.searchParams.set("redirect", redirectTarget);

  window.location.replace(nextUrl.toString());
}
