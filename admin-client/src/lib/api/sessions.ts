import {
  apiJsonApiCollection,
  apiJsonApiDocument,
  apiJsonApiWrite,
} from "./client";
import {
  adaptJsonApiCollection,
  adaptJsonApiResource,
} from "../jsonapi/adapters";
import type { JsonApiCollectionDocument, JsonApiDocument } from "./types";

export type SessionAttributes = {
  endTime: string;
  name: string;
  startTime: string;
};

export type SessionVm = {
  endTime: string;
  id: string;
  name: string;
  startTime: string;
};

export type SessionWritePayload = {
  data: {
    attributes: SessionAttributes;
    type: "sessions";
  };
};

export async function listSessions() {
  const document = await apiJsonApiCollection<SessionAttributes>("/sessions");

  return adaptJsonApiCollection(document, mapSessionResource);
}

export async function getCurrentSession() {
  const document =
    await apiJsonApiDocument<SessionAttributes>("/sessions/current");

  return adaptJsonApiResource(document, mapSessionResource);
}

export async function createSession(payload: SessionWritePayload) {
  return apiJsonApiWrite<JsonApiDocument<SessionAttributes>>("/sessions", {
    body: payload,
    method: "POST",
  });
}

export async function updateSession(id: string, payload: SessionWritePayload) {
  return apiJsonApiWrite<JsonApiDocument<SessionAttributes>>(
    `/sessions/${id}`,
    {
      body: payload,
      method: "PUT",
    },
  );
}

export async function deleteSession(id: string) {
  return apiJsonApiWrite<void>(`/sessions/${id}`, {
    method: "DELETE",
  });
}

function mapSessionResource(
  resource: JsonApiCollectionDocument<SessionAttributes>["data"][number],
): SessionVm {
  return {
    endTime: resource.attributes.endTime,
    id: resource.id,
    name: resource.attributes.name,
    startTime: resource.attributes.startTime,
  };
}
