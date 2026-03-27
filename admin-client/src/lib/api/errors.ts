import type { FetchError } from "ofetch";
import type { AppError, JsonApiErrorDocument } from "./types";

function isObject(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

function readString(value: unknown, fallback: string) {
  return typeof value === "string" && value.length > 0 ? value : fallback;
}

export function toAppError(error: unknown): AppError {
  if (isFetchError(error)) {
    return extractFetchError(error);
  }

  if (error instanceof Error) {
    return {
      detail: error.message,
      title: "Unexpected Error",
    };
  }

  return {
    detail: "An unexpected error occurred.",
    title: "Unexpected Error",
  };
}

export function toAppErrors(error: unknown): AppError[] {
  if (isFetchError(error)) {
    const data = error.data;

    if (isJsonApiErrorDocument(data)) {
      return data.errors.map((entry) => ({
        ...entry,
        detail: readString(entry.detail, "The request could not be completed."),
        title: readString(entry.title, "Request Failed"),
      }));
    }
  }

  return [toAppError(error)];
}

export function getPrimaryError(error: unknown): AppError {
  return toAppErrors(error)[0];
}

export function getUserFacingError(error: unknown): AppError {
  return toUserFacingError(getPrimaryError(error));
}

export function toUserFacingError(error: AppError): AppError {
  const status = toStatusNumber(error.status);
  const detail = readString(
    error.detail,
    "The request could not be completed.",
  );

  if (status === 401) {
    return {
      ...error,
      detail: "Your session is no longer valid. Please sign in again.",
      title: "Sign In Required",
    };
  }

  if (status === 403) {
    return {
      ...error,
      detail: "You do not have permission to perform that action.",
      title: "Permission Denied",
    };
  }

  if (status === 404) {
    return {
      ...error,
      detail: "The requested item could not be found.",
      title: "Not Found",
    };
  }

  if (status === 409) {
    return {
      ...error,
      detail:
        "That action could not be completed because the data changed. Refresh and try again.",
      title: "Conflict Detected",
    };
  }

  if (status === 422) {
    return {
      ...error,
      detail: isUserFriendlyDetail(detail)
        ? detail
        : "Some of the submitted information needs to be corrected before continuing.",
      title: "Check Your Input",
    };
  }

  if (typeof status === "number" && status >= 500) {
    return {
      ...error,
      detail:
        "Something went wrong on the server. Please try again in a moment.",
      title: "Server Error",
    };
  }

  if (!isUserFriendlyDetail(detail)) {
    return {
      ...error,
      detail: "The request could not be completed. Please try again.",
      title: readString(error.title, "Request Failed"),
    };
  }

  return {
    ...error,
    detail,
    title: readString(error.title, "Request Failed"),
  };
}

function extractFetchError(error: FetchError): AppError {
  const status = error.response?.status;
  const data = error.data;

  if (isJsonApiErrorDocument(data) && data.errors.length > 0) {
    const entry = data.errors[0];

    return {
      ...entry,
      detail: readString(entry.detail, "The request could not be completed."),
      status: entry.status ?? status,
      title: readString(entry.title, "Request Failed"),
    };
  }

  if (isObject(data) && typeof data.error === "string") {
    return {
      detail: data.error,
      status,
      title: "Request Failed",
    };
  }

  return {
    detail: error.message || "The request could not be completed.",
    status,
    title: "Request Failed",
  };
}

function isUserFriendlyDetail(detail: string) {
  const value = detail.trim();
  const normalized = value.toLowerCase();

  if (value.length === 0 || value.length > 180) {
    return false;
  }

  const technicalMarkers = [
    "exception",
    "stack",
    "trace",
    "sql",
    "syntax",
    "constraint",
    "panic",
    "timeout",
    "timed out",
    "failed to",
    "invalid character",
    "unexpected token",
    "internal server error",
  ];

  return !technicalMarkers.some((marker) => normalized.includes(marker));
}

function toStatusNumber(status: number | string | undefined) {
  if (typeof status === "number") {
    return status;
  }

  if (typeof status === "string" && status.length > 0) {
    const parsed = Number(status);

    return Number.isNaN(parsed) ? undefined : parsed;
  }

  return undefined;
}

function isFetchError(error: unknown): error is FetchError {
  return error instanceof Error && "response" in error;
}

function isJsonApiErrorDocument(value: unknown): value is JsonApiErrorDocument {
  return isObject(value) && Array.isArray(value.errors);
}
