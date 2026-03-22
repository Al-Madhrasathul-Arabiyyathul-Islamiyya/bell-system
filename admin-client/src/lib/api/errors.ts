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

function isFetchError(error: unknown): error is FetchError {
  return error instanceof Error && "response" in error;
}

function isJsonApiErrorDocument(value: unknown): value is JsonApiErrorDocument {
  return isObject(value) && Array.isArray(value.errors);
}
